<template>
   <UPopover>
      <UButton color="error" size="lg" label="Issues Detected" icon="i-lucide-octagon-alert" class="absolute left-2 top-0"/>
      <template #content>
         <div class="list">
            <ol start="1">
               <li v-for="(p,idx) in unitStore.problems" :key="`p${idx}`">
                  {{p}}
               </li>
            </ol>
         </div>
      </template>
   </UPopover>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
const unitStore = useUnitStore()
</script>

<style lang="scss" scoped>
div.list {
   padding: 10px;
   ol {
      margin: 0;
      font-weight: 100;
      text-align: left;
      li {
         white-space: nowrap;
         padding: 2px 8px;
      }
   }
}
</style>