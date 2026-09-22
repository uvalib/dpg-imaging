<template>
   <UModal v-model:open="isOpen" :modal="true" :dismissible="false" :close="false" :title=title >
      <UButton @click="show()" :label="buttonLabel" :disabled="!isOpenEnabled" :color="buttonColor"/>
      <template #body>
         <div class="content">
            <div v-if="validateComponents" class="validate">
               <WaitSpinner :overlay="false" message="Validating component settings..." />   
            </div>
            <template v-else>
               <template v-if="props.action=='finish' && isManuscript && (currStepName == 'Create Metadata' || currStepName == 'Finalize')">
                  <div class="row">
                     <label>Does this unit have components?</label>
                     <USelect v-model="hasComponents" :items="['Yes', 'No']" placeholder="Yes or no?" />
                  </div>
                  <div class="row" v-if="project.detail.containerType.hasFolders">
                     <label>Does this unit use folders?</label>
                     <USelect v-model="hasFolders" :items="['Yes', 'No']" placeholder="Yes or no?" />
                  </div>
               </template>
               <template v-if="props.action=='reject'">
                  <div class="reject-note">
                     Rejection requires the addition of a problem note that details the reason why it occurred.
                  </div>
                   <div class="row">
                     <label>Problem (select all that apply)</label>
                     <UCheckboxGroup v-model="problemIDs" :items="system.problemTypes" value-key="id" />
                  </div>
                  <div class="row">
                     <label for="note-text">Note</label>
                     <UTextarea v-model="note"/>
                  </div>
               </template>
               <div class="row">
                  <label for="time">Approximately how many minutes did you spend on this assignment?</label>
                  <UInputNumber v-model="stepMinutes" id="time" :min="1" :max="500" />
               </div>
            </template>
            <p class="error" v-if="error">{{error}}</p>
         </div>
      </template>
      <template #footer>
         <UButton @click="hide()" label="Cancel" color="secondary"/>
         <UButton @click="okClicked()" label="OK"/>
      </template>
   </UModal>
</template>

<script setup>
import { ref, computed } from 'vue'
import {useProjectStore} from '@/stores/project'
import {useSystemStore} from "@/stores/system"

const project = useProjectStore()
const system = useSystemStore()

const props = defineProps({
      action: {
         type: String,
         required: true,    
      },
   })

const isOpen = ref(false)
const hasComponents = ref(null)
const hasFolders = ref(null)
const validateComponents = ref(false)
const stepMinutes = ref(1)
const problemIDs = ref([])
const note = ref("")
const error = ref("")

const buttonColor = computed(() => {
   if (props.action == "reject") return "error"
   return "primary"
})

const buttonLabel = computed(() => {
   if ( props.action == "reject")  return "Reject"

   if ( currStepName.value == "Finailze" && project.hasError == true) return "Retry Finalize"
   return "Finish"
})

const isOpenEnabled = computed(()=> {
   if ( props.action == "reject")  return true 

   // finsh scan step has special requirements:
   if ( currStepName.value == "Scan" ) {
      // manuscripts must have container type set
      if ( project.detail.workflow.name == "Manuscript" && project.detail.containerType == null) return false

      // all others must have workstation set
      if ( project.detail.workstation.id == 0) return false
   }
   
   // finalize must have OCR info set
   if ( currStepName.value == "Finalize") {
      if ( project.detail.ocrHintID == 0) return false
      if ( project.detail.ocrHintID == 1 && project.detail.ocrLanguage == "") return false
   }

   return true
})

const title = computed(()=>{
   if ( props.action == "reject") return `Reject ${currStepName.value}`
   return `Finish ${currStepName.value}`
})

const currStepName = computed(()=>{
   if ( project.detail.currentStep ) {
      return project.detail.currentStep.name
   }
   return "Unknown"
})

const isManuscript = computed(() => {
   return project.detail.workflow.name == "Manuscript"
})

const okClicked = ( async () => {
   error.value = ""
   let checkFolders = false
   if (props.action == 'reject') {
      if ( problemIDs.value.length == 0) {
         error.value = "At least one problem is required"
         return
      }
      if ( note.value == "") {
         error.value = "A problem note is required"
         return
      }
      let data = {noteTypeID: 2, note: note.value, problemIDs: problemIDs.value}
      await project.addNote(data)
      await project.rejectStep( stepMinutes.value )
   } else {
      if ( isManuscript.value && (currStepName.value == 'Create Metadata' || currStepName.value == 'Finalize') ) {
         if ( hasComponents.value == null ) {
            error.value = "A response to the components question is required."
            return
         }
         if ( hasFolders.value == null ) {
            error.value = "A response to the folders question is required."
            return
         }

         checkFolders = (hasFolders.value == "Yes")

         if ( hasComponents.value == "Yes") {
            validateComponents.value = true
            await project.validateComponents()
            validateComponents.value = false  

            if ( project.hasMissingComponents == true ) {
               var msg = "The following images are missing component data: "
               msg += project.missingComponents.join(", ")
               let data = {noteTypeID: 2, note: msg, problemIDs: [4]}
               await project.addNote(data)
               hide()

               // after the dialog is closed, show a system error
               system.setError("Some images are missing component data. Please correct the problem before finishing this step.")
               return
            }  
         } 
      }
      // all ok, finish the step and close the modal
      await project.finishStep( stepMinutes.value, checkFolders )
   }

   hide()
})

const hide = (() => {
   isOpen.value=false
})

const show = ( () => {
    if ( props.action == "finish" && project.detail.assignments[0].durationMinutes > 0 ) {
      // this is a retry of a step and time has already already added. finish the step
      // with 0 time to indicate that time has already been recorded
      project.finishStep(0)
    } else {
      stepMinutes.value = 1
      hasComponents.value = null
      hasFolders.value = null
      validateComponents.value = false
      problemIDs.value = []
      note.value = ""
      error.value = ""
      if ( isManuscript.value && project.detail.containerType.hasFolders == false ) {
         hasFolders.value = "No"   
      }
      isOpen.value = true
   }
})
</script>

<style lang="scss" scoped>
.content, .row {
   display: flex;
   flex-direction: column;
   gap: 15px;
}
.row {
   gap: 5px;
}
.reject-note {
   padding: 10px;
   border: 1px solid var(--uvalib-blue-alt);
   background: var(--uvalib-blue-alt-light);
}
.error {
   padding: 0;
   margin: 0;
   text-align: center;
   color: var(--uvalib-red-emergency);
}
.validate {
   display: flex;
   flex-direction: column;
   align-items: center;
}
</style>
