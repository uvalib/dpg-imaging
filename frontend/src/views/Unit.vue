<template>
   <div class="unit">
      <WaitSpinner  v-if="unitStore.working" :overlay="true" message="Working..." />
      <ConfirmDialog group="delete">
         <template #message>
            <div style="display:flex; flex-direction: column; gap: 10px; align-items: flex-start;">
               <div>Delete the selected images? All data will be lost.</div>
               <div>This is not reversable.</div>
               <div>Are you sure?</div>
            </div>
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
            <div>Unit {{unitStore.unitID}}</div>
            <div class="divider"></div>
            <div class="small" >{{workingDir}}</div>
            <div class="small" >{{unitStore.masterFiles.length}} Images</div>
         </h3>
         <div class="back">
            <DPGButton icon="pi pi-angle-double-left" text label="Back to project" size="small" severity="secondary" @click="backClicked"/>
         </div>
      </div>
      <div class="master-files" ref="masterfiles">
         <MasterFilesList  v-if="unitStore.viewMode == 'list'" />
         <MasterFilesGrid  v-else />
      </div>
   </div>
</template>

<script setup>
import ProblemsDisplay from '@/components/ProblemsDisplay.vue'
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUnitStore} from "@/stores/unit"
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MasterFilesList from '@/components/unit/MasterFilesList.vue'
import MasterFilesGrid from '@/components/unit/MasterFilesGrid.vue'
import { useConfirm } from "primevue/useconfirm"
import KeyboardShortcutHelp from '@/components/KeyboardShortcutHelp.vue'

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

const backClicked = (() => {
   router.push( `/projects/${projectStore.detail.id}` )
})

function truncateTitle(title) {
   if (title.length < 200) return title
   return title.slice(0,200)+"..."
}

function paddedUnit() {
   let unitStr = ""+unitStore.unitID
   return unitStr.padStart(9,'0')
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
   if ( event.target.id == "edit-desc" || event.target.id == "title-input-box" ||
        unitStore.edit.pageNumber || unitStore.edit.metadata || unitStore.edit.component ) {
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
      unitStore.edit.pageNumber = true
   } else if (event.key == 'b' || event.key == 'f') {
      unitStore.edit.metadata = true
   } else if (event.key == 'k') {
      unitStore.edit.component = true
   }  else if (event.key == 'a') {
      unitStore.selectPage()
   }  else if (event.key == 'd') {
      handleDelete()
   }
})

onMounted( async () => {
   unitStore.lastURL = ""

   // setup keyboard litener for shortcuts
   window.addEventListener('keyup', keyboardHandler)
   // window.addEventListener("scroll", scrollHandler)

   if (projectStore.hasDetail == false) {
      await projectStore.getProject(route.params.id)
   }

   // set current page size and page, which is needed to get list of MF
   unitStore.pageSize = 20
   if ( route.query.pagesize ) {
      unitStore.pageSize = parseInt(route.query.pagesize, 10)
   }
   unitStore.currPage = 0
   if ( route.query.page ) {
      unitStore.currPage = parseInt(route.query.page, 10)
   }

   await unitStore.getMasterFiles( projectStore.detail )
   await unitStore.getMetadataPage( )
   if ( route.query.view ) {
      unitStore.viewMode = route.query.view
   }
})

onBeforeUnmount( async () => {
   window.removeEventListener('keyup', keyboardHandler)
})

</script>

<style lang="scss" scoped>
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
         text-align: left;
      }
   }
}
</style>
