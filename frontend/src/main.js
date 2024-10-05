import { createApp, markRaw } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import 'primeicons/primeicons.css'
import './assets/stylesheets/main.scss'

const app = createApp(App)

// Global component registration. All components can use these without import
import WaitSpinner from "@/components/WaitSpinner.vue"
app.component("WaitSpinner", WaitSpinner)

// Primevue setup
import PrimeVue from 'primevue/config'
import UVA from './assets/theme/uva'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'
import Tooltip from 'primevue/tooltip';

app.use(PrimeVue, {
   theme: {
      preset: UVA,
      options: {
         prefix: 'p',
         darkModeSelector: '.dpg-dark'
      }
   }
})

app.directive('tooltip', Tooltip)
app.use(ConfirmationService)
app.use(ToastService)

import Button from 'primevue/button'
import ConfirmDialog from 'primevue/confirmdialog'
app.component("DPGButton", Button)
app.component("ConfirmDialog", ConfirmDialog)

app.use( router )
app.use(createPinia().use( ({ store }) => {
   store.router = markRaw(router)
}))

// actually mount to DOM
app.mount('#app')