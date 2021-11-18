<template>
   <div class="panel">
      <dl v-if="!editing">
         <dt>Category:</dt>
         <dd>{{currProject.category.name}}
         </dd>
         <dt>Call Number:</dt>
         <dd>{{currProject.unit.metadata.callNumber}}</dd>
         <dt>Special Instructions:</dt>
         <dd>
            <span v-if="currProject.unit.specialInstructions">{{currProject.unit.specialInstructions}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>Condition:</dt>
         <dd>{{conditionText(currProject.itemCondition)}}</dd>
         <dt>Condition Notes:</dt>
         <dd>
            <span v-if="currProject.conditionNote">{{currProject.conditionNote}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>OCR Hint:</dt>
         <dd>
            <span v-if="currProject.unit.metadata.ocrHint.id > 0">{{currProject.unit.metadata.ocrHint.name}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>OCR Language Hint:</dt>
         <dd>
            <span v-if="currProject.unit.metadata.ocrLanguageHint">{{currProject.unit.metadata.ocrLanguageHint}}</span>
            <span v-else class="na">EMPTY</span>
         </dd>
         <dt>OCR Master Files:</dt>
         <dd>
            <span v-if="currProject.unit.ocrMasterFiles" class="yes-no">Yes</span>
            <span v-else class="yes-no">No</span>
         </dd>
      </dl>
      <table class="edit" v-else>
         <tr class="row">
            <td class="label"><label for="category">Category:</label></td>
            <td class="data">
               <select id="category" v-model="categoryID">
                  <option :value="0" disabled>Select a category</option>
                  <option v-for="c in categories" :key="`cat${c.id}`" :value="c.id">{{c.name}}</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="call-numbber">Call Number:</label></td>
            <td class="data">{{currProject.unit.metadata.callNumber}}</td>
         </tr>
         <tr class="row">
            <td class="label"><label for="instructions">Special Instructions:</label></td>
            <td class="data">
               <span v-if="currProject.unit.specialInstructions">{{currProject.unit.specialInstructions}}</span>
               <span v-else class="na">EMPTY</span>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="condition">Condition:</label></td>
            <td class="data">
               <select id="condition" v-model="condition">
                  <option :value="0">Good</option>
                  <option :value="1">Bad</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="notes">Condition Notes:</label></td>
            <td class="data"><textarea id="notes" v-model="note"></textarea></td>
         </tr>
         <tr class="row">
            <td class="label"><label for="ocr-hint">OCR Hint:</label></td>
            <td class="data">
               <select id="ocr-hint" v-model="ocrHintID" @change="hintChanged">
                  <option :value="0" disabled>Select an OCR hint</option>
                  <option v-for="h in ocrHints" :key="`ocr${h.id}`" :value="h.id">{{h.name}}</option>
               </select>
            </td>

         </tr>
         <tr class="row">
            <td class="label"><label :class="{disabled: !ocrCandidate}" for="ocr-language">OCR Language Hint:</label></td>
            <td class="data">
               <select id="ocr-language" v-model="ocrLangage" :class="{disabled: !ocrCandidate}" :disabled="!ocrCandidate">
                  <option value="" disabled>Select an OCR language hint</option>
                  <option v-for="h in ocrLanguageHints" :key="`lang${h.code}`" :value="h.code">{{h.language}}</option>
               </select>
            </td>
         </tr>
         <tr class="row">
            <td class="label"><label for="do-ocr" :class="{disabled: !ocrCandidate}">OCR Master Files:</label></td>
            <td class="data"><input type="checkbox" :class="{disabled: !ocrCandidate}" id="do-ocr" v-model="ocrMasterFiles" :disabled="!ocrCandidate"></td>
         </tr>
      </table>
      <div class="buttons" v-if="isOwner(computingID)">
         <DPGButton v-if="!editing" @click="editClicked">Edit</DPGButton>
         <template v-else>
            <DPGButton @click="cancelClicked">Cancel</DPGButton>
            <DPGButton @click="saveClicked">Save</DPGButton>
         </template>
      </div>
   </div>
</template>

<script>
import { mapState, mapGetters } from "vuex"
export default {
   data: function()  {
      return {
         editing: false,
         categoryID: 0,
         condition: 0,
         note: "",
         ocrHintID: 0,
         ocrLangage: "",
         ocrMasterFiles: false,
         ocrCandidate: true
      }
   },
   computed: {
      ...mapState({
         computingID: state => state.user.computeID,
         categories: state => state.categories,
         ocrHints: state => state.ocrHints,
         ocrLanguageHints: state => state.ocrLanguageHints,
      }),
      ...mapGetters({
         currProject: 'projects/currProject',
         isOwner: "projects/isOwner"
      }),
   },
   methods: {
      hintChanged() {
         this.ocrLangage = ""
         let hint = this.ocrHints.find(h => h.id == this.ocrHintID)
         this.ocrCandidate =  hint.ocrCandidate
         if ( !this.ocrCandidate) {
            this.ocrMasterFiles = false
         }
      },
      conditionText(condID) {
         if (condID == 0) return "Good"
         return "Bad"
      },
      editClicked() {
         this.editing = true
         this.categoryID = this.currProject.category.id
         this.condition = this.currProject.itemCondition
         this.note = this.currProject.conditionNote
         this.ocrHintID = this.currProject.unit.metadata.ocrHint.id
         this.ocrLangage = this.currProject.unit.metadata.ocrLanguageHint
      },
      cancelClicked() {
         this.editing = false
      },
      async saveClicked() {
         let data = {
            categoryID: this.categoryID,
            condition: this.condition,
            note: this.note,
            ocrHintID: this.ocrHintID,
            ocrLangage: this.ocrLangage,
            ocrMasterFiles: this.ocrMasterFiles
         }
         await this.$store.dispatch("projects/updateProject", data)
         this.editing = false
      }
   },
};
</script>

<style scoped lang="scss">
.panel {
   padding: 10px;
   text-align: left;
   .na {
      color: #999;
   }
   dl {
      margin: 10px 30px 0 30px;
      display: inline-grid;
      grid-template-columns: max-content 2fr;
      grid-column-gap: 10px;
      font-size: 0.9em;
      text-align: left;
      box-sizing: border-box;

      dt {
         font-weight: bold;
         text-align: right;
      }
      dd {
         margin: 0 0 10px 0;
         word-break: break-word;
         -webkit-hyphens: auto;
         -moz-hyphens: auto;
         hyphens: auto;
      }
   }
   .edit {
      font-size: 0.9em;
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 5px;
      td {
          padding: 5px 0px 5px 10px;
      }
      td.data {
         width: 100%;
         input, select {
            border-color: var(--uvalib-grey-light);
         }
         input[type=checkbox] {
            width: 15px;
            height: 15px;
         }
      }
      td.label {
         font-weight: bold;
         margin-right: 10px;
         text-align: right;
         white-space: nowrap;
         vertical-align: top;
      }
      .disabled {
         opacity: 0.5;
      }
      textarea {
         width: 100%;
         box-sizing: border-box;
         border-color: var(--uvalib-grey-light);
         border-radius: 5px;;
      }
   }
   .buttons {
      padding: 0;
      margin: 0;
      text-align: right;
      button {
         margin-left: 10px;
      }
   }
}
</style>
