package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type conatantVal struct {
	ID   int64  `json:"value"`
	Name string `json:"label"`
}

type constants struct {
	Categories []conatantVal `json:"categories"`
	Workflows  []conatantVal `json:"workflows"`
}

func (svc *serviceContext) getConstants(c *gin.Context) {
	out := constants{}

	if err := svc.DB.Table("categories").Order("name asc").Find(&out.Categories).Error; err != nil {
		log.Printf("ERROR: unable to get categories: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err := svc.DB.Table("workflows").Where("active=?", 1).Order("name asc").Find(&out.Workflows).Error; err != nil {
		log.Printf("ERROR: unable to get workflows: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) lookupProjectForUnit(c *gin.Context) {
	uid := c.Query("unit")
	if uid == "" {
		log.Printf("INFO: project lookup request is missing required unit param")
		c.String(http.StatusBadRequest, "missing required unit param")
		return
	}
	type lookupResp struct {
		Exists      bool   `json:"exists"`
		ProjectID   uint   `json:"projectID"`
		Workflow    string `json:"workflow"`
		CurrentStep string `json:"currentStep"`
		Finished    bool   `json:"finished"`
	}
	var proj project
	if err := svc.DB.Preload("CurrentStep").Preload("Workflow").Where("unit_id=?", uid).First(&proj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) == false {
			log.Printf("ERROR: lookup project for unit %s failed: %s", uid, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			log.Printf("INFO: no project exists for %s", uid)
			c.JSON(http.StatusOK, lookupResp{})
		}
		return
	}

	stepName := "None"
	if proj.CurrentStepID != nil {
		stepName = proj.CurrentStep.Name
	}
	c.JSON(http.StatusOK, lookupResp{Exists: true,
		ProjectID:   proj.ID,
		Workflow:    proj.Workflow.Name,
		CurrentStep: stepName,
		Finished:    proj.FinishedAt != nil})
}

type createProjectRequest struct {
	UnitID          int64  `json:"unitID"`
	WorkflowID      int64  `json:"workflowID"`
	ContainerTypeID int64  `json:"containerTypeID"`
	CategoryID      int64  `json:"categoryID"`
	Condition       int64  `json:"condition"`
	Notes           string `json:"notes"`
}

func (svc *serviceContext) createProject(c *gin.Context) {
	var req createProjectRequest
	if qpErr := c.ShouldBindJSON(&req); qpErr != nil {
		log.Printf("ERROR: invalid create project payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: received create project request: %v", req)

	var projCnt int64
	if err := svc.DB.Table("projects").Where("unit_id=?", req.UnitID).Count(&projCnt).Error; err != nil {
		log.Printf("ERROR: unable to determine if a project already exists for unit %d: %s", req.UnitID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if projCnt > 0 {
		log.Printf("INFO: unable to create project for unit %d as it already has a project", req.UnitID)
		c.String(http.StatusConflict, "a project already exists for this unit")
		return
	}

	log.Printf("INFO: lookup first step of new project for unit %d, workflow %d", req.UnitID, req.WorkflowID)
	var firstStep step
	if err := svc.DB.Where("workflow_id=? and step_type=0", req.WorkflowID).First(&firstStep).Error; err != nil {
		log.Printf("ERROR: unable to get first step for workflow %d: %s", req.WorkflowID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: create project for unit %d", req.UnitID)
	now := time.Now()
	newProj := project{
		WorkflowID:    uint(req.WorkflowID),
		UnitID:        uint(req.UnitID),
		CurrentStepID: &firstStep.ID,
		AddedAt:       &now,
		CategoryID:    uint(req.CategoryID),
		ItemCondition: uint(req.Condition),
		ConditionNote: req.Notes,
	}
	if req.ContainerTypeID != 0 {
		cID := uint(req.ContainerTypeID)
		newProj.ContainerTypeID = &cID
	}
	if err := svc.DB.Create(&newProj).Error; err != nil {
		log.Printf("ERROR: unable to create project for unit %d: %s", req.UnitID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: new project %d created for unit %d", newProj.ID, req.UnitID)
	c.String(http.StatusOK, fmt.Sprintf("%d", newProj.ID))
}

func (svc *serviceContext) cancelProject(c *gin.Context) {
	projID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	claims := getJWTClaims(c)
	log.Printf("INFO: %s requests cancelation of project %d", claims.ComputeID, projID)
	if claims.Role != "admin" && claims.Role != "supervisor" {
		c.String(http.StatusForbidden, "you cannot cancel this project")
		return
	}

	if err := svc.doProjectDelete(projID); err != nil {
		log.Printf("ERROR: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: project %d has been canceled and all associated data deleted", projID)
	c.String(http.StatusNotImplemented, "NO")
}

func (svc *serviceContext) failProject(c *gin.Context) {
	projID := c.Param("id")
	var req struct {
		Reason         string `json:"reason"`
		ProcessingMins uint   `json:"processingMins"`
		JobID          int64  `json:"jobID"` // optional
	}
	if qpErr := c.ShouldBindJSON(&req); qpErr != nil {
		log.Printf("ERROR: invalid fail project payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	log.Printf("INFO: fail project %s for reason [%s]", projID, req.Reason)

	var tgtProj project
	if err := svc.DB.Preload("CurrentStep").First(&tgtProj, projID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) == false {
			log.Printf("ERROR: unable to load project %s: %s", projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			log.Printf("INFO: project %s not found", projID)
			c.String(http.StatusNotFound, fmt.Sprintf("project %s not found", projID))
		}
		return
	}

	// lookup the active assignment
	log.Printf("INFO: get assignment for project %s, step %s", projID, tgtProj.CurrentStep.Name)
	var activeAssign assignment
	if err := svc.DB.Where("project_id=?", projID).Order("assigned_at DESC").Limit(1).First(&activeAssign).Error; err != nil {
		log.Printf("ERROR: unable to get active assignment for failed project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// fail the assign and increase time spent
	log.Printf("INFO: fail assignment for project %s, step %s and add duration %d mins", projID, tgtProj.CurrentStep.Name, req.ProcessingMins)
	activeAssign.DurationMinutes += req.ProcessingMins
	activeAssign.Status = 4 // error
	if err := svc.DB.Model(&activeAssign).Select("DurationMinutes", "Status").Updates(activeAssign); err != nil {
		log.Printf("ERROR: unable to update assignment %d to failed: %s", activeAssign.ID, err.Error)
	}

	// add a note describing the failure
	if tgtProj.CurrentStep.Name == "Finalization" {
		msg := fmt.Sprintf("<p>%s</p>", req.Reason)
		msg += "<p>Please manually correct the finalization problems. Once complete, press the Finish button to restart finalization.</p>"
		msg += fmt.Sprintf("<p>Error details <a href='%s/job_statuses/%d'>here</a></p>", svc.TrackSys.Client, req.JobID)
		svc.failStep(&tgtProj, "Finalization", msg)
	} else {
		msg := fmt.Sprintf("<p>%s failed</p><p>%s</p>", tgtProj.CurrentStep.Name, req.Reason)
		svc.failStep(&tgtProj, "Other", msg)
	}

	c.String(http.StatusOK, "ok")
}

func (svc *serviceContext) finishProject(c *gin.Context) {
	projID := c.Param("id")
	var req struct {
		ProcessingMins uint `json:"processingMins"`
	}
	if qpErr := c.ShouldBindJSON(&req); qpErr != nil {
		log.Printf("ERROR: invalid finish project payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	log.Printf("INFO: request to finish project %s with finalization duration %d mins", projID, req.ProcessingMins)

	var tgtProj project
	if err := svc.DB.Preload("CurrentStep").First(&tgtProj, projID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) == false {
			log.Printf("ERROR: unable to load project %s: %s", projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			log.Printf("INFO: project %s not found", projID)
			c.String(http.StatusNotFound, fmt.Sprintf("project %s not found", projID))
		}
		return
	}

	// stepType 1 is the end step. Must be on it to finish project
	log.Printf("INFO: validate current step is a final step for project %s", projID)
	if tgtProj.CurrentStep.StepType != 1 {
		log.Printf("ERROR: project %s is on non-final step %s and cannot be finished", projID, tgtProj.CurrentStep.Name)
		c.String(http.StatusPreconditionFailed, fmt.Sprintf("project is on non-final step %s and cannot be finished", tgtProj.CurrentStep.Name))
		return
	}

	log.Printf("INFO: get active assignment for project %s", projID)
	var activeAssign assignment
	if err := svc.DB.Where("project_id=?", tgtProj.ID).Order("assigned_at DESC").First(&activeAssign).Error; err != nil {
		log.Printf("ERROR: unable to get active assignment for project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	tgtProj.FinishedAt = &now
	tgtProj.OwnerID = nil
	tgtProj.CurrentStepID = nil

	// note: do this first so the calculation of total time below accounts for finalization time
	log.Printf("INFO: update project %s finalization assignment", projID)
	activeAssign.FinishedAt = &now
	activeAssign.Status = 2 // finished
	activeAssign.DurationMinutes = activeAssign.DurationMinutes + req.ProcessingMins
	if err := svc.DB.Model(&activeAssign).Select("FinishedAt", "Status", "DurationMinutes").Updates(activeAssign).Error; err != nil {
		log.Printf("ERROR: unable project %s finalization assignmnent with completion info: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("INFO: calculate total duration for project %s", projID)
	fields := []string{"FinishedAt", "OwnerID", "CurrentStepID"}
	sql := "select SUM(duration_minutes) as total from assignments where project_id=?"
	var total int64
	if err := svc.DB.Raw(sql, tgtProj.ID).Scan(&total).Error; err != nil {
		log.Printf("ERROR: unable to calculate project %s duration: %s", projID, err.Error())
	} else {
		tgtProj.TotalDurationMins = &total
		fields = append(fields, "TotalDurationMins")
		log.Printf("INFO: project %s total duration (minutes): %d", projID, total)
	}

	log.Printf("INFO: update project %s to reflect completed finalization", projID)
	if err := svc.DB.Model(tgtProj).Select(fields).Updates(tgtProj).Error; err != nil {
		log.Printf("ERROR: unable project %s completion info: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: project %s successfully marked as finished", projID)
	c.String(http.StatusOK, "ok")
}
