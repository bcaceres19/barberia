/*
 * Modelo de contenido de los mockups de `/panel` (agenda diaria).
 *
 * Mismo criterio que `tools/mockups/auth-eventos`: cada evento se declara una
 * sola vez y se renderiza en los dos viewports, de modo que escritorio y móvil
 * no puedan divergir en copy, orden ni estado.
 *
 * Los textos provienen del módulo real `apps/web/src/modules/agenda`
 * (`DailyAgendaPage.vue` y `model/dailyAgenda.ts`). Ningún texto se inventa:
 * si una cadena no existe en el código, procede de una `DEC-*` registrada.
 */

// Copys canónicos tomados del código real.
export const COPY = {
  // apps/web/src/modules/agenda/pages/DailyAgendaPage.vue
  title: 'Agenda',
  cta: 'Nuevo turno',
  barberLabel: 'Barbero',
  dateLabel: 'Fecha',
  prevDay: 'Anterior',
  nextDay: 'Siguiente',
  loadingBarbers: 'Cargando barberos…',
  loadErrorTitle: 'No pudimos cargar esta sección',
  loadErrorBody: 'Revisa tu conexión e inténtalo de nuevo.',
  retry: 'Reintentar',
  noBarbers:
    'Aún no tienes barberos registrados. Agrega uno en la sección "Barberos" para ver su agenda.',
  loadingAgenda: (name) => `Cargando la agenda de ${name}…`,
  updating: 'Actualizando…',
  notFoundTitle: 'Este barbero ya no está disponible',
  notFoundBody: 'Elige otro barbero en la lista.',
  agendaErrorTitle: 'No pudimos cargar la agenda de este barbero',
  agendaErrorBody: 'Revisa tu conexión e inténtalo de nuevo.',
  empty: (name, when) => `No hay turnos para ${name} ${when}.`,
  now: 'Ahora',

  // DEC-074/DEC-075: la zona horaria gobierna la navegación por fecha. El
  // texto es el de la propia pantalla cuando el dato no se pudo confirmar.
  timezoneErrorTitle: 'No pudimos confirmar la zona horaria',
  timezoneErrorBody:
    'La agenda sigue visible, pero la navegación por fecha queda bloqueada hasta recuperar ese dato.',

  // Cruce de medianoche: marca del eje, no una acción.
  dayChange: 'Cambio de día',
}

// apps/web/src/modules/agenda/model/dailyAgenda.ts
export const STATUS_LABELS = {
  confirmed: 'Confirmado',
  completed: 'Completado',
  cancelled_by_customer: 'Cancelado por el cliente',
  cancelled_by_barber: 'Cancelado por el barbero',
  no_show: 'No se presentó',
}

/*
 * El barbero deja de ser una cadena suelta: donde se ve o se elige, va con su
 * retrato. `photo` marca a quien tiene fotografía cargada; quien no la tiene
 * conserva el monograma derivado de su nombre, que es lo único que el
 * contrato vigente declara.
 */
const BARBER = { name: 'Julián Rodríguez', photo: true }

// Plantel completo para la lista desplegada. La mezcla es deliberada: el
// atlas debe mostrar convivir foto y monograma, porque en producción va a
// pasar siempre.
const ROSTER = [
  BARBER,
  { name: 'Andrés Beltrán', photo: true },
  { name: 'Camilo Restrepo', photo: false },
  { name: 'Tomás Iriarte', photo: true },
]
const SHOP = 'NAVA QA Local'
const TZ = 'America/Bogota'

const DATE_TODAY = 'Hoy, 3 de septiembre'
const DATE_TOMORROW = 'Viernes, 4 de septiembre'

// Turnos sintéticos del día. `start`/`end` en minutos desde medianoche: la
// posición en la línea temporal se calcula de ahí, así que una ficha no puede
// quedar desalineada de su propia hora.
const DAY_ENTRIES = [
  { start: 540, end: 585, name: 'Mateo Rojas', service: 'Corte clásico', status: 'confirmed' },
  { start: 630, end: 690, name: 'Samuel Díaz', service: 'Corte + barba', status: 'confirmed' },
  { start: 780, end: 810, name: 'Carlos Ruiz', service: 'Barba', status: 'completed' },
  {
    start: 930,
    end: 980,
    name: 'Daniel López',
    service: 'Fade premium',
    status: 'cancelled_by_customer',
  },
]

// Turno que cruza medianoche: 23:30 → 00:30 del día siguiente.
const NIGHT_ENTRIES = [
  {
    start: 1410,
    end: 1470,
    name: 'Laura Méndez',
    service: 'Corte clásico',
    status: 'confirmed',
    endsNextDay: 'Termina el sábado 5 de septiembre',
  },
]

const DAY_AXIS = { from: 480, to: 1080 } // 08:00 – 18:00
const NIGHT_AXIS = { from: 1320, to: 1560 } // 22:00 – 02:00

const NOW_TODAY = 675 // 11:15, dentro del turno de Samuel Díaz

/** Cascarón común. `chrome` describe lo que rodea al contenido del evento. */
const shell = (over = {}) => ({
  shop: SHOP,
  date: DATE_TODAY,
  timezone: TZ,
  barber: BARBER,
  options: ROSTER,
  open: false,
  showCta: true,
  showControls: true,
  datesEnabled: true,
  ...over,
})

export const PANEL = {
  screens: {
    // Agenda de hoy cargada. La línea temporal es decorativa y la lista de
    // abajo es la representación accesible equivalente: ambas se dibujan,
    // porque el implementador debe conservar las dos.
    '01-agenda-lista': {
      shell: shell(),
      axis: DAY_AXIS,
      now: NOW_TODAY,
      entries: DAY_ENTRIES,
    },

    // Carga inicial: todavía no se resolvió barbería, barberos ni zona, así
    // que no hay filtros, fecha ni acción que ofrecer.
    '02-carga-contexto': {
      shell: shell({ date: null, timezone: null, showCta: false, showControls: false }),
      state: { kind: 'loading', text: COPY.loadingBarbers },
    },

    // Fallo inicial de la sección: sin contexto resuelto tampoco hay CTA.
    '03-error-contexto': {
      shell: shell({ date: null, timezone: null, showCta: false, showControls: false }),
      state: {
        kind: 'alert',
        variant: 'warning',
        title: COPY.loadErrorTitle,
        body: COPY.loadErrorBody,
        action: COPY.retry,
      },
    },

    // Sin barberos activos no se inventa una selección ni se ofrece crear un
    // turno que no podría asignarse.
    '04-sin-barberos': {
      shell: shell({ date: null, timezone: null, showCta: false, showControls: false }),
      // Misma cadena del código, compuesta en dos partes: la primera es el
      // titular y la segunda dice qué hacer. No se reescribe el texto.
      state: {
        kind: 'note',
        text: 'Aún no tienes barberos registrados.',
        body: 'Agrega uno en la sección «Barberos» para ver su agenda.',
      },
    },

    // Primera carga de la agenda: barbero y fecha ya resueltos y visibles.
    '05-carga-agenda': {
      shell: shell({ showCta: false }),
      // Barbero y fecha ya resueltos: la espera conserva la geometría que
      // va a ocupar la agenda, para que la pantalla no salte al llegar.
      state: { kind: 'loading', text: COPY.loadingAgenda(BARBER.name), skeleton: true },
    },

    // Cambio de fecha: conserva la agenda anterior atenuada y anuncia la
    // actualización. El marcador "Ahora" desaparece porque la fecha mostrada
    // ya no es hoy.
    '06-actualizando-fecha': {
      shell: shell({ date: DATE_TOMORROW }),
      axis: DAY_AXIS,
      now: null,
      entries: DAY_ENTRIES,
      stale: true,
      status: COPY.updating,
    },

    // Error recuperable: barbero y fecha se conservan para reintentar sin
    // volver a elegirlos.
    '07-error-agenda': {
      shell: shell({ showCta: false }),
      state: {
        kind: 'alert',
        variant: 'warning',
        title: COPY.agendaErrorTitle,
        body: COPY.agendaErrorBody,
        action: COPY.retry,
      },
    },

    // El barbero solicitado dejó de estar disponible: la selección queda
    // vacía, porque conservarla contradiría el mensaje que pide elegir otro.
    '08-barbero-no-disponible': {
      shell: shell({ barber: null, showCta: false }),
      state: {
        kind: 'alert',
        variant: 'warning',
        title: COPY.notFoundTitle,
        body: COPY.notFoundBody,
      },
    },

    // Día válido sin turnos: se conserva la acción real de la pantalla, pero
    // una sola vez. Repetir "Nuevo turno" en el encabezado y dentro del estado
    // vacío duplica la misma acción primaria en una pantalla que no tiene otra
    // cosa que ofrecer, así que aquí vive donde está la atención.
    '09-dia-sin-turnos': {
      shell: shell({ showCta: false }),
      axis: DAY_AXIS,
      now: NOW_TODAY,
      entries: [],
      state: {
        kind: 'note',
        text: COPY.empty(BARBER.name, 'hoy'),
        action: COPY.cta,
      },
    },

    // La zona horaria no se pudo confirmar: la agenda sigue visible y solo se
    // bloquea la navegación por fecha, tal como dice el propio mensaje.
    '10-zona-horaria-no-disponible': {
      shell: shell({ timezone: null, datesEnabled: false }),
      axis: DAY_AXIS,
      now: null,
      entries: DAY_ENTRIES,
      alert: {
        variant: 'warning',
        title: COPY.timezoneErrorTitle,
        body: COPY.timezoneErrorBody,
        action: COPY.retry,
      },
    },

    // Selección de barbero: la lista desplegada muestra a cada persona con su
    // retrato, y convive quien tiene foto con quien conserva el monograma.
    '12-seleccion-barbero': {
      shell: shell({ open: true, showCta: false }),
      axis: DAY_AXIS,
      now: NOW_TODAY,
      entries: DAY_ENTRIES,
    },

    // Turno que cruza medianoche: el eje llega hasta las 02:00 y la marca de
    // cambio de día cae exactamente sobre las 00:00.
    '11-turno-nocturno': {
      shell: shell({ date: DATE_TOMORROW }),
      axis: NIGHT_AXIS,
      now: null,
      dayChangeAt: 1440,
      entries: NIGHT_ENTRIES,
    },
  },
}

// Destinos reales del dock (apps/web/src/modules/auth/components/AppNav.vue y
// el router). En móvil los secundarios se agrupan en "Más" sin desaparecer.
export const NAV_PRIMARY = [
  { label: 'Agenda', icon: 'agenda', active: true },
  { label: 'Servicios', icon: 'servicios' },
  { label: 'Barberos', icon: 'barberos' },
]
export const NAV_SECONDARY = [
  { label: 'Horarios', icon: 'horarios' },
  { label: 'Bloqueos', icon: 'bloqueos' },
  { label: 'Configuración', icon: 'configuracion' },
  { label: 'Servicios por barbero', icon: 'servicios-barbero' },
]
