import { test, expect } from '@playwright/test'

/**
 * Paso «Tus datos»: al elegir «Para otra persona» el formulario gana un campo
 * y crece. El documento no debe desplazarse; si el contenido no cabe, solo se
 * desplaza el interior de la hoja (formulario, resumen o confirmación).
 * Mismo patrón que reserva-publica-cliente.spec.ts: la página no llama a
 * ningún endpoint hasta confirmar, así que basta navegar a la ruta.
 */
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
const STARTS_AT = encodeURIComponent('2026-09-20T14:30:00Z')
const ROUTE = `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario/${STARTS_AT}/cliente`

const viewports = [
  { name: 'escritorio alto', width: 1280, height: 900 },
  { name: 'portátil', width: 1280, height: 800 },
  { name: 'móvil', width: 390, height: 844 },
]

for (const viewport of viewports) {
  test(`«Para otra persona» no genera scroll global (${viewport.name})`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto(ROUTE)

    const documentOverflow = () =>
      page.evaluate(
        () => document.documentElement.scrollHeight - document.documentElement.clientHeight,
      )

    await expect(page.getByLabel('Tu nombre')).toBeVisible()
    expect(await documentOverflow()).toBeLessThanOrEqual(1)

    await page.getByLabel('Para otra persona').check()
    await expect(page.getByLabel('Nombre de la persona atendida')).toBeVisible()
    // Espera a que termine la entrada escalonada de los campos.
    await page.waitForTimeout(1200)
    expect(await documentOverflow()).toBeLessThanOrEqual(1)

    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByLabel('Nombre de la persona atendida').fill('Luis Ríos')
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')

    // «Ver resumen» conserva su alto (no se comprime al caber justo) y queda
    // fijo y visible al pie de la hoja sin tener que desplazar.
    const pinned = page.getByRole('button', { name: 'Ver resumen' })
    await expect(pinned).toBeInViewport({ ratio: 1 })
    expect((await pinned.boundingBox())!.height).toBeGreaterThanOrEqual(40)

    // «Ver resumen» es alcanzable desplazando solo la hoja.
    const submit = page.getByRole('button', { name: 'Ver resumen' })
    await submit.scrollIntoViewIfNeeded()
    expect(await documentOverflow()).toBeLessThanOrEqual(1)
    const box = await submit.boundingBox()
    expect(box).not.toBeNull()
    expect(box!.y + box!.height).toBeLessThanOrEqual(viewport.height)
    await submit.click()

    // El resumen conserva la misma técnica: sin scroll global.
    await expect(page.getByRole('heading', { name: 'Revisa tus datos' })).toBeVisible()
    await page.waitForTimeout(800)
    expect(await documentOverflow()).toBeLessThanOrEqual(1)
  })
}
