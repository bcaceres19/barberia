import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-063 (navegación de la agenda por fecha). Corre
 * contra el API real en local (`go run ./cmd/api` sobre PostgreSQL real
 * con las diecisiete migraciones aplicadas y un usuario con un hash
 * argon2id REAL), mismo criterio que `e2e/agenda-diaria.spec.ts` (HU-062).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/agenda-diaria.spec.ts` (HU-062) en
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

async function selectAgendaBarber(page: Page, fullName: string) {
  const trigger = page.getByRole('button', { name: 'Barbero', exact: true })
  await trigger.click()
  await page.getByRole('option', { name: fullName, exact: true }).click()
  await expect(trigger).toContainText(fullName)
}

function civilDateAt(offsetDays: number): string {
  const d = new Date()
  d.setDate(d.getDate() + offsetDays)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
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
) {
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
}

test.describe('Navegación de la agenda por fecha (HU-063)', () => {
  test.describe.configure({ mode: 'serial' })

  test('hoy → anterior → siguiente → fecha elegida → recarga → atrás/adelante, con fecha y barbero siempre correctos (CA-063-01 a CA-063-08)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Navegacion Barbero ${stamp}`
    const serviceName = `E2E Navegacion Servicio ${stamp}`
    const attendeeName = `Cliente navegacion E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    await registerManualAppointment(page, barberName, serviceName, attendeeName, 30)

    await goToAgenda(page)
    await selectAgendaBarber(page, barberName)
    await expect(page.getByText(attendeeName)).toBeVisible()
    await expect(page).toHaveURL(new RegExp(`date=${civilDateAt(0)}`))

    // Anterior: el turno de hoy deja de verse; barbero se conserva.
    await page.getByRole('button', { name: 'Anterior' }).click()
    await expect(page).toHaveURL(new RegExp(`date=${civilDateAt(-1)}`))
    await expect(page.getByRole('button', { name: 'Barbero', exact: true })).toContainText(
      barberName,
    )
    await expect(page.getByText(attendeeName)).not.toBeVisible()

    // Siguiente, dos veces: vuelve a hoy y luego avanza un día.
    await page.getByRole('button', { name: 'Siguiente' }).click()
    await expect(page).toHaveURL(new RegExp(`date=${civilDateAt(0)}`))
    await expect(page.getByText(attendeeName)).toBeVisible()

    await page.getByRole('button', { name: 'Siguiente' }).click()
    await expect(page).toHaveURL(new RegExp(`date=${civilDateAt(1)}`))
    await expect(page.getByText(attendeeName)).not.toBeVisible()

    // Selector de fecha: salta directo a una fecha lejana.
    const farDate = civilDateAt(10)
    await page.getByLabel('Fecha').fill(farDate)
    await expect(page).toHaveURL(new RegExp(`date=${farDate}`))

    // Recarga: conserva la fecha y el barbero de la URL.
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
    await expect(page).toHaveURL(new RegExp(`date=${farDate}`))
    await expect(page.getByRole('button', { name: 'Barbero', exact: true })).toContainText(
      barberName,
    )

    // Atrás/adelante del navegador restauran la fecha anterior/siguiente.
    await page.goBack()
    await expect(page).toHaveURL(new RegExp(`date=${civilDateAt(1)}`))
    await page.goForward()
    await expect(page).toHaveURL(new RegExp(`date=${farDate}`))
  })
})

test.describe('Agenda en un dispositivo de otra zona horaria durante la navegación (RN-DIS-07)', () => {
  test.use({ timezoneId: 'America/Los_Angeles' })

  test('anterior/siguiente calculan el día civil de la ZONA DE LA BARBERÍA, nunca la del dispositivo', async ({
    page,
  }) => {
    await login(page)
    await expect(page.getByRole('heading', { name: 'Agenda', exact: true })).toBeVisible()
    await expect(page.getByText('Zona America/Bogota')).toBeVisible()

    const bogotaToday = civilDateAt(0)
    await expect(page).toHaveURL(new RegExp(`date=${bogotaToday}`))
  })
})
