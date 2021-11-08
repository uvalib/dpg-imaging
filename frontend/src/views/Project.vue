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
      margin-bottom: 15px;
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
}
</style>
