<template>
   <UModal v-model:open="open" :modal="true" :dismissible="false" :close="false" title="Set Page Numbers">
      <UButton @click="showClicked()" size="sm" color="secondary" label="Set Page Numbers" />
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
                  <label>Starting page number:</label>
                  <UInput id="start-page-num" v-model="startPage"  @keyup.enter="okPagesClicked"/>
               </span>
               <UCheckbox v-model="unnumberVerso" size="lg" label="Unnumbered Verso" />
            </div>   
            <p class="error" v-if="error">{{ error }}</p>
         </div>
      </template>
      <template #footer="{ close }">
         <UButton label="Cancel" color="secondary" @click="close" />
         <UButton label="OK" @click="okPagesClicked" />
      </template>
   </UModal>
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { ref, computed } from 'vue'
import { onKeyStroke } from '@vueuse/core'

const unitStore = useUnitStore()

const startPage = ref("1")
const unnumberVerso = ref(false)
const open = ref(false)
const startIdx = ref()
const endIdx = ref()
const error = ref("")

onKeyStroke('p', (e) => {
   if ( e.ctrlKey ) {
      showClicked()
   }
})

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   console.log(list)
   return list
})

const showClicked = (() => {
   open.value = true
   startPage.value = "1"
   unnumberVerso.value = false
   error.value = ""
   if (unitStore.rangeStartIdx > -1 ) {
      startIdx.value = unitStore.rangeStartIdx
   }
   if (unitStore.rangeEndIdx > -1 ) {
      endIdx.value = unitStore.rangeEndIdx
   }
})

const startChanged = (() => {
   error.value = ""
   unitStore.startFileSelected( startIdx.value )
})
const endChanged = (() => {
   error.value = ""
   unitStore.endFileSelected( endIdx.value )
})

const okPagesClicked = (() => {
   error.value = ""
   if ( startIdx.value == -1 || endIdx.value == -1) {
      error.value = "Start and end image must be selected"
      return
   }
   if (startPage.value == "") {
      error.value =  "Start page is required" 
      return
   }
   if (unnumberVerso.value && (unitStore.rangeEndIdx-unitStore.rangeStartIdx)%2 == 0) {
      error.value =  "An even number of pages is required for unnumbered verso"
      return
   }
   unitStore.updatePageNumbers(startPage.value, !unnumberVerso.value)
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
   padding: 10px;
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
      }
   }
}
</style>
