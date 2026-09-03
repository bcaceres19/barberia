import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-040 (horario laboral recurrente). Corre contra el API
 * real en local (`go run ./cmd/api` sobre PostgreSQL real con las trece
 * migraciones aplicadas y un usuario con un hash argon2id REAL, mismo
 * criterio que `e2e/servicios-por-barbero.spec.ts`).
 *
 * Variables de entorno:
 *   E2E_EMAIL / E2E_PASSWORD: credenciales válidas contra el API real
 *     (barbería A, "shopA").
 *   E2E_EMAIL_B / E2E_PASSWORD_B: credenciales válidas de una SEGUNDA
 *     barbería real ("shopB"), usadas solo para la prueba de aislamiento
 *     entre tenants.
 *   APP_BASE_URL: base del frontend (por defecto http://localhost:5173).
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

async function addWorkingHour(
  page: Page,
  {
    weekday,
    startsTime,
    durationMinutes,
  }: { weekday: string; startsTime: string; durationMinutes: string },
) {
  await page.getByRole('button', { name: 'Agregar tramo' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar tramo' })
  await dialog.getByLabel('Día').selectOption({ label: weekday })
  await dialog.getByLabel('Hora de inicio').fill(startsTime)
  await dialog.getByLabel('Duración (minutos)').fill(durationMinutes)
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  return dialog
}

test.describe('Horario laboral recurrente (HU-040)', () => {
  // Serial a propósito: varias pruebas comparten la MISMA barbería real
  // (shopA/shopB, sin aprovisionamiento de barberías nuevas en esta
  // historia); ejecutarlas en paralelo produciría una carrera entre sus
  // propias escrituras, no un defecto del producto.
  test.describe.configure({ mode: 'serial' })

  test('agregar un tramo persiste tras recargar, bajo su día correcto (CA-040-01/02)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Barbero Horario ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    const dialog = await addWorkingHour(page, {
      weekday: 'Lunes',
      startsTime: '08:00',
      durationMinutes: '240',
    })
    await expect(dialog).toBeHidden()
    await expect(page.getByText('08:00 · 240 min')).toBeVisible()

    await page.reload()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(page.getByText('08:00 · 240 min')).toBeVisible()
  })

  test('una jornada partida agrega dos tramos no solapados al mismo día (CA-040-02)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Jornada Partida ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    await addWorkingHour(page, { weekday: 'Martes', startsTime: '08:00', durationMinutes: '240' })
    await expect(page.getByText('08:00 · 240 min')).toBeVisible()

    await addWorkingHour(page, { weekday: 'Martes', startsTime: '14:00', durationMinutes: '240' })
    await expect(page.getByText('14:00 · 240 min')).toBeVisible()
    await expect(page.getByText('08:00 · 240 min')).toBeVisible()
  })

  test('un tramo que se solapa con otro existente se rechaza sin cerrar el diálogo (CA-040-04)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Solape ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })

    await addWorkingHour(page, {
      weekday: 'Miércoles',
      startsTime: '08:00',
      durationMinutes: '120',
    })
    await expect(page.getByText('08:00 · 120 min')).toBeVisible()

    const dialog = await addWorkingHour(page, {
      weekday: 'Miércoles',
      startsTime: '09:00',
      durationMinutes: '60',
    })
    await expect(dialog).toBeVisible()
    await expect(dialog.getByText('Este tramo se solapa con otro existente')).toBeVisible()
    await dialog.getByRole('button', { name: 'Cancelar' }).click()
    await expect(dialog).toBeHidden()

    // Nada nuevo se persistió: sigue habiendo un único tramo ese día.
    await page.reload()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(page.getByText('09:00 · 60 min')).not.toBeVisible()
  })

  test('editar un tramo cambia su intervalo, y retirarlo lo elimina de la lista', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Editar Retirar ${stamp}`

    await login(page, EMAIL, PASSWORD)
    await addBarber(page, barberName)

    await openSchedules(page)
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await addWorkingHour(page, { weekday: 'Jueves', startsTime: '08:00', durationMinutes: '60' })
    await expect(page.getByText('08:00 · 60 min')).toBeVisible()

    await page.getByRole('button', { name: /Editar tramo de Jueves/ }).click()
    const editDialog = page.getByRole('dialog', { name: 'Editar tramo' })
    await editDialog.getByLabel('Hora de inicio').fill('09:00')
    await editDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(editDialog).toBeHidden()
    await expect(page.getByText('09:00 · 60 min')).toBeVisible()
    await expect(page.getByText('08:00 · 60 min')).not.toBeVisible()

    await page.getByRole('button', { name: /Retirar tramo de Jueves/ }).click()
    await expect(page.getByText('09:00 · 60 min')).not.toBeVisible()

    // Persistencia real: recargar confirma que el retiro fue físico, no
    // solo un cambio visual.
    await page.reload()
    await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
    await expect(page.getByText('09:00 · 60 min')).not.toBeVisible()
  })

  test('un identificador real de otra barbería responde 404 al consultar su horario (RN-TEN-01)', async ({
    page,
  }) => {
    await login(page, EMAIL_B, PASSWORD_B)
    const stamp = Date.now()
    const nameB = `E2E Solo De B ${stamp}`
    await addBarber(page, nameB)

    const barberBId = await page.evaluate(async (name) => {
      const res = await fetch('/api/v1/private/barbers?limit=50', { credentials: 'include' })
      const body = (await res.json()) as { items: { id: string; fullName: string }[] }
      return body.items.find((b) => b.fullName === name)?.id ?? null
    }, nameB)
    expect(barberBId).toBeTruthy()

    // Cambia de sesión a la barbería A.
    await page.goto('/acceso')
    await login(page, EMAIL, PASSWORD)

    const crossTenantStatus = await page.evaluate(async (id) => {
      const res = await fetch(`/api/v1/private/barbers/${id}/working-hours`, {
        credentials: 'include',
      })
      return res.status
    }, barberBId)
    expect(crossTenantStatus).toBe(404)

    // El selector de A nunca ofrece el barbero de B.
    await openSchedules(page)
    await expect(
      page.getByLabel('Barbero', { exact: true }).locator('option', { hasText: nameB }),
    ).toHaveCount(0)
  })
})
