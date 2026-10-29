import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Ubicación de la barbería en la entrada pública (HU-102, DEC-130): con dirección geocodificada
 * la página muestra «Dónde estamos» con el mapa, la dirección en texto y el enlace a Google Maps;
 * sin ubicación (interruptor apagado o sin dirección) la sección no existe. En 320/360/768/1280 px,
 * sin scroll horizontal y con axe-core. Respuestas simuladas del API y del mapa (no dependen de
 * PostgreSQL ni de Google); la persistencia y el geocodificado los cubren las pruebas Go con dos
 * tenants. Capturas en e2e/evidence/ubicacion-reserva-publica/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'ubicacion-reserva-publica')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const SLUG = 'barberiacentro7k3m9q2x'
const ADDRESS = 'Calle 10 # 5-20, Bogotá'

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function mockProfile(page: Page, location: unknown) {
  await page.route(`**/api/v1/public/barbershops/${SLUG}`, (route) =>
    route.fulfill(
      json({
        name: 'Barbería Centro',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
        vocabulary: {
          businessTerm: 'barbería',
          businessTermGender: 'feminine',
          professionalTerm: 'barbero',
          professionalTermPlural: 'barberos',
          professionalTermGender: 'masculine',
        },
        location,
      }),
    ),
  )
  // El mapa real vive en el origen de Google: aquí solo se comprueba que la página lo pide.
  await page.route(/google\.com\/maps/, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'text/html',
      body: '<!doctype html><title>Mapa simulado</title><body style="margin:0;background:#dde">mapa</body>',
    }),
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
      window as unknown as {
        axe: { run: (ctx: unknown, opts: unknown) => Promise<{ violations: unknown[] }> }
      }
    ).axe.run(document, { iframes: false })
    return result.violations
  })
  expect(violations).toEqual([])
}

for (const viewport of viewports) {
  test(`muestra «Dónde estamos» con el mapa a ${viewport.name} px`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await mockProfile(page, { address: ADDRESS, latitude: 4.598056, longitude: -74.076111 })
    await page.goto(`/reservar/${SLUG}`)

    const section = page.locator('.pb-location')
    await expect(section.getByRole('heading', { name: 'Dónde estamos' })).toBeVisible()
    await expect(section.getByText(ADDRESS)).toBeVisible()
    await expect(section.locator('iframe')).toHaveAttribute('title', `Mapa de ${ADDRESS}`)
    await expect(section.getByRole('link', { name: /Abrir en Google Maps/ })).toHaveAttribute(
      'href',
      /google\.com\/maps/,
    )

    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `con-ubicacion-${viewport.name}.png`),
      fullPage: true,
    })
  })

  test(`no pinta la sección cuando no hay ubicación a ${viewport.name} px`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await mockProfile(page, null)
    await page.goto(`/reservar/${SLUG}`)

    await expect(page.getByRole('heading', { name: 'Barbería Centro' })).toBeVisible()
    await expect(page.locator('.pb-location')).toHaveCount(0)
    await expect(page.locator('iframe')).toHaveCount(0)

    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `sin-ubicacion-${viewport.name}.png`),
      fullPage: true,
    })
  })
}
