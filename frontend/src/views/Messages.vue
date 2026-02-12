<template>
   <div class="messages">
      <h2>Messages</h2>
      <DPGButton @click="createClicked" severity="secondary" label="Create Message"/>
      <Tabs value="inbox">
         <TabList>
            <Tab value="inbox">Inbox</Tab>
             <Tab value="sent">Sent</Tab>
         </TabList>
         <TabPanels>
            <TabPanel value="inbox">
               <div v-if="messageStore.inbox.length == 0">
                  <h3>You have no messages in your inbox</h3>
               </div>
               <DataTable v-else :value="messageStore.inbox" ref="inboxTable" dataKey="id"
                   showGridlines responsiveLayout="scroll" class="p-datatable-sm" :rowStyle="rowStyle"
                  :lazy="false" :paginator="messageStore.inbox.length > 15" :rows="15" :rowsPerPageOptions="[15,30,50]" removableSort
                  paginatorTemplate="FirstPageLink PrevPageLink CurrentPageReport NextPageLink LastPageLink RowsPerPageDropdown"
                  currentPageReportTemplate="{first} - {last} of {totalRecords}"
               >
                  <Column field="sentAt" header="Date" class="nowrap">
                     <template #body="slotProps">
                        {{formatDate(slotProps.data.sentAt)}}
                     </template>
                  </Column>
                  <Column field="fromID" header="From">
                     <template #body="slotProps">
                     <div>{{ system.getStaffMemberEmail(slotProps.data.fromID) }}</div>
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
                           <DPGButton @click="viewClicked(slotProps.data.id)" severity="secondary" size="small" label="View"/>
                           <DPGButton @click="deleteClicked(slotProps.data.id)" severity="danger" size="small" label="Delete"/>
                        </div>
                     </template>
                  </Column>
               </DataTable>
            </TabPanel>
            <TabPanel value="sent">
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
                        <div>{{ recipients(slotProps.data.recipients) }}</div>
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
         </TabPanels>
      </Tabs>
   </div>
</template>

<script setup>
import {useSystemStore} from "@/stores/system"
import {useMessageStore} from "@/stores/messages"
import { useUserStore } from "@/stores/user"
import Tabs from 'primevue/tabs'
import TabList from 'primevue/tablist'
import Tab from 'primevue/tab'
import TabPanels from 'primevue/tabpanels'
import TabPanel from 'primevue/tabpanel'
import DataTable from 'primevue/datatable'
import Column from 'primevue/column'
import { useDateFormat } from '@vueuse/core'

const system = useSystemStore()
const messageStore = useMessageStore()
const user = useUserStore()

const rowStyle = (data) => {
   let style = { fontWeight: '100'}
   data.recipients.forEach( r => {
      if (r.staffID == user.ID && r.read == false ) {
         style.fontWeight = 'bold'
      }
   })
   return style
}

const createClicked = (() => {
   messageStore.beginMessageCreate()
})

const recipients = ( list ) => {
   let out = []
   list.forEach( r => {
      out.push( system.getStaffMemberEmail(r.staffID) )
   })
   return out.join("; ")
}

const formatDate =(( date ) => {
   return useDateFormat(date, "YYYY-MM-DD hh:mm A")
})

const truncateMessage = ((msg) => {
   if (msg.length < 100) return msg
   return msg.slice(0,100)+"..."
})

const viewClicked = ((msgID) => {
   messageStore.viewMessage(user.ID, msgID)
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
     flex-flow: row nowrap;
     gap: 5px;
   }
   .email {
      margin-top: 5px;
      padding-left: 10px;
      font-size: 0.85em;
   }
   :deep(th.nowrap), :deep(td.nowrap) {
      white-space: nowrap !important;
   }
}
</style>