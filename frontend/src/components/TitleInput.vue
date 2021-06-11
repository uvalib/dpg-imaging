<template>
   <div class="title-input">
      <input id="title-input-box" type="text"
         :value="editVal"
         @input="handleInput($event.target.value)"
         @focus="showDropdown"
         @keypress.enter="acceptEdit" @keydown.tab="cancelEdit" @keydown.esc="cancelEdit"/>
      <div v-if="dropdownOpen" class="title-vocab-wrap">
         <ul>
            <li v-for="(v,idx) in vocab" :key="`v${idx}`">
               <span class="vocab" @click.stop.prevent="vocabSelected(v)">{{v}}</span>
            </li>
         </ul>
      </div>
   </div>
</template>

<script>
export default {
   props: ['value'],
   data() {
      return {
         editVal: this.value,
         dropdownOpen: false,
         vocab: ["Spine", "Front Cover", "Front Cover verso", "Back Cover", "Back Cover recto",
                 "Head", "Tail", "Fore-edge", "Front Paste-down Endpaper", "Front Free Endpaper page #",
                 "Blank Page", "Half-title", "Frontispiece", "Printer’s Imprint", "Copyright",
                 "Privilege", "Ad Lectorem",  "Table of Contents", "Titlepage", "Device",
                 "Epigraph", "Prologue/Preface", "Dedication", "Errata"]
      }
   },
   methods: {
      showDropdown() {
         this.dropdownOpen = true
      },
      hideDropdown() {
         this.dropdownOpen = false
      },
      acceptEdit() {
         this.hideDropdown()
         this.$emit("accepted")
      },
      cancelEdit() {
         this.hideDropdown()
         this.$emit("canceled")
      },
      vocabSelected(val) {
         this.editVal = val
         this.$emit('input', this.editVal)
         document.getElementById("title-input-box").focus()

      },
      handleInput(newVal) {
         this.editVal = newVal
         this.$emit('input', newVal)
      },
   },
   mounted() {
      let ele = document.getElementById("title-input-box")
      if ( ele ) {
         ele.focus()
         ele.select()
      }
   }
}
</script>

<style lang="scss" scoped>
.title-input {
   position: relative;
   .title-vocab-wrap {
      color: var(--uvalib-text);
      position: absolute;
      top: 20px;
      left: 0;
      z-index: 1000;
      background: white;
      padding: 10px;
      border: 1px solid var(--uvalib-grey);
      box-shadow: var(--box-shadow);
      border-radius: 0 5px 5px 5px;
      ul {
         list-style: none;
         margin: 0;
         padding: 0;
         li {
            .vocab {
               display: block;
            }
            white-space: nowrap;
            padding: 2px;
            &:hover {
               background: var(--uvalib-blue-alt-light);
               color: var(--uvalib-text);
            }
         }
      }
   }
}
</style>