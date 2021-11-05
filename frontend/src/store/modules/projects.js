import axios from 'axios'

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
   },
   getters: {
      totalPages: state => {
         return Math.ceil(state.total/state.pageSize)
      },
      currProject: state => {
         if (state.selectedProjectIdx == -1) {
            return {}
         }
         return state.projects[state.selectedProjectIdx]
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
      setProjects(state, data) {
         state.total = data.total
         state.pageSize = data.pageSize
         data.currPage = data.page
         state.projects.splice(0, state.projects.length)
         data.projects.forEach( p => state.projects.push(p))
      },
      selectProject(state, projID) {
         state.selectedProjectIdx = state.projects.findIndex( p => p.id = projID)
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
      getCandidates(ctx, projecdID) {
         ctx.commit("setWorking", true)
         axios.get(`/api/projects/${projecdID}/candidates`).then(response => {
            ctx.commit('setCandidates', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit("setWorking", false)
            ctx.commit("setError", e, {root: true})
         })
      },
      getProjects(ctx) {
         ctx.commit("setLoading", true, {root: true})
         axios.get(`/api/projects?page=${ctx.state.currPage}`).then(response => {
            ctx.commit('setProjects', response.data)
            ctx.commit("setLoading", false, {root: true})
         }).catch( e => {
            ctx.commit("setLoading", false, {root: true})
            ctx.commit("setError", e, {root: true})
         })
      },
      // this is only used when the project details page is loaded without a list of project data
      getProject(ctx, projectID) {
         ctx.commit("setLoading", true, {root: true})
         axios.get(`/api/projects/${projectID}`).then(response => {
            // set a fake list of projects containing only 1 project
            let projects = {total: 1, pageSize: 1, currPage: 1, projects: [response.data]}
            ctx.commit('setProjects', projects)
            ctx.commit('selectProject', projectID)
            ctx.commit("setLoading", false, {root: true})
         }).catch( e => {
            ctx.commit("setLoading", false, {root: true})
            ctx.commit("setError", e, {root: true})
         })
      },
      assignProject(ctx, {projectID, ownerID}) {
         ctx.commit("setLoading", true, {root: true})
         axios.post(`/api/projects/${projectID}/assign/${ownerID}`).then(response => {
            ctx.commit('updateProject', response.data)
            ctx.commit("setLoading", false, {root: true})
         }).catch( e => {
            ctx.commit("setLoading", false, {root: true})
            ctx.commit("setError", e, {root: true})
         })
      }
   }
}
export default projects