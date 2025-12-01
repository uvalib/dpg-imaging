<template>
   <div class="project">
      <h2>Digitization Project #{{route.params.id}}</h2>
      <WaitSpinner v-if="projectStore.working" :overlay="true" message="Working..." />
      <div v-if="projectStore.hasDetail" class="project-head">
         <h3>
            <a target="_blank" :href="metadataLink">{{detail.unit.metadata.title}}</a>
         </h3>
         <div class="row">
            <div>
               <label>Customer:</label>
               <span class="data">{{detail.unit.order.customer.firstName}} {{detail.unit.order.customer.lastName}}</span>
            </div>
            <div v-if="detail.unit.order.agency.id > 0">
               <label>Agency:</label>
               <span class="data">{{detail.unit.order.agency.name}}</span>
            </div>
            <div>
               <label>Intended Use:</label>
               <span class="data">{{detail.unit.intendedUse.description}}</span>
            </div>
         </div>
         <div class="row">
            <div>
               <label>Unit:</label>
               <a target="_blank" :href="`${systemStore.adminURL}/units/${detail.unit.id}`">{{detail.unit.id}}</a>
            </div>
            <div>
               <label>Order:</label>
               <a target="_blank" :href="`${systemStore.adminURL}/orders/${detail.unit.orderID}`">{{detail.unit.orderID}}</a>
            </div>
         </div>

         <div class="back">
            <DPGButton icon="pi pi-angle-double-left" text label="Back to projects" @click="backClicked" size="small" severity="secondary"/>
            <span v-if="projectStore.working == false" class="due">
               <label>Due:</label><span>{{projectStore.dueDate}}</span>
            </span>
         </div>
      </div>
      <div  v-if="projectStore.hasDetail" class="project-main">
         <ItemInfo />
         <Equipment v-if="projectStore.detail.workflow.name != 'Vendor'"/>
         <Workflow />
         <Notes />
         <History />
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
   return `${systemStore.adminURL}/metadata/${detail.value.unit.metadata.id}`
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
.project {
   position: relative;
   padding: 0;

   .due {
         color: var(--uvalib-text);
         font-size: 16px;
         font-weight: 500;
         background: var(--uvalib-blue-alt-light);
         border: 1px solid var(--uvalib-blue-alt);
         padding: 5px 15px;
         margin-left: auto;
      }

   label {
      font-weight: bold;
      margin-right: 5px;
   }

   .project-head {
      color: var(--uvalib-text);
      padding-bottom: 15px;
      border-bottom: 1px solid var(--uvalib-grey-light);
      position: relative;
      margin-bottom: 10px;
      h3  {
         max-width: 90%;
         text-align: center;
         font-weight: 500;
         font-size: 1.25em;
         margin: 5px auto 10px auto;
      }
      h4 {
         font-size: 0.9em;
         display: flex;
         flex-flow: row nowrap;
         justify-content: center;
         margin: 5px 0;

         label {
            margin-right: 10px;
            width: 100px;
            display: inline-block;
            text-align: right;
         }
         .data {
            font-weight: 500;
         }
         div {
            margin: 5px 0;
         }
      }
      .row  {
         display: flex;
         flex-flow: row nowrap;
         justify-content: center;
         padding: 5px 0 5px 0;
         label {
            margin-left: 15px;
         }
      }
      .back {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         align-items: center;
         padding: 0 10px;
      }
   }
   .project-main {
      padding: 20px 40px;
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 25px;
      align-items: start;
   }
}
</style>
