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
      async setPage(startPageNum) {
         this.currPage = startPageNum
         return this.getMetadataPage()
      },
      setPageSize(newSize) {
         this.pageSize = newSize
         this.setPage(1)
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
         this.viewMode = "list"
      },

      async getUnitMasterFiles(unit) {
         // dont try to reload a unit if the data is already present
         let intUnit = parseInt(unit, 10)
         if (this.currUnit == intUnit && this.masterFiles.length > 0 || this.working == true) return

         const system = useSystemStore()
         system.working = true
         this.clearUnitDetails()
         return axios.get(`/api/units/${unit}/masterfiles`).then(response => {
            this.currUnit = unit
            this.masterFiles.splice(0, this.masterFiles.length)
            response.data.masterFiles.forEach( mf =>{
               mf.error = ""
               this.masterFiles.push(mf)
            })
            this.problems = response.data.problems
            system.working = false
         }).catch( e => {
            system.error = e
            system.working = false
         })
      },

      async getMasterFileMetadata( masterFileIndex ) {
         let mf = this.masterFiles[masterFileIndex]
         if (!mf) return
         if (mf.resolution) return

         const system = useSystemStore()
         system.working = true
         let mdURL = `/api/units/${ this.currUnit}/masterfiles/metadata?file=${mf.path}`
         return axios.get(mdURL).then(response => {
            let md = response.data[0]
            mf.colorProfile = md.colorProfile
            mf.fileSize = md.fileSize
            mf.fileType = md.fileType
            mf.resolution = md.resolution
            mf.title = md.title
            mf.description = md.description
            mf.width = md.width
            mf.height = md.height
            mf.status = md.status
            mf.componentID = md.componentID
            system.working = false
         }).catch( e => {
            system.error = e
            system.working = false
         })
      },

      async getMetadataPage() {
         // must have a unit set to get metdata
         if (this.currUnit == "") return
         let startIdx = (this.currPage-1) * this.pageSize
         let endIdx = startIdx+this.pageSize-1
         if (endIdx >= this.masterFiles.length-1) {
            endIdx = this.masterFiles.length-1
         }
         let needsData = false
         for ( let i=startIdx; i<=endIdx; i++) {
            if ( !this.masterFiles[i].resolution ) {
               needsData = true
               break
            }
         }
         if (needsData == false ) {
            this.pageMasterFiles = this.masterFiles.slice(startIdx, startIdx+this.pageSize)
            return
         }

         const system = useSystemStore()
         system.working = true
         console.log("getMetadataPage")
         let mdURL = `/api/units/${ this.currUnit}/masterfiles/metadata?page=${this.currPage}&pagesize=${this.pageSize}`
         return axios.get(mdURL).then(response => {
            system.working = false
            response.data.forEach( md => {
               let mf = this.masterFiles.find( m => m.fileName == md.fileName)
               mf.colorProfile = md.colorProfile
               mf.fileSize = md.fileSize
               mf.fileType = md.fileType
               mf.resolution = md.resolution
               mf.title = md.title
               mf.description = md.description
               mf.width = md.width
               mf.height = md.height
               mf.status = md.status
               mf.box = md.box
               mf.folder = md.folder
               mf.componentID = md.componentID
            })
            this.pageMasterFiles = this.masterFiles.slice(startIdx, startIdx+this.pageSize)
         }).catch( e => {
            system.error = e
            system.working = false
         })
      },

      updatePageNumbers( start, verso ) {
         const system = useSystemStore()
         system.working = true
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
            data.push( {file: mf.path, field: "title", value: title})
            pageCnt+=1
         }

         axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
               let update = data.shift()
               this.masterFiles[i].title = update.value
            }
            system.working = false
            if (resp.data.success == false) {
               system.setError("Some images were not renumbered")
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.setError(e)
         })
      },

      async componentLink(componentID) {
         const system = useSystemStore()
         system.working = true
         let data = []
         for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
            let mf = this.masterFiles[i]
            data.push( {file: mf.path, field: "component", value: componentID})
         }
         return axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
               let update = data.shift()
               this.masterFiles[i].componentID = update.value
            }
            system.working = false
            if (resp.data.success == false) {
               system.setError("Some images could not be linked with component "+componentID)
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.setError(e)
         })
      },

      async updateMasterFileMetadata(file, field, value) {
         const system = useSystemStore()
         system.working = true
         if (field == "tag" && value == "none") {
            value = ""
         }
         return axios.post(`/api/units/${this.currUnit}/${file}/update?field=${field}&value=${encodeURIComponent(value.trim())}`).then( resp => {
            let tgtMF = this.masterFiles.find( mf => mf.fileName == file)
            if (field == "title") {
               tgtMF.title = value
            } else if (field == "description") {
               tgtMF.description = value
            } else if (field == "box") {
               tgtMF.box = value
            } else if (field == "folder") {
               tgtMF.folder = value
            } else if (field == "tag") {
               tgtMF.status = value
            }
            system.working = false
            if (resp.data.success == false) {
               system.setError("Unable to update image metadata")
               this.setMasterFileProblems(resp.data.problems)
            }
         }).catch( e => {
            system.setError(e)
         })
      },

      async rotateImage({file, dir}) {
         const system = useSystemStore()
         system.working = true
         return axios.post(`/api/units/${this.currUnit}/${file}/rotate?dir=${dir}`, {}).then(() => {
            system.working = false
         }).catch( e => {
            system.setError(e)
         })
      },

      renameAll() {
         const system = useSystemStore()
         system.working = true
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
            system.setError(e)
         })
      },

      lookupComponentID(componentID) {
         const system = useSystemStore()
         system.working = true
         axios.get(`/api/components/${componentID}`).then(response => {
            this.setComponentInfo(response.data)
            system.working = false
         }).catch( e => {
            system.setError(e)
         })
      }
   }
})