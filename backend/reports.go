package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type unitImageRec struct {
	ID         int64
	ImageCount int64
}

func (svc *serviceContext) getDeliveriesReport(c *gin.Context) {
	log.Printf("INFO: get deliveries report")
	tgtYear := c.Query("year")
	if tgtYear == "" {
		log.Printf("ERROR: year is required")
		c.String(http.StatusBadRequest, "year is required")
		return
	}

	// NOTE: in the response, the totals are an array of 12 counts. each count corresponds to a month
	months := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	var resp struct {
		Months []string `json:"months"`
		Total  []int    `json:"total"`
		Late   []int    `json:"late"`
		OnTime []int    `json:"onTime"`
	}
	resp.Months = months

	type order struct {
		ID            int64
		DateDue       time.Time
		DateCompleted time.Time
	}
	var completedOrders []order
	err := svc.DB.Joins("inner join units u on order_id = orders.id").
		Where("date_patron_deliverables_ready is not null").
		Where("order_status=?", "completed").
		Where("date_completed like ?", fmt.Sprintf("%s%%", tgtYear)).
		Distinct("orders.id", "date_due", "date_completed").
		Order("date_completed asc").
		Find(&completedOrders).Error
	if err != nil {
		log.Printf("ERROR: unable to get deliveries report: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// orders are sorted by completion date asc. track the current month and add the
	// count to the correct slot in the total/late/ontime arrays.
	// NOTE: the logic below does not account for skipped months as this should never happen
	currMonth := 0
	total := 0
	late := 0
	ontime := 0
	for _, o := range completedOrders {
		orderMonth := int(o.DateCompleted.Month())
		if currMonth != orderMonth {
			if currMonth > 0 {
				resp.Total = append(resp.Total, total)
				resp.Late = append(resp.Late, late)
				resp.OnTime = append(resp.OnTime, ontime)
			}
			total = 0
			late = 0
			ontime = 0
			currMonth = orderMonth
		}
		total++
		if o.DateCompleted.Before(o.DateDue) {
			ontime++
		} else {
			late++
		}
	}

	if total > 0 {
		resp.Total = append(resp.Total, total)
		resp.Late = append(resp.Late, late)
		resp.OnTime = append(resp.OnTime, ontime)
	}

	c.JSON(http.StatusOK, resp)
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

	var resp struct {
		Types    []string `json:"types"`
		Problems []int64  `json:"problems"`
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
		log.Printf("ERROR: unable to get productivity report: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	for _, p := range dbData {
		resp.Types = append(resp.Types, p.Type)
		resp.Problems = append(resp.Problems, p.Count)
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) getPageTimesReport(c *gin.Context) {
	log.Printf("INFO: get average page times report")
	workflowID := c.Query("workflow")
	startDate := c.Query("start")
	endDate := c.Query("end")

	log.Printf("INFO: get all catagories")
	type category struct {
		ID   int64
		Name string
	}
	var categories []category
	err := svc.DB.Find(&categories).Error
	if err != nil {
		log.Printf("ERROR: unable to get categories: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: get unit timing")
	type timingRec struct {
		ProjectID int64
		UnitID    int64
		Category  string
		TotalMins int64
	}

	var timings []timingRec
	err = svc.DB.Table("projects").Select("projects.id as project_id, projects.unit_id as unit_id, c.name as category, sum(duration_minutes) as total_mins").
		Joins("inner join assignments a on projects.id = a.project_id").
		Joins("inner join categories c on c.id = projects.category_id").
		Where("workflow_id=?", workflowID).
		Where("projects.finished_at >= ?", startDate).
		Where("projects.finished_at <= ?", endDate).
		Group("projects.id").Find(&timings).Error
	if err != nil {
		log.Printf("ERROR: unable to get project timing report: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("INFO: unit master file counts")
	unitImageCounts, uicErr := svc.getUnitImagesCount(workflowID, startDate, endDate)
	if uicErr != nil {
		log.Printf("ERROR: unable to get unit image count report: %s", uicErr.Error())
		c.String(http.StatusInternalServerError, uicErr.Error())
		return
	}

	log.Printf("INFO: generate report from collected stats")
	type Stats struct {
		TotalMins   int64   `json:"mins"`
		TotalImages int64   `json:"images"`
		TotalUnits  int64   `json:"units"`
		AvgPageTime float64 `json:"avgPageTime"`
	}

	// init response with blank stats rec for each category
	resp := make(map[string]*Stats)
	for _, c := range categories {
		resp[c.Name] = &Stats{}
	}

	// sum up unit / image counts for each category
	for _, t := range timings {
		tgtStats := resp[t.Category]
		for _, u := range unitImageCounts {
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

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) getRejectionsReport(c *gin.Context) {
	log.Printf("INFO: get average page times report")
	workflowID := c.Query("workflow")
	startDate := c.Query("start")
	endDate := c.Query("end")

	log.Printf("INFO: lookup finished/rejected projects counts")
	type projRec struct {
		ID        int64
		UnitID    int64
		StaffID   int64
		LastName  string
		FirstName string
		StepType  int64
		Status    int64
	}
	// NOTE: sort the results by project ID to prevent assignment data for multiple projects to be mixed.
	// With this in place, data can be iterated knowing that all projects changes happen in sequence.
	var projs []projRec
	err := svc.DB.Table("projects").Select("projects.id as id, projects.unit_id as unit_id, m.id as staff_id, last_name, first_name, s.step_type, a.status").
		Joins("inner join assignments a on a.project_id = projects.id").
		Joins("inner join staff_members m on m.id = staff_member_id").
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
		log.Printf("ERROR: unable to get rejection stats: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	unitImages, cntErr := svc.getUnitImagesCount(workflowID, startDate, endDate)
	if cntErr != nil {
		log.Printf("ERROR: unit master file counts: %s", cntErr.Error())
		c.String(http.StatusInternalServerError, cntErr.Error())
		return
	}

	log.Printf("INFO: merge db results into rejections report")
	type scansStats struct {
		Projects    int64   `json:"projects"`
		Images      int64   `json:"images"`
		Rejections  int64   `json:"rejections"`
		ProjectRate float64 `json:"projectRate"`
		ImageRate   float64 `json:"imageRate"`
	}
	type qaStats struct {
		Projects   int64   `json:"projects"`
		Rejections int64   `json:"rejections"`
		Rate       float64 `json:"rate"`
	}
	type respRec struct {
		StaffID   int64
		StaffName string      `json:"staffName"`
		Scans     *scansStats `json:"scans"`
		QA        *qaStats    `json:"qa"`
	}
	resp := make([]*respRec, 0)
	var scannerID int64
	for _, p := range projs {
		var rec *respRec
		if p.StepType == 0 {
			// scan step. track which user originally did the scan
			scannerID = p.StaffID
		}
		for _, exist := range resp {
			if exist.StaffID == p.StaffID {
				rec = exist
				break
			}
		}
		if rec == nil {
			rec = &respRec{StaffID: p.StaffID, StaffName: fmt.Sprintf("%s, %s", p.LastName, p.FirstName), Scans: &scansStats{}, QA: &qaStats{}}
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

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) getRatesReport(c *gin.Context) {
	log.Printf("INFO: get average page times report")
	workflowID := c.Query("workflow")
	startDate := c.Query("start")
	endDate := c.Query("end")

	log.Printf("INFO: lookup finished/rejected projects info")
	type projRec struct {
		ID              int64
		UnitID          int64
		StaffID         int64
		LastName        string
		FirstName       string
		StepName        string
		DurationMinutes int64
	}
	// NOTE: sort the results by project ID to prevent assignment data for multiple projects to be mixed.
	// With this in place, data can be iterated knowing that all projects changes happen in sequence.
	var projs []projRec
	err := svc.DB.Table("projects").
		Select("projects.id as id, projects.unit_id as unit_id, m.id as staff_id, last_name, first_name, s.name as step_name, a.duration_minutes").
		Joins("inner join assignments a on a.project_id = projects.id").
		Joins("inner join staff_members m on m.id = staff_member_id").
		Joins("inner join steps s on s.id = a.step_id").
		Where("a.status>=2").Where("a.status<=4"). // finished, rejected or error
		Where("a.duration_minutes is not null").   // valid duration
		Where("s.step_type != 2").                 // no error steps
		Where("projects.workflow_id=?", workflowID).
		Where("projects.finished_at >= ?", startDate).
		Where("projects.finished_at <= ?", endDate).
		Order("projects.id asc").Find(&projs).Error
	if err != nil {
		log.Printf("ERROR: unable to get rates stats: %s", err.Error())
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	unitImages, cntErr := svc.getUnitImagesCount(workflowID, startDate, endDate)
	if cntErr != nil {
		log.Printf("ERROR: unit master file counts for rates report: %s", cntErr.Error())
		c.String(http.StatusInternalServerError, cntErr.Error())
		return
	}

	log.Printf("INFO: merge db results into rates report")
	type rateStats struct {
		UnitIDs []int64 `json:"-"`
		Images  int64   `json:"images"`
		Minutes int64   `json:"minutes"`
		Rate    float64 `json:"rate"`
	}
	type respRec struct {
		StaffID   int64
		StaffName string     `json:"staffName"`
		Scans     *rateStats `json:"scans"`
		QA        *rateStats `json:"qa"`
	}
	resp := make([]*respRec, 0)
	var currUnitID int64
	var currUnitImageCnt int64
	scanSteps := []string{"Scan", "Process", "Create Metadata"}
	for _, p := range projs {

		// check for existing record for this user
		var rec *respRec
		for _, exist := range resp {
			if exist.StaffID == p.StaffID {
				rec = exist
				break
			}
		}

		// user rec not found; create a new one
		if rec == nil {
			rec = &respRec{StaffID: p.StaffID, StaffName: fmt.Sprintf("%s, %s", p.LastName, p.FirstName),
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

		unitCounted := false
		for _, uid := range tgtRateStats.UnitIDs {
			if uid == p.UnitID {
				unitCounted = true
				break
			}
		}
		if unitCounted == false {
			tgtRateStats.Images += currUnitImageCnt
			tgtRateStats.UnitIDs = append(tgtRateStats.UnitIDs, p.UnitID)
		}
		tgtRateStats.Minutes += p.DurationMinutes
		tgtRateStats.Rate = float64(tgtRateStats.Images) / float64(tgtRateStats.Minutes)
	}

	c.JSON(http.StatusOK, resp)
}

func (svc *serviceContext) getUnitImagesCount(workflowID string, startDate string, endDate string) ([]unitImageRec, error) {
	log.Printf("INFO: get unit masterfile counts")
	var unitImages []unitImageRec
	err := svc.DB.Table("units").Select("units.id, count(m.id) image_count").
		Joins("inner join projects p on unit_id = units.id").
		Joins("inner join master_files m on m.unit_id = units.id").
		Where("p.workflow_id=?", workflowID).
		Where("p.finished_at >= ?", startDate).
		Where("p.finished_at <= ?", endDate).
		Group("units.id").
		Find(&unitImages).Error
	if err != nil {
		return nil, err
	}
	return unitImages, nil
}
