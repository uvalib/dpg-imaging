<template>
   <div class="unit">
      <WaitSpinner v-if="working" :overlay="true" message="Loading master file..." />
      <template v-else>
         <h2>Unit {{currUnit}}</h2>
         <table class="unit-list">
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
      </template>
   </div>
</template>

<script>
import { mapState } from "vuex"
export default {
   name: "unit",
   computed: {
      ...mapState({
         working : state => state.working,
         masterFiles : state => state.masterFiles,
         currUnit: state => state.currUnit
      })
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
   padding: 15px;
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 30px;
   }
   table {
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
         padding-left: 0;
      }
   }
}
</style>
