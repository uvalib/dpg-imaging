<template>
   <div class="panel">
      <dl>
         <dt>Name:</dt>
         <dd>{{currProject.workflow.name}}</dd>
         <dt>Step:</dt>
         <dd>{{currProject.currentStep.description}}</dd>
         <dt>Owner:</dt>
         <dd>
            <span v-if="hasOwner(selectedProjectIdx)">{{currProject.owner.firstName}} {{currProject.owner.lastName}}</span>
            <span v-else class="na">Unassigned</span>
         </dd>
         <dt>Assigned:</dt>
         <dd>
            <span v-if="hasOwner(selectedProjectIdx)">{{assignedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Started:</dt>
         <dd>
            <span v-if="hasOwner(selectedProjectIdx) && startedAt">{{startedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Directory:</dt>
         <dd>{{workingDir}}</dd>
      </dl>
      <div class="workflow-btns time" v-if="timeEntry">
         <div class="time-form">
            <label for="time">Approximately how many minutes did you spend on this assignment?</label>
            <input id="time" type="number" v-model="stepMinutes"  @keyup.enter="timeEntered">
         </div>
         <div class="ok-cancel">
             <DPGButton @click="cancelFinish">Cancel</DPGButton>
             <DPGButton @click="timeEntered">OK</DPGButton>
         </div>
      </div>
      <div class="workflow-btns" v-else-if="isFinalizeRunning(selectedProjectIdx) == false && isFinished(selectedProjectIdx) == false">
         <template v-if="isOwner(userStore.computeID)">
            <DPGButton @click="viewerClicked" class="pad-right" v-if="isScanning == false">Open QA Viewer</DPGButton>
            <template v-if="isWorking(selectedProjectIdx) == false">
               <AssignModal v-if="(isOwner(userStore.computeID) || isSupervisor || isAdmin) "
                  :projectID="currProject.id" @assign="assignClicked" label="Reassign"/>
               <DPGButton v-if="inProgress(selectedProjectIdx) == false" @click="startStep">Start</DPGButton>
               <DPGButton v-if="inProgress(selectedProjectIdx) == true" :disabled="!isFinishEnabled" @click="finishClicked">
                  <template v-if="isFinalizing &&  hasError(selectedProjectIdx) == true">Retry Finalize</template>
                  <template v-else>Finish</template>
               </DPGButton>
               <DPGButton v-if="canReject(selectedProjectIdx)" class="reject"  @click="rejectStepClicked">Reject</DPGButton>
            </template>
         </template>
         <template v-else>
            <DPGButton @click="viewerClicked" class="pad-right" v-if="isScanning == false && (isAdmin || isSupervisor)">Open QA Viewer</DPGButton>
            <DPGButton v-if="isWorking(selectedProjectIdx) == false && hasOwner(selectedProjectIdx) == false" @click="claimClicked()"  class="pad-right">Claim</DPGButton>
            <AssignModal v-if="(isAdmin || isSupervisor)" :projectID="currProject.id" @assign="assignClicked"/>
         </template>
         <DPGButton v-if="hasOwner(selectedProjectIdx) && (isAdmin || isSupervisor)" @click="clearClicked()" class="pad-right">
            Clear Assignment
         </DPGButton>
      </div>
      <div class="workflow-message" v-if="isOwner(userStore.computeID) && workflowNote">
         {{workflowNote}}
      </div>
      <NoteModal id="problem-modal" :manual="true" :trigger="showRejectNote" :noteType="2"
         @closed="rejectCanceled" @submitted="rejectSubmitted"
         instructions="Rejection requires the addition of a problem note that details the reason why it occurred" />
   </div>
</template>

<script setup>
import date from 'date-and-time'
import AssignModal from "@/components/AssignModal.vue"
import NoteModal from '@/components/project/NoteModal.vue'
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import { ref, computed, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'

const router = useRouter()
const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const {
   currProject, selectedProjectIdx, isOwner, isAdmin, isSupervisor, hasOwner,
   isFinalizeRunning, isFinished, inProgress, isWorking, canReject, hasError
} = storeToRefs(projectStore)

const timeEntry = ref(false)
const stepMinutes = ref(0)
const action = ref("finish")
const showRejectNote = ref(false)

const currStepName = computed(()=>{
   return currProject.value.currentStep.name
})
const isFinalizing = computed(()=>{
   return currStepName.value == 'Finalize'
})
const isScanning = computed(()=>{
   return (currStepName.value == 'Scan' || currStepName.value == 'Process')
})
const isFinishEnabled = computed(()=>{
   if ( currStepName.value == "Scan" && currProject.value.workstation.id == 0) return false
   if ( currStepName.value == "Finalize") {
      if ( currProject.value.unit.metadata.ocrHint.id == 0) return false
      if ( currProject.value.unit.metadata.ocrHint.id == 1 && currProject.value.unit.metadata.ocrLanguageHint == "") return false
   }
   return true
})
const workingDir = computed(()=>{
   let unitDir =  unitDirectory(currProject.value.unit.id)
   if (currProject.value.currentStep.name == "Process" || currProject.value.currentStep.name == "Scan") {
      return `${systemStore.scanDir}/${unitDir}`
   }
   return `${systemStore.qaDir}/${unitDir}`
})
const assignedAt = computed(()=>{
   let stepID = currProject.value.currentStep.id
   let a = currProject.value.assignments.find( a => a.stepID == stepID)
   if (a && a.assignments)  {
      return date.format(new Date(a.assignedAt), "YYYY-MM-DD hh:mm A")
   }
   return ""
})
const startedAt = computed(()=>{
   let stepID = currProject.value.currentStep.id
   let a = currProject.value.assignments.find( a => a.stepID == stepID)
   if ( a && a.startedAt) return date.format(new Date(a.startedAt), "YYYY-MM-DD hh:mm A")
   return ""
})
const workflowNote = computed(()=>{
   if ( currStepName.value == "Scan" && currProject.value.workstation.id == 0) {
      return "Assignment cannot be finished until the workstation has been set."
   }
   if ( currStepName.value == "Finalize" && currProject.value.unit.metadata.ocrHint.id == 0) {
      return "Assignment cannot be finished until the OCR hint has been set."
   }
   if ( currProject.value.unit.metadata.ocrHint.id > 1 && currProject.value.unit.ocrMasterFiles == true) {
      return "Cannot OCR items that are not regular text."
   }
   if ( currProject.value.unit.metadata.ocrHint.id == 1 && currProject.value.unit.metadata.ocrLanguageHint == "" && currStepName.value == "Finalize") {
      return "Assignment cannot be finished until the OCR Language Hint has been set."
   }
   if ( currProject.value.unit.status == "error" ) {
      return "Finalization has failed. Correct the problem then click 'Retry Finalization'."
   }
   return ""
})

function clearClicked() {
   projectStore.assignProject({projectID: currProject.value.id, ownerID: 0} )
}

function rejectStepClicked() {
   action.value = "reject"
   showTimeEntry()
}

function assignClicked( info ) {
   projectStore.assignProject({projectID: currProject.value.id, ownerID: info.ownerID} )
}

function claimClicked() {
   projectStore.assignProject({projectID: currProject.value.id, ownerID: userStore.ID} )
}

function finishClicked() {
   action.value = "finish"
   if ( currProject.value.assignments[0].durationMinutes == 0) {
      showTimeEntry()
   } else {
      // send a 0 time to indicate that time has already been recorded
      projectStore.finishStep(0)
   }
}

function showTimeEntry() {
   timeEntry.value = true
   stepMinutes.value = 0
   nextTick( () => {
      let te = document.getElementById("time")
      if (te) {
         te.focus()
         te.select()
      }
   })
}

function timeEntered() {
   if ( action.value == "finish")  {
      projectStore.finishStep(stepMinutes.value)
      timeEntry.value = false
      stepMinutes.value = 0
   } else {
      showRejectNote.value = true
      timeEntry.value = false
   }
}

function rejectCanceled() {
   showRejectNote.value = false
   timeEntry.value = false
   stepMinutes.value = 0
}

function rejectSubmitted() {
   projectStore.rejectStep(stepMinutes.value)
   showRejectNote.value = false
   timeEntry.value = false
   stepMinutes.value = 0
}

function cancelFinish() {
   timeEntry.value = false
}

function viewerClicked() {
   router.push(`/projects/${currProject.value.id}/unit`)
}

function startStep() {
   projectStore.startStep()
}

function unitDirectory(unitID) {
   let ud = ""+unitID
   return ud.padStart(9, "0")
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
       .dpg-button.reject {
          background: rgb(178, 34, 34);
          color: white;
          &:hover {
             background: rgb(210, 60, 60);
          }
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
      color: var(--uvalib-red-emergency);
   }
}
</style>
