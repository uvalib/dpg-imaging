<template>
   <div class="panel">
      <dl>
         <dt>Name:</dt>
         <dd>{{currProject.workflow.name}}</dd>
         <dt>Step:</dt>
         <dd>{{currProject.currentStep.description}}</dd>
         <dt>Owner:</dt>
         <dd>
            <span v-if="hasOwner(projectIdx)">{{currProject.owner.firstName}} {{currProject.owner.lastName}}</span>
            <span v-else class="na">Unassigned</span>
         </dd>
         <dt>Assigned:</dt>
         <dd>
            <span v-if="hasOwner(projectIdx)">{{assignedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Started:</dt>
         <dd>
            <span v-if="hasOwner(projectIdx) && startedAt">{{startedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Directory:</dt>
         <dd>{{workingDir}}</dd>
      </dl>
      <div class="workflow-btns time" v-if="timeEntry">
         <div class="time-form">
            <label for="time">Approximately how many minutes did you spend on this assignment?</label>
            <input id="time" type="number" v-model="stepMinutes">
         </div>
         <div class="ok-cancel">
             <DPGButton @click="cancelFinish">Cancel</DPGButton>
             <DPGButton @click="finshTimeEntered">OK</DPGButton>
         </div>
      </div>
      <div class="workflow-btns" v-else>
         <template v-if="isOwner(computingID)">
            <DPGButton @click="viewerClicked" class="pad-right">Open QA Viewer</DPGButton>
            <AssignModal v-if="(isOwner(computingID) || isSupervisor || isAdmin) && isFinalizing(projectIdx) == false"
               :projectID="currProject.id" @assign="assignClicked" label="Reassign"/>
            <DPGButton v-if="inProgress(projectIdx) == false" @click="startStep">Start</DPGButton>
            <DPGButton v-if="canReject(projectIdx)" class="reject">Reject</DPGButton>
            <DPGButton v-if="inProgress(projectIdx) == true" :disabled="!isFinishEnabled" @click="finishClicked">Finish</DPGButton>
            <DPGButton v-if="onFinalizeStep(projectIdx) &&  hasError(projectIdx) == true">Retry Finalize</DPGButton>
         </template>
         <template v-else>
            <DPGButton v-if="hasOwner(projectIdx) == false" @click="claimClicked()">Claim</DPGButton>
            <AssignModal v-if="(isAdmin || isSupervisor) && isFinalizing(projectIdx) == false" :projectID="currProject.id" @assign="assignClicked"/>
         </template>
      </div>
      <div class="workflow-message" v-if="isOwner(computingID) && workflowNote">
         {{workflowNote}}
      </div>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
import date from 'date-and-time'
import AssignModal from "@/components/AssignModal"
export default {
   components: {
      AssignModal
   },
   data: function()  {
      return {
         timeEntry: false,
         stepMinutes: 0
      }
    },
   computed: {
      ...mapState({
         adminURL: state => state.adminURL,
         scanDir: state => state.scanDir,
         qaDir: state => state.qaDir,
         projectIdx: state => state.projects.selectedProjectIdx,
         computingID: state => state.user.computeID,
         userID : state => state.user.ID,
      }),
      ...mapGetters({
         currProject: 'projects/currProject',
         currentStepName: 'projects/currentStepName',
         isOwner: 'projects/isOwner',
         isAdmin: 'isAdmin',
         isSupervisor: 'isSupervisor',
         isFinalizing: 'projects/isFinalizing',
         inProgress: 'projects/inProgress',
         canReject: 'projects/canReject',
         onFinalizeStep: 'projects/onFinalizeStep',
         hasError: 'projects/hasError',
         hasOwner: 'projects/hasOwner',
      }),
      isFinishEnabled() {
         if ( this.currentStepName.name == "Scan" && this.currProject.workstation.id == 0) return false
         if ( this.currentStepName.name == "Finalize") {
            if ( this.currProject.unit.metadata.ocrHint.id == 0) return false
            if ( this.currProject.unit.metadata.ocrHint.id == 1 && this.currProject.unit.metadata.ocrLanguageHint == "") return false
         }
         return true
      },
      workingDir() {
         let unitDir =  this.unitDirectory(this.currProject.unit.id)
         if (this.currProject.currentStep.name == "Process" || this.currProject.currentStep.name == "Scan") {
            return `${this.scanDir}/${unitDir}`
         }
         return `${this.qaDir}/${unitDir}`
      },
      assignedAt() {
         let stepID = this.currProject.currentStep.id
         let a = this.currProject.assignments.find( a => a.stepID == stepID)
         return date.format(new Date(a.assignedAt), "YYYY-MM-DD hh:mm A")
      },
      startedAt() {
         let stepID = this.currProject.currentStep.id
         let a = this.currProject.assignments.find( a => a.stepID == stepID)
         if ( a.startedAt) return date.format(new Date(a.startedAt), "YYYY-MM-DD hh:mm A")
         return ""
      },
      workflowNote() {
         let projID = this.currProject.id
         if ( this.currentStepName(projID) == "Scan" && this.currProject.workstation.id == 0) {
            return "Assignment cannot be finished until the workstation has been set."
         }
         if ( this.currentStepName(projID) == "Finalize" && this.currProject.unit.metadata.ocrHint.id == 0) {
            return "Assignment cannot be finished until the OCR hint has been set."
         }
         if ( this.currProject.unit.metadata.ocrHint.id > 1 && this.currProject.unit.ocrMasterFiles == true) {
            return "Cannot OCR items that are not regular text."
         }
         if ( this.currProject.unit.metadata.ocrHint.id == 1 && this.currProject.unit.metadata.ocrLanguageHint == "") {
            return "Assignment cannot be finished until the OCR Language Hint has been set."
         }
         if ( this.currProject.unit.status == "error" ) {
            return "Finalization has failed. Correct the problem then click 'Retry Finalization'."
         }
         return ""
      }
   },
   methods: {
      assignClicked( info ) {
         this.$store.dispatch("projects/assignProject", {projectID: this.currProject.id, ownerID: info.ownerID} )
      },
      claimClicked() {
         this.$store.dispatch("projects/assignProject", {projectID: this.currProject.id, ownerID: this.userID} )
      },
      finishClicked() {
         if ( this.currProject.assignments[0].durationMinutes == 0) {
            this.timeEntry = true
            this.stepMinutes = 0
         } else {
            // send a 0 time to indicate that time has already been recorded
            this.$store.dispatch("projects/finishStep", 0)
         }
      },
      finshTimeEntered() {
         this.$store.dispatch("projects/finishStep", this.stepMinutes)
         this.timeEntry = false
         this.stepMinutes = 0
      },
      cancelFinish() {
         this.timeEntry = false
      },
      viewerClicked() {
         this.$router.push("/unit/"+this.currProject.unit.id)
      },
      startStep() {
         this.$store.dispatch("projects/startStep")
      },
      formatDate( date ) {
         return date.getUTCFullYear() + "/" +
            ("0" + (date.getUTCMonth()+1)).slice(-2) + "/" +
            ("0" + date.getUTCDate()).slice(-2) + " " +
            ("0" + date.getUTCHours()).slice(-2) + ":" +
            ("0" + date.getUTCMinutes()).slice(-2)
      },
      unitDirectory(unitID) {
         let ud = ""+unitID
         return ud.padStart(9, "0")
      },
   },
}
</script>

<style scoped lang="scss">
.panel {
   dl {
      margin: 10px 30px 0 30px;
      display: inline-grid;
      grid-template-columns: max-content 2fr;
      grid-column-gap: 10px;
      font-size: 0.9em;
      text-align: left;
      box-sizing: border-box;

      dt {
         font-weight: bold;
         text-align: right;
      }
      dd {
         margin: 0 0 10px 0;
         word-break: break-word;
         -webkit-hyphens: auto;
         -moz-hyphens: auto;
         hyphens: auto;
         .na {
            color: #999;
         }
      }
   }
   .pad-right {
      margin-right: 10px;
   }
   .workflow-btns {
      text-align: right;
      padding: 10px;
      border-top: 1px solid var(--uvalib-grey-light);
      .dpg-button {
         margin-left: 10px;
      }
   }
   .workflow-btns.time {
      text-align: left;
      .time-form {
         display: flex;
         flex-flow: flex nowrap;
         justify-content: flex-start;
         margin-bottom: 10px;
         align-items: center;
         font-size: 0.9em;
         label {
            white-space: nowrap;

         }
         input {
            flex-grow: 1;
            margin-left: 10px;
            border-color: var(--uvalib-grey-light);
         }
      }
      .ok-cancel {
         text-align: right;
      }
   }
   .workflow-message {
      padding: 10px;
      border-top: 1px solid var(--uvalib-grey-light);
      text-align: center;
   }
}
</style>
