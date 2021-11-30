
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Project from '../views/Project.vue'
import Units from '../views/Units.vue'    // old list of units in dpg-imaging dir. to be removed
import Unit from '../views/Unit.vue'
import Page from '../views/Page.vue'
import NotFound from '../views/NotFound.vue'
import Forbidden from '../views/Forbidden.vue'
import SignedOut from '../views/SignedOut.vue'
import VueCookies from 'vue-cookies'
import store from '../store'

const routes = [
   {
      path: '/',
      name: 'home',
      component: Home
   },
   {
      path: '/projects/:id',
      name: 'project',
      component: Project
   },
   {  // old list of units in dpg-imaging dir. to be removed
      path: '/units',
      name: 'units',
      component: Units
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
      path: '/forbidden',
      name: 'forbidden',
      component: Forbidden
   },
   {
      path: '/signedout',
      name: 'signedout',
      component: SignedOut
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
   if (to.path === '/granted') {
      let jwtStr = VueCookies.get("dpg_jwt")
      store.commit("setJWT", jwtStr)
      next( "/" )
   } else if (to.name !== 'not_found' && to.name !== 'forbidden' && to.name !== "signedout") {
      let jwtStr = localStorage.getItem('dpg_jwt')
      if ( jwtStr) {
         store.commit("setJWT", jwtStr)
         next()
      } else {
         window.location.href = "/authenticate"
      }
   } else {
      next()
   }
 })

export default router
