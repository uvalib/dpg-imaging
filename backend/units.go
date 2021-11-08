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
	Type       string `json:"type"`
}

type intendedUse struct {
	ID                    uint   `json:"id"`
	Description           string `json:"description"`
	DeliverableFormat     string `json:"deliverableFormat"`
	DeliverableResolution string `json:"deliverableResolution"`
}

type customer struct {
	ID        uint   `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type order struct {
	ID         uint     `json:"id"`
	CustomerID uint     `json:"customerID"`
	Customer   customer `gorm:"foreignKey:CustomerID" json:"customer"`
}

type unit struct {
	ID                  uint        `json:"id"`
	OrderID             uint        `json:"orderID"`
	Order               order       `gorm:"foreignKey:OrderID" json:"order"`
	MetadataID          uint        `json:"-"`
	Metadata            metadata    `gorm:"foreignKey:MetadataID" json:"metadata"`
	IntendedUseID       uint        `json:"-"`
	IntendedUse         intendedUse `gorm:"foreignKey:IntendedUseID" json:"intendedUse"`
	SpecialInstructions string      `json:"specialInstructions,omitempty"`
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
	resp := svc.DB.Preload(clause.Associations).Joins("inner join units on units.metadata_id = metadata.id").Where("units.id=?", uid).First(&md)
	if resp.Error != nil {
		return nil, fmt.Errorf("unable to get metadata for unit %s: %s", uid, resp.Error.Error())
	}

	return &md, nil
}
