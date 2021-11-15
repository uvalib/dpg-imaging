package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	ID              uint        `json:"id"`
	ProjectID       uint        `json:"projectID"`
	StepID          uint        `json:"stepID"`
	StaffMemberID   uint        `json:"-"`
	StaffMember     staffMember `gorm:"foreignKey:StaffMemberID" json:"staffMember"`
	AssignedAt      *time.Time  `json:"assignedAt,omitempty"`
	StartedAt       *time.Time  `json:"startedAt,omitempty"`
	FinishedAt      *time.Time  `json:"finishedAt,omitempty"`
	DurationMinutes uint        `json:"durationMinutes"`
	Status          uint        `json:"status"`
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

type equipment struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	SerialNumber string `json:"serialNumber"`
}

type workstation struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status uint   `json:"status"`
}

type problem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

type note struct {
	ID            uint        `json:"id"`
	ProjectID     uint        `json:"-"`
	StepID        uint        `json:"stepID"`
	NoteType      uint        `json:"type"`
	Note          string      `json:"text"`
	CreatedAt     *time.Time  `json:"createdAt,omitempty"`
	Problems      []problem   `gorm:"many2many:notes_problems"`
	StaffMemberID uint        `json:"-"`
	StaffMember   staffMember `gorm:"foreignKey:StaffMemberID" json:"staffMember"`
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
	DueOn             *time.Time    `json:"dueOn,omitempty"`
	AddedAt           *time.Time    `json:"addedAt,omitempty"`
	StartedAt         *time.Time    `json:"startedAt,omitempty"`
	FinishedAt        *time.Time    `json:"finishedAt,omitempty"`
	CategoryID        uint          `json:"-"`
	Category          category      `gorm:"foreignKey:CategoryID" json:"category"`
	CaptureResolution uint          `json:"captureResolution"`
	ResizedResolution uint          `json:"resizedResolution"`
	ResolutionNote    string        `json:"resolutionNote"`
	WorkstationID     uint          `json:"-"`
	Workstation       workstation   `json:"workstation"`
	ItemCondition     uint          `json:"itemCondition,omitempty"`
	ConditionNote     string        `json:"conditionNote,omitempty"`
	ContainerTypeID   uint          `json:"-"`
	ContainerType     containerType `gorm:"foreignKey:ContainerTypeID" json:"containerType,omitempty"`
	Notes             []note        `gorm:"foreignKey:ProjectID" json:"notes,omitempty"`
	Equipment         []equipment   `gorm:"many2many:project_equipment" json:"equipment,omitempty"`
}

func (svc *serviceContext) getProject(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	log.Printf("INFO: user %s is requesting project %s details", claims.ComputeID, projID)

	var proj project
	dbReq := svc.getBaseProjectQuery().Where("id=?", projID)
	resp := dbReq.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) getProjects(c *gin.Context) {
	claims := getJWTClaims(c)
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

	filters := []string{"me", "active", "unassigned", "finished"}
	filterQ := []string{
		fmt.Sprintf("owner_id=%d and finished_at is null", claims.UserID),
		"finished_at is null",
		"finished_at is null and owner_id is null",
		"finished_at is not null",
	}
	filter := c.Query("filter")
	if filter == "" {
		filter = "active"
	}
	filterIdx := -1
	for idx, f := range filters {
		if f == filter {
			filterIdx = idx
			break
		}
	}
	if filterIdx == -1 {
		log.Printf("ERROR: invalid filter %s specified", filter)
		c.String(http.StatusBadRequest, fmt.Sprintf("%s is an invalid filter", filter))
		return
	}
	log.Printf("INFO: user %s requests projects page %d", claims.ComputeID, page)

	type projResp struct {
		Total    int64     `json:"total"`
		Page     uint      `json:"page"`
		PageSize uint      `json:"pageSize"`
		Projects []project `json:"projects"`
	}
	out := projResp{Page: uint(page), PageSize: uint(pageSize)}

	cr := svc.DB.Model(&project{}).Where(filterQ[filterIdx]).Count(&out.Total)
	if cr.Error != nil {
		log.Printf("ERROR: unable to get count of projects: %s", cr.Error.Error())
		c.String(http.StatusInternalServerError, cr.Error.Error())
		return
	}

	resp := svc.getBaseProjectQuery().
		Offset(offset).Limit(pageSize).Order("due_on asc").
		Where(filterQ[filterIdx]).
		Find(&out.Projects)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get projects: %s", resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) assignProject(c *gin.Context) {
	projID := c.Param("id")
	userID := c.Param("uid")
	claims := getJWTClaims(c)
	log.Printf("INFO: user %s is assigning project %s to user ID %s", claims.ComputeID, projID, userID)

	log.Printf("INFO: looking up project %s", projID)
	var proj project
	pTx := svc.getBaseProjectQuery().Where("id=?", projID)
	resp := pTx.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: looking up new project %s owner %s", projID, userID)
	var owner staffMember
	resp = svc.DB.First(&owner, userID)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get new owner %s for project %s: %s", userID, projID, resp.Error.Error())
		c.String(http.StatusBadRequest, resp.Error.Error())
		return
	}

	if proj.OwnerID == owner.ID {
		log.Printf("INFO: project %d owner is already %d, no change needed", proj.ID, proj.OwnerID)
		c.JSON(http.StatusOK, proj)
		return
	}

	err := svc.canAssignProject(&owner, claims, &proj)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("INFO: user %s can assign project %d to %s; updating data", claims.ComputeID, proj.ID, owner.ComputingID)

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
	newA := assignment{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: owner.ID, AssignedAt: &now}
	resp = svc.DB.Create(&newA)
	if resp.Error != nil {
		log.Printf("ERROR: unable to create new assignment for project %d step %d: %s", proj.ID, proj.CurrentStepID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	resp = svc.DB.Exec("UPDATE projects set owner_id=? where id=?", owner.ID, proj.ID)
	if resp.Error != nil {
		log.Printf("ERROR: unable set project %d owner to %d: %s", proj.ID, owner.ID, resp.Error.Error())
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

	log.Printf("INFO: project %d claimed by user %d", proj.ID, owner.ID)
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) canAssignProject(assignee *staffMember, assigner *jwtClaims, proj *project) error {
	// admin/supervisor can caim or assign anything
	assigneeRole := assignee.roleString()
	if assigneeRole == "supervisor" || assigneeRole == "admin" || assigner.Role == "supervisor" || assigner.Role == "admin" {
		return nil
	}

	// OwnerType: [:any_owner, :prior_owner, :unique_owner, :original_owner, :supervisor_owner]

	// Any owner
	if proj.CurrentStep.OwnerType == 0 {
		return nil
	}

	// Prior owner
	if proj.CurrentStep.OwnerType == 1 {
		lastAssign := proj.Assignments[len(proj.Assignments)-1]
		if lastAssign.StaffMember.ComputingID != assignee.ComputingID {
			log.Printf("INFO: project %d requires prior owner %s to claim, not %s",
				proj.ID, lastAssign.StaffMember.ComputingID, assignee.ComputingID)
			return fmt.Errorf("This project requires prior owner %s, not %s", lastAssign.StaffMember.ComputingID, assignee.ComputingID)
		}
		return nil
	}

	// Unique owner
	if proj.CurrentStep.OwnerType == 2 {
		for _, a := range proj.Assignments {
			if a.StaffMember.ComputingID == assignee.ComputingID {
				log.Printf("INFO: project %d requires a unique owner, but %s has previously claimed it", proj.ID, assignee.ComputingID)
				return fmt.Errorf("This project requires a unique owner, and %s has previously owned it", assignee.ComputingID)
			}
		}
		return nil
	}

	// Priginal owner
	if proj.CurrentStep.OwnerType == 3 {
		origAssign := proj.Assignments[0]
		if origAssign.StaffMember.ComputingID != assignee.ComputingID {
			log.Printf("INFO: project %d requires original owner %s to claim, not %s",
				proj.ID, origAssign.StaffMember.ComputingID, assignee.ComputingID)
			return fmt.Errorf("This project can only be claimed by the original owner %s", origAssign.StaffMember.ComputingID)
		}
		return nil
	}

	// Supervisor owner
	if proj.CurrentStep.OwnerType == 4 {
		if assigneeRole == "supervisor" || assigneeRole == "admin" {
			return nil
		}
		log.Printf("INFO: project %d requries a supervisor owner, and %s is not a supervisor", proj.ID, assignee.ComputingID)
		return fmt.Errorf("This project requires a supervisor to claim, and %s is not one", assignee.ComputingID)
	}

	log.Printf("ERROR: unrecognized owner type: %d", proj.CurrentStep.OwnerType)
	return fmt.Errorf("%s cannot claim this project (internal error)", assignee.ComputingID)
}

func (svc *serviceContext) getBaseProjectQuery() (tx *gorm.DB) {
	return svc.DB.Preload(clause.Associations).
		Preload("Assignments", func(db *gorm.DB) *gorm.DB {
			return db.Order("assignments.assigned_at DESC")
		}).
		Preload("Notes", func(db *gorm.DB) *gorm.DB {
			return db.Order("notes.created_at DESC")
		}).
		Preload("Unit.Metadata").Preload("Unit.IntendedUse").
		Preload("Unit.Order").Preload("Unit.Order.Customer").
		Preload("Assignments.StaffMember").
		Preload("Notes.StaffMember").
		Preload("Workflow.Steps").
		Preload("Notes.Problems")
}
