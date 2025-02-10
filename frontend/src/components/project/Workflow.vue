<template>
   <ConfirmDialog position="top"/>
   <Panel header="Workflow" class="panel">
      <dl>
         <dt>Name:</dt>
         <dd>{{detail.workflow.name}}</dd>
         <dt>Step:</dt>
         <dd>{{detail.currentStep.description}}</dd>
         <dt>Owner:</dt>
         <dd>
            <span v-if="hasOwner">{{detail.owner.firstName}} {{detail.owner.lastName}}</span>
            <span v-else class="na">Unassigned</span>
         </dd>
         <dt>Assigned:</dt>
         <dd>
            <span v-if="hasOwner">{{assignedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Started:</dt>
         <dd>
            <span v-if="hasOwner && startedAt">{{startedAt}}</span>
            <span v-else class="na">N/A</span>
         </dd>
         <dt>Directory:</dt>
         <dd>{{workingDir}}</dd>
      </dl>
      <div class="workflow-btns time" v-if="timeEntry">
         <div v-if="validateComponents" class="validate">
            <div>Validating component settings...</div>
            <ProgressSpinner style="width: 40px; height: 40px" strokeWidth="5" />
         </div>
         <div v-else class="time-form" >
            <template v-if="isManuscript && detail.currentStep.name == 'Create Metadata'">
               <div  class="finish-info">
                  <label>Does this unit have components?</label>
                  <Select v-model="hasComponents" :options="['Yes', 'No']" placeholder="Yes or no?" @update:modelValue="componentChanged" />
               </div>
               <div class="sep"></div>
            </template>
            <div class="finish-info right">
               <label for="time">Approximately how many minutes did you spend on this assignment?</label>
               <div class="time-controls">
                  <InputNumber v-model="stepMinutes" inputId="time" :min="1" :max="500" />
                  <DPGButton @click="cancelFinish" severity="secondary" label="Cancel"/>
                  <DPGButton @click="timeEntered" label="OK" :disabled="isManuscript && hasComponents == null && detail.currentStep.name == 'Create Metadata'"/>
               </div>
            </div>
         </div>
      </div>
      <div class="finalizing" v-else-if="isFinalizeRunning" >
         <WaitSpinner :overlay="false" message="Finalization in progress..." />
      </div>
      <div class="workflow-btns" v-else-if="isFinished == false">
         <DPGButton @click="changeWorkflowClicked()" severity="secondary" v-if="projectStore.canChangeWorkflow &&  (isSupervisor || isAdmin)" label="Change Workflow"/>
         <DPGButton @click="viewerClicked" severity="secondary" v-if="isScanning == false && (isOwner(userStore.computeID) || isSupervisor || isAdmin)" label="Open QA Viewer"/>
         <DPGButton v-if="hasOwner && (isAdmin || isSupervisor)"
            @click="clearClicked()" severity="secondary" label="Clear Assignment"/>
         <template v-if="isOwner(userStore.computeID)">
            <template v-if="isWorking == false">
               <AssignModal v-if="(isOwner(userStore.computeID) || isSupervisor || isAdmin)" :projectID="detail.id" label="Reassign"/>
               <DPGButton v-if="inProgress == false" @click="startStep" label="Start"/>
               <DPGButton v-if="canReject" class="p-button-danger" @click="rejectStepClicked" label="Reject"/>
               <DPGButton v-if="inProgress == true" :disabled="!isFinishEnabled" @click="finishClicked">
                  <template v-if="isFinalizing &&  hasError == true">Retry Finalize</template>
                  <template v-else>Finish</template>
               </DPGButton>
            </template>
         </template>
         <template v-else>
            <DPGButton v-if="isWorking == false && (hasOwner == false || isAdmin ||isSupervisor)"
               @click="claimClicked()"  severity="secondary" label="Claim"/>
            <AssignModal v-if="(isAdmin || isSupervisor)" :projectID="detail.id" />
         </template>
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
         <p>Current workflow: {{ detail.workflow.name }}</p>
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
         <DPGButton @click="cancelWorkflowChange()" label="Cancel" severity="secondary"/>
         <span class="spacer"></span>
         <DPGButton @click="submitWorkflowChange()" label="Submit" :disabled="isWorkflowChangeDisabled"/>
      </template>
   </Dialog>
</template>

<script setup>
import { useDateFormat } from '@vueuse/core'
import AssignModal from "@/components/AssignModal.vue"
import NoteModal from '@/components/project/NoteModal.vue'
import { useProjectStore } from "@/stores/project"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { ref, computed, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import Select from 'primevue/select'
import Panel from 'primevue/panel'
import InputNumber from 'primevue/inputnumber'
import Dialog from 'primevue/dialog'
import ProgressSpinner from 'primevue/progressspinner'
import { useFocus } from '@vueuse/core'

const router = useRouter()
const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const {
   detail, isOwner, hasOwner, hasError,
   isFinalizeRunning, isFinished, inProgress, isWorking, canReject,
} = storeToRefs(projectStore)
const {isAdmin, isSupervisor} = storeToRefs(userStore)

const hasComponents = ref(null)
const validateComponents = ref(false)

const timeEntry = ref(false)
const time = ref()
const { focused: timeFocus } = useFocus(time)
const stepMinutes = ref(1)
const action = ref("finish")
const showRejectNote = ref(false)
const showWorkflowPicker = ref(false)
const selectedWorkflowIdx = ref(-1)
const selectedContainerTypeIdx = ref(-1)

const isManuscript = computed(() => {
   return detail.value.workflow.name == "Manuscript"
})
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
   return detail.value.currentStep.name
})
const isFinalizing = computed(()=>{
   return currStepName.value == 'Finalize'
})
const isScanning = computed(()=>{
   return (currStepName.value == 'Scan' || currStepName.value == 'Process')
})
const isFinishEnabled = computed(()=>{
   if ( detail.value.workflow.name == "Manuscript" && currStepName.value == "Scan" && detail.value.containerType == null) return false
   if ( currStepName.value == "Scan" && detail.value.workstation.id == 0) return false
   if ( currStepName.value == "Finalize") {
      if ( detail.value.unit.metadata.ocrHint.id == 0) return false
      if ( detail.value.unit.metadata.ocrHint.id == 1 && detail.value.unit.metadata.ocrLanguageHint == "") return false
   }
   return true
})
const workingDir = computed(()=>{
   let unitDir =  unitDirectory(detail.value.unit.id)
   if (detail.value.currentStep.name == "Process" || detail.value.currentStep.name == "Scan") {
      return `${systemStore.scanDir}/${unitDir}`
   }
   return `${systemStore.qaDir}/${unitDir}`
})
const assignedAt = computed(()=>{
   let currA = detail.value.assignments[0]
   if ( currA ) {
      return useDateFormat(currA.assignedAt, "YYYY-MM-DD hh:mm A")
   }
   return ""
})
const startedAt = computed(()=>{
   let currA = detail.value.assignments[0]
   if ( currA && currA.startedAt ) {
      return useDateFormat(currA.startedAt, "YYYY-MM-DD hh:mm A")
   }
   return ""
})
const workflowNote = computed(()=>{
   if ( currStepName.value == "Scan" && detail.value.workstation.id == 0) {
      return "Assignment cannot be finished until the workstation has been set."
   }
   if ( detail.value.workflow.name == "Manuscript" && currStepName.value == "Scan" && detail.value.containerType == null) {
      return "Assignment cannot be finished until container type is set."
   }
   if ( currStepName.value == "Finalize" && detail.value.unit.metadata.ocrHint.id == 0) {
      return "Assignment cannot be finished until the OCR hint has been set."
   }
   if ( detail.value.unit.metadata.ocrHint.id > 1 && detail.value.unit.ocrMasterFiles == true) {
      return "Cannot OCR items that are not regular text."
   }
   if ( detail.value.unit.metadata.ocrHint.id == 1 && detail.value.unit.metadata.ocrLanguageHint == "" && currStepName.value == "Finalize") {
      return "Assignment cannot be finished until the OCR Language Hint has been set."
   }
   if ( detail.value.unit.status == "error" ) {
      return "Finalization has failed. Correct the problem then click 'Retry Finalization'."
   }
   return ""
})

const componentChanged = ( async ()=> {
   if ( hasComponents.value == "Yes" ) {
      validateComponents.value = true
      await projectStore.validateComponents()
      validateComponents.value = false
      if ( projectStore.hasMissingComponents == true ) {
         systemStore.setError("Some images are missing component data. Please correct the problem before finishing this step.")
         cancelFinish()
         var msg = "The following images are missing component data: "
         msg += projectStore.missingComponents.join(", ")
         let data = {noteTypeID: 2, note: msg, problemIDs: [4]}
         projectStore.addNote(data)
      }
   }
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
   projectStore.assignProject({projectID: detail.value.id, ownerID: 0} )
}

function rejectStepClicked() {
   action.value = "reject"
   showTimeEntry()
}

function claimClicked() {
   projectStore.assignProject({projectID: detail.value.id, ownerID: userStore.ID} )
}

function finishClicked() {
   action.value = "finish"
   if ( detail.value.assignments[0].durationMinutes == 0) {
      showTimeEntry()
   } else {
      // send a 0 time to indicate that time has already been recorded
      projectStore.finishStep(0)
   }
}

function showTimeEntry() {
   timeEntry.value = true
   stepMinutes.value = 1
   nextTick( ()=> timeFocus.value = true )
}

function timeEntered() {
   timeEnterSuccess( intDuration )
}

const timeEnterSuccess = ((intDuration) => {
   if ( action.value == "finish")  {
      projectStore.finishStep( intDuration )
      timeEntry.value = false
      stepMinutes.value = 1
   } else {
      showRejectNote.value = true
      timeEntry.value = false
   }
})

function rejectCanceled() {
   showRejectNote.value = false
   timeEntry.value = false
   stepMinutes.value = 1
}

function rejectSubmitted() {
   let intDuration = parseInt(stepMinutes.value, 10)
   if ( isNaN(intDuration)) {
      systemStore.setError("Please enter a number")
      return
   }
   if (intDuration <= 0 ) {
      systemStore.setError("A non-zero duration is required")
      return
   }
   projectStore.rejectStep( intDuration )
   showRejectNote.value = false
   timeEntry.value = false
   stepMinutes.value = 1
}

function cancelFinish() {
   timeEntry.value = false
   hasComponents.value = null
}

function viewerClicked() {
   router.push(`/projects/${detail.value.id}/unit`)
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
   .finalizing {
      text-align: center;
      font-size: 0.75em;
   }
   .workflow-btns {
      padding: 0;
      margin-top: 10px;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      align-items: flex-start;
      gap: 10px;
   }
   .workflow-btns.time {
      text-align: left;
      .validate {
         border-top: 1px solid #e2e8f0;
         display: flex;
         flex-direction: column;
         align-items: center;
         gap: 1rem;
         font-size: 1.25rem;
         width: 100%;
         padding-top: 15px;
      }
      .time-form {
         width: 100%;
         text-align: right;
         margin-bottom: 10px;
         font-size: 0.9em;
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-end;
         border-top: 1px solid #e2e8f0;
         .sep {
            border-left: 1px solid #e2e8f0;
         }
         .finish-info {
            flex-basis: 100%;
            display: flex;
            flex-direction: column;
            align-items: flex-start;
            label {
               display: block;
               font-weight: bold;
               margin: 15px 0;
               max-width: 80%;
            }
            .time-controls {
               display: flex;
               flex-flow: row nowrap;
               justify-content: flex-end;
               gap: 10px;
               input {
                  width: 100px;
                  border-color: var(--uvalib-grey-light);
               }
            }
         }
         .finish-info.right {
            align-items: flex-end;
         }
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
