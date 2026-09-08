import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// Evidencia determinista de #191. Los datos viven solo dentro de Playwright:
// no inician sesión real ni contienen información de producción.
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'agenda-detalle',
  'fidelidad-191',
)

const detail = {
  id: 'turno-mateo',
  barberId: 'barbero-julian',
  barberFullName: 'Julián Rodríguez',
  attendeeName: 'Mateo Rojas',
  customerFullName: 'Mateo Rojas',
  customerPhone: '+57 321 456 7890',
  customerEmail: 'mateo.rojas@example.test',
  customerNote: 'Prefiere degradado bajo en los laterales.',
  startsAt: '2026-05-24T14:00:00Z',
  endsAt: '2026-05-24T14:45:00Z',
  status: 'confirmed',
  origin: 'manual',
  serviceName: 'Corte clásico',
  durationMinutes: 45,
  priceAmount: '28000.00',
  currency: 'COP',
  versionToken: 'opaque-version-token',
  createdAt: '2026-05-24T15:30:00Z',
}

const history = {
  items: [
    {
      id: 'event-created',
      eventType: 'appointment_created',
      actorType: 'system',
      actorLabel: 'Sistema',
      reason: null,
      occurredAt: '2026-05-24T15:30:00Z',
      changes: [],
    },
    {
      id: 'event-rescheduled',
      eventType: 'appointment_rescheduled',
      actorType: 'staff',
      actorLabel: 'Julián Rodríguez',
      reason: null,
      occurredAt: '2026-05-24T15:45:00Z',
      changes: [
        {
          fieldName: 'starts_at',
          previousValue: '24 may 2026, 08:00 AM',
          newValue: '24 may 2026, 09:00 AM',
        },
      ],
    },
  ],
  nextCursor: 'more-events',
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function installDetailMock(page: Page) {
  await page.route('**/api/v1/private/**', async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'shop-nava', name: 'NAVA QA Local' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
      return
    }
    if (url.pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json({ timezone: 'America/Bogota' }))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/history')) {
      await route.fulfill(json(history))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/reschedule')) {
      await route.fulfill(
        json(
          {
            type: 'https://nava.example/problems/version-conflict',
            title: 'Conflicto de versión',
            status: 409,
            detail: 'La representación del turno cambió.',
            code: 'version-conflict',
          },
          409,
        ),
      )
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo')) {
      await route.fulfill(json(detail))
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
}

async function openDetail(page: Page, width: number, height: number) {
  await page.setViewportSize({ width, height })
  await installDetailMock(page)
  await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
  await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
  await expect(page.getByText('Turno reprogramado')).toBeVisible()
  expect(await page.evaluate(() => window.innerWidth)).toBe(width)
  expect(await page.evaluate(() => window.innerHeight)).toBe(height)
}

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`fidelidad #191: detalle listo en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    )
    expect(overflow).toBe(false)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'detalle-listo.png'),
      animations: 'disabled',
    })
  })

  test(`fidelidad #191: diálogo de conflicto en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Reprogramar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Reprogramar turno' })
    await dialog.getByRole('button', { name: 'Confirmar' }).click()
    await expect(dialog.getByText('Este turno cambió mientras lo editabas')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'reprogramar-conflicto.png'),
      animations: 'disabled',
    })
  })
}

test('fidelidad #191: reflow, teclado y foco en los viewports obligatorios', async ({ page }) => {
  await installDetailMock(page)
  const viewports = [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    // El harness del repositorio aproxima el zoom de texto al 200 % con este viewport.
    { name: '1280-zoom200', width: 640, height: 450 },
  ]

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
    await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
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

    const action = page.getByRole('button', { name: 'Reprogramar turno' })
    await action.focus()
    await expect(action).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDir, 'responsive', viewport.name, 'foco-reprogramar.png'),
      animations: 'disabled',
    })
  }
})

test('fidelidad #191: movimiento reducido elimina la transición del diálogo', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openDetail(page, 1280, 900)
  await page.getByRole('button', { name: 'Reprogramar turno' }).click()
  await expect(page.getByRole('dialog', { name: 'Reprogramar turno' })).toBeVisible()
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-dialog') as HTMLElement).transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})
