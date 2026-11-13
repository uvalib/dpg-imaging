<template>
   <WaitSpinner v-if="searchStore.working" :overlay="true" message="Loading projects..." />
   <div class="row-wrap p-6 gap-5 relative">
      <SearchPanel />
      <template v-if="searchStore.working == false">
         <div class="text-2xl text-center grow pt-15" v-if="!searchStore.working && searchStore.projects.length == 0">
            No projects match your search criteria
         </div>
         <ul v-else class="row-wrap gap-6 grow">
            <li class="relative w-1 min-w-24/50 border border-brand-grey text-sm h-fit" v-for="(p,idx) in searchStore.projects" :key="`p${p.id}`" :class="{finished: p.finishedAt}">
               <div class="border-b border-brand-grey bg-brand-grey-200 column">
                  <div class="bg-brand-grey-100 p-2 border-b border-brand-grey row-between items-center">
                     <span class="row-left gap-2">
                        <label class="font-bold!">Date Due:</label><span>{{searchStore.dueDate(idx)}}</span>
                     </span>
                     <span class="row-right gap-2">
                        <span class="bg-brand-orange text-white px-2 font-bold!" v-if="isOverdue(idx) && !p.finishedAt">OVERDUE</span>
                        <UIcon v-if="searchStore.hasError(idx) && !p.finishedAt" name="i-lucide-alert-circle" 
                           class="size-7 bg-red-700 text-red-100 rounded-full" /> 
                     </span>
                     <span class="column" v-if="p.finishedAt">
                        <div class="row-right gap-1">
                           <label class="font-bold!">Finished:</label><span>{{p.finishedAt.split("T")[0]}}</span>
                        </div>
                        <div class="row-right gap-1" v-if="p.totalDuration">
                           <label class="font-bold!">Duration: </label>
                           <span>{{ p.totalDuration }} mins</span>
                        </div>
                     </span>
                  </div>
                  <router-link :to="`/projects/${p.id}`" class="border-none!">
                     <div class="column p-2">
                        <div class="row-between text-black!">
                           <span class="row-left gap-1">
                              <label class="font-bold!">Project:</label><span>{{p.id}}</span>
                           </span>
                           <span v-if="p.imageCount > 0" class="row-left gap-1">
                              <label class="font-bold!">Images:</label><span>{{p.imageCount}}</span>
                           </span>
                        </div>
                        <div class="text-black!">{{p.title}}</div>
                     </div>
                  </router-link>
               </div>
               <div class="row-between p-4 pb-0 text-md">
                  <dl class="m-0!">
                     <dt>Customer:</dt>
                     <dd class="mb-0!">{{ systemStore.getCustomerName(p.customerID) }} </dd>
                     <template v-if="p.agencyID > 0">
                        <dt>Agency:</dt>
                        <dd>{{ systemStore.getAgency(p.agencyID)}}</dd>
                     </template>
                     <dt>Call Number:</dt>
                     <dd>
                        <span v-if="p.callNumber">{{p.callNumber}}</span>
                        <span v-else class="na">N/A</span>
                     </dd>
                     <dt>Intended Use:</dt>
                     <dd>{{ p.intendedUse }}</dd>
                  </dl>
                  <dl>
                     <dt>Order:</dt>
                     <dd><a target="_blank" :href="`${systemStore.adminURL}/orders/${p.orderID}`">{{p.orderID}}</a></dd>
                     <dt>Unit:</dt>
                     <dd><a target="_blank" :href="`${systemStore.adminURL}/units/${p.unitID}`">{{p.unitID}}</a></dd>
                     <dt>Workflow:</dt>
                     <dd>{{p.workflow.name}}</dd>
                     <dt>Category:</dt>
                     <dd>{{p.category.name}}</dd>
                  </dl>
               </div>
               <div class="column p-2" v-if="p.specialInstructions">
                  <label class="font-bold!">Special Instructions:</label>
                  <div class="text-sm px-4">{{p.specialInstructions}}</div>
               </div>
               <div class="status" v-if="!p.finishedAt || p.finishedAt == ''">
                  <div class="row-between p-2 items-center gap-4">
                     <span v-if="searchStore.hasError(idx)" class="text-nowrap text-brand-red-B">{{searchStore.statusText(p.id)}}</span>
                     <span v-else class="text-nowrap">{{searchStore.statusText(p.id)}}</span>
                     <UProgress v-if="p.currentStep" :modelValue="searchStore.percentComplete(p.id)" size="lg"/>
                  </div>
                  <div class="row-between p-2">
                     <span class="row-left gap-2 items-center">
                        <UIcon name="i-lucide-user"/> 
                        <span v-if="!p.owner">Unassigned</span>
                        <span v-else>{{ownerInfo(p)}}</span>
                     </span>
                     <span class="row-right gap-2">
                        <UButton @click="deleteProjectClicked(p)" class="delete" color="error" v-if="userStore.isSupervisor || userStore.isAdmin" label="Delete"/>
                        <UButton v-if="canClaim(p)" @click="claimClicked(p.id)" color="secondary" label="Claim"/>
                        <AssignModal  v-if="canAssign" :projectID="p.id" @assigned="searchStore.getProjects()" />
                        <UButton  color="secondary" @click="viewClicked(p.id)" label="View"/>
                     </span>
                  </div>
               </div>
            </li>
         </ul>
      </template>
   </div>
</template>

<script setup>
import AssignModal from "@/components/AssignModal.vue"
import SearchPanel from "@/components/SearchPanel.vue"
import { useSearchStore } from "@/stores/search"
import { useSystemStore } from "@/stores/system"
import { useProjectStore } from "@/stores/project"
import { useUserStore } from "@/stores/user"
import { useRoute, useRouter } from 'vue-router'
import { onBeforeMount } from 'vue'
import { useConfirm } from "../composables/useConfirm"

const searchStore = useSearchStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const projectStore = useProjectStore()
const route = useRoute()
const router = useRouter()

onBeforeMount( () => {
   if ( route.query.workflow) {
      searchStore.search.workflow = parseInt(route.query.workflow,10)
   }
   if ( route.query.step) {
      searchStore.search.step = route.query.step
   }
   if ( route.query.owner) {
      searchStore.search.assignedTo = parseInt(route.query.owner,10)
   }
   if ( route.query.order) {
      searchStore.search.orderID = route.query.order
   }
   if ( route.query.unit) {
      searchStore.search.unitID = route.query.unit
   }
   if ( route.query.callnum) {
      searchStore.search.callNumber = route.query.callnum
   }
   if ( route.query.customer) {
      searchStore.search.customer = parseInt(route.query.customer,10)
   }
   if ( route.query.agency) {
      searchStore.search.agency = parseInt(route.query.agency,10)
   }
   if ( route.query.workstation) {
      searchStore.search.workstation = parseInt(route.query.workstation,10)
   }
   if ( route.query.filter) {
      searchStore.filter = route.query.filter
   }

   searchStore.getProjects()
})

const canClaim = ((p) => {
   if ( !p.owner ) return true
   if ( (userStore.isAdmin || userStore.isSupervisor ) && p.owner.computingID != userStore.computeID) return true
   return false
})

const claimClicked = ( async (projID) => {
   await projectStore.assignProject( projID, userStore.ID )
   searchStore.getProjects()
})

const viewClicked = ((projID) => {
   router.push(`/projects/${projID}`)
})

const canAssign = (() => {
   return (userStore.isAdmin || userStore.isSupervisor)
})

const ownerInfo = ((p) => {
   return `${p.owner.firstName} ${p.owner.lastName} (${p.owner.computingID})`
})

const isOverdue = ((projIdx) => {
   let due = new Date(searchStore.dueDate(projIdx))
   let now = new Date()
   return now > due
})

const deleteProjectClicked = (async (p) => {
   let note = `<span style='font-weight:bold'>Important</span>: any images associated with this project will be left<br/>in the processing directory for unit ${p.unitID}`
   let msg = `Delete project ${p.id}? This cannot be reversed.<br/>${note}`
   const resp = await useConfirm("Confirm Delete", msg, "Delete")
   if (resp) {
      await projectStore.deleteProject( p.id )
      window.location.reload() 
   } 
})
</script>

<style scoped lang="scss">
</style>
