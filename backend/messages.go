package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type message struct {
	ID         int64              `json:"id"`
	Subject    string             `json:"subject"`
	Message    string             `json:"message"`
	FromID     int64              `gorm:"column:from_id" json:"fromID"`
	Recipients []messageRecipient `gorm:"foreignKey:MessageID" json:"recipients"`
	SentAt     time.Time          `json:"sentAt"`
}

type messageRecipient struct {
	ID        int64 `json:"-"`
	MessageID int64 `json:"-"`
	StaffID   int64 `gorm:"column:staff_id" json:"staffID"`
	Read      bool  `json:"read"`
	Deleted   bool  `json:"deleted"`
}

func (svc *serviceContext) getMessages(c *gin.Context) {
	userID := c.Param("id")
	log.Printf("INFO: get messages for user %s", userID)
	var inbox []message
	err := svc.DB.Preload("Recipients").Joins("inner join message_recipients as r on message_id=messages.id").Where("r.deleted=? and r.staff_id=?", 0, userID).Find(&inbox).Error
	if err != nil {
		log.Printf("ERROR: unable to get messages for user %s: %s", userID, err.Error())
	}
	var sent []message
	err = svc.DB.Preload("Recipients").Joins("inner join message_recipients as r on message_id=messages.id").
		Group("messages.id").Where("r.deleted=? and from_id=?", 0, userID).Find(&sent).Error
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
	delQ := "update message_recipients set deleted=?, deleted_at=? where message_id=? and staff_id=?"
	if err := svc.DB.Exec(delQ, 1, time.Now(), msgID, userID).Error; err != nil {
		log.Printf("ERROR: unable to delete message %s: %s", msgID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "deleted")
}

func (svc *serviceContext) markMessageRead(c *gin.Context) {
	userID := c.Param("id")
	msgID := c.Param("msgid")
	log.Printf("INFO: mark user %s message %s as read", userID, msgID)
	readQ := "update message_recipients set `read`=? where message_id=? and staff_id=?"
	if err := svc.DB.Exec(readQ, 1, msgID, userID).Error; err != nil {
		log.Printf("ERROR: unable to mark message %s read: %s", msgID, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "read")
}

func (svc *serviceContext) sendMessage(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if userID == 0 {
		log.Printf("INFO: invalid user id %s in send message request", c.Param("id"))
		return
	}

	var msgRequest struct {
		To      []int64 `json:"to"`
		Subject string  `json:"subject"`
		Message string  `json:"message"`
	}

	qpErr := c.ShouldBindJSON(&msgRequest)
	if qpErr != nil {
		log.Printf("ERROR: invalid message payload from user %d: %s", userID, qpErr)
		c.String(http.StatusBadRequest, "Invalid request")
		return
	}

	log.Printf("INFO: user %d is sending a message with subject '%s' to recipients %v", userID, msgRequest.Subject, msgRequest.To)

	newMsg := message{Subject: msgRequest.Subject, Message: msgRequest.Message, SentAt: time.Now(), FromID: userID}
	if err := svc.DB.Create(&newMsg).Error; err != nil {
		log.Printf("ERROR: uable to create message %+v: %s", newMsg, err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	for _, toID := range msgRequest.To {
		recip := messageRecipient{
			MessageID: newMsg.ID,
			StaffID:   toID,
		}
		if err := svc.DB.Create(&recip).Error; err != nil {
			log.Printf("ERROR: unable to add recipient %d to message %d", toID, newMsg.ID)
		}
		newMsg.Recipients = append(newMsg.Recipients, recip)
	}

	c.JSON(http.StatusOK, newMsg)
}
