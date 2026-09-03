import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-010 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): normal,
 * foco, carga, error y bloqueado en 320/360/768/1280 px, más una
 * aproximación de zoom de texto 200% (viewport a la mitad del ancho,
 * técnica estándar de Playwright para simular zoom de navegador: el
 * layout resultante en la mitad del ancho es equivalente al de zoom 200%
 * en el ancho completo). Las capturas se guardan en `e2e/evidence/` (no
 * ignorado por git, a diferencia de `playwright-report/`/`test-results/`)
 * para quedar adjuntas como archivos reales del PR, no solo descritas.
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

const evidenceDir = path.join(path.dirname(fileURLToPath(import.meta.url)), 'evidence')

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  // Aproximación de zoom 200% sobre 1280: mitad de ancho.
  { name: '1280-zoom200', width: 640, height: 450 },
]

for (const viewport of viewports) {
  test.describe(`Evidencia ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test(`normal, foco y sin scroll horizontal (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/acceso')
      // Espera el contenido real (no solo el evento `load` de la
      // navegación) antes de la primera captura: la ruta se carga de
      // forma diferida y el primer paint puede llegar después de `load`.
      await expect(page.getByRole('heading', { name: 'Accede a NAVA' })).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'normal.png'),
        fullPage: true,
      })

      await page.getByLabel('Correo').focus()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'foco.png'),
      })

      // CA-010-06: sin desplazamiento horizontal en 320/360; se verifica
      // en todos los anchos capturados por consistencia.
      const hasHorizontalScroll = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(hasHorizontalScroll).toBe(false)
    })

    test(`carga (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/acceso')
      await page.getByLabel('Correo').fill(EMAIL)
      await page.getByLabel('Contraseña').fill(PASSWORD)

      await page.route('**/api/v1/public/auth/login', async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 400))
        await route.continue()
      })
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByRole('button', { name: 'Iniciando sesión…' })).toBeVisible()
      await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'carga.png') })
    })

    test(`error de credenciales (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/acceso')
      await page.getByLabel('Correo').fill(EMAIL)
      await page.getByLabel('Contraseña').fill('clave-incorrecta')
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByRole('alert').last()).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'error.png'),
        fullPage: true,
      })
    })

    test(`bloqueado por 429 (${viewport.name}px)`, async ({ page }) => {
      await page.goto('/acceso')
      await page.getByLabel('Correo').fill(EMAIL)
      await page.getByLabel('Contraseña').fill(PASSWORD)
      await page.route('**/api/v1/public/auth/login', async (route) => {
        await route.fulfill({
          status: 429,
          contentType: 'application/problem+json',
          headers: { 'Retry-After': '120' },
          body: JSON.stringify({
            title: 'Demasiadas solicitudes',
            status: 429,
            code: 'rate-limited',
            instance: 'evidence-429',
            requestId: 'evidence-429',
          }),
        })
      })
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByText('Demasiados intentos')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'bloqueado.png'),
        fullPage: true,
      })
    })
  })
}
