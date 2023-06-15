<template>
   <Panel class="panel">
      <template #header>
         <div class="panel-header">
            <span>Notes</span>
            <NoteModal id="note-modal" />
         </div>
      </template>
      <div v-if="!currProject.notes" class="none">
         There are no notes associated with this project
      </div>
      <div v-else class="note-card" v-for="n in currProject.notes" :key="`n${n.id}`" :class="noteTypeString(n.type).toLowerCase()">
         <div class="note-info">
            <div>
               <p class="note-date">{{formatDate(n.createdAt)}}</p>
               <p class="note-by">{{n.staffMember.firstName}} {{n.staffMember.lastName}}</p>
            </div>
            <div class="right">
               <p class="note-type">{{noteTypeString(n.type)}}</p>
               <p v-if="n.stepID > 0" class="note-step"><b>Step: </b>{{n.step.name}}</p>
            </div>
         </div>
         <div class="note-text">
            <div class="problems" v-if="n.problems && n.problems.length > 0">{{problemsString(n.problems)}}</div>
            <div v-html="n.text"></div>
         </div>
      </div>
   </Panel>
</template>

<script setup>
import {useProjectStore} from "@/stores/project"
import NoteModal from '@/components/project/NoteModal.vue'
import { storeToRefs } from 'pinia'
import dayjs from 'dayjs'
import Panel from 'primevue/panel'

const projectStore = useProjectStore()
const { currProject } = storeToRefs(projectStore)

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
   return dayjs(d).format("YYYY-MM-DD hh:mm A")
})
</script>

<style scoped lang="scss">
.panel {
   width: 46%;
   min-width: 600px;
   margin: 15px;
   display: inline-block;
   min-height: 100px;
   text-align: left;
   .panel-header {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      width: 100%;
   }

   .none {
      font-size: 1.15em;
      text-align: center;
      margin: 25px;
   }
   .note-card {
      background-color: white;
      border: 1px solid var(--uvalib-grey-light);
      border-radius: 0;
      margin: 5px 0 15px 0;
      padding: 8px;
      .note-info {
         display: flex;
         flex-flow: row nowrap;
         justify-content: space-between;
         font-size: 0.85em;
         border-bottom: 1px solid var(--uvalib-grey-light);
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
         font-size: 0.85em;
         padding: 10px 5px 5px 5px;
         :deep(p) {
            margin: 0 0 5px 0 !important;
         }
      }
   }
    div.note-card.comment {
      background-color: #ffe;
      border: 1px solid #cc9;
      color: #660;
      .note-info {
         border-color: #cc9;
      }
   }
   div.note-card.problem {
      background-color: #fee;
      border: 1px solid #daa;
       color: #700;
      .note-info {
         border-color: #daa;
      }
   }
   div.note-card.suggestion {
      background-color: #eef;
      border: 1px solid #aad;
      color: #007;
      .note-info {
         border-color: #aad;
      }
   }
}
</style>
