import axios from 'axios'
import { defineStore } from 'pinia'
import { useSystemStore } from './system'

export const useUnitStore = defineStore('unit', {
   state: () => ({
      currUnit: "",
      masterFiles: [],
      pageMasterFiles: [],
      viewMode: "list",
      rangeStartIdx: -1,
      rangeEndIdx: -1,
      editMode: "",
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
      currPage: 1,
   }),
   getters: {
      pageInfoURLs: state => {
         let out = []
         state.masterFiles.forEach( mf => out.push(mf.infoURL) )
         return out
      },
      masterFileInfo: state => {
         return (page) => {
            return state.masterFiles[page]
         }
      },
      pageStartIdx: state => {
         return (state.currPage-1)*state.pageSize
      },
      totalPages: state => {
         return Math.ceil(state.masterFiles.length / state.pageSize)
      },
      totalFiles: state => {
         return state.masterFiles.length
      }
   },
   actions: {
      selectAll() {
         this.rangeStartIdx = 0
         this.rangeEndIdx = this.masterFiles.length - 1
      },
      setPage(startPageNum) {
         this.pageMasterFiles.splice(0, this.pageMasterFiles.length)
         for (let idx = 0; idx<this.pageSize; idx++) {
            let mfIdx = (startPageNum-1)*this.pageSize + idx
            if (mfIdx < this.masterFiles.length ) {
               this.pageMasterFiles.push( this.masterFiles[mfIdx])
            }
         }
         this.currPage = startPageNum
      },

      applyMasterFileMetadataUpdate(data) {
         // data is an array of {file, title, description}
         data.forEach( d => {
            let mfIdx = this.masterFiles.findIndex( mf => mf.path == d.file)
            if (mfIdx > -1) {
               let mf = this.masterFiles[mfIdx]
               mf.title = d.title
               mf.description = d.description
               mf.status = d.status
               mf.componentID = d.componentID
               mf.error = ""
               this.masterFiles.splice(mfIdx,1, mf)
            }
         })
      },

      setMasterFileProblems(data) {
         data.forEach( p => {
            let mfIdx = this.masterFiles.findIndex( mf => mf.fileName == p.file)
            if (mfIdx > -1) {
               let mf = this.masterFiles[mfIdx]
               mf.error = p.problem
               this.masterFiles.splice(mfIdx,1, mf)
            }
         })
      },

      removeMasterFile(fn) {
         let mfIdx = this.masterFiles.findIndex( mf => mf.fileName == fn)
         if (mfIdx > -1) {
            this.masterFiles.splice(mfIdx,1)
         }
      },

      setComponentInfo(data) {
         this.component.title = data.title
         this.component.label = data.label
         this.component.desc = data.description
         this.component.date = data.date
         this.component.type = data.componentType.name
         this.component.valid = true
      },
      clearComponent() {
         this.component.valid = false
         this.component.title = ""
         this.component.label = ""
         this.component.desc = ""
         this.component.date = ""
         this.component.type = ""
      },
      clearUnitDetails() {
         this.masterFiles.splice(0, this.masterFiles.length)
         this.rangeStartIdx = -1
         this.rangeEndIdx = -1
         this.editMode = ""
         this.problems.splice(0, this.problems.length)
         this.currPage = 1
         this.viewMode = "list"
      },

      setPageSize(newSize) {
         this.pageSize = newSize
         this.setPage(1)
      },

      async getUnitMasterFiles(unit) {
         // dont try to reload a unit if the data is already present - unless an update is in process
         if (this.currUnit == unit && this.masterFiles.length > 0 && this.updating == false) return

         const system = useSystemStore()
         system.loading = true
         this.clearUnitDetails()
         return axios.get(`/api/units/${unit}/masterfiles`).then(response => {
            this.currUnit = unit
            this.masterFiles.splice(0, this.masterFiles.length)
            response.data.masterFiles.forEach( mf =>{
               mf.error = ""
               this.masterFiles.push(mf)
            })

            this.setPage(1)
            this.problems = response.data.problems
            system.loading = false
            system.updating = false
         }).catch( e => {
            system.error = e
            system.loading = false
            system.updating = false
         })
      },

      updatePageNumbers( {start, verso} ) {
         const system = useSystemStore()
         system.updating = true
         let data = []
         let page = parseInt(start,10)
         let pageCnt = 0
         for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
            let mf = this.masterFiles[i]

            // numbering is expected to begin on front, so verso is odd numbers
            let title = ""+page
            if (verso == false) {
               if (pageCnt%2 != 0) {
                  title = page+" verso"
                  page+=1
               }
            } else {
               page+=1
            }
            data.push( {file: mf.path, title: title, description: mf.description, status: mf.status, componentID: mf.componentID})
            pageCnt+=1
         }

         axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            this.applyMasterFileMetadataUpdate(data)
            system.updating = false
            if (resp.data.success == false) {
               system.setError("Some images were not renumbered")
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.error = e
         })
      },

      async componentLink(componentID) {
         const system = useSystemStore()
         system.updating = true
         let data = []
         for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
            let mf = this.masterFiles[i]
            data.push( {file: mf.path, title: mf.title.trim(), description: mf.description.trim(), status: mf.status, componentID: componentID})
         }
         return axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            this.applyMasterFileMetadataUpdate(data)
            system.updating = false
            if (resp.data.success == false) {
               system.setError("Some images could not be linked with component "+componentID)
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.error = e
         })
      },

      async updateMasterFileMetadata({file, title, description, status, componentID}) {
         const system = useSystemStore()
         system.updating = true
         let data = [{file: file, title: title.trim(), description: description.trim(), status: status, componentID: componentID}]
         return axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            this.applyMasterFileMetadataUpdate(data)
            system.updating = false
            if (resp.data.success == false) {
               system.setError("Unable to update image metadata")
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.error = e
            system.updating = false
         })
      },

      async setTag({file, tag}) {
         const system = useSystemStore()
         system.updating = true
         let mf = this.masterFiles.find( mf => mf.path == file)
         let status = tag
         if (tag == "none") {
            status = ""
         }
         let data = [{file: file, title: mf.title.trim(), description: mf.description.trim(), status: status}]
         return axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            this.applyMasterFileMetadataUpdate(data)
            system.updating = false
            if (resp.data.success == false) {
               system.setError("Unable to set tag on image")
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.error = e
         })
      },

      deleteMasterFile(mf) {
         const system = useSystemStore()
         system.updating = true
         axios.delete(`/api/units/${this.currUnit}/${mf}`).then(() => {
            this.removeMasterFile(mf)
            system.updating = false
            window.location.reload()
         }).catch( e => {
            system.error = e
         })
      },

      rotateImage({file, dir}) {
         const system = useSystemStore()
         system.updating = true
         axios.post(`/api/units/${this.currUnit}/${file}/rotate?dir=${dir}`, {}).then(() => {
            setTimeout( ()=>{
               system.updating = false
               window.location.reload()
            }, 250)
         }).catch( e => {
            system.error = e
         })
      },

      renameAll() {
         const system = useSystemStore()
         system.updating = true
         let data = []
         let pStartIdx = this.pageStartIdx
         let pEndIdx = pStartIdx + this.pageSize-1
         let numRegex = /^\d{4}$/
         this.masterFiles.forEach( (mf,idx) => {
            let originalFN = mf.fileName.toLowerCase()
            let originalPath = mf.path

            // if this MF is in the range of the acitve page, grab data for it from
            // the pageMasterFiles array instead as it may have been reordered with drag/drop
            if (idx >= pStartIdx && idx <= pEndIdx) {
               let pageMf = this.pageMasterFiles[idx-pStartIdx]
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
            let paddedUnit = `${this.currUnit}`.padStart(9,"0")
            if ( badName || mfPage !=  idx+1 || mfParts[0] != paddedUnit) {
               let newPg = `${idx+1}`
               newPg = newPg.padStart(4,'0')
               let newFN = `${paddedUnit}_${newPg}.tif`
               data.push({original: originalPath, new: newFN })
            }

         })
         axios.post(`/api/units/${this.currUnit}/rename`, data).then(() => {
            window.location.reload()
         }).catch( e => {
            system.error = e
         })
      },

      lookupComponentID(componentID) {
         const system = useSystemStore()
         system.updating = true
         axios.get(`/api/components/${componentID}`).then(response => {
            this.setComponentInfo(response.data)
            system.updating = false
         }).catch( e => {
            system.error = e
         })
      }
   }
})