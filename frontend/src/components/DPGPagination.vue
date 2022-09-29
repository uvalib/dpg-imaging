<template>
   <span class=pager>
      <span class="pages">
         <DPGButton2 class="p-button-text right-pad" :disabled="!prevAvailable" @click="$emit('first')" aria-label="first page">
            <i class="fas fa-angle-double-left"></i>
         </DPGButton2>
         <DPGButton2 class="p-button-text" :disabled="!prevAvailable" @click="$emit('prior')" aria-label="previous page">
            <i class="fas fa-angle-left"></i>
         </DPGButton2>
         <span class="page-info" @click="showPageJump">
            {{currPage}} of {{totalPages}}
         </span>
         <DPGButton2 class="p-button-text right-pad" :disabled="!nextAvailable" @click="$emit('next')" aria-label="next page">
            <i class="fas fa-angle-right"></i>
         </DPGButton2>
         <DPGButton2 class="p-button-text" :disabled="!nextAvailable" @click="$emit('last')" aria-label="last page">
            <i class="fas fa-angle-double-right"></i>
         </DPGButton2>
         <div class="page-jump" v-if="pageJumpOpen">
            <label>Jump to page</label>
            <div class="jump-body">
               <input id="page-jump" type="number" v-model="pageJump" :min="1" :max="totalPages"
                  @keydown.stop.prevent.enter="pageJumpSelected" @keydown.stop.prevent.esc="pageJumpCanceled"/>
            </div>
            <div class="button-bar">
               <DPGButton2 class="p-button-secondary right-margin" @click="pageJumpCanceled" label="Cancel"/>
               <DPGButton2 class="p-button-secondary" @click="pageJumpSelected" label="OK"/>
            </div>
         </div>
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
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
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
      font-size: 16px;
      font-weight: normal;
      color: var( --uvalib-text);
   }
   .page-info {
      margin: 0 10px;
      font-size: 0.85em;
      cursor: pointer;
   }
   .page-jump {
      background: white;
      padding: 0;
      border: 1px solid var(--uvalib-grey);
      box-shadow: var(--box-shadow);
      font-size: 0.9em;
      position: absolute;
      top: 20px;
      left: 5px;
      z-index: 10000;
      text-align: left;
      width: 130px;
      .jump-body {
         margin: 10px;
      }
      label {
         padding: 5px;
         background: var(--uvalib-blue-alt-light);
         width: 100%;
         display: block;
         box-sizing: border-box;
         border-bottom: 1px solid  var(--uvalib-blue-alt);
      }
      input {
         width: 100%;
         margin: 10px 0 0 0;
         display: block;
         box-sizing: border-box;
      }
      .button-bar {
         font-size: 0.85em;
         text-align: right;
         background: white;
         padding: 0 10px 10px 10px;
         .right-margin {
            margin-right: 5px;
         }
      }
   }
   .setup {
      margin-left: 10px;
      font-size: 0.9em;
      select {
         width: max-content;
         margin-left: 5px;
      }
   }
}
</style>