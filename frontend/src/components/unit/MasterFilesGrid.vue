<template>
   <div ref="gallery" class="gallery">
         <Card class="card" v-for="(element,index) in unitStore.masterFilesPage" :id="element.fileName" :key="element.fileName">
            <template #title>
               <div class="card-title">
                  <div class="card-sel">
                     <input type="checkbox" v-model="element.selected" @click="masterFileCheckboxClicked(index)"/>
                     <div class="file">
                        <span>{{element.fileName}}</span>
                        <span v-if="element.error">
                           <i class="image-err fas fa-exclamation-circle" @mouseover="hoverEnter(element.fileName)" @mouseleave="hoverExit()"></i>
                           <span v-if="showError==element.fileName" class="hover-error">{{element.error}}</span>
                        </span>
                     </div>
                  </div>
                  <i class="grip pi pi-arrows-alt"></i>
               </div>
            </template>
            <template #content>
               <div class="content">
                  <router-link :to="imageViewerURL(index)">
                     <img :src="element.mediumURL" v-if="unitStore.viewMode == 'medium'"/>
                     <img :src="element.largeURL" v-if="unitStore.viewMode == 'large'"/>
                  </router-link>
                  <TagPicker :masterFile="element" display="wide"/>
                  <div class="metadata">
                     <div class="row">
                        <label>Title</label>
                        <div tabindex="0" @focus.stop.prevent="editMetadata(element, 'title')" class="data editable" @click="editMetadata(element, 'title')">
                           <template v-if="isEditing(element, 'title')">
                              <TitleInput  @canceled="cancelEdit" @accepted="submitEdit(element)" v-model="newValue" @blur.stop.prevent="cancelEdit"/>
                           </template>
                           <template v-else>
                              <template v-if="element.title">{{element.title}}</template>
                              <span v-else class="undefined">Undefined</span>
                           </template>
                        </div>
                     </div>
                     <div class="row">
                        <label>Caption</label>
                        <div class="data editable" tabindex="0" @focus.stop.prevent="editMetadata(element, 'description')"  @click="editMetadata(element, 'description')">
                           <template v-if="isEditing(element, 'description')">
                              <input id="edit-desc" type="text" v-model="newValue" @keyup.enter="submitEdit(element)" @keyup.esc="cancelEdit" @blur.stop.prevent="cancelEdit"/>
                           </template>
                           <template v-else>
                              <template v-if="element.description">{{element.description}}</template>
                              <span v-else class="undefined">Undefined</span>
                           </template>
                        </div>
                     </div>
                     <div class="row" v-if="element.box && element.box != '<nil>'">
                        <label>Box</label>
                        <div class="data">{{element.box}}</div>
                     </div>
                     <div class="row" v-if="element.folder && element.folder != '<nil>'">
                        <label>Folder</label>
                        <div class="data">{{element.folder}}</div>
                     </div>
                     <div class="row" v-if="element.componentID">
                        <label>Component</label>
                        <div class="data">{{element.componentID}}</div>
                     </div>
                  </div>
               </div>
            </template>
         </Card>
      </div>
</template>

<script setup>
import { useSortable } from '@vueuse/integrations/useSortable'
import TagPicker from '@/components/TagPicker.vue'
import TitleInput from '@/components/TitleInput.vue'
import Card from 'primevue/card'
import { useProjectStore } from "@/stores/project"
import { useUnitStore } from "@/stores/unit"
import { ref, nextTick } from 'vue'

const projectStore = useProjectStore()
const unitStore = useUnitStore()

const gallery = ref()
const editMF = ref(null)
const newValue = ref("")
const editField = ref("")
const showError = ref("")

useSortable(gallery,  unitStore.masterFilesPage, {
  animation: 150,
  onUpdate: (e) => {
   unitStore.moveImage(e.oldIndex, e.newIndex)
  }
})

const masterFileCheckboxClicked = ((index) => {
   unitStore.masterFileSelected(index)
})

const imageViewerURL = ((pgIndex) => {
   return `/projects/${projectStore.detail.id}/unit/images/${unitStore.pageStartIdx+pgIndex+1}`
})
const hoverExit = (() => {
   showError.value = ""
})
const hoverEnter = ((f) => {
   showError.value = f
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
   nextTick( ()=> {
      let ele = null
      if ( field == "description") {
         ele = document.getElementById("edit-desc")
      }
      if ( field == "folder") {
         ele = document.getElementById("edit-folder")
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

const submitEdit = ( async (mf) => {
   await unitStore.updateMasterFileMetadata( mf.fileName, editField.value, newValue.value )
   editMF.value = null
})
</script>

<style lang="scss" scoped>
.undefined {
   font-style: italic;
}

.hover-error {
   position: absolute;
   white-space: nowrap;
   background: white;
   padding: 5px 10px;
   border: 2px solid var(--uvalib-red-dark);
   border-radius: 4px;
   box-shadow: var(--box-shadow);
   bottom: 5px;
   z-index: 10000;
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
      .card-title {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: center;
         gap: 20px;
         border-bottom: 1px solid var(--uvalib-grey-light);
         padding-bottom: 10px;
         margin-bottom: 10px;
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
            margin: 5px 0 0 15px;
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
}</style>
