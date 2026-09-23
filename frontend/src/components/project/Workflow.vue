<template>
   <UAccordion :items="[{label: 'Workflow', value: 'workflow'}]" defaultValue="workflow" class="panel">
      <template #body="{ }">
         <div  v-if="isFinished" class="finished">This project was completed {{ projectFinishedAt }}<br/>Workflow: {{ detail.workflow.name }}</div>
         <dl v-else>
            <dt>Name:</dt>
            <dd>{{detail.workflow.name}}</dd>
            <template v-if="detail.currentStep">
               <dt>Step:</dt>
               <dd>{{detail.currentStep.description}}</dd>
            </template>
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
         <div class="finalizing" v-if="isFinalizeRunning" >
            <WaitSpinner :overlay="false" message="Finalization in progress..." />
         </div>
         <div class="workflow-btns" v-else-if="isFinished == false">
            <UButton @click="deleteProjectClicked" class="delete" color="error" v-if="isSupervisor || isAdmin" label="Delete Project"/>
            <UButton @click="viewerClicked" color="secondary" v-if="isScanning == false && (isOwner(user.computeID) || isSupervisor || isAdmin)" label="Open QA Viewer"/>
            <UButton v-if="hasOwner && (isAdmin || isSupervisor)"
               @click="clearClicked()" color="secondary" label="Clear Assignment"/>
            <template v-if="isOwner(user.computeID)">
               <template v-if="isWorking == false">
                  <AssignModal v-if="(isOwner(user.computeID) || isSupervisor || isAdmin)" :projectID="detail.id" label="Reassign"/>
                  <UButton v-if="inProgress == false" @click="project.startStep()" label="Start"/>
                  <FinishStepModal v-if="canReject" action="reject" />
                  <FinishStepModal v-if="inProgress == true" action="finish" />
               </template>
            </template>
            <template v-else>
               <UButton v-if="isWorking == false && (hasOwner == false || isAdmin ||isSupervisor)"
                  @click="claimClicked()" color="secondary" label="Claim"/>
               <AssignModal v-if="(isAdmin || isSupervisor)" :projectID="detail.id" />
            </template>
         </div>
         <div class="workflow-message" v-if="isOwner(user.computeID) && workflowNote">
            {{workflowNote}}
         </div>
      </template>
   </UAccordion>
</template>

<script setup>
import { useDateFormat } from '@vueuse/core'
import { useProjectStore } from "@/stores/project"
import { useSystemStore } from "@/stores/system"
import { useUserStore } from "@/stores/user"
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useConfirm } from "@/composables/useConfirm"
import FinishStepModal from '@/components/project/FinishStepModal.vue'
import AssignModal from "@/components/AssignModal.vue"

const router = useRouter()
const project = useProjectStore()
const system = useSystemStore()
const user = useUserStore()

const {
   detail, isOwner, hasOwner,
   isFinalizeRunning, isFinished, inProgress, isWorking, canReject,
} = storeToRefs(project)

const {isAdmin, isSupervisor} = storeToRefs(user)

const currStepName = computed(()=>{
   if ( detail.value.currentStep ) {
      return detail.value.currentStep.name
   }
   return "Unknown"
})

const isScanning = computed(()=>{
   return (currStepName.value == 'Scan' || currStepName.value == 'Process')
})

const workingDir = computed(()=>{
   let unitDir =  unitDirectory(detail.value.unitID)
   if (detail.value.currentStep && (detail.value.currentStep.name == "Process" || detail.value.currentStep.name == "Scan")) {
      return `${system.scanDir}/${unitDir}`
   }
   return `${system.qaDir}/${unitDir}`
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

const deleteProjectClicked = (async () => {
   let note = `<span style='font-weight:bold'>Important</span>: any images associated with this project will be left<br/>in the processing directory for unit ${detail.value.unitID}`
   let msg = `Delete project ${detail.value.id}? This cannot be reversed.<br/>${note}`
   const resp = await useConfirm("Confirm Delete", msg, "Delete")
   if (resp) {
      await project.deleteProject( detail.value.id )
      window.location.reload() 
   } 
})

const clearClicked = (() => {
   project.assignProject( detail.value.id, 0 )
})

const claimClicked = (() => {
   project.assignProject( detail.value.id, user.ID )
})

const viewerClicked = (() => {
   router.push(`/projects/${detail.value.id}/unit`)
})

const unitDirectory = ((unitID) => {
   let ud = ""+unitID
   return ud.padStart(9, "0")
})
</script>

<style scoped lang="scss">
.panel {
   text-align: left;
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
      padding-top: 15px;
      border-top: 1px solid var(--uvalib-grey-light);
      .delete {
         margin-right: auto;
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
