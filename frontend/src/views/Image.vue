<template>
   <div class="viewer">
      <WaitSpinner v-if="systemStore.working" :overlay="true" message="Working..." />
      <div id="iiif-toolbar" class="toolbar">
         <TagPicker v-if="currMasterFile" :masterFile="currMasterFile" display="large" position="topright" class="top-right"/>
         <table class="info" v-if="projectStore.selectedProjectIdx > -1">
            <tr class="line">
               <td class="label">Image:</td>
               <td class="data" v-if="currMasterFile">
                  <span>{{currMasterFile.fileName}}</span>
                  <span class="detail">(Size: {{currMasterFile.width}} x {{currMasterFile.height}}, Resolution: {{currMasterFile.resolution}})</span>
               </td>
            </tr>
            <tr class="line">
               <td class="label">Title:</td>
               <td class="data editable" @click="editMetadata('title')" >
                  <TitleInput  v-if="isEditing('title')" @canceled="cancelEdit" @accepted="submitEdit" v-model="newTitle"/>
                  <template v-else>
                     <span  v-if="currMasterFile && currMasterFile.title">{{currMasterFile.title}}</span>
                     <span v-else class="undefined editable">Undefined</span>
                  </template>
               </td>
            </tr>
            <tr class="line">
               <td class="label">Caption:</td>
               <td class="data editable" @click="editMetadata('description')" >
                  <input  v-if="isEditing('description')" id="edit-desc" type="text" v-model="newDescription"
                     @keyup.enter="submitEdit()" @keyup.esc="cancelEdit" />
                  <template v-else>
                     <span v-if="currMasterFile && currMasterFile.description">{{currMasterFile.description}}</span>
                     <span v-else class="undefined editable">Undefined</span>
                  </template>
               </td>
            </tr>
         </table>
         <span class="toolbar-button group back">
            <i class="fas fa-angle-double-left back-button"></i>
            <span @click="router.back()">Back to unit</span>
         </span>
         <span class="paging group">
            <span id="previous" title="Previous" class="toolbar-button" :class="{disabled: prevDisabled()}"  @click="prevImage"><i class="fas fa-arrow-left"></i></span>
            <span class="page">{{page}} of {{unitStore.pageInfoURLs.length}}</span>
            <span id="next" title="Next" class="toolbar-button" :class="{disabled: nextDisabled()}" @click="nextImage"><i class="fas fa-arrow-right"></i></span>
         </span>
         <span class="zoom group">
            <span id="rotate-left" title="Rotate Left" class="toolbar-button"  @click="rotateImage('left')"><i class="fas fa-undo"></i></span>
            <span id="rotate-right" title="Rotate Right" class="toolbar-button"  @click="rotateImage('right')"><i class="rotated fas fa-undo"></i></span>
            <span id="zoom-in" title="Zoom in" class="toolbar-button"><i class="fas fa-search-plus"></i></span>
            <span class="page">{{Math.round(zoom*100)}} %</span>
            <span id="zoom-out" title="Zoom in" class="toolbar-button"><i class="fas fa-search-minus"></i></span>
            <span id="actual-size" title="Reset view" @click="viewActualSize" class="full toolbar-button">1:1</span>
            <span id="home" title="Reset view" class="toolbar-button"><i class="fas fa-home"></i></span>
            <span id="full-page" title="Full Screen" class="toolbar-button"><i class="fas fa-expand"></i></span>
         </span>
      </div>
      <div id="iiif-viewer"></div>
   </div>
</template>

<script setup>
import OpenSeadragon from "openseadragon"
import TagPicker from '../components/TagPicker.vue'
import TitleInput from '../components/TitleInput.vue'
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUnitStore} from "@/stores/unit"
import { computed, ref, onBeforeMount, onUnmounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const unitStore = useUnitStore()
const route = useRoute()
const router = useRouter()

// local data
var viewer = null
const page = ref(1)
const zoom = ref(50)
const newTitle = ref("")
const newDescription = ref("")
const editField = ref("")

const currMasterFile = computed(() => {
   return unitStore.masterFiles[page.value-1]
})

async function rotateImage(dir) {
   await unitStore.rotateImage({file: currMasterFile.value.fileName, dir: dir})
   viewer.world.resetItems()
   viewer.goToPage( viewer.currentPage() )
}
function viewActualSize() {
   viewer.viewport.zoomTo(viewer.viewport.imageToViewportZoom(1.0))
}
function isEditing(field = "all") {
   return editField.value == field
}
function editMetadata(field) {
   let mf = currMasterFile.value
   newTitle.value = mf.title
   newDescription.value = mf.description
   editField.value = field
   nextTick( ()=> {
      if ( field == "description") {
         let ele = document.getElementById("edit-desc")
         ele.focus()
         ele.select()
      }
   })
}
function cancelEdit() {
   editField.value = ""
}
async function submitEdit() {
   let mf = currMasterFile.value
   await unitStore.updateMasterFileMetadata(
      { file: mf.path, title: newTitle.value, description: newDescription.value,
        status: mf.status, componentID: mf.componentID
      }
   )
   editField.value = ""
}

async function nextImage() {
   page.value++
   let pgIdx = page.value-1
   await unitStore.getMasterFileMetadata( pgIdx )
   viewer.goToPage( pgIdx )
   let url = `/projects/${projectStore.currProject.id}/unit/images/${page.value}`
   router.replace(url)
   focusViewer()
}

function prevDisabled() {
   return page.value == 1
}
function nextDisabled() {
   return page.value == unitStore.totalFiles
}
async function prevImage() {
   if (page.value > 1) {
      page.value--
      let pgIdx = page.value-1
      await unitStore.getMasterFileMetadata( pgIdx )
      viewer.goToPage( pgIdx )
      let url = `/projects/${projectStore.currProject.id}/unit/images/${page.value}`
      router.replace(url)
      focusViewer()
   }
}

function focusViewer() {
   nextTick(()=>{
      let viewEle = document.getElementById("iiif-viewer")
      if (viewEle) {
         viewEle.querySelector('.openseadragon-canvas').focus()
      }
   })
}

onBeforeMount( async () => {
   page.value = parseInt(route.params.page, 10)
   let currPageIndex = page.value-1

   if (projectStore.selectedProjectIdx == -1) {
      await projectStore.getProject(route.params.id)
      await unitStore.getUnitMasterFiles(projectStore.currProject.unit.id)
      await unitStore.getMasterFileMetadata( currPageIndex )
   }
   nextTick(()=>{
      let hdr = document.getElementById("uva-header")
      let toolbar = document.getElementById("iiif-toolbar")
      let h = hdr.offsetHeight + toolbar.offsetHeight
      document.getElementById("iiif-viewer").style.top = `${h}px`
      viewer = OpenSeadragon({
         id: "iiif-viewer",
         toolbar: "iiif-toolbar",
         showNavigator: true,
         sequenceMode: true,
         preserveViewport: true,
         autoResize: true,
         visibilityRatio: 0.95,
         constrainDuringPan: true,
         imageSmoothingEnabled: false,
         maxZoomPixelRatio: 2.0,
         placeholderFillStyle: '#555555',
         navigatorPosition: "BOTTOM_RIGHT",
         zoomInButton:   "zoom-in",
         zoomOutButton:  "zoom-out",
         homeButton:     "home",
         fullPageButton: "full-page",
         showSequenceControl: false,
         tileSources: unitStore.pageInfoURLs,
         initialPage: currPageIndex
      })
      viewer.gestureSettingsMouse.clickToZoom = false
      viewer.addHandler("zoom", (data) => {
         zoom.value = viewer.viewport.viewportToImageZoom(data.zoom)
         focusViewer()
      })
      focusViewer()
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
   .load {
      margin-top: 5%;
   }
   .toolbar {
      padding: 10px;
      background: var(--uvalib-grey-light);
      position: relative;
      border-bottom: 1px solid var(--uvalib-grey);
      .top-right {
         position: absolute;
         top: 5px;
         right: 5px;
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
      input {
         box-sizing: border-box;
         width:100%;
         border-radius: 3px;
         padding: 3px 5px;
         border: 1px solid var(--uvalib-grey-light);
         outline: none;
         background: #f0f0f0;
      }

      .info {
         width: 100%;
         text-align: left;
         padding-bottom: 10px;
         margin-bottom: 10px;
         border-bottom: 1px solid var(--uvalib-grey);
         .line {
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
            padding: 3px 0;
         }
      }

      .group {
         display: inline-block;
         position: relative;
      }
      .rotated {
         transform: scaleX(-1);
      }

      .toolbar-button {
         padding: 5px 10px;
         display: inline-block;
         position: relative;
         cursor: pointer;
         &:hover {
            text-decoration: underline;
         }
      }
      .toolbar-button.disabled {
         opacity:0.2;
      }
      .back-button {
         padding: 5px 10px 5px 0;
         display: inline-block;
         position: relative;
      }

      .back {
         position: absolute;
         left: 10px;
         bottom: 7px;
         a {
            text-decoration: none;
            color: var(--uvalib-text);
         }
         &:hover {
            text-decoration: underline ;
         }
      }
      .zoom {
         position: absolute;
         right: 10px;
         bottom: 12px;
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
.not-found {
   display: inline-block;
   padding: 20px 50px;
   margin: 4% auto 0 auto;
   h2 {
      font-size: 1.5em;
      color: var(--uvalib-text);
   }
}
</style>
