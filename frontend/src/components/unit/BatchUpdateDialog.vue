<template>
   <DPGButton  severity="secondary" @click="showClicked">
      Set {{ props.title}}
   </DPGButton>
   <Dialog v-model:visible="showDialog" :modal="true" :header="`Batch Update ${props.title}`">
      <div class="panel">
         <div class="row">
            <span class="entry">
               <label>Start Image:</label>
               <Select v-model="unitStore.rangeStartIdx" @change="startChanged" filter placeholder="Select start page"
                  :options="masterFiles" optionLabel="label" optionValue="value" />
            </span>
            <span class="entry">
               <label>End Image:</label>
               <Select v-model="unitStore.rangeEndIdx" @change="endChanged" filter placeholder="Select end page"
                  :options="masterFiles" optionLabel="label" optionValue="value"/>
            </span>
            <DPGButton @click="selectAllClicked" severity="secondary" label="Select All"/>
         </div>
         <div class="row ">
            <span class="entry full">
               <label>{{props.title}}:</label>
               <input id="update-value" type="text" v-model="newValue"  @keyup.enter="okClicked"/>
            </span>
         </div>
      </div>
      <template #footer>
         <DPGButton @click="cancelEditClicked" severity="secondary" label="Cancel"/>
         <DPGButton @click="okClicked" label="OK"/>
      </template>
   </Dialog>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import { ref, computed } from 'vue'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'

const props = defineProps({
   title: {
      type: String,
      required: true
   },
   field: {
      type: String,
      required: true
   },
})

const unitStore = useUnitStore()

const newValue = ref("")
const showDialog = ref(false)

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   showDialog.value = true
   newValue.value = ""
})

const startChanged = (() => {
   unitStore.startFileSelected( unitStore.rangeStartIdx )
})
const endChanged = (() => {
   unitStore.endFileSelected( unitStore.rangeEndIdx )
})

const okClicked = ( () => {
   unitStore.batchUpdate( props.field, newValue.value )
   showDialog.value = false
})

const cancelEditClicked = (() => {
   showDialog.value = false
})

const selectAllClicked = (() => {
   unitStore.selectAll()
})
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   display: flex;
   flex-direction: column;
   gap: 20px;

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-end;
      justify-content: flex-start;
      gap: 10px;
      text-align: left;

      label {
         display: block;
         margin-bottom: 5px;
      }
      .entry.full {
         width: 100%;
      }
   }
}
</style>
