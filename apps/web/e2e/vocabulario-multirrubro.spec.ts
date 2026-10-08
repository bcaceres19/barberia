import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Vocabulario multirrubro en Horarios (#308, DEC-119): un negocio que llama a su
 * profesional «manicurista» y a sí mismo «estudio» no ve «barbero» ni «barbería» en
 * ninguna parte de la pantalla de Horarios, en 320/360/768/1280 px, sin scroll
 * horizontal y con axe-core. Corre con respuestas simuladas del API (no depende de
 * PostgreSQL); la persistencia del vocabulario la cubre configuracion-barberia.spec.ts.
 * Capturas en e2e/evidence/vocabulario-multirrubro/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'vocabulario-multirrubro')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const brand = {
  accent: 'brass',
  businessTerm: 'estudio',
  businessTermGender: 'masculine',
  professionalTerm: 'manicurista',
  professionalTermPlural: 'manicuristas',
  professionalTermGender: 'feminine',
  panelProfile: 'shop',
}

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function mockApi(page: Page, barbers: Array<{ id: string; fullName: string }>) {
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'unas', name: 'Estudio Lila' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
    } else if (pathname.endsWith('/settings/brand')) {
      await route.fulfill(json(brand))
    } else if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json({ timezone: 'America/Bogota' }))
    } else if (pathname.endsWith('/barbers')) {
      await route.fulfill(json({ items: barbers }))
    } else if (pathname.endsWith('/working-hours')) {
      await route.fulfill(json({ items: [], nextCursor: null }))
    } else if (pathname.endsWith('/holiday-calendar')) {
      await route.fulfill(json({ enabled: true }))
    } else if (pathname.endsWith('/schedule-exceptions')) {
      await route.fulfill(json({ items: [], nextCursor: null }))
    } else {
      await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
    }
  })
}

// El nombre propio de una persona puede contener la palabra; lo que se vigila es el
// texto de la interfaz, por eso los nombres de prueba no la llevan.
const FIXED_WORDS = /barber(o|a|os|as|ía|ías)\b/i

for (const viewport of viewports) {
  test(`Horarios con el vocabulario del negocio a ${viewport.name} px`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await mockApi(page, [
      { id: 'm-1', fullName: 'Laura Mejía' },
      { id: 'm-2', fullName: 'Paula Rojas' },
    ])
    await page.goto('/panel/horarios')
    await expect(page.getByRole('heading', { name: 'Horarios' })).toBeVisible()

    await expect(page.getByText('Horas en la zona horaria del estudio:')).toBeVisible()
    // El interruptor vive dentro de un acordeón cerrado: basta con que esté en el DOM.
    await expect(page.getByText('festivos colombianos de esta manicurista')).toBeAttached()
    await expect(page.locator('main')).not.toContainText(FIXED_WORDS)

    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    )
    expect(overflow).toBeLessThanOrEqual(0)

    await page.addScriptTag({ path: axeScriptPath })
    const violations = await page.evaluate(async () => {
      const result = await (
        window as unknown as { axe: { run: () => Promise<{ violations: unknown[] }> } }
      ).axe.run()
      return result.violations
    })
    expect(violations).toEqual([])

    await page.screenshot({
      path: path.join(evidenceDir, `horarios-${viewport.name}.png`),
      fullPage: true,
    })
  })
}

test('el estado vacío de Horarios habla de las manicuristas', async ({ page }) => {
  await page.setViewportSize({ width: 1280, height: 900 })
  await mockApi(page, [])
  await page.goto('/panel/horarios')

  await expect(page.getByText('Aún no tienes manicuristas registradas.')).toBeVisible()
  await expect(page.getByRole('link', { name: '“Manicuristas”' })).toBeVisible()
  await expect(page.locator('main')).not.toContainText(FIXED_WORDS)
})
