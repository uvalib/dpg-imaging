package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

func (svc *serviceContext) finishProjectStep(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var doneReq struct {
		DurationMins uint `json:"durationMins"`
	}

	qpErr := c.ShouldBindJSON(&doneReq)
	if qpErr != nil {
		log.Printf("ERROR: invalid finsih step payload: %v", qpErr)
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
		// first time, finish the step
		if proj.Unit.UnitStatus != "error" {
			err := svc.validateFinishStep(&proj)
			if err != nil {
				log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
				dbReq.First(&proj)
				c.JSON(http.StatusOK, proj)
				return
			}
		}
	} else {
		err := svc.validateFinishStep(&proj)
		if err != nil {
			log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
			dbReq.First(&proj)
			c.JSON(http.StatusOK, proj)
			return
		}
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
	if proj.Workflow.Name == "Manuscript" && proj.ContainerTypeID == 0 {
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
		return nil
	}

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

	err := os.Rename(srcDir, destDir)
	if err != nil {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Move %s to %s failed: %s</p>", srcDir, destDir, err.Error()))
		return fmt.Errorf("unable to move source %s to destination %s: %s", srcDir, destDir, err.Error())
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
	resp := svc.DB.Where("name = ?", problemName).First(&p)
	if resp.Error != nil {
		p.ID = 7 // other
	}
	svc.DB.Model(&newNote).Association("Problems").Append(&p)
}

func dirExist(tgtDir string) bool {
	if _, err := os.Stat(tgtDir); os.IsNotExist(err) {
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
