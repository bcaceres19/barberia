import { test, expect, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-095 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): carga,
 * selección de franja, navegación de día, vacío total y error "no
 * encontrado" en 320/360/768/1280 px, más una aproximación de zoom de texto
 * 200% (mismo criterio que
 * reserva-publica-barbero-evidencia-responsiva.spec.ts, HU-092). Sin
 * mockup exacto asignado a esta pantalla (composición libre dentro de
 * NAVA, DEC-078). Las capturas se guardan en `e2e/evidence/` (no ignorado
 * por git) para quedar adjuntas como archivos reales del PR.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'reserva-publica-horario',
)
const axeScriptPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'node_modules',
  'axe-core',
  'axe.min.js',
)

async function assertNoAxeViolations(page: import('@playwright/test').Page) {
  await page.addScriptTag({ path: axeScriptPath })
  const axeResults = (await page.evaluate(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const axeGlobal = (window as any).axe
    // color-contrast se desactiva: la paleta ya está verificada en la
    // tabla aprobada de estandar-diseno-visual.md §4.3 (mismo criterio que
    // reserva-publica-barbero-evidencia-responsiva.spec.ts).
    return axeGlobal.run(document, { rules: { 'color-contrast': { enabled: false } } })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  })) as { violations: any[] }
  expect(axeResults.violations, JSON.stringify(axeResults.violations, null, 2)).toEqual([])
}

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  // Aproximación de zoom 200% sobre 1280: mitad de ancho.
  { name: '1280-zoom200', width: 640, height: 450 },
]

const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'

async function fulfillAvailability(
  route: Route,
  body: {
    slots: Array<{ startsAt: string }>
    durationMinutes: number
    timezone: string
    slotGridMinutes: number
  },
) {
  await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) })
}

const availability = {
  slots: [
    { startsAt: '2026-09-15T19:00:00Z' },
    { startsAt: '2026-09-15T19:15:00Z' },
    { startsAt: '2026-09-16T14:00:00Z' },
  ],
  durationMinutes: 30,
  timezone: 'America/Bogota',
  slotGridMinutes: 15,
}

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`carga (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
        async (route) => {
          // Mismo margen que reserva-publica-barbero-evidencia-responsiva.spec.ts:
          // la ruta se carga de forma diferida y debe compilar/servir su
          // propio chunk bajo el servidor de desarrollo de Vite.
          await new Promise((resolve) => setTimeout(resolve, 1500))
          await fulfillAvailability(route, availability)
        },
      )
      await page.goto(
        `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
      )
      await expect(page.getByText('Cargando disponibilidad')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'carga.png'),
        fullPage: true,
      })
    })

    test(`día con franjas, sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
        (route) => fulfillAvailability(route, availability),
      )
      await page.goto(
        `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
      )
      await expect(page.getByRole('radio')).toHaveCount(2)
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)
      await assertNoAxeViolations(page)
    })

    test(`franja elegida (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
        (route) => fulfillAvailability(route, availability),
      )
      await page.goto(
        `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
      )
      await page.getByRole('radio').first().click()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'seleccionado.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`vacío total de la ventana (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
        (route) => fulfillAvailability(route, { ...availability, slots: [] }),
      )
      await page.goto(
        `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
      )
      await expect(page.getByText('No hay franjas disponibles')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'vacio.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`error de enlace no disponible, con foco en reintentar (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
        async (route) => {
          await route.fulfill({
            status: 404,
            contentType: 'application/problem+json',
            body: JSON.stringify({
              type: '/api/v1/problems/not-found',
              title: 'No encontrado',
              status: 404,
              detail: 'no existe una barbería pública con ese enlace',
              instance: 'evidence-hu095-404',
              code: 'not-found',
              requestId: 'evidence-hu095-404',
            }),
          })
        },
      )
      await page.goto(
        `/reservar/enlace-que-no-existe/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
      )
      await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'error.png'),
        fullPage: true,
      })

      await page.getByRole('button', { name: 'Reintentar' }).focus()
      await expect(page.getByRole('button', { name: 'Reintentar' })).toBeFocused()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'foco.png'),
        fullPage: true,
      })

      await assertNoAxeViolations(page)
    })
  })
}
