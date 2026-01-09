package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sys/unix"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ServiceContext contains common data used by all handlers
type serviceContext struct {
	Version              string
	ServiceURL           string
	ImagesDir            string
	ScanDir              string
	FinalizeDir          string
	IIIFURL              string
	TrackSys             tracksysCfg
	HTTPClient           *http.Client
	DB                   *gorm.DB
	DevAuthUser          string
	JWTKey               string
	BatchSize            int
	batchMutex           sync.Mutex
	BatchUnitsInProgress []string
}

// RequestError contains http status code and message for a failed HTTP request
type RequestError struct {
	StatusCode int
	Message    string
}

// InitializeService sets up the service context for all API handlers
func initializeService(version string, cfg *configData) *serviceContext {
	ctx := serviceContext{Version: version,
		ImagesDir:   cfg.imagesDir,
		IIIFURL:     cfg.iiifURL,
		ScanDir:     cfg.scanDir,
		FinalizeDir: cfg.finalizeDir,
		JWTKey:      cfg.jwtKey,
		ServiceURL:  cfg.serviceURL,
		TrackSys:    cfg.tracksys,
		DevAuthUser: cfg.devAuthUser,
		BatchSize:   10} // for all parallel processing. number of images processed per batch

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
			log.Printf("ERROR: unable to create tmp directory")
			log.Fatalf("unable to make tmp dir %s: %s", tmpDir, err.Error())
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
	serviceOK := true

	hcMap["database"] = hcResp{Healthy: true}
	sqlDB, err := svc.DB.DB()
	if err != nil {
		hcMap["database"] = hcResp{Healthy: false, Message: err.Error()}
		serviceOK = false
	} else {
		err := sqlDB.Ping()
		if err != nil {
			hcMap["database"] = hcResp{Healthy: false, Message: err.Error()}
		}
	}

	hcMap["images-dir"] = hcResp{Healthy: true}
	err = validatePath(svc.ImagesDir)
	if err != nil {
		hcMap["images-dir"] = hcResp{Healthy: false, Message: err.Error()}
		serviceOK = false
	}

	hcMap["scan-dir"] = hcResp{Healthy: true}
	err = validatePath(svc.ScanDir)
	if err != nil {
		hcMap["scan-dir"] = hcResp{Healthy: false, Message: err.Error()}
		serviceOK = false
	}

	hcMap["finalize-dir"] = hcResp{Healthy: true}
	err = validatePath(svc.FinalizeDir)
	if err != nil {
		hcMap["finalize-dir"] = hcResp{Healthy: false, Message: err.Error()}
		serviceOK = false
	}

	hcMap["service"] = hcResp{Healthy: serviceOK}

	c.JSON(http.StatusOK, hcMap)
}

func validatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", path)
	}
	// NOTE: logic pulled from https://stackoverflow.com/questions/20026320/how-to-tell-if-folder-exists-and-is-writable
	if unix.Access(path, unix.W_OK) != nil {
		return fmt.Errorf("%s is not writable", path)
	}
	return nil
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

func (svc *serviceContext) isBatchInProgress(tgtUnitID string) bool {
	svc.batchMutex.Lock()
	defer svc.batchMutex.Unlock()
	for _, uid := range svc.BatchUnitsInProgress {
		if uid == tgtUnitID {
			return true
		}
	}
	return false
}
func (svc *serviceContext) addBatchProcess(tgtUnitID string) {
	svc.batchMutex.Lock()
	defer svc.batchMutex.Unlock()
	svc.BatchUnitsInProgress = append(svc.BatchUnitsInProgress, tgtUnitID)
	log.Printf("INFO: added unit %s to batch processing list: %+v", tgtUnitID, svc.BatchUnitsInProgress)
}
func (svc *serviceContext) removeBatchProcess(tgtUnitID string) {
	svc.batchMutex.Lock()
	defer svc.batchMutex.Unlock()
	tgtIdx := -1
	for idx, uid := range svc.BatchUnitsInProgress {
		if uid == tgtUnitID {
			tgtIdx = idx
			break
		}
	}

	lastIdx := len(svc.BatchUnitsInProgress) - 1
	if tgtIdx == lastIdx {
		// this is the last item in the list; just shorten the slice
		svc.BatchUnitsInProgress = svc.BatchUnitsInProgress[:lastIdx]
	} else {
		// copy the last element into the target index, then shorten the slice
		svc.BatchUnitsInProgress[tgtIdx] = svc.BatchUnitsInProgress[lastIdx]
		svc.BatchUnitsInProgress = svc.BatchUnitsInProgress[:lastIdx]
	}
	log.Printf("INFO: removed unit %s from batch processing list: %+v", tgtUnitID, svc.BatchUnitsInProgress)
}

func (svc *serviceContext) getConfig(c *gin.Context) {
	log.Printf("INFO: get service configuration")
	type ocrData struct {
		Hints     []ocrHint         `json:"hints"`
		Languages []ocrLanguageHint `json:"languages"`
	}
	type cfgData struct {
		TrackSysURL    string          `json:"tracksysURL"`
		JobsURL        string          `json:"jobsURL"`
		QAImageDir     string          `json:"qaImageDir"`
		ScanDir        string          `json:"scanDir"`
		Agencies       []agency        `json:"agencies"`
		Customers      []customer      `json:"customers"`
		Staff          []staffMember   `json:"staff"`
		Workstations   []workstation   `json:"workstations"`
		Workflows      []workflow      `json:"workflows"`
		Categories     []category      `json:"categories"`
		ContainerTypes []containerType `json:"containerTypes"`
		Problems       []problem       `json:"problems"`
		OCR            ocrData         `json:"ocr"`
		Steps          []string        `json:"steps"`
	}
	resp := cfgData{
		TrackSysURL: svc.TrackSys.Client,
		JobsURL:     svc.TrackSys.Jobs,
		QAImageDir:  svc.ImagesDir,
		ScanDir:     svc.ScanDir,
	}

	log.Printf("INFO: load staff members")
	rawResp, err := svc.getRequest(fmt.Sprintf("%s/staff", svc.TrackSys.API))
	if err != nil {
		log.Printf("ERROR: unable to get staff members: %s", err.Message)
		c.String(err.StatusCode, err.Message)
		return
	}
	if err := json.Unmarshal(rawResp, &resp.Staff); err != nil {
		log.Printf("ERROR: unable to parse staff members: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: load customers")
	rawResp, err = svc.getRequest(fmt.Sprintf("%s/customers", svc.TrackSys.API))
	if err != nil {
		log.Printf("ERROR: unable to get customers: %s", err.Message)
		c.String(err.StatusCode, err.Message)
		return
	}
	if err := json.Unmarshal(rawResp, &resp.Customers); err != nil {
		log.Printf("ERROR: unable to parse customers: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: load agencies")
	rawResp, err = svc.getRequest(fmt.Sprintf("%s/agencies", svc.TrackSys.API))
	if err != nil {
		log.Printf("ERROR: unable to get agencies: %s", err.Message)
		c.String(err.StatusCode, err.Message)
		return
	}
	if err := json.Unmarshal(rawResp, &resp.Agencies); err != nil {
		log.Printf("ERROR: unable to parse agencies: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: load container types")
	rawResp, err = svc.getRequest(fmt.Sprintf("%s/containertypes", svc.TrackSys.API))
	if err != nil {
		log.Printf("ERROR: unable to get container types: %s", err.Message)
		c.String(err.StatusCode, err.Message)
		return
	}
	if err := json.Unmarshal(rawResp, &resp.ContainerTypes); err != nil {
		log.Printf("ERROR: unable to parse container types: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: load ocr hints")
	rawResp, err = svc.getRequest(fmt.Sprintf("%s/ocr", svc.TrackSys.API))
	if err != nil {
		log.Printf("ERROR: unable to get ocr info: %s", err.Message)
		c.String(err.StatusCode, err.Message)
		return
	}
	if err := json.Unmarshal(rawResp, &resp.OCR); err != nil {
		log.Printf("ERROR: unable to ocr: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: load categories")
	dbResp := svc.DB.Order("name asc").Find(&resp.Categories)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load categories: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load active workstations")
	dbResp = svc.DB.Where("status=?", 0).Order("name asc").Find(&resp.Workstations)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load workstations: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load workflows")
	dbResp = svc.DB.Order("name asc").Preload("Steps").Find(&resp.Workflows)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load workflows: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load problems")
	dbResp = svc.DB.Where("label != ?", "Finalization").
		Where("label != ?", "Filesystem").
		Where("label != ?", "Filename").
		Order("name asc").Find(&resp.Problems)
	if dbResp.Error != nil {
		log.Printf("ERROR: unable to load problems: %s", dbResp.Error.Error())
		c.String(http.StatusInternalServerError, dbResp.Error.Error())
		return
	}

	log.Printf("INFO: load step names")
	resp.Steps = make([]string, 0)
	if err := svc.DB.Raw("select distinct(steps.name) from steps inner join workflows w on w.id = workflow_id and active=1").Scan(&resp.Steps).Error; err != nil {
		log.Printf("ERROR: unable to load step names: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
}

func padLeft(str string, tgtLen int) string {
	for {
		if len(str) == tgtLen {
			return str
		}
		str = "0" + str
	}
}

func (svc *serviceContext) getRequest(url string) ([]byte, *RequestError) {
	log.Printf("GET request: %s", url)
	startTime := time.Now()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-type", "application/json")
	httpClient := svc.HTTPClient
	rawResp, rawErr := httpClient.Do(req)
	resp, err := handleAPIResponse(url, rawResp, rawErr)
	elapsed := time.Since(startTime)
	elapsedMS := int64(elapsed / time.Millisecond)

	if err != nil {
		log.Printf("ERROR: Failed response from GET %s - %d:%s. Elapsed Time: %d (ms)",
			url, err.StatusCode, err.Message, elapsedMS)
	} else {
		log.Printf("Successful response from GET %s. Elapsed Time: %d (ms)", url, elapsedMS)
	}
	return resp, err
}

func (svc *serviceContext) postRequest(url string, payload any) ([]byte, *RequestError) {
	log.Printf("POST request: %s", url)
	startTime := time.Now()
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Add("Content-type", "application/json")
	httpClient := svc.HTTPClient
	rawResp, rawErr := httpClient.Do(req)
	resp, err := handleAPIResponse(url, rawResp, rawErr)
	elapsed := time.Since(startTime)
	elapsedMS := int64(elapsed / time.Millisecond)

	if err != nil {
		log.Printf("ERROR: Failed response from POST %s - %d:%s. Elapsed Time: %d (ms)",
			url, err.StatusCode, err.Message, elapsedMS)
	} else {
		log.Printf("Successful response from POST %s. Elapsed Time: %d (ms)", url, elapsedMS)
	}
	return resp, err
}

func handleAPIResponse(logURL string, resp *http.Response, err error) ([]byte, *RequestError) {
	if err != nil {
		status := http.StatusBadRequest
		errMsg := err.Error()
		if strings.Contains(err.Error(), "Timeout") {
			status = http.StatusRequestTimeout
			errMsg = fmt.Sprintf("%s timed out", logURL)
		} else if strings.Contains(err.Error(), "connection refused") {
			status = http.StatusServiceUnavailable
			errMsg = fmt.Sprintf("%s refused connection", logURL)
		}
		return nil, &RequestError{StatusCode: status, Message: errMsg}
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		status := resp.StatusCode
		errMsg := string(bodyBytes)
		return nil, &RequestError{StatusCode: status, Message: errMsg}
	}

	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	return bodyBytes, nil
}

func exists(tgtDir string) bool {
	log.Printf("INFO: check existance of %s", tgtDir)
	_, err := os.Stat(tgtDir)
	if err != nil {
		log.Printf("INFO: %s does not exist", tgtDir)
		return false
	}
	return true
}

func findFile(basePath, tgtFileName string) string {
	fullPath := ""
	filepath.WalkDir(basePath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if entry.Name() == tgtFileName {
			fullPath = path
		}
		return nil
	})
	return fullPath
}
