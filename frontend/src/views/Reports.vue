<template>
   <h2>Digitization Reports</h2>
   <div class="control-bar">
      <span class="cfg">
         <span class="field">
            <label for="workflow-pick">Workflow:</label>
            <USelect id="workflow-pick" v-model="workflowID" :items="system.activeWorkflows" valueKey="id" labelKey="name" />
         </span>
         <span class="field">
            <label for="start">From:</label>
             <UPopover v-model:open="startOpen">
               <UButton color="secondary" trailing-icon="i-lucide-calendar" :label="startDate.toString()"/>
               <template #content>
                  <UCalendar v-model="startDate" class="p-2" @update:modelValue="startOpen = false"/>
               </template>
            </UPopover>
         </span>
         <span class="field">
            <label for="end">To:</label>
            <UPopover v-model:open="endOpen">
               <UButton color="secondary" trailing-icon="i-lucide-calendar" :label="endDate.toString()"/>
               <template #content>
                  <UCalendar v-model="endDate" class="p-2"  @update:modelValue="endOpen = false"/>
               </template>
            </UPopover>
         </span>
      </span>
      <UButton label="Generate Reports" @click="loadStats()"/>
   </div>
   <div class="reports">
      <div class="column">
          <PageTimeReport />
          <ProductivityReport />
      </div>
      <div class="column">
         <ProblemsReport />
         <RejectionsReport />
         <RatesReport />
      </div>
   </div>
</template>

<script setup>
import { onMounted, shallowRef, ref } from 'vue'
import {useReportStore} from '@/stores/reporting'
import {useSystemStore} from '@/stores/system'
import { today, getLocalTimeZone } from '@internationalized/date'
import PageTimeReport from '@/components/reports/PageTimeReport.vue'
import ProductivityReport from '@/components/reports/ProductivityReport.vue'
import ProblemsReport from '@/components/reports/ProblemsReport.vue'
import RejectionsReport from '@/components/reports/RejectionsReport.vue'
import RatesReport from '@/components/reports/RatesReport.vue'

const reportStore = useReportStore()
const system = useSystemStore()

const startOpen = ref(false)
const startDate = shallowRef( today( getLocalTimeZone() ).subtract({months: 3}) )
const endOpen = ref(false)
const endDate = shallowRef( today( getLocalTimeZone() ) )
const workflowID = ref(1)

onMounted( () => {
   loadStats()
})

const startPicked = (() => {
  startOpen.value = false
})

const loadStats = (() => {
   reportStore.clearStats()
   const start = startDate.value.toString()
   const end = endDate.value.toString()
   reportStore.getProductivityReport(workflowID.value, start, end)
   reportStore.getProblemsReport(workflowID.value, start, end)
   reportStore.getRateReports(workflowID.value, start, end)

})

</script>

<style scoped lang="scss">

h2 {
   display: flex;
   flex-flow: row wrap;
   justify-content: space-between;
   margin: 0 !important;
   background-color: var(--uvalib-grey-lightest);
   border-bottom: 1px solid var(--uvalib-grey-light);
   padding: 0.75rem 1rem;
}

.control-bar {
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-between;
   align-items: center;
   padding: 10px;
   border-bottom: 1px solid var(--uvalib-grey-light);
   margin-bottom: 15px;
   border-top: 1px solid var(--uvalib-grey-light);
   .cfg, .field{
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-start;
      align-items: center;
      gap: 20px;
   }
   .field {
      gap: 5px;
   }
}
.reports {
   margin: 10px 50px;
   display: flex;
   flex-flow: row wrap;
   .column {
      width: 48%;
   }
}
</style>