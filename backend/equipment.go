package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type equipmentStatus uint

type equipment struct {
	ID           uint            `json:"id"`
	Type         string          `json:"type"`
	Name         string          `json:"name"`
	SerialNumber string          `json:"serialNumber"`
	Status       equipmentStatus `json:"status"` // active=0, inactive=1, retired=2
	CreatedAt    time.Time       `json:"-"`
	UpdatedAt    time.Time       `json:"-"`
}

type workstation struct {
	ID           uint            `json:"id"`
	Name         string          `json:"name"`
	Status       equipmentStatus `json:"status"` // [:active, :inactive, :retired]
	Equipment    []equipment     `gorm:"many2many:workstation_equipment" json:"equipment"`
	ProjectCount uint            `gorm:"column:proj_cnt" json:"projectCount"`
	CreatedAt    time.Time       `json:"-"`
	UpdatedAt    time.Time       `json:"-"`
}

func (svc *serviceContext) getEquipment(c *gin.Context) {
	log.Printf("INfO: get all equipment")
	var resp struct {
		Workstations []workstation `json:"workstations"`
		Equipment    []equipment   `json:"equipment"`
	}

	// only count projects that are less than one year old that have been started, but not finished
	lastYear := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	projQ := fmt.Sprintf("(select count(*) from projects p where workstations.id = p.workstation_id and finished_at is null and started_at > '%s') as proj_cnt", lastYear)
	err := svc.DB.Preload("Equipment", "status != ?", 2).Select(projQ, "workstations.*").
		Where("status != ?", 2).
		Order("name asc").Find(&resp.Workstations).Error
	if err != nil {
		log.Printf("ERROR: unable to get workstations: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = svc.DB.Where("status!=?", 2).Order("type asc, name asc").Find(&resp.Equipment).Error
	if err != nil {
		log.Printf("ERROR: unable to get equipment: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) updateEquipment(c *gin.Context) {
	equipID := c.Param("id")
	var tgtEquip equipment
	err := svc.DB.Find(&tgtEquip, equipID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("INFO: equipment %s not found", equipID)
			c.String(http.StatusNotFound, fmt.Sprintf("equipment %s not found", equipID))
		} else {
			log.Printf("ERROR: unable to load equipment %s: %s", equipID, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}
	log.Printf("INFO: update equipment %d", tgtEquip.ID)
	var req struct {
		Name         string          `json:"name"`
		SerialNumber string          `json:"serialNumber"`
		Status       equipmentStatus `json:"status"`
	}
	err = c.BindJSON(&req)
	if err != nil {
		log.Printf("ERROR: invalid update equipment request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	tgtEquip.Name = req.Name
	tgtEquip.SerialNumber = req.SerialNumber
	tgtEquip.Status = req.Status

	err = svc.DB.Model(&tgtEquip).Select("Name", "SerialNumber", "Status").Updates(tgtEquip).Error
	if err != nil {
		log.Printf("ERROR: unable to update equipment %d: %s", tgtEquip.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, "updated")
}

func (svc *serviceContext) createWorkstation(c *gin.Context) {
	log.Printf("INFO: create workstation")
	var req struct {
		Name string `json:"name"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		log.Printf("ERROR: invalid create workstation request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	newWorkstation := workstation{Name: req.Name}
	err = svc.DB.Omit("ProjectCount").Create(&newWorkstation).Error
	if err != nil {
		log.Printf("ERROR: unable to create workstation: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	newWorkstation.Equipment = make([]equipment, 0)
	c.JSON(http.StatusOK, newWorkstation)
}

func (svc *serviceContext) createEquipment(c *gin.Context) {
	log.Printf("INFO: create equipment")
	var req struct {
		Type         string `json:"type"`
		Name         string `json:"name"`
		SerialNumber string `json:"serialNumber"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		log.Printf("ERROR: invalid create equipment request: %s", err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	newEquip := equipment{Type: req.Type, Name: req.Name, SerialNumber: req.SerialNumber}
	err = svc.DB.Create(&newEquip).Error
	if err != nil {
		log.Printf("ERROR: unable to create equipment: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, newEquip)
}

func (svc *serviceContext) updateWorkstation(c *gin.Context) {
	wsID := c.Param("id")
	tgtStatus, _ := strconv.ParseInt(c.Query("status"), 10, 8)
	var tgtWS workstation
	err := svc.DB.Find(&tgtWS, wsID).Error
	if err != nil {
		log.Printf("ERROR: unable to load workstation %s: %s", wsID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO: update workstation %d status to %d", tgtWS.ID, tgtStatus)
	tgtWS.Status = equipmentStatus(tgtStatus)
	err = svc.DB.Model(&tgtWS).Update("status", tgtStatus).Error
	if err != nil {
		log.Printf("ERROR: unable to update workstation %d status: %s", tgtWS.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "udpated")
}

func (svc *serviceContext) updateWorkstationSetup(c *gin.Context) {
	wsID := c.Param("id")
	var tgtWS workstation
	err := svc.DB.Find(&tgtWS, wsID).Error
	if err != nil {
		log.Printf("ERROR: unable to load workstation %s to save setup: %s", wsID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO: update workstation %d setup", tgtWS.ID)
	var req struct {
		Setup []equipment `json:"setup"`
	}
	err = c.BindJSON(&req)
	if err != nil {
		log.Printf("ERROR: invalid update workstation %d setup request: %s", tgtWS.ID, err.Error())
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("INFO: validate new setup for workstation %d", tgtWS.ID)
	var scannerCount, bodyCount, lensCount, backCount uint
	var addValues []string
	for _, equip := range req.Setup {
		addValues = append(addValues, fmt.Sprintf("(%d,%d)", tgtWS.ID, equip.ID))
		if equip.Type == "Scanner" {
			scannerCount++
		}
		if equip.Type == "Lens" {
			lensCount++
		}
		if equip.Type == "CameraBody" {
			bodyCount++
		}
		if equip.Type == "DigitalBack" {
			backCount++
		}
	}
	if scannerCount > 1 {
		log.Printf("INFO: invalid setup for workstation %d; more than one scanner", tgtWS.ID)
		c.String(http.StatusBadRequest, "a workstation can only have one scanner")
		return
	}
	if scannerCount == 1 {
		if len(req.Setup) > 1 {
			log.Printf("INFO: invalid setup for workstation %d; workstation can only have a camera assembly or a scanner, not both", tgtWS.ID)
			c.String(http.StatusBadRequest, "A workstation can only have a camera assembly or a scanner, not both")
			return
		}
	} else {
		if len(req.Setup) < 3 {
			log.Printf("INFO: invalid setup for workstation %d; setup is incomplete", tgtWS.ID)
			c.String(http.StatusBadRequest, "Incomplete camera assembly.")
			return
		}
		if bodyCount != 1 || backCount != 1 {
			log.Printf("INFO: invalid setup for workstation %d; camera must have 1 back and 1 body", tgtWS.ID)
			c.String(http.StatusBadRequest, "Camera assembly must have one back and one body")
			return
		}
		if lensCount > 2 {
			log.Printf("INFO: invalid setup for workstation %d; more than 2 lenses", tgtWS.ID)
			c.String(http.StatusBadRequest, "Camera assembly can have a maximum of two lenses")
			return
		}
	}
	log.Printf("INFO: workstation %d setup is valid; save changes", tgtWS.ID)
	err = svc.DB.Exec("delete from workstation_equipment where workstation_id=?", tgtWS.ID).Error
	if err != nil {
		log.Printf("ERROR: unable to clear workstation %d setup: %s", tgtWS.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	addSQL := "insert into workstation_equipment (workstation_id,equipment_id) values "
	addSQL += strings.Join(addValues, ",")
	err = svc.DB.Exec(addSQL).Error
	if err != nil {
		log.Printf("ERROR: unable to add workstation %d setup: %s", tgtWS.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: reload workstation %d and all equipment", tgtWS.ID)
	var resp struct {
		Workstation workstation `json:"workstation"`
		Equipment   []equipment `json:"equipment"`
	}

	err = svc.DB.Preload("Equipment", "status != ?", 2).Find(&resp.Workstation, tgtWS.ID).Error
	if err != nil {
		log.Printf("ERROR: unable to get workstations: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = svc.DB.Where("status!=?", 2).Order("type asc, name asc").Find(&resp.Equipment).Error
	if err != nil {
		log.Printf("ERROR: unable to get equipment: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
}
