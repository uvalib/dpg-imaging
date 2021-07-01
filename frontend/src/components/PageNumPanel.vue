<template>
   <div class="page-number panel">
      <h3>Set Page Numbering</h3>
      <div class="content">
         <span class="entry pad-right">
            <label>Start Image:</label>
            <select id="start-page" v-model="rangeStartIdx">
               <option disabled :value="-1">Select start page</option>
               <option v-for="(mf,idx) in masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
         <span class="entry  pad-right">
            <label>End Image:</label>
            <select id="end-page" v-model="rangeEndIdx">
               <option disabled :value="-1">Select end page</option>
               <option v-for="(mf,idx) in masterFiles" :value="idx" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
            <span class="entry">
            <label>Starting Page:</label>
            <input id="start-page-num" type="text" v-model="startPage"  @keyup.enter="okPagesClicked"/>
            </span>
      </div>
      <div class="panel-actions">
         <DPGButton @clicked="selectAllClicked" class="left">Select All</DPGButton>
         <label class="verso">Unnumbered Verso<input v-model="unnumberVerso" type="checkbox"/></label>
         <DPGButton @clicked="cancelEditClicked" class="right-pad">Cancel</DPGButton>
         <DPGButton @clicked="okPagesClicked">OK</DPGButton>
      </div>
   </div>
</template>

<script>
import { mapState } from "vuex"
import { mapFields } from 'vuex-map-fields'
export default {
   computed: {
      ...mapState({
         masterFiles : state => state.masterFiles,
         currUnit: state => state.currUnit,
      }),
      ...mapFields([
         "rangeStartIdx", "rangeEndIdx", "editMode"
      ]),
   },
   data() {
      return {
        startPage: "1",
        unnumberVerso: false
      }
   },
   methods: {
      cancelEditClicked() {
         this.$store.commit("clearError")
         this.editMode = ""
      },
      okPagesClicked() {
         this.$store.commit("clearError")
         if ( this.rangeStartIdx == -1 || this.rangeEndIdx == -1) {
            this.$store.commit("setError", "Start and end image must be selected")
            return
         }
         if (this.startPage == "") {
            this.$store.commit("setError", "Start page is required")
            return
         }
         if (this.unnumberVerso && (this.rangeEndIdx-this.rangeStartIdx)%2 == 0) {
            this.$store.commit("setError", "An even number of pages is required for unnumbered verso")
            return
         }
         this.$store.dispatch("updatePageNumbers", {start: this.startPage, verso: !this.unnumberVerso})
         this.editMode = ""
      },
      selectAllClicked() {
         this.$store.commit("selectAll")
      }
   }
}
</script>

<style lang="scss" scoped>
.panel {
   background: white;
   border-bottom: 1px solid var(--uvalib-grey);
   h3 {
      margin: 0;
      padding: 8px 0;
      font-size: 1em;
      background: var(--uvalib-blue-alt-light);
      border-bottom: 1px solid var(--uvalib-grey);
      font-weight: 500;
   }
   .panel-actions {
      padding: 0 0 20px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
      width: 50%;
      margin: 0 auto;
      .button {
         margin-left: 10px;
      }
      .left {
         margin-right: auto;
      }
   }
   .verso {
      cursor: pointer;
      margin-right: auto;
      label {
         vertical-align: middle;
      }
      input {
         width: 16px;
         height: 16px;
         margin-left: 10px;
         vertical-align: middle;
      }
   }
   .content {
      padding: 10px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      width: 50%;
      margin: 0 auto;
      .entry {
         flex-grow: 1;
         margin: 0;
         text-align: left;
         label {
            display: block;
            margin: 0 0 5px 0;
         }
      }
      .entry.pad-right {
         padding-right: 25px;
      }
   }
   .error {
      font-style: italic;
      color: var(--uvalib-red-emergency);
      margin: 0;
   }
}
</style>
