<template>
   <span class=pager>
      <span class="pages">
         <DPGButton class="p-button-text right-pad" :disabled="!prevAvailable" @click="$emit('first')" aria-label="first page">
            <i class="fas fa-angle-double-left"></i>
         </DPGButton>
         <DPGButton class="p-button-text" :disabled="!prevAvailable" @click="$emit('prior')" aria-label="previous page">
            <i class="fas fa-angle-left"></i>
         </DPGButton>
         <span class="page-info" @click="showPageJump">
            {{currPage}} of {{totalPages}}
         </span>
         <DPGButton class="p-button-text right-pad" :disabled="!nextAvailable" @click="$emit('next')" aria-label="next page">
            <i class="fas fa-angle-right"></i>
         </DPGButton>
         <DPGButton class="p-button-text" :disabled="!nextAvailable" @click="$emit('last')" aria-label="last page">
            <i class="fas fa-angle-double-right"></i>
         </DPGButton>
      </span>
      <span class="setup" v-if="sizePicker">
         <label>per page:</label>
         <select @change="pageSizeChanged">
            <option :value="20" :selected="pageSize==20">20</option>
            <option :value="50" :selected="pageSize==50">50</option>
            <option :value="75" :selected="pageSize==75">75</option>
         </select>
      </span>
   </span>
   <Dialog v-model:visible="pageJumpOpen" :modal="true" header="Jump to page">
      <input id="page-jump" type="number" v-model="pageJump" :min="1" :max="totalPages"
         @keydown.stop.prevent.enter="pageJumpSelected" @keydown.stop.prevent.esc="pageJumpCanceled"/>
      <template #footer>
         <DPGButton class="p-button-secondary right-margin" @click="pageJumpCanceled" label="Cancel"/>
         <DPGButton @click="pageJumpSelected" label="OK"/>
      </template>
   </Dialog>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import Dialog from 'primevue/dialog'

const props = defineProps({
   totalPages: {
      type: Number,
      required: true
   },
   sizePicker: {
      type: Boolean,
      default: false
   },
   currPage: {
      type: Number,
      required: true
   },
   pageSize: {
      type: Number,
      required: true
   },
})
const emit = defineEmits( ['next', 'prior', 'first', 'last', 'jump', 'size' ] )

const prevAvailable = computed( () => {
   return props.currPage >1
})
const nextAvailable = computed( () => {
   return props.currPage < props.totalPages
})

const pageJumpOpen = ref(false)
const pageJump = ref(1)

function pageSizeChanged(e) {
   emit("size", parseInt(e.target.value, 10) )
}

function pageJumpCanceled() {
   pageJumpOpen.value = false
}

function pageJumpSelected() {
   if (pageJump.value <= 0 || pageJump.value > props.totalPages) return
   let tgtPage = pageJump.value
   pageJumpOpen.value = false
   emit("jump", tgtPage)
}

function showPageJump() {
   pageJumpOpen.value = true
   nextTick( () =>{
      let ele = document.getElementById("page-jump")
      ele.focus()
      ele.select()
   })
}
</script>

<style lang="scss" scoped>
.pager {
   display: flex;
   flex-flow: row nowrap;
   justify-content: space-evenly;
   position: relative;

   .pages {
      display: flex;
      flex-flow: row nowrap;
      justify-content: flex-end;
      align-content: center;
      align-items: center;
      font-size: 1.25em;
      font-weight: normal;
      color: var( --uvalib-text);
   }
   .page-info {
      margin: 0 10px;
      font-size: 0.85em;
      cursor: pointer;
   }
   .setup {
      margin-left: 10px;
      select {
         width: max-content;
         margin-left: 5px;
      }
   }
}
</style>