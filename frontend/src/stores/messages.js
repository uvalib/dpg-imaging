import { defineStore } from 'pinia'
import { useSystemStore } from './system'
import axios from 'axios'

export const useMessageStore = defineStore('message', {
   state: () => ({
      inbox: [],
      sent: [],
      targetMessageID: -1,
      viewMesage: false,
      showCreateModal: false,
      newMessage:  {to: 0, subject: "", message: ""},
      isResponse: false,
      userID: -1,
      error: ""
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
      beginMessageCreate() {
         this.showCreateModal = true
         this.newMessage = {to: 0, subject: "", message: ""}
         this.error = ""
         this.isResponse = false
      },
      beginReply() {
         if ( this.targetMessageID == -1) return

         let srcMsg = this.inbox.find( m => m.id == this.targetMessageID)
         this.viewMesage = false

         this.isResponse = true
         this.showCreateModal = true
         this.newMessage = {to: srcMsg.from.id, subject: `RE: ${srcMsg.subject}`, message: ""}
         this.error = ""
      },
      cancelMessage() {
         this.showCreateModal = false
         this.error = ""
      },
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
         // the message modal can call this when OK is clicked or when the close callback is triggered
         // this can result in this call being made twice. Only handle the call when the target
         // message ID is not -1
         if ( this.targetMessageID == -1) return

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

      sendMessage() {
         if (this.newMessage.to == 0) {
            this.error = "Please select a recipient"
            return
         }
         if (this.newMessage.subject == "") {
            this.error = "Please set a subject"
            return
         }
         if (this.newMessage.message == "") {
            this.error = "Please add a message"
            return
         }

         if (this.isResponse) {
            let srcMsg = this.inbox.find( m => m.id == this.targetMessageID)
            this.newMessage.message = `${this.newMessage.message}\n\n>>>>\n${srcMsg.message}\n<<<<`
         }

         axios.post(`/api/user/${this.userID}/messages/send`, this.newMessage).then( resp => {
            this.error = ""
            this.showCreateModal = false
            this.sent.push(resp.data)
         }).catch( e => {
            this.error = e
         })
      }
   },
})