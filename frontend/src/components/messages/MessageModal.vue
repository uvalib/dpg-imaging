<template>
   <Dialog v-model:visible="messageStore.viewMesage" :modal="true" header="Message Viewer"  :closable="false" style="width: 650px">
      <dl>
         <dt>Sent:</dt>
         <dd>{{formatDate(message.sentAt)}}</dd>
         <dt>From:</dt>
         <dd>{{message.from.email}}</dd>
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
import {useMessageStore} from '@/stores/messages'
import Dialog from 'primevue/dialog'
import { useDateFormat } from '@vueuse/core'

const messageStore = useMessageStore()

const message = computed( () => {
   return messageStore.inbox.find(m => m.id == messageStore.targetMessageID)
})

const formatDate = (( date ) => {
   return useDateFormat(date, "YYYY-MM-DD hh:mm A")
})

const replyClicked = (() => {
   messageStore.beginReply()
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
      }
   }
</style>
