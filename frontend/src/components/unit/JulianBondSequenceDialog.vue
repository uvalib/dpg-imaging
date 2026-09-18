<template>
   <UModal v-model:open="open" :modal="true" :dismissible="false" :close="false" title="Add Sequence Number">
      <UButton @click="showClicked()" size="sm" color="secondary" label="Add Sequence" />
      <template #body>
         <div class="panel">
            <div class="info">
               <div class="info">This will append ", PJB ####" to the title of the selected images.<br/>If requested, it will replace any PJB info already present.</div>
               <div class="info">
                  <em>IMPORTANT</em>: sequence can only be added to images that have already been loaded.
                  You can accomplish this by paging through all images in the unit before requesting the seqence.
               </div>
            </div>
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
                  <UInput id="start-page-num" v-model="startSequence"  @keyup.enter="okClicked"/>
               </span>
               <UCheckbox v-model="overwrite" size="lg" label="Overwrite existing sequence?" />
            </div>   
            <p class="error" v-if="error">{{ error }}</p>
         </div>
      </template>
      <template #footer="{ close }">
         <UButton label="Cancel" color="secondary" @click="close" />
         <UButton label="OK" @click="okClicked" />
      </template>
   </UModal>
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { ref, computed } from 'vue'

const unitStore = useUnitStore()

const open = ref(false)
const startSequence = ref(1)
const overwrite = ref(false)
const startIdx = ref()
const endIdx = ref()
const error = ref("")

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   startSequence.value = 1
   error.value = ""
   if (unitStore.rangeStartIdx > -1 ) {
      startIdx.value = unitStore.rangeStartIdx
   }
   if (unitStore.rangeEndIdx > -1 ) {
      endIdx.value = unitStore.rangeEndIdx
   }
   open.value = true
})

const startChanged = (() => {
   error.value = ""
   unitStore.startFileSelected( startIdx.value )
})
const endChanged = (() => {
   error.value = ""
   unitStore.endFileSelected( endIdx.value )
})

const okClicked = ( async () => {
   error.value = ""
   if ( startIdx.value == -1 || endIdx.value == -1) {
      error.value = "Start and end image must be selected"
      return
   }
   if (startSequence.value == "") {
      error.value =  "Start sequence is required"
      return
   }

   await unitStore.updateJulianBondSequence(startSequence.value, overwrite.value)
   open.value = false
})

const selectAllClicked = (() => {
   unitStore.selectAll()
   startIdx.value = unitStore.rangeStartIdx
   endIdx.value = unitStore.rangeEndIdx
})
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   display: flex;
   flex-direction: column;
   gap: 20px;
   padding: 10px;
   em {
      font-weight: bold;
   }
   .error {
      margin: 0;
      padding: 0;
      color: var(--uvalib-red-emergency);
   }

   .info {
      text-align: left;
      margin: 0 0 10px 0;
      max-width: 450px;
   }

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-end;
      gap: 10px;
      :deep(label) {
         white-space: nowrap !important;
      }
   }
}
</style>
