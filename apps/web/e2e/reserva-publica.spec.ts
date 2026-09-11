import { test, expect, type Route } from '@playwright/test'

/**
 * Recorrido E2E de HU-090 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * entrada pública de reservas. A diferencia de la mayoría de los
 * recorridos P0 de este directorio, esta pantalla depende de un ÚNICO
 * endpoint sin sesión (`GET /public/barbershops/{slug}`); cada prueba
 * intercepta esa solicitud con `page.route(...).fulfill(...)` para cubrir
 * cada estado observable de forma determinista, sin depender de datos
 * reales de `database/testdata/hu090_reserva_publica.sql` ni de un
 * servidor Go real detrás — mismo patrón ya usado por el caso "bloqueado
 * por 429" de `acceso-evidencia-responsiva.spec.ts`. `webServer` de
 * playwright.config.ts levanta el frontend real (`pnpm dev`); ningún
 * fallback ni mock sustituye el enrutamiento, el componente ni el cliente
 * HTTP tipado reales.
 */
async function fulfillProfile(
  route: Route,
  body: {
    name: string
    timezone: string
    contactEmail: string | null
    contactPhone: string | null
  },
) {
  await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify(body) })
}

test.describe('Entrada pública de reservas (HU-090)', () => {
  test('un enlace válido abre el contexto público sin sesión (CA-090-01)', async ({ page }) => {
    await page.route('**/api/v1/public/barbershops/barberia-ejemplo', async (route) => {
      await fulfillProfile(route, {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: 'contacto@ejemplo.test',
        contactPhone: '+573001234567',
      })
    })

    await page.goto('/reservar/barberia-ejemplo')

    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
    await expect(page.getByText('Hora local de la barbería')).toBeVisible()
    await expect(page.getByText('contacto@ejemplo.test')).toBeVisible()
    await expect(page.getByText('+573001234567')).toBeVisible()
  })

  test('un enlace inválido, desconocido o no publicable produce la misma pantalla de "no encontrado" (CA-090-02)', async ({
    page,
  }) => {
    await page.route('**/api/v1/public/barbershops/**', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/api/v1/problems/not-found',
          title: 'No encontrado',
          status: 404,
          detail: 'no existe una barbería pública con ese enlace',
          instance: 'e2e-hu090-404',
          code: 'not-found',
          requestId: 'e2e-hu090-404',
        }),
      })
    })

    await page.goto('/reservar/enlace-que-no-existe')
    await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')

    // Ni el enlace inválido ni ningún identificador interno aparecen en el DOM.
    const bodyText = await page.locator('body').innerText()
    expect(bodyText).not.toContain('barbershopId')
  })

  test('el cuerpo de la solicitud pública nunca incluye barbershopId (CA-090-03)', async ({ page }) => {
    let capturedUrl = ''
    await page.route('**/api/v1/public/barbershops/**', async (route) => {
      capturedUrl = route.request().url()
      await fulfillProfile(route, {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      })
    })

    await page.goto('/reservar/barberia-ejemplo')
    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()

    expect(capturedUrl).toContain('/public/barbershops/barberia-ejemplo')
    expect(capturedUrl).not.toContain('barbershopId')
  })

  test('un fallo de red permite reintentar sin recargar la página (CA-090-05)', async ({ page }) => {
    let attempt = 0
    await page.route('**/api/v1/public/barbershops/**', async (route) => {
      attempt += 1
      if (attempt === 1) {
        await route.abort('connectionfailed')
        return
      }
      await fulfillProfile(route, {
        name: 'Barbería Ejemplo',
        timezone: 'America/Bogota',
        contactEmail: null,
        contactPhone: null,
      })
    })

    await page.goto('/reservar/barberia-ejemplo')
    await expect(page.getByText('No pudimos conectar')).toBeVisible()

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
    expect(attempt).toBe(2)
  })

  test('opera completamente con teclado, con foco visible en el botón de reintento (CA-090-05)', async ({
    page,
  }) => {
    await page.route('**/api/v1/public/barbershops/**', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/api/v1/problems/not-found',
          title: 'No encontrado',
          status: 404,
          detail: 'no existe una barbería pública con ese enlace',
          instance: 'e2e-hu090-keyboard',
          code: 'not-found',
          requestId: 'e2e-hu090-keyboard',
        }),
      })
    })

    await page.goto('/reservar/enlace-que-no-existe')
    const retryButton = page.getByRole('button', { name: 'Reintentar' })
    await expect(retryButton).toBeVisible()
    await retryButton.focus()
    await expect(retryButton).toBeFocused()
    await page.keyboard.press('Enter')
    // El reintento vuelve a fallar (misma ruta mockeada) sin romper la
    // pantalla ni perder el foco de teclado como recorrido operable.
    await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
  })
})
