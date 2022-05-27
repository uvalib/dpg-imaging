import { defineStore } from 'pinia'
import axios from 'axios'

export const useSystemStore = defineStore('system', {
	state: () => ({
      working: false,
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
      setError( e ) {
         this.error = e
         this.working = false
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
            this.working = false
         }).catch( e => {
            this.error =  e
            this.working = false
         })
      },
	}
})