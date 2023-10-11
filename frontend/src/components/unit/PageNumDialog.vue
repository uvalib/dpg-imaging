<template>
   <DPGButton  class="p-button-secondary right" @click="showClicked">
      Set Page Numbers
   </DPGButton>
   <Dialog v-model:visible="unitStore.edit.pageNumber" :modal="true" header="Set Page Numbering" @show="opened">
      <div class="panel">
         <div class="row">
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
         <div class="row">
            <span class="entry  pad-right">
               <label>Starting Page:</label>
               <input id="start-page-num" type="text" v-model="startPage"  @keyup.enter="okPagesClicked"/>
            </span>
            <label class="verso"><input v-model="unnumberVerso" type="checkbox"/>Unnumbered Verso</label>
         </div>
      </div>
      <div class="panel-actions">
         <DPGButton @click="cancelEditClicked" class="p-button-secondary right-pad" label="Cancel"/>
         <DPGButton @click="okPagesClicked" label="OK"/>
      </div>
   </Dialog>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import {useSystemStore} from "@/stores/system"
import Dialog from 'primevue/dialog'
import { ref, nextTick } from 'vue'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const startPage = ref("1")
const unnumberVerso = ref(false)

const showClicked = (() => {
   unitStore.edit.pageNumber = true
   startPage.value = "1"
   unnumberVerso.value = false
})

const opened = (() => {
   nextTick( () => {
      let ele = document.getElementById("start-page")
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

const cancelEditClicked = (() => {
   unitStore.edit.pageNumber = false
})

const okPagesClicked = (() => {
   systemStore.error = ""
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.setError( "Start and end image must be selected" )
      return
   }
   if (startPage.value == "") {
      systemStore.setError( "Start page is required" )
      return
   }
   if (unnumberVerso.value && (unitStore.rangeEndIdx-unitStore.rangeStartIdx)%2 == 0) {
      systemStore.setError( "An even number of pages is required for unnumbered verso")
      return
   }
   unitStore.updatePageNumbers(startPage.value, !unnumberVerso.value)
   unitStore.edit.pageNumber = false
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
      .entry.pad-right {
         margin-right: 10px;
      }
      .verso {
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
.panel-actions {
   padding: 0 0 10px 0;
   text-align: right;
}
</style>
