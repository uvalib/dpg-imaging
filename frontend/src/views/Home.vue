<template>
   <div class="home">
      <h2>
         <span>Digitization Projects</span>
      </h2>
      <WaitSpinner v-if="projectStore.working" :overlay="true" message="Loading projects..." />
      <div class="projects-content">
         <div class="toolbar">
            <div class="filter">
               <label for="me">
                  <input id="me" type="radio" value="me" name="filter" v-model="projectStore.filter" @change="filterChanged">
                  <span>Assigned to me <span class="count">({{projectStore.totals.me}})</span></span>
               </label>
               <label for="active">
                  <input id="active" type="radio" value="active" name="filter" v-model="projectStore.filter" @change="filterChanged">
                  <span>Active <span class="count">({{projectStore.totals.active}})</span></span>
               </label>
               <label for="unassigned">
                  <input id="unassigned" type="radio" value="unassigned" name="filter" v-model="projectStore.filter" @change="filterChanged">
                  <span>Unassigned <span class="count">({{projectStore.totals.unassigned}})</span></span>
               </label>
               <label for="finished">
                  <input id="finished" type="radio" value="finished" name="filter" v-model="projectStore.filter" @change="filterChanged">
                  <span>Finished <span class="count">({{projectStore.totals.finished}})</span></span>
               </label>
            </div>
            <div class="page-ctl" v-if="!projectStore.working && projectStore.projects.length>0">
               <DPGPagination :currPage="projectStore.currPage" :pageSize="projectStore.pageSize" :totalPages="projectStore.totalPages"
                  @next="nextClicked" @prior="priorClicked" @first="firstClicked" @last="lastClicked"
                  @jump="pageJumpClicked"
               />
            </div>
         </div>
         <div class="project-board">
            <SearchPanel />
            <div class="none" v-if="!projectStore.working && projectStore.projects.length == 0">
               No projects match your search criteria
            </div>
            <ul v-else class="projects">
               <li class="card" v-for="p in projectStore.projects" :key="`p${p.id}`">
                  <div class="top">
                     <div class="due">
                        <span>
                           <label>Date Due:</label><span>{{p.dueOn.split("T")[0]}}</span>
                        </span>
                        <span class="status-msg overdue" v-if="isOverdue(p) && !p.finishedAt">OVERDUE</span>
                        <span v-if="p.finishedAt"><label>Finished:</label><span>{{p.finishedAt.split("T")[0]}}</span></span>
                     </div>
                     <div class="title">
                        <router-link @click="projectStore.selectProject(p.id)" :to="`/projects/${p.id}`">{{p.unit.metadata.title}}</router-link>
                     </div>
                  </div>
                  <div class="data">
                     <dl>
                        <dt>Customer:</dt>
                        <dd>{{p.unit.order.customer.firstName}} {{p.unit.order.customer.lastName}}</dd>
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
                  <div class="status" v-if="!p.finishedAt || p.finishedAt == ''">
                     <div class="progress-panel">
                        <span>{{projectStore.statusText(p.id)}}</span>
                        <div class="progress-bar">
                           <div class="percentage" :style="{width: projectStore.percentComplete(p.id) }"></div>
                        </div>
                     </div>
                     <div class="owner-panel">
                        <span class="assignment">
                           <i class="user fas fa-user"></i>
                           <span v-if="!p.owner" class="unassigned">Unassigned</span>
                           <span v-else class="assigned">{{ownerInfo(p)}}</span>
                        </span>
                        <span class="owner-buttons">
                           <DPGButton v-if="canClaim(p)" @clicked="claimClicked(p.id)">Claim</DPGButton>
                           <AssignModal  v-if="canAssign" :projectID="p.id" @assign="assignClicked"/>
                        </span>
                     </div>
                  </div>
               </li>
            </ul>
         </div>
      </div>
   </div>
</template>

<script setup>
import AssignModal from "@/components/AssignModal.vue"
import SearchPanel from "@/components/SearchPanel.vue"
import {useProjectStore} from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import { useRoute } from 'vue-router'
import { onMounted } from 'vue'

const projectStore = useProjectStore()
const systemStore = useSystemStore()
const userStore = useUserStore()
const route = useRoute()

function filterChanged() {
   projectStore.getProjects()
}
function canClaim(p) {
   if ( !p.owner ) return true
   if ( (userStore.isAdmin || userStore.isSupervisor ) && p.owner.computingID != userStore.computeID) return true
   return false
}
function claimClicked(projID) {
   projectStore.assignProject( {projectID: projID, ownerID: userStore.ID} )
}
function canAssign() {
   return (userStore.isAdmin || userStore.isSupervisor)
}
function assignClicked( info ) {
   projectStore.assignProject( {projectID: info.projectID, ownerID: info.ownerID} )
}
function nextClicked() {
   projectStore.setCurrentPage(projectStore.currPage+1 )
}
function priorClicked() {
   projectStore.setCurrentPage(projectStore.currPage-1 )
}
function firstClicked() {
   projectStore.setCurrentPage( 1 )
}
function lastClicked() {
   projectStore.setCurrentPage(projectStore.totalPages )
}
function pageJumpClicked(p) {
   projectStore.setCurrentPage( p )
}
function ownerInfo(p) {
   return `${p.owner.firstName} ${p.owner.lastName} (${p.owner.computingID})`
}
function isOverdue(p) {
   let due =  new Date(p.dueOn)
   let now = new Date()
   return now > due
}

onMounted( async () => {
   if ( route.query.order ) {
      projectStore.orderID = route.query.order
      projectStore.getProjects()
   } else if ( route.query.unit ) {
      projectStore.unitID = route.query.unit
      projectStore.getProjects()
   } else {
      if (userStore.jwt != "" && projectStore.projects.length <= 1) {
         projectStore.getProjects()
      }
   }
})
</script>

<style scoped lang="scss">
.home {
   position: relative;
   padding: 25px;
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 50px;
      .old-units {
         font-size: 14px;
         position: absolute;
         top: 5px;
         left: 8px;
      }
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
      margin-bottom: 20px;
      border-top: 1px solid var(--uvalib-grey-light);
      border-bottom: 1px solid var(--uvalib-grey-light);
      display: flex;
      flex-flow: row;
      justify-content: flex-start;
      align-content: center;
      .filter {
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-start;
         align-items: center;

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

         .top {
            background: var(--uvalib-grey-lightest);
            border-bottom: 1px solid var(--uvalib-grey);
            color: var(--uvalib-text);
            .title {
               padding: 10px;
               a {
                  color: var(--uvalib-text) !important;
               }
            }
            .due {
               padding: 5px 5px 5px 10px;
               border-bottom: 1px solid var(--uvalib-grey);
               background: var(--uvalib-grey-light);
               display: flex;
               flex-flow: row nowrap;
               justify-content: space-between;
               align-items: center;
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
                  background: firebrick;
                  color: white;
                  border: 0;
               }
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

            .owner-buttons {
               display: flex;
               flex-flow: row nowrap;
               .dpg-button {
                  margin-right: 10px;
               }
            }

            .progress-panel {
               margin: 5px 0 15px 0;
               display: flex;
               flex-flow: row nowrap;
               justify-content: space-between;
               align-items: center;

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
