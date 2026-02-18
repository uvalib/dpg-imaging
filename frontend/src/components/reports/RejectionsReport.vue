<template>
   <div class="reports-card">
      <h3>Rejections</h3>
      <div v-if="reportStore.reports.loading" class="wait-wrap">
         <WaitSpinner/>
      </div>
      <div class="report">
         <table class="rejection-stats">
            <tbody>
               <tr>
                  <th class="top-header"></th>
                  <th class="top-header left-bar" colspan="5">Scans</th>
                  <th class="top-header left-bar" colspan="3">QA</th>
               </tr>
               <tr>
                  <th class="sub-header">
                     <span>Staff</span>
                  </th>
                  <th class="left-bar sub-header">
                     <span>Projects</span>
                  </th>
                  <th class="sub-header">
                     <span>Images</span>
                  </th>
                  <th class="sub-header">
                     <span>Rejects</span>
                  </th>
                  <th class="sub-header">
                     <span>Project Rate</span>
                  </th>
                  <th class="sub-header">
                     <span>Image Rate</span>
                  </th>
                  <th class="left-bar sub-header">
                     <span>Projects</span>
                  </th>
                  <th class="sub-header">
                     <span>Rejects</span>
                  </th>
                  <th class="sub-header">
                     <span>Rate</span>
                  </th>
               </tr>
               <tr v-for="r in reportStore.reports.rejections.data" :key="`reject${r.staffID}`">
                  <td class="left">{{ system.getStaffMemberName(r.staffID) }}</td>
                  <td class="left-bar">{{r.scans.projects}}</td>
                  <td>{{r.scans.images}}</td>
                  <td>{{r.scans.rejections}}</td>
                  <td>{{r.scans.projectRate.toFixed(2)}}%</td>
                  <td>{{r.scans.imageRate.toFixed(2)}}%</td>
                  <td class="left-bar">{{r.qa.projects}}</td>
                  <td>{{r.qa.rejections}}</td>
                  <td>{{r.qa.rate.toFixed(2)}}%</td>
               </tr>
            </tbody>
         </table>
         <p class="error" v-if="reportStore.reports.error">{{reportStore.reports.error}}</p>
      </div>
   </div>
</template>

<script setup>
import { useReportStore } from '@/stores/reporting'
import { useSystemStore } from '@/stores/system'
import WaitSpinner from '@/components/WaitSpinner.vue'

const reportStore = useReportStore()
const system = useSystemStore()
</script>

<style lang="scss" scoped>
.reports-card {
   margin: 10px;
   text-align: left;
   border: 1px solid var(--uvalib-grey-light);
   box-shadow: var(--box-shadow-light);
   position: relative;
   min-height: 300px;
   h3 {
      background: var(--uvalib-grey-lightest);
      font-size: 1em;
      text-align: left;
      margin:0;
      padding: 5px 10px;
      border-bottom: 1px solid var(--uvalib-grey-light);
      font-weight: 500;
   }
   .wait-wrap {
      text-align: center;
      padding: 30px 0 ;
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background-color:rgba(255,255,255, 0.6);
      z-index: 1000;
      div.spinner {
         margin-top: 25%;
      }
   }
   .report {
      padding: 10px;
   }
   .rejection-stats {
      width: 100%;
      font-size: 0.8em;
      margin-top: 0;
      background: white;
      border-top: none;
      border-collapse: collapse;
      border:1px solid var(--uvalib-grey-light);
      th {
         font-weight: normal;
         background: var(--uvalib-grey-lightest);
         padding: 5px;
         text-align: center;
      }
      th.sub-header {
         border-bottom: 1px solid var(--uvalib-grey-light);
      }
      td {
         padding: 5px;
         text-align: center;
      }
      td.left {
         text-align: left;
      }
      .top-header {
         border: none;
         text-align: center;
         margin-bottom: 0;
         padding-bottom: 0;
      }
      .left-bar {
         border-left:1px solid var(--uvalib-grey-light);

      }
   }
}
</style>

