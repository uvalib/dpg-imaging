<template>
   <UAccordion :items="[{label: 'Equipment', value: 'equip'}]" defaultValue="equip" class="panel">
      <template #body="{ }">
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
                     <USelect id="workstation" v-model="workstationID" :items="systemStore.workstations" 
                        value-key="id" label-key="name" class="w-full" placeholder="Select a workstation"/>
                  </td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="capture">Capture Resolution:</label></td>
                  <td><UInput id="capture" v-model="captureResolution" class="w-full"/></td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="resize">Resized Resolution:</label></td>
                  <td class="data"><UInput id="resize" v-model="resizedResolution" class="w-full"/></td>
               </tr>
               <tr class="row">
                  <td class="label"><label for="res-note">Resolution Note:</label></td>
                  <td class="data"><UTextarea id="res-note" v-model="resolutionNote" class="w-full"/></td>
               </tr>
            </tbody>
         </table>
         <div class="buttons" v-if="canEdit">
            <UButton v-if="!editing" @click="editClicked" color="secondary" label="Edit"/>
            <template v-else>
               <UButton @click="cancelClicked" label="Cancel" color="secondary"/>
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
import { ref, computed } from 'vue'
import { storeToRefs } from 'pinia'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()

const { detail } = storeToRefs(projectStore)

const editing = ref(false)
const workstationID = ref(null)
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
   if (workstationID.value == 0) workstationID.value = null
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
   text-align: left;

   .buttons {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 10px;
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
      }
      td.label {
         font-weight: bold;
         margin-right: 10px;
         text-align: right;
         vertical-align: top;
         white-space: nowrap;
      }
   }

   dl {
      font-size: 1em !important;
   }
   dd {
      table {
         width: 100%;
      }
   }
}
</style>
