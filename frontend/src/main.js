import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'

import WaitSpinner from "@/components/WaitSpinner"
Vue.component('WaitSpinner', WaitSpinner)

import ErrorMessage from "@/components/ErrorMessage"
Vue.component('ErrorMessage', ErrorMessage)

import DPGButton from "@/components/DPGButton"
Vue.component('DPGButton', DPGButton)

import ConfirmModal from "@/components/ConfirmModal"
Vue.component('ConfirmModal', ConfirmModal)

import '@fortawesome/fontawesome-free/css/all.css'

Vue.config.productionTip = false

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
