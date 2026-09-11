import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-093 (configuración pública de reserva y cancelación).
 * Corre contra el API real en local (`go run ./cmd/api` sobre PostgreSQL
 * real con las migraciones aplicadas, incluida
 * `20260911060000_add_barbershop_booking_policy.sql`, y un usuario con un
 * hash argon2id REAL), mismo criterio que `e2e/configuracion-barberia.spec.ts`
 * (HU-020).
 *
 * Variables de entorno:
 *   E2E_EMAIL / E2E_PASSWORD: credenciales válidas contra el API real.
 *   APP_BASE_URL: base del frontend (por defecto http://localhost:5173).
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

async function login(page: Page) {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
  await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel/)
}

test.describe('Configuración pública de reserva y cancelación (HU-093)', () => {
  // Serial a propósito: las tres pruebas escriben sobre la MISMA fila real
  // de barbershop (barbería única de prueba), mismo criterio que
  // configuracion-barberia.spec.ts: ejecutarlas en paralelo produciría una
  // carrera entre sus propias escrituras, no un defecto del producto.
  test.describe.configure({ mode: 'serial' })

  test('editar y guardar persiste tras recargar, con un token de versión renovado (CA-093-01, CA-093-02, CA-093-04)', async ({
    page,
  }) => {
    await login(page)

    await page.getByRole('link', { name: 'Reserva pública' }).click()
    await expect(page).toHaveURL(/\/panel\/reserva-publica$/)
    await expect(
      page.getByRole('heading', { name: 'Configuración de reserva pública' }),
    ).toBeVisible()

    await page.getByLabel('Anticipación mínima (minutos)').fill('90')
    await page.getByLabel('Ventana máxima (días)').fill('5')
    await page.getByLabel('Rejilla de horarios (minutos)').selectOption('30')
    await page.getByLabel('Plazo de cancelación del cliente (minutos)').fill('60')
    await page.getByLabel('El cliente puede cancelar vencido el plazo').uncheck()
    await page.getByLabel('Esa cancelación exige un motivo').uncheck()

    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Guardado')).toBeVisible()

    // Persistencia real: una recarga vuelve a consultar el servidor y
    // muestra exactamente lo guardado.
    await page.reload()
    await expect(page.getByLabel('Anticipación mínima (minutos)')).toHaveValue('90')
    await expect(page.getByLabel('Ventana máxima (días)')).toHaveValue('5')
    await expect(page.getByLabel('Rejilla de horarios (minutos)')).toHaveValue('30')
    await expect(page.getByLabel('Plazo de cancelación del cliente (minutos)')).toHaveValue('60')
    await expect(page.getByLabel('El cliente puede cancelar vencido el plazo')).not.toBeChecked()

    // Un segundo guardado inmediato (sin recargar de por medio) confirma
    // que la página renovó su propio versionToken tras la primera
    // escritura: si siguiera usando el token viejo, este PUT respondería
    // 409 en vez de 200 (CA-093-02).
    await page.getByLabel('Anticipación mínima (minutos)').fill('45')
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Guardado')).toBeVisible()
    await expect(page.getByText('La configuración cambió')).toHaveCount(0)

    // Restaura los defaults de DEC-083 para no dejar estado mutado entre
    // ejecuciones de esta suite.
    await page.getByLabel('Anticipación mínima (minutos)').fill('60')
    await page.getByLabel('Ventana máxima (días)').fill('3')
    await page.getByLabel('Rejilla de horarios (minutos)').selectOption('15')
    await page.getByLabel('Plazo de cancelación del cliente (minutos)').fill('20')
    await page.getByLabel('El cliente puede cancelar vencido el plazo').check()
    await page.getByLabel('Esa cancelación exige un motivo').check()
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Guardado')).toBeVisible()
  })

  test('una combinación incoherente se rechaza en el cliente sin llamar al servidor (CA-093-02)', async ({
    page,
  }) => {
    await login(page)
    await page.getByRole('link', { name: 'Reserva pública' }).click()
    await expect(page).toHaveURL(/\/panel\/reserva-publica$/)

    await page.getByLabel('El cliente puede cancelar vencido el plazo').uncheck()
    // "Esa cancelación exige un motivo" sigue marcado desde el default: la
    // combinación (cliente no puede cancelar tarde, pero se exige motivo)
    // es incoherente.
    await page.getByRole('button', { name: 'Guardar cambios' }).click()

    await expect(
      page.getByText('No puedes exigir motivo si el cliente no puede cancelar tarde.'),
    ).toBeVisible()
    // No se guardó nada: el checkbox marcado por el usuario sigue
    // desmarcado (CA-093-02, sin cambios ante un error de validación).
    await expect(page.getByLabel('El cliente puede cancelar vencido el plazo')).not.toBeChecked()
  })

  test('un conflicto real de versión permite recargar la representación vigente (CA-093-02)', async ({
    page,
    context,
  }) => {
    await login(page)
    await page.getByRole('link', { name: 'Reserva pública' }).click()
    await expect(page).toHaveURL(/\/panel\/reserva-publica$/)

    // Segunda pestaña con la MISMA sesión: escribe primero, invalidando el
    // token que la primera pestaña ya leyó.
    const secondPage = await context.newPage()
    await secondPage.goto('/panel/reserva-publica')
    await expect(
      secondPage.getByRole('heading', { name: 'Configuración de reserva pública' }),
    ).toBeVisible()
    await secondPage.getByLabel('Anticipación mínima (minutos)').fill('75')
    await secondPage.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(secondPage.getByText('Guardado')).toBeVisible()
    await secondPage.close()

    // La primera pestaña todavía tiene el token viejo: su escritura debe
    // rechazarse con el conflicto de versión, no aplicarse como último
    // escritor silencioso.
    await page.getByLabel('Anticipación mínima (minutos)').fill('80')
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('La configuración cambió')).toBeVisible()

    await page.getByRole('button', { name: 'Recargar' }).click()
    await expect(page.getByLabel('Anticipación mínima (minutos)')).toHaveValue('75')

    // Restaura el default.
    await page.getByLabel('Anticipación mínima (minutos)').fill('60')
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Guardado')).toBeVisible()
  })
})
