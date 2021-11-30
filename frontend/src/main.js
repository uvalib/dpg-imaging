import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'

const app = createApp(App)

// bind store and router to all componens as $store and $router
app.use(store)
app.use(router)

// Global component registration. All components can use these without import
import WaitSpinner from "@/components/WaitSpinner"
import ErrorMessage from "@/components/ErrorMessage"
import DPGButton from "@/components/DPGButton"
import DPGPagination from '@/components/DPGPagination.vue'
import ConfirmModal from "@/components/ConfirmModal"
app.component("WaitSpinner", WaitSpinner)
app.component("ErrorMessage", ErrorMessage)
app.component("DPGButton", DPGButton)
app.component("DPGPagination", DPGPagination)
app.component("ConfirmModal", ConfirmModal)

import '@fortawesome/fontawesome-free/css/all.css'

// actually mount to DOM
app.mount('#app')