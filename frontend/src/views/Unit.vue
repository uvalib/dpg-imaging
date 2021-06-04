<template>
   <div class="unit">
      <WaitSpinner v-if="updating" :overlay="true" message="Updating data..." />
      <div class="load" v-if="loading">
         <WaitSpinner v-if="loading" message="Loading master file..." />
      </div>
      <template v-else>
         <h2>Unit {{currUnit}}</h2>
         <div class="toolbar">
            <label>View:</label>
            <select id="layout" v-model="viewMode">
               <option value="list">List</option>
               <option value="medium">Gallery (medium)</option>
               <option value="large">Gallery (large)</option>
            </select>
         </div>
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
            <tr v-for="mf in masterFiles" :key="mf.fileName">
               <td class="thumb">
                  <router-link :to="`/unit/${currUnit}/page/${mf.fileName.replace('.tif','').split('_')[1]}`"><img :src="mf.thumbURL"/></router-link>
               </td>
               <td>{{mf.fileName}}</td>
               <td>{{mf.fileType}}</td>
               <td>{{mf.resolution}}</td>
               <td>{{mf.title}}</td>
               <td>{{mf.description}}</td>
               <td>{{mf.colorProfile}}</td>
               <td>{{mf.path}}</td>
            </tr>
         </table>
         <div class="gallery" :class="viewMode" v-else>
            <div class="card" v-for="mf in masterFiles" :key="mf.fileName">
               <router-link :to="`/unit/${currUnit}/page/${mf.fileName.replace('.tif','').split('_')[1]}`">
                  <img :src="mf.mediumURL" v-if="viewMode == 'medium'"/>
                  <img :src="mf.largeURL" v-if="viewMode == 'large'"/>
               </router-link>
               <div class="metadata">
                  <div class="row">
                     <i tabindex="0" @click="editMetadata(mf)" class="edit-md fas fa-edit" v-if="!isEditing(mf)"></i>
                     <label>File Name</label>
                     <div class="data">{{mf.fileName}}</div>
                  </div>
                  <div class="row">
                     <label>Title</label>
                     <div class="data">
                        <template v-if="isEditing(mf)">
                           <input id="edit-title" type="text" v-model="newTitle"  @keyup.enter="submitEdit(mf)" />
                        </template>
                        <template v-else>
                           <template v-if="mf.title">{{mf.title}}</template>
                           <span v-else class="undefined">Undefined</span>
                        </template>
                     </div>
                  </div>
                  <div class="row">
                     <label>Caption</label>
                     <div class="data">
                        <template v-if="isEditing(mf)">
                           <input id="edit-desc" type="text" v-model="newDescription" @keyup.enter="submitEdit(mf)" />
                        </template>
                        <template v-else>
                           <template v-if="mf.title">{{mf.description}}</template>
                           <span v-else class="undefined">Undefined</span>
                        </template>
                     </div>
                     <div class="edit-ctls">
                        <i tabindex="0" @click="cancelEdit" class="cancel fas fa-times-circle" v-if="isEditing(mf)"></i>
                        <i tabindex="0" @click="submitEdit(mf)" class="accept fas fa-check-circle" v-if="isEditing(mf)"></i>
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
export default {
   components: { WaitSpinner },
   name: "unit",
   computed: {
      ...mapState({
         loading : state => state.loading,
         updating : state => state.updating,
         masterFiles : state => state.masterFiles,
         currUnit: state => state.currUnit
      }),
      ...mapFields([
         'viewMode'
      ]),
   },
   data() {
      return {
        editMF: null,
        newTitle: "",
        newDescription: ""
      }
   },
   created() {
      this.$store.dispatch("getMasterFiles", this.$route.params.unit)
   },
   methods: {
      isEditing(mf) {
         return this.editMF == mf
      },
      editMetadata(mf) {
         this.editMF = mf
         this.newTitle = mf.title
         this.newDescription = mf.description
         this.$nextTick( ()=> {
            let ele = document.getElementById("edit-title")
            ele.focus()
            ele.select()
         })
      },
      cancelEdit() {
         this.editMF = null
      },
      submitEdit(mf) {
         this.editMF = null
         this.$store.dispatch("updateMetadata", {file: mf.fileName, title: this.newTitle, description: this.newDescription})
      }
   }
}
</script>

<style lang="scss">
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
   }
   div.gallery {
      display: flex;
      flex-flow: row wrap;
      padding: 15px;
      justify-content: space-evenly;
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
         padding: 10px;
         display: inline-block;
         margin: 15px;
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
            input {
               box-sizing: border-box;
               width:100%;
               border-radius: 3px;
               padding: 3px 5px;
               border: 1px solid var(--uvalib-grey-light);
            }
            .undefined {
               font-style: italic;
               color: var(--uvalib-grey);
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
      th {
         background-color: var(--uvalib-grey-lightest);
      }
      th,td {
         padding: 5px 10px;
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
}
</style>
