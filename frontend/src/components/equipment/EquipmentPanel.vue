<template>
   <UTable :data="props.equipment" :columns="cols" class="h-full pb-8" virtualize v-model:column-visibility="colVisibility"
   >
      <template #select-cell="{ row }">
         <UCheckbox size="lg" :value="row.original.id" :modelValue="isItemInUse(row.original.id)" 
            :disabled="isItemDisabled(row.original.id)" @click="equipmentClicked(row.original.id)"
         />  
      </template>
      <template #name-cell="{ row }">
         <UInput v-if="editInfo.rowIdx == row.index" v-model="newName" class="w-full"/>
         <span v-else>{{ row.original.name }}</span>
      </template>
      <template #serialNumber-cell="{ row }">
         <UInput v-if="editInfo.rowIdx == row.index" v-model="newSerial" class="w-full"/>
         <span v-else>{{ row.original.serialNumber }}</span>
      </template>
      <template #workstation-cell="{ row }">
         <span>{{workstation(row.original.id)}}</span>
      </template>
      <template #acts-cell="{ row }">
         <div class="row-acts">
            <template v-if="editInfo.rowIdx == row.index">
               <UButton label="Cancel" size="sm" color="secondary" @click="cancelEdit"/>
               <UButton label="Save" size="sm"  @click="saveChanges" />
            </template>
            <template v-else>
               <UButton label="Edit" size="sm" color="secondary" @click="editEquipment(row)" :disabled="isEditing"/>
               <UButton label="Retire" size="sm" color="error" @click="retireEquipment(row.original)" :disabled="isEditing || workstation(row.original.id) != 'N/A'"/>
            </template>
         </div>   
      </template>
   </UTable>
</template>

<script setup>
import { useEquipmentStore } from '@/stores/equipment'
import { ref,computed } from 'vue'
import { useConfirm } from "@/composables/useConfirm"

const props = defineProps({
   equipment: {
      type: Array,
      required: true
   }
})

const equipmentStore = useEquipmentStore()
const editInfo = ref({rowIdx: -1, equipment: null})
const newName = ref("")
const newSerial = ref("")

const cols = [
   {accessorKey: "select", header: ""}, 
   {accessorKey: "name", header: "Name"},  
   {accessorKey: "serialNumber", header: "Serial Number"},
   {accessorKey: "workstation", header: "Workstation"}, 
   {accessorKey: "acts", header: "Actions"}  
]
const colVisibility = computed(() => {
   if (equipmentStore.pending.workstationID == 0) {
      return {select: false}
   }
   return {}
})
const isEditing = computed(() => {
   return editInfo.value.equipment != null
})

const retireEquipment = ( async( equip ) => {
   const msg = `Retire ${equip.type} ${equip.name} with serial number ${equip.serialNumber}?`
   const resp = await useConfirm("Confirm Retire", msg, "Retire")
   if (resp) {
      equipmentStore.updateEquipment( equip.id, equip.name, equip.serialNumber, 2 ) // 2 is status retired
   } 
})

const editEquipment = (( row ) => {
   editInfo.value.rowIdx = row.index
   editInfo.value.equipment = row.original
   newName.value = row.original.name
   newSerial.value = row.original.serialNumber
})

const saveChanges = ( async () => {
   let equipID = editInfo.value.equipment.id
   let currStatus = editInfo.value.equipment.status
   await equipmentStore.updateEquipment( equipID, newName.value, newSerial.value, currStatus )
   cancelEdit()
})

const cancelEdit = (() => {
   editInfo.value = { rowIdx: -1, equipment: null}
})

const equipmentClicked = (( equipID ) => {
   equipmentStore.togglePendingEquipment( equipID )
})

const isItemInUse = ((equipID) => {
   if (equipmentStore.pending.workstationID == 0) return false 
   return equipmentStore.pending.equipment.some( pe => pe.id == equipID)
})
const isItemDisabled = ((equipID) => {
   let equipWS = workstation(equipID)
   if  ( equipWS == "N/A") return false

   let tgtWS = equipmentStore.workstations.find( ws => ws.id == equipmentStore.pending.workstationID)
   if (tgtWS) {
      return tgtWS.name != equipWS
   }
   return true
})

const workstation = (( equipID ) => {
   let wsName = ""
   equipmentStore.workstations.some( ws => {
      let equip = ws.equipment.find( e => e.id == equipID)
      if (equip ) {
         wsName = ws.name
      }
      return wsName != ""
   })
   if (wsName == "") {
      wsName = "N/A"
   }
   return wsName
})

</script>

<style lang="scss" scoped>
.row-acts {
   display: flex;
   flex-flow: row nowrap;
   gap: 5px;
}
</style>