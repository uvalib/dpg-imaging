<template>
   <div class="messages">
      <h2>Messages</h2>
      <div style="text-align: center;margin:15px 0">
         <UButton label="Create Message" color="secondary" @click="createClicked" />
      </div>
      <UTabs :items="tabs" variant="link" >
         <template #inbox>
            <div v-if="messageStore.inbox.length == 0">
               <h3>You have no messages in your inbox</h3>
            </div>
            <UTable v-else :data="inboxData" :columns="inboxCols" v-model:column-visibility="columnVisibility">
               <template #actions-cell="{ row }">
                  <UDropdownMenu :items="getInboxActions(row.original)">
                     <UButton
                        icon="i-lucide-ellipsis-vertical"
                        color="secondary"
                        aria-label="Actions"
                     />
                  </UDropdownMenu>
               </template>
            </UTable>
         </template>
         <template #sent>
            <div v-if="messageStore.sent.length == 0">
               <h3>You have no sent messages</h3>
            </div>
            <UTable v-else :data="sentData"  v-model:column-visibility="columnVisibility" />
         </template>
      </UTabs>
   </div>
</template>

<script setup>
import {useSystemStore} from "@/stores/system"
import {useMessageStore} from "@/stores/messages"
import { useUserStore } from "@/stores/user"
import { useDateFormat } from '@vueuse/core'
import { computed, ref } from "vue"

const system = useSystemStore()
const messageStore = useMessageStore()
const user = useUserStore()

const tabs = [
  {
    label: 'Inbox',
    slot: 'inbox'
  },
  {
    label: 'Sent',
    slot: 'sent'
  }
]

const sentData = computed(() => {
   let out = []
   messageStore.sent.forEach( m => {
      let read = "No"
      if  (m.read ) read = "Yes"
      out.push({id: m.id, read: read, date: formatDate(m.sentAt), to: recipients(m.recipients), 
         subject: m.subject, message: truncateMessage(m.message) })
   })
   return out
})
const recipients = ( list ) => {
   let out = []
   list.forEach( r => {
      out.push( system.getStaffMemberEmail(r.staffID) )
   })
   return out.join("; ")
}

const inboxData = computed(() => {
   let out = []
   messageStore.inbox.forEach( m => {
      let read = "No"
      m.recipients.some( r => {
         if (r.staffID == user.ID && r.read ) {
            read = "Yes"
         }
         return read == "Yes"
      })
      out.push({id: m.id, read: read, date: formatDate(m.sentAt), from: system.getStaffMemberEmail(m.fromID), 
         subject: m.subject, message: truncateMessage(m.message), recipients: m.recipients })
   })
   return out
})
const inboxCols = [
   {
      accessorKey: 'id',
   },
   {
      accessorKey: 'read',
      header: 'Read'
   },
   {
      accessorKey: 'date',
      header: 'Date'
   },
   {
      accessorKey: 'from',
      header: 'From'
   },
   {
      accessorKey: 'subject',
      header: 'Subject'
   },
   {
      accessorKey: 'message',
      header: 'Message'
   },
   {
      id: 'actions'
   }
]

const columnVisibility = ref({
  id: false, 
  recipients: false
})

const getInboxActions = ((msg) => {
   return [
      [
         {
            label: 'View',
            icon: 'i-lucide-view',
            onSelect: () => messageStore.viewMessage(user.ID, msg.id)
         }
      ],
      [
         {
            label: 'Delete',
            icon: 'i-lucide-trash',
            onSelect: () => messageStore.deleteMessage(msg.id)
         }
      ]
   ]
})

const createClicked = (() => {
   messageStore.beginMessageCreate()
})

const formatDate =(( date ) => {
   return useDateFormat(date, "YYYY-MM-DD hh:mm A")
})

const truncateMessage = ((msg) => {
   if (msg.length < 100) return msg
   return msg.slice(0,100)+"..."
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