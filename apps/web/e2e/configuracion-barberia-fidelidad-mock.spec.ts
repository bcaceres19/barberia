import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// El atlas usa valores ficticios; el mock conserva la evidencia aislada de
// cualquier contacto, sesión o barbería de un entorno real.
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'configuracion-barberia',
  'fidelidad-193',
)

const settings = {
  name: 'NAVA QA Local',
  timezone: 'America/Bogota',
  contactEmail: 'contacto@nava.qa',
  contactPhone: '+573001234567',
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function installSettingsMock(page: Page, patchStatus = 200) {
  await page.route('**/api/v1/private/**', async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({ barbershop: { id: 'shop-nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
      return
    }
    if (url.pathname.endsWith('/settings/barbershop') && route.request().method() === 'GET') {
      await route.fulfill(json(settings))
      return
    }
    if (url.pathname.endsWith('/settings/barbershop') && route.request().method() === 'PATCH') {
      if (patchStatus === 0) {
        await route.abort('failed')
        return
      }
      await route.fulfill(
        patchStatus === 200 ? json(settings) : json({ title: 'Datos inválidos' }, patchStatus),
      )
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
}

async function openSettings(page: Page, width: number, height: number, patchStatus = 200) {
  await page.setViewportSize({ width, height })
  await installSettingsMock(page, patchStatus)
  await page.goto('/panel/barberia')
  await expect(page.getByRole('heading', { name: 'Configuración de barbería' })).toBeVisible()
  expect(await page.evaluate(() => window.innerWidth)).toBe(width)
  expect(await page.evaluate(() => window.innerHeight)).toBe(height)
}

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`fidelidad #193: guardado en ${viewport.name}`, async ({ page }) => {
    await openSettings(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByRole('status')).toContainText('Guardado')
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'guardado.png'),
      animations: 'disabled',
    })
  })
}

test('fidelidad #193: error validable conserva el formulario', async ({ page }) => {
  await openSettings(page, 1280, 900, 422)
  await page.getByLabel('Nombre').fill('NAVA QA sin pérdida')
  await page.getByLabel('Zona horaria').fill('COT')
  await page.getByRole('button', { name: 'Guardar cambios' }).click()
  await expect(page.getByRole('alert')).toContainText('No pudimos guardar los cambios')
  await expect(page.getByLabel('Nombre')).toHaveValue('NAVA QA sin pérdida')
  await page.screenshot({
    path: path.join(evidenceDir, 'desktop', 'validacion.png'),
    animations: 'disabled',
  })
})

test('fidelidad #193: reflow, foco y zoom en los viewports obligatorios', async ({ page }) => {
  await openSettings(page, 1280, 900)
  for (const viewport of [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    { name: '1280-zoom200', width: 640, height: 450 },
  ]) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/barberia')
    expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
    expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)
    expect(
      await page.evaluate(() => {
        const content = document.querySelector('.private-shell__content')
        return (
          document.documentElement.scrollWidth > document.documentElement.clientWidth ||
          (!!content && content.scrollWidth > content.clientWidth)
        )
      }),
      `${viewport.name}: no hay desborde horizontal`,
    ).toBe(false)
    await page.getByLabel('Nombre').focus()
    await expect(page.getByLabel('Nombre')).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDir, 'responsive', viewport.name, 'foco-nombre.png'),
      animations: 'disabled',
    })
  }
})

test('fidelidad #193: movimiento reducido desactiva transiciones del formulario', async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openSettings(page, 1280, 900)
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-input') as HTMLElement).transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})
