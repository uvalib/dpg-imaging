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
   },
   mutations: {
      clearProjects(ctx) {
         ctx.projects.splice(0, ctx.projects.length)
      },
      setProjects(state, data) {
         state.total = data.total
         state.pageSize = data.pageSize
         data.currPage = data.page
         data.projects.forEach( p => state.projects.push(p))
      },
   },
   actions: {
      getProjects(ctx) {
         ctx.commit("setLoading", true, {root: true})
         ctx.commit("clearProjects")
         ctx.commit("clearUnitDetails", null, {root: true})
         axios.get(`/api/projects?page=${ctx.state.projectsPage}`).then(response => {
            ctx.commit('setProjects', response.data)
            ctx.commit("setLoading", false, {root: true})
         }).catch( e => {
            ctx.commit("handleError", e, {root: true})
         })
      },
   }
}
export default projects