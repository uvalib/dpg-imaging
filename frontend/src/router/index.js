
import { createRouter, createWebHistory } from 'vue-router'
import Home from '../views/Home.vue'
import Project from '../views/Project.vue'
import Unit from '../views/Unit.vue'
import Image from '../views/Image.vue'
import NotFound from '../views/NotFound.vue'
import Forbidden from '../views/Forbidden.vue'
import SignedOut from '../views/SignedOut.vue'
import VueCookies from 'vue-cookies'

import { useUserStore } from '@/stores/user'
import { useMessageStore } from '@/stores/messages'

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
      path: '/messages',
      name: 'messages',
      component: () => import('../views/Messages.vue')
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
   routes,
   scrollBehavior(_to, _from, _savedPosition) {
      return { top: 0 }
   },
})

router.beforeEach((to, _from, next) => {
   const userStore = useUserStore()
   const messageStore = useMessageStore()
   if (to.path === '/granted') {
      let jwtStr = VueCookies.get("dpg_jwt")
      userStore.setJWT(jwtStr)
      messageStore.getMessages( userStore.ID )
      let priorURL = localStorage.getItem('dpgImagingPriorURL')
      console.log("RESTORE LAST PATH "+to.fullPath)
      localStorage.removeItem("dpgImagingPriorURL")
      if ( priorURL && priorURL != "/granted" && priorURL != "/") {
         console.log("RESTORE "+priorURL)
         next(priorURL)
      } else {
         next("/")
      }
   } else if (to.name !== 'not_found' && to.name !== 'forbidden' && to.name !== "signedout") {
      console.log("SAVE LAST PATH "+to.fullPath)
      localStorage.setItem("dpgImagingPriorURL", to.fullPath)
      let jwtStr = localStorage.getItem('dpg_jwt')
      if ( jwtStr) {
         userStore.setJWT(jwtStr)
         messageStore.getMessages( userStore.ID )
         next()
      } else {
         window.location.href = "/authenticate"
      }
   } else {
      next()
   }
 })

export default router
