import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-020 (configuración básica de la barbería). Corre
 * contra el API real en local (`go run ./cmd/api` sobre PostgreSQL real con
 * las nueve migraciones aplicadas y un usuario con un hash argon2id REAL,
 * mismo criterio que `e2e/panel.spec.ts`).
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
  await expect(page).toHaveURL(/\/panel$/)
}

test.describe('Configuración básica de la barbería (HU-020)', () => {
  // Serial a propósito: ambas pruebas escriben sobre la MISMA fila real de
  // barbershop (barbería única de prueba, sin aprovisionamiento en esta
  // historia); ejecutarlas en paralelo produciría una carrera entre sus
  // propias escrituras, no un defecto del producto.
  test.describe.configure({ mode: 'serial' })

  test('editar y guardar actualiza la cabecera de inmediato y persiste tras recargar (CA-020-01, CA-020-02)', async ({
    page,
  }) => {
    await login(page)

    // El enlace del dock dice "Configuración" desde la Fase 2 del shell NAVA
    // (src/modules/settings/index.ts); el título de la propia pantalla sigue
    // siendo "Barbería".
    await page.getByRole('link', { name: 'Configuración' }).click()
    await expect(page).toHaveURL(/\/panel\/barberia$/)
    await expect(page.getByRole('heading', { name: 'Barbería' })).toBeVisible()

    // Nombre único por corrida: evita depender del estado dejado por una
    // ejecución anterior de este mismo recorrido.
    const newName = `Barbería E2E ${Date.now()}`
    await page.getByLabel('Nombre').fill(newName)
    await page.getByLabel('Zona horaria').fill('America/Bogota')
    await page.getByLabel('Correo de contacto').fill('contacto-e2e@ejemplo.test')
    await page.getByLabel('Teléfono de contacto').fill('+573009998877')

    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Guardado')).toBeVisible()

    // La cabecera refleja el nuevo nombre SIN recargar la aplicación
    // (CA-020-02).
    await expect(page.getByTestId('barbershop-name')).toHaveText(newName)

    // Persistencia real: una recarga vuelve a consultar el servidor y
    // muestra exactamente lo guardado, en la cabecera y en el formulario.
    await page.reload()
    await expect(page.getByTestId('barbershop-name')).toHaveText(newName)
    await expect(page.getByLabel('Nombre')).toHaveValue(newName)
    await expect(page.getByLabel('Correo de contacto')).toHaveValue('contacto-e2e@ejemplo.test')
    await expect(page.getByLabel('Teléfono de contacto')).toHaveValue('+573009998877')
  })

  test('una zona horaria no reconocida se rechaza sin perder los demás datos escritos (CA-020-03, CA-020-08)', async ({
    page,
  }) => {
    await login(page)
    await page.getByRole('link', { name: 'Configuración' }).click()
    await expect(page).toHaveURL(/\/panel\/barberia$/)

    const beforeName = await page.getByLabel('Nombre').inputValue()
    const attemptedName = `${beforeName} (intento con zona inválida)`
    await page.getByLabel('Nombre').fill(attemptedName)
    await page.getByLabel('Zona horaria').fill('COT')

    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('No pudimos guardar los cambios')).toBeVisible()

    // No perdió lo escrito (CA-020-08): el nombre sigue siendo el que el
    // barbero acaba de escribir, listo para corregir solo la zona.
    await expect(page.getByLabel('Nombre')).toHaveValue(attemptedName)

    // Cero escritura del lado del servidor: recargar muestra el nombre de
    // ANTES del intento, no el que se intentó guardar con la zona inválida.
    await page.reload()
    await expect(page.getByLabel('Nombre')).toHaveValue(beforeName)
  })
})
