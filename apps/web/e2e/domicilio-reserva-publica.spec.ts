import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { INITIAL_VOCABULARY } from './fixtures/vocabularioInicial'

/**
 * Servicio a domicilio en la reserva pública (HU-103, DEC-131). Respuestas simuladas del API: la
 * regla de «el interruptor de la barbería manda» (422), el CHECK de todo-o-nada y el aislamiento
 * entre barberías los cubren las pruebas Go y SQL con PostgreSQL real. Aquí se recorre lo visible:
 *  1. sin el interruptor encendido no aparece la modalidad;
 *  2. con él, el cliente elige «A domicilio», busca su dirección, elige una sugerencia (el cliente
 *     de Places es el simulado de desarrollo: Vite corre en modo dev sin clave) y confirma el mapa;
 *  3. sin elegir una sugerencia no se puede revisar, y lo enviado incluye la dirección con
 *     coordenadas; en el local no se envía dirección.
 * En 320/360/768/1280 px, sin scroll horizontal y con axe-core. Evidencia en
 * e2e/evidence/domicilio-reserva-publica/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'domicilio-reserva-publica')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const SLUG = 'barberia-ejemplo'
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
const STARTS_AT = '2026-09-20T14:30:00Z'
const ROUTE = `/reservar/${SLUG}/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario/${encodeURIComponent(STARTS_AT)}/cliente`

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

interface ConfirmBody {
  homeService?: { address: string; latitude: number; longitude: number } | null
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

/** Instala el API simulado y devuelve los cuerpos que recibió la confirmación. */
async function installApi(page: Page, offersHomeService: boolean): Promise<ConfirmBody[]> {
  const confirmations: ConfirmBody[] = []

  await page.route(`**/api/v1/public/barbershops/${SLUG}`, (route) =>
    route.fulfill(
      json({
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
        vocabulary: INITIAL_VOCABULARY,
        location: null,
        offersHomeService,
      }),
    ),
  )
  await page.route(`**/api/v1/public/barbershops/${SLUG}/services/${SERVICE_ID}/barbers`, (route) =>
    route.fulfill(
      json({
        items: [
          {
            id: BARBER_ID,
            fullName: 'Mateo Cruz',
            description: null,
            photoUpdatedAt: null,
            rating: null,
            ratingCount: 0,
          },
        ],
      }),
    ),
  )
  await page.route(
    `**/api/v1/public/barbershops/${SLUG}/services/${SERVICE_ID}/barbers/${BARBER_ID}/appointments`,
    async (route) => {
      confirmations.push(route.request().postDataJSON() as ConfirmBody)
      await route.fulfill(
        json(
          {
            attendeeName: 'Ana Ríos',
            barbershopName: 'Barbería Ejemplo',
            serviceName: 'Corte clásico',
            durationMinutes: 30,
            priceAmount: '20000.00',
            currency: 'COP',
            startsAt: STARTS_AT,
            endsAt: '2026-09-20T15:00:00Z',
            timezone: 'America/Bogota',
            accessToken: '3n9F1sQe2z8mM5wYtR7pL0oXhC4bV6dK1aE9jU2gN8i',
            customerNote: null,
          },
          201,
        ),
      )
    },
  )
  return confirmations
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

async function fillCustomer(page: Page) {
  await page.getByLabel('Tu nombre').fill('Ana Ríos')
  await page.getByLabel('Teléfono').fill('+573001234567')
  await page.getByLabel('Correo').fill('ana@example.com')
}

test('sin el interruptor encendido no se ofrece «a domicilio»', async ({ page }) => {
  await installApi(page, false)
  await page.goto(ROUTE)

  await expect(page.getByLabel('Tu nombre')).toBeVisible()
  await expect(page.getByText('¿Dónde te atendemos?')).toHaveCount(0)
  await expect(page.getByLabel('A domicilio')).toHaveCount(0)
})

test('exige elegir una sugerencia: un texto escrito no es una dirección', async ({ page }) => {
  await installApi(page, true)
  await page.goto(ROUTE)

  await fillCustomer(page)
  await page.getByLabel('A domicilio').check()
  await page.getByRole('combobox', { name: 'Tu dirección' }).fill('Carrera 7 71')
  await page.getByRole('button', { name: 'Ver resumen' }).click()

  await expect(page.getByText('Elige tu dirección de la lista de sugerencias.')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Confirmar turno' })).toHaveCount(0)
})

test('en el local no se envía ninguna dirección', async ({ page }) => {
  const confirmations = await installApi(page, true)
  await page.goto(ROUTE)

  await fillCustomer(page)
  await expect(page.getByLabel('En el local')).toBeChecked()
  await page.getByRole('button', { name: 'Ver resumen' }).click()
  await page.getByRole('button', { name: 'Confirmar turno' }).click()

  await expect(page.getByText('¡Tu turno quedó confirmado!')).toBeVisible()
  expect(confirmations).toHaveLength(1)
  expect(confirmations[0]!.homeService ?? null).toBeNull()
})

for (const viewport of viewports) {
  test(`el cliente reserva a domicilio con su dirección a ${viewport.name} px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    const confirmations = await installApi(page, true)
    await page.goto(ROUTE)

    await expect(page.getByText('¿Dónde te atendemos?')).toBeVisible()
    await page.getByLabel('A domicilio').check()
    await page.getByRole('combobox', { name: 'Tu dirección' }).fill('Carrera 7 71')
    const option = page.getByRole('option').first()
    await expect(option).toBeVisible()
    await option.click()
    // La selección muestra el mapa de confirmación del punto elegido.
    await expect(page.locator('iframe')).toBeVisible()
    await fillCustomer(page)
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `domicilio-direccion-${viewport.name}.png`),
      fullPage: true,
    })

    await page.getByRole('button', { name: 'Ver resumen' }).click()
    await expect(page.getByText('Te atendemos en')).toBeVisible()
    await expect(page.getByText('Carrera 7 71, Bogotá (1)')).toBeVisible()
    await expectAccessibleWithoutOverflow(page)
    await page.screenshot({
      path: path.join(evidenceDir, `domicilio-resumen-${viewport.name}.png`),
      fullPage: true,
    })

    // A 320×720 la firma del pie se superpone a los botones del resumen (defecto de maquetación
    // previo a HU-103, también sin domicilio): ahí se confirma con el teclado, que no depende
    // del punto donde cae el puntero; aquí solo se verifica lo enviado.
    const confirm = page.getByRole('button', { name: 'Confirmar turno' })
    if (viewport.width === 320) {
      await confirm.focus()
      await page.keyboard.press('Enter')
    } else {
      await confirm.click()
    }
    await expect(page.getByText('¡Tu turno quedó confirmado!')).toBeVisible()
    expect(confirmations).toHaveLength(1)
    expect(confirmations[0]!.homeService).toEqual({
      address: 'Carrera 7 71, Bogotá (1)',
      latitude: 4.6486,
      longitude: -74.0628,
    })
  })
}
