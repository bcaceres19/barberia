import { test, expect, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de `/recuperar-acceso` (HU-011,
 * Fase 3 del rediseño Tailored Grid, issue #188): panel de marca en tinta
 * junto al paso 1 ("Solicita tu código") en 320/360/768/1280 px, más zoom
 * 200% aproximado (mismo criterio que `e2e/acceso-evidencia-responsiva.spec.ts`).
 * Desde `DEC-092` el paso 1 ofrece dos botones —WhatsApp y Correo—: se captura
 * sin elegir, con cada canal elegido y con el error de cada campo.
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
  { name: '420', width: 420, height: 935 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1440', width: 1440, height: 1024 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

for (const viewport of viewports) {
  test.describe(`Evidencia recuperación ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    async function expectNoHorizontalScroll(page: Page) {
      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)
    }

    test(`paso 1 sin elegir, con cada canal y con foco visible (${viewport.name}px)`, async ({
      page,
    }) => {
      await page.goto('/recuperar-acceso')
      await expect(page.getByRole('heading', { name: 'Solicita tu código' })).toBeVisible()

      // Sin elegir: dos botones y ni campo ni acción de envío (CA-011-09).
      await expect(page.getByRole('button', { name: 'WhatsApp' })).toHaveAttribute(
        'aria-pressed',
        'false',
      )
      await expect(page.getByRole('button', { name: 'Enviar código' })).toHaveCount(0)
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })
      await expectNoHorizontalScroll(page)

      // WhatsApp elegido: mensaje solo de ese canal y su campo (CA-011-10).
      await page.getByRole('button', { name: 'WhatsApp' }).click()
      await expect(page.getByRole('button', { name: 'WhatsApp' })).toHaveAttribute(
        'aria-pressed',
        'true',
      )
      await expect(page.getByText('Escribe el número de WhatsApp de tu cuenta')).toBeVisible()
      await expect(page.getByLabel('WhatsApp', { exact: true })).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'whatsapp.png'),
        fullPage: true,
      })
      await expectNoHorizontalScroll(page)

      // Correo elegido: el mensaje y el campo de WhatsApp desaparecen.
      await page.getByRole('button', { name: 'Correo' }).click()
      await expect(page.getByText('Escribe el correo de tu cuenta')).toBeVisible()
      await expect(page.getByText('Escribe el número de WhatsApp de tu cuenta')).toHaveCount(0)
      await expect(page.getByLabel('Correo', { exact: true })).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'correo.png'),
        fullPage: true,
      })
      await expectNoHorizontalScroll(page)

      await page.getByRole('button', { name: 'WhatsApp' }).focus()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })
    })

    test(`error de validación de correo (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/recuperar-acceso')
      await page.getByRole('button', { name: 'Correo' }).click()
      await page.getByLabel('Correo', { exact: true }).fill('correo-invalido')
      await page.getByRole('button', { name: 'Enviar código' }).click()
      await expect(page.getByRole('alert').or(page.getByText(/correo válido/i))).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'error.png'),
        fullPage: true,
      })
    })

    test(`error de validación de WhatsApp (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/recuperar-acceso')
      await page.getByRole('button', { name: 'WhatsApp' }).click()
      await page.getByLabel('WhatsApp', { exact: true }).fill('3001234567')
      await page.getByRole('button', { name: 'Enviar código' }).click()
      await expect(page.getByText(/indicativo de tu país/i)).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'error-whatsapp.png'),
        fullPage: true,
      })
    })
  })
}
