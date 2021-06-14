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
      viewMode: "list",
      rangeStartIdx: -1,
      rangeEndIdx: -1,
      editMode: "",
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
      filenameSort(state) {
         state.masterFiles.sort( (a,b) => {
            if (a.fileName < b.fileName) return -1
            if (a.fileName > b.fileName) return 1
            return 0
         })
      },
      setError(state, msg) {
         state.error = true
         state.errorMessage = msg
      },
      clearError(state) {
         state.error = false
         state.errorMessage = ""
      },
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
               mf.status = d.status
               ctx.masterFiles.splice(mfIdx,1, mf)
            }
         })
      },

      removeMasterFile(ctx, fn) {
         let mfIdx = ctx.masterFiles.findIndex( mf => mf.fileName == fn)
         if (mfIdx > -1) {
            ctx.masterFiles.splice(mfIdx,1)
         }
      }
   },
   actions: {
      getUnits(ctx) {
         ctx.commit("setLoading", true)
         axios.get("/api/units").then(response => {
            ctx.commit('setUnits', response.data)
            ctx.commit("setLoading", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setLoading", false)
         })
      },
      async getUnitDetails(ctx, unit) {
         if (ctx.state.currUnit == unit && ctx.state.masterFiles.length > 0) return

         ctx.commit("setLoading", true)
         return axios.get(`/api/units/${unit}`).then(response => {
            ctx.commit('setMasterFiles', {unit: unit, masterFiles: response.data.masterFiles})
            ctx.commit("setLoading", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setLoading", false)
         })
      },

      updatePageNumbers(ctx, startPage) {
         ctx.commit("setUpdating", true)
         let data = []
         let page = parseInt(startPage,10)
         for (let i=ctx.state.rangeStartIdx; i<=ctx.state.rangeEndIdx; i++) {
            let mf = ctx.state.masterFiles[i]
            data.push( {file: mf.path, title: ""+page, description: mf.description, status: mf.status})
            page+=1
         }
         axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then(() => {
            ctx.commit('updateMetadata', data )
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      async updateMetadata(ctx, {file, title, description, status}) {
         ctx.commit("setUpdating", true)
         let data = [{file: file, title: title, description: description, status: status}]
         return axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then(() => {
            ctx.commit('updateMetadata', data )
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      async setTag(ctx, {file, tag}) {
         ctx.commit("setUpdating", true)
         let mf = ctx.state.masterFiles.find( mf => mf.path == file)
         let status = tag
         if (tag == "none") {
            status = ""
         }
         let data = [{file: file, title: mf.title, description: mf.description, status: status}]
         return axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then(() => {
            ctx.commit('updateMetadata', data )
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      deleteMasterFile(ctx, mf) {
         ctx.commit("setUpdating", true)
         axios.delete(`/api/units/${ctx.state.currUnit}/${mf}`).then(() => {
            ctx.commit('removeMasterFile', mf )
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      renameAll( ctx ) {
         ctx.commit("setUpdating", true)
         let data = []
         ctx.state.masterFiles.forEach( (mf,idx) => {
            let mfPage = parseInt(mf.fileName.toLowerCase().replace(".tif").split("_")[1],10)
            if (mfPage !=  idx+1) {
               let newPg = `${idx+1}`
               newPg = newPg.padStart(4,'0')
               let newFN = `${ctx.state.currUnit}_${newPg}.tif`
               data.push({original: mf.path, new: newFN })
            }

         })
         axios.post(`/api/units/${ctx.state.currUnit}/rename`, data).then(() => {
            window.location.reload()
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      }
   },
   modules: {
   }
})
