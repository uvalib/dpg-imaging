package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type workflow struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"isActive"`
	Steps       []step `gorm:"foreignKey:WorkflowID" json:"steps"`
}

type containerType struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	HasFolders bool   `json:"hasFolders"`
}

type assignStatusEnum uint

func (s assignStatusEnum) String() string {
	switch s {
	case Pending:
		return "pending"
	case Started:
		return "started"
	case Finished:
		return "finished"
	case Rejected:
		return "rejected"
	case Error:
		return "error"
	case Reassigned:
		return "reassigned"
	case Finalizing:
		return "finalizing"
	}
	return "unknown"
}

// Assignment status from rails enum: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing]
const (
	Pending    assignStatusEnum = 0
	Started    assignStatusEnum = 1
	Finished   assignStatusEnum = 2
	Rejected   assignStatusEnum = 3
	Error      assignStatusEnum = 4
	Reassigned assignStatusEnum = 5
	Finalizing assignStatusEnum = 6
)

type assignment struct {
	ID              uint             `json:"id"`
	ProjectID       uint             `json:"projectID"`
	StepID          uint             `json:"-"`
	Step            step             `gorm:"foreignKey:StepID" json:"step"`
	StaffMemberID   uint             `json:"-"`
	StaffMember     staffMember      `gorm:"foreignKey:StaffMemberID" json:"staffMember"`
	AssignedAt      *time.Time       `json:"assignedAt,omitempty"`
	StartedAt       *time.Time       `json:"startedAt,omitempty"`
	FinishedAt      *time.Time       `json:"finishedAt,omitempty"`
	DurationMinutes uint             `json:"durationMinutes"`
	Status          assignStatusEnum `json:"status"`
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

type projectEquipment struct {
	ID          uint       `json:"id"`
	ProjectID   uint       `json:"project_id"`
	EquipmentID uint       `json:"equipment_id"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

func (projectEquipment) TableName() string {
	return "project_equipment"
}

type workstation struct {
	ID        uint         `json:"id"`
	Name      string       `json:"name"`
	Status    uint         `json:"status"`
	Equipment []*equipment `gorm:"many2many:workstation_equipment" json:"equipment,omitempty"`
}

type project struct {
	ID                uint           `json:"id"`
	WorkflowID        uint           `json:"-"`
	Workflow          workflow       `gorm:"foreignKey:WorkflowID" json:"workflow"`
	UnitID            uint           `json:"-"`
	Unit              unit           `gorm:"foreignKey:UnitID" json:"unit"`
	OwnerID           *uint          `json:"-"`
	Owner             *staffMember   `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	ImageCount        int            `json:"imageCount"`
	Assignments       []*assignment  `gorm:"foreignKey:ProjectID" json:"assignments"`
	CurrentStepID     uint           `json:"-"`
	CurrentStep       step           `gorm:"foreignKey:CurrentStepID" json:"currentStep"`
	AddedAt           *time.Time     `json:"addedAt,omitempty"`
	StartedAt         *time.Time     `json:"startedAt,omitempty"`
	FinishedAt        *time.Time     `json:"finishedAt,omitempty"`
	TotalDurationMins *int64         `json:"totalDuration,omitempty"`
	CategoryID        uint           `json:"-"`
	Category          category       `gorm:"foreignKey:CategoryID" json:"category"`
	CaptureResolution uint           `json:"captureResolution"`
	ResizedResolution uint           `json:"resizedResolution"`
	ResolutionNote    string         `json:"resolutionNote"`
	WorkstationID     uint           `json:"-"`
	Workstation       workstation    `json:"workstation"`
	ItemCondition     uint           `json:"itemCondition"`
	ConditionNote     string         `json:"conditionNote,omitempty"`
	ContainerTypeID   *uint          `json:"-"`
	ContainerType     *containerType `gorm:"foreignKey:ContainerTypeID" json:"containerType"`
	Notes             []*note        `gorm:"foreignKey:ProjectID" json:"notes,omitempty"`
	Equipment         []*equipment   `gorm:"many2many:project_equipment" json:"equipment,omitempty"`
}

func (svc *serviceContext) getProject(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	log.Printf("INFO: user %s is requesting project %s details", claims.ComputeID, projID)

	var proj *project
	projQ := svc.DB.Model(&project{}).InnerJoins("Workflow").InnerJoins("Category").InnerJoins("Unit").
		Joins("Unit.Order").Joins("Unit.IntendedUse").Joins("Unit.Metadata").Joins("Unit.Metadata.OCRHint").
		Joins("Unit.Order.Customer").Joins("Unit.Order.Agency").Joins("ContainerType").
		Joins("Owner").Joins("CurrentStep").Preload("Equipment").Preload("Workstation")

	err := projQ.First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: get project %d assignments", proj.ID)
	err = svc.DB.Where("project_id=?", proj.ID).
		Joins("Step").Joins("StaffMember").
		Order("assigned_at DESC").Find(&proj.Assignments).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %d assignments: %s", proj.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: get project %d notes", proj.ID)
	err = svc.DB.Where("project_id=?", proj.ID).
		Joins("Step").Joins("StaffMember").Preload("Problems").
		Order("notes.created_at DESC").Find(&proj.Notes).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %d notes: %s", proj.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) updateProjecImageCount(c *gin.Context) {
	projID := c.Param("id")
	log.Printf("INFO: check project %s image count", projID)
	var proj project
	err := svc.DB.Preload("CurrentStep").Find(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	noCountSteps := []string{"Scan", "Process", "Create Metadata"}
	hasCount := !slices.Contains(noCountSteps, proj.CurrentStep.Name)

	if hasCount == false {
		log.Printf("INFO: project %s is on step %s, no image count yet", projID, proj.CurrentStep.Name)
		c.String(http.StatusOK, "ok")
		return
	}

	mfCnt := svc.getImageCount(proj.UnitID)
	if mfCnt != proj.ImageCount {
		log.Printf("INFO: update project %s image count to: %d", projID, mfCnt)
		proj.ImageCount = mfCnt
		err = svc.DB.Table("projects").Where("id = ?", proj.ID).Update("image_count", mfCnt).Error
		if err != nil {
			log.Printf("ERROR: unable to update project %s image count to %d: %s", projID, mfCnt, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		log.Printf("INFO: project %s has %d images; no update needed", projID, mfCnt)
	}
	c.String(http.StatusOK, fmt.Sprintf("%d", mfCnt))
}

func (svc *serviceContext) getImageCount(unitID uint) int {
	mfCnt := 0
	uidStr := padLeft(fmt.Sprintf("%d", unitID), 9)
	unitDir := path.Join(svc.ImagesDir, uidStr)
	mfRegex := regexp.MustCompile(`^\d{9}_\w{4,}\.tif$`)
	err := filepath.Walk(unitDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Printf("WARNING: directory traverse failed: %s", err.Error())
			return nil
		}

		if !f.IsDir() {
			fName := f.Name()
			if mfRegex.Match([]byte(fName)) {
				mfCnt++
				return nil
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("ERROR: unable to get image count for unit %d: %s", unitID, err.Error())
		mfCnt = 0
	}
	return mfCnt
}

func (svc *serviceContext) getProjectStatus(c *gin.Context) {
	projID := c.Param("id")
	var proj *project
	err := svc.DB.First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s for status check: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if proj.FinishedAt != nil {
		c.String(http.StatusOK, "finished")
	} else {
		var currAssgn assignment
		err = svc.DB.Where("project_id=?", proj.ID).Order("assigned_at desc").First(&currAssgn).Error
		if err != nil {
			log.Printf("ERROR: unable to get current assignemnt for project %s: %s", projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, currAssgn.Status.String())
	}
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

	filters := []string{"me", "active", "unassigned", "finished", "errors"}
	filterQ := []string{
		fmt.Sprintf("owner_id=%d and projects.finished_at is null", claims.UserID),
		"projects.finished_at is null and owner_id is not null",
		"projects.finished_at is null and owner_id is null",
		"projects.finished_at is not null",
		"(assignments.status=4 or assignstep.step_type=2) and projects.finished_at is null and assignments.finished_at is null",
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

	// ALWAYS exclude caneled projects. exclude done projects when filter is not finished.
	whereQ := " AND unit_status != 'canceled'"

	qWorkflow := c.Query("workflow")
	if qWorkflow != "" {
		id, _ := strconv.Atoi(qWorkflow)
		whereQ += fmt.Sprintf(" AND projects.workflow_id=%d", id)
	}
	qStep := c.Query("step")
	if qStep != "" {
		whereQ += fmt.Sprintf(" AND CurrentStep.name=\"%s\"", qStep)
	}
	qWorkstation := c.Query("workstation")
	if qWorkstation != "" {
		id, _ := strconv.Atoi(qWorkstation)
		whereQ += fmt.Sprintf(" AND workstation_id=%d", id)
	}
	qAssigned := c.Query("assigned")
	if qAssigned != "" {
		id, _ := strconv.Atoi(qAssigned)
		whereQ += fmt.Sprintf(" AND owner_id=%d", id)
	}
	qCallNum := c.Query("callnum")
	if qCallNum != "" {
		whereQ += fmt.Sprintf(" AND call_number like \"%%%s%%\"", qCallNum)
	}
	qCustomer := c.Query("customer")
	if qCustomer != "" {
		whereQ += fmt.Sprintf(" AND Unit__Order__Customer.last_name like \"%s%%\"", qCustomer)
	}
	qAgency := c.Query("agency")
	if qAgency != "" {
		id, _ := strconv.Atoi(qAgency)
		whereQ += fmt.Sprintf(" AND agency_id = %d", id)
	}
	qUnitID := c.Query("unit")
	if qUnitID != "" {
		id, _ := strconv.Atoi(qUnitID)
		whereQ += fmt.Sprintf(" AND unit_id = %d", id)
		log.Printf("INFO: query for unit %d", id)
	}
	qOrderID := c.Query("order")
	if qOrderID != "" {
		id, _ := strconv.Atoi(qOrderID)
		whereQ += fmt.Sprintf(" AND order_id = %d", id)
		log.Printf("INFO: query for order %d", id)
	}

	type projResp struct {
		TotalMe         int64      `json:"totalMe"`
		TotalActive     int64      `json:"totalActive"`
		TotalError      int64      `json:"totalError"`
		TotalUnassigned int64      `json:"totalUnassigned"`
		TotalFinished   int64      `json:"totalFinished"`
		Page            uint       `json:"page"`
		PageSize        uint       `json:"pageSize"`
		Projects        []*project `json:"projects"`
	}
	out := projResp{Page: uint(page), PageSize: uint(pageSize)}

	for idx, q := range filterQ {
		var total int64
		countQ := q + whereQ
		if idx != 3 {
			countQ += " AND unit_status != 'done'"
		}
		err = svc.getBaseSearchQuery().Where(countQ).Count(&total).Error
		if err != nil {
			log.Printf("WARNING: unable to get count of projects: %s", err.Error())
			total = 0
		}
		switch idx {
		case 0:
			out.TotalMe = total
		case 1:
			out.TotalActive = total
		case 2:
			out.TotalUnassigned = total
		case 3:
			out.TotalFinished = total
		default:
			out.TotalError = total
		}
	}

	orderStr := "date_due asc"
	if filter == "finished" {
		orderStr = "finished_at desc"
	}
	whereQ = filterQ[filterIdx] + whereQ
	if filterIdx != 3 {
		whereQ += " AND unit_status != 'done'"
	}
	err = svc.getBaseSearchQuery().
		Offset(offset).Limit(pageSize).Order(orderStr).
		Where(whereQ).
		Find(&out.Projects).Error
	if err != nil {
		log.Printf("ERROR: unable to get projects: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	for _, p := range out.Projects {
		err = svc.DB.Where("project_id=?", p.ID).
			Joins("Step").Joins("StaffMember").
			Order("assigned_at DESC").Find(&p.Assignments).Error
		if err != nil {
			log.Printf("ERROR: unable to get project %d assignments: %s", p.ID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) assignProject(c *gin.Context) {
	projID := c.Param("id")
	userID := c.Param("uid")
	claims := getJWTClaims(c)
	if userID == "0" {
		log.Printf("INFO: user %s is assigning clearing %s assignment", claims.ComputeID, projID)
	} else {
		log.Printf("INFO: user %s is assigning project %s to user ID %s", claims.ComputeID, projID, userID)
	}

	log.Printf("INFO: looking up project %s", projID)
	var proj project
	err := svc.DB.Joins("CurrentStep").Joins("Owner").First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s for reassign: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: lookup active assignent for project %s", projID)
	var activeAssign assignment
	err = svc.DB.Where("project_id=?", proj.ID).Joins("Step").Joins("StaffMember").
		Order("assigned_at DESC").Limit(1).Find(&activeAssign).Error
	if err != nil {
		log.Printf("ERROR: unable to get active assignment project %s reassign: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	out := struct {
		Owner       *staffMember `json:"owner"`
		Assignments []assignment `json:"assignments"`
		Notes       []note       `json:"notes"`
	}{}

	if userID == "0" {
		// In a clear POST, userID will be 0 and all owner lookups and error checks are skipped
		now := time.Now()
		msg := fmt.Sprintf("<p>Admin user canceled assignment to %s</p>", proj.Owner.ComputingID)
		newNote := note{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: *proj.OwnerID,
			NoteType: 0, Note: msg, CreatedAt: &now, UpdatedAt: &now}
		problemIDs := make([]uint, 0)
		_, err = svc.addNote(proj, newNote, problemIDs)
		if err != nil {
			log.Printf("ERROR: unable to add note to project %d: %s", proj.ID, err.Error())
			return
		}

		log.Printf("INFO: mark assignment %d as reassigned", activeAssign.ID)
		activeAssign.Status = StepReassigned
		err := svc.DB.Model(&activeAssign).Select("Status").Updates(activeAssign).Error
		if err != nil {
			log.Printf("ERROR: unable to mark project %d active assignment as reassigned: %s", proj.ID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		log.Printf("INFO: clear owner for project %s", projID)
		proj.OwnerID = nil
		proj.Owner = nil
		err = svc.DB.Model(&proj).Select("OwnerID").Updates(proj).Error
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		log.Printf("INFO: lookup new owner %s", projID)
		err = svc.DB.First(&out.Owner, userID).Error
		if err != nil {
			log.Printf("ERROR: unable to get new owner %s for project %s reassign: %s", userID, projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if proj.OwnerID != nil && *proj.OwnerID == out.Owner.ID {
			log.Printf("INFO: project %d owner is already %d, no change needed", proj.ID, *proj.OwnerID)
		} else {
			err = svc.canAssignProject(out.Owner, claims, &proj)
			if err != nil {
				c.String(http.StatusBadRequest, err.Error())
				return
			}

			log.Printf("INFO: user %s can assign project %d to %s; updating data", claims.ComputeID, proj.ID, out.Owner.ComputingID)

			// If someone else has this assignment, flag it as reassigned. Do not mark the finished time as it was never actually finished
			if proj.OwnerID != nil {
				log.Printf("INFO: marking assignment %d as reassigned", activeAssign.ID)
				activeAssign.Status = StepReassigned
				r := svc.DB.Model(&activeAssign).Select("Status").Updates(activeAssign)
				if r.Error != nil {
					log.Printf("ERROR: unable to mark project %d active assignment as reassigned: %s", proj.ID, r.Error.Error())
					c.String(http.StatusInternalServerError, r.Error.Error())
					return
				}
			}

			log.Printf("INFO: create assigmment for new owner %d", out.Owner.ID)
			now := time.Now()
			newA := assignment{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: out.Owner.ID, AssignedAt: &now}
			err = svc.DB.Create(&newA).Error
			if err != nil {
				log.Printf("ERROR: unable to create new assignment for project %d step %d: %s", proj.ID, proj.CurrentStepID, err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return
			}

			log.Printf("INFO: new owner %d for project %d", out.Owner.ID, proj.ID)
			proj.OwnerID = &out.Owner.ID
			proj.Owner = out.Owner
			err = svc.DB.Model(&proj).Select("OwnerID").Updates(proj).Error
			if err != nil {
				log.Printf("ERROR: unable set project %d owner to %d: %s", proj.ID, out.Owner.ID, err.Error())
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
			log.Printf("INFO: project %d assigned to user %d", proj.ID, out.Owner.ID)
		}
	}

	log.Printf("INFO: update data for reassigned project %d", proj.ID)
	err = svc.DB.Where("project_id=?", proj.ID).Joins("Step").Joins("StaffMember").Order("assigned_at DESC").Find(&out.Assignments).Error
	if err != nil {
		log.Printf("ERROR: unable to refresh assigned project %s assignments: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	err = svc.DB.Where("project_id=?", proj.ID).Joins("Step").Joins("StaffMember").Preload("Problems").Order("notes.created_at DESC").Find(&out.Notes).Error
	if err != nil {
		log.Printf("ERROR: unable to refresh assigned project %s notes: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) updateProject(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var updateData struct {
		ContainerTypeID uint           `json:"containerTypeID"`
		ContainerType   *containerType `json:"containerType"`
		CategoryID      uint           `json:"categoryID"`
		Category        category       `json:"category"`
		Condition       uint           `json:"condition"`
		Note            string         `json:"note"`
		OCRHintID       uint           `json:"ocrHintID"`
		OCRHint         ocrHint        `json:"ocrHint"`
		OCRLanguageHint string         `json:"ocrLangage"`
		OCRMasterFiles  bool           `json:"ocrMasterFiles"`
	}

	qpErr := c.ShouldBindJSON(&updateData)
	if qpErr != nil {
		log.Printf("ERROR: invalid update project %s payload: %v", projID, qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: user %s is updating project %s: %+v", claims.ComputeID, projID, updateData)

	log.Printf("INFO: lookup project %s", projID)
	var proj project
	err := svc.DB.Joins("Unit").First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s update: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if updateData.ContainerTypeID > 0 {
		log.Printf("INFO: lookup container type %d", updateData.ContainerTypeID)
		err = svc.DB.First(&updateData.ContainerType, updateData.ContainerTypeID).Error
		if err != nil {
			log.Printf("ERROR: unable to get container type %d project %s update: %s", updateData.ContainerTypeID, projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	log.Printf("INFO: lookup category %d", updateData.CategoryID)
	err = svc.DB.First(&updateData.Category, updateData.CategoryID).Error
	if err != nil {
		log.Printf("ERROR: unable to get category %d project %s update: %s", updateData.CategoryID, projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: lookup ocr hint %d", updateData.OCRHintID)
	err = svc.DB.First(&updateData.OCRHint, updateData.OCRHintID).Error
	if err != nil {
		log.Printf("ERROR: unable to get ocr hint %d project %s update: %s", updateData.OCRHintID, projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: update data for project %s", projID)
	proj.CategoryID = updateData.CategoryID
	proj.ItemCondition = updateData.Condition
	proj.ConditionNote = updateData.Note
	if updateData.ContainerTypeID > 0 && proj.Workflow.Name == "Manuscript" {
		proj.ContainerTypeID = &updateData.ContainerTypeID
	}
	err = svc.DB.Model(&proj).Select("ContainerTypeID", "CategoryID", "ItemCondition", "ConditionNote").Updates(proj).Error
	if err != nil {
		log.Printf("ERROR: unable to update data for project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: update OCR settings for project %s", projID)
	tgtUnit := unit{ID: proj.UnitID, OCRMasterFiles: updateData.OCRMasterFiles}
	err = svc.DB.Model(&tgtUnit).Select("OCRMasterFiles").Updates(tgtUnit).Error
	if err != nil {
		log.Printf("ERROR: unable to update unit OCR settings for project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if updateData.OCRHintID > 0 {
		tgtMD := metadata{ID: proj.Unit.MetadataID, OCRHintID: updateData.OCRHintID}
		err = svc.DB.Model(&tgtMD).Select("OCRHintID").Updates(tgtMD).Error
		if err != nil {
			log.Printf("ERROR: unable to update OCR Hint for project %s: %s", projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

	}
	if updateData.OCRLanguageHint != "" {
		tgtMD := metadata{ID: proj.Unit.MetadataID, OCRLanguageHint: updateData.OCRLanguageHint}
		err = svc.DB.Model(&tgtMD).Select("OCRLanguageHint").Updates(tgtMD).Error
		if err != nil {
			log.Printf("ERROR: unable to update OCR Language for project %s: %s", projID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(http.StatusOK, updateData)
}

func (svc *serviceContext) setProjectEquipment(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var equipPost struct {
		WorkstationID     uint   `json:"workstationID"`
		CaptureResolution uint   `json:"captureResolution"`
		ResizeResolution  uint   `json:"resizeResolution"`
		ResolutionNote    string `json:"resolutionNote"`
	}

	qpErr := c.ShouldBindJSON(&equipPost)
	if qpErr != nil {
		log.Printf("ERROR: invalid set equipment payload: %v", qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: user %s is setting project %s equipment: %+v", claims.ComputeID, projID, equipPost)

	var proj project
	err := svc.DB.First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s for equipment update: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: get current equipment for workstation %d", equipPost.WorkstationID)
	var ws workstation
	err = svc.DB.Preload("Equipment").First(&ws, equipPost.WorkstationID).Error
	if err != nil {
		log.Printf("ERROR: unable to get workstation %d: %s", equipPost.WorkstationID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: remove all equipment for project %s", projID)
	err = svc.DB.Model(&proj).Association("Equipment").Clear()
	if err != nil {
		log.Printf("ERROR: unable to clear existing equipment from project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: add equipment to project %s", projID)
	now := time.Now()
	for _, e := range ws.Equipment {
		pe := projectEquipment{ProjectID: proj.ID, EquipmentID: e.ID, CreatedAt: &now, UpdatedAt: &now}
		resp := svc.DB.Create(&pe)
		if resp.Error != nil {
			log.Printf("ERROR: unable to add equipment %s to project %s: %s", e.Name, projID, resp.Error.Error())
		}
	}

	log.Printf("INFO: set workstation for project %s", projID)
	proj.WorkstationID = equipPost.WorkstationID
	proj.Workstation = ws
	proj.CaptureResolution = equipPost.CaptureResolution
	proj.ResizedResolution = equipPost.ResizeResolution
	proj.ResolutionNote = equipPost.ResolutionNote
	err = svc.DB.Model(&proj).Select("WorkstationID", "CaptureResolution", "ResizedResolution", "ResolutionNote").Updates(proj).Error
	if err != nil {
		log.Printf("ERROR: unable to set workstation for project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: load updated equipment for project %d", proj.ID)
	err = svc.DB.Preload("Equipment").Preload("Workstation").First(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s for equipment update: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
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
		lastAssign := proj.Assignments[0]
		if lastAssign.StaffMember.ComputingID != assignee.ComputingID {
			log.Printf("INFO: project %d requires prior owner %s to claim, not %s",
				proj.ID, lastAssign.StaffMember.ComputingID, assignee.ComputingID)
			return fmt.Errorf("this project requires prior owner %s, not %s", lastAssign.StaffMember.ComputingID, assignee.ComputingID)
		}
		return nil
	}

	// Unique owner
	if proj.CurrentStep.OwnerType == 2 {
		for _, a := range proj.Assignments {
			if a.StaffMember.ComputingID == assignee.ComputingID {
				log.Printf("INFO: project %d requires a unique owner, but %s has previously claimed it", proj.ID, assignee.ComputingID)
				return fmt.Errorf("this project requires a unique owner, and %s has previously owned it", assignee.ComputingID)
			}
		}
		return nil
	}

	// Original owner
	if proj.CurrentStep.OwnerType == 3 {
		origAssign := proj.Assignments[len(proj.Assignments)-1]
		if origAssign.StaffMember.ComputingID != assignee.ComputingID {
			log.Printf("INFO: project %d requires original owner %s to claim, not %s",
				proj.ID, origAssign.StaffMember.ComputingID, assignee.ComputingID)
			return fmt.Errorf("this project can only be claimed by the original owner %s", origAssign.StaffMember.ComputingID)
		}
		return nil
	}

	// Supervisor owner
	if proj.CurrentStep.OwnerType == 4 {
		if assigneeRole == "supervisor" || assigneeRole == "admin" {
			return nil
		}
		log.Printf("INFO: project %d requries a supervisor owner, and %s is not a supervisor", proj.ID, assignee.ComputingID)
		return fmt.Errorf("this project requires a supervisor to claim, and %s is not one", assignee.ComputingID)
	}

	log.Printf("ERROR: unrecognized owner type: %d", proj.CurrentStep.OwnerType)
	return fmt.Errorf("%s cannot claim this project (internal error)", assignee.ComputingID)
}

func (svc *serviceContext) getBaseSearchQuery() (tx *gorm.DB) {
	return svc.DB.Model(&project{}).InnerJoins("Workflow").InnerJoins("Category").InnerJoins("Unit").
		Joins("Unit.Order").Joins("Unit.IntendedUse").Joins("Unit.Metadata").
		Joins("Unit.Order.Customer").Joins("Unit.Order.Agency").
		Joins("Owner").Joins("CurrentStep").
		Joins("LEFT OUTER JOIN assignments on assignments.project_id=projects.id").
		Joins("LEFT OUTER JOIN steps as assignstep on assignments.step_id=assignstep.id").
		Group("projects.id")
}
