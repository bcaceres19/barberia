import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-098 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): carga,
 * turno listo (estado terminal), enlace inválido y error de red, en
 * 320/360/768/1280 px, más una aproximación de zoom de texto 200% (mismo
 * criterio que reserva-publica-cliente-evidencia-responsiva.spec.ts,
 * HU-096). Sin mockup exacto asignado a esta pantalla (composición libre
 * dentro de NAVA, DEC-078). Las capturas se guardan en `e2e/evidence/`
 * (no ignorado por git) para quedar adjuntas como archivos reales del PR.
 */
const evidenceDir = path.join(path.dirname(fileURLToPath(import.meta.url)), 'evidence', 'mi-turno')
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
    // reserva-publica-cliente-evidencia-responsiva.spec.ts).
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

const ACCESS_TOKEN = '3n9F1sQe2z8mM5wYtR7pL0oXhC4bV6dK1aE9jU2gN8i'
const ROUTE = `/mi-turno/${ACCESS_TOKEN}`

const APPOINTMENT_VIEW = {
  barbershopName: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  attendeeName: 'Ana Ríos',
  serviceName: 'Corte clásico',
  durationMinutes: 30,
  barberName: 'Barbero Ejemplo',
  startsAt: '2027-03-15T15:00:00Z',
  endsAt: '2027-03-15T15:30:00Z',
  status: 'confirmed',
  cancellationDeadlineMinutes: 20,
  lateCancellationClientAllowed: true,
  lateCancellationReasonRequired: true,
}

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`turno listo, sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(APPOINTMENT_VIEW),
        })
      })

      await page.goto(ROUTE)
      await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)

      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'turno-listo.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`enlace inválido, con reintento visible y foco por teclado (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.route('**/api/v1/customer/appointments/**', async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/api/v1/problems/not-found',
            title: 'No encontrado',
            status: 404,
            detail: 'no existe un turno accesible con ese enlace',
            instance: 'e2e-hu098-evidencia-404',
            code: 'not-found',
            requestId: 'e2e-hu098-evidencia-404',
          }),
        })
      })

      await page.goto('/mi-turno/token-invalido')
      await expect(page.getByRole('alert')).toContainText('No pudimos abrir ese enlace')

      const retryButton = page.getByRole('button', { name: 'Reintentar' })
      await retryButton.focus()
      await expect(retryButton).toBeFocused()

      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'enlace-invalido.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`error de red con reintento (${viewport.name}px)`, async ({ page }) => {
      await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, async (route) => {
        await route.abort('failed')
      })

      await page.goto(ROUTE)
      await expect(page.getByText('No pudimos conectar')).toBeVisible()

      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'error-red.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })
  })
}
