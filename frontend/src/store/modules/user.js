import axios from 'axios'
import router from '../../router'

function parseJwt(token) {
   var base64Url = token.split('.')[1]
   var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
   var jsonPayload = decodeURIComponent(atob(base64).split('').map(function (c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
   }).join(''))

   return JSON.parse(jsonPayload);
}


const user = {
   namespaced: true,
   state: {
      jwt: "",
      firstName: "",
      lastName: "",
      role: "",
      computeID: "",
      ID: 0
   },
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
   },
   mutations: {
      signout(state) {
         localStorage.removeItem("dpg_jwt")
         state.jwt = ""
         state.firstName = ""
         state.lastName = ""
         state.role = ""
         state.computeID = ""
         state.ID = 0
      },
      setJWT(state, jwt) {
         if (jwt != state.jwt) {
            state.jwt = jwt
            localStorage.setItem("dpg_jwt", jwt)

            let parsed = parseJwt(jwt)
            state.ID = parsed.userID
            state.computeID = parsed.computeID
            state.firstName = parsed.firstName
            state.lastName = parsed.lastName
            state.role = parsed.role

            // add interceptor to put bearer token in header
            axios.interceptors.request.use( config => {
               config.headers['Authorization'] = 'Bearer ' + jwt
               return config
            }, error => {
               return Promise.reject(error)
            })

            // Catch 401 errors and redirect to an expired auth page
            axios.interceptors.response.use(
               res => res,
               err => {
                  if (err.config.url.match(/\/authenticate/) ) {
                     router.push( "/forbidden" )
                  } else {
                     if (err.response && err.response.status == 401 ) {
                        localStorage.removeItem("dpg_jwt")
                        state.jwt = ""
                        state.firstName = ""
                        state.lastName = ""
                        state.role = ""
                        state.computeID = ""
                        state.ID = 0
                        router.push( "/signedout?expired=1" )
                        return new Promise(() => { })
                     }
                  }
                  return Promise.reject(err)
               }
            )
         }
      },
   }
}
export default user