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

      <div class="metadata" v-if="projectStore.hasDetail">
         <KeyboardShortcutHelp />
         <h2>
            <ProblemsDisplay class="topleft" />
            <span class="title"><router-link :to="`/projects/${projectStore.detail.id}`">{{truncateTitle(title)}}</router-link></span>
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
            <router-link :to="`/projects/${projectStore.detail.id}`" class="link">Back to project</router-link>
         </span>
      </div>

      <div class="toolbar">
         <ViewMode />

         <DPGPagination  v-if="unitStore.masterFiles.length > 20" :currPage="unitStore.currPage" :pageSize="unitStore.pageSize"
            :totalPages="unitStore.totalPages" :sizePicker="true"
            @next="nextClicked" @prior="priorClicked" @first="firstClicked" @last="lastClicked"
            @jump="pageJumpClicked" @size="pageSizeChanged"
         />

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
      <template v-if="projectStore.hasDetail">
         <MasterFilesList  v-if="unitStore.viewMode == 'list'" />
         <MasterFilesGrid  v-else />
      </template>
   </div>
</template>

<script setup>
import ComponentPanel from '../components/ComponentPanel.vue'
import BatchUpdatePanel from '../components/BatchUpdatePanel.vue'
import PageNumPanel from '../components/PageNumPanel.vue'
import ProblemsDisplay from '../components/ProblemsDisplay.vue'
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUnitStore} from "@/stores/unit"
import { computed, ref, onBeforeMount, onBeforeUnmount, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import DPGPagination from '../components/DPGPagination.vue'
import MasterFilesList from '../components/unit/MasterFilesList.vue'
import MasterFilesGrid from '../components/unit/MasterFilesGrid.vue'
import { useConfirm } from "primevue/useconfirm"
import ViewMode from '../components/ViewMode.vue'
import KeyboardShortcutHelp from '../components/KeyboardShortcutHelp.vue'
import RenameFiles from '../components/RenameFiles.vue'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const unitStore = useUnitStore()
const route = useRoute()
const router = useRouter()
const confirm = useConfirm()

const title = computed(() => {
   let t = projectStore.detail.unit.metadata.title
   if ( t == "") {
      t = "Unknown"
   }
   return t
})
const callNumber = computed(() => {
   let t = projectStore.detail.unit.metadata.callNumber
   if ( t == "") {
      t = "Unknown"
   }
   return t
})
const workingDir = computed(()=>{
   let unitDir =  paddedUnit(projectStore.detail.unit.id)
   if (projectStore.detail.currentStep.name == "Process" || projectStore.detail.currentStep.name == "Scan") {
      return `${systemStore.scanDir}/${unitDir}`
   }
   return `${systemStore.qaDir}/${unitDir}`
})
const isManuscript = computed(() => {
   if ( projectStore.hasDetail == false) return false
   return projectStore.detail.workflow && projectStore.detail.workflow.name=='Manuscript'
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

function paddedUnit() {
   let unitStr = ""+unitStore.currUnit
   return unitStr.padStart(9,'0')
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

const handleDelete = (() => {
   confirm.require({
      group: 'delete',
      header: 'Confirm Image Delete',
      accept: () => {
         unitStore.deleteSelectedMasterFiles()
      }
   })
})

const keyboardHandler = ((event) => {
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
})

onBeforeMount( async () => {
   // setup keyboard litener for shortcuts
   window.addEventListener('keydown', keyboardHandler)

   if (projectStore.hasDetail == false) {
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

   await unitStore.getUnitMasterFiles( projectStore.detail.unit.id )
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
         display: flex;
         flex-flow: row wrap;
         justify-content: flex-end;
         align-content: center;
      }
   }
}
</style>
