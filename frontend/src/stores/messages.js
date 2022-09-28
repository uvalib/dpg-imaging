import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useMessageStore = defineStore('message', {
   state: () => ({
      messages: [],
   }),
   getters: {
   },
   actions: {
      getMessages(userID) {
         const system = useSystemStore()
         axios.get(`/api/user/${userID}/messages`).then(response => {
            this.messages = response.data
            system.working = false
         }).catch( e => {
            system.setError(e)
         })
      }
   },
})