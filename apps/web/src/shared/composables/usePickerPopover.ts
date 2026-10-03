import { computed, nextTick, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'

/** Panel flotante fuera del scroll del formulario: cabe en el viewport y
 * sigue al campo cuando se desplaza su diálogo, sin mover la página. */
export function usePickerPopover(
  open: Ref<boolean>,
  root: Ref<HTMLElement | null>,
  trigger: Ref<HTMLButtonElement | null>,
  panel: Ref<HTMLElement | null>,
  close: (restoreFocus?: boolean) => void,
) {
  const left = ref(8)
  const top = ref(8)
  const style = computed(() => ({ left: `${left.value}px`, top: `${top.value}px` }))
  function reposition() {
    if (!open.value || !trigger.value || !panel.value) return
    const field = trigger.value.getBoundingClientRect()
    panel.value.style.setProperty('--picker-trigger-width', `${field.width}px`)
    const width = panel.value.offsetWidth
    const height = Math.min(panel.value.scrollHeight, window.innerHeight - 16)
    left.value = Math.max(8, Math.min(field.left, window.innerWidth - width - 8))
    const below = field.bottom + 4
    const above = field.top - height - 4
    top.value = Math.max(
      8,
      Math.min(
        below + height <= window.innerHeight - 8 ? below : above,
        window.innerHeight - height - 8,
      ),
    )
  }
  function onOutside(event: MouseEvent) {
    if (!open.value || !(event.target instanceof Node)) return
    if (!root.value?.contains(event.target) && !panel.value?.contains(event.target)) close(false)
  }
  watch(open, async (value) => {
    if (value) {
      await nextTick()
      reposition()
    }
  })
  onMounted(() => {
    document.addEventListener('mousedown', onOutside)
    document.addEventListener('scroll', reposition, true)
    window.addEventListener('resize', reposition)
  })
  onUnmounted(() => {
    document.removeEventListener('mousedown', onOutside)
    document.removeEventListener('scroll', reposition, true)
    window.removeEventListener('resize', reposition)
  })
  return { style, reposition }
}
