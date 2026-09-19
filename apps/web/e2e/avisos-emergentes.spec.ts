import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Avisos emergentes tipo acordeón (DEC-095): recorrido "crear servicio" con la
 * API simulada, para no depender de PostgreSQL. Comprueba el aviso de
 * confirmación, el acordeón, la pausa del tiempo, su desaparición y la
 * ausencia de scroll horizontal y de violaciones axe en 320/360/768/1280 px
 * y zoom 200 %. Capturas en `e2e/evidence/avisos-emergentes/`.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'avisos-emergentes')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

function service(id: string, name: string, durationMinutes: number, price: string) {
  return {
    id,
    name,
    description: null,
    durationMinutes,
    price,
    currency: 'COP',
    isActive: true,
    deactivatedAt: null,
    createdAt: '2026-09-01T12:00:00Z',
    updatedAt: '2026-09-01T12:00:00Z',
  }
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function openCatalog(page: Page, width: number, height: number) {
  // El reloj simulado permite avanzar los 5 s de la confirmación sin esperar.
  await page.clock.install()
  await page.setViewportSize({ width, height })
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    const method = route.request().method()
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({ barbershop: { id: 'nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
      return
    }
    if (pathname.endsWith('/services') && method === 'GET') {
      await route.fulfill(
        json({ items: [service('corte', 'Corte clásico', 45, '28000.00')], nextCursor: null }),
      )
      return
    }
    if (pathname.endsWith('/services') && method === 'POST') {
      await route.fulfill(json(service('nuevo', 'Fade premium', 50, '35000.00'), 201))
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
  await page.goto('/panel/servicios')
  await expect(page.getByRole('heading', { name: 'Servicios' })).toBeVisible()
}

async function createService(page: Page) {
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
  await dialog.getByLabel('Nombre').fill('Fade premium')
  await dialog.getByLabel('Duración (minutos)').fill('50')
  await dialog.getByLabel('Precio (COP)').fill('35000.00')
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()
}

test('la confirmación aparece cerrada, se abre como acordeón, se pausa y desaparece sola', async ({
  page,
}) => {
  await openCatalog(page, 1280, 900)
  await createService(page)

  const region = page.getByRole('region', { name: 'Avisos' })
  const toast = region.getByRole('status')
  await expect(toast).toBeVisible()
  await expect(toast).toContainText('Confirmación')
  await expect(toast).toContainText('Servicio creado')
  await expect(toast.getByRole('button', { name: /Servicio creado/ })).toHaveAttribute(
    'aria-expanded',
    'false',
  )

  // El aviso nunca roba el foco: al cerrarse el diálogo vuelve a su disparador.
  await expect(page.getByRole('button', { name: 'Agregar servicio' })).toBeFocused()
  await expect(region.locator(':focus')).toHaveCount(0)
  await expect(toast.getByRole('button', { name: 'Descartar' })).toBeHidden()

  await toast.getByRole('button', { name: /Servicio creado/ }).click()
  await expect(toast.getByRole('button', { name: /Servicio creado/ })).toHaveAttribute(
    'aria-expanded',
    'true',
  )
  await expect(toast.getByText('«Fade premium» ya aparece en tu catálogo.')).toBeVisible()

  // Abierto, la cuenta atrás se detiene aunque pase mucho más que 5 s.
  await page.clock.runFor(30_000)
  await expect(toast).toBeVisible()

  // Cerrado otra vez y con el cursor fuera, corre lo que faltaba de los 5 s. Un
  // clic de ratón no deja el aviso retenido por foco.
  await toast.getByRole('button', { name: /Servicio creado/ }).click()
  await page.mouse.move(0, 0)
  await page.clock.runFor(5_100)
  await expect(toast).toBeHidden()
})

test('Escape descarta el aviso con foco y Descartar todos vacía la pila', async ({ page }) => {
  await openCatalog(page, 1280, 900)
  await createService(page)
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
  await dialog.getByLabel('Nombre').fill('Fade premium')
  await dialog.getByLabel('Duración (minutos)').fill('50')
  await dialog.getByLabel('Precio (COP)').fill('35000.00')
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()

  const region = page.getByRole('region', { name: 'Avisos' })
  await expect(region.getByRole('status')).toHaveCount(2)
  await expect(region.getByRole('button', { name: /Descartar todos/ })).toBeVisible()

  await region
    .getByRole('button', { name: /Servicio creado/ })
    .first()
    .focus()
  await page.keyboard.press('Escape')
  await expect(region.getByRole('status')).toHaveCount(1)

  await createService(page)
  await region.getByRole('button', { name: /Descartar todos/ }).click()
  await expect(region.getByRole('status')).toHaveCount(0)
})

test('reflow, foco visible y axe con la confirmación abierta en cada ancho', async ({ page }) => {
  await openCatalog(page, 1280, 900)
  await createService(page)
  const toast = page.getByRole('region', { name: 'Avisos' }).getByRole('status')
  await toast.getByRole('button', { name: /Servicio creado/ }).click()
  await page.addScriptTag({ path: axeScriptPath })

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await expect(toast).toBeVisible()

    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)

    // El aviso queda dentro del escenario del shell: entre cabecera y navegación.
    const box = await toast.boundingBox()
    expect(box).not.toBeNull()
    expect(box!.x).toBeGreaterThanOrEqual(0)
    expect(box!.x + box!.width).toBeLessThanOrEqual(viewport.width)

    await toast.getByRole('button', { name: 'Descartar' }).focus()
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'abierto-con-foco.png'),
      animations: 'disabled',
    })

    const violations = await page.evaluate(async () => {
      // @ts-expect-error axe se inyecta con addScriptTag
      const result = await window.axe.run(document.querySelector('[aria-label="Avisos"]'), {
        rules: { region: { enabled: false } },
      })
      return result.violations.map((v: { id: string }) => v.id)
    })
    expect(violations).toEqual([])
  }
})

test('respeta prefers-reduced-motion: sin animación de entrada ni traslación', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openCatalog(page, 1280, 900)
  await createService(page)

  const toast = page.getByRole('region', { name: 'Avisos' }).getByRole('status')
  await expect(toast).toBeVisible()
  const animation = await toast.evaluate((node) => getComputedStyle(node).animationName)
  expect(animation).toBe('none')
})
