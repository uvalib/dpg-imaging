<template>
   <div class="search">
      <h3>Search</h3>
      <div class="form">
         <label for="workflow">Workflow</label>
         <select id="workflow"  v-model="projectStore.search.workflow">
            <option :value="0">Any</option>
            <option v-for="w in systemStore.workflows" :key="`wf${w.id}`" :value="w.id">{{w.name}}</option>
         </select>

         <label for="assigned">Assigned To</label>
         <select id="assigned" v-model="projectStore.search.assignedTo">
            <option :value="0">Any</option>
            <option v-for="sm in systemStore.staffMembers" :key="`sm${sm.id}`" :value="sm.id">{{sm.lastName}}, {{sm.firstName}}</option>
         </select>

         <label for="unit">Order</label>
         <input id="unit" v-model="projectStore.search.orderID">

         <label for="unit">Unit</label>
         <input id="unit" v-model="projectStore.search.unitID">

         <label for="call">Call Number</label>
         <input id="call" v-model="projectStore.search.callNumber">

         <label for="customer">Customer Last Name</label>
         <input id="customer" v-model="projectStore.search.customer">

         <label for="agency">Agency</label>
         <select id="agency" v-model="projectStore.search.agency">
            <option :value="0">Any</option>
            <option v-for="a in systemStore.agencies" :key="`agency${a.id}`" :value="a.id">{{a.name}}</option>
         </select>

         <label for="workstation">Workstation</label>
         <select id="workstation" v-model="projectStore.search.workstation">
            <option :value="0">Any</option>
            <option v-for="ws in systemStore.workstations" :key="`ws${ws.id}`" :value="ws.id">{{ws.name}}</option>
         </select>
      </div>
      <div class="buttons">
         <DPGButton class="p-button-secondary" @click="resetSearch" label="Reset Search"/>
         <DPGButton class="p-button-secondary" @click="doSearch" label="Search"/>
      </div>
   </div>
</template>

<script setup>
import {useProjectStore} from '@/stores/project'
import {useSystemStore} from '@/stores/system'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const systemStore = useSystemStore()

async function resetSearch() {
   let query = Object.assign({}, route.query)
   if (query.order) {
      delete query.order
      await router.push({ query })
   } else if (query.unit) {
      delete query.unit
      await router.push({ query })
   }
   projectStore.resetSearch()
}
function doSearch() {
   projectStore.getProjects()
}
</script>

<style scoped lang="scss">
.search {
   width: 20%;
   min-width: 270px;
   border: 1px solid var(--uvalib-grey);
   box-shadow: rgba(0, 0, 0, 0.14) 0px 2px 2px 0px;
   h3 {
      text-align: center;
      padding: 5px;
      margin: 0;
      background: var(--uvalib-blue-alt-light);
      border-bottom: 1px solid var(--uvalib-grey);
      font-size: 1em;
      font-weight: normal;
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
