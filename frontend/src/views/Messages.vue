<template>
   <div class="messages">
      <h2>
         {{userStore.firstName}} {{userStore.lastName}} Messages
      </h2>
   </div>
</template>

<script setup>
import {useUserStore} from "@/stores/user"
import {useMessageStore} from "@/stores/messages"

const userStore = useUserStore()
const messageStore = useMessageStore()

</script>

<style scoped lang="scss">
   .messages {
      padding: 25px;
   }
</style>