<template>
   <div class="panel">
      <dl v-if="!editing">
         <template v-if="currProject.workflow.name == 'Manuscript'">
            <dt>Container Type:</dt>
            <dd>
               <span v-if="currProject.containerType && currProject.containerType.id > 0">{{currProject.containerType.name}}</span>
               <span v-else class="na">EMPTY</span>
            </dd>
         </template>
         <dt>Category:</dt>
         <dd>{{currProject.category.name}}
         </dd>
         <dt>Call Number:</dt>
         <dd>
            <span v-if="currProject.unit.metadata.callNumber">{{currProject.unit.metadata.callNumber}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Special Instructions:</dt>
         <dd>
            <span v-if="currProject.unit.specialInstructions">{{currProject.unit.specialInstructions}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Condition:</dt>
         <dd>{{conditionText(currProject.itemCondition)}}</dd>
         <dt>Condition Notes:</dt>
         <dd>
            <span v-if="currProject.conditionNote">{{currProject.conditionNote}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>OCR Hint:</dt>
         <dd>
            <span v-if="currProject.unit.metadata.ocrHint.id > 0">{{currProject.unit.metadata.ocrHint.name}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>OCR Language Hint:</dt>
         <dd>
            <span v-if="currProject.unit.metadata.ocrLanguageHint">{{currProject.unit.metadata.ocrLanguageHint}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>OCR Master Files:</dt>
         <dd>
            <span v-if="currProject.unit.ocrMasterFiles" class="yes-no">Yes</span>
            <span v-else class="yes-no">No</span>
         </dd>
      </dl>
      <table class="edit" v-else>
         <tr v-if="currProject.workflow.name == 'Manuscript'">
            <td class="label"><label for="container">Container Type:</label></td>
            <td class="data">
               <select id="container" v-model="containerTypeID">
                  <option :value="0" disabled>Select a container type</option>
                  <option v-for="c in systemStore.containerTypes" :key="`container-${c.id}`" :value="c.id">{{c.name}}</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="category">Category:</label></td>
            <td class="data">
               <select id="category" v-model="categoryID">
                  <option :value="0" disabled>Select a category</option>
                  <option v-for="c in systemStore.categories" :key="`cat${c.id}`" :value="c.id">{{c.name}}</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="call-numbber">Call Number:</label></td>
            <td class="data">{{currProject.unit.metadata.callNumber}}</td>
         </tr>
         <tr class="row">
            <td class="label"><label for="instructions">Special Instructions:</label></td>
            <td class="data">
               <span v-if="currProject.unit.specialInstructions">{{currProject.unit.specialInstructions}}</span>
               <span v-else class="na">EMPTY</span>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="condition">Condition:</label></td>
            <td class="data">
               <select id="condition" v-model="condition">
                  <option :value="0">Good</option>
                  <option :value="1">Bad</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="notes">Condition Notes:</label></td>
            <td class="data"><textarea id="notes" v-model="note"></textarea></td>
         </tr>
         <tr class="row">
            <td class="label"><label for="ocr-hint">OCR Hint:</label></td>
            <td class="data">
               <select id="ocr-hint" v-model="ocrHintID" @change="hintChanged">
                  <option :value="0" disabled>Select an OCR hint</option>
                  <option v-for="h in ocrHints" :key="`ocr${h.id}`" :value="h.id">{{h.name}}</option>
               </select>
            </td>

         </tr>
         <tr class="row">
            <td class="label"><label :class="{disabled: !ocrCandidate}" for="ocr-language">OCR Language Hint:</label></td>
            <td class="data">
               <select id="ocr-language" v-model="ocrLangage" :class="{disabled: !ocrCandidate}" :disabled="!ocrCandidate">
                  <option value="" disabled>Select an OCR language hint</option>
                  <option v-for="h in ocrLanguageHints" :key="`lang${h.code}`" :value="h.code">{{h.language}}</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="do-ocr" :class="{disabled: !ocrCandidate}">OCR Master Files:</label></td>
            <td class="data"><input type="checkbox" :class="{disabled: !ocrCandidate}" id="do-ocr" v-model="ocrMasterFiles" :disabled="!ocrCandidate"></td>
         </tr>
      </table>
      <div class="buttons" v-if="projectStore.isOwner(userStore.computeID)">
         <DPGButton v-if="!editing" @clicked="editClicked">Edit</DPGButton>
         <template v-else>
            <DPGButton @clicked="cancelClicked">Cancel</DPGButton>
            <DPGButton @clicked="saveClicked">Save</DPGButton>
         </template>
      </div>
   </div>
</template>

<script setup>
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import { ref } from 'vue'
import { storeToRefs } from 'pinia'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const { currProject } = storeToRefs(projectStore)

const editing = ref(false)
const categoryID = ref(0)
const containerTypeID = ref(0)
const condition = ref(0)
const note = ref("")
const ocrHintID = ref(0)
const ocrLangage = ref("")
const ocrMasterFiles = ref(false)
const ocrCandidate = ref(true)

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
   categoryID.value = currProject.category.id
   containerTypeID.value = 0
   if ( currProject.containerType ) {
      containerTypeID.value = currProject.containerType.id
   }
   condition.value = currProject.itemCondition
   note.value = currProject.conditionNote
   ocrHintID.value = currProject.unit.metadata.ocrHint.id
   ocrLangage.value = currProject.unit.metadata.ocrLanguageHint
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
   padding: 10px;
   text-align: left;
   .na {
      color: #999;
   }
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
      }
   }
   .edit {
      font-size: 0.9em;
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
      textarea {
         width: 100%;
         box-sizing: border-box;
         border-color: var(--uvalib-grey-light);
         border-radius: 5px;;
      }
   }
   .buttons {
      padding: 0;
      margin: 0;
      text-align: right;
      button {
         margin-left: 10px;
      }
   }
}
</style>
