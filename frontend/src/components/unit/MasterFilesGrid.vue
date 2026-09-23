<template>
   <div class="grid-view sticky z-50" :style="{top: headerHeight}">
      <div class="control-group">
         <ViewMode />
         <UPagination  v-if="unitStore.totalFiles > unitStore.pageSize" color="neutral" variant="ghost"
            v-model:page="unitStore.currPage" :items-per-page="unitStore.pageSize" 
            :total="unitStore.totalFiles" @update:page="pageChanged"
         />
         <USelect v-model="unitStore.pageSize" :items="['20','50','75']" @change="pageChanged()" />
      </div>
      <UnitActions />
   </div>
   <div class="gallery">
      <UCard v-for="(image,idx) in unitStore.masterFilesPage" :key="image.fileName" :id="image.fileName">
         <template #header>
            <div class="card-title">
               <UCheckbox :modelValue="unitStore.masterFiles[idx].selected" @update:modelValue="unitStore.masterFileSelected(idx)"
                  size="xl" :label="image.fileName"
               />
               <UIcon name="i-lucide-move" class="size-6 cursor-grab text-brand-grey"/>
            </div>
         </template>
         <div class="content">
            <RouterLink :to="imageViewerURL(image)" @click="imageClicked">
               <img :src="image.mediumURL" v-if="unitStore.viewMode == 'medium'" />
               <img :src="image.largeURL" v-if="unitStore.viewMode == 'large'" />
            </RouterLink>
            <TagPicker :masterFile="image" display="wide" />
            <div class="metadata">
               <div class="row">
                  <label>Title:</label>
                  <template v-if="editInfo.field=='title' && image.fileName == editInfo.fileName">
                     <TitlePicker v-model="editInfo.value" @cancel="cancelEdit" @submit="submitEdit"/>
                  </template>
                  <ULink v-else @click="startEdit(idx, 'title', image)">
                     <span v-if="image.title">{{  image.title }}</span>
                     <span v-else class="undefined">Undefined</span>
                  </ULink>
               </div>   
               <div class="row">
                  <label>Caption:</label>
                  <template v-if="editInfo.field=='description' && image.fileName == editInfo.fileName">
                     <TitlePicker v-model="editInfo.value" @cancel="cancelEdit" @submit="submitEdit"/>
                  </template>
                  <ULink v-else @click="startEdit(idx, 'description', image)">
                     <span v-if="image.description">{{  image.description }}</span>
                     <span v-else  class="undefined">Undefined</span>
                  </ULink>
               </div> 
               <template v-if="projectStore.isManuscript">
                  <div class="row">
                     <label>Location:</label>
                     <div class="data pl-2">{{ image.location }}</div>
                  </div> 
                  <div class="row" v-if="image.componentID">
                     <label>Component</label>
                     <div class="data pl-2">{{ image.componentID }}</div>
                  </div> 
               </template>
            </div>
         </div>
      </UCard>
   </div>
</template>

<script setup>
import { useSortable } from '@vueuse/integrations/useSortable'
import TagPicker from '@/components/TagPicker.vue'
import { useProjectStore } from "@/stores/project"
import { useUnitStore } from "@/stores/unit"
import { ref, computed } from 'vue'
import ViewMode from './ViewMode.vue'
import UnitActions from '@/components/unit/UnitActions.vue'
import { useRoute, useRouter } from 'vue-router'
import TitlePicker from '../TitlePicker.vue'
import { onKeyStroke } from '@vueuse/core'

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const unitStore = useUnitStore()

const editInfo = ref({fileName: "", field: "", orig: "", value: ""})

onKeyStroke(['>','.'], () => {
   const overlay =  document.querySelector('div[data-slot="overlay"]')
   if (overlay) return 

   if (unitStore.currPage < unitStore.totalPages && editInfo.value.field == "") {
      unitStore.currPage++
      pageChanged()
   }
})
onKeyStroke(['<',','], () => {
   const overlay =  document.querySelector('div[data-slot="overlay"]')
   if (overlay) return 
   
   if (unitStore.currPage > 1 && editInfo.value.field == "") {
      unitStore.currPage--
      pageChanged()
   }
})

// you cannot custruct tailwind class values dynamically, so you  cant do `top-${hdr.clientHeight}`. 
// Instead use this to bind an inline style 'top' param to stick the controls below the header
const headerHeight = computed(() => {
   let hdr = document.querySelector('header')
   return `${hdr.clientHeight}px`
})


useSortable('.gallery', unitStore.masterFilesPage, {
   animation: 150,
   onEnd: (evt) => {
      unitStore.moveImage(evt.oldIndex, evt.newIndex)
   }
})

const imageClicked = (() => {
   unitStore.lastURL = route.fullPath
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

const imageViewerURL = ((img) => {
   const idx = unitStore.masterFiles.findIndex( mf => mf.fileName == img.fileName)
   return `/projects/${projectStore.detail.id}/unit/images/${idx+1}`
})

const cancelEdit = (() => {
   editInfo.value = {fileName: "", field: "", orig: "", value: ""}
})
const startEdit = ((idx, field,image) => {
   editInfo.value = {fileName: image.fileName, field: field,  rowIndex: idx, orig: image[field], value: image[field]}
})
const submitEdit = (() => {
   if ( editInfo.value.orig != editInfo.value.value) {
      unitStore.updateMasterFileMetadata( editInfo.value.fileName, editInfo.value.field, editInfo.value.value)
   }
   editInfo.value = {fileName: "", field: "", orig: "", value: ""}
})
</script>

<style lang="scss" scoped>
.grid-view {
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
.undefined {
   font-style: italic;
}

div.gallery {
   padding: 15px;
   text-align: left;
   background: #e5e5e5;
   display: flex;
   flex-flow: row wrap;
   justify-content: flex-start;
   align-content: flex-start;
   gap: 10px;

   .card-title {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: center;
   }

   .content {
      display: flex;
      flex-direction: column;
      gap: 10px;
      .metadata {
         text-align: left;
         font-size: 0.9em;
         display: flex;
         flex-direction: column;
         align-items: flex-start;
         .row {
            display: flex;
            flex-direction: column;
            align-items: flex-start; 
         }
         label {
            font-weight: bold;
         }

         div.data {
            margin: 5px 0 0 0;
            text-align: left;
         }
      }
   }
}

div.gallery.medium {
   .card .metadata .data {
      max-width: 230px;
   }
}

div.gallery.large {
   .card .metadata .data {
      max-width: 380px;
   }
}

.editable {
   cursor: pointer;

   &:hover {
      text-decoration: underline;
      color: var(--uvalib-blue-alt) !important;
   }
}
</style>
