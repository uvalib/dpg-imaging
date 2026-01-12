<template>
   <DPGButton @click="show" label="Add Workstation" size="small"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Add Workstation" :style="{width: '400px'}" :closable="false">
      <div class="content">
         <FloatLabel variant="on">
            <InputText id="ws-name" type="text" v-model="workstationName" />
            <label for="ws-name">Name</label>
         </FloatLabel>
      </div>
      <div class="acts">
         <DPGButton @click="hide" label="Cancel" severity="secondary"/>
         <DPGButton label="Add" @click="addWorkstation" :disabled="workstationName.length==0" />
      </div>
   </Dialog>
</template>

<script setup>
import { ref } from 'vue'
import Dialog from 'primevue/dialog'
import InputText from 'primevue/inputtext'
import FloatLabel from 'primevue/floatlabel'
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
.acts {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-end;
   padding: 15px 0 10px 0;
   margin: 0;
   gap: 10px;
}
</style>
