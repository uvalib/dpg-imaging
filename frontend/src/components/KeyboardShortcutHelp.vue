<template>
   <div class="keyboard-shortcuts">
      <DPGButton @click="showMenu" severity="secondary" class="hint-trigger" label="Keyboard Shortcuts" icon="pi pi-question-circle" />
      <Popover ref="hints">
         <table >
            <tr><td class="act">Select All:</td><td>ctrl+a</td></tr>
            <tr><td class="act">Paging:</td><td>&lt;  &gt;</td></tr>
            <tr><td class="act">Delete:</td><td>ctrl+d</td></tr>
            <tr><td class="act">Rename:</td><td>ctrl+r</td></tr>
            <tr><td class="act">Page Numbers:</td><td>ctrl+p</td></tr>
            <tr v-if="isManuscript"><td class="act">Set Box:</td><td>ctrl+b</td></tr>
            <tr v-if="isManuscript"><td class="act">Set Folder:</td><td>ctrl+f</td></tr>
            <tr><td class="act">Component:</td><td>ctrl+k</td></tr>
            <tr><td class="act">Cancel Edit:</td><td>esc</td></tr>
         </table>
      </Popover>
   </div>
</template>

<script setup>
import Popover from 'primevue/popover'
import { ref, computed } from 'vue'
import {useProjectStore} from "@/stores/project"

const hints = ref()
const projectStore = useProjectStore()

const isManuscript = computed(() => {
   return projectStore.detail.workflow && projectStore.detail.workflow.name=='Manuscript'
})
function showMenu(event) {
   hints.value.toggle(event)
}
</script>

<style lang="scss" scoped>
.hint-trigger {
   float:right;
   z-index: 999;
   margin-right: 10px;
}

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