import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible del perfil «barbero individual» (#294, DEC-115):
 * Agenda (sin selector y con «mi día»), Servicios (columna «Lo ofrezco»), Mi perfil y la
 * sección «Perfil del panel» de Configuración, en 320/360/768/1280 px y en los dos modos
 * (Tinta y Marfil), sin scroll horizontal y con axe-core (incluido el contraste real de
 * color) sobre cada combinación. Corre con respuestas simuladas del API: el recorrido contra
 * el API real vive en perfil-barbero-individual.spec.ts. Las capturas van a
 * e2e/evidence/perfil-barbero-individual/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'perfil-barbero-individual')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const BARBER = 'barber-mateo'
const brand = {
  accent: 'brass',
  businessTerm: 'barbería',
  businessTermGender: 'feminine',
  professionalTerm: 'barbero',
  professionalTermPlural: 'barberos',
  professionalTermGender: 'masculine',
  panelProfile: 'solo',
}
const settings = {
  name: 'Mateo · Barbero',
  timezone: 'America/Bogota',
  contactEmail: 'mateo@ejemplo.test',
  contactPhone: '+573001234567',
}
const barberRow = {
  id: BARBER,
  fullName: 'Mateo Rojas',
  createdAt: '2026-08-23T15:04:05Z',
  updatedAt: '2026-08-23T15:04:05Z',
  photoUpdatedAt: null,
}
const serviceRow = (id: string, name: string, minutes: number, price: string) => ({
  id,
  name,
  description: null,
  durationMinutes: minutes,
  price,
  currency: 'COP',
  isActive: true,
  deactivatedAt: null,
  createdAt: '2026-08-24T15:04:05Z',
  updatedAt: '2026-08-24T15:04:05Z',
})
const services = [
  serviceRow('s-1', 'Corte clásico', 40, '30000.00'),
  serviceRow('s-2', 'Barba perfilada', 30, '20000.00'),
  serviceRow('s-3', 'Corte y barba', 60, '45000.00'),
  serviceRow('s-4', 'Diseño con navaja', 20, '15000.00'),
]
const offered = ['s-1', 's-2', 's-3']

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

function todayInBogota() {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'America/Bogota',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(new Date())
  const part = (type: string) => parts.find((p) => p.type === type)!.value
  return `${part('year')}-${part('month')}-${part('day')}`
}

// Turnos de hoy en la zona de la barbería: uno en curso y dos por venir, para que el
// resumen de «mi día» diga algo en cualquier hora en que se ejecute la spec.
function todayAgenda() {
  const now = Date.now()
  const slot = (
    id: string,
    name: string,
    service: string,
    fromNowMin: number,
    minutes: number,
  ) => ({
    id,
    attendeeName: name,
    startsAt: new Date(now + fromNowMin * 60_000).toISOString(),
    endsAt: new Date(now + (fromNowMin + minutes) * 60_000).toISOString(),
    status: 'confirmed',
    origin: 'manual',
    serviceName: service,
    durationMinutes: minutes,
    priceAmount: '30000.00',
    currency: 'COP',
  })
  return [
    slot('a-1', 'Cliente de prueba Uno', 'Corte clásico', -10, 40),
    slot('a-2', 'Cliente de prueba Dos', 'Corte y barba', 50, 60),
    slot('a-3', 'Cliente de prueba Tres', 'Barba perfilada', 150, 30),
  ]
}

async function mockApi(page: Page) {
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'mateo', name: 'Mateo · Barbero' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
    } else if (pathname.endsWith('/settings/brand')) {
      await route.fulfill(json(brand))
    } else if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json(settings))
    } else if (pathname.endsWith('/daily-agenda')) {
      await route.fulfill(json({ items: todayAgenda() }))
    } else if (pathname.endsWith(`/barbers/${BARBER}/services`)) {
      await route.fulfill(
        json({
          items: offered.map((serviceId) => ({
            barberId: BARBER,
            serviceId,
            createdAt: '2026-10-04T00:00:00Z',
          })),
          nextCursor: null,
        }),
      )
    } else if (pathname.endsWith('/barbers')) {
      await route.fulfill(
        json({
          items: [barberRow],
          page: 1,
          pageSize: 20,
          total: 1,
          totalPages: 1,
          nextCursor: null,
        }),
      )
    } else if (pathname.endsWith('/services')) {
      await route.fulfill(
        json({
          items: services,
          page: 1,
          pageSize: 20,
          total: services.length,
          totalPages: 1,
        }),
      )
    } else {
      await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
    }
  })
}

const screens = [
  {
    name: 'agenda',
    url: `/panel?barberId=${BARBER}&date=${todayInBogota()}`,
    ready: (page: Page) =>
      expect(page.locator('.daily-agenda-page__summary')).toContainText('turnos por atender'),
  },
  {
    name: 'servicios',
    url: '/panel/servicios',
    ready: async (page: Page) => {
      await expect(
        page.getByRole('switch', { name: 'Lo ofrezco: Corte clásico' }).first(),
      ).toBeVisible()
    },
  },
  {
    name: 'mi-perfil',
    url: '/panel/barberos',
    ready: (page: Page) => expect(page.getByRole('heading', { name: 'Mi perfil' })).toBeVisible(),
  },
  {
    name: 'configuracion-perfil',
    url: '/panel/barberia',
    ready: async (page: Page) => {
      const group = page.getByRole('radiogroup', { name: 'Perfil del panel' })
      await expect(group).toBeVisible()
      // La página abre en «Pantalla»: se lleva la sección nueva al encuadre antes de capturar.
      await page.locator('#configuracion-perfil').scrollIntoViewIfNeeded()
    },
  },
] as const

async function openScreen(
  page: Page,
  screen: (typeof screens)[number],
  width: number,
  height: number,
  theme: 'ink' | 'ivory',
) {
  await page.setViewportSize({ width, height })
  await page.addInitScript((stored) => {
    window.localStorage.setItem(
      'nava.appearance.v1',
      JSON.stringify({ theme: stored, textScale: 'normal', motion: 'reduced' }),
    )
  }, theme)
  await mockApi(page)
  await page.goto(screen.url)
  await screen.ready(page)
  // La entrada escalonada de la pantalla termina antes de medir o capturar.
  await page.waitForTimeout(900)
}

async function expectNoHorizontalScroll(page: Page) {
  const overflow = await page.evaluate(() => {
    const shell = document.querySelector('.private-shell__content') as HTMLElement
    return {
      document: document.documentElement.scrollWidth > document.documentElement.clientWidth,
      content: shell.scrollWidth > shell.clientWidth,
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

for (const screen of screens) {
  test.describe(screen.name, () => {
    for (const { theme, viewport } of combos) {
      test(`${theme} a ${viewport.name}: sin scroll horizontal y accesible`, async ({ page }) => {
        await openScreen(page, screen, viewport.width, viewport.height, theme)

        await page.screenshot({
          path: path.join(evidenceDir, theme, viewport.name, `${screen.name}.png`),
          fullPage: true,
        })
        await expectNoHorizontalScroll(page)
        expect(await axeViolations(page)).toEqual([])
      })
    }
  })
}

test('el dock del perfil solista cabe en 320 px con «Más» para lo secundario', async ({ page }) => {
  await openScreen(page, screens[0], 320, 720, 'ink')

  const mobile = page.locator('.app-nav__list--mobile .app-nav__label')
  await expect(mobile).toHaveText(['Agenda', 'Servicios', 'Mi perfil', 'Horarios', 'Más'])
  const dock = await page.locator('.app-nav').boundingBox()
  expect(dock!.y + dock!.height).toBeLessThanOrEqual(720 + 1)
})
