<template>
   <DPGButton  severity="secondary" @click="showClicked">
      Set {{ props.title}}
   </DPGButton>
   <Dialog v-model:visible="showDialog" :modal="true" :header="`Batch Update ${props.title}`" @show="opened" :closable="false">
      <div class="panel">
         <div class="row"  v-if="props.global == false">
            <span class="entry">
               <label>Start Image:</label>
               <Select v-model="unitStore.rangeStartIdx" @change="startChanged" filter placeholder="Select start page"
                  :options="masterFiles" optionLabel="label" optionValue="value" ref="pickstart" />
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
import { ref, nextTick, computed } from 'vue'
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
   global: {
      type: Boolean,
      default: false
   }
})

const unitStore = useUnitStore()

const newValue = ref("")
const showDialog = ref(false)
const pickstart = ref()

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const opened = (() => {
   nextTick( () => {
      pickstart.value.$el.focus()
   })
})

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
   unitStore.startFileSelected( unitStore.rangeStartIdx )
})
const endChanged = (() => {
   unitStore.endFileSelected( unitStore.rangeEndIdx )
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
