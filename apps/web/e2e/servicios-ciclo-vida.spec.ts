import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-024 (ciclo de vida de servicios: desactivar y
 * reactivar). Corre contra el API real en local (`go run ./cmd/api` sobre
 * PostgreSQL real con las doce migraciones aplicadas y un usuario con un
 * hash argon2id REAL), mismo criterio que `e2e/servicios.spec.ts`.
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

async function openServicios(page: Page) {
  await page.getByRole('link', { name: 'Servicios' }).click()
  await expect(page).toHaveURL(/\/panel\/servicios$/)
  await expect(page.getByRole('heading', { name: 'Servicios' })).toBeVisible()
}

async function createServiceViaUI(page: Page, name: string) {
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
  await dialog.getByLabel('Nombre').fill(name)
  await dialog.getByLabel('Duración (minutos)').fill('30')
  await dialog.getByLabel('Precio (COP)').fill('45000.00')
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()
}

test.describe('Ciclo de vida de servicios (HU-024)', () => {
  // Serial a propósito: varias pruebas comparten la MISMA barbería real
  // (shopA/shopB), mismo criterio que servicios.spec.ts.
  test.describe.configure({ mode: 'serial' })

  test('previsualizar, desactivar, confirmar persistencia y reactivar (CA-024-01, CA-024-02, CA-024-04, CA-024-05)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    const stamp = Date.now()
    const name = `Corte Ciclo De Vida E2E ${stamp}`
    await createServiceViaUI(page, name)

    const row = page.locator('li', { hasText: name })
    await expect(row.getByText('Activo')).toBeVisible()

    // CA-024-01: la advertencia consulta el impacto real (siempre 0 en B1,
    // DEC-069), nunca un valor por defecto sin consultar.
    await row.getByRole('button', { name: `Desactivar ${name}` }).click()
    const deactivateDialog = page.getByRole('dialog', { name: 'Desactivar servicio' })
    await expect(deactivateDialog.getByText('No hay citas futuras')).toBeVisible()

    // Cancelar no muta nada.
    await deactivateDialog.getByRole('button', { name: 'Cancelar' }).click()
    await expect(deactivateDialog).toBeHidden()
    await expect(row.getByText('Activo')).toBeVisible()

    // Confirmar desactiva de verdad (CA-024-02).
    await row.getByRole('button', { name: `Desactivar ${name}` }).click()
    await expect(deactivateDialog.getByText('No hay citas futuras')).toBeVisible()
    await deactivateDialog.getByRole('button', { name: 'Desactivar', exact: true }).click()
    await expect(deactivateDialog).toBeHidden()
    await expect(row.getByText('Inactivo')).toBeVisible()

    // Persistencia real, verificada tras recargar (nunca se borra la fila,
    // RN-SER-03): sigue apareciendo en el listado, ahora inactiva.
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Servicios' })).toBeVisible()
    await expect(row.getByText('Inactivo')).toBeVisible()

    const persisted = await page.evaluate(async (targetName) => {
      const listRes = await fetch('/api/v1/private/services?limit=50', { credentials: 'include' })
      const listBody = (await listRes.json()) as {
        items: { id: string; name: string; isActive: boolean; deactivatedAt: string | null }[]
      }
      return listBody.items.find((s) => s.name === targetName) ?? null
    }, name)
    expect(persisted?.isActive).toBe(false)
    expect(persisted?.deactivatedAt).not.toBeNull()

    // CA-024-05: reactivar vuelve a activo sin crear otra fila ni cambiar
    // duración/precio.
    await row.getByRole('button', { name: `Reactivar ${name}` }).click()
    const reactivateDialog = page.getByRole('dialog', { name: 'Reactivar servicio' })
    await reactivateDialog.getByRole('button', { name: 'Reactivar', exact: true }).click()
    await expect(reactivateDialog).toBeHidden()
    await expect(row.getByText('Activo')).toBeVisible()
    await expect(row.getByText('30 min')).toBeVisible()

    const reactivated = await page.evaluate(async (targetName) => {
      const res = await fetch('/api/v1/private/services?limit=50', { credentials: 'include' })
      const body = (await res.json()) as {
        items: {
          id: string
          name: string
          isActive: boolean
          deactivatedAt: string | null
          durationMinutes: number
        }[]
      }
      return body.items.find((s) => s.name === targetName) ?? null
    }, name)
    expect(reactivated?.isActive).toBe(true)
    expect(reactivated?.deactivatedAt).toBeNull()
    expect(reactivated?.durationMinutes).toBe(30)
  })

  test('un identificador real de otra barbería responde 404 al previsualizar, desactivar o reactivar (CA-024-07)', async ({
    page,
  }) => {
    await login(page, EMAIL_B, PASSWORD_B)
    await openServicios(page)

    const stamp = Date.now()
    const nameB = `Solo De B Ciclo De Vida E2E ${stamp}`
    await createServiceViaUI(page, nameB)

    const serviceBId = await page.evaluate(async (name) => {
      const res = await fetch('/api/v1/private/services?limit=50', { credentials: 'include' })
      const body = (await res.json()) as { items: { id: string; name: string }[] }
      return body.items.find((s) => s.name === name)?.id ?? null
    }, nameB)
    expect(serviceBId).toBeTruthy()

    await page.goto('/acceso')
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    const crossTenantResult = await page.evaluate(async (id) => {
      const impactRes = await fetch(`/api/v1/private/services/${id}/deactivation-impact`, {
        credentials: 'include',
      })
      const deactivateRes = await fetch(`/api/v1/private/services/${id}/deactivate`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Idempotency-Key': `e2e-cross-${crypto.randomUUID()}` },
      })
      const reactivateRes = await fetch(`/api/v1/private/services/${id}/reactivate`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Idempotency-Key': `e2e-cross-${crypto.randomUUID()}` },
      })
      return {
        impactStatus: impactRes.status,
        deactivateStatus: deactivateRes.status,
        reactivateStatus: reactivateRes.status,
      }
    }, serviceBId)

    expect(crossTenantResult.impactStatus).toBe(404)
    expect(crossTenantResult.deactivateStatus).toBe(404)
    expect(crossTenantResult.reactivateStatus).toBe(404)
  })
})
