
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

router.beforeEach( (to) => {
   console.log("BEFORE ROUTE "+to.path)
   const userStore = useUserStore()
   const noAuthRoutes = ["not_found", "forbidden", "signedout"]

   if (to.path === '/granted') {
      let jwtStr = VueCookies.get("dpg_jwt")
      userStore.setJWT(jwtStr)
      if ( userStore.isSignedIn  ) {
         console.log(`GRANTED [${jwtStr}]`)
         useMessageStore().getMessages( userStore.ID )
         let priorURL = localStorage.getItem('dpgImagingPriorURL')
         localStorage.removeItem("dpgImagingPriorURL")
         if ( priorURL && priorURL != "/granted" && priorURL != "/") {
            console.log("RESTORE "+priorURL)
            return priorURL
         }
         return "/"
      }
      return {name: "forbidden"}
   }

   if ( noAuthRoutes.includes(to.name)) {
      console.log("NOT A PROTECTED PAGE")
   } else {
      if (userStore.isSignedIn == false) {
         console.log("AUTHENTICATE")
         localStorage.setItem("dpgImagingPriorURL", to.fullPath)
         window.location.href = "/authenticate"
         return false   // cancel the original navigation
      } else {
         console.log(`REQUEST AUTHENTICATED PAGE WITH JWT`)
      }
   }
})

export default router
