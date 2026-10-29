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
    // `getBoundingClientRect` devuelve píxeles visuales, pero `left`/`top` de un
    // panel fixed se interpretan dentro del `zoom` de la página (tamaño de letra
    // «Grande» = 1,125): sin convertir, el panel queda corrido ese factor. El
    // factor real se mide con el propio panel (visual / CSS) en vez de leerlo.
    const box = panel.value.getBoundingClientRect()
    const zoom = panel.value.offsetWidth > 0 ? box.width / panel.value.offsetWidth : 1
    panel.value.style.setProperty('--picker-trigger-width', `${field.width / zoom}px`)
    const width = box.width
    const height = Math.min(panel.value.scrollHeight * zoom, window.innerHeight - 16)
    const visualLeft = Math.max(8, Math.min(field.left, window.innerWidth - width - 8))
    const below = field.bottom + 4
    const above = field.top - height - 4
    const visualTop = Math.max(
      8,
      Math.min(
        below + height <= window.innerHeight - 8 ? below : above,
        window.innerHeight - height - 8,
      ),
    )
    left.value = visualLeft / zoom
    top.value = visualTop / zoom
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
