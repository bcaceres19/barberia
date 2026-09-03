<script setup lang="ts">
// Pantalla "Servicios" (HU-022, HU-024): consultar y listar el catálogo de
// la barbería activa, crear un servicio, editarlo y cambiar su ciclo de
// vida (desactivar/reactivar). Estados discriminados: carga inicial, listo,
// vacío, error recuperable, guardando (mismo patrón que StaffPage.vue). Un
// error recuperable NUNCA borra lo que el barbero ya escribió; solo un
// guardado exitoso confirmado por el servidor cierra el diálogo. Nunca
// muestra asignaciones a barberos (HU-023), disponibilidad ni citas: fuera
// de alcance de esta historia.
import { onMounted, ref } from 'vue'
import {
  BaseAlert,
  BaseBadge,
  BaseButton,
  BaseDialog,
  BaseInput,
  PageHeader,
  RecordRow,
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
const nextCursor = ref<string | null>(null)
const loadingMore = ref(false)

// requestToken evita que una carga inicial obsoleta sobreescriba la lista
// con datos viejos si el barbero recarga la sección antes de que la
// primera respuesta llegue (mismo patrón que StaffPage.vue).
let requestToken = 0

async function load() {
  const token = ++requestToken
  loadStatus.value = 'loading'
  const outcome = await fetchServices()
  if (token !== requestToken) return

  if (outcome.kind === 'success') {
    services.value = outcome.page.items
    nextCursor.value = outcome.page.nextCursor
    loadStatus.value = 'ready'
    return
  }
  loadStatus.value = 'load-error'
}

onMounted(load)

function onRetryLoad() {
  void load()
}

async function onLoadMore() {
  if (loadingMore.value || !nextCursor.value) return
  loadingMore.value = true
  const outcome = await fetchServices(nextCursor.value)
  loadingMore.value = false

  if (outcome.kind === 'success') {
    // Concatena sin duplicar: el cursor de una página nunca repite un id ya
    // visto (backend, CA-022-01); esto solo evita un doble clic muy rápido
    // en "Cargar más" desde volver a insertar la misma página.
    const knownIDs = new Set(services.value.map((s) => s.id))
    for (const item of outcome.page.items) {
      if (!knownIDs.has(item.id)) services.value.push(item)
    }
    nextCursor.value = outcome.page.nextCursor
  }
}

// No concatena "$" (estandar-diseno-visual.md §10.3: "Dinero se formatea
// según la moneda definida por el contrato; no se concatena '$' ni se
// asumen pesos desde el componente"). service.price es un string decimal
// exacto que nunca se convierte a number en el cliente (docs/06-api/
// estandar-openapi.md §6); esta función no reformatea sus dígitos, solo
// añade la moneda real recibida del contrato.
function formatPrice(service: Service): string {
  return `${service.price} ${service.currency}`
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
  impactStatus.value = 'loading'
  affectedAppointments.value = 0
  lifecycleIdempotencyKey = newIdempotencyKey()
  isLifecycleOpen.value = true

  // CA-024-01: la advertencia solo muestra impacto respaldado por el
  // servidor, nunca un valor por defecto asumido en el cliente.
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

function openReactivateDialog(service: Service) {
  lifecycleTarget.value = service
  lifecycleIntent.value = 'reactivate'
  lifecycleStatus.value = 'idle'
  impactStatus.value = 'ready'
  lifecycleIdempotencyKey = newIdempotencyKey()
  isLifecycleOpen.value = true
}

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
      isLifecycleOpen.value = false
      lifecycleStatus.value = 'idle'
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
      isLifecycleOpen.value = false
      lifecycleStatus.value = 'idle'
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
// el cliente esperaba.
async function reloadAfterConflict(serviceId: string) {
  const outcome = await fetchServices()
  if (outcome.kind !== 'success') return
  const fresh = outcome.page.items.find((s) => s.id === serviceId)
  if (fresh) replaceServiceInList(fresh)
}
</script>

<template>
  <section class="catalog-page" aria-labelledby="catalog-page-title">
    <PageHeader title-id="catalog-page-title" title="Servicios">
      <template v-if="loadStatus === 'ready'" #actions>
        <BaseButton type="button" variant="primary" @click="openCreateDialog">
          Agregar servicio
        </BaseButton>
      </template>
    </PageHeader>

    <div
      v-if="loadStatus === 'loading'"
      class="catalog-page__state"
      role="status"
      aria-live="polite"
    >
      <p>Cargando el catálogo…</p>
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

    <template v-else>
      <p v-if="services.length === 0" class="catalog-page__empty">
        Aún no tienes servicios registrados. Agrega el primero para empezar.
      </p>

      <ul v-else class="catalog-page__list" aria-label="Servicios de la barbería">
        <RecordRow v-for="service in services" :key="service.id">
          <div class="catalog-page__item-heading">
            <span class="catalog-page__item-name">{{ service.name }}</span>
            <!-- No depende solo del color: el texto de la etiqueta ya
                 distingue el estado (CA-024-08). -->
            <BaseBadge
              :variant="service.isActive ? 'success' : 'neutral'"
              size="sm"
              dot
              role="status"
            >
              {{ service.isActive ? 'Activo' : 'Inactivo' }}
            </BaseBadge>
          </div>
          <span class="catalog-page__item-meta">
            {{ service.durationMinutes }} min · {{ formatPrice(service) }}
          </span>
          <span v-if="service.description" class="catalog-page__item-description">
            {{ service.description }}
          </span>
          <template #trailing>
            <div class="catalog-page__item-actions">
              <BaseButton
                type="button"
                variant="secondary"
                :aria-label="`Editar ${service.name}`"
                @click="openEditDialog(service)"
              >
                Editar
              </BaseButton>
              <BaseButton
                v-if="service.isActive"
                type="button"
                variant="danger"
                :aria-label="`Desactivar ${service.name}`"
                @click="openDeactivateDialog(service)"
              >
                Desactivar
              </BaseButton>
              <BaseButton
                v-else
                type="button"
                variant="secondary"
                :aria-label="`Reactivar ${service.name}`"
                @click="openReactivateDialog(service)"
              >
                Reactivar
              </BaseButton>
            </div>
          </template>
        </RecordRow>
      </ul>

      <div v-if="nextCursor" class="catalog-page__load-more">
        <BaseButton
          type="button"
          variant="secondary"
          :loading="loadingMore"
          :disabled="loadingMore"
          @click="onLoadMore"
        >
          Cargar más
        </BaseButton>
      </div>
    </template>

    <!-- Alta -->
    <BaseDialog
      v-model="isCreateOpen"
      title="Agregar servicio"
      size="sm"
      @close="onCreateDialogClosed"
    >
      <form class="catalog-page__form" novalidate @submit.prevent="onSubmitCreate">
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
          @update:model-value="onCreateNameInput"
        />
        <BaseInput
          :model-value="createForm.description"
          name="description"
          label="Descripción (opcional)"
          :maxlength="500"
          :disabled="createStatus === 'saving'"
          :error="createFieldErrors.description"
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

    <!-- Edición -->
    <BaseDialog v-model="isEditOpen" title="Editar servicio" size="sm" @close="onEditDialogClosed">
      <form class="catalog-page__form" novalidate @submit.prevent="onSubmitEdit">
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
          @update:model-value="onEditNameInput"
        />
        <BaseInput
          :model-value="editForm.description"
          name="description"
          label="Descripción (opcional)"
          :maxlength="500"
          :disabled="editStatus === 'saving'"
          :error="editFieldErrors.description"
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

    <!-- HU-024: desactivar/reactivar -->
    <BaseDialog
      v-model="isLifecycleOpen"
      :title="lifecycleIntent === 'deactivate' ? 'Desactivar servicio' : 'Reactivar servicio'"
      size="sm"
      @close="onLifecycleDialogClosed"
    >
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

        <template v-if="lifecycleIntent === 'deactivate'">
          <p v-if="impactStatus === 'loading'" role="status" aria-live="polite">
            Consultando el impacto real…
          </p>
          <template v-else-if="impactStatus === 'ready'">
            <p>
              <strong>{{ lifecycleTarget?.name }}</strong> dejará de ofrecerse. El servicio y su
              historial se conservan; nunca se borra.
            </p>
            <p v-if="affectedAppointments > 0">
              {{ affectedAppointments }} cita(s) futura(s) se verían afectadas. Ninguna se cancela
              ni se modifica automáticamente.
            </p>
            <p v-else>No hay citas futuras que se vean afectadas ahora mismo.</p>
          </template>
        </template>
        <template v-else>
          <p>
            <strong>{{ lifecycleTarget?.name }}</strong> vuelve a ofrecerse en el catálogo, sin
            crear otro registro ni cambiar duración, precio o asignaciones.
          </p>
        </template>

        <div class="catalog-page__dialog-actions">
          <BaseButton type="button" variant="secondary" @click="isLifecycleOpen = false">
            Cancelar
          </BaseButton>
          <BaseButton
            type="button"
            :variant="lifecycleIntent === 'deactivate' ? 'danger' : 'primary'"
            :loading="lifecycleStatus === 'saving'"
            :disabled="lifecycleStatus === 'saving' || impactStatus === 'loading'"
            @click="onConfirmLifecycle"
          >
            {{ lifecycleIntent === 'deactivate' ? 'Desactivar' : 'Reactivar' }}
          </BaseButton>
        </div>
      </div>
    </BaseDialog>
  </section>
</template>

<style scoped>
.catalog-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 640px;
  padding: var(--space-4);
  margin: 0 auto;
}

.catalog-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.catalog-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.catalog-page__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.catalog-page__item-heading {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.catalog-page__item-name {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 500;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.catalog-page__item-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}

.catalog-page__item-meta {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
  /* Cifras tabulares: duración y precio (estandar-diseno-visual.md §5.2, §10.3). */
  font-variant-numeric: tabular-nums;
}

.catalog-page__item-description {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.catalog-page__load-more {
  display: flex;
  justify-content: center;
}

.catalog-page__form {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.catalog-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-5);
  flex-wrap: wrap;
}

.catalog-page__lifecycle {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
</style>
