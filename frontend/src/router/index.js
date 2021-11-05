
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Project from '../views/Project.vue'
import Units from '../views/Units.vue'    // old list of units in dpg-imaging dir. to be removed
import Unit from '../views/Unit.vue'
import Page from '../views/Page.vue'
import NotFound from '../views/NotFound.vue'
import Forbidden from '../views/Forbidden.vue'
import VueCookies from 'vue-cookies'

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
      router.store.commit("setJWT", jwtStr)
      next( "/" )
   } else if (to.name !== 'not_found' && to.name !== 'forbidden') {
      let jwtStr = localStorage.getItem('dpg_jwt')
      if ( jwtStr) {
         router.store.commit("setJWT", jwtStr)
         next()
      } else {
         window.location.href = "/authenticate"
      }
   } else {
      next()
   }
 })

export default router
