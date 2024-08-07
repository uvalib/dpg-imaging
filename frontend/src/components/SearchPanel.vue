<template>
   <div class="search" id="search-panel">
      <h3>Search</h3>
      <div class="form">
         <div class="row">
            <label for="workflow">Workflow</label>
            <Select inputId="workflow" v-model="searchStore.search.workflow" :options="workflows"
               optionLabel="name" optionValue="id" @change="doSearch()" />
         </div>

         <div class="row">
            <label for="step">Step</label>
            <Select inputId="step" v-model="searchStore.search.step" :options="steps"
               optionLabel="name" optionValue="id" @change="doSearch()" />
            </div>

         <div class="row">
            <label for="assigned">Assigned To</label>
            <Select inputId="assigned" v-model="searchStore.search.assignedTo" :options="staffMembers"
               optionLabel="name" optionValue="id"
               filter autoFilterFocus resetFilterOnHide filterMatchMode="startsWith"
               @change="doSearch()" />
            </div>

         <div class="row">
            <label for="order">Order</label>
            <input id="order" v-model="searchStore.search.orderID" @keyup.enter="doSearch()">
         </div>

         <div class="row">
            <label for="unit">Unit</label>
            <input id="unit" v-model="searchStore.search.unitID" @keyup.enter="doSearch()">
         </div>

         <div class="row">
            <label for="call">Call Number</label>
            <input id="call" v-model="searchStore.search.callNumber" @keyup.enter="doSearch()">
         </div>

         <div class="row">
            <label for="customer">Customer Last Name</label>
            <input id="customer" v-model="searchStore.search.customer" @keyup.enter="doSearch()">
         </div>

         <div class="row">
            <label for="agency">Agency</label>
            <Select inputId="agency" v-model="searchStore.search.agency" :options="agencies"
               optionLabel="name" optionValue="id"
               filter autoFilterFocus resetFilterOnHide  filterMatchMode="startsWith"
               @change="doSearch()" />
            </div>

         <div class="row">
            <label for="workstation">Workstation</label>
            <Select inputId="workstation" v-model="searchStore.search.workstation" :options="workstations"
               optionLabel="name" optionValue="id" @change="doSearch()" />
         </div>
      </div>
      <div class="buttons">
         <DPGButton severity="secondary" @click="resetSearch" label="Reset Search"/>
         <DPGButton severity="secondary" @click="doSearch" label="Search"/>
      </div>
   </div>
</template>

<script setup>
import { useSearchStore } from '@/stores/search'
import { useSystemStore } from '@/stores/system'
import { useRoute, useRouter } from 'vue-router'
import Select from 'primevue/select'
import { computed } from 'vue'

const route = useRoute()
const router = useRouter()
const searchStore = useSearchStore()
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
const steps = computed( ()=> {
   let out = [ {name: "Any", id: "any"} ]
   systemStore.steps.forEach( sm => {
      out.push( { name: sm, id: sm})
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
   delete query.step
   delete query.owner
   delete query.order
   delete query.unit
   delete query.callnum
   delete query.customer
   delete query.agency
   delete query.workstation
   delete query.filter
   router.push({query})

   searchStore.resetSearch()
})

const doSearch = ( async () => {
   // this will reset the page number and trigger a search
   searchStore.setCurrentPage(0)

   let query = Object.assign({}, route.query)
   delete query.workflow
   if (searchStore.search.workflow != 0) {
      query.workflow = searchStore.search.workflow
   }

   delete query.step
   if (searchStore.search.step != "any") {
      query.step = searchStore.search.step
   }

   delete query.owner
   if (searchStore.search.assignedTo > 0 ) {
      query.owner = searchStore.search.assignedTo
   }

   delete query.order
   if (searchStore.search.orderID != "") {
      query.order = searchStore.search.orderID
   }

   delete query.unit
   if (searchStore.search.unitID != "") {
      query.unit = searchStore.search.unitID
   }

   delete query.callnum
   if (searchStore.search.callNumber != "") {
      query.callnum = searchStore.search.callNumber
   }

   delete query.customer
   if (searchStore.search.customer != "") {
      query.customer = searchStore.search.customer
   }

   delete query.agency
   if (searchStore.search.agency > 0) {
      query.agency = searchStore.search.agency
   }

   delete query.workstation
   if (searchStore.search.workstation > 0) {
      query.workstation =searchStore.search.workstation
   }
   await router.push({query})
   searchStore.lastSearchURL = route.fullPath
})
</script>

<style scoped lang="scss">
.search.pinned {
   position: fixed;
}

.search {
   border: 1px solid var(--uvalib-grey);
   border-radius: 4px;

   h3 {
      text-align: center;
      padding: 5px;
      margin: 0;
      background: var(--uvalib-blue-alt-light);
      border-bottom: 1px solid var(--uvalib-grey);
      font-size: 1em;
      font-weight: normal;
      border-radius: 4px 4px 0 0;
   }
   .form {
      background: white;
      min-height: 300px;
      text-align: left;
      padding: 15px;
      display: flex;
      flex-direction: column;
      row-gap: 15px;
      div.row {
         display: flex;
         flex-direction: column;
         row-gap: 2px;
      }
   }
   .buttons {
      border-radius: 4px;
      background: white;
      padding: 15px;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 10px;
   }
}
</style>
