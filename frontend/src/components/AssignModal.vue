<template>
   <UModal v-model:open="isOpen" :modal="true" :dismissible="false" title="Assign Project">
      <UButton @click="show" :label="props.label" color="secondary"/>
      <template #body>
         <UListbox  v-model="assignee" :items="staff" virtualize filter/>
      </template>
      <template #footer>
         <UButton @click="hide" label="Cancel" color="secondary"/>
         <UButton @click="assignClicked" label="Assign" :disabled="assignee == null"/>
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

const staff = computed( () => {
   let out = []
   systemStore.activeStaff.forEach( s => {
      out.push({label: `${s.lastName}, ${s.firstName}`, value: s})
   })
   return out
})

const assignClicked = ( async () => {
   // the listbox selection assigns an object to assignee with fields label and value. 
   // label is the user name value is the user data. Syntax to get it looks weird (double value)
   const staff = assignee.value.value
   await projectStore.assignProject( props.projectID, staff.id )
   hide()
   emit('assigned')
})

const hide = (() => {
   isOpen.value=false
})

const show = (() => {
   isOpen.value = true
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
