/*
 * Modelo de contenido de los mockups de `/panel/turnos/nuevo` (Nuevo turno).
 *
 * Mismo criterio que `tools/mockups/auth-eventos` y
 * `tools/mockups/panel-agenda-eventos`: cada evento se declara una sola vez
 * y se renderiza en los dos viewports, de modo que escritorio y móvil no
 * puedan divergir en copy, orden ni estado.
 *
 * Los textos provienen del módulo real
 * `apps/web/src/modules/agenda/pages/NewAppointmentPage.vue` (HU-061).
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
  noBarbers:
    'Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" antes de registrar turnos.',

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

  resumenTitle: 'Resumen del turno',
  resumenBarber: 'Barbero',
  resumenService: 'Servicio',
  resumenSchedule: 'Fecha y hora',
  resumenAttendee: 'Cliente atendido',
  resumenPhone: 'Teléfono',
  resumenTimezone: 'Zona horaria',

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
const DATE = '24/05/2025'
const TIME = '11:30 AM'
const NOTE = 'Cliente frecuente.'

/** Estado por defecto de cada campo: vacío, salvo que el evento lo llene. */
const fields = (over = {}) => ({
  barber: '',
  service: '',
  attendee: '',
  customerName: '',
  phone: '',
  email: '',
  date: '',
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

  // Formulario recién abierto: nada elegido todavía, sin resumen.
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

  // Formulario completo, con el resumen visible antes del envío.
  '08-formulario-completo': {
    page: 'form',
    timezone: TZ,
    fields: FILLED,
    resumen: true,
  },

  // Envío intentado con campos inválidos: error por campo (aquí, barbero y
  // servicio sin elegir) más la alerta global de validación.
  '09-error-validacion': {
    page: 'form',
    timezone: TZ,
    fields: fields({ attendee: ATTENDEE, customerName: CUSTOMER, date: DATE, time: TIME }),
    attempted: true,
    fieldErrors: { barber: 'Elige un barbero.', service: 'Elige un servicio.' },
    alert: { variant: 'danger', title: COPY.validationErrorTitle, body: COPY.validationErrorBody },
  },

  // Guardando: doble toque evitado, botón deshabilitado con "Guardando…".
  '10-guardando': {
    page: 'form',
    timezone: TZ,
    fields: FILLED,
    resumen: true,
    saving: true,
  },

  // Conflicto de horario (409): la hora recién se ocupó. El dato no se
  // pierde; el error vive también junto al campo de hora.
  '11-conflicto-horario': {
    page: 'form',
    timezone: TZ,
    fields: FILLED,
    resumen: true,
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
