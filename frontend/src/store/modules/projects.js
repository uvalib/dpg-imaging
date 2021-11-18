import axios from 'axios'
import { getField, updateField } from 'vuex-map-fields'

const projects = {
   namespaced: true,
   state: {
      projects: [],
      selectedProjectIdx: -1,
      total: 0,
      pageSize: 10,
      currPage: 1,
      candidates: [],
      working: false,
      filter: "active",
      search: {
         workflow: 0,
         workstation: 0,
         assignedTo: 0,
         agency: "",
         customer: "",
         callNumber: "",
      }
   },
   // NOTES : enums from tracksys models
   // assignment status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing]
   getters: {
      getField,
      canReject: state => projIdx => {
         if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
         let p = state.projects[projIdx]
         if (p.assignments.length == 0) return false
         let lastA = p.assignments[p.assignments.length-1]
         let step = p.workflow.steps.find( s => s.id = lastA.stepID)
         return step.failStepID > 0 && lastA.status == 1
      },
      onFinalizeStep: state => projIdx => {
         if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
         let p = state.projects[projIdx]
         if (p.assignments.length == 0) return false
         let lastA = p.assignments[p.assignments.length-1]
         let step = p.workflow.steps.find( s => s.id = lastA.stepID)
         return step.name == "Finalize"
      },
      isFinalizing: state => projIdx => {
         if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
         let p = state.projects[projIdx]
         if (p.assignments.length == 0) return false
         let lastA = p.assignments[p.assignments.length-1]
         return lastA.status == 6
      },
      hasError: state => projIdx => {
         if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
         let p = state.projects[projIdx]
         if (p.assignments.length == 0) return false
         let lastA = p.assignments[p.assignments.length-1]
         return lastA.status == 4
      },
      hasOwner: state => projIdx => {
         if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
         let p = state.projects[projIdx]
         return p.owner.id > 0
      },
      inProgress: state => projIdx => {
         if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
         let p = state.projects[projIdx]
         if (p.assignments.length == 0) return false
         let lastA = p.assignments[p.assignments.length-1]
         return  lastA.status == 1 ||  lastA.status == 4 ||  lastA.status == 6
      },
      isOwner: state => (computeID) => {
         if (state.selectedProjectIdx == -1) return false
         if (state.projects[state.selectedProjectIdx].owner.id == 0) return false
         return (state.projects[state.selectedProjectIdx].owner.computingID == computeID)
      },
      totalPages: state => {
         return Math.ceil(state.total/state.pageSize)
      },
      currProject: state => {
         if (state.selectedProjectIdx == -1) {
            return {}
         }
         return state.projects[state.selectedProjectIdx]
      },
      currentStepName: state => pID => {
         let p = state.projects.find( p => p.id == pID)
         if ( p ) {
            return p.currentStep.name
         }
         return "Unknown"
      },
      statusText: state => pID => {
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
      },
      percentComplete: state => pID => {
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
   },
   mutations: {
      updateField,
      setWorking( state, flag) {
         state.working = flag
      },
      setCandidates( state, data) {
         state.candidates.splice(0, state.candidates.length)
         data.forEach( a => state.candidates.push(a) )
      },
      clearProjects(state) {
         state.projects.splice(0, state.projects.length)
      },
      resetSearch(state) {
         state.search.workflow = 0
         state.search.workstation = 0
         state.search.assignedTo = 0
         state.search.agency = ""
         state.search.customer = ""
         state.search.callNumber = ""
         state.currPage = 1
      },
      setProjects(state, data) {
         state.total = data.total
         state.pageSize = data.pageSize
         data.currPage = data.page
         state.projects.splice(0, state.projects.length)
         data.projects.forEach( p => state.projects.push(p))
      },
      selectProject(state, projID) {
         state.selectedProjectIdx = state.projects.findIndex( p => p.id == projID)
      },
      updateProject(state, data) {
         let pIdx = state.projects.findIndex( p => p.id == data.id)
         if (pIdx > -1) {
            state.projects.splice(pIdx, 1, data)
         }
      },
      setPage(state, pg) {
         let maxPg = Math.ceil(state.total/state.pageSize)
         if (pg > 0 && pg <= maxPg) {
            state.currPage = pg
         }
      }
   },
   actions: {
      setCurrentPage(ctx, pg) {
         ctx.commit("setPage", pg)
         ctx.dispatch("getProjects")
      },
      resetSearch(ctx) {
         ctx.commit("resetSearch")
         ctx.dispatch("getProjects")
      },
      getProjects(ctx) {
         ctx.commit("setWorking", true)

         let qParam = []
         let s = ctx.state.search
         if (s.workflow != 0) {
            qParam.push(`workflow=${s.workflow}`)
         }
         if (s.workstation != 0) {
            qParam.push(`workstation=${s.workstation}`)
         }
         if (s.assignedTo != 0) {
            qParam.push(`assigned=${s.assignedTo}`)
         }
         if (s.agency != "") {
            qParam.push(`agency=${encodeURIComponent(s.agency)}`)
         }
         if (s.customer != "") {
            qParam.push(`customer=${encodeURIComponent(s.customer)}`)
         }
         if (s.callNumber != "") {
            qParam.push(`callnum=${encodeURIComponent(s.callNumber)}`)
         }

         let q = `/api/projects?page=${ctx.state.currPage}&filter=${ctx.state.filter}`
         if (qParam.length > 0 ) {
            q += `&${qParam.join("&")}`
         }

         axios.get(q).then(response => {
            ctx.commit('setProjects', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", true)
         })
      },
      // this is only used when the project details page is loaded without a list of project data
      async getProject(ctx, projectID) {
         ctx.commit("setWorking", true)
         return axios.get(`/api/projects/${projectID}`).then(response => {
            // set a fake list of projects containing only 1 project
            let projects = {total: 1, pageSize: 1, currPage: 1, projects: [response.data]}
            ctx.commit('setProjects', projects)
            ctx.commit('selectProject', projectID)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", false)
         })
      },
      async setEquipment(ctx, data) {
         // data contains { workstationID, captureResolution, resizeResolution, resolutionNote }
         ctx.commit("setWorking", true)
         let p = ctx.getters.currProject
         return axios.post(`/api/projects/${p.id}/equipment`, data).then(response => {
            ctx.commit('updateProject', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", false)
         })
      },
      async updateProject(ctx, data) {
         // data contains { categoryID, condition, note, ocrHintID, ocrLangage, ocrMasterFiles }
         ctx.commit("setWorking", true)
         let p = ctx.getters.currProject
         return axios.put(`/api/projects/${p.id}`, data).then(response => {
            ctx.commit('updateProject', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", false)
         })
      },
      async addNote(ctx, data) {
         // data contains { noteTypeID, note, problemIDs}
         ctx.commit("setWorking", true)
         let p = ctx.getters.currProject
         data.stepID = p.currentStep.id
         return axios.post(`/api/projects/${p.id}/note`, data).then(response => {
            ctx.commit('updateProject', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", false)
         })
      },
      startStep(ctx) {
         ctx.commit("setWorking", true)
         let p = ctx.getters.currProject
         axios.post(`/api/projects/${p.id}/start`).then(response => {
            ctx.commit('updateProject', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", false)
         })
      },
      assignProject(ctx, {projectID, ownerID}) {
         ctx.commit("setWorking", true)
         axios.post(`/api/projects/${projectID}/assign/${ownerID}`).then(response => {
            ctx.commit('updateProject', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setError", e, {root: true})
            ctx.commit("setWorking", false)
         })
      }
   }
}
export default projects