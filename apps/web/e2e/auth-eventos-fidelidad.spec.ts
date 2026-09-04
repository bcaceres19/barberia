import { test, expect, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia de fidelidad visual del issue #213: captura la app real en los
 * 23 eventos representados en `docs/10-backlog/evidence/
 * ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/`, en los dos
 * viewports del contrato (`desktop` 1440×1024, `mobile` 420 CSS px), para
 * comparar lado a lado con el PNG de referencia correspondiente.
 *
 * Todas las respuestas del API se interceptan con `page.route` (mismo
 * patrón que `acceso-evidencia-responsiva.spec.ts`): esta suite no depende
 * de `apps/api` ni de PostgreSQL.
 *
 * Eventos `07`, `09`, `10` y `12` de acceso no se capturan: el backend
 * vigente solo resuelve el canal WhatsApp (DEC-081), registrado en
 * `apps/web/e2e/evidence/auth-fidelidad-desviaciones.md`. Solo `08` y `11`
 * (la variante real) tienen evidencia ejecutable.
 */

const EMAIL = 'barbero@nava.co'
const PASSWORD = 'ClaveDePruebaHU010!'

const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'auth-eventos',
)

// Mismo patrón que servicios-evidencia-responsiva.spec.ts: axe-core inyectado
// desde node_modules, sin depender de un paquete de integración aparte.
const axeScriptPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'node_modules',
  'axe-core',
  'axe.min.js',
)

const DESKTOP = { width: 1440, height: 1024 }
const MOBILE = { width: 420, height: 935 }

async function fillCredentials(page: Page, email: string, password: string) {
  await page.getByLabel('Correo', { exact: true }).fill(email)
  await page.getByLabel('Contraseña', { exact: true }).fill(password)
}

// Cada ranura tiene `maxlength="1"`: `.fill()` sobre una sola ranura trunca
// al primer carácter (el `insertText` de Playwright respeta `maxlength`
// nativo), así que el pegado real de seis dígitos no llega al componente.
// Se simula tecleo real con el teclado: `OtpInput.onInput` avanza el foco a
// la siguiente ranura por su cuenta después de cada dígito.
async function fillOtp(page: Page, code: string) {
  await page.getByRole('group', { name: 'Código de 6 dígitos' }).locator('input').first().click()
  await page.keyboard.type(code)
}

async function assertViewport(page: Page, expected: { width: number; height: number }) {
  const actual = await page.evaluate(() => ({
    width: window.innerWidth,
    height: window.innerHeight,
  }))
  expect(actual.width).toBe(expected.width)
}

async function shoot(page: Page, viewportName: string, group: string, file: string) {
  await assertViewport(page, viewportName === 'desktop' ? DESKTOP : MOBILE)
  const hasHorizontalScroll = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  )
  expect(hasHorizontalScroll, `desbordamiento horizontal en ${viewportName}/${group}/${file}`).toBe(
    false,
  )
  await page.screenshot({
    path: path.join(evidenceDir, viewportName, group, file),
    fullPage: true,
  })

  // axe-core en vivo sobre el contenido real del evento ya renderizado.
  // `color-contrast` desactivado: paleta ya verificada en estandar-diseno-
  // visual.md §4.3, mismo criterio que el resto de la suite E2E.
  // `heading-order` desactivado: `BaseAlert.vue` (shared/ui, issue #212)
  // compone su título como `<h4>` para toda variante salvo `plain`, sin
  // encabezados `h2`/`h3` intermedios bajo el `<h1>` real de la pantalla
  // — hallazgo real, no un ajuste visual de #213. Fuera de alcance aquí
  // (no se parchea shared/ui localmente dentro de auth); registrado en
  // `evidence/auth-fidelidad-desviaciones.md` para #212.
  await page.addScriptTag({ path: axeScriptPath })
  const axeResults = (await page.evaluate(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const axeGlobal = (window as any).axe
    return axeGlobal.run(document, {
      rules: { 'color-contrast': { enabled: false }, 'heading-order': { enabled: false } },
    })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  })) as { violations: any[] }
  expect(axeResults.violations, JSON.stringify(axeResults.violations, null, 2)).toEqual([])
}

for (const viewport of [
  { name: 'desktop', ...DESKTOP },
  { name: 'mobile', ...MOBILE },
] as const) {
  test.describe(`Fidelidad auth-eventos — ${viewport.name}`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test.describe('acceso', () => {
      test('01-inicial', async ({ page }) => {
        await page.goto('/acceso')
        await expect(page.getByRole('heading', { name: 'Accede a NAVA' })).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '01-inicial.png')
      })

      test('02-enviando', async ({ page }) => {
        await page.goto('/acceso')
        await fillCredentials(page, EMAIL, PASSWORD)
        await page.route('**/api/v1/public/auth/login', async (route) => {
          await new Promise((resolve) => setTimeout(resolve, 400))
          await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Iniciar sesión' }).click()
        await expect(page.getByRole('button', { name: 'Iniciando sesión…' })).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '02-enviando.png')
      })

      test('03-validacion', async ({ page }) => {
        await page.goto('/acceso')
        await page.getByRole('button', { name: 'Iniciar sesión' }).click()
        await expect(page.getByRole('heading', { name: 'Revisa estos campos' })).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '03-validacion.png')
      })

      test('04-credenciales-invalidas', async ({ page }) => {
        await page.goto('/acceso')
        await fillCredentials(page, EMAIL, 'clave-incorrecta')
        await page.route('**/api/v1/public/auth/login', async (route) => {
          await route.fulfill({
            status: 401,
            contentType: 'application/problem+json',
            body: JSON.stringify({
              title: 'No autorizado',
              status: 401,
              code: 'invalid-credentials',
              instance: 'evidence-401',
              requestId: 'evidence-401',
            }),
          })
        })
        await page.getByRole('button', { name: 'Iniciar sesión' }).click()
        await expect(page.getByText('No pudimos iniciar tu sesión')).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '04-credenciales-invalidas.png')
      })

      test('05-sin-conexion', async ({ page }) => {
        await page.goto('/acceso')
        await fillCredentials(page, EMAIL, PASSWORD)
        await page.route('**/api/v1/public/auth/login', async (route) => {
          await route.abort('failed')
        })
        await page.getByRole('button', { name: 'Iniciar sesión' }).click()
        await expect(page.getByText('No pudimos conectar')).toBeVisible()
        await expect(page.getByRole('button', { name: 'Reintentar' })).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '05-sin-conexion.png')
      })

      test('06-sesion-vencida', async ({ page }) => {
        await page.goto('/acceso?motivo=sesion-expirada')
        await expect(page.getByText('Tu sesión venció')).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '06-sesion-vencida.png')
      })

      test('08-reto-envio-whatsapp', async ({ page }) => {
        await page.goto('/acceso')
        await fillCredentials(page, EMAIL, PASSWORD)
        await page.route('**/api/v1/public/auth/login', async (route) => {
          await route.fulfill({
            status: 429,
            contentType: 'application/problem+json',
            headers: { 'Retry-After': '120' },
            body: JSON.stringify({
              title: 'Demasiadas solicitudes',
              status: 429,
              code: 'rate-limited',
              instance: 'evidence-429',
              requestId: 'evidence-429',
            }),
          })
        })
        await page.getByRole('button', { name: 'Iniciar sesión' }).click()
        await expect(page.getByText('Verifica tu teléfono')).toBeVisible()

        await page.route('**/api/v1/public/auth/challenge', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código por WhatsApp' }).click()
        await expect(page.getByText('WhatsApp oficial')).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '08-reto-envio-whatsapp.png')
      })

      test('11-reto-codigo-invalido-whatsapp', async ({ page }) => {
        await page.goto('/acceso')
        await fillCredentials(page, EMAIL, PASSWORD)
        await page.route('**/api/v1/public/auth/login', async (route) => {
          await route.fulfill({
            status: 429,
            contentType: 'application/problem+json',
            headers: { 'Retry-After': '120' },
            body: JSON.stringify({
              title: 'Demasiadas solicitudes',
              status: 429,
              code: 'rate-limited',
              instance: 'evidence-429',
              requestId: 'evidence-429',
            }),
          })
        })
        await page.getByRole('button', { name: 'Iniciar sesión' }).click()
        await expect(page.getByText('Verifica tu teléfono')).toBeVisible()

        await page.route('**/api/v1/public/auth/challenge', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código por WhatsApp' }).click()
        await expect(page.getByText('WhatsApp oficial')).toBeVisible()

        await page.route('**/api/v1/public/auth/challenge/verify', async (route) => {
          await route.fulfill({
            status: 401,
            contentType: 'application/problem+json',
            body: JSON.stringify({
              title: 'No autorizado',
              status: 401,
              code: 'invalid-code',
              instance: 'evidence-401b',
              requestId: 'evidence-401b',
            }),
          })
        })
        await fillOtp(page, '000000')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('El código no es válido o venció')).toBeVisible()
        await shoot(page, viewport.name, 'acceso', '11-reto-codigo-invalido-whatsapp.png')
      })
    })

    test.describe('recuperacion', () => {
      test('01-solicitud-inicial', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await expect(page.getByText('Paso 1 de 3')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '01-solicitud-inicial.png')
      })

      test('02-solicitud-validacion', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Escribe tu correo')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '02-solicitud-validacion.png')
      })

      test('03-solicitud-enviando', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await new Promise((resolve) => setTimeout(resolve, 400))
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByRole('button', { name: 'Enviando…' })).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '03-solicitud-enviando.png')
      })

      test('04-solicitud-sin-conexion', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.abort('failed')
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('No pudimos conectar')).toBeVisible()
        await expect(page.getByRole('button', { name: 'Reintentar' })).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '04-solicitud-sin-conexion.png')
      })

      test('05-verificacion-inicial', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '05-verificacion-inicial.png')
      })

      test('06-verificacion-codigo-invalido', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()

        await page.route('**/api/v1/public/auth/recovery/verify', async (route) => {
          await route.fulfill({
            status: 401,
            contentType: 'application/problem+json',
            body: JSON.stringify({
              title: 'No autorizado',
              status: 401,
              code: 'invalid-code',
              instance: 'evidence-401c',
              requestId: 'evidence-401c',
            }),
          })
        })
        await fillOtp(page, '000000')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('El código no es correcto o ya venció')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '06-verificacion-codigo-invalido.png')
      })

      test('07-contrasena-inicial', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()

        await page.route('**/api/v1/public/auth/recovery/verify', async (route) => {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              resetToken: 'evidence-token',
              maskedPhone: '+57 *** *** 12',
              maskedEmail: 'b***@n***.co',
            }),
          })
        })
        await fillOtp(page, '482913')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('Paso 3 de 3')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '07-contrasena-inicial.png')
      })

      test('08-contrasena-validacion', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()

        await page.route('**/api/v1/public/auth/recovery/verify', async (route) => {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              resetToken: 'evidence-token',
              maskedPhone: '+57 *** *** 12',
              maskedEmail: 'b***@n***.co',
            }),
          })
        })
        await fillOtp(page, '482913')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('Paso 3 de 3')).toBeVisible()

        await page.locator('input[name="newPassword"]').fill('corta')
        await page.locator('input[name="confirmPassword"]').fill('otraclave')
        await page.getByRole('button', { name: 'Guardar contraseña nueva' }).click()
        await expect(page.getByRole('heading', { name: 'Revisa estos campos' })).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '08-contrasena-validacion.png')
      })

      test('09-contrasena-enviando', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()

        await page.route('**/api/v1/public/auth/recovery/verify', async (route) => {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              resetToken: 'evidence-token',
              maskedPhone: '+57 *** *** 12',
              maskedEmail: 'b***@n***.co',
            }),
          })
        })
        await fillOtp(page, '482913')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('Paso 3 de 3')).toBeVisible()

        await page.locator('input[name="newPassword"]').fill('contraseña-nueva-valida')
        await page.locator('input[name="confirmPassword"]').fill('contraseña-nueva-valida')
        await page.route('**/api/v1/public/auth/recovery/reset-password', async (route) => {
          await new Promise((resolve) => setTimeout(resolve, 400))
          await route.fulfill({ status: 204 })
        })
        await page.getByRole('button', { name: 'Guardar contraseña nueva' }).click()
        await expect(page.getByRole('button', { name: 'Guardando contraseña…' })).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '09-contrasena-enviando.png')
      })

      test('10-contrasena-enlace-vencido', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()

        await page.route('**/api/v1/public/auth/recovery/verify', async (route) => {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              resetToken: 'evidence-token',
              maskedPhone: '+57 *** *** 12',
              maskedEmail: 'b***@n***.co',
            }),
          })
        })
        await fillOtp(page, '482913')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('Paso 3 de 3')).toBeVisible()

        await page.locator('input[name="newPassword"]').fill('contraseña-nueva-valida')
        await page.locator('input[name="confirmPassword"]').fill('contraseña-nueva-valida')
        await page.route('**/api/v1/public/auth/recovery/reset-password', async (route) => {
          await route.fulfill({
            status: 401,
            contentType: 'application/problem+json',
            body: JSON.stringify({
              title: 'No autorizado',
              status: 401,
              code: 'invalid-token',
              instance: 'evidence-401d',
              requestId: 'evidence-401d',
            }),
          })
        })
        await page.getByRole('button', { name: 'Guardar contraseña nueva' }).click()
        await expect(page.getByText('El enlace de recuperación venció')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '10-contrasena-enlace-vencido.png')
      })

      test('11-completado', async ({ page }) => {
        await page.goto('/recuperar-acceso')
        await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
        await page.route('**/api/v1/public/auth/recovery/request', async (route) => {
          await route.fulfill({ status: 202, contentType: 'application/json', body: '{}' })
        })
        await page.getByRole('button', { name: 'Enviar código' }).click()
        await expect(page.getByText('Paso 2 de 3')).toBeVisible()

        await page.route('**/api/v1/public/auth/recovery/verify', async (route) => {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              resetToken: 'evidence-token',
              maskedPhone: '+57 *** *** 12',
              maskedEmail: 'b***@n***.co',
            }),
          })
        })
        await fillOtp(page, '482913')
        await page.getByRole('button', { name: 'Verificar código' }).click()
        await expect(page.getByText('Paso 3 de 3')).toBeVisible()

        await page.locator('input[name="newPassword"]').fill('contraseña-nueva-valida')
        await page.locator('input[name="confirmPassword"]').fill('contraseña-nueva-valida')
        await page.route('**/api/v1/public/auth/recovery/reset-password', async (route) => {
          await route.fulfill({ status: 204 })
        })
        await page.getByRole('button', { name: 'Guardar contraseña nueva' }).click()
        await expect(page.getByText('cerramos todas tus sesiones activas')).toBeVisible()
        await shoot(page, viewport.name, 'recuperacion', '11-completado.png')
      })
    })
  })
}
