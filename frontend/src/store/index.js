import Vue from 'vue'
import Vuex from 'vuex'
import axios from 'axios'
import { getField, updateField } from 'vuex-map-fields'

Vue.use(Vuex)

export default new Vuex.Store({
   state: {
      loading: false,
      updating: false,
      error: false,
      errorMessage: "",
      units: [],
      currUnit: "",
      masterFiles: [],
      viewMode: "list"
   },
   getters: {
      getField,
      pageInfoURLs: state => {
         let out = []
         state.masterFiles.forEach( mf => out.push(mf.infoURL) )
         return out
      },
      masterFileInfo: state => page => {
         return state.masterFiles[page]
      }
   },
   mutations: {
      updateField,
      setLoading(state, flag) {
         state.loading = flag
         if (flag == true ) {
            state.error = false
         }
      },
      setUpdating(state, flag) {
         state.updating = flag
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
      },
      updateMetadata( ctx, data) {
         // data is an array of {file, title, description}
         data.forEach( d => {
            let mfIdx = ctx.masterFiles.findIndex( mf => mf.path == d.file)
            if (mfIdx > -1) {
               let mf = ctx.masterFiles[mfIdx]
               mf.title = d.title
               mf.description = d.description
               ctx.masterFiles.splice(mfIdx,1, mf)
            }
         })
      }
   },
   actions: {
      getUnits(ctx) {
         ctx.commit("setLoading", true)
         axios.get("/api/units").then(response => {
            ctx.commit('setUnits', response.data)
            ctx.commit("setLoading", false)
         }).catch( e => {
            ctx.commit('setFailed', e)
            ctx.commit("setLoading", false)
         })
      },
      async getMasterFiles(ctx, unit) {
         if (ctx.state.currUnit == unit && ctx.state.masterFiles.length > 0) return

         ctx.commit("setLoading", true)
         return axios.get(`/api/units/${unit}`).then(response => {
            ctx.commit('setMasterFiles', {unit: unit, masterFiles: response.data})
            ctx.commit("setLoading", false)
         }).catch( e => {
            ctx.commit('setFailed', e)
            ctx.commit("setLoading", false)
         })
      },

      updateMetadata(ctx, {file, title, description}) {
         ctx.commit("setUpdating", true)
         let data = [{file: file, title: title, description: description}]
         axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then(() => {
            ctx.commit('updateMetadata', data )
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setFailed', e)
            ctx.commit("setUpdating", false)
            // TODO show error!!
         })
      }
   },
   modules: {
   }
})
