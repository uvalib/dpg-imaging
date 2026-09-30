import { defineStore } from 'pinia'
import axios from 'axios'

export const useReportStore = defineStore('report', {
	state: () => ({
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

		getProductivityReport( workflowID, startStr, endStr ) {
			let url = `/api/reports/productivity?workflow=${workflowID}&start=${startStr}&end=${endStr}`
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
		getProblemsReport( workflowID, startStr, endStr ) {
			let url = `/api/reports/problems?workflow=${workflowID}&start=${startStr}&end=${endStr}`
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
		getRateReports( workflowID, startStr, endStr ) {
			let url = `/api/reports/rates?workflow=${workflowID}&start=${startStr}&end=${endStr}`
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
