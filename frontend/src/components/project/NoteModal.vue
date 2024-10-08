<template>
   <DPGButton v-if="!manual"  label="Add Note" size="small" @click="showClicked" ref="notetrigger" />
   <Dialog v-model:visible="isOpen" :modal="true" header="Create Note" style="width:650px" @show="show" @afterHide="hidden">
      <div class="note-modal-content">
         <div class="instruct" v-if="instructions">{{instructions}}</div>
         <div class="row">
            <label>Note Type {{trigger}}</label>
            <select v-model="noteTypeID" :id="`${props.id}-type`" ref="typesel">
               <option :value="0">Comment</option>
               <option :value="1">Suggestion</option>
               <option :value="2">Problem</option>
               <option :value="3">Item Condition</option>
            </select>
         </div>
         <div class="row pad" v-if="noteTypeID==2">
            <label>Problem (select all that apply)</label>
            <label class="cb" v-for="p in systemStore.problemTypes" :key="p.label">
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
      <template #footer>
         <DPGButton @click="hide" severity="secondary" label="Cancel"/>
         <DPGButton autofocus @click="createClicked" label="Create"/>
      </template>
   </Dialog>
</template>

<script setup>
import { useSystemStore } from '@/stores/system'
import { useProjectStore } from '@/stores/project'
import { ref, watch } from 'vue'
import Dialog from 'primevue/dialog'
import { useFocus } from '@vueuse/core'

const systemStore = useSystemStore()
const projectStore = useProjectStore()

const emit = defineEmits( ['opened', 'closed', 'submitted' ] )

const props = defineProps({
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
})

const isOpen = ref(false)
const noteTypeID = ref(props.noteType) //[:comment, :suggestion, :problem, :item_condition
const note = ref("")
const problemIDs = ref([])
const error = ref("")
const notetrigger = ref()
const { focused: triggerFocus } = useFocus(notetrigger)
const typesel = ref()
const { focused: selFocus } = useFocus(typesel)

watch(() => props.trigger, (newtrigger) => {
   if (props.manual && newtrigger) {
      isOpen.value = true
   }
})

const createClicked = (() => {
   error.value = ""
    if ( note.value == "") {
      error.value = "Note text is required"
      return
   }
   if (noteTypeID.value == 2 && problemIDs.value.length == 0) {
      error.value = "At least one problem is required"
      return
   }
   let data = {noteTypeID: noteTypeID.value, note: note.value, problemIDs: problemIDs.value}
   projectStore.addNote(data)
   isOpen.value = false
   emit('submitted')
})

const hide = (() => {
   isOpen.value = false
})

const hidden = (() => {
   triggerFocus.value = true
   emit('closed')
})

const show = (() => {
   setTimeout( ()=> { selFocus.value = true }, 250 )
})

const showClicked = (() => {
   isOpen.value = true
   noteTypeID.value = 0
   note.value = ""
   problemIDs.value = []
   error.value = ""
})
</script>

<style lang="scss" scoped>
p.error {
   color: var(--uvalib-red-emergency);
   margin: 5px;
   text-align: center;
   font-weight: normal;
   font-style: italic;
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

   div.instruct {
      margin: 10px 0 20px;
      padding: 15px;
      border: 1px solid var(--uvalib-blue-alt);
      background: var(--uvalib-blue-alt-light);
   }

</style>
