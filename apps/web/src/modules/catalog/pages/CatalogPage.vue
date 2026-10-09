<script setup lang="ts">
// Pantalla "Servicios" (HU-022, HU-024): consultar y listar el catálogo de
// la barbería activa, crear un servicio, editarlo y cambiar su ciclo de
// vida (desactivar/reactivar). Estados discriminados: carga inicial, listo,
// vacío, error recuperable, guardando (mismo patrón que StaffPage.vue). Un
// error recuperable NUNCA borra lo que el barbero ya escribió; solo un
// guardado exitoso confirmado por el servidor cierra el diálogo. Nunca
// muestra asignaciones a barberos (HU-023), disponibilidad ni citas: fuera
// de alcance de esta historia. Cada cambio confirmado por el servidor añade un
// aviso emergente de confirmación (DEC-095); el resultado persistente sigue
// siendo la propia lista, y los errores siguen dentro del diálogo, junto al
// formulario que conservan.
import { computed, onMounted, onUnmounted, ref, useSlots, watch } from 'vue'
import { PAGE_MIN_HOLD_MS, useMinHoldLoading, useToast, useVocabulary } from '@/shared/composables'
import {
  BaseAlert,
  BaseBadge,
  BaseButton,
  BaseDialog,
  BaseInput,
  DiamondLoader,
  EmptyScene,
} from '@/shared/ui'
import {
  createService,
  deactivateService,
  fetchServices,
  previewDeactivation,
  reactivateService,
  updateService,
  type ServiceInput,
} from '../api/catalogApi'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type { Service } from '../model/service'
import {
  validateDescription,
  validateDurationMinutes,
  validateName,
  validatePrice,
} from '../validation/catalogValidation'

const toast = useToast()

// Columna opcional «Lo ofrezco» (DEC-115): `app` la aporta con el slot
// `service-offer` solo en el perfil de barbero individual. Sin el slot la tabla
// conserva exactamente su geometría de siempre. Este módulo no sabe qué hay
// dentro ni importa `barberServices`: solo le reserva el hueco.
// Se lee en cada render y NO en un `computed`: los slots no son reactivos, y el
// interruptor puede llegar después de montar (cuando termina de cargar qué ofrece).
const slots = useSlots()
const hasOfferColumn = () => !!slots['service-offer']

// Un servicio recién creado se anuncia hacia afuera para que `app` pueda,
// por ejemplo, ofrecérselo al único barbero sin que este módulo lo sepa.
const emit = defineEmits<{ 'service-created': [service: Service] }>()
// Palabras de la barbería (DEC-110): con los valores iniciales, el texto de siempre.
const v = useVocabulary()

type LoadStatus = 'loading' | 'ready' | 'load-error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'name-conflict'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

interface FormFieldErrors {
  name?: string
  description?: string
  durationMinutes?: string
  price?: string
}

function hasFieldError(errors: FormFieldErrors): boolean {
  return !!(errors.name || errors.description || errors.durationMinutes || errors.price)
}

const loadStatus = ref<LoadStatus>('loading')
const services = ref<Service[]>([])

// effectivePageSize sustituye a un PAGE_SIZE fijo (issue reportado
// 2026-09-28, primero "colócalos en una tabla con paginador y con su
// buscador... así con scroll infinito está feo", luego "que se adapte
// según la pantalla, que no haya scroll"): cuántas filas se piden por
// página ya no es un número fijo, se calcula para llenar exactamente el
// espacio vertical disponible dentro de .private-shell__content (el único
// contenedor con scroll propio del cascarón, PrivateShell.vue) — ni una
// fila de más (que forzaría scroll) ni de menos (que dejaría hueco vacío
// evitable). DEFAULT_PAGE_SIZE solo cubre la primera carga, antes de poder
// medir la pantalla real (ver computeFittingPageSize/adjustPageSizeToFit
// más abajo); coincide con el valor por defecto del backend
// (catalog.DefaultPageSize) para que ese primer pedido nunca dispare un
// clamping silencioso.
const DEFAULT_PAGE_SIZE = 20
const effectivePageSize = ref(DEFAULT_PAGE_SIZE)

// DYNAMIC_MIN_PAGE_SIZE = 1 (issue reportado 2026-09-28, "que se adapte
// según la pantalla, que no haya scroll"): "que no haya scroll" pesa más
// que "mostrar muchas filas" — en una ventana muy baja, es preferible una
// tabla de una sola fila por página a que reaparezca scroll. MAX_PAGE_SIZE
// coincide con el techo real del backend (catalog.MaxPageSize, DEC-103):
// pedir más no cambiaría nada.
const DYNAMIC_MIN_PAGE_SIZE = 1
const MAX_PAGE_SIZE = 50

// ROW_HEIGHT_*_PX debe coincidir EXACTAMENTE con el alto FIJO (no
// min-height) que .catalog-page__row impone por CSS más abajo: con un alto
// variable no habría forma de predecir cuántas filas caben en el espacio
// disponible, y volvería a aparecer scroll.
const ROW_HEIGHT_DESKTOP_PX = 80
const ROW_HEIGHT_MOBILE_PX = 100
const MOBILE_BREAKPOINT_QUERY = '(max-width: 640px)'

// typeof matchMedia === 'function': jsdom (CatalogPage.test.ts) no la
// implementa (issue detectado al verificar en vivo: un rechazo de promesa
// sin capturar por cada prueba, "window.matchMedia is not a function").
// Ningún navegador real carece de esta API; el resguardo es solo para ese
// entorno de pruebas, con el mismo resultado que ya asumía antes de que
// existiera este cálculo (ancho de escritorio).
function currentRowHeightPx(): number {
  if (typeof window.matchMedia !== 'function') return ROW_HEIGHT_DESKTOP_PX
  return window.matchMedia(MOBILE_BREAKPOINT_QUERY).matches
    ? ROW_HEIGHT_MOBILE_PX
    : ROW_HEIGHT_DESKTOP_PX
}

// tableListRef/paginationFooterRef anclan la fila real de datos y el pie
// de paginador (ver plantilla): computeFittingPageSize mide sus posiciones
// reales en pantalla, nunca asume un valor fijo para nada de lo que las
// rodea (encabezado, buscador, encabezado de columnas, hueco entre tabla y
// paginador), que puede cambiar de alto sin que esta función se entere.
const tableListRef = ref<HTMLElement | null>(null)
const paginationFooterRef = ref<HTMLElement | null>(null)

// findScrollableAncestor sube por el árbol real del DOM hasta el primer
// ancestro que el CSS marca como scrolleable (overflow-y auto/scroll) —
// nunca se acopla al nombre de clase de PrivateShell.vue (otro componente,
// otro módulo): a esta pantalla solo le importa "¿cuál es mi visor con
// scroll propio?", no cómo se llama. Se basa en la propiedad CSS
// declarada, no en si HAY overflow ahora mismo: una vez que el ajuste
// funcione y ya no sobre contenido, `scrollHeight > clientHeight` dejaría
// de ser cierto y ese criterio rompería los cálculos futuros (un resize
// que agranda la ventana, por ejemplo).
function findScrollableAncestor(el: HTMLElement): HTMLElement | null {
  let node = el.parentElement
  while (node) {
    const overflowY = window.getComputedStyle(node).overflowY
    if (overflowY === 'auto' || overflowY === 'scroll') return node
    node = node.parentElement
  }
  return null
}

// computeFittingPageSize calcula cuántas filas caben SIN que
// .private-shell__content necesite scroll. Fuerza scrollTop a 0 antes de
// medir: la posición real de tableListRef en el viewport depende de cuánto
// esté desplazado su contenedor, y esta pantalla quiere calcular para la
// posición de reposo (sin desplazar), que es además el estado al que
// aspira a volver.
// PAGE_BOTTOM_PADDING_PX debe coincidir con el padding-bottom real de
// .catalog-page (ver <style scoped>, "padding: 34px 32px 48px"). NO se
// puede medir en vivo como gapBelowList (abajo): .catalog-page tiene
// min-height: 100%, así que con pocas filas su borde inferior se estira
// hasta el final de .private-shell__content sin importar cuánto padding
// real tenga — medir "borde de .catalog-page menos borde del paginador" en
// ese caso da todo el hueco vacío estirado, no el padding, y hace que el
// cálculo se autodestruya (encontrado en vivo: en una pantalla baja, un
// hueco enorme se confundía con "padding", y el resultado subestimaba
// muchísimo cuántas filas caben).
const PAGE_BOTTOM_PADDING_PX = 48

function computeFittingPageSize(): number | null {
  const list = tableListRef.value
  const pagination = paginationFooterRef.value
  if (!list || !pagination) return null

  // Sin un ancestro con scroll propio no hay un viewport real contra el
  // cual ajustar nada (mismo caso, deliberado, de CatalogPage.test.ts: ahí
  // este componente se monta SOLO, sin el PrivateShell real que en la app
  // siempre lo envuelve) — caer a window.innerHeight como aproximación
  // sería adivinar un límite que no existe, no medir uno real.
  const scrollParent = findScrollableAncestor(list)
  if (!scrollParent) return null
  scrollParent.scrollTop = 0

  const viewportHeight = scrollParent.clientHeight
  const containerTop = scrollParent.getBoundingClientRect().top

  const listRect = list.getBoundingClientRect()
  const paginationRect = pagination.getBoundingClientRect()

  // gapBelowList (el hueco real entre el final de la lista y el inicio del
  // paginador) SÍ se mide en vivo, sin problema: es un gap entre dos
  // hermanos flex (.catalog-page__ready), no depende del min-height:100%
  // de un ancestro.
  const gapBelowList = paginationRect.top - listRect.bottom

  const available =
    viewportHeight -
    (listRect.top - containerTop) -
    gapBelowList -
    paginationRect.height -
    PAGE_BOTTOM_PADDING_PX
  const rows = Math.floor(available / currentRowHeightPx())

  return Math.min(Math.max(rows, DYNAMIC_MIN_PAGE_SIZE), MAX_PAGE_SIZE)
}

const currentPage = ref(1)
const totalItems = ref(0)
const totalPages = ref(1)
// searchTerm es lo que el barbero está escribiendo AHORA MISMO (v-model del
// campo); activeSearch es el término ya aplicado a la última carga, tras el
// debounce. Ambos difieren mientras la persona sigue escribiendo — solo
// activeSearch decide qué copy de "vacío" mostrar y qué parámetro viaja al
// backend.
const searchTerm = ref('')
const activeSearch = ref('')
// pageStatus cubre la carga de una página YA EN PANTALLA (cambiar de
// página o escribir en el buscador): a diferencia de loadStatus (carga
// inicial, con esqueleto de página completa), esto NUNCA reemplaza la tabla
// visible — solo la atenúa (ver :class en la plantilla) mientras llega la
// respuesta, para que teclear en el buscador no "parpadee" toda la pantalla
// en cada pulsación. Lo que cambia es el CUERPO de la tabla: sus filas se
// cambian por esqueletos (skeletonRowCount) hasta que llega el resultado.
const pageStatus = ref<'idle' | 'loading' | 'error'>('idle')

const isSearching = computed(() => activeSearch.value !== '')

// Cuántos esqueletos ocupan el lugar de las filas mientras carga una
// página/búsqueda: las mismas que había en pantalla, para que la tabla no
// cambie de alto (y el paginador no salte) al llegar el resultado; 3 si
// venía de una búsqueda sin coincidencias, que no tenía filas que igualar.
const skeletonRowCount = computed(() => Math.max(services.value.length, 3))

// resultsLabel acompaña al paginador con el dato real que un paginador
// numerado necesita para tener sentido (issue 2026-09-28, "con su
// buscador"): confirma cuántos coincidieron, no solo cuántas páginas hay.
const resultsLabel = computed(() => {
  const n = totalItems.value
  if (isSearching.value) {
    return `${n} ${n === 1 ? 'resultado' : 'resultados'} para «${activeSearch.value}»`
  }
  return `${n} ${n === 1 ? 'servicio' : 'servicios'}`
})

type PageToken = number | 'ellipsis'

// paginationWindow: con pocas páginas muestra todas (1..total); con muchas,
// ancla primera y última página y una vecindad de 1 alrededor de la
// vigente, con "…" en los huecos — el patrón numerado estándar de un
// paginador de tabla, sin que 30+ páginas desborden la fila de botones.
function paginationWindow(current: number, total: number): PageToken[] {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const tokens: PageToken[] = [1]
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  if (start > 2) tokens.push('ellipsis')
  for (let p = start; p <= end; p++) tokens.push(p)
  if (end < total - 1) tokens.push('ellipsis')
  tokens.push(total)
  return tokens
}

const pageTokens = computed(() => paginationWindow(currentPage.value, totalPages.value))

// useMinHoldLoading (issue reportado 2026-09-28: primero "espasmo cuando
// se cambia de pantallas", luego "se ve como se genera el objeto... quiero
// el tema de la carga para tapar ese evento"): el rombo/esqueleto se
// muestra DE INMEDIATO (nunca una pantalla en blanco que deje ver el
// contenido armándose), pero no se le permite ceder el paso al contenido
// real hasta que pasen al menos 400ms desde que empezó — así una carga
// rápida no lo retira a medio parpadeo. Solo protege la carga INICIAL
// (loadStatus); cambiar de página o buscar usa pageStatus, sin esqueleto.
const { start: startLoadingHold, hold: holdLoadingReveal } = useMinHoldLoading()

// PAGE_MIN_HOLD_MS (issue reportado 2026-09-29, "que al buscar un servicio
// se haga una animación y no que muestre el resultado, ya que se ve feo"):
// buscar o cambiar de página ya no intercambia las filas de golpe. Mientras
// llega la respuesta las filas se sustituyen por esqueletos con pulso, y el
// resultado entra escalonado. Este segundo hold evita que una respuesta
// inmediata deje el esqueleto en pantalla apenas un fotograma (un parpadeo
// peor que el corte original); es más corto que el de la carga inicial
// porque la persona ya está dentro de la pantalla y espera algo ágil.
const { start: startPageHold, hold: holdPageReveal } = useMinHoldLoading(PAGE_MIN_HOLD_MS)

// requestToken evita que una carga obsoleta (una página o búsqueda vieja
// que sigue en vuelo) sobreescriba la lista con datos que ya no corresponden
// a la página/búsqueda vigente (mismo patrón que StaffPage.vue, ahora
// también cubre cambiar de página o escribir rápido en el buscador).
let requestToken = 0
// lastRequestedPage es la página que "Reintentar" repite: la inicial
// (loadStatus) siempre es 1, pero un fallo de pageStatus debe reintentar la
// MISMA página que falló, no volver a la primera.
let lastRequestedPage = 1

async function load(page: number) {
  const token = ++requestToken
  lastRequestedPage = page
  const isInitialLoad = loadStatus.value !== 'ready'

  if (isInitialLoad) {
    loadStatus.value = 'loading'
    startLoadingHold()
  } else {
    pageStatus.value = 'loading'
    startPageHold()
  }

  const outcome = await fetchServices({
    page,
    pageSize: effectivePageSize.value,
    search: activeSearch.value || undefined,
  })
  if (token !== requestToken) return

  const apply = () => {
    if (outcome.kind === 'success') {
      services.value = outcome.page.items
      currentPage.value = outcome.page.page
      // effectivePageSize refleja lo que el SERVIDOR confirma haber usado
      // (ya clamped ahí también, DEC-103), no solo lo que este cliente
      // pidió: cualquier cálculo posterior de "cuántas filas caben" compara
      // contra la verdad ya confirmada.
      effectivePageSize.value = outcome.page.pageSize
      totalItems.value = outcome.page.total
      totalPages.value = outcome.page.totalPages
      loadStatus.value = 'ready'
      pageStatus.value = 'idle'
      // Sin filas no hay lista que medir (tableListRef nunca se monta): se
      // muestra el aviso de vacío directamente.
      if (outcome.page.items.length === 0) fitPending.value = false
      return
    }
    if (isInitialLoad) {
      loadStatus.value = 'load-error'
    } else {
      pageStatus.value = 'error'
    }
  }

  if (isInitialLoad) {
    holdLoadingReveal(apply)
  } else {
    holdPageReveal(apply)
  }
}

// adjustPageSizeToFit mide la pantalla real DESPUÉS de que la tabla ya
// está en el DOM (issue 2026-09-28, "que se adapte según la pantalla, que
// no haya scroll") y, si el resultado difiere de effectivePageSize, vuelve
// a pedir la página que contiene el mismo primer elemento que se veía antes
// del cambio — no simplemente "vuelve a la página 1" en cada resize, que
// sería desorientador si la persona ya había avanzado varias páginas.
async function adjustPageSizeToFit() {
  if (pageStatus.value === 'loading' || loadStatus.value !== 'ready') return
  const fitting = computeFittingPageSize()
  if (fitting === null || fitting === effectivePageSize.value) return

  const firstItemOffset = (currentPage.value - 1) * effectivePageSize.value
  effectivePageSize.value = fitting
  await load(Math.floor(firstItemOffset / fitting) + 1)
}

let resizeDebounceTimer: ReturnType<typeof setTimeout> | undefined

function onWindowResize() {
  if (resizeDebounceTimer !== undefined) clearTimeout(resizeDebounceTimer)
  resizeDebounceTimer = setTimeout(() => void adjustPageSizeToFit(), 150)
}

// El primer ajuste NO puede encadenarse tras `await load(1)` en onMounted
// ni reaccionar directamente a loadStatus (dos intentos descartados al
// verificar en vivo, issue 2026-09-28):
//   1. holdLoadingReveal (useMinHoldLoading) programa `apply` -el callback
//      que de verdad pone loadStatus en 'ready' y monta la tabla real- con
//      un setTimeout PROPIO que `load()` nunca espera: `await load(1)`
//      podía resolver mientras el DOM seguía mostrando el esqueleto.
//   2. Incluso reaccionando a loadStatus con flush: 'post', el <Transition
//      mode="out-in"> de la plantilla reproduce primero la salida de la
//      rama anterior antes de montar la nueva: loadStatus ya vale 'ready'
//      un instante antes de que el <ul> real llegue a existir en el DOM.
// Reaccionar directamente a tableListRef (el propio ref de plantilla) es lo
// único que refleja el momento real en que Vue insertó ese nodo: por eso
// se dispara ahí, no antes. hasPerformedInitialFit evita repetir el ajuste
// cada vez que una búsqueda hace que la lista se desmonte y remonte (0
// resultados y de vuelta) — el resize sigue siendo el único disparador
// después del primer ajuste real.
let hasPerformedInitialFit = false

// fitPending mantiene la tabla real invisible (opacity, sin quitarla del
// layout: hace falta medirla) hasta que el primer ajuste termina (issue
// reportado 2026-09-30, "al cargar la tabla se expande todo feo y después
// vuelve a su estado natural antes de mostrar los servicios"): la primera
// carga pide DEFAULT_PAGE_SIZE filas, que casi nunca caben, y mostrarlas
// para luego re-pedir la página con el tamaño medido dejaba ver la tabla
// creciendo, esqueletos de la misma altura y por fin el encogimiento.
const fitPending = ref(true)

// fitInitialPage resuelve ese primer ajuste SIN volver al servidor cuando
// puede: la página 1 con `fitting` filas son justo las primeras `fitting` de
// la página 1 ya descargada, así que basta recortarlas y recalcular
// totalPages. Solo si hacen falta más filas de las descargadas (pantalla muy
// alta y aún hay servicios sin traer) se cae al ajuste normal con re-pedido.
async function fitInitialPage() {
  const fitting = computeFittingPageSize()
  if (fitting === null || fitting === effectivePageSize.value) return

  if (fitting <= services.value.length || services.value.length >= totalItems.value) {
    services.value = services.value.slice(0, fitting)
    effectivePageSize.value = fitting
    totalPages.value = Math.max(1, Math.ceil(totalItems.value / fitting))
    return
  }
  await adjustPageSizeToFit()
}

watch(tableListRef, async (el) => {
  if (!el || hasPerformedInitialFit) return
  hasPerformedInitialFit = true
  await fitInitialPage()
  fitPending.value = false
})

onMounted(() => {
  void load(1)
  window.addEventListener('resize', onWindowResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', onWindowResize)
  if (resizeDebounceTimer !== undefined) clearTimeout(resizeDebounceTimer)
})

function onRetryLoad() {
  void load(lastRequestedPage)
}

function goToPage(page: number) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) return
  if (pageStatus.value === 'loading') return
  void load(page)
}

function onPrevPage() {
  goToPage(currentPage.value - 1)
}

function onNextPage() {
  goToPage(currentPage.value + 1)
}

// SEARCH_DEBOUNCE_MS: teclear no dispara una solicitud por pulsación — solo
// cuando la persona deja de escribir este tiempo (patrón estándar de
// buscador en vivo, ningún botón "Buscar" separado, igual que el mockup
// aprobado). requestToken (arriba) descarta la respuesta de un término ya
// obsoleto si dos búsquedas quedan en vuelo a la vez.
const SEARCH_DEBOUNCE_MS = 300
let searchDebounceTimer: ReturnType<typeof setTimeout> | undefined

function onSearchInput(value: string | number) {
  searchTerm.value = String(value)
  if (searchDebounceTimer !== undefined) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    const trimmed = searchTerm.value.trim()
    if (trimmed === activeSearch.value) return
    activeSearch.value = trimmed
    void load(1)
  }, SEARCH_DEBOUNCE_MS)
}

onUnmounted(() => {
  if (searchDebounceTimer !== undefined) clearTimeout(searchDebounceTimer)
})

// No concatena "$" (estandar-diseno-visual.md §10.3: "Dinero se formatea
// según la moneda definida por el contrato; no se concatena '$' ni se
// asumen pesos desde el componente"). service.price es un string decimal
// exacto que nunca se convierte a number en el cliente (docs/06-api/
// estandar-openapi.md §6); esta función no reformatea sus dígitos, solo
// añade la moneda real recibida del contrato.
function formatPrice(service: Service): string {
  return `${service.price} ${service.currency}`
}

// Monograma de la fila: primera letra del nombre, igual criterio que
// `initials()` en StaffPage.vue, pero de una sola letra (un servicio no
// tiene "nombre y apellido" que abreviar en dos). `name` es obligatorio y no
// vacío (validateName), así que siempre hay una letra real que mostrar.
function serviceInitial(name: string): string {
  return name.trim().charAt(0).toUpperCase()
}

// Ritmo propio del rombo de estado, igual técnica que
// DailyAgendaPage.vue#diamondStyle (issue #189): hash estable del id para
// que la duración/desfase del parpadeo no cambien al volver a pintar la
// lista y para que las filas no laten sincronizadas. Aquí no existe el
// concepto de "en proceso" de la agenda (un servicio solo está activo o
// inactivo), así que se usa un único rango, el mismo [6, 5] que Agenda
// aplica a sus turnos que no están en curso.
function diamondStyle(service: Service): Record<string, string> {
  let hash = 0
  for (const char of service.id) hash = (hash * 31 + char.charCodeAt(0)) >>> 0
  const spread = (hash % 1000) / 1000
  const phase = ((hash >>> 10) % 1000) / 1000
  const duration = 6 + spread * 5
  return {
    '--diamond-duration': `${duration.toFixed(1)}s`,
    '--diamond-delay': `-${(phase * duration).toFixed(1)}s`,
  }
}

// --- Detalle --------------------------------------------------------------

// detailTarget no se borra al cerrar: el diálogo sigue visible mientras dura
// su animación de salida y vaciarlo aquí lo dejaría en blanco a media salida.
const isDetailOpen = ref(false)
const detailTarget = ref<Service | null>(null)

function openDetailDialog(service: Service) {
  detailTarget.value = service
  isDetailOpen.value = true
}

function onDetailEdit() {
  const service = detailTarget.value
  if (!service) return
  isDetailOpen.value = false
  openEditDialog(service)
}

// --- Alta -----------------------------------------------------------------

const isCreateOpen = ref(false)
const createForm = ref<ServiceInput>({ name: '', description: '', durationMinutes: 30, price: '' })
const createDurationRaw = ref('30')
const createFieldErrors = ref<FormFieldErrors>({})
const createStatus = ref<SaveStatus>('idle')
const createAttempted = ref(false)
// Clave de idempotencia del intento lógico vigente (RN-IDE-01): se genera
// al abrir el diálogo y se REUTILIZA en cada reintento del mismo intento;
// solo un envío exitoso o cerrar y reabrir el diálogo la renueva (mismo
// intento lógico != mismo clic).
let createIdempotencyKey = newIdempotencyKey()

function openCreateDialog() {
  createForm.value = { name: '', description: '', durationMinutes: 30, price: '' }
  createDurationRaw.value = '30'
  createFieldErrors.value = {}
  createStatus.value = 'idle'
  createAttempted.value = false
  createIdempotencyKey = newIdempotencyKey()
  isCreateOpen.value = true
}

function onCreateDialogClosed() {
  createStatus.value = 'idle'
}

function validateCreateFields(): FormFieldErrors {
  return {
    name: validateName(createForm.value.name),
    description: validateDescription(createForm.value.description),
    durationMinutes: validateDurationMinutes(createDurationRaw.value),
    price: validatePrice(createForm.value.price),
  }
}

function revalidateCreateIfAttempted() {
  if (createAttempted.value) createFieldErrors.value = validateCreateFields()
}

function onCreateNameInput(value: string | number) {
  createForm.value.name = String(value)
  revalidateCreateIfAttempted()
}

function onCreateDescriptionInput(value: string | number) {
  createForm.value.description = String(value)
  revalidateCreateIfAttempted()
}

// BaseInput type="number" ya descarta todo lo que no sea dígito (sin signo
// ni letras); validateDurationMinutes sigue validando el rango al enviar.
function onCreateDurationInput(value: string | number) {
  createDurationRaw.value = String(value)
  revalidateCreateIfAttempted()
}

function onCreatePriceInput(value: string | number) {
  createForm.value.price = String(value)
  revalidateCreateIfAttempted()
}

async function onSubmitCreate() {
  if (createStatus.value === 'saving') return

  createAttempted.value = true
  const errors = validateCreateFields()
  createFieldErrors.value = errors
  if (hasFieldError(errors)) return

  createStatus.value = 'saving'
  const outcome = await createService(
    {
      name: createForm.value.name.trim(),
      description: createForm.value.description,
      durationMinutes: Number(createDurationRaw.value.trim()),
      price: createForm.value.price.trim(),
    },
    createIdempotencyKey,
  )

  switch (outcome.kind) {
    case 'success':
      services.value.unshift(outcome.service)
      isCreateOpen.value = false
      createStatus.value = 'idle'
      emit('service-created', outcome.service)
      toast.success('Servicio creado', {
        detail: `«${outcome.service.name}» ya aparece en tu catálogo.`,
      })
      return
    case 'validation-error':
      createStatus.value = 'validation-error'
      return
    case 'name-conflict':
      createStatus.value = 'name-conflict'
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

// --- Edición ----------------------------------------------------------

const isEditOpen = ref(false)
const editTarget = ref<Service | null>(null)
const editForm = ref<ServiceInput>({ name: '', description: '', durationMinutes: 30, price: '' })
const editDurationRaw = ref('30')
const editFieldErrors = ref<FormFieldErrors>({})
const editStatus = ref<SaveStatus>('idle')
const editAttempted = ref(false)

function openEditDialog(service: Service) {
  editTarget.value = service
  editForm.value = {
    name: service.name,
    description: service.description ?? '',
    durationMinutes: service.durationMinutes,
    price: service.price,
  }
  editDurationRaw.value = String(service.durationMinutes)
  editFieldErrors.value = {}
  editStatus.value = 'idle'
  editAttempted.value = false
  isEditOpen.value = true
}

function onEditDialogClosed() {
  editStatus.value = 'idle'
}

function validateEditFields(): FormFieldErrors {
  return {
    name: validateName(editForm.value.name),
    description: validateDescription(editForm.value.description),
    durationMinutes: validateDurationMinutes(editDurationRaw.value),
    price: validatePrice(editForm.value.price),
  }
}

function revalidateEditIfAttempted() {
  if (editAttempted.value) editFieldErrors.value = validateEditFields()
}

function onEditNameInput(value: string | number) {
  editForm.value.name = String(value)
  revalidateEditIfAttempted()
}

function onEditDescriptionInput(value: string | number) {
  editForm.value.description = String(value)
  revalidateEditIfAttempted()
}

function onEditDurationInput(value: string | number) {
  editDurationRaw.value = String(value)
  revalidateEditIfAttempted()
}

function onEditPriceInput(value: string | number) {
  editForm.value.price = String(value)
  revalidateEditIfAttempted()
}

async function onSubmitEdit() {
  if (editStatus.value === 'saving' || !editTarget.value) return

  editAttempted.value = true
  const errors = validateEditFields()
  editFieldErrors.value = errors
  if (hasFieldError(errors)) return

  editStatus.value = 'saving'
  const outcome = await updateService(editTarget.value.id, {
    name: editForm.value.name.trim(),
    description: editForm.value.description,
    durationMinutes: Number(editDurationRaw.value.trim()),
    price: editForm.value.price.trim(),
  })

  switch (outcome.kind) {
    case 'success': {
      // Reemplaza por id sin duplicar ni reordenar de forma inestable:
      // solo cambian los campos editados y updatedAt del elemento
      // existente, en su misma posición.
      const index = services.value.findIndex((s) => s.id === outcome.service.id)
      if (index !== -1) services.value[index] = outcome.service
      isEditOpen.value = false
      editStatus.value = 'idle'
      toast.success('Servicio actualizado', {
        detail: `Guardamos los cambios de «${outcome.service.name}».`,
      })
      return
    }
    case 'validation-error':
      editStatus.value = 'validation-error'
      return
    case 'name-conflict':
      editStatus.value = 'name-conflict'
      return
    case 'not-found':
      editStatus.value = 'not-found'
      return
    case 'network-error':
      editStatus.value = 'network-error'
      return
    case 'unexpected-error':
      editStatus.value = 'unexpected-error'
  }
}

// --- HU-024: desactivar/reactivar ------------------------------------------

type ImpactStatus = 'loading' | 'ready' | 'error'
type LifecycleStatus =
  | 'idle'
  | 'saving'
  | 'transition-conflict'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

const lifecycleTarget = ref<Service | null>(null)
const lifecycleIntent = ref<'deactivate' | 'reactivate' | null>(null)
const isLifecycleOpen = ref(false)
const lifecycleStatus = ref<LifecycleStatus>('idle')
const impactStatus = ref<ImpactStatus>('loading')
const affectedAppointments = ref(0)
// Una clave por intento lógico (RN-IDE-01), igual disciplina que
// createIdempotencyKey: se genera al abrir el diálogo y se REUTILIZA en
// cada reintento del mismo intento (network-error/unexpected-error);
// confirmar con éxito o cerrar el diálogo exige abrirlo de nuevo para
// obtener una clave nueva.
let lifecycleIdempotencyKey = newIdempotencyKey()

async function openDeactivateDialog(service: Service) {
  lifecycleTarget.value = service
  lifecycleIntent.value = 'deactivate'
  lifecycleStatus.value = 'idle'
  lifecycleIdempotencyKey = newIdempotencyKey()
  isLifecycleOpen.value = true
  await loadDeactivationImpact(service)
}

// CA-024-01: la advertencia solo muestra impacto respaldado por el
// servidor, nunca un valor por defecto asumido en el cliente. Se reutiliza al
// volver a "desactivar" dentro del mismo diálogo tras una reactivación.
async function loadDeactivationImpact(service: Service) {
  impactStatus.value = 'loading'
  affectedAppointments.value = 0

  const outcome = await previewDeactivation(service.id)
  if (lifecycleIntent.value !== 'deactivate' || lifecycleTarget.value?.id !== service.id) return

  if (outcome.kind === 'success') {
    affectedAppointments.value = outcome.affectedAppointments
    impactStatus.value = 'ready'
    return
  }
  if (outcome.kind === 'not-found') {
    lifecycleStatus.value = 'not-found'
    impactStatus.value = 'error'
    return
  }
  impactStatus.value = 'error'
}

// Un cambio confirmado NO cierra el diálogo (issue 2026-09-29, "si cambio de
// estado no significa que cierre la ventana... hasta que le dé a salir"): el
// diálogo pasa a mostrar el estado nuevo con el interruptor ya en su sitio,
// listo para revertirlo. Cada cambio es un intento lógico distinto, así que
// lleva su propia clave de idempotencia (RN-IDE-01).
function settleLifecycleAfterChange(updated: Service) {
  lifecycleTarget.value = updated
  lifecycleStatus.value = 'idle'
  lifecycleIdempotencyKey = newIdempotencyKey()
  if (updated.isActive) {
    lifecycleIntent.value = 'deactivate'
    void loadDeactivationImpact(updated)
  } else {
    lifecycleIntent.value = 'reactivate'
    impactStatus.value = 'ready'
  }
}

function openLifecycleDialog(service: Service) {
  if (service.isActive) void openDeactivateDialog(service)
  else openReactivateDialog(service)
}

function openReactivateDialog(service: Service) {
  lifecycleTarget.value = service
  lifecycleIntent.value = 'reactivate'
  lifecycleStatus.value = 'idle'
  impactStatus.value = 'ready'
  lifecycleIdempotencyKey = newIdempotencyKey()
  isLifecycleOpen.value = true
}

// Fechas del detalle y de "Inactivo desde…" del diálogo de reactivación. Zona
// fija de la moneda del contrato (COP, DEC-067) para que la fecha no dependa
// del huso del equipo.
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

const deactivatedSince = computed(() => formatDate(lifecycleTarget.value?.deactivatedAt))

// Posición del interruptor: refleja el estado actual y salta al destino en
// cuanto se pide el cambio (durante "saving"); si el servidor lo rechaza
// vuelve solo, porque lifecycleStatus deja de ser "saving".
const lifecycleSwitchOn = computed(() => {
  const activeNow = lifecycleIntent.value === 'deactivate'
  return lifecycleStatus.value === 'saving' ? !activeNow : activeNow
})

function onLifecycleDialogClosed() {
  lifecycleStatus.value = 'idle'
  lifecycleIntent.value = null
  lifecycleTarget.value = null
}

function replaceServiceInList(service: Service) {
  const index = services.value.findIndex((s) => s.id === service.id)
  if (index !== -1) services.value[index] = service
}

async function onConfirmDeactivate() {
  const target = lifecycleTarget.value
  if (!target || lifecycleStatus.value === 'saving') return

  lifecycleStatus.value = 'saving'
  const outcome = await deactivateService(target.id, lifecycleIdempotencyKey)

  switch (outcome.kind) {
    case 'success':
      replaceServiceInList(outcome.service)
      settleLifecycleAfterChange(outcome.service)
      toast.success('Servicio desactivado', { detail: `«${outcome.service.name}» quedó inactivo.` })
      return
    case 'not-found':
      lifecycleStatus.value = 'not-found'
      return
    case 'transition-conflict':
      // CA-024-04/Frontend §3: un conflicto de estado recarga el impacto
      // real, nunca asume éxito optimista.
      lifecycleStatus.value = 'transition-conflict'
      void reloadAfterConflict(target.id)
      return
    case 'idempotency-conflict':
      lifecycleStatus.value = 'idempotency-conflict'
      return
    case 'network-error':
      lifecycleStatus.value = 'network-error'
      return
    case 'unexpected-error':
      lifecycleStatus.value = 'unexpected-error'
  }
}

async function onConfirmReactivate() {
  const target = lifecycleTarget.value
  if (!target || lifecycleStatus.value === 'saving') return

  lifecycleStatus.value = 'saving'
  const outcome = await reactivateService(target.id, lifecycleIdempotencyKey)

  switch (outcome.kind) {
    case 'success':
      replaceServiceInList(outcome.service)
      settleLifecycleAfterChange(outcome.service)
      toast.success('Servicio reactivado', {
        detail: `«${outcome.service.name}» volvió a estar activo.`,
      })
      return
    case 'not-found':
      lifecycleStatus.value = 'not-found'
      return
    case 'transition-conflict':
      lifecycleStatus.value = 'transition-conflict'
      void reloadAfterConflict(target.id)
      return
    case 'idempotency-conflict':
      lifecycleStatus.value = 'idempotency-conflict'
      return
    case 'network-error':
      lifecycleStatus.value = 'network-error'
      return
    case 'unexpected-error':
      lifecycleStatus.value = 'unexpected-error'
  }
}

function onConfirmLifecycle() {
  if (lifecycleIntent.value === 'deactivate') void onConfirmDeactivate()
  else if (lifecycleIntent.value === 'reactivate') void onConfirmReactivate()
}

// reloadAfterConflict trae el estado real tras un conflicto de transición
// (CA-024-04): la lista refleja lo que el servidor confirma, nunca lo que
// el cliente esperaba. Repite la MISMA página/búsqueda que la persona tiene
// en pantalla (nunca vuelve silenciosamente a la página 1 sin filtro): con
// paginación real, el servicio afectado podría no estar en absoluto en la
// primera página, y esta función solo parchea esa fila puntual, no
// reemplaza la lista completa.
async function reloadAfterConflict(serviceId: string) {
  const outcome = await fetchServices({
    page: currentPage.value,
    pageSize: effectivePageSize.value,
    search: activeSearch.value || undefined,
  })
  if (outcome.kind !== 'success') return
  const fresh = outcome.page.items.find((s) => s.id === serviceId)
  if (fresh) replaceServiceInList(fresh)
}
</script>

<template>
  <section
    class="catalog-page"
    :class="{ 'catalog-page--offer': hasOfferColumn() }"
    aria-labelledby="catalog-page-title"
  >
    <header class="catalog-page__header">
      <div>
        <h1 id="catalog-page-title" class="catalog-page__title">Servicios</h1>
        <p class="catalog-page__subtitle">Catálogo de servicios en NAVA.</p>
      </div>
      <BaseButton
        v-if="loadStatus === 'ready'"
        type="button"
        variant="primary"
        class="catalog-page__create"
        @click="openCreateDialog"
      >
        Agregar servicio
      </BaseButton>
    </header>

    <!-- Buscador (issue reportado 2026-09-28, "colócalos en una tabla con
         paginador y con su buscador"): vive FUERA del <Transition> de abajo
         y solo depende de loadStatus, nunca de si la página actual tiene
         resultados — si dependiera de services.length, una búsqueda sin
         coincidencias escondería el propio campo que la persona necesita
         para corregirla o borrarla. Sin botón "Buscar": en vivo con
         debounce (ver onSearchInput), mismo criterio que el mockup
         aprobado por el propietario. -->
    <div v-if="loadStatus === 'ready'" class="catalog-page__toolbar">
      <BaseInput
        type="search"
        name="serviceSearch"
        label="Buscar"
        placeholder="Buscar servicio por nombre…"
        :model-value="searchTerm"
        class="catalog-page__search"
        @update:model-value="onSearchInput"
      />
    </div>

    <!-- Transition envuelve carga/error/listo en un único fundido (issue
         reportado 2026-09-28, "se ve como se genera el objeto... quiero el
         tema de la carga para tapar ese evento"): sin esto, el cambio de
         rombo/esqueleto a la tabla real es un corte instantáneo — con
         useMinHoldLoading ya no "parpadea", pero seguía siendo un corte
         seco. mode="out-in" espera a que lo anterior termine de
         desvanecerse antes de que entre lo nuevo, así nunca se ven ambos
         a la vez ni un salto de layout a medio fundido.

         Mismo tratamiento que AgendaSkeleton.vue (evento 05 del atlas
         panel-agenda-eventos) para el estado de carga: rombo NAVA en
         miniatura + la geometría del contenido por llegar, para que la
         lista real no "aparezca de golpe" (issue reportado 2026-09-28,
         "revisa la Agenda... hace algo antes de mostrar el componente").
         No se reutiliza AgendaSkeleton (documentado como local al módulo
         `agenda`, no promovido a shared/ui) ni el grid de
         `.catalog-page__row`: igual que el propio AgendaSkeleton, que
         tampoco reutiliza las clases reales de DailyAgendaPage.vue, esto
         es una aproximación de la forma con su propio layout responsive,
         no una réplica que dependa de los ganchos de grid-column/row que
         la fila real necesita solo en móvil. Se muestra DE INMEDIATO
         (useMinHoldLoading solo retrasa cuándo se le permite desaparecer,
         ver script). `<Transition>` exige exactamente un elemento hijo en
         el código fuente: este comentario vive antes, no entre medio. -->
    <Transition name="catalog-content" mode="out-in">
      <div
        v-if="loadStatus === 'loading'"
        class="catalog-page__state catalog-page__loading"
        role="status"
        aria-live="polite"
      >
        <DiamondLoader label="Cargando el catálogo…" layout="inline" />
        <ul class="catalog-page__list" aria-hidden="true">
          <li v-for="n in 3" :key="n" class="catalog-page__skeleton-row">
            <span class="catalog-page__skeleton-icon" />
            <span class="catalog-page__skeleton-text">
              <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--name" />
              <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--meta" />
            </span>
            <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--badge" />
            <span class="catalog-page__skeleton-actions">
              <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--button" />
              <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--button" />
            </span>
          </li>
        </ul>
      </div>

      <BaseAlert
        v-else-if="loadStatus === 'load-error'"
        variant="warning"
        title="No pudimos cargar el catálogo"
        role="alert"
      >
        Revisa tu conexión e inténtalo de nuevo.
        <template #action>
          <BaseButton variant="secondary" type="button" @click="onRetryLoad">Reintentar</BaseButton>
        </template>
      </BaseAlert>

      <!-- Un solo elemento, no <template>: <Transition> exige exactamente un
         hijo real por rama del v-if/else-if/else, y esta rama antes tenía
         hasta 3 hermanos sueltos (vacío/tabla, "Cargar más"). -->
      <div
        v-else
        class="catalog-page__ready"
        :class="{ 'catalog-page__ready--fitting': fitPending }"
      >
        <EmptyScene
          v-if="services.length === 0 && pageStatus !== 'loading'"
          :scene="isSearching ? 'search' : 'services'"
          class="catalog-page__empty"
        >
          <template #title>
            <template v-if="isSearching">
              No encontramos servicios que coincidan con «{{ activeSearch }}».
            </template>
            <template v-else>Aún no tienes servicios registrados.</template>
          </template>
          <template #hint>
            <template v-if="isSearching">Prueba con otra palabra o borra la búsqueda.</template>
            <template v-else>
              Crea el primero con su duración y precio; en cuanto lo guardes aparece en tu catálogo.
            </template>
          </template>
          <template v-if="!isSearching" #action>
            <BaseButton type="button" variant="primary" @click="openCreateDialog">
              Agregar servicio
            </BaseButton>
          </template>
        </EmptyScene>

        <div
          v-else
          class="catalog-page__table"
          :class="{
            'catalog-page__table--loading': pageStatus === 'loading',
            'catalog-page__table--offer': hasOfferColumn(),
          }"
          :aria-busy="pageStatus === 'loading'"
        >
          <div class="catalog-page__columns" aria-hidden="true">
            <span>Servicio</span><span>Duración</span><span>Precio (COP)</span><span>Estado</span>
            <span v-if="hasOfferColumn()">Lo ofrezco</span>
          </div>
          <!-- Mientras carga una búsqueda/página, esqueletos con la altura de
             una fila real (ver .catalog-page__skeleton-row--table); al llegar
             el resultado las filas montan de nuevo y entran escalonadas
             (--row-index, ver .catalog-page__row). El mismo <ul> conserva
             tableListRef para el cálculo de filas que caben. -->
          <ul
            ref="tableListRef"
            class="catalog-page__list"
            :aria-label="`Servicios ${v.ofTheBusiness}`"
          >
            <template v-if="pageStatus === 'loading'">
              <li
                v-for="n in skeletonRowCount"
                :key="`skeleton-${n}`"
                class="catalog-page__skeleton-row catalog-page__skeleton-row--table"
                aria-hidden="true"
              >
                <span class="catalog-page__skeleton-icon" />
                <span class="catalog-page__skeleton-text">
                  <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--name" />
                  <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--meta" />
                </span>
                <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--badge" />
                <span class="catalog-page__skeleton-actions">
                  <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--button" />
                  <span class="catalog-page__skeleton-bar catalog-page__skeleton-bar--button" />
                </span>
              </li>
            </template>
            <template v-else>
              <li
                v-for="(service, index) in services"
                :key="service.id"
                class="catalog-page__row"
                :style="{ '--row-index': index }"
              >
                <div class="catalog-page__item-heading">
                  <span class="catalog-page__item-icon" aria-hidden="true">{{
                    serviceInitial(service.name)
                  }}</span>
                  <div class="catalog-page__item-text">
                    <!-- El nombre es el disparador real del detalle (issue
                         2026-09-29, "que en el listado solo esté todo menos la
                         descripción y al hacer clic o tocar aparezca toda la
                         información"): un <button> nativo para teclado y lector
                         de pantalla; su ::after cubre TODA la fila, así clic o
                         toque en cualquier parte la abren, y las acciones
                         (z-index superior) siguen siendo botones independientes. -->
                    <button
                      type="button"
                      class="catalog-page__item-trigger"
                      aria-haspopup="dialog"
                      @click="openDetailDialog(service)"
                    >
                      <span class="catalog-page__item-name" :title="service.name">{{
                        service.name
                      }}</span>
                      <span class="catalog-page__item-chevron" aria-hidden="true">›</span>
                    </button>
                  </div>
                </div>
                <span class="catalog-page__item-meta catalog-page__item-duration">
                  {{ service.durationMinutes }} min
                </span>
                <span class="catalog-page__item-meta catalog-page__item-price">{{
                  formatPrice(service)
                }}</span>
                <BaseBadge
                  :variant="service.isActive ? 'success' : 'neutral'"
                  size="sm"
                  :label="service.isActive ? 'Activo' : 'Inactivo'"
                  :style="diamondStyle(service)"
                  role="status"
                >
                  {{ service.isActive ? 'Activo' : 'Inactivo' }}
                </BaseBadge>
                <div v-if="hasOfferColumn()" class="catalog-page__item-offer">
                  <slot v-if="service.isActive" name="service-offer" :service="service" />
                </div>
                <div class="catalog-page__item-actions">
                  <BaseButton
                    type="button"
                    variant="secondary"
                    :aria-label="`Editar ${service.name}`"
                    @click="openEditDialog(service)"
                  >
                    Editar
                  </BaseButton>
                  <!-- Un solo botón para ambos sentidos (issue 2026-09-29, "quita el
                   desactivar y deja ese reactivar ya que hace ambos"): el
                   diálogo decide desactivar o reactivar según service.isActive
                   y allí, en el botón de confirmación, sí se nombra el efecto. -->
                  <BaseButton
                    type="button"
                    variant="secondary"
                    class="catalog-page__state-action"
                    :aria-label="`Cambiar estado de ${service.name}`"
                    @click="openLifecycleDialog(service)"
                  >
                    Cambiar estado
                  </BaseButton>
                </div>
              </li>
            </template>
          </ul>
        </div>

        <!-- Pie de paginador (issue 2026-09-28, "colócalos en una tabla con
           paginador"): siempre visible en cuanto la carga inicial resuelve
           (incluida una búsqueda sin resultados, "0 resultados para «x»"),
           no solo cuando hay filas — el conteo real es información útil en
           ambos casos. La navegación numerada (nav) solo aparece con más de
           una página; con una sola página el conteo solo queda como
           confirmación silenciosa. -->
        <div ref="paginationFooterRef" class="catalog-page__pagination">
          <BaseAlert
            v-if="pageStatus === 'error'"
            variant="warning"
            title="No pudimos cargar esta página"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo.
            <template #action>
              <BaseButton variant="secondary" type="button" @click="onRetryLoad">
                Reintentar
              </BaseButton>
            </template>
          </BaseAlert>

          <div class="catalog-page__pagination-bar">
            <span class="catalog-page__pagination-count" role="status">{{ resultsLabel }}</span>
            <nav
              v-if="totalPages > 1"
              class="catalog-page__pagination-nav"
              aria-label="Paginación de servicios"
            >
              <BaseButton
                type="button"
                variant="secondary"
                :disabled="currentPage === 1 || pageStatus === 'loading'"
                aria-label="Página anterior"
                @click="onPrevPage"
              >
                ‹
              </BaseButton>
              <template v-for="(token, index) in pageTokens" :key="index">
                <span
                  v-if="token === 'ellipsis'"
                  class="catalog-page__pagination-ellipsis"
                  aria-hidden="true"
                >
                  …
                </span>
                <BaseButton
                  v-else
                  type="button"
                  :variant="token === currentPage ? 'primary' : 'secondary'"
                  :aria-current="token === currentPage ? 'page' : undefined"
                  :disabled="pageStatus === 'loading'"
                  @click="goToPage(token)"
                >
                  {{ token }}
                </BaseButton>
              </template>
              <BaseButton
                type="button"
                variant="secondary"
                :disabled="currentPage === totalPages || pageStatus === 'loading'"
                aria-label="Página siguiente"
                @click="onNextPage"
              >
                ›
              </BaseButton>
            </nav>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Detalle (issue 2026-09-29): la descripción completa vive aquí, no en la
         fila, porque puede ser larga. Va ANTES de los demás diálogos a
         propósito: al pulsar "Editar" se cierra este y se abre el de edición
         en el mismo ciclo, y BaseDialog guarda como "foco previo" el elemento
         activo al abrirse; con este orden el foco ya volvió a la fila cuando
         el de edición lo lee. Mismo tinte tinta/dorado y ficha con la
         inicial que los demás diálogos de la pantalla. -->
    <BaseDialog
      v-model="isDetailOpen"
      :title="detailTarget?.name"
      size="sm"
      content-class="catalog-page__detail-dialog"
    >
      <template #icon>
        <span class="catalog-page__detail-icon" aria-hidden="true">{{
          detailTarget ? serviceInitial(detailTarget.name) : ''
        }}</span>
      </template>
      <!-- catalog-page__lifecycle aporta solo el apilado, el color de texto
           sobre tinta y el par de botones dorado/fantasma ya calibrado para
           este fondo; el resto es propio del detalle. -->
      <div v-if="detailTarget" class="catalog-page__lifecycle catalog-page__detail">
        <p
          class="catalog-page__detail-status"
          :class="{ 'catalog-page__detail-status--active': detailTarget.isActive }"
          :style="{ '--row-index': 0, ...diamondStyle(detailTarget) }"
        >
          <span class="catalog-page__detail-diamond" aria-hidden="true" />
          {{ detailTarget.isActive ? 'Activo' : 'Inactivo' }}
        </p>

        <dl class="catalog-page__lifecycle-facts catalog-page__detail-facts">
          <div :style="{ '--row-index': 1 }">
            <dt>Duración</dt>
            <dd>
              {{ detailTarget.durationMinutes }}
              <small>min</small>
            </dd>
          </div>
          <div :style="{ '--row-index': 2 }">
            <dt>Precio</dt>
            <dd>
              {{ detailTarget.price }}
              <small>{{ detailTarget.currency }}</small>
            </dd>
          </div>
        </dl>

        <div class="catalog-page__detail-section" :style="{ '--row-index': 3 }">
          <h3 class="catalog-page__detail-label">Descripción</h3>
          <p v-if="detailTarget.description" class="catalog-page__detail-text">
            {{ detailTarget.description }}
          </p>
          <p v-else class="catalog-page__detail-text catalog-page__detail-text--empty">
            Este servicio no tiene descripción.
          </p>
        </div>

        <dl class="catalog-page__detail-meta" :style="{ '--row-index': 4 }">
          <div>
            <dt>Registrado</dt>
            <dd>{{ formatDate(detailTarget.createdAt) }}</dd>
          </div>
          <div v-if="detailTarget.deactivatedAt">
            <dt>Inactivo desde</dt>
            <dd>{{ formatDate(detailTarget.deactivatedAt) }}</dd>
          </div>
        </dl>
      </div>
      <!-- Pie fuera del área con scroll: con una descripción larga solo esta
           se desplaza y "Cerrar"/"Editar" nunca quedan fuera de la pantalla. -->
      <template #footer>
        <div class="catalog-page__lifecycle catalog-page__detail-footer">
          <BaseButton type="button" variant="secondary" @click="isDetailOpen = false">
            Cerrar
          </BaseButton>
          <BaseButton type="button" variant="primary" @click="onDetailEdit">Editar</BaseButton>
        </div>
      </template>
    </BaseDialog>

    <!-- Alta. contentClass="catalog-page__create-dialog" + #icon (issue
         reportado 2026-09-28, "ajusta esa ventana emergente, se ve como una
         simple ventana que aparece de la nada, sorpréndeme"; luego "aplica
         eso mismo para las pantallas emergentes de editar y eliminar", ver
         Editar y Desactivar/Reactivar más abajo — los tres diálogos
         comparten ya el mismo tinte, ficha con entrada animada y campos
         reglados). Ver estilos al final del bloque <style> para el porqué
         de cada valor (ninguno inventado: reutiliza --border-width-emphasis
         y el mismo cubic-bezier con rebote que ya usa BarberSelect.vue). -->
    <BaseDialog
      v-model="isCreateOpen"
      title="Agregar servicio"
      description="Aparece en tu catálogo activo en cuanto lo guardes."
      size="sm"
      content-class="catalog-page__create-dialog"
      @close="onCreateDialogClosed"
    >
      <template #icon>
        <span class="catalog-page__create-icon" aria-hidden="true">+</span>
      </template>
      <form
        class="catalog-page__form catalog-page__create-form"
        novalidate
        @submit.prevent="onSubmitCreate"
      >
        <BaseAlert
          v-if="createStatus === 'name-conflict'"
          variant="danger"
          title="Ese nombre ya está en uso"
          role="alert"
        >
          Ya existe un servicio activo con ese nombre. Usa otro nombre.
        </BaseAlert>
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

        <BaseInput
          :model-value="createForm.name"
          name="name"
          label="Nombre"
          required
          :maxlength="120"
          :disabled="createStatus === 'saving'"
          :error="createFieldErrors.name"
          :class="{ 'catalog-page__create-input--filled': !!createForm.name }"
          @update:model-value="onCreateNameInput"
        />
        <BaseInput
          :model-value="createForm.description"
          name="description"
          label="Descripción (opcional)"
          :maxlength="500"
          :disabled="createStatus === 'saving'"
          :error="createFieldErrors.description"
          :class="{ 'catalog-page__create-input--filled': !!createForm.description }"
          @update:model-value="onCreateDescriptionInput"
        />
        <BaseInput
          :model-value="createDurationRaw"
          type="number"
          name="durationMinutes"
          label="Duración (minutos)"
          required
          :min="1"
          :max="1440"
          :step="1"
          :disabled="createStatus === 'saving'"
          :error="createFieldErrors.durationMinutes"
          :class="{ 'catalog-page__create-input--filled': !!createDurationRaw }"
          @update:model-value="onCreateDurationInput"
        />
        <BaseInput
          :model-value="createForm.price"
          name="price"
          label="Precio (COP)"
          required
          placeholder="45000.00"
          hint="Precio informativo en pesos colombianos, mayor que cero."
          :disabled="createStatus === 'saving'"
          :error="createFieldErrors.price"
          :class="{ 'catalog-page__create-input--filled': !!createForm.price }"
          @update:model-value="onCreatePriceInput"
        />

        <div class="catalog-page__dialog-actions">
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

    <!-- Edición (issue 2026-09-28, "aplica eso mismo para las pantallas
         emergentes de editar y eliminar"): mismo tratamiento tinta/dorado
         que el diálogo de alta, arriba. La ficha del encabezado reutiliza
         serviceInitial() — la misma inicial que ya identifica a este
         servicio en la fila — en vez de un glifo genérico, así la ficha
         señala DE QUÉ servicio se trata, no solo que es "un" diálogo de
         edición. Ver estilos junto a .catalog-page__create-* y el bloque
         sin scope al final del archivo. -->
    <BaseDialog
      v-model="isEditOpen"
      title="Editar servicio"
      size="sm"
      content-class="catalog-page__edit-dialog"
      @close="onEditDialogClosed"
    >
      <template #icon>
        <span class="catalog-page__edit-icon" aria-hidden="true">{{
          editTarget ? serviceInitial(editTarget.name) : ''
        }}</span>
      </template>
      <form
        class="catalog-page__form catalog-page__edit-form"
        novalidate
        @submit.prevent="onSubmitEdit"
      >
        <BaseAlert
          v-if="editStatus === 'name-conflict'"
          variant="danger"
          title="Ese nombre ya está en uso"
          role="alert"
        >
          Ya existe un servicio activo con ese nombre. Usa otro nombre.
        </BaseAlert>
        <BaseAlert
          v-if="editStatus === 'not-found'"
          variant="warning"
          title="Este servicio ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-if="editStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
        </BaseAlert>
        <BaseAlert
          v-if="editStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
        </BaseAlert>

        <BaseInput
          :model-value="editForm.name"
          name="name"
          label="Nombre"
          required
          :maxlength="120"
          :disabled="editStatus === 'saving'"
          :error="editFieldErrors.name"
          :class="{ 'catalog-page__edit-input--filled': !!editForm.name }"
          @update:model-value="onEditNameInput"
        />
        <BaseInput
          :model-value="editForm.description"
          name="description"
          label="Descripción (opcional)"
          :maxlength="500"
          :disabled="editStatus === 'saving'"
          :error="editFieldErrors.description"
          :class="{ 'catalog-page__edit-input--filled': !!editForm.description }"
          @update:model-value="onEditDescriptionInput"
        />
        <BaseInput
          :model-value="editDurationRaw"
          type="number"
          name="durationMinutes"
          label="Duración (minutos)"
          required
          :min="1"
          :max="1440"
          :step="1"
          :disabled="editStatus === 'saving'"
          :error="editFieldErrors.durationMinutes"
          :class="{ 'catalog-page__edit-input--filled': !!editDurationRaw }"
          @update:model-value="onEditDurationInput"
        />
        <BaseInput
          :model-value="editForm.price"
          name="price"
          label="Precio (COP)"
          required
          placeholder="45000.00"
          hint="Precio informativo en pesos colombianos, mayor que cero."
          :disabled="editStatus === 'saving'"
          :error="editFieldErrors.price"
          :class="{ 'catalog-page__edit-input--filled': !!editForm.price }"
          @update:model-value="onEditPriceInput"
        />

        <div class="catalog-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isEditOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="submit"
            variant="primary"
            :loading="editStatus === 'saving'"
            :disabled="editStatus === 'saving'"
          >
            Guardar
          </BaseButton>
        </div>
      </form>
    </BaseDialog>

    <!-- HU-024: desactivar/reactivar (issue 2026-09-28, "aplica eso mismo
         para las pantallas emergentes de editar y eliminar"): mismo tinte
         base que alta/edición, pero la ficha y el filete del encabezado
         cambian a --color-danger-on-strong cuando la intención es
         "deactivate" (mismo rojo ya usado por el botón "Desactivar" de la
         fila) — el diálogo advierte visualmente antes de que el barbero
         lea una sola palabra. "reactivate" conserva el dorado neutro. -->
    <BaseDialog
      v-model="isLifecycleOpen"
      title="Estado del servicio"
      size="sm"
      :content-class="
        lifecycleIntent === 'deactivate'
          ? 'catalog-page__lifecycle-dialog catalog-page__lifecycle-dialog--danger'
          : 'catalog-page__lifecycle-dialog'
      "
      @close="onLifecycleDialogClosed"
    >
      <template #icon>
        <span
          class="catalog-page__lifecycle-icon"
          :class="{ 'catalog-page__lifecycle-icon--danger': lifecycleIntent === 'deactivate' }"
          aria-hidden="true"
          >{{ lifecycleTarget ? serviceInitial(lifecycleTarget.name) : '' }}</span
        >
      </template>
      <div class="catalog-page__lifecycle">
        <BaseAlert
          v-if="lifecycleStatus === 'not-found'"
          variant="warning"
          title="Este servicio ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
        </BaseAlert>
        <BaseAlert
          v-else-if="lifecycleStatus === 'transition-conflict'"
          variant="warning"
          title="El estado de este servicio cambió"
          role="alert"
        >
          Alguien más ya
          {{ lifecycleIntent === 'deactivate' ? 'lo desactivó' : 'lo reactivó' }}. Cierra este
          diálogo: la lista ya se actualizó con el estado real.
        </BaseAlert>
        <BaseAlert
          v-else-if="lifecycleStatus === 'idempotency-conflict'"
          variant="danger"
          title="No pudimos completar el intento anterior"
          role="alert"
        >
          Cierra este diálogo y vuelve a intentarlo.
        </BaseAlert>
        <BaseAlert
          v-else-if="lifecycleStatus === 'network-error'"
          variant="warning"
          title="No pudimos conectar"
          role="alert"
        >
          Revisa tu conexión e inténtalo de nuevo.
        </BaseAlert>
        <BaseAlert
          v-else-if="lifecycleStatus === 'unexpected-error'"
          variant="danger"
          title="Ocurrió un error inesperado"
          role="alert"
        >
          Inténtalo de nuevo en unos segundos.
        </BaseAlert>

        <template v-if="lifecycleIntent">
          <p
            v-if="lifecycleIntent === 'deactivate' && impactStatus === 'loading'"
            class="catalog-page__lifecycle-lead"
            role="status"
            aria-live="polite"
          >
            Consultando el impacto real…
          </p>
          <p v-else-if="lifecycleIntent === 'deactivate'" class="catalog-page__lifecycle-lead">
            <strong>{{ lifecycleTarget?.name }}</strong> dejará de estar disponible para nuevas
            reservas.
          </p>
          <p v-else class="catalog-page__lifecycle-lead">
            <strong>{{ lifecycleTarget?.name }}</strong> vuelve a ofrecerse en el catálogo.
          </p>

          <!-- El interruptor ES la acción (issue 2026-09-29, "quita ese botón de
               activar o desactivar y solo deja ese interruptor"): mover el
               interruptor confirma el cambio. La consecuencia (impacto real
               al desactivar) ya está a la vista en este mismo diálogo antes
               de poder tocarlo, y se deshabilita mientras se consulta o se
               guarda, así que no es un cambio silencioso. -->
          <button
            type="button"
            role="switch"
            class="catalog-page__lifecycle-switch"
            :class="{ 'catalog-page__lifecycle-switch--on': lifecycleSwitchOn }"
            :aria-checked="lifecycleSwitchOn"
            aria-label="Servicio activo"
            :aria-busy="lifecycleStatus === 'saving'"
            :disabled="lifecycleStatus === 'saving' || impactStatus === 'loading'"
            @click="onConfirmLifecycle"
          >
            <span class="catalog-page__lifecycle-state">Inactivo</span>
            <span class="catalog-page__lifecycle-track" aria-hidden="true">
              <span class="catalog-page__lifecycle-thumb"></span>
            </span>
            <span class="catalog-page__lifecycle-state catalog-page__lifecycle-state--on"
              >Activo</span
            >
          </button>
          <p class="catalog-page__lifecycle-hint">
            {{
              lifecycleIntent === 'deactivate'
                ? 'Mueve el interruptor para desactivarlo.'
                : 'Mueve el interruptor para reactivarlo.'
            }}
          </p>
          <dl class="catalog-page__lifecycle-facts">
            <div>
              <dt>Duración</dt>
              <dd>
                {{ lifecycleTarget?.durationMinutes }}
                <small>min</small>
              </dd>
            </div>
            <div>
              <dt>Precio</dt>
              <dd>
                {{ lifecycleTarget?.price }}
                <small>{{ lifecycleTarget?.currency }}</small>
              </dd>
            </div>
            <div v-if="deactivatedSince">
              <dt>Desde</dt>
              <dd>
                {{ deactivatedSince }}
                <small>inactivo</small>
              </dd>
            </div>
          </dl>

          <template v-if="lifecycleIntent === 'deactivate'">
            <template v-if="impactStatus === 'ready'">
              <div
                class="catalog-page__lifecycle-impact"
                :class="{ 'catalog-page__lifecycle-impact--warning': affectedAppointments > 0 }"
                role="status"
              >
                <span class="catalog-page__lifecycle-impact-icon" aria-hidden="true">{{
                  affectedAppointments > 0 ? '!' : '✓'
                }}</span>
                <p class="catalog-page__lifecycle-impact-text">
                  <template v-if="affectedAppointments > 0">
                    <strong>{{ affectedAppointments }}</strong>
                    {{
                      affectedAppointments === 1
                        ? 'cita futura no se cancela'
                        : 'citas futuras no se cancelan'
                    }}
                    automáticamente.
                  </template>
                  <template v-else>No hay citas futuras afectadas.</template>
                </p>
              </div>
              <p class="catalog-page__lifecycle-note">
                El historial se conserva; el servicio nunca se elimina.
              </p>
            </template>
          </template>
          <ul v-else class="catalog-page__lifecycle-checks">
            <li>Vuelve a aparecer para nuevas reservas.</li>
            <li>Conserva duración, precio y asignaciones.</li>
            <li>No se crea otro registro; el historial sigue intacto.</li>
          </ul>
        </template>

        <div class="catalog-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isLifecycleOpen = false">
            Salir
          </BaseButton>
        </div>
      </div>
    </BaseDialog>
  </section>
</template>

<style scoped>
/* Fundido entre carga/error/listo (ver <Transition> en la plantilla,
   issue 2026-09-28): --motion-duration-base, el mismo token que usa el
   resto de transiciones de estado del sistema (estandar-diseno-visual.md
   §12), no un valor suelto. */
.catalog-content-enter-active,
.catalog-content-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.catalog-content-enter-from,
.catalog-content-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .catalog-content-enter-active,
  .catalog-content-leave-active {
    transition: none;
  }
}

.catalog-page {
  /* Superficie tinta de punta a punta (estandar-diseno-visual.md §3, mismo
     canvas que DailyAgendaPage.vue y el cascarón AppHeader/AppNav): antes
     esta pantalla se quedó en el blanco previo a la Fase 2 del rediseño
     Tailored Grid y desentonaba del resto del panel privado (issue
     reportado por el propietario 2026-09-28, "el fondo está en blanco").
     BaseDialog/BaseAlert/BaseInput no se tocan: son superficies claras
     pensadas para flotar sobre cualquier fondo (mismo criterio que ya usan
     los diálogos y avisos de DailyAgendaPage.vue sobre tinta). */
  /* 820px (antes 760px, luego 580px): cada vez que la letra de la tabla
     sube un escalón (issues 2026-09-28 "la tabla se ve fea" y "la letra se
     ve pequeña") las columnas fijas necesitan más aire para no apretar
     precio/insignia/acciones. El resto de la pantalla sigue centrado como
     columna de lectura (no se estira a todo el ancho); solo se ensancha lo
     que la tabla necesita. */
  --catalog-width: 920px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
  max-width: none;
  padding: 34px 32px 48px;
  margin: 0 auto;
  color: var(--color-on-strong);
  /* Transparente: la tinta y el fondo animado los pone el cascarón. */
  background: transparent;
}

/* Con la columna «Lo ofrezco» (DEC-115) la tabla gana 120px para que el nombre del
   servicio no se recorte; sin ella conserva su ancho de siempre. */
.catalog-page--offer {
  --catalog-width: 940px;
}

.catalog-page__header,
.catalog-page__toolbar,
.catalog-page__table,
.catalog-page__list,
.catalog-page__state,
.catalog-page__empty,
.catalog-page__pagination,
.catalog-page > :deep(.base-alert) {
  width: min(100%, var(--catalog-width));
  margin-inline: auto;
}

/* Antes estos hermanos (vacío/tabla, "Cargar más") eran hijos directos del
   flex column de .catalog-page y heredaban su gap; envueltos en un solo
   elemento para <Transition> (ver plantilla), necesitan su propio flex
   para no perder ese espaciado. */
.catalog-page__ready {
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

/* Invisible pero con layout: el primer ajuste de filas mide la tabla real
   (ver fitPending en el script). */
.catalog-page__ready--fitting {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .catalog-page__ready {
    transition: none;
  }
}

.catalog-page__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
  padding-bottom: 16px;
  border-bottom: var(--border-width-normal) solid var(--color-field-strong-border);
}

/* h1 del sistema (antes 30px suelto), mismo escalón que DailyAgendaPage.vue
   usa para su propio título sobre tinta (issue 2026-09-28, "el título se ve
   muy pequeño"): 32px base, 40px desde 1024px — Catalog se había quedado
   más chico que su página hermana sin motivo. */
.catalog-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
}

@media (min-width: 1024px) {
  .catalog-page__title {
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

/* body-sm (antes 11px suelto): mismo reclamo del propietario sobre el
   subtítulo. Ya no encoge en el breakpoint móvil (ver media query de abajo,
   donde antes bajaba a 10px): a ese ancho es justo donde más se notaba
   "muy pequeño". */
.catalog-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
}

.catalog-page__create.base-button {
  height: 40px;
  padding-inline: 18px;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
}

.catalog-page__create :deep(.base-button__content)::before {
  content: '+';
  margin-right: 6px;
}

/* "Agregar servicio" es la acción primaria de la pantalla: el relleno tinta
   por defecto de BaseButton--primary (--color-action-primary) es el MISMO
   color que ahora tiene el fondo de la página (#101b2b), así que se
   fundiría con el canvas. Se levanta con el dorado de marca, igual criterio
   que ya usa BaseBadge--outline y el filete activo de AppNav sobre tinta:
   ningún color nuevo, solo los tokens de acento ya pensados para superficie
   oscura (--color-brand-accent-surface/--color-brand-accent-text). Se
   sobrescribe aquí, no en BaseButton, porque el resto de sus usos siguen
   sobre superficie clara. */
.catalog-page__create.base-button--primary {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.catalog-page__create.base-button--primary:hover:not(:disabled):not(.base-button--loading) {
  filter: brightness(92%);
}

.catalog-page__create.base-button--primary:active:not(:disabled):not(.base-button--loading) {
  filter: brightness(84%);
}

.catalog-page__state {
  padding: var(--space-4);
  color: var(--color-on-strong-muted);
}

/* layout="inline" (rombo chico a la izquierda, mismo criterio que
   AgendaSkeleton) en vez del rombo grande centrado: aquí no está solo,
   acompaña a las filas placeholder de abajo. */
.catalog-page__loading {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  padding: 0;
}

/* Filas placeholder: layout propio (no el grid de `.catalog-page__row`,
   que solo en móvil necesita ganchos de grid-column/row muy específicos
   para su insignia/acciones) — mismo criterio que AgendaSkeleton.vue, que
   tampoco reutiliza las clases reales de fila de DailyAgendaPage.vue: es
   una aproximación de la forma, con su propio responsive, no una réplica
   acoplada a los detalles internos de la fila real. */
.catalog-page__skeleton-row {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 80px;
  padding: 16px;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.catalog-page__skeleton-text {
  display: flex;
  flex: 1;
  min-width: 0;
  flex-direction: column;
  gap: 6px;
}

.catalog-page__skeleton-actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
}

/* Mismo `agenda-skeleton-pulse` que AgendaSkeleton.vue, aquí
   `catalog-skeleton-pulse`. */
.catalog-page__skeleton-icon,
.catalog-page__skeleton-bar {
  display: block;
  flex: 0 0 auto;
  background-color: color-mix(in srgb, var(--color-on-strong) 16%, transparent);
  border-radius: var(--radius-sm);
  animation: catalog-skeleton-pulse 1400ms ease-in-out infinite;
}

.catalog-page__skeleton-icon {
  width: 40px;
  height: 40px;
}

.catalog-page__skeleton-bar--name {
  width: 55%;
  max-width: 180px;
  height: 16px;
}

.catalog-page__skeleton-bar--meta {
  width: 96px;
  height: 12px;
  background-color: color-mix(in srgb, var(--color-on-strong) 10%, transparent);
}

.catalog-page__skeleton-bar--badge {
  width: 64px;
  height: 12px;
}

.catalog-page__skeleton-bar--button {
  width: 72px;
  height: 34px;
}

@keyframes catalog-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .catalog-page__skeleton-icon,
  .catalog-page__skeleton-bar {
    animation: none;
    opacity: 0.8;
  }
}

@media (max-width: 640px) {
  .catalog-page__skeleton-row {
    gap: 10px;
    min-height: 62px;
    padding: 9px 10px;
  }

  .catalog-page__skeleton-icon {
    width: 28px;
    height: 28px;
  }

  .catalog-page__skeleton-bar--name {
    max-width: 120px;
    height: 13px;
  }

  .catalog-page__skeleton-bar--meta {
    width: 70px;
    height: 10px;
  }

  .catalog-page__skeleton-bar--badge {
    width: 48px;
    height: 10px;
  }

  .catalog-page__skeleton-actions {
    flex-direction: column;
    gap: 4px;
  }

  .catalog-page__skeleton-bar--button {
    width: 52px;
    height: 22px;
  }
}

/* Panel hundido sobre tinta (--color-field-strong, tokens.css: "así leen
   como un hueco en la página y no como un velo claro sobre el azul"), mismo
   vocabulario que los campos de BarberSelect/AgendaDatePicker. */
.catalog-page__list {
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

.catalog-page__columns,
.catalog-page__row {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) 84px 124px 108px 232px;
  align-items: center;
  gap: 12px;
}

.catalog-page__columns {
  /* 16px (antes 14px, luego 10px): la letra de todo el cuerpo de la tabla
     subió un escalón más (issue 2026-09-28 "la letra se ve pequeña" otra
     vez), y el encabezado sigue el mismo --font-size-body para no quedar
     más chico que los datos que etiqueta. */
  min-height: 44px;
  padding: 0 16px;
  margin: 0;
  border: none;
  font-family: var(--font-sans);
  /* Encabezado reglado: versalitas espaciadas de latón, el mismo rótulo que
     usan los campos y las demás pantallas del panel (antes un gris en negrita
     de 16px que competía con los datos que etiqueta). */
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

/* Alto FIJO, no min-height (issue reportado 2026-09-28, "que se adapte
   según la pantalla, que no haya scroll"): calcular cuántas filas caben en
   la pantalla disponible (ver ROW_HEIGHT_PX/computeFittingPageSize en el
   script) exige que TODA fila mida exactamente lo mismo, tenga o no
   descripción — si el alto variara con el contenido, ese cálculo dejaría
   de predecir el alto real y volvería a aparecer scroll. */
.catalog-page__row {
  height: 80px;
  padding: 16px;
  overflow: hidden;
  border-top: var(--border-width-normal) solid var(--color-field-strong-border);
}

.catalog-page__item-heading {
  display: flex;
  align-items: center;
  min-width: 0;
  gap: 12px;
}

/* Monograma del servicio: reemplaza el glifo "▧" genérico (issue reportado
   2026-09-28, "se ve fea") por una ficha que encierra información real (la
   inicial ayuda a escanear una lista larga), mismo vocabulario de campo
   hundido que el resto de la pantalla sobre tinta, sin avatar circular
   (StaffPage lo usa para personas; un servicio no es una persona). */
.catalog-page__item-icon {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: var(--color-field-strong-raised);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
  color: var(--color-brand-accent-surface);
  font-family: var(--font-display);
  font-size: var(--font-size-body-lg);
  line-height: 1;
}

.catalog-page__item-text {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 2px;
}

.catalog-page__item-name {
  font-family: var(--font-family-base);
  /* body-lg, no body (issue 2026-09-28 "la letra se ve pequeña"): el nombre
     es el dato principal de la fila, un escalón por encima de duración/
     precio/descripción. */
  font-size: var(--font-size-body-lg);
  font-weight: 500;
  color: var(--color-on-strong);
  /* Una sola línea con "…" (issue 2026-09-28, "que se adapte según la
     pantalla, que no haya scroll"): un nombre largo ya no empuja el alto de
     la fila — el `title` del propio elemento (ver plantilla) muestra el
     texto completo al pasar el cursor, y un lector de pantalla sigue
     leyendo el texto real completo, nunca lo visualmente recortado. */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Disparador del detalle: botón sin cromo que hereda la tipografía del
   nombre. Su ::after se estira sobre toda la fila (.catalog-page__row es
   position: relative), de modo que cualquier clic o toque la abre; las
   acciones suben por z-index y conservan su propio clic. */
.catalog-page__item-trigger {
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

.catalog-page__item-trigger::after {
  content: '';
  position: absolute;
  inset: 0;
}

.catalog-page__item-trigger:focus-visible {
  outline: none;
}

.catalog-page__item-trigger:focus-visible::after {
  box-shadow: inset 0 0 0 2px var(--color-focus);
}

.catalog-page__item-chevron {
  flex: 0 0 auto;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-body-lg);
  line-height: 1;
  opacity: 0.55;
  transition:
    opacity var(--motion-duration-fast) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

.catalog-page__row {
  position: relative;
  transition: background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

/* Solo con puntero real: en táctil :hover se queda pegado tras el toque. */
@media (hover: hover) {
  .catalog-page__row:hover {
    background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  }

  .catalog-page__row:hover .catalog-page__item-name {
    color: var(--color-brand-accent-surface);
  }

  .catalog-page__row:hover .catalog-page__item-chevron {
    opacity: 1;
    transform: translateX(3px);
  }
}

.catalog-page__item-name {
  transition: color var(--motion-duration-fast) var(--motion-easing-standard);
}

/* Columna «Lo ofrezco» (DEC-115): una pista de 46px sin texto; el encabezado y el
   nombre accesible del interruptor dicen qué es. */
.catalog-page__table--offer .catalog-page__columns,
.catalog-page__table--offer .catalog-page__row {
  grid-template-columns: minmax(200px, 1fr) 80px 116px 100px 88px 204px;
}

.catalog-page__table--offer .catalog-page__columns > :nth-child(5) {
  white-space: nowrap;
}

.catalog-page__item-offer {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
}

.catalog-page__item-actions {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.catalog-page__item-actions :deep(.base-button) {
  height: 36px;
  padding-inline: 14px;
  font-size: var(--font-size-body-sm);
}

/* "Editar"/"Cambiar estado" usan BaseButton--secondary, calibrado para relleno
   blanco sobre superficie clara (estandar-diseno-visual.md §8.1). Mismo
   ajuste ya aplicado en DailyAgendaPage.vue (.daily-agenda-page__date-nav)
   para su BaseButton--ghost: se quita el relleno y se sube el latón claro,
   sin tocar el componente compartido porque sus demás usos sí están sobre
   superficie clara. */
.catalog-page__item-actions :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.catalog-page__item-actions
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.catalog-page__item-actions
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

/* "Cambiar estado" (issue 2026-09-29, "colócale un color distinto a ese
   botón"): mismo ghost sobre tinta que "Editar", pero en salvia
   (--color-success-on-strong, ya calculado para AA sobre tinta y el mismo
   verde del interruptor del diálogo que abre) para separarlo de la acción
   de edición (latón) sin recurrir al rojo, que sigue reservado a la
   confirmación destructiva. */
.catalog-page__item-actions :deep(.catalog-page__state-action.base-button--secondary) {
  color: var(--color-success-on-strong);
  border-color: color-mix(in srgb, var(--color-success-on-strong) 50%, transparent);
  border-bottom-color: var(--color-success-on-strong);
}

.catalog-page__item-actions
  :deep(
    .catalog-page__state-action.base-button--secondary:hover:not(:disabled):not(
        .base-button--loading
      )
  ) {
  background-color: color-mix(in srgb, var(--color-success-on-strong) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-success-on-strong) 50%, transparent);
}

.catalog-page__item-actions
  :deep(
    .catalog-page__state-action.base-button--secondary:active:not(:disabled):not(
        .base-button--loading
      )
  ) {
  background-color: color-mix(in srgb, var(--color-success-on-strong) 20%, transparent);
}

/* Insignia "Activo"/"Inactivo" como rombo (issue reportado 2026-09-28,
   "esos estados se ven feos, colócalos como diamantes... con animación
   como en agenda"): mismo rediseño local que ya usa DailyAgendaPage.vue
   (issue #189, pedido explícito del propietario 2026-09-04: "parecen
   botones") — se abandona la caja de BaseBadge por un rótulo: rombo de
   color + versalitas espaciadas. Sigue sin inventar un color nuevo: el
   rombo es `currentColor`, tomado del `color` que ya resuelve la variante
   (los mismos tintes on-strong de abajo). Se sobrescribe aquí, no en
   BaseBadge, porque el resto de sus usos en la app sí necesitan la caja
   con relleno sobre superficie clara. */
.catalog-page__row :deep(.base-badge) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: auto;
  padding: 0;
  background-color: transparent;
  border: none;
  border-radius: 0;
  font-size: var(--font-size-caption);
  font-weight: 600;
  line-height: var(--font-size-caption-line);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.catalog-page__row :deep(.base-badge)::before {
  content: '';
  flex-shrink: 0;
  width: 5px;
  height: 5px;
  /* Rombo (cuadrado girado 45°), igual que el marcador de agenda. */
  transform: rotate(45deg);
  background-color: currentColor;
  /* Ritmo propio por fila (ver diamondStyle): --diamond-duration/-delay
     vienen del `:style` de cada BaseBadge, con esta duración/desfase de
     respaldo si faltaran. */
  animation: catalog-diamond-blink var(--diamond-duration, 7s) ease-in-out var(--diamond-delay, 0s)
    infinite;
}

/* Parpadeo suave y esporádico: casi todo el ciclo queda quieto y solo un
   tramo corto baja la opacidad y la escala, sin movimiento lateral (mismos
   valores que agenda-diamond-blink en DailyAgendaPage.vue). */
@keyframes catalog-diamond-blink {
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

@media (prefers-reduced-motion: reduce) {
  .catalog-page__row :deep(.base-badge)::before {
    animation: none;
  }
}

/* Tintes YA calculados para AA sobre tinta (tokens.css §"Tintes de estado
   LEVANTADOS al canvas de tinta"), ningún color nuevo: el `color` de cada
   variant también es el que hereda el rombo vía currentColor. */
.catalog-page__row :deep(.base-badge--success) {
  color: var(--color-success-on-strong);
}

.catalog-page__row :deep(.base-badge--neutral) {
  color: var(--color-on-strong-muted);
}

.catalog-page__item-meta {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  color: var(--color-on-strong-muted);
  /* Cifras tabulares: duración y precio (estandar-diseno-visual.md §5.2, §10.3). */
  font-variant-numeric: tabular-nums;
  /* Nunca en dos líneas (issue reportado 2026-09-28, "los botones se
     desencajan todo raro"): .catalog-page__row tiene alto FIJO (ver
     ROW_HEIGHT_PX en el script) — un precio largo envolviendo a una
     segunda línea se recortaría en silencio en vez de solo desbordar. */
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Buscador (issue 2026-09-28, "con su buscador"): BaseInput type="search"
   reutilizado tal cual, sin campo nuevo — solo se acota su ancho (no tiene
   sentido que un solo input ocupe las mismas 820px que la tabla completa)
   y se reaplica el mismo tratamiento "reglado sobre tinta" que ya usan los
   demás BaseInput de esta pantalla (diálogos de alta/edición), porque
   shared/ui/BaseInput sigue calibrado para superficie clara
   (estandar-diseno-visual.md). */
.catalog-page__toolbar {
  display: flex;
}

.catalog-page__search {
  width: 100%;
  max-width: 320px;
}

.catalog-page__search :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  color: var(--color-on-strong);
}

.catalog-page__search :deep(.base-input__label) {
  color: var(--color-brand-accent-surface);
}

.catalog-page__search :deep(.base-input::placeholder) {
  color: var(--color-on-strong-muted);
  opacity: 1;
}

.catalog-page__search :deep(.base-input:hover:not(:disabled)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

/* Buscar o cambiar de página (issue 2026-09-29, "que al buscar un servicio
   se haga una animación y no que muestre el resultado, ya que se ve feo"):
   la tabla y su encabezado de columnas se quedan (nunca el esqueleto de
   página completa de la carga inicial, eso sería el parpadeo que
   useMinHoldLoading ya resolvió); solo el cuerpo pasa a esqueletos con
   pulso -misma altura que una fila real, así el paginador no salta- y el
   resultado entra fila por fila. */
.catalog-page__skeleton-row--table {
  height: 80px;
  min-height: 0;
  overflow: hidden;
}

/* Entrada escalonada: cada fila arranca `--row-index` pasos después que la
   anterior (tope de 8 pasos para que una página de 50 filas no tarde en
   terminar de aparecer). `both` mantiene la fila oculta durante su retraso. */
.catalog-page__row {
  animation: catalog-row-enter 320ms var(--motion-easing-standard) both;
  animation-delay: calc(min(var(--row-index, 0), 8) * 45ms);
}

@media (max-width: 640px) {
  /* Igual que .catalog-page__row en móvil (ROW_HEIGHT_MOBILE_PX). */
  .catalog-page__skeleton-row--table {
    height: 100px;
  }
}

@keyframes catalog-row-enter {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .catalog-page__row {
    animation: none;
  }
}

/* Pie de paginador (issue 2026-09-28, "con paginador"). */
.catalog-page__pagination {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.catalog-page__pagination-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}

.catalog-page__pagination-count {
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong-muted);
}

.catalog-page__pagination-nav {
  display: flex;
  align-items: center;
  gap: 6px;
}

.catalog-page__pagination-ellipsis {
  padding-inline: 4px;
  color: var(--color-on-strong-muted);
}

/* Botones del paginador: mismo par dorado-sólido (página vigente)/tinta-
   fantasma (el resto) que ya usa el CTA "Agregar servicio" y las acciones
   de fila — no los valores por defecto de BaseButton, calibrados para
   flotar sobre superficie clara. Cuadrados y compactos (34px, sin relleno
   horizontal de sobra): un paginador numerado vive de la repetición, no
   necesita el mismo padding que un botón de acción con texto largo. */
.catalog-page__pagination-nav :deep(.base-button) {
  height: 38px;
  min-width: 38px;
  padding-inline: 10px;
  font-size: var(--font-size-body-sm);
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.catalog-page__pagination-nav :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.catalog-page__pagination-nav
  :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.catalog-page__pagination-nav :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.catalog-page__pagination-nav
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.catalog-page__pagination-nav :deep(.base-button:disabled) {
  opacity: 0.4;
}

.catalog-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

/* Los tres diálogos de esta pantalla (issue 2026-09-28, alta: "se ve como
   una simple ventana que aparece de la nada, sorpréndeme", luego "sigue
   blanco, no respetas los colores del fondo... ajusta los inputs como los
   de Agenda"; edición/ciclo de vida: "aplica eso mismo para las pantallas
   emergentes de editar y eliminar"): ninguno usa ya la superficie clara
   por defecto de BaseDialog — los tres usan el mismo canvas de tinta que
   el resto de esta pantalla y el mismo tratamiento de campos "reglados
   sobre tinta" que NewAppointmentPage.vue (Agenda) ya estableció para sus
   propios formularios, en vez de flotar como una tarjeta blanca genérica.
   Nada de esto toca BaseDialog.vue ni BaseInput.vue para el resto de la
   app: solo aplica dentro de .catalog-page__create-dialog/__edit-dialog/
   __lifecycle-dialog y sus respectivas clases de formulario/contenedor
   (content-class y clase del <form>/<div>, ver plantilla).

   El fondo, el filete del encabezado, el título, la descripción y el botón
   de cerrar (elementos propios de BaseDialog.vue, no de esta plantilla) se
   sobrescriben en el bloque <style> SIN "scoped" al final del archivo, no
   aquí: BaseDialog renderiza su contenido en <Teleport to="body">, y Vue
   compila TODO `:deep()` anteponiendo el atributo de scope de ESTE
   componente como ancestro (`[data-v-hash] .foo`, documentado así por
   Vue), incluso cuando no se escribe un selector delante — ese ancestro ya
   no existe una vez teletransportado. Verificado en el navegador: esas
   reglas nunca se aplicaban aunque compilaban sin error (issue 2026-09-28,
   "sigue blanco"). Lo que SÍ sigue funcionando aquí abajo son los
   selectores anclados en clases que esta plantilla pone directamente en su
   propio contenido (.catalog-page__create-icon, .catalog-page__create-form
   :deep(...), y sus equivalentes __edit-* / __lifecycle-*): viajan con el
   elemento sin importar dónde lo monte Vue. */

/* Ficha del encabezado, los tres diálogos (issue 2026-09-28, "aplica eso
   mismo para las pantallas emergentes de editar y eliminar"): mismo par
   dorado/tinta que ya usa el botón "Agregar servicio" (fondo
   --color-brand-accent-surface, texto --color-brand-accent-text). Alta
   muestra "+"; Editar y Desactivar/Reactivar muestran serviceInitial() del
   servicio destino — la misma inicial que ya lo identifica en la fila —
   para que la ficha señale DE QUÉ servicio se trata, no solo la acción. */
.catalog-page__create-icon,
.catalog-page__edit-icon,
.catalog-page__detail-icon,
.catalog-page__lifecycle-icon {
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
  /* Entra un instante después de la tarjeta (mitad de --motion-duration-
     fast) para que se sienta como un segundo tiempo del mismo gesto, no
     dos animaciones independientes arrancando a la vez. */
  animation: catalog-dialog-icon-pop var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1)
    60ms both;
}

/* "Desactivar" tiñe su propia ficha con el mismo rojo ya calculado para AA
   sobre tinta (--color-danger-on-strong, el que ya usa el botón "Desactivar"
   de la fila) en vez del dorado neutro: advierte antes de leer una palabra.
   Mismo par fondo/texto invertido que el dorado (fondo claro, texto tinta),
   ningún color nuevo. */
.catalog-page__lifecycle-icon--danger {
  background-color: var(--color-danger-on-strong);
  color: var(--color-surface-strong);
}

/* Destello: un solo barrido diagonal, una vez, justo cuando la ficha
   termina de asentarse. En tinta el glint de DiamondLoader es dorado
   sobre un fondo transparente; aquí la ficha YA es dorada (o roja), así
   que el barrido usa el mismo triplete RGB de --color-on-strong que
   tokens.css ya reutiliza para sus propios veladores translúcidos
   (--color-field-strong-border es "color-mix(in srgb, var(--color-on-strong) 12%, transparent)") en vez de más
   color sobre color, que no se vería. UNA sola pasada, no en bucle: es el
   remate de abrir el diálogo, no un indicador de carga. */
.catalog-page__create-icon::after,
.catalog-page__edit-icon::after,
.catalog-page__detail-icon::after,
.catalog-page__lifecycle-icon::after {
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
  animation: catalog-dialog-icon-glint 480ms cubic-bezier(0.5, 0, 0.3, 1) 260ms both;
}

@keyframes catalog-dialog-icon-pop {
  from {
    opacity: 0;
    transform: scale(0.5) rotate(-20deg);
  }

  to {
    opacity: 1;
    transform: scale(1) rotate(0deg);
  }
}

@keyframes catalog-dialog-icon-glint {
  from {
    transform: translateX(-100%);
  }

  to {
    transform: translateX(100%);
  }
}

/* El rebote de apertura (transition-timing-function de .base-dialog--open)
   se desactiva junto con el resto en el bloque sin scope, al final del
   archivo — vive ahí por el mismo motivo de Teleport explicado arriba. */
@media (prefers-reduced-motion: reduce) {
  .catalog-page__create-icon,
  .catalog-page__create-icon::after,
  .catalog-page__edit-icon,
  .catalog-page__edit-icon::after,
  .catalog-page__detail-icon,
  .catalog-page__detail-icon::after,
  .catalog-page__lifecycle-icon,
  .catalog-page__lifecycle-icon::after {
    animation: none;
  }
}

/* Campos reglados sobre tinta, alta Y edición (issue 2026-09-28, "ajusta
   los inputs como están los inputs de la pantalla Agenda", luego "aplica
   eso mismo para... editar"): mismas variables que NewAppointmentPage.vue
   ya redefine para sus propios BaseInput sobre --color-surface-strong
   (docs/03-desarrollo/estandar-diseno-visual.md: shared/ui sigue calibrado
   para superficie clara; cada pantalla sobre tinta redefine estos tokens
   en su propio ámbito, no en BaseInput.vue). */
.catalog-page__create-form :deep(.base-input),
.catalog-page__edit-form :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  color: var(--color-on-strong);
}

/* Rótulo y asterisco en el dorado de marca, no en el latón oscuro
   (--color-accent-brass) ni en el rojo por defecto de BaseInput: ambos
   están calibrados para AA sobre superficie CLARA (tokens.css) y aquí
   apenas rozan 2.7:1 sobre tinta. --color-brand-accent-surface sí resuelve
   AA sobre --color-surface-strong (ya verificado para el filete del
   encabezado, arriba) y es el mismo dorado que el resto del diálogo. */
.catalog-page__create-form :deep(.base-input__label),
.catalog-page__edit-form :deep(.base-input__label) {
  color: var(--color-brand-accent-surface);
}

.catalog-page__create-form :deep(.base-input__required),
.catalog-page__edit-form :deep(.base-input__required) {
  margin-left: 2px;
  color: var(--color-brand-accent-surface);
}

.catalog-page__create-form :deep(.base-input__hint),
.catalog-page__edit-form :deep(.base-input__hint) {
  color: var(--color-on-strong-muted);
}

.catalog-page__create-form :deep(.base-input::placeholder),
.catalog-page__edit-form :deep(.base-input::placeholder) {
  color: var(--color-on-strong-muted);
  opacity: 1;
}

.catalog-page__create-form :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)),
.catalog-page__edit-form :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

/* El dorado del filete inferior marca el campo YA RESUELTO (mismo criterio
   que NewAppointmentPage.vue: ocho filetes dorados a la vez no distinguen
   lo hecho de lo pendiente); la clase la pone :class en cada BaseInput
   (ver plantilla) según si su propio v-model tiene contenido. Alta y
   edición usan cada una su propia clase --filled (no comparten estado). */
.catalog-page__create-form :deep(.catalog-page__create-input--filled .base-input),
.catalog-page__edit-form :deep(.catalog-page__edit-input--filled .base-input) {
  background-color: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.catalog-page__create-form :deep(.base-input:disabled),
.catalog-page__create-form :deep(.base-input--disabled),
.catalog-page__edit-form :deep(.base-input:disabled),
.catalog-page__edit-form :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.catalog-page__create-form :deep(.base-input--invalid),
.catalog-page__edit-form :deep(.base-input--invalid) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.catalog-page__create-form :deep(.base-input__error),
.catalog-page__edit-form :deep(.base-input__error) {
  color: var(--color-danger-on-strong);
}

/* Botones del pie, los tres diálogos: mismo par dorado-sólido/tinta-fantasma
   que ya usan el CTA "Agregar servicio" y las acciones de fila de esta
   misma pantalla sobre tinta — no los valores por defecto de BaseButton
   (calibrados para flotar sobre superficie clara, donde --primary es tinta
   sobre tinta y --secondary es un bloque blanco: ambos se romperían aquí).
   "Desactivar" (variant="danger" en la plantilla) es la única excepción:
   conserva su relleno rojo sólido normal — es la acción destructiva del
   diálogo de ciclo de vida, y aquí SÍ debe seguir leyendo como tal (ver
   comentario en la sección de item-actions, arriba, sobre el mismo
   criterio para la fila). */
.catalog-page__create-form :deep(.base-button),
.catalog-page__edit-form :deep(.base-button),
.catalog-page__lifecycle :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.catalog-page__create-form :deep(.base-button--primary),
.catalog-page__edit-form :deep(.base-button--primary),
.catalog-page__lifecycle :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.catalog-page__create-form
  :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)),
.catalog-page__edit-form
  :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)),
.catalog-page__lifecycle
  :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.catalog-page__create-form
  :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)),
.catalog-page__edit-form
  :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)),
.catalog-page__lifecycle
  :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  filter: brightness(84%);
}

.catalog-page__create-form :deep(.base-button--secondary),
.catalog-page__edit-form :deep(.base-button--secondary),
.catalog-page__lifecycle :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.catalog-page__create-form
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)),
.catalog-page__edit-form
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)),
.catalog-page__lifecycle
  :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
}

.catalog-page__create-form
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)),
.catalog-page__edit-form
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)),
.catalog-page__lifecycle
  :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

/* "Desactivar" (variant="danger"): ghost rojo, mismo tratamiento que ya usa
   la acción "Desactivar" de la fila (arriba) — sin relleno, filete inferior
   acentuado — para que la confirmación destructiva se lea consistente con
   el botón que la disparó, en vez del relleno rojo sólido genérico de
   BaseButton (pensado para superficie clara). */
.catalog-page__lifecycle :deep(.base-button--danger) {
  background-color: transparent;
  color: var(--color-danger-on-strong);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 50%, transparent);
  border-bottom-color: var(--color-danger-on-strong);
}

.catalog-page__lifecycle
  :deep(.base-button--danger:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 50%, transparent);
  filter: none;
}

.catalog-page__lifecycle
  :deep(.base-button--danger:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 20%, transparent);
  filter: none;
}

.catalog-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-5);
  flex-wrap: wrap;
}

/* color (issue detectado al verificar en vivo, 2026-09-28): sin esto, el
   párrafo hereda el texto casi negro por defecto de BaseDialog.vue
   (calibrado para su tarjeta blanca), invisible sobre el nuevo fondo de
   tinta de este diálogo. --color-on-strong es el mismo tono que ya usa el
   título del diálogo (bloque sin scope, al final del archivo). */
.catalog-page__lifecycle {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  color: var(--color-on-strong);
}

.catalog-page__lifecycle-lead {
  margin: 0;
  font-size: var(--font-size-body);
}

/* Impacto real (CA-024-01): antes era un segundo párrafo suelto, idéntico
   en peso visual a la frase anterior (issue reportado 2026-09-28, "el
   mensaje... lo veo muy sencillo") — el dato que de verdad importa para
   decidir (¿hay citas futuras?) se perdía entre dos oraciones largas.
   Ahora vive en su propia ficha, mismo lenguaje "hundido sobre tinta" que
   ya usa .catalog-page__list (--color-field-strong/-border), con un
   ícono de estado que resume el resultado antes de leer una palabra. */
.catalog-page__lifecycle-impact {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
}

.catalog-page__lifecycle-impact-icon {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
  border-radius: 50%;
  color: var(--color-on-strong-muted);
  font-family: var(--font-display);
  font-size: var(--font-size-body-sm);
  line-height: 1;
}

/* Sin citas futuras: "✓" en el tinte muted por defecto, la ficha no
   necesita más énfasis que el resto del diálogo — es la respuesta
   tranquilizadora. Con citas futuras: la ficha se tiñe del mismo rojo que
   ya usa el resto de este diálogo para "Desactivar" (ficha del encabezado,
   filete, botón), así el único dato realmente accionable destaca sin
   añadir un color nuevo. */
.catalog-page__lifecycle-impact--warning {
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 40%, transparent);
}

.catalog-page__lifecycle-impact--warning .catalog-page__lifecycle-impact-icon {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 16%, transparent);
  color: var(--color-danger-on-strong);
}

.catalog-page__lifecycle-impact-text {
  margin: 0;
  font-size: var(--font-size-body-sm);
}

.catalog-page__lifecycle-note {
  margin: 0;
  font-size: var(--font-size-caption);
  color: var(--color-on-strong-muted);
}

/* Diálogo "Estado del servicio", desactivar Y reactivar (issue 2026-09-29,
   "lo veo feo y sencillo", "deja ese... ya que hace ambos" y "solo deja ese
   interruptor"): un interruptor real que es la acción, la ficha del servicio
   y, al reactivar, tres garantías. Sin animaciones de entrada (issue
   2026-09-29, "al abrir esto se demora"): todo está a la vista desde el
   primer cuadro; solo el propio interruptor se desliza al tocarlo. Solo
   colores/espaciados ya existentes en tokens.css; mismo lenguaje "hundido
   sobre tinta" que __lifecycle-impact. */
.catalog-page__lifecycle-switch {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-4);
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
  color: var(--color-on-strong);
  font: inherit;
  font-size: var(--font-size-caption);
  letter-spacing: 0.08em;
  text-transform: uppercase;
  cursor: pointer;
}

.catalog-page__lifecycle-switch:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

.catalog-page__lifecycle-switch:disabled {
  cursor: progress;
}

.catalog-page__lifecycle-state {
  opacity: 1;
  transition: opacity var(--motion-duration-fast) var(--motion-easing-standard);
}

/* Sin interruptor activo, "Activo" queda apagado; con él, "Inactivo". */
.catalog-page__lifecycle-switch:not(.catalog-page__lifecycle-switch--on)
  .catalog-page__lifecycle-state--on,
.catalog-page__lifecycle-switch--on
  .catalog-page__lifecycle-state:not(.catalog-page__lifecycle-state--on) {
  opacity: 0.4;
}

.catalog-page__lifecycle-switch--on .catalog-page__lifecycle-state--on {
  color: var(--color-success-on-strong);
}

.catalog-page__lifecycle-track {
  position: relative;
  flex: 0 0 auto;
  width: 64px;
  height: 32px;
  background-color: color-mix(in srgb, var(--color-on-strong) 10%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-pill);
  transition: background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.catalog-page__lifecycle-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 24px;
  height: 24px;
  background-color: var(--color-on-strong-muted);
  border-radius: 50%;
  transition:
    transform var(--motion-duration-base) cubic-bezier(0.34, 1.56, 0.64, 1),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.catalog-page__lifecycle-switch--on .catalog-page__lifecycle-track {
  background-color: color-mix(in srgb, var(--color-success-on-strong) 22%, transparent);
}

.catalog-page__lifecycle-switch--on .catalog-page__lifecycle-thumb {
  transform: translateX(32px);
  background-color: var(--color-success-on-strong);
}

.catalog-page__lifecycle-hint {
  margin: calc(var(--space-2) * -1) 0 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  text-align: center;
}

.catalog-page__lifecycle-facts {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: minmax(0, 1fr);
  gap: var(--space-2);
  margin: 0;
}

.catalog-page__lifecycle-facts > div {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-3) var(--space-2);
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
}

.catalog-page__lifecycle-facts dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.catalog-page__lifecycle-facts dd {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-body);
  line-height: 1.2;
  overflow-wrap: anywhere;
}

.catalog-page__lifecycle-facts dd small {
  display: block;
  margin-top: var(--space-1);
  color: var(--color-on-strong-muted);
  font-family: inherit;
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
}

.catalog-page__lifecycle-checks {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: var(--font-size-body-sm);
}

.catalog-page__lifecycle-checks li {
  position: relative;
  padding-left: var(--space-6);
}

.catalog-page__lifecycle-checks li::before {
  content: '✓';
  position: absolute;
  top: 1px;
  left: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  background-color: color-mix(in srgb, var(--color-success-on-strong) 16%, transparent);
  border-radius: 50%;
  color: var(--color-success-on-strong);
  font-size: var(--font-size-caption);
  line-height: 1;
}

@media (prefers-reduced-motion: reduce) {
  .catalog-page__lifecycle-state,
  .catalog-page__lifecycle-track,
  .catalog-page__lifecycle-thumb {
    transition: none;
  }
}

/* Detalle del servicio (issue 2026-09-29). Reutiliza las fichas
   .catalog-page__lifecycle-facts (duración/precio); lo propio es el estado, la
   descripción y los datos de registro. Todo entra escalonado con --row-index
   (mismo recurso que las filas de la lista) un instante después de la tarjeta,
   para que el diálogo se arme en vez de aparecer de golpe. */
.catalog-page__detail-status,
.catalog-page__detail-facts > div,
.catalog-page__detail-section,
.catalog-page__detail-meta {
  animation: catalog-detail-enter 380ms var(--motion-easing-standard) both;
  animation-delay: calc(120ms + var(--row-index, 0) * 70ms);
}

.catalog-page__detail-status {
  display: inline-flex;
  align-items: center;
  align-self: flex-start;
  gap: 8px;
  margin: 0;
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.catalog-page__detail-status--active {
  color: var(--color-success-on-strong);
}

/* Mismo rombo que la insignia de la fila, con el ritmo propio del servicio
   (diamondStyle), para que el detalle lea como continuación de la fila. */
.catalog-page__detail-diamond {
  flex-shrink: 0;
  width: 7px;
  height: 7px;
  transform: rotate(45deg);
  background-color: currentColor;
  animation: catalog-diamond-blink var(--diamond-duration, 7s) ease-in-out var(--diamond-delay, 0s)
    infinite;
}

.catalog-page__detail-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-4);
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  /* Filete dorado a la izquierda, como cita: marca el texto largo como el
     contenido principal del detalle sin recurrir a otra caja de color. */
  border-left: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: var(--radius-sm);
}

.catalog-page__detail-label {
  margin: 0;
  color: var(--color-brand-accent-surface);
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

/* pre-wrap respeta los saltos de línea que la persona escribió; anywhere evita
   que una palabra sin espacios (una URL, 500 caracteres seguidos) ensanche el
   diálogo. */
/* Las cifras son el dato que se busca al abrir el detalle: un escalón sobre el
   cuerpo de las fichas de estado del diálogo de ciclo de vida. */
.catalog-page__detail-facts dd {
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
}

.catalog-page__detail-text {
  margin: 0;
  color: var(--color-on-strong);
  font-size: var(--font-size-body);
  line-height: 1.6;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.catalog-page__detail-text--empty {
  color: var(--color-on-strong-muted);
  font-style: italic;
}

.catalog-page__detail-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2) var(--space-5);
  margin: 0;
}

.catalog-page__detail-meta dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.catalog-page__detail-meta dd {
  margin: 2px 0 0;
  font-size: var(--font-size-body-sm);
  font-variant-numeric: tabular-nums;
}

.catalog-page__detail-footer {
  flex: 1;
  flex-flow: row wrap;
  justify-content: flex-end;
  gap: var(--space-3);
}

@keyframes catalog-detail-enter {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .catalog-page__detail-status,
  .catalog-page__detail-facts > div,
  .catalog-page__detail-section,
  .catalog-page__detail-meta,
  .catalog-page__detail-diamond,
  .catalog-page__item-chevron,
  .catalog-page__item-name,
  .catalog-page__row {
    animation: none;
    transition: none;
  }
}

/* Zona media 641–960px (issue reportado 2026-09-28, "los botones se
   desencajan todo raro"): entre el ancho de escritorio completo y el
   quiebre móvil de abajo, las columnas fijas (duración/precio/estado/
   acciones) ya no cabían junto al mínimo de 220px de la columna de
   nombre — el navegador forzaba ese mínimo y el resto de la fila se salía
   del ancho visible, recortando literalmente "Editar"/"Desactivar"
   (medido en vivo: a 850px ya no se veían, a 950px sí). Se acortan las
   columnas y los botones (más angostos, sin apilar: apilarlos exigiría
   una fila más alta que el resto de este quiebre, y el alto de fila es
   FIJO, ver ROW_HEIGHT_PX en el script) hasta que quepan de sobra junto al
   mínimo de la columna de nombre, verificado en vivo en todo el rango
   641–960px. */
@media (max-width: 960px) and (min-width: 641px) {
  .catalog-page {
    padding-inline: 20px;
  }

  .catalog-page__columns,
  .catalog-page__row {
    grid-template-columns: minmax(100px, 1fr) 64px 100px 88px 156px;
    gap: 8px;
  }

  .catalog-page__table--offer .catalog-page__columns,
  .catalog-page__table--offer .catalog-page__row {
    grid-template-columns: minmax(70px, 1fr) 56px 82px 70px 60px 140px;
  }

  .catalog-page__columns {
    letter-spacing: 0.06em;
  }

  .catalog-page__item-meta {
    font-size: var(--font-size-body-sm);
  }

  .catalog-page__item-actions {
    gap: 4px;
  }

  .catalog-page__item-actions :deep(.base-button) {
    height: 32px;
    padding-inline: 6px;
    font-size: var(--font-size-caption);
  }
}

@media (max-width: 640px) {
  .catalog-page {
    gap: 12px;
    padding: 16px 16px 28px;
  }

  .catalog-page__header {
    padding-bottom: 10px;
  }

  .catalog-page__search {
    max-width: none;
  }

  .catalog-page__pagination-bar {
    justify-content: center;
    text-align: center;
  }

  .catalog-page__pagination-nav {
    flex-wrap: wrap;
    justify-content: center;
  }

  .catalog-page__create.base-button {
    width: var(--control-height-icon);
    height: var(--control-height-icon);
    padding: 0;
    overflow: hidden;
    font-size: 0;
  }

  .catalog-page__create :deep(.base-button__content) {
    font-size: 0;
  }

  .catalog-page__create :deep(.base-button__content)::before {
    margin: 0;
    font-size: var(--font-size-title-item);
  }

  .catalog-page__columns {
    display: none;
  }

  /* 100px, no 62px (issue detectado al verificar en vivo, 2026-09-28,
     "que se adapte según la pantalla, que no haya scroll"): en móvil la
     fila apila DOS filas internas de grid (nombre+descripción arriba,
     duración+acciones abajo) — 62px alcanzaba de sobra cuando la
     descripción envolvía a varias líneas (min-height se estiraba), pero
     con alto FIJO (obligatorio para poder predecir cuántas filas caben,
     ver ROW_HEIGHT_MOBILE_PX en el script) recortaba el botón
     "Desactivar" por abajo. 100px es lo medido en vivo que ambas filas
     internas necesitan sin recortarse. */
  .catalog-page__row {
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 4px 10px;
    height: 100px;
    padding: 9px 10px;
  }

  .catalog-page__item-heading {
    gap: 9px;
  }

  .catalog-page__item-icon {
    width: 28px;
    height: 28px;
    font-size: var(--font-size-body-sm);
  }

  .catalog-page__item-name {
    font-size: var(--font-size-body-sm);
  }

  .catalog-page__item-duration,
  .catalog-page__item-price {
    grid-column: 1;
    /* 28px de ficha + 9px de espacio del heading (arriba): alinea bajo el
       nombre en vez de bajo la ficha. */
    margin-left: 37px;
    font-size: var(--font-size-caption);
  }

  .catalog-page__item-price {
    display: none;
  }

  /* Antes ambos compartían "1 / span 2" (misma celda que "Editar"/
     "Desactivar"): con dos elementos independientes centrados en la MISMA
     área, el texto de la insignia y los botones quedaban superpuestos
     (encontrado 2026-09-28 al revisar el nuevo alto de fila con
     descripción). Cada uno a su propia fila —insignia junto al nombre,
     acciones junto a la duración— evita el choque sin tocar el marcado. */
  .catalog-page__row > :deep(.base-badge) {
    grid-column: 2;
    grid-row: 1;
    justify-self: end;
  }

  .catalog-page__item-actions {
    grid-column: 2;
    grid-row: 2;
    gap: 6px;
  }

  /* La regla de seis columnas de «Lo ofrezco» pesa más que la de móvil de arriba: se
     devuelve aquí la cuadrícula de dos columnas de siempre. */
  .catalog-page__table--offer .catalog-page__row {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  /* «Lo ofrezco» (DEC-115): junto a la duración, en la fila de abajo, alineado a
     la derecha de su celda para no pisar la duración. */
  .catalog-page__item-offer {
    grid-column: 1;
    grid-row: 2;
    justify-self: end;
  }

  /* El interruptor ocupa su celda de forma explícita, así que la duración (que se
     colocaba sola en esa celda) pasaría a una fila implícita y se recortaría: se le
     da la misma fila. */
  .catalog-page__table--offer .catalog-page__item-duration {
    grid-row: 2;
  }

  .catalog-page__item-actions :deep(.base-button) {
    height: 34px;
    padding-inline: 10px;
    font-size: var(--font-size-caption);
  }
}
</style>

<style>
/* SIN "scoped" a propósito (issue 2026-09-28, "sigue blanco y no respetas
   los colores del fondo"; luego "aplica eso mismo para las pantallas
   emergentes de editar y eliminar"): BaseDialog.vue renderiza sus tres
   diálogos en <Teleport to="body">, así que su tarjeta, encabezado,
   título, descripción y botón de cerrar dejan de ser descendientes de
   .catalog-page en el DOM real. Un :deep() con alcance de componente
   (arriba, <style scoped>) sigue anteponiendo el atributo de scope de
   CatalogPage.vue como ancestro exigido (así lo compila Vue, con o sin un
   selector propio delante), y ese ancestro ya no existe tras el
   teletransporte — por eso esas reglas nunca llegaban a aplicarse pese a
   compilar sin error. Este bloque selecciona por nombre de clase real, sin
   atributo de scope, así que no depende de dónde Vue monte el nodo.
   .catalog-page__create-dialog / __edit-dialog / __lifecycle-dialog son
   nombres exclusivos de esta pantalla, no genéricos, así que el riesgo de
   colisión con otra hoja es nulo. El !important es necesario porque la
   regla propia de BaseDialog.vue tiene la MISMA especificidad (una clase +
   un atributo de scope) y el orden en que Vite inserta el <style> de cada
   componente en el documento final no está garantizado — sin !important,
   quién gana dependería de ese orden, no de cuál regla describe este
   diálogo en particular. Contenido a esta única pantalla (nunca a
   BaseDialog.vue ni a otro diálogo de la app). */
.catalog-page__create-dialog.base-dialog,
.catalog-page__edit-dialog.base-dialog,
.catalog-page__detail-dialog.base-dialog,
.catalog-page__lifecycle-dialog.base-dialog {
  background-color: var(--color-surface-strong) !important;
  border: var(--border-width-normal) solid var(--color-field-strong-border) !important;
}

/* Mismo cubic-bezier con rebote que ya usa BarberSelect.vue para su marca
   de opción elegida (0.34, 1.56, 0.64, 1): no es un valor nuevo, es el que
   este código base ya usa para "un elemento pequeño que llega y se
   asienta". */
.catalog-page__create-dialog.base-dialog--open,
.catalog-page__edit-dialog.base-dialog--open,
.catalog-page__detail-dialog.base-dialog--open,
.catalog-page__lifecycle-dialog.base-dialog--open {
  transition-timing-function: cubic-bezier(0.34, 1.56, 0.64, 1) !important;
}

@media (prefers-reduced-motion: reduce) {
  .catalog-page__create-dialog.base-dialog--open,
  .catalog-page__edit-dialog.base-dialog--open,
  .catalog-page__detail-dialog.base-dialog--open,
  .catalog-page__lifecycle-dialog.base-dialog--open {
    transition-timing-function: var(--motion-easing-standard) !important;
  }
}

/* Filete dorado (--border-width-emphasis, el mismo grosor que ya marca la
   fecha activa de AgendaDatePicker/DailyAgendaPage) en vez del filete gris
   por defecto de BaseDialog: --color-brand-accent-surface es correcto
   aquí porque el encabezado ahora vive sobre tinta. */
.catalog-page__create-dialog .base-dialog__header,
.catalog-page__edit-dialog .base-dialog__header,
.catalog-page__detail-dialog .base-dialog__header,
.catalog-page__lifecycle-dialog .base-dialog__header {
  border-bottom-width: var(--border-width-emphasis) !important;
  border-bottom-color: var(--color-brand-accent-surface) !important;
}

/* "Desactivar" tiñe el filete de rojo (mismo tinte que ya tiñe su ficha,
   arriba): la clase modificadora --danger viaja junto con
   .catalog-page__lifecycle-dialog en el mismo content-class (ver
   plantilla), así que ambas quedan en el elemento raíz del diálogo. */
.catalog-page__lifecycle-dialog--danger .base-dialog__header {
  border-bottom-color: var(--color-danger-on-strong) !important;
}

.catalog-page__create-dialog .base-dialog__title,
.catalog-page__edit-dialog .base-dialog__title,
.catalog-page__detail-dialog .base-dialog__title,
.catalog-page__lifecycle-dialog .base-dialog__title {
  color: var(--color-on-strong) !important;
}

.catalog-page__detail-dialog .base-dialog__title {
  overflow-wrap: anywhere;
}

.catalog-page__detail-dialog .base-dialog__footer {
  border-top-color: var(--color-field-strong-border) !important;
}

.catalog-page__create-dialog .base-dialog__description,
.catalog-page__edit-dialog .base-dialog__description,
.catalog-page__detail-dialog .base-dialog__description,
.catalog-page__lifecycle-dialog .base-dialog__description {
  color: var(--color-on-strong-muted) !important;
}

.catalog-page__create-dialog .base-dialog__close,
.catalog-page__edit-dialog .base-dialog__close,
.catalog-page__detail-dialog .base-dialog__close,
.catalog-page__lifecycle-dialog .base-dialog__close {
  color: var(--color-on-strong-muted) !important;
}

.catalog-page__create-dialog .base-dialog__close:hover,
.catalog-page__edit-dialog .base-dialog__close:hover,
.catalog-page__detail-dialog .base-dialog__close:hover,
.catalog-page__lifecycle-dialog .base-dialog__close:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent) !important;
  color: var(--color-on-strong) !important;
}

.catalog-page__create-dialog .base-dialog__close:focus-visible,
.catalog-page__edit-dialog .base-dialog__close:focus-visible,
.catalog-page__detail-dialog .base-dialog__close:focus-visible,
.catalog-page__lifecycle-dialog .base-dialog__close:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus) !important;
}
</style>
