package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (svc *serviceContext) adminMiddleware(c *gin.Context) {
	log.Printf("Authorize admin access to %s", c.Request.URL)
	tokenStr, err := getBearerToken(c.Request.Header.Get("Authorization"))
	if err != nil {
		log.Printf("Authentication failed: [%s]", err.Error())
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if tokenStr == "undefined" {
		log.Printf("Authentication failed; bearer token is undefined")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("Validating JWT auth token...")
	jwtClaims := jwtClaims{}
	_, jwtErr := jwt.ParseWithClaims(tokenStr, &jwtClaims, func(token *jwt.Token) (any, error) {
		return []byte(svc.JWTKey), nil
	})
	if jwtErr != nil {
		log.Printf("Authentication failed; token validation failed: %+v", jwtErr)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if jwtClaims.Role != "admin" {
		log.Printf("Authentication failed; admin required")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	log.Printf("got valid bearer token: [%s] for %s", tokenStr, jwtClaims.ComputeID)
	c.Set("jwt", tokenStr)
	c.Set("claims", jwtClaims)
	c.Next()
}

func (svc *serviceContext) adminCleanupOldProjects(c *gin.Context) {
	claims := getJWTClaims(c)
	limit, _ := strconv.ParseUint(c.Query("limit"), 10, 64)
	log.Printf("INFO: admin %s requests cleanup of projects older than 3 years", claims.ComputeID)
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
				log.Printf("INFO: max deletions of 500 projects reached")
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
