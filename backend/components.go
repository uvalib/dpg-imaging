package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (svc *serviceContext) getComponent(c *gin.Context) {
	cid := c.Param("id")
	log.Printf("INFO: lookup component %s", cid)
	cBytes, reqErr := svc.getRequest(fmt.Sprintf("%s/components/%s", svc.TrackSys.API, cid))
	if reqErr != nil {
		c.String(reqErr.StatusCode, reqErr.Message)
	}
	var cmp any
	if err := json.Unmarshal(cBytes, &cmp); err != nil {
		log.Printf("ERROR: unable to parse component response: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, cmp)
}
