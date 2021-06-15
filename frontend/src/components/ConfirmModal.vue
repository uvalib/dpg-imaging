<template>
   <div class="confirm-modal-wrapper">
      <DPGButton v-if="type=='button'" id="confirm-trigger" @click="show">{{label}}</DPGButton>
      <span class="txt-trigger" v-else id="confirm-trigger" @click="show">{{label}}</span>
      <div class="confirm-modal-dimmer" v-if="isOpen">
         <div role="dialog" aria-labelledby="confirm-modal-title" id="confirm-modal" class="confirm-modal">
            <div id="confirm-modal-title" class="confirm-modal-title">Confirm Action</div>
            <div class="confirm-modal-content">
               <slot />
               <p class="sure">Are you sure?</p>
            </div>
            <div class="confirm-modal-controls">
               <DPGButton id="close-confirm" @click="hide" @tabback="setFocus('ok-confirm')" :focusBackOverride="true">
                  No
               </DPGButton>
               <span class="spacer"></span>
               <DPGButton id="ok-confirm" @click="confirmClicked" @tabnext="setFocus('close-confirm')" :focusNextOverride="true">
                  Yes
               </DPGButton>
            </div>
         </div>
      </div>
   </div>
</template>

<script>
export default {
   props: {
      label: {
         type: String,
         required: true
      },
      type: {
         type: String,
         default: "button"
      }
   },
   data: function()  {
      return {
         isOpen: false
      }
   },
   methods: {
      confirmClicked() {
         this.hide()
         this.$nextTick( () => {
            this.$emit('confirmed')
         })
      },
      hide() {
         this.isOpen=false
         this.setFocus("confirm-trigger")
         this.$emit('closed')
      },
      show() {
         this.isOpen=true
         setTimeout(()=>{
            this.setFocus("close-confirm")
            this.$emit('opened')
         }, 150)
      },
      setFocus(id) {
         let ele = document.getElementById(id)
         if (ele ) {
            ele.focus()
         }
      },
   },
}
</script>

<style lang="scss" scoped>
.confirm-modal-dimmer {
   position: fixed;
   left: 0;
   top: 0;
   width: 100%;
   height: 100%;
   z-index: 1000;
   background: rgba(0, 0, 0, 0.2);
}
.txt-trigger {
   display: inline-block;
   cursor: pointer;
   width: 100%;
   &:hover {
      text-decoration: underline;
   }
}
div.confirm-modal {
   color: var(--uvalib-text);
   position: fixed;
   height: auto;
   z-index: 8000;
   background: white;
   top: 30%;
   left: 50%;
   transform: translate(-50%, -50%);
   box-shadow: var(--box-shadow);
   border-radius: 5px;
   min-width: 300px;
   border: 1px solid var(--uvalib-grey);

   .sure {
      text-align: right;
      padding: 0;
      margin: 10px 0 5px 0;
      font-weight: bold;
   }

   .spacer {
      display: inline-block;
      margin: 0 5px;
   }

   div.confirm-modal-content {
      padding: 20px 20px 0 20px;
      text-align: left;
      font-weight: normal;
   }
   div.confirm-modal-title {
      background:  var(--uvalib-blue-alt-light);
      font-size: 1.1em;
      color: var(--uvalib-text-dark);
      font-weight: 500;
      padding: 10px;
      border-radius: 5px 5px 0 0;
      border-bottom: 2px solid  var(--uvalib-blue-alt);
      text-align: left;
   }
   div.confirm-modal-controls {
      padding: 10px 20px 20px 20px;
      font-size: 0.9em;
      margin: 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
   }
}
</style>
