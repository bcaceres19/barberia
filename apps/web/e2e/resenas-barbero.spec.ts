import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { INITIAL_VOCABULARY } from './fixtures/vocabularioInicial'

/**
 * Reseñas internas por barbero (HU-100, DEC-128). Un único estado en memoria alimenta los tres
 * recorridos con respuestas simuladas del API (no dependen de PostgreSQL; la regla de «solo turno
 * completado», una reseña por turno, el promedio y RLS con dos tenants los cubren las pruebas Go):
 *  1. el cliente califica su turno completado en /mi-turno/{token};
 *  2. el visitante ve el promedio y las reseñas al elegir barbero en la reserva pública;
 *  3. quien administra oculta y vuelve a mostrar una reseña desde el detalle del barbero.
 * En 320/360/768/1280 px, sin scroll horizontal y con axe-core.
 * Capturas en e2e/evidence/resenas-barbero/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'resenas-barbero')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const SLUG = 'barberia-ejemplo'
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
const OTHER_BARBER_ID = 'a2222222-2222-2222-2222-222222222222'
const TOKEN = '3n9F1sQe2z8mM5wYtR7pL0oXhC4bV6dK1aE9jU2gN8i'

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

interface StoredReview {
  id: string
  rating: number
  comment: string | null
  hiddenAt: string | null
  createdAt: string
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

function summary(reviews: StoredReview[]) {
  const visible = reviews.filter((r) => r.hiddenAt === null)
  const rating = visible.length
    ? Math.round((visible.reduce((sum, r) => sum + r.rating, 0) / visible.length) * 10) / 10
    : null
  return { rating, ratingCount: visible.length }
}

/** Instala las rutas simuladas con un único estado de reseñas compartido. */
async function installApi(page: Page, initial: StoredReview[] = []) {
  const reviews: StoredReview[] = initial.map((r) => ({ ...r }))
  let myReview: StoredReview | null = null

  await page.route(`**/api/v1/public/barbershops/${SLUG}`, (route) =>
    route.fulfill(
      json({
        name: 'Barbería Ejemplo',
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
        location: null,
      }),
    ),
  )

  await page.route(`**/api/v1/public/barbershops/${SLUG}/services/${SERVICE_ID}/barbers`, (route) =>
    route.fulfill(
      json({
        items: [
          {
            id: BARBER_ID,
            fullName: 'Juan Pérez',
            description: 'Fade y barba.',
            photoUpdatedAt: null,
            ...summary(reviews),
          },
          {
            id: OTHER_BARBER_ID,
            fullName: 'Ana Gómez',
            description: null,
            photoUpdatedAt: null,
            rating: null,
            ratingCount: 0,
          },
        ],
      }),
    ),
  )

  await page.route(`**/api/v1/public/barbershops/${SLUG}/barbers/${BARBER_ID}/reviews`, (route) =>
    route.fulfill(
      json({
        ...summary(reviews),
        items: reviews
          .filter((r) => r.hiddenAt === null)
          .map(({ rating, comment, createdAt }) => ({ rating, comment, createdAt })),
      }),
    ),
  )

  await page.route(`**/api/v1/customer/appointments/${TOKEN}/review`, async (route) => {
    const body = route.request().postDataJSON() as { rating: number; comment: string | null }
    myReview = {
      id: 'b0000000-0000-0000-0000-000000000099',
      rating: body.rating,
      comment: body.comment,
      hiddenAt: null,
      createdAt: '2027-03-15T17:00:00Z',
    }
    reviews.unshift(myReview)
    await route.fulfill(
      json(
        { rating: myReview.rating, comment: myReview.comment, createdAt: myReview.createdAt },
        201,
      ),
    )
  })

  await page.route(`**/api/v1/customer/appointments/${TOKEN}`, (route) =>
    route.fulfill(
      json({
        barbershopName: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        attendeeName: 'Ana Ríos',
        serviceName: 'Corte clásico',
        durationMinutes: 30,
        barberName: 'Juan Pérez',
        startsAt: '2027-03-15T15:00:00Z',
        endsAt: '2027-03-15T15:30:00Z',
        status: 'completed',
        cancellationDeadlineMinutes: 20,
        lateCancellationClientAllowed: true,
        lateCancellationReasonRequired: true,
        versionToken: 'version-token-1',
        barbershopContactEmail: null,
        barbershopContactPhone: null,
        vocabulary: INITIAL_VOCABULARY,
        canReview: myReview === null,
        review: myReview
          ? { rating: myReview.rating, comment: myReview.comment, createdAt: myReview.createdAt }
          : null,
      }),
    ),
  )

  // Panel de administración.
  await page.route('**/api/v1/private/**', async (route) => {
    const url = new URL(route.request().url())
    const method = route.request().method()
    if (url.pathname.endsWith('/auth/session')) {
      return route.fulfill(
        json({ barbershop: { id: 'shop-nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
    }
    if (url.pathname.endsWith('/private/barbers') && method === 'GET') {
      const items = [
        {
          id: BARBER_ID,
          fullName: 'Juan Pérez',
          createdAt: '2026-05-01T10:00:00Z',
          updatedAt: '2026-05-01T10:00:00Z',
          photoUpdatedAt: null,
          description: null,
          ...summary(reviews),
        },
      ]
      // La pantalla pide la página numerada (?page=&pageSize=), no el cursor.
      return route.fulfill(
        json({ items, page: 1, pageSize: 20, total: items.length, totalPages: 1 }),
      )
    }
    const reviewMatch = url.pathname.match(/\/private\/barbers\/([^/]+)\/reviews(?:\/([^/]+))?$/)
    if (reviewMatch && method === 'GET') {
      return route.fulfill(json({ ...summary(reviews), items: reviews }))
    }
    if (reviewMatch?.[2] && method === 'PATCH') {
      const target = reviews.find((r) => r.id === reviewMatch[2])
      if (!target) return route.fulfill(json({ title: 'No encontrado' }, 404))
      const { hidden } = route.request().postDataJSON() as { hidden: boolean }
      target.hiddenAt = hidden ? (target.hiddenAt ?? '2027-03-16T10:00:00Z') : null
      return route.fulfill(json(target))
    }
    return route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })

  return reviews
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

const seed: StoredReview[] = [
  {
    id: 'b0000000-0000-0000-0000-000000000001',
    rating: 5,
    comment: 'Muy buen corte y puntual.',
    hiddenAt: null,
    createdAt: '2027-03-10T12:00:00Z',
  },
  {
    id: 'b0000000-0000-0000-0000-000000000002',
    rating: 1,
    comment: 'Comentario inapropiado.',
    hiddenAt: '2027-03-12T09:00:00Z',
    createdAt: '2027-03-11T12:00:00Z',
  },
]

for (const viewport of viewports) {
  test(`el cliente califica su turno completado a ${viewport.name} px`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await installApi(page)
    await page.goto(`/mi-turno/${TOKEN}`)

    await expect(page.getByRole('heading', { name: 'Califica a Juan Pérez' })).toBeVisible()
    // Sin elegir estrellas no se envía nada.
    await page.getByRole('button', { name: 'Enviar reseña' }).click()
    await expect(page.getByText('Elige de 1 a 5 estrellas.')).toBeVisible()
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `calificar-${viewport.name}.png`),
      fullPage: true,
    })

    await page.getByLabel('4 de 5 estrellas').check({ force: true })
    await page.getByLabel(/Comentario/).fill('Excelente atención.')
    await page.getByRole('button', { name: 'Enviar reseña' }).click()

    await expect(page.getByRole('heading', { name: 'Gracias por tu reseña' })).toBeVisible()
    await expect(page.getByText('Excelente atención.')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Enviar reseña' })).toHaveCount(0)
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `gracias-${viewport.name}.png`),
      fullPage: true,
    })
  })

  test(`el visitante ve el promedio y las reseñas al elegir barbero a ${viewport.name} px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await installApi(page, seed)
    await page.goto(`/reservar/${SLUG}/servicios/${SERVICE_ID}/barbero`)

    // Solo el barbero con reseñas visibles muestra promedio; la oculta no cuenta.
    await expect(page.getByRole('img', { name: '5,0 de 5 · 1 reseña' })).toBeVisible()
    await expect(page.getByRole('img', { name: /de 5/ })).toHaveCount(1)

    await page.getByRole('radio', { name: 'Juan Pérez' }).click()
    await page.getByRole('button', { name: 'Ver reseñas (1)' }).click()
    await expect(page.getByText('Muy buen corte y puntual.')).toBeVisible()
    await expect(page.getByText('Comentario inapropiado.')).toHaveCount(0)

    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `reserva-publica-${viewport.name}.png`),
      fullPage: true,
    })
  })

  test(`quien administra oculta y vuelve a mostrar una reseña a ${viewport.name} px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await installApi(page, seed)
    await page.goto('/panel/barberos')

    await page
      .getByRole('button', { name: /Juan Pérez/ })
      .first()
      .click()
    const dialog = page.getByRole('dialog', { name: 'Juan Pérez' })
    await expect(dialog.getByRole('img', { name: '5,0 de 5 · 1 reseña' })).toBeVisible()

    await dialog.getByRole('button', { name: 'Ver reseñas' }).click()
    await expect(dialog.getByText('Comentario inapropiado.')).toBeVisible()
    await expect(dialog.getByText(/Oculta desde/)).toBeVisible()
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `moderacion-${viewport.name}.png`),
      animations: 'disabled',
    })

    // Ocultar la visible: el promedio pasa a «Sin reseñas todavía».
    await dialog.getByRole('button', { name: /Ocultar la reseña de 5 estrellas/ }).click()
    await expect(dialog.getByRole('img', { name: 'Sin reseñas todavía' })).toBeVisible()

    // Y mostrar la oculta: vuelve al promedio y al conteo.
    await dialog.getByRole('button', { name: /Mostrar la reseña de 1 estrellas/ }).click()
    await expect(dialog.getByRole('img', { name: '1,0 de 5 · 1 reseña' }).first()).toBeVisible()
  })
}
