<template>
   <div class="box panel">
      <h3>Box Information</h3>
      <div class="content">
         <span class="entry">
            <label>Box:</label>
            <input id="box-id" type="text" v-model="box"  @keyup.enter="okClicked"/>
         </span>
      </div>
      <div class="panel-actions">
         <DPGButton @click="cancelEditClicked" class="right-pad">Cancel</DPGButton>
         <DPGButton @click="okClicked">OK</DPGButton>
      </div>
   </div>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import {useSystemStore} from "@/stores/system"
import { ref, onMounted, nextTick } from 'vue'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const box = ref("")

onMounted( async () => {
   nextTick( () => {
      let ele = document.getElementById("box-id")
      ele.focus()
   })
})

async function okClicked() {
   unitStore.selectAll()
   await unitStore.setLocation("box", box.value )
   cancelEditClicked()
}
function cancelEditClicked() {
   systemStore.error = ""
   unitStore.editMode = ""
}
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   border-bottom: 1px solid var(--uvalib-grey);
   h3 {
      margin: 0;
      padding: 8px 0;
      font-size: 1em;
      background: var(--uvalib-blue-alt-light);
      border-bottom: 1px solid var(--uvalib-grey);
      font-weight: 500;
   }
   .panel-actions {
      padding: 5px 0 20px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
      width: 50%;
      margin: 0 auto;
      .button {
         margin-left: 10px;
      }
      .left {
         margin-right: auto;
      }
   }
   .content {
      padding: 10px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      width: 50%;
      margin: 0 auto;
      .entry {
         flex-grow: 1;
         margin: 0;
         text-align: left;
         label {
            display: block;
            margin: 0 0 5px 0;
         }
      }
      .entry.pad-right {
         padding-right: 25px;
      }
   }
}
</style>
