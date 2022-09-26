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

func (svc *serviceContext) finalizeUnitData(rawUnitID string) (*finalizeResponse, error) {
	uid := padLeft(rawUnitID, 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)
	log.Printf("INFO: finalize unit %s", unitDir)

	pendingFilesCnt := 0
	chunkSize := 20
	fileCommands := make(map[string][]string)
	channel := make(chan []updateProblem)
	outstandingRequests := 0

	unitMD, uErr := svc.getUnitMetadata(uid)
	if uErr != nil {
		return nil, uErr
	}

	// walk the unit directory and generate masterFile info for each .tif
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}\.tif$`)
	tifRegex := regexp.MustCompile(`^.*\.tif$`)
	hiddenRegex := regexp.MustCompile(`^\..*`)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if f.IsDir() == true {
			return nil
		}
		fName := f.Name()
		if hiddenRegex.Match([]byte(fName)) || !tifRegex.Match([]byte(fName)) || !mfRegex.Match([]byte(fName)) {
			return nil
		}

		cmd := make([]string, 0)
		id := unitMD.CallNumber
		if id == "" {
			id = unitMD.PID
		}
		cmd = append(cmd, fmt.Sprintf("-iptc:MasterDocumentID=%s", fmt.Sprintf("UVA Library: %s", id)))
		cmd = append(cmd, fmt.Sprintf("-iptc:ObjectName=%s", fName))
		cmd = append(cmd, fmt.Sprintf("-iptc:ClassifyState=%s", ""))
		cmd = append(cmd, path)
		fileCommands[path] = cmd
		pendingFilesCnt++
		if pendingFilesCnt == chunkSize {
			outstandingRequests++
			log.Printf("INFO: finalize %s batch #%d of %d images", uid, outstandingRequests, chunkSize)
			go batchUpdateExifData(fileCommands, channel)
			fileCommands = make(map[string][]string)
			pendingFilesCnt = 0
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if pendingFilesCnt > 0 {
		outstandingRequests++
		log.Printf("INFO: finalize %s batch #%d of %d images", uid, outstandingRequests, chunkSize)
		go batchUpdateExifData(fileCommands, channel)
	}

	log.Printf("INFO: await all finalization updates for %s", uid)
	var resp finalizeResponse
	resp.Success = true
	resp.Problems = make([]updateProblem, 0)
	for outstandingRequests > 0 {
		errs := <-channel
		outstandingRequests--
		if len(errs) > 0 {
			log.Printf("ERROR: finalize %s failed: %+v", uid, errs)
			resp.Success = false
			resp.Problems = append(resp.Problems, errs...)
		}
	}
	log.Printf("INFO: all finalization updates for %s are done", uid)

	return &resp, nil
}

func (svc *serviceContext) updateImageMetadata(c *gin.Context) {
	uid := padLeft(c.Param("uid"), 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)
	fileName := c.Param("file")
	updateField := c.Query("field")
	updateValue := c.Query("value")
	log.Printf("INFO: update %s/%s %s to %s", unitDir, fileName, updateField, updateValue)

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
	log.Printf("INFO: lookup exif tag for field %s", fieldName)
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
	log.Printf("INFO: field %s matches exif tag  %s", fieldName, exifTag)
	return exifTag
}

func cleanupExifToolDups(fullPath string) {
	dupPath := fmt.Sprintf("%s_original", fullPath)
	os.Remove(dupPath)
	dupPath = fmt.Sprintf("%s_exiftool_tmp", fullPath)
	os.Remove(dupPath)
}

func (svc *serviceContext) updateMetadataBatch(c *gin.Context) {
	uid := padLeft(c.Param("uid"), 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)

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
	log.Printf("INFO: batch master file metadata in %s", unitDir)

	start := time.Now()
	pendingFilesCnt := 0
	chunkSize := 10
	fileCommands := make(map[string][]string)
	channel := make(chan []updateProblem)
	outstandingRequests := 0
	for _, change := range mdPost {
		cmd := make([]string, 0)
		exifTag := getExifTag(change.Field)
		cmd = append(cmd, fmt.Sprintf("-%s=%s", exifTag, change.Value))
		cmd = append(cmd, change.File)
		fileCommands[change.File] = cmd
		pendingFilesCnt++
		if pendingFilesCnt == chunkSize {
			outstandingRequests++
			log.Printf("INFO: update batch #%d of %d images", outstandingRequests, chunkSize)
			go batchUpdateExifData(fileCommands, channel)
			fileCommands = make(map[string][]string)
			pendingFilesCnt = 0
		}
	}

	if pendingFilesCnt > 0 {
		outstandingRequests++
		log.Printf("INFO: update batch #%d of %d images", outstandingRequests, chunkSize)
		go batchUpdateExifData(fileCommands, channel)
	}

	// wait for all metadata updates to complete
	var resp struct {
		Success  bool            `json:"success"`
		Problems []updateProblem `json:"problems"`
	}
	resp.Success = true
	resp.Problems = make([]updateProblem, 0)
	log.Printf("INFO: await all metadata updates")
	for outstandingRequests > 0 {
		errs := <-channel
		outstandingRequests--
		if len(errs) > 0 {
			resp.Success = false
			resp.Problems = append(resp.Problems, errs...)
		}
	}

	elapsed := time.Since(start)
	elapsedMS := int64(elapsed / time.Millisecond)
	log.Printf("INFO: updated  %d masterfiles in %dms", len(mdPost), elapsedMS)

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) renameFiles(c *gin.Context) {
	unit := padLeft(c.Param("uid"), 9)
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
	unit := padLeft(c.Param("uid"), 9)
	file := c.Param("file")
	rotateDirString := c.Query("dir")
	if rotateDirString == "" {
		rotateDirString = "right"
	}
	rotateDir := "90"
	if rotateDirString == "left" {
		rotateDir = "-90"
	}

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
	_, err = exec.Command("convert", cmd...).Output()
	if err != nil {
		log.Printf("ERROR: unable to rotate %s: %s", fullPath, err.Error())
		c.String(http.StatusInternalServerError, fmt.Sprintf("unable to rotate file: %s", err.Error()))
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

// batchUpdateExifData is called as a goroutine. It will update a batch of files, then return true on the channel
func batchUpdateExifData(fileCommands map[string][]string, channel chan []updateProblem) {
	errors := make([]updateProblem, 0)
	for tgtFile, command := range fileCommands {
		_, err := exec.Command("exiftool", command...).Output()
		if err != nil {
			log.Printf("ERROR: unable to update %s metadata with %v: %s", tgtFile, command, err.Error())
			errors = append(errors, updateProblem{File: path.Base(tgtFile), Problem: err.Error()})
		} else {
			cleanupExifToolDups(tgtFile)
		}
	}

	// notify caller that the block is done
	channel <- errors
}

func checkExifHeaders(cmdArray []string, checkLocation bool, channel chan []qaCheck) {
	out := make([]qaCheck, 0)
	cmd := exec.Command("exiftool", cmdArray...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR: unable to get qa metadata: %s", err.Error())
		channel <- out
	}

	var parsed []exifData
	err = json.Unmarshal(stdout, &parsed)
	if err != nil {
		log.Printf("ERROR: unable to parse qa metadata: %s", err.Error())
		channel <- out
	}

	for _, md := range parsed {
		errors := make([]string, 0)

		title := ""
		if md.Title != nil {
			title = fmt.Sprintf("%v", md.Title)
		}
		if title == "" {
			log.Printf("ERROR: %s is missing a title", md.SourceFile)
			errors = append(errors, "Missing title metadata")
		}

		if checkLocation {
			if md.Box == "" {
				log.Printf("ERROR: %s is missing a box", md.SourceFile)
				errors = append(errors, "Missing box metadata")
			}
			if md.Folder == "" {
				log.Printf("ERROR: %s is missing a folder", md.SourceFile)
				errors = append(errors, "Missing folder metadata")
			}
		}

		tc := qaCheck{File: path.Base(md.SourceFile), Valid: len(errors) == 0, Errors: errors}
		out = append(out, tc)
	}
	channel <- out
}

// getExifData will retrieve a batch of metadata for masterfiles in a goroutine. The list of metadata is returned.
func asyncGetExifData(cmdArray []string, channel chan []masterFileMetadata) {
	channel <- getExifData(cmdArray)
}

// getExifData will retrieve a batch of metadata for masterfiles in a goroutine. The list of metadata is returned.
func getExifData(cmdArray []string) []masterFileMetadata {
	out := make([]masterFileMetadata, 0)
	cmd := exec.Command("exiftool", cmdArray...)
	// log.Printf("INFO: %v", cmd)
	stdout, err := cmd.Output()
	if err != nil {
		log.Printf("WARNINIG: unable to get image metadata: %s", err.Error())
	} else {
		var parsed []exifData
		err = json.Unmarshal(stdout, &parsed)
		if err != nil {
			log.Printf("WARNING: unable to parse metadata: %s", err.Error())
		} else {
			for _, md := range parsed {
				mdRec := masterFileMetadata{}
				mdRec.FileName = path.Base(fmt.Sprintf("%v", md.SourceFile))
				if md.Title != nil {
					mdRec.Title = fmt.Sprintf("%v", md.Title)
				}
				if md.Description != nil {
					mdRec.Description = fmt.Sprintf("%v", md.Description)
				}
				if md.Component != nil {
					mdRec.ComponentID = fmt.Sprintf("%v", md.Component)
				}
				if md.Box != nil {
					mdRec.Box = fmt.Sprintf("%v", md.Box)
				}
				if md.Folder != nil {
					mdRec.Folder = fmt.Sprintf("%v", md.Folder)
				}
				if md.Resolution != nil {
					valType := fmt.Sprintf("%T", md.Resolution)
					if valType == "int" {
						mdRec.Resolution = md.Resolution.(int)
					} else if valType == "float64" {
						fRes := md.Resolution.(float64)
						mdRec.Resolution = int(fRes)
					} else {
						log.Printf("WARN: unsupported resolution type %s", valType)
						mdRec.Resolution = 0
					}
				}
				mdRec.ColorProfile = md.ColorProfile
				mdRec.FileSize = md.FileSize
				mdRec.FileType = md.FileType
				mdRec.Width = md.Width
				mdRec.Height = md.Height
				mdRec.Status = md.ClassifyState

				out = append(out, mdRec)
			}
		}
	}

	return out
}

func baseExifCmd() []string {
	out := []string{"-json", "-ImageWidth", "-ImageHeight",
		"-FileType", "-XResolution", "-FileSize", "-icc_profile:ProfileDescription", "-iptc:OwnerID",
		"-iptc:headline", "-iptc:caption-abstract", "-iptc:ClassifyState",
		"-iptc:ContentLocationName", "-iptc:Keywords"}
	return out
}

func qaExifCmd() []string {
	out := []string{"-json", "-iptc:headline", "-iptc:ContentLocationName", "-iptc:Keywords"}
	return out
}
