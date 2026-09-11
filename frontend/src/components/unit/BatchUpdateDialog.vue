<template>
   <UModal v-model:open="open" :modal="true" :dismissible="false" :close="false" :title="`Batch Update ${props.title}`">
      <UButton @click="showClicked()" size="sm" color="secondary" :label="`Set ${props.title}`" />
      <template #body>
         <div class="panel">
            <div class="row">
               <span class="entry">
                  <label>Start Image:</label>
                  <USelect v-model="startIdx" @change="startChanged" placeholder="Select start page" :items="masterFiles" />
               </span>
               <span class="entry">
                  <label>End Image:</label>
                  <USelect v-model="endIdx" @change="endChanged" filter placeholder="Select end page" :items="masterFiles"/>
               </span>
               <UButton @click="selectAllClicked" size="sm" color="secondary" label="Select All"/>
            </div>
            <div class="row">
               <span class="entry">
                  <label>{{ props.title }}:</label>
                  <UInput id="update-value" v-model="newValue" class="w-full" @keyup.enter="okClicked"/>
               </span>
            </div>   
            <p class="error" v-if="error">{{ error }}</p>
         </div>
      </template>
      <template #footer="{ close }">
         <UButton label="Cancel" size="sm" color="secondary" @click="close" />
         <UButton label="OK" size="sm" @click="okClicked" />
      </template>
   </UModal>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import { ref, computed } from 'vue'
import { onKeyStroke } from '@vueuse/core'

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
const startIdx = ref()
const endIdx = ref()
const error = ref("")
const open = ref(false)

onKeyStroke('b', (e) => {
   if ( e.ctrlKey && props.field=="box" ) {
      showClicked()
   }
})
onKeyStroke('f', (e) => {
   if ( e.ctrlKey && props.field=="folder" ) {
      showClicked()
   }
})
onKeyStroke('t', (e) => {
   if ( e.ctrlKey && props.field=="title" ) {
      showClicked()
   }
})

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   open.value = true
   newValue.value = ""
   if (unitStore.rangeStartIdx > -1 ) {
      startIdx.value = unitStore.rangeStartIdx
   }
   if (unitStore.rangeEndIdx > -1 ) {
      endIdx.value = unitStore.rangeEndIdx
   }
})

const startChanged = (() => {
   error.value = ""
   unitStore.startFileSelected( unitStore.rangeStartIdx )
})
const endChanged = (() => {
   error.value = ""
   unitStore.endFileSelected( unitStore.rangeEndIdx )
})

const okClicked = ( () => {
   error.value = ""
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      error.value = "Start and end image must be selected"
      return
   }
   unitStore.batchUpdate( props.field, newValue.value )
   open.value = false
})

const selectAllClicked = (() => {
   unitStore.selectAll()
})
</script>

<style lang="scss" scoped>
.panel {
   display: flex;
   flex-direction: column;
   gap: 20px;
   .error {
      margin: 0;
      padding: 0;
      color: var(--uvalib-red-emergency);
   }

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      gap: 15px;
      text-align: left;
      .entry {
         display: flex;
         flex-direction: column;
         width: 100%;
      }
   }
}
</style>
