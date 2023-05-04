<template>
   <div class="search">
      <h3>Search</h3>
      <div class="form">
         <label for="workflow">Workflow</label>
         <Dropdown id="workflow" v-model="projectStore.search.workflow" :options="workflows"
            optionLabel="name" optionValue="id" @change="doSearch()" />
         <label for="assigned">Assigned To</label>
         <Dropdown id="assigned" v-model="projectStore.search.assignedTo" :options="staffMembers"
            optionLabel="name" optionValue="id" filter  @change="doSearch()" />

         <label for="unit">Order</label>
         <input id="unit" v-model="projectStore.search.orderID" @keyup.enter="doSearch()">

         <label for="unit">Unit</label>
         <input id="unit" v-model="projectStore.search.unitID" @keyup.enter="doSearch()">

         <label for="call">Call Number</label>
         <input id="call" v-model="projectStore.search.callNumber" @keyup.enter="doSearch()">

         <label for="customer">Customer Last Name</label>
         <input id="customer" v-model="projectStore.search.customer" @keyup.enter="doSearch()">

         <label for="agency">Agency</label>
         <Dropdown id="agency" v-model="projectStore.search.agency" :options="agencies"
            optionLabel="name" optionValue="id" filter  @change="doSearch()" />

         <label for="workstation">Workstation</label>
         <Dropdown id="workflow" v-model="projectStore.search.workflow" :options="workflows"
            optionLabel="name" optionValue="id" @change="doSearch()" />
      </div>
      <div class="buttons">
         <DPGButton class="p-button-secondary" @click="resetSearch" label="Reset Search"/>
         <DPGButton class="p-button-secondary" @click="doSearch" label="Search"/>
      </div>
   </div>
</template>

<script setup>
import { useProjectStore } from '@/stores/project'
import {useSystemStore} from '@/stores/system'
import { useRoute, useRouter } from 'vue-router'
import Dropdown from 'primevue/dropdown'
import { computed } from 'vue'

const route = useRoute()
const router = useRouter()
const projectStore = useProjectStore()
const systemStore = useSystemStore()

const staffMembers = computed( () => {
   let out = [ {name: "Any", id: 0} ]
   systemStore.staffMembers.forEach( sm => {
      out.push( { name: `${sm.lastName}, ${sm.firstName}`, code: sm.id})
   })
   return out
})
const workflows = computed( () => {
   let out = [ {name: "Any", id: 0} ]
   systemStore.workflows.forEach( sm => {
      out.push( { name: sm.name, id: sm.id})
   })
   return out
})
const agencies = computed( () => {
   let out = [ {name: "Any", id: 0} ]
   systemStore.agencies.forEach( sm => {
      out.push( { name: sm.name, id: sm.id})
   })
   return out
})
const workstations = computed( () => {
   let out = [ {name: "Any", id: 0} ]
   systemStore.workstations.forEach( sm => {
      out.push( { name: sm.name, id: sm.id})
   })
   return out
})

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
:deep(span.p-dropdown-label.p-inputtext) {
   font-size: 0.9em;
   padding: 5px 8px;
   color: var(--uvalib-text);
}
:deep(div.p-dropdown-trigger) {
   width: auto;
   margin-right: 8px;
}
.search {
   width: 20%;
   min-width: 270px;
   border: 1px solid var(--uvalib-grey);
   box-shadow: rgba(0, 0, 0, 0.14) 0px 2px 2px 0px;
   div.p-dropdown.p-component {
      width: 100%;
      margin-bottom: 15px;
      margin-top: 2px;
      font-size: 0.9em;
   }

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
