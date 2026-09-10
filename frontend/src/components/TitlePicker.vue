<template>
   <UFieldGroup>
      <template v-if="showTitleVocab">
         <USelect :items="system.titleVocab" class="w-full" placeholder="Pick a title" v-model="model" @update:model-value="emit('submit')" @keydown.tab="emit('cancel')"/>
         <UButton icon="i-lucide-undo" size="xs" color="secondary" @click="toggleVocabClicked" />
      </template>
      <template v-else>
         <UInput v-model="model"  class="w-full"
            @keydown.enter="emit('submit')" @keydown.esc="emit('cancel')" @keydown.tab="emit('cancel')" />
         <UButton icon="i-lucide-search" size="xs" color="secondary" @click="toggleVocabClicked" />
      </template>
   </UFieldGroup>
</template>

<script setup>
import { ref } from "vue"
import { useSystemStore } from "@/stores/system"

const system = useSystemStore()
const showTitleVocab = ref(false)
const prior = ref("")

const model = defineModel()

const emit = defineEmits( ['cancel', 'submit'] )

const toggleVocabClicked = (() => {
   if (showTitleVocab.value == false ) {
      prior.value =  model.value
      model.value = ""
   } else {
      model.value = prior.value 
      prior.value = ""
   }
   showTitleVocab.value = !showTitleVocab.value
})
</script>

<style lang="scss" scoped>
</style>
