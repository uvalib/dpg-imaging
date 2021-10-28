<template>
   <div class="home">
      <h2>
         <span>Digitization Projects</span>
         <router-link class="old-units" to="/units">Old Units Page</router-link>
      </h2>
      <WaitSpinner v-if="loading" :overlay="true" message="Loading projects..." />
      <ul v-else class="projects">
          <li class="card" v-for="p in projects" :key="`p${p.id}`">
             <div class="top">
               <div class="due">
                  <span>
                     <label>Date Due:</label><span>{{p.dueOn.split("T")[0]}}</span>
                  </span>
                  <span class="status-msg overdue" v-if="isOverdue(p)">OVERDUE</span>
               </div>
               <div class="title">{{p.unit.metadata.title}}</div>
            </div>
             <div class="data">

                <dl>
                  <dt>Customer:</dt>
                  <dd>{{p.unit.order.customer.firstName}} {{p.unit.order.customer.lastName}}</dd>
                  <dt>Call Number:</dt>
                  <dd>{{p.unit.metadata.callNumber}}</dd>
                  <dt>ViU Number:</dt>
                  <dd>
                     <span v-if="p.viuNumber">{{p.viuNumber}}</span>
                     <span v-else class="na">N/A</span>
                  </dd>
                  <dt>Intended Use:</dt>
                  <dd>{{p.unit.intendedUse.description}}</dd>
                </dl>
                <dl class="right">
                  <dt>Order:</dt>
                  <dd><a target="_blank" :href="`${adminURL}/${p.unit.order.id}`">{{p.unit.order.id}}</a></dd>
                  <dt>Unit:</dt>
                  <dd><a target="_blank" :href="`${adminURL}/${p.unit.id}`">{{p.unit.id}}</a></dd>
                  <dt>Workflow:</dt>
                  <dd>{{p.workflow.name}}</dd>
                  <dt>Category:</dt>
                  <dd>{{p.category.name}}</dd>
                </dl>
             </div>
             <div class="status" v-if="isFinished(p) == false">
                <span class="assignment">
                   <i class="user fas fa-user"></i>
                   <span v-if="p.owner.id == 0" class="unassigned">Unassigned</span>
                   <span v-else class="assigned">{{ownerInfo(p)}}</span>
                </span>
             </div>
          </li>
      </ul>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
export default {
   name: "Home",
   components: {
   },
   computed: {
      ...mapState({
         loading : state => state.loading,
         projects : state => state.projects.projects,
         jwt : state => state.user.jwt,
         adminURL: state => state.adminURL
      }),
      ...mapGetters([
         'isAdmin',
         'isSupervisor',
      ])
   },
   methods: {
      ownerInfo(p) {
         return `${p.owner.firstName} ${p.owner.lastName} (${p.owner.computingID})`
      },
      isFinished(p) {
         return p.finishedAt != null
      },
      isOverdue(p) {
         let due =  new Date(p.dueOn)
         let now = new Date()
         return now > due
      }
   },
   created() {
      if (this.jwt != "") {
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
   .projects {
      list-style: none;
      margin: 0;
      padding: 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: center;
      .card {
         flex: 0 1 calc(25% - 1em);
         border: 1px solid var(--uvalib-grey);
         padding: 0;
         margin: 5px;
         position: relative;
         text-align: left;
         box-sizing: border-box;
         min-width: 45%;
         color: var(--uvalib-text);
         font-size: 0.9em;
         .top {
            background: var(--uvalib-grey-lightest);
            border-bottom: 1px solid var(--uvalib-grey);
            color: var(--uvalib-text);
            .title {
               padding: 10px;
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
            font-size: 0.9em;
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
</style>
