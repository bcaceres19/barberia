import { test, expect, type Route } from '@playwright/test'

/**
 * Recorrido E2E de HU-095 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * exploración pública de fechas y horarios. Mismo criterio que
 * reserva-publica-barbero.spec.ts (HU-092): un único endpoint sin sesión
 * (`GET /public/barbershops/{slug}/services/{serviceId}/barbers/{barberId}/availability`),
 * interceptado con `page.route(...).fulfill(...)` para cubrir cada estado
 * observable de forma determinista, sin depender de un servidor Go real
 * detrás. El endpoint no admite fecha como parámetro (HU-094): devuelve
 * todos los inicios válidos de la ventana pública en una sola respuesta,
 * agrupados por día civil en el cliente.
 */
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'

async function fulfillAvailability(
  route: Route,
  body: {
    slots: Array<{ startsAt: string }>
    durationMinutes: number
    timezone: string
    slotGridMinutes: number
  },
) {
  await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) })
}

const availability = {
  slots: [
    { startsAt: '2026-09-15T19:00:00Z' }, // 2026-09-15 14:00 America/Bogota
    { startsAt: '2026-09-15T19:15:00Z' }, // 2026-09-15 14:15 America/Bogota
    { startsAt: '2026-09-16T14:00:00Z' }, // 2026-09-16 09:00 America/Bogota
  ],
  durationMinutes: 30,
  timezone: 'America/Bogota',
  slotGridMinutes: 15,
}

test.describe('Exploración pública de fechas y horarios (HU-095)', () => {
  test('agrupa las franjas por día civil en la zona de la barbería y permite navegar entre días (CA-095-01)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await fulfillAvailability(route, availability)
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )

    await expect(page.getByText('martes, 15 de septiembre de 2026')).toBeVisible()
    await expect(page.getByRole('radio')).toHaveCount(2)

    await page.getByRole('button', { name: 'Siguiente' }).click()
    await expect(page.getByText('miércoles, 16 de septiembre de 2026')).toBeVisible()
    await expect(page.getByRole('radio')).toHaveCount(1)
    await expect(page.getByRole('button', { name: 'Siguiente' })).toBeDisabled()

    await page.getByRole('button', { name: 'Anterior' }).click()
    await expect(page.getByText('martes, 15 de septiembre de 2026')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Anterior' })).toBeDisabled()
  })

  test('elegir una franja anuncia fecha, hora, zona y duración juntas, sin insinuar una reserva (CA-095-02)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await fulfillAvailability(route, availability)
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )

    const firstSlot = page.getByRole('radio').first()
    await firstSlot.click()
    await expect(firstSlot).toHaveAttribute('aria-checked', 'true')

    const summary = page.locator('.availability__summary')
    await expect(summary).toContainText('martes, 15 de septiembre de 2026')
    await expect(summary).toContainText('America/Bogota')
    await expect(summary).toContainText('30 min')
    await expect(summary).not.toContainText(/reserv|confirm/i)
  })

  test('selecciona la franja enfocada con el teclado (Space/Enter)', async ({ page }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await fulfillAvailability(route, availability)
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )

    const options = page.getByRole('radio')
    await options.first().focus()
    await page.keyboard.press('ArrowRight')
    await page.keyboard.press('Enter')
    await expect(options.nth(1)).toHaveAttribute('aria-checked', 'true')
  })

  test('sin ninguna franja en toda la ventana pública, muestra un vacío distinto de carga o error (CA-095-04)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await fulfillAvailability(route, { ...availability, slots: [] })
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )
    await expect(
      page.getByText('No hay franjas disponibles en la ventana de reserva'),
    ).toBeVisible()
    await expect(page.getByRole('radiogroup')).toHaveCount(0)
  })

  test('un enlace inválido, desconocido o no publicable produce la misma pantalla de "no encontrado" (CA-090-02, reutilizado)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/api/v1/problems/not-found',
            title: 'No encontrado',
            status: 404,
            detail: 'no existe una barbería pública con ese enlace',
            instance: 'e2e-hu095-404',
            code: 'not-found',
            requestId: 'e2e-hu095-404',
          }),
        })
      },
    )

    await page.goto(
      `/reservar/enlace-que-no-existe/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )
    await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
  })

  test('un fallo de red permite reintentar sin recargar la página', async ({ page }) => {
    let attempt = 0
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        attempt += 1
        if (attempt === 1) {
          await route.abort('connectionfailed')
          return
        }
        await fulfillAvailability(route, availability)
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )
    await expect(page.getByText('No pudimos conectar')).toBeVisible()

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page.getByRole('radio')).toHaveCount(2)
    expect(attempt).toBe(2)
  })

  test('el botón "Continuar" de la selección de barbero navega a la exploración de fechas y horarios (HU-095)', async ({
    page,
  }) => {
    const ana = { id: BARBER_ID, fullName: 'Ana Gómez' }
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ items: [ana] }),
        })
      },
    )
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await fulfillAvailability(route, availability)
      },
    )

    await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
    await expect(page.getByText('Ana Gómez')).toBeVisible()
    await page.getByRole('link', { name: 'Continuar' }).click()

    await expect(page).toHaveURL(
      new RegExp(
        `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario$`,
      ),
    )
    await expect(page.getByRole('radio')).toHaveCount(2)
  })
})
