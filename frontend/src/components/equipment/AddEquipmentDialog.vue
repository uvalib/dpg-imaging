<template>
   <UModal v-model:open="isOpen" :modal="true" :dismissible="false" :close="false" title="Add Equipment">
      <UButton @click="show" label="Add Equipment"/>
      <template #body>
         <div class="form">
            <div class="field">
               <label for="equip-type">Type</label>
               <USelect v-model="equipType" id="equip-type" :items="equipmentTypes" placeholder="Select equipment type" />
            </div>
            <div class="field">
               <label for="equip-name">Name</label>
               <UInput v-model="name" id="equip-name"/>
            </div>
            <div class="field">
               <label for="equip-serial">Serial Number</label>
               <UInput v-model="serialNumber" id="equip-serial"/>
            </div>
         </div>
      </template>
      <template #footer>
         <UButton @click="hide" label="Cancel" color="secondary"/>
         <UButton @click="addEquipment" label="Add" :disabled="missingData" />
      </template>
   </UModal>
</template>

<script setup>
import { ref, computed } from 'vue'
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
   display: flex;
   flex-direction: column;
   gap: 10px;
   .field {
      text-align: left;
      display: flex;
      flex-direction: column;
   }
}
</style>
