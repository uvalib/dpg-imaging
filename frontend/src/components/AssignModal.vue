<template>
   <DPGButton @click="show" :label="props.label" severity="secondary"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Assign Project" :closable="false">
      <Listbox v-model="assignee" :options="staff" filter optionLabel="name" optionValue="value" />
      <p class="error">{{error}}</p>
      <template #footer>
         <DPGButton @click="hide" label="Cancel" severity="secondary"/>
         <span class="spacer"></span>
         <DPGButton @click="assignClicked" label="Assign"/>
      </template>
   </Dialog>
</template>

<script setup>
import { ref, computed } from 'vue'
import {useSystemStore} from '@/stores/system'
import {useProjectStore} from '@/stores/project'
import Dialog from 'primevue/dialog'
import Listbox from 'primevue/listbox'

const emit = defineEmits( ['assigned' ])

const systemStore = useSystemStore()
const projectStore = useProjectStore()

const props = defineProps({
      projectID: {
         type: Number,
         required: true,
      },
      label: {
         type: String,
         default: "Assign"
      }
   })

const isOpen = ref(false)
const assignee = ref()
const error = ref("")

const staff = computed( () => {
   let out = []
   systemStore.activeStaff.forEach( s => {
      out.push({name: `${s.lastName}, ${s.firstName}`, value: s})
   })
   return out
})

const assignClicked = ( async () => {
   error.value = ""
   if ( !assignee.value ) {
      error.value = "Please select a user"
      return
   }
   await projectStore.assignProject( {projectID: props.projectID, ownerID: assignee.value.id} )
   hide()
   emit('assigned')
})

const hide = (() => {
   isOpen.value=false
})

const show = (() => {
   isOpen.value = true
   error.value = ""
   assignee.value = null
})
</script>

<style lang="scss" scoped>
.error {
   padding: 0;
   margin: 0;
   text-align: center;
   color: var(--uvalib-red-emergency);
}
</style>
