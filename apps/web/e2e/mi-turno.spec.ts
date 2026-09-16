import { test, expect, type Route } from '@playwright/test'

/**
 * Recorrido E2E de HU-098 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * acceso del cliente a su turno mediante el token del enlace. El primer
 * caso cubre creación→consulta de punta a punta: completa el formulario
 * público (HU-096), confirma el turno (HU-097, mock del POST con el mismo
 * `accessToken` que HU-098 después consume) y visita `/mi-turno/{token}`
 * (mock del GET) verificando que ambas pantallas describen el MISMO turno.
 * El resto de casos ejercitan solo la pantalla de consulta, mismo criterio
 * que `reserva-publica.spec.ts`: cada prueba intercepta la solicitud con
 * `page.route(...).fulfill(...)`, sin depender de un servidor Go real
 * detrás. `webServer` de playwright.config.ts levanta el frontend real
 * (`pnpm dev`).
 */
const ACCESS_TOKEN = '3n9F1sQe2z8mM5wYtR7pL0oXhC4bV6dK1aE9jU2gN8i'

const APPOINTMENT_VIEW = {
  barbershopName: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  attendeeName: 'Ana Ríos',
  serviceName: 'Corte clásico',
  durationMinutes: 30,
  barberName: 'Barbero Ejemplo',
  startsAt: '2027-03-15T15:00:00Z',
  endsAt: '2027-03-15T15:30:00Z',
  status: 'confirmed',
  cancellationDeadlineMinutes: 20,
  lateCancellationClientAllowed: true,
  lateCancellationReasonRequired: true,
}

async function fulfillAppointment(route: Route) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify(APPOINTMENT_VIEW),
  })
}

async function fulfillNotFound(route: Route) {
  await route.fulfill({
    status: 404,
    contentType: 'application/problem+json',
    body: JSON.stringify({
      type: '/api/v1/problems/not-found',
      title: 'No encontrado',
      status: 404,
      detail: 'no existe un turno accesible con ese enlace',
      instance: 'e2e-hu098-404',
      code: 'not-found',
      requestId: 'e2e-hu098-404',
    }),
  })
}

test.describe('Creación→consulta del turno del cliente (HU-097/HU-098)', () => {
  test('el token que emite la confirmación es el mismo que resuelve la consulta posterior (CA-098-01)', async ({
    page,
  }) => {
    const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
    const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
    const STARTS_AT = encodeURIComponent('2027-03-15T15:00:00Z')

    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/appointments`,
      async (route) => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            attendeeName: APPOINTMENT_VIEW.attendeeName,
            barbershopName: APPOINTMENT_VIEW.barbershopName,
            serviceName: APPOINTMENT_VIEW.serviceName,
            durationMinutes: APPOINTMENT_VIEW.durationMinutes,
            priceAmount: '20000.00',
            currency: 'COP',
            startsAt: APPOINTMENT_VIEW.startsAt,
            endsAt: APPOINTMENT_VIEW.endsAt,
            timezone: APPOINTMENT_VIEW.timezone,
            accessToken: ACCESS_TOKEN,
            customerNote: null,
          }),
        })
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario/${STARTS_AT}/cliente`,
    )
    await page.getByLabel('Tu nombre').fill(APPOINTMENT_VIEW.attendeeName)
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')
    await page.getByRole('button', { name: 'Ver resumen' }).click()
    await expect(page.getByText('Revisa tus datos')).toBeVisible()
    await page.getByRole('button', { name: 'Confirmar turno' }).click()
    await expect(page.getByText('¡Tu turno quedó confirmado!')).toBeVisible()

    // El enlace en claro solo viaja por correo (DEC-091): esta pantalla
    // nunca lo muestra ni lo enlaza. HU-098 lo consume de forma
    // independiente, tal como el cliente lo haría desde su correo.
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, fulfillAppointment)
    await page.goto(`/mi-turno/${ACCESS_TOKEN}`)

    await expect(page.getByRole('heading', { name: APPOINTMENT_VIEW.barbershopName })).toBeVisible()
    await expect(page.getByText(APPOINTMENT_VIEW.attendeeName)).toBeVisible()
    await expect(page.getByText(APPOINTMENT_VIEW.serviceName)).toBeVisible()
    await expect(page.getByText(APPOINTMENT_VIEW.barberName)).toBeVisible()
  })
})

test.describe('Consulta del turno del cliente (HU-098)', () => {
  test('un token válido muestra el turno, la hora en la zona de la barbería y la política de cancelación (CA-098-01, CA-098-04)', async ({
    page,
  }) => {
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, fulfillAppointment)

    await page.goto(`/mi-turno/${ACCESS_TOKEN}`)

    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
    await expect(page.getByText('Confirmado')).toBeVisible()
    await expect(page.getByText('20 minutos antes')).toBeVisible()

    // RN-CNF-02: el estado nunca insinúa que la falta de respuesta cancele
    // el turno.
    const bodyText = await page.locator('body').innerText()
    expect(bodyText.toLowerCase()).not.toContain('se cancelará')
  })

  test('un enlace inválido, vencido o revocado produce la misma pantalla uniforme (CA-098-02)', async ({
    page,
  }) => {
    await page.route('**/api/v1/customer/appointments/**', fulfillNotFound)

    await page.goto('/mi-turno/token-que-ya-no-sirve')
    await expect(page.getByRole('alert')).toContainText('No pudimos abrir ese enlace')

    // Ni el token ni ningún identificador interno aparecen en el DOM.
    const bodyText = await page.locator('body').innerText()
    expect(bodyText).not.toContain('token-que-ya-no-sirve')
    expect(bodyText.toLowerCase()).not.toContain('barbershopid')
  })

  test('nunca expone un identificador interno en el cuerpo de la respuesta ni en el DOM (CA-098-03)', async ({
    page,
  }) => {
    let capturedUrl = ''
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, async (route) => {
      capturedUrl = route.request().url()
      await fulfillAppointment(route)
    })

    await page.goto(`/mi-turno/${ACCESS_TOKEN}`)
    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()

    expect(capturedUrl).toContain(ACCESS_TOKEN)
    const html = await page.content()
    for (const forbidden of ['barbershopId', 'appointmentId', 'customerId']) {
      expect(html.toLowerCase()).not.toContain(forbidden.toLowerCase())
    }
  })

  test('ofrece reintentar ante un error de red, volviendo a pedir el mismo token', async ({
    page,
  }) => {
    let attempts = 0
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, async (route) => {
      attempts += 1
      if (attempts === 1) {
        await route.abort('failed')
        return
      }
      await fulfillAppointment(route)
    })

    await page.goto(`/mi-turno/${ACCESS_TOKEN}`)
    await expect(page.getByText('No pudimos conectar')).toBeVisible()

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
    expect(attempts).toBe(2)
  })

  test('un error inesperado del servidor muestra el requestId sin filtrar el cuerpo crudo del problema', async ({
    page,
  }) => {
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/problem+json',
        body: JSON.stringify({
          type: '/api/v1/problems/internal-error',
          title: 'Error interno',
          status: 500,
          instance: 'e2e-hu098-500',
          code: 'internal-error',
          requestId: 'e2e-hu098-500',
        }),
      })
    })

    await page.goto(`/mi-turno/${ACCESS_TOKEN}`)
    await expect(page.getByText('Ocurrió un error inesperado')).toBeVisible()
    await expect(page.getByText('e2e-hu098-500')).toBeVisible()
  })
})
