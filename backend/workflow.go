package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (svc *serviceContext) startProjectStep(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	log.Printf("INFO: user %s is starting active step in project %s", claims.ComputeID, projID)
	var proj project
	dbReq := svc.getBaseProjectQuery().Where("projects.id=?", projID)
	resp := dbReq.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	startTime := time.Now()
	if proj.StartedAt == nil {
		log.Printf("INFO: setting project start time to %v", startTime)
		proj.StartedAt = &startTime
		r := svc.DB.Model(&proj).Select("started_at").Updates(proj)
		if r.Error != nil {
			log.Printf("ERROR: unable to update project %d start time: %s", proj.ID, r.Error.Error())
			c.String(http.StatusInternalServerError, r.Error.Error())
			return
		}
	}

	currA := proj.Assignments[0]
	log.Printf("INFO: start project %d assignment %d", proj.ID, currA.ID)
	currA.StartedAt = &startTime
	currA.Status = 1 // started
	r := svc.DB.Model(&currA).Select("StartedAt", "Status").Updates(currA)
	if r.Error != nil {
		log.Printf("ERROR: unable to update project %d step %d start time: %s", proj.ID, currA.StepID, r.Error.Error())
		c.String(http.StatusInternalServerError, r.Error.Error())
		return
	}
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) rejectProjectStep(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var doneReq struct {
		DurationMins uint `json:"durationMins"`
	}

	qpErr := c.ShouldBindJSON(&doneReq)
	if qpErr != nil {
		log.Printf("ERROR: invalid reject step payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: user %s is rejecting active step in project %s with duration %d", claims.ComputeID, projID, doneReq.DurationMins)
	var proj project
	dbReq := svc.getBaseProjectQuery().Where("projects.id=?", projID)
	resp := dbReq.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	currA := proj.Assignments[0]
	now := time.Now()
	currA.DurationMinutes = doneReq.DurationMins
	currA.FinishedAt = &now
	currA.Status = 3 // rejected
	resp = svc.DB.Model(&currA).Select("DurationMinutes", "FinishedAt", "Status").Updates(currA)
	if resp.Error != nil {
		log.Printf("ERROR: unable to reject assignment: %s", resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	// all errors go back to the scanner... the owner of the first step
	failStepID := proj.CurrentStep.FailStepID
	firstA := proj.Assignments[len(proj.Assignments)-1]
	proj.OwnerID = &firstA.StaffMemberID
	proj.CurrentStepID = failStepID
	svc.DB.Model(proj).Select("CurrentStepID", "OwnerID").Updates(proj)
	newAssign := assignment{ProjectID: proj.ID, StepID: failStepID, StaffMemberID: firstA.StaffMemberID, AssignedAt: &now}
	svc.DB.Create(&newAssign)

	dbReq.First(&proj)
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) finishProjectStep(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var doneReq struct {
		DurationMins uint `json:"durationMins"`
	}

	qpErr := c.ShouldBindJSON(&doneReq)
	if qpErr != nil {
		log.Printf("ERROR: invalid finish step payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: user %s is finishing active step in project %s with duration %d", claims.ComputeID, projID, doneReq.DurationMins)
	var proj project
	dbReq := svc.getBaseProjectQuery().Where("projects.id=?", projID)
	resp := dbReq.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	// First finish attempt includes a non-zero duration. Record it.
	// If a step fails and is corrected, 0 duration will be passed. Just
	// preserve the original duration. Requested by Sam P.
	currA := proj.Assignments[0]
	if doneReq.DurationMins > 0 {
		log.Printf("INFO: set duration %d for assignment %d", doneReq.DurationMins, currA.ID)
		currA.DurationMinutes = doneReq.DurationMins
		svc.DB.Model(&currA).Select("DurationMinutes").Updates(currA)
	}

	// is this the last step of a workflow?
	if proj.CurrentStep.StepType == 1 {
		if proj.Unit.UnitStatus != "error" {
			err := svc.validateFinishStep(&proj)
			if err != nil {
				log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
				dbReq.First(&proj)
				c.JSON(http.StatusOK, proj)
				return
			}
		}

		log.Printf("INFO: sending request to TrackSys to begin or restart finalization of unit %d", proj.UnitID)
		var fr struct {
			UnitID uint `json:"unit_id"`
		}
		fr.UnitID = proj.UnitID
		_, httpErr := svc.postRequest(fmt.Sprintf("%s/api/finalize", svc.TrackSysURL), fr)
		if httpErr != nil {
			currA.Status = 4 // error
			svc.DB.Model(&currA).Select("Status").Updates(currA)
			log.Printf("ERROR: finalize request failed: %s", httpErr.Message)
			c.String(http.StatusInternalServerError, httpErr.Message)
			return
		}

		currA.Status = 6 // finalizing
		svc.DB.Model(&currA).Select("Status").Updates(currA)
		c.JSON(http.StatusOK, proj)
		return
	}

	validateErr := svc.validateFinishStep(&proj)
	if validateErr != nil {
		log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, validateErr.Error())
		dbReq.First(&proj)
		c.JSON(http.StatusOK, proj)
		return
	}

	log.Printf("INFO: mark assignment %d finished", currA.ID)
	nowTimeStamp := time.Now()
	currA.FinishedAt = &nowTimeStamp
	currA.Status = 2 // finished
	resp = svc.DB.Model(&currA).Select("FinishedAt", "Status").Updates(currA)
	if resp.Error != nil {
		log.Printf("ERROR: unable to update project %d step %d finish time: %s", proj.ID, currA.StepID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	var nextStep step
	nextStepID := proj.CurrentStep.NextStepID
	log.Printf("INFO: advance to next step: %d", nextStepID)
	resp = svc.DB.Find(&nextStep, nextStepID)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %d next step %d: %s", proj.ID, nextStepID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: enforce next step %s owner type %d", nextStep.Name, nextStep.OwnerType)
	var err error
	if nextStep.OwnerType == 1 { // prior owner
		log.Printf("INFO: project %s workflow %s advancing to new step %s with current owner %s",
			projID, proj.Workflow.Name, nextStep.Name, proj.Owner.ComputingID)
		err = svc.nextStep(&proj, nextStepID, proj.OwnerID)
	} else if nextStep.OwnerType == 3 { // original owner
		firstA := proj.Assignments[len(proj.Assignments)-1]
		log.Printf("INFO: project %s workflow %s advancing to new step %s with originial owner %s",
			projID, proj.Workflow.Name, nextStep.Name, firstA.StaffMember.ComputingID)
		err = svc.nextStep(&proj, nextStepID, &firstA.StaffMemberID)
	} else {
		// any, unique or supervisor for this step. Someone must claim it, so set owner nil.
		log.Printf("INFO: project %s workflow %s advancing to new step %s with no owner set", projID, proj.Workflow.Name, nextStep.Name)
		proj.Owner = nil
		err = svc.nextStep(&proj, nextStepID, nil)
	}

	if err != nil {
		log.Printf("ERROR: unable to advance step: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// reload project to reflect changes and send result to client
	dbReq.First(&proj)
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) nextStep(proj *project, nextStepID uint, ownerID *uint) error {
	log.Printf("INFO: advance project %d to step %d", proj.ID, nextStepID)
	proj.CurrentStepID = nextStepID
	proj.OwnerID = ownerID
	resp := svc.DB.Model(proj).Select("CurrentStepID", "OwnerID").Updates(proj)
	if resp.Error != nil {
		return resp.Error
	}

	if ownerID != nil {
		log.Printf("INFO: assign step %d to staff %d", nextStepID, *ownerID)
		now := time.Now()
		newAssign := assignment{ProjectID: proj.ID, StepID: nextStepID, StaffMemberID: *ownerID, AssignedAt: &now}
		resp := svc.DB.Create(&newAssign)
		if resp.Error != nil {
			return resp.Error
		}
	}
	return nil
}

func (svc *serviceContext) validateFinishStep(proj *project) error {
	if proj.Workflow.Name == "Manuscript" && *proj.ContainerTypeID == 0 {
		svc.failStep(proj, "Other", "<p>This project is missing the required Container Type setting.</p>")
		return errors.New("manuscript is missing container type")
	}

	//  When finishing the final QA step, call finalize on the viewer to cleanup up and apply final metadata to each image
	if proj.CurrentStep.NextStepID > 0 && svc.nextStepName(proj) == "Finalize" {
		log.Printf("INFO: finishing final qa step; prep images for finalization in unit %d", proj.UnitID)
		resp, err := svc.finalizeUnitData(fmt.Sprintf("%d", proj.UnitID))
		if err != nil {
			log.Printf("ERROR: unable to prep unit [%d}] for finalization: %s", proj.UnitID, err.Error())
			msg := "<p>Prep for finalization failed</p>"
			msg += fmt.Sprintf("<p>DPG Imaging was unable to prep the unit for finalization: %s</p>", err.Error())
			svc.failStep(proj, "Other", msg)
			return fmt.Errorf("unable to prep unit for finalization: %s", err.Error())
		}

		if resp.Success == false {
			log.Printf("INFO: unit %d has finalize errors: %v", proj.UnitID, resp.Problems)
			msg := "<p>Prep for finalization failed</p>"
			for _, p := range resp.Problems {
				msg += fmt.Sprintf("<p>%s: %s</p>", p.File, p.Problem)
			}
			svc.failStep(proj, "Other", msg)
			return fmt.Errorf("unit %d has finalization problems", proj.UnitID)
		}
		log.Printf("INFO: unit %d data has been finalized", proj.UnitID)
	}

	// Make sure  directory is clean and in proper structure
	unitDir := padLeft(fmt.Sprintf("%d", proj.UnitID), 9)
	tgtDir := path.Join(svc.ImagesDir, unitDir)
	if proj.CurrentStep.Name == "Scan" || proj.CurrentStep.Name == "Process" {
		tgtDir = path.Join(svc.ScanDir, unitDir)
	}
	err := svc.validateDirectory(proj, tgtDir)
	if err != nil {
		return err
	}

	// Files get moved in two places; after Process and Finalization
	var moveErr error
	if proj.CurrentStep.Name == "Process" {
		srcDir := path.Join(svc.ScanDir, unitDir)
		tgtDir := path.Join(svc.ImagesDir, unitDir)
		moveErr = svc.moveFiles(proj, srcDir, tgtDir)
	}
	if proj.CurrentStep.Name == "Finalize" {
		srcDir := path.Join(svc.ImagesDir, unitDir)
		tgtDir := path.Join(svc.FinalizeDir, unitDir)
		moveErr = svc.moveFiles(proj, srcDir, tgtDir)
	}
	if moveErr != nil {
		return moveErr
	}

	log.Printf("INFO: project %d step %s successfully finished", proj.ID, proj.CurrentStep.Name)
	return nil
}

func (svc *serviceContext) validateDirectory(proj *project, tgtDir string) error {
	log.Printf("INFO: validate project %d directory %s", proj.ID, tgtDir)

	if dirExist(tgtDir) == false {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Directory %s does not exist</p>", tgtDir))
		return fmt.Errorf("%s does not exist", tgtDir)
	}

	// Scan and Process and Error steps have no checks other than directory existance
	if proj.CurrentStep.Name == "Scan" || proj.CurrentStep.Name == "Process" || proj.CurrentStep.StepType == 2 {
		log.Printf("INFO: scan, process and error steps have no validations")
		return nil
	}

	err := svc.validateDirectoryContent(proj, tgtDir)
	if err != nil {
		return err
	}
	err = svc.validateTifSequence(proj, tgtDir)
	if err != nil {
		return err
	}
	if proj.Workflow.Name == "Manuscript" {
		err = svc.validateDirectoryStructure(proj, tgtDir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (svc *serviceContext) validateDirectoryContent(proj *project, tgtDir string) error {
	log.Printf("INFO: validate contents of %s", tgtDir)
	err := filepath.WalkDir(tgtDir, func(fullPath string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}

		lcFN := strings.ToLower(entry.Name())
		if proj.Workflow.Name == "Manuscript" {
			if lcFN == "notes.txt" {
				log.Printf("INFO: found location notes file for manifest project %d", proj.ID)
				return nil
			}
		}

		if entry.Name() == ".DS_Store" {
			log.Printf("INFO: remove  %s", fullPath)
			os.Remove(fullPath)
		} else {
			if filepath.Ext(lcFN) != ".tif" {
				svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Unexpected file %s found</p>", fullPath))
				return fmt.Errorf("found unexpected file %s", fullPath)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf("INFO: %s content is valid", tgtDir)
	return nil
}

func (svc *serviceContext) validateTifSequence(proj *project, tgtDir string) error {
	log.Printf("INFO: validate count and tif sequence in %s", tgtDir)
	highest := -1
	cnt := 0
	err := filepath.WalkDir(tgtDir, func(fullPath string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}

		ext := filepath.Ext(entry.Name())
		noExtFN := strings.TrimSuffix(entry.Name(), ext)
		seqStr := strings.Split(noExtFN, "_")[1]
		seq, _ := strconv.Atoi(seqStr)
		cnt++
		if seq > highest {
			highest = seq
		}

		// make sure filename is well formed
		unitDir := padLeft(fmt.Sprintf("%d", proj.UnitID), 9)
		if unitDir != strings.Split(noExtFN, "_")[0] {
			log.Printf("ERROR: invalid name %s", fullPath)
			svc.failStep(proj, "Filename", fmt.Sprintf("<p>Found incorrectly named image file %s.</p>", fullPath))
			return fmt.Errorf("invalid filename %s", fullPath)
		}

		if proj.CurrentStep.Name == "Create Metadata" || proj.CurrentStep.Name == "Finalize" {
			// Make sure metadata is present at completion of create metadata and again as a final check on finalize
			cmdArray := []string{"-json", "-iptc:headline", fullPath}
			log.Printf("INFO: exif command %+v", cmdArray)
			stdout, err := exec.Command("exiftool", cmdArray...).Output()
			if err != nil {
				svc.failStep(proj, "Metadata", fmt.Sprintf("<p>Unable to extract metadata from %s.</p>", fullPath))
				return fmt.Errorf("unable to extract metadata from %s: %s", fullPath, err.Error())
			}
			log.Printf("INFO: exif response %s", stdout)
			var mfMD []exifData
			err = json.Unmarshal(stdout, &mfMD)
			if err != nil {
				svc.failStep(proj, "Metadata", fmt.Sprintf("<p>Unable to extract metadata from %s.</p>", fullPath))
				return fmt.Errorf("unable to extract metadata from %s: %s", fullPath, err.Error())
			}

			if mfMD[0].Title == nil {
				svc.failStep(proj, "Metadata", fmt.Sprintf("<p>Missing Tile metadata in %s.</p>", fullPath))
				return fmt.Errorf("Missing Tile metadata in %s", fullPath)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}
	if cnt == 0 {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>No image files found in %s.</p>", tgtDir))
		return fmt.Errorf("no files found in %s", tgtDir)
	}
	if highest != cnt {
		svc.failStep(proj, "Filename", fmt.Sprintf("<p>Number of image files does not match highest image sequence number %d.</p>", highest))
		return fmt.Errorf("count/sequence mismatch in %s", tgtDir)
	}

	log.Printf("INFO: %s sequence is valid", tgtDir)
	return nil
}

func (svc *serviceContext) validateDirectoryStructure(proj *project, tgtDir string) error {
	log.Printf("INFO: validate structure %s", tgtDir)
	// All manuscripts have a top level directory with any name. If ContainerType is
	// set to has_folders=true, this must contain folders only. all tif images reside in the folders.
	// If not, the top-level directory containes the tif images
	foundContainerDir := false
	foundFolders := false
	err := filepath.WalkDir(tgtDir, func(fullPath string, entry fs.DirEntry, err error) error {
		if err != nil || tgtDir == fullPath {
			return nil
		}

		// strip off base dir, leaving just the relative path (minus the starting /)
		relPath := fullPath[len(tgtDir)+1:]
		depth := len(strings.Split(relPath, "/"))
		log.Printf("INFO: validate %s, depth %d", relPath, depth)
		if entry.IsDir() {
			if depth > 2 {
				svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Too many subdirectories: %s</p>", relPath))
				return fmt.Errorf("too many subdirectories %s", relPath)
			}

			if depth == 2 {
				if proj.ContainerType.HasFolders == false {
					svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Folder directories not allowed in '%s' containers</p>", proj.ContainerType.Name))
					return fmt.Errorf("folders not allowed in %s", proj.ContainerType.Name)
				}
				foundFolders = true
			}

			if depth == 1 {
				if foundContainerDir == true {
					svc.failStep(proj, "Filesystem", "<p>There can only be one box directory</p>")
					return fmt.Errorf("multiple box directories found in %s", tgtDir)
				}
				foundContainerDir = true
			}
		} else {
			// Count slashes to figure out where in the directory tree this file resides.
			// NOTE: use -1 becase the filename is part of the relative path; EX: box/sample.tif
			// Validate based on folders flag in the container_type model
			if depth-1 < 1 {
				svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Files found in project root directory: %s</p>", fullPath))
				return fmt.Errorf("files found in root directory %s", fullPath)
			}
			if depth-1 == 1 && proj.ContainerType.HasFolders {
				svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Files found in box directory: %s</p>", relPath))
				return fmt.Errorf("files found in box directory %s", relPath)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Validate presence of folders based on container_type
	if foundFolders == false && proj.ContainerType.HasFolders == true {
		svc.failStep(proj, "Filesystem", "<p>No folder directories found</p>")
		return fmt.Errorf("missing required folders within %s", tgtDir)
	}

	log.Printf("INFO: %s structure is valid", tgtDir)
	return nil
}

func (svc *serviceContext) moveFiles(proj *project, srcDir string, destDir string) error {
	log.Printf("INFO: move project %d files from %s to %s", proj.ID, srcDir, destDir)
	if dirExist(srcDir) == false && dirExist(destDir) == false {
		svc.failStep(proj, "Filesystem", "<p>Neither start nor finsh directory exists</p>")
		return fmt.Errorf("neither source %s or destination %s exists", srcDir, destDir)
	}

	// Both exist without DELETE.ME; something is wrong. Fail
	if dirExist(srcDir) && dirExist(destDir) {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Both source %s and destination %s exist</p>", srcDir, destDir))
		return fmt.Errorf("both source %s and destination %s exists", srcDir, destDir)
	}

	// Source is gone but dest exists. No move needed
	if dirExist(srcDir) == false && dirExist(destDir) {
		log.Printf("source %s is missing (of has DELETE.ME) and destination %s already exists; no meve needed", srcDir, destDir)
		return nil
	}

	// See if there is an 'Output' directory for special handling. This is the directory where CaptureOne
	// places the generated .tif files. Treat it as the source location if it is present
	outputDir := path.Join(srcDir, "Output")
	if dirExist(outputDir) {
		log.Printf("INFO: output directory %s found; moving it to %s", outputDir, destDir)
		srcDir = outputDir
	}

	log.Printf("INFO: recursively copy %s to %s", srcDir, destDir)
	cmdArray := []string{"-R", srcDir, destDir}
	_, err := exec.Command("cp", cmdArray...).Output()
	if err != nil {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Move %s to %s failed: %s</p>", srcDir, destDir, err.Error()))
		return fmt.Errorf("unable to copy source %s to destination %s: %s", srcDir, destDir, err.Error())
	}

	log.Printf("INFO: cleanup original %s", srcDir)
	cmdArray = []string{"-r", srcDir}
	_, err = exec.Command("rm", cmdArray...).Output()
	if err != nil {
		log.Printf("WARN: unable to remove %s after copy to %s", srcDir, destDir)
	}

	log.Printf("INFO: files successfully moved to %s", destDir)
	return nil
}

func (svc *serviceContext) nextStepName(proj *project) string {
	nextStepID := proj.CurrentStep.NextStepID
	var nextStep step
	resp := svc.DB.Find(&nextStep, nextStepID)
	if resp.Error != nil {
		return "Unknown"
	}
	return nextStep.Name
}

func (svc *serviceContext) failStep(proj *project, problemName string, message string) {
	log.Printf("INFO: flag project %d step %s with an error", proj.ID, proj.CurrentStep.Name)
	currA := proj.Assignments[0]
	currA.Status = 4 // error
	svc.DB.Model(&currA).Select("Status").Updates(currA)

	log.Printf("INFO: adding problem(%s) note to project %d step %s", problemName, proj.ID, proj.CurrentStep.Name)
	now := time.Now()
	newNote := note{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: *proj.OwnerID,
		NoteType: 2, Note: message, CreatedAt: &now, UpdatedAt: &now}
	err := svc.DB.Model(&proj).Association("Notes").Append(&newNote)
	if err != nil {
		log.Printf("ERROR: unable to add note to project %d: %s", proj.ID, err.Error())
		return
	}

	var p problem
	resp := svc.DB.Where("label = ?", problemName).First(&p)
	if resp.Error != nil {
		p.ID = 7 // other
	}

	pq := "insert into notes_problems (note_id, problem_id) values "
	var vals []string
	vals = append(vals, fmt.Sprintf("(%d,%d)", newNote.ID, p.ID))

	pq += strings.Join(vals, ",")
	resp = svc.DB.Exec(pq)
	if resp.Error != nil {
		log.Printf("ERROR: unable to add problems to note: %s", resp.Error.Error())
	}
}

func dirExist(tgtDir string) bool {
	log.Printf("INFO: check existance of %s", tgtDir)
	_, err := os.Stat(tgtDir)
	if err != nil {
		log.Printf("ERROR: check %s failed: %s", tgtDir, err.Error())
		return false
	}
	return true
}

func fileExists(rootDir string, tgtFile string) bool {
	exist := false
	filepath.WalkDir(rootDir, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.Name() == tgtFile {
			exist = true
		}
		return nil
	})
	return exist
}
