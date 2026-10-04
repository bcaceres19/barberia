import { test, expect, type Page, type BrowserContext } from '@playwright/test'
import axe from 'axe-core'
import { existsSync, mkdirSync, readFileSync } from 'node:fs'
import path from 'node:path'

// #286: API y PostgreSQL reales. Datos dedicados en
// database/testdata/ui_bloqueos_barberos_286.sql; no comparte personas ni
// registros mutables con las otras suites.
const BARBER_A = '286b0001-0000-4000-8000-000000000001'
const BARBER_OTHER = '286b0001-0000-4000-8000-000000000002'
const BARBER_B = '286b0002-0000-4000-8000-000000000001'
const evidence = path.resolve('e2e/evidence/barberos/bloqueos-286')

async function login(page: Page, tenant: 1 | 2) {
  // Las sesiones de prueba se guardan solo en .auth (ignorado). Reutilizarlas
  // evita que los proyectos de navegador consuman el umbral de HU-007.
  const statePath = path.resolve(`.auth/bloqueos-286-${tenant}.json`)
  if (existsSync(statePath)) {
    const state = JSON.parse(readFileSync(statePath, 'utf8')) as Awaited<
      ReturnType<BrowserContext['storageState']>
    >
    await page.context().addCookies(state.cookies)
    const session = await page.request.get('/api/v1/private/auth/session')
    if (session.status() === 200) return
  }
  const response = await page.request.post('/api/v1/public/auth/login', {
    data: { email: `bloqueos.286.${tenant}@ejemplo.test`, password: 'PruebaBloqueos286!' },
  })
  expect(response.status()).toBe(200)
  mkdirSync(path.dirname(statePath), { recursive: true })
  await page.context().storageState({ path: statePath })
}
async function openPanel(page: Page) {
  await page.goto('/panel/barberos')
  await page.getByRole('button', { name: 'Bloquear a Alex de prueba', exact: true }).click()
  await expect(
    page.getByRole('dialog', { name: 'Bloqueos de Alex de prueba', exact: true }),
  ).toBeVisible()
  await expect(page.locator('.staff-page__blocks')).toHaveCount(0)
  await expect(page.getByRole('heading', { name: 'Tiempo fuera de agenda' })).toBeVisible()
  await expect(page.locator('.blocks-panel__timezone')).toBeVisible()
}
async function checkAxe(page: Page, selector?: string) {
  await expect(
    selector ? page.locator(selector) : page.locator('.base-dialog--open').last(),
  ).toHaveCSS('opacity', '1')
  // Mide el estado terminado: la entrada del panel y la salida del modal
  // cambian opacidad durante unos milisegundos. No se alteran animaciones.
  await page.evaluate(async () => {
    await document.fonts.ready
    await Promise.all(
      document
        .getAnimations()
        .filter((animation) => animation.effect?.getComputedTiming().iterations !== Infinity)
        .map((animation) => animation.finished.catch(() => undefined)),
    )
  })
  await page.evaluate(axe.source)
  const violations = await page.evaluate(
    async (selector) =>
      (
        await window.axe.run(
          (selector
            ? document.querySelector(selector)
            : Array.from(document.querySelectorAll('.base-dialog--open')).at(-1)) ?? document,
          {
            runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
          },
        )
      ).violations.map((v) => ({
        id: v.id,
        nodes: v.nodes.map((n) => ({ target: n.target, summary: n.failureSummary })),
      })),
    selector,
  )
  expect(violations).toEqual([])
}

async function pickBlockDate(page: Page, label: string, date: string) {
  const form = page.locator('.block-dialog.base-dialog--open')
  await form.getByRole('button', { name: new RegExp(label) }).click()
  const calendar = page.getByRole('dialog', { name: 'Elegir fecha', exact: true })
  const focused = await calendar.locator('[data-date][tabindex="0"]').getAttribute('data-date')
  if (!focused) throw new Error('Calendario sin día enfocado')
  const monthIndex = (value: string) => Number(value.slice(0, 4)) * 12 + Number(value.slice(5, 7))
  const delta = monthIndex(date) - monthIndex(focused)
  const years = Math.trunc(delta / 12)
  for (let index = 0; index < Math.abs(years); index++)
    await page.keyboard.press(years > 0 ? 'Shift+PageDown' : 'Shift+PageUp')
  const months = delta - years * 12
  for (let index = 0; index < Math.abs(months); index++)
    await page.keyboard.press(months > 0 ? 'PageDown' : 'PageUp')
  await calendar.locator(`[data-date="${date}"]`).click()
}
async function pickBlockTime(page: Page, label: string, time: string) {
  await page
    .locator('.block-dialog.base-dialog--open')
    .getByRole('button', { name: new RegExp(label) })
    .click()
  const picker = page.getByRole('dialog', { name: label, exact: true })
  await picker.getByRole('spinbutton', { name: 'Hora', exact: true }).fill(time.slice(0, 2))
  await picker.getByRole('spinbutton', { name: 'Minutos', exact: true }).fill(time.slice(3))
  await picker.getByRole('button', { name: 'Aplicar hora', exact: true }).click()
}

test('bloqueos dentro de Barberos: ciclo real, aislamiento, errores y evidencia responsive', async ({
  page,
  browser,
}, testInfo) => {
  test.setTimeout(120_000)
  const pageErrors: string[] = []
  page.on('pageerror', (error) => pageErrors.push(error.message))
  await login(page, 1)
  await openPanel(page)
  await expect(page.getByRole('link', { name: 'Bloqueos', exact: true })).toHaveCount(0)
  const marker = `Pausa de prueba ${testInfo.project.name} ${Date.now()}`
  let blockId: string | undefined
  let seriesId: string | undefined
  try {
    await page.getByRole('button', { name: 'Agregar bloqueo', exact: true }).click()
    const dialog = page.getByRole('dialog', { name: 'Agregar bloqueo puntual' })
    await expect(
      dialog.getByRole('form', { name: 'Bloqueo puntual para Alex de prueba' }),
    ).toBeVisible()
    await expect(dialog.locator('.block-form__identity')).toHaveCount(0)
    await dialog.getByRole('button', { name: 'Tipo de bloqueo', exact: true }).click()
    await dialog.getByRole('option', { name: 'Vacaciones', exact: true }).click()
    await pickBlockDate(page, 'Fecha de inicio', '2099-01-05')
    await pickBlockTime(page, 'Hora de inicio', '22:00')
    await pickBlockDate(page, 'Fecha de fin', '2099-01-06')
    await pickBlockTime(page, 'Hora de fin', '02:00')
    await dialog.getByLabel('Motivo (opcional)').fill(marker)
    mkdirSync(evidence, { recursive: true })
    await checkAxe(page)
    await page.screenshot({
      path: path.join(evidence, `${testInfo.project.name}-puntual-completo.png`),
    })
    const created = page.waitForResponse(
      (r) =>
        r.request().method() === 'POST' && r.url().endsWith(`/barbers/${BARBER_A}/time-blocks`),
    )
    await dialog.getByRole('button', { name: 'Guardar bloqueo' }).click()
    const result = await created
    expect(result.status()).toBe(201)
    blockId = (await result.json()).id
    await expect(dialog).not.toBeVisible()
    await expect(page.locator('.blocks-panel__record').filter({ hasText: marker })).toBeVisible()

    for (const notice of await page.getByRole('button', { name: /Descartar aviso/ }).all())
      await notice.click()
    await page.getByRole('button', { name: 'Agregar serie semanal', exact: true }).click()
    const weekly = page.getByRole('dialog', { name: 'Agregar serie semanal' })
    await weekly.getByRole('button', { name: 'Día de la semana', exact: true }).click()
    await weekly.getByRole('option', { name: 'Miércoles', exact: true }).click()
    await pickBlockTime(page, 'Hora de inicio', '12:00')
    await weekly.getByLabel('Duración (minutos)').fill('45')
    await pickBlockDate(page, 'Vigente desde', '2099-01-01')
    await weekly.getByLabel('Motivo (opcional)').fill(marker)
    await checkAxe(page)
    await page.screenshot({
      path: path.join(evidence, `${testInfo.project.name}-semanal-completo.png`),
    })
    const createdSeries = page.waitForResponse(
      (r) =>
        r.request().method() === 'POST' &&
        r.url().endsWith(`/barbers/${BARBER_A}/time-block-series`),
    )
    await weekly.getByRole('button', { name: 'Guardar serie' }).click()
    const seriesResponse = await createdSeries
    expect(seriesResponse.status()).toBe(201)
    seriesId = (await seriesResponse.json()).id
    await expect(weekly).not.toBeVisible()

    // Cambiar de profesional no mezcla registros; volver los consulta de verdad.
    await page.keyboard.press('Escape')
    await expect(
      page.getByRole('button', { name: 'Bloquear a Alex de prueba', exact: true }),
    ).toBeFocused()
    await page.getByRole('button', { name: 'Bloquear a Samuel de prueba', exact: true }).click()
    await expect(page.getByRole('heading', { name: 'Tiempo fuera de agenda' })).toBeVisible()
    await expect(page.getByText('Sin bloqueos puntuales')).toBeVisible()
    await expect(page.locator('.blocks-panel__record').filter({ hasText: marker })).toHaveCount(0)
    await page.keyboard.press('Escape')
    await page.getByRole('button', { name: 'Bloquear a Alex de prueba', exact: true }).click()
    await expect(page.locator('.blocks-panel__record').filter({ hasText: marker })).toHaveCount(2)

    const dismiss = page.getByRole('button', { name: /Descartar todos/ })
    if (await dismiss.isVisible()) await dismiss.click()
    await checkAxe(page)
    mkdirSync(evidence, { recursive: true })
    for (const viewport of [
      { width: 320, height: 740 },
      { width: 360, height: 800 },
      { width: 768, height: 1024 },
      { width: 1280, height: 900 },
      { width: 640, height: 450 },
    ]) {
      await page.setViewportSize(viewport)
      const size = await page.evaluate(() => ({
        width: innerWidth,
        height: innerHeight,
        overflow: document.documentElement.scrollWidth > innerWidth,
      }))
      expect(size.width).toBe(viewport.width)
      expect(size.height).toBe(viewport.height)
      expect(size.overflow).toBe(false)
      await page.locator('main').evaluate((el) => {
        el.scrollTop = 0
      })
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-equipo.png`),
        fullPage: true,
      })
      await checkAxe(page)
      expect(
        await page
          .locator('.blocks-workspace-dialog')
          .evaluate((el) => el.scrollWidth > el.clientWidth),
      ).toBe(false)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-panel.png`),
        fullPage: true,
      })
      const content = page.locator(
        '.blocks-workspace-dialog > .base-dialog__container > .base-dialog__content',
      )
      await content.evaluate((el) => {
        el.scrollTop = el.scrollHeight
      })
      await expect(
        page
          .locator('.blocks-panel__record')
          .last()
          .getByRole('button', { name: /Retirar/ }),
      ).toBeInViewport()
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-listas-fin.png`),
      })
      await content.evaluate((el) => {
        el.scrollTop = 0
      })
      await page.getByRole('button', { name: 'Agregar bloqueo', exact: true }).click()
      await expect(dialog).toBeVisible()
      await checkAxe(page)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-formulario.png`),
        fullPage: true,
      })
      await dialog.getByRole('button', { name: 'Tipo de bloqueo', exact: true }).press('ArrowDown')
      await expect(dialog.getByRole('option', { name: 'Emergencia', exact: true })).toBeFocused()
      await checkAxe(page)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-selector.png`),
      })
      await page.keyboard.press('Home')
      await page.keyboard.press('Enter')
      await expect(
        dialog.getByRole('button', { name: 'Tipo de bloqueo', exact: true }),
      ).toContainText('Descanso')
      await dialog.getByRole('button', { name: 'Tipo de bloqueo', exact: true }).click()
      await page.keyboard.press('Escape')
      await expect(dialog).toBeVisible()
      await expect(
        dialog.getByRole('button', { name: 'Tipo de bloqueo', exact: true }),
      ).toBeFocused()
      await dialog.getByRole('button', { name: 'Tipo de bloqueo', exact: true }).click()
      await page.keyboard.press('Tab')
      await expect(dialog.getByRole('button', { name: /Fecha de inicio/ })).toBeFocused()
      await dialog.getByRole('button', { name: /Fecha de inicio/ }).click()
      const calendar = page.getByRole('dialog', { name: 'Elegir fecha', exact: true })
      await expect(calendar).toBeVisible()
      await checkAxe(page, '.agenda-date-picker__dialog')
      const calendarBox = await calendar.boundingBox()
      expect(calendarBox!.x).toBeGreaterThanOrEqual(0)
      expect(calendarBox!.y).toBeGreaterThanOrEqual(0)
      expect(calendarBox!.x + calendarBox!.width).toBeLessThanOrEqual(viewport.width)
      expect(calendarBox!.y + calendarBox!.height).toBeLessThanOrEqual(viewport.height)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-calendario.png`),
      })
      await page.keyboard.press('ArrowRight')
      await page.keyboard.press('Escape')
      await expect(dialog).toBeVisible()
      await expect(dialog.getByRole('button', { name: /Fecha de inicio/ })).toBeFocused()
      await dialog.getByRole('button', { name: /Hora de inicio/ }).click()
      const timePicker = page.getByRole('dialog', { name: 'Hora de inicio', exact: true })
      await checkAxe(page, '.base-time-picker__dialog')
      const timeBox = await timePicker.boundingBox()
      const timeFieldBox = await dialog
        .getByRole('button', { name: /Hora de inicio/ })
        .boundingBox()
      expect(timeBox!.width).toBeCloseTo(Math.min(240, Math.max(192, timeFieldBox!.width)), 0)
      expect(timeBox!.height).toBeLessThan(220)
      expect(timeBox!.x).toBeGreaterThanOrEqual(0)
      expect(timeBox!.y).toBeGreaterThanOrEqual(0)
      expect(timeBox!.x + timeBox!.width).toBeLessThanOrEqual(viewport.width)
      expect(timeBox!.y + timeBox!.height).toBeLessThanOrEqual(viewport.height)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-hora.png`),
      })
      await timePicker.getByRole('spinbutton', { name: 'Hora', exact: true }).fill('09')
      const minutes = timePicker.getByRole('spinbutton', { name: 'Minutos', exact: true })
      await expect(minutes).toBeFocused()
      await expect(timePicker.getByRole('listbox')).toHaveCount(0)
      await minutes.fill('37')
      await page.keyboard.press('ArrowUp')
      await expect(minutes).toHaveValue('38')
      await minutes.fill('37')
      await checkAxe(page, '.base-time-picker__dialog')
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-minutos.png`),
      })
      await timePicker.getByRole('button', { name: 'Aplicar hora', exact: true }).click()
      await expect(dialog.getByRole('button', { name: /Hora de inicio/ })).toContainText('09:37')

      expect(await dialog.evaluate((el) => el.scrollWidth > el.clientWidth)).toBe(false)
      await page.keyboard.press('Escape')
      await expect(dialog).not.toBeVisible()
      await expect(page.getByRole('button', { name: 'Agregar bloqueo', exact: true })).toBeFocused()
    }
    await page.emulateMedia({ reducedMotion: 'reduce' })
    expect(
      await page
        .locator('.blocks-workspace-dialog')
        .evaluate((el) => parseFloat(getComputedStyle(el).transitionDuration)),
    ).toBeLessThanOrEqual(0.001)

    // Fallo de escritura: conserva campos e idempotencia para un reintento.
    await page.getByRole('button', { name: 'Agregar bloqueo', exact: true }).click()
    await pickBlockDate(page, 'Fecha de inicio', '2099-02-01')
    await pickBlockTime(page, 'Hora de inicio', '09:00')
    await pickBlockDate(page, 'Fecha de fin', '2099-02-01')
    await pickBlockTime(page, 'Hora de fin', '10:00')
    await dialog.getByLabel('Motivo (opcional)').fill('Datos conservados de prueba')
    await page.route(`**/barbers/${BARBER_A}/time-blocks`, (route) =>
      route.request().method() === 'POST' ? route.abort() : route.continue(),
    )
    await dialog.getByRole('button', { name: 'Guardar bloqueo' }).click()
    await expect(dialog.getByText(/Tus datos siguen aquí/)).toBeVisible()
    await expect(dialog.getByLabel('Motivo (opcional)')).toHaveValue('Datos conservados de prueba')
    await page.unroute(`**/barbers/${BARBER_A}/time-blocks`)
    await page.keyboard.press('Escape')

    const contextB = await browser.newContext({ baseURL: 'http://localhost:5173' })
    try {
      const other = await contextB.newPage()
      await login(other, 2)
      for (const route of [
        `/barbers/${BARBER_A}/time-blocks`,
        `/barbers/${BARBER_A}/time-block-series`,
      ]) {
        const response = await other.request.get(`/api/v1/private${route}`)
        expect(response.status()).toBe(404)
      }
      const own = await other.request.get(`/api/v1/private/barbers/${BARBER_B}/time-blocks`)
      expect(own.status()).toBe(200)
      expect(JSON.stringify(await own.json())).not.toContain(marker)
      const secondBarber = await page.request.get(
        `/api/v1/private/barbers/${BARBER_OTHER}/time-blocks`,
      )
      expect(JSON.stringify(await secondBarber.json())).not.toContain(marker)
    } finally {
      await contextB.close()
    }

    // Las filas se retiran inmediatamente con movimiento reducido. Reconsultar
    // evita que los índices de locator.all() apunten a la fila ya desplazada.
    const ownRecords = page.locator('.blocks-panel__record').filter({ hasText: marker })
    while (await ownRecords.count()) {
      const count = await ownRecords.count()
      await ownRecords
        .first()
        .getByRole('button', { name: /Retirar/ })
        .click()
      await expect(ownRecords).toHaveCount(count - 1)
    }
    await expect(page.locator('.blocks-panel__record').filter({ hasText: marker })).toHaveCount(0)
    await page.goto('/panel/bloqueos')
    await expect(page).toHaveURL(/\/panel\/barberos$/)
    expect(pageErrors).toEqual([])
  } finally {
    if (blockId)
      await page.request.delete(`/api/v1/private/barbers/${BARBER_A}/time-blocks/${blockId}`)
    if (seriesId)
      await page.request.delete(`/api/v1/private/barbers/${BARBER_A}/time-block-series/${seriesId}`)
  }
})

test('el calendario compartido conserva selección civil y recarga en la agenda', async ({
  page,
}, testInfo) => {
  await login(page, 1)
  await page.goto('/panel')
  const trigger = page.locator('#daily-agenda-date-picker')
  await expect(trigger).toBeEnabled()
  await trigger.click()
  const calendar = page.getByRole('dialog', { name: 'Elegir fecha', exact: true })
  await checkAxe(page, '.agenda-date-picker__dialog')
  await page.keyboard.press('ArrowRight')
  const date = await calendar.locator('[data-date][tabindex="0"]').getAttribute('data-date')
  await page.keyboard.press('Enter')
  await expect(calendar).not.toBeVisible()
  expect(new URL(page.url()).searchParams.get('date')).toBe(date)
  await expect(trigger).toContainText(date!)
  await page.reload()
  await expect(trigger).toContainText(date!)
  await trigger.click()
  await page.screenshot({
    path: path.join(evidence, `${testInfo.project.name}-agenda-calendario.png`),
  })
  await page.keyboard.press('Escape')
  await expect(trigger).toBeFocused()
})
