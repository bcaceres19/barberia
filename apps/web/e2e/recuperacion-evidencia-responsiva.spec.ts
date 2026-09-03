import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de `/recuperar-acceso` (HU-011,
 * Fase 3 del rediseño Tailored Grid, issue #188): panel de marca en tinta
 * junto al paso 1 ("Solicita tu código") en 320/360/768/1280 px, más zoom
 * 200% aproximado (mismo criterio que `e2e/acceso-evidencia-responsiva.spec.ts`).
 * No hay evidencia previa de esta ruta; los pasos 2 y 3 (verificación y
 * nueva contraseña) requieren un código real fuera de línea y ya tienen
 * cobertura de componente propia (`RecoveryVerifyStep.test.ts`,
 * `RecoveryResetStep.test.ts`); este archivo cubre la composición visual
 * del paso alcanzable sin backend adicional. Las capturas se guardan en
 * `e2e/evidence/recuperacion/`.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'recuperacion',
)

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

for (const viewport of viewports) {
  test.describe(`Evidencia recuperación ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`paso 1 sin scroll horizontal y con foco visible (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.goto('/recuperar-acceso')
      await expect(page.getByRole('heading', { name: 'Solicita tu código' })).toBeVisible()

      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)

      await page.getByLabel('Correo').focus()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })
    })

    test(`error de validación de correo (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/recuperar-acceso')
      await page.getByLabel('Correo').fill('correo-invalido')
      await page.getByRole('button', { name: 'Enviar código' }).click()
      await expect(page.getByRole('alert').or(page.getByText(/correo válido/i))).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'error.png'),
        fullPage: true,
      })
    })
  })
}
