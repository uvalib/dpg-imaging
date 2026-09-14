<template>
   <UPopover>
      <UButton color="secondary" label="Keyboard Shortcuts" icon="i-lucide-help-circle" size="sm" />
      <template #content>
         <div class="tips">
         <table >
            <tbody>
               <tr><td class="act">Select All:</td><td>ctrl+a</td></tr>
               <tr><td class="act">Paging:</td><td>&lt;  &gt;</td></tr>
               <tr><td class="act">Delete:</td><td>ctrl+d</td></tr>
               <tr><td class="act">Rename:</td><td>ctrl+r</td></tr>
               <tr><td class="act">Set Titles:</td><td>ctrl+t</td></tr>
               <tr><td class="act">Page Numbers:</td><td>ctrl+p</td></tr>
               <tr v-if="isManuscript"><td class="act">Set {{ projectStore.detail.containerType.name }}:</td><td>ctrl+b</td></tr>
               <tr v-if="isManuscript"><td class="act">Set Folders:</td><td>ctrl+f</td></tr>
               <tr><td class="act">Component:</td><td>ctrl+k</td></tr>
               <tr><td class="act">Cancel Edit:</td><td>esc</td></tr>
            </tbody>
         </table>
         </div>
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
.tips {
   padding: 10px 20px;
}
table {
   font-size: 0.8em;
   text-align: left;
   td.act {
      text-align: right;
      font-weight: bold;
      padding-right: 10px;
   }
}
</style>