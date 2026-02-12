<template>
   <Dialog v-model:visible="messageStore.viewMesage" :modal="true" header="Message Viewer"  :closable="false" style="width: 650px">
      <dl>
         <dt>Sent:</dt>
         <dd><pre>{{ formatDate(message.sentAt) }}</pre></dd>
         <dt>From:</dt>
         <dd><pre>{{ system.getStaffMemberEmail( message.fromID ) }}</pre></dd>
         <template v-if="message.recipients.length > 0">
            <dt>To:</dt>
            <dd><pre>{{ recipients }}</pre></dd>
         </template>
         <dt>Subject:</dt>
         <dd>{{message.subject}}</dd>
      </dl>
      <div class="msg">{{message.message}}</div>
      <template #footer>
         <DPGButton @click="replyClicked" label="Reply" severity="secondary"/>
         <DPGButton @click="hide" label="OK" severity="secondary"/>
      </template>
   </Dialog>
</template>

<script setup>
import { computed } from 'vue'
import { useMessageStore } from '@/stores/messages'
import { useSystemStore } from '@/stores/system'
import { useUserStore } from '@/stores/user'
import Dialog from 'primevue/dialog'
import { useDateFormat } from '@vueuse/core'

const messageStore = useMessageStore()
const system = useSystemStore()
const user = useUserStore()

const message = computed( () => {
   return messageStore.inbox.find(m => m.id == messageStore.targetMessageID)
})

const recipients = computed( () => {
   let out = []
   message.value.recipients.forEach(r => {
      let email = system.getStaffMemberEmail(r.staffID)
      out.push( email )
   })
   return out.join("\n")
})

const formatDate = (( date ) => {
   return useDateFormat(date, "YYYY-MM-DD hh:mm A")
})

const replyClicked = (() => {
   messageStore.beginReply( user.ID )
})

const hide =(() => {
   messageStore.markMessageRead()
})
</script>

<style lang="scss" scoped>
   div.msg {
      margin-top: 5px;
      padding: 15px;
      border-top: 1px solid var(--uvalib-grey-light);
      white-space: pre-wrap;
      text-align: left;
   }
   dl {
      margin: 10px 30px 0 30px;
      display: inline-grid;
      grid-template-columns: max-content 2fr;
      grid-column-gap: 10px;
      font-size: 0.9em;
      text-align: left;
      box-sizing: border-box;
      width: 100%;

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
         pre {
            margin: 0;
         }
      }
   }
</style>
