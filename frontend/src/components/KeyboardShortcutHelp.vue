<template>
   <div class="tag-picker">
      <DPGButton @click="showMenu" class="p-button-secondary hint-trigger" label="Keyboard Shortcuts" icon="pi pi-question-circle" />
      <OverlayPanel ref="hints">
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
      </OverlayPanel>
   </div>
</template>

<script setup>
import OverlayPanel from 'primevue/overlaypanel'
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
   position: absolute;
   right: 10px;
   top: 0px;
   z-index: 999;
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