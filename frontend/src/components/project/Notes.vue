<template>
   <div class="panel">
      <div v-if="currProject.notes.length == 0" class="none">
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
               <p v-if="n.stepID > 0" class="note-step"><b>Step: </b>{{lookupStepName(n.stepID)}}</p>
            </div>
         </div>
         <div class="note-text">

            <div v-html="n.text"></div>
         </div>
      </div>
   </div>
</template>

<script>
import { mapGetters } from "vuex";
import date from "date-and-time";
export default {
   computed: {
      ...mapGetters({
         currProject: "projects/currProject",
      }),
    },
   methods: {
      noteTypeString(typeID) {
         let types = ["COMMENT", "SUGGESTION", "PROBLEM", "ITEM CONDITION"]
         if ( typeID < 0 || typeID > types.length-1) return "COMMENT"
         return types[typeID]
      },
      problems( problems ) {
         let out = []
         problems.forEach( p => out.push(p.name))
         return out.join(", ")
      },
      lookupStepName(stepID) {
         let s = this.currProject.workflow.steps.find((s) => s.id == stepID);
         if (s) {
            return s.name;
         }
         return "Unknown";
      },
      formatDate(d) {
         return date.format(new Date(d), "YYYY-MM-DD hh:mm A");
      },
   },
};
</script>

<style scoped lang="scss">
.panel {
   padding: 10px;
   .none {
      font-size: 1.15em;
      text-align: center;
      margin-top: 25px;
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