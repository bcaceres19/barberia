import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-096 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): formulario
 * inicial, error de validación conservando datos, resumen para sí mismo y
 * resumen para otra persona, en 320/360/768/1280 px, más una aproximación
 * de zoom de texto 200% (mismo criterio que
 * reserva-publica-horario-evidencia-responsiva.spec.ts, HU-095). Sin
 * mockup exacto asignado a esta pantalla (composición libre dentro de
 * NAVA, DEC-078). Las capturas se guardan en `e2e/evidence/` (no ignorado
 * por git) para quedar adjuntas como archivos reales del PR.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'reserva-publica-cliente',
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
    // reserva-publica-horario-evidencia-responsiva.spec.ts).
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
const STARTS_AT = encodeURIComponent('2026-09-20T14:30:00Z')
const ROUTE = `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario/${STARTS_AT}/cliente`

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`formulario inicial, sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.goto(ROUTE)
      await expect(page.getByLabel('Tu nombre')).toBeVisible()

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)

      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'formulario.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`errores conservando datos escritos (${viewport.name}px)`, async ({ page }) => {
      await page.goto(ROUTE)
      await page.getByLabel('Tu nombre').fill('Ana Ríos')
      await page.getByRole('button', { name: 'Ver resumen' }).click()

      await expect(page.getByText('El teléfono es obligatorio.')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'errores.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test(`resumen para otra persona (${viewport.name}px)`, async ({ page }) => {
      await page.goto(ROUTE)
      await page.getByLabel('Tu nombre').fill('Ana Ríos')
      await page.getByLabel('Teléfono').fill('+573001234567')
      await page.getByLabel('Correo').fill('ana@example.com')
      await page.getByLabel('Para otra persona').check()
      await page.getByLabel('Nombre de la persona atendida').fill('Mateo Ruiz')
      await page.getByRole('button', { name: 'Ver resumen' }).click()

      await expect(page.getByRole('status')).toContainText('Mateo Ruiz')
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'resumen.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })
  })
}
