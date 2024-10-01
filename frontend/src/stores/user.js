import axios from 'axios'
import { defineStore } from 'pinia'
import { useJwt } from '@vueuse/integrations/useJwt'

export const useUserStore = defineStore('user', {
   state: () => ({
      jwt: "",
      firstName: "",
      lastName: "",
      role: "",
      computeID: "",
      ID: 0
   }),
   getters: {
      isAdmin: state => {
         return state.role == "admin"
      },
      isSupervisor: state => {
         return state.role == "supervisor"
      },
      signedInUser: state => {
         return `${state.firstName} ${state.lastName} (${state.computeID})`
      },
      isSignedIn: state => {
         return state.jwt != "" && state.computeID != ""
      }
   },
   actions: {
      signout() {
         localStorage.removeItem("dpg_jwt")
         localStorage.removeItem("dpgImagingPriorURL")
         this.jwt = ""
         this.firstName = ""
         this.lastName = ""
         this.role = ""
         this.computeID = ""
         this.ID = 0
      },
      setJWT(jwt) {
         if (jwt == this.jwt || jwt == "" || jwt == null || jwt == "null")  return

         this.jwt = jwt
         localStorage.setItem("dpg_jwt", jwt)

         const { payload } = useJwt(jwt)
         this.ID = payload.value.userID
         this.computeID = payload.value.computeID
         this.firstName = payload.value.firstName
         this.lastName = payload.value.lastName
         this.role = payload.value.role

         // add interceptor to put bearer token in header
         axios.interceptors.request.use(config => {
            config.headers['Authorization'] = 'Bearer ' + jwt
            return config
         }, error => {
            return Promise.reject(error)
         })

         // Catch 401 errors and redirect to an expired auth page
         axios.interceptors.response.use(
            res => res,
            err => {
               if (err.config.url.match(/\/authenticate/)) {
                  this.router.push("/forbidden")
               } else {
                  if (err.response && err.response.status == 401) {
                     localStorage.removeItem("dpg_jwt")
                     this.jwt = ""
                     this.firstName = ""
                     this.lastName = ""
                     this.role = ""
                     this.computeID = ""
                     this.ID = 0
                     this.router.push("/signedout?expired=1")
                     return new Promise(() => { })
                  }
               }
               return Promise.reject(err)
            }
         )
      },
   }
})