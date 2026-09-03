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
 *
 * Un solo inicio de sesión reutilizado entre los cinco breakpoints por
 * prueba (mismo criterio que `e2e/barberos-evidencia-responsiva.spec.ts`):
 * el umbral de intentos de HU-007 es compartido por todo este entorno de
 * pruebas. La versión anterior de este archivo iniciaba sesión una vez por
 * combinación prueba×ancho (20 inicios en segundos) y quedaba bloqueada por
 * la propia defensa contra abuso al ejecutarse localmente.
 *
 * Fase 2 del rediseño Tailored Grid (issue #187): el dock ya no lista los
 * seis destinos en línea plana; solo Agenda/Servicios/Barberos quedan
 * siempre visibles y el resto vive dentro de "Más" (BaseDialog). Se agrega
 * cobertura de ese menú (destinos secundarios, incluido "Bloqueos", que
 * antes no tenía entrada de navegación, y cierre de sesión).
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

async function login(page: import('@playwright/test').Page) {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
  await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
}

test('cascarón autenticado sin scroll horizontal ni pérdida de foco, en cada breakpoint', async ({
  page,
}) => {
  await login(page)

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

    await page.getByRole('link', { name: 'Agenda' }).focus()
    await expect(page.getByRole('link', { name: 'Agenda' })).toBeFocused()
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })
  }
})

test('el dock solo expone Agenda/Servicios/Barberos y "Más"; "Más" expone el resto y cierre de sesión, en cada breakpoint', async ({
  page,
}) => {
  await login(page)

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })

    const dock = page.getByRole('navigation', { name: 'Navegación principal' })
    await expect(dock.getByRole('link')).toHaveText(['Agenda', 'Servicios', 'Barberos'])

    await dock.getByRole('button', { name: 'Más' }).click()
    const dialog = page.getByRole('dialog', { name: 'Más' })
    await expect(dialog.getByRole('link', { name: 'Horarios' })).toBeVisible()
    await expect(dialog.getByRole('link', { name: 'Bloqueos' })).toBeVisible()
    await expect(dialog.getByRole('link', { name: 'Configuración' })).toBeVisible()
    await expect(dialog.getByRole('link', { name: 'Servicios por barbero' })).toBeVisible()
    await expect(dialog.getByRole('button', { name: 'Cerrar sesión' })).toBeVisible()
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'mas-abierto.png') })

    await page.keyboard.press('Escape')
    await expect(dialog).toBeHidden()
    await expect(dock.getByRole('button', { name: 'Más' })).toBeFocused()

    // Bloqueos no tenía entrada de navegación antes de esta fase: confirma
    // que ahora es alcanzable de punta a punta y no solo por URL directa.
    await dock.getByRole('button', { name: 'Más' }).click()
    await dialog.getByRole('link', { name: 'Bloqueos' }).click()
    await expect(page).toHaveURL(/\/panel\/bloqueos$/)
    await expect(dialog).toBeHidden()

    await page.goto('/panel')
    await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
  }
})
