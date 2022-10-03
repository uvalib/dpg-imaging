package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type message struct {
	ID        int64       `json:"id"`
	Subject   string      `json:"subject"`
	Message   string      `json:"message"`
	FromID    int64       `gorm:"column:from_id" json:"-"`
	From      staffMember `gorm:"foreignKey:FromID" json:"from"`
	ToID      int64       `gorm:"column:to_id" json:"-"`
	To        staffMember `gorm:"foreignKey:ToID" json:"to"`
	Read      bool        `json:"read"`
	Deleted   bool        `json:"deleted"`
	SentAt    time.Time   `json:"sentAt"`
	DeletedAt *time.Time  `json:"-"`
}

func (svc *serviceContext) getMessages(c *gin.Context) {
	userID := c.Param("id")
	log.Printf("INFO: get messages for user %s", userID)
	var inbox []message
	err := svc.DB.Preload("From").Preload("To").Where("deleted=?", 0).Where("to_id=?", userID).Find(&inbox).Error
	if err != nil {
		log.Printf("ERROR: unable to get messages for user %s: %s", userID, err.Error())
	}
	var sent []message
	err = svc.DB.Preload("From").Preload("To").Where("deleted=?", 0).Where("from_id=?", userID).Find(&sent).Error
	if err != nil {
		log.Printf("ERROR: unable to get sent messages for user %s: %s", userID, err.Error())
	}

	type msgResp struct {
		Inbox []message `json:"inbox"`
		Sent  []message `json:"sent"`
	}
	out := msgResp{Inbox: inbox, Sent: sent}
	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) deleteMessage(c *gin.Context) {
	userID := c.Param("id")
	msgID := c.Param("msgid")
	log.Printf("INFO: delete user %s message %s", userID, msgID)
	var msg message
	err := svc.DB.Find(&msg, msgID).Error
	if err != nil {
		log.Printf("ERROR: unable to get message %s for deletion: %s", msgID, err.Error())
		return
	}

	delTime := time.Now()
	msg.Deleted = true
	msg.DeletedAt = &delTime
	err = svc.DB.Model(&msg).Select("Deleted", "DeletedAt").Updates(msg).Error
	if err != nil {
		log.Printf("ERROR: unable to delete message %d: %s", msg.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) markMessageRead(c *gin.Context) {
	userID := c.Param("id")
	msgID := c.Param("msgid")
	log.Printf("INFO: mark user %s message %s as read", userID, msgID)
	var msg message
	err := svc.DB.Find(&msg, msgID).Error
	if err != nil {
		log.Printf("ERROR: unable to get message %s to mark as read: %s", msgID, err.Error())
		return
	}

	msg.Read = true
	err = svc.DB.Model(&msg).Select("Read").Updates(msg).Error
	if err != nil {
		log.Printf("ERROR: unable to mark message %d as read: %s", msg.ID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusOK, "read")
}

func (svc *serviceContext) sendMessage(c *gin.Context) {
	userID := c.Param("id")
	log.Printf("INFO: user %s is sending a message", userID)

	var msgRequest struct {
		To      uint   `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	qpErr := c.ShouldBindJSON(&msgRequest)
	if qpErr != nil {
		log.Printf("ERROR: invalid message payload from user %s: %s", userID, qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	var fromUser staffMember
	err := svc.DB.Find(&fromUser, userID).Error
	if err != nil {
		log.Printf("ERROR: sender %s not found: %s", userID, err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("%s is not a valid sender id", userID))
		return
	}

	var toUser staffMember
	err = svc.DB.Find(&toUser, msgRequest.To).Error
	if err != nil {
		log.Printf("ERROR: recipient %s not found: %s", userID, err.Error())
		c.String(http.StatusBadRequest, fmt.Sprintf("%s is not a valid recipient", userID))
		return
	}

	newMsg := message{Subject: msgRequest.Subject, Message: msgRequest.Message, SentAt: time.Now(), ToID: int64(toUser.ID), FromID: int64(fromUser.ID)}
	err = svc.DB.Create(&newMsg).Error
	if err != nil {
		log.Printf("ERROR: uable to create message %+v: %s", newMsg, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, newMsg)
}
