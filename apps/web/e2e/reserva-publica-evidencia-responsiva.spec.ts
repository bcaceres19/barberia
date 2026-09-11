import { test, expect, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-090 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): carga,
 * éxito, error "no encontrado" y foco en 320/360/768/1280 px, más una
 * aproximación de zoom de texto 200% (viewport a la mitad del ancho,
 * misma técnica que acceso-evidencia-responsiva.spec.ts). Cada solicitud
 * pública se intercepta con `page.route(...).fulfill(...)`: sin mockup
 * exacto asignado a esta pantalla (composición libre dentro de NAVA,
 * DEC-078), esta evidencia demuestra el layout real de la app, no una
 * comparación pixel a pixel contra una referencia. Las capturas se
 * guardan en `e2e/evidence/` (no ignorado por git) para quedar adjuntas
 * como archivos reales del PR.
 */
const evidenceDir = path.join(path.dirname(fileURLToPath(import.meta.url)), 'evidence', 'reserva-publica')
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
    // barberos-evidencia-responsiva.spec.ts y las pruebas de componente
    // con vitest-axe).
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

async function fulfillProfile(route: Route) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      name: 'Barbería Ejemplo',
      timezone: 'America/Bogota',
      contactEmail: 'contacto@ejemplo.test',
      contactPhone: '+573001234567',
    }),
  })
}

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`carga (${viewport.name}px)`, async ({ page }) => {
      await page.route('**/api/v1/public/barbershops/**', async (route) => {
        // 1500ms (no 400ms como en acceso-evidencia-responsiva.spec.ts):
        // esa prueba dispara la solicitud desde una SPA ya cargada (clic de
        // botón); esta navega con `page.goto` a una ruta cargada de forma
        // diferida (docs/03-desarrollo/estandar-frontend-vue.md §7.8), que
        // primero debe compilar/servir su propio chunk bajo el servidor de
        // desarrollo de Vite -mucho más lento bajo carga concurrente que
        // reutilizar un módulo ya en memoria-, así que necesita más margen
        // para que el estado "loading" siga visible cuando se comprueba.
        await new Promise((resolve) => setTimeout(resolve, 1500))
        await fulfillProfile(route)
      })
      await page.goto('/reservar/barberia-ejemplo')
      await expect(page.getByText('Abriendo tu barbería')).toBeVisible()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'carga.png'), fullPage: true })
    })

    test(`normal, foco y sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.route('**/api/v1/public/barbershops/**', fulfillProfile)
      await page.goto('/reservar/barberia-ejemplo')
      await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'normal.png'), fullPage: true })

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)

      await assertNoAxeViolations(page)
    })

    test(`error de enlace no disponible, con foco en reintentar (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.route('**/api/v1/public/barbershops/**', async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/api/v1/problems/not-found',
            title: 'No encontrado',
            status: 404,
            detail: 'no existe una barbería pública con ese enlace',
            instance: 'evidence-hu090-404',
            code: 'not-found',
            requestId: 'evidence-hu090-404',
          }),
        })
      })
      await page.goto('/reservar/enlace-que-no-existe')
      await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'error.png'), fullPage: true })

      await page.getByRole('button', { name: 'Reintentar' }).focus()
      await expect(page.getByRole('button', { name: 'Reintentar' })).toBeFocused()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png'), fullPage: true })

      await assertNoAxeViolations(page)
    })
  })
}
