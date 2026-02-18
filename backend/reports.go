package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

type unitImageRec struct {
	ID         int64
	ImageCount uint
}

// page time rates report structs
type pageTimeStats struct {
	TotalMins   int64   `json:"mins"`
	TotalImages uint    `json:"images"`
	TotalUnits  uint    `json:"units"`
	AvgPageTime float64 `json:"avgPageTime"`
}

// scan/qa rates structs
type rateStats struct {
	UnitIDs []int64 `json:"-"`
	Images  uint    `json:"images"`
	Minutes int64   `json:"minutes"`
	Rate    float64 `json:"rate"`
}

type ratesRespRec struct {
	StaffID int64      `json:"staffID"`
	Scans   *rateStats `json:"scans"`
	QA      *rateStats `json:"qa"`
}

// rejectection rates structs
type scansRejectStats struct {
	Projects    uint    `json:"projects"`
	Images      uint    `json:"images"`
	Rejections  int64   `json:"rejections"`
	ProjectRate float64 `json:"projectRate"`
	ImageRate   float64 `json:"imageRate"`
}
type qaRejectStats struct {
	Projects   uint    `json:"projects"`
	Rejections int64   `json:"rejections"`
	Rate       float64 `json:"rate"`
}
type rejectRespRec struct {
	StaffID int64             `json:"staffID"`
	Scans   *scansRejectStats `json:"scans"`
	QA      *qaRejectStats    `json:"qa"`
}

func (svc *serviceContext) getProductivityReport(c *gin.Context) {
	log.Printf("INFO: get productivity report")
	workflowID := c.Query("workflow")
	startDate := c.Query("start")
	endDate := c.Query("end")

	var resp struct {
		CompletedProjects int      `json:"completedProjects"`
		Types             []string `json:"types"`
		Productivity      []int64  `json:"productivity"`
	}
	type productivityRec struct {
		Type  string
		Count int64
	}
	var dbData []productivityRec
	err := svc.DB.Table("projects").Select("c.name as type, count(projects.id) as count").
		Joins("inner join categories c on c.id = category_id").
		Where("workflow_id=?", workflowID).
		Where("finished_at >= ? and finished_at <= ?", startDate, endDate).
		Group("c.id").Find(&dbData).Error
	if err != nil {
		log.Printf("ERROR: unable to get productivity report: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	for _, p := range dbData {
		resp.CompletedProjects += int(p.Count)
		resp.Types = append(resp.Types, p.Type)
		resp.Productivity = append(resp.Productivity, p.Count)
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) getProblemsReport(c *gin.Context) {
	log.Printf("INFO: get problems report")
	workflowID := c.Query("workflow")
	startDate := c.Query("start")
	endDate := c.Query("end")

	resp := struct {
		Types    []string `json:"types"`
		Problems []int64  `json:"problems"`
	}{
		Types:    make([]string, 0),
		Problems: make([]int64, 0),
	}
	type problemRec struct {
		Type  string
		Count int64
	}

	var dbData []problemRec
	err := svc.DB.Table("notes").Select("pb.label as type, count(notes.id) as count").
		Joins("inner join notes_problems np on np.note_id = notes.id").
		Joins("inner join problems pb on pb.id = np.problem_id").
		Joins("inner join projects p on project_id = p.id").
		Where("note_type=?", 2).Where("pb.label <> 'Finalization'").
		Where("finished_at is not null").Where("workflow_id=?", workflowID).
		Where("finished_at >= ? and finished_at <= ?", startDate, endDate).
		Group("problem_id").Find(&dbData).Error
	if err != nil {
		log.Printf("ERROR: unable to get problems projects: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	for _, p := range dbData {
		resp.Types = append(resp.Types, p.Type)
		resp.Problems = append(resp.Problems, p.Count)
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) getRateReports(c *gin.Context) {
	log.Printf("INFO: get data for three rate reports: average page times, rejections, rates")
	workflowID := c.Query("workflow")
	startDate := c.Query("start")
	endDate := c.Query("end")

	// all 3 reports need total masterfiles for all units in the date range from project
	// using the supplied workflow. Get this info first and share it with all reports (its expensive)
	unitImages, cntErr := svc.getUnitImagesCount(workflowID, startDate, endDate)
	if cntErr != nil {
		log.Printf("ERROR: unit master file counts for the rate reports: %s", cntErr.Error())
		c.String(http.StatusInternalServerError, cntErr.Error())
		return
	}

	var out struct {
		RatesReport    []ratesRespRec            `json:"ratesReport"`
		RejectReport   []rejectRespRec           `json:"rejectionsReport"`
		PageTimeRepors map[string]*pageTimeStats `json:"pageTimesReport"`
	}

	ratesReport, err := svc.getRatesReport(workflowID, startDate, endDate, unitImages)
	if err != nil {
		log.Printf("ERROR: unable to get rates report: %s", err)
	} else {
		out.RatesReport = ratesReport
	}
	rejReport, err := svc.getRejectionsReport(workflowID, startDate, endDate, unitImages)
	if err != nil {
		log.Printf("ERROR: unable to get rates report: %s", err)
	} else {
		out.RejectReport = rejReport
	}
	timeReport, err := svc.getPageTimesReport(workflowID, startDate, endDate, unitImages)
	if err != nil {
		log.Printf("ERROR: unable to get page times report: %s", err)
	} else {
		out.PageTimeRepors = timeReport
	}

	c.JSON(http.StatusOK, out)
}

func (svc *serviceContext) getPageTimesReport(workflowID, startDate, endDate string, unitImages []unitImageRec) (map[string]*pageTimeStats, error) {
	log.Printf("INFO: get average page times report")
	resp := make(map[string]*pageTimeStats)

	log.Printf("INFO: get all catagories")
	type category struct {
		ID   int64
		Name string
	}
	var categories []category
	if err := svc.DB.Find(&categories).Error; err != nil {
		return resp, err
	}

	log.Printf("INFO: get unit timing")
	type timingRec struct {
		ProjectID int64
		UnitID    int64
		Category  string
		TotalMins int64
	}

	var timings []timingRec
	if err := svc.DB.Table("projects").Select("projects.id as project_id, projects.unit_id as unit_id, c.name as category, sum(duration_minutes) as total_mins").
		Joins("inner join assignments a on projects.id = a.project_id").
		Joins("inner join categories c on c.id = projects.category_id").
		Where("workflow_id=?", workflowID).
		Where("projects.finished_at >= ?", startDate).
		Where("projects.finished_at <= ?", endDate).
		Group("projects.id").Find(&timings).Error; err != nil {
		return resp, err
	}

	log.Printf("INFO: generate page time report from collected stats")

	// init response with blank stats rec for each category
	for _, c := range categories {
		resp[c.Name] = &pageTimeStats{}
	}

	// sum up unit / image counts for each category
	for _, t := range timings {
		tgtStats := resp[t.Category]
		for _, u := range unitImages {
			if u.ID == t.UnitID {
				tgtStats.TotalUnits++
				tgtStats.TotalImages += u.ImageCount
				tgtStats.TotalMins += t.TotalMins
				break
			}
		}
	}

	// calculate average page time for each category
	for _, stats := range resp {
		if stats.TotalImages > 0 {
			stats.AvgPageTime = float64(stats.TotalMins) / float64(stats.TotalImages)
		}
	}

	return resp, nil
}

func (svc *serviceContext) getRejectionsReport(workflowID, startDate, endDate string, unitImages []unitImageRec) ([]rejectRespRec, error) {
	log.Printf("INFO: merge db results into rejections report")
	resp := make([]rejectRespRec, 0)

	log.Printf("INFO: lookup finished/rejected projects counts")
	type projRec struct {
		ID       int64
		UnitID   int64
		StaffID  int64
		StepType int64
		Status   int64
	}
	// NOTE: sort the results by project ID to prevent assignment data for multiple projects to be mixed.
	// With this in place, data can be iterated knowing that all projects changes happen in sequence.
	var projs []projRec
	err := svc.DB.Table("projects").Select("projects.id as id, projects.unit_id as unit_id, a.staff_member_id as staff_id, s.step_type, a.status").
		Joins("inner join assignments a on a.project_id = projects.id").
		Joins("inner join steps s on s.id = a.step_id").
		Where("a.status>=2").Where("a.status<=4"). // finished, rejected or error
		Where(
			svc.DB.Where("s.step_type = 0").Or("s.fail_step_id is not null"), // scan or any step that can be rejected
		).
		Where("projects.workflow_id=?", workflowID).
		Where("projects.finished_at >= ?", startDate).
		Where("projects.finished_at <= ?", endDate).
		Order("projects.id asc").Find(&projs).Error
	if err != nil {
		return resp, err
	}

	var scannerID int64
	for _, p := range projs {

		if p.StepType == 0 {
			// scan step. track which user originally did the scan
			scannerID = p.StaffID
		}
		var rec rejectRespRec
		for _, exist := range resp {
			if exist.StaffID == p.StaffID {
				rec = exist
				break
			}
		}
		if rec.StaffID == 0 {
			rec = rejectRespRec{StaffID: p.StaffID, Scans: &scansRejectStats{}, QA: &qaRejectStats{}}
			resp = append(resp, rec)
		}

		if p.StepType != 0 {
			// qa step
			rec.QA.Projects++
			if p.Status == 3 {
				log.Printf("INFO: rejection on project %d; scanner %d", p.ID, scannerID)
				// rejected. add one to this user qa rejects and one to the original scanner scan rejects
				rec.QA.Rejections++
				for _, test := range resp {
					if test.StaffID == scannerID {
						test.Scans.Rejections++
						break
					}
				}
			}
		} else {
			// scan step
			rec.Scans.Projects++
			for _, unitMfRec := range unitImages {
				if unitMfRec.ID == p.UnitID {
					rec.Scans.Images += unitMfRec.ImageCount
					break
				}
			}
		}
	}

	for _, v := range resp {
		if v.Scans.Projects > 0 {
			v.Scans.ProjectRate = (float64(v.Scans.Rejections) / float64(v.Scans.Projects)) * 100.0
		}
		if v.Scans.Images > 0 {
			v.Scans.ImageRate = (float64(v.Scans.Rejections) / float64(v.Scans.Images)) * 100.0
		}
		if v.QA.Projects > 0 {
			v.QA.Rate = (float64(v.QA.Rejections) / float64(v.QA.Projects)) * 100.0
		}
	}
	return resp, nil
}

func (svc *serviceContext) getRatesReport(workflowID, startDate, endDate string, unitImages []unitImageRec) ([]ratesRespRec, error) {
	log.Printf("INFO: lookup finished/rejected projects for rates report")
	resp := make([]ratesRespRec, 0)

	log.Printf("INFO: lookup finished/rejected projects info")
	type projRec struct {
		ID              int64
		UnitID          int64
		StaffID         int64
		StepName        string
		DurationMinutes int64
	}
	// NOTE: sort the results by project ID to prevent assignment data for multiple projects to be mixed.
	// With this in place, data can be iterated knowing that all projects changes happen in sequence.
	var projs []projRec
	err := svc.DB.Table("projects").
		Select("projects.id as id, projects.unit_id as unit_id, a.staff_member_id as staff_id, s.name as step_name, a.duration_minutes").
		Joins("inner join assignments a on a.project_id = projects.id").
		Joins("inner join steps s on s.id = a.step_id").
		Where("a.status>=2").Where("a.status<=4"). // finished, rejected or error
		Where("a.duration_minutes is not null").   // valid duration
		Where("s.step_type != 2").                 // no error steps
		Where("projects.workflow_id=?", workflowID).
		Where("projects.finished_at >= ?", startDate).
		Where("projects.finished_at <= ?", endDate).
		Order("projects.id asc").Find(&projs).Error
	if err != nil {
		return resp, err
	}

	log.Printf("INFO: merge db results into rates report")
	var currUnitID int64
	var currUnitImageCnt uint
	scanSteps := []string{"Scan", "Process", "Create Metadata"}
	for _, p := range projs {

		// check for existing record for this user
		var rec ratesRespRec
		for _, exist := range resp {
			if exist.StaffID == p.StaffID {
				rec = exist
				break
			}
		}

		// user rec not found; create a new one
		if rec.StaffID == 0 {
			rec = ratesRespRec{StaffID: p.StaffID,
				Scans: &rateStats{UnitIDs: make([]int64, 0)}, QA: &rateStats{UnitIDs: make([]int64, 0)}}
			resp = append(resp, rec)
		}

		// determine if this step is scan or QA and grab a pointer to teh correct user stats
		tgtRateStats := rec.QA
		for _, scanName := range scanSteps {
			if scanName == p.StepName {
				tgtRateStats = rec.Scans
				break
			}
		}

		// the first time a new unitID is encountered get the masterfile count
		if currUnitID != p.UnitID {
			currUnitID = p.UnitID
			for _, unitMfRec := range unitImages {
				if unitMfRec.ID == p.UnitID {
					currUnitImageCnt = unitMfRec.ImageCount
					break
				}
			}
		}

		unitCounted := slices.Contains(tgtRateStats.UnitIDs, p.UnitID)
		if unitCounted == false {
			tgtRateStats.Images += currUnitImageCnt
			tgtRateStats.UnitIDs = append(tgtRateStats.UnitIDs, p.UnitID)
		}
		tgtRateStats.Minutes += p.DurationMinutes
		tgtRateStats.Rate = float64(tgtRateStats.Images) / float64(tgtRateStats.Minutes)
	}

	return resp, nil
}

func (svc *serviceContext) getUnitImagesCount(workflowID string, startDate string, endDate string) ([]unitImageRec, error) {
	log.Printf("INFO: get unit masterfile counts")
	var unitImages []unitImageRec
	var unitIDs []int64
	unitQ := "select unit_id from projects where workflow_id=? and finished_at >? and finished_at <=?"
	if err := svc.DB.Raw(unitQ, workflowID, startDate, endDate).Scan(&unitIDs).Error; err != nil {
		return unitImages, err
	}
	log.Printf("INFO: %d units found for workflow %s between %s and %s", len(unitIDs), workflowID, startDate, endDate)
	for _, uID := range unitIDs {
		respBytes, reqErr := svc.getRequest(fmt.Sprintf("%s/units/%d", svc.TrackSys.API, uID))
		if reqErr != nil {
			log.Printf("ERROR: unable to get unit info for %d: %s", uID, reqErr.Message)
			continue
		}

		var unitMD unitInfo
		if err := json.Unmarshal(respBytes, &unitMD); err != nil {
			log.Printf("ERROR: unable to parse unit info for %d: %s", uID, err.Error())
			continue
		}
		unitImages = append(unitImages, unitImageRec{ID: uID, ImageCount: unitMD.MasterFileCount})
	}
	return unitImages, nil
}
