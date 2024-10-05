<template>
   <Select v-model="unitStore.viewMode" @change="viewModeChanged"
      :options="views" optionLabel="label" optionValue="value"
   />
</template>

<script setup>
import { useUnitStore } from "@/stores/unit"
import { useRoute, useRouter } from 'vue-router'
import Select from 'primevue/select'
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
   unitStore.pageSize = 20
   unitStore.currPage = 0
   let query = Object.assign({},route.query)
   query.view = unitStore.viewMode
   delete query.page
   delete query.pagesize
   router.push({query})
})
</script>

<style lang="scss" scoped>
</style>