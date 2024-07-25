<template>
   <DPGButton  severity="secondary" @click="showClicked">
      Link Component
   </DPGButton>
   <Dialog v-model:visible="unitStore.edit.component" :modal="true" header="Link Component" @show="opened" :closable="false">
      <div class="panel confirm" v-if="unitStore.component.valid">
         <table>
            <tr>
               <td class="label">Title:</td>
               <td class="data">{{formatData(unitStore.component.title)}}</td>
            </tr>
            <tr>
               <td class="label">Label:</td>
               <td class="data">{{formatData(unitStore.component.label)}}</td>
            </tr>
            <tr>
               <td class="label">Description:</td>
               <td class="data">{{formatData(unitStore.component.description)}}</td>
            </tr>
            <tr>
               <td class="label">Date:</td>
               <td class="data">{{formatData(unitStore.component.date)}}</td>
            </tr>
            <tr>
               <td class="label">Type:</td>
               <td class="data">{{formatData(unitStore.component.type)}}</td>
            </tr>
         </table>
         <p class="confirm">Link this component to selected images?</p>
      </div>
      <div class="panel" v-else>
         <div class="row">
            <span class="entr">
               <label>Start Image:</label>
               <Select v-model="unitStore.rangeStartIdx" @change="startChanged" filter placeholder="Select start page"
                  :options="masterFiles" optionLabel="label" optionValue="value" ref="pickstart" />
            </span>
            <span class="entry">
               <label>End Image:</label>
               <Select v-model="unitStore.rangeEndIdx" @change="endChanged" filter placeholder="Select end page"
                  :options="masterFiles" optionLabel="label" optionValue="value"/>
            </span>
            <DPGButton @click="selectAllClicked" severity="secondary" label="Select All"/>
         </div>
         <div class="row">
            <span class="entry full">
               <label>Component ID:</label>
               <input id="component-id" type="text" v-model="componentID"  @keyup.enter="okClicked"/>
            </span>
         </div>
      </div>
      <template #footer>
         <template  v-if="unitStore.component.valid">
            <DPGButton severity="secondary" @click="noLinkClicked" label="No"/>
            <DPGButton @click="linkConfirmed" label="Yes"/>
         </template>
         <template v-else>
            <DPGButton @click="unlinkClicked" severity="danger" class="left">Unlink</DPGButton>
            <DPGButton @click="cancelEditClicked" severity="secondary">Cancel</DPGButton>
            <DPGButton @click="okClicked" :loading="lookingUp">OK</DPGButton>
         </template>
      </template>
   </Dialog>
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { useSystemStore } from "@/stores/system"
import { ref, nextTick, computed } from 'vue'
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'

const unitStore = useUnitStore()
const systemStore = useSystemStore()
const componentID = ref("")
const lookingUp = ref(false)
const pickstart = ref()

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   unitStore.edit.component = true
   componentID.value = ""
})

const opened = (() => {
   nextTick( () => {
      pickstart.value.$el.focus()
   })
})

const startChanged = (() => {
   unitStore.startFileSelected( unitStore.rangeStartIdx )
})
const endChanged = (() => {
   unitStore.endFileSelected( unitStore.rangeEndIdx )
})

const formatData = (( value ) => {
   if (value && value != "" )  return value
   return "N/A"
})

const okClicked = (async () => {
   unitStore.clearComponent()
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.setError("Start and end image must be selected")
      return
   }
   if (componentID.value == "") {
      systemStore.setError("Component ID is required")
      return
   }
   lookingUp.value = true
   await unitStore.lookupComponentID(componentID.value)
   lookingUp.value = false
})

const noLinkClicked = (() => {
   unitStore.clearComponent()
   nextTick( () => {
      let ele = document.getElementById("component-id")
      ele.focus()
   })
})

const cancelEditClicked = (()=> {
   unitStore.clearComponent()
   unitStore.edit.component = false
})

const unlinkClicked= (() => {
   unitStore.clearComponent()
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.setError("Start and end image must be selected")
      return
   }
   unitStore.componentLink("")
   cancelEditClicked()
})

const linkConfirmed = ( () => {
   unitStore.componentLink(componentID.value)
   cancelEditClicked()
})

const selectAllClicked = (() => {
   unitStore.selectAll()
})
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   display: flex;
   flex-direction: column;
   gap: 20px;

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-end;
      justify-content: flex-start;
      gap: 10px;

      label {
         display: block;
         margin-bottom: 5px;
      }
      .entry.full {
         width: 100%;
      }
   }
}
.left {
   margin-right: auto;
}
</style>
