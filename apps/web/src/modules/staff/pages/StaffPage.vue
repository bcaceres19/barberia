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
        class="staff-page__create-button"
        @click="openCreateDialog"
      >
        <svg class="staff-page__button-icon" viewBox="0 0 24 24" aria-hidden="true">
          <path d="M12 5v14M5 12h14" />
        </svg>
        <span class="staff-page__create-label">Agregar barbero</span>
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
      <div v-if="barbers.length === 0" class="staff-page__empty">
        <span class="staff-page__empty-icon" aria-hidden="true">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4">
            <circle cx="12" cy="8" r="4" />
            <path d="M4 21c0-4 3.6-7 8-7s8 3 8 7" />
          </svg>
        </span>
        <p>Aún no tienes <span>barberos registrados.</span></p>
        <BaseButton type="button" variant="primary" @click="openCreateDialog">
          <svg class="staff-page__button-icon" viewBox="0 0 24 24" aria-hidden="true">
            <path d="M12 5v14M5 12h14" />
          </svg>
          Agregar barbero
        </BaseButton>
      </div>

      <div v-else class="staff-page__records">
        <div class="staff-page__column-labels" aria-hidden="true">
          <span>NOMBRE</span>
          <span>ACCIÓN</span>
        </div>
        <ul class="staff-page__list" aria-label="Barberos de la barbería">
          <li v-for="barber in barbers" :key="barber.id" class="staff-page__item">
            <span class="staff-page__item-avatar" aria-hidden="true">{{
              initials(barber.fullName)
            }}</span>
            <span class="staff-page__item-name">{{ barber.fullName }}</span>
            <button
              type="button"
              class="staff-page__edit-button"
              :aria-label="`Editar ${barber.fullName}`"
              @click="openRenameDialog(barber)"
            >
              <svg
                class="staff-page__edit-icon"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                aria-hidden="true"
              >
                <path d="m4 20 4.1-1 10-10a2.1 2.1 0 0 0-3-3l-10 10L4 20Z" />
                <path d="m13.8 7.2 3 3" />
              </svg>
              <span class="staff-page__edit-label">Editar</span>
              <svg
                class="staff-page__chevron"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                aria-hidden="true"
              >
                <path d="m9 5 7 7-7 7" />
              </svg>
            </button>
          </li>
        </ul>
      </div>

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
      content-class="staff-page__dialog"
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
      content-class="staff-page__dialog"
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
  align-items: center;
  width: 100%;
  gap: var(--space-6);
  max-width: none;
  min-height: 100%;
  padding: 42px 48px 56px;
  background-color: var(--color-surface);
  box-sizing: border-box;
}

.staff-page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: min(100%, 520px);
  gap: var(--space-4);
}

.staff-page__title {
  margin: 0;
  color: var(--color-text-primary);
  font-family: var(--font-display);
  font-size: 32px;
  font-weight: 400;
  line-height: 38px;
}

.staff-page__create-button {
  min-width: 166px;
}

.staff-page__button-icon {
  width: 18px;
  height: 18px;
  fill: none;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
}

.staff-page__state {
  padding: var(--space-4);
  color: var(--color-text-secondary);
}

.staff-page__empty {
  display: flex;
  min-height: 300px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-5);
  border: var(--border-width-normal) solid var(--color-border-subtle);
  background: var(--color-surface);
  color: var(--color-text-primary);
  text-align: center;
}

.staff-page__empty p {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: var(--font-size-h2-line);
}

.staff-page__empty p span {
  display: block;
}

.staff-page__empty-icon {
  display: grid;
  width: 64px;
  height: 64px;
  place-items: center;
  border-radius: 50%;
  background: var(--color-canvas);
}

.staff-page__empty-icon svg {
  width: 32px;
  height: 32px;
}

.staff-page__records {
  width: min(100%, 520px);
}

.staff-page__column-labels {
  display: grid;
  grid-template-columns: 1fr auto;
  padding: 0 var(--space-2) var(--space-2);
  color: var(--color-text-secondary);
  font-family: var(--font-family-base);
  font-size: 11px;
  line-height: 16px;
}

.staff-page__list {
  display: flex;
  flex-direction: column;
  padding: 0;
  margin: 0;
  border: var(--border-width-normal) solid var(--color-border-subtle);
  background: var(--color-surface);
  list-style: none;
}

.staff-page__item {
  display: flex;
  min-height: 92px;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  border-bottom: var(--border-width-normal) solid var(--color-border-subtle);
}

.staff-page__item:last-child {
  border-bottom: none;
}

/* Avatar operativo 40px (estandar-diseno-visual.md §6.2): monograma
   accesible, nunca un retrato ficticio. */
.staff-page__item-avatar {
  display: flex;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background-color: var(--color-canvas);
  color: var(--color-text-primary);
  font-family: var(--font-display);
  font-size: 16px;
  font-weight: 400;
}

.staff-page__item-name {
  flex: 1;
  min-width: 0;
  font-family: var(--font-display);
  font-size: 18px;
  font-weight: 400;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.staff-page__edit-button {
  display: inline-flex;
  min-width: var(--control-height);
  min-height: var(--control-height);
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: 0 var(--space-2);
  border: 0;
  background: transparent;
  color: var(--color-text-primary);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body-sm);
  cursor: pointer;
}

.staff-page__edit-button:hover {
  color: var(--color-action-primary-hover);
  background: var(--color-canvas);
}

.staff-page__edit-button:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface),
    0 0 0 4px var(--color-focus);
}

.staff-page__edit-icon {
  width: 18px;
  height: 18px;
}

.staff-page__chevron {
  display: none;
  width: 18px;
  height: 18px;
}

.staff-page__load-more {
  display: flex;
  width: min(100%, 520px);
  justify-content: center;
}

.staff-page__dialog-actions {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  margin-top: var(--space-5);
  flex-wrap: wrap;
}

.staff-page :deep(.staff-page__dialog .base-dialog__header) {
  padding: 22px 24px;
}

.staff-page :deep(.staff-page__dialog .base-dialog__title) {
  font-family: var(--font-display);
  font-size: 22px;
  font-weight: 400;
}

@media (max-width: 1023px) {
  .staff-page {
    padding: 28px 24px 32px;
  }
}

@media (max-width: 480px) {
  .staff-page {
    gap: var(--space-5);
    align-items: stretch;
    padding: 16px 0 28px;
  }

  .staff-page__header {
    padding: 0 var(--space-4);
  }

  .staff-page__title {
    font-size: 18px;
    line-height: 28px;
  }

  .staff-page__create-button {
    width: var(--control-height);
    min-width: var(--control-height);
    padding: 0;
    border-radius: 50%;
  }

  .staff-page__create-label,
  .staff-page__column-labels,
  .staff-page__edit-label,
  .staff-page__edit-icon {
    display: none;
  }

  .staff-page__records {
    border-top: var(--border-width-normal) solid var(--color-border-subtle);
  }

  .staff-page__list {
    border-width: 0 0 var(--border-width-normal);
  }

  .staff-page__item {
    min-height: 64px;
    gap: var(--space-3);
    padding: 8px var(--space-4);
  }

  .staff-page__item-avatar {
    width: 36px;
    height: 36px;
    font-size: 12px;
  }

  .staff-page__item-name {
    font-family: var(--font-family-base);
    font-size: 13px;
    line-height: 18px;
  }

  .staff-page__edit-button {
    width: var(--control-height);
    padding: 0;
    color: var(--color-text-secondary);
  }

  .staff-page__chevron {
    display: block;
  }

  .staff-page__empty {
    min-height: 320px;
    border-left: 0;
    border-right: 0;
  }

  .staff-page__dialog-actions {
    flex-wrap: nowrap;
  }

  .staff-page__dialog-actions :deep(.base-button) {
    flex: 1;
  }
}
</style>
