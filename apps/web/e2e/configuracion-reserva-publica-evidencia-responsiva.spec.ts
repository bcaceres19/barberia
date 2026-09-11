import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-093 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): pantalla
 * "Reserva pública" en 320/360/768/1280 px, más zoom 200% aproximado
 * (mismo criterio que
 * `e2e/configuracion-barberia-evidencia-responsiva.spec.ts`, HU-020). Se
 * inicia sesión UNA sola vez y se redimensiona la MISMA página entre
 * capturas. Las capturas se guardan en
 * `e2e/evidence/configuracion-reserva-publica/`.
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'configuracion-reserva-publica',
)

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

test('pantalla Reserva pública sin scroll horizontal y con foco visible por teclado en cada breakpoint', async ({
  page,
}) => {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
  await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel/)

  await page.getByRole('link', { name: 'Reserva pública' }).click()
  await expect(
    page.getByRole('heading', { name: 'Configuración de reserva pública' }),
  ).toBeVisible()

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })

    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'normal.png'),
      fullPage: true,
    })

    const hasHorizontalScroll = await page.evaluate(
      () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
    )
    expect(hasHorizontalScroll).toBe(false)

    await page.getByLabel('Anticipación mínima (minutos)').focus()
    await expect(page.getByLabel('Anticipación mínima (minutos)')).toBeFocused()
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })
  }
})
