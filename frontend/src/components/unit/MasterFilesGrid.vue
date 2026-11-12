<template>
   <div class="sticky z-50 row-between p-4 bg-white border-t border-b border-brand-grey-100" :style="{top: headerHeight}">
      <div class="row-left gap-2">
         <ViewMode />
         <UPagination  v-if="unitStore.totalFiles > unitStore.pageSize" color="neutral" variant="ghost"
            v-model:page="unitStore.currPage" :items-per-page="unitStore.pageSize" 
            :total="unitStore.totalFiles" @update:page="pageChanged"
         />
         <USelect v-model="unitStore.pageSize" :items="['20','50','75']" @change="pageChanged()" />
      </div>
      <UnitActions />
   </div>
   <div class="gallery row-wrap gap-6 p-6 bg-gray-100">
      <UCard v-for="(image,idx) in unitStore.masterFilesPage" :key="image.fileName" :id="image.fileName">
         <template #header>
            <div class="row-between">
               <UCheckbox :modelValue="unitStore.masterFiles[idx].selected" @update:modelValue="unitStore.masterFileSelected(idx)"
                  size="xl" :label="image.fileName"
               />
               <UIcon name="i-lucide-move" class="size-6 cursor-grab text-brand-grey"/>
            </div>
         </template>
         <div class="content" :class="{medium: unitStore.viewMode == 'medium', large: unitStore.viewMode == 'large'}">
            <RouterLink :to="imageViewerURL(image)" @click="imageClicked">
               <img :src="image.mediumURL" v-if="unitStore.viewMode == 'medium'" />
               <img :src="image.largeURL" v-if="unitStore.viewMode == 'large'" />
            </RouterLink>
            <TagPicker :masterFile="image" display="wide" />
            <div class="column gap-1 items-start pt-4">
               <div class="column items-start">
                  <label>Title:</label>
                  <template v-if="editInfo.field=='title' && image.fileName == editInfo.fileName">
                     <TitlePicker v-model="editInfo.value" @cancel="cancelEdit" @submit="submitEdit"/>
                  </template>
                  <ULink v-else @click="startEdit(idx, 'title', image)" class="text-left whitespace-break-spaces pl-2">
                     <template v-if="image.title">{{  image.title }}</template>
                     <template v-else class="italic">Undefined</template>
                  </ULink>
               </div>   
               <div class="column items-start">
                  <label>Caption:</label>
                  <template v-if="editInfo.field=='description' && image.fileName == editInfo.fileName">
                     <TitlePicker v-model="editInfo.value" @cancel="cancelEdit" @submit="submitEdit"/>
                  </template>
                  <ULink v-else @click="startEdit(idx, 'description', image)" class="text-left whitespace-break-spaces pl-2">
                     <template v-if="image.description">{{  image.description }}</template>
                     <template v-else  class="italic">Undefined</template>
                  </ULink>
               </div> 
               <template v-if="projectStore.isManuscript">
                  <div class="column items-start">
                     <label>Location:</label>
                     <div class="data pl-4">{{ image.location }}</div>
                  </div> 
                  <div class="column items-start" v-if="image.componentID">
                     <label>Component</label>
                     <div class="data pl-4">{{ image.componentID }}</div>
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
.content label {
   font-weight: bold;
   font-size: 0.9em;
}
.content.medium {
   max-width: 250px;
}
.content.large {
   max-width: 400px;
}
</style>
