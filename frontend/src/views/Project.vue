<template>
   <div class="project">
      <h2>Digitization Project #{{$route.params.id}}</h2>
      <WaitSpinner v-if="loading" :overlay="true" message="Loading projects..." />
      <template v-else>
         <div class="project-head">
            <h3><a target="_blank" :href="metadataLink">{{currProject.unit.metadata.title}}</a></h3>
            <h4>
               <div>
                  <label>Unit:</label>
                  <a target="_blank" :href="`${adminURL}/units/${currProject.unit.id}`">{{currProject.unit.id}}</a>
               </div>
               <div>
                  <label>Order:</label>
                  <a target="_blank" :href="`${adminURL}/orders/${currProject.unit.orderID}`">{{currProject.unit.orderID}}</a>
               </div>
               <div>
                  <label>Customer:</label>
                  <a target="_blank" :href="`${adminURL}/customers/${currProject.unit.order.customer.id}`">
                     {{currProject.unit.order.customer.firstName}} {{currProject.unit.order.customer.lastName}}
                  </a>
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
         margin: 5px auto 15px auto;
         .icon {
            display: inline-block;
            margin-left: 10px;
         }
      }
      h4 {
         font-size: 0.9em;
         label {
            margin-right: 5px;
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
