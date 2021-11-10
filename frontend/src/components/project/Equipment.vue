<template>
   <div class="panel">
      <dl>
         <dt>Workstation:</dt>
         <dd>
            <span v-if="currProject.workstation.id > 0">{{currProject.workstation.name}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <template v-if="currProject.workstation.id > 0">
            <dt>Setup:</dt>
            <dd>
               <table>
                  <tr v-for="e in currProject.equipment" :key="e.serialNumber">
                     <td>{{e.type}}</td>
                     <td>{{e.name}}</td>
                     <td>{{e.serialNumber}}</td>
                  </tr>
               </table>
            </dd>
         </template>
         <dt>Capture resolution:</dt>
         <dd>
            <span v-if="currProject.captureResolution">{{currProject.captureResolution}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Resized resolution:</dt>
         <dd>
            <span v-if="currProject.resizedResolution">{{currProject.resizedResolution}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Resolution note:</dt>
         <dd>
            <span v-if="currProject.resoltionNote">{{currProject.resoltionNote}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
      </dl>
   </div>
</template>

<script>
import { mapGetters } from "vuex";
export default {
   computed: {
      ...mapGetters({
         currProject: "projects/currProject",
      }),
    },
   methods: {

   },
};
</script>

<style scoped lang="scss">
.panel {
   padding: 10px;
   dl {
      margin: 10px 30px 0 30px;
      display: inline-grid;
      grid-template-columns: max-content 2fr;
      grid-column-gap: 10px;
      font-size: 0.9em;
      text-align: left;
      box-sizing: border-box;
      width: 100%;

      dt {
         font-weight: bold;
         text-align: right;
      }
      dd {
         margin: 0 0 10px 0;
         word-break: break-word;
         -webkit-hyphens: auto;
         -moz-hyphens: auto;
         hyphens: auto;
         .na {
            color: #999;
         }
         table {
            width: 100%;
            font-size: 0.8em;
         }
      }
   }
}
</style>
