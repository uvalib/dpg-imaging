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
	Status          uint        `json:"status"` //enum status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing]
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
	Workflow          workflow       `json:"workflow"`
	UnitID            uint           `json:"-"`
	Unit              unit           `gorm:"foreignKey:UnitID" json:"unit"`
	OwnerID           *uint          `json:"-"`
	Owner             *staffMember   `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Assignments       []*assignment  `gorm:"foreignKey:ProjectID" json:"assignments"`
	CurrentStepID     uint           `json:"-"`
	CurrentStep       step           `gorm:"foreignKey:CurrentStepID" json:"currentStep"`
	DueOn             *time.Time     `json:"dueOn,omitempty"`
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

	var proj project
	dbReq := svc.getBaseProjectQuery().Where("projects.id=?", projID)
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
	whereQ := " AND units.unit_status != 'canceled'"
	if filterIdx != 3 {
		whereQ += " AND units.unit_status != 'done'"
	}

	qWorkflow := c.Query("workflow")
	if qWorkflow != "" {
		id, _ := strconv.Atoi(qWorkflow)
		whereQ += fmt.Sprintf(" AND projects.workflow_id=%d", id)
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
		whereQ += fmt.Sprintf(" AND metadata.call_number like \"%%%s%%\"", qCallNum)
	}
	qCustomer := c.Query("customer")
	if qCustomer != "" {
		whereQ += fmt.Sprintf(" AND customers.last_name like \"%%%s%%\"", qCustomer)
	}
	qAgency := c.Query("agency")
	if qAgency != "" {
		id, _ := strconv.Atoi(qAgency)
		whereQ += fmt.Sprintf(" AND agencies.id = %d", id)
	}
	qUnitID := c.Query("unit")
	if qUnitID != "" {
		id, _ := strconv.Atoi(qUnitID)
		whereQ += fmt.Sprintf(" AND units.id = %d", id)
		log.Printf("INFO: query for unit %d", id)
	}
	qOrderID := c.Query("order")
	if qOrderID != "" {
		id, _ := strconv.Atoi(qOrderID)
		whereQ += fmt.Sprintf(" AND orders.id = %d", id)
		log.Printf("INFO: query for order %d", id)
	}

	type projResp struct {
		TotalMe         int64     `json:"totalMe"`
		TotalActive     int64     `json:"totalActive"`
		TotalError      int64     `json:"totalError"`
		TotalUnassigned int64     `json:"totalUnassigned"`
		TotalFinished   int64     `json:"totalFinished"`
		Page            uint      `json:"page"`
		PageSize        uint      `json:"pageSize"`
		Projects        []project `json:"projects"`
	}
	out := projResp{Page: uint(page), PageSize: uint(pageSize)}

	for idx, q := range filterQ {
		var total int64
		countQ := q + whereQ
		cr := svc.getBaseCountsQuery().Model(&project{}).Distinct("projects.id").Where(countQ).Count(&total)
		if cr.Error != nil {
			log.Printf("WARNING: unable to get count of projects: %s", cr.Error.Error())
			total = 0
		}
		if idx == 0 {
			out.TotalMe = total
		} else if idx == 1 {
			out.TotalActive = total
		} else if idx == 2 {
			out.TotalUnassigned = total
		} else if idx == 3 {
			out.TotalFinished = total
		} else {
			out.TotalError = total
		}
	}

	orderStr := "due_on asc"
	if filter == "finished" {
		orderStr = "finished_at desc"
	}
	whereQ = filterQ[filterIdx] + whereQ
	resp := svc.getBaseProjectQuery().Distinct().
		Offset(offset).Limit(pageSize).Order(orderStr).
		Where(whereQ).
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
	if userID == "0" {
		log.Printf("INFO: user %s is assigning clearing %s assignment", claims.ComputeID, projID)
	} else {
		log.Printf("INFO: user %s is assigning project %s to user ID %s", claims.ComputeID, projID, userID)
	}

	log.Printf("INFO: looking up project %s", projID)
	var proj project
	pTx := svc.getBaseProjectQuery().Where("projects.id=?", projID)
	resp := pTx.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	// In a clear POST, userID will be 0
	if userID == "0" {
		now := time.Now()
		msg := fmt.Sprintf("<p>Admin user canceled assignment to %s</p>", proj.Owner.ComputingID)
		newNote := note{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: *proj.OwnerID,
			NoteType: 0, Note: msg, CreatedAt: &now, UpdatedAt: &now}
		err := svc.DB.Model(&proj).Association("Notes").Append(&newNote)
		if err != nil {
			log.Printf("ERROR: unable to add note to project %d: %s", proj.ID, err.Error())
			return
		}

		activeAssign := proj.Assignments[0]
		log.Printf("INFO: mark assignment %d as reassigied", activeAssign.ID)
		activeAssign.Status = StepReassigned
		r := svc.DB.Model(&activeAssign).Select("Status").Updates(activeAssign)
		if r.Error != nil {
			log.Printf("ERROR: unable to mark project %d active assignment as reassigned: %s", proj.ID, r.Error.Error())
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
		c.JSON(http.StatusOK, proj)
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

	if proj.OwnerID != nil && *proj.OwnerID == owner.ID {
		log.Printf("INFO: project %d owner is already %d, no change needed", proj.ID, *proj.OwnerID)
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
	if proj.OwnerID != nil {
		activeAssign := proj.Assignments[0]
		log.Printf("INFO: marking assignment %d as reassigned", activeAssign.ID)
		activeAssign.Status = StepReassigned
		r := svc.DB.Model(&activeAssign).Select("Status").Updates(activeAssign)
		if r.Error != nil {
			log.Printf("ERROR: unable to mark project %d active assignment as reassigned: %s", proj.ID, r.Error.Error())
			c.String(http.StatusInternalServerError, r.Error.Error())
			return
		}
	}

	log.Printf("INFO: create assigmment for new owner %d", owner.ID)
	now := time.Now()
	newA := assignment{ProjectID: proj.ID, StepID: proj.CurrentStepID, StaffMemberID: owner.ID, AssignedAt: &now}
	resp = svc.DB.Create(&newA)
	if resp.Error != nil {
		log.Printf("ERROR: unable to create new assignment for project %d step %d: %s", proj.ID, proj.CurrentStepID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: new owner %d for project %d", owner.ID, proj.ID)
	proj.OwnerID = &owner.ID
	proj.Owner = nil
	resp = svc.DB.Model(&proj).Select("OwnerID").Updates(proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable set project %d owner to %d: %s", proj.ID, owner.ID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: project %d assigned to user %d", proj.ID, owner.ID)
	resp = pTx.First(&proj)
	c.JSON(http.StatusOK, proj)
}

func (svc *serviceContext) updateProject(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var updateData struct {
		ContainerTypeID uint   `json:"containerTypeID"`
		CategoryID      uint   `json:"categoryID"`
		Condition       uint   `json:"condition"`
		Note            string `json:"note"`
		OCRHintID       uint   `json:"ocrHintID"`
		OCRLanguageHint string `json:"ocrLangage"`
		OCRMasterFiles  bool   `json:"ocrMasterFiles"`
	}

	qpErr := c.ShouldBindJSON(&updateData)
	if qpErr != nil {
		log.Printf("ERROR: invalid update project %s payload: %v", projID, qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: user %s is updating project %s: %+v", claims.ComputeID, projID, updateData)

	var proj project
	dbReq := svc.getBaseProjectQuery().Where("projects.id=?", projID)
	resp := dbReq.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: update data for project %s", projID)
	proj.CategoryID = updateData.CategoryID
	proj.ItemCondition = updateData.Condition
	proj.ConditionNote = updateData.Note
	if updateData.ContainerTypeID > 0 && proj.Workflow.Name == "Manuscript" {
		proj.ContainerTypeID = &updateData.ContainerTypeID
	}
	r := svc.DB.Model(&proj).Select("ContainerTypeID", "CategoryID", "ItemCondition", "ConditionNote").Updates(proj)
	if r.Error != nil {
		log.Printf("ERROR: unable to update data for project %s: %s", projID, r.Error.Error())
		c.String(http.StatusInternalServerError, r.Error.Error())
		return
	}

	log.Printf("INFO: update OCR settings for project %s", projID)
	proj.Unit.OCRMasterFiles = updateData.OCRMasterFiles
	r = svc.DB.Model(&proj.Unit).Select("OCRMasterFiles").Updates(proj.Unit)
	if r.Error != nil {
		log.Printf("ERROR: unable to update unit OCR settings for project %s: %s", projID, r.Error.Error())
		c.String(http.StatusInternalServerError, r.Error.Error())
		return
	}
	if updateData.OCRHintID > 0 {
		proj.Unit.Metadata.OCRHintID = updateData.OCRHintID
		r = svc.DB.Model(&proj.Unit.Metadata).Select("OCRHintID").Updates(proj.Unit.Metadata)
		if r.Error != nil {
			log.Printf("ERROR: unable to update OCR Hint for project %s: %s", projID, r.Error.Error())
			c.String(http.StatusInternalServerError, r.Error.Error())
			return
		}

	}
	if updateData.OCRLanguageHint != "" {
		proj.Unit.Metadata.OCRLanguageHint = updateData.OCRLanguageHint
		r = svc.DB.Model(&proj.Unit.Metadata).Select("OCRLanguageHint").Updates(proj.Unit.Metadata)
		if r.Error != nil {
			log.Printf("ERROR: unable to update OCR Language for project %s: %s", projID, r.Error.Error())
			c.String(http.StatusInternalServerError, r.Error.Error())
			return
		}
	}

	dbReq.First(&proj)
	c.JSON(http.StatusOK, proj)
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
	dbReq := svc.getBaseProjectQuery().Where("projects.id=?", projID)
	resp := dbReq.First(&proj)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: get current equipment for workstation %d", equipPost.WorkstationID)
	var ws workstation
	resp = svc.DB.Preload("Equipment").First(&ws, equipPost.WorkstationID)
	if resp.Error != nil {
		log.Printf("ERROR: unable to get workstation %d: %s", equipPost.WorkstationID, resp.Error.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
		return
	}

	log.Printf("INFO: remove all equipment for project %s", projID)
	dbErr := svc.DB.Model(&proj).Association("Equipment").Clear()
	if dbErr != nil {
		log.Printf("ERROR: unable to clear existing equipment from project %s: %s", projID, dbErr.Error())
		c.String(http.StatusInternalServerError, resp.Error.Error())
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
	r := svc.DB.Model(&proj).Select("WorkstationID", "CaptureResolution", "ResizedResolution", "ResolutionNote").Updates(proj)
	if r.Error != nil {
		log.Printf("ERROR: unable to set workstation for project %s: %s", projID, r.Error.Error())
		c.String(http.StatusInternalServerError, r.Error.Error())
		return
	}

	dbReq.First(&proj)
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

	// Original owner
	if proj.CurrentStep.OwnerType == 3 {
		origAssign := proj.Assignments[len(proj.Assignments)-1]
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
	// NOTES:
	//  The Joins() calls all allow association table names to be used in nested where clauses by the caller.
	//  The Preload() calls preload all models with DB data
	//  LEFT OUTER joins are for data that is optional
	return svc.DB.Preload(clause.Associations).
		Joins("LEFT OUTER JOIN assignments on assignments.project_id=projects.id").
		Joins("LEFT OUTER JOIN steps as assignstep on assignments.step_id=assignstep.id").
		Joins("LEFT OUTER JOIN staff_members on assignments.staff_member_id=staff_members.id").
		Joins("INNER JOIN units on units.id=projects.unit_id").
		Joins("INNER JOIN metadata on metadata.id=units.metadata_id").
		Joins("INNER JOIN orders on orders.id=units.order_id").
		Joins("INNER JOIN customers on customers.id=orders.customer_id").
		Joins("LEFT OUTER JOIN agencies on agencies.id=orders.agency_id").
		Joins("LEFT OUTER JOIN notes on notes.project_id = projects.id").
		Preload("Assignments", func(db *gorm.DB) *gorm.DB { // specify the order of the associations in  Preload, not Joins
			return db.Order("assignments.assigned_at DESC")
		}).
		Preload("Notes", func(db *gorm.DB) *gorm.DB { // specify the order of the associations in  Preload, not Joins
			return db.Order("notes.created_at DESC")
		}).
		Preload("Assignments.StaffMember").      // explicitly preload nested assignment owner
		Preload("Unit." + clause.Associations).  // this is a shorthand to load all associations directy under unit
		Preload("Unit.Metadata.OCRHint").        // OCRHint is deeply nested, so need to preload explicitly
		Preload("Unit.Order.Customer").          // customer is deeply nested, so need to preload explicitly
		Preload("Unit.Order.Agency").            // agency is deeply nested, so need to preload explicitly
		Preload("Notes." + clause.Associations). // preload all associations under notes
		Preload("Workflow.Steps")                // explicitly preload nested workflow steps
}

func (svc *serviceContext) getBaseCountsQuery() (tx *gorm.DB) {
	return svc.DB.
		Joins("LEFT OUTER JOIN assignments on assignments.project_id=projects.id").
		Joins("LEFT OUTER JOIN steps as assignstep on assignments.step_id=assignstep.id").
		Joins("LEFT OUTER JOIN staff_members on assignments.staff_member_id=staff_members.id").
		Joins("INNER JOIN units on units.id=projects.unit_id").
		Joins("INNER JOIN metadata on metadata.id=units.metadata_id").
		Joins("INNER JOIN orders on orders.id=units.order_id").
		Joins("INNER JOIN customers on customers.id=orders.customer_id").
		Joins("LEFT OUTER JOIN agencies on agencies.id=orders.agency_id").
		Joins("LEFT OUTER JOIN notes on notes.project_id = projects.id")
}
