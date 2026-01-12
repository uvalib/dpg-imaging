<template>
   <DPGButton @click="show" label="Add Equipment"  size="small"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Add Equipment" :style="{width: '400px'}" :closable="false">
      <div class="form">
         <div class="field">
            <label for="equip-type">Type</label>
            <Select v-model="equipType" id="equip-type" :options="equipmentTypes" optionLabel="label" optionValue="value" placeholder="Select equipment type" />
         </div>
         <div class="field">
            <label for="equip-name">Name</label>
            <InputText v-model="name" id="equip-name"/>
         </div>
         <div class="field">
            <label for="equip-serial">Serial Number</label>
            <InputText v-model="serialNumber" id="equip-serial"/>
         </div>

         <div class="acts">
            <DPGButton @click="hide" label="Cancel" severity="secondary"/>
            <DPGButton @click="addEquipment" label="Add" :disabled="missingData" />
         </div>
      </div>
   </Dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import Dialog from 'primevue/dialog'
import FloatLabel from 'primevue/floatlabel'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import { useEquipmentStore } from '@/stores/equipment'

const equipmentStore = useEquipmentStore()

const isOpen = ref(false)
const name = ref("")
const serialNumber = ref("")
const equipType = ref()


const missingData = computed( () => {
   return ( name.value == "" || serialNumber.value == "" || equipType.value == null )
})
const equipmentTypes = computed( () => {
   return [
      {label: "Camera Body", value: "CameraBody"},
      {label: "Digital Back", value: "DigitalBack"},
      {label: "Lens", value: "Lens"},
      {label: "Scanner", value: "Scanner"},
   ]
})

const addEquipment= ( async () => {
   await equipmentStore.addEquipment( equipType.value, name.value, serialNumber.value)
   hide()
})

const hide = (() => {
   isOpen.value=false
})

const show = (() => {
   name.value = ""
   serialNumber.value = ""
   equipType.value = null
   isOpen.value = true

})
</script>

<style lang="scss" scoped>
.form {
   text-align: left;
   display: flex;
   flex-direction: column;
   gap: 20px;
   .field {
      text-align: left;
      display: flex;
      flex-direction: column;
      gap: 5px;
   }
}
.acts {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-end;
   padding: 15px 0 10px 0;
   margin: 0;
   gap: 10px;
}
</style>
