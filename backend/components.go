package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type componentType struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type component struct {
	ID              uint          `json:"id"`
	Title           string        `json:"title,omitempty"`
	Label           string        `json:"label,omitempty"`
	Description     string        `json:"description,omitempty"`
	Date            string        `json:"date,omitempty"`
	ComponentTypeID uint          `json:"-"`
	ComponentType   componentType `gorm:"foreignKey:ComponentTypeID" json:"componentType"`
}

func (svc *serviceContext) getComponent(c *gin.Context) {
	cid := c.Param("id")
	log.Printf("INFO: lookup component %s", cid)

	var cmp component
	resp := svc.DB.Preload(clause.Associations).First(&cmp, cid)
	if resp.Error != nil {
		if errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			log.Printf("INFO: component %s not found", cid)
			c.String(http.StatusNotFound, fmt.Sprintf("%s not found", cid))
		} else {
			log.Printf("ERROR: unable to get component %s: %s", cid, resp.Error.Error())
			c.String(http.StatusInternalServerError, resp.Error.Error())
		}
		return
	}
	c.JSON(http.StatusOK, cmp)
}
