<template>
   <div class="component panel">
      <h3>Component Link</h3>
      <div class="content">
         <span class="entry pad-right">
            <label>Start Image:</label>
            <select id="start-page" v-model="unitStore.rangeStartIdx">
               <option disabled :value="-1">Select start page</option>
               <option v-for="(mf,idx) in unitStore.masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
         <span class="entry pad-right">
            <label>End Image:</label>
            <select id="end-page" v-model="unitStore.rangeEndIdx">
               <option disabled :value="-1">Select end page</option>
               <option v-for="(mf,idx) in unitStore.masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
         <span class="entry">
            <label>Component ID:</label>
            <input id="component-id" type="text" v-model="componentID"  @keyup.enter="okClicked"/>
         </span>
      </div>
      <div class="panel-actions">
         <DPGButton2 @click="selectAllClicked" class="p-button-secondary left">Select All</DPGButton2>
         <DPGButton2 @click="unlinkClicked" class="p-button-secondary right-pad">Unlink</DPGButton2>
         <DPGButton2 @click="cancelEditClicked" class="p-button-secondary right-pad">Cancel</DPGButton2>
         <DPGButton2 @click="okClicked">OK</DPGButton2>
      </div>
      <Dialog v-model:visible="unitStore.component.valid" :modal="true" header="Confirm Component Link">
         <div class="component-modal-content">
            <table>
               <tr>
                  <td class="label">Title:</td>
                  <td class="data">{{formatData(unitStore.component.title)}}</td>
               </tr>
               <tr>
                  <td class="label">Label:</td>
                  <td class="data">{{formatData(unitStore.component.label)}}</td>
               </tr>
               <tr>
                  <td class="label">Description:</td>
                  <td class="data">{{formatData(unitStore.component.description)}}</td>
               </tr>
               <tr>
                  <td class="label">Date:</td>
                  <td class="data">{{formatData(unitStore.component.date)}}</td>
               </tr>
               <tr>
                  <td class="label">Type:</td>
                  <td class="data">{{formatData(unitStore.component.type)}}</td>
               </tr>
            </table>
            <p class="confirm">Link this component to selected images?</p>
         </div>
         <template #footer>
            <DPGButton2 class="p-button-secondary" @click="noLinkClicked" label="No"/>
            <span class="spacer"></span>
            <DPGButton2 @click="linkConfirmed" label="Yes"/>
         </template>
      </Dialog>
   </div>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import {useSystemStore} from "@/stores/system"
import { ref, onMounted, nextTick } from 'vue'
import Dialog from 'primevue/dialog'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const componentID = ref("")

onMounted( async () => {
   nextTick( () => {
      let ele = document.getElementById("start-page")
      ele.focus()
   })
})

function formatData( value ) {
   if (value && value != "" )  return value
   return "N/A"
}
function okClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.error = "Start and end image must be selected"
      return
   }
   if (componentID.value == "") {
      systemStore.error = "Component ID is required"
      return
   }
   unitStore.lookupComponentID(componentID.value)
}
function noLinkClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
}
function cancelEditClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
   unitStore.editMode = ""
}
async function unlinkClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.error = "Start and end image must be selected"
      return
   }
   await unitStore.componentLink("")
   cancelEditClicked()
}
async function linkConfirmed() {
   await unitStore.componentLink(componentID.value)
   cancelEditClicked()
}
function selectAllClicked() {
   unitStore.selectAll()
}
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   border-bottom: 1px solid var(--uvalib-grey);
   h3 {
      margin: 0;
      padding: 8px 0;
      font-size: 1em;
      background: var(--uvalib-blue-alt-light);
      border-bottom: 1px solid var(--uvalib-grey);
      font-weight: 500;
   }
   .panel-actions {
      padding: 5px 0 20px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
      width: 50%;
      margin: 0 auto;
      .button {
         margin-left: 10px;
      }
      .left {
         margin-right: auto;
      }
   }
   .content {
      padding: 10px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      width: 50%;
      margin: 0 auto;
      .entry {
         flex-grow: 1;
         margin: 0;
         text-align: left;
         label {
            display: block;
            margin: 0 0 5px 0;
         }
      }
      .entry.pad-right {
         padding-right: 25px;
      }
   }
}
:deep(table) {
   font-size: 0.9em;
   padding: 15px 15px 0 15px;
   td {
      padding: 3px;
      text-align: left;
   }
   td.label {
      font-weight: bold;
      text-align: right;
   }
}
:deep(p.confirm) {
   text-align: right;
}
</style>
