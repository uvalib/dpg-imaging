<template>
   <UAccordion :items="[{label: 'Workflow', value: 'workflow'}]" defaultValue="workflow" class="panel">
      <template #body="{ }">
         <div  v-if="isFinished" class="finished">This project was completed {{ projectFinishedAt }}<br/>Workflow: {{ detail.workflow.name }}</div>
         <dl v-else>
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
         <template v-if="timeEntry">
            <div v-if="validateComponents" class="validate">
               <WaitSpinner :overlay="false" message="Validating component settings..." />   
            </div>
            <div v-else class="workflow-btns time">
               <div v-if="isManuscript && detail.currentStep.name == 'Create Metadata'" class="finish-info">
                  <label>Does this unit have components?</label>
                  <USelect v-model="hasComponents" :items="['Yes', 'No']" placeholder="Yes or no?" @update:modelValue="componentChanged" />
               </div>
               <div class="finish-info right">
                  <label for="time">Approximately how many minutes did you spend on this assignment?</label>
                  <div class="time-controls">
                     <UInputNumber v-model="stepMinutes" id="time" :min="1" :max="500" />
                     <UButton @click="cancelFinish" color="secondary" label="Cancel"/>
                     <UButton @click="timeEntered" label="OK" :disabled="isManuscript && hasComponents == null && detail.currentStep.name == 'Create Metadata'"/>
                  </div>
               </div>
            </div>
         </template>
         <div class="finalizing" v-else-if="isFinalizeRunning" >
            <WaitSpinner :overlay="false" message="Finalization in progress..." />
         </div>
         <div class="workflow-btns" v-else-if="isFinished == false">
            <UButton @click="deleteProjectClicked" class="delete" color="error" v-if="isSupervisor || isAdmin" label="Delete Project"/>
            <UButton @click="viewerClicked" color="secondary" v-if="isScanning == false && (isOwner(userStore.computeID) || isSupervisor || isAdmin)" label="Open QA Viewer"/>
            <UButton v-if="hasOwner && (isAdmin || isSupervisor)"
               @click="clearClicked()" color="secondary" label="Clear Assignment"/>
            <template v-if="isOwner(userStore.computeID)">
               <template v-if="isWorking == false">
                  <AssignModal v-if="(isOwner(userStore.computeID) || isSupervisor || isAdmin)" :projectID="detail.id" label="Reassign"/>
                  <UButton v-if="inProgress == false" @click="startStep" label="Start"/>
                  <UButton v-if="canReject" class="p-button-danger" @click="rejectStepClicked" label="Reject"/>
                  <UButton v-if="inProgress == true" :disabled="!isFinishEnabled" @click="finishClicked">
                     <template v-if="isFinalizing &&  hasError == true">Retry Finalize</template>
                     <template v-else>Finish</template>
                  </UButton>
               </template>
            </template>
            <template v-else>
               <UButton v-if="isWorking == false && (hasOwner == false || isAdmin ||isSupervisor)"
                  @click="claimClicked()" color="secondary" label="Claim"/>
               <AssignModal v-if="(isAdmin || isSupervisor)" :projectID="detail.id" />
            </template>
         </div>
         <div class="workflow-message" v-if="isOwner(userStore.computeID) && workflowNote">
            {{workflowNote}}
         </div>
         <NoteModal id="problem-modal" :manual="true" :trigger="showRejectNote" :noteType="2"
            @closed="rejectCanceled" @submitted="rejectSubmitted"
            instructions="Rejection requires the addition of a problem note that details the reason why it occurred" />
      </template>
   </UAccordion>
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
import { useFocus } from '@vueuse/core'
import { useConfirm } from "@/composables/useConfirm"

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

const isManuscript = computed(() => {
   return detail.value.workflow.name == "Manuscript"
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
      if ( detail.value.ocrHintID == 0) return false
      if ( detail.value.ocrHintID == 1 && detail.value.ocrLanguage == "") return false
   }
   return true
})
const workingDir = computed(()=>{
   let unitDir =  unitDirectory(detail.value.unitID)
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
const projectFinishedAt = computed(()=>{
   if ( detail.value.finishedAt ) {
      return useDateFormat( detail.value.finishedAt, "YYYY-MM-DD hh:mm A")
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
   if ( currStepName.value == "Finalize" && detail.value.ocrHintID == 0) {
      return "Assignment cannot be finished until the OCR hint has been set."
   }
   if ( currStepName.value == "Finalize" && detail.value.ocrHintID == 1 && detail.value.ocrLanguage == "") {
      return "Assignment cannot be finished until the OCR language has been set."
   }
   if ( detail.value.ocrHintID > 1 && detail.value.ocrMasterFiles == true) {
      return "Cannot OCR items that are not regular text."
   }
   if ( detail.value.ocrHintID == 1 && detail.value.ocrLanguageHint == "" && currStepName.value == "Finalize") {
      return "Assignment cannot be finished until the OCR Language Hint has been set."
   }
   if ( detail.status == "error" ) {
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

const deleteProjectClicked = (async () => {
   let note = `<span style='font-weight:bold'>Important</span>: any images associated with this project will be left<br/>in the processing directory for unit ${detail.value.unitID}`
   let msg = `Delete project ${detail.value.id}? This cannot be reversed.<br/>${note}`
   const resp = await useConfirm("Confirm Delete", msg, "Delete")
   if (resp) {
      await projectStore.deleteProject( p.id )
      window.location.reload() 
   } 
})

const clearClicked = (() => {
   projectStore.assignProject({projectID: detail.value.id, ownerID: 0} )
})

const rejectStepClicked = (() => {
   action.value = "reject"
   showTimeEntry()
})

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

const timeEntered = (() =>{
   if ( action.value == "finish")  {
      projectStore.finishStep( stepMinutes.value )
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
   text-align: left;

   .pad-right {
      margin-right: 10px;
   }
   .finalizing {
      text-align: center;
   }
   .workflow-btns {
      padding: 0;
      margin-top: 10px;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      align-items: flex-start;
      gap: 10px;
      .delete {
         margin-right: auto;
      }
   }
   .validate {
      border-top: 1px solid #e2e8f0;
      display: flex;
      flex-direction: column;
      align-items: center;
      padding-top: 15px;
   }
   .workflow-btns.time {
      text-align: left;
      width: 100%;
      margin-bottom: 10px;
      font-size: 0.9em;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 10px;

      .finish-info {
         flex-basis: 100%;
         display: flex;
         flex-direction: column;
         justify-content: flex-end;
         label {
            display: block;
            font-weight: bold;
            margin: 10px 0;
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
   div.finished {
      text-align: center;
      font-size: 1.1em;
      font-weight: bold;
      padding: 25px 0;
   }
   .workflow-message {
      padding: 15px 0 0 0;
      margin-top: 15px;
      border-top: 1px solid var(--uvalib-grey-light);
      text-align: center;
      color: var(--uvalib-red-emergency);
   }
}
</style>
