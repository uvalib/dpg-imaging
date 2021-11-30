<template>
   <div class="search">
      <h3>Search</h3>
      <div class="form">
         <label for="workflow">Workflow</label>
         <select id="workflow"  v-model="tgtWorkflow">
            <option :value="0">Any</option>
            <option v-for="w in workflows" :key="`wf${w.id}`" :value="w.id">{{w.name}}</option>
         </select>

         <label for="assigned">Assigned To</label>
         <select id="assigned" v-model="tgtAssignedTo">
            <option :value="0">Any</option>
            <option v-for="sm in staff" :key="`sm${sm.id}`" :value="sm.id">{{sm.lastName}}, {{sm.firstName}}</option>
         </select>

         <label for="call">Call Number</label>
         <input id="call" v-model="tgtCallNumber">

         <label for="customer">Customer Last Name</label>
         <input id="customer" v-model="tgtCustomer">

         <label for="agency">Agency</label>
         <input id="agency" v-model="tgtAgency">

         <label for="workstation">Workstation</label>
         <select id="workstation" v-model="tgtWorkstation">
            <option :value="0">Any</option>
            <option v-for="ws in workstations" :key="`ws${ws.id}`" :value="ws.id">{{ws.name}}</option>
         </select>
      </div>
      <div class="buttons">
         <DPGButton @clicked="resetSearch">Reset Search</DPGButton>
         <DPGButton @clicked="doSearch">Search</DPGButton>
      </div>
   </div>
</template>

<script>
import { mapState } from "vuex"
import { mapFields } from 'vuex-map-fields'
export default {
   components: {
   },
   computed: {
      ...mapState({
         workstations : state => state.workstations,
         workflows : state => state.workflows,
         staff : state => state.staffMembers,
      }),
      ...mapFields({
         search: "projects.search",
         tgtWorkflow: "projects.search.workflow",
         tgtAssignedTo: "projects.search.assignedTo",
         tgtWorkstation: "projects.search.workstation",
         tgtCallNumber: "projects.search.callNumber",
         tgtCustomer: "projects.search.customer",
         tgtAgency: "projects.search.agency",
      })
   },
   methods: {
      resetSearch() {
         this.$store.dispatch("projects/resetSearch")
      },
      doSearch() {
         this.$store.dispatch("projects/getProjects")
      }
   },
};
</script>

<style scoped lang="scss">
.search {
   background: #f6f6f6;
   width: 20%;
   min-width: 270px;
   border: 1px solid var(--uvalib-grey);
   box-shadow: rgba(0, 0, 0, 0.14) 0px 2px 2px 0px;
   h3 {
      text-align: center;
      padding: 5px;
      margin: 0;
      background: var(--uvalib-grey-lightest);
      border-bottom: 1px solid var(--uvalib-grey);
      font-size: 1em;
   }
   .form {
      background: white;
      min-height: 300px;
      text-align: left;
      padding: 10px 10px 0 10px;
      label {
         display: block;
      }
      select, input{
         margin: 2px 0 15px 0;
      }
   }
   .buttons {
      background: white;
      padding: 0 10px 10px 10px;
      text-align: right;
      button {
         margin-left: 5px;
      }
   }
}
</style>
