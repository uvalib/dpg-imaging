package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type workflow struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Steps       []step `gorm:"foreignKey:WorkflowID" json:"steps"`
}

type containerType struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	HasFolders bool   `json:"hasFolders"`
}

type assignment struct {
	ID            uint        `json:"id"`
	ProjectID     uint        `json:"projectID"`
	StepID        uint        `json:"stepID"`
	StaffMemberID uint        `json:"-"`
	StaffMember   staffMember `gorm:"foreignKey:StaffMemberID" json:"staffMember"`
	AssignedAt    *time.Time  `json:"assignedAt,omitempty"`
	StartedAt     *time.Time  `json:"startedAt,omitempty"`
	FinishedAt    *time.Time  `json:"finishedAt,omitempty"`
	Status        uint        `json:"status,omitempty"`
}

type step struct {
	ID          uint   `json:"id"`
	WorkflowID  uint   `json:"workflowID"`
	StepType    uint   `json:"stepType"`
	Name        string `json:"name"`
	Description string `json:"description"`
	NextStepID  uint   `json:"nextStepID"`
	FailStepID  uint   `json:"failStepID"`
	OwnerType   uint   `json:"ownerType"`
}

type category struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type workstation struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status uint   `json:"status"`
}

type project struct {
	ID                uint          `json:"id"`
	WorkflowID        uint          `json:"-"`
	Workflow          workflow      `json:"workflow"`
	UnitID            uint          `json:"-"`
	Unit              unit          `gorm:"foreignKey:UnitID" json:"unit"`
	OwnerID           uint          `json:"-"`
	Owner             staffMember   `gorm:"foreignKey:OwnerID" json:"owner"`
	Assignments       []assignment  `gorm:"foreignKey:ProjectID" json:"assignments"`
	CurrentStepID     uint          `json:"-"`
	CurrentStep       step          `gorm:"foreignKey:CurrentStepID" json:"currentStep"`
	Priority          uint          `json:"priority"`
	DueOn             *time.Time    `json:"dueOn,omitempty"`
	ItemCondition     uint          `json:"itemCOndition,omitempty"`
	AddedAt           *time.Time    `json:"addedAt,omitempty"`
	StartedAt         *time.Time    `json:"startedAt,omitempty"`
	FinishedAt        *time.Time    `json:"finishedAt,omitempty"`
	CategoryID        uint          `json:"-"`
	Category          category      `gorm:"foreignKey:CategoryID" json:"category"`
	VIUNumber         string        `json:"viuNumner"`
	CaptureResolution uint          `json:"captureResolution"`
	ResizedResolution uint          `json:"resizedResolution"`
	ResolutionNote    string        `json:"resolutionNote"`
	WorkstationID     uint          `json:"-"`
	Workstation       workstation   `json:"workstation"`
	ConditionNote     string        `json:"conditionNote"`
	ContainerTypeID   uint          `json:"-"`
	ContainerType     containerType `gorm:"foreignKey:ContainerTypeID" json:"containerType"`
}

func (svc *serviceContext) getProjects(c *gin.Context) {
	log.Printf("INFO: get projects")
	pageSize := 20
	pageQ := c.Query("page")
	if pageQ == "" {
		pageQ = "1"
	}
	page, err := strconv.Atoi(pageQ)
	if err != nil {
		log.Printf("ERROR: invalid page %s specified, default to 1", pageQ)
		page = 1
	}
	offset := (page - 1) * pageSize

	type projResp struct {
		Total    int64     `json:"total"`
		Page     uint      `json:"page"`
		PageSize uint      `json:"pageSize"`
		Projects []project `json:"projects"`
	}
	out := projResp{Page: uint(page), PageSize: uint(pageSize)}

	cr := svc.DB.Model(&project{}).Where("finished_at is null").Count(&out.Total)
	if cr.Error != nil {
		log.Printf("ERROR: unable to get count of projects: %s", cr.Error.Error())
		c.String(http.StatusInternalServerError, cr.Error.Error())
		return
	}

	resp := svc.DB.Preload(clause.Associations).
		Preload("Unit.Metadata").Preload("Unit.IntendedUse"). // must preload nested explicitly
		Preload("Unit.Order").Preload("Unit.Order.Customer").
		Preload("Assignments.StaffMember").
		Preload("Workflow.Steps").
		Offset(offset).Limit(pageSize).
		Order("due_on asc").
		Where("finished_at is null").Find(&out.Projects)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get projects: %s", resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) claimProject(c *gin.Context) {
	projID := c.Param("id")
	claims, err := getJWTClaims(c)
	if err != nil {
		log.Printf("ERROR: request to claim project with invalid JWT: %s", err.Error())
		c.String(http.StatusUnauthorized, "unauthorized claim request")
		return
	}
	log.Printf("INFO: user %s is claiming project %s", claims.ComputeID, projID)
	var proj project
	pTx := svc.DB.Preload(clause.Associations).
		Preload("Unit.Metadata").Preload("Unit.IntendedUse"). // must preload nested explicitly
		Preload("Unit.Order").Preload("Unit.Order.Customer").
		Preload("Assignments.StaffMember").
		Preload("Workflow.Steps").
		Where("id=?", projID)
	resp := pTx.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	err = svc.canClaimProject(claims, &proj)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("INFO: user %s can claim project %d; updating data", claims.ComputeID, proj.ID)
	// If someone else has this assignment, flag it as reassigned. Do not mark the finished time as it was never actually finished
	if proj.OwnerID > 0 {
		activeAssign := proj.Assignments[len(proj.Assignments)-1]
		activeAssign.Status = 5
		r := svc.DB.Save(&activeAssign)
		if r.Error != nil {
			log.Printf("ERROR: unable to mark project %d active assignment as reassigned: %s", proj.ID, r.Error.Error())
			c.String(http.StatusInternalServerError, r.Error.Error())
			return
		}
	}

	now := time.Now()
	newA := assignment{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: claims.UserID, AssignedAt: &now}
	resp = svc.DB.Create(&newA)
	if resp.Error != nil {
		log.Printf("ERROR: unable to create new assignment for project %d step %d: %s", proj.ID, proj.CurrentStepID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	resp = svc.DB.Exec("UPDATE projects set owner_id=? where id=?", claims.UserID, proj.ID)
	if resp.Error != nil {
		log.Printf("ERROR: unable set project %d owner to %d: %s", proj.ID, claims.UserID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	// reload the project
	resp = pTx.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to reload project %d: %s", proj.ID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: project %d claimed by user %s", proj.ID, claims.ComputeID)
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) getProjectCandidates(c *gin.Context) {
	projID := c.Param("id")
	var proj project
	resp := svc.DB.Preload(clause.Associations).Where("id=?", projID).First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: project %s not found", projID)
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}
	candidates, err := svc.projectCandidates(&proj)
	if err != nil {
		log.Printf("ERROR: unable to get candidates for prokect %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, candidates)
}

func (svc *serviceContext) projectCandidates(proj *project) ([]staffMember, error) {
	out := make([]staffMember, 0)
	resp := svc.DB.
		Distinct("staff_members.id", "computing_id", "first_name", "last_name", "role").
		Joins("inner join staff_skills on staff_member_id=staff_members.id").
		Where("(role <= 1 and is_active=1) or (category_id=? and role=2 and is_active=1)", proj.CategoryID).
		Order("last_name asc").
		Find(&out)
	if resp.Error != nil {
		return nil, resp.Error
	}
	return out, nil
}

func (svc *serviceContext) canClaimProject(userClaims *jwtClaims, proj *project) error {
	// Admins can caim anything
	if userClaims.Role == "supervisor" || userClaims.Role == "admin" {
		return nil
	}

	if proj.OwnerID != 0 && !(userClaims.Role == "supervisor" || userClaims.Role == "admin") {
		log.Printf("INFO: user %s has already claimed project %d", proj.Owner.ComputingID, proj.ID)
		return fmt.Errorf(fmt.Sprintf("%s has already claimed this project", proj.Owner.ComputingID))
	}

	candidates, err := svc.projectCandidates(proj)
	if err != nil {
		log.Printf("ERROR: unable to get candidates for project category %d: %s", proj.CategoryID, err.Error())
		candidates = make([]staffMember, 0)
	}
	canClaim := false
	for _, sm := range candidates {
		if sm.ComputingID == userClaims.ComputeID {
			canClaim = true
			break
		}
	}
	if canClaim == false {
		log.Printf("INFO: user %s does not have the skills to claim project %d", userClaims.ComputeID, proj.ID)
		return fmt.Errorf("%s does not have the required skills to claim this project", proj.Owner.ComputingID)
	}

	// OwnerType: [:any_owner, :prior_owner, :unique_owner, :original_owner, :supervisor_owner]
	if proj.CurrentStep.OwnerType == 0 {
		// any owner
		return nil
	}
	if proj.CurrentStep.OwnerType == 4 && !(userClaims.Role == "supervisor" || userClaims.Role == "admin") {
		log.Printf("INFO: project %d requries a supervisor owner, and %s is not a supervisor", proj.ID, userClaims.ComputeID)
		return fmt.Errorf("This project requires a supervisor to claim")
	}

	// unique owner?
	if proj.CurrentStep.OwnerType == 2 {
		for _, a := range proj.Assignments {
			if a.StaffMember.ComputingID == userClaims.ComputeID {
				log.Printf("INFO: project %d requires a unique owner, but %s has previously claimed it", proj.ID, userClaims.ComputeID)
				return fmt.Errorf("This project requires a unique owner")
			}
		}
		return nil
	}

	// prior owner?
	if proj.CurrentStep.OwnerType == 1 {
		lastAssign := proj.Assignments[len(proj.Assignments)-1]
		if lastAssign.StaffMember.ComputingID != userClaims.ComputeID {
			log.Printf("INFO: project %d requires prior owner %s to claim, not %s",
				proj.ID, lastAssign.StaffMember.ComputingID, userClaims.ComputeID)
			return fmt.Errorf("This project requires a unique owner")
		}
		return nil
	}

	// original owner?
	if proj.CurrentStep.OwnerType == 3 {
		origAssign := proj.Assignments[0]
		if origAssign.StaffMember.ComputingID != userClaims.ComputeID {
			log.Printf("INFO: project %d requires original owner %s to claim, not %s",
				proj.ID, origAssign.StaffMember.ComputingID, userClaims.ComputeID)
			return fmt.Errorf("This project can only be cleimed by the original owner %s", origAssign.StaffMember.ComputingID)
		}
		return nil
	}

	return fmt.Errorf("You cannot claim this project")
}
