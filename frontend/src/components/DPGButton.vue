<template>
   <button tabindex="0" class="dpg-button" :class="{icon: mode=='icon', disabled: disabled}"
      @keydown.exact.tab="tabNext"
      @keydown.shift.tab="tabBack"
      @click.prevent.stop="clicked" @keydown.prevent.stop.enter="clicked" @keydown.space.prevent.stop="clicked">
      <slot></slot>
   </button>
</template>

<script setup>
const props = defineProps({
   focusNextOverride: {
      type: Boolean,
      default: false
   },
   focusBackOverride: {
      type: Boolean,
      default: false
   },
   disabled: {
      type: Boolean,
      default: false
   },
   mode: {
      type: String,
      default: "button"
   },
})

const emit = defineEmits( ['click', 'tabback', 'tabnext' ])

function clicked() {
   if (!props.disabled) {
      emit('click')
   }
}

function tabBack(event) {
   if (props.focusBackOverride ) {
      event.stopPropagation()
      event.preventDefault()
      emit('tabback')
   }
}

function tabNext(event) {
   if (props.focusNextOverride ) {
      event.stopPropagation()
      event.preventDefault()
      emit('tabnext')
   }
}

</script>

<style lang="scss" scoped>
.dpg-button {
   border-radius: 5px;
   font-weight: normal;
   border: 1px solid var(--uvalib-grey);
   padding: 2px 12px 3px 12px;
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
.dpg-button.disabled {
   cursor: default !important;
   opacity: 0.5;
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