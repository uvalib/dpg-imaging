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
         <UNavigationMenu v-if="userStore.isSignedIn" highlight content-orientation="vertical" :items="menuItems" />

         <template #right>
            <div class="site-link">
               <RouterLink to="/">DPG Imaging</RouterLink>
               <p class="version">{{ systemStore.version }}</p>
            </div>
         </template>

         <template v-if="userStore.isSignedIn"  #body>
            <UNavigationMenu :items="menuItems" orientation="vertical" highlight class="-mx-2.5" />
         </template>
         
         <!-- This adds a section below the left/default/right. 
              The theme needs to be update to make header root h-auto to -include it in the uverall height -->
         <template v-if="route.path == '/'" #bottom>
            <div class="toolbar">
               <URadioGroup orientation="horizontal" size="lg" color="info" v-model="searchStore.filter" value-key="id" :items="filters" @update:modelValue="filterChanged"/>
               <div class="page-ctl" v-if="!searchStore.working && searchStore.projects.length>0">
                  <DPGPagination :currPage="searchStore.currPage" :pageSize="searchStore.pageSize" :totalPages="searchStore.totalPages"
                     @next="nextClicked" @prior="priorClicked" @first="firstClicked" @last="lastClicked"
                     @jump="pageJumpClicked"
                  />
               </div>
            </div>
         </template> 
      </UHeader>

      <UMain>
         <router-view v-if="systemStore.initializing==false"/>
      </UMain>

      <div v-if="systemStore.initializing" style="margin-top:5%">
         <WaitSpinner :overlay="true" message="Initializing sysem..." />
      </div>

      <UModal v-model:open="systemStore.showError" :modal="true" :dismissible="false" title="System Error">
         <template #body>
            <div style="text-align: left" v-html="systemStore.error"></div>
         </template>
      </UModal>
      <MessageModal />
      <CreateMessageModal />
   </UApp>
</template>

<script setup>
import UvaLibraryLogo from "@/components/UvaLibraryLogo.vue"
import {useSystemStore} from "@/stores/system"
import {useUserStore} from "@/stores/user"
import {useMessageStore} from "@/stores/messages"
import { useSearchStore } from "@/stores/search"
import { useRouter, useRoute } from 'vue-router'
import { onMounted, computed } from 'vue'
import MessageModal from "./components/messages/MessageModal.vue"
import CreateMessageModal from "./components/messages/CreateMessageModal.vue"

const systemStore = useSystemStore()
const userStore = useUserStore()
const messageStore = useMessageStore()
const searchStore = useSearchStore()
const router = useRouter()
const route = useRoute()

const filters = computed( () => {
   const out = [
      {id: "me", label: `Assigned to me (${searchStore.totals.me})`},
      {id: "active", label: `Active (${searchStore.totals.active})`},
      {id: "errors", label: `Problems (${searchStore.totals.errors})`},
      {id: "unassigned", label: `Unassigned (${searchStore.totals.unassigned})`},
      {id: "finished", label: `Finished (${searchStore.totals.finished})`},
   ]
   return out
})

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


const filterChanged = ( async () => {
   searchStore.filterChanged()
   let query = Object.assign({}, route.query)
   query.filter = searchStore.filter
   await router.push({query})
   searchStore.lastSearchURL = route.fullPath
   searchStore.getProjects()
})
const nextClicked = (() => {
   searchStore.setCurrentPage(searchStore.currPage+1 )
})

const priorClicked = (() => {
   searchStore.setCurrentPage(searchStore.currPage-1 )
})
const firstClicked = (() => {
   searchStore.setCurrentPage( 1 )
})

const lastClicked = (() => {
   searchStore.setCurrentPage(searchStore.totalPages )
})

const pageJumpClicked = ((p) => {
   searchStore.setCurrentPage( p )
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

.toolbar {
   padding: 5px 10px;
   background: var(--uvalib-grey-lightest);
   display: flex;
   flex-flow: row;
   justify-content: space-between;
   align-items: center;
   min-height: 50px;
}

</style>
