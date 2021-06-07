package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type masterFileInfo struct {
	FileName     string `json:"fileName"`
	Path         string `json:"path"`
	ThumbURL     string `json:"thumbURL"`
	MediumURL    string `json:"mediumURL"`
	LargeURL     string `json:"largeURL"`
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
	ColorProfile string      `json:"ICCProfileName"`
	FileSize     string      `json:"FileSize"`
	FileType     string      `json:"FileType"`
	Resolution   int         `json:"XResolution"`
	Title        interface{} `json:"Headline"`
	Description  interface{} `json:"Caption-Abstract"`
	Width        int         `json:"ImageWidth"`
	Height       int         `json:"ImageHeight"`
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

	// walk the unit directory and generate masterFile info for each .tif
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}.tif$`)
	out := make([]*masterFileInfo, 0)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("ERROR: directory traverse failed: %s", err.Error())
			return err
		}
		if f.IsDir() == false {
			log.Printf("INFO: Found .tif image %s", path)

			// NOTE: path param is the full path to the tif - including the filename
			// Strip the base image dir and file name to get relative path
			fName := f.Name()
			relPath := strings.Replace(path, fmt.Sprintf("%s/", svc.ImagesDir), "", 1)
			relPath = strings.Replace(relPath, fmt.Sprintf("/%s", fName), "", 1)

			if mfRegex.Match([]byte(fName)) {
				mf := masterFileInfo{FileName: fName, Path: path}
				fName = strings.ReplaceAll(fName, ".tif", "")

				pathID := strings.Replace(relPath, "/", "%2F", -1)
				mf.ThumbURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/30,/0/default.jpg", svc.IIIFURL, pathID, fName)
				mf.MediumURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/250,/0/default.jpg", svc.IIIFURL, pathID, fName)
				mf.LargeURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/400,/0/default.jpg", svc.IIIFURL, pathID, fName)
				mf.InfoURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/info.json", svc.IIIFURL, pathID, fName)

				out = append(out, &mf)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("ERROR: unable to walk contents of %s: %s", unitDir, err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	// use exiftool to get metadata for master files in batches of 10
	currIdx := 0
	pendingFilesCnt := 0
	cmdArray := baseExifCmd()
	for _, mf := range out {
		cmdArray = append(cmdArray, mf.Path)
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

func (svc *serviceContext) updateMetadata(c *gin.Context) {
	dir := c.Param("dir")
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, dir)
	log.Printf("INFO: update master file metadata in %s", unitDir)

	var mdPost []struct {
		File        string `json:"file"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	qpErr := c.ShouldBindJSON(&mdPost)
	if qpErr != nil {
		log.Printf("ERROR: invalid updateMetadata payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	for _, change := range mdPost {
		titleEdit := fmt.Sprintf("-iptc:headline=%s", change.Title)
		descEdit := fmt.Sprintf("-iptc:caption-abstract=%s", change.Description)
		log.Printf("INFO: exiftool %s %s %s", titleEdit, descEdit, change.File)
		_, err := exec.Command("exiftool", titleEdit, descEdit, change.File).Output()
		if err != nil {
			log.Printf("ERROR: unable to update %s metadata: %s", change.File, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		dupPath := fmt.Sprintf("%s_original", change.File)
		removeErr := os.Remove(dupPath)
		if removeErr != nil {
			log.Printf("WARNING: unable to remove backup file %s:%s", dupPath, removeErr.Error())
		}
	}

	c.String(http.StatusOK, "updated")
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
				title := ""
				if md.Title != nil {
					title = fmt.Sprintf("%v", md.Title)
				}
				desc := ""
				if md.Description != nil {
					desc = fmt.Sprintf("%v", md.Description)
				}
				files[currIdx].ColorProfile = md.ColorProfile
				files[currIdx].Description = desc
				files[currIdx].FileSize = md.FileSize
				files[currIdx].FileType = md.FileType
				files[currIdx].Resolution = md.Resolution
				files[currIdx].Title = title
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
