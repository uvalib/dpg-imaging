<template>
   <div class="home">
      <h2>
         <span>Digitization Projects</span>
      </h2>
      <WaitSpinner v-if="working" :overlay="true" message="Loading projects..." />
      <div class="projects-content">
         <div class="toolbar">
            <div class="filter">
               <label for="me">
                  <input id="me" type="radio" value="me" name="filter" v-model="filter" @change="filterChanged">
                  <span>Assigned to me</span>
               </label>
               <label for="active">
                  <input id="active" type="radio" value="active" name="filter" v-model="filter" @change="filterChanged">
                  <span>Active</span>
               </label>
               <label for="unassigned">
                  <input id="unassigned" type="radio" value="unassigned" name="filter" v-model="filter" @change="filterChanged">
                  <span>Unassigned</span>
               </label>
               <label for="finished">
                  <input id="finished" type="radio" value="finished" name="filter" v-model="filter" @change="filterChanged">
                  <span>Finished</span>
               </label>
            </div>
            <div class="page-ctl" v-if="!working && projects.length>0">
               <DPGPagination :currPage="currPage" :pageSize="pageSize" :totalPages="totalPages"
                  @next="nextClicked" @prior="priorClicked" @first="firstClicked" @last="lastClicked"
                  @jump="pageJumpClicked"
               />
            </div>
         </div>
         <div class="project-board">
            <SearchPanel />
            <div class="none" v-if="!working && projects.length == 0">
               No projects match your search criteria
            </div>
            <ul v-else class="projects">
               <li class="card" v-for="p in projects" :key="`p${p.id}`">
                  <div class="top">
                     <div class="due">
                        <span>
                           <label>Date Due:</label><span>{{p.dueOn.split("T")[0]}}</span>
                        </span>
                        <span class="status-msg overdue" v-if="isOverdue(p) && !p.finishedAt">OVERDUE</span>
                        <span v-if="p.finishedAt"><label>Finished:</label><span>{{p.finishedAt.split("T")[0]}}</span></span>
                     </div>
                     <div class="title">
                        <router-link @click="selectProject(p.id)" :to="`/projects/${p.id}`">{{p.unit.metadata.title}}</router-link>
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
                        <dd><a target="_blank" :href="`${adminURL}/orders/${p.unit.order.id}`">{{p.unit.order.id}}</a></dd>
                        <dt>Unit:</dt>
                        <dd><a target="_blank" :href="`${adminURL}/units/${p.unit.id}`">{{p.unit.id}}</a></dd>
                        <dt>Workflow:</dt>
                        <dd>{{p.workflow.name}}</dd>
                        <dt>Category:</dt>
                        <dd>{{p.category.name}}</dd>
                     </dl>
                  </div>
                  <div class="status" v-if="!p.finishedAt || p.finishedAt == ''">
                     <div class="progress-panel">
                        <span>{{statusText(p.id)}}</span>
                        <div class="progress-bar">
                           <div class="percentage" :style="{width: percentComplete(p.id) }"></div>
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
                           <assign-modal  v-if="canAssign" :projectID="p.id" @assign="assignClicked"/>
                        </span>
                     </div>
                  </div>
               </li>
            </ul>
         </div>
      </div>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
import { mapFields } from 'vuex-map-fields'
import AssignModal from "@/components/AssignModal"
import SearchPanel from "@/components/SearchPanel"
export default {
   name: "home",
   components: {
      AssignModal, SearchPanel
   },
   computed: {
      ...mapState({
         working : state => state.projects.working,
         projects : state => state.projects.projects,
         currPage : state => state.projects.currPage,
         pageSize : state => state.projects.pageSize,
         jwt : state => state.user.jwt,
         userComputingID : state => state.user.computeID,
         userID : state => state.user.ID,
         adminURL: state => state.adminURL,
      }),
      ...mapGetters({
         totalPages: 'projects/totalPages',
         isAdmin: 'user/isAdmin',
         isSupervisor: 'user/isSupervisor',
         statusText: 'projects/statusText',
         percentComplete: 'projects/percentComplete',
         isFinished: 'projects/isFinished'
      }),
      ...mapFields({
        filter: 'projects.filter',
      })
   },
   methods: {
      filterChanged() {
         this.$store.dispatch("projects/getProjects")
      },
      selectProject(id) {
         this.$store.commit("projects/selectProject", id)
      },
      canClaim(p) {
         if ( !p.owner ) return true
         if ( (this.isAdmin || this.isSupervisor ) && p.owner.computingID != this.userComputingID) return true
         return false
      },
      claimClicked(projID) {
         this.$store.dispatch("projects/assignProject", {projectID: projID, ownerID: this.userID} )
      },
      canAssign() {
         return (this.isAdmin || this.isSupervisor)
      },
      assignClicked( info ) {
         this.$store.dispatch("projects/assignProject", {projectID: info.projectID, ownerID: info.ownerID} )
      },
      nextClicked() {
         this.$store.dispatch("projects/setCurrentPage", this.currPage+1 )
      },
      priorClicked() {
         this.$store.dispatch("projects/setCurrentPage", this.currPage-1 )
      },
      firstClicked() {
         this.$store.dispatch("projects/setCurrentPage", 1 )
      },
      lastClicked() {
         this.$store.dispatch("projects/setCurrentPage", this.totalPages )
      },
      pageJumpClicked(p) {
         this.$store.dispatch("projects/setCurrentPage", p )
      },
      ownerInfo(p) {
         return `${p.owner.firstName} ${p.owner.lastName} (${p.owner.computingID})`
      },
      isOverdue(p) {
         let due =  new Date(p.dueOn)
         let now = new Date()
         return now > due
      }
   },
   created() {
      if (this.jwt != "" && this.projects.length <= 1) {
         this.$store.dispatch("projects/getProjects")
      }
   },
};
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
