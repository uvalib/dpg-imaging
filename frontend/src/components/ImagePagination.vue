<template>
   <span class=pager>
      <span class="pages">
         <DPGButton mode="icon" :disabled="!prevAvailable" @clicked="firstClicked" aria-label="first page">
            <i class="fas fa-angle-double-left"></i>
         </DPGButton>
         <DPGButton mode="icon" :disabled="!prevAvailable"  @clicked="priorClicked" aria-label="previous page">
            <i class="fas fa-angle-left"></i>
         </DPGButton>
         <span class="page-info" @clicked="showPageJump">
            {{currPage+1}} of {{totalPages}}
         </span>
         <DPGButton mode="icon"  :disabled="!nextAvailable" @clicked="nextClicked" aria-label="next page">
            <i class="fas fa-angle-right"></i>
         </DPGButton>
         <DPGButton mode="icon" :disabled="!nextAvailable" @clicked="lastClicked" aria-label="last page">
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
         <select v-model="currPageSize" @change="pageSizeChanged">
            <option :value="20">20</option>
            <option :value="50">50</option>
            <option :value="75">75</option>
         </select>
      </span>
   </span>
</template>

<script>
import { mapGetters, mapState } from "vuex"
export default {
   computed: {
      ...mapGetters([
        'pageStartIdx',
        'totalPages',
        'totalFiles'
      ]),
      ...mapState({
         currPage : state => state.currPage,
         pageSize : state => state.pageSize,
      }),
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
         pageJump: 1,
         currPageSize: this.pageSize
      }
   },
   methods: {
      pageSizeChanged() {
         this.$store.dispatch("setPageSize", this.currPageSize)
         let query = Object.assign({}, this.$route.query)
         query.pagesize = this.currPageSize
         query.page = this.currPage+1
         this.$router.push({query})
      },
      pageJumpCanceled() {
         this.pageJumpOpen = false
      },
      pageJumpSelected() {
         if (this.pageJump <= 0 || this.pageJump > this.totalPages) return
         let tgtPage = (this.pageJump-1)
         this.$store.commit("setPage", tgtPage)
         this.pageChanged()
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
         this.$store.commit("setPage", this.currPage-1)
         this.pageChanged()
      },
      nextClicked() {
         this.$store.commit("setPage", this.currPage+1)
         this.pageChanged()
      },
      lastClicked() {
         this.$store.commit("setPage", this.totalPages-1)
         this.pageChanged()
      },
      firstClicked() {
         this.$store.commit("setPage", 0)
         this.pageChanged()
      },
      pageChanged() {
         let query = Object.assign({}, this.$route.query)
         query.page = this.currPage+1
         this.$router.push({query})
      }
   },
   mounted() {
      this.currPageSize = this.pageSize
      if ( this.$route.query.pagesize ) {
         let ps = parseInt(this.$route.query.pagesize, 10)
         this.currPageSize = ps
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