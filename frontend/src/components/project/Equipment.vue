<template>
   <Panel header="Equipment" class="panel">
      <dl v-if="!editing">
         <dt>Workstation:</dt>
         <dd>
            <span v-if="detail.workstation.id > 0">{{detail.workstation.name}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <template v-if="detail.workstation.id > 0">
            <dt>Setup:</dt>
            <dd>
               <table>
                  <tr v-for="e in detail.equipment" :key="e.serialNumber">
                     <td>{{e.type}}</td>
                     <td>{{e.name}}</td>
                     <td>{{e.serialNumber}}</td>
                  </tr>
               </table>
            </dd>
         </template>
         <dt>Capture resolution:</dt>
         <dd>
            <span v-if="detail.captureResolution">{{detail.captureResolution}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Resized resolution:</dt>
         <dd>
            <span v-if="detail.resizedResolution">{{detail.resizedResolution}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Resolution note:</dt>
         <dd>
            <span v-if="detail.resolutionNote">{{detail.resolutionNote}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
      </dl>
      <table class="edit" v-else>
         <tbody>
            <tr class="row">
               <td class="label"><label for="workstation">Workstation:</label></td>
               <td class="data">
                  <select id="workstation" v-model="workstationID" ref="workstation">
                     <option disabled :value="0">Choose a workstation</option>
                     <option v-for="ws in systemStore.workstations" :key="`ws${ws.id}`" :value="ws.id">{{ws.name}}</option>
                  </select>
               </td>
            </tr>
            <tr class="row">
               <td class="label"><label for="capture">Capture Resolution:</label></td>
               <td><input id="capture" type="text" v-model="captureResolution"></td>
            </tr>
            <tr class="row">
               <td class="label"><label for="resize">Resized Resolution:</label></td>
               <td class="data"><input id="resize" type="text" v-model="resizedResolution"></td>
            </tr>
            <tr class="row">
               <td class="label"><label for="res-note">Resolution Note:</label></td>
               <td class="data"><textarea id="res-note" v-model="resolutionNote"></textarea></td>
            </tr>
         </tbody>
      </table>
      <div class="buttons" v-if="canEdit">
         <DPGButton v-if="!editing" @click="editClicked" severity="secondary" label="Edit"/>
         <template v-else>
            <DPGButton @click="cancelClicked" label="Cancel" severity="secondary"/>
            <DPGButton @click="saveClicked" label="Save"/>
         </template>
      </div>
   </Panel>
</template>

<script setup>
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import { ref, computed, nextTick } from 'vue'
import { storeToRefs } from 'pinia'
import Panel from 'primevue/panel'
import { useFocus } from '@vueuse/core'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()

const { detail } = storeToRefs(projectStore)

const editing = ref(false)
const workstation = ref()
const { focused: wsFocus } = useFocus(workstation)
const workstationID = ref(0)
const captureResolution = ref("")
const resizedResolution = ref("")
const resolutionNote = ref("")

const canEdit = computed(() => {
   if (projectStore.isOwner(userStore.computeID) == false) return false
   if (projectStore.isFinalizeRunning || projectStore.isFinished || projectStore.isWorking) {
      return false
   }
   return true
})

const editClicked = (() => {
   workstationID.value = detail.value.workstation.id
   captureResolution.value = ""
   resizedResolution.value = ""
   if ( detail.value.captureResolution) {
      captureResolution.value = detail.value.captureResolution
   }
   if (detail.value.resizedResolution) {
      resizedResolution.value = detail.value.resizedResolution
   }
   resolutionNote.value = detail.value.resolutionNote
   editing.value = true
   nextTick( ()=> wsFocus.value = true )
})

const cancelClicked =(() => {
   editing.value = false
})

const saveClicked = ( async () => {
   let data = {
      workstationID: workstationID.value,
      captureResolution: parseInt(captureResolution.value, 10),
      resizeResolution: parseInt(resizedResolution.value, 10),
      resolutionNote: resolutionNote.value
   }
   await projectStore.setEquipment(data)
   editing.value = false
})
</script>

<style scoped lang="scss">
.panel {
   width: 46%;
   min-width: 600px;
   margin: 15px;
   display: inline-block;
   min-height: 100px;
   text-align: left;

   .buttons {
      padding: 0;
      margin: 0;
      text-align: right;
      button {
         margin-left: 10px;
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
      }
      td.label {
         font-weight: bold;
         margin-right: 10px;
         text-align: right;
         vertical-align: top;
         white-space: nowrap;
      }
   }

   dd {
      table {
         width: 100%;
         font-size: 0.8em;
      }
   }
}
</style>
