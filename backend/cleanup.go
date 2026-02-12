package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (svc *serviceContext) cleanupOldProjects(c *gin.Context) {
	limit, _ := strconv.ParseUint(c.Query("limit"), 10, 64)
	log.Printf("INFO: cleanup of projects older than 3 years requested")
	if limit > 0 {
		log.Printf("INFO: limit to %d deletions", limit)
	}

	threeYearsAgo := time.Now().AddDate(-3, 0, 0)
	dateStr := threeYearsAgo.Format("2006-01-02")
	log.Printf("INFO: cutoff date is %s", dateStr)
	sql := "select id,finished_at from projects where finished_at is not null and finished_at < ? order by finished_at asc"
	var projs []struct {
		ID         int64
		FinishedAt time.Time
	}
	if err := svc.DB.Raw(sql, dateStr).Scan(&projs).Error; err != nil {
		log.Printf("ERROR: unable to get old projects: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: %d projects older than %s found", len(projs), dateStr)
	go func() {
		delCnt := 0
		failCnt := 0
		for _, proj := range projs {
			log.Printf("INFO: remove project %d finished at %s", proj.ID, proj.FinishedAt.Format("2006-01-02"))
			if err := svc.doProjectDelete(proj.ID); err != nil {
				log.Printf("ERROR: unable to delete project %d: %s", proj.ID, err.Error())
				failCnt++
			} else {
				delCnt++
			}

			time.Sleep(100 * time.Millisecond)
			if limit > 0 && delCnt >= int(limit) {
				log.Printf("INFO: deletion limit of %d projects reached", limit)
				break

			} else if delCnt >= 500 {
				log.Printf("INFO: max deletions of 500 projects reached")
				break
			}
		}
		log.Printf("INFO: finished processing %d projects older than %s. %d deleted, %d failed", len(projs), dateStr, delCnt, failCnt)
	}()

	c.String(http.StatusOK, "process started to retire %d completed projects older than %s", len(projs), dateStr)
}

func (svc *serviceContext) cleanupDeletedMessages(c *gin.Context) {
	log.Printf("INFO: cleanup deleted messages older than 2 months")
	deleteThreshold := time.Now().AddDate(0, -2, 0)
	dateStr := deleteThreshold.Format("2006-01-02")

	// NOTES: a message can have multiple recipients. Each recipient is in the message_recipients
	// table and each can independently mark a message as deleted. When this happens, only their reference
	// to the message is removed. If a message has no recipients, the it can be deleted.
	log.Printf("INFO: scan for message recipients to delete")
	var delMessageIDs []int64
	delQ := "select message_id from message_recipients where deleted=? and deleted_at < ?"
	if err := svc.DB.Raw(delQ, 1, dateStr).Scan(&delMessageIDs).Error; err != nil {
		log.Printf("ERROR: unable to get count of old messages: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if len(delMessageIDs) == 0 {
		log.Printf("INFO: there are no messages to delete")
		c.String(http.StatusOK, "no messages to delete")
		return
	}

	log.Printf("INFO: delete %d records for recipients that marked a message as deleted", len(delMessageIDs))
	if err := svc.DB.Exec("DELETE from message_recipients where message_id in ?", delMessageIDs).Error; err != nil {
		log.Printf("ERROR: unable to delete message recipients: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	cntQ := "select m.id, count(r.id) as cnt from messages m left join message_recipients r on r.message_id=m.id where m.id in ? group by(m.id)"
	var msgRespCnts []struct {
		ID  int64
		Cnt int64
	}
	if err := svc.DB.Raw(cntQ, delMessageIDs).Scan(&msgRespCnts).Error; err != nil {
		log.Printf("ERROR: unable to get message recipient counts: %s", err.Error())
		c.JSON(http.StatusOK, fmt.Sprintf("%d messages deleted", len(delMessageIDs)))
		return
	}

	var noRecipIDs []int64
	for _, msgCnt := range msgRespCnts {
		if msgCnt.Cnt == 0 {
			noRecipIDs = append(noRecipIDs, msgCnt.ID)
		}
	}

	if len(noRecipIDs) > 0 {
		log.Printf("INFO: delete %d messages with no recipients", len(delMessageIDs))
		if err := svc.DB.Exec("DELETE from messages where id in ?", noRecipIDs).Error; err != nil {
			log.Printf("ERROR: unable to delete messages: %s", err.Error())
		}
	}

	c.JSON(http.StatusOK, fmt.Sprintf("%d messages deleted", len(delMessageIDs)))
}
