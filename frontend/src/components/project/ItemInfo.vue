<template>
   <UAccordion :items="[{label: 'Item Information', value: 'info'}]" defaultValue="info" class="panel">
      <template #body="{ }">
         <dl v-if="!editing">
            <template v-if="detail.workflow.name == 'Manuscript'">
               <dt>Container Type:</dt>
               <dd>
                  <span v-if="detail.containerType && detail.containerType.id > 0">{{detail.containerType.name}}</span>
                  <span v-else class="na">EMPTY</span>
               </dd>
            </template>
            <dt>Category:</dt>
            <dd>{{detail.category.name}}
            </dd>
            <dt>Call Number:</dt>
            <dd>
               <span v-if="detail.callNumber">{{detail.callNumber}}</span>
               <span v-else class="na">EMPTY</span>
            </dd>
            <dt>Special Instructions:</dt>
            <dd>
               <span v-if="detail.specialInstructions">{{detail.specialInstructions}}</span>
               <span v-else class="na">EMPTY</span>
            </dd>
            <dt>Condition:</dt>
            <dd>{{conditionText(detail.itemCondition)}}</dd>
            <dt>Condition Notes:</dt>
            <dd>
               <span v-if="detail.conditionNote">{{detail.conditionNote}}</span>
               <span v-else class="na">EMPTY</span>
            </dd>
            <dt>OCR Hint:</dt>
            <dd>
               <span v-if="detail.ocrHintID > 0">{{ systemStore.getOCRHint(detail.ocrHintID ) }}</span>
               <span v-else class="na">EMPTY</span>
            </dd>
            <dt>OCR Language Hint:</dt>
            <dd>
               <span v-if="detail.ocrLanguage">{{ systemStore.getOCRLanguageHint(detail.ocrLanguage) }}</span>
               <span v-else class="na">EMPTY</span>
            </dd>
            <dt>OCR Master Files:</dt>
            <dd>
               <span v-if="detail.ocrMasterFiles" class="yes-no">Yes</span>
               <span v-else class="yes-no">No</span>
            </dd>
         </dl>
         <table class="edit" v-else>
            <tbody>
               <tr v-if="detail.workflow.name == 'Manuscript'">
                  <td class="label"><label for="container">Container Type:</label></td>
                  <td class="data">
                     <USelect id="container" v-model="containerTypeID" :items="systemStore.containerTypes" 
                        value-key="id" label-key="name" class="w-full"/>
                  </td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="category">Category:</label></td>
                  <td class="data">
                     <USelect id="category" v-model="categoryID" :items="systemStore.categories" 
                        value-key="id" label-key="name" class="w-full"/>
                  </td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="call-numbber">Call Number:</label></td>
                  <td class="data">{{detail.callNumber}}</td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="instructions">Special Instructions:</label></td>
                  <td class="data">
                     <span v-if="detail.specialInstructions">{{detail.specialInstructions}}</span>
                     <span v-else class="na">EMPTY</span>
                  </td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="condition">Condition:</label></td>
                  <td class="data">
                     <USelect id="condition" v-model="condition" :items="conditions" class="w-full"/>
                  </td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="notes">Condition Notes:</label></td>
                  <td class="data"><UTextarea id="notes" v-model="note" class="w-full"/></td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="ocr-hint">OCR Hint:</label></td>
                  <td class="data">
                     <USelect id="ocr-hint" v-model="ocrHintID" :items="systemStore.ocrHints" @change="hintChanged" 
                        value-key="id" label-key="name" class="w-full" placeholder="Select an OCR hint"/>
                  </td>

               </tr>
               <tr class="row">
                  <td class="label"><label :class="{disabled: !ocrCandidate}" for="ocr-language">OCR Language Hint:</label></td>
                  <td class="data">
                     <USelectMenu id="ocr-language" :items=" systemStore.ocrLanguageHints" v-model="ocrLangage" :disabled="!ocrCandidate"
                        value-key="code" label-key="language" class="w-full" placeholder="Select an OCR language hint" virtualize />
                  </td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="do-ocr" :class="{disabled: !ocrCandidate}">OCR Master Files:</label></td>
                  <td class="data"><UCheckbox size="lg" id="do-ocr" v-model="ocrMasterFiles" :disabled="!ocrCandidate"/></td>
               </tr>
            </tbody>
         </table>
         <div class="buttons" v-if="canEdit">
            <UButton v-if="!editing" @click="editClicked" color="secondary" label="Edit"/>
            <template v-else>
               <UButton @click="cancelClicked" color="secondary" label="Cancel"/>
               <UButton @click="saveClicked" label="Save"/>
            </template>
         </div>
      </template>
   </UAccordion>
</template>

<script setup>
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import { ref, computed, nextTick } from 'vue'
import { storeToRefs } from 'pinia'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const { detail } = storeToRefs(projectStore)

const conditions = [{label: "Good", value: 0}, {label: "Bad", value: 1}]

const editing = ref(false)
const categoryID = ref(0)
const containerTypeID = ref(0)
const condition = ref(0)
const note = ref("")
const ocrHintID = ref(0)
const ocrLangage = ref("")
const ocrMasterFiles = ref(false)
const ocrCandidate = ref(true)

const canEdit = computed(() => {
   if ( projectStore.detail.workflow.name == 'Vendor' ) return false
   if ( projectStore.isOwner(userStore.computeID) == false ) return false
   if ( projectStore.isFinalizeRunning || projectStore.isFinished || projectStore.isWorking ) {
      return false
   }
   return true
})


function hintChanged() {
   ocrLangage.value = ""
   let hint = systemStore.ocrHints.find(h => h.id == ocrHintID.value)
   ocrCandidate.value =  hint.ocrCandidate
   if ( !ocrCandidate.value) {
      ocrMasterFiles.value = false
   }
}

function conditionText(condID) {
   if (condID == 0) return "Good"
   return "Bad"
}

function editClicked() {
   editing.value = true
   categoryID.value = detail.value.category.id
   containerTypeID.value = 0
   if ( detail.value.containerType ) {
      containerTypeID.value = detail.value.containerType.id
   }
   condition.value = detail.value.itemCondition
   note.value = detail.value.conditionNote
   ocrHintID.value = detail.value.ocrHintID
   ocrLangage.value = detail.value.ocrLanguage
   if (ocrHintID.value != 1) {
      ocrLangage.value = ""
   }
}

function cancelClicked() {
   editing.value = false
}

async function saveClicked() {
   let data = {
      containerTypeID: containerTypeID.value,
      categoryID: categoryID.value,
      condition: condition.value,
      note: note.value,
      ocrHintID: ocrHintID.value,
      ocrLangage: ocrLangage.value,
      ocrMasterFiles: ocrMasterFiles.value
   }
   await projectStore.updateProject(data)
   editing.value = false
}
</script>

<style scoped lang="scss">
.panel {
   text-align: left;
   
   dl {
      font-size: 1em !important;
   }

   .edit {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 5px;
      td {
          padding: 5px 0px 5px 10px;
      }
      td.data {
         width: 100%;
         input, select {
            border-color: var(--uvalib-grey-light);
         }
         input[type=checkbox] {
            width: 15px;
            height: 15px;
         }
      }
      td.label {
         font-weight: bold;
         margin-right: 10px;
         text-align: right;
         white-space: nowrap;
         vertical-align: top;
      }
      .disabled {
         opacity: 0.5;
      }
   }
   .buttons {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 10px;
   }
}
</style>
