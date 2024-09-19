<template>
   <div id="app">
      <div class="header" role="banner" id="uva-header">
         <div class="main-header">
            <div class="library-link">
               <a target="_blank" href="https://library.virginia.edu">
                  <UvaLibraryLogo />
               </a>
            </div>
            <div class="site-link">
               <router-link to="/">DPG Imaging</router-link>
               <p class="version">{{ systemStore.version }}</p>
            </div>
         </div>
         <div class="user-banner" v-if="userStore.jwt">
            <div class="acts">
               <span class="signed-in-as">{{ userStore.signedInUser }}</span>
               <DPGButton icon="pi pi-envelope" @click="messagesClicked" :label="`${messageStore.unreadMessageCount}`" size="small"/>
               <DPGButton icon="pi pi-sign-out" @click="signout" label="Sign out" size="small"/>
            </div>
         </div>
      </div>
      <router-view />
      <Dialog v-model:visible="systemStore.showError" :modal="true" header="System Error" @hide="errorClosed()" class="error">
         {{systemStore.error}}
         <template #footer>
            <DPGButton @click="errorClosed()" label="OK" severity="secondary"/>
         </template>
      </Dialog>
      <MessageModal />
      <CreateMessageModal />
      <ScrollTop />
   </div>
</template>

<script setup>
import UvaLibraryLogo from "@/components/UvaLibraryLogo.vue"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import {useMessageStore} from "@/stores/messages"
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'
import Dialog from 'primevue/dialog'
import MessageModal from "./components/messages/MessageModal.vue"
import CreateMessageModal from "./components/messages/CreateMessageModal.vue"
import ScrollTop from 'primevue/scrolltop'

const systemStore = useSystemStore()
const userStore = useUserStore()
const messageStore = useMessageStore()
const router = useRouter()

const errorClosed = (() => {
   systemStore.setError("")
})

const messagesClicked = (() => {
   router.push("/messages")
})

const signout = (() => {
   userStore.signout()
   router.push("/signedout")
})

onMounted( async () => {
   systemStore.getVersion()
   await systemStore.getConfig()
})
</script>

<style lang="scss">
div.header {
   background-color: var(--uvalib-brand-blue);
   color: white;
   padding: 1vw 20px;
   text-align: left;
   position: relative;
   box-sizing: border-box;

   p.version {
      margin: 5px 0 0 0;
      font-size: 0.5em;
      text-align: right;
      padding: 0;
      opacity: 0.5;
   }
   div.library-link {
      width: 220px;
      order: 0;
      flex: 0 1 auto;
      align-self: flex-start;
   }
   div.site-link {
      border: 0;
      font-size: 1.5em;
      a {
         color: white !important;
         text-decoration: none;
         &:hover {
            text-decoration: underline;
         }
      }
   }

   .main-header {
      display: flex;
      flex-direction: row;
      flex-wrap: nowrap;
      justify-content: space-between;
      align-content: stretch;
      align-items: center;
   }
   .user-banner {
      text-align: right;
      padding: 10px 0 0 0;
      font-size: 0.9em;
      margin: 0;

      .acts {
         margin: 0;
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-end;
         align-items: center;
         gap: 10px;

         .signed-in-as {
            display: inline-block;
         }
      }
   }
}

</style>
