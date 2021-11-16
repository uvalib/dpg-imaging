<template>
   <div class="assign-modal-wrapper">
      <DPGButton id="assign-trigger" @clicked="show">Assign</DPGButton>
      <div class="assign-modal-dimmer" v-if="isOpen">
         <div role="dialog" aria-labelledby="assign-modal-title" id="assign-modal" class="assign-modal">
            <div id="assign-modal-title" class="assign-modal-title">Assign Project</div>
            <div class="assign-modal-content">
               <div v-if="working" class="spinner-wrap">
                  <WaitSpinner :overlay="false" message="Loading candidates..." />
               </div>
               <div v-else class="candidate-scroller">
                  <div class="val" v-for="(c,idx) in candidates" :key="c.id" :class="{selected: idx == selectedIdx}" @click="selectCandidate(idx)">
                     <span class="candidate">{{c.lastName}}, {{c.firstName}}</span> ({{c.computingID}})
                  </div>
               </div>
            </div>
            <p class="error">{{error}}</p>
            <div class="assign-modal-controls">
               <DPGButton id="close-assign" @clicked="hide" @tabback="setFocus('ok-assign')" :focusBackOverride="true">
                  Cancel
               </DPGButton>
               <span class="spacer"></span>
               <DPGButton id="ok-assign" @clicked="assignClicked" @tabnext="setFocus('close-assign')" :focusNextOverride="true">
                  Assign
               </DPGButton>
            </div>
         </div>
      </div>
   </div>
</template>

<script>
import { mapState } from "vuex"
export default {
   emits: ['assign', 'closed', 'opened' ],
   props: {
      projectID: {
         type: Number,
         required: true,
      },
   },
   data: function()  {
      return {
         isOpen: false,
         selectedIdx: -1,
         error: ""
      }
   },
   computed: {
      ...mapState({
         working: state => state.projects.working,
         candidates: state => state.staffMembers,
      })
   },
   methods: {
      selectCandidate(idx) {
         this.selectedIdx = idx
      },
      assignClicked() {
         this.error = ""
         if ( this.selectedIdx == -1) {
            this.error = "Please select a user"
            return
         }
         this.hide()
         this.$nextTick( () => {
            let userID = this.candidates[this.selectedIdx].id
            this.$emit('assign', {projectID: this.projectID, ownerID: userID})
         })
      },
      hide() {
         this.isOpen=false
         this.setFocus("assign-trigger")
         this.$emit('closed')
      },
      show() {
         this.isOpen=true
         setTimeout(()=>{
            this.setFocus("close-assign")
            this.$emit('opened')
         }, 150)
         this.error = ""
         this.selectedIdx = -1
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
.assign-modal-wrapper {
   button {
      height: 100%;
   }
}
.error {
   padding: 5px 10px;
   margin: 0;
   text-align: center;
   color: var(--uvalib-red-emergency);
}
.spinner-wrap {
   text-align: center;
   margin-bottom: 25px;;
}
.assign-modal-dimmer {
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
div.assign-modal {
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

   .spacer {
      display: inline-block;
      margin: 0 5px;
   }

   div.assign-modal-content {
      padding: 10px 10px 0 10px;
      text-align: left;
      font-weight: normal;
      .candidate-scroller {
         max-height: 300px;
         overflow: scroll;
         padding: 0;
         margin:  10px;
         border: 1px solid var(--uvalib-grey-light);
         border-radius: 4px;
         .val {
            padding: 2px 10px 3px 10px;
            cursor: pointer;
            display: flex;
            flex-flow: row nowrap;
            justify-content: space-between;
            &:hover  {
               background: var(--uvalib-blue-alt-light);
            }
         }
         .val.selected {
            background: var(--uvalib-blue-alt);
            color: white;
         }
         .candidate {
            font-weight: bold;
         }
      }
   }
   div.assign-modal-title {
      background:  var(--uvalib-blue-alt-light);
      font-size: 1.1em;
      color: var(--uvalib-text-dark);
      font-weight: 500;
      padding: 10px;
      border-radius: 5px 5px 0 0;
      border-bottom: 2px solid  var(--uvalib-blue-alt);
      text-align: left;
   }
   div.assign-modal-controls {
      padding: 10px 20px 20px 20px;
      font-size: 0.9em;
      margin: 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
   }
}
</style>
