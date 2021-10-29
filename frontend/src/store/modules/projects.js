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
         let p = state.projects.find( p => p.id == pID)
         num_steps = p.workflow.length*3 // each step has 3 parts, assigned, in-process and done

         // curr_step = 0
         // step_ids = []
         // self.assignments.joins(:step).each do |a|
         //    if !a.step.error? && !step_ids.include?(a.step.id) && !a.reassigned? && !a.error?
         //       # Rejections generally count as a completion as they finish the step. Per team, reject moves to a rescan.
         //       # When rescan is done, the workflow proceeds to the step AFTER the one that was rejected.
         //       # The exception to this is the last step. If rejected, completing the rescan returns
         //       # to that step, not the next. This is the case we need to skip when computing percentage complete.
         //       next if a.rejected? && a.step.fail_step.next_step == a.step

         //       step_ids << a.step.id
         //       curr_step +=1  # if an assignment is here, that is the first count: Assigned
         //       curr_step +=1 if !a.started_at.nil?    # Started
         //       curr_step +=1 if !a.finished_at.nil?   # Finished
         //    end
         // end
         // percent =  (curr_step.to_f/num_steps.to_f*100).to_i
         // if finished? && percent != 100
         //    Rails.logger.error("Project #{self.id} is finished, but percentage reported as #{percent}")
         //    percent = 100
         // end
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