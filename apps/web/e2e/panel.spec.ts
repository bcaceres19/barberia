import { test, expect, type Page, type BrowserContext } from '@playwright/test'

/**
 * Recorrido E2E de HU-012 (docs/03-desarrollo/estrategia-pruebas.md,
 * cascarón del panel privado). Corre contra el API real en local
 * (`go run ./cmd/api` sobre PostgreSQL real con las seis migraciones
 * aplicadas y un usuario con un hash argon2id REAL, mismo criterio que
 * `e2e/acceso.spec.ts`).
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

function sessionCookie(context: BrowserContext) {
  return context.cookies().then((cookies) => cookies.find((c) => c.name === 'barberia_session'))
}

test.describe('Cascarón del panel privado (HU-012)', () => {
  test('sin sesión, cualquier ruta privada redirige a acceso y conserva el destino (CA-012-02)', async ({
    page,
  }) => {
    await page.goto('/panel')
    await expect(page).toHaveURL(/\/acceso\?redirect=%2Fpanel$|\/acceso\?redirect=\/panel$/)

    await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
    await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()

    await expect(page).toHaveURL(/\/panel$/)
  })

  test('una sesión real sobrevive a cerrar y reabrir el navegador (CA-012-01)', async ({
    browser,
  }) => {
    const context = await browser.newContext()
    const page = await context.newPage()
    await login(page)
    const storageState = await context.storageState()
    await context.close()

    // "Reabrir el navegador": un contexto completamente nuevo que solo
    // reutiliza el estado persistido (cookie HttpOnly incluida), sin
    // ningún estado de JavaScript en memoria.
    const freshContext = await browser.newContext({ storageState })
    const freshPage = await freshContext.newPage()
    await freshPage.goto('/panel')

    await expect(freshPage).toHaveURL(/\/panel$/)
    await expect(freshPage.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()
    await freshContext.close()
  })

  test('una recarga directa de /panel conserva la sesión (CA-012-01)', async ({ page }) => {
    await login(page)
    await page.reload()
    await expect(page).toHaveURL(/\/panel$/)
    await expect(page.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()
  })

  test('la cabecera muestra siempre la barbería activa real (CA-012-04)', async ({ page }) => {
    await login(page)
    await expect(page.getByTestId('barbershop-name')).not.toHaveText('')
    const name = await page.getByTestId('barbershop-name').textContent()
    expect(name?.trim().length ?? 0).toBeGreaterThan(0)
  })

  test('la navegación entre acceso y panel no recarga la aplicación completa (CA-012-06)', async ({
    page,
  }) => {
    // El `load` de la navegación INICIAL a /acceso no cuenta (algunos
    // navegadores, p. ej. Firefox, emiten más de un evento `load` para esa
    // primera carga por razones ajenas al enrutamiento SPA); el listener se
    // registra recién después de que esa carga ya asentó, así que solo
    // captura una recarga completa causada por la transición misma.
    await page.goto('/acceso')
    await expect(page.getByRole('heading', { name: 'Inicia sesión' })).toBeVisible()

    let fullLoads = 0
    page.on('load', () => {
      fullLoads += 1
    })

    await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
    await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    await expect(page).toHaveURL(/\/panel$/)

    // La transición a /panel tras el login es navegación SPA
    // (router.push), sin una segunda carga completa del documento.
    expect(fullLoads).toBe(0)
  })

  test('una pérdida de conexión ofrece Reintentar y luego entra normalmente (CA-012-05)', async ({
    page,
  }) => {
    await page.goto('/acceso')
    await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
    await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)

    let attempt = 0
    await page.route('**/api/v1/private/auth/session', async (route) => {
      attempt += 1
      if (attempt === 1) {
        await route.abort('connectionfailed')
        return
      }
      await route.continue()
    })

    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    await expect(page).toHaveURL(/\/panel$/)
    await expect(page.getByText('No pudimos conectar')).toBeVisible()

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page.getByRole('heading', { name: 'Agenda de hoy' })).toBeVisible()
  })

  test('cerrar sesión invalida el servidor y vuelve a acceso sin bucle (CA-012-07)', async ({
    page,
    context,
  }) => {
    await login(page)
    const before = await sessionCookie(context)
    expect(before).toBeTruthy()

    await page.getByRole('button', { name: 'Cerrar sesión' }).click()
    await expect(page).toHaveURL(/\/acceso$/)

    // El material anterior queda revocado en el SERVIDOR, no solo
    // limpiado en el navegador: reinyectarlo manualmente en un contexto
    // nuevo (simulando un token capturado antes) no autentica.
    if (before) {
      const freshContext = await context.browser()!.newContext()
      await freshContext.addCookies([{ ...before, expires: before.expires }])
      const freshPage = await freshContext.newPage()
      await freshPage.goto('/panel')
      await expect(freshPage).toHaveURL(/\/acceso/)
      await freshContext.close()
    }
  })
})
