<template>
   <span class="view-mode">
      <label>View:</label>
      <select id="layout" v-model="unitStore.viewMode" @change="viewModeChanged">
         <option value="list">List</option>
         <option value="medium">Gallery (medium)</option>
         <option value="large">Gallery (large)</option>
      </select>
   </span>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const unitStore = useUnitStore()

function viewModeChanged() {
   let query = Object.assign({},route.query)
   query.view = unitStore.viewMode
   router.push({query})
}
</script>

<style lang="scss" scoped>
.view-mode {
   display: flex;
   flex-flow: row nowrap;
   justify-content: flex-start;
   align-items: baseline;
   gap: 5px;

   label {
      display: inline-block;
      font-weight: bold;
   }
   select {
      border: 1px solid var(--uvalib-grey);
   }
}
</style>