<template>
   <div class="title-picker">
      <InputGroup v-if="showTitleVocab">
         <Select :options="system.titleVocab" autofocus placeholder="Pick a title" v-model="model" @update:model-value="emit('submit')" @keydown.tab="emit('cancel')"/>
         <InputGroupAddon>
            <DPGButton icon="pi pi-undo" severity="secondary" variant="text" @click="toggleVocabClicked" />
         </InputGroupAddon>
      </InputGroup>
      <InputGroup v-else>
         <InputText v-model="model" :readonly="showTitleVocab" autofocus
            @keydown.enter="emit('submit')" @keydown.esc="emit('cancel')" @keydown.tab="emit('cancel')" />
         <InputGroupAddon>
            <DPGButton icon="pi pi-search" severity="secondary" variant="text" @click="toggleVocabClicked" />
         </InputGroupAddon>
      </InputGroup>
   </div>
</template>

<script setup>
import { ref } from "vue"
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import { useSystemStore } from "@/stores/system"

const system = useSystemStore()
const showTitleVocab = ref(false)

const model = defineModel()

const emit = defineEmits( ['cancel', 'submit'] )

const toggleVocabClicked = (() => {
   showTitleVocab.value = !showTitleVocab.value
})

</script>
