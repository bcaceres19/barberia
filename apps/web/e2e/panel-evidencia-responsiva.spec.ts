import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible del cascarón autenticado (HU-012) y,
 * desde HU-062, de la agenda diaria real que ahora abre en `/panel`
 * (docs/03-desarrollo/estandar-diseno-visual.md §16 y estrategia-pruebas.md
 * §5.4): foco de navegación en 320/360/768/1280 px, más zoom 200%
 * aproximado (mismo criterio que `e2e/acceso-evidencia-responsiva.spec.ts`).
 * Las capturas se guardan en `e2e/evidence/panel/` (no ignorado por git).
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

const evidenceDir = path.join(path.dirname(fileURLToPath(import.meta.url)), 'evidence', 'panel')

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

for (const viewport of viewports) {
  test.describe(`Evidencia panel ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`cascarón autenticado sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/acceso')
      await page.getByLabel('Correo').fill(EMAIL)
      await page.getByLabel('Contraseña').fill(PASSWORD)
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()

      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })

      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)
    })

    test(`navegación operable con teclado y foco visible (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.goto('/acceso')
      await page.getByLabel('Correo').fill(EMAIL)
      await page.getByLabel('Contraseña').fill(PASSWORD)
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()

      await page.getByRole('link', { name: 'Panel' }).focus()
      await expect(page.getByRole('link', { name: 'Panel' })).toBeFocused()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })
    })
  })
}
