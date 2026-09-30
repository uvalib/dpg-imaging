<template>
   <UModal v-model:open="isOpen" :modal="true" :dismissible="false" :close="false" title="Add Workstation">
      <UButton @click="show" label="Add Workstation"/>
      <template #body>
         <label for="ws-name">Name</label>
         <UInput id="ws-name" v-model="workstationName" class="w-full" autofocus/>
      </template>
      <template #footer>
         <UButton @click="hide" label="Cancel" color="secondary"/>
         <UButton @click="addWorkstation" label="Add"/>
      </template>
   </UModal>
</template>

<script setup>
import { ref } from 'vue'
import { useEquipmentStore } from '@/stores/equipment'

const equipmentStore = useEquipmentStore()

const isOpen = ref(false)
const workstationName = ref("")

const addWorkstation= (async () => {
   await equipmentStore.addWorkstation( workstationName.value)
   hide()
})

const hide = (() => {
   isOpen.value=false
})

const show = (() => {
   workstationName.value = ""
   isOpen.value = true
})
</script>

<style lang="scss" scoped>
.content {
   padding: 5px 0 0 0;
}
</style>
