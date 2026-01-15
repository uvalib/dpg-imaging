import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useSearchStore = defineStore('search', {
   state: () => ({
      projects: [],
      totals: {
         me: 0,
         active: 0,
         errors: 0,
         unassigned: 0,
         finished: 0
      },
      pageSize: 20,
      currPage: 1,
      working: false,
      filter: "active",
      search: {
         workflow: 0,
         workstation: 0,
         step: "any",
         assignedTo: 0,
         agency: 0,
         customer: 0,
         callNumber: "",
         unitID: "",
         orderID: ""
      },
      lastSearchURL: "/",
   }),
   getters: {
      dueDate: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return "Unknown"
            let p = state.projects[projIdx]
            return  p.unit.order.dateDue.split("T")[0]
         }
      },
      hasError: state => {
         return (projIdx) => {
            if (projIdx < 0 || projIdx > state.projects.length-1 ) return false
            let p = state.projects[projIdx]
            if ( p.assignments == null || p.assignments === undefined) return false
            if (p.assignments.length == 0) return false
            let currA = p.assignments[0]
            if (currA.status == 4) return true
            if (currA.step.stepType == 2) return true


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
         } else if (state.filter == "errors") {
            total = state.totals.errors
         }

         return Math.ceil(total/state.pageSize)
      },

      statusText: state => {
         return (pID) => {
            let p = state.projects.find( p => p.id == pID)
            if (p.finishedAt != null) {
               return `Finished at ${p.finishedAt.split("T")[0]}`
            }
            let out = `${p.currentStep.name}: `
            let a = p.assignments.find(a => a.step.id == p.currentStep.id)
            if ( a ) {
               if (a.finishedAt != null) {
                  if (a.status == 4) {
                     out += " Failed"
                  } else {
                     out += " Finished"
                  }
               }
               if (a.startedAt != null) {
                  if (a.status == 4) {
                     out += " Failed"
                  } else {
                     out += " In progress"
                  }
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
            const system = useSystemStore()
            let p = state.projects.find( p => p.id == pID)
            let tgtWorkflow = system.workflows.find( w => w.id == p.workflow.id)
            if (tgtWorkflow == null ) {
               console.log("no workflow found for id "+ p.workflow.id)
               console.log(system.workflows)
               return
            }
            let nonErrSteps =  tgtWorkflow.steps.filter( s => s.stepType != 2)
            let numSteps = nonErrSteps.length*3 // each non-error step has 3 parts, assigned, in-process and done

            let stepCount = 0
            let stepIDs = []
            p.assignments.forEach( a => {
               // only check steps once, don't care about error steps (stepType == 2)  and don't count reassigns or errors
               if ( stepIDs.includes(a.step.id) ) return
               if (a.step.stepType == 2 ) return
               if ( a.status == 5 || a.status == 4) return

               stepIDs.push(a.step.id)          // make sure each step only gets counted once
               stepCount++                      // if an assignment is here, that is the first count: Assigned
               if (a.startedAt) stepCount++     // Started
               if (a.finishedAt) stepCount++    // Finished
            })

            let percent = Math.round((stepCount/numSteps)*100.0)
            percent = Math.min(100, percent)
            return percent+"%"
         }
      }
   },
   actions: {
      setProjects(data) {
         const system = useSystemStore()
         this.totals.me  = data.totalMe
         this.totals.active  = data.totalActive
         this.totals.unassigned  = data.totalUnassigned
         this.totals.finished  = data.totalFinished
         this.totals.errors  = data.totalError
         this.pageSize = data.pageSize
         data.currPage = data.page
         this.projects = []
         data.projects.forEach( p => {
            p.owner = system.getStaffMember(p.ownerID)
            delete p.ownerID
            this.projects.push(p)
         })
      },

      updateOwner(projectID, newOwner) {
         let p = this.projects.find( proj => proj.id = projectID)
         if ( p ) {
            p.owner = newOwner
         }
      },

      changeFilter( newFilter ) {
         this.pageSize =  20
         this.currPage = 1
         this.filter = newFilter
      },

      setPage(pg) {
         let total = this.totals.active
         if (this.filter == "me") {
            total = this.totals.me
         } else if (this.filter == "unassigned") {
            total = this.totals.unassigned
         } else if (this.filter == "finished") {
            total = this.totals.finished
         } else if (this.filter == "errors") {
            total = this.totals.errors
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
         this.search.step = "any"
         this.search.assignedTo = 0
         this.search.agency = 0
         this.search.customer = 0
         this.search.callNumber = ""
         this.search.unitID = ""
         this.search.orderID = ""
         this.currPage = 1
         this.lastSearchURL = "/"
         this.filter = "active"
         this.getProjects()
      },
      getProjects() {
         this.working = true

         let qParam = []
         if (this.search.workflow != 0) {
            qParam.push(`workflow=${this.search.workflow}`)
         }
         if (this.search.step != "any") {
            qParam.push(`step=${this.search.step}`)
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

         axios.get(q).then(response => {
            this.setProjects(response.data)
            this.working = false
         }).catch( e => {
            const system = useSystemStore()
            system.error = e
            this.working = false
         })
      },
   }
})