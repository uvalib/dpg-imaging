<template>
   <div class="list-view sticky z-50" :style="{top: headerHeight}">
      <div class="control-group">
         <ViewMode />
         <UPagination  v-if="unitStore.masterFiles.length>0" 
            v-model:page="unitStore.currPage" :items-per-page="unitStore.pageSize" 
            :total="unitStore.totalFiles" @update:page="pageChanged"
         />
      </div>
      <UnitActions />
   </div>
   <UTable :data="masterFilesPage" :columns="columns" v-model:column-visibility="columnVisibility" :ui="{tbody: 'mf-tbody'}">
      <template #image-cell="{ row }">
         <RouterLink  @click="imageClicked" :to="`/projects/${projectStore.detail.id}/unit/images/${row.index+1}`">
            <img :src="row.original.thumbURL"/>   
         </RouterLink>
      </template>

      <template #title-cell="{ row }">
         <template v-if="editInfo.field=='title' && row.original.fileName == editInfo.fileName">
            <TitlePicker v-model="editInfo.value" @cancel="cancelEdit" @submit="submitEdit"/>
         </template>
         <ULink v-else @click="startEdit('title', row)">
            <span v-if="row.original.title">{{  row.original.title }}</span>
            <span v-else class="undefined">Undefined</span>
         </ULink>
      </template>

      <template #description-cell="{ row }">
         <template v-if="editInfo.field=='description' && row.original.fileName == editInfo.fileName">
            <UInput v-model="editInfo.value" class="w-full" autofocus
               @keydown.enter="submitEdit" @keydown.esc="cancelEdit" @keydown.tab="cancelEdit" />
         </template>
         <ULink v-else @click="startEdit('description', row)">
            <span v-if="row.original.description">{{  row.original.description }}</span>
            <span v-else class="undefined">Undefined</span>
         </ULink>
      </template>

      <template #box-cell="{ row }">
         <template v-if="editInfo.field=='box' && row.original.fileName == editInfo.fileName">
            <UInput v-model="editInfo.value" class="w-full" autofocus 
               @keydown.enter="submitEdit" @keydown.esc="cancelEdit" @keydown.tab="cancelEdit" />
         </template>
         <ULink v-else @click="startEdit('box', row)">
            <span v-if="row.original.box">{{  row.original.box }}</span>
            <span v-else class="undefined">Undefined</span>
         </ULink>
      </template>

      <template #folder-cell="{ row }">
         <template v-if="editInfo.field=='folder' && row.original.fileName == editInfo.fileName">
            <UInput v-model="editInfo.value" class="w-full" autofocus
               @keydown.enter="submitEdit" @keydown.esc="cancelEdit" @keydown.tab="cancelEdit" />
         </template>
         <ULink v-else @click="startEdit('folder', row)">
            <span v-if="row.original.folder">{{  row.original.folder }}</span>
            <span v-else class="undefined">Undefined</span>
         </ULink>
      </template>

      <template #component-cell="{ row }">
          <template v-if="editInfo.field=='component' && row.original.fileName == editInfo.fileName">
            <UInput v-model="editInfo.value" class="w-full" autofocus
               @keydown.enter="submitEdit" @keydown.esc="cancelEdit" @keydown.tab="cancelEdit" />
         </template>
         <ULink v-else @click="startEdit('component', row)"> 
            <span v-if="row.original.component">{{  row.original.component }}</span>
            <span v-else class="undefined">Undefined</span>
         </ULink>
      </template>

      <template #reorder-cell="{ }">
         <UIcon name="i-lucide-arrow-down-up" class="size-6 cursor-grab text-brand-grey"/>
      </template>
   </UTable>
</template>

<script setup>
import TagPicker from '@/components/TagPicker.vue'
import { useProjectStore } from "@/stores/project"
import { useUnitStore } from "@/stores/unit"
import ViewMode from '@/components/ViewMode.vue'
import UnitActions from '@/components/unit/UnitActions.vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import TitlePicker from "@/components/TitlePicker.vue"
import { computed, h, ref, watch  } from 'vue'
import UCheckboxCell from './UCheckboxCell.vue'
import { useSortable } from '@vueuse/integrations/useSortable'
import { onKeyStroke } from '@vueuse/core'

// NOTES: 
//   h is short for hyperscript: javascript which produces html dynamically and injects them into the DOM.
//   Because of the way nuxt-ui components are handled (injected at build time in templates and setup),
//   using UCheckbox (or any nuxt component) in a column def cell reder will fill. The component cannot be resolved.
//   To fix, create a dummy wrapper around the component and directly include it instead (like UCheckboxCell above).
//   Then it use it in the cell render logic and it will be resolved.

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const unitStore = useUnitStore()

const editInfo = ref({fileName: "", field: "", orig: "", value: ""})
const masterFilesPage = ref([])

watch(() => unitStore.masterFilesUpdated, (updated) => {
   if ( updated == true) {
      updateMasterFilesPage()
   }
})
watch(() => [unitStore.currPage, unitStore.pageSize], () => {
   updateMasterFilesPage()
})

const updateMasterFilesPage = (() => {
   // Pagination is wierd here; the store holds all file references, but maybe not all metadata.
   // The table just shows a subset of the total list. This function uses curr page num
   // and page size to get an array of masterfiles for the current page.
   masterFilesPage.value = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      if (idx >= unitStore.currStartIndex && masterFilesPage.value.length <= unitStore.pageSize) {
         masterFilesPage.value.push(mf)  
      }
   })  
})

useSortable('.mf-tbody', masterFilesPage.value, {
  animation: 150,
  onEnd: (evt) => {
    unitStore.moveImage(evt.oldIndex, evt.newIndex)
  }
})

onKeyStroke(['>','.'], () => {
   if (unitStore.currPage < unitStore.totalPages && editInfo.value.field == "") {
      unitStore.currPage++
      pageChanged()
   }
})
onKeyStroke(['<',','], () => {
   if (unitStore.currPage > 1 && editInfo.value.field == "") {
      unitStore.currPage--
      pageChanged()
   }
})

const columns = [
   {
      id: 'select',
      cell: ({ row }) => 
         h(UCheckboxCell, {
            modelValue: unitStore.masterFiles[row.index].selected,
            'onUpdate:modelValue': () => unitStore.masterFileSelected(row.index)
         })
   },
   {
      id: 'image',
   },
   {
      id: 'tag',
      header: "Tag",
      cell: ({ row }) => 
         h(TagPicker, { masterFile: row.original })
   },
   {
      accessorKey: 'fileName',
      header: "File Name"
   },
   {
      accessorKey: 'title',
      header: "Title"
   },
   {
      accessorKey: 'description',
      header: "Caption"
   },
   {
      accessorKey: 'box',
      header: "Box"
   },
   {
      accessorKey: 'folder',
      header: "Folder"
   },
   {
      accessorKey: 'component',
      header: "Component",
   },
   {
      accessorKey: 'fileSize',
      header: "Size"
   },
   {
      header: "Resolution",
      cell: ({ row }) => `${row.original.width} x ${row.original.height}`
   },
   {
      accessorKey: 'colorProfile',
      header: "Color Profile"
   },
   {
      accessorKey: "reorder", header: ""
   }
]

const columnVisibility = computed(() => {
   if (projectStore.isManuscript == false) {
      return {
         box: false, 
         folder: false
      }
   }
   return {}
})

// you cannot custruct tailwind class values dynamically, so you  cant do `top-${hdr.clientHeight}`. 
// Instead use this to bind an inline style 'top' param to stick the controls below the header
const headerHeight = computed(() => {
   let hdr = document.querySelector('header')
   return `${hdr.clientHeight}px`
})

const cancelEdit = (() => {
   editInfo.value = {fileName: "", field: "", orig: "", value: ""}
})
const startEdit = ((field,row) => {
   editInfo.value = {fileName: row.original.fileName, field: field,  rowIndex: row.index, orig: row.original[field], value: row.original[field]}
})
const submitEdit = (() => {
   console.log(editInfo.value.value)
   if ( editInfo.value.orig != editInfo.value.value) {
      unitStore.updateMasterFileMetadata( editInfo.value.fileName, editInfo.value.field, editInfo.value.value)
   }
   editInfo.value = {fileName: "", field: "", orig: "", value: ""}
})

// The pageNum info is already in the store; just set it in the URL and request metadata
const pageChanged = (()=> {
   unitStore.deselectAll()
   let query = Object.assign({}, route.query)
   query.pagesize = unitStore.pageSize
   query.page = unitStore.currPage
   router.push({query})
   unitStore.getMetadataPage()
})

const imageClicked = (() => {
   unitStore.lastURL = route.fullPath
})
</script>

<style lang="scss" scoped>
.list-view {
   padding: 15px;
   background: white;
   border-top: 1px solid var(--uvalib-grey-light);
   border-bottom: 1px solid var(--uvalib-grey-light);
   display: flex;
   flex-flow: row wrap;
   justify-content: space-between;
   align-items: center;
   .control-group {
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      align-items: center;
      gap: 10px;   
   }
}
.filename {
   display: flex;
   flex-flow: row nowrap;
   gap: 10px;
   align-items: center;
   i.image-err {
      font-size: 1.15em;
      color: var(--uvalib-red-emergency);
      cursor: pointer;
   }
}
.undefined {
   font-style: italic;
   color: var(--uvalib-grey);
}
.nowrap {
   white-space: nowrap;
}
</style>