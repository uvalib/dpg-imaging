<template>
   <div class="viewer">
      <WaitSpinner v-if="unitStore.working" :overlay="true" message="Working..." />
      <div id="iiif-header" class="toolbar" v-show="fullScreen == false">
         <div class="header-top">
            <table class="info" v-if="projectStore.hasDetail">
               <tbody>
                  <tr class="line">
                     <td class="label">Image:</td>
                     <td class="data" v-if="currMasterFile">
                        <span>{{currMasterFile.fileName}}</span>
                        <span class="detail">(Size: {{currMasterFile.width}} x {{currMasterFile.height}}, Resolution: {{currMasterFile.resolution}})</span>
                     </td>
                  </tr>
                  <tr class="line">
                     <td class="label">Title:</td>
                     <td class="data" @click="editMetadata('title')" >
                        <template v-if="isEditing('title')">
                           <TitlePicker v-model="newValue" @cancel="cancelEdit" @submit="submitEdit"/>
                        </template>
                        <template v-else>
                           <span v-if="currMasterFile && currMasterFile.title" class="editable">{{currMasterFile.title}}</span>
                           <span v-else class="undefined editable">Undefined</span>
                        </template>
                     </td>
                  </tr>
                  <tr class="line">
                     <td class="label">Caption:</td>
                     <td class="data editable" @click="editMetadata('description')" >
                        <input  v-if="isEditing('description')" id="edit-desc" type="text" v-model="newValue"
                           @keyup.enter="submitEdit()" @keydown.stop.prevent.esc="cancelEdit" @keydown.tab="cancelEdit"/>
                        <template v-else>
                           <span v-if="currMasterFile && currMasterFile.description">{{currMasterFile.description}}</span>
                           <span v-else class="undefined editable">Undefined</span>
                        </template>
                     </td>
                  </tr>
                  <tr class="line">
                     <td class="label">Keybard Shortcuts:</td>
                     <td class="data">Pan Image: w,a,s,d or arrow keys. Pagination: &lt; prior, &gt; next. Full Screen Toggle: z. 100% Zoom: 1.</td>
                  </tr>
               </tbody>
            </table>
            <TagPicker v-if="currMasterFile" :masterFile="currMasterFile" display="large" />
         </div>
         <div class="acts">
            <span class="back">
               <UButton icon="i-lucide-arrow-left" label="Back to unit" @click="backClicked" size="sm" color="secondary"/>
            </span>
            <span class="paging group">
               <UButton icon="i-lucide-arrow-left" @click="prevImage" size="sm" color="secondary" :disabled="prevDisabled"/>
               <span class="page">{{page}} of {{unitStore.pageInfoURLs.length}}</span>
               <UButton icon="i-lucide-arrow-right" @click="nextImage" size="sm" color="secondary" :disabled="nextDisabled"/>
            </span>
            <span class="zoom group">
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
   </div>
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
         showFullPageControl: false
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

<style scoped lang="scss">
.viewer {
   :deep(.openseadragon-container) {
      background: #555555 !important;
   }
   .toolbar {
      background: #fafafa;
      position: relative;
      display: flex;
      flex-direction: column;
      .header-top {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: flex-start;
         padding: 10px 10px 0 10px;
      }
      .undefined {
         font-style: italic;
         color: var(--uvalib-grey);
      }
      .editable {
         display: inline-block;
         cursor: pointer;
         &:hover {
            text-decoration: underline;
            color: var(--uvalib-blue-alt) !important;
         }
      }
      .acts {
         padding: 5px 10px;
         border-top: 1px solid var(--uvalib-grey-light);
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         background: #f0f0f0;
         .group {
            display: flex;
            flex-flow: row nowrap;
            align-items: center;
            gap: 10px;
         }
         .back {
            width: 25%;
            display: flex;
            flex-flow: row nowrap;
            align-items: center;
            justify-content: flex-start;
         }
         .zoom {
            gap: 5px;
         }
      }

      .info {
         text-align: left;
         padding-bottom: 10px;
         margin-bottom: 0;

         .line {
            padding: 3px 0;

            td {
               padding: 5px;
            }
            td.label {
               font-weight: bold;
               text-align: right;
               width: max-content;
               white-space: nowrap;
            }
            td.data {
               text-align: left;
               width:100%;
               .detail {
                  font-weight: 100;
                  font-size: 0.8em;
                  margin-left: 15px;
               }
            }
         }
      }
   }
}
#iiif-viewer {
   position: absolute;
   width: 100%;
   bottom: 0;
   top: 120px;
   background: black;
}
</style>
