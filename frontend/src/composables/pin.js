import { onMounted, ref, watch } from 'vue'
import { useWindowScroll, useElementBounding } from '@vueuse/core'

export function usePinnable( pinID, scrollID, pinCallback ) {
   const { y } = useWindowScroll()
   const scrollBody = ref()
   const toolbar = ref(null)
   const toolbarBounds = ref()
   const pinnedY = ref(-1)

   watch(y, (newY) => {
      if ( pinnedY.value < 0) {
         if ( toolbarBounds.value.top <= 0 ) {
            pinnedY.value = y.value+toolbarBounds.value.top
            toolbar.value.classList.add("sticky")
            toolbar.value.style.width = `${toolbarBounds.value.width}px`
            scrollBody.value.style.top = `${toolbarBounds.value.height}px`
            if ( pinCallback) {
               pinCallback(true, toolbarBounds.value.height )
            } else {
               scrollBody.value.style.top = `${toolbarBounds.value.height}px`
            }
         }
      } else {
         if ( newY <=  pinnedY.value) {
            pinnedY.value = -1
            toolbar.value.classList.remove("sticky")
            toolbar.value.style.width = `auto`
            scrollBody.value.style.top = `0px`
            if ( pinCallback ) {
               pinCallback(false, toolbarBounds.value.height)
            } else {
               scrollBody.value.style.top = `auto`
            }
         }
      }
   })

   onMounted( () => {
      if ( scrollID ) {
         scrollBody.value = document.getElementById(scrollID)
         toolbar.value = document.getElementById(pinID)
      } else {
         // When scrollID is not specified, pinnable is being used on a datatable/dataview to
         // pin the pagination toolbar. need to look up classes that are children of the
         // datatable ID. Paginatior class is p-datatable-paginator-top, table body is either
         // p-datatable-table-container (DataTable) or p-dataview-content (DataView)
         toolbar.value = document.querySelector(`#${pinID} .p-paginator`)
         let sb = document.querySelector(`#${pinID} .p-datatable-table-container`)
         if ( !sb ) {
            sb = document.querySelector(`#${pinID} .p-dataview-content`)
         }
         scrollBody.value = sb
         scrollBody.value.style.position = 'relative'
      }

      toolbarBounds.value = useElementBounding( toolbar )
   })

   return {}
}