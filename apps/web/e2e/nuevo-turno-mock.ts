import type { Page, Route } from '@playwright/test'

/**
 * Doble de API y datos sintéticos de `/panel/turnos/nuevo` (issue #190).
 *
 * Vive fuera de `src/` a propósito: estos datos existen SOLO dentro de
 * Playwright, para revisar la composición de cada estado de la pantalla sin
 * iniciar sesión ni tocar API, PostgreSQL, OTP o rate limiting locales. Lo
 * comparten la evidencia de fidelidad (`nuevo-turno-fidelidad-mock.spec.ts`) y
 * la evidencia responsive (`nuevo-turno-evidencia-responsiva.spec.ts`) para
 * que ambas describan exactamente el mismo turno.
 *
 * No termina en `.spec.ts`, así que el `testMatch` por defecto de Playwright
 * no lo recoge como suite.
 */
export type Scenario =
  | 'context-loading'
  | 'context-error'
  | 'no-barbers'
  | 'empty-form'
  | 'services-loading'
  | 'services-empty'
  | 'services-error'
  | 'filled-form'
  | 'validation-error'
  | 'saving'
  | 'schedule-conflict'
  | 'created'

export type Deferred = { release: () => void }

// Datos ficticios (RN-DAT-01/02), los mismos que declara
// `tools/mockups/nuevo-turno-eventos/content.mjs`.
export const BARBER = { id: 'barber-camilo', fullName: 'Camilo Torres' }
export const SERVICE = { id: 'service-corte', name: 'Corte clásico' }
export const ATTENDEE = 'Diego Molina'
export const CUSTOMER = 'Diego Molina'
// El atlas escribe el teléfono con espacios; `validateCustomerPhone` exige
// `^\+[1-9][0-9]{7,14}$`, así que la evidencia usa la forma que el propio
// placeholder del campo declara (ver el archivo de desviaciones del PR).
export const PHONE = '+573001234567'
export const EMAIL = 'diego.molina@email.com'
export const NOTE = 'Cliente frecuente.'
export const CIVIL_DATE = '2026-09-04'
export const TIME = '11:30'

function deferred(): { promise: Promise<void>; deferred: Deferred } {
  let release = () => {}
  const promise = new Promise<void>((resolve) => {
    release = resolve
  })
  return { promise, deferred: { release } }
}

async function respondJson(route: Route, body: unknown, status = 200) {
  await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) })
}

const PENDING_SCENARIOS: readonly Scenario[] = ['context-loading', 'services-loading', 'saving']

export async function installNewAppointmentMock(
  page: Page,
  scenario: Scenario,
): Promise<Deferred | null> {
  const pending = PENDING_SCENARIOS.includes(scenario) ? deferred() : null

  await page.route('**/api/v1/**', async (route) => {
    const { pathname } = new URL(route.request().url())

    if (pathname === '/api/v1/private/auth/session') {
      await respondJson(route, {
        barbershop: { id: 'shop-mock', name: 'NAVA QA Local' },
        expiresAt: '2026-09-05T00:00:00Z',
      })
      return
    }

    if (pathname === '/api/v1/private/barbers') {
      if (scenario === 'context-loading' && pending) await pending.promise
      if (scenario === 'context-error') {
        await respondJson(route, { title: 'No disponible' }, 500)
        return
      }
      await respondJson(route, { items: scenario === 'no-barbers' ? [] : [BARBER] })
      return
    }

    if (pathname === '/api/v1/private/settings/barbershop') {
      if (scenario === 'context-loading' && pending) await pending.promise
      await respondJson(route, { timezone: 'America/Bogota' })
      return
    }

    // Servicios asignados al barbero (HU-023): solo identificadores, el
    // nombre se cruza contra el catálogo real.
    if (pathname.includes('/barbers/') && pathname.endsWith('/services')) {
      if (scenario === 'services-loading' && pending) await pending.promise
      if (scenario === 'services-error') {
        await respondJson(route, { title: 'No disponible' }, 500)
        return
      }
      if (scenario === 'services-empty') {
        await respondJson(route, { items: [] })
        return
      }
      await respondJson(route, {
        items: [{ barberId: BARBER.id, serviceId: SERVICE.id, createdAt: '2026-09-01T12:00:00Z' }],
      })
      return
    }

    if (pathname === '/api/v1/private/services') {
      await respondJson(route, {
        items: [
          {
            id: SERVICE.id,
            name: SERVICE.name,
            durationMinutes: 30,
            priceAmount: '45000.00',
            currency: 'COP',
            isActive: true,
          },
        ],
      })
      return
    }

    if (pathname === '/api/v1/private/appointments' && route.request().method() === 'POST') {
      if (scenario === 'saving' && pending) await pending.promise
      if (scenario === 'schedule-conflict') {
        await respondJson(route, { detail: 'el barbero ya tiene una cita en ese intervalo' }, 409)
        return
      }
      await respondJson(
        route,
        {
          id: 'appt-mock',
          barberId: BARBER.id,
          serviceId: SERVICE.id,
          attendeeName: ATTENDEE,
          startsAt: `${CIVIL_DATE}T${TIME}:00-05:00`,
          endsAt: `${CIVIL_DATE}T12:00:00-05:00`,
          serviceName: SERVICE.name,
          durationMinutes: 30,
          priceAmount: '45000.00',
          currency: 'COP',
        },
        201,
      )
      return
    }

    await route.abort('blockedbyclient')
  })

  return pending?.deferred ?? null
}

/** Abre el listbox de barbero y elige el único del mock. */
export async function chooseBarber(page: Page) {
  await page.getByRole('button', { name: 'Barbero', exact: true }).click()
  await page.getByRole('option', { name: BARBER.fullName }).click()
}

/** Rellena persona atendida, cliente y fecha/hora (sin barbero ni servicio). */
export async function fillCustomerAndSchedule(page: Page) {
  await page.getByLabel('Persona atendida').fill(ATTENDEE)
  await page.getByLabel('Nombre del cliente').fill(CUSTOMER)
  await page.getByLabel('Fecha del turno').fill(CIVIL_DATE)
  await page.getByLabel('Hora del turno').fill(TIME)
}

export async function fillCompleteForm(page: Page) {
  await chooseBarber(page)
  await page.getByLabel('Servicio', { exact: true }).selectOption({ label: SERVICE.name })
  await fillCustomerAndSchedule(page)
  await page.getByLabel('Teléfono (opcional)').fill(PHONE)
  await page.getByLabel('Correo (opcional)').fill(EMAIL)
  await page.getByLabel('Nota (opcional)').fill(NOTE)
  // El atlas dibuja el formulario en reposo: sin este blur la evidencia
  // mostraría el anillo de foco del último campo escrito, que no es un estado
  // del evento sino un artefacto de cómo se llenó.
  await page.locator('#new-appointment-note').blur()
}
