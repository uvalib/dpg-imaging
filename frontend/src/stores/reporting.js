import { defineStore } from 'pinia'
import axios from 'axios'
import dayjs from 'dayjs'

export const useReportStore = defineStore('report', {
	state: () => ({
		workflowID: 1,
		startDate: null,
		endDate: null,
		productivity: {
			loading: false,
			labels: [],
			datasets: [],
			totalCompleted: 0,
			error: ""
		},
		problems: {
			loading: false,
			labels: [],
			datasets: [],
			totalProjects: 0,
			error: ""
		},
		reports: {
			loading: false,
			error: "",
			pageTimes: {
				labels: [],
				datasets: [],
				raw: []
			},
			rejections: {
				data: [],
			},
			rates: {
				data: [],
			}
		}
	}),
	getters: {
	},
	actions: {
		init() {
			var oldDate = new Date()
			oldDate.setMonth(oldDate.getMonth() - 3)
			this.startDate  = oldDate
			this.endDate = new Date()
		},

		clearStats() {
			this.productivity.loading = false
			this.productivity.datasets = []
			this.productivity.totalCompleted = 0
			this.productivity.error = ""

			this.problems.loading = false
			this.problems.datasets = []
			this.problems.totalProjects = 0

			this.reports.loading = false
			this.reports.pageTimes.datasets = []
			this.reports.raw = []
			this.reports.rejections.data = []
			this.reports.rates.data = []
		},

		getProductivityReport( workflowID, start, end ) {
			let url = `/api/reports/productivity?workflow=${workflowID}&start=${dayjs(start).format("YYYY-MM-DD")}&end=${dayjs(end).format("YYYY-MM-DD")}`
			this.productivity.loading = true
			axios.get(url).then(response => {
				this.productivity.labels = response.data.types
				let prodDataset = [{data: response.data.productivity, backgroundColor: "#44aacc"}]
				this.productivity.datasets = prodDataset
				this.productivity.totalCompleted = response.data.completedProjects
				this.productivity.loading = false
				this.productivity.error = ""
			}).catch(e => {
            this.productivity.error = e
				this.productivity.loading = false
         })
		},
		getProblemsReport( workflowID, start, end ) {
			let url = `/api/reports/problems?workflow=${workflowID}&start=${dayjs(start).format("YYYY-MM-DD")}&end=${dayjs(end).format("YYYY-MM-DD")}`
			this.problems.loading = true
			this.problems.error = ""
			axios.get(url).then(response => {
				this.problems.labels = response.data.types
				let dataset = {data: response.data.problems, backgroundColor: "#cc4444"}
				this.problems.datasets = []
				this.problems.datasets.push(dataset)
				this.problems.totalProjects = response.data.totalProjects
				this.problems.loading = false
			}).catch(e => {
            this.problems.error = e
				this.problems.loading = false
         })
		},
		getRateReports( workflowID, start, end ) {
			let url = `/api/reports/rates?workflow=${workflowID}&start=${dayjs(start).format("YYYY-MM-DD")}&end=${dayjs(end).format("YYYY-MM-DD")}`
			this.reports.loading = true
			this.reports.error = ""
			axios.get(url).then(response => {
				this.reports.rates.data = response.data.ratesReport
				this.reports.rejections.data = response.data.rejectionsReport

				this.reports.pageTimes.labels = []
				this.reports.pageTimes.datasets = []
				this.reports.pageTimes.raw = []
				let timeDS = {data: [], backgroundColor: "#44aacc"}
				for (const [category, stats] of Object.entries(response.data.pageTimesReport)) {
					this.reports.pageTimes.labels.push(category)
					timeDS.data.push(stats.avgPageTime)
					let row = {category: category, units: stats.units, totalMins: stats.mins,
						totalPages: stats.images, avgPageTime:  Number.parseFloat(stats.avgPageTime).toFixed(2)}
					this.reports.pageTimes.raw.push(row)
				}
				this.reports.pageTimes.datasets.push(timeDS)

				this.reports.loading = false
			}).catch(e => {
            this.reports.error = e
				this.reports.loading = false
         })
		},
	}
})
