import { createServer } from 'node:http'

// Servidor efímero de demostración para `/panel`. Solo lo consume Vite en
// desarrollo; no pertenece a src/, al cliente HTTP ni al backend de producto.
//
// Los turnos se generan relativos a la fecha PEDIDA (query `date`, o "hoy"
// en America/Bogota si no llega ninguna) en vez de una fecha fija: una
// fecha fija queda "vieja" apenas pasa un día real y el turno más viejo
// termina el día anterior al que se está viendo, lo que rompe por completo
// el cálculo de la línea temporal (bounds/posición) porque esos turnos ya
// no son adyacentes al día mostrado. Con fecha relativa la demo sigue
// siendo válida sin importar cuándo se ejecute.
const barbers = [
  { id: 'barber-julian', fullName: 'Julián Rodríguez' },
  { id: 'barber-andres', fullName: 'Andrés Beltrán' },
  { id: 'barber-camilo', fullName: 'Camilo Restrepo' },
  { id: 'barber-tomas', fullName: 'Tomás Iriarte' },
]

const BOGOTA_OFFSET = '-05:00'

function todayInBogota() {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'America/Bogota',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(new Date())
  const year = parts.find((p) => p.type === 'year').value
  const month = parts.find((p) => p.type === 'month').value
  const day = parts.find((p) => p.type === 'day').value
  return `${year}-${month}-${day}`
}

// Instante ISO para `hhmm` del día civil `civilDate` en America/Bogota
// (desfase fijo, sin horario de verano en esa zona).
function instantAt(civilDate, hhmm) {
  return `${civilDate}T${hhmm}:00${BOGOTA_OFFSET}`
}

const appointment = (
  id,
  attendeeName,
  civilDate,
  startHHMM,
  endHHMM,
  status = 'confirmed',
  serviceName = 'Corte clásico',
) => ({
  id,
  attendeeName,
  startsAt: instantAt(civilDate, startHHMM),
  endsAt: instantAt(civilDate, endHHMM),
  status,
  origin: 'manual',
  serviceName,
  durationMinutes: 45,
  priceAmount: '45000.00',
  currency: 'COP',
})

function bogotaNowHHMM() {
  const parts = new Intl.DateTimeFormat('en-GB', {
    timeZone: 'America/Bogota',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).formatToParts(new Date())
  const hour = Number(parts.find((p) => p.type === 'hour').value)
  const minute = Number(parts.find((p) => p.type === 'minute').value)
  return hour * 60 + minute
}

function hhmm(minutesOfDay) {
  const clamped = Math.min(23 * 60 + 59, Math.max(0, minutesOfDay))
  const hour = Math.floor(clamped / 60)
  const minute = clamped % 60
  return `${String(hour).padStart(2, '0')}:${String(minute).padStart(2, '0')}`
}

// Un turno extra, anclado a la hora real de Bogotá, solo para el día de
// hoy: sin esto, el marcador "Ahora" del carril solo aparece si alguien
// prueba la demo entre las 08:00 y las 19:00 (el rango de los cuatro
// turnos fijos de abajo) — fuera de ese horario simplemente no hay nada
// que mostrar, aunque el cálculo esté bien (issue #189, reporte en vivo:
// "falta el coso naranja"). Con este turno el marcador siempre cae dentro
// del rango visible, sin importar cuándo se abra la demo.
function nowAnchoredAppointment(civilDate) {
  const nowMinutes = bogotaNowHHMM()
  return appointment(
    'turno-ahora',
    'Turno en curso',
    civilDate,
    hhmm(nowMinutes - 20),
    hhmm(nowMinutes + 25),
    'confirmed',
    'Demo en vivo',
  )
}

function agendaFor(civilDate) {
  const items = [
    appointment('turno-mateo', 'Mateo Rojas', civilDate, '09:00', '09:45'),
    appointment(
      'turno-samuel',
      'Samuel Díaz',
      civilDate,
      '10:30',
      '11:30',
      'confirmed',
      'Corte + barba',
    ),
    appointment('turno-carlos', 'Carlos Ruiz', civilDate, '13:00', '13:30', 'completed', 'Barba'),
    appointment(
      'turno-daniel',
      'Daniel López',
      civilDate,
      '15:30',
      '16:20',
      'cancelled_by_customer',
      'Fade premium',
    ),
  ]
  if (civilDate === todayInBogota()) {
    items.push(nowAnchoredAppointment(civilDate))
  }
  return items
}

function respond(response, status, body) {
  response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' })
  response.end(JSON.stringify(body))
}

const DEMO_EMAIL = 'demo@nava.test'
const DEMO_PASSWORD = 'Demo1234!'

function readJsonBody(request) {
  return new Promise((resolve) => {
    let raw = ''
    request.on('data', (chunk) => {
      raw += chunk
    })
    request.on('end', () => {
      try {
        resolve(raw ? JSON.parse(raw) : {})
      } catch {
        resolve({})
      }
    })
  })
}

createServer(async (request, response) => {
  const url = new URL(request.url ?? '/', 'http://localhost')
  const { pathname } = url

  if (request.method === 'POST' && pathname.endsWith('/public/auth/login')) {
    const body = await readJsonBody(request)
    if (body.email === DEMO_EMAIL && body.password === DEMO_PASSWORD) {
      respond(response, 200, { expiresAt: '2099-01-01T00:00:00Z' })
      return
    }
    respond(response, 401, { title: 'Credenciales inválidas' })
    return
  }

  if (pathname === '/api/v1/private/auth/session') {
    respond(response, 200, {
      barbershop: { id: 'shop-mock', name: 'Taller NAVA' },
      expiresAt: '2099-01-01T00:00:00Z',
    })
    return
  }
  if (pathname === '/api/v1/private/barbers') {
    respond(response, 200, { items: barbers })
    return
  }
  if (pathname === '/api/v1/private/settings/barbershop') {
    respond(response, 200, { timezone: 'America/Bogota' })
    return
  }
  if (pathname.endsWith('/appointments/daily-agenda')) {
    const requestedDate = url.searchParams.get('date') || todayInBogota()
    respond(response, 200, { items: agendaFor(requestedDate) })
    return
  }

  respond(response, 404, { title: 'Mock endpoint not found' })
}).listen(5177, () => {
  console.log('Mock API de agenda disponible en http://localhost:5177')
})
