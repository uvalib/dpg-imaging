import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex)

export default new Vuex.Store({
   state: {
      working: false,
      error: false,
   },
   mutations: {
      setWorking(state, flag) {
         state.working = flag
         if (flag == true ) {
            state.error = false
         }
      },
   },
   actions: {
   },
   modules: {
   }
})
