<template>
   <UApp :toaster="toaster">
      <UHeader mode="slideover">
         <template #title>
            <div class="library-link">
               <a target="_blank" href="https://library.virginia.edu">
                  <UvaLibraryLogo />
               </a>
            </div>
         </template>

         <!-- this is the main menu. shows up in the center of the header if size allows -->
         <UNavigationMenu highlight content-orientation="vertical" :items="menuItems" />

         <template #right>
            <div class="site-link">
               <RouterLink to="/">DPG Imaging</RouterLink>
               <p class="version">{{ systemStore.version }}</p>
            </div>
         </template>

         <template #body>
            <UNavigationMenu :items="menuItems" orientation="vertical" highlight class="-mx-2.5" />
         </template>
      </UHeader>

      <main>
         <router-view v-if="systemStore.initializing==false"/>
         <div v-else style="margin-top:5%">
            <WaitSpinner :overlay="false" message="Initializing sysem..." />
         </div>
      </main>

      <Dialog v-model:visible="systemStore.showError" :modal="true" header="System Error" @hide="errorClosed()" class="error">
         <div style="text-align: left" v-html="systemStore.error"></div>
         <template #footer>
            <DPGButton @click="errorClosed()" label="OK" severity="secondary"/>
         </template>
      </Dialog>
      <MessageModal />
      <CreateMessageModal />
      <ScrollTop />
   </UApp>
</template>

<script setup>
import UvaLibraryLogo from "@/components/UvaLibraryLogo.vue"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import {useMessageStore} from "@/stores/messages"
import { useRouter } from 'vue-router'
import { onMounted, computed } from 'vue'
import Dialog from 'primevue/dialog'
import MessageModal from "./components/messages/MessageModal.vue"
import CreateMessageModal from "./components/messages/CreateMessageModal.vue"
import ScrollTop from 'primevue/scrolltop'

const systemStore = useSystemStore()
const userStore = useUserStore()
const messageStore = useMessageStore()
const router = useRouter()

const toaster = { duration: 5000, position: "top-center" }

const menuItems = computed(() => {
   let menu = [ {label: "Projects", icon: 'i-lucide-house', to: "/"} ]
   if ( userStore.isAdmin || userStore.isSupervisor ) {
      menu.push( {label: "Equipment", icon: 'i-lucide-settings', to: "/equipment"} ) 
      menu.push( {label: "Reports", icon: 'i-lucide-chart-line', to: "/reports"} ) 
   }
   let msgLabel = "Messages"
   if ( messageStore.unreadMessageCount(userStore.ID) > 0 ) {
      msgLabel += ` (${messageStore.unreadMessageCount(userStore.ID)})`
   }
   let userMenu = { label: userStore.signedInUser, icon: "i-lucide-user-round", 
      children: [
         {label: msgLabel, icon: 'i-lucide-mail', to: "/messages"},
         {label: "Sign out", icon: 'i-lucide-log-out',  onSelect: () => signout()} 
      ]
   }
   if ( messageStore.unreadMessageCount(userStore.ID) > 0 ) {
      userMenu.chip = { color: "info"}
   }
   menu.push(userMenu)
   return menu
})

const errorClosed = (() => {
   systemStore.clearError()
})

const messagesClicked = (() => {
   router.push("/messages")
})


const signout = (() => {
   userStore.signout()
   router.push("/signedout")
})

const homeClicked = (() => {
   router.push("/")
})

onMounted( async () => {
   systemStore.getVersion()
   await systemStore.getConfig()
})
</script>

<style lang="scss">

p.version {
   margin: 0;
   font-size: 0.5em;
   text-align: right;
   padding: 0;
}
div.library-link {
   width: 220px;
}
   
div.site-link {
   font-size: 1.3em;
}

</style>
