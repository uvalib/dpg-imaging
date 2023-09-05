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
	"sync"
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
	ID         uint      `json:"id"`
	CustomerID uint      `json:"customerID"`
	Customer   customer  `gorm:"foreignKey:CustomerID" json:"customer"`
	AgencyID   uint      `json:"-"`
	Agency     agency    `gorm:"foreignKey:AgencyID" json:"agency"`
	DateDue    time.Time `json:"dateDue"`
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
	FileName  string `json:"fileName"`
	Path      string `json:"path"`
	ThumbURL  string `json:"thumbURL"`
	MediumURL string `json:"mediumURL"`
	LargeURL  string `json:"largeURL"`
	InfoURL   string `json:"infoURL"`
}

type masterFileMetadata struct {
	FileName     string `json:"fileName"`
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
	Box          string `json:"box"`
	Folder       string `json:"folder"`
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

	log.Printf("INFO: get all master files from %s", unitDir)

	// walk the unit directory and generate masterFile info for each .tif
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}\.tif$`)
	tifRegex := regexp.MustCompile(`^.*\.tif$`)
	hiddenRegex := regexp.MustCompile(`^\..*`)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("WARNING: directory traverse failed: %s", err.Error())
			return nil
		}

		if !f.IsDir() {
			fName := f.Name()
			if hiddenRegex.Match([]byte(fName)) {
				log.Printf("INFO: skipping hidden file %s", fName)
				return nil
			}

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

	log.Printf("INFO: found %d files in %s", len(out.MasterFiles), unitDir)
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

	elapsed := time.Since(start)
	elapsedMS := int64(elapsed / time.Millisecond)
	log.Printf("INFO: got %d masterfiles for unit %s in %dms", len(out.MasterFiles), uidStr, elapsedMS)

	c.JSON(http.StatusOK, out)
}

// getMasterFilesMetadata will retrieve the metadata for one page of master files
func (svc *serviceContext) getMasterFilesMetadata(c *gin.Context) {
	uidStr := padLeft(c.Param("uid"), 9)
	tgtFile := c.Query("file")
	if tgtFile != "" {
		log.Printf("INFO: get metadata for masterfile %s", tgtFile)
		mfMD, err := getExifData(tgtFile)
		if err != nil {
			log.Printf("ERROR: unable to get %s metadata: %s", tgtFile, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.JSON(http.StatusOK, mfMD)
		}
		return
	}

	currPage, _ := strconv.Atoi(c.Query("page"))
	if currPage == 0 {
		currPage = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("pagesize"))
	if pageSize == 0 {
		pageSize = 20
	}
	unitDir := path.Join(svc.ImagesDir, uidStr)
	startSeqNum := (currPage-1)*pageSize + 1
	log.Printf("INFO: get %d metadata records from %s starting from masterfile index %d", pageSize, unitDir, startSeqNum)

	// use exiftool to get metadata for master files on the current page only
	startTime := time.Now()
	var mdWG sync.WaitGroup
	tgtFiles := make([]string, 0)
	mdChannel := make(chan masterFileMetadata)

	for i := 0; i < pageSize; i++ {
		mfSeqNum := startSeqNum + i
		filename := fmt.Sprintf("%s_%04d.tif", uidStr, mfSeqNum)
		fullPath := findFile(unitDir, filename)
		if fullPath != "" {
			tgtFiles = append(tgtFiles, fullPath)
			if len(tgtFiles) == svc.BatchSize {
				mdWG.Add(1)
				filesCopy := make([]string, len(tgtFiles))
				copy(filesCopy, tgtFiles)
				go func() {
					defer mdWG.Done()
					getExifMetadataBatch(filesCopy, mdChannel)
				}()
				tgtFiles = make([]string, 0)
			}
		}
	}

	if len(tgtFiles) > 0 {
		mdWG.Add(1)
		go func() {
			defer mdWG.Done()
			getExifMetadataBatch(tgtFiles, mdChannel)
		}()
	}

	go func() {
		log.Printf("INFO: await all exif metadata responses")
		mdWG.Wait()
		close(mdChannel)
		elapsed := time.Since(startTime)
		log.Printf("INFO: all hexif metadata requests have completed in %.3f sec", elapsed.Seconds())
	}()

	out := make([]masterFileMetadata, 0)
	for mdResp := range mdChannel {
		out = append(out, mdResp)
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) deleteFiles(c *gin.Context) {
	rawUnitID := c.Param("uid")
	uidStr := padLeft(rawUnitID, 9)
	unitDir := path.Join(svc.ImagesDir, uidStr)
	if svc.isBatchInProgress(rawUnitID) {
		log.Printf("WARNING: request to delete files from unit %s rejected because it is already being processed", rawUnitID)
		c.String(http.StatusConflict, "this unit is currently being processed by another user")
		return
	}

	var delReq struct {
		Filenames []string `json:"filenames"`
	}
	qpErr := c.ShouldBindJSON(&delReq)
	if qpErr != nil {
		log.Printf("ERROR: invalid delete images payload for unit %s: %s", uidStr, qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	svc.addBatchProcess(rawUnitID)
	defer svc.removeBatchProcess(rawUnitID)

	for _, fn := range delReq.Filenames {
		delPath := path.Join(unitDir, fn)
		log.Printf("INFO: delete %s", delPath)
		err := os.Remove(delPath)
		if err != nil {
			log.Printf("ERROR: unable to delete %s: %s", delPath, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.String(http.StatusOK, "deleted")
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

func (svc *serviceContext) finalizeUnitData(rawUnitID string) (*finalizeResponse, error) {
	uid := padLeft(rawUnitID, 9)
	unitDir := fmt.Sprintf("%s/%s", svc.ImagesDir, uid)
	log.Printf("INFO: finalize unit %s", unitDir)

	unitMD, uErr := svc.getUnitMetadata(uid)
	if uErr != nil {
		return nil, uErr
	}

	commandsBatch := make([]exifFileCommands, 0)
	errChannel := make(chan updateProblem)
	var batchWG sync.WaitGroup

	// walk the unit directory and generate masterFile info for each .tif
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}\.tif$`)
	tifRegex := regexp.MustCompile(`^.*\.tif$`)
	hiddenRegex := regexp.MustCompile(`^\..*`)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if f.IsDir() {
			return nil
		}
		fName := f.Name()
		if hiddenRegex.Match([]byte(fName)) || !tifRegex.Match([]byte(fName)) || !mfRegex.Match([]byte(fName)) {
			return nil
		}

		cmd := exifFileCommands{File: path, Commands: make([]string, 0)}
		id := unitMD.CallNumber
		if id == "" {
			id = unitMD.PID
		}
		cmd.Commands = append(cmd.Commands, fmt.Sprintf("-iptc:MasterDocumentID=%s", fmt.Sprintf("UVA Library: %s", id)))
		cmd.Commands = append(cmd.Commands, fmt.Sprintf("-iptc:ObjectName=%s", fName))
		cmd.Commands = append(cmd.Commands, fmt.Sprintf("-iptc:ClassifyState=%s", ""))
		cmd.Commands = append(cmd.Commands, path)
		commandsBatch = append(commandsBatch, cmd)

		if len(commandsBatch) == svc.BatchSize {
			batchWG.Add(1)
			cmdCopy := make([]exifFileCommands, len(commandsBatch))
			copy(cmdCopy, commandsBatch)
			go func() {
				defer batchWG.Done()
				batchUpdateExifData(cmdCopy, errChannel)
			}()
			commandsBatch = make([]exifFileCommands, 0)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(commandsBatch) > 0 {
		batchWG.Add(1)
		go func() {
			defer batchWG.Done()
			batchUpdateExifData(commandsBatch, errChannel)
		}()
	}

	var resp finalizeResponse
	resp.Success = true
	resp.Problems = make([]updateProblem, 0)

	go func() {
		log.Printf("INFO: await all finalization updates for %s", uid)
		batchWG.Wait()
		close(errChannel)
	}()

	for problem := range errChannel {
		resp.Success = false
		resp.Problems = append(resp.Problems, problem)
	}

	log.Printf("INFO: all finalization updates for %s are done", uid)

	return &resp, nil
}
