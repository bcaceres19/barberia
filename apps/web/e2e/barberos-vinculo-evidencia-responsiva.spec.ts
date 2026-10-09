import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia responsive/accesible del vínculo «Este soy yo» (#322, DEC-100) en la ficha de
 * un barbero: sin vincular, vinculado y el aviso de un barbero ya tomado, en 320/360/768/1280
 * px y en Tinta y Marfil, sin scroll horizontal y con axe-core (incluido el contraste real).
 * Corre con respuestas simuladas del API que conservan el vínculo entre solicitudes; las capturas
 * van a e2e/evidence/barberos-vinculo/. El recorrido contra la base real vive en las pruebas
 * de integración de Go (cmd/api/staff_link_integration_test.go).
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'barberos-vinculo')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const brand = {
  accent: 'brass',
  businessTerm: 'barbería',
  businessTermGender: 'feminine',
  professionalTerm: 'barbero',
  professionalTermPlural: 'barberos',
  professionalTermGender: 'masculine',
  panelProfile: 'shop',
}
const settings = {
  name: 'Barbería de prueba',
  timezone: 'America/Bogota',
  contactEmail: 'equipo@ejemplo.test',
  contactPhone: '+573001234567',
}
const barberRow = (id: string, fullName: string) => ({
  id,
  fullName,
  createdAt: '2026-08-23T15:04:05Z',
  updatedAt: '2026-08-23T15:04:05Z',
  photoUpdatedAt: null,
})
const barbers = [barberRow('b-mateo', 'Mateo Rojas'), barberRow('b-lucia', 'Lucía Pérez')]
// Lucía pertenece a otra persona: pedirla responde 409.
const TAKEN_BARBER = 'b-lucia'

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function mockApi(page: Page, initialLinked: string | null) {
  let linked = initialLinked
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    const method = route.request().method()
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'demo', name: 'Barbería de prueba' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
    } else if (pathname.endsWith('/settings/brand')) {
      await route.fulfill(json(brand))
    } else if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json(settings))
    } else if (pathname.endsWith('/me/barber')) {
      if (method === 'GET') {
        const mine = barbers.find((b) => b.id === linked)
        await route.fulfill(
          mine ? json(mine) : json({ title: 'No encontrado', status: 404, code: 'not_found' }, 404),
        )
      } else if (method === 'PUT') {
        const { barberId } = route.request().postDataJSON() as { barberId: string }
        if (barberId === TAKEN_BARBER) {
          await route.fulfill(json({ title: 'Conflicto', status: 409, code: 'conflict' }, 409))
        } else {
          linked = barberId
          await route.fulfill(json(barbers.find((b) => b.id === barberId)))
        }
      } else {
        linked = null
        await route.fulfill({ status: 204 })
      }
    } else if (pathname.endsWith('/barbers')) {
      await route.fulfill(
        json({
          items: barbers,
          page: 1,
          pageSize: 20,
          total: barbers.length,
          totalPages: 1,
          nextCursor: null,
        }),
      )
    } else {
      await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
    }
  })
}

async function openDetail(
  page: Page,
  name: string,
  width: number,
  height: number,
  theme: 'ink' | 'ivory',
  initialLinked: string | null,
) {
  await page.setViewportSize({ width, height })
  await page.addInitScript((stored) => {
    window.localStorage.setItem(
      'nava.appearance.v1',
      JSON.stringify({ theme: stored, textScale: 'normal', motion: 'reduced' }),
    )
  }, theme)
  await mockApi(page, initialLinked)
  await page.goto('/panel/barberos')
  await page
    .getByRole('button', { name: new RegExp(name) })
    .first()
    .click()
  await expect(page.getByRole('dialog', { name })).toBeVisible()
  // La entrada escalonada de la ficha termina antes de medir o capturar.
  await page.waitForTimeout(900)
}

async function expectNoHorizontalScroll(page: Page) {
  const overflow = await page.evaluate(() => ({
    document: document.documentElement.scrollWidth > document.documentElement.clientWidth,
    dialog: (() => {
      const d = document.querySelector(
        '.base-dialog--open .base-dialog__panel',
      ) as HTMLElement | null
      return d ? d.scrollWidth > d.clientWidth : false
    })(),
  }))
  expect(overflow).toEqual({ document: false, dialog: false })
}

async function axeViolations(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  return page.evaluate(async () => {
    const results = await (
      window as unknown as {
        axe: {
          run: (
            c: Document,
            o: object,
          ) => Promise<{ violations: Array<{ id: string; nodes: Array<{ target: unknown }> }> }>
        }
      }
    ).axe.run(document, {
      runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
    })
    return results.violations.map((v) => ({ id: v.id, targets: v.nodes.map((n) => n.target) }))
  })
}

const combos = [
  ...viewports.map((viewport) => ({ theme: 'ink' as const, viewport })),
  { theme: 'ivory' as const, viewport: viewports[1] },
  { theme: 'ivory' as const, viewport: viewports[3] },
]

for (const { theme, viewport } of combos) {
  test(`${theme} a ${viewport.name}: sin vincular, vinculado y barbero ya tomado`, async ({
    page,
  }) => {
    await openDetail(page, 'Mateo Rojas', viewport.width, viewport.height, theme, null)
    const dialog = page.getByRole('dialog', { name: 'Mateo Rojas' })
    await expect(dialog).toContainText('Sin vincular')

    await page.screenshot({
      path: path.join(evidenceDir, theme, viewport.name, 'sin-vincular.png'),
      fullPage: true,
    })
    await expectNoHorizontalScroll(page)
    expect(await axeViolations(page)).toEqual([])

    await dialog.getByRole('button', { name: 'Este soy yo' }).click()
    await expect(dialog).toContainText('Eres tú')
    await expect(dialog.getByRole('button', { name: 'Ya no soy yo' })).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDir, theme, viewport.name, 'vinculado.png'),
      fullPage: true,
    })
    await expectNoHorizontalScroll(page)
    expect(await axeViolations(page)).toEqual([])

    await dialog.getByRole('button', { name: 'Ya no soy yo' }).click()
    await expect(dialog).toContainText('Sin vincular')
  })

  test(`${theme} a ${viewport.name}: un barbero de otra persona explica el conflicto`, async ({
    page,
  }) => {
    await openDetail(page, 'Lucía Pérez', viewport.width, viewport.height, theme, null)
    const dialog = page.getByRole('dialog', { name: 'Lucía Pérez' })
    await dialog.getByRole('button', { name: 'Este soy yo' }).click()
    await expect(dialog.getByRole('alert')).toContainText('ya está vinculado a otro usuario')

    await page.screenshot({
      path: path.join(evidenceDir, theme, viewport.name, 'conflicto.png'),
      fullPage: true,
    })
    await expectNoHorizontalScroll(page)
    expect(await axeViolations(page)).toEqual([])
    // El botón sigue disponible para reintentar y el foco no se pierde.
    await expect(dialog.getByRole('button', { name: 'Este soy yo' })).toBeEnabled()
  })
}
