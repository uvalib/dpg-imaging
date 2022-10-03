<template>
   <div class="project">
      <h2>
         <span>Digitization Project #{{route.params.id}}</span>
         <span v-if="projectStore.working == false && currProject.dueOn" class="due">
            <label>Due:</label><span>{{currProject.dueOn.split("T")[0]}}</span>
         </span>
      </h2>
      <WaitSpinner v-if="projectStore.working" :overlay="true" message="Working..." />
      <template v-if="projectStore.selectedProjectIdx >=0">
         <div class="project-head">
            <h3>
               <a target="_blank" :href="metadataLink">{{currProject.unit.metadata.title}}</a>
            </h3>
            <h4 class="proj-data">
               <div class="column right-pad">
                  <div>
                     <label>Unit:</label>
                     <a target="_blank" :href="`${systemStore.adminURL}/units/${currProject.unit.id}`">{{currProject.unit.id}}</a>
                  </div>
                  <div>
                     <label>Order:</label>
                     <a target="_blank" :href="`${systemStore.adminURL}/orders/${currProject.unit.orderID}`">{{currProject.unit.orderID}}</a>
                  </div>
               </div>
               <div class="column">
                  <div>
                     <label>Customer:</label>
                     <a target="_blank" :href="`${systemStore.adminURL}/customers/${currProject.unit.order.customer.id}`">
                        {{currProject.unit.order.customer.firstName}} {{currProject.unit.order.customer.lastName}}
                     </a>
                  </div>
                  <div>
                     <label>Intended Use:</label>
                     <span class="data">{{currProject.unit.intendedUse.description}}</span>
                  </div>
               </div>
            </h4>
            <span class="back">
               <i class="fas fa-angle-double-left back-button"></i>
               <router-link to="/">Back to projects</router-link>
            </span>
         </div>
         <div class="project-main">
            <ItemInfo />
            <Equipment />
            <div class="double">
               <Workflow />
               <History />
            </div>
            <Notes />
         </div>
      </template>
   </div>
</template>

<script setup>
import ItemInfo from "@/components/project/ItemInfo.vue"
import Workflow from "@/components/project/Workflow.vue"
import History from "@/components/project/History.vue"
import Notes from "@/components/project/Notes.vue"
import Equipment from "@/components/project/Equipment.vue"
import {useSystemStore} from "@/stores/system"
import {useProjectStore} from "@/stores/project"
import { onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'

const systemStore = useSystemStore()
const projectStore = useProjectStore()
const route = useRoute()

const { currProject } = storeToRefs(projectStore)

const metadataLink = computed(() => {
   let mdType = "sirsi_metadata"
   if (currProject.value.unit.metadata.type == "XmlMetadata") {
      mdType = "xml_metadata"
   }
   return `${systemStore.adminURL}/${mdType}/${currProject.value.unit.metadata.id}`
})

onMounted( async () => {
   if (projectStore.selectedProjectIdx == -1) {
      await projectStore.getProject(route.params.id)
   }
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
      margin-bottom: 20px;
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
         .column {
            text-align: left;
         }
         .column.right-pad {
             margin-right: 25px;
         }
         label {
            margin-right: 10px;
            width: 100px;
            display: inline-block;
            text-align: right;
         }
         .data {
            font-weight: 500;
         }
         margin: 5px 0;
         div {
            margin: 5px 0;
         }
      }
      .back {
         position: absolute;
         left: 0px;
         bottom: 10px;
         a {
            font-weight: normal;
            text-decoration: none;
            color: var(--uvalib-text) !important;
            display: inline-block;
            margin-left: 5px;
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
