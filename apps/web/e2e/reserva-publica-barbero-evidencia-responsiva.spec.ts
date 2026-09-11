import { test, expect, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-092 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): carga,
 * preselección con un barbero, elección con varios, vacío y error "no
 * encontrado" en 320/360/768/1280 px, más una aproximación de zoom de texto
 * 200% (mismo criterio que
 * reserva-publica-catalogo-evidencia-responsiva.spec.ts, HU-091). Sin
 * mockup exacto asignado a esta pantalla (composición libre dentro de
 * NAVA, DEC-078). Las capturas se guardan en `e2e/evidence/` (no ignorado
 * por git) para quedar adjuntas como archivos reales del PR.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'reserva-publica-barbero',
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
    // reserva-publica-catalogo-evidencia-responsiva.spec.ts).
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
const ana = { id: 'a1111111-1111-1111-1111-111111111111', fullName: 'Ana Gómez' }
const luis = { id: 'b2222222-2222-2222-2222-222222222222', fullName: 'Luis Rojas' }

async function fulfillBarbers(route: Route, items: Array<{ id: string; fullName: string }>) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ items }),
  })
}

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`carga (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers**`,
        async (route) => {
          // Mismo margen que reserva-publica-catalogo-evidencia-responsiva.spec.ts:
          // la ruta se carga de forma diferida y debe compilar/servir su
          // propio chunk bajo el servidor de desarrollo de Vite.
          await new Promise((resolve) => setTimeout(resolve, 1500))
          await fulfillBarbers(route, [ana, luis])
        },
      )
      await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
      await expect(page.getByText('Cargando barberos')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'carga.png'),
        fullPage: true,
      })
    })

    test(`preselección con un barbero, sin scroll horizontal (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers**`,
        (route) => fulfillBarbers(route, [ana]),
      )
      await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
      await expect(page.getByText('Ana Gómez')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'preseleccionado.png'),
        fullPage: true,
      })

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)
      await assertNoAxeViolations(page)
    })

    test(`elección explícita con varios barberos (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers**`,
        (route) => fulfillBarbers(route, [ana, luis]),
      )
      await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
      await expect(page.getByText('Ana Gómez')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)

      await page.getByRole('radio', { name: /Ana Gómez/ }).click()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'seleccionado.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`estado vacío (${viewport.name}px)`, async ({ page }) => {
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers**`,
        (route) => fulfillBarbers(route, []),
      )
      await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
      await expect(
        page.getByText('Este servicio no tiene barberos disponibles en este momento.'),
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
      await page.route(
        `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers**`,
        async (route) => {
          await route.fulfill({
            status: 404,
            contentType: 'application/problem+json',
            body: JSON.stringify({
              type: '/api/v1/problems/not-found',
              title: 'No encontrado',
              status: 404,
              detail: 'no existe una barbería pública con ese enlace',
              instance: 'evidence-hu092-404',
              code: 'not-found',
              requestId: 'evidence-hu092-404',
            }),
          })
        },
      )
      await page.goto(`/reservar/enlace-que-no-existe/servicios/${SERVICE_ID}/barbero`)
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
