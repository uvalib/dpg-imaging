<template>
   <button tabindex="0" class="dpg-button"
      @click.prevent.stop="clicked" @keydown.prevent.stop.enter="clicked" @keydown.space.prevent.stop="clicked">
      <slot></slot>
   </button>
</template>

<script>
export default {
   methods: {
      clicked() {
         this.$nextTick( () => {
            this.$emit('click')
         })
      },
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
</style>