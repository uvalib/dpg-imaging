<template>
   <div class="messages">
      <h2>Messages</h2>
      <div class="acts">
         <DPGButton @click="createClicked" class="p-button-secondary" label="Create Message"/>
      </div>
      <TabView>
         <TabPanel header="Inbox">
            <div v-if="messageStore.inbox.length == 0">
               <h3>You have no messages in your inbox</h3>
            </div>
            <DataTable v-else :value="messageStore.inbox" ref="inboxTable" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
               :lazy="false" :paginator="messageStore.inbox.length > 15" :rows="15" :rowsPerPageOptions="[15,30,50]" removableSort
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
            >
               <Column field="read" header=""  class="icon">
                  <template #body="slotProps">
                     <i v-if="slotProps.data.read" class="far fa-envelope-open"></i>
                     <i v-else class="far fa-envelope"></i>
                  </template>
               </Column>
               <Column field="sentAt" header="Date" class="nowrap">
                  <template #body="slotProps">
                     {{formatDate(slotProps.data.sentAt)}}
                  </template>
               </Column>
               <Column field="from" header="From">
                  <template #body="slotProps">
                     <div>{{slotProps.data.from.firstName}} {{slotProps.data.from.lastName}}</div>
                     <div class="email">{{slotProps.data.from.email}}</div>
                  </template>
               </Column>
               <Column field="subject" header="Subject"/>
               <Column field="message" header="Message" class="grow">
                  <template #body="slotProps">
                     {{truncateMessage(slotProps.data.message)}}
                  </template>
               </Column>
               <Column header="">
                  <template #body="slotProps">
                     <DPGButton @click="viewClicked(slotProps.data.id)" class="small p-button-secondary" label="View"/>
                     <DPGButton @click="deleteClicked(slotProps.data.id)" class="small p-button-danger" label="Delete"/>
                  </template>
               </Column>
            </DataTable>
         </TabPanel>
         <TabPanel header="Sent">
            <div v-if="messageStore.sent.length == 0">
               <h3>You have no sent messages</h3>
            </div>
            <DataTable v-else :value="messageStore.sent" ref="inboxTable" dataKey="id"
               stripedRows showGridlines responsiveLayout="scroll" class="p-datatable-sm"
               :lazy="false" :paginator="messageStore.sent.length > 15" :rows="15" :rowsPerPageOptions="[15,30,50]" removableSort
               paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
               currentPageReportTemplate="{first} - {last} of {totalRecords}"
            >
               <Column field="read" header="Read"  class="icon">
                  <template #body="slotProps">
                     <span v-if="slotProps.data.read">Yes</span>
                     <span e-else>No</span>
                  </template>
               </Column>
               <Column field="sentAt" header="Date" class="nowrap">
                  <template #body="slotProps">
                     {{formatDate(slotProps.data.sentAt)}}
                  </template>
               </Column>
               <Column field="to" header="To">
                  <template #body="slotProps">
                     <div>{{slotProps.data.to.firstName}} {{slotProps.data.to.lastName}}</div>
                     <div class="email">{{slotProps.data.to.email}}</div>
                  </template>
               </Column>
               <Column field="subject" header="Subject"/>
               <Column field="message" header="Message" class="grow">
                  <template #body="slotProps">
                     {{truncateMessage(slotProps.data.message)}}
                  </template>
               </Column>
            </DataTable>
         </TabPanel>
      </TabView>
      <Dialog v-model:visible="composeOpen" :modal="true" header="New Message" @hide="cancelMessage" style="width:650px;">
         <div class="row">
            <label>To</label>
            <select v-model="newMessage.to">
               <option disabled :value="0">Select a recipient</option>
               <option v-for="sm in systemStore.staffMembers" :key="`sm${sm.id}`" :value="sm.id">{{sm.lastName}}, {{sm.firstName}}</option>
            </select>
         </div>
         <div class="row">
            <label>Subject</label>
            <input  v-model="newMessage.subject" />
         </div>
         <div class="row">
            <label>Message</label>
            <textarea :rows="10" v-model="newMessage.message"></textarea>
         </div>
         <p class="error">{{error}}</p>
         <template #footer>
            <DPGButton @click="cancelMessage" label="Cancel" class="p-button-secondary right-pad"/>
            <DPGButton @click="sendMessage" label="Send"/>
         </template>
      </Dialog>
   </div>
</template>

<script setup>
import {useMessageStore} from "@/stores/messages"
import {useSystemStore} from "@/stores/system"
import TabView from 'primevue/tabview'
import TabPanel from 'primevue/tabpanel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import Dialog from 'primevue/dialog'
import dayjs from 'dayjs'
import { ref } from 'vue'

const messageStore = useMessageStore()
const systemStore = useSystemStore()

const composeOpen = ref(false)
const newMessage = ref({to: 0, subject: "", message: ""})
const error = ref("")

function createClicked() {
   newMessage.value = {to: 0, subject: "", message: ""}
   composeOpen.value = true
   error.value = ""
}
function cancelMessage() {
   composeOpen.value = false
}
function sendMessage() {
   if (newMessage.value.to == 0) {
      error.value = "Please select a recipient"
      return
   }
   if (newMessage.value.subject == "") {
      error.value = "Please set a subject"
      return
   }
   if (newMessage.value.message == "") {
      error.value = "Please add a message"
      return
   }
   messageStore.sendMessage( newMessage.value )
   composeOpen.value = false
}

function formatDate( date ) {
   return dayjs(date).format("YYYY-MM-DD hh:mm A")
}
function truncateMessage(msg) {
   if (msg.length < 100) return msg
   return msg.slice(0,100)+"..."
}
function viewClicked(msgID) {
   messageStore.viewMessageID = msgID
   messageStore.viewMesage = true
}
function deleteClicked(msgID) {
   messageStore.deleteMessage(msgID)
}

</script>

<style scoped lang="scss">
div.row {
   margin: 0 0 15px 0;
   label {
      font-weight: 500;
      margin-bottom: 5px;
      display: inline-block;
   }
}
p.error {
   color: var(--uvalib-red-emergency);
}
.messages {
   padding: 25px;
   .acts {
      margin: 15px 0;
   }
   .email {
      margin-top: 5px;
      padding-left: 10px;
      font-size: 0.85em;
   }
   :deep(td.icon) {
      text-align: center;
      width: 50px;
      i {
         font-size: 1.25em;
      }
   }
   :deep(td.nowrap) {
      white-space: nowrap;
   }
   :deep(td.grow) {
      width:100%
   }
   button.p-button-secondary.small,button.p-button-danger.small {
      font-size: 0.8em;
      padding: 3px 10px;
      width: 100%;
      margin-bottom: 5px;
   }
}
</style>