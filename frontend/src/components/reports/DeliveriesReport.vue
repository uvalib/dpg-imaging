<template>
   <div class="reports-card">
      <h3>Patron Deliveries</h3>
      <div  v-if="reportStore.deliveries.loading" class="wait-wrap">
         <WaitSpinner message="Loading Patron Deliveries report..."/>
      </div>
      <div class="report">
         <Chart type="line" :data="reportStore.deliveries" :options="options"/>
         <p class="error" v-if="reportStore.deliveries.error">{{reportStore.deliveries.error}}</p>
         <div class="controls">
            <span class="year-picker">
               <label>Year:<input v-model="tgtYear"></label>
            </span>
            <button @click="loadStats">Generate</button>
         </div>
      </div>
   </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {useReportStore} from '@/stores/reporting'
import WaitSpinner from '@/components/WaitSpinner.vue'
import Chart from 'primevue/chart'

const tgtYear = ref( new Date().getFullYear() )

const options = ref({
   responsive: true,
   plugins: {
      legend: {
         position: 'top',
      },
   },
})

const reportStore = useReportStore()

function loadStats() {
   reportStore.getDeliveriesReport(tgtYear.value)
}

onMounted( () => {
   console.log("EIFUHELIRUHFUIFREFPIFU")
   if (reportStore.deliveries.datasets.length == 0) {
      reportStore.getDeliveriesReport(tgtYear.value)
   }
})
</script>

<style lang="scss" scoped>
.reports-card {
   margin: 10px;
   text-align: left;
   border: 1px solid var(--uvalib-grey-light);
   box-shadow: var(--box-shadow-light);
   position: relative;
   min-height: 360px;
   width: 100%;
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
      .controls {
         border-top: 1px solid var(--uvalib-grey-lightest);
         display: flex;
         flex-flow: row wrap;
         justify-content: flex-end;
         padding-top: 15px;
         margin-top: 5px;
         input {
            margin: 0 10px;
            width: 85px;
            color: var(--uvalib-text);
            text-align: center;
         }
      }
   }
}
</style>

