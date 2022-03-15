<template>
   <div class="message-dimmer">
      <div class="messsage-box">
         <div class="message" role="alertdialog" aria-modal="true"
            aria-labelledby="msgtitle" aria-describedby="msgbody"
         >
            <div class="bar">
               <span tabindex="-1" id="msgtitle" class="title">DPG Imaging Error</span>
            </div>
            <div class="message-body" id="msgbody" v-html="message"></div>
            <div class="controls">
               <DPGButton id="close-error" @clicked="dismiss">OK</DPGButton>
            </div>
         </div>
      </div>
   </div>
</template>

<script setup>
import {useSystemStore} from "@/stores/system"
import { onMounted, nextTick } from 'vue'

const systemStore = useSystemStore()

function dismiss() {
   systemStore.error = ""
}

onMounted( () => {
   nextTick( () =>{
      let ele = document.getElementById("close-error")
      ele.focus()
   })
})
</script>

<style lang="scss" scoped>
.message-dimmer {
   position: fixed;
   left: 0;
   top: 0;
   width: 100%;
   height: 100%;
   z-index: 1000;
   background: rgba(0, 0, 0, 0.2);
}
div.messsage-box {
   position: fixed;
   left: 0;
   right: 0;
   z-index: 9999;
   top: 25%;

   .details {
      text-align: left;
      padding: 0 30px 20px 30px;
   }

   .message {
      display: inline-block;
      text-align: left;
      background: white;
      padding: 0px;
      box-shadow:  var(--box-shadow);
      min-width: 20%;
      max-width: 80%;
      border-radius: 5px;
      border: 1px solid var(--uvalib-grey);
      .bar {
         padding: 5px;
         color: white;
         font-weight: bold;
         display: flex;
         flex-flow: row nowrap;
         align-items: center;
         justify-content: space-between;
         background-color: var( --uvalib-red-dark);
         border-bottom: 2px solid var( --uvalib-red-darker);
         border-radius: 5px 5px 0 0;
         font-size: 1.1em;
         padding: 10px;
      }

      .message-body {
         text-align: left;
         padding: 20px 30px 0 30px;
         font-weight: normal;
         opacity: 1;
         visibility: visible;
         text-align: left;
         word-break: break-word;
         -webkit-hyphens: auto;
         -moz-hyphens: auto;
         hyphens: auto;
         color: var(--uvalib-primary-text);
      }

      .controls {
         padding: 15px 10px 10px 0;
         font-size: 0.9em;
         text-align: right;
      }
   }
}

</style>
