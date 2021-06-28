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
	"time"

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
	ComponentID  string `json:"componentID"`
}

type exifData struct {
	ColorProfile  string      `json:"ProfileDescription"`
	FileSize      string      `json:"FileSize"`
	FileType      string      `json:"FileType"`
	Resolution    int         `json:"XResolution"`
	Title         interface{} `json:"Headline"`
	Description   interface{} `json:"Caption-Abstract"`
	Width         int         `json:"ImageWidth"`
	Height        int         `json:"ImageHeight"`
	ClassifyState string      `json:"ClassifyState"`
	OwnerID       interface{} `json:"OwnerID"`
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

type updateProblem struct {
	File    string `json:"file"`
	Problem string `json:"problem"`
}

func padLeft(str string, tgtLen int) string {
	for {
		if len(str) == tgtLen {
			return str
		}
		str = "0" + str
	}
}

func (svc *serviceContext) getComponent(c *gin.Context) {
	cid := c.Param("id")
	log.Printf("INFO: lookup component %s", cid)
	qs := `select title,label,date,content_desc,name from components c inner join component_types t on t.id = c.component_type_id where c.id={:cid}`
	q := svc.DB.NewQuery(qs)
	q.Bind(dbx.Params{"cid": cid})

	var dbInfo struct {
		Title       sql.NullString `db:"title"`
		Label       sql.NullString `db:"label"`
		Description sql.NullString `db:"content_desc"`
		Date        sql.NullString `db:"date"`
		Type        string         `db:"name"`
	}
	err := q.One(&dbInfo)
	if err != nil {
		log.Printf("ERROR: component %s not found: %s", cid, err.Error())
		c.String(http.StatusNotFound, "component not found")
		return
	}

	out := make(map[string]string)
	out["type"] = dbInfo.Type
	out["title"] = ""
	out["label"] = ""
	out["description"] = ""
	out["date"] = ""
	if dbInfo.Title.Valid {
		out["title"] = dbInfo.Title.String
	}
	if dbInfo.Label.Valid {
		out["label"] = dbInfo.Label.String
	}
	if dbInfo.Description.Valid {
		out["description"] = dbInfo.Description.String
	}
	if dbInfo.Date.Valid {
		out["date"] = dbInfo.Date.String
	}

	c.JSON(http.StatusOK, out)
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
	start := time.Now()
	unitMD, uErr := svc.getUnitMetadata(uidStr)
	if uErr != nil {
		log.Printf("ERROR: %s", uErr.Error())
		out.Problems = append(out.Problems, "No metadata record found for this unit")
		out.Metadata = &unitMetadata{Title: "Unknown", CallNumber: "Unknown"}
	} else {
		out.Metadata = unitMD
	}

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

	// use exiftool to get metadata for master files in batches of 10
	currIdx := 0
	pendingFilesCnt := 0
	chunkSize := 10
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
		resp.Problems = append(resp.Problems, errs...)
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
	backUpDir := path.Join(svc.ImagesDir, unit, "tmp")
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

// updateExifData is called as a goroutine. It will update a batch of files, then return true on the channel
func updateExifData(fileCommands map[string][]string, channel chan []updateProblem) {
	log.Printf("INFO: batch %+v", fileCommands)
	errors := make([]updateProblem, 0)
	for tgtFile, command := range fileCommands {
		log.Printf("INFO: exiftool %v", command)
		_, err := exec.Command("exiftool", command...).Output()
		if err != nil {
			log.Printf("ERROR: unable to update %s metadata: %s", tgtFile, err.Error())
			errors = append(errors, updateProblem{File: path.Base(tgtFile), Problem: err.Error()})
		} else {
			dupPath := fmt.Sprintf("%s_original", tgtFile)
			removeErr := os.Remove(dupPath)
			if removeErr != nil {
				log.Printf("WARNING: unable to remove backup file %s:%s", dupPath, removeErr.Error())
			}
		}
	}

	// notify caller that the block is done
	channel <- errors
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
				title := ""
				if md.Title != nil {
					title = fmt.Sprintf("%v", md.Title)
				}
				desc := ""
				if md.Description != nil {
					desc = fmt.Sprintf("%v", md.Description)
				}
				component := ""
				if md.OwnerID != nil {
					component = fmt.Sprintf("%v", md.OwnerID)
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
				files[currIdx].ComponentID = component
				currIdx++
			}
		}
	}

	// notify caller that the block is done
	log.Printf("INFO: exif batch from index %d done", startIdx)
	channel <- true
}

func baseExifCmd() []string {
	out := []string{"-json", "-ImageWidth", "-ImageHeight",
		"-FileType", "-XResolution", "-FileSize", "-icc_profile:ProfileDescription", "-iptc:OwnerID",
		"-iptc:headline", "-iptc:caption-abstract", "-iptc:ClassifyState"}
	return out
}
