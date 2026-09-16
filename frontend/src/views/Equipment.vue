<template>
   <h2>
      <span>Manage Equipment</span>
      <div class="actions" >
         <AddWorkstationDialog />
         <AddEquipmentDialog />
      </div>
   </h2>
   <div class="equipment">
      <div class="columns">
         <UCard title="Workstations" class="grow max-h-150 scroll-smooth" :ui="{header: 'bg-brand-grey-200 pl-3! pt-2 pb-2'}">
            <UTable :data="equipmentStore.workstations" :columns="wsCols" v-model:row-selection="wsRowSelection"
               :rowSelectionOptions="{enableMultiRowSelection: false}" @select="workstationSelected" 
               v-model:column-visibility="hideCheckbox" class="h-full pb-8" virtualize
            >
               <template #select-cell="{ row }">
                  <UCheckbox :modelValue="row.getIsSelected()"/>
               </template>
               <template #status-cell="{ row }">
                  <span :class="statusClass(row.original.status)"></span>
               </template>
               <template #actions-cell="{ row }">
                  <div  class="row-acts">
                     <UButton size="sm" v-if="row.original.status==0" label="Deactivate"  color="secondary" @click="deactivateWorkstation(row.original.id)"/>
                     <UButton size="sm" v-else label="Activate" color="secondary" @click="activateWorkstation(row.original.id)"/>
                     <UButton size="sm" label="Retire"  color="error" @click="retireWorkstation(row.original.id)" :disabled="row.original.projectCount > 0"/>
                  </div>
               </template>
            </UTable>
         </UCard>
         <UCard :title="setupHeader" class="grow max-h-150 scroll-smooth" :ui="{header: 'bg-brand-grey-200 pl-3! pt-2 pb-2'}">
            <UTable :data="equipmentStore.pendingEquipment.equipment" :columns="equipCols" class="h-full">
               <template #body-bottom>
                  <div v-if="equipmentStore.selectedWorkstation" class="row-acts pt-3">
                     <UButton label="Clear Setup" color="secondary" @click="clearSetup" :disabled="clearAllDisabled"/>
                     <UButton label="Save Setup Changes" color="secondary" @click="saveSetup"
                        :disabled="!(equipmentStore.pendingEquipment.changed==true && equipmentStore.pendingEquipment.equipment.length > 0)"/>
                  </div>
               </template>
            </UTable>
         </UCard>
      </div>
      <div class="columns">
         <UCard title="Equipment" class="grow" :ui="{header: 'bg-brand-grey-200 pl-3! pt-2 pb-2'}">
            <UTabs :items="tabs" variant="link"  >
               <template #bodies>
                  <EquipmentPanel :equipment="equipmentStore.cameraBodies" />
               </template>
               <template #lenses>
                  <EquipmentPanel :equipment="equipmentStore.lenses" />
               </template>
               <template #backs>
                  <EquipmentPanel :equipment="equipmentStore.digitalBacks" />
               </template>
               <template #scanners>
                  <EquipmentPanel :equipment="equipmentStore.scanners" />
               </template>
            </UTabs>
         </UCard>
      </div>
   </div> 
</template>

<script setup>
import { onBeforeMount, ref, computed } from 'vue'
import { useEquipmentStore } from '@/stores/equipment'
import { useConfirm } from "@/composables/useConfirm"
import EquipmentPanel from '@/components/equipment/EquipmentPanel.vue'
import AddWorkstationDialog from '@/components/equipment/AddWorkstationDialog.vue'
import AddEquipmentDialog from '@/components/equipment/AddEquipmentDialog.vue'

const equipmentStore = useEquipmentStore()

const wsRowSelection = ref({})

const wsCols = [
   {accessorKey: "select", header: ""}, 
   {accessorKey: "status", header: "Active"}, 
   {accessorKey: "name", header: "Name"}, 
   {accessorKey: "projectCount", header: "Projects"}, 
   {accessorKey: "actions", header: "Actions"}, 
]
const hideCheckbox = ref({select: false})

const equipCols = [
   {accessorKey: "type", header: "Type"},   
   {accessorKey: "name", header: "Name"},   
   {accessorKey: "serialNumber", header: "Serial Number"},   
]
const tabs = [
  { label: 'Camera Bodies', slot: 'bodies' },
  { label: 'Lenses', slot: 'lenses' },
  { label: 'Digital Backs', slot: 'backs' },
  { label: 'Scanners', slot: 'scanners' },
]

const setupHeader = computed(() => {
   const selWs = equipmentStore.selectedWorkstation 
   if ( selWs ) {
      return `${selWs.name} Setup`
   }
   return "Workstation Setup"
})

const clearAllDisabled = computed( () => {
   const selWs = equipmentStore.selectedWorkstation 
   if ( selWs ) {
      return selWs.equipment.length == 0
   }
   return true
})

onBeforeMount( async () => {
   equipmentStore.getEquipment()
})

const statusClass = ((statusID) => {
   if (statusID == 1) {
      return "ws-status inactive"
   }
   return "ws-status active"
})

const workstationSelected = (( _e, row) => {
   const selWS = equipmentStore.selectedWorkstation 
   if ( selWS && selWS.id == row.original.id )  { 
      equipmentStore.deselectWorkstation( )
   }
   else if ( selWS == null || (selWS && selWS.id != row.original.id) ) {
      equipmentStore.selectWorkstation( row.original.id )
   }
   row.toggleSelected(!row.getIsSelected())
})

const deactivateWorkstation = (( wsID ) => {
   equipmentStore.deactivateWorkstation(wsID)
})

const activateWorkstation = (( wsID ) => {
   equipmentStore.activateWorkstation(wsID)
})

const retireWorkstation = ( async ( wsID ) => {
   const ws = equipmentStore.workstations.find(ws => ws.id == wsID)
   const resp = await useConfirm("Confirm Retire", `Retire workstation '${ws.name}'?`, "Retire")
   if (resp) {
      equipmentStore.retireWorkstation(wsID)
   } 
})

const clearSetup = (() => {
   equipmentStore.clearSetup()
})

const saveSetup = (() => {
   equipmentStore.saveSetup()
})
</script>

<style scoped lang="scss">

h2 {
   display: flex;
   flex-flow: row wrap;
   justify-content: space-between;
   align-items: center;
   margin: 0 !important;
   background-color: var(--uvalib-grey-lightest);
   border-bottom: 1px solid var(--uvalib-grey-light);
   padding: 0.75rem 1rem;
   .actions {
      display: flex;
      flex-flow: row nowrap;
      align-items: center;
      gap: 5px;
   }
}

.equipment {
   .columns {
      padding: 20px;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      gap: 20px;

      h3 {
         text-align: center;
         color: var(--uvalib-text);
         font-weight: 500;
         font-size: 1em;
      }

      span.ws-status {
         width: 20px;
         height: 20px;
         display: inline-block;
         border-radius: 20px;
         background: var(--uvalib-green);
      }

      span.ws-status.inactive {
         background: var(--uvalib-grey-light);
      }
   }
   .row-acts {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      gap: 5px;
   }
}

</style>
