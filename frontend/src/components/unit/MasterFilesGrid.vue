<template>
   <DataView :value="unitStore.masterFiles" id="mf-grid" paginatorPosition="top"
      :lazy="false" :rows="unitStore.pageSize" :first="unitStore.currStartIndex" :rowsPerPageOptions="[20,50,75]" paginator
      paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
      currentPageReportTemplate="{currentPage} of {totalPages}" @page="pageChanged"
      stateStorage="session" stateKey="dpg-paging"
   >
      <template #paginatorstart>
         <ViewMode />
      </template>
      <template #paginatorend>
         <UnitActions />
      </template>
      <template #list="slotProps">
         <div ref="gallery" class="gallery">
            <Card class="card" v-for="image in slotProps.items" :key="image.fileName" :id="image.fileName">
               <template #title>
                  <div class="card-title">
                     <div class="card-sel">
                        <input type="checkbox" v-model="image.selected" @click="masterFileCheckboxClicked(image)" />
                        <div class="file">
                           <span>{{ image.fileName }}</span>
                           <i v-if="image.error" class="image-err pi pi-exclamation-circle" v-tooltip.bottom="{ value: image.error, autoHide: false }"></i>
                        </div>
                     </div>
                     <i class="grip pi pi-arrows-alt"></i>
                  </div>
               </template>
               <template #content>
                  <div class="content">
                     <router-link :to="imageViewerURL(image)" @click="imageClicked">
                        <img :src="image.mediumURL" v-if="unitStore.viewMode == 'medium'" />
                        <img :src="image.largeURL" v-if="unitStore.viewMode == 'large'" />
                     </router-link>
                     <TagPicker :masterFile="image" display="wide" />
                     <div class="metadata">
                        <div class="row">
                           <label>Title</label>
                           <div tabindex="0" @focus.stop.prevent="editMetadata(image, 'title')"  class="data editable" @click="editMetadata(image, 'title')">
                              <template v-if="isEditing(image, 'title')">
                                 <Select id="title-edit" v-model="newValue" fluid editable :options="system.titleVocab" @keydown.enter="submitEdit(image)" @keydown.tab="cancelEdit"/>
                              </template>
                              <template v-else>
                                 <template v-if="image.title">{{ image.title }}</template>
                                 <span v-else class="undefined">Undefined</span>
                              </template>
                           </div>
                        </div>
                        <div class="row">
                           <label>Caption</label>
                           <div class="data editable" tabindex="0"
                              @focus.stop.prevent="editMetadata(image, 'description')"
                              @click="editMetadata(image, 'description')">
                              <template v-if="isEditing(image, 'description')">
                                 <input id="edit-desc" type="text" v-model="newValue" @keyup.enter="submitEdit(image)"
                                    @keydown.stop.prevent.esc="cancelEdit" @blur.stop.prevent="cancelEdit" />
                              </template>
                              <template v-else>
                                 <template v-if="image.description">{{ image.description }}</template>
                                 <span v-else class="undefined">Undefined</span>
                              </template>
                           </div>
                        </div>
                        <template v-if="projectStore.isManuscript">
                           <div class="row" v-if="image.box && image.box != '<nil>'">
                              <label>Box</label>
                              <div class="data">{{ image.box }}</div>
                           </div>
                           <div class="row" v-if="image.folder && image.folder != '<nil>'">
                              <label>Folder</label>
                              <div class="data">{{ image.folder }}</div>
                           </div>
                        </template>
                        <div class="row" v-if="image.componentID">
                           <label>Component</label>
                           <div class="data">{{ image.componentID }}</div>
                        </div>
                     </div>
                  </div>
               </template>
            </Card>
         </div>
      </template>
   </DataView>
</template>

<script setup>
import DataView from 'primevue/DataView'
import { useSortable } from '@vueuse/integrations/useSortable'
import TagPicker from '@/components/TagPicker.vue'
import Card from 'primevue/card'
import { useProjectStore } from "@/stores/project"
import { useUnitStore } from "@/stores/unit"
import { useSystemStore } from "@/stores/system"
import { ref, nextTick } from 'vue'
import Select from 'primevue/select'
import ViewMode from '@/components/ViewMode.vue'
import UnitActions from '@/components/unit/UnitActions.vue'
import { useRoute, useRouter } from 'vue-router'
import { usePinnable } from '@/composables/pin'

usePinnable("mf-grid")

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const unitStore = useUnitStore()
const system = useSystemStore()

const gallery = ref()
const editMF = ref(null)
const newValue = ref("")
const editField = ref("")

useSortable(gallery, unitStore.masterFiles, {
   animation: 150,
   onUpdate: (e) => {
      let pageStartIdx = unitStore.currPage * unitStore.pageSize
      unitStore.moveImage(pageStartIdx+e.oldIndex, pageStartIdx+e.newIndex)
   }
})

const imageClicked = (() => {
   unitStore.lastURL = route.fullPath
})

const pageChanged = ((event) => {
   unitStore.deselectAll()
   unitStore.pageSize = event.rows
   unitStore.currPage = event.page
   let query = Object.assign({}, route.query)
   query.pagesize = unitStore.pageSize
   query.page = unitStore.currPage
   router.push({query})
   unitStore.getMetadataPage()
})

const masterFileCheckboxClicked = ((img) => {
   const idx = unitStore.masterFiles.findIndex( mf => mf.fileName == img.fileName)
   unitStore.masterFileSelected(idx)
})

const imageViewerURL = ((img) => {
   const idx = unitStore.masterFiles.findIndex( mf => mf.fileName == img.fileName)
   return `/projects/${projectStore.detail.id}/unit/images/${idx+1}`
})

const isEditing = ((mf, field) => {
   return editMF.value == mf && editField.value == field
})
const editMetadata = ((mf, field) => {
   editMF.value = mf
   editField.value = field
   if (field == "title") {
      newValue.value = mf.title
   }
   if (field == "description") {
      newValue.value = mf.description
   }
   if (field == "folder") {
      newValue.value = mf.folder
   }
   nextTick(() => {
      let ele = null
      if (field == "description") {
         ele = document.getElementById("edit-desc")
      }
      if (field == "folder") {
         ele = document.getElementById("edit-folder")
      }
      if ( field == "title") {
         ele = document.querySelector("#title-edit .p-select-label")
      }
      if (ele) {
         ele.focus()
         ele.select()
      }
   })
})
const cancelEdit = (() => {
   editMF.value = null
})

const submitEdit = (async (mf) => {
   await unitStore.updateMasterFileMetadata(mf.fileName, editField.value, newValue.value)
   editMF.value = null
})
</script>

<style lang="scss" scoped>
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

   .card {
      position: relative;
      padding: 0;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 5px;

      .card-title {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: center;
         gap: 20px;
         border-bottom: 1px solid var(--uvalib-grey-light);
         padding-bottom: 10px;
         margin-bottom: 10px;

         .file {
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

         .grip {
            font-size: 1.15em;
            color: #aaa;
            cursor: grab;
         }
      }

      .card-sel {
         padding: 0;
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: center;
         font-size: 0.8em;
         gap: 5px;

         input[type=checkbox] {
            width: 20px;
            height: 20px;
         }
      }

      .content {
         display: flex;
         flex-direction: column;
         gap: 10px;
      }

      .metadata {
         text-align: left;
         font-size: 0.9em;
         display: flex;
         flex-direction: column;
         gap: 10px;

         label {
            font-weight: bold;
         }

         div.data {
            margin: 5px 0 0 0;
            text-align: left;
         }
      }

      img {
         background-image: url('/src/assets/dots.gif');
         background-repeat: no-repeat;
         background-position: center center;
         background-color: #f5f5f5;
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
