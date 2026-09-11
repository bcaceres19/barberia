import { test, expect, type Route } from '@playwright/test'

/**
 * Recorrido E2E de HU-091 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * catálogo público de servicios. Mismo criterio que
 * reserva-publica.spec.ts (HU-090): un único endpoint sin sesión (`GET
 * /public/barbershops/{slug}/services`), interceptado con
 * `page.route(...).fulfill(...)` para cubrir cada estado observable de
 * forma determinista, sin depender de `database/testdata/
 * hu091_catalogo_publico.sql` ni de un servidor Go real detrás.
 */
async function fulfillServices(
  route: Route,
  items: Array<{
    id: string
    name: string
    description: string | null
    durationMinutes: number
    price: string
    currency: string
  }>,
) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ items, nextCursor: null }),
  })
}

const corteClasico = {
  id: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
  name: 'Corte clásico',
  description: 'Corte con máquina y tijera, incluye lavado.',
  durationMinutes: 30,
  price: '45000.00',
  currency: 'COP',
}
const barba = {
  id: '1a2b3c4d-5e6f-4708-9a0b-1c2d3e4f5061',
  name: 'Barba',
  description: null,
  durationMinutes: 20,
  price: '20000.00',
  currency: 'COP',
}

test.describe('Catálogo público de servicios (HU-091)', () => {
  test('un enlace válido lista solo los servicios activos y asignados (CA-091-01)', async ({
    page,
  }) => {
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo/services**', async (route) => {
      await fulfillServices(route, [corteClasico, barba])
    })

    await page.goto('/reservar/barberia-ejemplo/servicios')

    await expect(page.getByText('Corte clásico')).toBeVisible()
    await expect(page.getByText('30 min')).toBeVisible()
    await expect(page.getByText('45000.00 COP')).toBeVisible()
    await expect(page.getByText('Barba')).toBeVisible()
    await expect(page.getByText('20 min')).toBeVisible()
    await expect(page.getByText('20000.00 COP')).toBeVisible()
  })

  test('una barbería sin servicios públicos todavía muestra el estado vacío (CA-091-04)', async ({
    page,
  }) => {
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo/services**', async (route) => {
      await fulfillServices(route, [])
    })

    await page.goto('/reservar/barberia-ejemplo/servicios')
    await expect(
      page.getByText('Esta barbería todavía no tiene servicios disponibles para reservar.'),
    ).toBeVisible()
  })

  test('un enlace inválido, desconocido o no publicable produce la misma pantalla de "no encontrado" (CA-090-02, reutilizado)', async ({
    page,
  }) => {
    await page.route('**/api/v1/public/barbershops/**/services**', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/api/v1/problems/not-found',
          title: 'No encontrado',
          status: 404,
          detail: 'no existe una barbería pública con ese enlace',
          instance: 'e2e-hu091-404',
          code: 'not-found',
          requestId: 'e2e-hu091-404',
        }),
      })
    })

    await page.goto('/reservar/enlace-que-no-existe/servicios')
    await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
  })

  test('un fallo de red permite reintentar sin recargar la página', async ({ page }) => {
    let attempt = 0
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo/services**', async (route) => {
      attempt += 1
      if (attempt === 1) {
        await route.abort('connectionfailed')
        return
      }
      await fulfillServices(route, [corteClasico])
    })

    await page.goto('/reservar/barberia-ejemplo/servicios')
    await expect(page.getByText('No pudimos conectar')).toBeVisible()

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page.getByText('Corte clásico')).toBeVisible()
    expect(attempt).toBe(2)
  })

  test('selecciona un servicio con clic y con teclado (CA-091-04)', async ({ page }) => {
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo/services**', async (route) => {
      await fulfillServices(route, [corteClasico, barba])
    })

    await page.goto('/reservar/barberia-ejemplo/servicios')

    const corteOption = page.getByRole('radio', { name: /Corte clásico/ })
    const barbaOption = page.getByRole('radio', { name: /Barba/ })

    await corteOption.click()
    await expect(corteOption).toHaveAttribute('aria-checked', 'true')
    await expect(barbaOption).toHaveAttribute('aria-checked', 'false')

    await corteOption.focus()
    await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Enter')
    await expect(barbaOption).toHaveAttribute('aria-checked', 'true')
    await expect(corteOption).toHaveAttribute('aria-checked', 'false')
  })

  test('el botón "Reservar un turno" de la entrada pública navega al catálogo de servicios', async ({
    page,
  }) => {
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          name: 'Barbería Ejemplo',
          timezone: 'America/Bogota',
          contactEmail: null,
          contactPhone: null,
        }),
      })
    })
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo/services**', async (route) => {
      await fulfillServices(route, [corteClasico])
    })

    await page.goto('/reservar/barberia-ejemplo')
    await page.getByRole('link', { name: 'Reservar un turno' }).click()

    await expect(page).toHaveURL(/\/reservar\/barberia-ejemplo\/servicios$/)
    await expect(page.getByText('Corte clásico')).toBeVisible()
  })
})
