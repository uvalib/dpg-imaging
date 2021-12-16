package main

import (
	"encoding/json"
	"fmt"
	"io"
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

type exifData struct {
	ColorProfile  string      `json:"ProfileDescription"`
	FileSize      string      `json:"FileSize"`
	FileType      string      `json:"FileType"`
	Resolution    interface{} `json:"XResolution"`
	Title         interface{} `json:"Headline"`
	Description   interface{} `json:"Caption-Abstract"`
	Width         int         `json:"ImageWidth"`
	Height        int         `json:"ImageHeight"`
	ClassifyState string      `json:"ClassifyState"`
	OwnerID       interface{} `json:"OwnerID"`
}

type exifTitle struct {
	SourceFile string      `json:"SourceFile"`
	Headline   interface{} `json:"Headline"`
}

type updateProblem struct {
	File    string `json:"file"`
	Problem string `json:"problem"`
}

type titleCheck struct {
	File  string `json:"file"`
	Valid bool   `json:"valid"`
}

func (svc *serviceContext) finalizeUnitRequest(c *gin.Context) {
	uid := c.Param("uid")
	start := time.Now()
	resp, err := svc.finalizeUnitData(uid)
	if err != nil {
		log.Printf("ERROR: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	elapsed := time.Since(start)
	elapsedMS := int64(elapsed / time.Millisecond)

	log.Printf("INFO: finalized  unit %s in %dms", uid, elapsedMS)
	c.JSON(http.StatusOK, resp)
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
			go updateExifData(fileCommands, channel)
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
		go updateExifData(fileCommands, channel)
	}

	log.Printf("INFO: await all finalization updates")
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

	return &resp, nil
}

func (svc *serviceContext) updateMetadata(c *gin.Context) {
	uid := padLeft(c.Param("uid"), 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)
	log.Printf("INFO: update master file metadata in %s", unitDir)

	var mdPost []struct {
		File        string `json:"file"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		ComponentID string `json:"componentID"`
	}

	qpErr := c.ShouldBindJSON(&mdPost)
	if qpErr != nil {
		log.Printf("ERROR: invalid updateMetadata payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	start := time.Now()
	pendingFilesCnt := 0
	chunkSize := 10
	fileCommands := make(map[string][]string)
	channel := make(chan []updateProblem)
	outstandingRequests := 0
	for _, change := range mdPost {
		cmd := make([]string, 0)
		cmd = append(cmd, fmt.Sprintf("-iptc:headline=%s", change.Title))
		cmd = append(cmd, fmt.Sprintf("-iptc:caption-abstract=%s", change.Description))
		cmd = append(cmd, fmt.Sprintf("-iptc:ClassifyState=%s", change.Status))
		cmd = append(cmd, fmt.Sprintf("-iptc:OwnerID=%s", change.ComponentID))
		cmd = append(cmd, change.File)
		fileCommands[change.File] = cmd
		pendingFilesCnt++
		if pendingFilesCnt == chunkSize {
			outstandingRequests++
			log.Printf("INFO: update batch #%d of %d images", outstandingRequests, chunkSize)
			go updateExifData(fileCommands, channel)
			fileCommands = make(map[string][]string)
			pendingFilesCnt = 0
		}
	}

	if pendingFilesCnt > 0 {
		outstandingRequests++
		log.Printf("INFO: update batch #%d of %d images", outstandingRequests, chunkSize)
		go updateExifData(fileCommands, channel)
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

func (svc *serviceContext) deleteFile(c *gin.Context) {
	unit := padLeft(c.Param("uid"), 9)
	file := c.Param("file")
	unitDir := path.Join(svc.ImagesDir, unit)
	log.Printf("INFO: delete %s from %s", file, unitDir)
	delFilePath := ""
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err == nil {
			if f.IsDir() == false && f.Name() == file {
				delFilePath = path
				return io.EOF
			}
		}
		return nil
	})
	if err != nil && err != io.EOF {
		log.Printf("ERROR: unable to traverse %s for file delete: %s", unitDir, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if delFilePath == "" {
		log.Printf("ERROR: unable to delete %s; file not found", file)
		c.String(http.StatusNotFound, fmt.Sprintf("unable to delete %s: file not found", file))
		return

	}
	log.Printf("INFO: delete %s", delFilePath)
	err = os.Remove(delFilePath)
	if err != nil {
		log.Printf("ERROR: unable to delete %s: %s", delFilePath, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("deleted %s", delFilePath))
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
	fullPath := path.Join(svc.ImagesDir, unit, file)

	// grab the currrent data inthe exif heassers ad the rotate command wipes it all
	cmdArray := []string{"-json", "-iptc:OwnerID", "-iptc:headline", "-iptc:caption-abstract", "-iptc:ClassifyState", fullPath}
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
	if origMD[0].OwnerID != nil {
		cmd = append(cmd, fmt.Sprintf("-iptc:OwnerID=%v", origMD[0].OwnerID))
	}
	if origMD[0].Title != nil {
		cmd = append(cmd, fmt.Sprintf("-iptc:headline=%v", origMD[0].Title))
	}
	if origMD[0].Description != nil {
		cmd = append(cmd, fmt.Sprintf("-iptc:caption-abstract=%v", origMD[0].Description))
	}
	cmd = append(cmd, fmt.Sprintf("-iptc:ClassifyState=%v", origMD[0].ClassifyState))
	cmd = append(cmd, fullPath)
	_, err = exec.Command("exiftool", cmd...).Output()
	if err != nil {
		log.Printf("WARNING: unable to restore metadata to %s after rotation: %s", file, err.Error())
	} else {
		dupPath := fmt.Sprintf("%s_original", fullPath)
		os.Remove(dupPath)
	}

	c.String(http.StatusOK, "rotated")
}

// updateExifData is called as a goroutine. It will update a batch of files, then return true on the channel
func updateExifData(fileCommands map[string][]string, channel chan []updateProblem) {
	errors := make([]updateProblem, 0)
	for tgtFile, command := range fileCommands {
		_, err := exec.Command("exiftool", command...).Output()
		if err != nil {
			log.Printf("ERROR: unable to update %s metadata with [exiftool %v]: %s", tgtFile, command, err.Error())
			errors = append(errors, updateProblem{File: path.Base(tgtFile), Problem: err.Error()})
		} else {
			dupPath := fmt.Sprintf("%s_original", tgtFile)
			os.Remove(dupPath)
			dupPath = fmt.Sprintf("%s_exiftool_tmp", tgtFile)
			os.Remove(dupPath)
		}
	}

	// notify caller that the block is done
	channel <- errors
}

func checkExifHeaders(cmdArray []string, channel chan []titleCheck) {
	log.Printf("INFO: check exif header for %d files", len(cmdArray))
	out := make([]titleCheck, 0)
	stdout, err := exec.Command("exiftool", cmdArray...).Output()
	if err != nil {
		log.Printf("ERROR: unable to get title metadata: %s", err.Error())
		channel <- out
	}
	var parsed []exifTitle
	err = json.Unmarshal(stdout, &parsed)
	if err != nil {
		log.Printf("ERROR: unable to parse title metadata: %s", err.Error())
		channel <- out
	}
	for _, md := range parsed {
		valid := false
		if md.Headline != nil {
			check := fmt.Sprintf("%v", md.Headline)
			if check != "" {
				valid = true
			} else {
				log.Printf("INFO: %s is missing a title", md.SourceFile)
			}
		}
		tc := titleCheck{File: md.SourceFile, Valid: valid}
		out = append(out, tc)
	}
	channel <- out
}

// getExifData will retrieve a batch of metadata for masterfiles in a goroutine. When complete return true
func getExifData(cmdArray []string, files []*masterFileInfo, startIdx int, channel chan bool) {
	currIdx := startIdx
	stdout, err := exec.Command("exiftool", cmdArray...).Output()
	if err != nil {
		log.Printf("WARNINIG: unable to get image metadata: %s", err.Error())
	} else {
		var parsed []exifData
		err = json.Unmarshal(stdout, &parsed)
		if err != nil {
			log.Printf("WARNING: unable to parse metadata: %s", err.Error())
		} else {
			for _, md := range parsed {
				if md.Title != nil {
					files[currIdx].Title = fmt.Sprintf("%v", md.Title)
				}
				if md.Description != nil {
					files[currIdx].Description = fmt.Sprintf("%v", md.Description)
				}
				if md.OwnerID != nil {
					files[currIdx].ComponentID = fmt.Sprintf("%v", md.OwnerID)
				}
				if md.Resolution != nil {
					valType := fmt.Sprintf("%T", md.Resolution)
					if valType == "int" {
						files[currIdx].Resolution = md.Resolution.(int)
					} else if valType == "float64" {
						fRes := md.Resolution.(float64)
						files[currIdx].Resolution = int(fRes)
					} else {
						log.Printf("WARN: unsupported resolution type %s", valType)
						files[currIdx].Resolution = 0
					}
				}
				files[currIdx].ColorProfile = md.ColorProfile
				files[currIdx].FileSize = md.FileSize
				files[currIdx].FileType = md.FileType
				files[currIdx].Width = md.Width
				files[currIdx].Height = md.Height
				files[currIdx].Status = md.ClassifyState
				currIdx++
			}
		}
	}

	// notify caller that the block is done
	channel <- true
}

func baseExifCmd() []string {
	out := []string{"-json", "-ImageWidth", "-ImageHeight",
		"-FileType", "-XResolution", "-FileSize", "-icc_profile:ProfileDescription", "-iptc:OwnerID",
		"-iptc:headline", "-iptc:caption-abstract", "-iptc:ClassifyState"}
	return out
}

func titleExifCmd() []string {
	out := []string{"-json", "-iptc:headline"}
	return out
}
