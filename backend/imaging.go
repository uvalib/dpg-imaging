package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	dbx "github.com/go-ozzo/ozzo-dbx"
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
	Status       string `json:"status"`
}

type exifData struct {
	ColorProfile  string      `json:"ICCProfileName"`
	FileSize      string      `json:"FileSize"`
	FileType      string      `json:"FileType"`
	Resolution    int         `json:"XResolution"`
	Title         interface{} `json:"Headline"`
	Description   interface{} `json:"Caption-Abstract"`
	Width         int         `json:"ImageWidth"`
	Height        int         `json:"ImageHeight"`
	ClassifyState string      `json:"ClassifyState"`
}

type unitMetadata struct {
	ID         int64         `db:"id" json:"id"`
	CallNumber string        `db:"call_number" json:"callNumber"`
	Title      string        `db:"title" json:"title"`
	ProjectID  sql.NullInt64 `db:"project" json:"-"`
	ProjectURL string        `json:"projectURL"`
}

type unitData struct {
	Metadata    *unitMetadata     `json:"metadata"`
	MasterFiles []*masterFileInfo `json:"masterFiles"`
	Problems    []string          `json:"problems"`
}

func padLeft(str string, tgtLen int) string {
	for {
		if len(str) == tgtLen {
			return str
		}
		str = "0" + str
	}
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

func (svc *serviceContext) getUnitMetadata(uidStr string) (*unitMetadata, error) {
	log.Printf("INFO: get metadata for unit %s", uidStr)
	uid, convErr := strconv.Atoi(uidStr)
	if convErr != nil {
		return nil, fmt.Errorf("unit id %s is not valid", uidStr)
	}

	qSQL := `select m.id, call_number, title, p.id as project from metadata m
		inner join units u on u.metadata_id = m.id
		left outer join projects p on p.unit_id=u.id
		where u.id={:uid}`
	q := svc.DB.NewQuery(qSQL)
	q.Bind(dbx.Params{"uid": uid})
	var dbInfo unitMetadata
	err := q.One(&dbInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to get metadata for unit %d: %s", uid, err.Error())
	}

	if dbInfo.ProjectID.Valid {
		dbInfo.ProjectURL = fmt.Sprintf("%s/admin/projects/%d", svc.TrackSysURL, dbInfo.ProjectID.Int64)
	}

	return &dbInfo, nil
}

func (svc *serviceContext) getUnitDetails(c *gin.Context) {
	uidStr := padLeft(c.Param("uid"), 9)
	unitDir := path.Join(svc.ImagesDir, uidStr)
	out := unitData{MasterFiles: make([]*masterFileInfo, 0), Problems: make([]string, 0)}

	log.Printf("INFO: get details for unit %s", uidStr)
	unitMD, uErr := svc.getUnitMetadata(uidStr)
	if uErr != nil {
		log.Printf("ERROR: %s", uErr.Error())
		c.String(http.StatusInternalServerError, uErr.Error())
		return
	}
	out.Metadata = unitMD

	log.Printf("INFO: get master files from %s", unitDir)

	// walk the unit directory and generate masterFile info for each .tif
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}\.tif$`)
	tifRegex := regexp.MustCompile(`^.*\.tif$`)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("WARNING: directory traverse failed: %s", err.Error())
			return nil
		}

		if f.IsDir() == false {
			fName := f.Name()
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
			mf.ThumbURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/30,/0/default.jpg", svc.IIIFURL, pathID, fName)
			mf.MediumURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/250,/0/default.jpg", svc.IIIFURL, pathID, fName)
			mf.LargeURL = fmt.Sprintf("%s/iiif/2/%s%%2F%s/full/400,/0/default.jpg", svc.IIIFURL, pathID, fName)
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

	lastMF := out.MasterFiles[len(out.MasterFiles)-1]
	lastSeq := strings.Split(lastMF.FileName, "_")[1]
	lastSeq = strings.Replace(lastSeq, ".tif", "", 1)
	lastSeqNum, _ := strconv.Atoi(lastSeq)
	if len(out.MasterFiles) != lastSeqNum {
		out.Problems = append(out.Problems, "Last masterfile number doesn't match count")
	}

	// use exiftool to get metadata for master files in batches of 10
	currIdx := 0
	pendingFilesCnt := 0
	cmdArray := baseExifCmd()
	for _, mf := range out.MasterFiles {
		cmdArray = append(cmdArray, mf.Path)
		pendingFilesCnt++
		if pendingFilesCnt == 10 {
			getExifData(cmdArray, out.MasterFiles, currIdx)
			cmdArray = baseExifCmd()
			currIdx = pendingFilesCnt
			pendingFilesCnt = 0
		}
	}

	if pendingFilesCnt > 0 {
		getExifData(cmdArray, out.MasterFiles, currIdx)
	}

	c.JSON(http.StatusOK, out)
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
		statusEdit := fmt.Sprintf("-iptc:ClassifyState=%s", change.Status)
		log.Printf("INFO: exiftool %s %s %s", titleEdit, descEdit, change.File)
		_, err := exec.Command("exiftool", titleEdit, descEdit, statusEdit, change.File).Output()
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
	backUpDir := path.Join(svc.ImagesDir, unit, "tmp")
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

	// second pass moves the renambed files back where they belong
	for _, rn := range rnPost {
		tmpFile := path.Join(backUpDir, rn.NewName)
		origDir, _ := path.Split(rn.Original)
		renamed := path.Join(origDir, rn.NewName)
		log.Printf("INFO: move tmp %s to %s", tmpFile, renamed)
		err := os.Rename(tmpFile, renamed)
		if err != nil {
			log.Printf("ERROR: unable to restore %s from %s: %s", renamed, tmpFile, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// last, cleanup tmp
	os.Remove(backUpDir)

	c.String(http.StatusOK, "ok")
}

func (svc *serviceContext) deleteFile(c *gin.Context) {
	unit := padLeft(c.Param("uid"), 9)
	file := c.Param("file")
	unitDir := path.Join(svc.ImagesDir, unit)
	log.Printf("INFO: delete %s from unit %s", file, unit)

	delFilePath := ""
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("WARNING: delete file directory traverse failed: %s", err.Error())
		} else {
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

	log.Printf("INFO: delete %s", delFilePath)
	err = os.Remove(delFilePath)
	if err != nil {
		log.Printf("ERROR: unable to delete %s: %s", delFilePath, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, fmt.Sprintf("deleted %s", delFilePath))
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
				files[currIdx].Status = md.ClassifyState
				currIdx++
			}
		}
	}
}

func baseExifCmd() []string {
	out := []string{"-json", "-ImageWidth", "-ImageHeight",
		"-FileType", "-XResolution", "-FileSize", "-ICCProfileName",
		"-iptc:headline", "-iptc:caption-abstract", "-iptc:ClassifyState"}
	return out
}
