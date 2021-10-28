import axios from 'axios'

const old = {
   namespaced: true,
   state: {
      units: [],
   },
   getters: {
   },
   mutations: {
      setUnits(state, data) {
         state.units.splice(0, state.units.length)
         data.sort()
         data.forEach( u => {
            state.units.push( u )
         })
      },
   },
   actions: {
      getUnits(ctx) {
         ctx.commit("setLoading", true, {root: true})
         ctx.commit("clearUnitDetails", null, {root: true})
         axios.get("/api/units").then(response => {
            ctx.commit('setUnits', response.data)
            ctx.commit("setLoading", false, {root: true})
         }).catch( e => {
            ctx.commit("handleError", e, {root: true})
         })
      },
   }
}
export default old