<template>
   <div class="reports-card">
      <h3>Problems</h3>
      <div  v-if="reportStore.problems.loading" class="wait-wrap">
         <WaitSpinner message="Loading Problems report..."/>
      </div>
      <div class="report">
         <Chart type="bar" :data="reportStore.problems" :options="options"/>
         <p class="error" v-if="reportStore.problems.error">{{reportStore.problems.error}}</p>
      </div>
   </div>
</template>

<script setup>
import { ref } from 'vue'
import {useReportStore} from '@/stores/reporting'
import WaitSpinner from '@/components/WaitSpinner.vue'
import Chart from 'primevue/chart'

const options = ref({
      responsive: true,
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
      },
    });

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

