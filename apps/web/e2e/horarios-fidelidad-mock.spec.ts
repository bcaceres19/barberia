import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'horarios',
  'fidelidad-196',
)

const workingHours = [
  workingHour('m-1', 1, '09:00', 240),
  workingHour('m-2', 1, '14:00', 240),
  workingHour('t-1', 2, '09:00', 480),
  workingHour('j-1', 4, '10:00', 240),
  workingHour('j-2', 4, '15:00', 180),
  workingHour('v-1', 5, '09:00', 480),
]

function workingHour(id: string, isoWeekday: number, startsTime: string, durationMinutes: number) {
  return {
    id,
    isoWeekday,
    startsTime,
    durationMinutes,
    createdAt: '2026-09-09T12:00:00Z',
    updatedAt: '2026-09-09T12:00:00Z',
  }
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function openSchedules(page: Page, width: number, height: number) {
  await page.setViewportSize({ width, height })
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({ barbershop: { id: 'nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
      return
    }
    if (pathname.endsWith('/barbers')) {
      await route.fulfill(json({ items: [{ id: 'barber-nava', fullName: 'Barbero NAVA' }] }))
      return
    }
    if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json({ timezone: 'America/Bogota' }))
      return
    }
    if (pathname.endsWith('/working-hours')) {
      await route.fulfill(json({ items: workingHours, nextCursor: null }))
      return
    }
    if (pathname.endsWith('/holiday-calendar')) {
      await route.fulfill(json({ enabled: true }))
      return
    }
    if (pathname.endsWith('/schedule-exceptions')) {
      await route.fulfill(
        json({
          items: [
            {
              id: 'exception-1',
              effectiveDate: '2099-12-08',
              isClosed: true,
              reason: 'Capacitación interna',
              segments: [],
              createdAt: '2026-09-09T12:00:00Z',
              updatedAt: '2026-09-09T12:00:00Z',
            },
          ],
          nextCursor: null,
        }),
      )
      return
    }
    if (pathname.endsWith('/schedule/colombian-holidays')) {
      await route.fulfill(json({ items: [{ date: '2099-12-25', name: 'Navidad' }] }))
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
  await page.goto('/panel/horarios')
  await expect(page.getByRole('heading', { name: 'Horarios' })).toBeVisible()
  await expect(page.getByText('09:00 · 240 min').first()).toBeVisible()
}

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`fidelidad #196: horarios en ${viewport.name}`, async ({ page }) => {
    await openSchedules(page, viewport.width, viewport.height)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'horarios.png'),
      fullPage: true,
      animations: 'disabled',
    })
  })
}

test('fidelidad #196: foco y reflow obligatorio', async ({ page }) => {
  await openSchedules(page, 1280, 900)
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
    await page.getByRole('button', { name: 'Agregar tramo' }).focus()
    await expect(page.getByRole('button', { name: 'Agregar tramo' })).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDir, 'responsive', viewport.name, 'foco-agregar.png'),
      animations: 'disabled',
    })
  }
})
