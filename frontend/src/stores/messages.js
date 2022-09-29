import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useMessageStore = defineStore('message', {
   state: () => ({
      inbox: [],
      sent: [],
      targetMessageID: -1,
      viewMesage: false,
      userID: -1
   }),
   getters: {
      unreadMessageCount: state => {
         let cnt = 0
         state.inbox.forEach( m=> {
            if (m.read == false) {
               cnt++
            }
         })
         return cnt
      }
   },
   actions: {
      getMessages(userID) {
         const system = useSystemStore()
         this.userID = userID
         axios.get(`/api/user/${userID}/messages`).then(response => {
            this.inbox = response.data.inbox
            this.sent = response.data.sent
            this.targetMessageID = -1
            this.viewMesage = false
            this.inbox.some( m => {
               if (m.read == false ) {
                  this.targetMessageID = m.id
                  this.viewMesage = true
               }
               return this.viewMesage == true
            })
         }).catch( e => {
            system.setError(e)
         })
      },
      markMessageRead() {
         const system = useSystemStore()
         axios.post(`/api/user/${this.userID}/messages/${this.targetMessageID}/read`).then( () => {
            let m = this.inbox.find( m => m.id == this.targetMessageID)
            m.read = true
            this.targetMessageID = -1
            this.viewMesage = false
         }).catch( e => {
            system.setError(e)
         })
      },
      deleteMessage(id) {
         const system = useSystemStore()
         axios.post(`/api/user/${this.userID}/messages/${id}/delete`).then( () => {
            let idx = this.inbox.findIndex( m => m.id == id)
            if (idx > -1) {
               this.inbox.splice(idx,1)
            }
         }).catch( e => {
            system.setError(e)
         })
      },
      sendMessage(msg) {
         const system = useSystemStore()
         axios.post(`/api/user/${this.userID}/messages/send`, msg).then( resp => {
            this.sent.push(resp.data)
         }).catch( e => {
            system.setError(e)
         })
      }
   },
})