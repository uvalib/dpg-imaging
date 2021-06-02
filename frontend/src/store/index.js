import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'

Vue.use(Vuex)

export default new Vuex.Store({
   state: {
      working: false,
      error: false,
      errorMessage: "",
      units: [],
      currUnit: "",
      masterFiles: []
   },
   getters: {
      pageInfoURLs: state => {
         let out = []
         state.masterFiles.forEach( mf => out.push(mf.infoURL) )
         return out
      }
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
      },
      setMasterFiles(ctx, {unit, masterFiles}) {
         ctx.currUnit = unit
         ctx.masterFiles.splice(0, ctx.masterFiles.length)
         masterFiles.forEach( mf =>{
            ctx.masterFiles.push(mf)
         })
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
      },
      async getMasterFiles(ctx, unit) {
         if (ctx.currUnit == unit && ctx.masterFiles.length > 0) return

         ctx.commit("setWorking", true)
         return axios.get(`/api/units/${unit}`).then(response => {
            ctx.commit('setMasterFiles', {unit: unit, masterFiles: response.data})
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
