package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type problem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
}

type note struct {
	ID            uint        `json:"id"`
	ProjectID     uint        `json:"-"`
	StepID        uint        `json:"-"`
	Step          step        `gorm:"foreignKey:StepID" json:"step"`
	NoteType      uint        `json:"type"`
	Note          string      `json:"text"`
	CreatedAt     *time.Time  `json:"createdAt,omitempty"`
	UpdatedAt     *time.Time  `json:"-"`
	Problems      []problem   `gorm:"many2many:notes_problems"  json:"problems"`
	StaffMemberID uint        `json:"-"`
	StaffMember   staffMember `gorm:"foreignKey:StaffMemberID" json:"staffMember"`
}

func (svc *serviceContext) addNoteRequest(c *gin.Context) {
	projID := c.Param("id")
	claims := getJWTClaims(c)
	var noteReq struct {
		StepID     uint   `json:"stepID"`
		TypeID     uint   `json:"noteTypeID"`
		Note       string `json:"note"`
		ProblemIDs []uint `json:"problemIDs"`
	}

	qpErr := c.ShouldBindJSON(&noteReq)
	if qpErr != nil {
		log.Printf("ERROR: invalid note payload for project %s: %s", qpErr.Error(), projID)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}
	log.Printf("INFO: user %s is adding a note to project %s: %+v", claims.ComputeID, projID, noteReq)

	var proj project
	err := svc.DB.Find(&proj, projID).Error
	if err != nil {
		log.Printf("ERROR: unable to get project %s: %s", projID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	newNote := note{ProjectID: proj.ID, StepID: noteReq.StepID, StaffMemberID: claims.UserID, NoteType: noteReq.TypeID, Note: noteReq.Note}
	notes, err := svc.addNote(proj, newNote, noteReq.ProblemIDs)
	if err != nil {
		log.Printf("ERROR: add note to project %d failed: %s", proj.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, notes)
}

func (svc *serviceContext) addNote(proj project, newNote note, problemIDs []uint) ([]note, error) {
	log.Printf("INFO: add note to project %d", proj.ID)
	var notes []note
	err := svc.DB.Model(&proj).Association("Notes").Append(&newNote)
	if err != nil {
		return nil, err
	}

	if len(problemIDs) > 0 {
		pq := "insert into notes_problems (note_id, problem_id) values "
		var vals []string
		for _, pid := range problemIDs {
			vals = append(vals, fmt.Sprintf("(%d,%d)", newNote.ID, pid))
		}
		pq += strings.Join(vals, ",")
		resp := svc.DB.Exec(pq)
		if resp.Error != nil {
			log.Printf("ERROR: unable to add problems to note: %s", resp.Error.Error())
		}
	}

	err = svc.DB.Where("project_id=?", proj.ID).
		Joins("Step").Joins("StaffMember").Preload("Problems").
		Order("notes.created_at DESC").Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}
