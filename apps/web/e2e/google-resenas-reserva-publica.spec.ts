import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { INITIAL_VOCABULARY } from './fixtures/vocabularioInicial'

/**
 * Reseñas y estrellas de Google en la entrada pública (HU-101, DEC-129). Respuestas simuladas del
 * API (la consulta real a Google, su caché de 6 horas y la clave de servidor las cubren las pruebas
 * Go con un Google falso): con ficha y datos la página muestra la calificación, la atribución y
 * las reseñas bajo demanda; sin ficha, con la estructura vacía o ante un fallo del servidor, la
 * sección no existe y no aparece ningún error. En 320/360/768/1280 px, sin scroll horizontal y con
 * axe-core. Capturas en e2e/evidence/google-resenas-reserva-publica/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'google-resenas-reserva-publica')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const SLUG = 'barberia-ejemplo'

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

const withListing = {
  rating: 4.7,
  ratingCount: 128,
  mapsUrl: 'https://maps.google.com/?cid=1234567890',
  reviews: [
    {
      authorName: 'Camila R.',
      authorUrl: 'https://www.google.com/maps/contrib/100000000000000000000',
      rating: 5,
      text: 'Excelente atención y muy puntuales.',
      relativeTime: 'hace un mes',
    },
    {
      authorName: 'Andrés M.',
      authorUrl: null,
      rating: 4,
      text: null,
      relativeTime: 'hace 3 meses',
    },
  ],
}

const empty = { rating: null, ratingCount: 0, mapsUrl: null, reviews: [] }

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function installApi(page: Page, reviews: { status: number; body: unknown }) {
  await page.route(`**/api/v1/public/barbershops/${SLUG}`, (route) =>
    route.fulfill(
      json({
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
        vocabulary: INITIAL_VOCABULARY,
        location: null,
        offersHomeService: false,
      }),
    ),
  )
  await page.route(`**/api/v1/public/barbershops/${SLUG}/reviews`, (route) =>
    route.fulfill(json(reviews.body, reviews.status)),
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

for (const [name, response] of [
  [
    'la estructura vacía (sin ficha, sin credencial o consulta fallida)',
    { status: 200, body: empty },
  ],
  ['un error del servidor', { status: 500, body: { title: 'Error interno' } }],
] as const) {
  test(`oculta la sección sin mostrar ningún error ante ${name}`, async ({ page }) => {
    await installApi(page, response)
    await page.goto(`/reservar/${SLUG}`)

    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
    await expect(page.getByText('Opiniones en Google')).toHaveCount(0)
    await expect(page.getByRole('alert')).toHaveCount(0)
  })
}

for (const viewport of viewports) {
  test(`muestra la calificación y las reseñas de Google a ${viewport.name} px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await installApi(page, { status: 200, body: withListing })
    await page.goto(`/reservar/${SLUG}`)

    const section = page.getByRole('region', { name: 'Opiniones en Google' })
    await expect(section).toBeVisible()
    await expect(section.getByRole('img', { name: '4,7 de 5 estrellas en Google' })).toBeVisible()
    await expect(section.getByText('128 calificaciones en Google')).toBeVisible()
    // Las reseñas son opcionales de leer: cerradas hasta que la persona las pide.
    await expect(section.getByText('Excelente atención y muy puntuales.')).toHaveCount(0)
    await section.scrollIntoViewIfNeeded()
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `google-resumen-${viewport.name}.png`),
      fullPage: true,
    })

    await section.getByRole('button', { name: 'Ver reseñas (2)' }).click()
    await expect(section.getByText('Excelente atención y muy puntuales.')).toBeVisible()
    await expect(section.getByRole('link', { name: 'Camila R.' })).toHaveAttribute(
      'href',
      withListing.reviews[0]!.authorUrl,
    )
    await expect(section.getByText('Andrés M.')).toBeVisible()
    const attribution = section.getByRole('link', { name: /Ver la ficha en Google Maps/ })
    await expect(attribution).toHaveAttribute('href', withListing.mapsUrl)
    await expect(attribution).toHaveAttribute('rel', /noopener/)
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `google-reseñas-${viewport.name}.png`),
      fullPage: true,
    })
  })
}
