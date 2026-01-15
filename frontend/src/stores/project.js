import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useProjectStore = defineStore('project', {
   state: () => ({
      detail: null,
      statusCheckIntervalID: -1,
      working: true,
      missingComponents: []
   }),
   getters: {
      hasMissingComponents: state => {
         if (state.detail == null) return false
         if (state.detail.workflow == null) return false
         return state.detail.workflow.name == "Manuscript" && state.missingComponents.length > 0
      },
      isManuscript: state => {
         if (state.detail == null) return false
         return state.detail.workflow && state.detail.workflow.name=='Manuscript'
      },
      hasDetail: state => {
         return state.detail != null
      },
      dueDate: state => {
         if ( state.detail == null ) return ""
         if (!state.detail.unit ) return ""
         return  state.detail.unit.order.dateDue.split("T")[0]
      },
      // NOTES : enums from tracksys models
      // assignment status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing]
      canReject: state => {
         if ( state.detail == null ) return false
         if ( state.detail.assignments == null || state.detail.assignments === undefined) return false
         if ( state.detail.assignments.length == 0) return false
         let currA = state.detail.assignments[0]
         return currA.step.failStepID > 0 && currA.status == 1
      },
      isFinalizeRunning: state => {
         if ( state.detail == null ) return false
         if ( state.detail.assignments == null || state.detail.assignments === undefined) return false
         if ( state.detail.assignments.length == 0 ) return false
         let currA = state.detail.assignments[0]
         return currA.status == 6 // finalizing
      },
      isFinished: state => {
         if ( state.detail == null ) return false
         return Object.hasOwn(state.detail, 'finishedAt') && state.detail.finishedAt != ""
      },
      hasError: state => {
         if ( state.detail == null ) return false
         if ( state.detail.assignments == null || state.detail.assignments === undefined) return false
         if (state.detail.assignments.length == 0) return false
         let currA = state.detail.assignments[0]
         if (currA.status == 4) return true
         if (currA.step.stepType == 2) return true
         return false
      },
      hasOwner: state => {
         if ( state.detail == null ) return false
         return (state.detail.owner !== undefined && state.detail.owner != null)
      },
      inProgress: state => {
         if ( state.detail == null ) return false
         if ( state.detail.assignments == null || state.detail.assignments === undefined) return false
         if (state.detail.assignments.length == 0) return false
         let currA = state.detail.assignments[0]
         return  currA.status == 1 ||  currA.status == 4 ||  currA.status == 6 ||  currA.status == 7
      },
      isWorking: state => {
         if ( state.detail == null ) return false
         if ( state.detail.assignments == null || state.detail.assignments === undefined) return false
         if ( state.detail.assignments.length == 0) return false
         let currA = state.detail.assignments[0]
         return currA.status == 7
      },
      isOwner: state => {
         return (computeID) => {
            if ( state.detail == null ) return false
            if ( state.detail.owner ) {
               return state.detail.owner.computingID == computeID
            }
            return false
         }
      },
   },
   actions: {
      async getProject(projectID) {
         const system = useSystemStore()
         this.working = true
         await axios.get(`/api/projects/${projectID}`).then(response => {
            this.detail = response.data
            this.detail.containerType = system.getContainerType(this.detail.containerTypeID)
            delete this.detail.containerTypeID
            this.detail.owner = system.getStaffMember(this.detail.ownerID)
            delete this.detail.ownerID
            this.working = false
            axios.put(`/api/projects/${projectID}/images/count`)
         }).catch( e => {

            system.setError( e )
            this.working = false
         })
      },
      cancelStatusPolling() {
         if (this.statusCheckIntervalID > -1) {
            clearInterval( this.statusCheckIntervalID )
            this.statusCheckIntervalID = -1
         }
      },
      async setEquipment(data) {
         // data contains { workstationID, captureResolution, resizeResolution, resolutionNote }
         this.working = true
         return axios.post(`/api/projects/${this.detail.id}/equipment`, data).then(response => {
            this.detail.equipment = response.data.equipment
            this.detail.workstation = response.data.workstation
            this.detail.captureResolution = response.data.captureResolution
            this.detail.resizedResolution = response.data.resizedResolution
            this.detail.resolutionNote = response.data.resolutionNote
            this.working = false
         }).catch( e => {
            const system = useSystemStore()
            system.setError( e )
            this.working = false
         })
      },

      async updateProject(data) {
         // data contains { categoryID, containerTypeID, condition, note, ocrHintID, ocrLangage, ocrMasterFiles }
         this.working = true
         const system = useSystemStore()
         return axios.put(`/api/projects/${this.detail.id}`, data).then(response => {
            console.log("GOT RESPONSE FOR UPDATE")
            console.log(response.data)
            this.detail.category = system.getCategory( response.data.categoryID )
            this.detail.containerType = system.getContainerType( response.data.containerTypeID )
            this.detail.itemCondition = response.data.condition
            this.detail.conditionNote = response.data.note
            this.detail.unit.ocrMasterFiles = response.data.ocrMasterFiles
            this.detail.unit.metadata.ocrHint = system.getOCRHint( response.data.ocrHintID )
            this.detail.unit.metadata.ocrLanguageHint = response.data.ocrLangage
            this.working = false
         }).catch( e => {
            console.error(e)
            system.setError( e )
            this.working = false
         })
      },
      async addNote(data) {
         // data contains { noteTypeID, note, problemIDs}
         this.working = true
         data.stepID = this.detail.currentStep.id
         return axios.post(`/api/projects/${this.detail.id}/note`, data).then(response => {
            this.detail.notes = response.data
            this.working = false
         }).catch( e => {
            const system = useSystemStore()
            system.setError( e )
            this.working = false
         })
      },
      startStep() {
         this.working = true
         axios.post(`/api/projects/${this.detail.id}/start`).then(response => {
            this.detail.startedAt = response.data.startedAt
            this.detail.assignments =  response.data.assignments
            this.working = false
         }).catch( e => {
            const system = useSystemStore()
            system.setError( e )
            this.working = false
         })
      },
      finishStep(durationMins) {
         this.working = true
         let isFinalize = (this.detail.assignments[0].step.name == "Finalize")
         axios.post(`/api/projects/${this.detail.id}/finish`, {durationMins: durationMins} ).then(response => {
            this.detail.owner = response.data.owner
            this.detail.currentStep = response.data.currentStep
            this.detail.assignments = response.data.assignments
            this.detail.notes = response.data.notes
            this.working = false
            if ( isFinalize ) {
               this.pollProjectStatus()
            }
         }).catch( e => {
            const system = useSystemStore()
            system.setError( e )
            this.working = false
         })
      },
      async validateComponents() {
         if (this.detail.workflow.name != "Manuscript" ) {
            // not manuscript, nothing to do
            return
         }
         this.missingComponents = []
         await axios.get(  `/api/units/${this.detail.unit.id}/validate/components` ).then(response => {
            if (response.data.valid == false) {
               this.missingComponents = response.data.missing
            }
         }).catch( e => {
            const system = useSystemStore()
            system.setError( e)
         })
      },
      pollProjectStatus() {
         if ( this.isFinalizeRunning == false) return

         const system = useSystemStore()
         this.statusCheckIntervalID = setInterval( ()=>{
            axios.get(`/api/projects/${this.detail.id}/status`).then(response => {
               if (response.data == "error" || response.data == "finished") {
                  this.getProject( this.detail.id )
                  clearInterval(this.statusCheckIntervalID )
                  this.statusCheckIntervalID = -1
               }
            }).catch( e => {
               system.setError( e )
               clearInterval(this.statusCheckIntervalID )
               this.statusCheckIntervalID = -1
            })
         }, 5000)
      },
      rejectStep(durationMins) {
         this.working = true
         axios.post(`/api/projects/${this.detail.id}/reject`, {durationMins: durationMins} ).then(response => {
            this.detail.owner = response.data.owner
            this.detail.currentStep = response.data.currentStep
            this.detail.assignments = response.data.assignments
            this.detail.notes = response.data.notes
            this.working = false
         }).catch( e => {
            const system = useSystemStore()
            system.setError( e )
            this.working = false
         })
      },
      async assignProject({projectID, ownerID}) {
         this.working = true
         const system = useSystemStore()
         return axios.post(`/api/projects/${projectID}/assign/${ownerID}`).then(response => {
            console.log(response.data)
            this.detail.notes = response.data.notes
            this.detail.assignments = response.data.assignments
            this.detail.owner = null
            if ( response.data.ownerID > 0) {
               this.detail.owner = system.getStaffMember( response.data.ownerID )
            }
            this.working = false
         }).catch( e => {

            system.setError( e )
            this.working = false
         })
      }
   }
})