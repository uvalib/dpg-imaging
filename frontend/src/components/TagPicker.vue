<template>
   <div class="tag-picker">
      <span tabindex="0" @click="showMenu" @keydown.enter="showMenu" class="tag current" :class="[masterFile.status, props.display]"></span>
      <Popover ref="picker" :dismissable="true">
         <ul>
            <li @click.stop.prevent="selectTag('rescan')">
               <span class="tag rescan"></span>
               <span class="label">Rescan</span>
            </li>
            <li @click.stop.prevent="selectTag('good')">
               <span class="tag good"></span>
               <span class="label">Good</span>
            </li>
            <li @click.stop.prevent="selectTag('recrop')">
               <span class="tag recrop"></span>
               <span class="label">Recrop/Rotate</span>
            </li>
            <li @click.stop.prevent="selectTag('color')">
               <span class="tag color"></span>
               <span class="label">Color Issues</span>
            </li>
            <li @click.stop.prevent="selectTag('other')">
               <span class="tag other"></span>
               <span class="label">Other issues</span>
            </li>
            <li @click.stop.prevent="selectTag('placeholder')">
               <span class="tag placeholder"></span>
               <span class="label">Placeholder</span>
            </li>
            <li @click.stop.prevent="selectTag('none')">
               <span class="tag none">X</span>
               <span class="label">Remove Tag</span>
            </li>
         </ul>
      </Popover>
   </div>
</template>

<script setup>
import Popover from 'primevue/popover'
import {useUnitStore} from "@/stores/unit"
import { ref } from 'vue'

const unitStore = useUnitStore()
const picker = ref()

const props = defineProps({
   masterFile: {
      type: Object,
      required: true
   },
   display: {
      type: String,
      default: "default"
   },
})

function showMenu(event) {
   picker.value.toggle(event)
}
function hideMenu() {
   picker.value.toggle()
}
async function selectTag( tag ) {
   await unitStore.updateMasterFileMetadata( props.masterFile.fileName, "tag", tag )
   hideMenu()
}
</script>

<style lang="scss" scoped>

ul {
   list-style: none;
   margin: 0;
   padding: 0;
   li {
      white-space: nowrap;
      display: flex;
      flex-flow: row nowrap;
      align-content: center;
      padding: 10px;
      cursor: pointer;
      border-radius: 5px;

      .label {
         display: block;
         margin: 0 10px;
         position: relative;
         top: 3px;
      }

      &:hover {
         background: var(--uvalib-blue-alt-light);
      }
   }
}

   .current {
      cursor: pointer;
      display: none;
   }

   .tag.large {
      width: 50px;
      height: 50px;
   }
   .tag.wide {
      width: 100%;
      box-sizing: border-box;
   }
   .tag {
      display: block;
      width: 20px;
      height: 20px;
      border: 1px solid var(--uvalib-grey);
      background: white;
      border-radius: 15px;
   }
   .tag.wide {
      border-radius: 0;
   }
   .tag.rescan {
      background: var(--uvalib-red);
      border-color: var(--uvalib-red-darkest);
   }
   .tag.good {
      background: var(--uvalib-green);
      border-color: var(--uvalib-green-dark);
   }
   .tag.recrop {
      background: var(--uvalib-blue-alt);
      border-color: var(--uvalib-blue-alt-dark);
   }
   .tag.color {
      background: magenta;
      border-color: darkmagenta;
   }
   .tag.other {
      background: var(--uvalib-yellow);
      border-color: var(--uvalib-yellow-dark);
   }
   .tag.placeholder {
      background: var(--uvalib-teal);
      border-color: var(--uvalib-teal-dark);
   }
   .tag.none {
      text-align: center;
      font-size: 20px;
      font-weight: 100;
      color: var(--uvalib-grey-light);
   }
</style>