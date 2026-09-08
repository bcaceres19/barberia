/*
 * Modelo de contenido de los mockups de `/panel/turnos/:appointmentId`
 * (detalle, historial y reprogramación).
 *
 * Mismo criterio que `tools/mockups/auth-eventos`,
 * `tools/mockups/panel-agenda-eventos` y `tools/mockups/nuevo-turno-eventos`:
 * cada evento se declara una sola vez y se renderiza en los dos viewports, de
 * modo que escritorio y móvil no puedan divergir en copy, orden ni estado.
 *
 * Los textos provienen del módulo real
 * `apps/web/src/modules/agenda/pages/AppointmentDetailPage.vue` (HU-064) y de
 * `apps/web/src/modules/agenda/model/appointmentDetail.ts` /
 * `model/dailyAgenda.ts`. Ningún texto se inventa: si una cadena no existe en
 * el código, procede del contrato de error real del backend o de una `DEC-*`
 * registrada.
 */

// Copys canónicos tomados del código real.
export const COPY = {
  back: 'Volver a la agenda',

  loadingDetail: 'Cargando el detalle del turno…',
  notFoundTitle: 'Este turno ya no está disponible',
  notFoundBody: 'Puede haberse eliminado o pertenecer a otra barbería.',
  loadErrorTitle: 'No pudimos cargar este turno',
  loadErrorBody: 'Revisa tu conexión e inténtalo de nuevo.',
  retry: 'Reintentar',

  reschedule: 'Reprogramar turno',

  // Rótulos del `<dl>` real, en el orden del template.
  factTime: 'Hora',
  factAttendee: 'Persona atendida',
  factService: 'Servicio',
  factBarber: 'Barbero',
  factCustomer: 'Cliente que reservó',
  factContact: 'Contacto',
  factNote: 'Nota del cliente',
  // El template anexa «· Zona X» dentro del mismo `dd`. Aquí baja a su propia
  // línea, así que el punto medio que la separaba del rango sobra.
  zoneSuffix: (tz) => `Zona ${tz}`,

  historyTitle: 'Historial',
  historyLoading: 'Cargando historial…',
  historyErrorTitle: 'No pudimos cargar el historial',
  historyErrorBody: 'Revisa tu conexión e inténtalo de nuevo.',
  historyEmpty: 'Sin eventos registrados todavía.',
  loadMore: 'Cargar más',

  // Diálogo de reprogramación (HU-065, T2).
  dialogTitle: 'Reprogramar turno',
  dialogCurrent: (range) => `Horario actual: ${range}`,
  dialogDateLabel: 'Nueva fecha',
  dialogTimeLabel: 'Nueva hora',
  dialogPreview: (from, to) => `Nuevo horario aproximado: ${from} – ${to}`,
  cancel: 'Cancelar',
  confirm: 'Confirmar',
  reload: 'Recargar',

  // `RescheduleStatus` del código real.
  versionConflictTitle: 'Este turno cambió mientras lo editabas',
  versionConflictBody:
    'Alguien más lo modificó. Recarga para ver los datos vigentes antes de reprogramar.',
  invalidStateTitle: 'Este turno ya no se puede reprogramar',
  invalidStateBody: 'Su estado cambió mientras lo editabas.',
  conflictTitle: 'No pudimos reprogramar el turno',
  networkErrorTitle: 'No pudimos conectar',
  networkErrorBody: 'Revisa tu conexión e inténtalo de nuevo. No perdiste lo que elegiste.',

  // `detail` real de un 409 de agenda: apperr.Conflict en
  // apps/api/internal/modules/booking/postgres/repository.go.
  conflictDetail: 'el barbero ya tiene una cita en ese intervalo',
}

// apps/web/src/modules/agenda/model/dailyAgenda.ts
export const STATUS_LABELS = {
  confirmed: 'Confirmado',
  completed: 'Completado',
  cancelled_by_customer: 'Cancelado por el cliente',
  cancelled_by_barber: 'Cancelado por el barbero',
  no_show: 'No se presentó',
}

// APPOINTMENT_STATUS_BADGE_VARIANT del mismo archivo.
export const STATUS_VARIANT = {
  confirmed: 'info',
  completed: 'success',
  cancelled_by_customer: 'neutral',
  cancelled_by_barber: 'danger',
  no_show: 'warning',
}

// apps/web/src/modules/agenda/model/appointmentDetail.ts
export const HISTORY_EVENT_LABELS = {
  appointment_created: 'Turno creado',
  appointment_rescheduled: 'Turno reprogramado',
  appointment_service_changed: 'Servicio cambiado',
  appointment_completed: 'Turno completado',
  appointment_cancelled_by_customer: 'Cancelado por el cliente',
  appointment_cancelled_by_barber: 'Cancelado por el barbero',
  appointment_no_show: 'Cliente no se presentó',
  appointment_status_corrected: 'Estado corregido',
}

export const HISTORY_FIELD_LABELS = {
  status: 'Estado',
  starts_at: 'Hora de inicio',
  ends_at: 'Hora de fin',
  barber_id: 'Barbero',
  service_id: 'Servicio',
  service_name_snapshot: 'Servicio',
  duration_minutes_snapshot: 'Duración',
  price_amount_snapshot: 'Precio',
  cancellation_reason: 'Motivo de cancelación',
}

// Datos ficticios (RN-DAT-01/02): ningún dato real de barbería o cliente.
const SHOP = 'NAVA QA Local'
const TZ = 'America/Bogota'

/*
 * El turno es el mismo Mateo Rojas de las 09:00 que ya aparece en la agenda de
 * `panel-agenda-eventos`, y su barbero es el mismo Julián Rodríguez de aquel
 * atlas: quien llega aquí lo hace tocando esa fila, y «Volver a la agenda»
 * regresa exactamente a ese día y barbero (HU-063).
 */
const ATTENDEE = 'Mateo Rojas'
const CUSTOMER = 'Mateo Rojas'
const BARBER = 'Julián Rodríguez'
const SERVICE = 'Corte clásico'
const DURATION = 45
const PRICE = '28.000'
const CURRENCY = 'COP'
const PHONE = '+57 321 456 7890'
const EMAIL = 'mateo.rojas@email.com'
const NOTE = 'Prefiere degradado bajo en los laterales.'

/*
 * Instantes en UTC, como los devuelve la API. Todas las cadenas visibles se
 * derivan de ellos con el MISMO `Intl.DateTimeFormat` que usan
 * `formatInstantInTimezone` (`es-CO`, `dateStyle: 'medium'` +
 * `timeStyle: 'short'`) y su variante de solo hora, así que ningún rótulo del
 * atlas puede desalinearse del que produce el código. `render.mjs` vuelve a
 * calcularlos dentro de Chromium y aborta si alguno difiere.
 */
const STARTS_AT = '2026-09-04T14:00:00Z' // 09:00 en America/Bogota
const ENDS_AT = '2026-09-04T14:45:00Z' // 09:45
const NEW_STARTS_AT = '2026-09-04T16:30:00Z' // 11:30
const NEW_ENDS_AT = '2026-09-04T17:15:00Z' // 12:15
const PREV_STARTS_AT = '2026-09-04T13:00:00Z' // 08:00
const PREV_ENDS_AT = '2026-09-04T13:45:00Z' // 08:45
const FIRST_STARTS_AT = '2026-09-04T20:00:00Z' // 15:00
const FIRST_ENDS_AT = '2026-09-04T20:45:00Z' // 15:45

const instant = (iso) =>
  new Intl.DateTimeFormat('es-CO', {
    timeZone: TZ,
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(iso))

const clock = (iso) =>
  new Intl.DateTimeFormat('es-CO', { timeZone: TZ, timeStyle: 'short' }).format(new Date(iso))

/*
 * `timeRangeLabel` del código: instante completo de inicio, guion largo y solo
 * la hora del fin. La zona se anexa aparte, como hace el template.
 */
const range = (from, to) => `${instant(from)} – ${clock(to)}`

// Los instantes que el generador verifica contra Chromium.
export const INSTANTS = {
  timezone: TZ,
  values: {
    start: { iso: STARTS_AT, style: 'instant', text: instant(STARTS_AT) },
    end: { iso: ENDS_AT, style: 'clock', text: clock(ENDS_AT) },
    newStart: { iso: NEW_STARTS_AT, style: 'instant', text: instant(NEW_STARTS_AT) },
    newEnd: { iso: NEW_ENDS_AT, style: 'clock', text: clock(NEW_ENDS_AT) },
    prevStart: { iso: PREV_STARTS_AT, style: 'instant', text: instant(PREV_STARTS_AT) },
    prevEnd: { iso: PREV_ENDS_AT, style: 'instant', text: instant(PREV_ENDS_AT) },
    firstStart: { iso: FIRST_STARTS_AT, style: 'instant', text: instant(FIRST_STARTS_AT) },
    firstEnd: { iso: FIRST_ENDS_AT, style: 'instant', text: instant(FIRST_ENDS_AT) },
  },
}

/** Turno vigente. `status` gobierna la insignia y la acción disponible. */
const appointment = (over = {}) => ({
  attendeeName: ATTENDEE,
  customerFullName: CUSTOMER,
  barberFullName: BARBER,
  serviceName: SERVICE,
  durationMinutes: DURATION,
  priceAmount: PRICE,
  currency: CURRENCY,
  customerPhone: PHONE,
  customerEmail: EMAIL,
  customerNote: NOTE,
  status: 'confirmed',
  timeRange: range(STARTS_AT, ENDS_AT),
  timezone: TZ,
  ...over,
})

const RESCHEDULED = appointment({ timeRange: range(NEW_STARTS_AT, NEW_ENDS_AT) })

/*
 * Historial en el orden real de la API (`ORDER BY occurred_at, id`): del
 * evento más antiguo al más reciente. Los campos de cada evento salen en el
 * orden real de `loadHistoryChanges` (`ORDER BY history_id, field_name`), por
 * eso «Hora de fin» aparece ANTES que «Hora de inicio»: es alfabético
 * (`ends_at` < `starts_at`), no cronológico.
 */
const CREATED = {
  eventType: 'appointment_created',
  actorLabel: CUSTOMER,
  occurredAt: instant('2026-09-02T15:30:00Z'),
  changes: [],
}

const RESCHEDULE_1 = {
  eventType: 'appointment_rescheduled',
  actorLabel: BARBER,
  occurredAt: instant('2026-09-02T22:40:00Z'),
  changes: [
    { field: 'ends_at', from: instant(FIRST_ENDS_AT), to: instant(PREV_ENDS_AT) },
    { field: 'starts_at', from: instant(FIRST_STARTS_AT), to: instant(PREV_STARTS_AT) },
  ],
}

const RESCHEDULE_2 = {
  eventType: 'appointment_rescheduled',
  actorLabel: BARBER,
  occurredAt: instant('2026-09-03T21:05:00Z'),
  changes: [
    { field: 'ends_at', from: instant(PREV_ENDS_AT), to: instant(ENDS_AT) },
    { field: 'starts_at', from: instant(PREV_STARTS_AT), to: instant(STARTS_AT) },
  ],
}

// Evento que agrega la reprogramación recién confirmada (evento 16).
const RESCHEDULE_3 = {
  eventType: 'appointment_rescheduled',
  actorLabel: BARBER,
  occurredAt: instant('2026-09-04T13:12:00Z'),
  changes: [
    { field: 'ends_at', from: instant(ENDS_AT), to: instant(NEW_ENDS_AT) },
    { field: 'starts_at', from: instant(STARTS_AT), to: instant(NEW_STARTS_AT) },
  ],
}

const COMPLETED_EVENT = {
  eventType: 'appointment_completed',
  actorLabel: BARBER,
  occurredAt: instant('2026-09-04T14:52:00Z'),
  changes: [],
}

const CANCELLED_EVENT = {
  eventType: 'appointment_cancelled_by_customer',
  actorLabel: CUSTOMER,
  occurredAt: instant('2026-09-03T23:20:00Z'),
  reason: 'Viaje de trabajo.',
  changes: [],
}

const HISTORY_FULL = [CREATED, RESCHEDULE_1, RESCHEDULE_2]

/** Diálogo de reprogramación. `date`/`time` son los valores de los controles
 *  nativos (`type="date"` en es-CO y `type="time"` en 24 h). */
const dialog = (over = {}) => ({
  date: '04/09/2026',
  time: '09:00',
  previewFrom: '09:00',
  previewTo: '09:45',
  ...over,
})

// Diálogo con la nueva hora ya elegida (11:30): la vista previa la calcula
// `addMinutesToTimeString` sumando los 45 minutos del snapshot.
const EDITED = dialog({ time: '11:30', previewFrom: '11:30', previewTo: '12:15' })

export const SCREENS = {
  // Carga inicial del detalle: esqueleto con la geometría de la ficha y del
  // historial que van a llegar.
  '01-carga-detalle': {
    page: 'loading',
  },

  // 404: el turno se eliminó o pertenece a otra barbería (RN-TEN-01: un
  // único mensaje, sin distinguir la causa).
  '02-turno-no-disponible': {
    page: 'not-found',
  },

  // Fallo recuperable al cargar el detalle.
  '03-error-detalle': {
    page: 'load-error',
  },

  // Detalle completo de un turno confirmado con su historial cargado. Es la
  // referencia de la pantalla.
  '04-detalle-confirmado': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: HISTORY_FULL },
  },

  // Detalle ya resuelto, historial todavía cargando: la ficha no espera al
  // historial (son dos peticiones distintas).
  '05-historial-cargando': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'loading' },
  },

  // Turno sin eventos registrados todavía.
  '06-historial-vacio': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: [] },
  },

  // Falla la carga del historial: la ficha sigue completa y solo la columna
  // del historial ofrece reintentar.
  '07-error-historial': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'error' },
  },

  // Historial paginado: la primera página no agota el registro y el eje
  // continúa hacia «Cargar más».
  '08-historial-paginado': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: [CREATED, RESCHEDULE_1], more: true },
  },

  // Terminal: completado. No hay ninguna acción propia (T3 y la corrección
  // de estado no existen), así que el encabezado queda sin botón.
  '09-turno-completado': {
    page: 'detail',
    appointment: appointment({ status: 'completed' }),
    history: { status: 'ready', items: [CREATED, RESCHEDULE_1, RESCHEDULE_2, COMPLETED_EVENT] },
  },

  // Terminal: cancelado por el cliente. El motivo del evento se muestra tal
  // como llega en `reason`.
  '10-turno-cancelado': {
    page: 'detail',
    appointment: appointment({ status: 'cancelled_by_customer' }),
    history: { status: 'ready', items: [CREATED, RESCHEDULE_1, CANCELLED_EVENT] },
  },

  // Diálogo recién abierto: fecha y hora vienen precargadas con el horario
  // vigente, así que la vista previa repite ese mismo horario.
  '11-dialogo-reprogramar': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: HISTORY_FULL },
    dialog: dialog(),
  },

  // Envío en curso: campos deshabilitados y «Confirmar» bloqueado.
  '12-dialogo-guardando': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: HISTORY_FULL },
    dialog: { ...EDITED, saving: true },
  },

  // 409 de agenda: el `detail` del problema es el texto real del backend.
  '13-dialogo-conflicto-agenda': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: HISTORY_FULL },
    dialog: {
      ...EDITED,
      alert: { variant: 'danger', title: COPY.conflictTitle, body: COPY.conflictDetail },
    },
  },

  // Versión obsoleta (412/`version-conflict`): la única salida es recargar el
  // detalle vigente, así que la alerta lleva su propia acción.
  '14-dialogo-version-obsoleta': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: HISTORY_FULL },
    dialog: {
      ...EDITED,
      alert: {
        variant: 'warning',
        title: COPY.versionConflictTitle,
        body: COPY.versionConflictBody,
        action: COPY.reload,
      },
    },
  },

  // Sin conexión: lo elegido se conserva, tal como dice el propio mensaje.
  '15-dialogo-sin-conexion': {
    page: 'detail',
    appointment: appointment(),
    history: { status: 'ready', items: HISTORY_FULL },
    dialog: {
      ...EDITED,
      alert: {
        variant: 'warning',
        title: COPY.networkErrorTitle,
        body: COPY.networkErrorBody,
      },
    },
  },

  // Éxito: el diálogo se cierra y el detalle se recarga. La confirmación es
  // el dato nuevo —hora vigente y un evento más en el historial—, no una
  // alerta: el código no emite ninguna.
  '16-reprogramacion-aplicada': {
    page: 'detail',
    appointment: RESCHEDULED,
    history: { status: 'ready', items: [...HISTORY_FULL, RESCHEDULE_3] },
  },
}

// Destinos reales del dock (mismos que los atlas hermanos).
export const NAV_PRIMARY = [
  { label: 'Agenda', icon: 'agenda' },
  { label: 'Servicios', icon: 'servicios' },
  { label: 'Barberos', icon: 'barberos' },
]
export const NAV_SECONDARY = [
  { label: 'Horarios', icon: 'horarios' },
  { label: 'Bloqueos', icon: 'bloqueos' },
  { label: 'Configuración', icon: 'configuracion' },
  { label: 'Servicios por barbero', icon: 'servicios-barbero' },
]

export const SHOP_NAME = SHOP
