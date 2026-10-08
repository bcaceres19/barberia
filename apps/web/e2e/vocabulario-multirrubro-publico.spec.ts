import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Vocabulario multirrubro en lo que ve el cliente (#309, DEC-119): un estudio que llama
 * «manicurista» a su profesional no muestra «barbero» ni «barbería» en la selección pública
 * ni en el acceso al turno, en 320/360/768/1280 px, sin scroll horizontal y con axe-core.
 * Respuestas simuladas del API (no depende de PostgreSQL); la lectura real del vocabulario
 * la cubren las pruebas Go con dos tenants. Capturas en e2e/evidence/vocabulario-multirrubro/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'vocabulario-multirrubro')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const SLUG = 'estudiolilak7x2m9q4'
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const TOKEN = 'tokenDePruebaDelTurno0123456789abcdefghijk'

const nails = {
  businessTerm: 'estudio de uñas',
  businessTermGender: 'masculine',
  professionalTerm: 'manicurista',
  professionalTermPlural: 'manicuristas',
  professionalTermGender: 'feminine',
}

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

// Los nombres propios no llevan la palabra: lo que se vigila es el texto de la interfaz.
const FIXED_WORDS = /barber(o|a|os|as|ía|ías)\b/i

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function mockPublicApi(page: Page) {
  await page.route(`**/api/v1/public/barbershops/${SLUG}`, (route) =>
    route.fulfill(
      json({
        name: 'Estudio Lila',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
        vocabulary: nails,
      }),
    ),
  )
  await page.route(
    `**/api/v1/public/barbershops/${SLUG}/services/${SERVICE_ID}/barbers**`,
    (route) =>
      route.fulfill(
        json({
          items: [
            { id: 'a1111111-1111-1111-1111-111111111111', fullName: 'Laura Mejía' },
            { id: 'b2222222-2222-2222-2222-222222222222', fullName: 'Paula Rojas' },
          ],
        }),
      ),
  )
}

async function expectAccessibleWithoutOverflow(page: Page) {
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
}

for (const viewport of viewports) {
  test(`selección pública con el vocabulario del negocio a ${viewport.name} px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await mockPublicApi(page)
    await page.goto(`/reservar/${SLUG}/servicios/${SERVICE_ID}/barbero`)

    await expect(page.getByRole('heading', { name: 'Elige tu manicurista' })).toBeVisible()
    await expect(page.getByRole('radiogroup', { name: 'Manicuristas disponibles' })).toBeVisible()
    await expect(page.locator('.pb-progress')).toContainText('Manicurista')
    await expect(page.locator('body')).not.toContainText(FIXED_WORDS)

    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `reserva-publica-${viewport.name}.png`),
      fullPage: true,
    })
  })

  test(`acceso al turno con el vocabulario del negocio a ${viewport.name} px`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.route(`**/api/v1/customer/appointments/${TOKEN}`, (route) =>
      route.fulfill(
        json({
          barbershopName: 'Estudio Lila',
          timezone: 'America/Bogota',
          attendeeName: 'Ana Ríos',
          serviceName: 'Manicure semipermanente',
          durationMinutes: 60,
          barberName: 'Laura Mejía',
          startsAt: '2027-03-15T15:00:00Z',
          endsAt: '2027-03-15T16:00:00Z',
          status: 'cancelled_by_barber',
          cancellationDeadlineMinutes: 20,
          lateCancellationClientAllowed: false,
          lateCancellationReasonRequired: false,
          vocabulary: nails,
        }),
      ),
    )
    await page.goto(`/mi-turno/${TOKEN}`)

    await expect(page.getByText('Cancelado por la manicurista')).toBeVisible()
    await expect(page.getByText('contacta directamente al estudio de uñas')).toBeVisible()
    await expect(page.locator('body')).not.toContainText(FIXED_WORDS)

    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `mi-turno-${viewport.name}.png`),
      fullPage: true,
    })
  })
}
