<template>
   <div class="viewer">
      <WaitSpinner v-if="loading" :overlay="true" message="Loading viewer..." />
      <template v-else>
         <div id="iiif-toolbar" class="toolbar">
            <div class="info">
               <div class="line">
                  <label>File Name:</label>
                  <span class="data">{{currMasterFile.fileName}}</span>
               </div>
               <div class="line">
                  <label>Title:</label>
                  <span class="data">{{currMasterFile.title}}</span>
               </div>
               <div class="line">
                  <label>Caption:</label>
                  <span class="data">{{currMasterFile.description}}</span>
               </div>
            </div>
            <span class="toolbar-button group back">
               <i class="fas fa-angle-double-left back-button"></i>
               <router-link :to="`/unit/${currUnit}`">Back to Unit</router-link>
            </span>
            <span class="paging group">
               <span id="previous" title="Previous" class="toolbar-button"><i class="fas fa-arrow-left"></i></span>
               <span class="page">{{page}} of {{pageInfoURLs.length}}</span>
               <span id="next" title="Next" class="toolbar-button"><i class="fas fa-arrow-right"></i></span>
            </span>
            <span class="zoom group">
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
export default {
   name: "Page",
   computed: {
      ...mapState({
         loading : state => state.loading,
         currUnit: state => state.currUnit
      }),
      ...mapGetters([
        'pageInfoURLs',
        'masterFileInfo'
      ]),
      currMasterFile() {
         return this.masterFileInfo( this.page-1)
      }
   },
   data() {
      return {
        viewer: null,
        page: 1,
        zoom: 50
      }
   },
   methods: {
      viewActualSize() {
         this.viewer.viewport.zoomTo(this.viewer.viewport.imageToViewportZoom(1.0))
      }
   },
   async created() {
      await this.$store.dispatch("getMasterFiles", this.$route.params.unit)
      this.$nextTick( ()=>{
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
            this.$router.replace(`/unit/${this.currUnit}/page/${this.page}`)
         })
         this.viewer.addHandler("zoom", (data) => {
            this.zoom = this.viewer.viewport.viewportToImageZoom(data.zoom)
         })
      })
   }
}
</script>

<style lang="scss">
.viewer {
   background: black;
   .toolbar {
      padding: 10px;
      background: var(--uvalib-grey-light);
      position: relative;
      border-bottom: 1px solid var(--uvalib-grey);

      .info {
         text-align: left;
         padding-bottom: 10px;
         margin-bottom: 10px;
         border-bottom: 1px solid var(--uvalib-grey);
         .line {
            label {
               font-weight: bold;
               width: 100px;
               text-align: right;
               display: inline-block;
            }
            .data {
               margin-left: 10px;
            }
            padding: 3px 0;
         }
      }

      .group {
         display: inline-block;
         position: relative;
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
.extra-tools {
   z-index: 1000;
   position: absolute;
   left: 12px;
   top: 12px;
   font-size: 1.1em;
   color: #222;
   cursor: pointer;
   .dl-text {
      margin-left: 5px;
      font-weight: 500;
      &:hover {
         text-decoration: underline;
      }
   }
}
</style>
