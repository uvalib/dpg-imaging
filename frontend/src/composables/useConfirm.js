import ConfirmModal from "../components/ConfirmModal.vue"

export async function useConfirm( title, message, button ) {
   const confirm =  useOverlay().create(ConfirmModal, {destroyOnClose: true})
   const opened = confirm.open({title: title, message: message, button: button})
   return await opened.result
}