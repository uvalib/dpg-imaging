import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from '../views/Home.vue'
import Viewer from '../views/Viewer.vue'

Vue.use(VueRouter)

const routes = [
   {
      path: '/',
      name: 'Home',
      component: Home
   },
   {
      path: '/view/:unit',
      name: 'viewer',
      component: Viewer
   },
]

const router = new VueRouter({
   mode: 'history',
   base: process.env.BASE_URL,
   routes
})

export default router
