<template>
   <h2 class="text-left! font-medium! m-0! px-4 py-2 bg-brand-grey-200">Digitization Reports</h2>
   <div class="row-between p-3 border-t border-b border-brand-grey-100">
      <span class="row-left items-center gap-4">
         <span class="row-left items-center gap-2">
            <label for="workflow-pick">Workflow:</label>
            <USelect id="workflow-pick" v-model="workflowID" :items="system.activeWorkflows" valueKey="id" labelKey="name" />
         </span>
         <span class="row-left items-center gap-2">
            <label for="start">From:</label>
             <UPopover v-model:open="startOpen">
               <UButton color="secondary" trailing-icon="i-lucide-calendar" :label="startDate.toString()"/>
               <template #content>
                  <UCalendar v-model="startDate" class="p-2" @update:modelValue="startOpen = false"/>
               </template>
            </UPopover>
         </span>
         <span class="row-left items-center gap-2">
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
   <div class="row-left gap-4 p-4">
      <div class="column basis-1/2">
          <PageTimeReport />
          <ProductivityReport />
      </div>
      <div class="column basis-1/2">
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
</style>