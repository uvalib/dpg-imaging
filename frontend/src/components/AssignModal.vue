<template>
   <DPGButton @click="show" :label="props.label" class="p-button-secondary"/>
   <Dialog v-model:visible="isOpen" :modal="true" header="Assign Project">
      <div class="candidate-scroller">
         <div class="val" v-for="(c,idx) in systemStore.staffMembers" :key="c.id" :class="{selected: idx == selectedIdx}" @click="selectCandidate(idx)">
            <span class="candidate">{{c.lastName}}, {{c.firstName}}</span> ({{c.computingID}})
         </div>
      </div>
      <p class="error">{{error}}</p>
      <template #footer>
         <DPGButton @click="hide" label="Cancel" class="p-button-secondary"/>
         <span class="spacer"></span>
         <DPGButton @click="assignClicked" label="Assign"/>
      </template>
   </Dialog>
</template>

<script setup>
import { ref } from 'vue'
import {useSystemStore} from '@/stores/system'
import {useProjectStore} from '@/stores/project'
import Dialog from 'primevue/dialog'

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
const selectedIdx = ref(-1)
const error = ref("")

const selectCandidate = ((idx) => {
   selectedIdx.value = idx
})

const assignClicked = ( async () => {
   error.value = ""
   if ( selectedIdx.value == -1) {
      error.value = "Please select a user"
      return
   }
   let userID = systemStore.staffMembers[selectedIdx.value].id
   await projectStore.assignProject( {projectID: props.projectID, ownerID: userID} )
   hide()
   emit('assigned')
})

function hide() {
   isOpen.value=false
}
function show() {
   isOpen.value = true
   error.value = ""
   selectedIdx.value = -1
}
</script>

<style lang="scss" scoped>
.error {
   padding: 0;
   margin: 0;
   text-align: center;
   color: var(--uvalib-red-emergency);
}

.candidate-scroller {
   max-height: 300px;
   overflow: scroll;
   padding: 0;
   margin:  0;
   border: 1px solid var(--uvalib-grey-light);
   border-radius: 4px;
   .val {
      padding: 2px 10px 3px 10px;
      cursor: pointer;
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      &:hover  {
         background: var(--uvalib-blue-alt-light);
      }
   }
   .val.selected {
      background: var(--uvalib-blue-alt);
      color: white;
   }
   .candidate {
      font-weight: bold;
   }
}
</style>
