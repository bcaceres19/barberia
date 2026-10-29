import { test, expect, type Page, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { INITIAL_VOCABULARY } from './fixtures/vocabularioInicial'

/**
 * Recorrido E2E y evidencia responsive/accesible de HU-099 (T5, cancelación
 * pública; docs/03-desarrollo/estrategia-pruebas.md §5.3 y §5.4): el cliente
 * cancela su turno desde `/mi-turno/:token`. Mismo criterio que
 * `mi-turno.spec.ts`: cada solicitud se intercepta con `page.route(...)`, sin
 * servidor Go real. La política real (antes/exacto/después, atomicidad,
 * liberación y carreras) se prueba con PostgreSQL en
 * `apps/api/internal/modules/booking/postgres/cancel_customer_repository_test.go`;
 * aquí se comprueba que la interfaz envía lo correcto, explica cada desenlace y
 * que, tras cancelar, la franja vuelve a aparecer al reconsultar la
 * disponibilidad (CA-099-04). Las capturas quedan en `e2e/evidence/mi-turno-cancelacion/`.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'mi-turno-cancelacion',
)
const axeScriptPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'node_modules',
  'axe-core',
  'axe.min.js',
)

const ACCESS_TOKEN = '3n9F1sQe2z8mM5wYtR7pL0oXhC4bV6dK1aE9jU2gN8i'
const ROUTE = `/mi-turno/${ACCESS_TOKEN}`
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
const STARTS_AT = '2027-03-15T15:00:00Z'
// 14:40, el límite de un plazo de 20 minutos; el reloj del dispositivo solo da la pista local.
const BEFORE_DEADLINE = new Date('2027-03-15T13:00:00Z')
const AFTER_DEADLINE = new Date('2027-03-15T14:50:00Z')

const APPOINTMENT_VIEW = {
  barbershopName: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  attendeeName: 'Ana Ríos',
  serviceName: 'Corte clásico',
  durationMinutes: 30,
  barberName: 'Barbero Ejemplo',
  startsAt: STARTS_AT,
  endsAt: '2027-03-15T15:30:00Z',
  status: 'confirmed',
  cancellationDeadlineMinutes: 20,
  lateCancellationClientAllowed: true,
  lateCancellationReasonRequired: true,
  versionToken: 'version-token-1',
  barbershopContactEmail: 'contacto@barberia-ejemplo.test',
  barbershopContactPhone: '+573001234567',
  vocabulary: INITIAL_VOCABULARY,
}

const json = (route: Route, status: number, body: unknown, contentType = 'application/json') =>
  route.fulfill({ status, contentType, body: JSON.stringify(body) })

const problem = (status: number, code: string, detail: string) => ({
  type: `/api/v1/problems/${code}`,
  title: 'x',
  status,
  detail,
  instance: 'e2e-hu099',
  code,
  requestId: 'e2e-hu099',
})

async function mockAppointment(page: Page, view: Record<string, unknown> = APPOINTMENT_VIEW) {
  await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, (route) =>
    json(route, 200, view),
  )
}

async function assertNoAxeViolations(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  const results = (await page.evaluate(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (window as any).axe.run(document, { rules: { 'color-contrast': { enabled: false } } })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  })) as { violations: any[] }
  expect(results.violations, JSON.stringify(results.violations, null, 2)).toEqual([])
}

test.describe('Cancelación pública del turno (HU-099)', () => {
  test('cancela dentro de plazo, el enlace queda revocado y la franja vuelve a ofrecerse (CA-099-01, CA-099-04)', async ({
    page,
  }) => {
    await page.clock.setFixedTime(BEFORE_DEADLINE)
    let cancelled = false
    let captured: { headers: Record<string, string>; body: unknown } | undefined

    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, (route) =>
      cancelled
        ? json(
            route,
            404,
            problem(404, 'not-found', 'no existe un turno accesible con ese enlace'),
            'application/problem+json',
          )
        : json(route, 200, APPOINTMENT_VIEW),
    )
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}/cancel`, (route) => {
      captured = { headers: route.request().headers(), body: route.request().postDataJSON() }
      cancelled = true
      return json(route, 200, { status: 'cancelled_by_customer' })
    })
    const availability = (slots: string[]) =>
      page.route(
        `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
        (route) =>
          json(route, 200, {
            slots: slots.map((startsAt) => ({ startsAt })),
            durationMinutes: 30,
            timezone: 'America/Bogota',
            slotGridMinutes: 15,
          }),
      )

    await page.goto(ROUTE)
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    // La consecuencia se dice ANTES de confirmar, y el foco entra al campo de motivo.
    await expect(page.getByText('No se puede deshacer')).toBeVisible()
    await expect(page.getByLabel(/Motivo/)).toBeFocused()
    await page.getByLabel(/Motivo/).fill('Me surgió un imprevisto')
    await page.getByRole('button', { name: 'Sí, cancelar mi turno' }).click()

    await expect(page.getByText('Cancelaste tu turno')).toBeVisible()
    await expect(page.getByText('Cancelado por el cliente')).toBeVisible()
    await expect(page.getByRole('button', { name: /Cancelar turno|Sí, cancelar/ })).toHaveCount(0)
    expect(captured?.headers['if-match']).toBe('version-token-1')
    expect(captured?.headers['idempotency-key']).toMatch(/^[0-9a-f-]{36}$/)
    expect(captured?.body).toEqual({ reason: 'Me surgió un imprevisto' })

    // DEC-089: el enlace ya no sirve.
    await page.reload()
    await expect(page.getByRole('alert')).toContainText('No pudimos abrir ese enlace')

    // CA-099-04: al reconsultar la disponibilidad pública la franja liberada vuelve.
    await availability([STARTS_AT])
    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )
    await expect(page.getByRole('radio')).toHaveCount(1)
  })

  test('fuera de plazo con motivo obligatorio no envía en blanco y luego sí con motivo (CA-099-02, CA-099-03)', async ({
    page,
  }) => {
    await page.clock.setFixedTime(AFTER_DEADLINE)
    await mockAppointment(page)
    const bodies: unknown[] = []
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}/cancel`, (route) => {
      bodies.push(route.request().postDataJSON())
      return json(route, 200, { status: 'cancelled_by_customer' })
    })

    await page.goto(ROUTE)
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    await expect(page.getByText('(obligatorio)')).toBeVisible()

    await page.getByRole('button', { name: 'Sí, cancelar mi turno' }).click()
    await expect(page.getByRole('alert').filter({ hasText: 'motivo' })).toBeVisible()
    expect(bodies).toHaveLength(0)
    await expect(page.getByLabel(/Motivo/)).toBeFocused()

    await page.getByLabel(/Motivo/).fill('No llego a tiempo')
    await page.getByRole('button', { name: 'Sí, cancelar mi turno' }).click()
    await expect(page.getByText('Cancelaste tu turno')).toBeVisible()
    expect(bodies).toEqual([{ reason: 'No llego a tiempo' }])
  })

  test('si la política no permite cancelar tarde no hay botón y se explica a quién contactar (CA-099-06)', async ({
    page,
  }) => {
    await page.clock.setFixedTime(AFTER_DEADLINE)
    await mockAppointment(page, {
      ...APPOINTMENT_VIEW,
      lateCancellationClientAllowed: false,
      lateCancellationReasonRequired: false,
    })

    await page.goto(ROUTE)
    await expect(page.getByText('El plazo para cancelar por tu cuenta ya venció')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancelar turno' })).toHaveCount(0)
    await expect(page.getByRole('link', { name: '+573001234567' })).toHaveAttribute(
      'href',
      'tel:+573001234567',
    )
    await expect(page.getByRole('link', { name: 'contacto@barberia-ejemplo.test' })).toBeVisible()
  })

  test('el servidor manda: un 422 de plazo vencido reemplaza el formulario y deja el turno confirmado', async ({
    page,
  }) => {
    await page.clock.setFixedTime(BEFORE_DEADLINE)
    await mockAppointment(page)
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}/cancel`, (route) =>
      json(
        route,
        422,
        problem(422, 'cancellation-window-closed', 'el plazo para cancelar ya venció'),
        'application/problem+json',
      ),
    )

    await page.goto(ROUTE)
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    await page.getByRole('button', { name: 'Sí, cancelar mi turno' }).click()

    await expect(page.getByText('El plazo para cancelar por tu cuenta ya venció')).toBeVisible()
    await expect(page.getByText('Confirmado')).toBeVisible()
    await expect(page.getByRole('link', { name: '+573001234567' })).toBeVisible()
  })

  test('un 409 de versión recarga el turno y avisa que cambió (CA-099-05)', async ({ page }) => {
    await page.clock.setFixedTime(BEFORE_DEADLINE)
    let reads = 0
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}`, (route) => {
      reads += 1
      return json(
        route,
        200,
        reads === 1 ? APPOINTMENT_VIEW : { ...APPOINTMENT_VIEW, status: 'cancelled_by_barber' },
      )
    })
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}/cancel`, (route) =>
      json(
        route,
        409,
        problem(409, 'version-conflict', 'el turno cambió'),
        'application/problem+json',
      ),
    )

    await page.goto(ROUTE)
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    await page.getByRole('button', { name: 'Sí, cancelar mi turno' }).click()

    await expect(page.getByText('Tu turno cambió mientras tanto')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancelar turno' })).toHaveCount(0)
  })

  test('se completa solo con teclado y "Mantener mi turno" no envía nada', async ({ page }) => {
    await page.clock.setFixedTime(BEFORE_DEADLINE)
    await mockAppointment(page)
    let posts = 0
    await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}/cancel`, (route) => {
      posts += 1
      return json(route, 200, { status: 'cancelled_by_customer' })
    })

    await page.goto(ROUTE)
    await page.getByRole('button', { name: 'Cancelar turno' }).focus()
    await page.keyboard.press('Enter')
    await expect(page.getByLabel(/Motivo/)).toBeFocused()
    await page.getByRole('button', { name: 'Mantener mi turno' }).focus()
    await page.keyboard.press('Enter')
    await expect(page.getByLabel(/Motivo/)).toHaveCount(0)
    expect(posts).toBe(0)
  })
})

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  // Aproximación de zoom 200% sobre 1280: mitad de ancho.
  { name: '1280-zoom200', width: 640, height: 450 },
]

for (const viewport of viewports) {
  test.describe(`Evidencia cancelación ${viewport.name}px`, () => {
    test.use({
      viewport: { width: viewport.width, height: viewport.height },
      contextOptions: { reducedMotion: 'reduce' },
    })

    async function noHorizontalScroll(page: Page) {
      const overflow = await page.evaluate(
        () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
      )
      expect(overflow).toBe(false)
    }

    test('formulario de confirmación con la consecuencia', async ({ page }) => {
      await page.clock.setFixedTime(AFTER_DEADLINE)
      await mockAppointment(page)
      await page.goto(ROUTE)
      await page.getByRole('button', { name: 'Cancelar turno' }).click()
      await expect(page.getByText('No se puede deshacer')).toBeVisible()
      await noHorizontalScroll(page)
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'confirmar.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test('rechazo por política con el contacto seguro', async ({ page }) => {
      await page.clock.setFixedTime(AFTER_DEADLINE)
      await mockAppointment(page, {
        ...APPOINTMENT_VIEW,
        lateCancellationClientAllowed: false,
        lateCancellationReasonRequired: false,
      })
      await page.goto(ROUTE)
      await expect(page.getByText('El plazo para cancelar por tu cuenta ya venció')).toBeVisible()
      await noHorizontalScroll(page)
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'rechazo-contacto.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })

    test('estado cancelado', async ({ page }) => {
      await page.clock.setFixedTime(BEFORE_DEADLINE)
      await mockAppointment(page)
      await page.route(`**/api/v1/customer/appointments/${ACCESS_TOKEN}/cancel`, (route) =>
        json(route, 200, { status: 'cancelled_by_customer' }),
      )
      await page.goto(ROUTE)
      await page.getByRole('button', { name: 'Cancelar turno' }).click()
      await page.getByRole('button', { name: 'Sí, cancelar mi turno' }).click()
      await expect(page.getByText('Cancelaste tu turno')).toBeVisible()
      await noHorizontalScroll(page)
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'cancelado.png'),
        fullPage: true,
      })
      await assertNoAxeViolations(page)
    })
  })
}
