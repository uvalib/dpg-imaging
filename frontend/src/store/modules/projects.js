import axios from 'axios'

const projects = {
   namespaced: true,
   state: {
      projects: [],
      total: 0,
      pageSize: 10,
      currPage: 1,
   },
   getters: {
      totalPages: state => {
         return Math.ceil(state.total/state.pageSize)
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
         console.log("PROJECT "+pID+" PERCENT: "+percent)
         return percent+"%"
      }
   },
   mutations: {
      clearProjects(state) {
         state.projects.splice(0, state.projects.length)
      },
      setProjects(state, data) {
         state.total = data.total
         state.pageSize = data.pageSize
         data.currPage = data.page
         state.projects.splice(0, state.projects.length)
         data.projects.forEach( p => state.projects.push(p))
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
      getProjects(ctx) {
         ctx.commit("setLoading", true, {root: true})
         axios.get(`/api/projects?page=${ctx.state.currPage}`).then(response => {
            ctx.commit('setProjects', response.data)
            ctx.commit("setLoading", false, {root: true})
         }).catch( e => {
            ctx.commit("handleError", e, {root: true})
         })
      },
   }
}
export default projects