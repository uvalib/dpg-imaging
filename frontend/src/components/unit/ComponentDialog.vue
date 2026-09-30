<template>
   <UModal v-model:open="open" :modal="true" :dismissible="false" :close="false" title="Link Component">
      <UButton @click="showClicked()" size="sm" color="secondary" label="Link Component" />
      <template #body>
         <div class="panel confirm" v-if="unitStore.component.valid">
            <table>
               <tbody>
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
               </tbody>
            </table>
            <p class="confirm">Link this component to selected images?</p>
         </div>
         <div v-else class="panel">
            <div class="row">
               <span class="entry">
                  <label>Start Image:</label>
                  <USelect v-model="startIdx" @change="startChanged" placeholder="Select start page" :items="masterFiles" />
               </span>
               <span class="entry">
                  <label>End Image:</label>
                  <USelect v-model="endIdx" @change="endChanged" filter placeholder="Select end page" :items="masterFiles"/>
               </span>
               <UButton @click="selectAllClicked" size="sm" color="secondary" label="Select All"/>
            </div>
            <div class="row">
               <span class="entry">
                  <label>Component ID:</label>
                  <UInput id="component-id" v-model="componentID"  @keyup.enter="okClicked"/>
               </span>
            </div>   
            <p class="error" v-if="error">{{ error }}</p>
         </div>
      </template>
      <template #footer="{ close }">
         <template  v-if="unitStore.component.valid">
            <UButton @click="noLinkClicked" color="secondary" label="No"/>
            <UButton @click="linkConfirmed"  label="Yes" />
         </template>
         <template v-else>
            <UButton label="Unlink" color="error" class="left" @click="unlinkClicked" />
            <UButton label="Cancel" color="secondary" @click="close" />
            <UButton label="OK" @click="okClicked"  :loading="lookingUp" :disabled="lookingUp" />
         </template>
      </template>
   </UModal>
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { ref, computed } from 'vue'
import { onKeyStroke } from '@vueuse/core'

const unitStore = useUnitStore()
const componentID = ref("")
const lookingUp = ref(false)
const open = ref(false)
const startIdx = ref()
const endIdx = ref()
const error = ref("")

onKeyStroke('k', (e) => {
   if ( e.ctrlKey ) {
      showClicked()
   }
})

const masterFiles = computed( () => {
   let list = []
   unitStore.masterFiles.forEach( (mf,idx) => {
      list.push({ value: idx, label: mf.fileName })
   })
   return list
})

const showClicked = (() => {
   componentID.value = ""
   if (unitStore.rangeStartIdx > -1 ) {
      startIdx.value = unitStore.rangeStartIdx
   }
   if (unitStore.rangeEndIdx > -1 ) {
      endIdx.value = unitStore.rangeEndIdx
   }
   open.value = true
})

const startChanged = (() => {
   error.value = ""
   unitStore.startFileSelected( startIdx.value )
})
const endChanged = (() => {
   error.value = ""
   unitStore.endFileSelected( endIdx.value )
})

const formatData = (( value ) => {
   if (value && value != "" )  return value
   return "N/A"
})

const okClicked = (async () => {
   error.value = ""
   unitStore.clearComponent()
   if ( startIdx.value == -1 || endIdx.value == -1) {
      error.value = "Start and end image must be selected"
      return
   }
   if (componentID.value == "") {
      error.value = "Component ID is required"
      return
   }
   lookingUp.value = true
   await unitStore.lookupComponentID(componentID.value)
   lookingUp.value = false
})

const noLinkClicked = (() => {
   unitStore.clearComponent()
})

const cancelEditClicked = (()=> {
   unitStore.clearComponent()
   open.value = false
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
   startIdx.value = unitStore.rangeStartIdx
   endIdx.value = unitStore.rangeEndIdx
})
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   display: flex;
   flex-direction: column;
   gap: 20px;
   padding: 10px;
   
   .error {
      margin: 0;
      padding: 0;
      color: var(--uvalib-red-emergency);
   }

   td.data {
      text-align: left;
   }
   td.label {
      text-align: right;
      font-weight: bold;
      padding-right: 10px;
   }

   .row {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      align-items: flex-end;
      justify-content: flex-start;
      gap: 10px;
      text-align: left;

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
