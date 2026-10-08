<script setup lang="ts">
// Pantalla "Barberos" (HU-021, DEC-104): consultar y listar el equipo de la
// barbería activa, agregar un barbero, renombrarlo y ponerle (o quitarle) una
// fotografía. Una barbería unipersonal y una de cuatro personas usan
// exactamente el mismo componente y el mismo estado de datos (CA-021-01/02): la
// lista con 1 elemento y la lista con 4 no tienen ninguna rama especial.
// Estados discriminados: carga inicial, listo, vacío, error recuperable,
// guardando (trabajo requerido §4.3/§4.4). Un error recuperable NUNCA borra lo
// que el barbero ya escribió ni la foto que ya eligió; solo un guardado exitoso
// confirmado por el servidor cierra el diálogo. Cada guardado confirmado añade
// un aviso emergente (DEC-095); los errores siguen dentro del diálogo, junto al
// formulario. Mismo lenguaje visual que Servicios (tinta, latón, tabla hundida,
// entrada escalonada, carga con el rombo) con el retrato del barbero como
// protagonista.
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { PAGE_MIN_HOLD_MS, useMinHoldLoading, useToast, useVocabulary } from '@/shared/composables'
import {
  BarberAvatar,
  BaseAlert,
  EmptyScene,
  BaseButton,
  BaseDialog,
  BaseInput,
  DiamondLoader,
} from '@/shared/ui'
import {
  barberPhotoUrl,
  createBarber,
  fetchBarberPage,
  removeBarberPhoto,
  renameBarber,
  uploadBarberPhoto,
} from '../api/staffApi'
import BarberPhotoField from '../components/BarberPhotoField.vue'
import MyProfileCard from '../components/MyProfileCard.vue'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type { Barber } from '../model/barber'
import { NO_PHOTO_CHANGE, type PhotoDraft } from '../model/photoDraft'
import { capitalize, isSoloProfile } from '@/shared/model'
import { validateFullName } from '../validation/staffValidation'

const toast = useToast()
// Palabras de la barbería (DEC-110): con los valores iniciales, el mismo texto de siempre.
const v = useVocabulary()

// Frases decorativas del rombo de carga: las de la casa, con voz de equipo.
const LOADING_PHRASES = [
  'Reuniendo al equipo',
  'Afilando la navaja',
  'Alineando los turnos',
  'Todo a su hora',
] as const

type LoadStatus = 'loading' | 'ready' | 'load-error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'idempotency-conflict'
  | 'not-found'
  | 'photo-error'
  | 'network-error'
  | 'unexpected-error'

const loadStatus = ref<LoadStatus>('loading')
const barbers = ref<Barber[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const totalItems = ref(0)
const totalPages = ref(1)
const pageLoading = ref(false)
const pageFailed = ref(false)
const justAddedId = ref<string | null>(null)
const listRef = ref<HTMLElement | null>(null)
const footerRef = ref<HTMLElement | null>(null)
const pageRef = ref<HTMLElement | null>(null)
const fitPending = ref(true)
const viewportWidth = ref(window.innerWidth)
const { start: startLoadingHold, hold: holdLoadingReveal } = useMinHoldLoading()
const { start: startPageHold, hold: holdPageReveal } = useMinHoldLoading(PAGE_MIN_HOLD_MS)
let requestToken = 0
let lastRequestedPage = 1
let pendingPaginationFocus = false
let resizeTimer: ReturnType<typeof setTimeout> | undefined
let observer: ResizeObserver | undefined

async function load(page = 1, initial = false) {
  if (initial) pendingPaginationFocus = false
  else if (
    document.activeElement instanceof HTMLElement &&
    document.activeElement.closest('.staff-page__pagination-nav')
  )
    pendingPaginationFocus = true
  const token = ++requestToken
  lastRequestedPage = page
  if (initial) {
    loadStatus.value = 'loading'
    startLoadingHold()
  } else {
    pageLoading.value = true
    pageFailed.value = false
    startPageHold()
  }
  const outcome = await fetchBarberPage(page, pageSize.value)
  if (token !== requestToken) return
  const reveal = () => {
    if (token !== requestToken) return
    pageLoading.value = false
    if (outcome.kind === 'success') {
      barbers.value = outcome.page.items
      currentPage.value = outcome.page.page
      pageSize.value = outcome.page.pageSize
      totalItems.value = outcome.page.total
      totalPages.value = outcome.page.totalPages
      loadStatus.value = 'ready'
      // La tarjeta del perfil individual no pagina ni mide filas de tabla
      // (DEC-115): sin esto, fitPending se quedaría en true para siempre
      // porque listRef nunca se monta en ese caso.
      if (barbers.value.length === 0 || soloCard.value) fitPending.value = false
    } else if (initial) loadStatus.value = 'load-error'
    else pageFailed.value = true
    if (pendingPaginationFocus) {
      void nextTick(() => {
        if (token !== requestToken) return
        const active = document.activeElement
        if (
          active === document.body ||
          (active instanceof HTMLElement && active.closest('.staff-page__pagination-nav'))
        ) {
          footerRef.value?.querySelector<HTMLButtonElement>('[aria-current="page"]')?.focus()
        }
        pendingPaginationFocus = false
      })
    }
  }
  if (initial) holdLoadingReveal(reveal)
  else holdPageReveal(reveal)
}

function scrollParent(el: HTMLElement): HTMLElement | null {
  let node = el.parentElement
  while (node) {
    if (['auto', 'scroll'].includes(getComputedStyle(node).overflowY)) return node
    node = node.parentElement
  }
  return null
}

// DEC-107: mide el visor real, las filas y el pie; no oculta overflow para
// aparentar que caben filas. El padding cambia en móvil y también se mide.
async function fitPage() {
  await nextTick()
  const list = listRef.value,
    footer = footerRef.value,
    page = pageRef.value
  if (!list || !footer || !page) {
    fitPending.value = false
    return
  }
  const parent = scrollParent(list)
  if (!parent) {
    fitPending.value = false
    return
  }
  const row = list.querySelector<HTMLElement>('.staff-page__row, .staff-page__skeleton-row')
  const height = row?.getBoundingClientRect().height
  if (!height) {
    fitPending.value = false
    return
  }
  const listRect = list.getBoundingClientRect(),
    footerRect = footer.getBoundingClientRect()
  const gap = Math.max(0, footerRect.top - listRect.bottom)
  const bottomPadding = parseFloat(getComputedStyle(page).paddingBottom)
  const available =
    parent.clientHeight -
    (listRect.top - parent.getBoundingClientRect().top + parent.scrollTop) -
    gap -
    footerRect.height -
    bottomPadding -
    2
  const fitting = Math.max(1, Math.min(50, Math.floor(available / height)))
  if (fitting !== pageSize.value) {
    const firstIndex = (currentPage.value - 1) * pageSize.value
    pageSize.value = fitting
    const targetPage = Math.floor(firstIndex / fitting) + 1
    // Recortar la primera carga evita un segundo lote grande en pantalla.
    if (currentPage.value === 1 && !pageLoading.value && fitting <= barbers.value.length) {
      barbers.value = barbers.value.slice(0, fitting)
      totalPages.value = Math.max(1, Math.ceil(totalItems.value / fitting))
    } else {
      await load(targetPage)
    }
  }
  parent.scrollTop = 0
  fitPending.value = false
}

function scheduleFit() {
  viewportWidth.value = window.innerWidth
  if (resizeTimer !== undefined) clearTimeout(resizeTimer)
  resizeTimer = setTimeout(() => {
    void fitPage()
  }, 120)
}
watch(
  listRef,
  async (el) => {
    if (!el) return
    await fitPage()
    if (typeof ResizeObserver !== 'undefined') {
      observer?.disconnect()
      observer = new ResizeObserver(scheduleFit)
      const parent = scrollParent(el)
      if (parent) observer.observe(parent)
      const header = pageRef.value?.querySelector('.staff-page__header')
      if (header) observer.observe(header)
    }
  },
  { flush: 'post' },
)
onMounted(() => {
  void load(1, true)
  window.addEventListener('resize', scheduleFit)
})
onUnmounted(() => {
  ++requestToken
  observer?.disconnect()
  window.removeEventListener('resize', scheduleFit)
  if (resizeTimer !== undefined) clearTimeout(resizeTimer)
})
function onRetryLoad() {
  void load(1, true)
}
function rowIndex(index: number): number {
  return index
}
function goToPage(page: number) {
  if (page < 1 || page > totalPages.value || page === currentPage.value || pageLoading.value) return
  void load(page)
}
// Perfil de barbero individual (DEC-115): esta pantalla es la ficha de quien trabaja
// solo. `soloView` cambia los textos; `soloCard` (un único barbero ya cargado) además
// quita lo que solo tiene sentido con un equipo: «Agregar», encabezados de tabla y
// paginador. Con más de un barbero la pantalla vuelve a ser la de siempre.
const soloView = computed(
  () => isSoloProfile.value && !(loadStatus.value === 'ready' && totalItems.value > 1),
)
const soloCard = computed(
  () => soloView.value && loadStatus.value === 'ready' && totalItems.value === 1,
)

const countLabel = computed(
  () =>
    `${totalItems.value} ${totalItems.value === 1 ? v.value.professional : v.value.professionals}`,
)
const pageTokens = computed(() => {
  const total = totalPages.value,
    current = currentPage.value
  if (total <= 5) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = new Set(
    viewportWidth.value <= 640
      ? [1, total, current]
      : [1, total, current, Math.max(1, current - 1), Math.min(total, current + 1)],
  )
  const tokens: (number | 'ellipsis')[] = []
  for (const p of [...pages].sort((a, b) => a - b)) {
    const last = tokens[tokens.length - 1]
    if (typeof last === 'number' && p - last > 1) tokens.push('ellipsis')
    tokens.push(p)
  }
  return tokens
})

// Fotografía (DEC-104): la URL solo existe cuando el barbero tiene una.
function photoOf(barber: Barber): string | null {
  return barberPhotoUrl(barber)
}

function hasPhoto(barber: Barber): boolean {
  return barber.photoUpdatedAt !== null
}

// Monograma de la ficha del encabezado de un diálogo: la primera letra del
// nombre, la misma inicial que abre el retrato sin foto de la fila.
function firstInitial(fullName: string): string {
  return fullName.trim().charAt(0).toUpperCase()
}

// Ritmo propio del rombo de estado, igual técnica que Servicios/Agenda: hash
// estable del id para que la duración y el desfase del parpadeo no cambien al
// volver a pintar la lista y las filas no latan sincronizadas.
function diamondStyle(barber: Barber): Record<string, string> {
  let hash = 0
  for (const char of barber.id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  const spread = (hash % 1000) / 1000
  const phase = ((hash >>> 10) % 1000) / 1000
  const duration = 6 + spread * 5
  return {
    '--diamond-duration': `${duration.toFixed(1)}s`,
    '--diamond-delay': `-${(phase * duration).toFixed(1)}s`,
  }
}

// Fecha de alta. Zona fija de la moneda del contrato (COP, DEC-067), igual que
// Servicios, para que la fecha no dependa del huso del equipo.
function formatDate(iso: string | null | undefined): string {
  if (!iso) return ''
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return ''
  return new Intl.DateTimeFormat('es-CO', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    timeZone: 'America/Bogota',
  }).format(date)
}

function replaceInList(barber: Barber) {
  const index = barbers.value.findIndex((b) => b.id === barber.id)
  if (index !== -1) barbers.value[index] = barber
}

// --- Detalle --------------------------------------------------------------

// detailTarget no se borra al cerrar: el diálogo sigue visible mientras dura su
// animación de salida y vaciarlo aquí lo dejaría en blanco a media salida.
const isDetailOpen = ref(false)
const detailTarget = ref<Barber | null>(null)

function openDetailDialog(barber: Barber) {
  detailTarget.value = barber
  isDetailOpen.value = true
}

function onDetailEdit() {
  const barber = detailTarget.value
  if (!barber) return
  isDetailOpen.value = false
  openRenameDialog(barber)
}

// --- Alta -----------------------------------------------------------------

const isCreateOpen = ref(false)
const createFullName = ref('')
const createFieldError = ref<string | undefined>(undefined)
const createStatus = ref<SaveStatus>('idle')
const createAttempted = ref(false)
const createPhotoDraft = ref<PhotoDraft>(NO_PHOTO_CHANGE)
// Clave de idempotencia del intento lógico vigente (RN-IDE-01): se genera
// al abrir el diálogo y se REUTILIZA en cada reintento del mismo intento;
// solo un envío exitoso o cerrar y reabrir el diálogo la renueva (mismo
// intento lógico != mismo clic).
let createIdempotencyKey = newIdempotencyKey()

function openCreateDialog() {
  createFullName.value = ''
  createFieldError.value = undefined
  createStatus.value = 'idle'
  createAttempted.value = false
  createPhotoDraft.value = NO_PHOTO_CHANGE
  createIdempotencyKey = newIdempotencyKey()
  isCreateOpen.value = true
}

function onCreateDialogClosed() {
  // Cerrar sin guardar también es un intento lógico terminado: la próxima
  // apertura genera una clave nueva (ya cubierto por openCreateDialog,
  // aquí solo se limpia el estado visual para que no "parpadee" un error
  // viejo si se reabre).
  createStatus.value = 'idle'
}

function onCreateFullNameInput(value: string | number) {
  createFullName.value = String(value)
  if (createAttempted.value) createFieldError.value = validateFullName(createFullName.value)
}

async function onSubmitCreate() {
  if (createStatus.value === 'saving') return

  createAttempted.value = true
  const error = validateFullName(createFullName.value)
  createFieldError.value = error
  if (error) return

  createStatus.value = 'saving'
  const outcome = await createBarber(createFullName.value.trim(), createIdempotencyKey)

  switch (outcome.kind) {
    case 'success':
      await onBarberCreated(outcome.barber)
      return
    case 'validation-error':
      createStatus.value = 'validation-error'
      return
    case 'idempotency-conflict':
      createStatus.value = 'idempotency-conflict'
      return
    case 'network-error':
      createStatus.value = 'network-error'
      return
    case 'unexpected-error':
      createStatus.value = 'unexpected-error'
  }
}

// El barbero ya existe en el servidor: aunque la foto falle, el alta está
// hecha. Por eso el diálogo se cierra igual y el fallo de la foto se avisa
// aparte, con la salida (Editar) a la vista, en vez de dejar al barbero
// atrapado en un formulario que ya no puede repetir el alta.
async function onBarberCreated(created: Barber) {
  let barber = created
  const draft = createPhotoDraft.value
  let photoFailed = false

  if (draft.kind === 'new') {
    const photo = await uploadBarberPhoto(created.id, draft.blob)
    if (photo.kind === 'success') barber = photo.barber
    else photoFailed = true
  }

  totalItems.value += 1
  totalPages.value = Math.max(1, Math.ceil(totalItems.value / pageSize.value))
  await load(totalPages.value)
  justAddedId.value = barber.id
  isCreateOpen.value = false
  createStatus.value = 'idle'
  createPhotoDraft.value = NO_PHOTO_CHANGE

  if (photoFailed) {
    toast.warning(
      `${v.value.Professional} agregad${v.value.professionalEnding}, pero no pudimos guardar la foto`,
      {
        detail: 'Ábrelo con «Editar» para intentarlo de nuevo.',
      },
    )
    return
  }
  toast.success(`${v.value.Professional} agregad${v.value.professionalEnding}`, {
    detail: 'Ya aparece en tu equipo.',
  })
}

// --- Edición (nombre y foto) ------------------------------------------------

const isRenameOpen = ref(false)
const renameTarget = ref<Barber | null>(null)
const renameFullName = ref('')
const renameFieldError = ref<string | undefined>(undefined)
const renameStatus = ref<SaveStatus>('idle')
const renameAttempted = ref(false)
const renamePhotoDraft = ref<PhotoDraft>(NO_PHOTO_CHANGE)

function openRenameDialog(barber: Barber) {
  renameTarget.value = barber
  renameFullName.value = barber.fullName
  renameFieldError.value = undefined
  renameStatus.value = 'idle'
  renameAttempted.value = false
  renamePhotoDraft.value = NO_PHOTO_CHANGE
  isRenameOpen.value = true
}

function onRenameDialogClosed() {
  renameStatus.value = 'idle'
}

function onRenameFullNameInput(value: string | number) {
  renameFullName.value = String(value)
  if (renameAttempted.value) renameFieldError.value = validateFullName(renameFullName.value)
}

type StepFailure = 'validation-error' | 'not-found' | 'network-error' | 'unexpected-error'

// Un fallo de la foto (formato, tamaño) no es un error de campo del nombre: se
// muestra con su propio mensaje. Los demás se traducen igual que el nombre.
function failureStatus(kind: StepFailure, isPhoto: boolean): SaveStatus {
  if (kind === 'validation-error') return isPhoto ? 'photo-error' : 'validation-error'
  return kind
}

// Guardar aplica en orden lo que cambió: primero el nombre y luego la foto. Cada
// paso confirmado actualiza al barbero de la lista y el objetivo del diálogo,
// así un reintento tras un fallo a medias (el nombre ya quedó, la foto no)
// nunca repite lo que ya se guardó. Sin cambios no envía nada.
async function onSubmitRename() {
  const target = renameTarget.value
  if (renameStatus.value === 'saving' || !target) return

  renameAttempted.value = true
  const error = validateFullName(renameFullName.value)
  renameFieldError.value = error
  if (error) return

  const fullName = renameFullName.value.trim()
  const nameChanged = fullName !== target.fullName
  const draft = renamePhotoDraft.value
  if (!nameChanged && draft.kind === 'none') {
    isRenameOpen.value = false
    return
  }

  renameStatus.value = 'saving'
  let current = target

  if (nameChanged) {
    const renamed = await renameBarber(current.id, fullName)
    if (renamed.kind !== 'success') {
      renameStatus.value = failureStatus(renamed.kind, false)
      return
    }
    current = renamed.barber
    renameTarget.value = current
    replaceInList(current)
  }

  if (draft.kind === 'new') {
    const uploaded = await uploadBarberPhoto(current.id, draft.blob)
    if (uploaded.kind !== 'success') {
      renameStatus.value = failureStatus(uploaded.kind, true)
      return
    }
    current = uploaded.barber
    renameTarget.value = current
    replaceInList(current)
  } else if (draft.kind === 'remove') {
    const removed = await removeBarberPhoto(current.id)
    if (removed.kind !== 'success') {
      renameStatus.value = failureStatus(removed.kind, true)
      return
    }
    current = { ...current, photoUpdatedAt: null }
    renameTarget.value = current
    replaceInList(current)
  }

  isRenameOpen.value = false
  renameStatus.value = 'idle'
  renamePhotoDraft.value = NO_PHOTO_CHANGE

  if (nameChanged && draft.kind === 'none') {
    toast.success('Nombre actualizado', {
      detail: `Guardamos el nuevo nombre ${v.value.ofTheProfessional}.`,
    })
  } else if (!nameChanged && draft.kind === 'new') {
    toast.success('Foto actualizada', { detail: 'El nuevo retrato ya aparece en tu equipo.' })
  } else if (!nameChanged && draft.kind === 'remove') {
    toast.success('Foto quitada', {
      detail: `${capitalize(v.value.theProfessional)} vuelve a mostrarse con su monograma.`,
    })
  } else {
    toast.success(`${v.value.Professional} actualizad${v.value.professionalEnding}`, {
      detail: `Guardamos los cambios ${v.value.ofTheProfessional}.`,
    })
  }
}
</script>

<template>
  <section ref="pageRef" class="staff-page" aria-labelledby="staff-page-title">
    <header class="staff-page__header">
      <div>
        <h1 id="staff-page-title" class="staff-page__title">
          {{ soloView ? 'Mi perfil' : v.Professionals }}
        </h1>
        <p class="staff-page__subtitle">
          {{ soloView ? 'Tu ficha en NAVA: nombre, foto y bloqueos.' : 'Tu equipo en NAVA.' }}
        </p>
      </div>
      <!-- Una acción principal por región: con el equipo vacío, el CTA vive en el
           propio estado vacío y no se repite en la cabecera. -->
      <BaseButton
        v-if="loadStatus === 'ready' && barbers.length > 0 && !soloCard"
        type="button"
        variant="primary"
        class="staff-page__create"
        @click="openCreateDialog"
      >
        Agregar {{ v.professional }}
      </BaseButton>
    </header>

    <!-- Un solo fundido entre carga/error/listo (mismo criterio que Servicios):
         sin esto el paso del rombo a la lista real sería un corte seco.
         mode="out-in" espera a que lo anterior termine de desvanecerse. -->
    <Transition name="staff-content" mode="out-in">
      <div
        v-if="loadStatus === 'loading'"
        class="staff-page__state staff-page__loading"
        role="status"
        aria-live="polite"
      >
        <DiamondLoader label="Cargando el equipo…" layout="inline" :phrases="LOADING_PHRASES" />
        <ul class="staff-page__list" aria-hidden="true">
          <li v-for="n in 3" :key="n" class="staff-page__skeleton-row">
            <span class="staff-page__skeleton-portrait" />
            <span class="staff-page__skeleton-text">
              <span class="staff-page__skeleton-bar staff-page__skeleton-bar--name" />
              <span class="staff-page__skeleton-bar staff-page__skeleton-bar--meta" />
            </span>
            <span class="staff-page__skeleton-bar staff-page__skeleton-bar--button" />
          </li>
        </ul>
      </div>

      <BaseAlert
        v-else-if="loadStatus === 'load-error'"
        variant="warning"
        title="No pudimos cargar el equipo"
        role="alert"
      >
        Revisa tu conexión e inténtalo de nuevo.
        <template #action>
          <BaseButton variant="secondary" type="button" @click="onRetryLoad">Reintentar</BaseButton>
        </template>
      </BaseAlert>

      <div
        v-else
        class="staff-page__ready"
        :class="{ 'staff-page__ready--fitting': fitPending }"
        :aria-busy="pageLoading"
      >
        <EmptyScene v-if="barbers.length === 0" scene="team" class="staff-page__empty">
          <template #title>{{
            soloView ? 'Aún no tienes tu perfil.' : `Aún no tienes ${v.professionalsRegistered}.`
          }}</template>
          <template #hint>
            {{
              soloView
                ? 'Créalo con tu nombre y, si quieres, tu foto: tu agenda y tus horarios cuelgan de él.'
                : 'Agrega al primero de tu equipo: con su nombre y, si quieres, su foto.'
            }}
          </template>
          <template #action>
            <BaseButton type="button" variant="primary" @click="openCreateDialog">
              {{ soloView ? 'Crear mi perfil' : `Agregar ${v.professional}` }}
            </BaseButton>
          </template>
        </EmptyScene>

        <!-- Perfil individual (DEC-115): su propia tarjeta, no una fila de
             tabla sin cabecera. StaffPage sigue siendo dueño de los datos y
             del diálogo de edición; la tarjeta solo presenta. -->
        <MyProfileCard
          v-if="soloCard"
          :full-name="barbers[0]!.fullName"
          :photo-url="photoOf(barbers[0]!)"
          :has-photo="hasPhoto(barbers[0]!)"
          :since-label="formatDate(barbers[0]!.createdAt)"
          :role-label="v.Professional"
          @edit="openRenameDialog(barbers[0]!)"
        >
          <template #actions>
            <slot name="barber-actions" :barber="barbers[0]!" />
          </template>
        </MyProfileCard>

        <div v-else class="staff-page__table">
          <div class="staff-page__columns" aria-hidden="true">
            <span>{{ v.Professional }}</span
            ><span>En NAVA desde</span><span>Foto</span><span>Acción</span>
          </div>
          <ul
            ref="listRef"
            class="staff-page__list"
            :aria-label="`${v.Professionals} ${v.ofTheBusiness}`"
          >
            <li
              v-for="(barber, index) in pageLoading ? [] : barbers"
              :key="barber.id"
              class="staff-page__row"
              :class="{
                'staff-page__row--new': barber.id === justAddedId,
              }"
              :style="{ '--row-index': rowIndex(index) }"
            >
              <div class="staff-page__item-heading">
                <BarberAvatar
                  class="staff-page__portrait"
                  size="row"
                  :full-name="barber.fullName"
                  :photo-url="photoOf(barber)"
                />
                <div class="staff-page__item-text">
                  <!-- El nombre es el disparador real del detalle: un <button>
                       nativo para teclado y lector de pantalla; su ::after cubre
                       TODA la fila, así clic o toque en cualquier parte la abren,
                       y la acción (z-index superior) sigue siendo un botón
                       independiente. -->
                  <button
                    type="button"
                    class="staff-page__item-trigger"
                    aria-haspopup="dialog"
                    @click="openDetailDialog(barber)"
                  >
                    <span class="staff-page__item-name" :title="barber.fullName">{{
                      barber.fullName
                    }}</span>
                    <span class="staff-page__item-chevron" aria-hidden="true">›</span>
                  </button>
                  <span class="staff-page__item-since-inline">
                    Desde {{ formatDate(barber.createdAt) }}
                  </span>
                </div>
              </div>
              <span class="staff-page__item-meta staff-page__item-since">
                {{ formatDate(barber.createdAt) }}
              </span>
              <span
                class="staff-page__photo-state"
                :class="{ 'staff-page__photo-state--has': hasPhoto(barber) }"
                :style="diamondStyle(barber)"
              >
                {{ hasPhoto(barber) ? 'Con foto' : 'Sin foto' }}
              </span>
              <div class="staff-page__item-actions">
                <slot name="barber-actions" :barber="barber" />
                <BaseButton
                  type="button"
                  variant="secondary"
                  :aria-label="`Editar ${barber.fullName}`"
                  @click="openRenameDialog(barber)"
                >
                  Editar
                </BaseButton>
              </div>
            </li>
            <!-- Mientras llega la siguiente página, esqueletos con la forma de
                 una fila en el lugar donde van a entrar. -->
            <template v-if="pageLoading">
              <li
                v-for="n in Math.min(pageSize, Math.max(1, barbers.length))"
                :key="`skeleton-${n}`"
                class="staff-page__skeleton-row staff-page__skeleton-row--page"
                aria-hidden="true"
              >
                <span class="staff-page__skeleton-portrait" />
                <span class="staff-page__skeleton-text">
                  <span class="staff-page__skeleton-bar staff-page__skeleton-bar--name" />
                  <span class="staff-page__skeleton-bar staff-page__skeleton-bar--meta" />
                </span>
                <span class="staff-page__skeleton-bar staff-page__skeleton-bar--button" />
              </li>
            </template>
          </ul>
        </div>

        <div v-if="barbers.length > 0 && !soloCard" ref="footerRef" class="staff-page__footer">
          <BaseAlert
            v-if="pageFailed"
            variant="warning"
            title="No pudimos cargar esta página"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo.
          </BaseAlert>
          <div class="staff-page__footer-bar">
            <span class="staff-page__count" role="status">{{ countLabel }}</span>
            <nav
              class="staff-page__pagination-nav"
              :aria-label="`Paginación de ${v.professionals}`"
            >
              <BaseButton
                type="button"
                variant="secondary"
                aria-label="Página anterior"
                :disabled="currentPage === 1 || pageLoading"
                @click="goToPage(currentPage - 1)"
                >‹</BaseButton
              >
              <template v-for="(token, index) in pageTokens" :key="`${token}-${index}`">
                <span
                  v-if="token === 'ellipsis'"
                  class="staff-page__pagination-ellipsis"
                  aria-hidden="true"
                  >…</span
                >
                <BaseButton
                  v-else
                  type="button"
                  :variant="token === currentPage ? 'primary' : 'secondary'"
                  :aria-label="`Página ${token}`"
                  :aria-current="token === currentPage ? 'page' : undefined"
                  :disabled="pageLoading"
                  @click="goToPage(token)"
                  >{{ token }}</BaseButton
                >
              </template>
              <BaseButton
                type="button"
                variant="secondary"
                aria-label="Página siguiente"
                :disabled="currentPage === totalPages || pageLoading"
                @click="goToPage(currentPage + 1)"
                >›</BaseButton
              >
            </nav>
            <BaseButton
              v-if="pageFailed"
              type="button"
              variant="secondary"
              @click="load(lastRequestedPage)"
              >Reintentar página</BaseButton
            >
          </div>
        </div>
      </div>
    </Transition>

    <!-- Detalle: ficha del barbero con su retrato grande. Va ANTES de los demás
         diálogos a propósito: al pulsar "Editar" se cierra este y se abre el de
         edición en el mismo ciclo, y BaseDialog guarda como "foco previo" el
         elemento activo al abrirse; con este orden el foco ya volvió a la fila
         cuando el de edición lo lee. -->
    <BaseDialog
      v-model="isDetailOpen"
      :title="detailTarget?.fullName"
      size="sm"
      content-class="staff-page__dialog staff-page__detail-dialog"
    >
      <div v-if="detailTarget" class="staff-page__ink staff-page__detail">
        <div class="staff-page__detail-hero" :style="{ '--row-index': 0 }">
          <span class="staff-page__detail-frame">
            <BarberAvatar
              size="hero"
              :full-name="detailTarget.fullName"
              :photo-url="photoOf(detailTarget)"
            />
          </span>
          <p class="staff-page__detail-role">{{ v.Professional }}</p>
        </div>

        <dl class="staff-page__facts">
          <div :style="{ '--row-index': 1 }">
            <dt>En NAVA desde</dt>
            <dd>{{ formatDate(detailTarget.createdAt) }}</dd>
          </div>
          <div :style="{ '--row-index': 2 }">
            <dt>Foto</dt>
            <dd>{{ hasPhoto(detailTarget) ? 'Con foto' : 'Sin foto' }}</dd>
          </div>
        </dl>
      </div>
      <!-- Pie fuera del área con scroll: "Cerrar"/"Editar" nunca quedan fuera de
           la pantalla. -->
      <template #footer>
        <div class="staff-page__ink staff-page__detail-footer">
          <BaseButton type="button" variant="secondary" @click="isDetailOpen = false">
            Cerrar
          </BaseButton>
          <BaseButton type="button" variant="primary" @click="onDetailEdit">Editar</BaseButton>
        </div>
      </template>
    </BaseDialog>

    <!-- Alta -->
    <BaseDialog
      v-model="isCreateOpen"
      :title="`Agregar ${v.professional}`"
      description="Aparece en tu equipo en cuanto lo guardes."
      size="md"
      content-class="staff-page__dialog staff-page__create-dialog"
      @close="onCreateDialogClosed"
    >
      <template #icon>
        <span class="staff-page__dialog-chip" aria-hidden="true">+</span>
      </template>
      <form
        class="staff-page__ink staff-page__form staff-page__create-form"
        novalidate
        @submit.prevent="onSubmitCreate"
      >
        <BaseAlert
          v-if="createStatus === 'idempotency-conflict'"
          variant="danger"
          title="No pudimos completar el intento anterior"
          role="alert"
        >
          Inténtalo de nuevo.
        </BaseAlert>
        <BaseAlert
          v-if="createStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="createStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <BarberPhotoField
          v-model:draft="createPhotoDraft"
          :full-name="createFullName || `Nuev${v.professionalEnding} ${v.professional}`"
          :current-url="null"
          :disabled="createStatus === 'saving'"
        />

        <BaseInput
          :model-value="createFullName"
          name="fullName"
          label="Nombre"
          required
          :maxlength="120"
          :disabled="createStatus === 'saving'"
          :error="createFieldError"
          :class="{ 'staff-page__input--filled': !!createFullName }"
          @update:model-value="onCreateFullNameInput"
        />

        <div class="staff-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isCreateOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="createStatus === 'saving'"
            :disabled="createStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>

    <!-- Edición -->
    <BaseDialog
      v-model="isRenameOpen"
      :title="`Editar ${v.professional}`"
      size="md"
      content-class="staff-page__dialog staff-page__edit-dialog"
      @close="onRenameDialogClosed"
    >
      <template #icon>
        <span class="staff-page__dialog-chip" aria-hidden="true">{{
          renameTarget ? firstInitial(renameTarget.fullName) : ''
        }}</span>
      </template>
      <form
        class="staff-page__ink staff-page__form staff-page__rename-form"
        novalidate
        @submit.prevent="onSubmitRename"
      >
        <BaseAlert
          v-if="renameStatus === 'not-found'"
          variant="warning"
          :title="`${capitalize(v.thisProfessional)} ya no está disponible`"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-if="renameStatus === 'photo-error'"
          variant="warning"
          title="No pudimos guardar la foto"
          role="alert"
        >
          Prueba con otra imagen o inténtalo de nuevo. Lo que escribiste sigue aquí.
        </BaseAlert>
        <BaseAlert
          v-if="renameStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="renameStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <BarberPhotoField
          v-model:draft="renamePhotoDraft"
          :full-name="renameFullName || renameTarget?.fullName || ''"
          :current-url="renameTarget ? photoOf(renameTarget) : null"
          :disabled="renameStatus === 'saving'"
        />

        <BaseInput
          :model-value="renameFullName"
          name="fullName"
          label="Nombre"
          required
          :maxlength="120"
          :disabled="renameStatus === 'saving'"
          :error="renameFieldError"
          :class="{ 'staff-page__input--filled': !!renameFullName }"
          @update:model-value="onRenameFullNameInput"
        />

        <div class="staff-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isRenameOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="renameStatus === 'saving'"
            :disabled="renameStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>
  </section>
</template>

<style scoped>
.staff-page__ready--fitting {
  visibility: hidden;
}
.staff-page__pagination-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.staff-page__pagination-ellipsis {
  padding-inline: 4px;
  color: var(--color-on-strong-muted);
}

/* Botones del paginador: mismo par dorado-sólido (página vigente)/tinta-
   fantasma (el resto) que ya usa el CTA "Agregar servicio" y las acciones
   de fila — no los valores por defecto de BaseButton, calibrados para
   flotar sobre superficie clara. Cuadrados y compactos (34px, sin relleno
   horizontal de sobra): un paginador numerado vive de la repetición, no
   necesita el mismo padding que un botón de acción con texto largo. */
.staff-page__pagination-nav :deep(.base-button) {
  height: 38px;
  min-width: 38px;
  padding-inline: 10px;
  font-size: var(--font-size-body-sm);
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.staff-page__pagination-nav :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.staff-page__pagination-nav
  :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.staff-page__pagination-nav :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.staff-page__pagination-nav
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.staff-page__pagination-nav :deep(.base-button:disabled) {
  opacity: 0.4;
}

/* Fundido entre carga/error/listo: --motion-duration-base, el mismo token que
   el resto de transiciones de estado (estandar-diseno-visual.md §12). */
.staff-content-enter-active,
.staff-content-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.staff-content-enter-from,
.staff-content-leave-to {
  opacity: 0;
}

/* Superficie tinta de punta a punta (estandar-diseno-visual.md §3), el mismo
   canvas que Agenda y Servicios; columna de lectura de 820px centrada. */
.staff-page {
  --staff-width: 920px;
  --staff-row-height: 84px;

  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
  padding: 34px 32px 48px;
  color: var(--color-on-strong);
  /* Transparente: la tinta y el fondo animado los pone el cascarón. */
  background: transparent;
  box-sizing: border-box;
}

.staff-page__header,
.staff-page__state,
.staff-page__ready,
.staff-page > :deep(.base-alert) {
  width: min(100%, var(--staff-width));
  margin-inline: auto;
}

.staff-page__ready {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.staff-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: 16px;
  border-bottom: var(--border-width-normal) solid var(--color-field-strong-border);
}

.staff-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
}

@media (min-width: 1024px) {
  .staff-page__title {
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

.staff-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
}

/* CTA "Agregar barbero": el relleno tinta de BaseButton--primary es el MISMO
   color que el fondo de la página, así que se levanta con el dorado de marca
   (mismo criterio que "Agregar servicio"). */
.staff-page__create.base-button {
  height: 40px;
  padding-inline: 18px;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
}

.staff-page__create :deep(.base-button__content)::before {
  content: '+';
  margin-right: 6px;
}

.staff-page__create.base-button--primary {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.staff-page__create.base-button--primary:hover:not(:disabled):not(.base-button--loading) {
  filter: brightness(92%);
}

.staff-page__create.base-button--primary:active:not(:disabled):not(.base-button--loading) {
  filter: brightness(84%);
}

.staff-page__state {
  padding: var(--space-4);
  color: var(--color-on-strong-muted);
}

.staff-page__loading {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: 0;
}

/* Esqueletos: la forma de una fila real (retrato cuadrado, dos líneas de texto,
   un botón) con el pulso de Servicios/Agenda. */
.staff-page__skeleton-row {
  display: flex;
  align-items: center;
  gap: 16px;
  height: var(--staff-row-height);
  padding: 16px;
  overflow: hidden;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.staff-page__skeleton-text {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}

.staff-page__skeleton-portrait,
.staff-page__skeleton-bar {
  display: block;
  flex: 0 0 auto;
  background-color: color-mix(in srgb, var(--color-on-strong) 16%, transparent);
  border-radius: 2px;
  animation: staff-skeleton-pulse 1400ms ease-in-out infinite;
}

.staff-page__skeleton-portrait {
  width: 52px;
  height: 52px;
}

.staff-page__skeleton-bar--name {
  width: 55%;
  max-width: 180px;
  height: 16px;
}

.staff-page__skeleton-bar--meta {
  width: 96px;
  height: 12px;
  background-color: color-mix(in srgb, var(--color-on-strong) 10%, transparent);
}

.staff-page__skeleton-bar--button {
  width: 72px;
  height: 34px;
}

.staff-page__skeleton-row--page {
  animation: staff-row-enter 280ms var(--motion-easing-standard) both;
}

@keyframes staff-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

/* Panel hundido sobre tinta (--color-field-strong): el mismo hueco en la página
   que los campos de Agenda y la tabla de Servicios. */
.staff-page__list {
  display: flex;
  flex-direction: column;
  padding: 0;
  margin: 0;
  list-style: none;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
  overflow: hidden;
}

.staff-page__columns,
.staff-page__row {
  display: grid;
  grid-template-columns: minmax(180px, 1fr) 128px 108px 212px;
  align-items: center;
  gap: 12px;
}

.staff-page__columns {
  min-height: 44px;
  padding: 0 16px;
  font-family: var(--font-sans);
  /* Encabezado reglado: versalitas espaciadas de latón, igual que el de
     Servicios y los rótulos de campo del panel. */
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.staff-page__columns span:last-child {
  text-align: right;
}

.staff-page__row {
  position: relative;
  height: var(--staff-row-height);
  padding: 16px;
  overflow: hidden;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
  /* Entrada escalonada: cada fila arranca --row-index pasos después de la
     anterior (tope de 8). `both` la mantiene oculta durante su retraso. */
  animation: staff-row-enter 320ms var(--motion-easing-standard) both;
  animation-delay: calc(min(var(--row-index, 0), 8) * 45ms);
}

@keyframes staff-row-enter {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Barbero recién agregado: un velo de latón que se apaga, detrás del contenido. */
.staff-page__row--new::before {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: linear-gradient(
    90deg,
    color-mix(in srgb, var(--color-brand-accent-surface) 34%, transparent),
    color-mix(in srgb, var(--color-brand-accent-surface) 0%, transparent) 70%
  );
  animation: staff-row-flash 1800ms var(--motion-easing-standard) 200ms both;
}

@keyframes staff-row-flash {
  from {
    opacity: 1;
  }

  to {
    opacity: 0;
  }
}

.staff-page__item-heading {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 14px;
}

.staff-page__portrait {
  transition: transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1);
}

.staff-page__item-text {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.staff-page__item-name {
  overflow: hidden;
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-lg);
  font-weight: 500;
  color: var(--color-on-strong);
  white-space: nowrap;
  text-overflow: ellipsis;
  transition: color var(--motion-duration-fast) var(--motion-easing-standard);
}

/* Disparador del detalle: botón sin cromo que hereda la tipografía del nombre.
   Su ::after se estira sobre toda la fila (.staff-page__row es position:
   relative), de modo que cualquier clic o toque la abre; la acción sube por
   z-index y conserva su propio clic. */
.staff-page__item-trigger {
  display: flex;
  align-items: center;
  min-width: 0;
  max-width: 100%;
  gap: 6px;
  padding: 0;
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.staff-page__item-trigger::after {
  content: '';
  position: absolute;
  inset: 0;
}

.staff-page__item-trigger:focus-visible {
  outline: none;
}

.staff-page__item-trigger:focus-visible::after {
  box-shadow: inset 0 0 0 2px var(--color-focus);
}

.staff-page__item-chevron {
  flex: 0 0 auto;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-body-lg);
  line-height: 1;
  opacity: 0.55;
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

/* Solo con puntero real: en táctil :hover se queda pegado tras el toque. */
@media (hover: hover) {
  .staff-page__row:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  }

  .staff-page__row:hover .staff-page__item-name {
    color: var(--color-brand-accent-surface);
  }

  .staff-page__row:hover .staff-page__item-chevron {
    opacity: 1;
    transform: translateX(3px);
  }

  .staff-page__row:hover .staff-page__portrait {
    transform: scale(1.06);
  }
}

.staff-page__item-since-inline {
  display: none;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-variant-numeric: tabular-nums;
}

.staff-page__item-meta {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong-muted);
  font-variant-numeric: tabular-nums;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Estado de la foto como rombo con rótulo (mismo lenguaje que Activo/Inactivo
   de Servicios y los estados de la agenda): sin caja, versalitas espaciadas y
   un rombo que late con el ritmo propio de la fila. */
.staff-page__photo-state {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: var(--font-size-caption-line);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
}

.staff-page__photo-state::before {
  content: '';
  flex-shrink: 0;
  width: 5px;
  height: 5px;
  transform: rotate(45deg);
  background-color: currentColor;
  animation: staff-diamond-blink var(--diamond-duration, 7s) ease-in-out var(--diamond-delay, 0s)
    infinite;
}

.staff-page__photo-state--has {
  color: var(--color-success-on-strong);
}

@keyframes staff-diamond-blink {
  0%,
  62%,
  100% {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }

  80% {
    opacity: 0.55;
    transform: rotate(45deg) scale(0.85);
  }
}

.staff-page__item-actions {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

/* "Editar" usa BaseButton--secondary, calibrado para relleno blanco sobre
   superficie clara: sobre tinta se quita el relleno y se sube el latón claro,
   igual que las acciones de fila de Servicios. */
.staff-page__item-actions :deep(.base-button) {
  height: 36px;
  padding-inline: 14px;
  font-size: var(--font-size-body-sm);
}

.staff-page__item-actions :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.staff-page__item-actions
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.staff-page__item-actions
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

/* Pie de la tabla: conteo a la izquierda, "Cargar más" a la derecha. */
.staff-page__footer {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.staff-page__footer-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.staff-page__count {
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong-muted);
}

.staff-page__load-more.base-button {
  height: 34px;
  padding-inline: 14px;
  font-size: var(--font-size-caption);
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.staff-page__load-more.base-button--secondary {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.staff-page__load-more.base-button--secondary:hover:not(:disabled):not(.base-button--loading) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

/* Diálogos (alta, edición, detalle): mismo tinte, filete de latón y campos
   reglados que los de Servicios. Solo aplica dentro de .staff-page__ink (el
   contenido que esta plantilla pone en el diálogo); el fondo, el encabezado y
   el botón de cerrar —que BaseDialog renderiza en <Teleport to="body">— se
   sobrescriben en el bloque sin "scoped" del final. */
.staff-page__dialog-chip {
  position: relative;
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  overflow: hidden;
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-radius: var(--radius-md);
  font-family: var(--font-display);
  font-size: var(--font-size-title-item);
  line-height: 1;
  animation: staff-chip-pop var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1) 60ms both;
}

/* Un solo barrido de destello cuando la ficha termina de asentarse. */
.staff-page__dialog-chip::after {
  content: '';
  position: absolute;
  inset: -40% -60%;
  background: linear-gradient(
    75deg,
    transparent 40%,
    color-mix(in srgb, var(--color-on-strong) 50%, transparent) 50%,
    transparent 60%
  );
  transform: translateX(-100%);
  animation: staff-chip-glint 480ms cubic-bezier(0.5, 0, 0.3, 1) 260ms both;
}

@keyframes staff-chip-pop {
  from {
    opacity: 0;
    transform: scale(0.5) rotate(-20deg);
  }

  to {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}

@keyframes staff-chip-glint {
  from {
    transform: translateX(-100%);
  }

  to {
    transform: translateX(100%);
  }
}

.staff-page__ink {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  color: var(--color-on-strong);
}

.staff-page__ink :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  color: var(--color-on-strong);
}

.staff-page__ink :deep(.base-input__label),
.staff-page__ink :deep(.base-input__required) {
  color: var(--color-brand-accent-surface);
}

.staff-page__ink :deep(.base-input__required) {
  margin-left: 2px;
}

.staff-page__ink :deep(.base-input__hint) {
  color: var(--color-on-strong-muted);
}

.staff-page__ink :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

/* El filete dorado inferior marca el campo YA RESUELTO (mismo criterio que
   Agenda y Servicios). */
.staff-page__ink :deep(.staff-page__input--filled .base-input) {
  background-color: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.staff-page__ink :deep(.base-input:disabled),
.staff-page__ink :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.staff-page__ink :deep(.base-input--invalid) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.staff-page__ink :deep(.base-input__error) {
  color: var(--color-danger-on-strong);
}

/* Botones del pie: par latón-sólido / tinta-fantasma calibrado para este fondo
   (BaseButton está pensado para flotar sobre superficie clara). */
.staff-page__ink :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.staff-page__ink :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.staff-page__ink :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.staff-page__ink :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  filter: brightness(84%);
}

.staff-page__ink :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.staff-page__ink :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.staff-page__ink :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

.staff-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-2);
  flex-wrap: wrap;
}

/* Detalle: retrato grande en un marco de esquinas de latón, el nombre ya está en
   el encabezado del diálogo. Todo entra escalonado con --row-index un instante
   después de la tarjeta, para que la ficha se arme en vez de aparecer de golpe. */
.staff-page__detail-hero,
.staff-page__facts > div {
  animation: staff-detail-enter 380ms var(--motion-easing-standard) both;
  animation-delay: calc(120ms + var(--row-index, 0) * 70ms);
}

.staff-page__detail-hero {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) 0 var(--space-2);
}

.staff-page__detail-frame {
  position: relative;
  display: inline-flex;
  padding: 10px;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 2px;
}

.staff-page__detail-frame::before,
.staff-page__detail-frame::after {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  border: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
}

.staff-page__detail-frame::before {
  top: -5px;
  left: -5px;
  border-right: 0;
  border-bottom: 0;
}

.staff-page__detail-frame::after {
  right: -5px;
  bottom: -5px;
  border-top: 0;
  border-left: 0;
}

.staff-page__detail-role {
  margin: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.staff-page__facts {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: minmax(0, 1fr);
  gap: var(--space-2);
  margin: 0;
}

.staff-page__facts > div {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-3);
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
}

.staff-page__facts dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.staff-page__facts dd {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h3);
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
  overflow-wrap: anywhere;
}

.staff-page__detail-footer {
  flex: 1;
  flex-flow: row wrap;
  justify-content: flex-end;
  gap: var(--space-3);
}

@keyframes staff-detail-enter {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Zona media 641–960px: las columnas fijas ya no caben junto al mínimo de la
   de nombre; se acortan (sin apilar: el alto de fila es fijo). */
@media (max-width: 960px) and (min-width: 641px) {
  .staff-page__columns,
  .staff-page__row {
    grid-template-columns: minmax(120px, 1fr) 100px 96px 196px;
    gap: 8px;
  }

  .staff-page__item-meta {
    font-size: var(--font-size-body-sm);
  }
}

@media (max-width: 640px) {
  .staff-page {
    --staff-row-height: 104px;
    gap: 12px;
    padding: 16px 16px 28px;
  }

  .staff-page__header {
    padding-bottom: 10px;
  }

  /* CTA compacto: un "+" en un círculo (el nombre accesible sigue siendo el
     texto del botón). */
  .staff-page__create.base-button {
    width: 32px;
    height: 32px;
    padding: 0;
    overflow: hidden;
    font-size: 0;
  }

  .staff-page__create :deep(.base-button__content) {
    font-size: 0;
  }

  .staff-page__create :deep(.base-button__content)::before {
    margin: 0;
    font-size: var(--font-size-title-item);
  }

  .staff-page__columns,
  .staff-page__photo-state,
  .staff-page__item-since {
    display: none;
  }

  /* En móvil el alta va bajo el nombre y el estado de la foto se ve en el detalle. */
  .staff-page__item-since-inline {
    display: block;
  }

  .staff-page__row {
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: 44px 30px;
    gap: var(--space-2);
    height: var(--staff-row-height);
    padding: 10px 12px;
  }

  .staff-page__skeleton-row {
    height: var(--staff-row-height);
    gap: 12px;
    padding: 10px 12px;
  }

  .staff-page__skeleton-portrait {
    width: 44px;
    height: 44px;
  }

  .staff-page__item-heading {
    gap: 12px;
  }

  .staff-page__item-name {
    font-size: var(--font-size-body);
  }

  .staff-page__item-actions :deep(.base-button) {
    height: 30px;
    padding-inline: 10px;
  }

  .staff-page__footer-bar {
    justify-content: center;
    text-align: center;
  }

  .staff-page__facts {
    grid-auto-flow: row;
  }

  .staff-page__dialog-actions :deep(.base-button) {
    flex: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .staff-content-enter-active,
  .staff-content-leave-active,
  .staff-page__portrait,
  .staff-page__item-name,
  .staff-page__item-chevron {
    transition: none;
  }

  .staff-page__row,
  .staff-page__row--new::before,
  .staff-page__skeleton-row--page,
  .staff-page__photo-state::before,
  .staff-page__dialog-chip,
  .staff-page__dialog-chip::after,
  .staff-page__detail-hero,
  .staff-page__facts > div {
    animation: none;
  }

  .staff-page__skeleton-portrait,
  .staff-page__skeleton-bar {
    animation: none;
    opacity: 0.8;
  }
}
</style>

<style>
/* SIN "scoped" a propósito, por el mismo motivo que en Servicios: BaseDialog.vue
   renderiza su tarjeta, encabezado, título, descripción y botón de cerrar en
   <Teleport to="body">, así que dejan de ser descendientes de .staff-page en el
   DOM real y un :deep() con alcance de componente nunca los alcanza. Estas
   reglas seleccionan por el nombre de clase real (.staff-page__dialog es
   exclusivo de esta pantalla). El !important es necesario porque la regla
   propia de BaseDialog tiene la misma especificidad y el orden de inserción de
   los <style> no está garantizado. Contenido a esta única pantalla. */
.staff-page__dialog.base-dialog {
  background-color: var(--color-surface-strong) !important;
  border: var(--border-width-normal) solid var(--color-field-strong-border) !important;
}

/* Mismo rebote de apertura que ya usan los diálogos de Servicios y
   BarberSelect: un elemento que llega y se asienta. */
.staff-page__dialog.base-dialog--open {
  transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1) !important;
}

@media (prefers-reduced-motion: reduce) {
  .staff-page__dialog.base-dialog--open {
    transition-timing-function: var(--motion-easing-standard) !important;
  }
}

.staff-page__dialog .base-dialog__header {
  border-bottom-width: var(--border-width-emphasis) !important;
  border-bottom-color: var(--color-brand-accent-surface) !important;
}

.staff-page__dialog .base-dialog__title {
  color: var(--color-on-strong) !important;
  overflow-wrap: anywhere;
}

.staff-page__dialog .base-dialog__description {
  color: var(--color-on-strong-muted) !important;
}

.staff-page__detail-dialog .base-dialog__footer {
  border-top-color: var(--color-field-strong-border) !important;
}

.staff-page__dialog .base-dialog__close {
  color: var(--color-on-strong-muted) !important;
}

.staff-page__dialog .base-dialog__close:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent) !important;
  color: var(--color-on-strong) !important;
}

.staff-page__dialog .base-dialog__close:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus) !important;
}
</style>
