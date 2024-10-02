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

<script setup>
import { ref } from 'vue'

const props = defineProps( ['modelValue'] )
const emit = defineEmits( ['update:modelValue', 'accepted', 'canceled'] )

const editVal = ref(props.modelValue)
const dropdownOpen = ref(false)
const vocab = ref(["Spine", "Front Cover", "Front Cover verso", "Back Cover", "Back Cover recto",
         "Head", "Tail", "Fore-edge", "Front Paste-down Endpaper", "Front Free Endpaper",
         "Rear Paste-down Endpaper", "Rear Free Endpaper", "Plate",
         "Blank Page", "Half-title", "Frontispiece", "Printer's Imprint", "Copyright",
         "Privilege", "Ad Lectorem",  "Table of Contents", "Titlepage", "Device",
         "Epigraph", "Prologue/Preface", "Dedication", "Errata"].sort())

function showDropdown() {
   dropdownOpen.value = true
}
function hideDropdown() {
   dropdownOpen.value = false
}
function acceptEdit() {
   hideDropdown()
   emit("accepted")
}
function cancelEdit() {
   hideDropdown()
   emit("canceled")
}
function vocabSelected(val) {
   editVal.value = val
   emit('update:modelValue', val)
}
function handleInput(newVal) {
   editVal.value = newVal
   emit('update:modelValue', newVal)
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
            white-space: nowrap;
            padding: 2px;

            .vocab {
               display: block;
            }
            &:hover {
               background: var(--uvalib-blue-alt-light);
               color: var(--uvalib-text);
            }
         }
      }
   }
}
</style>