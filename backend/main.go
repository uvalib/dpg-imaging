package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// Version of the service
const Version = "4.0.3"

func main() {
	// Load cfg
	log.Printf("===> DPG Imaging Service is starting up <===")
	cfg := getConfiguration()
	svc := initializeService(Version, cfg)

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()

	// Set routes and start server
	router.Use(cors.Default())
	router.GET("/config", svc.getConfig)
	router.GET("/version", svc.getVersion)
	router.GET("/healthcheck", svc.healthCheck)
	router.GET("/authenticate", svc.authenticate)
	api := router.Group("/api")
	{
		api.GET("/components/:id", svc.authMiddleware, svc.getComponent)

		api.GET("/projects", svc.authMiddleware, svc.getProjects)
		api.GET("/projects/:id", svc.authMiddleware, svc.getProject)
		api.PUT("/projects/:id", svc.authMiddleware, svc.updateProject)
		api.POST("/projects/:id/assign/:uid", svc.authMiddleware, svc.assignProject)
		api.POST("/projects/:id/equipment", svc.authMiddleware, svc.setProjectEquipment)
		api.POST("/projects/:id/note", svc.authMiddleware, svc.addNote)
		api.POST("/projects/:id/start", svc.authMiddleware, svc.startProjectStep)
		api.POST("/projects/:id/finish", svc.authMiddleware, svc.finishProjectStep)
		api.POST("/projects/:id/reject", svc.authMiddleware, svc.rejectProjectStep)

		api.GET("/units/:uid/masterfiles", svc.authMiddleware, svc.getUnitMasterFiles)
		api.POST("/units/:uid/update", svc.authMiddleware, svc.updateMetadata)
		api.POST("/units/:uid/rename", svc.authMiddleware, svc.renameFiles)
		api.DELETE("/units/:uid/:file", svc.authMiddleware, svc.deleteFile)
		api.POST("/units/:uid/:file/rotate", svc.authMiddleware, svc.rotateFile)
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
