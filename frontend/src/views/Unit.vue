<template>
   <WaitSpinner  v-if="unitStore.working" :overlay="true" message="Working..." />
   <div class="column items-center" v-if="projectStore.hasDetail">
      <h2>
         <router-link :to="`/projects/${projectStore.detail.id}`">{{truncateTitle(title)}}</router-link>
      </h2>
      <div class="column items-center gap-1">
         <div class="row-left gap-2">
            <label class="font-semibold!">Unit:</label><a target="_blank" :href="`${systemStore.adminURL}/units/${unitStore.unitID}`">{{unitStore.unitID}}</a>
         </div>
         <div class="row-left gap-2">
            <label class="font-semibold!">Directory:</label>{{workingDir}}
         </div>
      </div>
      <div class="row-between px-3 w-full pb-4">
         <UButton icon="i-lucide-arrow-left" label="Back to project" @click="backClicked" size="sm" color="secondary"/>
         <ProblemsDisplay    v-if="unitStore.problems.length > 0"/>
         <KeyboardShortcutHelp />
      </div>
   </div>
   <div class="master-files" ref="masterfiles">
      <div v-if="unitStore.masterFiles.length == 0" class="text-center text-xl border-t border-brand-grey-100 pt-10">
         No images found
      </div> 
      <template v-else>
         <MasterFilesList  v-if="unitStore.viewMode == 'list'" />
         <MasterFilesGrid  v-else />
      </template>
   </div>
</template>

<script setup>
import ProblemsDisplay from '@/components/unit/ProblemsDisplay.vue'
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUnitStore} from "@/stores/unit"
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import MasterFilesList from '@/components/unit/MasterFilesList.vue'
import MasterFilesGrid from '@/components/unit/MasterFilesGrid.vue'
import KeyboardShortcutHelp from '@/components/unit/KeyboardShortcutHelp.vue'
import { onKeyStroke } from '@vueuse/core'
import { useConfirm } from "@/composables/useConfirm"

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
   let files = []
   for (let i=unitStore.rangeStartIdx; i<=unitStore.rangeEndIdx; i++) {
      let mf = unitStore.masterFiles[i]
      files.push( mf.fileName )
   }
   let msg = `<div style="display:flex; flex-direction: column; gap: 5px; align-items: flex-start;">
               <div>Delete these images?<div>
               <div style="margin: 0 0 15px 10px; font-size: 0.9em">${files.join(", ")}</div>
               <div>All data will be list. This is not reversable.</div>
               <div>Are you sure?</div>
            </div>`
   const resp = await useConfirm("Confirm Image Delete", msg, "Delete")
   if (resp) {
      unitStore.deleteMasterFiles( files )
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
</style>
