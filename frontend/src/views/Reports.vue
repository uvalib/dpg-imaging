<template>
   <h2>Digitization Reports</h2>
   <div class="control-bar">
      <span class="cfg">
         <span class="field">
            <label for="workflow-pick">Workflow:</label>
            <select id="workflow-pick" v-model="reportStore.workflowID">
               <option v-for="w in system.activeWorkflows" :value="w.id" :key="`wf${w.id}`">{{w.name}}</option>
            </select>
         </span>

         <span class="field">
            <label for="start">From:</label>
            <DatePicker id="start" v-model="reportStore.startDate" showIcon :showOnFocus="false" dateFormat="yy-mm-dd"/>
         </span>
         <span class="field">
            <label for="end">To:</label>
            <DatePicker id="end" v-model="reportStore.endDate" showIcon :showOnFocus="false" dateFormat="yy-mm-dd"/>
         </span>
      </span>
      <DPGButton label="Generate Reports" @click="loadStats" size="small"/>
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
import { onMounted } from 'vue'
import {useReportStore} from '@/stores/reporting'
import {useSystemStore} from '@/stores/system'
import DatePicker from 'primevue/datepicker'
import PageTimeReport from '@/components/reports/PageTimeReport.vue'
import ProductivityReport from '@/components/reports/ProductivityReport.vue'
import ProblemsReport from '@/components/reports/ProblemsReport.vue'
import RejectionsReport from '@/components/reports/RejectionsReport.vue'
import RatesReport from '@/components/reports/RatesReport.vue'

const reportStore = useReportStore()
const system = useSystemStore()

onMounted( () => {
   reportStore.init()
   loadStats()
})

const loadStats = (() => {
   reportStore.getProductivityReport(reportStore.workflowID, reportStore.startDate, reportStore.endDate)
   reportStore.getProblemsReport(reportStore.workflowID, reportStore.startDate, reportStore.endDate)
   reportStore.getRateReports(reportStore.workflowID, reportStore.startDate, reportStore.endDate)

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