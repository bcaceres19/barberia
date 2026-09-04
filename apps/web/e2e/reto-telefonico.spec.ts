import { test, expect, type Page } from '@playwright/test'
import { randomBytes } from 'node:crypto'
import { readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'

/**
 * Recorrido E2E de HU-007 (DEC-061/DEC-062): cruzar el umbral de intentos
 * exige el reto telefónico antes de evaluar la contraseña, y verificarlo
 * con el código correcto desbloquea el acceso de inmediato, sin esperar el
 * escalamiento de 24 horas.
 *
 * Requiere, además de lo que ya exige e2e/acceso.spec.ts (`apps/api`
 * corriendo contra PostgreSQL real con las siete migraciones aplicadas y
 * un usuario con hash argon2id real):
 *
 *   - APP_AUTH_HMAC_SECRET, APP_LOGIN_THROTTLE_* / APP_PHONE_CHALLENGE_*
 *     (o sus valores por defecto: umbral 5, ventana 900 s).
 *   - APP_TRUSTED_PROXIES=127.0.0.1/32 — permite que este archivo aísle
 *     cada prueba con una IP sintética propia vía X-Forwarded-For (ver
 *     `withIsolatedIP` abajo), para no compartir el contador de
 *     login_throttle con el resto de la suite E2E (acceso.spec.ts,
 *     panel.spec.ts) ni entre pruebas de este mismo archivo corriendo en
 *     paralelo. Sin esta variable, el servidor ignora la cabecera por
 *     completo (comportamiento correcto y seguro por defecto) y estas
 *     pruebas comparten IP con el resto de la suite.
 *   - APP_PHONE_CHALLENGE_CAPTURE_FILE=<ruta> — el remitente real de
 *     WhatsApp es DEC-066 (decisión de HU-008, fuera de alcance de
 *     HU-007); en su lugar, `auth.CapturingPhoneCodeSender` escribe
 *     {phone, code} en esta ruta para que la prueba pueda leer el código
 *     real sin depender de un proveedor real ("terceros se interceptan").
 *     Restringido por main.go a APP_ENVIRONMENT=local/test.
 *
 * Comando reproducible (PowerShell, dos terminales):
 *   $env:APP_ENVIRONMENT="local"; $env:APP_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/postgres?sslmode=disable";
 *   $env:APP_AUTH_HMAC_SECRET="una-clave-de-prueba-local-de-al-menos-32-caracteres";
 *   $env:APP_TRUSTED_PROXIES="127.0.0.1/32";
 *   $env:APP_PHONE_CHALLENGE_CAPTURE_FILE="$env:TEMP\hu007-e2e-capture.json";
 *   go run ./cmd/api   # (desde apps/api)
 *   pnpm run test:e2e -- reto-telefonico   # (desde apps/web)
 */
// Mismo usuario y misma contraseña por defecto que acceso.spec.ts/
// panel.spec.ts (ClaveDePruebaHU010!): las tres suites comparten la misma
// fila real de staff_credential; usar una contraseña por defecto distinta
// aquí rompería silenciosamente las otras suites en cuanto alguien
// regenere el hash contra ESTE archivo en vez del original.
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'
const CAPTURE_FILE =
  process.env.APP_PHONE_CHALLENGE_CAPTURE_FILE ?? path.join(tmpdir(), 'hu007-e2e-capture.json')

function syntheticIP(): string {
  const octets = Array.from({ length: 3 }, () => randomBytes(1)[0])
  return `10.${octets[0]}.${octets[1]}.${octets[2]}`
}

/** Aísla las solicitudes de auth de esta página bajo una IP sintética
 * propia (ver comentario de cabecera): solo tiene efecto si el servidor
 * confía en 127.0.0.1 vía APP_TRUSTED_PROXIES. */
async function withIsolatedIP(page: Page): Promise<string> {
  const ip = syntheticIP()
  await page.route('**/api/v1/public/auth/**', async (route) => {
    await route.continue({ headers: { ...route.request().headers(), 'x-forwarded-for': ip } })
  })
  return ip
}

async function fillCredentials(page: Page, email: string, password: string) {
  await page.getByLabel('Correo', { exact: true }).fill(email)
  await page.getByLabel('Contraseña', { exact: true }).fill(password)
}

// `OtpInput` (issue #213) no expone un único control asociado a "Código de
// 6 dígitos" vía `<label>`: es un grupo de seis casillas, cada una con su
// propio `aria-label` ("Dígito N de 6") y `maxlength="1"` nativo.
// Verificado contra el navegador real (2026-09-04): `.fill()` sobre la
// primera casilla NO reproduce un pegado — el navegador trunca el valor a
// 1 carácter por el `maxlength` antes de que `onInput` lo lea. Se llena
// cada casilla con su propio dígito: mismo camino de producción
// (`onInput`, rama de un solo carácter) que un usuario tecleando dígito
// por dígito, mismo estado final que un pegado.
async function fillOtp(page: Page, code: string) {
  const inputs = page.getByRole('group', { name: 'Código de 6 dígitos' }).locator('input')
  for (let i = 0; i < code.length; i += 1) {
    await inputs.nth(i).fill(code[i])
  }
}

async function readCapturedCode(): Promise<string> {
  const raw = await readFile(CAPTURE_FILE, 'utf-8')
  const parsed = JSON.parse(raw) as { phone: string; code: string }
  return parsed.code
}

test.describe('Defensa escalonada contra abuso del acceso (HU-007)', () => {
  // Serial a propósito: más de una prueba de este archivo solicita el reto
  // telefónico, y el servidor de prueba escribe el código capturado en UN
  // solo archivo compartido (APP_PHONE_CHALLENGE_CAPTURE_FILE es una ruta
  // fija por proceso, no por solicitud). Ejecutarlas en paralelo (el modo
  // por defecto del proyecto) haría que una prueba leyera el código de la
  // solicitud de OTRA prueba. Cada prueba sigue aislada por su propia IP
  // sintética (`withIsolatedIP`); solo el orden de ejecución se serializa.
  test.describe.configure({ mode: 'serial' })

  test.beforeEach(async () => {
    await rm(CAPTURE_FILE, { force: true }).catch(() => {})
  })

  test('las primeras cinco solicitudes se evalúan con normalidad (CA-007-01)', async ({ page }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')

    for (let i = 0; i < 5; i += 1) {
      await fillCredentials(page, EMAIL, 'contraseña-incorrecta-a-proposito')
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByRole('alert').last()).toContainText('Revisa tu correo y contraseña')
    }
  })

  test('la sexta solicitud exige el reto telefónico y bloquea la contraseña hasta completarlo (CA-007-01, CA-007-02, CA-007-03)', async ({
    page,
  }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')

    // Cinco solicitudes normales primero, para llegar exactamente a la
    // sexta con esta IP aislada (DEC-061: la sexta escala, no la quinta).
    for (let i = 0; i < 5; i += 1) {
      await fillCredentials(page, EMAIL, 'contraseña-incorrecta-a-proposito')
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
      await expect(page.getByRole('alert').last()).toContainText('Revisa tu correo y contraseña')
    }

    // Sexta solicitud, incluso con la contraseña CORRECTA: debe bloquearse
    // por el reto, no evaluarse (CA-007-02 exige que la contraseña nunca
    // se evalúe en esta rama).
    await fillCredentials(page, EMAIL, PASSWORD)
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()

    await expect(page.getByText('Demasiados intentos')).toBeVisible()
    await expect(page.getByText('Verifica tu teléfono')).toBeVisible()
    // Sigue en /acceso: la contraseña correcta no bastó (CA-007-02).
    await expect(page).toHaveURL(/\/acceso$/)
  })

  test('completar el reto con el código correcto reintenta el login y llega a /panel (CA-007-02)', async ({
    page,
  }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')

    for (let i = 0; i < 6; i += 1) {
      await fillCredentials(page, EMAIL, i < 5 ? 'contraseña-incorrecta-a-proposito' : PASSWORD)
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    }
    await expect(page.getByText('Verifica tu teléfono')).toBeVisible()

    await page.getByRole('button', { name: 'Enviar código por WhatsApp' }).click()
    await expect(page.getByText('Si tu cuenta existe')).toBeVisible()

    // Terceros interceptados (cabecera de este archivo): el código real se
    // lee de la captura de prueba, nunca de un WhatsApp real.
    await expect
      .poll(
        async () => {
          try {
            return await readCapturedCode()
          } catch {
            return null
          }
        },
        { timeout: 10_000 },
      )
      .not.toBeNull()
    const code = await readCapturedCode()

    await fillOtp(page, code)
    await page.getByRole('button', { name: 'Verificar código' }).click()

    await expect(page).toHaveURL(/\/panel$/)
    await expect(page.getByRole('heading', { name: 'Agenda' })).toBeVisible()
  })

  test('un código incorrecto no desbloquea el acceso y no revela la causa (CA-007-03, no enumeración)', async ({
    page,
  }) => {
    await withIsolatedIP(page)
    await page.goto('/acceso')

    for (let i = 0; i < 6; i += 1) {
      await fillCredentials(page, EMAIL, 'contraseña-incorrecta-a-proposito')
      await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    }
    await expect(page.getByText('Verifica tu teléfono')).toBeVisible()

    await page.getByRole('button', { name: 'Enviar código por WhatsApp' }).click()
    await expect(page.getByText('Si tu cuenta existe')).toBeVisible()

    await fillOtp(page, '000000')
    await page.getByRole('button', { name: 'Verificar código' }).click()

    await expect(page.getByText('El código no es válido o venció')).toBeVisible()
    await expect(page).toHaveURL(/\/acceso$/)
  })
})
