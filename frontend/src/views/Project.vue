<template>
   <div class="project">
      <h2>
         <span>Digitization Project #{{$route.params.id}}</span>
         <span v-if="working == false" class="due">
            <label>Due:</label><span>{{currProject.dueOn.split("T")[0]}}</span>
         </span>
      </h2>
      <WaitSpinner v-if="working" :overlay="true" :message="`Loading project ${$route.params.id}...`" />
      <template v-else>
         <div class="project-head">
            <h3>
               <a target="_blank" :href="metadataLink">{{currProject.unit.metadata.title}}</a>
            </h3>
            <h4 class="proj-data">
               <div class="column right-pad">
                  <div>
                     <label>Unit:</label>
                     <a target="_blank" :href="`${adminURL}/units/${currProject.unit.id}`">{{currProject.unit.id}}</a>
                  </div>
                  <div>
                     <label>Order:</label>
                     <a target="_blank" :href="`${adminURL}/orders/${currProject.unit.orderID}`">{{currProject.unit.orderID}}</a>
                  </div>
               </div>
               <div class="column">
                  <div>
                     <label>Customer:</label>
                     <a target="_blank" :href="`${adminURL}/customers/${currProject.unit.order.customer.id}`">
                        {{currProject.unit.order.customer.firstName}} {{currProject.unit.order.customer.lastName}}
                     </a>
                  </div>
                  <div>
                     <label>Intended Use:</label>
                     <span class="data">{{currProject.unit.intendedUse.description}}</span>
                  </div>
               </div>
            </h4>
            <span class="back">
               <i class="fas fa-angle-double-left back-button"></i>
               <router-link to="/">Back to Projects</router-link>
            </span>
         </div>
         <div class="project-main">
            <div class="info-block">
               <h4>Item Information</h4>
               <dl>
                  <dt>Category:</dt>
                  <dd>{{currProject.category.name}}
                  </dd>
                  <dt>Call Number:</dt>
                  <dd>{{currProject.unit.metadata.callNumber}}</dd>
                  <dt>Special Instructions:</dt>
                  <dd>
                     <span v-if="currProject.unit.specialInstructions">{{currProject.unit.specialInstructions}}</span>
                     <span v-else class="na">EMPTY</span>
                  </dd>
                  <dt>Condition:</dt>
                  <dd>{{conditionText(currProject.ItemCondition)}}</dd>
                  <dt>Condition Notes:</dt>
                  <dd>
                     <span v-if="currProject.conditionNote">{{currProject.conditionNote}}</span>
                     <span v-else class="na">EMPTY</span>
                  </dd>
               </dl>
            </div>
            <div class="info-block">
               <h4>Equipment</h4>
            </div>
            <div class="info-block">
               <h4>OCR Settings</h4>
               <dl>
                  <dt>OCR Hint:</dt>
                  <dd>
                     <span v-if="currProject.unit.metadata.ocrHint.id > 0">{{currProject.unit.metadata.ocrHint.name}}</span>
                     <span v-else class="na">EMPTY</span>
                  </dd>
                  <dt>OCR Language Hint:</dt>
                  <dd>
                     <span v-if="currProject.unit.metadata.ocrLanguageHint">{{currProject.unit.metadata.ocrLanguageHint}}</span>
                     <span v-else class="na">EMPTY</span>
                  </dd>
                  <dt>OCR Master Files:</dt>
                  <dd>
                     <span v-if="currProject.unit.ocrMasterFiles" class="yes-no">Yes</span>
                     <span v-else class="yes-no">No</span>
                  </dd>
               </dl>
            </div>
            <div class="info-block">
               <Workflow />
            </div>
            <div class="info-block">
               <h4>History</h4>
               <div class="timing">
                  <span>
                     <label>Date started:</label>
                     <span v-if="currProject.startedAt">{{formatDate(currProject.startedAt)}}</span>
                     <span v-else class="na">N/A</span>
                  </span>
                  <span>
                     <label>Total work time:</label>
                     <span v-if="currProject.startedAt">{{totalWorkTime}}</span>
                     <span v-else class="na">N/A</span>
                  </span>
               </div>
               <div class="history-wrap">
                  <table class="history">
                     <tr>
                        <th>Date</th><th>Step</th><th>Activity</th><th>Owner</th>
                     </tr>
                     <template v-for="a in currProject.assignments" :key="`a${a.id}`">
                        <template v-if="lookupStepName(a.stepID) != 'Unknown'">
                           <template v-if="a.finishedAt">
                              <!-- status: [:pending, :started, :finished, :rejected, :error, :reassigned, :finalizing] -->
                              <tr :class="{success: a.status != 3, reject: a.status==3}">
                                 <td>{{formatDate(a.finishedAt)}}</td>
                                 <td>{{lookupStepName(a.stepID)}}</td>
                                 <td>
                                    <span>
                                       <template v-if="a.status==3">Rejected</template>
                                       <template v-else>Finished</template>
                                       <template v-if="a.status != 5"> <!-- not reassigned -->
                                          <br/>{{a.durationMinutes}} mins
                                       </template>
                                    </span>
                                 </td>
                                 <td>{{a.staffMember.firstName}} {{a.staffMember.lastName}}</td>
                              </tr>
                           </template>
                           <template v-if="a.startedAt">
                              <tr :class="{error: a.status == 4, finalize: a.status == 6}">
                                 <td>{{formatDate(a.startedAt)}}</td>
                                 <td>{{lookupStepName(a.stepID)}}</td>
                                 <td v-if="a.status == 4">Error</td>
                                 <td v-else-if="a.status == 6">Finalizing...</td>
                                 <td v-else>Started</td>
                                 <td>{{a.staffMember.firstName}} {{a.staffMember.lastName}}</td>
                              </tr>
                           </template>

                           <tr>
                              <td>{{formatDate(a.assignedAt)}}</td>
                              <td>{{lookupStepName(a.stepID)}}</td>
                              <td>Assigned</td>
                              <td>{{a.staffMember.firstName}} {{a.staffMember.lastName}}</td>
                           </tr>

                        </template>
                     </template>
                     <tr class="create">
                        <td>{{formatDate(currProject.addedAt)}}</td>
                        <td>Project #{{currProject.id}}</td>
                        <td>Created</td>
                        <td></td>
                     </tr>
                  </table>
               </div>
            </div>
            <div class="info-block">
               <h4>Notes</h4>
            </div>
         </div>
      </template>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
import Workflow from "@/components/project/Workflow"
import date from 'date-and-time'
export default {
   name: "project",
   components: {
      Workflow
   },
   computed: {
      ...mapState({
         working : state => state.projects.working,
         adminURL: state => state.adminURL,
         scanDir: state => state.scanDir,
         qaDir: state => state.qaDir,
         selectedProjectIdx: state => state.projects.selectedProjectIdx,
         computingID: state => state.user.computeID
      }),
      ...mapGetters({
         currProject: 'projects/currProject',
         isOwner: 'projects/isOwner',
         isAdmin: 'isAdmin',
         isSupervisor: 'isSupervisor',
      }),
      metadataLink() {
         let mdType = "sirsi_metadata"
         if (this.currProject.unit.metadata.type == "XmlMetadata") {
            mdType = "xml_metadata"
         }
         return `${this.adminURL}/${mdType}/${this.currProject.unit.metadata.id}`
      },
      totalWorkTime() {
         let mins = 0
         this.currProject.assignments.forEach( a => {
            console.log(a.durationMinutes)
            mins += a.durationMinutes
         })
         console.log("totalMins "+mins)
         let h = mins/60
         mins -= (h*60)
         return `${(""+h).padStart(2,"0")}:${(""+mins).padStart(2,"0")}`
      }
   },
   methods: {
      lookupStepName( stepID) {
         let s = this.currProject.workflow.steps.find( s => s.id == stepID)
         if (s) {
            return s.name
         }
         return "Unknown"
      },
      formatDate( d ) {
         return date.format(new Date(d), "YYYY-MM-DD hh:mm A")
      },
      conditionText(condID) {
         if (condID == 0) return "Good"
         return "Bad"
      }
   },
  async beforeMount() {
      if (this.selectedProjectIdx == -1) {
         console.log("REQ PROJ")
         await this.$store.dispatch("projects/getProject", this.$route.params.id)
         console.log("DONE PROJ")
      }
   },
};
</script>

<style scoped lang="scss">
.project {
   position: relative;
   padding: 25px;
   h2 {
      color: var(--uvalib-brand-orange);
      margin-bottom: 25px;
      position: relative;
      .due {
         position: absolute;
         right: 0;
         color: var(--uvalib-text);
         font-size: 16px;
         font-weight: 500;
         background: var(--uvalib-blue-alt-light);
         border: 1px solid var(--uvalib-blue-alt);
         padding: 5px 15px;
      }
   }
   label {
      font-weight: bold;
      margin-right: 5px;
   }

   .project-head {
      color: var(--uvalib-text);
      padding-bottom: 15px;
      border-bottom: 1px solid var(--uvalib-grey-light);
      position: relative;
      margin-bottom: 20px;
      h3  {
         max-width: 90%;
         text-align: center;
         font-weight: 500;
         font-size: 1.25em;
         margin: 5px auto 10px auto;
         .icon {
            display: inline-block;
            margin-left: 10px;
         }
      }
      h4 {
         font-size: 0.9em;
         display: flex;
         flex-flow: row nowrap;
         justify-content: center;
         .column {
            text-align: left;
         }
         .column.right-pad {
             margin-right: 25px;
         }
         label {
            margin-right: 10px;
            width: 100px;
            display: inline-block;
            text-align: right;
         }
         .data {
            font-weight: 500;
         }
         margin: 5px 0;
         div {
            margin: 5px 0;
         }
      }
      .back {
         position: absolute;
         left: 0px;
         bottom: 10px;
         a {
            font-weight: normal;
            text-decoration: none;
            color: var(--uvalib-text);
            display: inline-block;
            margin-left: 5px;
            &:hover {
               text-decoration: underline ;
            }
         }
      }
   }
   .project-main {
      display: flex;
      flex-flow: row wrap;
      justify-content: center;

      .info-block {
         width: 46%;
         min-width: 600px;
         border: 1px solid var(--uvalib-grey-light);
         margin: 15px;
         display: inline-block;
         min-height: 100px;
         box-shadow: 0 1px 3px rgba(0,0,0,.06), 0 1px 2px rgba(0,0,0,.12);
         text-align: left;
         .na {
            color: #999;
         }

         .timing {
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;
            padding: 10px 10px 0 10px;
            font-size: 0.8em;
            border: none;
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

            tr.create td {
               background: var(--uvalib-blue-alt-light);
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
            }
            tr.finalize {
               background-color: #5a5;
               color: white;
            }
         }
         h4 {
            text-align: center;
            color: var(--uvalib-text);
            font-size: 1em;
            margin: 0;
            padding: 5px;
            background: var(--uvalib-grey-lightest);
            border-bottom: 1px solid var(--uvalib-grey-light);
         }
         dl {
            margin: 10px 30px 0 30px;
            display: inline-grid;
            grid-template-columns: max-content 2fr;
            grid-column-gap: 10px;
            font-size: 0.9em;
            text-align: left;
            box-sizing: border-box;

            dt {
               font-weight: bold;
               text-align: right;
            }
            dd {
               margin: 0 0 10px 0;
               word-break: break-word;
               -webkit-hyphens: auto;
               -moz-hyphens: auto;
               hyphens: auto;
            }
         }
      }
   }
}
</style>
