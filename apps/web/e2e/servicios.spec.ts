import { test, expect, type Page } from '@playwright/test'

/**
 * Recorrido E2E de HU-022 (catálogo básico de servicios). Corre contra el
 * API real en local (`go run ./cmd/api` sobre PostgreSQL real con las once
 * migraciones aplicadas y un usuario con un hash argon2id REAL), mismo
 * criterio que `e2e/barberos.spec.ts`.
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
  await page.getByLabel('Correo').fill(email)
  await page.getByLabel('Contraseña').fill(password)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel$/)
}

async function openServicios(page: Page) {
  await page.getByRole('link', { name: 'Servicios' }).click()
  await expect(page).toHaveURL(/\/panel\/servicios$/)
  await expect(page.getByRole('heading', { name: 'Servicios' })).toBeVisible()
}

async function createServiceViaUI(
  page: Page,
  name: string,
  { duration = '30', price = '45000.00' }: { duration?: string; price?: string } = {},
) {
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
  await dialog.getByLabel('Nombre').fill(name)
  await dialog.getByLabel('Duración (minutos)').fill(duration)
  await dialog.getByLabel('Precio (COP)').fill(price)
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()
}

test.describe('Catálogo básico de servicios (HU-022)', () => {
  // Serial a propósito: varias pruebas comparten la MISMA barbería real
  // (shopA/shopB, sin aprovisionamiento de barberías nuevas en esta
  // historia); ejecutarlas en paralelo produciría una carrera entre sus
  // propias escrituras, no un defecto del producto.
  test.describe.configure({ mode: 'serial' })

  test('crear, listar, editar y recargar conserva el servicio con sus nuevos valores (CA-022-01, CA-022-02, CA-022-04, CA-022-05)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    const stamp = Date.now()
    const originalName = `Corte E2E ${stamp}`
    await createServiceViaUI(page, originalName, { duration: '30', price: '45000.00' })
    await expect(page.getByText(originalName)).toBeVisible()

    // Esta barbería real acumula servicios de corridas anteriores (sin
    // DELETE disponible para limpiar entre corridas, mismo criterio que
    // e2e/barberos.spec.ts): cada comprobación de texto se ancla a la FILA
    // del servicio recién creado (identificada por su nombre único con
    // timestamp), nunca a un `getByText` global que podría resolver varias
    // filas con la misma duración o precio.
    const row = page.locator('li', { hasText: originalName })
    await expect(row.getByText('30 min')).toBeVisible()

    await row.getByRole('button', { name: `Editar ${originalName}` }).click()
    const editDialog = page.getByRole('dialog', { name: 'Editar servicio' })
    await expect(editDialog.getByLabel('Nombre')).toHaveValue(originalName)
    await editDialog.getByLabel('Duración (minutos)').fill('45')
    await editDialog.getByLabel('Precio (COP)').fill('50000.00')
    await editDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(editDialog).toBeHidden()

    await expect(row.getByText('45 min')).toBeVisible()
    await expect(row.getByText(/50000\.00 COP/)).toBeVisible()

    // Persistencia real, verificada contra el servidor tras recargar (mismo
    // criterio que e2e/barberos.spec.ts: esta barbería real acumula datos
    // de corridas anteriores, sin DELETE disponible para limpiar; se
    // confirma con una lectura directa al recurso por id).
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Servicios' })).toBeVisible()

    const persisted = await page.evaluate(async (name) => {
      const listRes = await fetch('/api/v1/private/services?limit=50', { credentials: 'include' })
      const listBody = (await listRes.json()) as { items: { id: string; name: string }[] }
      const found = listBody.items.find((s) => s.name === name)
      if (!found) return null
      const getRes = await fetch(`/api/v1/private/services/${found.id}`, { credentials: 'include' })
      return (await getRes.json()) as { durationMinutes: number; price: string }
    }, originalName)

    expect(persisted?.durationMinutes).toBe(45)
    expect(persisted?.price).toBe('50000.00')
  })

  test('un precio de cero o un nombre vacío se rechazan sin persistir, y el diálogo conserva la interacción (CA-022-03, CA-022-04)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    await page.getByRole('button', { name: 'Agregar servicio' }).click()
    const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog.getByText('Escribe el nombre del servicio.')).toBeVisible()
    await expect(dialog).toBeVisible()

    const stamp = Date.now()
    await dialog.getByLabel('Nombre').fill(`Precio Cero E2E ${stamp}`)
    await dialog.getByLabel('Duración (minutos)').fill('30')
    await dialog.getByLabel('Precio (COP)').fill('0')
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog.getByText('El precio debe ser mayor que cero.')).toBeVisible()
    await expect(dialog).toBeVisible()
    await expect(dialog.getByLabel('Nombre')).toHaveValue(`Precio Cero E2E ${stamp}`)
  })

  test('duraciones 25, 30, 45 y 90 minutos se aceptan sin catálogo cerrado (CA-022-03)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    const stamp = Date.now()
    for (const duration of ['25', '30', '45', '90']) {
      const name = `Duración ${duration} E2E ${stamp}`
      await createServiceViaUI(page, name, { duration })
      await expect(page.getByText(name)).toBeVisible()
    }
  })

  test('un nombre ya usado por otro servicio activo responde con conflicto sin crear una segunda fila (CA-022-04, DEC-067)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    const stamp = Date.now()
    const name = `Nombre Duplicado E2E ${stamp}`
    await createServiceViaUI(page, name)
    await expect(page.getByText(name)).toBeVisible()

    await page.getByRole('button', { name: 'Agregar servicio' }).click()
    const dialog = page.getByRole('dialog', { name: 'Agregar servicio' })
    await dialog.getByLabel('Nombre').fill(name)
    await dialog.getByLabel('Duración (minutos)').fill('45')
    await dialog.getByLabel('Precio (COP)').fill('1.00')
    await dialog.getByRole('button', { name: 'Guardar' }).click()

    await expect(dialog.getByText('Ese nombre ya está en uso')).toBeVisible()
    await expect(dialog).toBeVisible()

    // Solo una fila con ese nombre exacto (la primera).
    const matches = await page.evaluate(async (targetName) => {
      const res = await fetch('/api/v1/private/services?limit=50', { credentials: 'include' })
      const body = (await res.json()) as { items: { name: string }[] }
      return body.items.filter((s) => s.name === targetName).length
    }, name)
    expect(matches).toBe(1)
  })

  test('un identificador real de otra barbería responde 404 al consultarlo o editarlo, y nunca aparece en el listado propio (CA-022-06)', async ({
    page,
  }) => {
    await login(page, EMAIL_B, PASSWORD_B)
    await openServicios(page)

    const stamp = Date.now()
    const nameB = `Solo De B E2E ${stamp}`
    await createServiceViaUI(page, nameB)

    const serviceBId = await page.evaluate(async (name) => {
      const res = await fetch('/api/v1/private/services?limit=50', { credentials: 'include' })
      const body = (await res.json()) as { items: { id: string; name: string }[] }
      const found = body.items.find((s) => s.name === name)
      return found?.id ?? null
    }, nameB)
    expect(serviceBId).toBeTruthy()

    await page.goto('/acceso')
    await login(page, EMAIL, PASSWORD)
    await openServicios(page)

    const crossTenantResult = await page.evaluate(async (id) => {
      const getRes = await fetch(`/api/v1/private/services/${id}`, { credentials: 'include' })
      const patchRes = await fetch(`/api/v1/private/services/${id}`, {
        method: 'PATCH',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'Secuestrado Desde A' }),
      })
      return { getStatus: getRes.status, patchStatus: patchRes.status }
    }, serviceBId)

    expect(crossTenantResult.getStatus).toBe(404)
    expect(crossTenantResult.patchStatus).toBe(404)

    await expect(page.getByText(nameB)).toHaveCount(0)
  })
})
