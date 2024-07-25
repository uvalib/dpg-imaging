<template>
   <Dialog v-model:visible="messageStore.showCreateModal" :modal="true" header="New Message"
      @hide="cancelMessage" style="width:650px;" :closable="false">
      <div class="content">
         <div class="row">
            <label>To</label>
            <Select id="assigned" v-model="messageStore.newMessage.to" :options="staffMembers"
               optionLabel="name" optionValue="id"
               filter autoFilterFocus resetFilterOnHide filterMatchMode="startsWith"
               placeholder="Select a recipient" />
         </div>
         <div class="row">
            <label>Subject</label>
            <input  v-model="messageStore.newMessage.subject" />
         </div>
         <div class="row">
            <label v-if="messageStore.isResponse">Response</label>
            <label v-else>Message</label>
            <textarea :rows="10" v-model="messageStore.newMessage.message"></textarea>
         </div>
         <div class="row" v-if="messageStore.isResponse">
            <label>Original Message</label>
            <textarea :rows="5" v-model="message.message" readonly disabled></textarea>
         </div>
         <p class="error" v-if="messageStore.error">{{messageStore.error}}</p>
      </div>
      <template #footer>
         <DPGButton @click="cancelMessage" label="Cancel" severity="secondary"/>
         <DPGButton @click="sendMessage" label="Send"/>
      </template>
   </Dialog>
</template>

<script setup>
import { useMessageStore } from "@/stores/messages"
import { useSystemStore } from "@/stores/system"
import Dialog from 'primevue/dialog'
import Select from 'primevue/select'
import { computed } from 'vue'

const messageStore = useMessageStore()
const systemStore = useSystemStore()

const staffMembers = computed( () => {
   let out = []
   systemStore.staffMembers.forEach( sm => {
      out.push( { name: `${sm.lastName}, ${sm.firstName}`, id: sm.id})
   })
   return out
})

const message = computed( () => {
   return messageStore.inbox.find(m => m.id == messageStore.targetMessageID)
})

const cancelMessage = (() => {
   messageStore.cancelMessage()
})

const sendMessage = (() => {
   messageStore.sendMessage()
})
</script>

<style scoped lang="scss">
.content {
   display: flex;
   flex-direction: column;
   row-gap: 15px;;
   div.row {
      label {
         font-weight: 500;
         margin-bottom: 5px;
         display: block;
      }
      div.p-dropdown.p-component  {
         width: 100%;
      }
      input {
         padding: 10px;
      }
   }
   p.error {
      padding: 0;
      margin:0;
      text-align: center;
      color: var(--uvalib-red-emergency);
   }
}
</style>