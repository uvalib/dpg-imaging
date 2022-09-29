import { createApp, markRaw } from 'vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'

// create pinia store and give all stores access to the router as this.router
const pinia = createPinia()
pinia.use(({ store }) => {
   store.router = markRaw(router)
})

const app = createApp(App)
app.use( router )
app.use( pinia )

// Global component registration. All components can use these without import
import WaitSpinner from "@/components/WaitSpinner.vue"
app.component("WaitSpinner", WaitSpinner)

// Primevue setup
import PrimeVue from 'primevue/config'
import ConfirmationService from 'primevue/confirmationservice'
import ToastService from 'primevue/toastservice'

app.use(PrimeVue)
app.use(ConfirmationService)
app.use(ToastService)

import Button from 'primevue/button'
import ConfirmDialog from 'primevue/confirmdialog'
app.component("DPGButton2", Button)
app.component("ConfirmDialog", ConfirmDialog)

import 'primevue/resources/themes/saga-blue/theme.css'
import 'primevue/resources/primevue.min.css'
import 'primeicons/primeicons.css'

// Styles
import '@fortawesome/fontawesome-free/css/all.css'
import './assets/stylesheets/uva-colors.css'
import './assets/stylesheets/main.scss'
import './assets/stylesheets/styleoverrides.scss'

// actually mount to DOM
app.mount('#app')