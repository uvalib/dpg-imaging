<template>
   <DPGButton  severity="secondary" @click="showClicked">
      Add Sequence
   </DPGButton>
   <Dialog v-model:visible="showDialog" :modal="true" header="Add Sequence Number">
      <div class="panel">
         <div class="info">This will append ", PJB ####" to the title of the selected images.<br/>If requested, it will replace any PJB info already present.</div>
         <div class="info">
            <b>IMPORTANT</b>: sequence can only be added to images that have already been loaded.
            You can accomplish this by paging through all images in the unit before requesting the seqence.
         </div>
         <div class="row">
            <span class="entry pad-right">
               <label>Start Image:</label>
               <Select v-model="unitStore.rangeStartIdx" @change="startChanged" filter placeholder="Select start page"
                  :options="masterFiles" optionLabel="label" optionValue="value" />
            </span>
            <span class="entry  pad-right">
               <label>End Image:</label>
               <Select v-model="unitStore.rangeEndIdx" @change="endChanged" filter placeholder="Select end page"
                  :options="masterFiles" optionLabel="label" optionValue="value"/>
            </span>
            <DPGButton @click="selectAllClicked" severity="secondary" label="Select All"/>
         </div>
         <div class="row">
            <span class="entry  pad-right">
               <label>Starting Sequence Number:</label>
               <input id="start-page-num" type="text" v-model="startSequence"  @keyup.enter="okClicked"/>
            </span>
            <label class="overwrite"><input v-model="overwrite" type="checkbox"/>Overwrite existing sequence?</label>
         </div>
      </div>
      <template #footer>
         <DPGButton @click="cancelEditClicked" severity="secondary" label="Cancel"/>
         <DPGButton @click="okClicked" label="OK"/>
      </template>
   </Dialog>
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { useSystemStore } from "@/stores/system"
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import { ref, computed } from 'vue'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const showDialog = ref(false)
const startSequence = ref(1)
const overwrite = ref(false)

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   startSequence.value = 1
   showDialog.value = true
})

const startChanged = (() => {
   unitStore.startFileSelected( unitStore.rangeStartIdx )
})
const endChanged = (() => {
   unitStore.endFileSelected( unitStore.rangeEndIdx )
})

const cancelEditClicked = (() => {
   showDialog.value = false
})

const okClicked = ( async () => {
   systemStore.error = ""
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.setError( "Start and end image must be selected" )
      return
   }
   if (startSequence.value == "") {
      systemStore.setError( "Start sequence is required" )
      return
   }

   await unitStore.updateJulianBondSequence(startSequence.value, overwrite.value)
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

   .info {
      text-align: left;
      margin: 0 0 10px 0;
      max-width: 450px;
   }

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-end;
      gap: 10px;
      text-align: left;
      label {
         display: block;
         margin-bottom: 5px;
      }
      .overwrite {
         cursor: pointer;
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-evenly;
         align-items: center;
         label {
            vertical-align: middle;
            display: inline-block;
         }
         input {
            width: 20px;
            height:  20px;
            margin-right: 10px;
            vertical-align: middle;
         }
      }
   }
}
</style>
