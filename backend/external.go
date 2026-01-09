package main

import (
	"errors"
	"log"
	"net/http"

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
	c.String(http.StatusNotImplemented, "NO")
}

func (svc *serviceContext) cancelProject(c *gin.Context) {
	c.String(http.StatusNotImplemented, "NO")
}
