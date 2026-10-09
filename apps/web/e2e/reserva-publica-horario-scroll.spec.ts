import { test, expect } from '@playwright/test'

/**
 * Paso «Horario» con muchas horas el mismo día: el scroll debe vivir solo en
 * la lista de horas. Cabecera, tira de días y barra de acción se quedan a la
 * vista y el documento no se desplaza. Mismo criterio de intercepción que
 * reserva-publica-horario.spec.ts: un único endpoint público simulado.
 */
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'

// 09:00–17:45 America/Bogota cada 15 min (36 franjas) en un único día.
const manySlots = Array.from({ length: 36 }, (_, index) => {
  const minutes = 14 * 60 + index * 15
  const hh = String(Math.floor(minutes / 60)).padStart(2, '0')
  const mm = String(minutes % 60).padStart(2, '0')
  return { startsAt: `2026-09-15T${hh}:${mm}:00Z` }
})

const viewports = [
  { name: 'escritorio alto', width: 1280, height: 900 },
  { name: 'portátil', width: 1280, height: 800 },
  { name: 'móvil', width: 390, height: 844 },
]

for (const viewport of viewports) {
  test(`con muchas horas solo scrollea la lista de horas (${viewport.name})`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.route(
      `**/api/v1/public/barbershops/barberia-ejemplo/services/${SERVICE_ID}/barbers/${BARBER_ID}/availability**`,
      async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            slots: manySlots,
            durationMinutes: 30,
            timezone: 'America/Bogota',
            slotGridMinutes: 15,
          }),
        })
      },
    )

    await page.goto(
      `/reservar/barberia-ejemplo/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`,
    )

    const options = page.getByRole('radio')
    await expect(options).toHaveCount(manySlots.length)
    const list = page.getByRole('radiogroup', { name: 'Horas disponibles' })

    const metrics = () =>
      page.evaluate(() => {
        const root = document.documentElement
        const slots = document.querySelector('.pb-slots') as HTMLElement
        return {
          documentOverflow: root.scrollHeight - root.clientHeight,
          listOverflow: slots.scrollHeight - slots.clientHeight,
        }
      })

    // Sin elegir nada: el documento no se desplaza y la lista sí.
    const before = await metrics()
    expect(before.documentOverflow).toBeLessThanOrEqual(1)
    expect(before.listOverflow).toBeGreaterThan(0)

    // La última hora es alcanzable desplazando solo la lista y se puede elegir.
    await options.last().scrollIntoViewIfNeeded()
    await options.last().click()
    await expect(options.last()).toHaveAttribute('aria-checked', 'true')

    // Con una franja elegida aparece la barra de acción: sigue sin haber
    // scroll de documento y la barra queda dentro de la ventana.
    const cta = page.getByRole('link', { name: /continuar/i })
    await expect(cta).toBeVisible()
    const after = await metrics()
    expect(after.documentOverflow).toBeLessThanOrEqual(1)
    const box = await cta.boundingBox()
    expect(box).not.toBeNull()
    expect(box!.y + box!.height).toBeLessThanOrEqual(viewport.height)

    // El scroll se queda en la lista: la rueda sobre las horas no mueve el documento.
    await list.hover()
    await page.mouse.wheel(0, 400)
    expect(await page.evaluate(() => window.scrollY)).toBe(0)
  })
}
