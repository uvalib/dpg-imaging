<template>
   <UModal :modal="true" :close="false" :dismissible="false" :title="props.title">
      <template #body>
         <div v-html="props.message"></div>
      </template>
      <template #footer>
         <UButton @click="emit('close', false)" label="Cancel" color="secondary"/>
         <UButton @click="emit('close', true)" :label="props.button"/>
      </template>
   </UModal>
</template>

<script setup>
const emit = defineEmits( ['close' ])

const props = defineProps({
   title: {
      type: String,
      required: true,
   },
   message: {
      type: String,
      required: true,
   },
   button: {
      type: String,
      default: "OK"
   }
})
</script>

<style lang="scss" scoped>
</style>
