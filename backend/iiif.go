package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

type iiifMasterFile struct {
	FileName string
	Width    int
	Height   int
}

type iiifData struct {
	IIIFURL     string
	URL         string
	Unit        string
	MasterFiles []iiifMasterFile
}

func (svc *serviceContext) getIIIFManifest(c *gin.Context) {
	dir := c.Param("unit")
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, dir)
	log.Printf("INFO: get IIIF manifest from %s", unitDir)
	_, err := os.Stat(unitDir)
	if os.IsNotExist(err) {
		log.Printf("INFO: %s does not exist", dir)
		c.String(http.StatusNotFound, fmt.Sprintf("%s not found", dir))
		return
	}

	files, err := ioutil.ReadDir(unitDir)
	if err != nil {
		log.Printf("ERROR: unable to list contents of %s: %s", unitDir, err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	mfRegex := regexp.MustCompile(`^\d{9}_\d{4}.tif$`)
	iiifManURL := fmt.Sprintf("%s/api/iiif/%s", svc.ServiceURL, dir)
	data := iiifData{URL: iiifManURL, IIIFURL: svc.IIIFURL, Unit: dir}
	for _, f := range files {
		fName := f.Name()
		if mfRegex.Match([]byte(fName)) {
			fullPath := fmt.Sprintf("%s/%s", unitDir, fName)
			fName = strings.ReplaceAll(fName, ".tif", "")
			mf := iiifMasterFile{FileName: fName}

			stdout, err := exec.Command("exiftool", "-json", "-ImageWidth", "-ImageHeight", fullPath).Output()
			if err != nil {
				log.Printf("WARNINIG: unable to get width/height of %s:%s", fullPath, err.Error())
			} else {
				var parsed []struct {
					Source string `json:"SourceFile"`
					Width  int    `json:"ImageWidth"`
					Height int    `json:"ImageHeight"`
				}
				err = json.Unmarshal(stdout, &parsed)
				if err != nil {
					log.Printf("WARNING: unable to parse exiftool response: %s", err.Error())
				} else {
					mf.Width = parsed[0].Width
					mf.Height = parsed[0].Height
				}
			}
			data.MasterFiles = append(data.MasterFiles, mf)
		}
	}

	var outBuffer bytes.Buffer
	tplErr := svc.IIIFManTemplate.Execute(&outBuffer, data)
	if tplErr != nil {
		log.Printf("ERROR: unable to render IIIF metadata for %s: %s", dir, tplErr.Error())
		c.String(http.StatusInternalServerError, tplErr.Error())
		return
	}

	c.Header("content-type", "application/json; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.String(http.StatusOK, outBuffer.String())
}
