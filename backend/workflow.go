package main

import (
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
		r := svc.DB.Model(&proj).Update("started_at", startTime)
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
			err := svc.finishStep(&proj)
			if err != nil {
				log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
				c.String(http.StatusBadGateway, err.Error())
				return
			}
		}
	} else {
		err := svc.finishStep(&proj)
		if err != nil {
			log.Printf("ERROR: unable to finish project %s step %s: %s", projID, proj.CurrentStep.Name, err.Error())
			c.String(http.StatusBadGateway, err.Error())
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
		// TODO ADD ERROR NOTE TO PROJECT
		return
	}

	var nextStep step
	nextStepID := proj.CurrentStep.NextStepID
	resp = svc.DB.Find(&nextStep, nextStepID)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %d next step %d: %s", proj.ID, nextStepID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		// TODO ADD ERROR NOTE TO PROJECT
		return
	}

	log.Printf("INFO: enforce next step %s owner type %d", nextStep.Name, nextStep.OwnerType)
	if nextStep.OwnerType == 1 { // prior owner
		log.Printf("INFO: project %s workflow %s advancing to new step %s with current owner %s",
			projID, proj.Workflow.Name, nextStep.Name, proj.Owner.ComputingID)
		//    self.update(current_step: new_step)
		//    Assignment.create(project: self, staff_member: self.owner, step: new_step)
	} else if nextStep.OwnerType == 3 { // original owner
		firstA := proj.Assignments[len(proj.Assignments)-1]
		log.Printf("INFO: project %s workflow %s advancing to new step %s with originial owner %s",
			projID, proj.Workflow.Name, nextStep.Name, firstA.StaffMember.ComputingID)
		//    self.update(current_step: new_step, owner: first_owner)
		//    Assignment.create(project: self, staff_member: first_owner, step: new_step)
	} else {
		// any, unique or supervisor for this step. Someone must claim it, so set owner nil.
		log.Printf("INFO: project %s workflow %s advancing to new step %s with no owner set", projID, proj.Workflow.Name, nextStep.Name)
		//    self.update(current_step: new_step, owner: nil)
	}

	c.String(http.StatusNotImplemented, "not yet")
}

func (svc *serviceContext) nextStep(proj *project, nextStepID uint, ownerID uint) error {
	proj.CurrentStepID = nextStepID
	return nil
}

func (svc *serviceContext) finishStep(proj *project) error {
	return nil
}
