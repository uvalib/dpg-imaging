<template>
   <UModal v-model:open="isOpen" :modal="true" :dismissible="false" :close="false" title="Create Note">
      <UButton label="Add Note" @click="showClicked" />
      <template #body>
         <div class="note-modal-content">
            <div class="row">
               <label>Note Type</label>
               <USelect v-model="noteTypeID" :items="noteTypes" autofocus  class="w-full"/>
            </div>
            <div class="row" v-if="noteTypeID==2">
               <label>Problem (select all that apply)</label>
                <UCheckboxGroup v-model="problemIDs" :items="systemStore.problemTypes" value-key="id" />
            </div>
            <div class="row">
               <label for="note-text">Note Text</label>
               <UTextarea v-model="note" class="w-full"/>
            </div>
         </div>
         <p class="error" v-if="error">{{error}}</p>
      </template>
      <template #footer>
         <UButton @click="hide" color="secondary" label="Cancel"/>
         <UButton @click="createClicked" label="Create"/>
      </template>
   </UModal>
</template>

<script setup>
import { useSystemStore } from '@/stores/system'
import { useProjectStore } from '@/stores/project'
import { ref } from 'vue'

const systemStore = useSystemStore()
const projectStore = useProjectStore()

const noteTypes = [
   {value: 0, label: "Comment"},
   {value: 1, label: "Suggestion"},
   {value: 2, label: "Problem"},
   {value: 3, label: "Item Condition"},
]

const isOpen = ref(false)
const noteTypeID = ref(0) 
const note = ref("")
const problemIDs = ref([])
const error = ref("")

const createClicked = (() => {
   error.value = ""
    if ( note.value == "") {
      error.value = "Note text is required"
      return
   }
   if (noteTypeID.value == 2 && problemIDs.value.length == 0) {
      error.value = "At least one problem is required"
      return
   }
   let data = {noteTypeID: noteTypeID.value, note: note.value, problemIDs: problemIDs.value}
   projectStore.addNote(data)
   isOpen.value = false
})

const hide = (() => {
   isOpen.value = false
})

const showClicked = (() => {
   isOpen.value = true
   noteTypeID.value = 0
   note.value = ""
   problemIDs.value = []
   error.value = ""
})
</script>

<style lang="scss" scoped>
p.error {
   color: var(--uvalib-red-emergency);
   margin: 5px;
   text-align: center;
   font-weight: normal;
   font-style: italic;
}

div.note-modal-content {
   text-align: left;
   font-weight: normal;
   display: flex;
   flex-direction: column;
   gap: 15px;
   label {
      display: block;
      font-weight: bold;
      margin-bottom: 5px;
      font-size: 0.9em;
   }
}
</style>
