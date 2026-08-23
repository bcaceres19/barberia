import { test, expect, type Page } from '@playwright/test'
import { readFile, rm } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import path from 'node:path'

/**
 * Recorrido E2E de HU-011 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * recuperación de acceso de punta a punta contra el backend real de HU-008
 * (integrado en `main`, PR #63) y la pantalla real de esta historia.
 *
 * Requiere, además de lo que ya exige `e2e/acceso.spec.ts` (`apps/api`
 * corriendo contra PostgreSQL real con las siete migraciones aplicadas y
 * un usuario con hash argon2id real):
 *
 *   - APP_AUTH_HMAC_SECRET (al menos 32 caracteres).
 *   - APP_RECOVERY_CAPTURE_FILE=<ruta> — el remitente real es `DEC-066`
 *     (Meta WhatsApp Cloud API + Resend); en su lugar,
 *     `auth.CapturingRecoveryCodeSender` escribe {phone, email, code} en
 *     esta ruta para que la prueba lea el código real sin depender de un
 *     proveedor real ("terceros se interceptan"). Restringido por
 *     `cmd/api/main.go` a APP_ENVIRONMENT=local/test.
 *   - Para la prueba de código vencido: APP_RECOVERY_CODE_EXPIRES_SECONDS=2
 *     (por defecto son 900 s; un valor bajo evita esperar 15 minutos
 *     reales). El resto de la suite tolera cualquier valor por encima de
 *     los pocos segundos que tarda cada prueba.
 *
 * Comando reproducible (PowerShell, dos terminales):
 *   $env:APP_ENVIRONMENT="local"; $env:APP_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/postgres?sslmode=disable";
 *   $env:APP_AUTH_HMAC_SECRET="una-clave-de-prueba-local-de-al-menos-32-caracteres";
 *   $env:APP_RECOVERY_CAPTURE_FILE="$env:TEMP\hu011-e2e-capture.json";
 *   $env:APP_RECOVERY_CODE_EXPIRES_SECONDS="2";
 *   go run ./cmd/api   # (desde apps/api)
 *   pnpm run test:e2e -- recuperacion   # (desde apps/web)
 *
 * Cuentas: `dueno.b@ejemplo.test` (código válido) y `duena.a@ejemplo.test`
 * (código vencido) — una por prueba. `acceso.spec.ts`/`reto-telefonico.spec.ts`
 * no llaman ningún endpoint de recuperación, así que no compiten por su
 * cooldown; separar las dos pruebas de ESTE archivo sí importa: el
 * cooldown de reenvío de `DEC-064` es de 60 s por cuenta, y ambas pruebas
 * corren en modo serial una detrás de otra sobre el mismo PostgreSQL.
 * Usar la misma cuenta en las dos haría que la segunda solicitud cayera
 * dentro del cooldown de la primera y nunca generara un código nuevo.
 */
const VALID_CODE_EMAIL = process.env.E2E_RECOVERY_EMAIL ?? 'dueno.b@ejemplo.test'
const EXPIRED_CODE_EMAIL = process.env.E2E_RECOVERY_EXPIRED_EMAIL ?? 'duena.a@ejemplo.test'
const NEW_PASSWORD = 'ClaveDeRecuperacionHU011!'
const CAPTURE_FILE =
  process.env.APP_RECOVERY_CAPTURE_FILE ?? path.join(tmpdir(), 'hu011-e2e-capture.json')

async function readCapturedCode(): Promise<string> {
  const raw = await readFile(CAPTURE_FILE, 'utf-8')
  const parsed = JSON.parse(raw) as { phone: string; email: string; code: string }
  return parsed.code
}

async function waitForCapturedCode(): Promise<string> {
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
  return readCapturedCode()
}

async function requestRecovery(page: Page, email: string) {
  await page.goto('/recuperar-acceso')
  await page.getByLabel('Correo').fill(email)
  await page.getByRole('button', { name: 'Enviar código' }).click()
  await expect(page.getByText('Paso 2 de 3')).toBeVisible()
}

test.describe('Recuperación de acceso (HU-011)', () => {
  // Serial: ambas pruebas solicitan un código real sobre la MISMA cuenta y
  // leen el mismo archivo de captura compartido por proceso (mismo motivo
  // que reto-telefonico.spec.ts).
  test.describe.configure({ mode: 'serial' })

  test.beforeEach(async () => {
    await rm(CAPTURE_FILE, { force: true }).catch(() => {})
  })

  test('recorrido completo con código válido: solicitar, verificar y establecer contraseña (CA-011-01, CA-011-02, CA-011-05)', async ({
    page,
  }) => {
    await requestRecovery(page, VALID_CODE_EMAIL)

    const code = await waitForCapturedCode()
    await page.getByLabel('Código de 6 dígitos').fill(code)
    await page.getByRole('button', { name: 'Verificar código' }).click()

    await expect(page.getByText('Paso 3 de 3')).toBeVisible()
    // Destino enmascarado, nunca completo (CA-011-02/CA-008-06).
    await expect(page.getByText(VALID_CODE_EMAIL)).toHaveCount(0)
    await expect(page.getByText(/\*{2,}/)).toBeVisible()

    await page.locator('input[name="newPassword"]').fill(NEW_PASSWORD)
    await page.locator('input[name="confirmPassword"]').fill(NEW_PASSWORD)
    await page.getByRole('button', { name: 'Guardar contraseña nueva' }).click()

    await expect(page.getByText('cerramos todas tus sesiones activas')).toBeVisible()

    // Confirma la contraseña nueva contra el login real (CA-008-05: la
    // anterior deja de autenticar, la nueva sí).
    await page.getByRole('button', { name: 'Ir al acceso' }).click()
    await expect(page).toHaveURL(/\/acceso$/)
    await page.getByLabel('Correo').fill(VALID_CODE_EMAIL)
    await page.getByLabel('Contraseña').fill(NEW_PASSWORD)
    await page.getByRole('button', { name: 'Iniciar sesión' }).click()
    await expect(page).toHaveURL(/\/panel$/)
  })

  test('un código vencido produce el mismo mensaje uniforme que uno incorrecto, sin revelar el motivo (CA-011-03, CT-007)', async ({
    page,
  }) => {
    await requestRecovery(page, EXPIRED_CODE_EMAIL)
    await waitForCapturedCode()

    // Espera a que el código expire. Requiere
    // APP_RECOVERY_CODE_EXPIRES_SECONDS bajo (ver cabecera de este archivo);
    // con el valor por defecto de producción (900 s) esta prueba no es
    // practicable en una suite E2E y debe omitirse.
    await page.waitForTimeout(3_000)
    const expiredCode = await readCapturedCode()

    await page.getByLabel('Código de 6 dígitos').fill(expiredCode)
    await page.getByRole('button', { name: 'Verificar código' }).click()

    await expect(page.getByText('El código no es válido')).toBeVisible()
    await expect(page.getByText('Paso 2 de 3')).toBeVisible()
  })
})
