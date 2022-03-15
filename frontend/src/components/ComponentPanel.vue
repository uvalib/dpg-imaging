<template>
   <div class="component panel">
      <h3>Component Link</h3>
      <div class="content">
         <span class="entry pad-right">
            <label>Start Image:</label>
            <select id="start-page" v-model="unitStore.rangeStartIdx">
               <option disabled :value="-1">Select start page</option>
               <option v-for="(mf,idx) in unitStore.masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
         <span class="entry pad-right">
            <label>End Image:</label>
            <select id="end-page" v-model="unitStore.rangeEndIdx">
               <option disabled :value="-1">Select end page</option>
               <option v-for="(mf,idx) in unitStore.masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
         <span class="entry">
            <label>Component ID:</label>
            <input id="component-id" type="text" v-model="componentID"  @keyup.enter="okClicked"/>
         </span>
      </div>
      <div class="panel-actions">
         <DPGButton @clicked="selectAllClicked" class="left">Select All</DPGButton>
         <DPGButton @clicked="unlinkClicked" class="right-pad">Unlink</DPGButton>
         <DPGButton @clicked="cancelEditClicked" class="right-pad">Cancel</DPGButton>
         <DPGButton @clicked="okClicked">OK</DPGButton>
      </div>
      <div class="component-dimmer" v-if="component.valid">
         <div role="dialog" aria-labelledby="component-modal-title" id="component-modal" class="component-modal">
            <div id="component-modal-title" class="component-modal-title">Confirm Component Link</div>
            <div class="component-modal-content">
               <table>
                  <tr>
                     <td class="label">Title:</td>
                     <td class="data">{{formatData(component.title)}}</td>
                  </tr>
                  <tr>
                     <td class="label">Label:</td>
                     <td class="data">{{formatData(component.label)}}</td>
                  </tr>
                  <tr>
                     <td class="label">Description:</td>
                     <td class="data">{{formatData(component.description)}}</td>
                  </tr>
                  <tr>
                     <td class="label">Date:</td>
                     <td class="data">{{formatData(component.date)}}</td>
                  </tr>
                  <tr>
                     <td class="label">Type:</td>
                     <td class="data">{{formatData(component.type)}}</td>
                  </tr>
               </table>
               <p class="confirm">Link this component to selected images?</p>
            </div>
            <div class="component-modal-controls">
               <DPGButton id="close-confirm" @clicked="noLinkClicked" @tabback="setFocus('ok-confirm')" :focusBackOverride="true">
                  No
               </DPGButton>
               <span class="spacer"></span>
               <DPGButton id="ok-confirm" @clicked="linkConfirmed" @tabnext="setFocus('close-confirm')" :focusNextOverride="true">
                  Yes
               </DPGButton>
            </div>
         </div>
      </div>
   </div>
</template>

<script setup>
import {useUnitStore} from "@/stores/unit"
import {useSystemStore} from "@/stores/system"
import { ref } from 'vue'

const unitStore = useUnitStore()
const systemStore = useSystemStore()

const componentID = ref("")

function formatData( value ) {
   if (value && value != "" )  return value
   return "N/A"
}
function okClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.error = "Start and end image must be selected"
      return
   }
   if (this.componentID == "") {
      systemStore.error = "Component ID is required"
      return
   }
   unitStore.lookupComponentID(componentID.value)
}
function noLinkClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
}
function cancelEditClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
   unitStore.editMode = ""
}
async function unlinkClicked() {
   systemStore.error = ""
   unitStore.clearComponent()
   if ( unitStore.rangeStartIdx == -1 || unitStore.rangeEndIdx == -1) {
      systemStore.error = "Start and end image must be selected"
      return
   }
   await unitStore.componentLink("")
   cancelEditClicked()
}
async function linkConfirmed() {
   await unitStore.componentLink(componentID.value)
   cancelEditClicked()
}
function selectAllClicked() {
   unitStore.selectAll()
}
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   border-bottom: 1px solid var(--uvalib-grey);
   h3 {
      margin: 0;
      padding: 8px 0;
      font-size: 1em;
      background: var(--uvalib-blue-alt-light);
      border-bottom: 1px solid var(--uvalib-grey);
      font-weight: 500;
   }
   .panel-actions {
      padding: 5px 0 20px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
      width: 50%;
      margin: 0 auto;
      .button {
         margin-left: 10px;
      }
      .left {
         margin-right: auto;
      }
   }
   .content {
      padding: 10px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      width: 50%;
      margin: 0 auto;
      .entry {
         flex-grow: 1;
         margin: 0;
         text-align: left;
         label {
            display: block;
            margin: 0 0 5px 0;
         }
      }
      .entry.pad-right {
         padding-right: 25px;
      }
   }
   .error {
      font-style: italic;
      color: var(--uvalib-red-emergency);
      margin: 0;
   }
   .component-dimmer {
      position: fixed;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      z-index: 1000;
      background: rgba(0, 0, 0, 0.2);
      div.component-modal {
         color: var(--uvalib-text);
         position: fixed;
         height: auto;
         z-index: 8000;
         background: white;
         top: 30%;
         left: 50%;
         transform: translate(-50%, -50%);
         box-shadow: var(--box-shadow);
         border-radius: 5px;
         min-width: 450px;
         max-width: 50%;
         border: 1px solid var(--uvalib-grey);
         div.component-modal-title {
            background:  var(--uvalib-blue-alt-light);
            font-size: 1.1em;
            color: var(--uvalib-text-dark);
            font-weight: 500;
            padding: 10px;
            border-radius: 5px 5px 0 0;
            border-bottom: 2px solid  var(--uvalib-blue-alt);
            text-align: left;
         }
         .component-modal-content {
            font-size: 0.9em;
            padding: 15px 15px 0 15px;
            .confirm {
               text-align: right;
            }
            td {
               padding: 3px;
               text-align: left;
            }
            td.label {
               font-weight: bold;
               text-align: right;
            }
         }
         div.component-modal-controls {
            padding: 10px 20px 20px 20px;
            font-size: 0.9em;
            margin: 0;
            display: flex;
            flex-flow: row wrap;
            justify-content: flex-end;
            .spacer {
               display: inline-block;
               margin: 0 5px;
            }
         }
      }
   }
}
</style>
