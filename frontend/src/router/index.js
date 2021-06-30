
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Unit from '../views/Unit.vue'
import Page from '../views/Page.vue'
import NotFound from '../views/NotFound.vue'
import Forbidden from '../views/Forbidden.vue'
import VueCookies from 'vue-cookies'
import store from '../store'

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
      // NOTES: the granted route doesn't have a visual representation. It just stores token and redirects
      path: '/granted',
      beforeEnter: async (_to, _from, next) => {
         let jwtStr = VueCookies.get("dpg_jwt")
         store.commit("setJWT", jwtStr)
         next( "/" )
      }
   },
   {
      path: '/forbidden',
      name: 'forbidden',
      component: Forbidden
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

router.beforeEach((to, _from, next) => {
   if (to.path !== '/granted') {
      let jwtStr = localStorage.getItem('dpg_jwt')
      if ( jwtStr) {
         store.commit("setJWT", jwtStr)
         next()
      } else {
         window.location.href = "/authenticate"
      }
   }
   else {
      next()
   }
 })

export default router
