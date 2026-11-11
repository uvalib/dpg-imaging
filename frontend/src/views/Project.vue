<template>
   <h2 class="mb-0!">Digitization Project #{{route.params.id}}</h2>
   <WaitSpinner v-if="projectStore.working" :overlay="true" message="Working..." />
   <div v-if="projectStore.hasDetail" class="column gap-1 items-center border-b border-brand-grey-100 pb-4">
      
         <a target="_blank" :href="metadataLink">{{detail.title}}</a>
      
      <div class="row-center gap-6">
         <div class="row-left gap-2">
            <label class="font-bold!">Customer:</label>
            <span class="data">{{ systemStore.getCustomerName(detail.customerID) }}</span>
         </div>
         <div v-if="detail.agencyID > 0" class="row-left gap-2">
            <label class="font-bold!">Agency:</label>
            <span class="data">{{ systemStore.getAgency(detail.agencyID) }}</span>
         </div>
         <div class="row-left gap-2">
            <label class="font-bold!">Intended Use:</label>
            <span class="data">{{detail.intendedUse}}</span>
         </div>
      </div>
      <div class="row-center gap-6">
         <div class="row-left gap-2">
            <label class="font-bold!">Unit:</label>
            <a target="_blank" :href="`${systemStore.adminURL}/units/${detail.unitID}`">{{detail.unitID}}</a>
         </div>
         <div class="row-left gap-2">
            <label class="font-bold!">Order:</label>
            <a target="_blank" :href="`${systemStore.adminURL}/orders/${detail.orderID}`">{{detail.orderID}}</a>
         </div>
      </div>

      <div class="row-between px-4 w-full">
         <UButton icon="i-lucide-arrow-left" label="Back to projects" @click="backClicked" size="sm" color="secondary"/>
         <span v-if="projectStore.working == false" class="row-left gap-2 bg-brand-blue-alt-300 px-2 py-0.5 border border-brand-blue-alt">
            <label class="font-bold!">Due:</label><span>{{projectStore.dueDate}}</span>
         </span>
      </div>
   </div>
   <div  v-if="projectStore.hasDetail" class="project-main">
      <div class="column gap-6">
         <ItemInfo />
         <Workflow />
         <History />
      </div>
      <div class="column gap-6">
         <Equipment v-if="projectStore.detail.workflow.name != 'Vendor'"/>
         <Notes />
      </div>
      
   </div>
</template>

<script setup>
import ItemInfo from "@/components/project/ItemInfo.vue"
import Workflow from "@/components/project/Workflow.vue"
import History from "@/components/project/History.vue"
import Notes from "@/components/project/Notes.vue"
import Equipment from "@/components/project/Equipment.vue"
import { useSystemStore } from "@/stores/system"
import { useProjectStore } from "@/stores/project"
import { useSearchStore } from "@/stores/search"
import { onMounted, onBeforeUnmount, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'

const systemStore = useSystemStore()
const projectStore = useProjectStore()
const searchStore = useSearchStore()
const route = useRoute()
const router = useRouter()

const { detail } = storeToRefs(projectStore)

const metadataLink = computed(() => {
   return `${systemStore.adminURL}/metadata/${detail.value.metadataID}`
})

onMounted( async () => {
   await projectStore.getProject(route.params.id)
   projectStore.pollProjectStatus()
})

onBeforeUnmount( async () => {
   projectStore.cancelStatusPolling()
})

const backClicked = (() => {
   router.push( searchStore.lastSearchURL )
})

</script>

<style scoped lang="scss">
.project-main {
   padding: 20px 40px;
   display: grid;
   grid-template-columns: repeat(2, 1fr);
   gap: 20px;
   align-items: start;
}
</style>
