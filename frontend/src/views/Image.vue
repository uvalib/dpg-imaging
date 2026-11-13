<template>
   <WaitSpinner v-if="unitStore.working" :overlay="true" message="Working..." />
   <div id="iiif-header" v-show="fullScreen == false">
      <div class="row-between gap-4 p-2">
         <table v-if="projectStore.hasDetail" class="table-fixed grow">
            <tbody>
               <tr>
                  <td class="label">Image:</td>
                  <td class="row-left gap-2 items-center" v-if="currMasterFile">
                     <span>{{currMasterFile.fileName}}</span>
                     <span class="text-sm">(Size: {{currMasterFile.width}} x {{currMasterFile.height}}, Resolution: {{currMasterFile.resolution}})</span>
                  </td>
               </tr>
               <tr>
                  <td class="label">Title:</td>
                  <td>
                     <TitlePicker v-if="isEditing('title')" v-model="newValue" @cancel="cancelEdit" @submit="submitEdit"/>
                     <ULink v-else @click="editMetadata('title')" class="text-left whitespace-break-spaces p-0">
                        <template v-if="currMasterFile?.title">{{  currMasterFile.title }}</template>
                        <template v-else class="italic">Undefined</template>
                     </ULink>
                  </td>
               </tr>
               <tr>
                  <td class="label">Caption:</td>
                  <td>
                     <UInput v-if="isEditing('description')" v-model="newValue" @keydown.enter="submitEdit" 
                        @keydown.esc="cancelEdit" @keydown.tab="submitEdit" class="w-full" autofocus/>
                     <ULink v-else @click="editMetadata('description')" class="text-left whitespace-break-spaces p-0">
                        <template v-if="currMasterFile?.description">{{  currMasterFile.description }}</template>
                        <template v-else class="italic">Undefined</template>
                     </ULink>
                  </td>
               </tr>
               <tr>
                  <td class="label">Keybard Shortcuts:</td>
                  <td class="data">Pan Image: w,a,s,d or arrow keys. Pagination: &lt; prior, &gt; next. Full Screen Toggle: z. 100% Zoom: 1.</td>
               </tr>
            </tbody>
         </table>
         <TagPicker v-if="currMasterFile" :masterFile="currMasterFile" display="large" />
      </div>
      <div class="row-between border-t border-brand-grey-100 p-2 items-center bg-gray-50">
         <UButton icon="i-lucide-arrow-left" label="Back to unit" @click="backClicked" size="sm" color="secondary"/>
         <span class="row-left gap-2 items-center">
            <UButton icon="i-lucide-arrow-left" @click="prevImage" size="sm" color="secondary" :disabled="prevDisabled"/>
            <span class="page">{{page}} of {{unitStore.pageInfoURLs.length}}</span>
            <UButton icon="i-lucide-arrow-right" @click="nextImage" size="sm" color="secondary" :disabled="nextDisabled"/>
         </span>
         <span class="row-left gap-2 items-center">
            <UButton id="rotate-left" icon="i-lucide-rotate-ccw" @click="rotateImage('left')" color="secondary"/>
            <UButton id="rotate-left" icon="i-lucide-rotate-cw" @click="rotateImage('right')" color="secondary"/>
            <UButton id="zoom-in" icon="i-lucide-zoom-in" color="secondary"/>
            <span class="page">{{Math.round(zoom*100)}} %</span>
            <UButton id="zoom-out" icon="i-lucide-zoom-out" color="secondary"/>
            <UButton id="actual-size" label="1:1" color="secondary" @click="viewActualSize"/>
            <UButton id="home" icon="i-lucide-home" color="secondary" />
         </span>
      </div>
   </div>
   <div id="iiif-viewer"></div>
</template>

<script setup>
import OpenSeadragon from "openseadragon"
import TagPicker from '../components/TagPicker.vue'
import {useProjectStore} from "@/stores/project"
import {useUnitStore} from "@/stores/unit"
import { computed, ref, onBeforeMount, onUnmounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import TitlePicker from "../components/TitlePicker.vue"
import { onKeyStroke } from '@vueuse/core'

const projectStore = useProjectStore()
const unitStore = useUnitStore()
const route = useRoute()
const router = useRouter()

// local data
var viewer = null
const page = ref(1)
const zoom = ref(50)
const editField = ref("")
const newValue = ref("")
const fullScreen = ref(false)
const viewerTop = ref()
const viewerH = ref()

const currMasterFile = computed(() => {
   return unitStore.masterFiles[page.value-1]
})

const prevDisabled = computed(() => {
   return page.value == 1
})

const nextDisabled = computed(() => {
   return page.value == unitStore.totalFiles
})

const backClicked = (() => {
   if (unitStore.lastURL != "") {
      router.push( unitStore.lastURL )
   } else {
      router.push( `/projects/${projectStore.detail.id}/unit` )
   }
})

const rotateImage = ( async (dir) => {
   await unitStore.rotateImage({file: currMasterFile.value.fileName, dir: dir})
   viewer.world.resetItems()
   viewer.goToPage( viewer.currentPage() )
})

const viewActualSize = (() => {
   viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(1.0))
})

const isEditing = ((field = "all") => {
   return editField.value == field
})

const editMetadata = ((field) => {
   let mf = currMasterFile.value
   if (field == "title") {
      newValue.value = mf.title
   }
   if (field == "description") {
      newValue.value = mf.description
   }
   editField.value = field
})

const cancelEdit= (() => {
   editField.value = ""
})

const submitEdit = ( async () => {
   let mf = currMasterFile.value
   await unitStore.updateMasterFileMetadata( mf.fileName, editField.value, newValue.value )
   editField.value = ""
})

const nextImage = ( async () => {
   page.value++
   let pgIdx = page.value-1
   await unitStore.getMasterFileMetadata( pgIdx )
   viewer.goToPage( pgIdx )
   let url = `/projects/${projectStore.detail.id}/unit/images/${page.value}`
   router.replace(url)
   focusViewer()
})

const prevImage = ( async () => {
   if (page.value > 1) {
      page.value--
      let pgIdx = page.value-1
      await unitStore.getMasterFileMetadata( pgIdx )
      viewer.goToPage( pgIdx )
      let url = `/projects/${projectStore.detail.id}/unit/images/${page.value}`
      router.replace(url)
      focusViewer()
   }
})

const focusViewer = (() => {
   nextTick(()=>{
      let viewEle = document.getElementById("iiif-viewer")
      if (viewEle) {
         viewEle.querySelector('.openseadragon-canvas').focus()
      }
   })
})

onKeyStroke(['>','.'], () => {
   if (nextDisabled.value == false && editField.value == "") {
      nextImage()
      viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(zoom.value))
      setTimeout( () => {
         viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(zoom.value))
      }, 255) // 255 is just longer that the .25 sec animation timer
   }
})

onKeyStroke(['<',','], () => {
   if (prevDisabled.value == false && editField.value == "") {
      prevImage()
      viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(zoom.value))
      setTimeout( () => {
         viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(zoom.value))
      }, 255) // 255 is just longer that the .25 sec animation timer
   }
})

onKeyStroke('1', () => {
   if ( editField.value == "" ) {
      viewActualSize()
   }
})

onKeyStroke('z', () => {
   if ( editField.value == "" ) {
      let ele = document.getElementById("iiif-viewer")
      fullScreen.value = !fullScreen.value
      let origZoom = zoom.value
      viewer.setFullPage( fullScreen.value )
      if (fullScreen.value ) {
         ele.style.top = `0px`
         ele.style.height = `100%`
      }
      focusViewer()
      setTimeout( () => {
         if ( fullScreen.value == false ) {
            ele.style.top = `${viewerTop.value}px`
            ele.style.height = `${viewerH.value}px`
         }
         if (zoom.value != origZoom)  {
            zoom.value = origZoom
            viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(origZoom))
         }
      }, 50)
   }
})


onBeforeMount( async () => {
   page.value = parseInt(route.params.page, 10)
   let currPageIndex = page.value-1

   if (projectStore.hasDetail == false) {
      await projectStore.getProject(route.params.id)
      await unitStore.getMasterFiles( projectStore.detail )
      await unitStore.getMasterFileMetadata( currPageIndex )
   }
   nextTick(()=>{
      let hdr = document.getElementById("uva-header")
      let toolbar = document.getElementById("iiif-header")
      viewerTop.value = hdr.offsetHeight + toolbar.offsetHeight
      let ele =  document.getElementById("iiif-viewer")
      ele.style.top = `${viewerTop.value}px`
      viewer = OpenSeadragon({
         id: "iiif-viewer",
         animationTime: 0.25,
         showNavigator: true,
         sequenceMode: true,
         preserveViewport: true,
         autoResize: true,
         visibilityRatio: 0.95,
         constrainDuringPan: true,
         imageSmoothingEnabled: false,
         smoothTileEdgesMinZoom: Infinity,
         maxZoomPixelRatio: 2.0,
         placeholderFillStyle: '#555555',
         navigatorPosition: "BOTTOM_RIGHT",
         zoomInButton:   "zoom-in",
         zoomOutButton:  "zoom-out",
         homeButton:     "home",
         showSequenceControl: false,
         tileSources: unitStore.pageInfoURLs,
         initialPage: currPageIndex,
         showFullPageControl: false,
         navigatorDrawer: "canvas",
         crossOriginPolicy: "Anonymous",
      })
      viewer.gestureSettingsMouse.clickToZoom = false
      viewer.addHandler("zoom", (data) => {
         zoom.value = viewer.viewport.viewportToImageZoom(data.zoom)
         focusViewer()
      })
      focusViewer()
      nextTick( ()=> {
         viewerH.value = ele.offsetHeight
      })
   })
})
onUnmounted( async () => {
   if (viewer) {
      viewer.destroy()
   }
})
</script>

<style scoped lang="css">
@reference "@/assets/main.css";

.label {
   @apply font-bold! text-right pr-2 max-w-22;
}
:deep(.openseadragon-container) {
   background: #555555 !important;
}

#iiif-viewer {
   position: absolute;
   width: 100%;
   bottom: 0;
   top: 120px;
   background: black;
}
</style>
