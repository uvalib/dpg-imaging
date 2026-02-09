<template>
   <div class="reports-card">
      <h3>Average Page Completion Time</h3>
      <div v-if="reportStore.pageTimes.loading" class="wait-wrap">
         <WaitSpinner message="Loading Page Completion Time report..."/>
      </div>
      <div class="report">
         <Chart type="bar" :data="reportStore.pageTimes" :options="options"/>
         <table class="raw-data">
            <tbody>
               <tr>
                  <th>Category</th>
                  <th>Units</th>
                  <th>Total Mins</th>
                  <th>Total Pages</th>
                  <th>Avg. Mins</th>
               </tr>
               <tr v-for="(r,idx) in reportStore.pageTimes.raw" :key="`raw${idx}`">
                  <td>{{r.category}}</td>
                  <td>{{r.units}}</td>
                  <td>{{r.totalMins}}</td>
                  <td>{{r.totalPages}}</td>
                  <td>{{r.avgPageTime}}</td>
               </tr>
            </tbody>
         </table>
         <p class="error" v-if="reportStore.pageTimes.error">{{reportStore.pageTimes.error}}</p>
      </div>
   </div>
</template>

<script setup>
import { ref } from 'vue'
import {useReportStore} from '@/stores/reporting'
import WaitSpinner from '@/components/WaitSpinner.vue'
import Chart from 'primevue/chart'

const options = ref({
   title: {
      display: false,
   },
   legend: {
      display: false
   },
   plugins: {
      legend: {
         display: false,
      },
      colors: {
         enabled: false
      }
   },
})

const reportStore = useReportStore()
</script>

<style lang="scss" scoped>
.reports-card {
   margin: 10px;
   text-align: left;
   border: 1px solid var(--uvalib-grey-light);
   box-shadow: var(--box-shadow-light);
   position: relative;
   min-height: 360px;
   h3 {
      background: var(--uvalib-grey-lightest);
      font-size: 1em;
      text-align: left;
      margin:0;
      padding: 5px 10px;
      border-bottom: 1px solid var(--uvalib-grey-light);
      font-weight: 500;;
   }
   .raw-data {
      border-collapse: collapse;
      width: 100%;
      margin: 10px 0 0 0;
      font-size: 0.9em;
      th {
         border-bottom: 1px solid var(--uvalib-grey-light);
      }
      th, td {
         padding: 5px 0;
      }
   }
   .wait-wrap {
      text-align: center;
      padding: 75px 0;
      position: absolute;
      left: 0;
      right: 0;
   }
   .report {
      padding: 10px;
   }
}
</style>

