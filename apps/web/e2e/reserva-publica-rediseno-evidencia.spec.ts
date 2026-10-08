import { test, expect, type Page, type Route } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { INITIAL_VOCABULARY } from './fixtures/vocabularioInicial'

/**
 * Evidencia del rediseño de la reserva pública (DEC-111): el recorrido
 * completo (entrada → servicio → barbero → horario → datos → confirmación)
 * sobre el lienzo de tinta de NAVA, en 320/360/768/1280/1440 px y a 200 % de
 * zoom de texto. Modo de conformidad: identidad guiada (sin mockup exacto
 * asignado, DEC-078). Cada captura espera a que terminen las entradas
 * animadas; la verificación accesible (axe-core, ahora CON `color-contrast`:
 * la reserva ya no usa la paleta clara que el estándar daba por medida) se
 * corre con movimiento reducido para medir el estado final y no un
 * fotograma intermedio. Las capturas quedan en `e2e/evidence/` para
 * adjuntarse al PR.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'reserva-publica', 'rediseno')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

const SLUG = 'barberia-ejemplo'
const SERVICE_ID = '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4'
const SERVICE_2_ID = '1a2b3c4d-5e6f-4708-9a0b-1c2d3e4f5061'
const BARBER_ID = 'a1111111-1111-1111-1111-111111111111'
const BARBER_2_ID = 'a2222222-2222-2222-2222-222222222222'
const STARTS_AT = '2026-09-15T19:00:00Z'

const json = (route: Route, body: unknown, status = 200) =>
  route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) })

interface MockOptions {
  barbers?: Array<{ id: string; fullName: string }>
  services?: unknown[]
  confirm?: 'ok' | 'conflict'
}

/** Intercepta los cinco endpoints públicos: ningún dato real, ningún
 * servidor Go detrás (mismo criterio que el resto de specs `reserva-publica-*`). */
async function mockApi(page: Page, options: MockOptions = {}) {
  const base = `**/api/v1/public/barbershops/${SLUG}`
  await page.route(base, (route) =>
    json(route, {
      name: 'Barbería Ejemplo',
      timezone: 'America/Bogota',
      contactEmail: 'contacto@ejemplo.test',
      contactPhone: '+573001234567',
      vocabulary: INITIAL_VOCABULARY,
    }),
  )
  await page.route(`${base}/services**`, (route) =>
    json(route, {
      items: options.services ?? [
        {
          id: SERVICE_ID,
          name: 'Corte clásico',
          description: 'Corte con máquina y tijera, incluye lavado.',
          durationMinutes: 30,
          price: '45000.00',
          currency: 'COP',
        },
        {
          id: SERVICE_2_ID,
          name: 'Barba y toalla caliente',
          description: null,
          durationMinutes: 20,
          price: '20000.00',
          currency: 'COP',
        },
        {
          id: '3b3b3b3b-3b3b-4b3b-8b3b-3b3b3b3b3b3b',
          name: 'Corte + barba con perfilado y tratamiento facial de cortesía',
          description:
            'Servicio completo con asesoría de imagen, lavado, masaje capilar y acabado con productos de la casa.',
          durationMinutes: 60,
          price: '75000.00',
          currency: 'COP',
        },
      ],
      nextCursor: null,
    }),
  )
  await page.route(`${base}/services/*/barbers**`, (route) =>
    json(route, {
      items: options.barbers ?? [
        { id: BARBER_ID, fullName: 'Julián Rodríguez' },
        { id: BARBER_2_ID, fullName: 'Luis Martínez' },
        { id: 'a3333333-3333-3333-3333-333333333333', fullName: 'Santiago Gómez' },
      ],
    }),
  )
  await page.route(`${base}/services/*/barbers/*/availability**`, (route) => {
    const slots: Array<{ startsAt: string }> = []
    for (const day of ['15', '16', '17', '18', '19', '21', '22']) {
      for (const hour of [14, 15, 16, 17, 19, 20, 21, 22]) {
        slots.push({ startsAt: `2026-09-${day}T${hour}:00:00Z` })
        if (hour % 2 === 0) slots.push({ startsAt: `2026-09-${day}T${hour}:30:00Z` })
      }
    }
    return json(route, {
      slots,
      durationMinutes: 30,
      timezone: 'America/Bogota',
      slotGridMinutes: 30,
    })
  })
  await page.route(`${base}/services/*/barbers/*/appointments**`, (route) => {
    if (options.confirm === 'conflict') {
      return json(
        route,
        {
          type: '/api/v1/problems/slot-conflict',
          title: 'Conflicto',
          status: 409,
          code: 'slot-conflict',
          requestId: 'e2e-rediseno',
          alternatives: [
            { startsAt: '2026-09-15T20:00:00Z' },
            { startsAt: '2026-09-16T14:00:00Z' },
          ],
        },
        409,
      )
    }
    return json(
      route,
      {
        attendeeName: 'Ana Ríos',
        barbershopName: 'Barbería Ejemplo',
        serviceName: 'Corte clásico',
        durationMinutes: 30,
        priceAmount: '45000.00',
        currency: 'COP',
        startsAt: STARTS_AT,
        endsAt: '2026-09-15T19:30:00Z',
        timezone: 'America/Bogota',
        accessToken: 'token-de-prueba',
        customerNote: null,
      },
      201,
    )
  })
}

async function settle(page: Page, ms = 1900) {
  await page.waitForTimeout(ms)
}

async function snap(page: Page, viewport: string, name: string) {
  await settle(page)
  await page.screenshot({ path: path.join(evidenceDir, viewport, `${name}.png`), fullPage: true })
}

/** Captura de la ventana (no de la página completa) con el final del
 * contenido a la vista: la barra de acción es `sticky` al pie de la
 * ventana, así que una captura de página completa la dejaría a media altura
 * y no mostraría lo que la persona ve de verdad. */
async function snapBottom(page: Page, viewport: string, name: string) {
  await settle(page)
  await page.evaluate(() =>
    window.scrollTo({ top: document.body.scrollHeight, behavior: 'instant' }),
  )
  await settle(page, 900)
  await page.screenshot({ path: path.join(evidenceDir, viewport, `${name}.png`) })
}

async function assertNoHorizontalScroll(page: Page) {
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
  )
  expect(overflow).toBe(false)
}

/** axe-core con `color-contrast` activo. Se corre con movimiento reducido
 * para que ninguna entrada animada deje un fotograma a medio opacar. */
async function assertAccessible(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  const results = (await page.evaluate(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return (window as any).axe.run(document)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  })) as { violations: any[] }
  expect(results.violations, JSON.stringify(results.violations, null, 2)).toEqual([])
}

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1440', width: 1440, height: 900 },
  // Aproximación de zoom de texto 200 % sobre 1280: mitad de ancho.
  { name: '1280-zoom200', width: 640, height: 450 },
]

const horarioRoute = `/reservar/${SLUG}/servicios/${SERVICE_ID}/barbero/${BARBER_ID}/horario`
const clienteRoute = `${horarioRoute}/${encodeURIComponent(STARTS_AT)}/cliente`

for (const viewport of viewports) {
  test.describe(`Rediseño reserva pública ${viewport.name}px`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } })

    test('recorrido: entrada, servicio, barbero y horario', async ({ page }) => {
      await mockApi(page)

      await page.goto(`/reservar/${SLUG}`)
      await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '01-entrada')

      await page.getByRole('link', { name: 'Reservar un turno' }).click()
      await expect(page.getByRole('heading', { name: 'Elige tu servicio' })).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '02-servicios')

      await page.getByRole('radio', { name: /Corte clásico/ }).click()
      await expect(page.getByRole('link', { name: 'Continuar' })).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snapBottom(page, viewport.name, '03-servicio-elegido')

      await page.getByRole('link', { name: 'Continuar' }).click()
      await expect(page.getByRole('heading', { name: 'Elige tu barbero' })).toBeVisible()
      await snap(page, viewport.name, '04-barberos')

      await page.getByRole('radio', { name: /Luis Martínez/ }).click()
      await snapBottom(page, viewport.name, '05-barbero-elegido')

      await page.getByRole('link', { name: 'Continuar' }).click()
      await expect(page.getByRole('heading', { name: 'Elige fecha y hora' })).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '06-horario')

      await page.getByRole('radio').nth(2).click()
      await expect(page.getByRole('link', { name: 'Continuar' })).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snapBottom(page, viewport.name, '07-horario-elegido')
    })

    test('datos, resumen, conflicto y confirmación', async ({ page }) => {
      await mockApi(page)
      await page.goto(clienteRoute)
      await expect(page.getByLabel('Tu nombre')).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '08-datos')

      await page.getByLabel('Tu nombre').fill('Ana Ríos')
      await page.getByRole('button', { name: 'Ver resumen' }).click()
      await expect(page.getByText('El teléfono es obligatorio.')).toBeVisible()
      await snap(page, viewport.name, '09-datos-errores')

      await page.getByLabel('Teléfono').fill('+573001234567')
      await page.getByLabel('Correo').fill('ana@example.com')
      await page.getByLabel('Para otra persona').check()
      await page.getByLabel('Nombre de la persona atendida').fill('Mateo Ruiz')
      await page.getByRole('button', { name: 'Ver resumen' }).click()
      await expect(page.getByRole('button', { name: 'Confirmar turno' })).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '10-resumen')

      await page.getByRole('button', { name: 'Confirmar turno' }).click()
      await expect(page.getByText('¡Tu turno quedó confirmado!')).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '11-confirmado')
    })

    test('estados: carga, enlace no encontrado y vacío', async ({ page }) => {
      await page.route(`**/api/v1/public/barbershops/${SLUG}/services**`, async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 1500))
        await json(route, { items: [], nextCursor: null })
      })
      await page.goto(`/reservar/${SLUG}/servicios`)
      await expect(page.getByText('Cargando servicios…')).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, '12-carga.png'),
        fullPage: true,
      })
      await expect(page.getByText('todavía no tiene servicios disponibles')).toBeVisible()
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '13-vacio')

      await page.route('**/api/v1/public/barbershops/enlace-inexistente', (route) =>
        json(
          route,
          {
            type: '/api/v1/problems/not-found',
            title: 'No encontrado',
            status: 404,
            code: 'not-found',
          },
          404,
        ),
      )
      await page.goto('/reservar/enlace-inexistente')
      await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
      await assertNoHorizontalScroll(page)
      await snap(page, viewport.name, '14-no-encontrado')
    })
  })
}

test.describe('Rediseño reserva pública: accesibilidad con contraste (movimiento reducido)', () => {
  test.use({
    viewport: { width: 1280, height: 900 },
    contextOptions: { reducedMotion: 'reduce' },
  })

  test('cada pantalla y estado del recorrido pasa axe-core, color-contrast incluido', async ({
    page,
  }) => {
    await mockApi(page)

    await page.goto(`/reservar/${SLUG}`)
    await expect(page.getByRole('heading', { name: 'Barbería Ejemplo' })).toBeVisible()
    await assertAccessible(page)

    await page.getByRole('link', { name: 'Reservar un turno' }).click()
    await expect(page.getByRole('heading', { name: 'Elige tu servicio' })).toBeVisible()
    await assertAccessible(page)
    await page.getByRole('radio', { name: /Corte clásico/ }).click()
    await assertAccessible(page)

    await page.getByRole('link', { name: 'Continuar' }).click()
    await expect(page.getByRole('heading', { name: 'Elige tu barbero' })).toBeVisible()
    await assertAccessible(page)
    await page.getByRole('radio', { name: /Luis Martínez/ }).click()
    await assertAccessible(page)

    await page.getByRole('link', { name: 'Continuar' }).click()
    await expect(page.getByRole('heading', { name: 'Elige fecha y hora' })).toBeVisible()
    await assertAccessible(page)
    await page.getByRole('radio').nth(2).click()
    await assertAccessible(page)

    await page.getByRole('link', { name: 'Continuar' }).click()
    await expect(page.getByLabel('Tu nombre')).toBeVisible()
    await assertAccessible(page)
    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByRole('button', { name: 'Ver resumen' }).click()
    await assertAccessible(page)
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')
    await page.getByRole('button', { name: 'Ver resumen' }).click()
    await assertAccessible(page)
    await page.getByRole('button', { name: 'Confirmar turno' }).click()
    await expect(page.getByText('¡Tu turno quedó confirmado!')).toBeVisible()
    await assertAccessible(page)
  })

  test('estados de página: no encontrado, sin conexión y error inesperado', async ({ page }) => {
    await page.route('**/api/v1/public/barbershops/enlace-inexistente', (route) =>
      json(
        route,
        {
          type: '/api/v1/problems/not-found',
          title: 'No encontrado',
          status: 404,
          code: 'not-found',
        },
        404,
      ),
    )
    await page.goto('/reservar/enlace-inexistente')
    await expect(page.getByRole('alert')).toContainText('No encontramos ese enlace')
    await assertAccessible(page)

    await page.route('**/api/v1/public/barbershops/sin-red', (route) => route.abort())
    await page.goto('/reservar/sin-red')
    await expect(page.getByRole('alert')).toContainText('No pudimos conectar')
    await assertAccessible(page)

    await page.route('**/api/v1/public/barbershops/con-fallo', (route) =>
      json(
        route,
        {
          type: '/api/v1/problems/internal',
          title: 'Error interno',
          status: 500,
          detail: 'fallo inesperado',
          instance: 'req-e2e-1',
          code: 'internal',
          requestId: 'req-e2e-1',
        },
        500,
      ),
    )
    await page.goto('/reservar/con-fallo')
    await expect(page.getByRole('alert')).toContainText('Ocurrió un error inesperado')
    await expect(page.getByRole('alert')).toContainText('req-e2e-1')
    await assertAccessible(page)
  })

  test('conflicto de horario con alternativas y barbero único', async ({ page }) => {
    await mockApi(page, { confirm: 'conflict' })
    await page.goto(clienteRoute)
    await page.getByLabel('Tu nombre').fill('Ana Ríos')
    await page.getByLabel('Teléfono').fill('+573001234567')
    await page.getByLabel('Correo').fill('ana@example.com')
    await page.getByRole('button', { name: 'Ver resumen' }).click()
    await page.getByRole('button', { name: 'Confirmar turno' }).click()
    await expect(page.getByText('Ese horario se acaba de ocupar.')).toBeVisible()
    await assertAccessible(page)

    await mockApi(page, { barbers: [{ id: BARBER_ID, fullName: 'Julián Rodríguez' }] })
    await page.goto(`/reservar/${SLUG}/servicios/${SERVICE_ID}/barbero`)
    await expect(page.getByText('Te atenderá')).toBeVisible()
    await assertAccessible(page)
  })
})

test.describe('Rediseño reserva pública: teclado y movimiento', () => {
  test.use({ viewport: { width: 1280, height: 900 } })

  test('el foco de teclado es visible y las flechas mueven la selección', async ({ page }) => {
    await mockApi(page)
    await page.goto(`/reservar/${SLUG}/servicios`)
    await expect(page.getByRole('heading', { name: 'Elige tu servicio' })).toBeVisible()
    await page.getByRole('radio', { name: /Corte clásico/ }).focus()
    await page.keyboard.press('ArrowDown')
    await page.keyboard.press('Space')
    await expect(page.getByRole('radio', { name: /Barba y toalla/ })).toHaveAttribute(
      'aria-checked',
      'true',
    )
    await snap(page, '1280', '15-foco-teclado')
  })

  test('con movimiento reducido no queda ninguna animación corriendo en el contenido', async ({
    page,
  }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await mockApi(page)
    await page.goto(`/reservar/${SLUG}/servicios`)
    await expect(page.getByRole('heading', { name: 'Elige tu servicio' })).toBeVisible()
    await page.getByRole('radio', { name: /Corte clásico/ }).click()
    const running = await page.evaluate(() =>
      document
        .getAnimations()
        .filter((animation) => animation.playState === 'running')
        .map((animation) => (animation as CSSAnimation).animationName ?? 'transición'),
    )
    expect(running).toEqual([])
  })

  test('volver atrás conserva el recorrido y el progreso retrocede', async ({ page }) => {
    await mockApi(page)
    await page.goto(`/reservar/${SLUG}/servicios/${SERVICE_ID}/barbero`)
    await expect(page.getByRole('heading', { name: 'Elige tu barbero' })).toBeVisible()
    await expect(page.getByText('Paso 2 de 4')).toBeVisible()
    await page.getByRole('link', { name: /Servicios/ }).click()
    await expect(page.getByRole('heading', { name: 'Elige tu servicio' })).toBeVisible()
    await expect(page.getByText('Paso 1 de 4')).toBeVisible()
  })
})
