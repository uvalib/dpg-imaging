<template>
   <button tabindex="0" class="dpg-button" :class="{icon: mode=='icon'}"
      @keydown.exact.tab="tabNext($event)"
      @keydown.shift.tab="tabBack($event)"
      @click.prevent.stop="clicked" @keydown.prevent.stop.enter="clicked" @keydown.space.prevent.stop="clicked">
      <slot></slot>
   </button>
</template>

<script>
export default {
   props: {
      focusNextOverride: {
         type: Boolean,
         default: false
      },
      focusBackOverride: {
         type: Boolean,
         default: false
      },
      mode: {
         type: String,
         default: "button"
      }
   },
   methods: {
      clicked() {
         this.$nextTick( () => {
            this.$emit('click')
         })
      },
      tabBack(event) {
         if (this.focusBackOverride ) {
            event.stopPropagation()
            event.preventDefault()
            this.$emit('tabback')
         }
      },
      tabNext(event ) {
         if (this.focusNextOverride ) {
            event.stopPropagation()
            event.preventDefault()
            this.$emit('tabnext')
         }
      }
   }
}
</script>

<style lang="scss" scoped>
.dpg-button {
   border-radius: 5px;
   font-weight: normal;
   border: 1px solid var(--uvalib-grey);
   padding: 2px 12px;
   background: var(--uvalib-grey-lightest);
   cursor: pointer;
   font-size: 0.9em;
   transition: all 0.25s ease-out;
   outline: 0;
   &:hover {
      background: #fafafa;
   }
   &:focus {
      background: #f0f0ff;
      border: 1px solid var(--uvalib-blue-alt);
   }
}
.dpg-button.icon {
   border: none;
   background: transparent;
   padding: 2px 4px;
   font-size: 1em;
   &:hover {
      color: var(--uvalib-blue-alt);
   }
}
</style>