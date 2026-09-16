<template>
   <UTable :data="props.equipment" :columns="cols" class="h-full pb-8" virtualize
   >
      <template #select-cell="{ row }">
         {{ row.original.id }}
         <!-- <UCheckbox :value="slotProps.data" v-model="equipmentStore.pendingEquipment.equipment" :disabled="isItemDisabled(row.original.id)" @click="equipmentClicked"/>     -->
      </template>
      <template #workstation-cell="{ row }">
         <span class="workstation">{{workstation(row.original.id)}}</span>
      </template>
      <template #acts-cell="{ row }">
         <div class="row-acts">
               <template v-if="editingRows.length == 1 && editingRows[0].id == row.original.id">
                  <UButton label="Cancel" size="sm" color="secondary" @click="cancelEdit"/>
                  <UButton label="Save" size="sm" color="secondary" @click="saveChanges" :disabled="workstation(row.original.id) != 'N/A'"/>
               </template>
               <template v-else>
                  <UButton label="Edit" size="sm" color="secondary" @click="editEquipment(row.original)" :disabled="editingRows.length > 0"/>
                  <UButton label="Retire" size="sm" color="error" @click="retireEquipment(row.original.id)" :disabled="editingRows.length > 0 || workstation(row.original.id) != 'N/A'"/>
               </template>
            </div>   
      </template>
   </UTable>
   <!-- <DataTable :value="props.equipment" ref="equipmentTable" dataKey="id"
      stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
      :lazy="false" :paginator="false" :rows="props.equipment.length"
      v-model:editingRows="editingRows" editMode="row"
   >
      <template #empty>No equipment found</template>
      <Column header="" headerStyle="width: 3em" v-if="equipmentStore.pendingEquipment.workstationID > 0">
         <template #body="slotProps">
            <Checkbox :value="slotProps.data" v-model="equipmentStore.pendingEquipment.equipment" :disabled="isItemDisabled(slotProps.data.id)" @click="equipmentClicked"/>
         </template>
      </Column>
      <Column field="name" header="Name" class="e-wide">
         <template #editor>
            <InputText v-model="newName" autofocus />
         </template>
      </Column>
      <Column field="serialNumber" header="Serial Number" class="e-wide">
         <template #editor>
            <InputText v-model="newSerial" />
         </template>
      </Column> -->
</template>

<script setup>
import { useEquipmentStore } from '@/stores/equipment'
import { ref,computed } from 'vue'
import { useConfirm } from "primevue/useconfirm"

const confirm = useConfirm()

const props = defineProps({
   equipment: {
      type: Array,
      required: true
   }
})

const equipmentStore = useEquipmentStore()
const editingRows = ref([])
const newName = ref("")
const newSerial = ref("")

const cols = [
   {accessorKey: "select", header: "S"}, 
   {accessorKey: "name", header: "Name"},  
   {accessorKey: "serialNumber", header: "Serial Number"},
   {accessorKey: "workstation", header: "Workstation"}, 
   {accessorKey: "acts", header: "Actions"}  
]
const colVisibility = computed(() => {
   if (equipmentStore.pendingEquipment.workstationID == 0) {
      return {select: false}
   }
   return {}
})

function retireEquipment( equipID ) {
   let tgtE = equipmentStore.equipment.find( e => e.id == equipID)
   confirm.require({
      message: `Retire equipment name: '${tgtE.name}, serial number: ${tgtE.serialNumber}'?`,
      header: 'Confirm Retire',
      icon: 'pi pi-question-circle',
      rejectProps: {
         label: 'Cancel',
         severity: 'secondary'
      },
      acceptProps: {
         label: 'Retire'
      },
      accept: () => {
         equipmentStore.updateEquipment( equipID, tgtE.name, tgtE.serialNumber, 2 )
      }
   })
}
function editEquipment( equip ) {
   editingRows.value = [equip]
   newName.value = equip.name
   newSerial.value = equip.serialNumber
}
async function saveChanges() {
   let equipID = editingRows.value[0].id
   let currStatus = editingRows.value[0].status
   await equipmentStore.updateEquipment( equipID, newName.value, newSerial.value, currStatus )
   editingRows.value = []
}
function cancelEdit() {
   editingRows.value = []
}

function equipmentClicked() {
   equipmentStore.pendingEquipment.changed = true
}

function isItemDisabled(equipID) {
   let equipWS = workstation(equipID)
   if  ( equipWS == "N/A") return false
   let tgtWS = equipmentStore.workstations.find( ws => ws.id == equipmentStore.pendingEquipment.workstationID)
   if (tgtWS) {
      return tgtWS.name != equipWS
   }
   return true
}

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