<script setup lang="ts">
// Pantalla "Barberos" (HU-021): consultar y listar el equipo de la
// barbería activa, agregar un barbero y renombrarlo. Una barbería
// unipersonal y una de cuatro personas usan exactamente el mismo
// componente y el mismo estado de datos (CA-021-01/02): la lista con 1
// elemento y la lista con 4 no tienen ninguna rama especial. Estados
// discriminados: carga inicial, listo, vacío, error recuperable, guardando
// (trabajo requerido §4.3/§4.4). Un error recuperable NUNCA borra lo que el
// barbero ya escribió; solo un guardado exitoso confirmado por el servidor
// cierra el diálogo.
import { onMounted, ref } from 'vue'
import { BaseAlert, BaseButton, BaseDialog, BaseInput } from '@/shared/ui'
import { createBarber, fetchBarbers, renameBarber } from '../api/staffApi'
import { newIdempotencyKey } from '../model/idempotencyKey'
import type { Barber } from '../model/barber'
import { validateFullName } from '../validation/staffValidation'

type LoadStatus = 'loading' | 'ready' | 'load-error'
type SaveStatus =
  | 'idle'
  | 'saving'
  | 'validation-error'
  | 'idempotency-conflict'
  | 'not-found'
  | 'network-error'
  | 'unexpected-error'

const loadStatus = ref<LoadStatus>('loading')
const barbers = ref<Barber[]>([])
const nextCursor = ref<string | null>(null)
const loadingMore = ref(false)

// requestToken evita que una carga inicial obsoleta sobreescriba la lista
// con datos viejos si el barbero recarga la sección antes de que la
// primera respuesta llegue (mismo patrón que SettingsPage.vue).
let requestToken = 0

async function load() {
  const token = ++requestToken
  loadStatus.value = 'loading'
  const outcome = await fetchBarbers()
  if (token !== requestToken) return

  if (outcome.kind === 'success') {
    barbers.value = outcome.page.items
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

// Monograma accesible (especificacion-frontend-nava.md §4.4/§6): "Avatares
// reales solo si el producto incorpora una fuente y política para fotos.
// Hasta entonces se usa monograma accesible o ninguna imagen." Toma la
// primera letra del primer y del último término del nombre completo (o
// solo la primera si es un único término); decorativo, el nombre visible
// de la fila ya da el nombre accesible.
function initials(fullName: string): string {
  const parts = fullName.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return ''
  if (parts.length === 1) return parts[0]!.charAt(0).toUpperCase()
  return (parts[0]!.charAt(0) + parts[parts.length - 1]!.charAt(0)).toUpperCase()
}

async function onLoadMore() {
  if (loadingMore.value || !nextCursor.value) return
  loadingMore.value = true
  const outcome = await fetchBarbers(nextCursor.value)
  loadingMore.value = false

  if (outcome.kind === 'success') {
    // Concatena sin duplicar: el cursor de una página nunca repite un id
    // ya visto (backend, CA-021-02); esto solo evita un doble clic muy
    // rápido en "Cargar más" desde volver a insertar la misma página.
    const knownIDs = new Set(barbers.value.map((b) => b.id))
    for (const item of outcome.page.items) {
      if (!knownIDs.has(item.id)) barbers.value.push(item)
    }
    nextCursor.value = outcome.page.nextCursor
  }
}

// --- Alta -----------------------------------------------------------------

const isCreateOpen = ref(false)
const createFullName = ref('')
const createFieldError = ref<string | undefined>(undefined)
const createStatus = ref<SaveStatus>('idle')
const createAttempted = ref(false)
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
      barbers.value.unshift(outcome.barber)
      isCreateOpen.value = false
      createStatus.value = 'idle'
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

// --- Renombrado -------------------------------------------------------

const isRenameOpen = ref(false)
const renameTarget = ref<Barber | null>(null)
const renameFullName = ref('')
const renameFieldError = ref<string | undefined>(undefined)
const renameStatus = ref<SaveStatus>('idle')
const renameAttempted = ref(false)

function openRenameDialog(barber: Barber) {
  renameTarget.value = barber
  renameFullName.value = barber.fullName
  renameFieldError.value = undefined
  renameStatus.value = 'idle'
  renameAttempted.value = false
  isRenameOpen.value = true
}

function onRenameDialogClosed() {
  renameStatus.value = 'idle'
}

function onRenameFullNameInput(value: string | number) {
  renameFullName.value = String(value)
  if (renameAttempted.value) renameFieldError.value = validateFullName(renameFullName.value)
}

async function onSubmitRename() {
  if (renameStatus.value === 'saving' || !renameTarget.value) return

  renameAttempted.value = true
  const error = validateFullName(renameFullName.value)
  renameFieldError.value = error
  if (error) return

  renameStatus.value = 'saving'
  const outcome = await renameBarber(renameTarget.value.id, renameFullName.value.trim())

  switch (outcome.kind) {
    case 'success': {
      // Reemplaza por id sin duplicar ni reordenar de forma inestable
      // (trabajo requerido §4.5): solo cambia fullName/updatedAt del
      // elemento existente, en su misma posición.
      const index = barbers.value.findIndex((b) => b.id === outcome.barber.id)
      if (index !== -1) barbers.value[index] = outcome.barber
      isRenameOpen.value = false
      renameStatus.value = 'idle'
      return
    }
    case 'validation-error':
      renameStatus.value = 'validation-error'
      return
    case 'not-found':
      renameStatus.value = 'not-found'
      return
    case 'network-error':
      renameStatus.value = 'network-error'
      return
    case 'unexpected-error':
      renameStatus.value = 'unexpected-error'
  }
}
</script>

<template>
  <section class="staff-page" aria-labelledby="staff-page-title">
    <header class="staff-page__header">
      <h1 id="staff-page-title" class="staff-page__title">Barberos</h1>
      <BaseButton
        v-if="loadStatus === 'ready'"
        type="button"
        variant="primary"
        @click="openCreateDialog"
      >
        Agregar barbero
      </BaseButton>
    </header>

    <div v-if="loadStatus === 'loading'" class="staff-page__state" role="status" aria-live="polite">
      <p>Cargando el equipo…</p>
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

    <template v-else>
      <p v-if="barbers.length === 0" class="staff-page__empty">
        Aún no tienes barberos registrados. Agrega el primero para empezar.
      </p>

      <ul v-else class="staff-page__list" aria-label="Barberos de la barbería">
        <li v-for="barber in barbers" :key="barber.id" class="staff-page__item">
          <span class="staff-page__item-identity">
            <span class="staff-page__item-avatar" aria-hidden="true">{{
              initials(barber.fullName)
            }}</span>
            <span class="staff-page__item-name">{{ barber.fullName }}</span>
          </span>
          <BaseButton
            type="button"
            variant="secondary"
            :aria-label="`Editar ${barber.fullName}`"
            @click="openRenameDialog(barber)"
          >
            Editar
          </BaseButton>
        </li>
      </ul>

      <div v-if="nextCursor" class="staff-page__load-more">
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
      title="Agregar barbero"
      size="sm"
      @close="onCreateDialogClosed"
    >
      <form class="staff-page__create-form" novalidate @submit.prevent="onSubmitCreate">
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
          :model-value="createFullName"
          name="fullName"
          label="Nombre"
          required
          :maxlength="120"
          :disabled="createStatus === 'saving'"
          :error="createFieldError"
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
      title="Editar barbero"
      size="sm"
      @close="onRenameDialogClosed"
    >
      <form class="staff-page__rename-form" novalidate @submit.prevent="onSubmitRename">
        <BaseAlert
          v-if="renameStatus === 'not-found'"
          variant="warning"
          title="Este barbero ya no está disponible"
          role="alert"
        >
          Cierra este diálogo y recarga la lista.
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

        <BaseInput
          :model-value="renameFullName"
          name="fullName"
          label="Nombre"
          required
          :maxlength="120"
          :disabled="renameStatus === 'saving'"
          :error="renameFieldError"
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
.staff-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  max-width: 640px;
  padding: var(--space-4);
  margin: 0 auto;
}

.staff-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  flex-wrap: wrap;
}

.staff-page__title {
  margin: 0;
  font-size: var(--font-size-h1);
  line-height: var(--font-size-h1-line);
  font-weight: var(--font-weight-h1);
  color: var(--color-text-primary);
}

.staff-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.staff-page__empty {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.staff-page__list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  padding: 0;
  margin: 0;
  list-style: none;
}

.staff-page__item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  /* Fila 64-72px en escritorio, ficha 64px mínimo en móvil (§8.2, §7.2). */
  min-height: 64px;
  padding: var(--space-4);
  background-color: var(--color-surface);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  border-radius: var(--radius-md);
}

.staff-page__item-identity {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  min-width: 0;
}

/* Avatar operativo 40px (estandar-diseno-visual.md §6.2): monograma
   accesible, nunca un retrato ficticio. */
.staff-page__item-avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: var(--color-action-soft);
  color: var(--color-action-primary);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  font-weight: 600;
}

.staff-page__item-name {
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.staff-page__load-more {
  display: flex;
  justify-content: center;
}

.staff-page__dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-5);
  flex-wrap: wrap;
}
</style>
