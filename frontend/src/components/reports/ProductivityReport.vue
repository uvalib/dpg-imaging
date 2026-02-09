<template>
   <div class="reports-card">
      <h3>Productivity</h3>
       <div  v-if="reportStore.productivity.loading" class="wait-wrap">
         <WaitSpinner message="Loading Productivity report..."/>
      </div>
      <div  class="report">
         <Chart type="bar" :data="reportStore.productivity" :options="options" />
         <div class="total">
            <label>Total Completed Projects:</label><span class="total">{{reportStore.productivity.totalCompleted}}</span>
         </div>
         <p class="error" v-if="reportStore.productivity.error">{{reportStore.productivity.error}}</p>
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
   div.total {
      margin: 10px 0 5px 0;
      text-align: center;
      span.total {
         margin-left: 10px;
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

