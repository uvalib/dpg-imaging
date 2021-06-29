
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Unit from '../views/Unit.vue'
import Page from '../views/Page.vue'
import NotFound from '../views/NotFound.vue'

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
      path: '/:pathMatch(.*)*',
      name: "not_found",
      component: NotFound
   }
]

const router = createRouter({
   history: createWebHistory( process.env.BASE_URL ),
   routes
})

export default router
