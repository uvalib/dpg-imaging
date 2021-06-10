<template>
   <div class="unit">
      <WaitSpinner v-if="updating" :overlay="true" message="Updating data..." />
      <div class="load" v-if="loading">
         <WaitSpinner v-if="loading" message="Loading master file..." />
      </div>
      <template v-else>
         <h2>Unit {{currUnit}}</h2>
         <div class="toolbar">
            <span class="viewe-mode">
               <label>View:</label>
               <select id="layout" v-model="viewMode">
                  <option value="list">List</option>
                  <option value="medium">Gallery (medium)</option>
                  <option value="large">Gallery (large)</option>
               </select>
            </span>
            <span class="actions">
               <DPGButton id="sort" class="right-pad" @click="resetSort">File Name Sort</DPGButton>
               <ConfirmModal label="Batch Rename" class="right-pad" @confirmed="renameAll">
                  <div>All files will be renamed to match the following format:</div>
                  <code>{{currUnit}}_0001.tif</code>
               </ConfirmModal>
               <DPGButton id="set-titles" @click="setPageNumbersClicked" class="button">Set Page Numbers</DPGButton>
            </span>
         </div>
         <PageNumPanel v-if="editMode == 'page'" />
         <table class="unit-list" v-if="viewMode == 'list'">
            <thead>
               <tr>
                  <th></th>
                  <th>File Name</th>
                  <th>File Type</th>
                  <th>Resolution</th>
                  <th>Title</th>
                  <th>Caption</th>
                  <th>Color Profile</th>
                  <th>Path</th>
               </tr>
            </thead>
            <draggable v-model="masterFiles" tag="tbody"  @start="dragStarted" >
               <tr v-for="mf in masterFiles" :key="mf.fileName" :id="mf.fileName"
                  @mousedown="fileSelected(mf.fileName, $event)"
                  @contextmenu.prevent="showContextMenu($event)"
               >
                  <td class="thumb">
                     <router-link :to="`/unit/${currUnit}/page/${mf.fileName.replace('.tif','').split('_')[1]}`"><img :src="mf.thumbURL"/></router-link>
                  </td>
                  <td>{{mf.fileName}}</td>
                  <td>{{mf.fileType}}</td>
                  <td>{{mf.resolution}}</td>
                  <td @click="editMetadata(mf, 'title')" class="editable">
                     <span  v-if="!isEditing(mf, 'title')"  class="editable">
                        <span v-if="mf.title">{{mf.title}}</span>
                        <span v-else class="undefined">Undefined</span>
                     </span>
                     <input v-else id="edit-title" type="text" v-model="newTitle"
                        @keyup.enter="submitEdit(mf)" @keyup.esc="cancelEdit" />
                  </td>
                  <td @click="editMetadata(mf, 'description')" class="editable" >
                     <span  v-if="!isEditing(mf, 'description')" class="editable">
                        <span v-if="mf.description">{{mf.description}}</span>
                        <span v-else class="undefined">Undefined</span>
                     </span>
                     <input v-else id="edit-desc" type="text" v-model="newDescription"
                        @keyup.enter="submitEdit(mf)"  @keyup.esc="cancelEdit" />
                  </td>
                  <td>{{mf.colorProfile}}</td>
                  <td>{{mf.path}}</td>
               </tr>
            </draggable>
         </table>
         <div class="gallery" :class="viewMode" v-else>
            <draggable v-model="masterFiles" @start="dragStarted" >
               <div class="card" v-for="mf in masterFiles" :key="mf.fileName"  @mousedown="fileSelected(mf.fileName, $event)" :id="mf.fileName">
                  <router-link :to="`/unit/${currUnit}/page/${mf.fileName.replace('.tif','').split('_')[1]}`">
                     <img :src="mf.mediumURL" v-if="viewMode == 'medium'"/>
                     <img :src="mf.largeURL" v-if="viewMode == 'large'"/>
                  </router-link>
                  <div class="metadata">
                     <div class="row">
                        <label>File Name</label>
                        <div class="data">{{mf.fileName}}</div>
                     </div>
                     <div class="row">
                        <label>Title</label>
                        <div class="data editable" @click="editMetadata(mf, 'title')">
                           <template v-if="isEditing(mf, 'title')">
                              <input id="edit-title" type="text" v-model="newTitle"  @keyup.enter="submitEdit(mf)" @keyup.esc="cancelEdit" />
                           </template>
                           <template v-else>
                              <template v-if="mf.title">{{mf.title}}</template>
                              <span v-else class="undefined">Undefined</span>
                           </template>
                        </div>
                     </div>
                     <div class="row">
                        <label>Caption</label>
                        <div class="data editable" @click="editMetadata(mf, 'description')">
                           <template v-if="isEditing(mf, 'description')">
                              <input id="edit-desc" type="text" v-model="newDescription" @keyup.enter="submitEdit(mf)" @keyup.esc="cancelEdit" />
                           </template>
                           <template v-else>
                              <template v-if="mf.description">{{mf.description}}</template>
                              <span v-else class="undefined">Undefined</span>
                           </template>
                        </div>
                     </div>
                  </div>
               </div>
            </draggable>
         </div>
      </template>
      <div class="popupmenu" id="popupmenu" v-show="menuVisible">
         <ul>
            <li @click="setPageNumbersClicked">Set Page Numbers</li>
            <li>Rename Image</li>
            <li>Delete Image</li>
         </ul>
      </div>
   </div>
</template>

<script>
import { mapState } from "vuex"
import { mapFields } from 'vuex-map-fields'
import PageNumPanel from '../components/PageNumPanel.vue'
import draggable from 'vuedraggable'
export default {
   components: {PageNumPanel, draggable },
   name: "unit",
   computed: {
      ...mapState({
         loading : state => state.loading,
         updating : state => state.updating,
         currUnit: state => state.currUnit
      }),
      ...mapFields([
         'viewMode', "rangeStartIdx", "rangeEndIdx", "editMode", "masterFiles"
      ]),
   },
   data() {
      return {
        editMF: null,
        newTitle: "",
        newDescription: "",
        editField: "",
        menuVisible: false,
      }
   },
   created() {
      this.$store.dispatch("getMasterFiles", this.$route.params.unit)
   },
   methods: {
      resetSort() {
         this.$store.commit("filenameSort")
      },
      showContextMenu(e) {
         let m = document.getElementById("popupmenu")
         m.style.left = e.pageX+"px"
         m.style.top = e.pageY+"px"
         this.menuVisible =  true
         this.$nextTick( () => {
            let mW = m.offsetWidth
            let mH = m.offsetHeight
            if ( mW +  e.pageX > window.innerWidth) {
               m.style.left = (e.pageX - mW) + "px";
            }
            if ( mH +  e.pageY > window.innerHeight) {
               m.style.top = (e.pageY - mH) + "px";
            }
         })
      },
      renameAll() {
         this.$store.dispatch("renameAll")
      },
      setPageNumbersClicked() {
         this.editMode = "page"
         this.menuVisible =  false
         this.$nextTick( () => {
            let p = document.getElementById("start-page-num")
            p.focus()
            p.select()
         })
      },
      dragStarted() {
         let eles=document.getElementsByClassName("selected")
         while (eles[0]) {
            eles[0].classList.remove('selected')
         }
      },
      fileSelected(fn, e) {
         this.menuVisible = false
         if ( e.ctrlKey ) return

         if ( e.shiftKey ) {
            // start of by considering this the end of a range
            this.rangeEndIdx = this.masterFiles.findIndex( mf => mf.fileName == fn)
            if ( this.rangeStartIdx > this.rangeEndIdx ) {
               // if not, swap indexes
               let t = this.rangeEndIdx
               this.rangeEndIdx =  this.rangeStartIdx
               this.rangeStartIdx = t
            }

            // get all of the masterfiels in the range and select them
            for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
               let tgt = this.masterFiles[i].fileName
               let tgtEle = document.getElementById(tgt)
               if (tgtEle.classList.contains("selected") == false) {
                  tgtEle.classList.add("selected")
               }
            }
         } else {
            // grab selected element and set a flag if it is not currently selected
            this.rangeStartIdx = -1
            let tgtEle = document.getElementById(fn)
            let selectIt = (tgtEle.classList.contains("selected") == false)

            // clear all selected classes
            let eles=document.getElementsByClassName("selected")
            while (eles[0]) {
               eles[0].classList.remove('selected')
            }

            // if the just-clicked element needs to be selected, select it now
            if (selectIt) {
               document.getElementById(fn).classList.add("selected")
               this.rangeStartIdx =  this.masterFiles.findIndex( mf => mf.fileName == fn)
            }
         }

      },
      isEditing(mf, field = "all") {
         return this.editMF == mf && this.editField == field
      },
      editMetadata(mf, field) {
         this.editMF = mf
         this.newTitle = mf.title
         this.newDescription = mf.description
         this.editField = field
         this.$nextTick( ()=> {
            let ele = document.getElementById("edit-title")
            if ( field == "description") {
               ele = document.getElementById("edit-desc")
            }
            ele.focus()
            ele.select()
         })
      },
      cancelEdit() {
         this.editMF = null
      },
      submitEdit(mf) {
         this.editMF = null
         this.$store.dispatch("updateMetadata", {file: mf.path, title: this.newTitle, description: this.newDescription})
      }
   }
}
</script>

<style lang="scss" scoped>
.unit {
   padding: 0;
   .load {
      margin-top: 10%;
   }
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 25px;
   }
   .toolbar {
      display: flex;
      flex-flow: row wrap;
      padding: 10px;
      background: var(--uvalib-grey-light);
      border-bottom: 1px solid var(--uvalib-grey);
      border-top: 1px solid var(--uvalib-grey);
      label {
         font-weight: bold;
         margin-right: 10px;
      }
      select {
         width:max-content;
      }
      .actions {
         margin-left: auto;
         display: flex;
         flex-flow: row wrap;
         justify-content: flex-end;
         align-content: center;
         .right-pad {
            margin-right: 10px;
         }
      }
   }
   .selected {
      background: var(--uvalib-blue-alt-light) !important;
      td {
         border-bottom: 1px solid  var(--uvalib-blue-alt) !important;
      }
   }
   .undefined {
      font-style: italic;
   }
   div.gallery {
      display: flex;
      flex-flow: row wrap;
      padding: 15px;
      text-align: left;
      justify-content: flex-start;
      align-content: flex-start;
      background: #e5e5e5;

      .edit-md {
         float: right;
         cursor: pointer;
         font-size: 1.25em;
         font-weight: bold;
         &:hover {
            color: var(--uvalib-blue-alt);
         }
      }
      .edit-ctls {
         position: relative;
         top: 10px;
         right: -5px;
         font-size: 1.25em;
         text-align: right;
         i {
            display: inline-block;
         }
         .cancel {
            color: var(--uvalib-red-darker);
            margin-right: 5px;
            &:hover {
               color: var(--uvalib-red-emergency);
            }
         }
         .accept {
            color: var(--uvalib-green-dark);
            &:hover {
               color: var(--uvalib-green-lightest);
            }
         }
      }

      div.card {
         position: relative;
         border: 1px solid var(--uvalib-grey-light);
         padding: 20px;
         display: inline-block;
         margin: 5px;
         background: white;
         box-shadow:  0 1px 3px rgba(0,0,0,.06), 0 1px 2px rgba(0,0,0,.12);
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
            background-image: url('~@/assets/dots.gif');
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
      img {
         min-width: 250px;
         min-height: 390px;
      }
   }
   div.gallery.large {
      .card .metadata .data  {
         max-width: 380px;
      }
      img {
         min-width: 400px;
         min-height: 590px;
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
         padding: 5px;
         text-align: left;
         border-bottom: 1px solid var(--uvalib-grey-lightest);
      }
      td.thumb {
         padding: 5px 5px 2px 5px;
         img {
            border:1px solid var(--uvalib-grey);
         }
      }
      th {
         border-bottom: 1px solid var(--uvalib-grey);
      }
      tr {
         &:hover {
            background:var(  --uvalib-grey-lightest);
         }
      }
   }
   .editable {
      cursor: pointer;
      &:hover {
         text-decoration: underline;
         color: var(--uvalib-blue-alt) !important;
      }
   }
   .popupmenu {
      position: absolute;
      background: var(--uvalib-blue-alt);
      box-shadow: 0 3px 6px rgba(0, 0, 0, 0.16), 0 3px 6px rgba(0, 0, 0, 0.23);
      padding: 5px;
      top: 50px;
      left: 50px;
      text-align: left;
      border-radius: 5px;
      ul {
         border-radius: 5px;
         background: white;
         list-style: none;
         margin: 0;
         padding: 0;
         border: 1px solid var(--uvalib-blue-alt-dark);
         li {
            border-radius: 5px;
            padding: 4px 15px 4px 5px;
            white-space: nowrap;
            cursor:pointer;
            &:hover {
               background: var(--uvalib-blue-alt-light);
            }
         }
      }
   }
}
</style>
