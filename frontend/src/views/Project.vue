<template>
   <div class="project">
      <h2>Digitization Project #{{$route.params.id}}</h2>
       <WaitSpinner v-if="loading" :overlay="true" message="Loading projects..." />
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
   }
}
</style>
