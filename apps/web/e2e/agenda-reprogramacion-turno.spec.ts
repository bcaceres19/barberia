import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-065 (reprogramación auditada de un turno, T2). Corre
 * contra el API real en local (`go run ./cmd/api` sobre PostgreSQL real con
 * las diecisiete migraciones aplicadas y un usuario con un hash argon2id
 * REAL), mismo criterio que `e2e/agenda-detalle-historial.spec.ts` (HU-064).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/agenda-detalle-historial.spec.ts` (HU-064) en
 * `apps/web/README.md`.
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
  await page.getByLabel('Correo').fill(EMAIL)
  await page.getByLabel('Contraseña').fill(PASSWORD)
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
  await expect(page.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()
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

async function barberIdFromUrl(page: Page): Promise<string> {
  const url = new URL(page.url())
  return url.searchParams.get('barberId') ?? ''
}

test.describe('Reprogramación de un turno (HU-065, T2)', () => {
  test.describe.configure({ mode: 'serial' })

  test('reprograma a otro día, la agenda anterior queda vacía, la nueva lo muestra, y el historial registra un único evento (CA-065-01 a CA-065-08)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Reprog Barbero ${stamp}`
    const serviceName = `E2E Reprog Servicio ${stamp}`
    const attendeeName = `Cliente reprog E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    // Dos días completos más adelante: deja margen para que "el nuevo día"
    // (originalDate + 1) siga siendo un instante futuro real al reprogramar.
    const { date: originalDate } = await registerManualAppointment(
      page,
      barberName,
      serviceName,
      attendeeName,
      2 * 24 * 60 + 30,
    )
    const { date: newDate, time: newTime } = nearFutureCivilDateTime(3 * 24 * 60 + 30)

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByLabel('Fecha').fill(originalDate)
    await expect(page.getByText(attendeeName)).toBeVisible()
    const barberId = await barberIdFromUrl(page)

    await page.getByText(attendeeName).click()
    await expect(page).toHaveURL(/\/panel\/turnos\/[^/]+\?/)

    // Reprograma a otro día (T2, DEC-076): el diálogo se prellena con el
    // horario vigente y el envío confirma contra la representación real.
    await page.getByRole('button', { name: 'Reprogramar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Reprogramar turno' })
    await expect(dialog).toBeVisible()
    await dialog.getByLabel('Nueva fecha').fill(newDate)
    await dialog.getByLabel('Nueva hora').fill(newTime)
    await dialog.getByRole('button', { name: 'Confirmar' }).click()
    await expect(dialog).toBeHidden()

    // El detalle recargado muestra el turno en su nuevo horario y el
    // historial registra exactamente un evento de reprogramación
    // (CA-065-05 a CA-065-07).
    await expect(page.getByRole('heading', { name: attendeeName })).toBeVisible()
    await expect(page.getByText('Turno reprogramado')).toBeVisible()
    await expect(page.getByText('Turno reprogramado')).toHaveCount(1)

    // "Volver" apunta a la fecha NUEVA (el turno vigente ya no está en la
    // fecha original, trabajo requerido §3.5): la anterior queda vacía y la
    // nueva lo muestra.
    await page.getByRole('link', { name: 'Volver a la agenda' }).click()
    await expect(page).toHaveURL(new RegExp(`date=${newDate}`))
    await expect(page.getByLabel('Barbero')).toHaveValue(barberId)
    await expect(page.getByText(attendeeName)).toBeVisible()

    await page.getByLabel('Fecha').fill(originalDate)
    await expect(page.getByText(attendeeName)).not.toBeVisible()
  })

  test('reintentar el mismo envío tras un error de red no duplica el evento de reprogramación (idempotencia)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Reprog Idem Barbero ${stamp}`
    const serviceName = `E2E Reprog Idem Servicio ${stamp}`
    const attendeeName = `Cliente reprog idem E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    await registerManualAppointment(page, barberName, serviceName, attendeeName, 2 * 24 * 60 + 30)
    const { date: newDate, time: newTime } = nearFutureCivilDateTime(3 * 24 * 60 + 45)

    await goToAgenda(page)
    await page.getByRole('link', { name: 'Nuevo turno' }).click()
    await page.goBack()
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByText(attendeeName).click()

    await page.getByRole('button', { name: 'Reprogramar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Reprogramar turno' })
    await dialog.getByLabel('Nueva fecha').fill(newDate)
    await dialog.getByLabel('Nueva hora').fill(newTime)

    // Simula una caída de red justo tras el envío: el navegador queda sin
    // conexión antes de que la respuesta llegue, forzando el camino de
    // reintento con la MISMA clave de idempotencia (mismo criterio que
    // e2e HU-004/HU-021 de "doble envío no duplica").
    await page.context().setOffline(true)
    await dialog.getByRole('button', { name: 'Confirmar' }).click()
    await expect(page.getByText('No pudimos conectar')).toBeVisible()
    await page.context().setOffline(false)

    await dialog.getByRole('button', { name: 'Confirmar' }).click()
    await expect(dialog).toBeHidden()

    await expect(page.getByText('Turno reprogramado')).toHaveCount(1)
  })
})
