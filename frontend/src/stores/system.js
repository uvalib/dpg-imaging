import { defineStore } from 'pinia'
import axios from 'axios'

export const useSystemStore = defineStore('system', {
	state: () => ({
      working: false,
		version: "unknown",
		error: "",
      showError: false,
      staffMembers: [],
      agencies: [],
      workstations: [],
      workflows: [],
      steps: [],
      categories: [],
      containerTypes: [],
      ocrHints: [],
      ocrLanguageHints: [],
      problemTypes: [],
      adminURL: "",
      jobsURL: "",
      qaDir: "",
      scanDir: "",
	}),
	getters: {
	},
	actions: {
      setError( e ) {
         this.error = e
         this.working = false
         if (this.error && this.error.length > 0) {
            this.showError = true
         } else {
            this.showError = false
         }
      },
		getVersion() {
         axios.get("/version").then(response => {
            this.version = `v${response.data.version}-${response.data.build}`
         }).catch(e => {
            this.error = e
         })
      },
      getConfig() {
         this.working = true
         axios.get("/config").then(response => {
            this.jobsURL = response.data.jobsURL
            this.adminURL = response.data.tracksysURL
            this.qaDir =  response.data.qaImageDir
            this.scanDir =  response.data.scanDir
            this.staffMembers = response.data.staff
            this.workflows = response.data.workflows
            this.steps = response.data.steps
            this.workstations = response.data.workstations
            this.categories = response.data.categories
            this.ocrHints = response.data.ocrHints
            this.ocrLanguageHints = response.data.ocrLanguageHints
            this.problemTypes = response.data.problems
            this.containerTypes = response.data.containerTypes
            this.agencies = response.data.agencies
            this.working = false
         }).catch( e => {
            this.error =  e
            this.working = false
         })
      },
	}
})