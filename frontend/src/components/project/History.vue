<template>
   <Panel header="History" class="panel" toggleable>
      <div class="timing">
         <span>
            <label>Date started:</label>
            <span v-if="projectStore.detail.startedAt">{{formatDate(projectStore.detail.startedAt)}}</span>
            <span v-else class="na">N/A</span>
         </span>
         <span>
            <label>Total work time:</label>
            <span v-if="projectStore.detail.startedAt">{{totalWorkTime}}</span>
            <span v-else class="na">N/A</span>
         </span>
      </div>
      <div class="history-wrap">
         <table class="history"><tbody>
            <tr>
               <th>Date</th><th>Step</th><th>Activity</th><th>Owner</th>
            </tr>
            <template v-for="a in projectStore.detail.assignments" :key="`a${a.id}`">
               <template v-if="a.step.name != 'Unknown'">
                  <template v-if="a.finishedAt">
                     <!-- status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing, :working] -->
                     <tr :class="{success: a.status != 3, reject: a.status==3}">
                        <td>{{formatDate(a.finishedAt)}}</td>
                        <td>{{a.step.name}}</td>
                        <td>
                           <span>
                              <template v-if="a.status==3">Rejected</template>
                              <template v-else>Finished</template>
                              <template v-if="a.status != 5"> <!-- not reassigned -->
                                 <br/>{{a.durationMinutes}} mins
                              </template>
                           </span>
                        </td>
                        <td>{{ system.getStaffMemberName(a.staffMemberID) }}</td>
                     </tr>
                  </template>
                  <template v-if="a.startedAt">
                     <tr :class="{error: a.status == 4, finalize: a.status == 6, working: a.status == 7}">
                        <td>{{formatDate(a.startedAt)}}</td>
                        <td>{{a.step.name}}</td>
                        <td v-if="a.status == 4">Error</td>
                        <td v-else-if="a.status == 6">Finalizing...</td>
                        <td v-else-if="a.status == 7">Working...</td>
                        <td v-else>Started</td>
                        <td>{{ system.getStaffMemberName(a.staffMemberID) }}</td>
                     </tr>
                  </template>
                  <tr v-else :class="{reassign: a.status == 5}">
                     <td>{{formatDate(a.assignedAt)}}</td>
                     <td>{{a.step.name}}</td>
                     <td v-if="a.status == 5">Reassigned</td>
                     <td v-else>Assigned</td>
                     <td>{{ system.getStaffMemberName(a.staffMemberID) }}</td>
                  </tr>

               </template>
            </template>
            <tr class="create">
               <td>{{formatDate(projectStore.detail.addedAt)}}</td>
               <td>Project #{{projectStore.detail.id}}</td>
               <td>Created</td>
               <td></td>
            </tr>
         </tbody></table>
      </div>
   </Panel>
</template>

<script setup>
import { useDateFormat } from '@vueuse/core'
import { useProjectStore } from "@/stores/project"
import {useSystemStore} from "@/stores/system"
import { computed } from 'vue'
import Panel from 'primevue/panel'

const projectStore = useProjectStore()
const system = useSystemStore()

const totalWorkTime = computed(() => {
   let mins = 0
   projectStore.detail.assignments.forEach( a => {
      mins += a.durationMinutes
   })
   let h = 0
   if (mins > 60) {
      h = Math.trunc(mins/60)
      mins -= (h*60)
   }
   return `${(""+h).padStart(2,"0")}:${(""+mins).padStart(2,"0")}`
})

const formatDate = ( (d) => {
   return useDateFormat(d, "YYYY-MM-DD hh:mm A")
})
</script>

<style scoped lang="scss">
.panel {
   text-align: left;
   .timing {
      display: flex;
      flex-flow: row nowrap;
      justify-content: space-between;
      padding: 10px 10px 0 10px;
      font-size: 0.8em;
      border: none;
      label {
         font-weight: bold;
         margin-right: 5px;;
      }
   }
   .history-wrap {
      padding: 10px;
   }
   .history {
      font-size: 0.8em;
      width: 100%;
      border-collapse: collapse;
      border: 1px solid var(--uvalib-grey-light);
      td, th {
         padding:4px 10px;
      }
      th {
         border-bottom: 1px solid var(--uvalib-grey-light);
         border-top: 1px solid var(--uvalib-grey-light);
         background: var(--uvalib-grey-lightest);
         padding: 10px;
      }

      tr.working td {
         background: var(--uvalib-teal-lightest);
      }
      tr.create td {
         background: var(--uvalib-teal-darker);
         color: white;
      }
      tr.success td {
         background: #C3F3CF;
      }
      tr.reject td {
         background: #F3CFCF;
      }
      tr.error td{
         background: #a33;
         color: white;
      }
      tr.reassign td {
         background: lightgoldenrodyellow;
         color: var(--uvalib-text);
      }
      tr.finalize {
         background-color: #5a5;
         color: white;
      }
   }
}
</style>
