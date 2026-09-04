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

function agendaFor(civilDate) {
  return [
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
}

function respond(response, status, body) {
  response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' })
  response.end(JSON.stringify(body))
}

createServer((request, response) => {
  const url = new URL(request.url ?? '/', 'http://localhost')
  const { pathname } = url

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
