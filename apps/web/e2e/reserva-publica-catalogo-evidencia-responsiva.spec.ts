import { test, expect, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-091 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): carga,
 * lista con servicios, selección, vacío y error "no encontrado" en
 * 320/360/768/1280 px, más una aproximación de zoom de texto 200% (mismo
 * criterio que reserva-publica-evidencia-responsiva.spec.ts, HU-090). Sin
 * mockup exacto asignado a esta pantalla (composición libre dentro de
 * NAVA, DEC-078). Las capturas se guardan en `e2e/evidence/` (no ignorado
 * por git) para quedar adjuntas como archivos reales del PR.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'reserva-publica-catalogo',
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
    // reserva-publica-evidencia-responsiva.spec.ts).
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

const corteClasico = {
  id: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
  name: 'Corte clásico',
  description: 'Corte con máquina y tijera, incluye lavado.',
  durationMinutes: 30,
  price: '45000.00',
  currency: 'COP',
}
const barba = {
  id: '1a2b3c4d-5e6f-4708-9a0b-1c2d3e4f5061',
  name: 'Barba',
  description: null,
  durationMinutes: 20,
  price: '20000.00',
  currency: 'COP',
}

async function fulfillServices(route: Route) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ items: [corteClasico, barba], nextCursor: null }),
  })
}

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`carga (${viewport.name}px)`, async ({ page }) => {
      await page.route('**/api/v1/public/barbershops/**/services**', async (route) => {
        // Mismo margen que reserva-publica-evidencia-responsiva.spec.ts:
        // la ruta se carga de forma diferida y debe compilar/servir su
        // propio chunk bajo el servidor de desarrollo de Vite.
        await new Promise((resolve) => setTimeout(resolve, 1500))
        await fulfillServices(route)
      })
      await page.goto('/reservar/barberia-ejemplo/servicios')
      await expect(page.getByText('Cargando servicios')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'carga.png'),
        fullPage: true,
      })
    })

    test(`lista, selección y sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.route('**/api/v1/public/barbershops/**/services**', fulfillServices)
      await page.goto('/reservar/barberia-ejemplo/servicios')
      await expect(page.getByText('Corte clásico')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)
      await assertNoAxeViolations(page)

      await page.getByRole('radio', { name: /Corte clásico/ }).click()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'seleccionado.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`catálogo vacío (${viewport.name}px)`, async ({ page }) => {
      await page.route('**/api/v1/public/barbershops/**/services**', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ items: [], nextCursor: null }),
        })
      })
      await page.goto('/reservar/barberia-ejemplo/servicios')
      await expect(
        page.getByText('Esta barbería todavía no tiene servicios disponibles para reservar.'),
      ).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'vacio.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`error de enlace no disponible, con foco en reintentar (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.route('**/api/v1/public/barbershops/**/services**', async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/api/v1/problems/not-found',
            title: 'No encontrado',
            status: 404,
            detail: 'no existe una barbería pública con ese enlace',
            instance: 'evidence-hu091-404',
            code: 'not-found',
            requestId: 'evidence-hu091-404',
          }),
        })
      })
      await page.goto('/reservar/enlace-que-no-existe/servicios')
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
