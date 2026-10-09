<script setup lang="ts">
// Pantalla «Google Calendar» (issue #325, DEC-099, DEC-101, DEC-122): el barbero
// conecta, ve el estado, sincroniza ahora y desconecta SU Google Calendar, y
// elige la anticipación del recordatorio. La publicación es de NAVA hacia
// Google: lo que se cambie allí no altera la agenda de NAVA. La pantalla nunca
// maneja tokens: el flujo OAuth ocurre entre el navegador, Google y el API, y la
// vuelta la completa GoogleCalendarCallbackPage.
//
// Estados (derivados del recurso, no campos de la API): no disponible, falta
// vincular el barbero, no conectado, conectado, sincronizando (cambios en cola),
// requiere reconexión y error de sincronización. Las acciones se bloquean mientras
// una está en curso, así un doble clic nunca la repite.
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useToast, useVocabulary } from '@/shared/composables'
import { BaseAlert, BaseBadge, BaseButton, BaseDialog, BaseInput, DiamondLoader } from '@/shared/ui'
import {
  disconnect,
  fetchConnection,
  startConnect,
  syncNow,
  updateReminder,
} from '../api/googleCalendarApi'
import {
  viewOf,
  type GoogleCalendarConnection,
  type GoogleCalendarView,
} from '../model/googleCalendar'
import {
  MAX_REMINDER_MINUTES,
  MIN_REMINDER_MINUTES,
  validateReminderMinutes,
} from '../validation/reminder'

const toast = useToast()
const v = useVocabulary()

const LOADING_PHRASES = [
  'Revisando tu calendario',
  'Alineando los turnos',
  'Todo a su hora',
] as const
// Mientras hay cambios en cola se vuelve a consultar el estado hasta que se vacíe.
const POLL_MS = 5000

type LoadStatus = 'loading' | 'ready' | 'load-error'
type Busy = 'connect' | 'sync' | 'disconnect' | 'reminder' | null

const loadStatus = ref<LoadStatus>('loading')
const connection = ref<GoogleCalendarConnection | null>(null)
const busy = ref<Busy>(null)
const actionError = ref<string | null>(null)
const redirecting = ref(false)
const isDisconnectOpen = ref(false)

// Formulario del recordatorio.
const useDefaultReminder = ref(true)
const reminderText = ref('30')
const reminderError = ref<string | undefined>(undefined)

const view = computed<GoogleCalendarView | null>(() =>
  connection.value ? viewOf(connection.value) : null,
)
const isConnectedLike = computed(
  () => view.value === 'connected' || view.value === 'syncing' || view.value === 'sync-error',
)
const canConnect = computed(
  () => view.value === 'not-connected' || view.value === 'reauth' || view.value === 'sync-error',
)

let pollTimer: ReturnType<typeof setTimeout> | undefined
let disposed = false

function applyConnection(next: GoogleCalendarConnection) {
  connection.value = next
  if (next.reminderMinutes === null) {
    useDefaultReminder.value = true
  } else {
    useDefaultReminder.value = false
    reminderText.value = String(next.reminderMinutes)
  }
  schedulePoll()
}

function schedulePoll() {
  clearTimeout(pollTimer)
  if (disposed || !connection.value || connection.value.pendingSyncJobs === 0) return
  pollTimer = setTimeout(() => void refresh(), POLL_MS)
}

async function load() {
  loadStatus.value = 'loading'
  const outcome = await fetchConnection()
  if (disposed) return
  if (outcome.kind === 'success') {
    applyConnection(outcome.connection)
    loadStatus.value = 'ready'
  } else {
    loadStatus.value = 'load-error'
  }
}

// Actualiza el estado sin volver a mostrar la carga: lo usan las acciones y el sondeo.
async function refresh() {
  const outcome = await fetchConnection()
  if (disposed) return
  if (outcome.kind === 'success') applyConnection(outcome.connection)
}

onMounted(() => void load())
onUnmounted(() => {
  disposed = true
  clearTimeout(pollTimer)
})

function failureMessage(kind: 'network-error' | 'unexpected-error', what: string): string {
  return kind === 'network-error'
    ? `No pudimos ${what}. Revisa tu conexión e inténtalo de nuevo.`
    : `No pudimos ${what}. Inténtalo de nuevo en un momento.`
}

async function onConnect() {
  if (busy.value) return
  busy.value = 'connect'
  actionError.value = null
  const outcome = await startConnect()
  if (outcome.kind === 'success') {
    // Google toma la pantalla: no se libera `busy` para que nadie vuelva a pulsar.
    redirecting.value = true
    window.location.assign(outcome.authorizationUrl)
    return
  }
  busy.value = null
  if (outcome.kind === 'unavailable') {
    actionError.value = 'Ahora no se puede conectar. Revisa el estado de abajo.'
    await refresh()
  } else {
    actionError.value = failureMessage(outcome.kind, 'abrir Google')
  }
}

async function onSync() {
  if (busy.value) return
  busy.value = 'sync'
  actionError.value = null
  const outcome = await syncNow()
  if (outcome.kind === 'success') {
    toast.success('Sincronizando con Google Calendar', {
      detail: 'Tus cambios pendientes se publicarán en unos segundos.',
    })
    await refresh()
  } else if (outcome.kind === 'not-connected' || outcome.kind === 'unavailable') {
    actionError.value = 'No hay una conexión activa con Google Calendar.'
    await refresh()
  } else {
    actionError.value = failureMessage(outcome.kind, 'sincronizar')
  }
  busy.value = null
}

function openDisconnect() {
  if (busy.value) return
  actionError.value = null
  isDisconnectOpen.value = true
}

async function onConfirmDisconnect() {
  if (busy.value) return
  busy.value = 'disconnect'
  const outcome = await disconnect()
  busy.value = null
  isDisconnectOpen.value = false
  if (outcome.kind === 'success') {
    toast.success('Google Calendar desconectado', {
      detail: 'Los eventos que ya están en tu calendario se quedan allí.',
    })
    await refresh()
  } else {
    actionError.value = failureMessage(outcome.kind, 'desconectar')
  }
}

function onReminderInput(value: string | number) {
  reminderText.value = String(value)
  reminderError.value = undefined
}

async function onSaveReminder() {
  if (busy.value || !connection.value) return
  actionError.value = null
  let minutes: number | null = null
  if (!useDefaultReminder.value) {
    const result = validateReminderMinutes(reminderText.value)
    if (!result.ok) {
      reminderError.value = result.message
      return
    }
    minutes = result.minutes
  }
  reminderError.value = undefined
  busy.value = 'reminder'
  const outcome = await updateReminder(minutes)
  busy.value = null
  if (outcome.kind === 'success') {
    toast.success('Recordatorio guardado', {
      detail:
        minutes === null
          ? 'Tus eventos usarán los recordatorios predeterminados de Google.'
          : `Te avisaremos ${minutes} min antes de cada turno.`,
    })
    await refresh()
  } else if (outcome.kind === 'validation-error') {
    reminderError.value = `El aviso debe ser de ${MIN_REMINDER_MINUTES} a ${MAX_REMINDER_MINUTES} minutos.`
  } else if (outcome.kind === 'not-connected') {
    actionError.value = 'Conecta Google Calendar para guardar el recordatorio.'
    await refresh()
  } else {
    actionError.value = failureMessage(outcome.kind, 'guardar el recordatorio')
  }
}

const dateFormat = new Intl.DateTimeFormat('es-CO', { dateStyle: 'medium', timeStyle: 'short' })
function formatMoment(iso: string | null): string {
  if (!iso) return ''
  const date = new Date(iso)
  return Number.isNaN(date.getTime()) ? '' : dateFormat.format(date)
}

const badge = computed(() => {
  switch (view.value) {
    case 'connected':
      return { variant: 'success' as const, label: 'Conectado' }
    case 'syncing':
      return { variant: 'info' as const, label: 'Sincronizando' }
    case 'reauth':
      return { variant: 'warning' as const, label: 'Requiere reconexión' }
    case 'sync-error':
      return { variant: 'danger' as const, label: 'Error de sincronización' }
    case 'unavailable':
      return { variant: 'neutral' as const, label: 'No disponible' }
    default:
      return { variant: 'neutral' as const, label: 'No conectado' }
  }
})

const lastSyncLabel = computed(() => {
  const when = formatMoment(connection.value?.lastSyncedAt ?? null)
  return when ? `Última sincronización: ${when}` : 'Aún no hay eventos sincronizados.'
})
</script>

<template>
  <section class="gcal-page" aria-labelledby="gcal-page-title">
    <header class="gcal-page__header">
      <h1 id="gcal-page-title" class="gcal-page__title">Google Calendar</h1>
      <p class="gcal-page__subtitle">
        Ve tu agenda de NAVA en el calendario que ya usas y recibe tu aviso de proximidad.
      </p>
      <span class="gcal-page__rule" aria-hidden="true"
        ><span class="gcal-page__rule-line" /><span class="gcal-page__rule-diamond" /><span
          class="gcal-page__rule-line"
      /></span>
    </header>

    <div v-if="loadStatus === 'loading'" class="gcal-page__state" role="status" aria-live="polite">
      <DiamondLoader label="Cargando Google Calendar…" layout="inline" :phrases="LOADING_PHRASES" />
    </div>

    <BaseAlert
      v-else-if="loadStatus === 'load-error'"
      class="gcal-page__wide"
      variant="warning"
      role="alert"
    >
      <strong>No pudimos cargar tu Google Calendar.</strong> Revisa tu conexión e inténtalo de
      nuevo.
      <template #action>
        <BaseButton variant="secondary" type="button" @click="load">Reintentar</BaseButton>
      </template>
    </BaseAlert>

    <div v-else-if="connection && view" class="gcal-page__layout gcal-ink">
      <!-- Sin `title`: BaseAlert lo dibuja como h4 y saltaría niveles bajo el h1/h2. -->
      <BaseAlert v-if="actionError" class="gcal-page__wide" variant="danger" role="alert">
        <strong>No se pudo completar.</strong> {{ actionError }}
      </BaseAlert>

      <!-- Estado y acciones -->
      <section class="gcal-panel" aria-labelledby="gcal-state-title">
        <header class="gcal-panel__header">
          <h2 id="gcal-state-title" class="gcal-panel__title">Tu conexión</h2>
          <BaseBadge :variant="badge.variant" :dot="true" size="md" outline>{{
            badge.label
          }}</BaseBadge>
        </header>

        <div class="gcal-panel__body">
          <p v-if="view === 'unavailable'" class="gcal-copy">
            La conexión con Google Calendar no está disponible en este entorno. El resto de NAVA
            funciona igual.
          </p>

          <template v-else-if="view === 'needs-barber'">
            <p class="gcal-copy">
              Para publicar tu agenda NAVA necesita saber cuál {{ v.professional }} eres. Abre tu
              ficha en {{ v.Professionals }} y pulsa «Este soy yo».
            </p>
            <RouterLink :to="{ name: 'staff-barberos' }" class="gcal-link">
              <span class="gcal-link__copy">
                <span class="gcal-link__title">Ir a {{ v.Professionals }}</span>
                <span class="gcal-link__detail">Elige tu ficha y vincúlala con tu usuario.</span>
              </span>
              <span aria-hidden="true">→</span>
            </RouterLink>
          </template>

          <template v-else-if="view === 'not-connected'">
            <p class="gcal-copy">
              Aún no has conectado tu calendario. Al conectar, tus turnos y bloqueos aparecerán como
              eventos de tu Google Calendar.
            </p>
          </template>

          <template v-else-if="view === 'reauth'">
            <p class="gcal-copy">
              Google ya no acepta el permiso de NAVA (lo retiraste o caducó). Vuelve a conectar para
              seguir publicando tu agenda. Tus turnos no se pierden.
            </p>
          </template>

          <template v-else>
            <dl class="gcal-facts">
              <div>
                <dt>Cuenta de Google</dt>
                <dd>{{ connection.accountEmail ?? 'Cuenta conectada' }}</dd>
              </div>
              <div>
                <dt>Calendario</dt>
                <dd>Principal</dd>
              </div>
            </dl>
            <p class="gcal-copy" aria-live="polite">
              <template v-if="view === 'syncing'">
                Sincronizando {{ connection.pendingSyncJobs }}
                {{ connection.pendingSyncJobs === 1 ? 'cambio' : 'cambios' }}…
              </template>
              <template v-else>{{ lastSyncLabel }}</template>
            </p>
            <BaseAlert v-if="view === 'sync-error'" variant="danger">
              <strong>Error de sincronización.</strong> Algunos cambios no se pudieron publicar en
              Google. Tus turnos están a salvo en NAVA; pulsa «Sincronizar ahora» para reintentar.
            </BaseAlert>
          </template>
        </div>

        <footer
          v-if="view !== 'unavailable' && view !== 'needs-barber'"
          class="gcal-panel__actions"
        >
          <BaseButton
            v-if="canConnect"
            type="button"
            :variant="view === 'sync-error' ? 'secondary' : 'primary'"
            :disabled="busy !== null || redirecting"
            @click="onConnect"
          >
            {{
              redirecting
                ? 'Abriendo Google…'
                : view === 'reauth' || view === 'sync-error'
                  ? 'Volver a conectar'
                  : 'Conectar Google Calendar'
            }}
          </BaseButton>
          <BaseButton
            v-if="isConnectedLike"
            type="button"
            variant="primary"
            :disabled="busy !== null"
            @click="onSync"
          >
            {{ busy === 'sync' ? 'Sincronizando…' : 'Sincronizar ahora' }}
          </BaseButton>
          <BaseButton
            v-if="isConnectedLike || view === 'reauth'"
            type="button"
            variant="secondary"
            :disabled="busy !== null"
            @click="openDisconnect"
          >
            Desconectar
          </BaseButton>
        </footer>
      </section>

      <!-- Recordatorio (solo con una conexión viva) -->
      <section v-if="isConnectedLike" class="gcal-panel" aria-labelledby="gcal-reminder-title">
        <header class="gcal-panel__header">
          <h2 id="gcal-reminder-title" class="gcal-panel__title">Recordatorio</h2>
        </header>
        <form class="gcal-panel__body" novalidate @submit.prevent="onSaveReminder">
          <p class="gcal-copy">
            Google te avisa antes de cada turno en tu calendario. NAVA no envía un aviso propio.
          </p>
          <label class="gcal-check">
            <input v-model="useDefaultReminder" type="checkbox" :disabled="busy !== null" />
            <span>Usar los recordatorios predeterminados de Google</span>
          </label>
          <BaseInput
            v-if="!useDefaultReminder"
            :model-value="reminderText"
            name="reminderMinutes"
            type="number"
            label="Minutos de anticipación"
            :hint="`De ${MIN_REMINDER_MINUTES} a ${MAX_REMINDER_MINUTES}, por ejemplo 30`"
            :error="reminderError"
            :disabled="busy !== null"
            @update:model-value="onReminderInput"
          />
          <div class="gcal-panel__actions gcal-panel__actions--inline">
            <BaseButton type="submit" variant="primary" :disabled="busy !== null">
              {{ busy === 'reminder' ? 'Guardando…' : 'Guardar recordatorio' }}
            </BaseButton>
          </div>
        </form>
      </section>

      <!-- Cómo funciona: visible salvo cuando la integración no está disponible (no hay nada
           que explicar y, sin ningún control enfocable, la región desplazable fallaría axe). -->
      <section v-if="view !== 'unavailable'" class="gcal-panel" aria-labelledby="gcal-how-title">
        <header class="gcal-panel__header">
          <h2 id="gcal-how-title" class="gcal-panel__title">Cómo funciona</h2>
        </header>
        <ul class="gcal-panel__body gcal-list">
          <li>
            <strong>NAVA manda.</strong> La publicación va de NAVA hacia Google. Lo que cambies o
            borres en Google Calendar no modifica tus turnos, tus bloqueos ni tu disponibilidad.
          </li>
          <li>
            <strong>Si borras un evento por error,</strong> NAVA lo vuelve a crear en unos minutos.
            Para dejar de publicar un turno, cancélalo desde NAVA.
          </li>
          <li>
            <strong>Tus clientes reciben la invitación.</strong> Si un cliente dejó su correo,
            Google le envía una invitación al turno y los avisos de cambio o cancelación. Lo que
            responda allí no cambia tu agenda.
          </li>
          <li>
            Al desconectar, los eventos que ya están en tu calendario se quedan; NAVA deja de
            publicar nuevos.
          </li>
        </ul>
      </section>
    </div>

    <BaseDialog
      v-model="isDisconnectOpen"
      title="¿Desconectar Google Calendar?"
      size="sm"
      content-class="gcal-dialog"
    >
      <div class="gcal-ink gcal-dialog__body">
        <p class="gcal-copy">
          NAVA dejará de publicar tus turnos y bloqueos y retirará su permiso en Google. Los eventos
          que ya están en tu calendario se quedan allí y tus turnos no cambian.
        </p>
      </div>
      <template #footer>
        <div class="gcal-ink gcal-dialog__footer">
          <BaseButton
            type="button"
            variant="secondary"
            :disabled="busy === 'disconnect'"
            @click="isDisconnectOpen = false"
          >
            Cancelar
          </BaseButton>
          <BaseButton
            type="button"
            variant="danger"
            :disabled="busy === 'disconnect'"
            @click="onConfirmDisconnect"
          >
            {{ busy === 'disconnect' ? 'Desconectando…' : 'Desconectar' }}
          </BaseButton>
        </div>
      </template>
    </BaseDialog>
  </section>
</template>

<style scoped>
/* Superficie tinta de punta a punta, el mismo canvas que Configuración,
   Agenda y Barberos (estandar-diseno-visual.md §3). */
.gcal-page {
  --gcal-width: 760px;

  display: flex;
  flex-direction: column;
  gap: 24px;
  min-height: 100%;
  padding: 34px 32px 48px;
  color: var(--color-on-strong);
  background: transparent;
  box-sizing: border-box;
}

.gcal-page__header,
.gcal-page__state,
.gcal-page__layout,
.gcal-page > :deep(.base-alert) {
  width: min(100%, var(--gcal-width));
  margin-inline: auto;
}

.gcal-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  font-weight: var(--font-weight-h1);
  line-height: var(--font-size-h1-line);
}

@media (min-width: 1024px) {
  .gcal-page__title {
    font-size: var(--font-size-title-page);
    line-height: 46px;
  }
}

.gcal-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
}

.gcal-page__rule {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
}

.gcal-page__rule-line {
  height: 1px;
  flex: 1;
  background: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
}

.gcal-page__rule-diamond {
  width: 7px;
  height: 7px;
  background: var(--color-brand-accent-surface);
  transform: rotate(45deg);
}

.gcal-page__state {
  color: var(--color-on-strong-muted);
}

.gcal-page__layout {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.gcal-panel {
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-top: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 3px;
}

.gcal-panel__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 24px 14px;
  border-bottom: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 35%, transparent);
}

.gcal-panel__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h3);
  font-weight: var(--font-weight-h1);
  line-height: 1.2;
}

.gcal-panel__body {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 18px 24px 20px;
  margin: 0;
}

.gcal-panel__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 0 24px 20px;
}

.gcal-panel__actions--inline {
  padding: 0;
}

.gcal-copy {
  margin: 0;
  font-size: var(--font-size-body);
  line-height: 1.5;
  color: var(--color-on-strong);
}

.gcal-facts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin: 0;
}

.gcal-facts > div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 14px;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
  min-width: 0;
}

.gcal-facts dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.gcal-facts dd {
  margin: 0;
  overflow-wrap: anywhere;
  font-family: var(--font-display);
  font-size: var(--font-size-h3);
  line-height: 1.2;
}

.gcal-check {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  font-size: var(--font-size-body);
  cursor: pointer;
}

.gcal-list {
  padding-left: 40px;
  list-style: disc;
  line-height: 1.5;
}

.gcal-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  min-height: 44px;
  padding: 14px 16px;
  color: inherit;
  text-decoration: none;
  background-color: var(--color-field-strong);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
}

.gcal-link:hover,
.gcal-link:focus-visible {
  border-color: var(--color-brand-accent-surface);
}

.gcal-link__copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.gcal-link__title {
  font-weight: 600;
}

.gcal-link__detail {
  font-size: var(--font-size-caption);
  color: var(--color-on-strong-muted);
}

.gcal-dialog__body {
  padding: 4px 0;
}

.gcal-dialog__footer {
  display: flex;
  flex: 1;
  flex-flow: row wrap;
  justify-content: flex-end;
  gap: var(--space-3);
}

/* --- Campos y botones sobre tinta: mismo tratamiento que Servicios, Barberos y
   Configuración (el relleno claro por defecto de BaseInput/BaseButton está pensado
   para superficie clara). --- */

.gcal-ink :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  height: 44px;
  color: var(--color-on-strong);
}

.gcal-ink :deep(.base-input__label),
.gcal-ink :deep(.base-input__required) {
  color: var(--color-brand-accent-surface);
}

.gcal-ink :deep(.base-input__hint) {
  color: var(--color-on-strong-muted);
}

.gcal-ink :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

.gcal-ink :deep(.base-input:disabled),
.gcal-ink :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.gcal-ink :deep(.base-input--invalid) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.gcal-ink :deep(.base-input__error) {
  color: var(--color-danger-on-strong);
}

.gcal-ink :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.gcal-ink :deep(.base-button--primary) {
  background-color: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.gcal-ink :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  filter: brightness(92%);
}

.gcal-ink :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  filter: brightness(84%);
}

.gcal-ink :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.gcal-ink :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

/* «Desconectar» confirmado (variant="danger"): ghost rojo, como «Desactivar» en Servicios. */
.gcal-ink :deep(.base-button--danger) {
  background-color: transparent;
  color: var(--color-danger-on-strong);
  border-color: color-mix(in srgb, var(--color-danger-on-strong) 50%, transparent);
  border-bottom-color: var(--color-danger-on-strong);
}

.gcal-ink :deep(.base-button--danger:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 12%, transparent);
  filter: none;
}

.gcal-ink :deep(.base-button:disabled) {
  opacity: 0.4;
}

.gcal-check input {
  appearance: none;
  display: grid;
  place-content: center;
  flex: 0 0 auto;
  width: 22px;
  height: 22px;
  margin: 0;
  background: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 40%, transparent);
  border-radius: 3px;
  cursor: pointer;
}

.gcal-check input:checked {
  background: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.gcal-check input:checked::after {
  width: 10px;
  height: 6px;
  margin-top: -2px;
  content: '';
  border-bottom: 2px solid var(--color-brand-accent-text);
  border-left: 2px solid var(--color-brand-accent-text);
  transform: rotate(-45deg);
}

.gcal-check input:focus-visible {
  outline: none;
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus);
}

@media (max-width: 640px) {
  .gcal-page {
    padding: 24px 16px 40px;
  }

  .gcal-panel__header,
  .gcal-panel__body {
    padding-inline: 16px;
  }

  .gcal-panel__actions {
    padding-inline: 16px;
  }

  .gcal-panel__actions > * {
    flex: 1 1 100%;
  }
}
</style>

<style>
/* SIN "scoped" a propósito (igual que Barberos y Servicios): BaseDialog dibuja su tarjeta,
   encabezado y botón de cerrar en <Teleport to="body">, fuera del alcance del componente.
   Se selecciona por la clase propia de esta pantalla. El !important es necesario porque la
   regla de BaseDialog tiene la misma especificidad y el orden de los <style> no está garantizado. */
.gcal-dialog.base-dialog {
  background-color: var(--color-surface-strong) !important;
  border: var(--border-width-normal) solid var(--color-field-strong-border) !important;
}

.gcal-dialog .base-dialog__header {
  border-bottom-width: var(--border-width-emphasis) !important;
  border-bottom-color: var(--color-brand-accent-surface) !important;
}

.gcal-dialog .base-dialog__title {
  color: var(--color-on-strong) !important;
  overflow-wrap: anywhere;
}

.gcal-dialog .base-dialog__footer {
  border-top-color: var(--color-field-strong-border) !important;
}

.gcal-dialog .base-dialog__close {
  color: var(--color-on-strong-muted) !important;
}

.gcal-dialog .base-dialog__close:hover {
  background-color: color-mix(in srgb, var(--color-on-strong) 8%, transparent) !important;
  color: var(--color-on-strong) !important;
}

.gcal-dialog .base-dialog__close:focus-visible {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-focus) !important;
}
</style>
