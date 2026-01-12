<template>
   <DataTable :value="props.equipment" ref="equipmentTable" dataKey="id"
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
      </Column>
      <Column field="" header="Workstation" class="e-wide">
         <template #body="slotProps">
            <span class="workstation">{{workstation(slotProps.data.id)}}</span>
         </template>
      </Column>
      <Column header="Actions">
         <template #body="slotProps">
            <div class="row-acts">
               <template v-if="editingRows.length == 1 && editingRows[0].id == slotProps.data.id">
                  <DPGButton label="Cancel"  severity="secondary" @click="cancelEdit"/>
                  <DPGButton label="Save"  severity="secondary" @click="saveChanges" :disabled="workstation(slotProps.data.id) != 'N/A'"/>
               </template>
               <template v-else>
                  <DPGButton label="Edit"  severity="secondary" @click="editEquipment(slotProps.data)" :disabled="editingRows.length > 0"/>
                  <DPGButton label="Retire"  severity="secondary" @click="retireEquipment(slotProps.data.id)" :disabled="editingRows.length > 0 || workstation(slotProps.data.id) != 'N/A'"/>
               </template>
            </div>
         </template>
      </Column>
   </DataTable>
</template>

<script setup>
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useEquipmentStore } from '@/stores/equipment'
import Checkbox from 'primevue/checkbox'
import InputText from 'primevue/inputtext'
import { ref } from 'vue'
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

function workstation( equipID ) {
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
}

</script>

<style lang="scss" scoped>
.e-wide {
   width: 30%;
}
</style>