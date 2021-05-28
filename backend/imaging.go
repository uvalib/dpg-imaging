package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

type masterFileInfo struct {
	FileName  string `json:"fileName"`
	ThumbURL  string `json:"thumbURL"`
	MediumURL string `json:"mediumURL"`
}
type masterFiles struct {
	IIIFManifest string           `json:"iiifManifest"`
	Files        []masterFileInfo `json:"masterFiles"`
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

	mfRegex := regexp.MustCompile(`^\d{9}_\d{4}.tif$`)
	out := masterFiles{Files: make([]masterFileInfo, 0)}
	for _, f := range files {
		fName := f.Name()
		if mfRegex.Match([]byte(fName)) {
			mf := masterFileInfo{FileName: fName}
			out.Files = append(out.Files, mf)
		}
	}

	c.JSON(http.StatusOK, out)
}
