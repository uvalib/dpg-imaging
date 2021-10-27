<template>
   <div class="home">
      <h2>Digitization Projects</h2>
      <WaitSpinner v-if="loading" :overlay="true" message="Loading projects..." />
      <ul v-else class="projects">
          <li class="card" v-for="p in projects" :key="`p${p.id}`">
             <div class="top">
                <div class="title">{{p.unit.metadata.title}}</div>
                <div class="due"><label>Date Due:</label><span>{{p.dueOn.split("T")[0]}}</span></div>
            </div>
             <div class="data">
                <dl>
                  <dt>Order:</dt>
                  <dd><a target="_blank" :href="`${adminURL}/${p.unit.order.id}`">{{p.unit.order.id}}</a></dd>
                </dl>
                <dl>
                  <dt>Customer:</dt>
                  <dd>{{p.unit.order.customer.firstName}} {{p.unit.order.customer.lastName}}</dd>
                </dl>
             </div>
             <div class="status">
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
import { mapState } from "vuex"
export default {
   name: "Home",
   components: {
   },
   computed: {
      ...mapState({
         loading : state => state.loading,
         projects : state => state.projects,
         jwt : state => state.user.jwt,
         adminURL: state => state.adminURL

      })
   },
   methods: {
      ownerInfo(p) {
         return `${p.owner.firstName} ${p.owner.lastName} (${p.owner.computingID})`
      }
   },
   created() {
      if (this.jwt != "") {
         this.$store.dispatch("getProjects")
      }
   },
};
</script>

<style scoped lang="scss">
.home {
   padding: 25px;
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 50px;
   }
   .projects {
      list-style: none;
      margin: 0;
      padding: 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: center;
      .card {
         border-radius: 5px;
         flex: 0 1 calc(25% - 1em);
         border: 1px solid var(--uvalib-grey);
         padding: 0;
         margin: 5px;
         position: relative;
         text-align: left;
         box-sizing: border-box;
         min-width: 45%;
         color: var(--uvalib-text);
         .top {
            border-radius: 5px 5px 0 0;
            background: var(--uvalib-grey-lightest);
            padding: 5px 10px;
            border-bottom: 1px solid var(--uvalib-grey);
            color: var(--uvalib-text);
            .title {
               margin-bottom: 10px;
            }
            .due {
               label {
                  font-weight: bold;
                  margin-right: 5px;
               }
            }
         }
         .data {
            padding: 10px;
            display: flex;
            flex-flow: row nowrap;
            justify-content: flex-start;
            font-size: 0.9em;
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
