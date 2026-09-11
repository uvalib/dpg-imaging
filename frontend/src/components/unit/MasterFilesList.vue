<template>
   <div class="list-view sticky z-50" :style="{top: headerHeight}">
      <ViewMode />
      <UPagination  v-if="unitStore.masterFiles.length>0" 
         v-model:page="unitStore.currPage" :items-per-page="unitStore.pageSize" 
         :total="unitStore.totalFiles" @update:page="pageChanged"
      />
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
   <!-- <DataTable :value="unitStore.masterFiles" ref="mfTable" id="mf-table" dataKey="fileName"
         stripedRows size="small" paginatorPosition="top"
         :lazy="false" :rows="unitStore.pageSize" :first="unitStore.currStartIndex" :rowsPerPageOptions="[20,50,75]" paginator
         paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
         currentPageReportTemplate="{currentPage} of {totalPages}"  @page="pageChanged"
         editMode="cell" @cell-edit-complete="onCellEditComplete" @rowReorder="onRowReorder"
   >
      <template #paginatorstart>
         <ViewMode />
      </template>
      <template #paginatorend>
         <UnitActions />
      </template>
      <Column headerStyle="width: 3rem">
         <template #body="slotProps">
            <input type="checkbox" style="width: 20px;height: 20px" v-model="slotProps.data.selected" @click="masterFileCheckboxClicked(slotProps.data)"/>
         </template>
      </Column>
      <Column headerStyle="width: 70px">
         <template #body="slotProps">
            <div class="centered">
               <router-link :to="imageViewerURL(slotProps.data)" @click="imageClicked"><img :src="slotProps.data.thumbURL"/></router-link>
            </div>
         </template>
      </Column>
      <Column header="Tag" headerStyle="width: 60px">
         <template #body="slotProps">
            <TagPicker :masterFile="slotProps.data" />
         </template>
      </Column>
      <Column header="File Name" field="fileName">
         <template #body="slotProps">
            <div class="filename">
               <span>{{ slotProps.data.fileName }}</span>
               <i v-if="slotProps.data.error" class="image-err pi pi-exclamation-circle" v-tooltip.bottom="{ value: slotProps.data.error, autoHide: false }"></i>
            </div>
         </template>
      </Column>
      <Column header="Title" field="title">
         <template #body="slotProps"><span class="editable">{{ slotProps.data.title }}</span></template>
         <template #editor="{ data, field }">
            <TitlePicker v-model="data[field]"/>
         </template>
      </Column>
      <Column header="Caption" field="description">
         <template #body="slotProps"><span class="editable">{{ slotProps.data.description }}</span></template>
         <template #editor="{ data, field }">
            <InputText v-model="data[field]" fluid />
         </template>
      </Column>
      <template v-if="projectStore.isManuscript">
         <Column :header="projectStore.detail.containerType.name" field="box" class="nowrap">
            <template #body="slotProps">
               <span  v-if="slotProps.data.box" class="editable">{{ slotProps.data.box }}</span>
               <span  v-else class="editable undefined">Undefined</span>
            </template>
            <template #editor="{ data, field }">
               <InputText v-model="data[field]" fluid />
            </template>
         </Column>
         <Column v-if="projectStore.detail.containerType.hasFolders" header="Folder" field="folder" class="nowrap">
            <template #body="slotProps">
               <span  v-if="slotProps.data.folder" class="editable">{{ slotProps.data.folder }}</span>
               <span  v-else class="editable undefined">Undefined</span>
            </template>
            <template #editor="{ data, field }">
               <InputText v-model="data[field]" fluid />
            </template>
         </Column>
      </template>
      <Column header="Component" field="component" class="nowrap">
         <template #body="slotProps">
            <span v-if="slotProps.data.componentID">{{slotProps.data.componentID}}</span>
            <span v-else class="undefined">N/A</span>
         </template>
      </Column>
      <Column header="Size" class="nowrap">
         <template #body="slotProps">{{slotProps.data.width}} x {{slotProps.data.height}}</template>
      </Column>
      <Column header="Resolution" field="resolution" class="nowrap"/>
      <Column header="Color Profile" field="colorProfile" class="nowrap"/>
      <Column rowReorder headerStyle="width: 3rem" :reorderableColumn="false" />
   </DataTable> -->
</template>

<script setup>
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import TagPicker from '@/components/TagPicker.vue'
import InputText from 'primevue/inputtext'
import { useProjectStore } from "@/stores/project"
import { useUnitStore } from "@/stores/unit"
import ViewMode from '@/components/ViewMode.vue'
import UnitActions from '@/components/unit/UnitActions.vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import TitlePicker from "@/components/TitlePicker.vue"
import { computed, h, ref, watch  } from 'vue'
import UCheckboxCell from './UCheckboxCell.vue'
import { useSortable } from '@vueuse/integrations/useSortable'

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
   justify-content: flex-start;
   gap: 10px;
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