<template>
   <span class=pager>
      <span class="pages">
         <DPGButton mode="icon" :disabled="!prevAvailable" @click="firstClicked" aria-label="first page">
            <i class="fas fa-angle-double-left"></i>
         </DPGButton>
         <DPGButton mode="icon" :disabled="!prevAvailable"  @click="priorClicked" aria-label="previous page">
            <i class="fas fa-angle-left"></i>
         </DPGButton>
         <span class="page-info" @click="showPageJump">
            {{currPage+1}} of {{totalPages}}
         </span>
         <DPGButton mode="icon"  :disabled="!nextAvailable" @click="nextClicked" aria-label="next page">
            <i class="fas fa-angle-right"></i>
         </DPGButton>
         <DPGButton mode="icon" :disabled="!nextAvailable" @click="lastClicked" aria-label="last page">
            <i class="fas fa-angle-double-right"></i>
         </DPGButton>
         <div class="page-jump" v-if="pageJumpOpen">
            <label>Jump to page</label>
            <input id="page-jump" type="number" v-model="pageJump" :min="1" :max="totalPages"
               @keydown.stop.prevent.enter="pageJumpSelected" @keydown.stop.prevent.esc="pageJumpCanceled"/>
         </div>
      </span>
      <span class="setup">
         <label>per page:</label>
         <select v-model="pageSize" >
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="75">75</option>
         </select>
      </span>
   </span>
</template>

<script>
import { mapGetters } from "vuex"
import { mapFields } from 'vuex-map-fields'
export default {
   computed: {
      ...mapGetters([
        'pageStartIdx',
        'totalPages',
        'totalFiles'
      ]),
      ...mapFields([
         'currPage', "pageSize",
      ]),
      prevAvailable() {
         return this.currPage > 0
      },
      nextAvailable() {
         return this.currPage < (this.totalPages-1)
      }
   },
   data() {
      return {
         pageJumpOpen: false,
         pageJump: 1
      }
   },
   methods: {
      pageJumpCanceled() {
         this.pageJumpOpen = false
      },
      pageJumpSelected() {
         if (this.pageJump <= 0 || this.pageJump > this.totalPages) return
         this.currPage = (this.pageJump-1)
         this.pageJumpOpen = false
       },
      showPageJump() {
         this.pageJumpOpen = true
         this.$nextTick( () =>{
            let ele = document.getElementById("page-jump")
            ele.focus()
            ele.select()
         })
      },
      priorClicked() {
         this.currPage--
      },
      nextClicked() {
         this.currPage++
      },
      lastClicked() {
         this.currPage = this.totalPages-1
      },
      firstClicked() {
         this.currPage = 0
      }
   }
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
      padding: 10px;
      border-radius: 4px;
      border: 1px solid var(--uvalib-grey);
      box-shadow: var(--box-shadow);
      font-size: 0.9em;
      position: absolute;
      top: 0px;
      left: 45px;
      input {
         width: 95px;
         margin: 5px 0 0 0;
         display: block;
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