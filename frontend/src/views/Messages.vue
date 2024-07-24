<template>
   <div class="messages">
      <h2>Messages</h2>
      <DPGButton @click="createClicked" severity="secondary" label="Create Message"/>
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
                     <div class="acts">
                        <DPGButton @click="viewClicked(slotProps.data.id)" severity="secondary" label="View"/>
                        <DPGButton @click="deleteClicked(slotProps.data.id)" severity="danger" label="Delete"/>
                     </div>
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
   </div>
</template>

<script setup>
import {useMessageStore} from "@/stores/messages"
import TabView from 'primevue/tabview'
import TabPanel from 'primevue/tabpanel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import dayjs from 'dayjs'

const messageStore = useMessageStore()

const createClicked = (() => {
   messageStore.beginMessageCreate()
})

const formatDate =(( date ) => {
   return dayjs(date).format("YYYY-MM-DD hh:mm A")
})

const truncateMessage = ((msg) => {
   if (msg.length < 100) return msg
   return msg.slice(0,100)+"..."
})

const viewClicked = ((msgID) => {
   messageStore.targetMessageID = msgID
   messageStore.viewMesage = true
})

const deleteClicked = ((msgID) => {
   messageStore.deleteMessage(msgID)
})
</script>

<style scoped lang="scss">
.messages {
   padding: 25px;
   .acts {
     display: flex;
     flex-direction: column;
     gap: 5px;
     button {
      font-size: 0.85em;
     }
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
}
</style>