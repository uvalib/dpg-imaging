<template>
   <div class="home">
      <h2>
         <span>Digitization Projects</span>
      </h2>
      <WaitSpinner v-if="searchStore.working" :overlay="true" message="Loading projects..." />
      <div class="toolbar">
         <div class="filter">
            <label for="me">
               <input id="me" type="radio" value="me" name="filter" v-model="activeFilter" @change="filterChanged">
               <span>Assigned to me <span class="count">({{searchStore.totals.me}})</span></span>
            </label>
            <label for="active">
               <input id="active" type="radio" value="active" name="filter" v-model="activeFilter" @change="filterChanged">
               <span>Active <span class="count">({{searchStore.totals.active}})</span></span>
            </label>
            <label for="errors">
               <input id="errors" type="radio" value="errors" name="filter" v-model="activeFilter" @change="filterChanged">
               <span>Problems <span class="count">({{searchStore.totals.errors}})</span></span>
            </label>
            <label for="unassigned">
               <input id="unassigned" type="radio" value="unassigned" name="filter" v-model="activeFilter" @change="filterChanged">
               <span>Unassigned <span class="count">({{searchStore.totals.unassigned}})</span></span>
            </label>
            <label for="finished">
               <input id="finished" type="radio" value="finished" name="filter" v-model="activeFilter" @change="filterChanged">
               <span>Finished <span class="count">({{searchStore.totals.finished}})</span></span>
            </label>
         </div>
         <div class="page-ctl" v-if="!searchStore.working && searchStore.projects.length>0">
            <DPGPagination :currPage="searchStore.currPage" :pageSize="searchStore.pageSize" :totalPages="searchStore.totalPages"
               @next="nextClicked" @prior="priorClicked" @first="firstClicked" @last="lastClicked"
               @jump="pageJumpClicked"
            />
         </div>
      </div>

      <div class="project-board">
         <SearchPanel />
         <div class="none" v-if="!searchStore.working && searchStore.projects.length == 0">
            No projects match your search criteria
         </div>
         <ul v-else class="projects">
            <li class="card" v-for="(p,idx) in searchStore.projects" :key="`p${p.id}`">
               <div class="top">
                  <div class="due">
                     <span>
                        <label>Date Due:</label><span>{{searchStore.dueDate(idx)}}</span>
                     </span>
                     <span>
                        <span class="status-msg overdue" v-if="isOverdue(idx) && !p.finishedAt">OVERDUE</span>
                        <i v-if="searchStore.hasError(idx)" class="error-icon fas fa-exclamation-triangle"></i>
                     </span>
                     <span v-if="p.finishedAt">
                        <label>Finished:</label><span>{{p.finishedAt.split("T")[0]}}</span>
                        <div class="time" v-if="p.totalDuration">
                           <label>Duration:</label>
                           <span>{{ p.totalDuration }} mins</span>
                        </div>
                     </span>
                  </div>
                  <router-link :to="`/projects/${p.id}`">
                     <div class="title">
                        <div class="project-id"><label>Project:</label><span>{{p.id}}</span></div>
                        <div>{{p.unit.metadata.title}}</div>
                     </div>
                  </router-link>
               </div>
               <div class="data">
                  <dl>
                     <dt>Customer:</dt>
                     <dd>{{p.unit.order.customer.lastName}}, {{p.unit.order.customer.firstName}} </dd>
                     <template v-if="p.unit.order.agency.id > 0">
                        <dt>Agency:</dt>
                        <dd>{{p.unit.order.agency.name}}</dd>
                     </template>
                     <dt>Call Number:</dt>
                     <dd>
                        <span v-if="p.unit.metadata.callNumber">{{p.unit.metadata.callNumber}}</span>
                        <span v-else class="na">N/A</span>
                     </dd>
                     <dt>Intended Use:</dt>
                     <dd>{{p.unit.intendedUse.description}}</dd>
                  </dl>
                  <dl class="right">
                     <dt>Order:</dt>
                     <dd><a target="_blank" :href="`${systemStore.adminURL}/orders/${p.unit.order.id}`">{{p.unit.order.id}}</a></dd>
                     <dt>Unit:</dt>
                     <dd><a target="_blank" :href="`${systemStore.adminURL}/units/${p.unit.id}`">{{p.unit.id}}</a></dd>
                     <dt>Workflow:</dt>
                     <dd>{{p.workflow.name}}</dd>
                     <dt>Category:</dt>
                     <dd>{{p.category.name}}</dd>
                  </dl>
               </div>
               <div class="special-instructions" v-if="p.unit.specialInstructions">
                  <label>Special Instructions:</label>
                  <p>{{p.unit.specialInstructions}}</p>
               </div>
               <div class="status" v-if="!p.finishedAt || p.finishedAt == ''">
                  <div class="progress-panel">
                     <span :class="{error: searchStore.hasError(idx)}">{{searchStore.statusText(p.id)}}</span>
                     <div class="progress-bar">
                        <div class="percentage" :style="{width: searchStore.percentComplete(p.id) }"></div>
                     </div>
                  </div>
                  <div class="owner-panel">
                     <span class="assignment">
                        <i class="user fas fa-user"></i>
                        <span v-if="!p.owner" class="unassigned">Unassigned</span>
                        <span v-else class="assigned">{{ownerInfo(p)}}</span>
                     </span>
                     <span class="owner-buttons">
                        <DPGButton v-if="canClaim(p)" @click="claimClicked(p.id)" class="p-button-secondary right-pad" label="Claim"/>
                        <AssignModal  v-if="canAssign" :projectID="p.id" @assigned="searchStore.getProjects()" />
                        <DPGButton  class="view p-button-secondary" @click="viewClicked(p.id)" label="View"/>
                     </span>
                  </div>
               </div>
            </li>
         </ul>
      </div>
   </div>
</template>

<script setup>
import DPGPagination from "../components/DPGPagination.vue"
import AssignModal from "@/components/AssignModal.vue"
import SearchPanel from "@/components/SearchPanel.vue"
import { useSearchStore } from "@/stores/search"
import { useSystemStore } from "@/stores/system"
import { useProjectStore } from "@/stores/project"
import { useUserStore } from "@/stores/user"
import { useRoute, useRouter } from 'vue-router'
import { onBeforeMount, ref } from 'vue'

const searchStore = useSearchStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const projectStore = useProjectStore()
const route = useRoute()
const router = useRouter()

const activeFilter = ref("active")

onBeforeMount( () => {
   if ( route.query.workflow) {
      searchStore.search.workflow = parseInt(route.query.workflow,10)
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
      searchStore.search.customer = route.query.customer
   }
   if ( route.query.agency) {
      searchStore.search.agency = parseInt(route.query.agency,10)
   }
   if ( route.query.workstation) {
      searchStore.search.workstation = parseInt(route.query.workstation,10)
   }
   if ( route.query.filter) {
      activeFilter.value = route.query.filter
      searchStore.filter = route.query.filter
   }

   searchStore.getProjects()
})

const filterChanged = ( async () => {
   let query = Object.assign({}, route.query)
   query.filter = activeFilter.value
   searchStore.changeFilter(activeFilter.value)
   await router.push({query})
   searchStore.lastSearchURL = route.fullPath

   searchStore.getProjects()
})

const canClaim = ((p) => {
   if ( !p.owner ) return true
   if ( (userStore.isAdmin || userStore.isSupervisor ) && p.owner.computingID != userStore.computeID) return true
   return false
})

const claimClicked = ( async (projID) => {
   await projectStore.assignProject( {projectID: projID, ownerID: userStore.ID} )
   searchStore.getProjects()
})

const viewClicked = ((projID) => {
   router.push(`/projects/${projID}`)
})

const canAssign = (() => {
   return (userStore.isAdmin || userStore.isSupervisor)
})

const nextClicked = (() => {
   searchStore.setCurrentPage(searchStore.currPage+1 )
})

const priorClicked = (() => {
   searchStore.setCurrentPage(searchStore.currPage-1 )
})
const firstClicked = (() => {
   searchStore.setCurrentPage( 1 )
})

const lastClicked = (() => {
   searchStore.setCurrentPage(searchStore.totalPages )
})

const pageJumpClicked = ((p) => {
   searchStore.setCurrentPage( p )
})

const ownerInfo = ((p) => {
   return `${p.owner.firstName} ${p.owner.lastName} (${p.owner.computingID})`
})

const isOverdue = ((projIdx) => {
   let due = new Date(searchStore.dueDate(projIdx))
   let now = new Date()
   return now > due
})
</script>

<style scoped lang="scss">
.home {
   position: relative;
   padding: 0;
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 20px;
   }
   .none {
      font-size: 1.25em;
      text-align: center;
      margin-top: 50px;
      flex-grow: 1;
   }
   .toolbar {
      padding: 5px 10px;
      text-align: right;
      background: var(--uvalib-grey-lightest);
      border-top: 1px solid var(--uvalib-grey);
      border-bottom: 1px solid var(--uvalib-grey);
      display: flex;
      flex-flow: row;
      justify-content: flex-start;
      align-content: center;
      .filter {
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: baseline;

         label {
            display: flex;
            flex-flow: row nowrap;
            align-items: center;
            margin: 0;
            padding: 0;
            margin-right: 25px;
            cursor: pointer;
            span {
               display: inline-block;
               font-size: 0.9em;
               position: relative;
               top: 2px;
            }
            .count {
               display: inline-block;
               top: 0;
               color: var(--uvalib-grey);
            }
            &:hover {
               text-decoration: underline;
            }
         }
         input {
            cursor: pointer;
            margin-right: 8px;
            display: inline-block;
            width: 15px;
            height: 15px;
         }
      }

      .page-ctl {
         margin-left: auto;
         display: inline-block;
      }
   }
   .project-board {
      display: flex;
      flex-flow: row nowrap;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: flex-start;
      padding: 25px 25px 0 25px;
   }
   .projects {
      list-style: none;
      margin: 0;
      padding: 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: center;
      flex-grow: 1;
      background: transparent;

      .card {
         flex: 0 1 calc(25% - 1em);
         border: 1px solid var(--uvalib-grey);
         padding: 0;
         margin: 0px 10px 20px 10px;
         position: relative;
         text-align: left;
         box-sizing: border-box;
         min-width: 48%;
         color: var(--uvalib-text);
         font-size: 0.9em;
         box-shadow: rgba(0, 0, 0, 0.14) 0px 2px 2px 0px;
         background: white;
         padding-bottom: 110px;

         .top {
            border-bottom: 1px solid var(--uvalib-grey);
            color: var(--uvalib-text);
            .title {
               padding: 10px;
               color: var(--uvalib-text) !important;
               background:var(--uvalib-grey-lightest);
               .project-id {
                  label { font-weight: bold; margin-right: 5px;}
                  margin-bottom: 10px;
                  padding-bottom: 10px;
                  border-bottom: 1px solid var(--uvalib-grey-light);
               }
            }
            .due {
               padding: 5px 5px 5px 10px;
               border-bottom: 1px solid var(--uvalib-grey);
               background: var(--uvalib-grey-light);
               display: flex;
               flex-flow: row nowrap;
               justify-content: space-between;
               align-items: flex-start;
               .error-icon {
                  display: inline-block;
                  color: var(--uvalib-red-emergency);
                  margin-left: 5px;
                  font-size: 1.15em;
               }
               label {
                  font-weight: bold;
                  margin-right: 5px;
               }
               .status-msg {
                  background: white;
                  padding: 2px 10px;
                  border: 1px solid var(--uvalib-grey);
               }
               .overdue {
                  font-weight: bold;
                  background: var(--uvalib-brand-orange);
                  color: white;
                  border: 0;
                  border-radius: 5px;
               }
            }
         }
         .special-instructions{
            margin: 0 30px;
            font-size: 0.9em;
            label {
               display: block;
               font-weight: bold !important;
            }
            p {
               padding:0;
               margin: 5px 0;
            }
         }
         .data {
            padding: 10px;
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-start;
            align-items: flex-start;
            font-size: 0.9em;
            dl.right {
               margin-left: 50px;
            }
            dl {
               margin-left: 25px;
               display: inline-grid;
               grid-template-columns: max-content 2fr;
               grid-column-gap: 5px;
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
                  .na {
                     color: #999;
                  }
               }
            }
         }
         .status {
            padding: 10px;
            border-top: 1px solid var(--uvalib-grey-lightest);
            position: absolute;
            width: 100%;
            box-sizing: border-box;
            bottom: 0;

            .owner-buttons {
               display: flex;
               flex-flow: row nowrap;
               .dpg-button {
                  margin-right: 10px;
               }
               .view {
                  margin-right: 0;
                  margin-left: 10px;
               }
            }

            .progress-panel {
               margin: 5px 0 15px 0;
               display: flex;
               flex-flow: row nowrap;
               justify-content: space-between;
               align-items: center;
               .error {
                  color: var(--uvalib-red-darker);
                  font-weight: bold;
               }

               .progress-bar {
                  border: 1px solid var(--uvalib-grey-light);
                  background: white;
                  height: 20px;
                  margin-left: 15px;
                  flex-grow: 1;
                  .percentage {
                     background: var(--uvalib-green);
                     height: 100%;
                  }
               }
            }
            .owner-panel {
               display: flex;
               flex-flow: row nowrap;
               justify-content: space-between;

               .assignment {
                  .user {
                     margin-right: 10px;
                  }
                  .unassigned {
                     font-weight: 100;
                     color: #999;
                  }
               }
            }
         }
      }
   }
}
</style>
