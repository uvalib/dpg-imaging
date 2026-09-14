<template>
   <USelect v-model="unitStore.viewMode" @change="viewModeChanged" :items="views" />
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { useRoute, useRouter } from 'vue-router'
import { ref } from 'vue'

const route = useRoute()
const router = useRouter()
const unitStore = useUnitStore()

const views = ref([
   {label: "View: List", value: "list"},
   {label: "View: Gallery (medium)", value: "medium"},
   {label: "View: Gallery (large)", value: "large"},
])

const viewModeChanged = (() => {
   unitStore.deselectAll()
   let query = Object.assign({},route.query)
   query.view = unitStore.viewMode
   router.push({query})
})
</script>

<style lang="scss" scoped>
</style>