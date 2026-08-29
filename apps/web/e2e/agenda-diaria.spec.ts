import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-062 (agenda diaria de hoy). Corre contra el API real
 * en local (`go run ./cmd/api` sobre PostgreSQL real con las diecisiete
 * migraciones aplicadas y un usuario con un hash argon2id REAL), mismo
 * criterio que `e2e/nuevo-turno.spec.ts` (HU-061).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/nuevo-turno.spec.ts` (HU-061), `e2e/horarios.spec.ts`
 * (HU-040) y `e2e/excepciones-festivos.spec.ts` (HU-041) en
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

test.describe('Agenda diaria de hoy (HU-062)', () => {
  test.describe.configure({ mode: 'serial' })

  test('al abrir el panel, la agenda muestra la fecha completa y la zona, con selector obligatorio de barbero (CA-062-01, DEC-074)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Agenda Barbero ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await goToAgenda(page)

    await expect(page.locator('#daily-agenda-barber-select')).toBeVisible()
    // La fecha completa y la zona aparecen siempre juntas, nunca solo una
    // hora suelta (CA-062-01).
    await expect(page.getByText(/\d{4}/)).toBeVisible()
    await expect(page.getByText('Zona')).toBeVisible()
  })

  test('un turno recién creado aparece en la agenda de su barbero, con estado Confirmado (CA-062-03)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Agenda Turno Barbero ${stamp}`
    const serviceName = `E2E Agenda Turno Servicio ${stamp}`
    const attendeeName = `Cliente agenda E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    await registerManualAppointment(page, barberName, serviceName, attendeeName, 30)

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberName })

    await expect(page.getByText(attendeeName)).toBeVisible()
    await expect(page.getByText(serviceName)).toBeVisible()
    await expect(page.getByText('Confirmado')).toBeVisible()
  })

  test('cambiar de barbero en el selector muestra la agenda de ese barbero, no la del anterior (DEC-074)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberOne = `E2E Agenda Uno ${stamp}`
    const barberTwo = `E2E Agenda Dos ${stamp}`
    const serviceName = `E2E Agenda Cambio Servicio ${stamp}`
    const attendeeOne = `Cliente uno E2E ${stamp}`

    await login(page)
    await addBarber(page, barberOne)
    await addBarber(page, barberTwo)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberOne, serviceName)
    await registerManualAppointment(page, barberOne, serviceName, attendeeOne, 45)

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberOne })
    await expect(page.getByText(attendeeOne)).toBeVisible()

    await page.getByLabel('Barbero').selectOption({ label: barberTwo })
    await expect(page.getByText(attendeeOne)).not.toBeVisible()
    await expect(page.getByText(`No hay turnos para ${barberTwo} hoy`)).toBeVisible()
  })

  test('"Nuevo turno" desde la agenda navega al formulario real de HU-061', async ({ page }) => {
    await login(page)
    await goToAgenda(page)
    await page.getByRole('button', { name: 'Nuevo turno' }).click()
    await expect(page).toHaveURL(/\/panel\/turnos\/nuevo$/)
    await expect(page.getByRole('heading', { name: 'Nuevo turno' })).toBeVisible()
  })
})

test.describe('Agenda diaria en un dispositivo de otra zona horaria (CA-062-01, RN-DIS-07)', () => {
  // El dispositivo simula America/Los_Angeles; la barbería de las
  // credenciales de prueba está configurada en otra zona (America/Bogota,
  // ver database/testdata/dos_barberias.sql). La fecha/hora mostradas deben
  // seguir siendo las de la ZONA DE LA BARBERÍA, nunca las del dispositivo.
  test.use({ timezoneId: 'America/Los_Angeles' })

  test('la agenda muestra la zona configurada de la barbería, no la del dispositivo', async ({
    page,
  }) => {
    await login(page)
    await expect(page.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()
    await expect(page.getByText('Zona America/Bogota')).toBeVisible()
  })
})
