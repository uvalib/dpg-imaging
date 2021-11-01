package main

import (
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
