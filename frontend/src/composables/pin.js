import { onMounted, ref, watch } from 'vue'
import { useWindowScroll, useElementBounding } from '@vueuse/core'

export function usePinnable( pinID, scrollID, pinCallback ) {
   const { y } = useWindowScroll()
   const toolbar = ref(null)
   const toolbarBounds = ref()
   const pinnedY = ref(-1)

   watch(y, (newY) => {
      const scrollBody = document.getElementById(scrollID)
      if ( pinnedY.value < 0) {
         if ( toolbarBounds.value.top <= 0 ) {
            pinnedY.value = y.value+toolbarBounds.value.top
            toolbar.value.classList.add("sticky")
            toolbar.value.style.width = `${toolbarBounds.value.width}px`
            scrollBody.style.top = `${toolbarBounds.value.height}px`
            pinCallback(true, toolbarBounds.value.height )
         }
      } else {
         if ( newY <=  pinnedY.value) {
            pinnedY.value = -1
            toolbar.value.classList.remove("sticky")
            toolbar.value.style.width = `auto`
            scrollBody.style.top = `0px`
            pinCallback(false, toolbarBounds.value.height)
         }
      }
   })

   onMounted( () => {
      toolbar.value = document.getElementById(pinID)
      toolbarBounds.value = useElementBounding( toolbar )
   })

   return {}
}