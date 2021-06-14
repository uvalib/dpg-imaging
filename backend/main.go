package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// Version of the service
const Version = "1.0.0"

func main() {
	// Load cfg
	log.Printf("===> DPG Imaging Service is staring up <===")
	cfg := getConfiguration()
	svc := initializeService(Version, cfg)

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()

	// Set routes and start server
	router.Use(cors.Default())
	router.GET("/version", svc.getVersion)
	router.GET("/healthcheck", svc.healthCheck)
	api := router.Group("/api")
	{
		api.GET("/units", svc.getUnits)
		api.GET("/units/:uid", svc.getUnitDetails)
		api.POST("/units/:uid/update", svc.updateMetadata)
		api.POST("/units/:uid/rename", svc.renameFiles)
		api.DELETE("/units/:uid/:file", svc.deleteFile)
	}

	// Note: in dev mode, this is never actually used. The front end is served
	// by yarn and it proxies all requests to the API to the routes above
	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	// add a catchall route that renders the index page.
	// based on no-history config setup info here:
	//    https://router.vuejs.org/guide/essentials/history-mode.html#example-server-configurations
	router.NoRoute(func(c *gin.Context) {
		c.File("./public/index.html")
	})

	portStr := fmt.Sprintf(":%d", cfg.port)
	log.Printf("INFO: start DPG Imaging Service on port %s with CORS support enabled", portStr)
	log.Fatal(router.Run(portStr))
}
