import { createApp, markRaw } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import './assets/main.css'

const app = createApp(App)

import VueCookies from 'vue3-cookies'
app.use(VueCookies)

// Global component registration. All components can use these without import
import WaitSpinner from "@/components/WaitSpinner.vue"
app.component("WaitSpinner", WaitSpinner)

// NuxtUI defaukted to LIGHT mode
import ui from '@nuxt/ui/vue-plugin'
import { useColorMode } from '@vueuse/core'
useColorMode().value = 'light'
app.use(ui)

// Per some suggestions on vue / pinia git hub issue reports, create and add pinia support LAST
// and use the chained form of the setup. This to avid problems where the vuew dev tools fail to
// include pinia in the tools
app.use( router )

app.use(createPinia().use( ({ store }) => {
   store.router = markRaw(router)
}))

// actually mount to DOM
app.mount('#app')