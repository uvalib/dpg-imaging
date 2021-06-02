import Vue from 'vue'
import VueRouter from 'vue-router'
import Home from '../views/Home.vue'
import Unit from '../views/Unit.vue'
import Page from '../views/Page.vue'
import NotFound from '../views/NotFound.vue'

Vue.use(VueRouter)

const routes = [
   {
      path: '/',
      name: 'Home',
      component: Home
   },
   {
      path: '/unit/:unit',
      name: 'unit',
      component: Unit
   },
   {
      path: '/unit/:unit/page/:page',
      name: 'page',
      component: Page
   },
   {
      path: "*",
      name: "not_found",
      component: NotFound
   }
]

const router = new VueRouter({
   mode: 'history',
   base: process.env.BASE_URL,
   routes
})

export default router
