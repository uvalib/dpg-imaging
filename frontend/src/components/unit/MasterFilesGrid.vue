<template>
   <Sortable :list="unitStore.masterFilesPage" item-key="fileName" @end="moveImage" class="gallery" :class="unitStore.viewMode" >
      <template #item="{element, index}">
         <Card class="card" :id="element.fileName">
            <template #content>
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
               <router-link :to="imageViewerURL(index)">
                  <img :src="element.mediumURL" v-if="unitStore.viewMode == 'medium'"/>
                  <img :src="element.largeURL" v-if="unitStore.viewMode == 'large'"/>
               </router-link>
               <div class="tag">
                  <TagPicker :masterFile="element" display="wide"/>
               </div>
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
                  <div class="row" v-if="element.box">
                     <label>Box</label>
                     <div class="data">{{element.box}}</div>
                  </div>
                  <div class="row" v-if="element.folder">
                     <label>Folder</label>
                     <div class="data">{{element.folder}}</div>
                  </div>
                  <div class="row" v-if="element.componentID">
                     <label>Component</label>
                     <div class="data">{{element.componentID}}</div>
                  </div>
               </div>
            </template>
         </Card>
      </template>
   </Sortable>
</template>

<script setup>
import { Sortable } from "sortablejs-vue3"
import TagPicker from '@/components/TagPicker.vue'
import TitleInput from '@/components/TitleInput.vue'
import Card from 'primevue/card'
import { useProjectStore } from "@/stores/project"
import { useUnitStore } from "@/stores/unit"
import { ref, nextTick } from 'vue'

const projectStore = useProjectStore()
const unitStore = useUnitStore()

const editMF = ref(null)
const newValue = ref("")
const editField = ref("")
const showError = ref("")

const moveImage = ((event) => {
   unitStore.moveImage(event.oldIndex, event.newIndex)
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

   .card {
      position: relative;
      padding: 0 20px 20px 20px;
      margin: 5px;

      input[type=checkbox] {
         width: 20px;
         height: 20px;
      }

      .card-sel {
         padding: 0 0 20px 0;
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: end;

         input[type=checkbox] {
            margin: 0;
            padding: 0;
            display: inline-block;
            margin-right: 15px;
         }
      }

      .metadata {
         text-align: left;
         font-size: 0.9em;
         padding: 0 5px 5px 5px;

         label {
            font-weight: bold;
            display: block;
            margin-top: 5px;
         }

         div.data {
            margin: 5px 0 0 15px;
            font-weight: normal;
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
