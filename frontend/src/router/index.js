
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Project from '../views/Project.vue'
import Unit from '../views/Unit.vue'
import Image from '../views/Image.vue'
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
   {
      path: '/projects/:id/unit',
      name: 'unit',
      component: Unit
   },
   {
      path: '/projects/:id/unit/images/:page',
      name: 'image',
      component: Image
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
   history: createWebHistory(import.meta.env.BASE_URL),
   routes
})

router.beforeEach((to, _from, next) => {
   if (to.path === '/granted') {
      let jwtStr = VueCookies.get("dpg_jwt")
      store.commit("user/setJWT", jwtStr)
      next( "/" )
   } else if (to.name !== 'not_found' && to.name !== 'forbidden' && to.name !== "signedout") {
      let jwtStr = localStorage.getItem('dpg_jwt')
      if ( jwtStr) {
         store.commit("user/setJWT", jwtStr)
         next()
      } else {
         window.location.href = "/authenticate"
      }
   } else {
      next()
   }
 })

export default router
