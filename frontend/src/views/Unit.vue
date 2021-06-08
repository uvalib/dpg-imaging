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
               <span tabindex="0" id="rename" class="button">Batch Rename</span>
               <span tabindex="0" id="set-titles" @click="setPageNumbersClicked" class="button">Set Page Numbers</span>
            </span>
         </div>
         <PageNumPanel v-if="editMode == 'page'" />
         <table class="unit-list" v-if="viewMode == 'list'">
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
            <tr v-for="mf in masterFiles" :key="mf.fileName" @mousedown.prevent="fileSelected(mf.fileName, $event)" :id="mf.fileName">
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
         </table>
         <div class="gallery" :class="viewMode" v-else>
            <div class="card" v-for="mf in masterFiles" :key="mf.fileName"  @mousedown.prevent="fileSelected(mf.fileName, $event)" :id="mf.fileName">
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
         </div>
      </template>
   </div>
</template>

<script>
import { mapState } from "vuex"
import { mapFields } from 'vuex-map-fields'
import WaitSpinner from '../components/WaitSpinner.vue'
import PageNumPanel from '../components/PageNumPanel.vue'
export default {
   components: { WaitSpinner,PageNumPanel },
   name: "unit",
   computed: {
      ...mapState({
         loading : state => state.loading,
         updating : state => state.updating,
         masterFiles : state => state.masterFiles,
         currUnit: state => state.currUnit
      }),
      ...mapFields([
         'viewMode', "rangeStart", "rangeEnd", "editMode"
      ]),
   },
   data() {
      return {
        editMF: null,
        newTitle: "",
        newDescription: "",
        editField: "",
      }
   },
   created() {
      this.$store.dispatch("getMasterFiles", this.$route.params.unit)
   },
   methods: {
      setPageNumbersClicked() {
         this.editMode = "page"
         this.$nextTick( () => {
            document.getElementById("start-page").focus()
         })
      },
      fileSelected(fn, e) {
         if (e.shiftKey) {
            let startNum = parseInt(this.rangeStart.replace(".tif","").split("_")[1],10)
            let endNum = parseInt(fn.replace(".tif","").split("_")[1],10)
            this.rangeEnd = fn
            if ( this.rangeStart > fn) {
               let t = endNum
               endNum = startNum
               startNum = t
               this.rangeEnd = this.rangeStart
               this.rangeStart = fn
            }
            for (let i=startNum; i<=endNum; i++) {
               let numStr = ""+i
               let tgt = this.currUnit+"_"+numStr.padStart(4,0)+".tif"
               let tgtEle = document.getElementById(tgt)
               if (tgtEle.classList.contains("selected") == false) {
                  tgtEle.classList.add("selected")
               }
            }
         } else {
            this.rangeStart = ""
            let tgtEle = document.getElementById(fn)
            let selectIt = (tgtEle.classList.contains("selected") == false)
            let eles=document.getElementsByClassName("selected")
            while (eles[0]) {
               eles[0].classList.remove('selected')
            }
            if (selectIt) {
               document.getElementById(fn).classList.add("selected")
               this.rangeStart = fn
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
      }
      #rename {
         margin-right: 10px;
      }
   }
   .selected {
      background:  var(--uvalib-yellow-light) !important;
   }
   .undefined {
      font-style: italic;
      color: var(--uvalib-grey);
   }
   div.gallery {
      display: flex;
      flex-flow: row wrap;
      padding: 15px;
      justify-content: flex-start;
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
         padding-left: 5px;
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
}
</style>
