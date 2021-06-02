package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type masterFileInfo struct {
	FileName     string `json:"fileName"`
	Path         string `json:"path"`
	ThumbURL     string `json:"thumbURL"`
	MediumURL    string `json:"mediumURL"`
	InfoURL      string `json:"infoURL"`
	ColorProfile string `json:"colorProfile"`
	FileSize     string `json:"fileSize"`
	FileType     string `json:"fileType"`
	Resolution   int    `json:"resolution"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type exifData struct {
	ColorProfile string `json:"ICCProfileName"`
	FileSize     string `json:"FileSize"`
	FileType     string `json:"FileType"`
	Resolution   int    `json:"XResolution"`
	Title        string `json:"Headline"`
	Description  string `json:"Caption-Abstract"`
	Width        int    `json:"ImageWidth"`
	Height       int    `json:"ImageHeight"`
}

func (svc *serviceContext) getUnits(c *gin.Context) {
	log.Printf("INFO: get available units from %s", svc.ImagesDir)
	files, err := ioutil.ReadDir(svc.ImagesDir)
	if err != nil {
		log.Printf("ERROR: unable to list contents of images directory: %s", err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	unitRegex := regexp.MustCompile(`^\d{9}$`)
	out := make([]string, 0)
	for _, f := range files {
		fName := f.Name()
		if unitRegex.Match([]byte(fName)) {
			out = append(out, fName)
		}
	}
	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) getMasterFiles(c *gin.Context) {
	dir := c.Param("dir")
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, dir)
	log.Printf("INFO: get master files from %s", unitDir)

	files, err := ioutil.ReadDir(unitDir)
	if err != nil {
		log.Printf("ERROR: unable to list contents of %s: %s", unitDir, err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}.tif$`)
	out := make([]*masterFileInfo, 0)
	for _, f := range files {
		fName := f.Name()
		if mfRegex.Match([]byte(fName)) {
			mf := masterFileInfo{FileName: fName, Path: unitDir}

			// FIXME FOR BOX/FOLDER STRUCTURE
			fName = strings.ReplaceAll(fName, ".tif", "")
			mf.ThumbURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/30,/0/default.jpg", svc.IIIFURL, dir, fName)
			mf.MediumURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/175,/0/default.jpg", svc.IIIFURL, dir, fName)
			mf.InfoURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/info.json", svc.IIIFURL, dir, fName)
			out = append(out, &mf)
		}
	}

	currIdx := 0
	pendingFilesCnt := 0
	cmdArray := baseExifCmd()
	for _, mf := range out {
		fullPath := fmt.Sprintf("%s/%s", mf.Path, mf.FileName)
		cmdArray = append(cmdArray, fullPath)
		pendingFilesCnt++
		if pendingFilesCnt == 10 {
			getExifData(cmdArray, out, currIdx)
			cmdArray = baseExifCmd()
			currIdx = pendingFilesCnt
			pendingFilesCnt = 0
		}
	}

	if pendingFilesCnt > 0 {
		getExifData(cmdArray, out, currIdx)
	}

	c.JSON(http.StatusOK, out)
}

func getExifData(cmdArray []string, files []*masterFileInfo, startIdx int) {
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
				files[currIdx].ColorProfile = md.ColorProfile
				files[currIdx].Description = md.Description
				files[currIdx].FileSize = md.FileSize
				files[currIdx].FileType = md.FileType
				files[currIdx].Resolution = md.Resolution
				files[currIdx].Title = md.Title
				files[currIdx].Width = md.Width
				files[currIdx].Height = md.Height
				currIdx++
			}
		}
	}
}

func baseExifCmd() []string {
	out := []string{"-json", "-ImageWidth", "-ImageHeight",
		"-FileType", "-XResolution", "-FileSize", "-ICCProfileName",
		"-iptc:headline", "-iptc:caption-abstract"}
	return out
}
