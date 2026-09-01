import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-064 (detalle e historial de un turno). Corre contra
 * el API real en local (`go run ./cmd/api` sobre PostgreSQL real con las
 * diecisiete migraciones aplicadas y un usuario con un hash argon2id REAL),
 * mismo criterio que `e2e/agenda-navegacion-fecha.spec.ts` (HU-063).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/agenda-navegacion-fecha.spec.ts` (HU-063) en
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
  customerPhone: string,
  customerNote: string,
  minutesFromNow: number,
): Promise<{ date: string }> {
  const { date, time } = nearFutureCivilDateTime(minutesFromNow)
  await page.getByRole('link', { name: 'Nuevo turno' }).click()
  await expect(page).toHaveURL(/\/panel\/turnos\/nuevo$/)
  await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
  await page.getByLabel('Servicio', { exact: true }).selectOption({ label: serviceName })
  await page.getByLabel('Persona atendida').fill(attendeeName)
  await page.getByLabel('Nombre del cliente').fill(attendeeName)
  await page.getByLabel('Teléfono (opcional)').fill(customerPhone)
  await page.getByLabel('Nota (opcional)').fill(customerNote)
  await page.getByLabel('Fecha del turno').fill(date)
  await page.getByLabel('Hora del turno').fill(time)
  await page.getByRole('button', { name: 'Registrar turno' }).click()
  await expect(page.getByText('Turno registrado')).toBeVisible()
  return { date }
}

test.describe('Detalle e historial de un turno (HU-064)', () => {
  test.describe.configure({ mode: 'serial' })

  test('desde una fecha no actual abre el turno, verifica snapshots e historial, y vuelve a la misma fecha/barbero (CA-064-01 a CA-064-08)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Detalle Barbero ${stamp}`
    const serviceName = `E2E Detalle Servicio ${stamp}`
    const attendeeName = `Cliente detalle E2E ${stamp}`
    const customerPhone = '+573009998877'
    const customerNote = 'Nota de prueba E2E'

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    // Un día completo más adelante: "una fecha no actual" real, no "hoy".
    const { date } = await registerManualAppointment(
      page,
      barberName,
      serviceName,
      attendeeName,
      customerPhone,
      customerNote,
      24 * 60 + 30,
    )

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByLabel('Fecha').fill(date)
    await expect(page).toHaveURL(new RegExp(`date=${date}`))
    await expect(page.getByText(attendeeName)).toBeVisible()

    const barberId = await barberIdFromUrl(page)

    await page.getByText(attendeeName).click()
    await expect(page).toHaveURL(/\/panel\/turnos\/[^/]+\?/)

    // Detalle: snapshots de servicio, contacto y nota (CA-064-01 a CA-064-04).
    await expect(page.getByRole('heading', { name: attendeeName })).toBeVisible()
    await expect(page.getByText(serviceName)).toBeVisible()
    await expect(page.getByText('20000.00')).toBeVisible()
    await expect(page.getByText(customerPhone)).toBeVisible()
    await expect(page.getByText(customerNote)).toBeVisible()
    await expect(page.getByText(barberName)).toBeVisible()

    // Historial: al menos el evento de creación (CA-064-05).
    await expect(page.getByText('Turno creado')).toBeVisible()

    // Volver conserva la misma fecha y barbero (trabajo requerido §3.5).
    await page.getByRole('link', { name: 'Volver a la agenda' }).click()
    await expect(page).toHaveURL(new RegExp(`date=${date}`))
    await expect(page.getByLabel('Barbero')).toHaveValue(barberId)
    await expect(page.getByText(attendeeName)).toBeVisible()
  })

  async function barberIdFromUrl(page: Page): Promise<string> {
    const url = new URL(page.url())
    return url.searchParams.get('barberId') ?? ''
  }
})

test.describe('Detalle de un turno inexistente (HU-064)', () => {
  test('un appointmentId inexistente muestra un estado explícito, no un error genérico', async ({
    page,
  }) => {
    await login(page)
    await page.goto('/panel/turnos/00000000-0000-0000-0000-000000000000')
    await expect(page.getByText('Este turno ya no está disponible')).toBeVisible()
  })
})
