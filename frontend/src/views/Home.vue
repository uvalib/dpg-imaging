<template>
   <div class="home">
      <h2>Welcome to the DPG Imaging Service</h2>
      <WaitSpinner v-if="working" :overlay="true" message="Loading units..." />
      <template v-else>
         <p>Please select a unit from the list below</p>
         <ul class="units">
            <li v-for="unit in units" :key="unit">
               <router-link :to="`/unit/${unit}`">{{unit}}</router-link>
            </li>
         </ul>
      </template>
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
         working : state => state.working,
         units : state => state.units,
      })
   },
   methods: {
   },
   created() {
      this.$store.dispatch("getUnits")
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
   p {
      margin: 25px 0;
      font-weight: bold;
   }
   ul.units {
      list-style: none;
      padding: 0; margin: 0;
      li {
         padding: 5px 0;
         cursor: pointer;
         a {
            color: var(--uvalib-text);
            text-decoration: none;
            &:hover {
               text-decoration: underline;
            }
         }
      }

   }
}
</style>
