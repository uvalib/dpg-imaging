import { defineStore } from 'pinia'
import axios from 'axios'

export const useSystemStore = defineStore('system', {
	state: () => ({
      working: false,
		version: "unknown",
		error: "",
      showError: false,
      customers: [],
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
      titleVocab: [
         "Spine", "Front Cover", "Front Cover verso", "Back Cover", "Back Cover recto",
         "Head", "Tail", "Fore-edge", "Front Paste-down Endpaper", "Front Free Endpaper",
         "Rear Paste-down Endpaper", "Rear Free Endpaper", "Plate",
         "Blank Page", "Half-title", "Frontispiece", "Printer's Imprint", "Copyright",
         "Privilege", "Ad Lectorem",  "Table of Contents", "Titlepage", "Device",
         "Epigraph", "Prologue/Preface", "Dedication", "Errata"].sort()
	}),
	getters: {
	},
	actions: {
      setError( e ) {
         this.error = e
         if (e.response && e.response.data) {
            this.error = e.response.data
         }
         this.showError = true
         this.working = false
      },
      clearError() {
         this.error = ""
         this.showError = false
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
            this.customers = response.data.customers
            this.staffMembers = response.data.staff
            this.workflows = response.data.workflows
            this.steps = response.data.steps
            this.workstations = response.data.workstations
            this.categories = response.data.categories
            this.ocrHints = response.data.ocr.hints
            this.ocrLanguageHints = response.data.ocr.languages
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