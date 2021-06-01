import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

export default new Vuex.Store({
   state: {
      working: false,
      error: false,
      errorMessage: "",
      units: []
   },
   mutations: {
      setWorking(state, flag) {
         state.working = flag
         if (flag == true ) {
            state.error = false
         }
      },
      setUnits(state, data) {
         state.units.splice(0, state.units.length)
         data.sort()
         data.forEach( u => {
            state.units.push( u )
         })
      },
      setFailed(state, err) {
         state.error = true
         state.errorMessage = err
      }
   },
   actions: {
      getUnits(ctx) {
         ctx.commit("setWorking", true)
         axios.get("/api/units").then(response => {
            ctx.commit('setUnits', response.data)
            ctx.commit("setWorking", false)
         }).catch( e => {
            ctx.commit('setFailed', e)
            ctx.commit("setWorking", false)
         })
      }
   },
   modules: {
   }
})
