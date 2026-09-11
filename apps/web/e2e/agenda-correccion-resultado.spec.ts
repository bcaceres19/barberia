import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-068 (corrección auditada de un resultado terminal,
 * T8). Corre contra el API real en local (`go run ./cmd/api` sobre
 * PostgreSQL real con las dieciséis migraciones aplicadas y un usuario con
 * un hash argon2id REAL), mismo criterio que
 * `e2e/agenda-cancelacion-turno.spec.ts` (HU-066).
 *
 * La cita se lleva a `cancelled_by_barber` (T6, "Cancelar turno") antes de
 * corregirla: es el único camino real de la interfaz para obtener un
 * terminal sin esperar a que `startsAt` pase (el formulario de alta exige
 * un inicio futuro, y "Marcar como atendido"/"que no asistió" solo
 * aparecen una vez que ya pasó). La corrección de destino usada aquí
 * (`cancelled_by_barber` -> `cancelled_by_customer`) no cruza la frontera
 * temporal de CA-068-04 porque ninguno de los dos ocupa agenda. La matriz
 * completa de doce transiciones -- incluidas `completed`/`no_show`, la
 * frontera temporal y el cruce de agenda -- ya está cubierta contra
 * PostgreSQL real en
 * apps/api/internal/modules/booking/postgres/correct_repository_test.go
 * (TestCorrectAppointmentStatus_TwelveTerminalTransitions_*,
 * TestCorrectAppointmentStatus_CancelledToOccupying_*); el componente
 * cubre completed -> no_show con ambos eventos visibles
 * (AppointmentDetailPage.test.ts).
 *
 * Pendiente de ejecución contra Chromium real en esta entrega (no forma
 * parte de los checks de CI, que solo corren `test:unit`), mismo estado
 * documentado que `e2e/agenda-cancelacion-turno.spec.ts` (HU-066) en
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

test.describe('Corrección auditada de un resultado terminal (HU-068, T8)', () => {
  test.describe.configure({ mode: 'serial' })

  test('corrige una cita cancelada hacia otro terminal, ambos eventos quedan en el historial y confirmed nunca aparece como destino (CA-068-01 a CA-068-03, CA-068-08)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Correct Barbero ${stamp}`
    const serviceName = `E2E Correct Servicio ${stamp}`
    const attendeeName = `Cliente correct E2E ${stamp}`

    await login(page)
    await addBarber(page, barberName)
    await addService(page, serviceName)
    await assignServiceToBarber(page, barberName, serviceName)
    const { date } = await registerManualAppointment(
      page,
      barberName,
      serviceName,
      attendeeName,
      2 * 24 * 60 + 30,
    )

    await goToAgenda(page)
    await page.getByLabel('Barbero').selectOption({ label: barberName })
    await page.getByLabel('Fecha').fill(date)
    await page.getByText(attendeeName).click()
    await expect(page).toHaveURL(/\/panel\/turnos\/[^/]+\?/)

    // Deja el turno en un terminal real (T6) antes de corregirlo: único
    // camino de la interfaz sin esperar a que starts_at pase.
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    const cancelDialog = page.getByRole('dialog', { name: 'Cancelar turno' })
    await cancelDialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(cancelDialog).toBeHidden()
    await expect(page.getByText('Cancelado por el barbero')).toBeVisible()

    // CA-068-08: la acción de corrección aparece sobre un terminal.
    await page.getByRole('button', { name: 'Corregir resultado' }).click()
    const correctDialog = page.getByRole('dialog', { name: 'Corregir resultado' })
    await expect(correctDialog).toBeVisible()

    // CA-068-01: confirmed nunca es una opción de destino, ni siquiera
    // corrigiendo una cita cancelada -- una cancelada no "revive" por T8.
    const destinationSelect = correctDialog.getByLabel('Nuevo resultado')
    const destinationOptionLabels = await destinationSelect.locator('option').allTextContents()
    expect(destinationOptionLabels).not.toContain('Confirmado')

    await destinationSelect.selectOption({ label: 'Cancelado por el cliente' })
    await correctDialog
      .getByLabel('Motivo de la corrección')
      .fill('El barbero canceló por error; en realidad canceló el cliente.')
    await correctDialog.getByRole('button', { name: 'Confirmar corrección' }).click()
    await expect(correctDialog).toBeHidden()

    // El detalle recargado muestra el nuevo estado terminal; el historial
    // conserva el evento original (cancelled_by_barber) Y agrega el de
    // corrección, nunca reemplaza ni oculta el anterior (CA-068-03).
    await expect(page.getByRole('heading', { name: attendeeName })).toBeVisible()
    await expect(page.getByText('Cancelado por el cliente')).toBeVisible()
    await expect(page.getByText('Cancelado por el barbero', { exact: false })).toBeVisible()
    await expect(
      page.getByText('El barbero canceló por error; en realidad canceló el cliente.'),
    ).toBeVisible()
    await expect(page.getByRole('button', { name: 'Confirmado' })).toHaveCount(0)
  })

  test('reintentar el mismo envío tras un error de red no duplica el evento de corrección (idempotencia)', async ({
    page,
  }) => {
    const stamp = Date.now()
    const barberName = `E2E Correct Idem Barbero ${stamp}`
    const serviceName = `E2E Correct Idem Servicio ${stamp}`
    const attendeeName = `Cliente correct idem E2E ${stamp}`

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
    const cancelDialog = page.getByRole('dialog', { name: 'Cancelar turno' })
    await cancelDialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(cancelDialog).toBeHidden()

    await page.getByRole('button', { name: 'Corregir resultado' }).click()
    const correctDialog = page.getByRole('dialog', { name: 'Corregir resultado' })
    await correctDialog.getByLabel('Nuevo resultado').selectOption({ label: 'No se presentó' })
    await correctDialog.getByLabel('Motivo de la corrección').fill('Motivo de prueba idempotente.')

    // Simula una caída de red justo tras el envío: el navegador queda sin
    // conexión antes de que la respuesta llegue, forzando el camino de
    // reintento con la MISMA clave de idempotencia (mismo criterio que el
    // E2E de cancelación de HU-066).
    await page.context().setOffline(true)
    await correctDialog.getByRole('button', { name: 'Confirmar corrección' }).click()
    await expect(page.getByText('No pudimos conectar')).toBeVisible()
    await page.context().setOffline(false)

    await correctDialog.getByRole('button', { name: 'Confirmar corrección' }).click()
    await expect(correctDialog).toBeHidden()

    await expect(page.getByText('No se presentó', { exact: false })).toHaveCount(2) // insignia + evento de historial
  })
})
