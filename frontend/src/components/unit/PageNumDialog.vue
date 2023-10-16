<template>
   <DPGButton  class="p-button-secondary right" @click="showClicked">
      Set Page Numbers
   </DPGButton>
   <Dialog v-model:visible="unitStore.edit.pageNumber" :modal="true" header="Set Page Numbers" @show="opened">
      <div class="panel">
         <div class="row">
            <span class="entry pad-right">
               <label>Start Image:</label>
               <Dropdown v-model="unitStore.rangeStartIdx" @change="startChanged" filter placeholder="Select start page"
                  :options="masterFiles" optionLabel="label" optionValue="value" ref="pickstart" />
            </span>
            <span class="entry  pad-right">
               <label>End Image:</label>
               <Dropdown v-model="unitStore.rangeEndIdx" @change="endChanged" filter placeholder="Select end page"
                  :options="masterFiles" optionLabel="label" optionValue="value"/>
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
import Dropdown from 'primevue/dropdown'
import { ref, nextTick, computed } from 'vue'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const startPage = ref("1")
const unnumberVerso = ref(false)
const pickstart = ref()

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   unitStore.edit.pageNumber = true
   startPage.value = "1"
   unnumberVerso.value = false
})

const opened = (() => {
   nextTick( () => {
      pickstart.value.$el.focus()
   })
})

const startChanged = (() => {
   unitStore.startFileSelected( unitStore.rangeStartIdx )
})
const endChanged = (() => {
   unitStore.endFileSelected( unitStore.rangeEndIdx )
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
         display: block;
         margin-bottom: 5px;
      }
      button {
         padding: 8px 16px;
      }
      input[type=text] {
         padding: 8px;
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
