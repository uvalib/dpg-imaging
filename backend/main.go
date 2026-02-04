package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

// Version of the service
const Version = "5.10.0"

func main() {
	// Load cfg
	log.Printf("===> DPG Imaging Service is starting up <===")
	cfg := getConfiguration()
	svc := initializeService(Version, cfg)

	// log.Printf("HACK THE PROJECT UPDATES =============================")
	// var projs []project
	// cnt := 0
	// var errors []uint
	// var missing []uint
	// if err := svc.DB.Where("title is null").Find(&projs).Error; err != nil {
	// 	log.Fatal(err.Error())
	// }

	// qStr := "select order_id, customer_id, agency_id, title, call_number from units u inner join orders o on o.id = order_id inner join metadata m on m.id = metadata_id where u.id=?"
	// for _, tgtProj := range projs {
	// 	log.Printf("ADD METADATA TO PROJECT %d", tgtProj.ID)
	// 	var info struct {
	// 		OrderID    uint
	// 		CustomerID uint
	// 		AgencyID   *uint
	// 		CallNumber string
	// 		Title      string
	// 	}
	// 	if err := svc.DB.Raw(qStr, tgtProj.ID).Scan(&info).Error; err != nil {
	// 		log.Fatal(err.Error())
	// 	}
	// 	if info.OrderID == 0 {
	// 		log.Printf("NO UNIT/ORDER FOR PROJECT %d", tgtProj.ID)
	// 		missing = append(missing, tgtProj.ID)
	// 		continue
	// 	}

	// 	fields := make([]string, 0)
	// 	if info.Title != tgtProj.Title {
	// 		tgtProj.Title = info.Title
	// 		fields = append(fields, "Title")
	// 	}
	// 	if info.CallNumber != tgtProj.CallNumber {
	// 		tgtProj.CallNumber = info.CallNumber
	// 		fields = append(fields, "CallNumber")
	// 	}
	// 	if info.OrderID != tgtProj.OrderID {
	// 		tgtProj.OrderID = info.OrderID
	// 		fields = append(fields, "OrderID")
	// 	}
	// 	if info.CustomerID != tgtProj.CustomerID {
	// 		tgtProj.CustomerID = info.CustomerID
	// 		fields = append(fields, "CustomerID")
	// 	}
	// 	if info.AgencyID != nil {
	// 		if *info.AgencyID != tgtProj.AgencyID {
	// 			tgtProj.AgencyID = *info.AgencyID
	// 			fields = append(fields, "AgencyID")
	// 		}
	// 	}

	// 	if len(fields) > 0 {
	// 		if err := svc.DB.Model(&tgtProj).Select(fields).Updates(tgtProj).Error; err != nil {
	// 			log.Fatalf("unable to update project %d metadata: %s", tgtProj.ID, err.Error())
	// 		}
	// 	}

	// 	cnt++
	// }
	// log.Printf("UPDATE %d projects", cnt)
	// if len(errors) > 0 {
	// 	log.Printf("FAILED UPDATES: %v", errors)
	// }
	// if len(missing) > 0 {
	// 	log.Printf("MNISSING UNITS %v", missing)
	// }
	// log.Fatal("DONE")

	// Set routes and start server
	gin.SetMode(gin.ReleaseMode)
	gin.DisableConsoleColor()
	router := gin.Default()
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowCredentials = true
	corsCfg.AddAllowHeaders("Authorization")
	router.Use(cors.New(corsCfg))

	// Set routes and start server
	router.GET("/config", svc.getConfig)
	router.GET("/version", svc.getVersion)
	router.GET("/healthcheck", svc.healthCheck)
	router.GET("/authenticate", svc.authenticate)

	// external API used by TrackSys / dpg-jobs
	router.GET("/constants", svc.getConstants)
	router.GET("/projects/lookup", svc.lookupProjectForUnit)

	api := router.Group("/api", svc.authMiddleware)
	{
		// external calls used by TS/jobs
		api.POST("/projects/create", svc.createProject)
		api.POST("/projects/:id/cancel", svc.cancelProject)
		api.POST("/projects/:id/done", svc.finishProject)
		api.POST("/projects/:id/fail", svc.failProject)
		api.POST("/projects/:id/update", svc.updateProjectMetadata)

		api.GET("/components/:id", svc.getComponent)

		// equipment management
		api.GET("/equipment", svc.getEquipment)
		api.POST("/equipment", svc.createEquipment)
		api.POST("/equipment/:id/update", svc.updateEquipment)
		api.POST("/workstation", svc.createWorkstation)
		api.POST("/workstation/:id/update", svc.updateWorkstation)
		api.POST("/workstation/:id/setup", svc.updateWorkstationSetup)

		api.GET("/projects", svc.getProjects)
		api.GET("/projects/:id", svc.getProject)
		api.PUT("/projects/:id", svc.updateProject)
		api.DELETE("/projects/:id", svc.deleteProject)
		api.PUT("/projects/:id/images/count", svc.updateProjecImageCount)
		api.GET("/projects/:id/status", svc.getProjectStatus)
		api.POST("/projects/:id/assign/:uid", svc.assignProject)
		api.POST("/projects/:id/equipment", svc.setProjectEquipment)
		api.POST("/projects/:id/note", svc.addNoteRequest)
		api.POST("/projects/:id/start", svc.startProjectStep)
		api.POST("/projects/:id/finish", svc.finishProjectStep)
		api.POST("/projects/:id/reject", svc.rejectProjectStep)

		api.GET("/units/:uid/validate/components", svc.validateComponentSettings)
		api.GET("/units/:uid/masterfiles", svc.getUnitMasterFiles)
		api.GET("/units/:uid/masterfiles/metadata", svc.getMasterFilesMetadata)
		api.POST("/units/:uid/update", svc.updateMetadataBatch)       // this is protected by BatchUnitsInProgress
		api.POST("/units/:uid/rename", svc.renameFiles)               // this is protected by BatchUnitsInProgress
		api.POST("/units/:uid/delete", svc.deleteFiles)               // this is protected by BatchUnitsInProgress
		api.POST("/units/:uid/:file/rotate", svc.rotateFile)          // this is protected by BatchUnitsInProgress
		api.POST("/units/:uid/:file/update", svc.updateImageMetadata) // this is protected by BatchUnitsInProgress

		api.GET("/user/:id/messages", svc.getMessages)
		api.POST("/user/:id/messages/:msgid/delete", svc.deleteMessage)
		api.POST("/user/:id/messages/:msgid/read", svc.markMessageRead)
		api.POST("/user/:id/messages/send", svc.sendMessage)
	}

	cleanup := router.Group("/cleanup")
	{
		cleanup.POST("/finished-projects", svc.cleanupOldProjects)
		cleanup.POST("/deleted-messages", svc.cleanupDeletedMessages)
	}

	// Note: in dev mode, this is never actually used. The front end is served
	// by NPM and it proxies all requests to the API to the routes above
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
