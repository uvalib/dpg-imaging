<template>
   <div class="project">
      <h2>
         <span>Digitization Project #{{$route.params.id}}</span>
         <span v-if="!loading" class="due"><label>Due:</label><span>{{currProject.dueOn.split("T")[0]}}</span></span>
      </h2>
      <WaitSpinner v-if="loading" :overlay="true" message="Loading projects..." />
      <template v-else>
         <div class="project-head">
            <h3>
               <a target="_blank" :href="metadataLink">{{currProject.unit.metadata.title}}</a>
            </h3>
            <h4 class="proj-data">
               <div class="column right-pad">
                  <div>
                     <label>Unit:</label>
                     <a target="_blank" :href="`${adminURL}/units/${currProject.unit.id}`">{{currProject.unit.id}}</a>
                  </div>
                  <div>
                     <label>Order:</label>
                     <a target="_blank" :href="`${adminURL}/orders/${currProject.unit.orderID}`">{{currProject.unit.orderID}}</a>
                  </div>
               </div>
               <div class="column">
                  <div>
                     <label>Customer:</label>
                     <a target="_blank" :href="`${adminURL}/customers/${currProject.unit.order.customer.id}`">
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
               <router-link to="/">Back to Projects</router-link>
            </span>
         </div>
         <div class="project-main">
            <div class="info-block">
               <h4>Item Information</h4>
               <dl>
                  <dt>Category:</dt>
                  <dd>{{currProject.category.name}}
                  </dd>
                  <dt>Call Number:</dt>
                  <dd>{{currProject.unit.metadata.callNumber}}</dd>
                  <dt>Special Instructions:</dt>
                  <dd>
                     <span v-if="currProject.unit.specialInstructions">{{currProject.unit.specialInstructions}}</span>
                     <span v-else class="na">N/A</span>
                  </dd>
                  <dt>Condition:</dt>
                  <dd>{{conditionText(currProject.ItemCondition)}}</dd>
                  <dt>Condition Notes:</dt>
                  <dd>
                     <span v-if="currProject.conditionNote">{{currProject.conditionNote}}</span>
                     <span v-else class="na">N/A</span>
                  </dd>
               </dl>
            </div>
            <div class="info-block">
               <h4>Equipment</h4>
            </div>
            <div class="info-block workflow">
               <h4>Workflow</h4>
            </div>
            <div class="info-block">
               <h4>History</h4>
            </div>
            <div class="info-block">
               <h4>Notes</h4>
            </div>
         </div>
      </template>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
export default {
   name: "project",
   components: {
   },
   computed: {
      ...mapState({
         loading : state => state.loading,
         projects : state => state.projects.projects,
         adminURL: state => state.adminURL,
         selectedProjectIdx: state => state.projects.selectedProjectIdx,
      }),
      ...mapGetters({
         currProject: 'projects/currProject',
         isAdmin: 'isAdmin',
         isSupervisor: 'isSupervisor',
         statusText: 'projects/statusText',
         percentComplete: 'projects/percentComplete'
      }),
      metadataLink() {
         let mdType = "sirsi_metadata"
         if (this.currProject.unit.metadata.type == "XmlMetadata") {
            mdType = "xml_metadata"
         }
         return `${this.adminURL}/${mdType}/${this.currProject.unit.metadata.id}`
      }
   },
   methods: {
      conditionText(condID) {
         if (condID == 0) return "Good"
         return "Bad"
      }
   },
   created() {
      if (this.selectedProjectIdx == -1) {
         this.$store.dispatch("projects/getProject", this.$route.params.id)
      }
   },
};
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
         label {
            font-weight: bold;
            margin-right: 5px;
         }
      }
   }
   .project-head {
      color: var(--uvalib-text);
      padding-bottom: 15px;
      border-bottom: 1px solid var(--uvalib-grey-light);
      position: relative;
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
            color: var(--uvalib-text);
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

      .info-block.workflow {
         width: 95%;
      }
      .info-block {
         width: 46%;
         min-width: 600px;
         border: 1px solid var(--uvalib-grey-light);
         margin: 15px;
         display: inline-block;
         min-height: 100px;
          box-shadow: 0 1px 3px rgba(0,0,0,.06), 0 1px 2px rgba(0,0,0,.12);

         h4 {
            color: var(--uvalib-text);
            font-size: 1em;
            margin: 0;
            padding: 5px;
            background: var(--uvalib-grey-lightest);
            border-bottom: 1px solid var(--uvalib-grey-light);
         }
         dl {
            margin: 10px 30px;
            display: inline-grid;
            grid-template-columns: max-content 2fr;
            grid-column-gap: 10px;
            font-size: 0.9em;
            text-align: left;

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
   }
}
</style>
