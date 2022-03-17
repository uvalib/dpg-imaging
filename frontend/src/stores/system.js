import { defineStore } from 'pinia'
import axios from 'axios'

export const useSystemStore = defineStore('system', {
	state: () => ({
      loading: false,
      updating: false,
		version: "unknown",
		error: "",
      staffMembers: [],
      agencies: [],
      workstations: [],
      workflows: [],
      categories: [],
      containerTypes: [],
      ocrHints: [],
      ocrLanguageHints: [],
      problemTypes: [],
      adminURL: "",
      qaDir: "",
      scanDir: "",
	}),
	getters: {
	},
	actions: {
		getVersion() {
         axios.get("/version").then(response => {
            this.version = `v${response.data.version}-${response.data.build}`
         }).catch(e => {
            this.error = e
         })
      },
      getConfig() {
         this.loading = true
         axios.get("/config").then(response => {
            this.adminURL = response.data.tracksysURL
            this.qaDir =  response.data.qaImageDir
            this.scanDir =  response.data.scanDir
            this.staffMembers = response.data.staff
            this.workflows= response.data.workflows
            this.workstations = response.data.workstations
            this.categories = response.data.categories
            this.ocrHints = response.data.ocrHints
            this.ocrLanguageHints = response.data.ocrLanguageHints
            this.problemTypes = response.data.problems
            this.containerTypes = response.data.containerTypes
            this.agencies = response.data.agencies
            this.loading = false
         }).catch( e => {
            this.error =  e
            this.loading = true
         })
      },
	}
})