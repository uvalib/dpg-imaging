<template>
   <div class="tag-picker">
      <span tabindex="0" @click="showMenu" @keydown.enter="showMenu" class="tag current" :class="masterFile.status"></span>
      <div class="popup-list" v-if="menuOpen">
         <div class="title">
            <span>Select a tag</span>
            <span tabindex="0" @click.stop.prevent="hideMenu()" class="close">X</span>
         </div>
         <ul>
            <li @click.stop.prevent="selectTag('rescan')">
               <span class="tag rescan"></span>
               <span class="label">Rescan</span>
            </li>
            <li @click.stop.prevent="selectTag('good')">
               <span class="tag good"></span>
               <span class="label">Good</span>
            </li>
            <li @click.stop.prevent="selectTag('recrop')">
               <span class="tag recrop"></span>
               <span class="label">Recrop/Rotate</span>
            </li>
            <li @click.stop.prevent="selectTag('color')">
               <span class="tag color"></span>
               <span class="label">Color Issues</span>
            </li>
            <li @click.stop.prevent="selectTag('other')">
               <span class="tag other"></span>
               <span class="label">Other issues</span>
            </li>
            <li @click.stop.prevent="selectTag('placeholder')">
               <span class="tag placeholder"></span>
               <span class="label">Placeholder</span>
            </li>
            <li @click.stop.prevent="selectTag('none')">
               <span class="tag none">X</span>
               <span class="label">Remove Tag</span>
            </li>
         </ul>
      </div>
   </div>
</template>

<script>
export default {
   props: {
      masterFile: {
         type: Object,
         required: true
      },
   },
   data() {
      return {
        menuOpen: false,
      }
   },
   methods: {
      showMenu() {
         this.menuOpen = true
      },
      hideMenu() {
         this.menuOpen = false
      },
      async selectTag( tag ) {
         await this.$store.dispatch("setTag", {file: this.masterFile.path, tag: tag})
         this.hideMenu()
      }
   }
}
</script>

<style lang="scss" scoped>
.tag-picker {
   position: relative;
   .popup-list {
      position: absolute;
      background: white;
      z-index: 1000;
      border: 1px solid var(--uvalib-blue-alt-dark);
      box-shadow: var(--box-shadow);
      top: 0;

      .title {
         background: var(--uvalib-grey-light);
         padding: 5px;
         border-bottom: 1px solid var(--uvalib-grey);
         display: flex;
         flex-flow: flex nowrap;
         justify-content: space-between;
         .close {
            background: var(--uvalib-red-dark);
            font-size: 16px;
            color: white;
            font-weight: bolder;
            border: 1px solid var(--uvalib-red-darker);
            height: 16px;
            width: 16px;
            text-align: center;
            border-radius: 10px;
            cursor: pointer;
            &:hover {
               background: firebrick;
            }
         }
      }

      ul {
         list-style: none;
         margin: 0;
         padding: 0;
         li {
            white-space: nowrap;
            display: flex;
            flex-flow: row nowrap;
            align-content: center;
            margin: 10px;
            cursor: pointer;
            border: 1px solid  white;

            .label {
               display: block;
               margin: 0 10px;
               position: relative;
               top: 3px;
            }

            &:hover {
               background: var(--uvalib-blue-alt-light);
               border: 1px solid var(--uvalib-blue-alt);

            }
         }
      }
   }
   .current {
      cursor: pointer;
      display: none;
   }

   .tag {
      display: block;
      width: 20px;
      height: 20px;
      border: 1px solid var(--uvalib-grey-light);
      background: white;
   }
   .tag.rescan {
      background: var(--uvalib-red);
      border-color: var(--uvalib-red-darkest);
   }
   .tag.good {
      background: var(--uvalib-green);
      border-color: var(--uvalib-green-dark);
   }
   .tag.recrop {
      background: var(--uvalib-blue-alt);
      border-color: var(--uvalib-blue-alt-dark);
   }
   .tag.color {
      background: magenta;
      border-color: darkmagenta;
   }
   .tag.other {
      background: var(--uvalib-yellow);
      border-color: var(--uvalib-yellow-dark);
   }
   .tag.placeholder {
      background: var(--uvalib-teal);
      border-color: var(--uvalib-teal-dark);
   }
   .tag.none {
      text-align: center;
      font-size: 20px;
      font-weight: 100;
      color: var(--uvalib-grey-light);
   }
}
</style>