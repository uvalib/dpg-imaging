<template>
   <DPGButton  class="p-button-secondary right" @click="showClicked">
      Set {{ props.title}}
   </DPGButton>
   <Dialog v-model:visible="showDialog" :modal="true" :header="`Batch Update ${props.title}`">
      <div class="panel">
         <div class="row"  v-if="props.global == false">
            <span class="entry pad-right">
               <label>Start Image:</label>
               <select id="start-page" v-model="unitStore.rangeStartIdx" @change="startChanged">
                  <option disabled :value="-1">Select start page</option>
                  <option v-for="(mf,idx) in unitStore.masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
               </select>
            </span>
            <span class="entry  pad-right">
               <label>End Image:</label>
               <select id="end-page" v-model="unitStore.rangeEndIdx" @change="endChanged">
                  <option disabled :value="-1">Select end page</option>
                  <option v-for="(mf,idx) in unitStore.masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
               </select>
            </span>
            <DPGButton @click="selectAllClicked" class="p-button-secondary left" label="Select All"/>
         </div>
         <div class="row ">
            <span class="entry full">
               <label>{{props.title}}:</label>
               <input id="update-value" type="text" v-model="newValue"  @keyup.enter="okClicked"/>
            </span>
         </div>
      </div>
      <div class="panel-actions">
         <DPGButton @click="cancelEditClicked" class="p-button-secondary right-pad" label="Cancel"/>
         <DPGButton @click="okClicked" label="OK"/>
      </div>
   </Dialog>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import { ref, nextTick } from 'vue'
import Dialog from 'primevue/dialog'

const props = defineProps({
   title: {
      type: String,
      required: true
   },
   field: {
      type: String,
      required: true
   },
   global: {
      type: Boolean,
      default: false
   }
})

const unitStore = useUnitStore()

const newValue = ref("")
const showDialog = ref(false)

const showClicked = (() => {
   showDialog.value = true
   newValue.value = ""
   nextTick( () => {
      let ele = document.getElementById("start-page")
      if (props.global) {
         ele = document.getElementById("update-value")
      }
      ele.focus()
   })
})

const startChanged = (() => {
   let pageIndex = unitStore.rangeStartIdx % unitStore.pageSize
   unitStore.deselectAll()
   unitStore.masterFileSelected( pageIndex )
})
const endChanged = (() => {
   let pageIndex = unitStore.rangeEndIdx % unitStore.pageSize
   unitStore.masterFileSelected( pageIndex )
})

const okClicked = ( () => {
   if ( props.global) {
      unitStore.selectAll()
   }
   cancelEditClicked()
   unitStore.batchUpdate( props.field, newValue.value )
})

const cancelEditClicked = (() => {
   showDialog.value = false
})

const selectAllClicked = (() => {
   unitStore.selectAll()
})
</script>

<style lang="scss" scoped>
button.p-button-secondary.right {
   margin-right: 10px;
}
.panel {
   background: white;

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-end;
      margin-bottom: 20px;
      label {
         display: inline-block;
         margin-bottom: 5px;
      }
      input {
         width: 100%;
      }
      .entry.pad-right {
         margin-right: 10px;
      }
      .entry.full {
         width: 100%;
      }
   }
}
.panel-actions {
   padding: 0 0 10px 0;
   text-align: right;
}
</style>
