<template>
   <div class="note-modal-wrapper">
      <DPGButton :id="`${id}-trigger`" v-if="!manual" mode="icon" @clicked="show"><i class="fas fa-plus-circle"></i></DPGButton>
      <div class="note-modal-dimmer" v-if="isOpen">
         <div role="dialog" :aria-labelledby="`${id}-title`" :id="id" class="note-modal">
            <div :id="`${id}-title`" class="note-modal-title">Create Note</div>
            <div class="note-modal-content">
               <div class="instruct" v-if="instructions">{{instructions}}</div>
               <div class="row">
                  <label>Note Type {{trigger}}</label>
                  <select v-model="noteTypeID" :id="`${id}-type`">
                     <option :value="0">Comment</option>
                     <option :value="1">Suggestion</option>
                     <option :value="2">Problem</option>
                     <option :value="3">Item Condition</option>
                  </select>
               </div>
               <div class="row pad" v-if="noteTypeID==2">
                  <label>Problem (select all that apply)</label>
                  <label class="cb" v-for="p in systemStore.problems" :key="p.label">
                     <input type="checkbox" :value="p.id" v-model="problemIDs" />
                     {{p.name}}
                  </label>
               </div>
               <div class="row pad">
                  <label for="note-text">Note Text</label>
                  <textarea rows="5" v-model="note"></textarea>
               </div>
            </div>
            <p class="error" v-if="error">{{error}}</p>
            <div class="note-modal-controls">
               <DPGButton :id="`${id}-close`" @clicked="hide" :focusBackOverride="true">
                  Cancel
               </DPGButton>
               <span class="spacer"></span>
               <DPGButton :id="`${id}-ok`" @clicked="createClicked" @tabnext="okNextTab" :focusNextOverride="true">
                  Create
               </DPGButton>
            </div>
         </div>
      </div>
   </div>
</template>

<script>
import { mapStores } from 'pinia'
import { useSystemStore } from '@/stores/system'
export default {
   emits: ['opened', 'closed', 'submitted' ],
   props: {
      id: {
         type: String,
         required: true
      },
      trigger: {
         type: Boolean,
         default: false,
      },
      manual: {
         type: Boolean,
         default: false
      },
      noteType: {
         type: Number,
         default: 0
      },
      instructions: {
         type: String,
         default: ""
      }
   },
   watch: {
       trigger(newtrigger) {
          if (this.manual && newtrigger) {
            this.isOpen = true,
            setTimeout(()=>{
               this.setFocus(`${this.id}-close`)
               this.$emit('opened')
            }, 150)
          }
       }
   },
   data: function()  {
      return {
         isOpen: false,
         noteTypeID: this.noteType, //[:comment, :suggestion, :problem, :item_condition]
         note: "",
         problemIDs: [],
         error: ""
      }
   },
   computed: {
      ...mapStores(useSystemStore),
   },
   methods: {
      okNextTab() {
         this.setFocus(`${this.id}-type`)
      },
      createClicked() {
         this.error = ""
         if ( this.note == "") {
            this.error = "Note text is required"
            return
         }
         if ( this.noteTypeID == 2 && this.problemIDs.length == 0) {
            this.error = "At least one problem is required"
            return
         }
         let data = {noteTypeID: this.noteTypeID, note: this.note, problemIDs: this.problemIDs}
         this.$store.dispatch("projects/addNote", data)
         this.isOpen=false
         this.setFocus(`${this.id}-trigger`)
         this.$emit('submitted')
      },
      hide() {
         this.$emit('closed')
         this.isOpen=false
         this.setFocus(`${this.id}-trigger`)
      },
      show() {
         this.isOpen=true
         this.noteTypeID = 0
         this.note = ""
         this.problemIDs = []
         setTimeout(()=>{
            this.setFocus(`${this.id}-close`)
            this.$emit('opened')
         }, 150)
         this.error = ""
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
.note-modal-wrapper {
   margin-left: auto;
   button {
      height: 100%;
   }
}
#note-trigger {
   color: var(--uvalib-grey);
}
p.error {
   color: var(--uvalib-red-emergency);
   margin: 5px;
   text-align: center;
   font-weight: normal;
   font-style: italic;
}
.note-modal-dimmer {
   position: fixed;
   left: 0;
   top: 0;
   width: 100%;
   height: 100%;
   z-index: 1000;
   background: rgba(0, 0, 0, 0.2);
}
div.note-modal {
   color: var(--uvalib-text);
   position: fixed;
   height: auto;
   z-index: 8000;
   background: white;
   top: 40%;
   left: 50%;
   transform: translate(-50%, -50%);
   box-shadow: var(--box-shadow);
   border-radius: 5px;
   min-width: 300px;
   max-width: 400px;
   border: 1px solid var(--uvalib-grey);

   .spacer {
      display: inline-block;
      margin: 0 5px;
   }

   div.note-modal-content {
      padding: 10px 10px 0 10px;
      text-align: left;
      font-weight: normal;
      .row.pad {
         margin-top: 20px;
      }
      label {
         display: block;
         font-weight: bold;
         margin-bottom: 5px;
         font-size: 0.9em;
      }
      label.cb {
         font-weight: normal;
         input[type=checkbox] {
            width: auto;
            margin-left: 25px;
            margin-right: 5px;
         }
      }
      textarea, select  {
         border-color: var(--uvalib-grey-light);
         border-radius: 5px;
         box-sizing: border-box;
         width: 100%;
      }
      textarea {
         padding: 5px;
      }
   }
   div.note-modal-title {
      background:  var(--uvalib-blue-alt-light);
      font-size: 1.1em;
      color: var(--uvalib-text-dark);
      font-weight: 500;
      padding: 10px;
      border-radius: 5px 5px 0 0;
      border-bottom: 2px solid  var(--uvalib-blue-alt);
      text-align: left;
   }
   div.note-modal-controls {
      padding: 10px 10px 10px 10px;
      font-size: 0.9em;
      margin: 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
   }
   div.instruct {
      margin: 10px 0 20px;
      padding: 15px;
      border: 1px solid var(--uvalib-blue-alt);
      background: var(--uvalib-blue-alt-light);
   }
}
</style>
