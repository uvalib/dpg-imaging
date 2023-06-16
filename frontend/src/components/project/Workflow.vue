<template>
   <ConfirmDialog position="top"/>
   <Panel header="Workflow" class="panel">
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
             <DPGButton @click="cancelFinish" class="p-button-secondary" label="Cancel"/>
             <DPGButton @click="timeEntered" label="OK"/>
         </div>
      </div>
      <div class="workflow-btns" v-else-if="isFinalizeRunning(selectedProjectIdx) == false && isFinished(selectedProjectIdx) == false">
         <DPGButton @click="changeWorkflowClicked()" class="p-button-secondary" v-if="projectStore.canChangeWorkflow &&  (isSupervisor || isAdmin)" label="Change Workflow"/>
         <span class="right-buttons">
            <DPGButton @click="viewerClicked" class="p-button-secondary" v-if="isScanning == false && (isOwner(userStore.computeID) || isSupervisor || isAdmin)" label="Open QA Viewer"/>
            <DPGButton v-if="hasOwner(selectedProjectIdx) && (isAdmin || isSupervisor)"
               @click="clearClicked()" class="p-button-secondary pad-right" label="Clear Assignment"/>
            <template v-if="isOwner(userStore.computeID)">
               <template v-if="isWorking(selectedProjectIdx) == false">
                  <AssignModal v-if="(isOwner(userStore.computeID) || isSupervisor || isAdmin)" :projectID="currProject.id" label="Reassign"/>
                  <DPGButton v-if="inProgress(selectedProjectIdx) == false" @click="startStep" label="Start"/>
                  <DPGButton v-if="canReject(selectedProjectIdx)" class="p-button-danger" @click="rejectStepClicked" label="Reject"/>
                  <DPGButton v-if="inProgress(selectedProjectIdx) == true" :disabled="!isFinishEnabled" @click="finishClicked">
                     <template v-if="isFinalizing &&  hasError(selectedProjectIdx) == true">Retry Finalize</template>
                     <template v-else>Finish</template>
                  </DPGButton>
               </template>
            </template>
            <template v-else>
               <DPGButton v-if="isWorking(selectedProjectIdx) == false && (hasOwner(selectedProjectIdx) == false || isAdmin ||isSupervisor)"
                  @click="claimClicked()"  class="p-button-secondary pad-right" label="Claim"/>
               <AssignModal v-if="(isAdmin || isSupervisor)" :projectID="currProject.id" />
            </template>
         </span>
      </div>
      <div class="workflow-message" v-if="isOwner(userStore.computeID) && workflowNote">
         {{workflowNote}}
      </div>
      <NoteModal id="problem-modal" :manual="true" :trigger="showRejectNote" :noteType="2"
         @closed="rejectCanceled" @submitted="rejectSubmitted"
         instructions="Rejection requires the addition of a problem note that details the reason why it occurred" />
   </Panel>
   <Dialog v-model:visible="showWorkflowPicker" :modal="true" header="Change Workflow" style="width:300px">
      <div class="workflow-picker">
         <p>Current workflow: {{ currProject.workflow.name }}</p>
         <p>Select a new workflow:</p>
         <div class="workflow-list">
            <div class="workflow-val" v-for="(w,idx) in activeWorkflows" :key="w.id"
               :class="{selected: idx == selectedWorkflowIdx}" @click="selectWorkflow(idx)"
            >
               {{  w.name }}
            </div>
         </div>
         <div class="container-types" v-if="isManuscriptSelected">
            <p>Select container type:</p>
            <div class="workflow-list">
               <div class="workflow-val" v-for="(ct,idx) in systemStore.containerTypes" :key="ct.id"
                  :class="{selected: idx == selectedContainerTypeIdx}" @click="selectContainerType(idx)"
               >
                  {{  ct.name }}
               </div>
            </div>
         </div>
      </div>
      <template #footer>
         <DPGButton @click="cancelWorkflowChange()" label="Cancel" class="p-button-secondary"/>
         <span class="spacer"></span>
         <DPGButton @click="submitWorkflowChange()" label="Submit" :disabled="isWorkflowChangeDisabled"/>
      </template>
   </Dialog>
</template>

<script setup>
import dayjs from 'dayjs'
import AssignModal from "@/components/AssignModal.vue"
import NoteModal from '@/components/project/NoteModal.vue'
import { useProjectStore } from "@/stores/project"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { ref, computed, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import Panel from 'primevue/panel'
import Dialog from 'primevue/dialog'
import { useConfirm } from "primevue/useconfirm"

const confirm = useConfirm()

const router = useRouter()
const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const {
   currProject, selectedProjectIdx, isOwner, hasOwner, hasError,
   isFinalizeRunning, isFinished, inProgress, isWorking, canReject,
} = storeToRefs(projectStore)
const {isAdmin, isSupervisor} = storeToRefs(userStore)

const timeEntry = ref(false)
const stepMinutes = ref(0)
const action = ref("finish")
const showRejectNote = ref(false)
const showWorkflowPicker = ref(false)
const selectedWorkflowIdx = ref(-1)
const selectedContainerTypeIdx = ref(-1)

const activeWorkflows = computed(() => {
   return systemStore.workflows.filter( wf => wf.isActive == true)
})

const isWorkflowChangeDisabled = computed(() => {
   if ( selectedWorkflowIdx.value == -1 ) return true
   if (isManuscriptSelected.value) {
      if ( selectedContainerTypeIdx.value == -1 ) return true
   }
   return false
})
const isManuscriptSelected = computed(()=>{
   if ( selectedWorkflowIdx.value == -1 ) return false
   let wf = activeWorkflows.value[selectedWorkflowIdx.value]
   if (wf) {
      return wf.name == "Manuscript"
   }
   return false
})
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
   if ( currProject.value.workflow.name == "Manuscript" && currStepName.value == "Scan" && currProject.value.containerType == null) return false
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
   if (a && a.assignedAt)  {
      return dayjs(a.assignedAt).format("YYYY-MM-DD hh:mm A")
   }
   return ""
})
const startedAt = computed(()=>{
   let stepID = currProject.value.currentStep.id
   let a = currProject.value.assignments.find( a => a.stepID == stepID)
   if ( a && a.startedAt) return dayjs(a.startedAt).format("YYYY-MM-DD hh:mm A")
   return ""
})
const workflowNote = computed(()=>{
   if ( currStepName.value == "Scan" && currProject.value.workstation.id == 0) {
      return "Assignment cannot be finished until the workstation has been set."
   }
   if ( currProject.value.workflow.name == "Manuscript" && currStepName.value == "Scan" && currProject.value.containerType == null) {
      return "Assignment cannot be finished until container type is set."
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

function changeWorkflowClicked() {
   selectedWorkflowIdx.value = -1
   selectedContainerTypeIdx.value = -1
   showWorkflowPicker.value = true
}
function cancelWorkflowChange() {
   showWorkflowPicker.value = false
}
async function submitWorkflowChange()  {
   let workflowID = activeWorkflows.value[selectedWorkflowIdx.value].id
   let containerTypeID = 0
   if ( isManuscriptSelected.value ) {
      containerTypeID = systemStore.containerTypes[selectedContainerTypeIdx.value].id
   }
   await projectStore.changeWorkflow( workflowID, containerTypeID )
   showWorkflowPicker.value = false
}
function selectWorkflow( idx ) {
   selectedWorkflowIdx.value = idx
   selectedContainerTypeIdx.value = -1
}
function selectContainerType( idx ) {
   selectedContainerTypeIdx.value = idx
}

function clearClicked() {
   projectStore.assignProject({projectID: currProject.value.id, ownerID: 0} )
}

function rejectStepClicked() {
   action.value = "reject"
   showTimeEntry()
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
   let threshold = 100
   if (currProject.value.currentStep.name == "Scan" || currProject.value.currentStep.name == "First QA") {
      threshold = 500
   }
   if (stepMinutes.value == 0 ) {
      systemStore.setError("A non-zero duration is required")
      return
   }
   if (stepMinutes.value > threshold ) {
      confirm.require({
         message: `The duration entered (${stepMinutes.value} minutes) is very large. Are you sure?`,
         header: 'Confirm Duration',
         rejectClass: 'p-button-secondary',
         accept: () => {
            timeEnterSuccess()
         }
      })
   } else {
      timeEnterSuccess()
   }
}

const timeEnterSuccess = (() => {
   if ( action.value == "finish")  {
      projectStore.finishStep(stepMinutes.value)
      timeEntry.value = false
      stepMinutes.value = 0
   } else {
      showRejectNote.value = true
      timeEntry.value = false
   }
})

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
   width: 100%;
   box-sizing: border-box;
   margin: 15px 0;
   display: inline-block;
   min-height: 100px;
   text-align: left;

   .pad-right {
      margin-right: 10px;
   }
   .workflow-btns {
      padding: 0;
      margin-top: 10px;
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      button.p-button {
         margin-left: 10px;
      }
      .right-buttons {
         margin-left: auto;
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
      padding: 15px 0 0 0;
      margin-top: 15px;
      border-top: 1px solid var(--uvalib-grey-light);
      text-align: center;
      color: var(--uvalib-red-emergency);
   }
}
div.workflow-picker {
   p {
      margin: 0 0 10px 0;
      color: #999;
   }
   .workflow-list {
      margin: 0 0 15px 0;
      background: white;
      border: 1px solid var(--uvalib-grey-light);
      padding: 5px 0;
      .workflow-val {
         padding: 2px 10px 3px 10px;
         cursor: pointer;
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         &:hover  {
            background: var(--uvalib-blue-alt-light);
         }
      }
      .workflow-val.selected {
         background: var(--uvalib-blue-alt);
         color: white;
      }
   }
}
</style>
