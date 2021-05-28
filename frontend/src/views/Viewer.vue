<template>
   <div class="viewer">
      <WaitSpinner v-if="working" :overlay="true" message="Loading viewer..." />
      <template v-else>
         <div id="tify-viewer"></div>
      </template>
   </div>
</template>

<script>
import { mapState } from "vuex"
export default {
   name: "Viewer",
   computed: {
      ...mapState({
         working : state => state.working,
      })
   },
   async created() {
      let iiifManURL = window.location.href
      iiifManURL = iiifManURL.split("/view")[0]
      iiifManURL = `${iiifManURL}/api/iiif/${this.$route.params.unit}`

      window.tifyOptions = {
         container: '#tify-viewer',
         immediateRender: false,
         manifest: iiifManURL,
         stylesheet: '/tify_mods.css',
         title: null,
      }
      await import ('tify/dist/tify.css')
      await import ('tify/dist/tify.js')
   },
   methods: {
   }
}
</script>

<style lang="scss">
#tify-viewer {
   position: absolute;
   width: 100%;
   bottom: 0;
   top: 100px;
}
::v-deep .tify-header_column.-controls.-visible  {
   display: none !important;
}
.not-found {
   display: inline-block;
   padding: 20px 50px;
   margin: 4% auto 0 auto;
   h2 {
      font-size: 1.5em;
      color: var(--uvalib-text);
   }
}
.extra-tools {
   z-index: 1000;
   position: absolute;
   left: 12px;
   top: 12px;
   font-size: 1.1em;
   color: #222;
   cursor: pointer;
   .dl-text {
      margin-left: 5px;
      font-weight: 500;
      &:hover {
         text-decoration: underline;
      }
   }
}
</style>
