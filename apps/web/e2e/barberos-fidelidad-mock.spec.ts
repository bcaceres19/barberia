import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// Datos únicamente de Playwright: prueban la composición sin requerir una
// barbería, sesión ni información personal reales.
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'barberos',
  'fidelidad-192',
)

const barbers = [
  {
    id: 'julian',
    fullName: 'Julián Rodríguez',
    createdAt: '2026-05-01T10:00:00Z',
    updatedAt: '2026-05-01T10:00:00Z',
  },
  {
    id: 'luis',
    fullName: 'Luis Martínez',
    createdAt: '2026-05-02T10:00:00Z',
    updatedAt: '2026-05-02T10:00:00Z',
  },
  {
    id: 'santiago',
    fullName: 'Santiago Gómez',
    createdAt: '2026-05-03T10:00:00Z',
    updatedAt: '2026-05-03T10:00:00Z',
  },
  {
    id: 'camilo',
    fullName: 'Camilo Torres',
    createdAt: '2026-05-04T10:00:00Z',
    updatedAt: '2026-05-04T10:00:00Z',
  },
]

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function installStaffMock(page: Page, empty = false) {
  await page.route('**/api/v1/private/**', async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({ barbershop: { id: 'shop-nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
      return
    }
    if (url.pathname.endsWith('/barbers') && route.request().method() === 'GET') {
      await route.fulfill(
        json({ items: empty ? [] : barbers, nextCursor: empty ? null : 'page-2' }),
      )
      return
    }
    if (url.pathname.endsWith('/barbers') && route.request().method() === 'POST') {
      await route.abort('failed')
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
}

async function openStaff(page: Page, width: number, height: number, empty = false) {
  await page.setViewportSize({ width, height })
  await installStaffMock(page, empty)
  await page.goto('/panel/barberos')
  await expect(page.getByRole('heading', { name: 'Barberos' })).toBeVisible()
  expect(await page.evaluate(() => window.innerWidth)).toBe(width)
  expect(await page.evaluate(() => window.innerHeight)).toBe(height)
}

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`fidelidad #192: lista en ${viewport.name}`, async ({ page }) => {
    await openStaff(page, viewport.width, viewport.height)
    await expect(page.getByText('Julián Rodríguez')).toBeVisible()
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'lista.png'),
      animations: 'disabled',
    })
  })

  test(`fidelidad #192: error de alta en ${viewport.name}`, async ({ page }) => {
    await openStaff(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Agregar barbero' }).click()
    const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
    await dialog.getByLabel('Nombre').fill('Andrés Valencia')
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog.getByText('No pudimos conectar')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'alta-error.png'),
      animations: 'disabled',
    })
  })
}

test('fidelidad #192: vacío, reflow, teclado y foco en los viewports obligatorios', async ({
  page,
}) => {
  await openStaff(page, 1280, 900, true)
  await expect(page.getByText('Aún no tienes barberos registrados.')).toBeVisible()
  await page.screenshot({
    path: path.join(evidenceDir, 'desktop', 'vacio.png'),
    animations: 'disabled',
  })

  const viewports = [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    // El harness del repositorio aproxima el zoom de texto al 200 % así.
    { name: '1280-zoom200', width: 640, height: 450 },
  ]

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/barberos')
    await expect(page.getByRole('heading', { name: 'Barberos' })).toBeVisible()
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
    const action = page.locator('.staff-page__create-button')
    await action.focus()
    await expect(action).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDir, 'responsive', viewport.name, 'foco-agregar.png'),
      animations: 'disabled',
    })
  }
})

test('fidelidad #192: movimiento reducido elimina la transición del diálogo', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openStaff(page, 1280, 900)
  await page.getByRole('button', { name: 'Agregar barbero' }).click()
  await expect(page.getByRole('dialog', { name: 'Agregar barbero' })).toBeVisible()
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-dialog') as HTMLElement)
            .transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})
