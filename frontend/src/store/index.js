import { createStore } from 'vuex'
import axios from 'axios'
import { getField, updateField } from 'vuex-map-fields'

export default createStore({
   state: {
      loading: false,
      updating: false,
      error: false,
      errorMessage: "",
      units: [],
      currUnit: "",
      masterFiles: [],
      pageMasterFiles: [],
      viewMode: "list",
      rangeStartIdx: -1,
      rangeEndIdx: -1,
      editMode: "",
      callNumber: "",
      title: "",
      projectURL: "",
      problems: [],
      component: {
         valid: false,
         title: "",
         label: "",
         desc: "",
         date: "",
         type: "",
      },
      pageSize: 20,
      currPage: 0,
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
      },
      pageStartIdx: state => {
         return state.currPage*state.pageSize
      },
      totalPages: state => {
         return Math.ceil(state.masterFiles.length / state.pageSize)
      },
      totalFiles: state => {
         return state.masterFiles.length
      }
   },
   mutations: {
      updateField,
      selectAll(state) {
         state.rangeStartIdx = 0
         state.rangeEndIdx = state.masterFiles.length - 1
      },
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
      setUnitMetadata(state, data) {
         state.callNumber = data.callNumber
         state.title = data.title
         state.projectURL = data.projectURL
      },
      setUnitProblems(state, problems) {
         state.problems.splice(0, state.problems.length)
         problems.forEach( p => state.problems.push(p) )
      },
      setPage(state, startPageNum) {
         state.pageMasterFiles.splice(0, state.pageMasterFiles.length)
         for (let idx = 0; idx<state.pageSize; idx++) {
            let mfIdx = startPageNum*state.pageSize + idx
            if (mfIdx < state.masterFiles.length ) {
               state.pageMasterFiles.push( state.masterFiles[mfIdx])
            }
         }
         state.currPage = startPageNum
      },
      setMasterFiles(ctx, {unit, masterFiles}) {
         ctx.currUnit = unit
         ctx.masterFiles.splice(0, ctx.masterFiles.length)
         masterFiles.forEach( mf =>{
            mf.error = ""
            ctx.masterFiles.push(mf)
         })
      },
      updateMasterFileMetadata( ctx, data) {
         // data is an array of {file, title, description}
         data.forEach( d => {
            let mfIdx = ctx.masterFiles.findIndex( mf => mf.path == d.file)
            if (mfIdx > -1) {
               let mf = ctx.masterFiles[mfIdx]
               mf.title = d.title
               mf.description = d.description
               mf.status = d.status
               mf.componentID = d.componentID
               mf.error = ""
               ctx.masterFiles.splice(mfIdx,1, mf)
            }
         })
      },

      setMasterFileProblems(ctx, data) {
         data.forEach( p => {
            let mfIdx = ctx.masterFiles.findIndex( mf => mf.fileName == p.file)
            if (mfIdx > -1) {
               let mf = ctx.masterFiles[mfIdx]
               mf.error = p.problem
               ctx.masterFiles.splice(mfIdx,1, mf)
            }
         })
      },

      removeMasterFile(ctx, fn) {
         let mfIdx = ctx.masterFiles.findIndex( mf => mf.fileName == fn)
         if (mfIdx > -1) {
            ctx.masterFiles.splice(mfIdx,1)
         }
      },

      setComponentInfo(ctx, data) {
         ctx.component.title = data.title
         ctx.component.label = data.label
         ctx.component.desc = data.description
         ctx.component.date = data.date
         ctx.component.type = data.type
         ctx.component.valid = true
      },
      clearComponent(ctx) {
         ctx.component.valid = false
         ctx.component.title = ""
         ctx.component.label = ""
         ctx.component.desc = ""
         ctx.component.date = ""
         ctx.component.type = ""
      },
      clearUnitDetails(ctx) {
         ctx.masterFiles.splice(0, ctx.masterFiles.length)
         ctx.rangeStartIdx = -1
         ctx.rangeEndIdx = -1
         ctx.editMode = ""
         ctx.callNumber = "Unknown"
         ctx.title = "Unknown"
         ctx.projectURL = ""
         ctx.problems.splice(0, ctx.problems.length)
         ctx.currPage = 0
      }
   },
   actions: {
      getUnits(ctx) {
         ctx.commit("setLoading", true)
         ctx.commit("clearUnitDetails")
         axios.get("/api/units").then(response => {
            ctx.commit('setUnits', response.data)
            ctx.commit("setLoading", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setLoading", false)
         })
      },
      async getUnitDetails(ctx, unit) {
         // dont try to reload a unit if the data is already present - unless an update is in process
         if (ctx.state.currUnit == unit && ctx.state.masterFiles.length > 0 && ctx.state.updating == false) return

         ctx.commit("setLoading", true)
         ctx.commit("clearUnitDetails")
         return axios.get(`/api/units/${unit}`).then(response => {
            ctx.commit('setMasterFiles', {unit: unit, masterFiles: response.data.masterFiles})
            ctx.commit('setPage', 0)
            ctx.commit('setUnitMetadata', response.data.metadata)
            ctx.commit('setUnitProblems', response.data.problems)
            ctx.commit("setLoading", false)
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setLoading", false)
            ctx.commit("setUpdating", false)
         })
      },

      updatePageNumbers(ctx, startPage) {
         ctx.commit("setUpdating", true)
         let data = []
         let page = parseInt(startPage,10)
         for (let i=ctx.state.rangeStartIdx; i<=ctx.state.rangeEndIdx; i++) {
            let mf = ctx.state.masterFiles[i]
            data.push( {file: mf.path, title: ""+page, description: mf.description, status: mf.status, componentID: mf.componentID})
            page+=1
         }
         axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then( resp => {
            ctx.commit('updateMasterFileMetadata', data )
            ctx.commit("setUpdating", false)
            if (resp.data.success == false) {
               ctx.commit('setError', "Some images were not renumbered")
               ctx.commit('setMasterFileProblems', resp.data.problems)
            }
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      async componentLink(ctx, componentID) {
         ctx.commit("setUpdating", true)
         let data = []
         for (let i=ctx.state.rangeStartIdx; i<=ctx.state.rangeEndIdx; i++) {
            let mf = ctx.state.masterFiles[i]
            data.push( {file: mf.path, title: mf.title, description: mf.description, status: mf.status, componentID: componentID})
         }
         return axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then( resp => {
            ctx.commit('updateMasterFileMetadata', data )
            ctx.commit("setUpdating", false)
            if (resp.data.success == false) {
               ctx.commit('setError', "Some images could not be linked with component "+componentID)
               ctx.commit('setMasterFileProblems', resp.data.problems)
            }
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      async updateMasterFileMetadata(ctx, {file, title, description, status, componentID}) {
         ctx.commit("setUpdating", true)
         let data = [{file: file, title: title, description: description, status: status, componentID: componentID}]
         return axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then( resp => {
            ctx.commit('updateMasterFileMetadata', data )
            ctx.commit("setUpdating", false)
            if (resp.data.success == false) {
               ctx.commit('setError', "Unable to update image metadata")
               ctx.commit('setMasterFileProblems', resp.data.problems)
            }
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
         return axios.post(`/api/units/${ctx.state.currUnit}/update`, data).then( resp => {
            ctx.commit('updateMasterFileMetadata', data )
            ctx.commit("setUpdating", false)
            if (resp.data.success == false) {
               ctx.commit('setError', "Unable to set tag on image")
               ctx.commit('setMasterFileProblems', resp.data.problems)
            }
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

      rotateImage(ctx, mf) {
         ctx.commit("setUpdating", true)
         axios.post(`/api/units/${ctx.state.currUnit}/${mf}/rotate`, {}).then(() => {
            setTimeout( ()=>{
               ctx.commit("setUpdating", false)
               window.location.reload()
            }, 250)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },


      renameAll( ctx ) {
         ctx.commit("setUpdating", true)
         let data = []
         let pStartIdx = ctx.state.currPage*ctx.state.pageSize
         let pEndIdx = pStartIdx + ctx.state.pageSize-1
         let numRegex = /^\d{4}$/
         ctx.state.masterFiles.forEach( (mf,idx) => {
            let originalFN = mf.fileName.toLowerCase()
            let originalPath = mf.path

            // if this MF is in the range of the acitve page, grab data for it from
            // the pageMasterFiles array instead as it may have been reordered with drag/drop
            if (idx >= pStartIdx && idx <= pEndIdx) {
               let pageMf = ctx.state.pageMasterFiles[idx-pStartIdx]
               originalFN = pageMf.fileName.toLowerCase()
               originalPath = pageMf.path
            }

            originalFN = originalFN.replace(".tif","")
            let mfParts = originalFN.split("_")
            let mfPage = parseInt(mfParts[1],10)
            let badName = false
            if ( !mfParts[1].match(numRegex) ) {
               badName = true
            }

            // detect either sequence issues, bad name, or unit prefix issues
            if ( badName || mfPage !=  idx+1 || mfParts[0] != ctx.state.currUnit) {
               let newPg = `${idx+1}`
               newPg = newPg.padStart(4,'0')
               let newFN = `${ctx.state.currUnit}_${newPg}.tif`
               data.push({original: originalPath, new: newFN })
            }

         })
         axios.post(`/api/units/${ctx.state.currUnit}/rename`, data).then(() => {
            ctx.dispatch('getUnitDetails', ctx.state.currUnit)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      },

      lookupComponentID(ctx, componentID) {
         ctx.commit("setUpdating", true)
         axios.get(`/api/components/${componentID}`).then(response => {
            ctx.commit('setComponentInfo', response.data)
            ctx.commit("setUpdating", false)
         }).catch( e => {
            ctx.commit('setError', e)
            ctx.commit("setUpdating", false)
         })
      }
   },
   modules: {
   }
})
