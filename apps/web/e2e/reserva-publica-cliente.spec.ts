import { test, expect } from '@playwright/test'

/**
 * Recorrido E2E de HU-096 (docs/03-desarrollo/estrategia-pruebas.md §5.3):
 * datos del cliente y persona atendida. A diferencia de las demás rutas
 * públicas, esta página no llama a ningún endpoint (HU-097, que persiste,
 * está bloqueada por DP-PUB-05/DP-PUB-06/CT-011): no hace falta
 * `page.route(...)`, solo navegar directamente a la ruta con el contexto
 * ya elegido en la URL (mismo patrón que las demás rutas del asistente).
 */
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
const STARTS_AT = encodeURIComponent('2026-09-20T14:30:00Z')
const ROUTE = `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario/${STARTS_AT}/cliente`

test.describe('Datos del cliente y persona atendida (HU-096)', () => {
  test('exige nombre, teléfono y correo, y conserva lo ya escrito ante un error (CA-096-02, CA-096-03)', async ({
    page,
  }) => {
    await page.goto(ROUTE)

    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByRole('button', { name: 'Ver resumen' }).click()

    await expect(page.getByText('El teléfono es obligatorio.')).toBeVisible()
    await expect(page.getByText('El correo es obligatorio.')).toBeVisible()
    await expect(page.getByLabel('Tu nombre')).toHaveValue('Ana Ríos')
  })

  test('reserva para sí mismo resuelve la persona atendida sin pedirla dos veces (CA-096-01)', async ({
    page,
  }) => {
    await page.goto(ROUTE)

    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')
    await expect(page.getByLabel('Nombre de la persona atendida')).toHaveCount(0)

    await page.getByRole('button', { name: 'Ver resumen' }).click()

    await expect(page.getByRole('status')).toContainText('Cliente')
    await expect(page.getByRole('status')).toContainText('Ana Ríos')
  })

  test('reserva para otra persona exige un nombre adicional no vacío (CA-096-01)', async ({
    page,
  }) => {
    await page.goto(ROUTE)

    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')
    await page.getByLabel('Para otra persona').check()
    await page.getByRole('button', { name: 'Ver resumen' }).click()

    await expect(page.getByText('El nombre de la persona atendida es obligatorio.')).toBeVisible()

    await page.getByLabel('Nombre de la persona atendida').fill('Mateo Ruiz')
    await page.getByRole('button', { name: 'Ver resumen' }).click()

    const summary = page.getByRole('status')
    await expect(summary).toContainText('Ana Ríos')
    await expect(summary).toContainText('Mateo Ruiz')
  })

  test('permite volver a editar sin perder los datos ya escritos', async ({ page }) => {
    await page.goto(ROUTE)

    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')
    await page.getByRole('button', { name: 'Ver resumen' }).click()

    await page.getByRole('button', { name: 'Editar' }).click()

    await expect(page.getByLabel('Tu nombre')).toHaveValue('Ana Ríos')
    await expect(page.getByLabel('Teléfono')).toHaveValue('+573001234567')
    await expect(page.getByLabel('Correo')).toHaveValue('ana@example.com')
  })
})
