<template>
   <UAccordion :items="[{label: 'Notes', value: 'notes'}]" defaultValue="notes" class="panel">
      <template #body="{ }">
         <div v-if="!detail.notes" class="none">
            There are no notes associated with this project
         </div>
         <div v-else class="notes">
            <div class="note-card" v-for="n in detail.notes" :key="`n${n.id}`" :class="noteTypeString(n.type).toLowerCase()">
               <div class="note-info">
                  <div>
                     <p class="note-date">{{formatDate(n.createdAt)}}</p>
                     <p class="note-by">{{ system.getStaffMemberName(n.staffMemberID) }}</p>
                  </div>
                  <div class="right">
                     <p class="note-type">{{noteTypeString(n.type)}}</p>
                     <p v-if="n.step.id > 0" class="note-step"><b>Step: </b>{{n.step.name}}</p>
                  </div>
               </div>
               <div class="note-text">
                  <div class="problems" v-if="n.problems && n.problems.length > 0">{{problemsString(n.problems)}}</div>
                  <div v-html="n.text"></div>
               </div>
            </div>
         </div>
         <div class="buttons">
            <NoteModal v-if="!detail.finishedAt" />
         </div>
      </template>
   </UAccordion>
</template>

<script setup>
import {useSystemStore} from "@/stores/system"
import {useProjectStore} from "@/stores/project"
import NoteModal from '@/components/project/NoteModal.vue'
import { storeToRefs } from 'pinia'
import { useDateFormat } from '@vueuse/core'

const projectStore = useProjectStore()
const system = useSystemStore()
const { detail } = storeToRefs(projectStore)

const problemsString = ((probs) => {
   let out = []
   probs.forEach(p => out.push(p.label) )
   return out.join(", ")
})

const noteTypeString =((typeID) => {
   let types = ["COMMENT", "SUGGESTION", "PROBLEM", "ITEM CONDITION"]
   if ( typeID < 0 || typeID > types.length-1) return "COMMENT"
   return types[typeID]
})

const formatDate =((d) => {
   return useDateFormat(d, "YYYY-MM-DD hh:mm A")
})
</script>

<style scoped lang="scss">
.panel {
   text-align: left;

   .buttons {
      margin-top: 15px;
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      gap: 10px;
      padding-top: 15px;
      border-top: 1px solid var(--uvalib-grey-light);
   }

   .none {
      font-size: 1.15em;
      text-align: center;
      margin: 25px;
   }
   .notes {
      display: flex;
      flex-direction: column;
      gap: 15px;
      max-height: 800px;
      overflow-y: scroll;
   }
   .note-card {
      background-color: white;
      border: 1px solid var(--uvalib-grey-light);
      border-radius: 0;
      padding: 8px;
      color: var(--uvalib-text-dark);
      .note-info {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         border-bottom: 1px solid var(--uvalib-text-dark);
         padding-bottom: 5px;
         margin-bottom: 5px;
         p {
            padding:0;
            margin:0;
         }
         .right {
            text-align: right;
         }
      }
      .problems {
         font-weight: bold;
         margin-bottom: 5px;
      }
      .note-text {
         padding: 10px 5px 5px 5px;
         :deep(p) {
            margin: 0 0 5px 0 !important;
         }
      }
   }
    div.note-card.condition {
      background-color: var(--uvalib-grey-lightest);
      border: 1px solid var(--uvalib-grey);
   }
    div.note-card.comment {
      background-color: var(--uvalib-yellow-light);
      border: 1px solid var(--uvalib-yellow-dark);
   }
   div.note-card.problem {
      background-color: var(--uvalib-red-lightest);
      border: 1px solid var(--uvalib-red-darker);
   }
   div.note-card.suggestion {
      background-color: var(--uvalib-blue-alt-light);
      border: 1px solid var(--uvalib-blue-alt-dark);
   }
}
</style>
