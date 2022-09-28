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
import ErrorMessage from "@/components/ErrorMessage.vue"
import DPGButton from "@/components/DPGButton.vue"
import ConfirmModal from "@/components/ConfirmModal.vue"
app.component("WaitSpinner", WaitSpinner)
app.component("ErrorMessage", ErrorMessage)
app.component("DPGButton", DPGButton)
app.component("ConfirmModal", ConfirmModal)

// Styles
import '@fortawesome/fontawesome-free/css/all.css'
import './assets/stylesheets/uva-colors.css'
import './assets/stylesheets/main.scss'

// actually mount to DOM
app.mount('#app')