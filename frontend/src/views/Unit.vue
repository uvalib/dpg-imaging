<template>
   <div class="unit">
      <WaitSpinner  v-if="unitStore.working" :overlay="true" message="Working..." />
      <div class="metadata" v-if="projectStore.hasDetail">
         <h2>
            <ProblemsDisplay class="topleft" />
            <span class="title"><router-link :to="`/projects/${projectStore.detail.id}`">{{truncateTitle(title)}}</router-link></span>
         </h2>
         <h3>
            <div>{{callNumber}}</div>
            <div>Unit {{unitStore.unitID}}</div>
            <div class="small" >{{workingDir}}</div>
            <div class="small" >{{unitStore.masterFiles.length}} Images</div>
         </h3>
         <div class="back">
            <UButton icon="i-lucide-arrow-left" label="Back to project" @click="backClicked" size="sm" color="secondary"/>
            <KeyboardShortcutHelp />
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
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MasterFilesList from '@/components/unit/MasterFilesList.vue'
import MasterFilesGrid from '@/components/unit/MasterFilesGrid.vue'
import KeyboardShortcutHelp from '@/components/KeyboardShortcutHelp.vue'
import { onKeyStroke } from '@vueuse/core'
import { useConfirm } from "../composables/useConfirm"

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const unitStore = useUnitStore()
const route = useRoute()
const router = useRouter()

onKeyStroke('d', (e) => {
   if ( e.ctrlKey ) {
      handleDelete()
   }
})
onKeyStroke('a', (e) => {
   if ( e.ctrlKey ) {
      unitStore.selectAll()
   }
})

const title = computed(() => {
   let t = projectStore.detail.title
   if ( t == "") {
      t = "Unknown"
   }
   return t
})

const callNumber = computed(() => {
   let t = projectStore.detail.callNumber
   if ( t == "") {
      t = "Unknown"
   }
   return t
})

const workingDir = computed(()=>{
   let unitDir =  paddedUnit(projectStore.detail.unitID)
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

const handleDelete = ( async () => {
   let msg = `<div style="display:flex; flex-direction: column; gap: 5px; align-items: flex-start;">
               <div>Delete the selected images? All data will be lost.</div>
               <div>This is not reversable.</div>
               <div>Are you sure?</div>
            </div>`
   const resp = await useConfirm("Confirm Image Delete", msg, "Delete")
   if (resp) {
      unitStore.deleteSelectedMasterFiles()
   } 
})

onMounted( async () => {
   unitStore.lastURL = ""

   if (projectStore.hasDetail == false) {
      await projectStore.getProject(route.params.id)
   }

   // set current page size and page, which is needed to get list of MF
   unitStore.pageSize = 20
   if ( route.query.pagesize ) {
      unitStore.pageSize = parseInt(route.query.pagesize, 10)
   }
   unitStore.currPage = 1
   if ( route.query.page ) {
      unitStore.currPage = parseInt(route.query.page, 10)
   }

   await unitStore.getMasterFiles( projectStore.detail )
   await unitStore.getMetadataPage( )
   if ( route.query.view ) {
      unitStore.viewMode = route.query.view
   }
})

</script>

<style lang="scss" scoped>
.unit {
   padding: 0;
   text-align: center;

   .metadata {
      margin-bottom: 15px;
      position: relative;
      .small {
         font-size: 0.95em;
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
          margin: 5px 0;
          font-weight: normal;
          .divider {
            border-bottom: 1px solid var(--uvalib-grey-light);
            margin: 10px auto 20px auto;
            width: 50%;
          }
      }
      .back {
         padding: 0 15px;
         text-align: left;
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
      }
   }
}
</style>
