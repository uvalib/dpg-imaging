<template>
   <div class="panel">
      <dl>
         <dt>Name:</dt>
         <dd>{{currProject.workflow.name}}</dd>
         <dt>Step:</dt>
         <dd>{{currProject.currentStep.description}}</dd>
         <dt>Owner:</dt>
         <dd>
            <span v-if="currProject.owner.id > 0">{{currProject.owner.firstName}} {{currProject.owner.lastName}}</span>
            <span v-else class="na">Unassigned</span>
         </dd>
         <dt>Assigned:</dt>
         <dd>
            <span v-if="currProject.owner.id > 0">{{assignedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Started:</dt>
         <dd>
            <span v-if="currProject.owner.id > 0 && startedAt">{{startedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Directory:</dt>
         <dd>{{workingDir}}</dd>
      </dl>
      <div class="workflow-btns">
         <template v-if="isOwner(computingID)">
            <DPGButton >Open QA Viewer</DPGButton>
            <DPGButton v-if="(isOwner(computingID) || isSupervisor || isAdmin) && isFinalizing(projectIdx) == false">Reassign</DPGButton>
            <DPGButton v-if="inProgress(projectIdx) == false">Start</DPGButton>
            <DPGButton v-if="canReject(projectIdx)" class="reject">Reject</DPGButton>
            <DPGButton v-if="hasError(projectIdx) == false && inProgress(projectIdx) == true">Finish</DPGButton>
            <DPGButton v-if="onFinalizeStep(projectIdx) &&  hasError(projectIdx) == true">Retry Finalize</DPGButton>
         </template>
         <template v-else>
            <DPGButton v-if="hasOwner(projectIdx) == false">Claim</DPGButton>
            <DPGButton v-if="(isAdmin || isSupervisor) && isFinalizing(projectIdx) == false">Assign</DPGButton>
         </template>
      </div>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
import date from 'date-and-time'
export default {
   components: {
   },
   computed: {
      ...mapState({
         adminURL: state => state.adminURL,
         scanDir: state => state.scanDir,
         qaDir: state => state.qaDir,
         projectIdx: state => state.projects.selectedProjectIdx,
         computingID: state => state.user.computeID
      }),
      ...mapGetters({
         currProject: 'projects/currProject',
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
      }
   },
   methods: {
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
   .workflow-btns {
      text-align: right;
      padding: 10px;
      border-top: 1px solid var(--uvalib-grey-light);
      .dpg-button {
         margin-left: 10px;
      }
   }
}
</style>
