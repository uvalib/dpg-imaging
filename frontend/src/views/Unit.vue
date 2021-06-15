<template>
   <div class="unit">
      <WaitSpinner v-if="updating" :overlay="true" message="Updating data..." />
      <div class="load" v-if="loading">
         <WaitSpinner v-if="loading" message="Loading master file..." />
      </div>
      <template v-else>
         <div class="metadata">
            <h2>
               <ProblemsDisplay class="topleft" />
               <span>{{title}}</span>
            </h2>
            <h3>
               <div>{{callNumber}}</div>
               <div>Unit {{currUnit}}</div>
            </h3>
            <div><a :href="projectURL" target="_blank">TrackSys Project<i class="link fas fa-external-link-alt"></i></a></div>
         </div>
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
               <DPGButton id="set-titles" @click="setPageNumbersClicked" class="button right-pad">Set Page Numbers</DPGButton>
               <DPGButton id="set-titles" @click="componentLinkClicked" class="button">Component Link</DPGButton>
            </span>
         </div>
         <PageNumPanel v-if="editMode == 'page'" />
         <ComponentPanel v-if="editMode == 'component'" />
         <table class="unit-list" v-if="viewMode == 'list'">
            <thead>
               <tr>
                  <th></th>
                  <th>Tag</th>
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
                  @contextmenu.prevent="showContextMenu(mf.fileName, $event)"
               >
                  <td class="thumb">
                     <router-link :to="`/unit/${currUnit}/page/${mf.fileName.replace('.tif','').split('_')[1]}`"><img :src="mf.thumbURL"/></router-link>
                  </td>
                  <td><TagPicker :masterFile="mf" /></td>
                  <td>{{mf.fileName}}</td>
                  <td>{{mf.fileType}}</td>
                  <td>{{mf.resolution}}</td>
                  <td @click="editMetadata(mf, 'title')" class="editable">
                     <span  v-if="!isEditing(mf, 'title')"  class="editable">
                        <span v-if="mf.title">{{mf.title}}</span>
                        <span v-else class="undefined">Undefined</span>
                     </span>
                     <TitleInput v-else @canceled="cancelEdit" @accepted="submitEdit(mf)" v-model="newTitle"/>
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
         <draggable v-else v-model="masterFiles" @start="dragStarted" class="gallery" :class="viewMode" >
            <div class="card" v-for="mf in masterFiles" :key="mf.fileName"
               @mousedown="fileSelected(mf.fileName, $event)"
               @contextmenu.prevent="showContextMenu(mf.fileName, $event)"
               :id="mf.fileName"
            >
               <router-link :to="`/unit/${currUnit}/page/${mf.fileName.replace('.tif','').split('_')[1]}`">
                  <img :src="mf.mediumURL" v-if="viewMode == 'medium'"/>
                  <img :src="mf.largeURL" v-if="viewMode == 'large'"/>
               </router-link>
               <div class="tag">
                  <TagPicker :masterFile="mf" display="wide"/>
               </div>
               <div class="metadata">
                  <div class="row">
                     <label>Image</label>
                     <div class="data">{{mf.fileName}}</div>
                     <div class="data">{{mf.width}} x {{mf.height}}, {{mf.resolution}}</div>
                  </div>
                  <div class="row">
                     <label>Title</label>
                     <div class="data editable" @click="editMetadata(mf, 'title')">
                        <template v-if="isEditing(mf, 'title')">
                           <TitleInput  @canceled="cancelEdit" @accepted="submitEdit(mf)" v-model="newTitle"/>
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
      </template>
      <div class="popupmenu" id="popupmenu" v-show="menuVisible">
         <ul @keydown.esc="clearState">
            <li tabindex="0" @click.stop.prevent="setPageNumbersClicked" class="menuitem"
               @keydown.exact.shift.tab.stop.prevent="focusLastmenu()"
            >
               Set Page Numbers
            </li>
            <li tabindex="0" @click.stop.prevent="componentLinkClicked" class="menuitem">
               Component Link
            </li>
            <li tabindex="0" class="menuitem">
               <ConfirmModal label="Delete Image" type="text" @confirmed="deleteSelected" @closed="menuVisible=false" >
                  <div>Delete image {{rightClickedMF}}? This cannot be reversed.</div>
               </ConfirmModal>
            </li>
            <li tabindex="0" @click="menuVisible = false" class="menuitem" @keydown.exact.tab.stop.prevent="focusFirstmenu()">
               Close Menu
            </li>
         </ul>
      </div>
   </div>
</template>

<script>
import { mapState } from "vuex"
import { mapFields } from 'vuex-map-fields'
import ComponentPanel from '../components/ComponentPanel.vue'
import PageNumPanel from '../components/PageNumPanel.vue'
import TagPicker from '../components/TagPicker.vue'
import TitleInput from '../components/TitleInput.vue'
import ProblemsDisplay from '../components/ProblemsDisplay.vue'
import draggable from 'vuedraggable'
export default {
   components: {PageNumPanel, draggable, TagPicker, TitleInput, ProblemsDisplay, ComponentPanel },
   name: "unit",
   computed: {
      ...mapState({
         loading : state => state.loading,
         updating : state => state.updating,
         currUnit: state => state.currUnit,
         title: state => state.title,
         callNumber: state => state.callNumber,
         projectURL: state => state.projectURL,
      }),
      ...mapFields([
         'viewMode', "rangeStartIdx", "rangeEndIdx", "editMode", "masterFiles"
      ]),
   },
   data() {
      return {
        editMF: null,
        rightClickedMF: "",
        newTitle: "",
        newDescription: "",
        editField: "",
        menuVisible: false,
      }
   },
   created() {
      this.$store.dispatch("getUnitDetails", this.$route.params.unit)
   },
   methods: {
      clearState() {
          this.rightClickedMF = ""
          this.menuVisible= false
          this.rangeStartIdx = -1
          this.rangeEndIdx = -1
      },
      deleteSelected() {
         this.$store.dispatch("deleteMasterFile", this.rightClickedMF)
      },
      resetSort() {
         this.$store.commit("filenameSort")
      },
      showContextMenu(fileName, e) {
         this.rightClickedMF = fileName
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
            let items = document.getElementsByClassName("menuitem")
            if (items.length > 0) {
               items[0].focus()
            }
         })
      },
      focusLastmenu() {
          let items = document.getElementsByClassName("menuitem")
            if (items.length > 0) {
               items[items.length-1].focus()
            }
      },
      focusFirstmenu() {
          let items = document.getElementsByClassName("menuitem")
            if (items.length > 0) {
               items[0].focus()
            }
      },
      renameAll() {
         this.$store.dispatch("renameAll")
      },
      componentLinkClicked() {
         console.log("COMPONENT CLICK")
         this.editMode = "component"
         this.menuVisible =  false
         this.$nextTick( () => {
            let p = document.getElementById("component-id")
            p.focus()
            p.select()
         })
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

            let eles=document.getElementsByClassName("selected")
            while (eles[0]) {
               eles[0].classList.remove('selected')
            }

            for (let i=this.rangeStartIdx; i<=this.rangeEndIdx; i++) {
               let tgt = this.masterFiles[i].fileName
               let tgtEle = document.getElementById(tgt)
               tgtEle.classList.add("selected")
            }
         } else {
            // grab selected element and set a flag if it is not currently selected
            this.rangeStartIdx = -1
            this.rangeEndIdx = -1
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
            if ( field == "description") {
               let ele = document.getElementById("edit-desc")
               ele.focus()
               ele.select()
            }
         })
      },
      cancelEdit() {
         this.editMF = null
      },
      async submitEdit(mf) {
         await this.$store.dispatch("updateMasterFileMetadata", {file: mf.path, title: this.newTitle, description: this.newDescription, status: mf.status})
         this.editMF = null
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
   .metadata {
      margin-bottom: 15px;
      position: relative;
      .topleft {
         position: absolute;
         top:0;
         left: 0px;
      }
      h2 {
         color: var(--uvalib-brand-orange);
         margin: 10px 0;
      }
      h3 {
          margin: 5px 0;
          font-weight: normal;
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
   .undefined {
      font-style: italic;
   }
   div.gallery {
      padding: 15px;
      text-align: left;
      background: #e5e5e5;

      display: flex;
      flex-flow: row wrap;
      justify-content: flex-start;
      align-content: flex-start;

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
      div.card.selected {
         background: var(--uvalib-blue-alt-light);
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
            background: aliceblue;
         }
      }
      tr.selected {
         &:hover {
            background: aliceblue;
         }
      }
      .selected {
         background: var(--uvalib-blue-alt-light);
         td {
            border-bottom: 1px solid  var(--uvalib-blue-alt-light);
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
      padding: 0;
      top: 50px;
      left: 50px;
      text-align: left;
      ul {
         background: white;
         list-style: none;
         margin: 0;
         padding: 0;
         border: 1px solid var(--uvalib-blue-alt-dark);
         li {
            padding: 4px 15px 4px 5px;
            white-space: nowrap;
            outline: 0;
            cursor:pointer;
            &:hover {
               background: var(--uvalib-blue-alt-light);
            }
         }
      }
   }
}
</style>
