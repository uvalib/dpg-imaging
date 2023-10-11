<template>
   <DPGButton @click="renameClicked" class="p-button-secondary right-pad" label="Rename All"/>
   <ConfirmDialog>
      <template #message>
         <div>All files will be renamed to match the following format:</div>
         <code>{{paddedUnit()}}_0001.tif - {{paddedUnit()}}_nnnn.tif</code>
      </template>
   </ConfirmDialog>
</template>

<script setup>
import { useConfirm } from "primevue/useconfirm"
import {useUnitStore} from "@/stores/unit"
import { onBeforeMount, onBeforeUnmount } from 'vue'

const confirm = useConfirm()
const unitStore = useUnitStore()

onBeforeMount( async () => {
   // setup keyboard litener for shortcuts
   window.addEventListener('keydown', keyboardHandler)
})

onBeforeUnmount( async () => {
   window.removeEventListener('keydown', keyboardHandler)
})

function keyboardHandler(event) {
   if ( !event.ctrlKey ) return
   if (event.key == 'r') {
      renameClicked()
   }
}

function renameClicked() {
   confirm.require({
      header: 'Confirm Rename',
      accept: () => {
         unitStore.renameAll()
      }
   })
}

function paddedUnit() {
   let unitStr = ""+unitStore.currUnit
   return unitStr.padStart(9,'0')
}
</script>
