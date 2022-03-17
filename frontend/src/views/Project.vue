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
            <div class="info-block">
               <h4>Item Information</h4>
               <ItemInfo />
            </div>
            <div class="info-block">
               <h4>Equipment</h4>
               <Equipment />
            </div>
            <div class="double">
               <div class="info-block nested">
                  <h4>Workflow</h4>
                  <Workflow />
               </div>
               <div class="info-block nested">
                  <h4>History</h4>
                  <History />
               </div>
            </div>

            <div class="info-block">
               <h4 class="notes">
                  <span>Notes</span>
                  <NoteModal id="note-modal" />
               </h4>
               <Notes />
            </div>
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
import NoteModal from '@/components/project/NoteModal.vue'
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
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 25px;
      position: relative;
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
         .icon {
            display: inline-block;
            margin-left: 10px;
         }
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
      .info-block.nested {
         width: 100%;
         box-sizing: border-box;
         margin: 15px 0;
      }
      .info-block {
         width: 46%;
         min-width: 600px;
         border: 1px solid var(--uvalib-grey-light);
         margin: 15px;
         display: inline-block;
         min-height: 100px;
         box-shadow: 0 1px 3px rgba(0,0,0,.06), 0 1px 2px rgba(0,0,0,.12);
         text-align: left;
         .na {
            color: #999;
         }
         h4 {
            text-align: center;
            color: var(--uvalib-text);
            font-size: 1em;
            margin: 0;
            padding: 5px;
            background: var(--uvalib-grey-lightest);
            border-bottom: 1px solid var(--uvalib-grey-light);
         }
         h4.notes {
            display: flex;
            flex-flow: row nowrap;
            justify-content: center;
            align-items: center;
            span {
               flex-grow: 1;
            }
         }
         dl {
            margin: 10px 30px 0 30px;
            display: inline-grid;
            grid-template-columns: max-content 2fr;
            grid-column-gap: 10px;
            font-size: 0.9em;
            text-align: left;
            box-sizing: border-box;

            dt {
               font-weight: bold;
               text-align: right;
            }
            dd {
               margin: 0 0 10px 0;
               word-break: break-word;
               -webkit-hyphens: auto;
               -moz-hyphens: auto;
               hyphens: auto;
            }
         }
      }
   }
}
</style>
