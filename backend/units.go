package main

import (
	"fmt"
	"log"

	"gorm.io/gorm/clause"
)

type ocrHint struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OCRCandidate bool   `json:"ocrCandidate"`
}

type ocrLanguageHint struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type metadata struct {
	ID              uint    `json:"id"`
	PID             string  `gorm:"column:pid" json:"pid"`
	CallNumber      string  `json:"callNumber,omitempty"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	OCRHintID       uint    `json:"-"`
	OCRHint         ocrHint `gorm:"foreignKey:OCRHintID" json:"ocrHint"`
	OCRLanguageHint string  `json:"ocrLanguageHint"`
}

type intendedUse struct {
	ID                    uint   `json:"id"`
	Description           string `json:"description"`
	DeliverableFormat     string `json:"deliverableFormat"`
	DeliverableResolution string `json:"deliverableResolution"`
}

type agency struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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
	OCRMasterFiles      bool        `json:"ocrMasterFiles"`
	UnitStatus          string      `json:"status"`
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
