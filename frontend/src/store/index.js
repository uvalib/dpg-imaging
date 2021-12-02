import { createStore } from 'vuex'
import axios from 'axios'
import { getField, updateField } from 'vuex-map-fields'

import projects from './modules/projects'
import user from './modules/user'
import units from './modules/units'

export default createStore({
   state: {
      version: "unknown",
      staffMembers: [],
      agencies: [],
      workstations: [],
      workflows: [],
      categories: [],
      containerTypes: [],
      ocrHints: [],
      ocrLanguageHints: [],
      problemTypes: [],
      adminURL: "",
      qaDir: "",
      scanDir: "",
      loading: false,
      updating: false,
      error: false,
      errorMessage: "",
   },
   getters: {
      getField,
   },
   mutations: {
      updateField,
      setConfig(state, data) {
         state.adminURL = data.tracksysURL
         state.qaDir =  data.qaImageDir
         state.scanDir =  data.scanDir
         state.staffMembers.splice(0, state.staffMembers.length)
         data.staff.forEach( s=> state.staffMembers.push(s))
         state.workflows.splice(0, state.workflows.length)
         data.workflows.forEach( w=> state.workflows.push(w))
         state.workstations.splice(0, state.workstations.length)
         data.workstations.forEach( w=> state.workstations.push(w))
         state.categories.splice(0, state.categories.length)
         data.categories.forEach( w=> state.categories.push(w))
         state.ocrHints.splice(0, state.ocrHints.length)
         data.ocrHints.forEach( w=> state.ocrHints.push(w))
         state.ocrLanguageHints.splice(0, state.ocrLanguageHints.length)
         data.ocrLanguageHints.forEach( w=> state.ocrLanguageHints.push(w))
         state.problemTypes.splice(0, state.problemTypes.length)
         data.problems.forEach( p=> state.problemTypes.push(p))
         state.containerTypes.splice(0, state.containerTypes.length)
         data.containerTypes.forEach( w=> state.containerTypes.push(w))
         state.agencies.splice(0, state.agencies.length)
         data.agencies.forEach( w=> state.agencies.push(w))
      },
      setVersion(state, data) {
         state.version = `${data.version}-${data.build}`
      },
      setError(state, err) {
         state.error = true
         state.updating = false
         state.loading = false
         if (err.response && err.response.data) {
            state.errorMessage = err.response.data
         } else {
            state.errorMessage = err
         }
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
   },
   actions: {
      getVersion(ctx) {
         axios.get("/version").then(response => {
            ctx.commit('setVersion', response.data)
         }).catch( e => {
            ctx.commit("setError", e)
            ctx.commit("setLoading", false)
         })
      },
      getConfig(ctx) {
         ctx.commit("setLoading", true)
         axios.get("/config").then(response => {
            ctx.commit('setConfig', response.data)
            ctx.commit("setLoading", false)
         }).catch( e => {
            ctx.commit("setError", e)
            ctx.commit("setLoading", false)
         })
      },
   },
   modules: {
      projects: projects,
      user: user,
      units: units,
   }
})
