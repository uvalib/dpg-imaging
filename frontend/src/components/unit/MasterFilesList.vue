<template>
   <table class="unit-list">
      <thead>
         <tr>
            <th></th>
            <th></th>
            <th>Tag</th>
            <th>File Name</th>
            <th>Title</th>
            <th>Caption</th>
            <template v-if="isManuscript">
               <th>Box</th>
               <th>Folder</th>
            </template>
            <th>Component</th>
            <th>Size</th>
            <th>Resolution</th>
            <th>Color Profile</th>
            <th></th>
         </tr>
      </thead>
      <tbody ref="masterfiles">
         <tr v-for="(element,index) in unitStore.masterFilesPage" :id="element.fileName" :key="element.fileName">
            <td><input type="checkbox" v-model="element.selected" @click="masterFileCheckboxClicked(index)"/></td>
            <td class="thumb">
               <router-link :to="imageViewerURL(index)"><img :src="element.thumbURL"/></router-link>
            </td>
            <td><TagPicker :masterFile="element" /></td>
            <td class="hover-container">
               <span>{{element.fileName}}</span>
               <span v-if="element.error">
                  <i class="image-err fas fa-exclamation-circle" @mouseover="hoverEnter(element.fileName)" @mouseleave="hoverExit()"></i>
                  <span v-if="showError==element.fileName" class="hover-error">{{element.error}}</span>
               </span>
            </td>
            <td @click="editMetadata(element, 'title')" class="editable nowrap" tabindex="0" @focus.stop.prevent="editMetadata(element, 'title')" >
               <span v-if="!isEditing(element, 'title')"  class="editable">
                  <span v-if="element.title">{{element.title}}</span>
                  <span v-else class="undefined">Undefined</span>
               </span>
               <TitleInput v-else @canceled="cancelEdit" @accepted="submitEdit(element)" v-model="newValue"  @blur.stop.prevent="cancelEdit"/>
            </td>
            <td @click="editMetadata(element, 'description')" class="editable nowrap" >
               <span  tabindex="0" @focus.stop.prevent="editMetadata(element, 'description')" v-if="!isEditing(element, 'description')" class="editable">
                  <span v-if="element.description">{{element.description}}</span>
                  <span v-else class="undefined">Undefined</span>
               </span>
               <input v-else id="edit-desc" type="text" v-model="newValue"
                  @keyup.enter="submitEdit(element)"  @keyup.esc="cancelEdit"  @blur.stop.prevent="cancelEdit"/>
            </td>
            <template v-if="isManuscript">
               <td>
               <span v-if="element.box">{{element.box}}</span>
               <span v-else class="undefined">Undefined</span>
            </td>
               <td @click="editMetadata(element, 'folder')" class="editable nowrap" tabindex="0" @focus.stop.prevent="editMetadata(element, 'folder')" >
                  <span v-if="!isEditing(element, 'folder')"  class="editable">
                     <span v-if="element.folder">{{element.folder}}</span>
                     <span v-else class="undefined">Undefined</span>
                  </span>
                  <input v-else id="edit-folder" type="text" v-model="newValue" @keyup.enter="submitEdit(element)"  @keyup.esc="cancelEdit"  @blur.stop.prevent="cancelEdit"/>
               </td>
            </template>
            <td>
               <span v-if="element.componentID">{{element.componentID}}</span>
               <span v-else>N/A</span>
            </td>
            <td class="nowrap">{{element.width}} x {{element.height}}</td>
            <td>{{element.resolution}}</td>
            <td class="nowrap">{{element.colorProfile}}</td>
            <td class="grip"><i class="pi pi-bars"></i></td>
         </tr>
      </tbody>
   </table>
</template>

<script setup>
import TagPicker from '@/components/TagPicker.vue'
import TitleInput from '@/components/TitleInput.vue'
import { useProjectStore } from "@/stores/project"
import {useUnitStore} from "@/stores/unit"
import { computed, ref, nextTick } from 'vue'
import { useSortable } from '@vueuse/integrations/useSortable'

const projectStore = useProjectStore()
const unitStore = useUnitStore()

const masterfiles = ref()
const editMF = ref(null)
const newValue = ref("")
const editField = ref("")
const showError = ref("")

useSortable(masterfiles,  unitStore.masterFilesPage, {
  animation: 150,
  handle: '.grip',
  onUpdate: (e) => {
   unitStore.moveImage(e.oldIndex, e.newIndex)
  }
})

const isManuscript = computed(() => {
   if ( projectStore.hasDetail == false) return false
   return projectStore.detail.workflow && projectStore.detail.workflow.name=='Manuscript'
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
   .hover-container {
      position: relative;
      .image-err {
         color: var(--uvalib-red);
         margin: 0 5px;
         cursor: pointer;
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
   }
   table.unit-list {
      border-collapse: collapse;
      width: 100%;
      font-size: 0.9em;
      input[type=checkbox] {
         width: 20px;
         height: 20px;
      }

      th {
         background-color: var(--uvalib-grey-lightest);
      }
      th,td {
         padding: 5px 10px;
         text-align: left;
         border-bottom: 1px solid var(--uvalib-grey-lightest);
         cursor:  default;
      }
      td.thumb {
         padding: 5px 5px 2px 5px;
         img {
            border:1px solid var(--uvalib-grey);
         }
      }
      td.grip {
         color: var(--uvalib-grey);
         cursor:  grab;
      }
      th {
         border-bottom: 1px solid var(--uvalib-grey-light);
      }
   }
   .editable {
      cursor: pointer;
      &:hover {
         text-decoration: underline;
         color: var(--uvalib-blue-alt) !important;
      }
   }
   .nowrap {
      white-space: nowrap;
   }
</style>
