<template>
   <div class="viewer">
      <WaitSpinner v-if="updating" :overlay="true" message="Updating data..." />
       <div class="load" v-if="loadingUnit || loadingProject || masterFiles == false">
         <WaitSpinner message="Loading viewer..." />
      </div>
      <template v-else>
         <div id="iiif-toolbar" class="toolbar">
            <TagPicker :masterFile="currMasterFile" display="large" position="topright" class="top-right"/>
            <table class="info">
               <tr class="line">
                  <td class="label">Image:</td>
                  <td class="data">
                     <span>{{currMasterFile.fileName}}</span>
                     <span class="detail">(Size: {{currMasterFile.width}} x {{currMasterFile.height}}, Resolution: {{currMasterFile.resolution}})</span>
                  </td>
               </tr>
               <tr class="line">
                  <td class="label">Title:</td>
                  <td class="data editable" @click="editMetadata('title')" >
                     <TitleInput  v-if="isEditing('title')" @canceled="cancelEdit" @accepted="submitEdit" v-model="newTitle"/>
                     <template v-else>
                        <span  v-if="currMasterFile.title">{{currMasterFile.title}}</span>
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
                        <span v-if="currMasterFile.description">{{currMasterFile.description}}</span>
                        <span v-else class="undefined editable">Undefined</span>
                     </template>
                  </td>
               </tr>
            </table>
            <span class="toolbar-button group back">
               <i class="fas fa-angle-double-left back-button"></i>
               <span @click="$router.back()">Back to unit</span>
            </span>
            <span class="paging group">
               <span id="previous" title="Previous" class="toolbar-button"><i class="fas fa-arrow-left"></i></span>
               <span class="page">{{page}} of {{pageInfoURLs.length}}</span>
               <span id="next" title="Next" class="toolbar-button"><i class="fas fa-arrow-right"></i></span>
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
      </template>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
import OpenSeadragon from "openseadragon"
import TagPicker from '../components/TagPicker.vue'
import TitleInput from '../components/TitleInput.vue'
export default {
   name: "Image",
   components: {
      TagPicker, TitleInput
   },
   computed: {
      ...mapState({
         loadingProject: state => state.projects.working,
         loadingUnit : state => state.loading,
         updating : state => state.updating,
         currUnit: state => state.units.currUnit,
         selectedProjectIdx: state  => state.projects.selectedProjectIdx,
         masterFiles:  state => state.units.masterFiles
      }),
      ...mapGetters({
         pageInfoURLs: 'units/pageInfoURLs',
         masterFileInfo: 'units/masterFileInfo',
         currProject: 'projects/currProject',
      }),
      hasMasterFiles() {
         return this.masterFiles.length > 0
      },
      currMasterFile() {
         return this.masterFileInfo( this.page-1)
      }
   },
   data() {
      return {
        viewer: null,
        page: 1,
        zoom: 50,
        newTitle: "",
        newDescription: "",
        editField: ""
      }
   },
   methods: {
      rotateImage(dir) {
         this.$store.dispatch("units/rotateImage", {file: this.currMasterFile.fileName, dir: dir})
      },
      viewActualSize() {
         this.viewer.viewport.zoomTo(this.viewer.viewport.imageToViewportZoom(1.0))
      },
      isEditing(field = "all") {
         return this.editField == field
      },
      editMetadata(field) {
         let mf = this.currMasterFile
         this.newTitle = mf.title
         this.newDescription = mf.description
         this.editField = field
         this.$nextTick( ()=> {
            if ( field == "description") {
               let ele = document.getElementById("edit-desc")
               ele.focus()
               ele.select()
            }
         })
      },
      cancelEdit() {
         this.editField = ""
      },
      async submitEdit() {
         let mf = this.currMasterFile
         await this.$store.dispatch("units/updateMasterFileMetadata",
            { file: mf.path, title: this.newTitle, description: this.newDescription,
              status: mf.status, componentID: mf.componentID } )
         this.editField = ""

      }
   },
   async beforeMount() {
      if (this.selectedProjectIdx == -1) {
         await this.$store.dispatch("projects/getProject", this.$route.params.id)
         await this.$store.dispatch("units/getUnitMasterFiles", this.currProject.unit.id)
      }
      this.$nextTick(()=>{
         let hdr = document.getElementById("uva-header")
         let toolbar = document.getElementById("iiif-toolbar")
         let h = hdr.offsetHeight + toolbar.offsetHeight
         document.getElementById("iiif-viewer").style.top = `${h}px`
         this.page = parseInt(this.$route.params.page, 10)
         let pageIdx = this.page-1
         this.viewer = OpenSeadragon({
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
            nextButton:     "next",
            previousButton: "previous",
            tileSources: this.pageInfoURLs,
            initialPage: pageIdx
         })
         this.viewer.addHandler("page", (data) => {
            this.page = data.page + 1
            let url = `/projects/${this.currProject.id}/unit/images/${this.page}`
            this.$router.replace(url)
         })
         this.viewer.addHandler("zoom", (data) => {
            this.zoom = this.viewer.viewport.viewportToImageZoom(data.zoom)
         })
         this.viewer.forceRedraw()
      })
   },
   unmounted() {
      if (this.viewer) {
         this.viewer.destroy()
      }
   }
}
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
