package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type ocrHint struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OCRCandidate bool   `json:"ocrCandidate"`
}

type ocrLanguageHint struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type metadata struct {
	ID              uint    `json:"id"`
	PID             string  `gorm:"column:pid" json:"pid"`
	CallNumber      string  `json:"callNumber,omitempty"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	OCRHintID       uint    `json:"-"`
	OCRHint         ocrHint `gorm:"foreignKey:OCRHintID" json:"ocrHint"`
	OCRLanguageHint string  `json:"ocrLanguageHint"`
}

type intendedUse struct {
	ID                    uint   `json:"id"`
	Description           string `json:"description"`
	DeliverableFormat     string `json:"deliverableFormat"`
	DeliverableResolution string `json:"deliverableResolution"`
}

type agency struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type customer struct {
	ID        uint   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type order struct {
	ID         uint     `json:"id"`
	CustomerID uint     `json:"customerID"`
	Customer   customer `gorm:"foreignKey:CustomerID" json:"customer"`
}

type unit struct {
	ID                  uint        `json:"id"`
	OrderID             uint        `json:"orderID"`
	Order               order       `gorm:"foreignKey:OrderID" json:"order"`
	MetadataID          uint        `json:"-"`
	Metadata            metadata    `gorm:"foreignKey:MetadataID" json:"metadata"`
	IntendedUseID       uint        `json:"-"`
	IntendedUse         intendedUse `gorm:"foreignKey:IntendedUseID" json:"intendedUse"`
	SpecialInstructions string      `json:"specialInstructions,omitempty"`
	OCRMasterFiles      bool        `json:"ocrMasterFiles"`
	UnitStatus          string      `json:"status"`
}

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
	Status       string `json:"status"`
	ComponentID  string `json:"componentID"`
}

type unitMasterfiles struct {
	MasterFiles []*masterFileInfo `json:"masterFiles"`
	Problems    []string          `json:"problems"`
}

func (svc *serviceContext) getUnitMasterFiles(c *gin.Context) {
	uidStr := padLeft(c.Param("uid"), 9)
	unitDir := path.Join(svc.ImagesDir, uidStr)
	out := unitMasterfiles{MasterFiles: make([]*masterFileInfo, 0), Problems: make([]string, 0)}
	start := time.Now()

	log.Printf("INFO: get master files from %s", unitDir)

	// walk the unit directory and generate masterFile info for each .tif
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}\.tif$`)
	tifRegex := regexp.MustCompile(`^.*\.tif$`)
	hiddenRegex := regexp.MustCompile(`^\..*`)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("WARNING: directory traverse failed: %s", err.Error())
			return nil
		}

		if f.IsDir() == false {
			fName := f.Name()
			if hiddenRegex.Match([]byte(fName)) {
				log.Printf("INFO: skipping hidden file %s", fName)
				return nil
			}
			log.Printf("INFO: Found file %s", path)

			if !tifRegex.Match([]byte(fName)) {
				log.Printf("INFO: %s is not a tif; skipping", fName)
				out.Problems = append(out.Problems, fmt.Sprintf("%s is not an image file", path))
				return nil
			}

			// NOTE: path param is the full path to the tif - including the filename
			// Strip the base image dir and file name to get relative path
			relPath := strings.Replace(path, fmt.Sprintf("%s/", svc.ImagesDir), "", 1)
			relPath = strings.Replace(relPath, fmt.Sprintf("/%s", fName), "", 1)

			if !mfRegex.Match([]byte(fName)) {
				log.Printf("INFO: %s is named incorrectly", fName)
				out.Problems = append(out.Problems, fmt.Sprintf("%s is named incorrectly", path))
			}

			mf := masterFileInfo{FileName: fName, Path: path}
			fName = strings.ReplaceAll(fName, ".tif", "")

			if strings.Split(fName, "_")[0] != uidStr {
				out.Problems = append(out.Problems, fmt.Sprintf("%s doesn't match unit number", path))
			}

			pathID := strings.Replace(relPath, "/", "%2F", -1)
			mf.ThumbURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/!30,45/0/default.jpg", svc.IIIFURL, pathID, fName)
			mf.MediumURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/!250,375/0/default.jpg", svc.IIIFURL, pathID, fName)
			mf.LargeURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/!400,600/0/default.jpg", svc.IIIFURL, pathID, fName)
			mf.InfoURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/info.json", svc.IIIFURL, pathID, fName)

			out.MasterFiles = append(out.MasterFiles, &mf)
			return nil

		}

		return nil
	})

	if err != nil {
		log.Printf("ERROR: unable to walk contents of %s: %s", unitDir, err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	if len(out.MasterFiles) > 0 {
		lastMF := out.MasterFiles[len(out.MasterFiles)-1]
		lastSeq := strings.Split(lastMF.FileName, "_")[1]
		lastSeq = strings.Replace(lastSeq, ".tif", "", 1)
		lastSeqNum, _ := strconv.Atoi(lastSeq)
		if len(out.MasterFiles) != lastSeqNum {
			out.Problems = append(out.Problems, "Last image number doesn't match count")
		}
	} else {
		out.Problems = append(out.Problems, "No images found")
	}

	// use exiftool to get metadata for master files in batches
	currIdx := 0
	pendingFilesCnt := 0
	chunkSize := 20
	cmdArray := baseExifCmd()
	channel := make(chan bool)
	outstandingRequests := 0
	for _, mf := range out.MasterFiles {
		cmdArray = append(cmdArray, mf.Path)
		pendingFilesCnt++
		if pendingFilesCnt == chunkSize {
			outstandingRequests++
			log.Printf("INFO: get batch #%d of %d exif metadata starting at %d", outstandingRequests, chunkSize, currIdx)
			go getExifData(cmdArray, out.MasterFiles, currIdx, channel)
			cmdArray = baseExifCmd()
			pendingFilesCnt = 0
			currIdx += chunkSize
		}
	}

	if pendingFilesCnt > 0 {
		log.Printf("INFO: get batch #%d of %d exif metadata starting at %d", outstandingRequests, chunkSize, currIdx)
		outstandingRequests++
		go getExifData(cmdArray, out.MasterFiles, currIdx, channel)
	}

	// wait for all metadata to complete
	log.Printf("INFO: await all metadata")
	for outstandingRequests > 0 {
		<-channel
		outstandingRequests--
	}

	elapsed := time.Since(start)
	elapsedMS := int64(elapsed / time.Millisecond)
	log.Printf("INFO: got %d masterfiles for unit %s in %dms", len(out.MasterFiles), uidStr, elapsedMS)

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) getUnitMetadata(uid string) (*metadata, error) {
	log.Printf("INFO: get metadata for unit %s", uid)
	var md metadata
	resp := svc.DB.Preload(clause.Associations).Joins("inner join units on units.metadata_id = metadata.id").Where("units.id=?", uid).First(&md)
	if resp.Error != nil {
		return nil, fmt.Errorf("unable to get metadata for unit %s: %s", uid, resp.Error.Error())
	}

	return &md, nil
}
