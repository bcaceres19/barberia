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
import GoogleCalendarEmblem from '../components/GoogleCalendarEmblem.vue'
import GoogleCalendarSteps from '../components/GoogleCalendarSteps.vue'
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
  receivedAt.value = Date.now()
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
const relativeFormat = new Intl.RelativeTimeFormat('es', { numeric: 'auto' })
function formatMoment(iso: string | null): string {
  if (!iso) return ''
  const date = new Date(iso)
  return Number.isNaN(date.getTime()) ? '' : dateFormat.format(date)
}

// «hace 5 min»: se calcula al recibir el estado, no corre un reloj propio.
const receivedAt = ref(Date.now())
function formatRelative(iso: string | null): string {
  if (!iso) return ''
  const diffSeconds = Math.round((new Date(iso).getTime() - receivedAt.value) / 1000)
  if (Number.isNaN(diffSeconds)) return ''
  const abs = Math.abs(diffSeconds)
  // Un instante futuro (reloj desfasado) se lee como «ahora» en vez de «dentro de…».
  if (diffSeconds > -60) return 'hace un momento'
  if (abs < 3600) return relativeFormat.format(Math.round(diffSeconds / 60), 'minute')
  if (abs < 86400) return relativeFormat.format(Math.round(diffSeconds / 3600), 'hour')
  return relativeFormat.format(Math.round(diffSeconds / 86400), 'day')
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

const pendingLabel = computed(() => {
  const count = connection.value?.pendingSyncJobs ?? 0
  return `${count} ${count === 1 ? 'cambio' : 'cambios'}`
})

// Texto del encabezado de estado: dice qué pasa y qué sigue, sin jerga técnica.
const headline = computed(() => {
  switch (view.value) {
    case 'unavailable':
      return 'Todavía no está disponible'
    case 'needs-barber':
      return `Primero, dinos qué ${v.value.professional} eres`
    case 'reauth':
      return 'Hay que volver a conectar Google'
    case 'sync-error':
      return 'Algunos cambios no llegaron a Google'
    case 'syncing':
      return 'Publicando tus cambios'
    case 'connected':
      return 'Tu agenda ya está en Google Calendar'
    default:
      return 'Lleva tu agenda a Google Calendar'
  }
})

const lead = computed(() => {
  switch (view.value) {
    case 'unavailable':
      return 'La conexión con Google Calendar no está disponible en este entorno. El resto de NAVA funciona igual.'
    case 'needs-barber':
      return `Para publicar tu agenda, NAVA necesita saber cuál ${v.value.professional} eres. Abre tu ficha en ${v.value.Professionals} y pulsa «Este soy yo»: toma un clic.`
    case 'reauth':
      return 'Retiraste el permiso de NAVA o caducó. Tus turnos no se pierden: vuelve a conectar y seguirán publicándose.'
    case 'sync-error':
      return 'Tus turnos están a salvo en NAVA. Pulsa «Sincronizar ahora» para reintentar el envío.'
    case 'syncing':
      return `Estamos enviando ${pendingLabel.value} a Google. Tarda unos segundos.`
    case 'connected':
      return 'Cada turno nuevo, cambio o cancelación se publica solo, y tus clientes reciben su invitación.'
    default:
      return 'Tus turnos y bloqueos aparecerán solos en tu calendario, y tus clientes recibirán la invitación por correo.'
  }
})

// Recorrido hacia la conexión: solo mientras todavía no hay una conexión viva.
const steps = computed(() => [
  { title: `Vincula tu ficha`, detail: `Dinos qué ${v.value.professional} eres.` },
  { title: 'Conecta tu Google', detail: 'Das tu permiso una sola vez.' },
  { title: 'Listo', detail: 'Tus turnos se publican solos.' },
])
const showSteps = computed(
  () => view.value === 'needs-barber' || view.value === 'not-connected' || view.value === 'reauth',
)
const currentStep = computed(() => (view.value === 'needs-barber' ? 1 : 2))

// Recordatorio: dos modos y atajos de minutos habituales.
const reminderMode = computed({
  get: () => (useDefaultReminder.value ? 'default' : 'custom'),
  set: (mode: 'default' | 'custom') => {
    useDefaultReminder.value = mode === 'default'
    reminderError.value = undefined
  },
})
const REMINDER_PRESETS = [
  { minutes: 10, label: '10 min' },
  { minutes: 15, label: '15 min' },
  { minutes: 30, label: '30 min' },
  { minutes: 60, label: '1 hora' },
  { minutes: 120, label: '2 horas' },
] as const
function pickPreset(minutes: number) {
  if (busy.value) return
  reminderText.value = String(minutes)
  reminderError.value = undefined
}

const HOW_IT_WORKS = [
  {
    title: 'NAVA manda',
    text: 'Google es solo una vista. Lo que cambies o borres allí no modifica tus turnos, bloqueos ni disponibilidad.',
  },
  {
    title: 'Si borras un evento',
    text: 'NAVA lo vuelve a crear en unos minutos. Para dejar de publicar un turno, cancélalo desde NAVA.',
  },
  {
    title: 'Tus clientes, invitados',
    text: 'Si dejaron su correo, Google les envía la invitación y los avisos de cambio o cancelación. Lo que respondan no cambia tu agenda.',
  },
  {
    title: 'Si desconectas',
    text: 'Los eventos que ya están en tu calendario se quedan. NAVA solo deja de publicar nuevos.',
  },
] as const
</script>

<template>
  <section class="gcal-page" aria-labelledby="gcal-page-title">
    <header class="gcal-page__header nv-rise">
      <h1 id="gcal-page-title" class="gcal-page__title">Google Calendar</h1>
      <p class="gcal-page__subtitle">
        Tus turnos y bloqueos de NAVA, en el calendario que ya usas.
      </p>
      <span class="gcal-page__rule" aria-hidden="true"
        ><span class="gcal-page__rule-line nv-wipe" /><span
          class="gcal-page__rule-diamond nv-pop" /><span class="gcal-page__rule-line nv-wipe"
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
      <section
        class="gcal-hero nv-rise"
        :class="`gcal-hero--${view}`"
        style="--i: 1"
        aria-labelledby="gcal-state-title"
      >
        <GoogleCalendarEmblem class="gcal-hero__emblem" :view="view" />

        <div class="gcal-hero__main">
          <BaseBadge
            :variant="badge.variant"
            :dot="true"
            size="md"
            outline
            class="gcal-hero__badge"
            >{{ badge.label }}</BaseBadge
          >
          <!-- `key`: al cambiar de estado el texto entra de nuevo en vez de cambiar a golpe. -->
          <div :key="view" class="gcal-hero__copy nv-rise">
            <h2 id="gcal-state-title" class="gcal-hero__title">{{ headline }}</h2>
            <p class="gcal-copy" aria-live="polite">{{ lead }}</p>
          </div>

          <RouterLink
            v-if="view === 'needs-barber'"
            :to="{ name: 'staff-barberos' }"
            class="gcal-link nv-lift"
          >
            <span class="gcal-link__copy">
              <span class="gcal-link__title">Ir a {{ v.Professionals }}</span>
              <span class="gcal-link__detail">Elige tu ficha y vincúlala con tu usuario.</span>
            </span>
            <span class="gcal-link__arrow" aria-hidden="true">→</span>
          </RouterLink>

          <dl v-if="isConnectedLike" class="gcal-facts">
            <div class="nv-rise" style="--i: 2">
              <dt>Cuenta de Google</dt>
              <dd>{{ connection.accountEmail ?? 'Cuenta conectada' }}</dd>
            </div>
            <div class="nv-rise" style="--i: 3">
              <dt>Calendario</dt>
              <dd>Principal</dd>
            </div>
            <div class="nv-rise" style="--i: 4">
              <dt>Última sincronización</dt>
              <dd v-if="connection.lastSyncedAt">
                {{ formatMoment(connection.lastSyncedAt) }}
                <small>{{ formatRelative(connection.lastSyncedAt) }}</small>
              </dd>
              <dd v-else>Aún no hay eventos sincronizados.</dd>
            </div>
          </dl>
        </div>

        <footer v-if="view !== 'unavailable' && view !== 'needs-barber'" class="gcal-hero__actions">
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
            <span class="gcal-action">
              <svg
                class="gcal-action__icon"
                :class="{ 'gcal-action__icon--spin': busy === 'sync' || view === 'syncing' }"
                viewBox="0 0 20 20"
                width="16"
                height="16"
                aria-hidden="true"
                focusable="false"
              >
                <path d="M16 10a6 6 0 0 1-10.5 4M4 10a6 6 0 0 1 10.5-4M14.5 3v3h-3M5.5 17v-3h3" />
              </svg>
              {{ busy === 'sync' ? 'Sincronizando…' : 'Sincronizar ahora' }}
            </span>
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

      <!-- Recorrido hacia la conexión -->
      <GoogleCalendarSteps
        v-if="showSteps"
        :steps="steps"
        :current="currentStep"
        label="Pasos para conectar Google Calendar"
      />

      <!-- Recordatorio (solo con una conexión viva) -->
      <section
        v-if="isConnectedLike"
        class="gcal-panel nv-rise"
        style="--i: 2"
        aria-labelledby="gcal-reminder-title"
      >
        <header class="gcal-panel__header">
          <h2 id="gcal-reminder-title" class="gcal-panel__title">Recordatorio</h2>
        </header>
        <form class="gcal-panel__body" novalidate @submit.prevent="onSaveReminder">
          <p class="gcal-copy">
            Google te avisa antes de cada turno en tu calendario. NAVA no envía un aviso propio.
          </p>

          <fieldset class="gcal-modes" :disabled="busy !== null">
            <legend class="gcal-sr-only">Cuándo quieres que te avise Google</legend>
            <label class="gcal-mode" :class="{ 'gcal-mode--on': reminderMode === 'default' }">
              <input v-model="reminderMode" type="radio" name="reminderMode" value="default" />
              <span class="gcal-mode__dot" aria-hidden="true" />
              <span class="gcal-mode__copy">
                <span class="gcal-mode__title">Avisos de Google</span>
                <span class="gcal-mode__detail"
                  >Usa los recordatorios que ya tienes en tu calendario.</span
                >
              </span>
            </label>
            <label class="gcal-mode" :class="{ 'gcal-mode--on': reminderMode === 'custom' }">
              <input v-model="reminderMode" type="radio" name="reminderMode" value="custom" />
              <span class="gcal-mode__dot" aria-hidden="true" />
              <span class="gcal-mode__copy">
                <span class="gcal-mode__title">Avisarme antes de cada turno</span>
                <span class="gcal-mode__detail">Tú eliges cuántos minutos de anticipación.</span>
              </span>
            </label>
          </fieldset>

          <Transition name="gcal-expand">
            <div v-if="!useDefaultReminder" class="gcal-custom">
              <div class="gcal-chips" role="group" aria-label="Atajos de anticipación">
                <button
                  v-for="preset in REMINDER_PRESETS"
                  :key="preset.minutes"
                  type="button"
                  class="gcal-chip"
                  :class="{ 'gcal-chip--on': reminderText === String(preset.minutes) }"
                  :aria-pressed="reminderText === String(preset.minutes)"
                  :disabled="busy !== null"
                  @click="pickPreset(preset.minutes)"
                >
                  {{ preset.label }}
                </button>
              </div>
              <BaseInput
                :model-value="reminderText"
                name="reminderMinutes"
                type="number"
                label="Minutos de anticipación"
                :hint="`De ${MIN_REMINDER_MINUTES} a ${MAX_REMINDER_MINUTES}, por ejemplo 30`"
                :error="reminderError"
                :disabled="busy !== null"
                @update:model-value="onReminderInput"
              />
            </div>
          </Transition>

          <div class="gcal-panel__actions gcal-panel__actions--inline">
            <BaseButton type="submit" variant="primary" :disabled="busy !== null">
              {{ busy === 'reminder' ? 'Guardando…' : 'Guardar recordatorio' }}
            </BaseButton>
          </div>
        </form>
      </section>

      <!-- Cómo funciona: visible salvo cuando la integración no está disponible (no hay nada
           que explicar y, sin ningún control enfocable, la región desplazable fallaría axe). -->
      <section v-if="view !== 'unavailable'" class="gcal-how" aria-labelledby="gcal-how-title">
        <h2 id="gcal-how-title" class="gcal-how__title nv-rise" style="--i: 3">Cómo funciona</h2>
        <ul class="gcal-how__grid">
          <li
            v-for="(item, index) in HOW_IT_WORKS"
            :key="item.title"
            class="gcal-how__item nv-rise"
            :style="{ '--i': index + 4 }"
          >
            <span class="gcal-how__num" aria-hidden="true">{{ index + 1 }}</span>
            <strong class="gcal-how__name">{{ item.title }}</strong>
            <span class="gcal-how__text">{{ item.text }}</span>
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
  gap: 16px;
  padding: 18px 24px 22px;
  margin: 0;
}

.gcal-panel__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.gcal-panel__actions--inline {
  padding: 0;
}

.gcal-copy {
  margin: 0;
  font-size: var(--font-size-body);
  line-height: 1.55;
  color: var(--color-on-strong);
}

.gcal-sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}

/* --- Encabezado de estado --------------------------------------------------- */

.gcal-hero {
  /* El tono lo fija el estado; lo lee también el emblema. */
  --gcal-tone: var(--color-brand-accent-surface);

  position: relative;
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  column-gap: 28px;
  row-gap: 20px;
  align-items: center;
  padding: 26px 28px 24px;
  overflow: hidden;
  background:
    radial-gradient(
      ellipse 60% 120% at 0% 0%,
      color-mix(in srgb, var(--gcal-tone) 14%, transparent),
      transparent 70%
    ),
    color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-top: var(--border-width-emphasis) solid var(--gcal-tone);
  border-radius: 3px;
  transition:
    border-color 400ms var(--motion-ease-out, ease-out),
    background-color 400ms var(--motion-ease-out, ease-out);
}

.gcal-hero--connected {
  --gcal-tone: var(--color-success-on-strong);
}

.gcal-hero--syncing {
  --gcal-tone: var(--color-info-on-strong);
}

.gcal-hero--reauth {
  --gcal-tone: var(--color-warning-on-strong);
}

.gcal-hero--sync-error {
  --gcal-tone: var(--color-danger-on-strong);
}

.gcal-hero--unavailable {
  --gcal-tone: var(--color-on-strong-muted);
}

.gcal-hero__emblem {
  width: 112px;
  height: 112px;
}

/* El badge outline neutro desaparece sobre tinta: texto y borde toman el tono del estado. */
.gcal-hero__badge {
  color: var(--color-on-strong);
  background: color-mix(in srgb, var(--gcal-tone) 14%, transparent);
  border-color: color-mix(in srgb, var(--gcal-tone) 70%, transparent);
}

.gcal-hero__main {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.gcal-hero__copy {
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-width: 0;
}

.gcal-hero__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2, 26px);
  font-weight: var(--font-weight-h1);
  line-height: 1.15;
  overflow-wrap: anywhere;
}

.gcal-hero__copy .gcal-copy {
  max-width: 56ch;
  color: var(--color-on-strong-muted);
}

.gcal-hero__actions {
  display: flex;
  flex-wrap: wrap;
  grid-column: 1 / -1;
  gap: 12px;
  padding-top: 18px;
  border-top: var(--border-width-normal) solid color-mix(in srgb, var(--gcal-tone) 28%, transparent);
}

.gcal-action {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.gcal-action__icon {
  flex: 0 0 auto;
  fill: none;
  stroke: currentcolor;
  stroke-width: 1.7;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.gcal-action__icon--spin {
  animation: gcal-spin 1s linear infinite;
}

.gcal-facts {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  width: 100%;
  margin: 4px 0 0;
}

/* El correo es lo más largo: ocupa su propia fila para no partirse en dos líneas. */
.gcal-facts > div:first-child {
  grid-column: 1 / -1;
}

.gcal-facts > div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  padding: 12px 14px;
  background-color: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-left: 2px solid var(--gcal-tone);
  border-radius: var(--radius-sm);
}

.gcal-facts dt {
  color: var(--color-on-strong-muted);
  font-size: var(--font-size-caption);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.gcal-facts dd {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin: 0;
  overflow-wrap: anywhere;
  font-size: var(--font-size-body);
  line-height: 1.3;
}

.gcal-facts dd small {
  font-size: var(--font-size-caption);
  color: var(--color-on-strong-muted);
}

.gcal-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
  max-width: 460px;
  min-height: 44px;
  padding: 14px 16px;
  color: inherit;
  text-decoration: none;
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-radius: var(--radius-sm);
}

.gcal-link:hover,
.gcal-link:focus-visible {
  border-color: var(--color-brand-accent-surface);
}

.gcal-link:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
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

.gcal-link__arrow {
  color: var(--color-brand-accent-surface);
  transition: transform 220ms var(--motion-ease-out, ease-out);
}

.gcal-link:hover .gcal-link__arrow,
.gcal-link:focus-visible .gcal-link__arrow {
  transform: translateX(4px);
}

/* --- Recordatorio ----------------------------------------------------------- */

.gcal-modes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  min-width: 0;
  padding: 0;
  margin: 0;
  border: 0;
}

.gcal-mode {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-height: 44px;
  padding: 14px 16px;
  cursor: pointer;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
  transition:
    background-color 220ms var(--motion-ease-out, ease-out),
    border-color 220ms var(--motion-ease-out, ease-out),
    transform 220ms var(--motion-ease-out, ease-out);
}

.gcal-mode:hover {
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 60%, transparent);
}

.gcal-mode--on {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border-color: var(--color-brand-accent-surface);
}

/* El radio nativo sigue siendo el control; se oculta solo a la vista. */
.gcal-mode input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  cursor: pointer;
  opacity: 0;
}

.gcal-mode:has(input:focus-visible) {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}

.gcal-mode__dot {
  position: relative;
  flex: 0 0 auto;
  width: 20px;
  height: 20px;
  margin-top: 1px;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 40%, transparent);
  border-radius: 50%;
  transition: border-color 220ms var(--motion-ease-out, ease-out);
}

.gcal-mode__dot::after {
  position: absolute;
  inset: 4px;
  content: '';
  background: var(--color-brand-accent-surface);
  border-radius: 50%;
  transform: scale(0);
  transition: transform 260ms var(--motion-ease-spring, ease-out);
}

.gcal-mode--on .gcal-mode__dot {
  border-color: var(--color-brand-accent-surface);
}

.gcal-mode--on .gcal-mode__dot::after {
  transform: scale(1);
}

.gcal-mode__copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.gcal-mode__title {
  font-weight: 600;
  line-height: 1.3;
}

.gcal-mode__detail {
  font-size: var(--font-size-caption);
  line-height: 1.4;
  color: var(--color-on-strong-muted);
}

.gcal-custom {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 16px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border-left: 2px solid var(--color-brand-accent-surface);
}

.gcal-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.gcal-chip {
  min-height: 36px;
  padding: 0 14px;
  font: inherit;
  font-size: var(--font-size-body-sm);
  color: var(--color-on-strong);
  cursor: pointer;
  background: transparent;
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  border-radius: 999px;
  transition:
    background-color 200ms var(--motion-ease-out, ease-out),
    border-color 200ms var(--motion-ease-out, ease-out),
    color 200ms var(--motion-ease-out, ease-out),
    transform 160ms var(--motion-ease-out, ease-out);
}

.gcal-chip:hover:not(:disabled) {
  border-color: var(--color-brand-accent-surface);
}

.gcal-chip:active:not(:disabled) {
  transform: scale(0.96);
}

.gcal-chip--on {
  color: var(--color-brand-accent-text);
  background: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.gcal-chip:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}

.gcal-chip:disabled {
  cursor: default;
  opacity: 0.45;
}

.gcal-expand-enter-active,
.gcal-expand-leave-active {
  transition:
    opacity 240ms var(--motion-ease-out, ease-out),
    transform 240ms var(--motion-ease-out, ease-out);
}

.gcal-expand-enter-from,
.gcal-expand-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* --- Cómo funciona ---------------------------------------------------------- */

.gcal-how {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 6px;
}

.gcal-how__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h3);
  font-weight: var(--font-weight-h1);
  line-height: 1.2;
}

.gcal-how__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  padding: 0;
  margin: 0;
  list-style: none;
}

.gcal-how__item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  column-gap: 12px;
  row-gap: 4px;
  padding: 16px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: var(--radius-sm);
  transition:
    border-color 240ms var(--motion-ease-out, ease-out),
    transform 240ms var(--motion-ease-out, ease-out);
}

@media (hover: hover) {
  .gcal-how__item:hover {
    border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
    transform: translateY(-2px);
  }
}

.gcal-how__num {
  grid-row: span 2;
  display: grid;
  place-content: center;
  width: 28px;
  height: 28px;
  font-family: var(--font-display);
  font-size: var(--font-size-body-sm);
  color: var(--color-brand-accent-surface);
  border: var(--border-width-normal) solid
    color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
  border-radius: 50%;
}

.gcal-how__name {
  align-self: center;
  line-height: 1.3;
}

.gcal-how__text {
  grid-column: 2;
  font-size: var(--font-size-body-sm);
  line-height: 1.5;
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

@keyframes gcal-spin {
  to {
    transform: rotate(360deg);
  }
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

@media (max-width: 640px) {
  .gcal-page {
    padding: 24px 16px 40px;
  }

  .gcal-panel__header,
  .gcal-panel__body {
    padding-inline: 16px;
  }

  .gcal-hero {
    grid-template-columns: 1fr;
    justify-items: start;
    padding: 20px 16px;
  }

  .gcal-hero__emblem {
    width: 84px;
    height: 84px;
  }

  .gcal-hero__actions {
    width: 100%;
  }

  .gcal-hero__actions > :deep(.base-button),
  .gcal-panel__actions > * {
    flex: 1 1 100%;
  }

  .gcal-modes,
  .gcal-how__grid,
  .gcal-facts {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .gcal-action__icon--spin {
    animation: none;
  }

  .gcal-hero,
  .gcal-mode,
  .gcal-mode__dot::after,
  .gcal-chip,
  .gcal-how__item,
  .gcal-link__arrow,
  .gcal-expand-enter-active,
  .gcal-expand-leave-active {
    transition: none;
  }

  .gcal-how__item:hover,
  .gcal-link:hover .gcal-link__arrow {
    transform: none;
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

<style>
/* Animaciones reducidas desde Configuración (DEC-112): mismo efecto que prefers-reduced-motion. */
:root[data-motion='reduced'] .gcal-action__icon--spin {
  animation: none;
}

:root[data-motion='reduced'] .gcal-hero,
:root[data-motion='reduced'] .gcal-mode,
:root[data-motion='reduced'] .gcal-mode__dot::after,
:root[data-motion='reduced'] .gcal-chip,
:root[data-motion='reduced'] .gcal-how__item,
:root[data-motion='reduced'] .gcal-link__arrow,
:root[data-motion='reduced'] .gcal-expand-enter-active,
:root[data-motion='reduced'] .gcal-expand-leave-active {
  transition: none;
}
</style>
