package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type metadata struct {
	ID         uint   `json:"id"`
	PID        string `gorm:"column:pid" json:"pid"`
	CallNumber string `json:"callNumber,omitempty"`
	Title      string `json:"title"`
}

type intendedUse struct {
	ID                    uint   `json:"id"`
	Description           string `json:"description"`
	DeliverableFormat     string `json:"deliverableFormat"`
	DeliverableResolution string `json:"deliverableResolution"`
}

type unit struct {
	ID            uint        `json:"id"`
	OrderID       uint        `json:"orderID"`
	MetadataID    uint        `json:"-"`
	Metadata      metadata    `gorm:"foreignKey:MetadataID" json:"metadata"`
	IntendedUseID uint        `json:"-"`
	IntendedUse   intendedUse `gorm:"foreignKey:IntendedUseID" json:"intendedUse"`
}

func (svc *serviceContext) getQAUnits(c *gin.Context) {
	log.Printf("INFO: get available units from %s", svc.ImagesDir)
	files, err := ioutil.ReadDir(svc.ImagesDir)
	if err != nil {
		log.Printf("ERROR: unable to list contents of images directory: %s", err.Error())
		c.String(http.StatusInternalServerError, "unable to find units")
		return
	}

	// get only directories that match naming requirements for a unit; 9 digits.
	unitRegex := regexp.MustCompile(`^\d{9}$`)
	out := make([]string, 0)
	for _, f := range files {
		fName := f.Name()
		if unitRegex.Match([]byte(fName)) {
			out = append(out, fName)
		}
	}
	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) getUnitMetadata(uid string) (*metadata, error) {
	log.Printf("INFO: get metadata for unit %s", uid)
	var md metadata
	resp := svc.GDB.Preload(clause.Associations).Joins("inner join units on units.metadata_id = metadata.id").Where("units.id=?", uid).First(&md)
	if resp.Error != nil {
		return nil, fmt.Errorf("unable to get metadata for unit %s: %s", uid, resp.Error.Error())
	}

	return &md, nil
}
