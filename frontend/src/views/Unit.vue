<template>
   <div class="unit">
      <WaitSpinner  v-if="systemStore.working" :overlay="true" message="Working..." />
      <ConfirmDialog group="delete">
         <template #message>
            <div>Delete the selected images? All data will be lost.</div>
            <div>This is not reversable.</div>
            <div class="sure">Are you sure?</div>
         </template>
      </ConfirmDialog>

      <div class="metadata" v-if="projectStore.selectedProjectIdx > -1">
         <KeyboardShortcutHelp />
         <h2>
            <ProblemsDisplay class="topleft" />
            <span class="title"><router-link :to="`/projects/${projectStore.currProject.id}`">{{truncateTitle(title)}}</router-link></span>
         </h2>
         <h3>
            <div>{{callNumber}}</div>
            <div>Unit {{unitStore.currUnit}}</div>
            <div class="divider"></div>
            <div class="small" >{{workingDir}}</div>
            <div class="small" >{{unitStore.masterFiles.length}} Images</div>
         </h3>
         <span class="back">
            <i class="fas fa-angle-double-left back-button"></i>
            <router-link :to="`/projects/${projectStore.currProject.id}`" class="link">Back to project</router-link>
         </span>
      </div>

      <div class="toolbar">
         <ViewMode />
         <span class="actions">
            <RenameFiles />
            <DPGButton @click="setPageNumbersClicked" class="p-button-secondary right-pad" label="Set Page Numbers"/>
            <DPGButton @click="titleClicked" class="p-button-secondary right-pad" label="Set Title"/>
            <DPGButton @click="descClicked" class="p-button-secondary right-pad" label="Set Caption"/>
            <template v-if="isManuscript">
               <DPGButton @click="boxClicked" class="p-button-secondary right-pad" label="Set Box"/>
               <DPGButton @click="folderClicked" class="p-button-secondary right-pad" label="Set Folder"/>
            </template>
            <DPGButton @click="componentLinkClicked" class="p-button-secondary" label="Set Component"/>
         </span>
      </div>
      <PageNumPanel v-if="unitStore.editMode == 'page'" />
      <ComponentPanel v-if="unitStore.editMode == 'component'" />
      <BatchUpdatePanel v-if="showBatchUpdate" :title="batchUpdateTitle" :field="batchUpdateField" :global="unitStore.editMode=='box'" />
      <table class="unit-list" v-if="unitStore.viewMode == 'list'">
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
         <draggable v-model="unitStore.pageMasterFiles" tag="tbody"  @start="dragStarted" item-key="fileName">
            <template #item="{element, index}">
               <tr :id="element.fileName" :key="element.fileName">
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
                  <td class="grip"><i class="fas fa-grip-lines"></i></td>
               </tr>
            </template>
         </draggable>
      </table>

      <draggable v-else v-model="unitStore.pageMasterFiles" @start="dragStarted" class="gallery" :class="unitStore.viewMode" item-key="fileName">
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
      </draggable>
      <div class="footer" v-if="unitStore.masterFiles.length > 20">
         <DPGPagination :currPage="unitStore.currPage" :pageSize="unitStore.pageSize"
            :totalPages="unitStore.totalPages" :sizePicker="true"
            @next="nextClicked" @prior="priorClicked" @first="firstClicked" @last="lastClicked"
            @jump="pageJumpClicked" @size="pageSizeChanged"
         />
      </div>
   </div>
</template>

<script setup>
import ComponentPanel from '../components/ComponentPanel.vue'
import BatchUpdatePanel from '../components/BatchUpdatePanel.vue'
import PageNumPanel from '../components/PageNumPanel.vue'
import TagPicker from '../components/TagPicker.vue'
import TitleInput from '../components/TitleInput.vue'
import ProblemsDisplay from '../components/ProblemsDisplay.vue'
import draggable from 'vuedraggable'
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUnitStore} from "@/stores/unit"
import { computed, ref, onBeforeMount, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DPGPagination from '../components/DPGPagination.vue'
import { useConfirm } from "primevue/useconfirm"
import Card from 'primevue/card'
import ViewMode from '../components/ViewMode.vue'
import KeyboardShortcutHelp from '../components/KeyboardShortcutHelp.vue'
import RenameFiles from '../components/RenameFiles.vue'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const unitStore = useUnitStore()
const route = useRoute()
const router = useRouter()
const confirm = useConfirm()

// local data
const editMF = ref(null)
const newValue = ref("")
const editField = ref("")
const showError = ref("")

// computed
const title = computed(() => {
   let t = projectStore.currProject.unit.metadata.title
   if ( t == "") {
      t = "Unknown"
   }
   return t
})
const callNumber = computed(() => {
   let t = projectStore.currProject.unit.metadata.callNumber
   if ( t == "") {
      t = "Unknown"
   }
   return t
})
const workingDir = computed(()=>{
   let unitDir =  paddedUnit(projectStore.currProject.unit.id)
   if (projectStore.currProject.currentStep.name == "Process" || projectStore.currProject.currentStep.name == "Scan") {
      return `${systemStore.scanDir}/${unitDir}`
   }
   return `${systemStore.qaDir}/${unitDir}`
})
const isManuscript = computed(() => {
   return projectStore.currProject.workflow && projectStore.currProject.workflow.name=='Manuscript'
})
const showBatchUpdate = computed(() => {
   return ( unitStore.editMode == "box" || unitStore.editMode == "folder" || unitStore.editMode == "title" || unitStore.editMode == "description")
})
const batchUpdateTitle = computed(() => {
   if ( unitStore.editMode == "box") {
      return "Box"
   }
   if ( unitStore.editMode == "folder") {
      return "Folder"
   }
   if ( unitStore.editMode == "title") {
      return "Title"
   }
   if ( unitStore.editMode == "description") {
      return "Description"
   }

   return ""
})
const batchUpdateField = computed(() => {
   return unitStore.editMode
})

function truncateTitle(title) {
   if (title.length < 200) return title
   return title.slice(0,200)+"..."
}
function masterFileCheckboxClicked(index) {
   unitStore.masterFileSelected(index)
}

function paddedUnit() {
   let unitStr = ""+unitStore.currUnit
   return unitStr.padStart(9,'0')
}
function imageViewerURL(pgIndex) {
   return `/projects/${projectStore.currProject.id}/unit/images/${unitStore.pageStartIdx+pgIndex+1}`
}
async function priorClicked() {
   unitStore.setPage(unitStore.currPage-1)
   pageChanged()
}
async function nextClicked() {
   await unitStore.setPage(unitStore.currPage+1)
   pageChanged()
}
async function lastClicked() {
   await unitStore.setPage(unitStore.totalPages)
   pageChanged()
}
async function firstClicked() {
   await unitStore.setPage(1)
   pageChanged()
}
async function pageJumpClicked(pg) {
   await unitStore.setPage(pg)
   pageChanged()
}
function pageSizeChanged(sz) {
   unitStore.setPageSize(sz)
   let query = Object.assign({}, route.query)
   query.pagesize = unitStore.pageSize
   query.page = unitStore.currPage
   router.push({query})
}
function pageChanged() {
   let query = Object.assign({}, route.query)
   query.page = unitStore.currPage
   router.push({query})
}
function hoverExit() {
   showError.value = ""
}
function hoverEnter(f) {
   showError.value = f
}
function boxClicked() {
   unitStore.editMode = "box"
}
function folderClicked() {
   unitStore.editMode = "folder"
}
function componentLinkClicked() {
   unitStore.editMode = "component"
}
function titleClicked() {
   unitStore.editMode = "title"
}
function descClicked() {
   unitStore.editMode = "description"
}
function setPageNumbersClicked() {
   unitStore.editMode = "page"
}
function dragStarted() {
   let eles=document.getElementsByClassName("selected")
   while (eles[0]) {
      eles[0].classList.remove('selected')
   }
}
function isEditing(mf, field) {
   return editMF.value == mf && editField.value == field
}
function editMetadata(mf, field) {
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
}
function cancelEdit() {
   editMF.value = null
}
async function submitEdit(mf) {
   await unitStore.updateMasterFileMetadata( mf.fileName, editField.value, newValue.value )
   editMF.value = null
}

function handleDelete() {
   confirm.require({
      group: 'delete',
      header: 'Confirm Image Delete',
      accept: () => {
         unitStore.deleteSelectedMasterFiles()
      }
   })
}


function keyboardHandler(event) {
   if (event.key == 'Escape') {
      systemStore.error = ""
      unitStore.editMode = ""
      return
   }

   if (event.target.id == "edit-desc" || event.target.id == "title-input-box") {
      return
   }

   if ( event.key == ',' || event.key == '<') {
      if (unitStore.currPage > 1) {
         priorClicked()
         return
      }
   }
   if ( event.key == '.' || event.key == '>') {
      if (unitStore.currPage < unitStore.totalPages) {
         nextClicked()
         return
      }
   }

   if ( !event.ctrlKey ) return

   if (event.key == 'p') {
      setPageNumbersClicked()
   } else if (event.key == 'b') {
      boxClicked()
   } else if (event.key == 'f') {
      folderClicked()
   } else if (event.key == 'k') {
      componentLinkClicked()
   }  else if (event.key == 'a') {
      unitStore.selectPage()
   }  else if (event.key == 'd') {
      handleDelete()
   }
}

onBeforeMount( async () => {
   // setup keyboard litener for shortcuts
   window.addEventListener('keydown', keyboardHandler)

   if (projectStore.selectedProjectIdx == -1) {
      await projectStore.getProject(route.params.id)
   }

   // set current page size and page, which is needed to get list of MF
   if ( route.query.pagesize ) {
      let ps = parseInt(route.query.pagesize, 10)
      unitStore.setPageSize(ps)
   } else {
       unitStore.setPageSize(20)
   }
   if ( route.query.page ) {
      let pg = parseInt(route.query.page, 10)
      unitStore.setPage(pg)
   } else {
      unitStore.setPage(1)
   }

   await unitStore.getUnitMasterFiles( projectStore.currProject.unit.id )
   await unitStore.getMetadataPage()
   if ( route.query.view ) {
      unitStore.viewMode = route.query.view
   }
})
onBeforeUnmount( async () => {
   window.removeEventListener('keydown', keyboardHandler)
})
</script>

<style lang="scss" scoped>
   .sure {
      text-align: right;
      margin-top: 15px;
   }
.unit {
   padding: 0;
   input[type=checkbox] {
      width: 20px;
      height: 20px;
   }
   .load {
      margin-top: 15%;
   }
   .metadata {
      margin-bottom: 15px;
      position: relative;
      .small {
         font-size: 0.85em;
         margin-bottom: 5px;
      }
      .topleft {
         position: absolute;
         top:0;
         left: 0px;
      }
      h2 {
         .title {
            display: block;
            margin: 0 200px;
            a {
               color: inherit !important;
               font-weight: inherit !important;
               font-size: inherit !important;
            }
         }
      }
      h3 {
          margin: 5px 0 25px 0;
          font-weight: normal;
          .divider {
            border-bottom: 1px solid var(--uvalib-grey-light);
            margin: 10px auto 20px auto;
            width: 50%;
          }
      }
      a {
         display: inline-block;
         margin-top: 8px;
         font-weight: bold;
         text-decoration: none;
         cursor: pointer;
         color: var(--uvalib-blue-alt);
         .link {
            display: inline-block;
            margin-left: 8px;
         }
         &:hover {
            text-decoration: underline;
         }
      }
      .back {
         position: absolute;
         left: 10px;
         bottom: -15px;
         .link {
            font-weight: normal;
            text-decoration: none;
            color: var(--uvalib-text) !important;
            display: inline-block;
            margin-left: 5px;
            cursor: pointer;
            &:hover {
               text-decoration: underline ;
            }
         }
      }
   }

   .footer {
      display: flex;
      flex-flow: row wrap;
      justify-content: center;
      align-content: center;
      padding: 10px;
      border-top: 1px solid var(--uvalib-grey-light);
   }
   .toolbar {
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      align-content: center;
      padding: 10px;
      background: var(--uvalib-grey-light);
      border-bottom: 1px solid var(--uvalib-grey);
      border-top: 1px solid var(--uvalib-grey);

      .actions {
         margin-left: auto;
         display: flex;
         flex-flow: row wrap;
         justify-content: flex-end;
         align-content: center;
      }
   }
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

         .card-sel {
            padding: 0 0 20px 0;
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-start;
            align-items: end;

            input[type=checkbox] {
               margin:0;
               padding:0;
               display: inline-block;
               margin-right: 15px;
            }
         }
         .metadata {
            text-align: left;
            font-size: 0.9em;
            padding: 0 5px 5px 5px;
            label{
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
            background-repeat:no-repeat;
            background-position: center center;
            background-color: #f5f5f5;
         }
      }
   }

   div.gallery.medium {
      .card .metadata .data  {
         max-width: 230px;
      }
   }
   div.gallery.large {
      .card .metadata .data  {
         max-width: 380px;
      }
   }
   table.unit-list {
      border-collapse: collapse;
      width: 100%;
      font-size: 0.9em;
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
         border-bottom: 1px solid var(--uvalib-grey);
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
}
</style>
