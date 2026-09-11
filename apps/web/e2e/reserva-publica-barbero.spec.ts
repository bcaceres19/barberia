import { test, expect, type Route } from '@playwright/test'

/**
 * Recorrido E2E de HU-092 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * selección pública de barbero. Mismo criterio que
 * reserva-publica-catalogo.spec.ts (HU-091): un único endpoint sin sesión
 * (`GET /public/barbershops/{slug}/services/{serviceId}/barbers`),
 * interceptado con `page.route(...).fulfill(...)` para cubrir cada estado
 * observable de forma determinista, sin depender de `database/testdata/
 * hu092_seleccion_barbero.sql` ni de un servidor Go real detrás.
 */
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'

async function fulfillBarbers(route: Route, items: Array<{ id: string; fullName: string }>) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ items }),
  })
}

const ana = { id: 'a1111111-1111-1111-1111-111111111111', fullName: 'Ana Gómez' }
const luis = { id: 'b2222222-2222-2222-2222-222222222222', fullName: 'Luis Rojas' }

test.describe('Selección pública de barbero (HU-092)', () => {
  test('con un único barbero elegible, se preselecciona sin exigir un paso adicional (CA-092-01)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        await fulfillBarbers(route, [ana])
      },
    )

    await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)

    await expect(page.getByText('Ana Gómez')).toBeVisible()
    await expect(page.getByRole('radiogroup')).toHaveCount(0)
  })

  test('con varios barberos elegibles, exige una elección explícita (CA-092-01)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        await fulfillBarbers(route, [ana, luis])
      },
    )

    await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)

    const anaOption = page.getByRole('radio', { name: /Ana Gómez/ })
    const luisOption = page.getByRole('radio', { name: /Luis Rojas/ })
    await expect(anaOption).toHaveAttribute('aria-checked', 'false')
    await expect(luisOption).toHaveAttribute('aria-checked', 'false')

    await anaOption.click()
    await expect(anaOption).toHaveAttribute('aria-checked', 'true')
    await expect(luisOption).toHaveAttribute('aria-checked', 'false')

    await anaOption.focus()
    await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Enter')
    await expect(luisOption).toHaveAttribute('aria-checked', 'true')
    await expect(anaOption).toHaveAttribute('aria-checked', 'false')
  })

  test('sin barberos elegibles, muestra un estado vacío distinto de carga o error (CA-092-03)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        await fulfillBarbers(route, [])
      },
    )

    await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
    await expect(
      page.getByText('Este servicio no tiene barberos disponibles en este momento.'),
    ).toBeVisible()
  })

  test('un enlace inválido, desconocido o no publicable produce la misma pantalla de "no encontrado" (CA-090-02, reutilizado)', async ({
    page,
  }) => {
    await page.route(
      `**/api/v1/public/barbershops/**/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        await route.fulfill({
          status: 404,
          contentType: 'application/problem+json',
          body: JSON.stringify({
            type: '/api/v1/problems/not-found',
            title: 'No encontrado',
            status: 404,
            detail: 'no existe una barbería pública con ese enlace',
            instance: 'e2e-hu092-404',
            code: 'not-found',
            requestId: 'e2e-hu092-404',
          }),
        })
      },
    )

    await page.goto(`/reservar/enlace-que-no-existe/servicios/${SERVICE_ID}/barbero`)
    await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
  })

  test('un fallo de red permite reintentar sin recargar la página', async ({ page }) => {
    let attempt = 0
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        attempt += 1
        if (attempt === 1) {
          await route.abort('connectionfailed')
          return
        }
        await fulfillBarbers(route, [ana])
      },
    )

    await page.goto(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero`)
    await expect(page.getByText('No pudimos conectar')).toBeVisible()

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page.getByText('Ana Gómez')).toBeVisible()
    expect(attempt).toBe(2)
  })

  test('el botón "Continuar" del catálogo navega a la selección de barbero del servicio elegido (HU-092)', async ({
    page,
  }) => {
    const corteClasico = {
      id: SERVICE_ID,
      name: 'Corte clásico',
      description: null,
      durationMinutes: 30,
      price: '45000.00',
      currency: 'COP',
    }
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo/services**', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ items: [corteClasico], nextCursor: null }),
      })
    })
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers**`,
      async (route) => {
        await fulfillBarbers(route, [ana])
      },
    )

    await page.goto('/reservar/barberia-ejemplo/servicios')
    await page.getByRole('radio', { name: /Corte clásico/ }).click()
    await page.getByRole('link', { name: 'Continuar' }).click()

    await expect(page).toHaveURL(
      new RegExp(`/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero$`),
    )
    await expect(page.getByText('Ana Gómez')).toBeVisible()
  })
})
