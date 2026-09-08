<template>
   <UPopover>
      <UButton color="secondary" class="hint-trigger" label="Keyboard Shortcuts" icon="i-lucide-help-circle" />
      <template #content>
         <table >
            <tbody>
               <tr><td class="act">Select All:</td><td>ctrl+a</td></tr>
               <tr><td class="act">Paging:</td><td>&lt;  &gt;</td></tr>
               <tr><td class="act">Delete:</td><td>ctrl+d</td></tr>
               <tr><td class="act">Rename:</td><td>ctrl+r</td></tr>
               <tr><td class="act">Page Numbers:</td><td>ctrl+p</td></tr>
               <tr v-if="isManuscript"><td class="act">Set {{ projectStore.detail.containerType.name }}:</td><td>ctrl+b</td></tr>
               <tr v-if="isManuscript"><td class="act">Set Folder:</td><td>ctrl+f</td></tr>
               <tr><td class="act">Component:</td><td>ctrl+k</td></tr>
               <tr><td class="act">Cancel Edit:</td><td>esc</td></tr>
            </tbody>
         </table>
         </template>
   </UPopover>
</template>

<script setup>
import { computed } from 'vue'
import {useProjectStore} from "@/stores/project"

const projectStore = useProjectStore()

const isManuscript = computed(() => {
   return projectStore.detail.workflow && projectStore.detail.workflow.name=='Manuscript'
})
</script>

<style lang="scss" scoped>
table {
   font-size: 0.75em;
   text-align: left;
   td.act {
      text-align: right;
      font-weight: bold;
      padding-right: 10px;
   }
}
</style>