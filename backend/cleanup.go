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

	log.Printf("INFO: scan for messages to delete")
	var delCount int64
	if err := svc.DB.Table("messages").Where("deleted=? and deleted_at < ?", 1, dateStr).Count(&delCount).Error; err != nil {
		log.Printf("ERROR: unable to get count of old messages: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("INFO: delete %d deleted messages", delCount)
	if err := svc.DB.Exec("DELETE from messages where deleted=? and deleted_at < ?", 1, dateStr).Error; err != nil {
		log.Printf("ERROR: unable to delete messages: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, fmt.Sprintf("%d messages deleted", delCount))
}
