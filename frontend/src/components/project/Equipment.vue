<template>
   <div class="panel">
      <dl v-if="!editing">
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
            <span v-if="currProject.resolutionNote">{{currProject.resolutionNote}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
      </dl>
      <table class="edit" v-else>
         <tr class="row">
            <td class="label"><label for="workstation">Workstation:</label></td>
            <td class="data">
               <select id="workstation" v-model="workstationID">
                  <option disabled :value="0">Choose a workstation</option>
                  <option v-for="ws in workstations" :key="`ws${ws.id}`" :value="ws.id">{{ws.name}}</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="capture">Capture Resolution:</label></td>
            <td><input id="capture" type="text" v-model="captureResolution"></td>
         </tr>
         <tr class="row">
            <td class="label"><label for="resize">Resized Resolution:</label></td>
            <td class="data"><input id="resize" type="text" v-model="resizedResolution"></td>
         </tr>
         <tr class="row">
            <td class="label"><label for="res-note">Resolution Note:</label></td>
            <td class="data"><textarea id="res-note" v-model="resolutionNote"></textarea></td>
         </tr>
      </table>
      <div class="buttons" v-if="isOwner(computingID)">
         <DPGButton v-if="!editing" @click="editClicked">Edit</DPGButton>
         <template v-else>
            <DPGButton @click="cancelClicked">Cancel</DPGButton>
            <DPGButton @click="saveClicked">Save</DPGButton>
         </template>
      </div>
   </div>
</template>

<script>
import { mapGetters, mapState } from "vuex";
export default {
   data: function()  {
      return {
         editing: false,
         workstationID: 0,
         captureResolution: "",
         resizedResolution: "",
         resolutionNote: ""
      }
   },
   computed: {
      ...mapState({
         workstations: state => state.workstations,
         computingID: state => state.user.computeID,
      }),
      ...mapGetters({
         currProject: "projects/currProject",
         isOwner: "projects/isOwner"
      }),
   },
   methods: {
      editClicked() {
         this.workstationID = 0
         this.captureResolution = ""
         this.resizedResolution = ""
         this.resolutionNote = ""
         this.editing = true
      },
      cancelClicked() {
         this.editing = false
      },
      async saveClicked() {
         let data = {
            workstationID: this.workstationID,
            captureResolution: parseInt(this.captureResolution, 10),
            resizeResolution: parseInt(this.resizedResolution, 10),
            resolutionNote: this.resolutionNote
         }
         await this.$store.dispatch("projects/setEquipment", data)
         this.editing = false
      }
   },
};
</script>

<style scoped lang="scss">
.panel {
   padding: 10px;
   .buttons {
      padding: 0;
      margin: 0;
      text-align: right;
      button {
         margin-left: 10px;
      }
   }
   .edit {
      font-size: 0.9em;
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 5px;
      td {
          padding: 5px 0px 5px 10px;
      }
      td.data {
         width: 100%;
         input, select {
            border-color: var(--uvalib-grey-light);
         }
      }
      td.label {
         font-weight: bold;
         margin-right: 10px;
         text-align: right;
         vertical-align: top;
         white-space: nowrap;
      }
      textarea {
         width: 100%;
         box-sizing: border-box;
         border-color: var(--uvalib-grey-light);
         border-radius: 5px;
         padding: 5px;
      }
   }
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
