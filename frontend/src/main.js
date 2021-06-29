import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'

import WaitSpinner from "@/components/WaitSpinner"
import ErrorMessage from "@/components/ErrorMessage"
import DPGButton from "@/components/DPGButton"
import ConfirmModal from "@/components/ConfirmModal"

const app = createApp(App)
app.use(store)
app.use(router)

app.component("WaitSpinner", WaitSpinner)
app.component("ErrorMessage", ErrorMessage)
app.component("DPGButton", DPGButton)
app.component("ConfirmModal", ConfirmModal)



import '@fortawesome/fontawesome-free/css/all.css'

// actually mount to DOM
app.mount('#app')