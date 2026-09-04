import { createServer } from 'node:http'

// Servidor efímero de demostración para `/panel`. Solo lo consume Vite en
// desarrollo; no pertenece a src/, al cliente HTTP ni al backend de producto.
const barbers = [
  { id: 'barber-julian', fullName: 'Julián Rodríguez' },
  { id: 'barber-andres', fullName: 'Andrés Beltrán' },
  { id: 'barber-camilo', fullName: 'Camilo Restrepo' },
  { id: 'barber-tomas', fullName: 'Tomás Iriarte' },
]

const appointment = (
  id,
  attendeeName,
  startsAt,
  endsAt,
  status = 'confirmed',
  serviceName = 'Corte clásico',
) => ({
  id,
  attendeeName,
  startsAt,
  endsAt,
  status,
  origin: 'manual',
  serviceName,
  durationMinutes: 45,
  priceAmount: '45000.00',
  currency: 'COP',
})

const agenda = [
  appointment('turno-mateo', 'Mateo Rojas', '2026-09-03T14:00:00Z', '2026-09-03T14:45:00Z'),
  appointment(
    'turno-samuel',
    'Samuel Díaz',
    '2026-09-03T15:30:00Z',
    '2026-09-03T16:30:00Z',
    'confirmed',
    'Corte + barba',
  ),
  appointment(
    'turno-carlos',
    'Carlos Ruiz',
    '2026-09-03T18:00:00Z',
    '2026-09-03T18:30:00Z',
    'completed',
    'Barba',
  ),
  appointment(
    'turno-daniel',
    'Daniel López',
    '2026-09-03T20:30:00Z',
    '2026-09-03T21:20:00Z',
    'cancelled_by_customer',
    'Fade premium',
  ),
]

function respond(response, status, body) {
  response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8' })
  response.end(JSON.stringify(body))
}

createServer((request, response) => {
  const { pathname } = new URL(request.url ?? '/', 'http://localhost')

  if (pathname === '/api/v1/private/auth/session') {
    respond(response, 200, {
      barbershop: { id: 'shop-mock', name: 'Taller NAVA' },
      expiresAt: '2026-09-04T00:00:00Z',
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
    respond(response, 200, { items: agenda })
    return
  }

  respond(response, 404, { title: 'Mock endpoint not found' })
}).listen(5177, () => {
  console.log('Mock API de agenda disponible en http://localhost:5177')
})
