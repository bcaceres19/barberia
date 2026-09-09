import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'servicios-por-barbero',
  'fidelidad-195',
)

const barber = { id: 'barber-nava', fullName: 'Barbero NAVA' }
const services = [
  { id: 'corte', name: 'Corte clásico' },
  { id: 'barba', name: 'Corte + barba' },
  { id: 'fade', name: 'Fade premium' },
  { id: 'solo-barba', name: 'Barba' },
]
const initiallyAssigned = new Set(['corte', 'barba', 'fade'])

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function openAssignments(page: Page, width: number, height: number) {
  await page.setViewportSize({ width, height })
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({ barbershop: { id: 'nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
      return
    }
    if (pathname.endsWith('/barbers') && route.request().method() === 'GET') {
      await route.fulfill(json({ items: [barber], nextCursor: null }))
      return
    }
    if (pathname.endsWith('/barbers/barber-nava/services') && route.request().method() === 'GET') {
      await route.fulfill(
        json({
          items: [...initiallyAssigned].map((serviceId) => ({
            barberId: barber.id,
            serviceId,
            createdAt: '2026-09-08T12:00:00Z',
          })),
          nextCursor: null,
        }),
      )
      return
    }
    if (pathname.endsWith('/services') && route.request().method() === 'GET') {
      await route.fulfill(json({ items: services, nextCursor: null }))
      return
    }
    if (
      pathname.endsWith('/barbers/barber-nava/services/corte') &&
      route.request().method() === 'DELETE'
    ) {
      await route.fulfill(json({ code: 'conflict' }, 409))
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
  await page.goto('/panel/servicios-por-barbero')
  await expect(page.getByRole('heading', { name: 'Servicios por barbero' })).toBeVisible()
  await expect(page.getByRole('checkbox', { name: 'Corte clásico' })).toBeChecked()
}

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`fidelidad #195: asignaciones en ${viewport.name}`, async ({ page }) => {
    await openAssignments(page, viewport.width, viewport.height)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'asignaciones.png'),
      animations: 'disabled',
    })
  })
}

test('fidelidad #195: conflicto, foco y reflow obligatorio', async ({ page }) => {
  await openAssignments(page, 1280, 900)
  await page.getByRole('checkbox', { name: 'Corte clásico' }).click()
  await expect(page.getByRole('alert')).toContainText(
    'No puedes retirar la última asignación activa de este servicio',
  )
  await expect(page.getByRole('checkbox', { name: 'Corte clásico' })).toBeChecked()
  await page.screenshot({
    path: path.join(evidenceDir, 'desktop', 'conflicto.png'),
    animations: 'disabled',
  })

  for (const viewport of [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    { name: '1280-zoom200', width: 640, height: 450 },
  ]) {
    await page.setViewportSize(viewport)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await page.getByLabel('Barbero', { exact: true }).focus()
    await expect(page.getByLabel('Barbero', { exact: true })).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDir, 'responsive', viewport.name, 'foco-barbero.png'),
      animations: 'disabled',
    })
  }
})
