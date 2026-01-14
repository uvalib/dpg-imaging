package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

func (svc *serviceContext) extAuthMiddleware(c *gin.Context) {
	log.Printf("Authorize external access to %s", c.Request.URL)
	tokenStr, err := getBearerToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Printf("Authentication failed: [%s]", err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if tokenStr == "undefined" {
		log.Printf("Authentication failed; bearer token is undefined")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("Validating JWT auth token with tracksys key...")
	jwtClaims := jwtClaims{}
	_, jwtErr := jwt.ParseWithClaims(tokenStr, &jwtClaims, func(token *jwt.Token) (any, error) {
		return []byte(svc.TrackSys.JWTKey), nil
	})
	if jwtErr != nil {
		log.Printf("Authentication failed; token validation failed: %+v", jwtErr)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("got valid bearer token for external auth: [%s] for %v", tokenStr, jwtClaims)
	c.Next()
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
	uid := c.Param("uid")
	type lookupResp struct {
		Exists    bool `json:"exists"`
		ProjectID uint `json:"projectID"`
	}
	var proj project
	if err := svc.DB.Where("unit_id=?", uid).First(&proj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) == false {
			log.Printf("ERROR: lookup project for unit %s failed: %s", uid, err.Error())
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			log.Printf("INFO: no project exists for %s", uid)
			c.JSON(http.StatusOK, lookupResp{})
		}
		return
	}

	c.JSON(http.StatusOK, lookupResp{Exists: true, ProjectID: proj.ID})
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
		CurrentStepID: firstStep.ID,
		AddedAt:       &now,
		CategoryID:    uint(req.CategoryID),
		ItemCondition: uint(req.Condition),
		ConditionNote: req.Notes,
	}
	if req.ContainerTypeID != 0 {
		cID := uint(req.ContainerTypeID)
		newProj.ContainerTypeID = cID
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
	c.String(http.StatusNotImplemented, "NO")
}
