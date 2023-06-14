<template>
   <div class="search">
      <h3>Search</h3>
      <div class="form">
         <label for="workflow">Workflow</label>
         <Dropdown id="workflow" v-model="projectStore.search.workflow" :options="workflows"
            optionLabel="name" optionValue="id" @change="doSearch()" />
         <label for="assigned">Assigned To</label>
         <Dropdown id="assigned" v-model="projectStore.search.assignedTo" :options="staffMembers"
            optionLabel="name" optionValue="id"
            filter autoFilterFocus resetFilterOnHide filterMatchMode="startsWith"
            @change="doSearch()" />

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
            optionLabel="name" optionValue="id"
            filter autoFilterFocus resetFilterOnHide  filterMatchMode="startsWith"
            @change="doSearch()" />

         <label for="workstation">Workstation</label>
         <Dropdown id="workstation" v-model="projectStore.search.workstation" :options="workstations"
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
      out.push( { name: `${sm.lastName}, ${sm.firstName}`, id: sm.id})
   })
   return out
})
const workflows = computed( () => {
   let out = [ {name: "Any", id: 0} ]
   systemStore.workflows.forEach( sm => {
      if (sm.isActive ) {
         out.push( { name: sm.name, id: sm.id})
      }
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

const resetSearch = ( async() => {
   let query = Object.assign({}, route.query)
   delete query.workflow
   delete query.owner
   delete query.order
   delete query.unit
   delete query.callnum
   delete query.customer
   delete query.agency
   delete query.workstation
   delete query.filter
   router.push({query})

   projectStore.resetSearch()
})

const doSearch = ( async () => {
   projectStore.getProjects()

   let query = Object.assign({}, route.query)
   delete query.workflow
   if (projectStore.search.workflow != 0) {
      query.workflow = projectStore.search.workflow
   }

   delete query.owner
   if (projectStore.search.assignedTo > 0 ) {
      console.log("set owner")
      query.owner = projectStore.search.assignedTo
   } else {
      console.log("fr")
   }

   delete query.order
   if (projectStore.search.orderID != "") {
      query.order = projectStore.search.orderID
   }

   delete query.unit
   if (projectStore.search.unitID != "") {
      query.unit = projectStore.search.unitID
   }

   delete query.callnum
   if (projectStore.search.callNumber != "") {
      query.callnum = projectStore.search.callNumber
   }

   delete query.customer
   if (projectStore.search.customer != "") {
      query.customer = projectStore.search.customer
   }

   delete query.agency
   if (projectStore.search.agency > 0) {
      query.agency = projectStore.search.agency
   }

   delete query.workstation
   if (projectStore.search.workstation > 0) {
      query.workstation =projectStore.search.workstation
   }
   await router.push({query})
   projectStore.lastSearchURL = route.fullPath
})
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
