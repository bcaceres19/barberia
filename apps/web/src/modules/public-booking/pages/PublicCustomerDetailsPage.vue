<script setup lang="ts">
// Página de captura de datos del cliente y persona atendida (HU-096,
// CA-096-01 a CA-096-06) y confirmación pública concurrente (HU-097,
// CA-097-01 a CA-097-07). Sin mockup asignado: composición libre dentro de
// NAVA / Tailored Grid (DEC-078), alojada en el cascarón de la reserva
// pública (DEC-111): lo que se rellena vive en una hoja de papel marfil
// sobre el lienzo de tinta. Valida/normaliza en el cliente con las
// mismas reglas que el backend aplicará (identity.go), conserva los datos
// ante un error de validación o conflicto (CA-096-03, CA-097-03) y resuelve
// `attendeeName` sin pedirlo dos veces cuando el cliente reserva para sí
// mismo (CA-096-01, RN-RES-02).
import { computed, nextTick, onMounted, ref, useId, watch } from 'vue'
import { useRouter } from 'vue-router'
import { BarberAvatar, BaseAlert, BaseButton, BaseDialog, BaseInput } from '@/shared/ui'
import type { LatLng } from '@/shared/model/mapPicker'
import { isPlacesAvailable, type PlaceSelection } from '@/shared/model/placesAutocomplete'
import HomeLocationField from '../components/HomeLocationField.vue'
import PhoneField from '../components/PhoneField.vue'
import { confirmPublicAppointment } from '../api/confirmPublicAppointmentApi'
import { listPublicBarbers } from '../api/listPublicBarbersApi'
import { useBookingChrome } from '../model/bookingChrome'
import { newIdempotencyKey } from '../model/idempotencyKey'
import { loadPublicProfile } from '../model/publicProfile'
import type { PublicBarber } from '../model/publicBarberListOutcome'
import { publicBarberPhotoUrl } from '../model/publicBarberPhotoUrl'
import { DEFAULT_PHONE_COUNTRY, composePhone, phoneCountry } from '../model/phoneCountries'
import { usePublicVocabulary } from '../model/publicVocabulary'
import type {
  ConfirmedPublicAppointment,
  PublicAppointmentAlternative,
} from '../model/confirmPublicAppointmentOutcome'
import {
  ATTENDEE_NAME_MAX_LENGTH,
  CUSTOMER_NOTE_MAX_LENGTH,
  validateAttendeeName,
  validateCustomerEmail,
  validateCustomerFullName,
  validateCustomerNote,
  validateCustomerPhoneParts,
} from '../validation/customerIdentityValidation'

interface Props {
  /** Identificador del enlace público, tal como llega del parámetro de ruta
   * `:slug`. */
  slug: string
  /** Servicio ya elegido (HU-091). */
  serviceId: string
  /** Barbero ya elegido (HU-092). */
  barberId: string
  /** Franja ya elegida (HU-095), instante ISO absoluto. El servidor la
   * revalida por completo al confirmar (CA-097-02); esta página solo la
   * conserva y la envía tal cual. */
  startsAt: string
}
const props = defineProps<Props>()
const v = usePublicVocabulary()
const router = useRouter()
const gemId = `seal-gem-${useId()}`
// Único por instancia: el diálogo vive en el <body> y no debe repetir ids.
const conflictAltsLabelId = `pb-conflict-alts-${useId()}`

// El cascarón dibuja el progreso: al confirmar, se completa.
const chrome = useBookingChrome()

type AttendeeChoice = 'self' | 'other'

const attendeeChoice = ref<AttendeeChoice>('self')
const fullName = ref('')
// El prefijo vive en su propia casilla; al API viaja el E.164 compuesto.
const phoneCountryIso = ref(DEFAULT_PHONE_COUNTRY)
const phoneNational = ref('')
const phone = computed(() => composePhone(phoneCountryIso.value, phoneNational.value))
const email = ref('')
const note = ref('')
const attendeeName = ref('')

// El barbero elegido, para mostrarlo en el resumen con su foto. La página puede
// abrirse directamente desde un enlace, así que se busca por id en la lista
// pública del servicio. Es solo presentación: si no llega, el resumen sigue
// siendo completo (el servidor valida el barbero al confirmar).
const barber = ref<PublicBarber | null>(null)
onMounted(async () => {
  const outcome = await listPublicBarbers(props.slug, props.serviceId)
  if (outcome.kind === 'success') {
    barber.value = outcome.barbers.find((item) => item.id === props.barberId) ?? null
  }
})

// Servicio a domicilio (DEC-131, HU-103): el paso solo existe si la barbería lo ofrece Y hay
// forma de elegir una dirección con el autocompletado en este entorno. Se lee fresco al
// abrir la página; el servidor lo revalida al confirmar (CA-103-05), así que un valor viejo
// aquí solo produciría un 422 explicado, nunca un turno equivocado.
type ServiceMode = 'local' | 'home'
const offersHomeService = ref(false)
const serviceMode = ref<ServiceMode>('local')
const homePlace = ref<PlaceSelection | null>(null)
// Dónde abre el mapa de «a domicilio»: la barbería, si publicó su ubicación (DEC-138).
const homeMapStart = ref<LatLng | null>(null)
onMounted(async () => {
  const outcome = await loadPublicProfile(props.slug)
  if (outcome.kind === 'success') {
    offersHomeService.value = outcome.profile.offersHomeService && isPlacesAvailable()
    const where = outcome.profile.location
    homeMapStart.value = where ? { latitude: where.latitude, longitude: where.longitude } : null
  }
})
// Sin la opción visible, el turno siempre es en el local aunque el estado local diga otra cosa.
const isHomeService = computed(() => offersHomeService.value && serviceMode.value === 'home')

const attempted = ref(false)
const reviewing = ref(false)

// selectedStartsAt permite reintentar con una alternativa (RN-CON-05,
// DEC-090) sin recargar la página ni perder el resto del formulario.
const selectedStartsAt = ref(props.startsAt)

type ConfirmState = 'idle' | 'submitting' | 'confirmed' | 'schedule-conflict' | 'error'
const confirmState = ref<ConfirmState>('idle')
const confirmedAppointment = ref<ConfirmedPublicAppointment | null>(null)
const alternatives = ref<PublicAppointmentAlternative[]>([])
const errorMessage = ref('')
// El servidor las entrega por cercanía; al cliente le sirven en orden de reloj.
const sortedAlternatives = computed(() =>
  [...alternatives.value].sort((a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt)),
)
// idempotencyKey se conserva mientras el reintento es la MISMA solicitud
// (red/error inesperado, RN-IDE-01): un contenido distinto (editar, elegir
// otra alternativa) siempre pide una clave nueva.
const idempotencyKey = ref<string | null>(null)

const formRef = ref<HTMLElement | null>(null)
const summaryRef = ref<HTMLElement | null>(null)
const actionsRef = ref<HTMLElement | null>(null)
const confirmedRef = ref<HTMLElement | null>(null)

const fieldErrors = computed(() => ({
  fullName: validateCustomerFullName(fullName.value),
  phone: validateCustomerPhoneParts(phoneCountry(phoneCountryIso.value).dial, phoneNational.value),
  email: validateCustomerEmail(email.value),
  note: validateCustomerNote(note.value),
  attendeeName:
    attendeeChoice.value === 'other' ? validateAttendeeName(attendeeName.value) : undefined,
  // DEC-138: basta una ubicación marcada (mapa, «Usar mi ubicación» o una sugerencia elegida);
  // un texto libre sin ninguna de ellas no es una ubicación.
  home:
    isHomeService.value && !homePlace.value
      ? 'Marca tu ubicación en el mapa, usa tu ubicación o elige una dirección de la lista.'
      : undefined,
}))

const hasErrors = computed(() => Object.values(fieldErrors.value).some((message) => !!message))

// CA-096-01: para sí mismo, el nombre atendido se deriva del cliente sin
// pedirlo dos veces.
const resolvedAttendeeName = computed(() =>
  attendeeChoice.value === 'self' ? fullName.value.trim() : attendeeName.value.trim(),
)

async function handleSubmit() {
  attempted.value = true
  if (hasErrors.value) {
    await nextTick()
    const firstErrorField = document.querySelector<HTMLElement>('[aria-invalid="true"]')
    firstErrorField?.focus()
    return
  }
  reviewing.value = true
  await nextTick()
  summaryRef.value?.focus()
}

// Tras un fallo el botón se había deshabilitado y el foco cayó al <body>; además, en móvil el
// aviso entra bajo el borde de la ventana y empuja los botones fuera de ella. Se devuelve el
// foco a «Confirmar turno» (reintento por teclado) y se lleva a la vista aviso y botones.
watch(confirmState, async (state) => {
  if (state !== 'error') return
  await nextTick()
  actionsRef.value
    ?.querySelector<HTMLElement>('button:last-of-type')
    ?.focus({ preventScroll: true })
  actionsRef.value?.scrollIntoView?.({ block: 'end' })
})

async function editAgain() {
  reviewing.value = false
  confirmState.value = 'idle'
  alternatives.value = []
  errorMessage.value = ''
  idempotencyKey.value = null
  await nextTick()
  formRef.value?.querySelector<HTMLElement>('input, textarea')?.focus()
}

// Fecha larga y hora por separado: el resumen da a la hora el protagonismo.
function formatLocalDay(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('es-CO', { dateStyle: 'full' })
  } catch {
    return iso
  }
}

function formatLocalDayShort(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString('es-CO', {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
    })
  } catch {
    return iso
  }
}

function formatLocalTime(iso: string): string {
  try {
    return new Date(iso).toLocaleTimeString('es-CO', { timeStyle: 'short' })
  } catch {
    return ''
  }
}

// Piezas del talón de confirmación. Se calculan en la zona del navegador, igual
// que el resto de fechas de esta pantalla.
function leafParts(iso: string): { month: string; day: string; weekday: string } {
  const date = new Date(iso)
  return {
    month: date.toLocaleDateString('es-CO', { month: 'short' }).replace('.', ''),
    day: String(date.getDate()),
    weekday: date.toLocaleDateString('es-CO', { weekday: 'long' }),
  }
}

// «Hoy», «Mañana» o «En N días»: cuenta días de calendario, no bloques de 24 h.
function untilLabel(iso: string): string {
  const target = new Date(iso)
  const today = new Date()
  const days = Math.round(
    (new Date(target.getFullYear(), target.getMonth(), target.getDate()).getTime() -
      new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime()) /
      86_400_000,
  )
  if (days <= 0) return 'Hoy'
  if (days === 1) return 'Mañana'
  return `En ${days} días`
}

function formatMoney(amount: string, currency: string): string {
  const value = Number(amount)
  if (!Number.isFinite(value)) return `${amount} ${currency}`
  try {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency,
      maximumFractionDigits: 0,
    }).format(value)
  } catch {
    return `${amount} ${currency}`
  }
}

// CA-097-07: bloquea doble toque (confirmState guarda 'submitting' hasta
// que el intento termina) y conserva los datos del formulario en
// cualquier desenlace (nunca se limpian fuera de 'confirmed').
async function handleConfirm() {
  if (confirmState.value === 'submitting') return

  if (!idempotencyKey.value) {
    idempotencyKey.value = newIdempotencyKey()
  }

  confirmState.value = 'submitting'
  errorMessage.value = ''

  const outcome = await confirmPublicAppointment(
    props.slug,
    props.serviceId,
    props.barberId,
    {
      startsAt: selectedStartsAt.value,
      fullName: fullName.value.trim(),
      phone: phone.value,
      email: email.value.trim(),
      note: note.value.trim() || null,
      forSomeoneElse: attendeeChoice.value === 'other',
      attendeeName: attendeeChoice.value === 'other' ? attendeeName.value.trim() : null,
      homeService:
        isHomeService.value && homePlace.value
          ? {
              address: homePlace.value.address,
              latitude: homePlace.value.latitude,
              longitude: homePlace.value.longitude,
            }
          : null,
    },
    idempotencyKey.value,
  )

  switch (outcome.kind) {
    case 'success':
      confirmedAppointment.value = outcome.appointment
      confirmState.value = 'confirmed'
      chrome.completed.value = true
      idempotencyKey.value = null
      await nextTick()
      confirmedRef.value?.focus()
      return
    case 'schedule-conflict':
      alternatives.value = outcome.alternatives
      confirmState.value = 'schedule-conflict'
      idempotencyKey.value = null
      return
    case 'not-found':
      errorMessage.value =
        'Este enlace de reserva ya no está disponible. Vuelve a intentarlo desde el inicio.'
      confirmState.value = 'error'
      idempotencyKey.value = null
      return
    case 'validation-error':
      errorMessage.value = outcome.detail
      confirmState.value = 'error'
      idempotencyKey.value = null
      return
    case 'idempotency-conflict':
      errorMessage.value = 'Hubo un problema al procesar tu confirmación. Intenta de nuevo.'
      confirmState.value = 'error'
      idempotencyKey.value = null
      return
    case 'network-error':
      errorMessage.value =
        'No pudimos conectar con el servidor. Verifica tu conexión e intenta de nuevo.'
      confirmState.value = 'error'
      return
    case 'unexpected-error':
      errorMessage.value = 'Ocurrió un error inesperado. Intenta de nuevo.'
      confirmState.value = 'error'
      return
  }
}

// El conflicto obliga a editar el turno: o se elige una de las horas libres
// cercanas (el resumen se actualiza y hay que volver a confirmar) o se vuelve
// al calendario. Los datos del formulario siguen en esta página mientras tanto.
function pickAnotherSlot() {
  void router.push({
    name: 'reserva-publica-horario',
    params: { slug: props.slug, serviceId: props.serviceId, barberId: props.barberId },
  })
}

function chooseAlternative(startsAt: string) {
  selectedStartsAt.value = startsAt
  alternatives.value = []
  confirmState.value = 'idle'
  idempotencyKey.value = null
}
</script>

<template>
  <main
    class="pb-page pb-page--center"
    :class="confirmState === 'confirmed' ? 'pb-page--confirmed' : 'pb-page--scroll-form'"
  >
    <span v-if="confirmState === 'confirmed'" class="customer-details__seal" aria-hidden="true">
      <svg class="customer-details__gem" viewBox="0 0 64 64" fill="none" focusable="false">
        <defs>
          <clipPath :id="`${gemId}-clip`">
            <path d="M32 4 60 32 32 60 4 32Z" />
          </clipPath>
          <linearGradient :id="`${gemId}-glint`" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="#e6f5ea" stop-opacity="0" />
            <stop offset="0.5" stop-color="#e6f5ea" stop-opacity="0.85" />
            <stop offset="1" stop-color="#e6f5ea" stop-opacity="0" />
          </linearGradient>
        </defs>
        <path class="customer-details__gem-inner" d="M32 17 47 32 32 47 17 32Z" />
        <path class="customer-details__gem-facets" d="M32 4V60M4 32H60M18 18 46 46M46 18 18 46" />
        <path class="customer-details__gem-outline" d="M32 4 60 32 32 60 4 32Z" pathLength="100" />
        <g :clip-path="`url(#${gemId}-clip)`">
          <g transform="skewX(-22)">
            <rect
              class="customer-details__gem-glint"
              x="-4"
              y="0"
              width="18"
              height="64"
              :fill="`url(#${gemId}-glint)`"
            />
          </g>
        </g>
      </svg>
      <i
        v-for="spark in 8"
        :key="spark"
        class="customer-details__spark"
        :style="{
          '--spark-angle': `${(spark - 1) * 45}deg`,
          '--spark-delay': `${0.7 + (spark % 3) * 0.06}s`,
        }"
      ></i>
    </span>
    <p class="pb-eyebrow">
      {{
        confirmState === 'confirmed'
          ? 'Reserva completa'
          : reviewing
            ? 'Última revisión'
            : 'Casi listo'
      }}
    </p>
    <h1 class="pb-title">{{ confirmState === 'confirmed' ? 'Todo listo' : 'Tus datos' }}</h1>
    <div class="pb-rule" aria-hidden="true"></div>

    <section
      class="pb-sheet customer-details"
      :class="{ 'customer-details--ticket': confirmState === 'confirmed' }"
    >
      <form
        v-if="!reviewing"
        ref="formRef"
        class="pb-sheet__form customer-details__rise"
        novalidate
        @submit.prevent="handleSubmit"
      >
        <!-- Los campos viven en su propia región: con poco alto es lo único que
             se desplaza y el botón de enviar queda siempre a la vista. -->
        <div class="customer-details__fields">
          <fieldset class="customer-details__fieldset customer-details__span">
            <legend class="customer-details__legend">¿Para quién es el turno?</legend>
            <div
              class="customer-details__radio-group"
              role="radiogroup"
              aria-label="¿Para quién es el turno?"
            >
              <label class="customer-details__radio">
                <input v-model="attendeeChoice" type="radio" name="attendee-choice" value="self" />
                <span>Para mí</span>
              </label>
              <label class="customer-details__radio">
                <input v-model="attendeeChoice" type="radio" name="attendee-choice" value="other" />
                <span>Para otra persona</span>
              </label>
            </div>
          </fieldset>

          <fieldset
            v-if="offersHomeService"
            class="customer-details__fieldset customer-details__span"
          >
            <legend class="customer-details__legend">¿Dónde te atendemos?</legend>
            <div
              class="customer-details__radio-group"
              role="radiogroup"
              aria-label="¿Dónde te atendemos?"
            >
              <label class="customer-details__radio">
                <input v-model="serviceMode" type="radio" name="service-mode" value="local" />
                <span>En el local</span>
              </label>
              <label class="customer-details__radio">
                <input v-model="serviceMode" type="radio" name="service-mode" value="home" />
                <span>A domicilio</span>
              </label>
            </div>
            <HomeLocationField
              v-if="serviceMode === 'home'"
              v-model="homePlace"
              class="customer-details__address customer-details__rise"
              :start="homeMapStart"
              :error="attempted ? fieldErrors.home : undefined"
            />
          </fieldset>

          <BaseInput
            v-model="fullName"
            :class="{ 'customer-details__span': attendeeChoice === 'self' }"
            type="text"
            label="Tu nombre"
            autocomplete="name"
            required
            :error="attempted ? fieldErrors.fullName : undefined"
          />

          <BaseInput
            v-if="attendeeChoice === 'other'"
            v-model="attendeeName"
            class="customer-details__rise"
            type="text"
            label="Nombre de la persona atendida"
            autocomplete="off"
            required
            :maxlength="ATTENDEE_NAME_MAX_LENGTH"
            :error="attempted ? fieldErrors.attendeeName : undefined"
          />

          <PhoneField
            v-model:country="phoneCountryIso"
            v-model:national="phoneNational"
            required
            :error="attempted ? fieldErrors.phone : undefined"
          />

          <BaseInput
            v-model="email"
            type="email"
            label="Correo"
            autocomplete="email"
            required
            :error="attempted ? fieldErrors.email : undefined"
          />

          <div class="customer-details__field customer-details__span">
            <label for="customer-details-note" class="customer-details__label"
              >Nota (opcional)</label
            >
            <textarea
              id="customer-details-note"
              v-model="note"
              class="customer-details__textarea"
              rows="3"
              :aria-invalid="attempted && !!fieldErrors.note"
              :aria-describedby="
                attempted && fieldErrors.note
                  ? 'customer-details-note-error'
                  : 'customer-details-note-counter'
              "
            />
            <div class="customer-details__field-foot">
              <span
                v-if="attempted && fieldErrors.note"
                id="customer-details-note-error"
                class="customer-details__error"
                role="alert"
              >
                {{ fieldErrors.note }}
              </span>
              <span v-else />
              <span id="customer-details-note-counter" class="customer-details__counter">
                {{ note.length }}/{{ CUSTOMER_NOTE_MAX_LENGTH }}
              </span>
            </div>
          </div>
        </div>

        <BaseButton type="submit" variant="primary" size="lg" class="customer-details__submit"
          >Ver resumen</BaseButton
        >
      </form>

      <!-- Resumen y confirmación real (HU-096/HU-097, CA-097-07): exige
           confirmación explícita antes de reservar (RN-DIS-03), nunca
           insinúa que el turno ya quedó reservado hasta que el servidor lo
           confirme. -->
      <div
        v-else-if="confirmState !== 'confirmed'"
        ref="summaryRef"
        class="customer-details__summary customer-details__rise"
        role="status"
        tabindex="-1"
      >
        <h2 class="pb-sheet__heading">Revisa tus datos</h2>

        <!-- La cita es lo que el cliente viene a comprobar: va arriba, grande,
             como el talón de un turno. -->
        <div class="customer-details__when">
          <span class="customer-details__when-label">Tu turno</span>
          <span class="customer-details__when-time">{{ formatLocalTime(selectedStartsAt) }}</span>
          <span class="customer-details__when-day">{{ formatLocalDay(selectedStartsAt) }}</span>
          <span v-if="barber" class="customer-details__who">
            <BarberAvatar
              size="row"
              :full-name="barber.fullName"
              :photo-url="publicBarberPhotoUrl(slug, barber)"
            />
            <span class="customer-details__who-text">
              <span class="customer-details__when-label">{{ v.Professional }}</span>
              <span class="customer-details__who-name">{{ barber.fullName }}</span>
            </span>
          </span>
        </div>

        <!-- Región desplazable: con poco alto es lo único que se mueve y los
             botones quedan siempre a la vista. -->
        <!-- Foco y nombre: con la dirección a domicilio la región puede desbordar y, sin
             contenido enfocable dentro, el teclado no podría desplazarla. -->
        <div
          class="customer-details__summary-body"
          tabindex="0"
          role="region"
          aria-label="Resumen del turno"
        >
          <dl class="customer-details__summary-list">
            <div class="customer-details__summary-row">
              <dt>Cliente</dt>
              <dd>{{ fullName }}</dd>
            </div>
            <!-- Solo cuando el turno es de otra persona: para uno mismo repetiría
               el nombre del cliente. -->
            <div v-if="attendeeChoice === 'other'" class="customer-details__summary-row">
              <dt>Turno para</dt>
              <dd>{{ resolvedAttendeeName }}</dd>
            </div>
            <div class="customer-details__summary-row">
              <dt>Teléfono</dt>
              <dd>{{ phoneCountry(phoneCountryIso).dial }} {{ phoneNational }}</dd>
            </div>
            <div class="customer-details__summary-row">
              <dt>Correo</dt>
              <dd>{{ email }}</dd>
            </div>
            <div v-if="isHomeService && homePlace" class="customer-details__summary-row">
              <dt>Te atendemos en</dt>
              <dd>{{ homePlace.address }}</dd>
            </div>
            <div v-if="note.trim()" class="customer-details__summary-row">
              <dt>Nota</dt>
              <dd>{{ note }}</dd>
            </div>
          </dl>

          <!-- DEC-122: si el negocio publica su agenda en Google Calendar, Google invita al
             cliente al turno en el correo que dio. El aviso es condicional porque esta
             pantalla no sabe si el negocio lo usa; nunca promete una invitación. -->
          <p class="customer-details__invite-note">
            Si el negocio usa Google Calendar, recibirás una invitación a este turno en
            <strong>{{ email }}</strong
            >.
          </p>
        </div>

        <!-- Fuera de la región desplazable: dentro, el aviso quedaba bajo su borde
             visible y quien confirmó solo veía el botón activo de nuevo (F-RQ-01, #429). -->
        <BaseAlert v-if="confirmState === 'error'" variant="danger" role="alert">
          {{ errorMessage }}
        </BaseAlert>

        <div ref="actionsRef" class="customer-details__actions">
          <BaseButton
            type="button"
            variant="secondary"
            size="lg"
            :disabled="confirmState === 'submitting'"
            @click="editAgain"
          >
            Editar
          </BaseButton>
          <BaseButton
            type="button"
            variant="primary"
            size="lg"
            :loading="confirmState === 'submitting'"
            @click="handleConfirm"
          >
            Confirmar turno
          </BaseButton>
        </div>
      </div>

      <!-- Confirmación real del servidor (CA-097-01, CA-097-07): el correo
           con el enlace de acceso ya se envió; el token en claro solo
           existe en esta respuesta (DEC-089), nunca se vuelve a mostrar ni
           se dibuja aquí. Se presenta como un boleto de turno. -->
      <div
        v-else
        ref="confirmedRef"
        class="customer-details__confirmed ticket"
        role="status"
        tabindex="-1"
      >
        <p class="ticket__title">¡Tu turno quedó confirmado!</p>

        <article v-if="confirmedAppointment" class="ticket__card">
          <div class="ticket__main">
            <p class="ticket__shop">{{ confirmedAppointment.barbershopName }}</p>
            <p class="ticket__service">{{ confirmedAppointment.serviceName }}</p>
            <p class="ticket__meta">
              <span>{{ confirmedAppointment.durationMinutes }} min</span>
              <span aria-hidden="true">·</span>
              <span>{{
                formatMoney(confirmedAppointment.priceAmount, confirmedAppointment.currency)
              }}</span>
            </p>
            <span class="ticket__stamp" aria-hidden="true">Confirmado</span>
          </div>

          <div class="ticket__tear" aria-hidden="true"></div>

          <div class="ticket__stub">
            <div class="ticket__when">
              <span class="ticket__leaf" aria-hidden="true">
                <span class="ticket__leaf-month">{{
                  leafParts(confirmedAppointment.startsAt).month
                }}</span>
                <span class="ticket__leaf-day">{{
                  leafParts(confirmedAppointment.startsAt).day
                }}</span>
              </span>
              <span class="ticket__when-text">
                <span class="ticket__weekday">{{
                  leafParts(confirmedAppointment.startsAt).weekday
                }}</span>
                <span class="ticket__time">{{
                  formatLocalTime(confirmedAppointment.startsAt)
                }}</span>
                <span class="ticket__until">{{ untilLabel(confirmedAppointment.startsAt) }}</span>
              </span>
            </div>

            <div v-if="barber" class="ticket__who">
              <BarberAvatar
                size="row"
                :full-name="barber.fullName"
                :photo-url="publicBarberPhotoUrl(slug, barber)"
              />
              <span class="ticket__who-text">
                <span class="ticket__label">{{ v.Professional }}</span>
                <span class="ticket__who-name">{{ barber.fullName }}</span>
              </span>
            </div>
          </div>

          <p v-if="attendeeChoice === 'other'" class="ticket__for">
            <span class="ticket__label">Turno para</span>
            <span>{{ confirmedAppointment.attendeeName }}</span>
          </p>

          <span class="ticket__bars" aria-hidden="true"></span>
        </article>

        <p class="customer-details__note">
          Te enviamos un correo a <strong>{{ email }}</strong> con el enlace para consultar tu turno
          cuando quieras.
        </p>
      </div>
    </section>

    <!-- Horario ocupado: un diálogo modal y sin salida lateral (ni Escape ni
         clic fuera), para que el cliente no confirme sin haber revisado el
         nuevo turno. -->
    <BaseDialog
      :model-value="confirmState === 'schedule-conflict'"
      title="Ese horario se acaba de ocupar"
      size="sm"
      content-class="pb-conflict-dialog"
      :close-on-backdrop="false"
      :close-on-escape="false"
      :show-close="false"
    >
      <template #icon>
        <span class="pb-conflict__mark" aria-hidden="true">!</span>
      </template>

      <p class="pb-conflict__lead">
        Alguien reservó las <strong>{{ formatLocalTime(selectedStartsAt) }}</strong> del
        <strong>{{ formatLocalDay(selectedStartsAt) }}</strong> justo antes que tú. Elige otra hora
        para seguir con tu turno.
      </p>

      <template v-if="alternatives.length > 0">
        <p :id="conflictAltsLabelId" class="pb-conflict__label">Estas horas siguen libres</p>
        <ul class="pb-conflict__alts" :aria-labelledby="conflictAltsLabelId">
          <li v-for="alt in sortedAlternatives" :key="alt.startsAt">
            <button type="button" class="pb-conflict__alt" @click="chooseAlternative(alt.startsAt)">
              <span class="pb-conflict__alt-time">{{ formatLocalTime(alt.startsAt) }}</span>
              <span class="pb-conflict__alt-day">{{ formatLocalDayShort(alt.startsAt) }}</span>
            </button>
          </li>
        </ul>
      </template>
      <p v-else class="pb-conflict__lead">
        No quedan horarios libres cerca de ese. Elige otro día en el calendario.
      </p>

      <template #footer>
        <BaseButton
          type="button"
          :variant="alternatives.length > 0 ? 'secondary' : 'primary'"
          @click="pickAnotherSlot"
        >
          Elegir otro horario
        </BaseButton>
      </template>
    </BaseDialog>
  </main>
</template>

<style scoped>
/* Todo lo que sigue vive sobre el panel de tinta levantada (`.pb-sheet`), el
   mismo lienzo de las demás pantallas de la reserva: los campos, el resumen
   y la confirmación se recomponen con los tokens de tinta y latón. `BaseInput`
   y `BaseButton` se pensaron para superficies claras, así que aquí se les
   ajustan sus variables y colores sin tocar el componente compartido. */
.customer-details__rise {
  animation: pb-rise 0.55s var(--pb-ease, ease) 0.1s backwards;
}

/* Los campos del formulario entran escalonados, uno tras otro. */
.customer-details__fields > * {
  animation: pb-rise 0.6s var(--pb-ease, ease) calc(0.35s + var(--i, 0) * 0.07s) backwards;
}

.customer-details__fields > :nth-child(2) {
  --i: 1;
}

.customer-details__fields > :nth-child(3) {
  --i: 2;
}

.customer-details__fields > :nth-child(4) {
  --i: 3;
}

.customer-details__fields > :nth-child(5) {
  --i: 4;
}

.customer-details__fields > :nth-child(6) {
  --i: 5;
}

/* Rejilla de dos columnas: nombre y persona atendida, teléfono y correo van
   en pareja; el selector, la nota y el nombre a solas ocupan el ancho. */
.customer-details__fields {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--space-5) var(--space-4);
}

@media (min-width: 560px) {
  .customer-details__fields {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .customer-details__span {
    grid-column: 1 / -1;
  }
}

.customer-details__fieldset {
  min-width: 0;
  padding: 0;
  margin: 0;
  border: none;
}

/* Rótulo reglado: versalitas de latón, el mismo de BaseInput. */
.customer-details__legend,
.customer-details__label {
  padding: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

.customer-details__legend {
  margin: 0 0 var(--space-2);
}

/* ── Campos de BaseInput sobre tinta ───────────────────────────────────── */

.customer-details :deep(.base-input) {
  --input-bg: color-mix(in srgb, #000 46%, var(--color-surface-strong));
  --input-border-color: var(--pb-line-strong);
  --input-border-base-color: color-mix(in srgb, var(--pb-brass) 75%, transparent);
  --input-focus-ring:
    0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--pb-brass),
    0 0 22px color-mix(in srgb, var(--pb-brass) 35%, transparent);

  color: var(--pb-text);
}

.customer-details :deep(.base-input::placeholder) {
  color: var(--pb-muted);
  opacity: 1;
}

.customer-details :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  background-color: color-mix(in srgb, #000 38%, var(--color-surface-strong));
  border-color: color-mix(in srgb, var(--pb-brass) 60%, transparent);
  border-bottom-color: var(--pb-brass);
}

.customer-details :deep(.base-input:focus-visible) {
  background-color: color-mix(in srgb, #000 52%, var(--color-surface-strong));
  border-color: var(--pb-brass);
  border-bottom-color: var(--pb-brass);
}

.customer-details :deep(.base-input:-webkit-autofill),
.customer-details :deep(.base-input:-webkit-autofill:hover),
.customer-details :deep(.base-input:-webkit-autofill:focus) {
  -webkit-text-fill-color: var(--pb-text);
  caret-color: var(--pb-text);
}

.customer-details :deep(.base-input__label) {
  letter-spacing: 0.12em;
  color: var(--pb-brass);
}

.customer-details :deep(.base-input__required) {
  color: var(--color-danger-on-strong);
}

.customer-details :deep(.base-input--invalid),
.customer-details :deep(.base-input--invalid:hover) {
  border-color: var(--color-danger-on-strong);
  border-bottom-color: var(--color-danger-on-strong);
}

.customer-details :deep(.base-input--invalid:focus-visible) {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-danger-on-strong);
}

.customer-details :deep(.base-input__hint) {
  color: var(--pb-muted);
}

.customer-details :deep(.base-input__error),
.customer-details__error {
  color: var(--color-danger-on-strong);
}

/* «Para mí / Para otra persona» como control segmentado. El `<input>` nativo
   sigue siendo el control (foco, flechas, formulario); solo se oculta su
   caja y la etiqueta hace de segmento. */
.customer-details__radio-group {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(128px, 1fr));
  gap: var(--space-2);
}

.customer-details__radio {
  position: relative;
  display: flex;
  min-height: var(--control-height);
  cursor: pointer;
}

.customer-details__radio input {
  position: absolute;
  inset: 0;
  z-index: 1;
  width: 100%;
  height: 100%;
  margin: 0;
  cursor: pointer;
  opacity: 0;
}

.customer-details__radio span {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  padding: 0 var(--space-4);
  font-size: var(--font-size-body);
  font-weight: 500;
  color: var(--pb-text);
  text-align: center;
  background-color: transparent;
  border: var(--border-width-normal) solid var(--pb-line-strong);
  border-bottom: var(--border-width-emphasis) solid var(--pb-line-strong);
  border-radius: 2px;
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    color var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-base) var(--pb-ease, ease);
}

.customer-details__radio:hover input:not(:checked) + span {
  background-color: var(--pb-wash);
  border-color: var(--pb-brass);
  transform: translateY(-1px);
}

.customer-details__radio input:checked + span {
  font-weight: 600;
  color: var(--pb-on-brass);
  background-color: var(--pb-brass);
  border-color: var(--pb-brass);
  border-bottom-color: color-mix(in srgb, var(--pb-brass) 70%, #000);
}

/* El segmento elegido también lleva un rombo: no depende solo del relleno. */
.customer-details__radio input:checked + span::before {
  content: '';
  width: 7px;
  height: 7px;
  margin-right: var(--space-2);
  background-color: var(--pb-on-brass);
  transform: rotate(45deg);
  animation: pb-pop 0.4s cubic-bezier(0.3, 1.5, 0.5, 1);
}

.customer-details__radio input:focus-visible + span {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--pb-brass);
}

.customer-details__field {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

/* Área de nota: el mismo campo reglado que BaseInput (recuesto oscuro, filete
   perimetral y línea base de latón de 2px). */
.customer-details__textarea {
  padding: var(--space-3) var(--space-4);
  font-family: var(--font-family-base);
  font-size: var(--font-size-body);
  line-height: var(--font-size-body-line);
  color: var(--pb-text);
  height: calc(var(--font-size-body-line) * 3 + var(--space-3) * 2);
  resize: none;
  background-color: color-mix(in srgb, #000 46%, var(--color-surface-strong));
  border: var(--border-width-normal) solid var(--pb-line-strong);
  border-bottom: var(--border-width-emphasis) solid
    color-mix(in srgb, var(--pb-brass) 75%, transparent);
  border-radius: 2px;
  outline: none;
  transition:
    border-color var(--motion-duration-fast) var(--motion-easing-standard),
    box-shadow var(--motion-duration-fast) var(--motion-easing-standard),
    background-color var(--motion-duration-fast) var(--motion-easing-standard);
}

.customer-details__textarea:hover {
  background-color: color-mix(in srgb, #000 38%, var(--color-surface-strong));
  border-color: color-mix(in srgb, var(--pb-brass) 60%, transparent);
  border-bottom-color: var(--pb-brass);
}

.customer-details__textarea:focus-visible {
  background-color: color-mix(in srgb, #000 52%, var(--color-surface-strong));
  border-color: var(--pb-brass);
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--pb-brass),
    0 0 22px color-mix(in srgb, var(--pb-brass) 35%, transparent);
}

.customer-details__textarea[aria-invalid='true'] {
  border-color: var(--color-danger-on-strong);
  border-bottom-color: var(--color-danger-on-strong);
}

.customer-details__field-foot {
  display: flex;
  gap: var(--space-2);
  justify-content: space-between;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
}

.customer-details__counter {
  color: var(--pb-muted);
}

/* ── Botones de BaseButton en latón sobre tinta ────────────────────────── */

.customer-details :deep(.base-button) {
  font-weight: 600;
}

.customer-details :deep(.base-button--primary) {
  color: var(--pb-on-brass);
  background-color: var(--pb-brass);
  border-color: var(--pb-brass);
  border-bottom-width: var(--border-width-emphasis);
  border-bottom-color: color-mix(in srgb, var(--pb-brass) 70%, #000);
}

.customer-details :deep(.base-button--primary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--pb-brass) 86%, #fff);
  border-color: color-mix(in srgb, var(--pb-brass) 86%, #fff);
  border-bottom-color: color-mix(in srgb, var(--pb-brass) 70%, #000);
  box-shadow: 0 10px 24px -12px color-mix(in srgb, var(--pb-brass) 70%, transparent);
}

.customer-details :deep(.base-button--primary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--pb-brass) 88%, #000);
  border-color: color-mix(in srgb, var(--pb-brass) 88%, #000);
}

.customer-details :deep(.base-button--secondary),
.customer-details :deep(.base-button--soft) {
  color: var(--pb-brass);
  background-color: transparent;
  border-color: var(--pb-brass);
  border-bottom-width: var(--border-width-emphasis);
}

.customer-details :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)),
.customer-details :deep(.base-button--soft:hover:not(:disabled):not(.base-button--loading)) {
  background-color: var(--pb-wash-strong);
  border-color: var(--pb-brass);
}

.customer-details :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)),
.customer-details :deep(.base-button--soft:active:not(:disabled):not(.base-button--loading)) {
  background-color: var(--pb-wash-strong);
}

.customer-details :deep(.base-button:focus-visible) {
  box-shadow:
    0 0 0 2px var(--color-surface-strong),
    0 0 0 4px var(--color-on-strong);
}

.customer-details__submit {
  flex: none;
  width: 100%;
  margin-top: var(--space-2);
  animation: pb-rise 0.6s var(--pb-ease, ease) 0.8s backwards;
}

.customer-details__summary,
.customer-details__confirmed {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  outline: none;
}

.customer-details__summary-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.customer-details__summary .pb-sheet__heading {
  margin-bottom: 0;
}

/* Talón del turno: hora en grande sobre la fecha, con filete de latón. */
.customer-details__when {
  display: grid;
  gap: var(--space-1);
  padding: var(--space-4) var(--space-5);
  background-image: linear-gradient(
    to right,
    color-mix(in srgb, var(--pb-brass) 12%, transparent),
    transparent 80%
  );
  border: 1px solid var(--pb-line);
  border-left: 3px solid var(--pb-brass);
  border-radius: 2px;
}

.customer-details__when-label {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

/* Cifras en la sans: la serif de títulos dibuja el 1 como una «l». */
.customer-details__when-time {
  font-family: var(--font-sans);
  font-size: 36px;
  font-weight: 500;
  line-height: 44px;
  font-variant-numeric: tabular-nums;
  color: var(--pb-text);
}

.customer-details__when-day {
  color: var(--pb-soft);
}

/* El día llega en minúscula («viernes, 30 de octubre…»): solo la inicial. */
.customer-details__when-day::first-letter {
  text-transform: uppercase;
}

/* Barbero elegido: retrato y nombre al pie del talón, separados por un filete. */
.customer-details__who {
  display: flex;
  gap: var(--space-3);
  align-items: center;
  padding-top: var(--space-3);
  margin-top: var(--space-2);
  border-top: 1px solid var(--pb-line);
}

/* En ancho, el barbero pasa a la derecha del talón: la hora y él se leen a la
   vez y la hoja gana alto para los datos. */
@media (min-width: 560px) {
  .customer-details__when {
    grid-template-columns: 1fr auto;
    column-gap: var(--space-6);
  }

  .customer-details__when > :not(.customer-details__who) {
    grid-column: 1;
  }

  .pb-page--scroll-form .customer-details__who,
  .customer-details__who {
    grid-row: 1 / span 3;
    grid-column: 2;
    align-self: center;
    padding: 0 0 0 var(--space-6);
    margin: 0;
    border-top: 0;
    border-left: 1px solid var(--pb-line);
  }
}

.customer-details__who-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.customer-details__who-name {
  overflow-wrap: anywhere;
  font-weight: 500;
  color: var(--pb-text);
}

.customer-details__summary-list {
  display: flex;
  flex-direction: column;
  margin: 0;
  border-top: 1px solid var(--pb-line-strong);
}

.customer-details__invite-note {
  margin: 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--pb-muted);
  overflow-wrap: anywhere;
}

.customer-details__summary-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1) var(--space-6);
  justify-content: space-between;
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--pb-line);
  animation: pb-rise 0.5s var(--pb-ease, ease) backwards;
}

.customer-details__summary-row:nth-child(2) {
  animation-delay: 0.08s;
}

.customer-details__summary-row:nth-child(3) {
  animation-delay: 0.16s;
}

.customer-details__summary-row:nth-child(4) {
  animation-delay: 0.24s;
}

.customer-details__summary-row:nth-child(5) {
  animation-delay: 0.32s;
}

.customer-details__summary-row:nth-child(n + 6) {
  animation-delay: 0.4s;
}

.customer-details__summary-row dt {
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 24px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

.customer-details__summary-row dd {
  min-width: 0;
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--pb-text);
  text-align: right;
}

/* Avisos (error y conflicto) sobre tinta: baño tenue del color de estado y
   filete lateral levantado al lienzo oscuro. */
.customer-details :deep(.base-alert) {
  color: var(--pb-text);
  background-color: color-mix(
    in srgb,
    var(--color-danger-on-strong) 12%,
    var(--color-surface-strong)
  );
  border-left-color: var(--color-danger-on-strong);
}

.customer-details :deep(.base-alert__title),
.customer-details :deep(.base-alert__text) {
  color: var(--pb-text);
}

.customer-details :deep(.base-alert__status) {
  color: var(--pb-brass);
}

.customer-details__actions {
  display: grid;
  grid-template-columns: 1fr;
  gap: var(--space-3);
}

.customer-details__actions :deep(.base-button) {
  width: 100%;
}

@media (min-width: 480px) {
  .customer-details__actions {
    grid-template-columns: 1fr 1.6fr;
  }
}

.customer-details__note {
  max-width: 46ch;
  margin: 0;
  color: var(--pb-soft);
  text-align: center;
  overflow-wrap: anywhere;
}

.customer-details__note strong {
  font-weight: 600;
  color: var(--pb-text);
}

/* ── Confirmación: boleto de turno ───────────────────────────────────────── */

/* La hoja de papel deja paso al boleto: sin panel, sin cinta métrica. */
.pb-sheet.customer-details--ticket {
  padding: 0;
  margin-top: var(--space-6);
  background: none;
  border: 0;
  box-shadow: none;
  animation: none;
}

.customer-details--ticket::before {
  display: none;
}

.ticket {
  align-items: center;
  gap: var(--space-5);
}

/* Cabecera de la confirmación: sello, rótulo y título centrados sobre el boleto. */
.pb-page--confirmed {
  align-items: center;
}

.pb-page--confirmed .pb-eyebrow {
  justify-content: center;
  margin-top: var(--space-5);
}

.pb-page--confirmed .pb-eyebrow::after {
  content: '';
  width: 28px;
  height: 1px;
  background-color: currentcolor;
}

.pb-page--confirmed .pb-title {
  font-size: clamp(48px, 13vw, 76px);
  text-align: center;
}

.pb-page--confirmed .pb-rule {
  display: none;
}

.pb-page--confirmed .customer-details {
  width: 100%;
}

/* Halo de latón tras el sello. */
.pb-page--confirmed .customer-details__seal::before {
  content: '';
  position: absolute;
  inset: -34px;
  background-image: radial-gradient(
    circle,
    color-mix(in srgb, var(--color-success-on-strong) 24%, transparent),
    transparent 68%
  );
  animation: ticket-halo 1.4s ease-out 0.8s backwards;
}

.ticket__title {
  margin: 0;
  font-size: var(--font-size-body-lg, 18px);
  text-align: center;
  color: var(--pb-soft);
  animation: pb-rise 0.7s var(--pb-ease, ease) 0.9s backwards;
}

.ticket__card {
  position: relative;
  width: 100%;
  max-width: 560px;
  overflow: hidden;
  background-color: color-mix(in srgb, #000 26%, var(--color-surface-strong));
  background-image: linear-gradient(
    to bottom,
    color-mix(in srgb, var(--pb-brass) 7%, transparent),
    transparent 160px
  );
  border: 1px solid var(--pb-line);
  border-top: 2px solid var(--pb-brass);
  border-radius: 4px;
  box-shadow:
    0 36px 60px -30px rgb(0 0 0 / 80%),
    0 0 0 1px color-mix(in srgb, var(--pb-brass) 6%, transparent);
  transform-origin: 50% 0;
  animation: ticket-drop 0.95s cubic-bezier(0.2, 0.8, 0.2, 1) 1.1s backwards;
}

/* Un destello de latón cruza el boleto una sola vez al aterrizar. */
.ticket__card::after {
  content: '';
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image: linear-gradient(
    105deg,
    transparent 40%,
    color-mix(in srgb, var(--pb-brass) 22%, transparent) 50%,
    transparent 60%
  );
  transform: translateX(-110%);
  animation: ticket-shine 1.3s ease-in-out 1.9s;
}

.ticket__main {
  position: relative;
  display: grid;
  gap: var(--space-1);
  padding: var(--space-6) var(--space-6) var(--space-5);
}

.ticket__shop,
.ticket__label {
  margin: 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--pb-brass);
}

.ticket__service {
  margin: var(--space-1) 0 0;
  font-family: var(--font-display);
  font-size: clamp(30px, 8vw, 40px);
  line-height: 1.05;
  color: var(--pb-text);
  overflow-wrap: anywhere;
}

.ticket__meta {
  display: flex;
  gap: var(--space-2);
  margin: var(--space-2) 0 0;
  color: var(--pb-soft);
  font-variant-numeric: tabular-nums;
}

/* Sello de tinta: cae con un golpe y se queda un poco ladeado. */
.ticket__stamp {
  position: absolute;
  right: var(--space-6);
  bottom: var(--space-5);
  padding: 3px var(--space-3);
  font-family: var(--font-sans);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--color-success-on-strong);
  border: 2px solid currentcolor;
  border-radius: 3px;
  outline: 1px solid color-mix(in srgb, currentcolor 45%, transparent);
  outline-offset: 2px;
  opacity: 0.92;
  transform: rotate(-7deg);
  animation: ticket-stamp 0.5s cubic-bezier(0.3, 1.4, 0.5, 1) 2.1s backwards;
}

/* Perforación: línea de puntos con una muesca a cada lado. */
.ticket__tear {
  position: relative;
  height: 0;
  border-top: 2px dashed var(--pb-line-strong);
}

.ticket__tear::before,
.ticket__tear::after {
  content: '';
  position: absolute;
  top: -12px;
  width: 22px;
  height: 22px;
  background-color: var(--color-surface-strong);
  border: 1px solid var(--pb-line);
  border-radius: 50%;
}

.ticket__tear::before {
  left: -12px;
}

.ticket__tear::after {
  right: -12px;
}

.ticket__stub {
  display: grid;
  gap: var(--space-5);
  padding: var(--space-6);
}

.ticket__when {
  display: flex;
  gap: var(--space-4);
  align-items: center;
}

/* Hoja de calendario: banda de latón con el mes y el día en grande. */
.ticket__leaf {
  display: grid;
  flex: none;
  width: 68px;
  overflow: hidden;
  text-align: center;
  border: 1px solid var(--pb-brass);
  border-radius: 3px;
  animation: pb-pop 0.6s cubic-bezier(0.3, 1.5, 0.5, 1) 1.7s backwards;
}

.ticket__leaf-month {
  padding: 3px 0;
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: var(--pb-on-brass);
  background-color: var(--pb-brass);
}

.ticket__leaf-day {
  padding: 2px 0 6px;
  font-family: var(--font-display);
  font-size: 38px;
  line-height: 44px;
  color: var(--pb-text);
}

.ticket__when-text {
  display: grid;
  gap: 2px;
  justify-items: start;
}

.ticket__weekday {
  color: var(--pb-soft);
  text-transform: capitalize;
}

.ticket__time {
  font-family: var(--font-sans);
  font-size: 30px;
  font-weight: 500;
  line-height: 36px;
  font-variant-numeric: tabular-nums;
  color: var(--pb-text);
}

.ticket__until {
  padding: 2px var(--space-2);
  margin-top: var(--space-1);
  font-family: var(--font-sans);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--pb-brass);
  border: 1px solid color-mix(in srgb, var(--pb-brass) 60%, transparent);
  border-radius: 2px;
}

.ticket__who {
  display: flex;
  gap: var(--space-3);
  align-items: center;
  min-width: 0;
}

.ticket__who-text {
  display: grid;
  gap: 2px;
  min-width: 0;
}

.ticket__who-name {
  font-weight: 500;
  color: var(--pb-text);
  overflow-wrap: anywhere;
}

.ticket__for {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-1) var(--space-4);
  align-items: baseline;
  justify-content: space-between;
  padding: var(--space-3) var(--space-6);
  margin: 0;
  color: var(--pb-text);
  border-top: 1px solid var(--pb-line);
}

/* Franja de barras decorativa al pie, como el código de un boleto. No codifica
   nada: la credencial del turno nunca se dibuja (DEC-089). */
.ticket__bars {
  display: block;
  height: 26px;
  margin: 0 var(--space-6) var(--space-5);
  opacity: 0.4;
  background-image: repeating-linear-gradient(
    90deg,
    var(--pb-text) 0 2px,
    transparent 2px 5px,
    var(--pb-text) 5px 6px,
    transparent 6px 10px,
    var(--pb-text) 10px 13px,
    transparent 13px 15px,
    var(--pb-text) 15px 16px,
    transparent 16px 21px
  );
}

@media (min-width: 560px) {
  .ticket__stub {
    grid-template-columns: 1fr auto;
    align-items: center;
  }

  .ticket__who {
    padding-left: var(--space-6);
    border-left: 1px solid var(--pb-line);
  }
}

/* Ventanas bajas: el boleto se compacta para que quepa sin desplazar la página. */
@media (max-height: 900px) {
  .pb-page--confirmed {
    padding-top: var(--space-4);
    padding-bottom: var(--space-6);
  }

  .pb-page--confirmed .customer-details__seal {
    width: 72px;
    height: 72px;
  }

  .pb-page--confirmed .pb-eyebrow {
    margin-top: var(--space-2);
  }

  .pb-page--confirmed .pb-title {
    font-size: clamp(40px, 11vw, 56px);
  }

  .pb-page--confirmed .customer-details--ticket {
    margin-top: var(--space-4);
  }

  .ticket {
    gap: var(--space-3);
  }

  .ticket__main {
    padding: var(--space-5) var(--space-5) var(--space-4);
  }

  .ticket__stub {
    gap: var(--space-4);
    padding: var(--space-4) var(--space-5);
  }

  .ticket__for {
    padding: var(--space-2) var(--space-5);
  }

  .ticket__bars {
    display: none;
  }

  .ticket__stub {
    padding-bottom: var(--space-5);
  }
}

@media (max-height: 800px) {
  .pb-page--confirmed {
    padding-top: var(--space-2);
    padding-bottom: var(--space-4);
  }

  .pb-page--confirmed .customer-details__seal {
    width: 56px;
    height: 56px;
  }

  .pb-page--confirmed .pb-title {
    font-size: clamp(36px, 10vw, 48px);
  }

  .pb-page--confirmed .customer-details--ticket {
    margin-top: var(--space-2);
  }

  .ticket__main {
    padding-top: var(--space-4);
  }

  .ticket__stub {
    padding-top: var(--space-3);
    padding-bottom: var(--space-4);
  }
}

@keyframes ticket-drop {
  from {
    opacity: 0;
    transform: translateY(-34px) rotate(-1.6deg) scale(0.97);
  }
  60% {
    opacity: 1;
    transform: translateY(6px) rotate(0.4deg) scale(1);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes ticket-stamp {
  from {
    opacity: 0;
    transform: rotate(-16deg) scale(2.2);
  }
  to {
    opacity: 0.92;
    transform: rotate(-7deg) scale(1);
  }
}

@keyframes ticket-shine {
  to {
    transform: translateX(110%);
  }
}

@keyframes ticket-halo {
  from {
    opacity: 0;
    transform: scale(0.4);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

/* Sello de confirmación: el rombo se traza, el check se dibuja y ocho
   chispas de latón salen disparadas. Después queda quieto. */
.customer-details__seal {
  position: relative;
  display: block;
  align-self: center;
  width: 104px;
  height: 104px;
}

/* Piedra verde de «confirmado»: la misma gema tallada del cargador NAVA
   (contorno, rombo interior, facetas y un destello), pero en verde y sin
   repetirse. El contorno se talla una vez, aparecen las facetas, cruza el
   destello y la piedra queda quieta. */
.customer-details__gem {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: visible;
}

.customer-details__gem-outline {
  stroke: var(--color-success-on-strong);
  stroke-width: 2;
  stroke-linejoin: miter;
  stroke-dasharray: 100;
  animation: gem-draw 0.95s cubic-bezier(0.55, 0, 0.2, 1) 0.15s backwards;
}

.customer-details__gem-facets {
  stroke: var(--color-success-on-strong);
  stroke-width: 0.8;
  opacity: 0.5;
  animation: gem-facets 0.6s ease-out 0.85s backwards;
}

.customer-details__gem-inner {
  fill: color-mix(in srgb, var(--color-success-on-strong) 30%, transparent);
  stroke: color-mix(in srgb, var(--color-success-on-strong) 75%, transparent);
  stroke-width: 1;
  transform-box: fill-box;
  transform-origin: center;
  animation: gem-inner 0.7s cubic-bezier(0.3, 1.4, 0.5, 1) 0.8s backwards;
}

.customer-details__gem-glint {
  transform: translateX(96px);
  animation: gem-glint 1s cubic-bezier(0.5, 0, 0.3, 1) 1.5s backwards;
}

.customer-details__spark {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 5px;
  height: 5px;
  margin: -2.5px 0 0 -2.5px;
  background-color: var(--color-success-on-strong);
  opacity: 0;
  transform: rotate(var(--spark-angle)) translateY(0) rotate(45deg);
  animation: seal-spark 1s cubic-bezier(0.2, 0.7, 0.2, 1) var(--spark-delay, 0.8s) backwards;
}

@keyframes gem-draw {
  from {
    stroke-dashoffset: 100;
  }
  to {
    stroke-dashoffset: 0;
  }
}

@keyframes gem-facets {
  from {
    opacity: 0;
  }
}

@keyframes gem-inner {
  from {
    opacity: 0;
    transform: scale(0.2);
  }
}

@keyframes gem-glint {
  from {
    transform: translateX(-40px);
  }
}

@keyframes seal-spark {
  0% {
    opacity: 1;
    transform: rotate(var(--spark-angle)) translateY(0) rotate(45deg) scale(1);
  }
  100% {
    opacity: 0;
    transform: rotate(var(--spark-angle)) translateY(-74px) rotate(45deg) scale(0.4);
  }
}

@media (prefers-reduced-motion: reduce) {
  .customer-details__rise,
  .customer-details__fields > *,
  .customer-details__submit,
  .customer-details__summary-row,
  .customer-details__gem-outline,
  .customer-details__gem-facets,
  .customer-details__gem-inner,
  .customer-details__gem-glint,
  .customer-details__radio input:checked + span::before {
    animation: none;
  }

  .customer-details__spark,
  .ticket__card::after {
    display: none;
  }

  .ticket__card,
  .ticket__title,
  .ticket__leaf,
  .ticket__stamp,
  .pb-page--confirmed .customer-details__seal::before {
    animation: none;
  }

  .customer-details__radio span,
  .customer-details__textarea {
    transition: none;
  }

  .customer-details__radio:hover input:not(:checked) + span {
    transform: none;
  }
}
</style>
