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
               <p class="version">v{{ systemStore.version }}</p>
            </div>
         </div>
         <div class="user-banner" v-if="userStore.jwt">
            <div>
               <label>Signed in as:</label><span class="user">{{ userStore.signedInUser }}</span>
            </div>
            <div class="acts">
               <div class="messages">
                  <router-link to="/messages">
                     <span class="cnt">{{messageStore.messages.length}}</span>
                     <i class="mail fas fa-envelope"></i>
                  </router-link>
               </div>
               <span class="signout" @click="signout">Sign out</span>
            </div>
         </div>
      </div>
      <router-view />
      <Dialog v-model:visible="systemStore.showError" :modal="true" header="System Error" @hide="errorClosed()" class="error">
         {{systemStore.error}}
      </Dialog>
      <ScrollToTop />
   </div>
</template>

<script setup>
import UvaLibraryLogo from "@/components/UvaLibraryLogo.vue"
import ScrollToTop from "@/components/ScrollToTop.vue"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import {useMessageStore} from "@/stores/messages"
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'
import Dialog from 'primevue/dialog'

const systemStore = useSystemStore()
const userStore = useUserStore()
const messageStore = useMessageStore()
const router = useRouter()

function errorClosed() {
   systemStore.setError("")
}

function signout() {
   userStore.signout()
   router.push("signedout")
}

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
      font-size: 0.8em;
      margin: 0;
      label {
         font-weight: bold;
         margin-right: 5px;
      }
      .user {
         font-weight: 100;
      }
      .acts {
         margin-top: 10px;
         display: flex;
         flex-flow: row nowrap;
         justify-content: flex-end;
         align-items: center;

         .messages {
            .mail, .cnt {
               display: inline-block;
               margin-right: 10px;
               font-size: 1.5em;
               color: white;
            }
         }
         .signout {
            display: inline-block;
            margin-left: 10px;
            cursor: pointer;
            border: 1px solid var(--uvalib-brand-blue-light);
            padding: 2px 9px;
            border-radius: 3px;
            background: var(--uvalib-brand-blue-light);
            &:hover {
               background: var(--uvalib-brand-blue-lighter);
            }
         }
      }
   }
}

</style>
