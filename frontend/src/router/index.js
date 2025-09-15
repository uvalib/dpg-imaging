import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useMessageStore } from '@/stores/messages'
import { useCookies } from "vue3-cookies"

const routes = [
   {
      path: '/',
      name: 'home',
      component: () => import('../views/Home.vue')
   },
   {
      path: '/projects/:id',
      name: 'project',
      component: () => import('../views/Project.vue')
   },
   {
      path: '/projects/:id/unit',
      name: 'unit',
      component: () => import('../views/Unit.vue')
   },
   {
      path: '/projects/:id/unit/images/:page',
      name: 'image',
      component: () => import('../views/Image.vue')
   },
   {
      path: '/messages',
      name: 'messages',
      component: () => import('../views/Messages.vue')
    },
   {
      path: '/forbidden',
      name: 'forbidden',
      component: () => import('../views/Forbidden.vue')
   },
   {
      path: '/signedout',
      name: 'signedout',
      component: () => import('../views/SignedOut.vue')
   },
   {
      path: '/:pathMatch(.*)*',
      name: "not_found",
      component: () => import('../views/NotFound.vue')
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
   const { cookies } = useCookies()
   const noAuthRoutes = ["not_found", "forbidden", "signedout"]

   if (to.path === '/granted') {
      const jwtStr = cookies.get("dpg_jwt")
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
