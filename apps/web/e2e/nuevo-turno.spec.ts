import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-061 (creación manual de turnos). Corre contra el API
 * real en local (`go run ./cmd/api` sobre PostgreSQL real con las
 * diecisiete migraciones aplicadas y un usuario con un hash argon2id REAL,
 * mismo criterio que `e2e/servicios-por-barbero.spec.ts`).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/excepciones-festivos.spec.ts` (HU-041) y
 * `e2e/horarios.spec.ts` (HU-040) en `apps/web/README.md`.
 *
 * Variables de entorno:
 *   E2E_EMAIL / E2E_PASSWORD: credenciales válidas contra el API real
 *     (barbería A, "shopA").
 *   APP_BASE_URL: base del frontend (por defecto http://localhost:5173).
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

async function login(page: Page, email: string, password: string) {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(email)
  await page.getByLabel('Contraseña', { exact: true }).fill(password)
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

async function openNewAppointment(page: Page) {
  await page.getByRole('link', { name: 'Nuevo turno' }).click()
  await expect(page).toHaveURL(/\/panel\/turnos\/nuevo$/)
  await expect(page.getByRole('heading', { name: 'Nuevo turno' })).toBeVisible()
}

// nearFutureCivilDateTime construye una fecha/hora civil dentro de la
// próxima hora: CA-061-01/RN-DIS-04 exigen que un turno manual pueda
// registrarse fuera de la anticipación mínima pública (60 min, DEC-018),
// así que un instante deliberadamente cercano demuestra la excepción.
function nearFutureCivilDateTime(minutesFromNow: number): { date: string; time: string } {
  const d = new Date(Date.now() + minutesFromNow * 60_000)
  const pad = (n: number) => String(n).padStart(2, '0')
  return {
    date: `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`,
    time: `${pad(d.getHours())}:${pad(d.getMinutes())}`,
  }
}

test.describe('Creación manual de turnos (HU-061)', () => {
  test.describe.configure({ mode: 'serial' })

  test('el barbero registra un turno manual dentro de la próxima hora, fuera de la rejilla pública (CA-061-01, RN-DIS-04)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Turno Barbero ${stamp}`
    const serviceName = `E2E Turno Servicio ${stamp}`
    const { date, time } = nearFutureCivilDateTime(5)

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)

    await openNewAppointment(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(
      page.getByLabel('Servicio', { exact: true }).locator('option', { hasText: serviceName }),
    ).toHaveCount(1)
    await page.getByLabel('Servicio', { exact: true }).selectOption({ label: serviceName })
    await page.getByLabel('Persona atendida').fill('Cliente de prueba E2E')
    await page.getByLabel('Nombre del cliente').fill('Cliente de prueba E2E')
    await page.getByLabel('Fecha del turno').fill(date)
    await page.getByLabel('Hora del turno').fill(time)
    await page.getByRole('button', { name: 'Registrar turno' }).click()

    await expect(page.getByText('Turno registrado')).toBeVisible()
    await expect(page.getByText(serviceName)).toBeVisible()
  })

  test('un turno manual que se cruza con uno existente del mismo barbero se rechaza (RN-CON-01)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Cruce Barbero ${stamp}`
    const serviceName = `E2E Cruce Servicio ${stamp}`
    const { date, time } = nearFutureCivilDateTime(120)

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)

    async function submitOnce(attendeeName: string) {
      await openNewAppointment(page)
      await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
      await page.getByLabel('Servicio', { exact: true }).selectOption({ label: serviceName })
      await page.getByLabel('Persona atendida').fill(attendeeName)
      await page.getByLabel('Nombre del cliente').fill(attendeeName)
      await page.getByLabel('Fecha del turno').fill(date)
      await page.getByLabel('Hora del turno').fill(time)
      await page.getByRole('button', { name: 'Registrar turno' }).click()
    }

    await submitOnce('Primer cliente E2E')
    await expect(page.getByText('Turno registrado')).toBeVisible()

    await submitOnce('Segundo cliente E2E')
    await expect(page.getByText('el barbero ya tiene una cita en ese intervalo')).toBeVisible()
  })

  test('un servicio no asignado al barbero elegido no aparece en el selector (DEC-072)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Sin Asignar Barbero ${stamp}`
    const unassignedService = `E2E Sin Asignar Servicio ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)
    await addService(page, unassignedService)
    // Deliberadamente SIN assignServiceToBarber: el servicio existe pero no
    // está asignado a este barbero.

    await openNewAppointment(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(
      page
        .getByLabel('Servicio', { exact: true })
        .locator('option', { hasText: unassignedService }),
    ).toHaveCount(0)
  })
})
