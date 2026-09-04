import { test, expect, type Page } from '@playwright/test'
import { randomBytes } from 'node:crypto'

/**
 * Recorrido E2E de HU-010 (docs/03-desarrollo/estrategia-pruebas.md §5.3,
 * recorrido P0 #4 "barbero inicia sesión"). Corre contra el API real en
 * local (`go run ./cmd/api` sobre PostgreSQL real con las migraciones
 * aplicadas y un usuario con un hash argon2id REAL, no el hash ficticio de
 * `database/testdata/hu005_credenciales_sesiones.sql`): ver el comando
 * reproducible y la fixture en el resumen de verificación de este PR.
 *
 * Variables de entorno:
 *   E2E_EMAIL / E2E_PASSWORD: credenciales válidas contra el API real.
 *   APP_BASE_URL: base del frontend (por defecto http://localhost:5173).
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

async function fillCredentials(page: Page, email: string, password: string) {
  await page.getByLabel('Correo', { exact: true }).fill(email)
  await page.getByLabel('Contraseña', { exact: true }).fill(password)
}

// Este archivo por sí solo emite más de cinco solicitudes reales de login
// (umbral por defecto de DEC-061, APP_LOGIN_THROTTLE_THRESHOLD=5): sin
// aislar cada prueba en su propia IP sintética, la quinta o sexta prueba
// tropieza con el reto telefónico de HU-007 contra su propia cuenta, no
// contra un escenario real de abuso. Mismo patrón que
// `reto-telefonico.spec.ts` (`withIsolatedIP`), verificado en vivo
// (2026-09-04): sin esto, `login_throttle` se comparte entre TODAS las
// pruebas de este archivo (y de `recuperacion.spec.ts` si corre en la
// misma invocación) porque todas comparten el mismo peer real.
function syntheticIP(): string {
  const octets = Array.from({ length: 3 }, () => randomBytes(1)[0])
  return `10.${octets[0]}.${octets[1]}.${octets[2]}`
}

async function withIsolatedIP(page: Page): Promise<void> {
  const ip = syntheticIP()
  await page.route('**/api/v1/public/auth/**', async (route) => {
    await route.continue({ headers: { ...route.request().headers(), 'x-forwarded-for': ip } })
  })
}

test.describe('Acceso del barbero (HU-010)', () => {
  test('accede con credenciales válidas y llega a /panel (CA-010-01)', async ({ page }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')
    await fillCredentials(page, EMAIL, PASSWORD)
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()

    await expect(page).toHaveURL(/\/panel$/)
    await expect(page.getByRole('heading', { name: 'Agenda' })).toBeVisible()
  })

  test('credenciales inválidas son indistinguibles entre correo inexistente y contraseña incorrecta (CA-010-02)', async ({
    page,
  }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')
    await fillCredentials(page, EMAIL, 'contraseña-incorrecta-a-proposito')
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    const wrongPasswordMessage = await page.getByRole('alert').last().textContent()
    await expect(page).toHaveURL(/\/acceso$/)
    // El correo escrito se conserva (CA-010-02); la contraseña no.
    await expect(page.getByLabel('Correo', { exact: true })).toHaveValue(EMAIL)
    await expect(page.getByLabel('Contraseña', { exact: true })).toHaveValue('')

    await fillCredentials(page, 'correo-que-no-existe@ejemplo.test', 'cualquier-cosa')
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    const unknownEmailMessage = await page.getByRole('alert').last().textContent()

    expect(unknownEmailMessage).toBe(wrongPasswordMessage)
  })

  test('un fallo de red conserva lo escrito y permite reintentar sin recargar (CA-010-03)', async ({
    page,
  }) => {
    await page.goto('/acceso')
    await fillCredentials(page, EMAIL, PASSWORD)

    // Simula un fallo de red real interceptando exactamente la primera
    // solicitud de login; la segunda (el reintento) llega al API real. La
    // IP sintética va en este mismo handler (no en `withIsolatedIP`,
    // registrado antes): Playwright invoca el handler más reciente
    // primero, así que uno externo nunca se alcanzaría aquí.
    const ip = syntheticIP()
    let attempt = 0
    await page.route('**/api/v1/public/auth/login', async (route) => {
      attempt += 1
      if (attempt === 1) {
        await route.abort('connectionfailed')
        return
      }
      await route.continue({ headers: { ...route.request().headers(), 'x-forwarded-for': ip } })
    })

    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    await expect(page.getByRole('button', { name: 'Reintentar' })).toBeVisible()
    // Los datos siguen ahí sin recargar la página.
    await expect(page.getByLabel('Correo', { exact: true })).toHaveValue(EMAIL)
    await expect(page.getByLabel('Contraseña', { exact: true })).toHaveValue(PASSWORD)

    await page.getByRole('button', { name: 'Reintentar' }).click()
    await expect(page).toHaveURL(/\/panel$/)
  })

  test('el botón se deshabilita durante el envío y un doble toque no produce dos solicitudes (CA-010-04)', async ({
    page,
  }) => {
    await page.goto('/acceso')
    await fillCredentials(page, EMAIL, PASSWORD)

    const ip = syntheticIP()
    let requestCount = 0
    await page.route('**/api/v1/public/auth/login', async (route) => {
      requestCount += 1
      // Retrasa la respuesta real lo suficiente para poder observar el
      // estado "enviando" y disparar un segundo toque mientras está en
      // curso, sin sustituir el recorrido real por un mock de resultado.
      await new Promise((resolve) => setTimeout(resolve, 300))
      await route.continue({ headers: { ...route.request().headers(), 'x-forwarded-for': ip } })
    })

    // Selector estable (no por nombre accesible, que cambia a "Iniciando
    // sesión…" apenas se hace clic): el mismo botón se usa para el primer
    // toque y el segundo.
    const submitButton = page.locator('form button[type="submit"]')
    await submitButton.click()
    await expect(submitButton).toBeDisabled()
    await expect(submitButton).toHaveText('Iniciando sesión…')
    // Segundo toque mientras la solicitud sigue en curso: el atributo
    // `disabled` nativo del botón ya bloquea el clic del navegador
    // (defensa real, no solo de prueba); `force` demuestra que, incluso
    // saltándose esa barrera del navegador, la guardia de LoginPage
    // (estado `submitting`) sigue evitando una segunda solicitud.
    await submitButton.click({ force: true, timeout: 2000 }).catch(() => {})

    await expect(page).toHaveURL(/\/panel$/)
    expect(requestCount).toBe(1)
  })

  test('opera completamente con teclado, con foco visible y orden coherente (CA-010-05)', async ({
    page,
  }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')
    await page.getByLabel('Correo', { exact: true }).focus()
    await expect(page.getByLabel('Correo', { exact: true })).toBeFocused()
    await page.keyboard.type(EMAIL)
    await page.keyboard.press('Tab')
    await expect(page.getByLabel('Contraseña', { exact: true })).toBeFocused()
    await page.keyboard.type(PASSWORD)
    await page.keyboard.press('Tab')
    // El control de mostrar/ocultar contraseña (mockup 02-acceso-recuperacion.png)
    // es el siguiente elemento enfocable real antes del envío.
    await expect(page.getByRole('button', { name: 'Mostrar contraseña' })).toBeFocused()
    await page.keyboard.press('Tab')
    await expect(page.getByRole('button', { name: 'Iniciar sesión' })).toBeFocused()
    await page.keyboard.press('Enter')

    await expect(page).toHaveURL(/\/panel$/)
  })

  test('el enlace de recuperación es visible, semántico y navegable (CA-010-08)', async ({
    page,
  }) => {
    await page.goto('/acceso')
    const link = page.getByRole('link', { name: '¿Olvidaste tu contraseña?' })
    await expect(link).toBeVisible()
    await expect(link).toHaveAttribute('href', '/recuperar-acceso')
    await link.click()
    await expect(page).toHaveURL(/\/recuperar-acceso$/)
    // No es una ruta rota: la respuesta HTTP de navegación es real.
    await expect(page.getByRole('heading', { name: 'Solicita tu código' })).toBeVisible()
  })

  test('sin sesión, /panel redirige al acceso conservando el destino (DEC-056, generalizado por HU-012/CA-012-02)', async ({
    page,
  }) => {
    await page.goto('/panel')
    await expect(page).toHaveURL(/\/acceso\?redirect=%2Fpanel$|\/acceso\?redirect=\/panel$/)
  })

  test('un 429 documentado explica el bloqueo sin recargar la página (trabajo requerido §8)', async ({
    page,
  }) => {
    await page.goto('/acceso')
    await fillCredentials(page, EMAIL, PASSWORD)
    await page.route('**/api/v1/public/auth/login', async (route) => {
      await route.fulfill({
        status: 429,
        contentType: 'application/problem+json',
        headers: { 'Retry-After': '120' },
        body: JSON.stringify({
          type: '/api/v1/problems/rate-limited',
          title: 'Demasiadas solicitudes',
          status: 429,
          code: 'rate-limited',
          instance: 'e2e-429',
          requestId: 'e2e-429',
        }),
      })
    })

    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    await expect(page.getByText('Demasiados intentos')).toBeVisible()
    await expect(page).toHaveURL(/\/acceso$/)
  })
})
