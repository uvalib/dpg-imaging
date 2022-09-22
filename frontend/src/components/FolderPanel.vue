<template>
   <div class="folder panel">
      <h3>Folder Information</h3>
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
            <label>Folder:</label>
            <input id="folder-id" type="text" v-model="folder"  @keyup.enter="okClicked"/>
         </span>
      </div>
      <div class="panel-actions">
         <DPGButton @click="selectAllClicked" class="left">Select All</DPGButton>
         <DPGButton @click="cancelEditClicked" class="right-pad">Cancel</DPGButton>
         <DPGButton @click="okClicked">OK</DPGButton>
      </div>
   </div>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import {useSystemStore} from "@/stores/system"
import { ref } from 'vue'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const folder = ref("")

async function okClicked() {
   await unitStore.setLocation( "folder", folder.value )
   cancelEditClicked()
}
function cancelEditClicked() {
   systemStore.error = ""
   unitStore.editMode = ""
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
</style>
