import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { openAsidePanel, pickDate, pickTime, selectBarber } from './horarios-controles'

/**
 * Evidencia visual responsive/accesible de la pantalla Horarios rediseñada
 * (tablero semanal de latón sobre tinta, mismo lenguaje que Barberos y
 * Servicios): 320/360/768/1280 px más zoom 200% aproximado, sin scroll
 * horizontal, foco visible, diálogos de tramo y de excepción, y axe-core.
 * Corre con respuestas simuladas del API (sin sesión ni datos reales): el
 * recorrido contra el API real vive en horarios.spec.ts y
 * excepciones-festivos.spec.ts. Las capturas van a e2e/evidence/horarios/.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'horarios',
  'rediseno',
)
const axeScriptPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'node_modules',
  'axe-core',
  'axe.min.js',
)

const workingHours = [
  workingHour('m-1', 1, '09:00', 240),
  workingHour('m-2', 1, '14:00', 240),
  workingHour('t-1', 2, '09:00', 480),
  workingHour('j-1', 4, '10:00', 240),
  workingHour('j-2', 4, '15:00', 180),
  workingHour('v-1', 5, '09:00', 480),
]

function workingHour(id: string, isoWeekday: number, startsTime: string, durationMinutes: number) {
  return {
    id,
    isoWeekday,
    startsTime,
    durationMinutes,
    createdAt: '2026-09-09T12:00:00Z',
    updatedAt: '2026-09-09T12:00:00Z',
  }
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function openSchedules(page: Page, width: number, height: number) {
  await page.setViewportSize({ width, height })
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({ barbershop: { id: 'nava', name: 'NAVA' }, expiresAt: '2099-01-01T00:00:00Z' }),
      )
      return
    }
    if (pathname.endsWith('/barbers')) {
      await route.fulfill(
        json({
          items: [
            { id: 'barber-nava', fullName: 'Barbero NAVA' },
            { id: 'barber-dos', fullName: 'Segundo Barbero' },
          ],
        }),
      )
      return
    }
    if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json({ timezone: 'America/Bogota' }))
      return
    }
    if (pathname.endsWith('/working-hours')) {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON() as {
          isoWeekday: number
          startsTime: string
          durationMinutes: number
        }
        await route.fulfill(
          json(workingHour('nuevo', body.isoWeekday, body.startsTime, body.durationMinutes), 201),
        )
        return
      }
      await route.fulfill(json({ items: workingHours, nextCursor: null }))
      return
    }
    if (pathname.endsWith('/holiday-calendar')) {
      await route.fulfill(json({ enabled: true }))
      return
    }
    if (pathname.endsWith('/schedule-exceptions')) {
      if (route.request().method() === 'POST') {
        const body = route.request().postDataJSON() as {
          effectiveDate: string
          isClosed: boolean
          reason: string | null
        }
        await route.fulfill(
          json(
            {
              id: 'exception-nueva',
              effectiveDate: body.effectiveDate,
              isClosed: body.isClosed,
              reason: body.reason,
              segments: [],
              createdAt: '2026-09-09T12:00:00Z',
              updatedAt: '2026-09-09T12:00:00Z',
            },
            201,
          ),
        )
        return
      }
      await route.fulfill(
        json({
          items: [
            {
              id: 'exception-1',
              effectiveDate: '2099-12-08',
              isClosed: true,
              reason: 'Capacitación interna',
              segments: [],
              createdAt: '2026-09-09T12:00:00Z',
              updatedAt: '2026-09-09T12:00:00Z',
            },
            {
              id: 'exception-2',
              effectiveDate: '2099-12-24',
              isClosed: false,
              reason: null,
              segments: [{ startsTime: '10:00', durationMinutes: 180 }],
              createdAt: '2026-09-09T12:00:00Z',
              updatedAt: '2026-09-09T12:00:00Z',
            },
          ],
          nextCursor: null,
        }),
      )
      return
    }
    if (pathname.endsWith('/schedule/colombian-holidays')) {
      await route.fulfill(
        json({
          items: [
            { date: '2099-12-08', name: 'Día de la Inmaculada Concepción' },
            { date: '2099-12-25', name: 'Navidad' },
          ],
        }),
      )
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
  await page.goto('/panel/horarios')
  await expect(page.getByRole('heading', { name: 'Horarios' })).toBeVisible()
  await expect(page.getByText('09:00 · 240 min').first()).toBeVisible()
}

async function expectNoAxeViolations(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  // color-contrast se desactiva: la paleta ya está verificada en la tabla
  // aprobada de estandar-diseno-visual.md §4.3 (mismo criterio que las demás
  // evidencias responsive).
  const results = (await page.evaluate(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const axeGlobal = (window as any).axe
    return axeGlobal.run(document, { rules: { 'color-contrast': { enabled: false } } })
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  })) as { violations: any[] }
  expect(results.violations, JSON.stringify(results.violations, null, 2)).toEqual([])
}

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

test('horarios: sin scroll horizontal, foco visible y sin violaciones axe en cada breakpoint', async ({
  page,
}) => {
  await openSchedules(page, 1280, 900)
  for (const viewport of viewports) {
    await page.setViewportSize(viewport)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    // Mide el estado terminado: las barras y los días entran escalonados.
    await page.waitForTimeout(1400)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'horarios.png'),
      fullPage: true,
    })
    await page.getByRole('button', { name: 'Agregar tramo' }).focus()
    await expect(page.getByRole('button', { name: 'Agregar tramo' })).toBeFocused()
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco-agregar.png') })
    await expectNoAxeViolations(page)
  }
})

test('horarios: el diálogo de tramo y el de excepción se leen en móvil y escritorio', async ({
  page,
}) => {
  await openSchedules(page, 1280, 900)
  for (const viewport of [viewports[0]!, viewports[3]!]) {
    await page.setViewportSize(viewport)

    await page.getByRole('button', { name: 'Agregar un tramo el Martes' }).click()
    const tramo = page.getByRole('dialog', { name: 'Agregar tramo' })
    await expect(tramo).toBeVisible()
    await expect(tramo.getByRole('radio', { name: 'Martes' })).toBeChecked()
    await page.waitForTimeout(700)
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'dialogo-tramo.png') })
    await expectNoAxeViolations(page)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await tramo.getByRole('button', { name: 'Cancelar' }).click()
    await expect(tramo).toBeHidden()

    await page.getByRole('button', { name: 'Agregar excepción', exact: true }).click()
    const excepcion = page.getByRole('dialog', { name: 'Agregar excepción' })
    await expect(excepcion).toBeVisible()
    await excepcion.getByLabel('Abierto con tramos especiales').check()
    await page.waitForTimeout(700)
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'dialogo-excepcion.png') })
    await expectNoAxeViolations(page)
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth),
    ).toBe(false)
    await excepcion.getByRole('button', { name: 'Cancelar' }).click()
    await expect(excepcion).toBeHidden()
  }
})

test('horarios: el selector de barbero se maneja con teclado y cambia el horario', async ({
  page,
}) => {
  await openSchedules(page, 1280, 900)
  const trigger = page.getByRole('button', { name: /Barbero/ }).first()
  await trigger.focus()
  await page.keyboard.press('ArrowDown')
  await expect(page.getByRole('listbox')).toBeVisible()
  await page.keyboard.press('ArrowDown')
  await page.keyboard.press('Enter')
  await expect(page.getByRole('listbox')).toBeHidden()
  await expect(trigger).toContainText('Segundo Barbero')
})

test('horarios: con movimiento reducido no hay animaciones de entrada', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openSchedules(page, 1280, 900)
  const animated = await page.evaluate(() =>
    document
      .getAnimations()
      .filter((a) => a.effect?.getComputedTiming().iterations === Infinity)
      .map((a) => (a as CSSAnimation).animationName ?? ''),
  )
  expect(animated).toEqual([])
})

test('horarios: el alta de un tramo y de una excepción funciona con los controles propios', async ({
  page,
}) => {
  await openSchedules(page, 1280, 900)

  await selectBarber(page, 'Segundo Barbero')

  await page.getByRole('button', { name: 'Agregar tramo' }).click()
  const tramo = page.getByRole('dialog', { name: 'Agregar tramo' })
  await tramo.getByRole('radio', { name: 'Miércoles' }).click()
  await pickTime(tramo, page, 'Hora de inicio', '11:30')
  await tramo.getByLabel('Duración (minutos)').fill('90')
  await expect(tramo.getByText('Miércoles · 11:30 – 13:00 · 1 h 30 min')).toBeVisible()
  await tramo.getByRole('button', { name: 'Guardar' }).click()
  await expect(tramo).toBeHidden()
  await expect(page.getByText('11:30 · 90 min')).toBeVisible()

  await page.getByRole('button', { name: 'Agregar excepción', exact: true }).click()
  const excepcion = page.getByRole('dialog', { name: 'Agregar excepción' })
  await pickDate(page, 'schedules-create-date', '2099-12-30')
  await excepcion.getByLabel('Motivo (opcional)').fill('Inventario')
  await excepcion.getByRole('button', { name: 'Guardar' }).click()
  await expect(excepcion).toBeHidden()
  await expect(
    page.getByRole('list', { name: 'Excepciones de jornada' }).getByText('2099-12-30'),
  ).toBeVisible()
})

// Escritorio: la pantalla cabe sin desplazarse. El contenedor que se desplaza
// es el <main> del cascarón; si su contenido cabe, no hay scroll vertical.
const noScrollViewports = [
  { name: '1280', width: 1280, height: 900 },
  { name: '1440', width: 1440, height: 820 },
  { name: '1920', width: 1920, height: 1080 },
]

test('horarios: en escritorio la semana y el panel caben sin desplazarse', async ({ page }) => {
  await openSchedules(page, 1280, 900)
  for (const viewport of noScrollViewports) {
    await page.setViewportSize(viewport)
    await page.waitForTimeout(1400)
    const overflow = await page.evaluate(() => {
      const content = document.querySelector('.private-shell__content')!
      const board = document.querySelector('.week-board')!
      return {
        page: content.scrollHeight - content.clientHeight,
        board: board.scrollHeight - board.clientHeight,
      }
    })
    expect(overflow, `sin scroll a ${viewport.name}`).toEqual({ page: 0, board: 0 })
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'sin-scroll.png') })
  }
})

test('horarios: el acordeón abre un apartado a la vez y oculta el resto', async ({ page }) => {
  await openSchedules(page, 1280, 900)
  const holidays = page.getByRole('button', { name: 'Calendario de festivos colombianos' })
  const exceptions = page.getByRole('button', { name: /^Excepciones de jornada/ })
  const upcoming = page.getByRole('button', { name: 'Próximos festivos colombianos' })

  // Abre primero el de excepciones: es lo que más se edita.
  await expect(exceptions).toHaveAttribute('aria-expanded', 'true')
  await expect(holidays).toHaveAttribute('aria-expanded', 'false')
  await expect(page.getByRole('button', { name: 'Agregar excepción', exact: true })).toBeVisible()
  await expect(page.getByLabel(/Cerrar automáticamente los festivos/)).toBeHidden()

  await holidays.click()
  await expect(holidays).toHaveAttribute('aria-expanded', 'true')
  await expect(exceptions).toHaveAttribute('aria-expanded', 'false')
  await expect(page.getByLabel(/Cerrar automáticamente los festivos/)).toBeVisible()
  await expect(page.getByRole('button', { name: 'Agregar excepción', exact: true })).toBeHidden()
  await page.waitForTimeout(600)
  await page.screenshot({ path: path.join(evidenceDir, '1280', 'acordeon-festivos.png') })
  await expectNoAxeViolations(page)

  await upcoming.click()
  await expect(upcoming).toHaveAttribute('aria-expanded', 'true')
  await expect(page.getByText('Navidad')).toBeVisible()
  await page.waitForTimeout(600)
  await page.screenshot({ path: path.join(evidenceDir, '1280', 'acordeon-proximos.png') })
  await expectNoAxeViolations(page)

  // Plegar el abierto deja todos cerrados, y se maneja con teclado.
  await upcoming.focus()
  await page.keyboard.press('Enter')
  await expect(upcoming).toHaveAttribute('aria-expanded', 'false')
  await openAsidePanel(page, 'Excepciones de jornada')
  await expect(page.getByRole('list', { name: 'Excepciones de jornada' })).toBeVisible()

  // En móvil el acordeón pasa bajo el tablero y crece con su contenido.
  await page.setViewportSize({ width: 360, height: 800 })
  await page.getByRole('complementary', { name: 'Festivos y excepciones' }).scrollIntoViewIfNeeded()
  await page.waitForTimeout(600)
  await page.screenshot({ path: path.join(evidenceDir, '360', 'acordeon.png') })
  expect(await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)).toBe(
    false,
  )
})
