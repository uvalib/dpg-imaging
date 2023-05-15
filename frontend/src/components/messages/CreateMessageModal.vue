<template>
   <Dialog v-model:visible="messageStore.showCreateModal" :modal="true" header="New Message" @hide="cancelMessage" style="width:650px;">
      <div class="row">
         <label>To</label>
         <Dropdown id="assigned" v-model="messageStore.newMessage.to" :options="staffMembers"
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
      <template #footer>
         <DPGButton @click="cancelMessage" label="Cancel" class="p-button-secondary right-pad"/>
         <DPGButton @click="sendMessage" label="Send"/>
      </template>
   </Dialog>
</template>

<script setup>
import { useMessageStore } from "@/stores/messages"
import { useSystemStore } from "@/stores/system"
import Dialog from 'primevue/dialog'
import Dropdown from 'primevue/dropdown'
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
div.row {
   margin: 0 0 15px 0;
   label {
      font-weight: 500;
      margin-bottom: 5px;
      display: inline-block;
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
</style>