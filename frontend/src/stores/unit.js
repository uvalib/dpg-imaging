import axios from 'axios'
import { defineStore } from 'pinia'
import { useSystemStore } from './system'

export const useUnitStore = defineStore('unit', {
   state: () => ({
      currUnit: "",
      masterFiles: [],
      viewMode: "list",
      rangeStartIdx: -1,
      rangeEndIdx: -1,
      edit: {
         pageNumber: false,
         component: false,
         metadata: false,
      },
      problems: [],
      component: {
         valid: false,
         title: "",
         label: "",
         desc: "",
         date: "",
         type: "",
      },
      lastURL: "",
      currPage: 0,
      pageSize: 20,
      containerType: null
   }),
   getters: {
      pageInfoURLs: state => {
         return state.masterFiles.map( mf => mf.infoURL )
      },
      totalFiles: state => {
         return state.masterFiles.length
      },
      currStartIndex: state => {
         return state.currPage * state.pageSize
      }
   },
   actions: {
      moveImage( fromIndex, toIndex ) {
         let img = this.masterFiles.splice(fromIndex, 1)[0]
         this.masterFiles.splice(toIndex, 0, img)
      },
      selectPage() {
         this.rangeStartIdx = this.currStartIndex
         this.rangeEndIdx = this.rangeStartIdx + this.pageSize
         if (this.rangeEndIdx >= this.masterFiles.length) {
            this.rangeEndIdx = this.masterFiles.length-1
         }
         for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
            this.masterFiles[i].selected = true
         }
      },
      startFileSelected(idx) {
         if (this.rangeStartIdx > -1 || this.rangeStartIdx != idx)  {
            this.deselectAll()
         }
         this.rangeEndIdx = idx
         this.rangeStartIdx = idx
         this.masterFiles[idx].selected = true
      },
      endFileSelected(idx) {
         if ( idx <= this.rangeStartIdx) {
            this.startFileSelected(idx)
         } else {
            this.rangeEndIdx = idx
            for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
               this.masterFiles[i].selected = true
            }
         }
      },
      selectAll() {
         this.rangeStartIdx = 0
         this.rangeEndIdx = this.masterFiles.length - 1
         this.masterFiles.forEach( mf => mf.selected = true)
      },
      deselectAll() {
         this.rangeStartIdx = -1
         this.rangeEndIdx = - 1
         this.masterFiles.forEach( mf => mf.selected = false)
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

      masterFileSelected(idx) {
         if (this.masterFiles[idx].selected) {
            this.masterFiles.forEach( mf => mf.selected = false)
            this.rangeStartIdx = -1
            this.rangeEndIdx = -1
         } else {
            let priorSelIdx = this.masterFiles.findIndex( mf => mf.selected)
            if (priorSelIdx == -1) {
               this.masterFiles[idx].selected = true
               this.rangeStartIdx = idx
               this.rangeEndIdx = idx
            } else {
               if (priorSelIdx < idx) {
                  this.rangeStartIdx = priorSelIdx
                  this.rangeEndIdx = idx
               } else {
                  this.rangeStartIdx = idx
                  this.rangeEndIdx = priorSelIdx
               }
               for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
                  this.masterFiles[i].selected = true
               }
            }
         }
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
               mf.selected = false
               this.masterFiles.push(mf)
            })
            this.problems = response.data.problems
            system.working = false
         }).catch( e => {
            system.error = e
            system.working = false
         })
      },

      // NOTES: this is only used from the image view
      async getMasterFileMetadata( masterFileIndex ) {
         let mf = this.masterFiles[masterFileIndex]
         if (!mf) return
         if (mf.resolution) return

         const system = useSystemStore()
         system.working = true
         let mdURL = `/api/units/${ this.currUnit}/masterfiles/metadata?file=${mf.path}`
         return axios.get(mdURL).then(response => {
            this.setImageMetadata(response.data)
            system.working = false
         }).catch( e => {
            system.error = e
            system.working = false
         })
      },

      setImageMetadata( md ) {
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
         mf.location = "Undefined"
         mf.box = ""
         mf.folder = ""
         if (md.location ) {
               mf.location = md.location
         }
         if ( md.box && md.box != "<nil>" ) {
            mf.box = md.box
            if ( this.containerType.hasFolders && md.folder ) {
               mf.folder = md.folder
            }
            // for existing projects before this update, location will not be set
            // generate it from the box/folder info
            if (mf.location == "Undefined") {
               mf.location = `${ this.containerType.name} ${mf.box}`
               if ( mf.folder) {
                  mf.location += `, Folder ${mf.folder}`
               }
            }
         }
         mf.componentID = md.componentID
      },

      async getMetadataPage() {
         console.log("GET PAGEINDEX "+this.currPage+" sz "+this.pageSize)
         if (this.currUnit == "") return

         let startIdx = this.currPage * this.pageSize
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
            return
         }

         const system = useSystemStore()
         system.working = true
         let mdURL = `/api/units/${ this.currUnit}/masterfiles/metadata?page=${this.currPage+1}&pagesize=${this.pageSize}`
         return axios.get(mdURL).then(response => {
            system.working = false
            response.data.forEach( md => {
               this.setImageMetadata( md )
            })
         }).catch( e => {
            system.error = e
            system.working = false
         })
      },

      deleteSelectedMasterFiles() {
         const system = useSystemStore()
         system.working = true
         let data = []
         for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
            let mf = this.masterFiles[i]
            data.push( mf.fileName )
         }
         axios.post(`/api/units/${this.currUnit}/delete`, {filenames: data}).then( () => {
            data.forEach( fn => {
               let idx = this.masterFiles.findIndex( mf => mf.fileName == fn)
               if (idx > -1 ) {
                  this.masterFiles.splice(idx, 1)
               }
            })
            system.working = false
         }).catch( e => {
            system.setError(e)
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

      async batchUpdate(field, value) {
         const system = useSystemStore()
         system.working = true
         let data = []
         for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
            let mf = this.masterFiles[i]
            data.push( {file: mf.path, field: field, value: value})
         }
         return axios.post(`/api/units/${this.currUnit}/update`, data).then( resp => {
            for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
               let update = data.shift()
               if (field == "folder") {
                  this.masterFiles[i].folder = update.value
               } else if (field == "box") {
                  this.masterFiles[i].box = update.value
               } else if (field == "title") {
                  this.masterFiles[i].title = update.value
               } else if (field == "description") {
                  this.masterFiles[i].description = update.value
               }
               if ( field == "folder" || field == "box" ) {
                  this.masterFiles[i].location = `${ this.containerType.name} ${this.masterFiles[i].box}`
                  if ( this.masterFiles[i].folder) {
                     this.masterFiles[i].location += `, Folder ${this.masterFiles[i].folder}`
                  }
               }
            }
            system.working = false
            if (resp.data.success == false) {
               system.setError("Folder assignment failed for some images")
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
            this.setImageMetadata(resp.data)
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
         let numRegex = /^\d{4}$/
         this.masterFiles.forEach( (mf,idx) => {
            let originalFN = mf.fileName.toLowerCase()
            let originalPath = mf.path

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

      async lookupComponentID(componentID) {
         const system = useSystemStore()
         return axios.get(`/api/components/${componentID}`).then(response => {
            this.setComponentInfo(response.data)
         }).catch( e => {
            system.setError(e)
         })
      }
   }
})