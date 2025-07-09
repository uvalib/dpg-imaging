package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type finalizeResponse struct {
	Success  bool            `json:"success"`
	Problems []updateProblem `json:"problems"`
}

type exifMapping struct {
	FieldName string `json:"field"`
	ExifTag   string `json:"exifTag"`
}

type exifData struct {
	SourceFile    string      `json:"SourceFile"`
	ColorProfile  string      `json:"ProfileDescription"`
	FileSize      string      `json:"FileSize"`
	FileType      string      `json:"FileType"`
	Resolution    interface{} `json:"XResolution"`
	Title         interface{} `json:"Headline"`
	Description   interface{} `json:"Caption-Abstract"`
	Width         int         `json:"ImageWidth"`
	Height        int         `json:"ImageHeight"`
	ClassifyState string      `json:"ClassifyState"`       // tag
	Box           interface{} `json:"Keywords"`            // box
	Folder        interface{} `json:"ContentLocationName"` // folder
	Component     interface{} `json:"OwnerID"`             // component
}

type updateProblem struct {
	File    string `json:"file"`
	Problem string `json:"problem"`
}

type qaCheck struct {
	File   string   `json:"file"`
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

type exifFileCommands struct {
	File     string
	Commands []string
}

func (svc *serviceContext) cleanupImageFilenames(c *gin.Context) {
	uid := padLeft(c.Param("uid"), 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)

	log.Printf("INFO: rename or remove  .tif_orignal files in %s", unitDir)
	renameCnt := 0
	originalTifRegex := regexp.MustCompile(`^.*\.tif_original$`)
	err := filepath.Walk(unitDir, func(fullPath string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("ERROR: directory %s traverse failed: %s", unitDir, err.Error())
			return nil
		}

		if f.IsDir() {
			log.Printf("INFO: directory %s", f.Name())
			return nil
		}

		fName := f.Name()
		if originalTifRegex.Match([]byte(fName)) {
			fixedName := strings.Replace(fName, "tif_original", "tif", 1)
			fixedFullPath := path.Join(unitDir, fixedName)
			log.Printf("INFO: rename %s to %s", fullPath, fixedFullPath)
			if exists(fixedFullPath) {
				log.Printf("ERROR: cannot rename %s: %s already exists", fName, fixedFullPath)
				return nil
			}
			err = os.Rename(fullPath, fixedFullPath)
			if err != nil {
				log.Printf("ERROR: unable to rename %s: %s", fullPath, err.Error())
				return nil
			}
			renameCnt++
		}

		return nil
	})
	if err != nil {
		log.Printf("ERROR: unable to walk contents of %s: %s", unitDir, err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	log.Printf("INFO: renamed %d files", renameCnt)
	c.String(http.StatusOK, "ok")
}

func (svc *serviceContext) updateImageMetadata(c *gin.Context) {
	rawUnitID := c.Param("uid")
	uid := padLeft(rawUnitID, 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)
	fileName := c.Param("file")
	updateField := c.Query("field")
	updateValue := c.Query("value")
	log.Printf("INFO: update %s/%s %s to %s", unitDir, fileName, updateField, updateValue)

	if svc.isBatchInProgress(rawUnitID) {
		log.Printf("WARNING: request to update metadata for file from unit %s rejected because it is already being processed", rawUnitID)
		c.String(http.StatusConflict, "this unit is currently being processed by another user")
		return
	}

	svc.addBatchProcess(rawUnitID)
	defer svc.removeBatchProcess(rawUnitID)

	exifTag := getExifTag(updateField)
	if exifTag == "" {
		log.Printf("ERROR: invalid update field %s", updateField)
		c.String(http.StatusBadRequest, fmt.Sprintf("%s is not a valid update field", updateField))
		return
	}

	tgtFile := path.Join(unitDir, fileName)
	cmd := []string{fmt.Sprintf("-%s=%s", exifTag, updateValue), tgtFile}
	log.Printf("INFO: update command %v", cmd)
	_, err := exec.Command("exiftool", cmd...).Output()
	if err != nil {
		log.Printf("ERROR: exiftool %v failed: %s", cmd, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	cleanupExifToolDups(tgtFile)

	c.String(http.StatusOK, "ok")
}

func getExifTag(fieldName string) string {
	// log.Printf("INFO: lookup exif tag for field %s", fieldName)
	fields := []exifMapping{{FieldName: "title", ExifTag: "iptc:headline"}, {FieldName: "description", ExifTag: "iptc:caption-abstract"},
		{FieldName: "box", ExifTag: "iptc:Keywords"}, {FieldName: "folder", ExifTag: "iptc:ContentLocationName"},
		{FieldName: "tag", ExifTag: "iptc:ClassifyState"}, {FieldName: "component", ExifTag: "iptc:OwnerID"},
	}
	exifTag := ""
	for _, fv := range fields {
		if fieldName == fv.FieldName {
			exifTag = fv.ExifTag
			break
		}
	}
	// log.Printf("INFO: field %s matches exif tag  %s", fieldName, exifTag)
	return exifTag
}

func cleanupExifToolDups(fullPath string) {
	dupPath := fmt.Sprintf("%s_original", fullPath)
	os.Remove(dupPath)
	dupPath = fmt.Sprintf("%s_exiftool_tmp", fullPath)
	os.Remove(dupPath)
}

func (svc *serviceContext) updateMetadataBatch(c *gin.Context) {
	rawUnitID := c.Param("uid")
	uid := padLeft(rawUnitID, 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)
	if svc.isBatchInProgress(rawUnitID) {
		log.Printf("WARNING: request to batch update unit %s rejected because it is already being processed", rawUnitID)
		c.String(http.StatusConflict, "this unit is currently being processed by another user")
		return
	}

	var mdPost []struct {
		File  string `json:"file"`
		Field string `json:"field"`
		Value string `json:"value"`
	}

	qpErr := c.ShouldBindJSON(&mdPost)
	if qpErr != nil {
		log.Printf("ERROR: invalid updateMetadata payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	start := time.Now()
	commandsBatch := make([]exifFileCommands, 0)
	errChannel := make(chan updateProblem)
	var updateWG sync.WaitGroup
	svc.addBatchProcess(rawUnitID)
	defer svc.removeBatchProcess(rawUnitID)

	log.Printf("INFO: batch master file metadata in %s with batch size %d", unitDir, svc.BatchSize)
	for _, change := range mdPost {
		cmd := exifFileCommands{File: change.File, Commands: make([]string, 0)}
		exifTag := getExifTag(change.Field)
		cmd.Commands = append(cmd.Commands, fmt.Sprintf("-%s=%s", exifTag, change.Value))
		cmd.Commands = append(cmd.Commands, change.File)
		commandsBatch = append(commandsBatch, cmd)
		if len(commandsBatch) == svc.BatchSize {
			updateWG.Add(1)
			cmdCopy := make([]exifFileCommands, len(commandsBatch))
			copy(cmdCopy, commandsBatch)
			go func() {
				defer updateWG.Done()
				batchUpdateExifData(cmdCopy, errChannel)
			}()
			commandsBatch = make([]exifFileCommands, 0)
		}
	}

	if len(commandsBatch) > 0 {
		updateWG.Add(1)
		go func() {
			defer updateWG.Done()
			batchUpdateExifData(commandsBatch, errChannel)
		}()
	}

	var resp struct {
		Success  bool            `json:"success"`
		Problems []updateProblem `json:"problems"`
	}
	resp.Success = true
	resp.Problems = make([]updateProblem, 0)

	go func() {
		log.Printf("INFO: await all metadata updates and collect any problems in the response")
		updateWG.Wait()
		close(errChannel)
		log.Printf("INFO: all batches complete")
	}()

	for problem := range errChannel {
		resp.Success = false
		resp.Problems = append(resp.Problems, problem)
	}

	elapsed := time.Since(start)
	elapsedMS := int64(elapsed / time.Millisecond)
	log.Printf("INFO: updated  %d masterfiles in %dms", len(mdPost), elapsedMS)

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) renameFiles(c *gin.Context) {
	rawUnitID := c.Param("uid")
	unit := padLeft(rawUnitID, 9)
	if svc.isBatchInProgress(rawUnitID) {
		log.Printf("WARNING: request to rename files for unit %s rejected because it is already being processed", rawUnitID)
		c.String(http.StatusConflict, "this unit is currently being processed by another user")
		return
	}

	var rnPost []struct {
		Original string `json:"original"`
		NewName  string `json:"new"`
	}
	log.Printf("INFO: rename files for unit %s", unit)

	qpErr := c.ShouldBindJSON(&rnPost)
	if qpErr != nil {
		log.Printf("ERROR: invalid rename payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	// create a working dir in the root of the unit directory to hold the original files to be renamed
	svc.addBatchProcess(rawUnitID)
	defer svc.removeBatchProcess(rawUnitID)
	backUpDir := path.Join(svc.ImagesDir, "tmp", unit)
	_, existErr := os.Stat(backUpDir)
	if existErr == nil {
		log.Printf("INFO: working directory %s already exists, removing it", backUpDir)
		err := os.RemoveAll(backUpDir)
		if err != nil {
			log.Printf("ERROR: unable to remove %s: %s", backUpDir, err.Error())
			c.String(http.StatusInternalServerError, "unable to cleanup old working directory")
			return
		}
	}
	err := os.Mkdir(backUpDir, 0777)
	if err != nil {
		log.Printf("ERROR: unable to make backup dir %s: %s", backUpDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// do this is two passes to avoid conflicts... move the files to tmp with new name first
	for _, rn := range rnPost {
		tmpFile := path.Join(backUpDir, rn.NewName)
		log.Printf("INFO: rename %s to %s", rn.Original, tmpFile)
		err := os.Rename(rn.Original, tmpFile)
		if err != nil {
			log.Printf("ERROR: unable to rename %s: %s", rn.Original, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// second pass moves the renamed files back where they belong
	for _, rn := range rnPost {
		tmpFile := path.Join(backUpDir, rn.NewName)
		origDir, _ := path.Split(rn.Original)
		renamed := path.Join(origDir, rn.NewName)

		/// if renamed already exists, something is wrong! leave as-is and abort before files are lost or overwritten
		_, existErr := os.Stat(renamed)
		if existErr == nil {
			log.Printf("ERROR: renamed file %s already exists. Abort to avoid data loss", renamed)
			c.String(http.StatusInternalServerError, fmt.Sprintf("Renamed file %s already exists! Rename aborted. Manual corrections are required.", renamed))
			return
		}

		log.Printf("INFO: move tmp %s to %s", tmpFile, renamed)
		err := os.Rename(tmpFile, renamed)
		if err != nil {
			log.Printf("ERROR: unable to restore %s from %s: %s", renamed, tmpFile, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// last, cleanup tmp
	log.Printf("INFO: cleaning up working dorectory %s", backUpDir)
	err = os.RemoveAll(backUpDir)
	if err != nil {
		log.Printf("ERROR: unable to clean up %s: %s", backUpDir, err.Error())
	}

	c.String(http.StatusOK, "ok")
}

func (svc *serviceContext) rotateFile(c *gin.Context) {
	rawUnitID := c.Param("uid")
	unit := padLeft(rawUnitID, 9)
	file := c.Param("file")
	rotateDirString := c.Query("dir")
	if rotateDirString == "" {
		rotateDirString = "right"
	}
	rotateDir := "90"
	if rotateDirString == "left" {
		rotateDir = "-90"
	}

	if svc.isBatchInProgress(rawUnitID) {
		log.Printf("WARNING: request to rotate file from unit %s rejected because it is already being processed", rawUnitID)
		c.String(http.StatusConflict, "this unit is currently being processed by another user")
		return
	}

	svc.addBatchProcess(rawUnitID)
	defer svc.removeBatchProcess(rawUnitID)

	basePath := path.Join(svc.ImagesDir, unit)
	log.Printf("INFO: looking for image %s in unit dir %s", file, basePath)
	fullPath := findFile(basePath, file)
	if fullPath == "" {
		log.Printf("ERROR: unable to find %s found in unit dir %s", file, basePath)
		c.String(http.StatusBadRequest, fmt.Sprintf("%s not found", file))
		return
	}

	// grab the currrent data inthe exif headers ad the rotate command wipes it all
	cmdArray := []string{"-json", "-iptc:OwnerID", "-iptc:headline", "-iptc:caption-abstract",
		"-iptc:ClassifyState", "-iptc:ContentLocationName", "-iptc:Keywords", fullPath}
	stdout, err := exec.Command("exiftool", cmdArray...).Output()
	if err != nil {
		log.Printf("ERROR: unable to get %s metadata before rotation: %s", file, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	var origMD []exifData
	err = json.Unmarshal(stdout, &origMD)
	if err != nil {
		log.Printf("ERROR: unable to parse %s metadata before rotation: %s", file, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: rotate %s", fullPath)
	cmd := []string{fullPath, "-rotate", rotateDir, fullPath}
	rotateOut, err := exec.Command("magick", cmd...).CombinedOutput()
	if err != nil {
		log.Printf("ERROR: unable to rotate %s: %s", fullPath, rotateOut)
		c.String(http.StatusInternalServerError, fmt.Sprintf("unable to rotate file: %s", rotateOut))
		return
	}

	log.Printf("INFO: restoring metadata after rotaion of %s", file)
	cmd = make([]string, 0)
	if origMD[0].Component != nil {
		cmd = append(cmd, fmt.Sprintf("-iptc:OwnerID=%v", origMD[0].Component))
	}
	if origMD[0].Title != nil {
		cmd = append(cmd, fmt.Sprintf("-iptc:headline=%v", origMD[0].Title))
	}
	if origMD[0].Description != nil {
		cmd = append(cmd, fmt.Sprintf("-iptc:caption-abstract=%v", origMD[0].Description))
	}
	cmd = append(cmd, fmt.Sprintf("-iptc:ClassifyState=%v", origMD[0].ClassifyState))
	cmd = append(cmd, fmt.Sprintf("-iptc:ContentLocationName=%v", origMD[0].Folder))
	cmd = append(cmd, fmt.Sprintf("-iptc:Keywords=%v", origMD[0].Box))
	cmd = append(cmd, fullPath)
	_, err = exec.Command("exiftool", cmd...).Output()
	if err != nil {
		log.Printf("WARNING: unable to restore metadata to %s after rotation: %s", file, err.Error())
	} else {
		cleanupExifToolDups(fullPath)
	}

	c.String(http.StatusOK, "rotated")
}

func batchUpdateExifData(fileCommands []exifFileCommands, channel chan updateProblem) {
	log.Printf("INFO: start batch of %d update commands", len(fileCommands))
	startTime := time.Now()
	for _, fc := range fileCommands {
		cmd := exec.Command("exiftool", fc.Commands...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("ERROR: unable to update %s metadata with %v: %s", fc.File, cmd, out)
			channel <- updateProblem{File: path.Base(fc.File), Problem: err.Error()}
		} else {
			cleanupExifToolDups(fc.File)
		}
	}
	elapsed := time.Since(startTime)
	log.Printf("INFO: batch of %d update commands has finished in %d ms", len(fileCommands), elapsed.Milliseconds())
}

func checkExifHeaders(files []string, checkLocation bool, channel chan updateProblem) {
	log.Printf("INFO: start batch of %d validate commands", len(files))
	startTime := time.Now()
	cmdArray := []string{"-json", "-iptc:headline", "-iptc:ContentLocationName", "-iptc:Keywords"}
	cmdArray = append(cmdArray, files...)
	cmd := exec.Command("exiftool", cmdArray...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR: unable to get qa metadata: %s", err.Error())
		channel <- updateProblem{File: "all", Problem: err.Error()}
	} else {

		var parsed []exifData
		json.Unmarshal(stdout, &parsed)
		for _, md := range parsed {
			title := ""
			if md.Title != nil {
				title = fmt.Sprintf("%v", md.Title)
			}
			if title == "" {
				log.Printf("ERROR: %s is missing a title", md.SourceFile)
				channel <- updateProblem{File: md.SourceFile, Problem: "Missing title metadata"}
			}

			if checkLocation {
				if md.Box == "" {
					log.Printf("ERROR: %s is missing a box", md.SourceFile)
					channel <- updateProblem{File: md.SourceFile, Problem: "Missing box metadata"}
				}
				if md.Folder == "" {
					log.Printf("ERROR: %s is missing a folder", md.SourceFile)
					channel <- updateProblem{File: md.SourceFile, Problem: "Missing folder metadata"}
				}
			}
		}
	}
	elapsed := time.Since(startTime)
	log.Printf("INFO: batch of %d validate commands has finished in %d ms", len(files), elapsed.Milliseconds())
}

func getExifMetadataBatch(tgtFiles []string, channel chan masterFileMetadata) {
	log.Printf("INFO: start get metadata batch of %d files", len(tgtFiles))
	startTime := time.Now()
	cmdArray := baseExifCmd()
	cmdArray = append(cmdArray, tgtFiles...)
	cmd := exec.Command("exiftool", cmdArray...)
	cmdOut, err := cmd.Output()
	if err != nil {
		log.Printf("WARNINIG: unable to get image metadata: %s", err.Error())
		return
	}

	var parsed []exifData
	json.Unmarshal(cmdOut, &parsed)
	for _, exifMD := range parsed {
		mdRec := parseExifResponse(&exifMD)
		channel <- mdRec
	}

	elapsed := time.Since(startTime)
	log.Printf("INFO: get metadata batch of %d files has finished in %d ms", len(tgtFiles), elapsed.Milliseconds())
}

func getExifData(tgtFile string) (*masterFileMetadata, error) {
	log.Printf("INFO: get exif metadata for %s", tgtFile)
	cmdArray := baseExifCmd()
	cmdArray = append(cmdArray, tgtFile)
	cmd := exec.Command("exiftool", cmdArray...)

	cmdOut, err := cmd.Output()
	if err != nil {
		log.Printf("WARNINIG: unable to get image metadata: %s", err.Error())
		return nil, fmt.Errorf("%s", cmdOut)
	}

	var parsed []exifData
	json.Unmarshal(cmdOut, &parsed)
	md := parsed[0]
	mdRec := parseExifResponse(&md)

	return &mdRec, nil
}

func parseExifResponse(exifMD *exifData) masterFileMetadata {
	mdRec := masterFileMetadata{}
	mdRec.FileName = path.Base(fmt.Sprintf("%v", exifMD.SourceFile))
	if exifMD.Title != nil {
		mdRec.Title = fmt.Sprintf("%v", exifMD.Title)
	}
	if exifMD.Description != nil {
		mdRec.Description = fmt.Sprintf("%v", exifMD.Description)
	}
	if exifMD.Component != nil {
		mdRec.ComponentID = fmt.Sprintf("%v", exifMD.Component)
	}
	if exifMD.Box != nil {
		mdRec.Box = fmt.Sprintf("%v", exifMD.Box)
	}
	if exifMD.Folder != nil {
		mdRec.Folder = fmt.Sprintf("%v", exifMD.Folder)
	}
	if exifMD.Resolution != nil {
		valType := fmt.Sprintf("%T", exifMD.Resolution)
		switch valType {
		case "int":
			mdRec.Resolution = exifMD.Resolution.(int)
		case "float64":
			fRes := exifMD.Resolution.(float64)
			mdRec.Resolution = int(fRes)
		default:
			log.Printf("WARN: unsupported resolution type %s", valType)
			mdRec.Resolution = 0
		}
	}
	mdRec.ColorProfile = exifMD.ColorProfile
	mdRec.FileSize = exifMD.FileSize
	mdRec.FileType = exifMD.FileType
	mdRec.Width = exifMD.Width
	mdRec.Height = exifMD.Height
	mdRec.Status = exifMD.ClassifyState
	return mdRec
}

func baseExifCmd() []string {
	out := []string{"-json", "-ImageWidth", "-ImageHeight",
		"-FileType", "-XResolution", "-FileSize", "-icc_profile:ProfileDescription", "-iptc:OwnerID",
		"-iptc:headline", "-iptc:caption-abstract", "-iptc:ClassifyState",
		"-iptc:ContentLocationName", "-iptc:Keywords"}
	return out
}
