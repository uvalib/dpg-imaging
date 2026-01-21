<template>
   <div class="title-picker">
      <InputGroup v-if="showTitleVocab">
         <Select :options="system.titleVocab" placeholder="Pick a title" v-model="model"/>
         <InputGroupAddon>
            <DPGButton icon="pi pi-trash" severity="secondary" variant="text" @click="model=''" />
            <DPGButton icon="pi pi-undo" severity="secondary" variant="text" @click="toggleVocabClicked" />
         </InputGroupAddon>
      </InputGroup>
      <InputGroup v-else>
         <InputText v-model="model" :readonly="showTitleVocab" />
         <InputGroupAddon>
            <DPGButton icon="pi pi-trash" severity="secondary" variant="text" @click="model=''" />
            <DPGButton icon="pi pi-search" severity="secondary" variant="text" @click="toggleVocabClicked" />
         </InputGroupAddon>
      </InputGroup>
   </div>
</template>

<script setup>
import { computed, ref } from "vue"
import InputGroup from 'primevue/inputgroup'
import InputGroupAddon from 'primevue/inputgroupaddon'
import Select from 'primevue/select'
import InputText from 'primevue/inputtext'
import { useSystemStore } from "@/stores/system"

const system = useSystemStore()
const showTitleVocab = ref(false)

const model = defineModel()

const titleIcon = computed(() => {
   if ( showTitleVocab.value ) return "pi pi-times"
   return "pi pi-search"
})

const toggleVocabClicked = (() => {
   showTitleVocab.value = !showTitleVocab.value

})


</script>
