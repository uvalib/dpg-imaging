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
}

type containerType struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	HasFolders bool   `json:"hasFolders"`
}

type staffMember struct {
	ID          uint   `json:"id"`
	ComputingID string `json:"computingID"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
}

type step struct {
	ID          uint `json:"id"`
	StepType    uint
	Name        string
	Description string
	NextStepID  uint
	FailStepID  uint
	OwnerType   uint
}

type category struct {
	ID   uint `json:"id"`
	Name string
}

type workstation struct {
	ID     uint `json:"id"`
	Name   string
	Statis uint
}

type project struct {
	ID                uint          `json:"id"`
	WorkflowID        uint          `json:"-"`
	Workflow          workflow      `json:"workflow"`
	UnitID            uint          `json:"-"`
	Unit              unit          `gorm:"foreignKey:UnitID" json:"unit"`
	OwnerID           uint          `json:"-"`
	Owner             staffMember   `gorm:"foreignKey:OwnerID" json:"owner"`
	CurrentStepID     uint          `json:"-"`
	Step              containerType `gorm:"foreignKey:CurrentStepID" json:"step"`
	Priority          uint
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
	pageSize := 10
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
	var out []project
	resp := svc.GDB.Preload(clause.Associations).
		Preload("Unit.Metadata").Preload("Unit.IntendedUse"). // must preload nested explicitly
		Offset(offset).Limit(pageSize).Find(&out)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get projects: %s", resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}
	c.JSON(http.StatusOK, out)
}
