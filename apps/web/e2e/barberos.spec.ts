import { test, expect, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Recorrido E2E de HU-021 (registro y listado de barberos). Corre contra
 * el API real en local (`go run ./cmd/api` sobre PostgreSQL real con las
 * diez migraciones aplicadas y un usuario con un hash argon2id REAL, mismo
 * criterio que `e2e/configuracion-barberia.spec.ts`).
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
  await expect(page).toHaveURL(/\/panel(\?.*)?$/)
}

async function openBarberos(page: Page) {
  await page.getByRole('link', { name: 'Barberos' }).click()
  await expect(page).toHaveURL(/\/panel\/barberos$/)
  await expect(page.getByRole('heading', { name: 'Barberos' })).toBeVisible()
}

test.describe('Registro y listado de barberos (HU-021)', () => {
  // Serial a propósito: varias pruebas comparten la MISMA barbería real
  // (shopA/shopB, sin aprovisionamiento de barberías nuevas en esta
  // historia); ejecutarlas en paralelo produciría una carrera entre sus
  // propias escrituras, no un defecto del producto.
  test.describe.configure({ mode: 'serial' })

  test('agregar un barbero y luego tres más produce cuatro recursos nuevos por el mismo camino de UI (CA-021-01, CA-021-02)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openBarberos(page)

    const stamp = Date.now()
    const names = [
      `Barbero E2E Uno ${stamp}`,
      `Barbero E2E Dos ${stamp}`,
      `Barbero E2E Tres ${stamp}`,
      `Barbero E2E Cuatro ${stamp}`,
    ]

    // El primer alta demuestra el camino "un barbero" (CA-021-01): el
    // mismo formulario, el mismo botón, la misma lista.
    for (const name of names) {
      await page.getByRole('button', { name: 'Agregar barbero' }).click()
      const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
      await dialog.getByLabel('Nombre').fill(name)
      await dialog.getByRole('button', { name: 'Guardar' }).click()
      await expect(dialog).toBeHidden()
      await expect(page.getByText(name)).toBeVisible()
    }

    // Las cuatro filas nuevas coexisten en la lista, cada una como un
    // recurso distinto (CA-021-02): ningún alta posterior reemplazó o
    // fusionó una anterior.
    for (const name of names) {
      await expect(page.getByText(name)).toBeVisible()
    }
  })

  test('renombrar persiste tras recargar, sin duplicar el recurso (CA-021-04)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openBarberos(page)

    const stamp = Date.now()
    const originalName = `Antes De Renombrar ${stamp}`
    const newName = `Después De Renombrar ${stamp}`

    await page.getByRole('button', { name: 'Agregar barbero' }).click()
    const createDialog = page.getByRole('dialog', { name: 'Agregar barbero' })
    await createDialog.getByLabel('Nombre').fill(originalName)
    await createDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(createDialog).toBeHidden()
    await expect(page.getByText(originalName)).toBeVisible()

    const row = page.locator('li', { hasText: originalName })
    await row.getByRole('button', { name: `Editar ${originalName}` }).click()
    const editDialog = page.getByRole('dialog', { name: 'Editar barbero' })
    await expect(editDialog.getByLabel('Nombre')).toHaveValue(originalName)
    await editDialog.getByLabel('Nombre').fill(newName)
    await editDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(editDialog).toBeHidden()

    await expect(page.getByText(newName)).toBeVisible()
    await expect(page.getByText(originalName, { exact: true })).toHaveCount(0)

    // Persistencia real, verificada contra el servidor: esta barbería real
    // ya acumuló muchos barberos de corridas E2E anteriores (sin DELETE
    // disponible para limpiar entre corridas), así que el recién renombrado
    // -el más reciente por created_at- puede caer fuera de la primera
    // página por defecto (orden ascendente, CA-021-02) tras recargar; en
    // una base efímera de CI esto no ocurriría. Se confirma la persistencia
    // real con una lectura directa al recurso por id, no con la vista de
    // la primera página.
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Barberos' })).toBeVisible()

    const renamed = await page.evaluate(async (name) => {
      const listRes = await fetch('/api/v1/private/barbers?limit=50', { credentials: 'include' })
      const listBody = (await listRes.json()) as { items: { id: string; fullName: string }[] }
      const found = listBody.items.find((b) => b.fullName === name)
      if (!found) return null
      const getRes = await fetch(`/api/v1/private/barbers/${found.id}`, { credentials: 'include' })
      return (await getRes.json()) as { id: string; fullName: string }
    }, newName)

    expect(renamed?.fullName).toBe(newName)
    await expect(page.getByText(originalName, { exact: true })).toHaveCount(0)
  })

  test('la fotografía se sube, se ve en la lista, persiste al recargar y se quita (DEC-104)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openBarberos(page)

    const name = `Retrato E2E ${Date.now()}`
    await page.getByRole('button', { name: 'Agregar barbero' }).click()
    const createDialog = page.getByRole('dialog', { name: 'Agregar barbero' })
    await createDialog.getByLabel('Nombre').fill(name)
    // Una imagen sintética (sin personas reales) que el navegador recorta y
    // reduce antes de enviarla.
    await createDialog
      .locator('[data-testid="barber-photo-input"]')
      .setInputFiles(
        path.join(
          path.dirname(fileURLToPath(import.meta.url)),
          'fixtures',
          'retrato-sintetico.jpg',
        ),
      )
    await expect(createDialog.getByText('Foto lista.')).toBeVisible()
    await createDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(createDialog).toBeHidden()

    const row = page.locator('li', { hasText: name })
    // La imagen se sirve de verdad desde el API (no un icono de imagen rota).
    const photo = row.locator('img.barber-avatar__photo')
    await expect(photo).toBeVisible()
    await expect
      .poll(() => photo.evaluate((img: HTMLImageElement) => img.complete && img.naturalWidth))
      .toBeGreaterThan(0)

    // El servidor guarda la fotografía y la sirve con su tipo real. (Sin
    // recargar: esta barbería acumula barberos de corridas anteriores y el
    // recién creado puede quedar fuera de la primera página, ver renombrar.)
    const served = async (fullName: string) =>
      page.evaluate(async (target) => {
        const res = await fetch('/api/v1/private/barbers?limit=50', { credentials: 'include' })
        const body = (await res.json()) as {
          items: { id: string; fullName: string; photoUpdatedAt: string | null }[]
        }
        const barber = body.items.find((b) => b.fullName === target)
        if (!barber) return null
        const photoRes = await fetch(`/api/v1/private/barbers/${barber.id}/photo`, {
          credentials: 'include',
        })
        return {
          hasVersion: barber.photoUpdatedAt !== null,
          status: photoRes.status,
          type: photoRes.headers.get('content-type'),
        }
      }, fullName)
    expect(await served(name)).toEqual({ hasVersion: true, status: 200, type: 'image/jpeg' })

    // Quitarla devuelve el monograma.
    const editButton = page
      .locator('li', { hasText: name })
      .getByRole('button', { name: `Editar ${name}` })
    await editButton.click()
    const editDialog = page.getByRole('dialog', { name: 'Editar barbero' })
    await editDialog.getByRole('button', { name: 'Quitar foto' }).click()
    await editDialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(editDialog).toBeHidden()
    await expect(page.locator('li', { hasText: name }).locator('img')).toHaveCount(0)
    expect(await served(name)).toEqual({
      hasVersion: false,
      status: 404,
      type: 'application/problem+json',
    })
  })

  test('un nombre vacío se rechaza sin persistir, y el diálogo conserva la interacción (CA-021-03)', async ({
    page,
  }) => {
    await login(page, EMAIL, PASSWORD)
    await openBarberos(page)

    await page.getByRole('button', { name: 'Agregar barbero' }).click()
    const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
    await dialog.getByRole('button', { name: 'Guardar' }).click()

    await expect(dialog.getByText('Escribe el nombre del barbero.')).toBeVisible()
    await expect(dialog).toBeVisible()
  })

  test('un identificador real de otra barbería responde 404 al consultarlo o editarlo, y nunca aparece en el listado propio (CA-021-05)', async ({
    page,
  }) => {
    // Crea un barbero real en la barbería B para tener un id real que
    // consultar desde el contexto de A.
    await login(page, EMAIL_B, PASSWORD_B)
    await openBarberos(page)

    const stamp = Date.now()
    const nameB = `Solo De B ${stamp}`
    await page.getByRole('button', { name: 'Agregar barbero' }).click()
    const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
    await dialog.getByLabel('Nombre').fill(nameB)
    await dialog.getByRole('button', { name: 'Guardar' }).click()
    await expect(dialog).toBeHidden()

    const barberBId = await page.evaluate(async (name) => {
      const res = await fetch('/api/v1/private/barbers?limit=50', { credentials: 'include' })
      const body = (await res.json()) as { items: { id: string; fullName: string }[] }
      const found = body.items.find((b) => b.fullName === name)
      return found?.id ?? null
    }, nameB)
    expect(barberBId).toBeTruthy()

    // Cambia de sesión a la barbería A (misma pestaña: cierra sesión no
    // existe en este flujo, así que se navega directamente a /acceso e
    // inicia sesión de nuevo; la cookie anterior se reemplaza).
    await page.goto('/acceso')
    await login(page, EMAIL, PASSWORD)
    await openBarberos(page)

    const crossTenantResult = await page.evaluate(async (id) => {
      const getRes = await fetch(`/api/v1/private/barbers/${id}`, { credentials: 'include' })
      const patchRes = await fetch(`/api/v1/private/barbers/${id}`, {
        method: 'PATCH',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ fullName: 'Secuestrado Desde A' }),
      })
      return { getStatus: getRes.status, patchStatus: patchRes.status }
    }, barberBId)

    expect(crossTenantResult.getStatus).toBe(404)
    expect(crossTenantResult.patchStatus).toBe(404)

    // El barbero de B nunca aparece en el listado de A.
    await expect(page.getByText(nameB)).toHaveCount(0)
  })
})
