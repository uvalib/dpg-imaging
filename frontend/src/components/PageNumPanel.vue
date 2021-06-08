<template>
   <div class="page-number panel">
      <h3>Set Page Numbering</h3>
      <div class="content">
         <span class="entry">
            <label>Start Image:</label>
            <select id="start-page" v-model="rangeStart">
               <option disabled value="">Select start page</option>
               <option v-for="mf in masterFiles" :value="mf.fileName" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
         <span class="entry">
            <label>End Image:</label>
            <select id="end-page" v-model="rangeEnd">
               <option disabled value="">Select end page</option>
               <option v-for="mf in masterFiles" :value="mf.fileName" :key="`start-${mf.fileName}`">{{mf.fileName}}</option>
            </select>
         </span>
            <span class="entry">
            <label>Starting Page:</label>
            <input id="start-page-num" type="text" v-model="startPage"  @keyup.enter="okPagesClicked"/>
            </span>
      </div>
      <p class="error" v-if="error" v-html="errorMessage"></p>
      <div class="panel-actions">
         <span tabindex="0" class="button" @click="cancelEditClicked">Cancel</span>
         <span tabindex="0" class="button" @click="okPagesClicked">OK</span>
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
         error: state => state.error,
         errorMessage: state => state.errorMessage,
      }),
      ...mapFields([
         "rangeStart", "rangeEnd", "editMode"
      ]),
   },
   data() {
      return {
        startPage: "1",
      }
   },
   methods: {
      cancelEditClicked() {
         this.$store.commit("clearError")
         this.editMode = ""
      },
      okPagesClicked() {
         this.$store.commit("clearError")
         if ( this.rangeStart == "" || this.rangeEnd == "") {
            this.$store.commit("setError", "Start and end image must be selected")
            return
         }
         this.$store.dispatch("updatePageNumbers", this.startPage)
         this.editMode = ""
      },
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
      padding: 0 10px 10px 0;
      display: flex;
      flex-flow: row wrap;
      justify-content: flex-end;
      width: 50%;
      margin: 0 auto;
      .button {
         margin-left: 10px;
      }
   }
   .content {
      padding: 10px;
      display: flex;
      flex-flow: row wrap;
      justify-content: space-between;
      width: 50%;
      margin: 0 auto;
      .entry {
         flex-grow: 1;
         margin: 0 10px;
         text-align: left;
         label {
            display: block;
            margin: 0 0 5px 0;
         }
      }
   }
   .error {
      font-style: italic;
      color: var(--uvalib-red-emergency);
      margin: 0;
   }
}
</style>
