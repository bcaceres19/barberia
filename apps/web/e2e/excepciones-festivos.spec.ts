import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-041 (excepciones de jornada y festivos colombianos).
 * Corre contra el API real en local (`go run ./cmd/api` sobre PostgreSQL
 * real con las catorce migraciones aplicadas y un usuario con un hash
 * argon2id REAL), mismo criterio que `e2e/horarios.spec.ts` (HU-040).
 *
 * Variables de entorno: las mismas que horarios.spec.ts.
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'
const EMAIL_B = process.env.E2E_EMAIL_B ?? 'dueno.b@ejemplo.test'
const PASSWORD_B = process.env.E2E_PASSWORD_B ?? 'ClaveDePruebaHU010!'

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

async function openSchedules(page: Page) {
  await page.getByRole('link', { name: 'Horarios' }).click()
  await expect(page).toHaveURL(/\/panel\/horarios$/)
  await expect(page.getByRole('heading', { name: 'Horarios' })).toBeVisible()
}

async function addException(
  page: Page,
  { effectiveDate, closed }: { effectiveDate: string; closed: boolean },
) {
  await page.getByRole('button', { name: 'Agregar excepción' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar excepción' })
  await dialog.getByLabel('Fecha').fill(effectiveDate)
  if (!closed) {
    await dialog.getByLabel('Abierto con tramos especiales').check()
  }
  return dialog
}

test.describe('Excepciones de jornada y festivos (HU-041)', () => {
  // Serial a propósito: varias pruebas comparten la MISMA barbería real
  // (shopA/shopB), mismo criterio que horarios.spec.ts.
  test.describe.configure({ mode: 'serial' })

  test('activar el calendario de festivos persiste tras recargar (CA-041-01/02)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Festivos ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    const toggle = page.getByLabel(
      'Cerrar automáticamente los festivos colombianos de este barbero',
    )
    await expect(toggle).not.toBeChecked()
    await toggle.check()
    await expect(toggle).toBeChecked()

    await page.reload()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(toggle).toBeChecked()
  })

  test('agregar una excepción cerrada persiste tras recargar (CA-041-04)', async ({ page }) => {
    const stamp = Date.now()
    const barberName = `E2E Excepcion Cerrada ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    const dialog = await addException(page, { effectiveDate: '2027-12-08', closed: true })
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog).toBeHidden()
    await expect(page.getByText('2027-12-08')).toBeVisible()
    await expect(page.getByText('Cerrado')).toBeVisible()

    await page.reload()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(page.getByText('2027-12-08')).toBeVisible()
  })

  test('agregar una excepción abierta con un tramo especial persiste (CA-041-04)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Excepcion Abierta ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    const dialog = await addException(page, { effectiveDate: '2027-07-20', closed: false })
    await dialog.getByLabel('Hora de inicio').fill('09:00')
    await dialog.getByLabel('Duración (minutos)').fill('180')
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog).toBeHidden()
    await expect(page.getByText('09:00 (180 min)')).toBeVisible()
  })

  test('una fecha duplicada se rechaza sin cerrar el diálogo (CA-041-05)', async ({ page }) => {
    const stamp = Date.now()
    const barberName = `E2E Fecha Duplicada ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    const first = await addException(page, { effectiveDate: '2027-12-25', closed: true })
    await first.getByRole('button', { name: 'Guardar' }).click()
    await expect(first).toBeHidden()

    const second = await addException(page, { effectiveDate: '2027-12-25', closed: true })
    await second.getByRole('button', { name: 'Guardar' }).click()
    await expect(second).toBeVisible()
    await expect(second.getByText('Ya existe una excepción para esa fecha')).toBeVisible()
    await second.getByRole('button', { name: 'Cancelar' }).click()
    await expect(second).toBeHidden()
  })

  test('editar una excepción cambia su fecha, y retirarla la elimina de la lista', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Editar Retirar Excepcion ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    const dialog = await addException(page, { effectiveDate: '2027-05-01', closed: true })
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog).toBeHidden()
    await expect(page.getByText('2027-05-01')).toBeVisible()

    await page.getByRole('button', { name: /Editar excepción del 2027-05-01/ }).click()
    const editDialog = page.getByRole('dialog', { name: 'Editar excepción' })
    await editDialog.getByLabel('Fecha').fill('2027-05-02')
    await editDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(editDialog).toBeHidden()
    await expect(page.getByText('2027-05-02')).toBeVisible()
    await expect(page.getByText('2027-05-01')).not.toBeVisible()

    await page.getByRole('button', { name: /Retirar excepción del 2027-05-02/ }).click()
    await expect(page.getByText('2027-05-02')).not.toBeVisible()

    await page.reload()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(page.getByText('2027-05-02')).not.toBeVisible()
  })

  test('un identificador real de otra barbería responde 404 al consultar el calendario de festivos (RN-TEN-01)', async ({
    page,
  }) => {
    await login(page, EMAIL_B, PASSWORD_B)
    const stamp = Date.now()
    const nameB = `E2E Festivos Solo De B ${stamp}`
    await addBarber(page, nameB)

    const barberBId = await page.evaluate(async (name) => {
      const res = await fetch('/api/v1/private/barbers?limit=50', { credentials: 'include' })
      const body = (await res.json()) as { items: { id: string; fullName: string }[] }
      return body.items.find((b) => b.fullName === name)?.id ?? null
    }, nameB)
    expect(barberBId).toBeTruthy()

    await page.goto('/acceso')
    await login(page, EMAIL, PASSWORD)

    const crossTenantStatus = await page.evaluate(async (id) => {
      const res = await fetch(`/api/v1/private/barbers/${id}/holiday-calendar`, {
        credentials: 'include',
      })
      return res.status
    }, barberBId)
    expect(crossTenantStatus).toBe(404)
  })
})
