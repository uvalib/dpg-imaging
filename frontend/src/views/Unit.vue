<template>
   <div class="unit">
      <WaitSpinner v-if="working" :overlay="true" message="Loading master file..." />
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
                     <label>File Name</label>
                     <div class="data">{{mf.fileName}}</div>
                  </div>
                  <div class="row">
                     <label>Title</label>
                     <div class="data">
                        <template v-if="mf.title">{{mf.title}}</template>
                        <span v-else class="undefined">Undefined</span>
                     </div>
                  </div>
                  <div class="row">
                     <label>Caption</label>
                     <div class="data">
                        <template v-if="mf.title">{{mf.description}}</template>
                        <span v-else class="undefined">Undefined</span>
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
export default {
   name: "unit",
   computed: {
      ...mapState({
         working : state => state.working,
         masterFiles : state => state.masterFiles,
         currUnit: state => state.currUnit
      }),
      ...mapFields([
         'viewMode'
      ]),
   },
   created() {
      this.$store.dispatch("getMasterFiles", this.$route.params.unit)
   },
   methods: {
   }
}
</script>

<style lang="scss">
.unit {
   padding: 0;
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

      div.card {
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
            .undefined {
               font-style: italic;
               color: var(--uvalib-grey);
            }
         }
      }
   }
   div.gallery.medium .card .metadata .data  {
      max-width: 230px;
   }
   div.gallery.large .card .metadata .data  {
      max-width: 380px;
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
