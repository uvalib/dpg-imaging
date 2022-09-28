package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type message struct {
	ID      int64       `json:"id"`
	Subject string      `json:"subject"`
	Message string      `json:"message"`
	FromID  int64       `gorm:"column:from_id" json:"-"`
	From    staffMember `gorm:"foreignKey:FromID" json:"from"`
	ToID    int64       `gorm:"column:to_id" json:"-"`
	To      staffMember `gorm:"foreignKey:ToID" json:"to"`
	Read    bool        `json:"read"`
	Deleted bool        `json:"deleted"`
	SentAt  time.Time   `json:"sentAt"`
}

func (svc *serviceContext) getMessages(c *gin.Context) {
	userID := c.Param("id")
	log.Printf("INFO: get messages for user %s", userID)
	var out []message
	err := svc.DB.Preload("From").Where("deleted=?", 0).Where("to_id=?", userID).Find(&out).Error
	if err != nil {
		log.Printf("ERROR: unable to get messages for user %s: %s", userID, err.Error())
	}
	c.JSON(http.StatusOK, out)
}
