<template>
   <DataTable :value="unitStore.masterFiles" ref="mfTable" id="mf-table" dataKey="fileName"
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
   </DataTable>
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
import { useRoute, useRouter } from 'vue-router'
import { usePinnable } from '@/composables/pin'
import TitlePicker from "@/components/TitlePicker.vue"

usePinnable("mf-table")

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const unitStore = useUnitStore()

const imageClicked = (() => {
   unitStore.lastURL = route.fullPath
})

const pageChanged = ((event) => {
   unitStore.deselectAll()
   unitStore.pageSize = event.rows
   unitStore.currPage = event.page
   let query = Object.assign({}, route.query)
   query.pagesize = event.rows
   query.page = event.page
   router.push({query})
   unitStore.getMetadataPage()
})

const onRowReorder = ( (event) => {
   unitStore.masterFiles = event.value
})

const masterFileCheckboxClicked = ((img) => {
   const idx = unitStore.masterFiles.findIndex( mf => mf.fileName == img.fileName)
   unitStore.masterFileSelected(idx)
})

const imageViewerURL = ((img) => {
   const idx = unitStore.masterFiles.findIndex( mf => mf.fileName == img.fileName)
   return `/projects/${projectStore.detail.id}/unit/images/${idx+1}`
})

const onCellEditComplete = ( (event) => {
   let { data, newValue, field } = event
   if ( data[field] != newValue) {
      unitStore.updateMasterFileMetadata( data.fileName, field, newValue)
   }
})
</script>

<style lang="scss" scoped>
.centered {
   display: flex;
   flex-flow: row nowrap;
   justify-content: center;
   padding: 5px 5px 2px 5px;
   img {
      border:1px solid var(--uvalib-grey);
   }
}
.editable {
   cursor: pointer;
   &:hover {
      text-decoration: underline;
      color: var(--uvalib-blue-alt) !important;
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
   color: var(--uvalib-grey-light);
}
.nowrap {
   white-space: nowrap;
}
</style>