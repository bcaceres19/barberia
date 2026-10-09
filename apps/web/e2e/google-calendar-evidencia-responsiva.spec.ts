import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Recorrido y evidencia responsive/accesible de la pantalla «Google Calendar»
 * (#325, DEC-099/DEC-101/DEC-122), con respuestas simuladas del API que conservan
 * el estado entre solicitudes:
 *  - recorrido: conectar (la redirección a Google se simula navegando a la URL de
 *    retorno), sincronizar, guardar el recordatorio y desconectar;
 *  - evidencia: cada estado en 320/360/768/1280 px y en Tinta y Marfil, sin scroll
 *    horizontal y con axe-core (incluido el contraste real).
 * El recorrido contra Google real y la base real lo cubren las pruebas de Go. Las
 * capturas van a e2e/evidence/google-calendar/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'google-calendar')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const brand = {
  accent: 'brass',
  businessTerm: 'barbería',
  businessTermGender: 'feminine',
  professionalTerm: 'barbero',
  professionalTermPlural: 'barberos',
  professionalTermGender: 'masculine',
  panelProfile: 'shop',
}
const settings = {
  name: 'Barbería de prueba',
  timezone: 'America/Bogota',
  contactEmail: 'equipo@ejemplo.test',
  contactPhone: '+573001234567',
}

interface Connection {
  enabled: boolean
  barberLinked: boolean
  status: string
  accountEmail: string | null
  reminderMinutes: number | null
  connectedAt: string | null
  lastSyncedAt: string | null
  pendingSyncJobs: number
  failedSyncJobs: number
}

const disconnected: Connection = {
  enabled: true,
  barberLinked: true,
  status: 'not_connected',
  accountEmail: null,
  reminderMinutes: null,
  connectedAt: null,
  lastSyncedAt: null,
  pendingSyncJobs: 0,
  failedSyncJobs: 0,
}
const connected: Connection = {
  ...disconnected,
  status: 'connected',
  accountEmail: 'barbero@ejemplo.test',
  reminderMinutes: 30,
  connectedAt: '2026-10-10T15:00:00Z',
  lastSyncedAt: '2026-10-10T16:30:00Z',
}

const states: Array<{ name: string; connection: Connection }> = [
  { name: 'no-conectado', connection: disconnected },
  { name: 'conectado', connection: connected },
  { name: 'sincronizando', connection: { ...connected, pendingSyncJobs: 3 } },
  { name: 'error-de-sincronizacion', connection: { ...connected, failedSyncJobs: 2 } },
  { name: 'requiere-reconexion', connection: { ...disconnected, status: 'reauth_required' } },
  { name: 'falta-vincular', connection: { ...disconnected, barberLinked: false } },
  { name: 'no-disponible', connection: { ...disconnected, enabled: false } },
]

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

interface Mock {
  calls: string[]
  connection: Connection
}

async function mockApi(page: Page, initial: Connection): Promise<Mock> {
  const mock: Mock = { calls: [], connection: { ...initial } }
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    const method = route.request().method()
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'demo', name: 'Barbería de prueba' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
    } else if (pathname.endsWith('/settings/brand')) {
      await route.fulfill(json(brand))
    } else if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json(settings))
    } else if (pathname.endsWith('/integrations/google-calendar/connect')) {
      mock.calls.push('connect')
      // Google devolvería al barbero a la pantalla de retorno con state y code.
      await route.fulfill(
        json({
          authorizationUrl:
            'http://localhost:5173/panel/barberia/google-calendar/callback?state=estado-1&code=codigo-1',
        }),
      )
    } else if (pathname.endsWith('/integrations/google-calendar/callback')) {
      const body = route.request().postDataJSON() as { state: string; code?: string }
      mock.calls.push(`callback:${body.state}:${body.code ?? ''}`)
      mock.connection = {
        ...connected,
        lastSyncedAt: null,
        reminderMinutes: null,
        pendingSyncJobs: 2,
      }
      await route.fulfill(json({ result: 'connected' }))
    } else if (pathname.endsWith('/integrations/google-calendar/sync')) {
      mock.calls.push('sync')
      mock.connection = {
        ...mock.connection,
        pendingSyncJobs: 0,
        lastSyncedAt: '2026-10-10T17:00:00Z',
      }
      await route.fulfill(json({ pendingSyncJobs: 0, failedSyncJobs: 0 }))
    } else if (pathname.endsWith('/integrations/google-calendar')) {
      if (method === 'GET') {
        await route.fulfill(json(mock.connection))
      } else if (method === 'PATCH') {
        const { reminderMinutes } = route.request().postDataJSON() as {
          reminderMinutes: number | null
        }
        mock.calls.push(`reminder:${reminderMinutes}`)
        mock.connection = { ...mock.connection, reminderMinutes }
        await route.fulfill(json(mock.connection))
      } else {
        mock.calls.push('disconnect')
        mock.connection = { ...disconnected, status: 'disconnected' }
        await route.fulfill({ status: 204 })
      }
    } else {
      await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
    }
  })
  return mock
}

async function openPage(
  page: Page,
  initial: Connection,
  width: number,
  height: number,
  theme: 'ink' | 'ivory',
  url = '/panel/barberia/google-calendar',
  heading = 'Google Calendar',
): Promise<Mock> {
  await page.setViewportSize({ width, height })
  await page.addInitScript((stored) => {
    window.localStorage.setItem(
      'nava.appearance.v1',
      JSON.stringify({ theme: stored, textScale: 'normal', motion: 'reduced' }),
    )
  }, theme)
  const mock = await mockApi(page, initial)
  await page.goto(url)
  await expect(page.getByRole('heading', { name: heading, level: 1 })).toBeVisible()
  // La entrada de la pantalla termina antes de medir o capturar.
  await page.waitForTimeout(500)
  return mock
}

async function expectNoHorizontalScroll(page: Page) {
  const overflow = await page.evaluate(() => {
    const shell = document.querySelector('.private-shell__content') as HTMLElement | null
    return {
      document: document.documentElement.scrollWidth > document.documentElement.clientWidth,
      content: shell ? shell.scrollWidth > shell.clientWidth : false,
    }
  })
  expect(overflow).toEqual({ document: false, content: false })
}

async function axeViolations(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  return page.evaluate(async () => {
    const results = await (
      window as unknown as {
        axe: {
          run: (
            c: Document,
            o: object,
          ) => Promise<{ violations: Array<{ id: string; nodes: Array<{ target: unknown }> }> }>
        }
      }
    ).axe.run(document, {
      runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
    })
    return results.violations.map((v) => ({ id: v.id, targets: v.nodes.map((n) => n.target) }))
  })
}

// Tinta en los cuatro anchos; Marfil en el angosto y el ancho, que son los extremos.
const combos = [
  ...viewports.map((viewport) => ({ theme: 'ink' as const, viewport })),
  { theme: 'ivory' as const, viewport: viewports[1] },
  { theme: 'ivory' as const, viewport: viewports[3] },
]

for (const state of states) {
  test.describe(state.name, () => {
    for (const { theme, viewport } of combos) {
      test(`${theme} a ${viewport.name}: sin scroll horizontal y accesible`, async ({ page }) => {
        await openPage(page, state.connection, viewport.width, viewport.height, theme)
        await page.screenshot({
          path: path.join(evidenceDir, theme, viewport.name, `${state.name}.png`),
          fullPage: true,
        })
        await expectNoHorizontalScroll(page)
        expect(await axeViolations(page)).toEqual([])
      })
    }
  })
}

test('recorrido: conectar, sincronizar, guardar el recordatorio y desconectar', async ({
  page,
}) => {
  const mock = await openPage(page, disconnected, 1280, 900, 'ink')

  // Conectar: el navegador va a «Google» (simulado) y vuelve con state y code.
  await page.getByRole('button', { name: 'Conectar Google Calendar' }).click()
  await expect(page).toHaveURL(/\/panel\/barberia\/google-calendar$/)
  await expect(page.getByText('barbero@ejemplo.test')).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Tu conexión' })).toBeVisible()
  // La URL de retorno no conserva el código OAuth ni el state.
  expect(page.url()).not.toContain('codigo-1')
  expect(page.url()).not.toContain('estado-1')
  expect(mock.calls.filter((c) => c.startsWith('callback'))).toEqual(['callback:estado-1:codigo-1'])

  // Con cambios en cola se muestra «Sincronizando» y «Sincronizar ahora» los publica.
  await expect(page.getByText('Sincronizando 2 cambios')).toBeVisible()
  await page.getByRole('button', { name: 'Sincronizar ahora' }).click()
  await expect(page.getByText('Última sincronización')).toBeVisible()
  expect(mock.calls).toContain('sync')

  // Recordatorio: validación y guardado.
  await page.getByLabel('Usar los recordatorios predeterminados de Google').uncheck()
  await page.getByLabel('Minutos de anticipación').fill('99999')
  await page.getByRole('button', { name: 'Guardar recordatorio' }).click()
  await expect(page.getByText('de 0 a 40320')).toBeVisible()
  await page.getByLabel('Minutos de anticipación').fill('45')
  await page.getByRole('button', { name: 'Guardar recordatorio' }).click()
  await expect.poll(() => mock.calls.includes('reminder:45')).toBe(true)
  await page.screenshot({
    path: path.join(evidenceDir, 'recorrido', 'conectado.png'),
    fullPage: true,
  })

  // Desconectar pide confirmación; cancelar no hace nada.
  await page.getByRole('button', { name: 'Desconectar' }).click()
  const dialog = page.getByRole('dialog', { name: '¿Desconectar Google Calendar?' })
  await expect(dialog).toBeVisible()
  await dialog.getByRole('button', { name: 'Cancelar' }).click()
  expect(mock.calls).not.toContain('disconnect')
  await page.getByRole('button', { name: 'Desconectar' }).click()
  await dialog.getByRole('button', { name: 'Desconectar' }).click()
  await expect(page.getByText('No conectado')).toBeVisible()
  expect(mock.calls).toContain('disconnect')
})

test('la pantalla de retorno limpia la URL y avisa si la autorización venció', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.route('**/api/v1/private/integrations/google-calendar/callback', (route) =>
    route.fulfill(json({ title: 'Solicitud inválida', status: 400, code: 'invalid-request' }, 400)),
  )
  await mockApi(page, disconnected)
  await page.route('**/api/v1/private/integrations/google-calendar/callback', (route) =>
    route.fulfill(json({ title: 'Solicitud inválida', status: 400, code: 'invalid-request' }, 400)),
  )
  await page.goto('/panel/barberia/google-calendar/callback?state=viejo&code=secreto')
  await expect(page.getByText('La autorización venció.')).toBeVisible()
  expect(page.url()).not.toContain('secreto')
  expect(await axeViolations(page)).toEqual([])
  await page.screenshot({ path: path.join(evidenceDir, 'recorrido', 'retorno-vencido.png') })
  await page.getByRole('link', { name: 'Volver a Google Calendar' }).click()
  await expect(page.getByRole('heading', { name: 'Google Calendar', level: 1 })).toBeVisible()
})

test('Configuración enlaza con la pantalla de Google Calendar', async ({ page }) => {
  await openPage(page, disconnected, 1280, 900, 'ink', '/panel/barberia', 'Configuración')
  await page
    .getByRole('link', { name: /Google Calendar/ })
    .first()
    .click()
  await expect(page).toHaveURL(/\/panel\/barberia\/google-calendar$/)
})
