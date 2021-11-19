package main

import (
	"errors"
	"log"
	"net/http"
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
	log.Printf("INFO: user %s is finishing active step in project %s", claims.ComputeID, projID)

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
				c.String(http.StatusBadRequest, err.Error())
				return
			}
		}
	} else {
		err := svc.validateFinishStep(&proj)
		if err != nil {
			log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
			c.String(http.StatusBadRequest, err.Error())
			return
		}
	}

	log.Printf("INFO: advance to next step")
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
		c.String(http.StatusInternalServerError, resp.Error.Error())
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

	// TODO lots

	log.Printf("INFO: project %d step %s successfully finished", proj.ID, proj.CurrentStep.Name)
	return nil
}

func (svc *serviceContext) failStep(proj *project, problemName string, message string) {
	log.Printf("INFO: flag project %d step %s with an error", proj.ID, proj.CurrentStep.Name)
	currA := proj.Assignments[0]
	currA.Status = 4 // error
	svc.DB.Model(&currA).Select("Status").Updates(currA)

	log.Printf("INFO: adding problem(%s) note to project %d step %s", problemName, proj.ID, proj.CurrentStep.Name)
	newNote := note{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: *proj.OwnerID, NoteType: 2, Note: message}
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
