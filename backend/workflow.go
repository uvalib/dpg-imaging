package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Assignment status codes
const (
	StepPending    = 0 // not started
	StepStarted    = 1 // start has been clicked
	StepFinished   = 2 // finish has been clicked and all processing/validations are complete
	StepRejected   = 3 // QA rejection
	StepError      = 4 // any kind of error
	StepReassigned = 5 // step has been reassigned to a new owner
	StepFinalizing = 6 // finalization is in process
	StepWorking    = 7 // finish has been clicked, step validations in-progress
)

func (svc *serviceContext) getProjectInfo(projID string) (*project, error) {
	log.Printf("INFO: look up basic info for project %s", projID)
	var tgtProject *project
	// FIXME lookup OWNER AND ContainerType
	err := svc.DB.Preload("CurrentStep").Preload("Owner").Preload("Workflow").First(&tgtProject, projID).Error
	if err != nil {
		return nil, err
	}

	// TODO lookup STAFF
	log.Printf("INFO: get project %d assignments", tgtProject.ID)
	err = svc.DB.Where("project_id=?", tgtProject.ID).Joins("Step").Joins("StaffMember").
		Order("assigned_at DESC").Find(&tgtProject.Assignments).Error
	if err != nil {
		return nil, err
	}

	log.Printf("INFO: get project %d notes", tgtProject.ID)
	err = svc.DB.Where("project_id=?", tgtProject.ID).Joins("Step").Joins("StaffMember").Preload("Problems").
		Order("notes.created_at DESC").Find(&tgtProject.Notes).Error
	if err != nil {
		return nil, err
	}

	return tgtProject, nil
}

func (svc *serviceContext) startProjectStep(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	log.Printf("INFO: user %s is looking for project %s to start active step", claims.ComputeID, projID)
	proj, err := svc.getProjectInfo(projID)
	if err != nil {
		log.Printf("ERROR: unable to getp project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO: user %s is starting [%s] for project %s", claims.ComputeID, proj.CurrentStep.Name, projID)

	startTime := time.Now()
	if proj.StartedAt == nil {
		log.Printf("INFO: setting project %s step %s start time to %v", projID, proj.CurrentStep.Name, startTime)
		proj.StartedAt = &startTime
		err := svc.DB.Model(&proj).Select("started_at").Updates(proj).Error
		if err != nil {
			log.Printf("ERROR: unable to update project %d start time: %s", proj.ID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	currA := proj.Assignments[0]
	log.Printf("INFO: start project %d assignment %d", proj.ID, currA.ID)
	currA.StartedAt = &startTime
	currA.Status = StepStarted
	err = svc.DB.Model(&currA).Select("StartedAt", "Status").Updates(currA).Error
	if err != nil {
		log.Printf("ERROR: unable to update project %d step %d start time: %s", proj.ID, currA.StepID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
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
	proj, err := svc.getProjectInfo(projID)
	if err != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	currA := proj.Assignments[0]
	now := time.Now()
	currA.DurationMinutes = doneReq.DurationMins
	currA.FinishedAt = &now
	currA.Status = StepRejected
	err = svc.DB.Model(&currA).Select("DurationMinutes", "FinishedAt", "Status").Updates(currA).Error
	if err != nil {
		log.Printf("ERROR: unable to reject assignment: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
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

	proj, _ = svc.getProjectInfo(projID)
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) changeProjectWorkflow(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var req struct {
		Workflow      uint `json:"workflow"`
		ContainerType uint `json:"containerType"`
	}

	qpErr := c.ShouldBindJSON(&req)
	if qpErr != nil {
		log.Printf("ERROR: invalid update workflow payload: %v", qpErr)
		c.String(http.StatusBadRequest, qpErr.Error())
		return
	}

	log.Printf("INFO: user %s is changing project %s workflow to %d and container type %d", claims.ComputeID, projID, req.Workflow, req.ContainerType)
	var proj project
	err := svc.DB.Joins("CurrentStep").Joins("Owner").First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s for workflow change: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// init the respose
	out := struct {
		Workflow      workflow       `json:"workflow"`
		CurrentStep   step           `json:"step"`
		ContainerType *containerType `json:"containerType"`
		Assignments   []assignment   `json:"assignments"`
	}{}

	// load target workflow and associated steps
	err = svc.DB.Preload("Steps").First(&out.Workflow, req.Workflow).Error
	if err != nil {
		log.Printf("ERROR: unable to load new workflow %d: %s", req.Workflow, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	out.CurrentStep = out.Workflow.Steps[0]

	// load container type if non-zero container type specified
	if req.ContainerType > 0 {
		log.Printf("INFO: lookup container type %d", req.ContainerType)
		err = svc.DB.First(&out.ContainerType, req.ContainerType).Error
		if err != nil {
			log.Printf("ERROR: unable to load new container type %d: %s", req.ContainerType, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// update project workflow and first step ID
	proj.WorkflowID = req.Workflow
	proj.CurrentStepID = out.CurrentStep.ID
	if out.ContainerType != nil {
		log.Printf("INFO: update container type in project %d to %d", proj.ID, out.ContainerType.ID)
		proj.ContainerTypeID = &out.ContainerType.ID
	} else {
		proj.ContainerTypeID = nil
	}
	err = svc.DB.Debug().Model(&proj).Select("workflow_id", "current_step_id", "container_type_id").Updates(proj).Error
	if err != nil {
		log.Printf("ERROR: unable to update workflow for project %d to %d: %s", proj.ID, req.Workflow, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// update all assignments step_id, then refresh assignments list
	err = svc.DB.Model(assignment{}).Where("project_id = ?", proj.ID).Updates(assignment{StepID: proj.CurrentStepID}).Error
	if err != nil {
		log.Printf("ERROR: unable to update project %d assignment steps to %d: %s", proj.ID, proj.CurrentStepID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO lookup staff NOTE: client already has a list of staff and customers... could just sent ID and do lookup on client
	err = svc.DB.Where("project_id=?", proj.ID).Joins("Step").Joins("StaffMember").Order("assigned_at DESC").Find(&out.Assignments).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %d assignments: %s", proj.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, out)
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

	proj, err := svc.getProjectInfo(projID)
	if err != nil {
		log.Printf("ERROR: unable to getp project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO: user %s is finishing [%s] step in project [%s] with duration %d", claims.ComputeID, proj.CurrentStep.Name, projID, doneReq.DurationMins)
	noCountSteps := []string{"Scan", "Process", "Create Metadata"}
	updateCount := !slices.Contains(noCountSteps, proj.CurrentStep.Name)

	// First finish attempt includes a non-zero duration. Record it.
	// If a step fails and is corrected, 0 duration will be passed. Just
	// preserve the original duration. Requested by Sam P.
	currA := proj.Assignments[0]
	if currA.Status == StepWorking {
		log.Printf("ERROR: user %s attempt to finish [%s] step in project [%s] that is already in process", claims.ComputeID, proj.CurrentStep.Name, projID)
		c.String(http.StatusBadRequest, "step finish is already in-process")
		return
	}
	if currA.Status != StepStarted && currA.Status != StepError {
		log.Printf("ERROR: user %s attempt to finish [%s] step in project [%s] that is not started", claims.ComputeID, proj.CurrentStep.Name, projID)
		c.String(http.StatusBadRequest, "step has not been started")
		return
	}
	if doneReq.DurationMins > 0 {
		log.Printf("INFO: set duration %d for assignment %d", doneReq.DurationMins, currA.ID)
		currA.DurationMinutes = doneReq.DurationMins
		svc.DB.Model(&currA).Select("DurationMinutes").Updates(currA)
	}

	log.Printf("INFO: mark step [%s] in project [%s] as working", proj.CurrentStep.Name, projID)
	currA.Status = StepWorking
	svc.DB.Model(&currA).Select("Status").Updates(currA)

	// is this the last step of a workflow?
	if proj.CurrentStep.StepType == 1 {
		currA.Status = StepFinalizing
		svc.DB.Model(&currA).Select("Status").Updates(currA)

		go func() {
			if proj.Unit.UnitStatus != "error" {
				err := svc.validateFinishStep(proj)
				if err != nil {
					log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
					return
				}
			}

			log.Printf("INFO: sending request to dpg-jobs to begin or restart finalization of unit %d", proj.UnitID)
			_, httpErr := svc.postRequest(fmt.Sprintf("%s/units/%d/finalize", svc.TrackSys.Jobs, proj.UnitID), nil)
			if httpErr != nil {
				log.Printf("ERROR: finalize request failed: %s", httpErr.Message)
				msg := fmt.Sprintf("<p>Request to start finalization failed: %s</p>", httpErr.Message)
				svc.failStep(proj, "Other", msg)
				return
			}
		}()

		proj, _ = svc.getProjectInfo(projID)
		c.JSON(http.StatusOK, proj)
		return
	}

	validateErr := svc.validateFinishStep(proj)
	if validateErr != nil {
		log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, validateErr.Error())
		proj, _ = svc.getProjectInfo(projID)
		c.JSON(http.StatusOK, proj)
		return
	}

	log.Printf("INFO: mark assignment %d finished", currA.ID)
	nowTimeStamp := time.Now()
	currA.FinishedAt = &nowTimeStamp
	currA.Status = StepFinished
	err = svc.DB.Model(&currA).Select("FinishedAt", "Status").Updates(currA).Error
	if err != nil {
		log.Printf("ERROR: unable to update project %d step %d finish time: %s", proj.ID, currA.StepID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	var nextStep step
	nextStepID := proj.CurrentStep.NextStepID
	log.Printf("INFO: advance to next step: %d", nextStepID)
	err = svc.DB.First(&nextStep, nextStepID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %d next step %d: %s", proj.ID, nextStepID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// FIXME lookup compoute ID for owner and assignee
	log.Printf("INFO: enforce next step %s owner type %d", nextStep.Name, nextStep.OwnerType)
	switch nextStep.OwnerType {
	case 1: // prior owner
		log.Printf("INFO: project %s workflow %s advancing to new step %s with current owner %s",
			projID, proj.Workflow.Name, nextStep.Name, proj.Owner.ComputingID)
		err = svc.nextStep(proj, nextStepID, proj.OwnerID)
	case 3: // original owner
		firstA := proj.Assignments[len(proj.Assignments)-1]
		log.Printf("INFO: project %s workflow %s advancing to new step %s with originial owner %s",
			projID, proj.Workflow.Name, nextStep.Name, firstA.StaffMember.ComputingID)
		err = svc.nextStep(proj, nextStepID, &firstA.StaffMemberID)
	default:
		// any, unique or supervisor for this step. Someone must claim it, so set owner nil.
		log.Printf("INFO: project %s workflow %s advancing to new step %s with no owner set", projID, proj.Workflow.Name, nextStep.Name)
		proj.Owner = nil
		err = svc.nextStep(proj, nextStepID, nil)
	}

	if err != nil {
		log.Printf("ERROR: unable to advance step: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if updateCount {
		log.Printf("INFO: finishing this step triggers an image count update")
		mfCnt := svc.getImageCount(proj.UnitID)
		if mfCnt != proj.ImageCount {
			log.Printf("INFO: update project %s image count to %d", projID, mfCnt)
			proj.ImageCount = mfCnt
			err = svc.DB.Table("projects").Where("id = ?", proj.ID).Update("image_count", mfCnt).Error
			if err != nil {
				log.Printf("ERROR: unable to update project %s image count to %d: %s", projID, mfCnt, err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		} else {
			log.Printf("INFO: project %d already has image count %d; no update needed", proj.ID, mfCnt)
		}
	}

	// reload project to reflect changes and send result to client
	proj, _ = svc.getProjectInfo(projID)
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
	log.Printf("INFO: validate project [%d] step [%s] finish", proj.ID, proj.CurrentStep.Name)

	isManuscript := proj.Workflow.Name == "Manuscript"
	if isManuscript && proj.ContainerTypeID == nil {
		// container type is required for manuscript workflows
		svc.failStep(proj, "Other", "<p>This project is missing the required Container Type setting.</p>")
		return errors.New("manuscript is missing container type")
	}

	//  When finishing the final QA step, call finalize on the viewer to cleanup up and apply final metadata to each image
	if proj.CurrentStep.NextStepID > 0 && svc.nextStepName(proj) == "Finalize" {
		log.Printf("INFO: finishing final qa step; prep images for finalization in unit %d", proj.UnitID)
		resp, err := svc.finalizeUnitData(proj)
		if err != nil {
			log.Printf("ERROR: unable to prep unit [%d}] for finalization: %s", proj.UnitID, err.Error())
			msg := "<p>Prep for finalization failed</p>"
			msg += fmt.Sprintf("<p>DPG Imaging was unable to prep the unit for finalization: %s</p>", err.Error())
			svc.failStep(proj, "Other", msg)
			return fmt.Errorf("unable to prep unit for finalization: %s", err.Error())
		}

		if !resp.Success {
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
	imagesMovedToFinalize := false
	unitDir := padLeft(fmt.Sprintf("%d", proj.UnitID), 9)
	tgtDir := path.Join(svc.ImagesDir, unitDir)
	if proj.CurrentStep.Name == "Scan" || proj.CurrentStep.Name == "Process" {
		tgtDir = path.Join(svc.ScanDir, unitDir)
	}

	// finalize is a special case. In it, files are moved to the finalization directory. Since this step can be retried,
	// the files may be in finalization instead of the standard imaging directory. Handle that here:
	if proj.CurrentStep.Name == "Finalize" {
		// see if the starting directory is present. If not switch to finalization.
		// At this point, no validation is done. Just checking where files reside
		if !exists(tgtDir) {
			log.Printf("INFO: finalization files do not exist at %s, checking alternate location", tgtDir)
			tgtDir = path.Join(svc.FinalizeDir, unitDir)
			imagesMovedToFinalize = true
		}
	}

	err := svc.validateDirectory(proj, tgtDir)
	if err != nil {
		return err
	}

	// Files get moved in two places; after Process and Finalization. Handle the case of a retried finalize when files have already been moved
	switch proj.CurrentStep.Name {
	case "Process":
		srcDir := path.Join(svc.ScanDir, unitDir)
		tgtDir := path.Join(svc.ImagesDir, unitDir)
		moveErr := svc.moveFiles(proj, srcDir, tgtDir)
		if moveErr != nil {
			return moveErr
		}
	case "Finalize":
		if !imagesMovedToFinalize {
			srcDir := path.Join(svc.ImagesDir, unitDir)
			tgtDir := path.Join(svc.FinalizeDir, unitDir)
			moveErr := svc.moveFiles(proj, srcDir, tgtDir)
			if moveErr != nil {
				log.Printf("ERROR: move files from %s to %s failed: %s", srcDir, tgtDir, moveErr.Error())
			}
		} else {
			log.Printf("INFO: files for project %d step %s have already been moved to the finalization directory", proj.ID, proj.CurrentStep.Name)
		}
	}

	log.Printf("INFO: project %d step %s successfully finished", proj.ID, proj.CurrentStep.Name)
	return nil
}

func (svc *serviceContext) validateDirectory(proj *project, tgtDir string) error {
	log.Printf("INFO: validate project %d directory %s", proj.ID, tgtDir)

	if !exists(tgtDir) {
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
	err = svc.validateImages(proj, tgtDir)
	if err != nil {
		return err
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

		if entry.Name() == ".DS_Store" || strings.Index(entry.Name(), ".smbdelete") == 0 {
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

func (svc *serviceContext) validateImages(proj *project, tgtDir string) error {
	log.Printf("INFO: validate images info in %s", tgtDir)
	highest := -1
	cnt := 0
	isManuscript := proj.Workflow.Name == "Manuscript"

	qaFiles := make([]string, 0)
	errChannel := make(chan updateProblem)
	var checkWG sync.WaitGroup
	startTime := time.Now()

	err := filepath.WalkDir(tgtDir, func(fullPath string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}

		lcFN := strings.ToLower(entry.Name())
		if lcFN == "notes.txt" {
			if !isManuscript {
				log.Printf("ERROR: found notes.text for non-manuscript workflow")
				svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Found unexpected notes: %s</p>", fullPath))
				return fmt.Errorf("unexpected %s", fullPath)
			}
			log.Printf("INFO: found notes.txt for Manuscript workflow")
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
			qaFiles = append(qaFiles, fullPath)
			if len(qaFiles) == svc.BatchSize {
				log.Printf("INFO: check headers on batch of %d files", len(qaFiles))
				filesCopy := make([]string, len(qaFiles))
				copy(filesCopy, qaFiles)
				checkWG.Add(1)
				go func() {
					defer checkWG.Done()
					checkExifHeaders(filesCopy, isManuscript, errChannel)
				}()
				qaFiles = make([]string, 0)
			}
		}

		return nil
	})

	if len(qaFiles) > 0 {
		log.Printf("INFO: check headers on final batch of %d files", len(qaFiles))
		checkWG.Add(1)
		go func() {
			defer checkWG.Done()
			checkExifHeaders(qaFiles, isManuscript, errChannel)
		}()
	}

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

	go func() {
		log.Printf("INFO: await all metadata checks and collect any problems in the response")
		checkWG.Wait()
		close(errChannel)
		elapsed := time.Since(startTime)
		log.Printf("INFO: all header check requests have completed for %s in %.3f sec", tgtDir, elapsed.Seconds())
	}()

	errorMsg := ""
	for problem := range errChannel {
		if problem.File == "all" {
			svc.failStep(proj, "Metadata", "<p>Unable to extract metadata from images.</p>")
			return fmt.Errorf("unable to extract metadata from images")
		}
		errorMsg += fmt.Sprintf("<li>%s - %s</li>", path.Base(problem.File), problem.Problem)
	}

	if errorMsg != "" {
		svc.failStep(proj, "Metadata", fmt.Sprintf("The following errors were found: <ul>%s<ul>", errorMsg))
		return fmt.Errorf("one or mor images has metadata errors")
	}

	log.Printf("INFO: %s images and metadata are valid", tgtDir)
	return nil
}

func (svc *serviceContext) moveFiles(proj *project, srcDir string, destDir string) error {
	log.Printf("INFO: move project %d files from %s to %s", proj.ID, srcDir, destDir)
	srcExist := exists(srcDir)
	destExist := exists(destDir)
	if !srcExist && !destExist {
		svc.failStep(proj, "Filesystem", "<p>Neither start nor finsh directory exists</p>")
		return fmt.Errorf("neither source %s or destination %s exists", srcDir, destDir)
	}

	// Both exist something is wrong. Fail
	if srcExist && destExist {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Both source %s and destination %s exist</p>", srcDir, destDir))
		return fmt.Errorf("both source %s and destination %s exist", srcDir, destDir)
	}

	// Source is gone but dest exists. No move needed
	if !srcExist && destExist {
		log.Printf("source %s is missing and destination %s exists; no move needed", srcDir, destDir)
		return nil
	}

	// See if there is an 'Output' directory for special handling. This is the directory where CaptureOne
	// places the generated .tif files. Treat it as the source location if it is present
	outputDir := path.Join(srcDir, "Output")
	if exists(outputDir) {
		log.Printf("INFO: output directory %s found; moving it to %s", outputDir, destDir)
		srcDir = outputDir
	}

	log.Printf("INFO: recursively copy %s to %s", srcDir, destDir)
	startTime := time.Now()
	cmdArray := []string{"-R", srcDir, destDir}
	_, err := exec.Command("cp", cmdArray...).Output()
	if err != nil {
		svc.failStep(proj, "Filesystem", fmt.Sprintf("<p>Move %s to %s failed: %s</p>", srcDir, destDir, err.Error()))
		return fmt.Errorf("unable to copy source %s to destination %s: %s", srcDir, destDir, err.Error())
	}
	elapsed := time.Since(startTime)
	elapsedMS := int64(elapsed / time.Millisecond)
	log.Printf("INFO: copy from %s to %s completed in %d ms", srcDir, destDir, elapsedMS)

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
	resp := svc.DB.First(&nextStep, nextStepID)
	if resp.Error != nil {
		return "Unknown"
	}
	return nextStep.Name
}
