<template>
   <span class="actions">
      <RenameFilesDialog />
      <PageNumDialog />
      <BatchUpdateDialog title="Title" field="title" />
      <BatchUpdateDialog title="Caption" field="description" />
      <template v-if="projectStore.isManuscript">
         <BatchUpdateDialog :title="projectStore.detail.containerType.name" field="box" />
         <BatchUpdateDialog v-if="projectStore.detail.containerType.hasFolders" title="Folder" field="folder" />
      </template>
      <ComponentDialog />
   </span>
</template>

<script setup>
import { useProjectStore } from "@/stores/project"
import ComponentDialog from '@/components/unit/ComponentDialog.vue'
import BatchUpdateDialog from '@/components/unit/BatchUpdateDialog.vue'
import PageNumDialog from '@/components/unit/PageNumDialog.vue'
import RenameFilesDialog from '@/components/unit/RenameFilesDialog.vue'

const projectStore = useProjectStore()

</script>

<style lang="scss" scoped>
.actions {
   display: flex;
   flex-flow: row wrap;
   gap: 10px;
}
</style>