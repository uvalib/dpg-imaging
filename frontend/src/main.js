import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'

const app = createApp(App)

// provide store access to the router and router to store
store.router = router
router.store = store

// bind store and router to all componens as $store and $router
app.use(store)
app.use(router)

// Global component registration. All components can use these without import
import WaitSpinner from "@/components/WaitSpinner"
import ErrorMessage from "@/components/ErrorMessage"
import DPGButton from "@/components/DPGButton"
import ConfirmModal from "@/components/ConfirmModal"
app.component("WaitSpinner", WaitSpinner)
app.component("ErrorMessage", ErrorMessage)
app.component("DPGButton", DPGButton)
app.component("ConfirmModal", ConfirmModal)

import '@fortawesome/fontawesome-free/css/all.css'

// actually mount to DOM
app.mount('#app')