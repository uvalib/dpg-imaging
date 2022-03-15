import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useProjectStore = defineStore('project', {
   state: () => ({
      system: useSystemStore(),
      projects: [],
      selectedProjectIdx: -1,
      totals: {
         me: 0,
         active: 0,
         unassigned: 0,
         finished: 0
      },
      pageSize: 10,
      currPage: 1,
      candidates: [],
      working: false,
      filter: "active",
      search: {
         workflow: 0,
         workstation: 0,
         assignedTo: 0,
         agency: 0,
         customer: "",
         callNumber: "",
         unitID: "",
         orderID: ""
      }
   }),
   getters: {
      // NOTES : enums from tracksys models
      // assignment status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing]
      canReject: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if (p.assignments.length == 0) return false
            let currA = p.assignments[0]
            let step = p.workflow.steps.find( s => s.id == currA.stepID)
            return step.failStepID > 0 && currA.status == 1
         }
      },
      isFinalizeRunning: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if (p.assignments.length == 0) return false
            let currA = p.assignments[0]
            return currA.status == 6 // finalizing
         }
      },
      isFinished: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if (p) {
               return Object.hasOwn(p, 'finishedAt') && p.finishedAt != ""
            }
            return false
         }
      },
      hasError: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if (p.assignments.length == 0) return false
            let currA = p.assignments[0]
            return currA.status == 4
         }
      },
      hasOwner: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            return p.owner !== undefined
         }
      },
      inProgress: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if (p.assignments.length == 0) return false
            let currA = p.assignments[0]
            return  currA.status == 1 ||  currA.status == 4 ||  currA.status == 6 ||  currA.status == 7
         }
      },
      isWorking: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if (p.assignments.length == 0) return false
            let currA = p.assignments[0]
            return currA.status == 7
         }
      },
      isOwner: state => {
         return (computeID) => {
            if ( state.selectedProjectIdx == -1 ) return false
            if ( state.projects[state.selectedProjectIdx].owner ) {
               return (state.projects[state.selectedProjectIdx].owner.computingID == computeID)
            }
            return false
         }
      },
      totalPages: state => {
         let total = state.totals.active
         if (state.filter == "me") {
            total = state.totals.me
         } else if (state.filter == "unassigned") {
            total = state.totals.unassigned
         } else if (state.filter == "finished") {
            total = state.totals.finished
         }
         return Math.ceil(total/state.pageSize)
      },
      currProject: state => {
         if (state.selectedProjectIdx == -1) {
            return {}
         }
         return state.projects[state.selectedProjectIdx]
      },

      statusText: state => {
         return (pID) => {
            let p = state.projects.find( p => p.id == pID)
            if (p.finishedAt != null) {
               return `Finished at ${p.finishedAt.split("T")[0]}`
            }
            let out = `${p.currentStep.name}: `
            let a = p.assignments.find(a => a.stepID == p.currentStep.id)
            if ( a ) {
               if (a.startedAt != null) {
                  out += " In progress"
               } else {
                  out += " Not started"
               }
            } else {
               out += " Not assigned"
            }
            return out
         }
      },

      percentComplete: state => {
         return (pID) => {
            // NOTE: enum values...
            //    assign status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing]
            //    step types [:start, :end, :error, :normal]
            let p = state.projects.find( p => p.id == pID)
            let nonErrSteps =  p.workflow.steps.filter( s => s.stepType != 2)
            let numSteps = nonErrSteps.length*3 // each non-error step has 3 parts, assigned, in-process and done

            let stepCount = 0
            let stepIDs = []
            p.assignments.forEach( a => {
               if ( stepIDs.includes(a.stepID) ) return

               let step = p.workflow.steps.find( s => s.id == a.stepID )
               if (!step) return

               if (step.stepType != 2 && a.status != 4 && a.status != 5) {
                  // Rejections generally count as a completion as they finish the step. Per team, reject moves to a rescan.
                  // When rescan is done, the workflow proceeds to the step AFTER the one that was rejected.
                  // The exception to this is the last step. If rejected, completing the rescan returns
                  // to that step, not the next. This is the case we need to skip when computing percentage complete.
                  let failStep =  p.workflow.steps.find( s => s.id == step.failStepID )
                  if (a.status == 3 && failStep.nextStepID == a.stepID) return

                  stepIDs.push(a.stepID)           // make sure each step only gets counted once
                  stepCount++                      // if an assignment is here, that is the first count: Assigned
                  if (a.startedAt) stepCount++     // Started
                  if (a.finishedAt) stepCount++    // Finished
               }
            })
            let percent = Math.round((stepCount/numSteps)*100.0)
            return percent+"%"
         }
      }
   },
   actions: {
      setProjects(data) {
         this.totals.me  = data.totalMe
         this.totals.active  = data.totalActive
         this.totals.unassigned  = data.totalUnassigned
         this.totals.finished  = data.totalFinished
         this.pageSize = data.pageSize
         data.currPage = data.page
         this.projects.splice(0, this.projects.length)
         data.projects.forEach( p => this.projects.push(p))
      },
      selectProject(projID) {
         this.selectedProjectIdx = this.projects.findIndex( p => p.id == projID)
      },
      updateProjectData(data) {
         let pIdx = this.projects.findIndex( p => p.id == data.id)
         if (pIdx > -1) {
            this.projects.splice(pIdx, 1, data)
         }
      },
      setPage(pg) {
         let total = this.totals.active
         if (this.filter == "me") {
            total = this.totals.me
         } else if (this.filter == "unassigned") {
            total = this.totals.unassigned
         } else if (this.filter == "finished") {
            total = this.totals.finished
         }
         let maxPg = Math.ceil(total/this.pageSize)
         if (pg > 0 && pg <= maxPg) {
            this.currPage = pg
         }
      },
      setCurrentPage(pg) {
         this.setPage(pg)
         this.getProjects()
      },
      resetSearch() {
         this.search.workflow = 0
         this.search.workstation = 0
         this.search.assignedTo = 0
         this.search.agency = ""
         this.search.customer = ""
         this.search.callNumber = ""
         this.search.unitID = ""
         this.search.orderID = ""
         this.currPage = 1
         this.getProjects()
      },
      getProjects() {
         this.working = true

         let qParam = []
         if (this.search.workflow != 0) {
            qParam.push(`workflow=${this.search.workflow}`)
         }
         if (this.search.workstation != 0) {
            qParam.push(`workstation=${this.search.workstation}`)
         }
         if (this.search.assignedTo != 0) {
            qParam.push(`assigned=${this.search.assignedTo}`)
         }
         if (this.search.agency != "") {
            qParam.push(`agency=${encodeURIComponent(this.search.agency)}`)
         }
         if (this.search.customer != "") {
            qParam.push(`customer=${encodeURIComponent(this.search.customer)}`)
         }
         if (this.search.callNumber != "") {
            qParam.push(`callnum=${encodeURIComponent(this.search.callNumber)}`)
         }
         if (this.search.unitID != "") {
            qParam.push(`unit=${this.search.unitID}`)
         }
         if (this.search.orderID != "") {
            qParam.push(`order=${this.search.orderID}`)
         }

         let q = `/api/projects?page=${this.currPage}&filter=${this.filter}`
         if (qParam.length > 0 ) {
            q += `&${qParam.join("&")}`
         }

         console.log("GET PROJECTS "+q)
         axios.get(q).then(response => {
            this.setProjects(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      async getProject(projectID) {
         this.working = true
         return axios.get(`/api/projects/${projectID}`).then(response => {
            // set a fake list of projects containing only 1 project
            let projects = {total: 1, pageSize: 1, currPage: 1, projects: [response.data]}
            this.setProjects(projects)
            this.selectProject(projectID)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      async setEquipment(data) {
         // data contains { workstationID, captureResolution, resizeResolution, resolutionNote }
         this.working = true
         return axios.post(`/api/projects/${this.currentProject.id}/equipment`, data).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },

      async updateProject(data) {
         // data contains { categoryID, condition, note, ocrHintID, ocrLangage, ocrMasterFiles }
         this.working = true
         return axios.put(`/api/projects/${this.currProject.id}`, data).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      async addNote(data) {
         // data contains { noteTypeID, note, problemIDs}
         this.working = true
         let p = this.currProject
         data.stepID = p.currentStep.id
         return axios.post(`/api/projects/${p.id}/note`, data).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      startStep() {
         this.working = true
         axios.post(`/api/projects/${this.currProject.id}/start`).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      finishStep(durationMins) {
         this.working = true
         axios.post(`/api/projects/${this.currProject.id}/finish`, {durationMins: durationMins} ).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      rejectStep(durationMins) {
         this.working = true
         axios.post(`/api/projects/${this.currProject.id}/reject`, {durationMins: durationMins} ).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      },
      assignProject({projectID, ownerID}) {
         this.working = true
         axios.post(`/api/projects/${projectID}/assign/${ownerID}`).then(response => {
            this.updateProjectData(response.data)
            this.working = false
         }).catch( e => {
            this.system.error = e
            this.working = false
         })
      }
   }
})