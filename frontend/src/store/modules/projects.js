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