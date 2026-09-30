<template>
   <UModal v-model:open="open" :modal="true" :dismissible="false" :close="false" title="Confirm Rename">
      <UButton @click="open=true" size="sm" color="secondary" label="Rename All" />
      <template #body>
         <div style="display:flex; flex-direction: column; gap: 10px; align-items: flex-start;">
            <div>All files will be renamed to match the following format:</div>
            <code>{{paddedUnit}}_0001.tif - {{paddedUnit}}_nnnn.tif</code>
         </div>
      </template>
      <template #footer="{ close }">
         <UButton label="Cancel" color="secondary" @click="close" />
         <UButton label="Rename" @click="unitStore.renameAll()" />
      </template>
   </UModal>
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { ref, computed } from 'vue'
import { onKeyStroke } from '@vueuse/core'

const open = ref(false)
const unitStore = useUnitStore()

onKeyStroke('r', (e) => {
   if ( e.ctrlKey ) {
      open.value = true
   }
})

const paddedUnit = computed(() => {
   let unitStr = ""+unitStore.unitID
   return unitStr.padStart(9,'0')
})
</script>
