<template>
   <div class="project">
      <h2>
         <span>Digitization Project #{{route.params.id}}</span>
         <span v-if="projectStore.working == false" class="due">
            <label>Due:</label><span>{{projectStore.dueDate}}</span>
         </span>
      </h2>
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
         <div class="row right-pad">
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
            <i class="fas fa-angle-double-left back-button"></i>
            <span class="text-button" @click="backClicked">Back to projects</span>
         </div>
      </div>
      <div  v-if="projectStore.hasDetail" class="project-main">
         <ItemInfo />
         <Equipment />
         <div class="double">
            <Workflow />
            <History />
         </div>
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
   padding: 25px;

   .due {
      position: absolute;
      right: 0;
      color: var(--uvalib-text);
      font-size: 16px;
      font-weight: 500;
      background: var(--uvalib-blue-alt-light);
      border: 1px solid var(--uvalib-blue-alt);
      padding: 5px 15px;
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
         text-align: left;
         .back-button {
            color: var(--color-link);
         }
         .text-button {
            font-weight: normal;
            text-decoration: none;
            color: var(--color-link);
            display: inline-block;
            margin-left: 5px;
            cursor: pointer;
            &:hover {
               text-decoration: underline ;
            }
         }
      }
   }
   .project-main {
      display: flex;
      flex-flow: row wrap;
      justify-content: center;
      align-items: flex-start;

      .double {
         width: 46%;
         min-width: 600px;
      }
   }
}
</style>
