import { expect, test, type Page, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// Evidencia visual del issue #189. Estos datos existen SOLO dentro de
// Playwright: permiten revisar la composición de cada estado del atlas sin
// iniciar sesión ni tocar API, PostgreSQL, OTP o rate limiting locales.
type Scenario =
  | 'agenda'
  | 'context-loading'
  | 'context-error'
  | 'no-barbers'
  | 'agenda-loading'
  | 'updating'
  | 'agenda-error'
  | 'barber-not-found'
  | 'empty-day'
  | 'timezone-unavailable'
  | 'night-shift'
  | 'barber-selection'

type Deferred = { release: () => void }

const evidenceRoot = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'panel',
  'fidelidad-189',
  'mock',
)

const barbers = [
  { id: 'barber-julian', fullName: 'Julián Rodríguez' },
  { id: 'barber-lucia', fullName: 'Lucía Pérez' },
  { id: 'barber-sol', fullName: 'Sol' },
]

const agendaEntries = [
  entry('turno-mateo', 'Mateo Rojas', '2026-09-03T14:00:00Z', '2026-09-03T14:45:00Z'),
  entry(
    'turno-samuel',
    'Samuel Díaz',
    '2026-09-03T15:30:00Z',
    '2026-09-03T16:30:00Z',
    'confirmed',
    'Corte + barba',
  ),
  entry(
    'turno-carlos',
    'Carlos Ruiz',
    '2026-09-03T18:00:00Z',
    '2026-09-03T18:30:00Z',
    'completed',
    'Barba',
  ),
  entry(
    'turno-daniel',
    'Daniel López',
    '2026-09-03T20:30:00Z',
    '2026-09-03T21:20:00Z',
    'cancelled_by_customer',
    'Fade premium',
  ),
]

const nightEntry = entry(
  'turno-nocturno',
  'Diana Mora',
  '2026-09-04T04:30:00Z',
  '2026-09-04T05:30:00Z',
)

function entry(
  id: string,
  attendeeName: string,
  startsAt: string,
  endsAt: string,
  status: 'confirmed' | 'completed' | 'cancelled_by_customer' = 'confirmed',
  serviceName = 'Corte clásico',
) {
  return {
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
  }
}

function deferred(): { promise: Promise<void>; deferred: Deferred } {
  let release = () => {}
  const promise = new Promise<void>((resolve) => {
    release = resolve
  })
  return { promise, deferred: { release } }
}

async function respondJson(route: Route, body: unknown, status = 200) {
  await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) })
}

async function installPanelMock(page: Page, scenario: Scenario): Promise<Deferred | null> {
  const pending =
    scenario === 'context-loading' || scenario === 'agenda-loading' || scenario === 'updating'
      ? deferred()
      : null
  let agendaRequests = 0

  await page.route('**/api/v1/**', async (route) => {
    const requestUrl = new URL(route.request().url())
    const { pathname } = requestUrl

    if (pathname === '/api/v1/private/auth/session') {
      await respondJson(route, {
        barbershop: { id: 'shop-mock', name: 'Taller NAVA' },
        expiresAt: '2026-09-04T00:00:00Z',
      })
      return
    }

    if (pathname === '/api/v1/private/barbers') {
      if (scenario === 'context-loading' && pending) {
        await pending.promise
      }
      if (scenario === 'context-error') {
        await respondJson(route, { title: 'No disponible' }, 500)
        return
      }
      await respondJson(route, { items: scenario === 'no-barbers' ? [] : barbers })
      return
    }

    if (pathname === '/api/v1/private/settings/barbershop') {
      if (scenario === 'context-loading' && pending) {
        await pending.promise
      }
      if (scenario === 'timezone-unavailable') {
        await respondJson(route, { title: 'No disponible' }, 500)
        return
      }
      await respondJson(route, { timezone: 'America/Bogota' })
      return
    }

    if (pathname.endsWith('/appointments/daily-agenda')) {
      agendaRequests += 1
      if (scenario === 'agenda-loading' && pending) {
        await pending.promise
      }
      if (scenario === 'updating' && agendaRequests > 1 && pending) {
        await pending.promise
      }
      if (scenario === 'agenda-error') {
        await respondJson(route, { title: 'No disponible' }, 500)
        return
      }
      if (scenario === 'barber-not-found') {
        await respondJson(route, { title: 'No encontrado' }, 404)
        return
      }
      if (scenario === 'empty-day') {
        await respondJson(route, { items: [] })
        return
      }
      if (scenario === 'night-shift') {
        await respondJson(route, { items: [nightEntry] })
        return
      }
      await respondJson(route, { items: agendaEntries })
      return
    }

    await route.abort('blockedbyclient')
  })

  return pending?.deferred ?? null
}

async function waitForScenario(page: Page, scenario: Scenario) {
  switch (scenario) {
    case 'context-loading':
      await expect(page.getByText('Cargando barberos')).toBeVisible()
      return
    case 'context-error':
      await expect(page.getByText('No pudimos cargar esta sección')).toBeVisible()
      return
    case 'no-barbers':
      await expect(page.getByText('Aún no tienes barberos registrados')).toBeVisible()
      return
    case 'agenda-loading':
      await expect(page.locator('.agenda-skeleton')).toBeVisible()
      return
    case 'agenda-error':
      await expect(page.getByText('No pudimos cargar la agenda de este barbero')).toBeVisible()
      return
    case 'barber-not-found':
      await expect(page.getByText('Este barbero ya no está disponible')).toBeVisible()
      return
    case 'empty-day':
      await expect(page.getByText('No hay turnos para Julián Rodríguez hoy')).toBeVisible()
      return
    case 'timezone-unavailable':
      await expect(page.getByRole('button', { name: 'Anterior' })).toBeDisabled()
      return
    case 'night-shift':
      await expect(page.locator('.daily-agenda-page__timeline-mark--day-change')).toBeAttached()
      return
    case 'barber-selection':
      await page.getByRole('button', { name: 'Barbero', exact: true }).click()
      await expect(page.getByRole('listbox', { name: 'Barbero' })).toBeVisible()
      return
    case 'agenda':
    case 'updating':
      await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
  }
}

const scenarios: { readonly file: string; readonly state: Scenario }[] = [
  { file: '01-agenda-lista', state: 'agenda' },
  { file: '02-carga-contexto', state: 'context-loading' },
  { file: '03-error-contexto', state: 'context-error' },
  { file: '04-sin-barberos', state: 'no-barbers' },
  { file: '05-carga-agenda', state: 'agenda-loading' },
  { file: '06-actualizando-fecha', state: 'updating' },
  { file: '07-error-agenda', state: 'agenda-error' },
  { file: '08-barbero-no-disponible', state: 'barber-not-found' },
  { file: '09-dia-sin-turnos', state: 'empty-day' },
  { file: '10-zona-horaria-no-disponible', state: 'timezone-unavailable' },
  { file: '11-turno-nocturno', state: 'night-shift' },
  { file: '12-seleccion-barbero', state: 'barber-selection' },
]

for (const viewport of [
  { name: 'desktop', width: 1440, height: 1024 },
  { name: 'mobile', width: 420, height: 935 },
] as const) {
  test.describe(`fidelidad mock de /panel en ${viewport.name}`, () => {
    if (viewport.name === 'mobile') {
      test.use({ deviceScaleFactor: 2 })
    }

    for (const scenario of scenarios) {
      test(`${scenario.file} usa datos mock sin backend`, async ({ page }) => {
        await page.setViewportSize(viewport)
        await page.clock.install({ time: new Date('2026-09-03T16:15:00Z') })
        const pending = await installPanelMock(page, scenario.state)
        await page.goto('/panel', { waitUntil: 'domcontentloaded' })
        await waitForScenario(page, scenario.state)

        if (scenario.state === 'updating') {
          await page.getByRole('button', { name: 'Siguiente' }).click()
          await expect(page.getByText('Actualizando…')).toBeVisible()
        }

        expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
        expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)
        await page.screenshot({
          path: path.join(evidenceRoot, viewport.name, `${scenario.file}.png`),
          fullPage: true,
        })

        pending?.release()
      })
    }
  })
}

test('vista interactiva local de agenda mock', async ({ page }) => {
  test.skip(
    process.env.SHOW_PANEL_MOCK !== '1',
    'Solo se abre cuando se solicita una revisión local.',
  )

  await page.setViewportSize({ width: 1440, height: 1024 })
  await page.clock.install({ time: new Date('2026-09-03T16:15:00Z') })
  await installPanelMock(page, 'agenda')
  await page.goto('/panel')
  await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
  await page.pause()
})
