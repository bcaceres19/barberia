/*
 * Modelo de contenido de los mockups de `/panel/turnos/nuevo` (Nuevo turno).
 *
 * Mismo criterio que `tools/mockups/auth-eventos` y
 * `tools/mockups/panel-agenda-eventos`: cada evento se declara una sola vez
 * y se renderiza en los dos viewports, de modo que escritorio y móvil no
 * puedan divergir en copy, orden ni estado.
 *
 * Los textos provienen del módulo real
 * `apps/web/src/modules/agenda/pages/NewAppointmentPage.vue` (HU-061) y de
 * `apps/web/src/modules/agenda/validation/appointmentValidation.ts`.
 * Ningún texto se inventa: si una cadena no existe en el código, procede de
 * una `DEC-*` registrada.
 */

// Copys canónicos tomados del código real.
export const COPY = {
  title: 'Nuevo turno',
  timezoneLine: (tz) => `Horas en la zona horaria de la barbería: ${tz}`,
  loadingBarbers: 'Cargando barberos…',
  loadErrorTitle: 'No pudimos cargar esta sección',
  loadErrorBody: 'Revisa tu conexión e inténtalo de nuevo.',
  retry: 'Reintentar',
  // Misma cadena del código, partida por su punto: titular en serif y
  // cuerpo, igual que el evento `04` de `panel-agenda-eventos`.
  noBarbersTitle: 'Aún no tienes barberos registrados.',
  noBarbersBody: 'Agrega uno en la sección "Barberos" antes de registrar turnos.',

  section1Title: 'Selecciona',
  section1Hint: 'Elige al barbero y el servicio.',
  barberLabel: 'Barbero',
  barberPlaceholder: 'Elige un barbero',
  serviceLabel: 'Servicio',
  servicePlaceholderNoBarber: 'Elige primero un barbero',
  servicePlaceholder: 'Elige un servicio',
  servicesLoading: 'Cargando servicios…',
  servicesEmpty: 'Este barbero no tiene servicios activos asignados.',
  servicesError: 'No pudimos cargar los servicios de este barbero.',

  section2Title: 'Persona atendida',
  section2Hint: 'Indica quién recibirá el servicio.',
  attendeeLabel: 'Persona atendida',
  customerNameLabel: 'Nombre del cliente',
  phoneLabel: 'Teléfono (opcional)',
  phonePlaceholder: '+573001234567',
  emailLabel: 'Correo (opcional)',
  noContactHint: 'Sin teléfono ni correo, el cliente no recibirá recordatorios.',

  section3Title: 'Fecha y hora',
  section3Hint: 'Define cuándo será el turno.',
  dateLabel: 'Fecha del turno',
  timeLabel: 'Hora del turno',

  section4Title: 'Nota (opcional)',
  section4Hint: 'Agrega información adicional si es necesario.',
  noteLabel: 'Nota (opcional)',
  // `validateCustomerNote` rechaza más de 500 caracteres: el contador del
  // mockup refleja ese límite real, no uno inventado.
  noteMax: 500,

  // Resumen real (`hasSummaryContent` + `<dl>` de la página): título
  // «Resumen» y solo estas cuatro entradas, cada una visible únicamente
  // cuando su dato ya está elegido.
  resumenTitle: 'Resumen',
  resumenBarber: 'Barbero',
  resumenService: 'Servicio',
  resumenAttendee: 'Persona atendida',
  resumenSchedule: 'Fecha y hora',

  submit: 'Registrar turno',
  submitting: 'Guardando…',

  // saveStatus del código real: 'conflict' | 'validation-error' |
  // 'idempotency-conflict' | 'not-found' | 'network-error' | 'unexpected-error'
  validationErrorTitle: 'Revisa los datos del turno.',
  validationErrorBody: 'Corrige los campos marcados para continuar.',
  startsAtConflict: 'Ese horario acaba de ocuparse.',
  conflictTitle: 'Ese horario acaba de ocuparse',
  conflictBody: 'Elige otra hora para continuar.',
  idempotencyConflict:
    'Este intento ya estaba en curso o cambió mientras se procesaba. Recarga la página e inténtalo de nuevo.',
  notFound: 'El barbero o el servicio elegidos ya no están disponibles.',
  networkError: 'No pudimos registrar el turno. Tus datos se conservaron; inténtalo de nuevo.',

  successTitle: 'Turno registrado',
  successSummary: (c) =>
    `${c.attendeeName} · ${c.serviceName} (${c.durationMinutes} min) · ${c.priceAmount} ${c.currency}`,
  successNote:
    'El turno quedó confirmado en la agenda del barbero. Las notificaciones al cliente todavía no están disponibles (B5 las agrega más adelante).',
  successAgain: 'Registrar otro turno',
}

// Datos ficticios (RN-DAT-01/02): ningún dato real de barbería o cliente.
const SHOP = 'NAVA QA Local'
const TZ = 'America/Bogota'
const BARBER = 'Camilo Torres'
const SERVICE = 'Corte clásico'
const ATTENDEE = 'Diego Molina'
const CUSTOMER = 'Diego Molina'
const PHONE = '+57 300 123 4567'
const EMAIL = 'diego.molina@email.com'
const NOTE = 'Cliente frecuente.'

/*
 * Fecha del turno de ejemplo. Es un día futuro respecto de la fecha del atlas
 * hermano (`panel-agenda-eventos` muestra «Hoy, 3 de septiembre»): registrar
 * un turno en el pasado contradiría la pantalla.
 *
 * `date` es lo que dibuja el control nativo `type="date"` en es-CO y `time`
 * el valor de `type="time"` (24 h), que es exactamente el que el resumen
 * concatena en el código. `dateFull` usa el mismo `Intl.DateTimeFormat` que
 * `formatCivilDateFull`, así que el rótulo del resumen no puede desalinearse
 * del día real.
 */
const CIVIL_DATE = '2026-09-04'
const DATE = '04/09/2026'
const TIME = '11:30'
const DATE_FULL = new Intl.DateTimeFormat('es-CO', { timeZone: 'UTC', dateStyle: 'full' }).format(
  new Date(`${CIVIL_DATE}T12:00:00Z`),
)

/** Estado por defecto de cada campo: vacío, salvo que el evento lo llene. */
const fields = (over = {}) => ({
  barber: '',
  service: '',
  attendee: '',
  customerName: '',
  phone: '',
  email: '',
  date: '',
  dateFull: '',
  time: '',
  note: '',
  ...over,
})

const FILLED = fields({
  barber: BARBER,
  service: SERVICE,
  attendee: ATTENDEE,
  customerName: CUSTOMER,
  phone: PHONE,
  email: EMAIL,
  date: DATE,
  dateFull: DATE_FULL,
  time: TIME,
  note: NOTE,
})

export const SCREENS = {
  // Carga inicial: todavía no se resolvieron los barberos.
  '01-carga-contexto': {
    page: 'loading',
  },

  // Fallo al cargar barberos/zona horaria.
  '02-error-contexto': {
    page: 'load-error',
  },

  // Sin barberos activos: no hay formulario que ofrecer.
  '03-sin-barberos': {
    page: 'empty',
  },

  // Formulario recién abierto: nada elegido todavía. Único evento sin
  // resumen, porque `hasSummaryContent` todavía es falso.
  '04-formulario-vacio': {
    page: 'form',
    timezone: TZ,
    fields: fields(),
  },

  // Barbero elegido: servicios cargando. El campo de servicio queda
  // deshabilitado mientras tanto (código real: `servicesStatus === 'loading'`).
  '05-cargando-servicios': {
    page: 'form',
    timezone: TZ,
    fields: fields({ barber: BARBER }),
    servicesStatus: 'loading',
  },

  // El barbero elegido no tiene servicios activos asignados.
  '06-sin-servicios-asignados': {
    page: 'form',
    timezone: TZ,
    fields: fields({ barber: BARBER }),
    servicesStatus: 'empty',
  },

  // Falla la carga de servicios de ese barbero.
  '07-error-servicios': {
    page: 'form',
    timezone: TZ,
    fields: fields({ barber: BARBER }),
    servicesStatus: 'error',
  },

  // Formulario completo, con el resumen ya cerrado antes del envío.
  '08-formulario-completo': {
    page: 'form',
    timezone: TZ,
    fields: FILLED,
  },

  // Envío intentado con campos inválidos: error por campo (aquí, barbero y
  // servicio sin elegir) más la alerta global de validación.
  '09-error-validacion': {
    page: 'form',
    timezone: TZ,
    fields: fields({
      attendee: ATTENDEE,
      customerName: CUSTOMER,
      date: DATE,
      dateFull: DATE_FULL,
      time: TIME,
    }),
    attempted: true,
    fieldErrors: { barber: 'Elige un barbero.', service: 'Elige un servicio.' },
    alert: { variant: 'danger', title: COPY.validationErrorTitle, body: COPY.validationErrorBody },
  },

  // Guardando: doble toque evitado, botón deshabilitado con "Guardando…".
  '10-guardando': {
    page: 'form',
    timezone: TZ,
    fields: FILLED,
    saving: true,
  },

  // Conflicto de horario (409): la hora recién se ocupó. El dato no se
  // pierde; el error vive también junto al campo de hora.
  '11-conflicto-horario': {
    page: 'form',
    timezone: TZ,
    fields: FILLED,
    attempted: true,
    fieldErrors: { time: COPY.startsAtConflict },
    alert: { variant: 'danger', title: COPY.conflictTitle, body: COPY.conflictBody },
  },

  // Éxito persistente: la cita quedó confirmada. HU-062 (agenda diaria) no
  // enlaza todavía, así que el resumen es el único destino de la pantalla.
  '12-turno-registrado': {
    page: 'success',
    created: {
      attendeeName: ATTENDEE,
      serviceName: SERVICE,
      durationMinutes: 30,
      priceAmount: '$45.000',
      currency: 'COP',
    },
  },
}

// Destinos reales del dock (mismos que panel-agenda-eventos).
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
