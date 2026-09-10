import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-066 (cancelación de un turno por el barbero, T6).
 * Corre contra el API real en local (`go run ./cmd/api` sobre PostgreSQL
 * real con las diecisiete migraciones aplicadas y un usuario con un hash
 * argon2id REAL), mismo criterio que `e2e/agenda-reprogramacion-turno.spec.ts`
 * (HU-065).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/agenda-reprogramacion-turno.spec.ts` (HU-065) en
 * `apps/api/README.md`.
 *
 * Variables de entorno:
 *   E2E_EMAIL / E2E_PASSWORD: credenciales válidas contra el API real
 *     (barbería A, "shopA").
 *   APP_BASE_URL: base del frontend (por defecto http://localhost:5173).
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

async function login(page: Page) {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
  await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel$/)
}

async function addBarber(page: Page, fullName: string) {
  await page.getByRole('link', { name: 'Barberos' }).click()
  await expect(page).toHaveURL(/\/panel\/barberos$/)
  await page.getByRole('button', { name: 'Agregar barbero' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
  await dialog.getByLabel('Nombre').fill(fullName)
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()
}

async function addService(page: Page, name: string) {
  await page.getByRole('link', { name: 'Servicios', exact: true }).click()
  await expect(page).toHaveURL(/\/panel\/servicios$/)
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
  await dialog.getByLabel('Nombre').fill(name)
  await dialog.getByLabel('Duración (minutos)').fill('30')
  await dialog.getByLabel('Precio (COP)').fill('20000.00')
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()
}

async function assignServiceToBarber(page: Page, barberName: string, serviceName: string) {
  await page.getByRole('link', { name: 'Servicios por barbero' }).click()
  await expect(page).toHaveURL(/\/panel\/servicios-por-barbero$/)
  await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
  await page.getByRole('checkbox', { name: serviceName }).check()
  await expect(page.getByRole('checkbox', { name: serviceName })).toBeChecked()
}

async function goToAgenda(page: Page) {
  await page.getByRole('link', { name: 'Panel' }).click()
  await expect(page).toHaveURL(/\/panel$/)
  await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
}

function nearFutureCivilDateTime(minutesFromNow: number): { date: string; time: string } {
  const d = new Date(Date.now() + minutesFromNow * 60_000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return {
    date: `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`,
    time: `${pad(d.getHours())}:${pad(d.getMinutes())}`,
  }
}

async function registerManualAppointment(
  page: Page,
  barberName: string,
  serviceName: string,
  attendeeName: string,
  minutesFromNow: number,
): Promise<{ date: string; time: string }> {
  const { date, time } = nearFutureCivilDateTime(minutesFromNow)
  await page.getByRole('link', { name: 'Nuevo turno' }).click()
  await expect(page).toHaveURL(/\/panel\/turnos\/nuevo$/)
  await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
  await page.getByLabel('Servicio', { exact: true }).selectOption({ label: serviceName })
  await page.getByLabel('Persona atendida').fill(attendeeName)
  await page.getByLabel('Nombre del cliente').fill(attendeeName)
  await page.getByLabel('Fecha del turno').fill(date)
  await page.getByLabel('Hora del turno').fill(time)
  await page.getByRole('button', { name: 'Registrar turno' }).click()
  await expect(page.getByText('Turno registrado')).toBeVisible()
  return { date, time }
}

test.describe('Cancelación de un turno por el barbero (HU-066, T6)', () => {
  test.describe.configure({ mode: 'serial' })

  test('cancela una cita futura, el historial registra un único evento, la agenda libera la franja y una cita nueva puede ocuparla (CA-066-01 a CA-066-08)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Cancel Barbero ${stamp}`
    const serviceName = `E2E Cancel Servicio ${stamp}`
    const attendeeName = `Cliente cancel E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    const { date, time } = await registerManualAppointment(
      page,
      barberName,
      serviceName,
      attendeeName,
      2 * 24 * 60 + 30,
    )

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByLabel('Fecha').fill(date)
    await expect(page.getByText(attendeeName)).toBeVisible()

    await page.getByText(attendeeName).click()
    await expect(page).toHaveURL(/\/panel\/turnos\/[^/]+\?/)

    // Cancela desde el detalle (T6, CA-066-08): consecuencia explícita antes
    // de confirmar.
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
    await expect(dialog).toBeVisible()
    await expect(dialog.getByText('no se puede deshacer')).toBeVisible()
    await dialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(dialog).toBeHidden()

    // El detalle recargado muestra el estado terminal y el historial
    // registra exactamente un evento de cancelación (CA-066-01, CA-066-07);
    // la acción ya no se ofrece sobre un turno terminal (CA-066-08).
    await expect(page.getByRole('heading', { name: attendeeName })).toBeVisible()
    await expect(page.getByText('Cancelado por el barbero')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Cancelar turno' })).toHaveCount(0)
    await expect(page.getByText('Cancelado por el barbero', { exact: false })).toHaveCount(2) // insignia + evento de historial

    // La agenda de esa fecha ya no muestra el turno cancelado (CA-066-03:
    // deja de participar en la exclusión de inmediato).
    await page.getByRole('link', { name: 'Volver a la agenda' }).click()
    await expect(page).toHaveURL(new RegExp(`date=${date}`))
    await expect(page.getByText(attendeeName)).not.toBeVisible()

    // Una cita nueva puede ocupar legítimamente la franja liberada.
    const newAttendeeName = `Cliente franja liberada E2E ${stamp}`
    await page.getByRole('link', { name: 'Nuevo turno' }).click()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await page.getByLabel('Servicio', { exact: true }).selectOption({ label: serviceName })
    await page.getByLabel('Persona atendida').fill(newAttendeeName)
    await page.getByLabel('Nombre del cliente').fill(newAttendeeName)
    await page.getByLabel('Fecha del turno').fill(date)
    await page.getByLabel('Hora del turno').fill(time)
    await page.getByRole('button', { name: 'Registrar turno' }).click()
    await expect(page.getByText('Turno registrado')).toBeVisible()

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByLabel('Fecha').fill(date)
    await expect(page.getByText(newAttendeeName)).toBeVisible()
  })

  // La repetición de una cancelación por el mismo barbero (CA-066-04, un
  // reintento con clave nueva sobre una cita ya cancelled_by_barber) no es
  // observable desde esta pantalla: el botón "Cancelar turno" desaparece en
  // cuanto el turno deja de estar confirmed (CA-066-08), así que la UI real
  // nunca ofrece una segunda cancelación sobre el mismo turno. Ese
  // desenlace ya está cubierto contra PostgreSQL real en
  // apps/api/internal/modules/booking/postgres/cancel_repository_test.go
  // (TestCancelByBarber_RepeatedByBarber_IsSuccessNoOpNoDuplicateEvent).

  test('reintentar el mismo envío tras un error de red no duplica el evento de cancelación (idempotencia)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Cancel Idem Barbero ${stamp}`
    const serviceName = `E2E Cancel Idem Servicio ${stamp}`
    const attendeeName = `Cliente cancel idem E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    await registerManualAppointment(page, barberName, serviceName, attendeeName, 2 * 24 * 60 + 30)

    await goToAgenda(page)
    await page.getByRole('link', { name: 'Nuevo turno' }).click()
    await page.goBack()
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByText(attendeeName).click()

    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })

    // Simula una caída de red justo tras el envío: el navegador queda sin
    // conexión antes de que la respuesta llegue, forzando el camino de
    // reintento con la MISMA clave de idempotencia (mismo criterio que el
    // E2E de reprogramación de HU-065).
    await page.context().setOffline(true)
    await dialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(page.getByText('No pudimos conectar')).toBeVisible()
    await page.context().setOffline(false)

    await dialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(dialog).toBeHidden()

    await expect(page.getByText('Cancelado por el barbero', { exact: false })).toHaveCount(2)
  })
})
