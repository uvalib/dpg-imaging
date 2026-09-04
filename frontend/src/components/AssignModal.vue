<template>
   <UButton @click="show" :label="props.label" color="secondary"/>
   <UModal v-model:open="isOpen" :modal="true" :dismissible="false" title="Assign Project">
      <template #body>
         <!-- <Listbox v-model="assignee" :options="staff" filter optionLabel="name" optionValue="value" /> -->
         <UListbox  v-model="assignee" :items="staff" virtualize filter />
         <p class="error">{{error}}</p>
      </template>
      <template #footer>
         <UButton @click="hide" label="Cancel" color="secondary"/>
         <UButton @click="assignClicked" label="Assign"/>
      </template>
   </UModal>
</template>

<script setup>
import { ref, computed } from 'vue'
import {useSystemStore} from '@/stores/system'
import {useProjectStore} from '@/stores/project'

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
      out.push({label: `${s.lastName}, ${s.firstName}`, value: s})
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
