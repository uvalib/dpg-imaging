package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext contains common data used by all handlers
type serviceContext struct {
	Version     string
	ServiceURL  string
	ImagesDir   string
	ScanDir     string
	IIIFURL     string
	TrackSysURL string
	HTTPClient  *http.Client
	DB          *gorm.DB
	DevAuthUser string
	JWTKey      string
}

// InitializeService sets up the service context for all API handlers
func initializeService(version string, cfg *configData) *serviceContext {
	ctx := serviceContext{Version: version,
		ImagesDir:   cfg.imagesDir,
		IIIFURL:     cfg.iiifURL,
		ScanDir:     cfg.scanDir,
		JWTKey:      cfg.jwtKey,
		ServiceURL:  cfg.serviceURL,
		TrackSysURL: cfg.tracksysURL,
		DevAuthUser: cfg.devAuthUser}

	log.Printf("INFO: connecting to DB...")
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true",
		cfg.db.User, cfg.db.Pass, cfg.db.Host, cfg.db.Name)
	gdb, err := gorm.Open(mysql.Open(connectStr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	ctx.DB = gdb
	log.Printf("INFO: DB Connection established")

	log.Printf("INFO: create tmp directory for working files...")
	tmpDir := path.Join(ctx.ImagesDir, "tmp")
	_, existErr := os.Stat(tmpDir)
	if existErr != nil {
		err := os.Mkdir(tmpDir, 0777)
		if err != nil {
			log.Fatal(fmt.Sprintf("unable to make tmp dir %s: %s", tmpDir, err.Error()))
		}
		log.Printf("INFO: tmp directory created")
	} else {
		log.Printf("INFO: tmp directory already exists")
	}

	log.Printf("INFO: create HTTP client...")
	defaultTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 600 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	ctx.HTTPClient = &http.Client{
		Transport: defaultTransport,
		Timeout:   5 * time.Second,
	}
	log.Printf("INFO: HTTP Client created")
	return &ctx
}

func (svc *serviceContext) healthCheck(c *gin.Context) {
	type hcResp struct {
		Healthy bool   `json:"healthy"`
		Message string `json:"message,omitempty"`
	}
	hcMap := make(map[string]hcResp)
	hcMap["circulation"] = hcResp{Healthy: true}

	c.JSON(http.StatusOK, hcMap)
}

func (svc *serviceContext) getVersion(c *gin.Context) {
	build := "unknown"

	// cos our CWD is the bin directory
	files, _ := filepath.Glob("../buildtag.*")
	if len(files) == 1 {
		build = strings.Replace(files[0], "../buildtag.", "", 1)
	}

	vMap := make(map[string]string)
	vMap["version"] = Version
	vMap["build"] = build
	c.JSON(http.StatusOK, vMap)
}

func (svc *serviceContext) getConfig(c *gin.Context) {
	log.Printf("INFO: get service configuration")
	type cfgData struct {
		TrackSysURL      string            `json:"tracksysURL"`
		QAImageDir       string            `json:"qaImageDir"`
		ScanDir          string            `json:"scanDir"`
		Staff            []staffMember     `json:"staff"`
		Workstations     []workstation     `json:"workstations"`
		Workflows        []workflow        `json:"workflows"`
		Categories       []category        `json:"categories"`
		OCRHints         []ocrHint         `json:"ocrHints"`
		OCRLanguageHints []ocrLanguageHint `json:"ocrLanguageHints"`
	}
	resp := cfgData{TrackSysURL: svc.TrackSysURL,
		QAImageDir: svc.ImagesDir,
		ScanDir:    svc.ScanDir,
	}

	log.Printf("UNFO: load staff members")
	dbResp := svc.DB.Where("role<=? and is_active=?", 2, 1).Order("last_name asc").Find(&resp.Staff)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to get staff members: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load categories")
	dbResp = svc.DB.Order("name asc").Find(&resp.Categories)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load categories: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load workstations")
	dbResp = svc.DB.Order("name asc").Find(&resp.Workstations)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load workstations: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load workflows")
	dbResp = svc.DB.Order("name asc").Where("active=1").Find(&resp.Workflows)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load workflows: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load ocr hints")
	dbResp = svc.DB.Order("name asc").Find(&resp.OCRHints)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load ocr hints: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load ocr language hints")
	resp.OCRLanguageHints = make([]ocrLanguageHint, 0)
	f, err := os.Open("./data/languages.csv")
	if err != nil {
		log.Printf("ERROR: unable to load ocr language hints: %s", dbResp.Error.Error())
	} else {
		defer f.Close()
		csvReader := csv.NewReader(f)
		langRecs, err := csvReader.ReadAll()
		if err != nil {
			log.Printf("ERROR: unable to parse languages file: %s", err.Error())
		} else {
			for _, rec := range langRecs {
				resp.OCRLanguageHints = append(resp.OCRLanguageHints, ocrLanguageHint{Code: rec[0], Language: rec[1]})
			}
		}
	}

	c.JSON(http.StatusOK, resp)
}
