<template>
   <div class="problems-wrap" v-if="unitStore.problems.length > 0">
      <span class="problems-btn" v-if="unitStore.problems" @click="toggleList()">
         {{unitStore.problems.length}} <span v-if="unitStore.problems.length==1">issue</span><span v-else>issues</span> detected
         <i class="caret fas fa-angle-down" :class="{rotate: opened}"></i>
      </span>
      <div class="list" v-if="opened">
         <ol start="1">
            <li v-for="(p,idx) in unitStore.problems" :key="`p${idx}`">
               {{p}}
            </li>
         </ol>
      </div>
   </div>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import {ref } from 'vue'

const unitStore = useUnitStore()
const opened = ref(false)

function toggleList() {
   opened.value = !opened.value
}
</script>

<style lang="scss" scoped>
div.problems-wrap {
   position: relative;
   span.problems-btn {
      white-space: nowrap;
      font-size: 14px;
      color: white;
      background: var(--uvalib-red-darker);
      padding: 5px 12px;
      border-radius: 4px;
      position: absolute;
      left: 10px;
      top: 0;
      cursor: pointer;
      &:hover {
         background: var(--uvalib-red-emergency);
      }
      .caret {
         margin-left: 5px;
      }
      .caret.rotate {
         transform: rotate(180deg);
      }
   }
   div.list {
      background: white;
      padding: 5px;
      border: 1px solid var(--uvalib-grey);
      box-shadow: var(--box-shadow);
      z-index: 100;
      position: absolute;
      font-size: 14px;
      left: 10px;
      top: 24px;

      ol {
         color: var(--uvalib-text);
         margin: 0;
         padding-left: 20px;
         font-weight: 100;
         text-align: left;
         li {
            white-space: nowrap;
            padding: 2px 8px;
         }
      }
   }
}
</style>