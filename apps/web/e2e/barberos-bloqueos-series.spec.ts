import { test, expect, type Locator, type Page, type BrowserContext } from '@playwright/test'
import axe from 'axe-core'
import { existsSync, mkdirSync, readFileSync } from 'node:fs'
import path from 'node:path'

// #100: bloqueos por fechas explícitas, excepciones y edición de serie por
// alcance, sobre API y PostgreSQL reales. Reutiliza los datos dedicados de
// #286 (database/testdata/ui_bloqueos_barberos_286.sql); cada ejecución crea
// sus propias series con un rótulo único y las retira al terminar.
const BARBER_A = '286b0001-0000-4000-8000-000000000001'
const evidence = path.resolve('e2e/evidence/barberos/bloqueos-100')
const API = `/api/v1/private/barbers/${BARBER_A}`

test.use({ contextOptions: { reducedMotion: 'reduce' } })

async function login(page: Page) {
  // Una sola sesión por ejecución: HU-007 limita los accesos repetidos.
  const statePath = path.resolve('.auth/bloqueos-286-1.json')
  if (existsSync(statePath)) {
    const state = JSON.parse(readFileSync(statePath, 'utf8')) as Awaited<
      ReturnType<BrowserContext['storageState']>
    >
    await page.context().addCookies(state.cookies)
    const session = await page.request.get('/api/v1/private/auth/session')
    if (session.status() === 200) return
  }
  const response = await page.request.post('/api/v1/public/auth/login', {
    data: { email: 'bloqueos.286.1@ejemplo.test', password: 'PruebaBloqueos286!' },
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
  await expect(page.locator('.blocks-panel__timezone')).toBeVisible()
  await expect(page.getByRole('button', { name: 'Agregar bloqueo por fechas' })).toBeEnabled()
}

async function checkAxe(page: Page, selector = '.base-dialog--open:last-of-type') {
  const target = page.locator('.base-dialog--open').last()
  await expect(target).toHaveCSS('opacity', '1')
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
    async (scope) =>
      (
        await window.axe.run(
          (scope === 'last-dialog'
            ? Array.from(document.querySelectorAll('.base-dialog--open')).at(-1)
            : document.querySelector(scope)) ?? document,
          { runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] } },
        )
      ).violations.map((v) => ({
        id: v.id,
        nodes: v.nodes.map((n) => ({ target: n.target, summary: n.failureSummary })),
      })),
    selector === '.base-dialog--open:last-of-type' ? 'last-dialog' : selector,
  )
  expect(violations).toEqual([])
}

async function pickDate(page: Page, scope: Locator, label: RegExp | string, date: string) {
  await scope.getByRole('button', { name: label }).first().click()
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

async function choose(scope: Locator, label: string, option: string) {
  await scope.getByRole('button', { name: label, exact: true }).click()
  await scope.getByRole('option', { name: option, exact: true }).click()
}

async function listSeries(page: Page) {
  const body = await page.evaluate(async (url) => {
    const response = await fetch(`${url}/time-block-series?limit=50`)
    return response.json()
  }, API)
  return body.items as {
    id: string
    recurrenceKind: string
    reason: string | null
    effectiveFrom: string
    effectiveUntil: string | null
    durationMinutes: number
    dates: { blockDate: string }[]
    exceptions: { excludedDate: string; reason: string | null }[]
  }[]
}

async function dismissNotices(page: Page) {
  for (const notice of await page.getByRole('button', { name: /Descartar aviso/ }).all())
    await notice.click()
}

test('series por fechas, excepciones y edición por alcance con API real', async ({
  page,
}, testInfo) => {
  test.setTimeout(180_000)
  const pageErrors: string[] = []
  page.on('pageerror', (error) => pageErrors.push(error.message))
  const marker = `Serie de prueba ${testInfo.project.name} ${Date.now()}`
  const created: string[] = []
  mkdirSync(evidence, { recursive: true })

  await login(page)
  await openPanel(page)
  try {
    // --- Alta por fechas explícitas -----------------------------------
    await page.getByRole('button', { name: 'Agregar bloqueo por fechas' }).click()
    const form = page.getByRole('dialog', { name: 'Agregar bloqueo por fechas' })
    await expect(form).toBeVisible()
    const scope = page.locator('.block-dialog.base-dialog--open')
    await pickDate(page, scope, /Vigente desde/, '2099-12-01')
    await choose(scope, 'Fin de la vigencia', 'Hasta una fecha')
    await pickDate(page, scope, /Vigente hasta/, '2099-12-31')
    await pickDate(page, scope, /Elegir una fecha/, '2099-12-16')
    await form.getByRole('button', { name: 'Añadir a la lista' }).click()
    await pickDate(page, scope, /Elegir una fecha/, '2099-12-15')
    await form.getByRole('button', { name: 'Añadir a la lista' }).click()
    await expect(
      form.getByRole('list', { name: 'Fechas elegidas' }).getByRole('listitem'),
    ).toHaveCount(2)
    await form.getByLabel('Motivo (opcional)').fill(marker)

    // Una fecha fuera de la vigencia se explica antes de enviar nada.
    await pickDate(page, scope, /Elegir una fecha/, '2100-01-05')
    await form.getByRole('button', { name: 'Añadir a la lista' }).click()
    await form.getByRole('button', { name: 'Guardar bloqueo' }).click()
    await expect(form.getByText(/queda fuera de la vigencia/)).toBeVisible()
    await form.getByRole('button', { name: /^Quitar la fecha .*2100/ }).click()

    await checkAxe(page)
    await page.screenshot({ path: path.join(evidence, `${testInfo.project.name}-alta-fechas.png`) })
    const createdResponse = page.waitForResponse(
      (r) => r.request().method() === 'POST' && r.url().endsWith(`${API}/time-block-series`),
    )
    await form.getByRole('button', { name: 'Guardar bloqueo' }).click()
    const dateListResponse = await createdResponse
    expect(dateListResponse.status()).toBe(201)
    const dateListId = (await dateListResponse.json()).id as string
    created.push(dateListId)
    await expect(form).not.toBeVisible()
    const dateListRecord = page
      .locator('.blocks-panel__record')
      .filter({ hasText: marker })
      .filter({ hasText: '2 fechas' })
    await expect(dateListRecord).toBeVisible()
    await dismissNotices(page)

    // --- Fechas y excepciones de la serie por fechas --------------------
    await dateListRecord.getByRole('button', { name: /^Fechas y excepciones de la serie/ }).click()
    const instances = page.getByRole('dialog', { name: 'Fechas y excepciones' })
    await expect(instances).toBeVisible()
    await pickDate(page, scope, /Agregar una fecha/, '2099-12-20')
    const addDate = page.waitForResponse(
      (r) =>
        r.request().method() === 'POST' &&
        r.url().endsWith(`/time-block-series/${dateListId}/dates`),
    )
    await instances.getByRole('button', { name: 'Agregar fecha' }).click()
    expect((await addDate).status()).toBe(204)
    await expect(
      instances.getByRole('list', { name: 'Fechas de la serie' }).getByRole('listitem'),
    ).toHaveCount(3)

    // Repetir una fecha es un 409 determinista que la pantalla explica.
    await pickDate(page, scope, /Agregar una fecha/, '2099-12-20')
    await instances.getByRole('button', { name: 'Agregar fecha' }).click()
    await expect(instances.getByText('Esa fecha ya está en la serie.')).toBeVisible()

    await pickDate(page, scope, /Instancia que no se bloquea/, '2099-12-16')
    await instances.getByLabel('Motivo de la excepción (opcional)').fill('Libró ese día')
    const addException = page.waitForResponse(
      (r) =>
        r.request().method() === 'POST' &&
        r.url().endsWith(`/time-block-series/${dateListId}/exceptions`),
    )
    await instances.getByRole('button', { name: 'Agregar excepción' }).click()
    expect((await addException).status()).toBe(204)
    await expect(
      instances.getByRole('list', { name: 'Excepciones de la serie' }).getByText('Libró ese día'),
    ).toBeVisible()

    const removeDate = page.waitForResponse(
      (r) =>
        r.request().method() === 'DELETE' &&
        r.url().endsWith(`/time-block-series/${dateListId}/dates/2099-12-15`),
    )
    await instances.getByRole('button', { name: /^Quitar la fecha .*15/ }).click()
    expect((await removeDate).status()).toBe(204)
    await checkAxe(page)
    await page.screenshot({
      path: path.join(evidence, `${testInfo.project.name}-fechas-excepciones.png`),
    })

    const afterEdits = (await listSeries(page)).find((item) => item.id === dateListId)!
    expect(afterEdits.dates.map((item) => item.blockDate).sort()).toEqual([
      '2099-12-16',
      '2099-12-20',
    ])
    expect(afterEdits.exceptions).toEqual([
      expect.objectContaining({ excludedDate: '2099-12-16', reason: 'Libró ese día' }),
    ])

    const restore = page.waitForResponse(
      (r) => r.request().method() === 'DELETE' && r.url().endsWith(`/exceptions/2099-12-16`),
    )
    await instances.getByRole('button', { name: /^Restaurar la instancia del/ }).click()
    expect((await restore).status()).toBe(204)
    await expect(instances.getByText('Ninguna instancia está exceptuada.')).toBeVisible()
    await instances.getByRole('button', { name: 'Cerrar', exact: true }).click()
    await expect(instances).not.toBeVisible()
    await expect(
      dateListRecord.getByRole('button', { name: /^Fechas y excepciones de la serie/ }),
    ).toBeFocused()
    await expect(dateListRecord).toContainText('2 fechas')

    // --- Serie semanal: editar toda la serie y «esta y las siguientes» --
    await page.getByRole('button', { name: 'Agregar serie semanal', exact: true }).click()
    const weekly = page.getByRole('dialog', { name: 'Agregar serie semanal' })
    await choose(scope, 'Día de la semana', 'Miércoles')
    await scope.getByRole('button', { name: /Hora de inicio/ }).click()
    const timePicker = page.getByRole('dialog', { name: 'Hora de inicio', exact: true })
    await timePicker.getByRole('spinbutton', { name: 'Hora', exact: true }).fill('12')
    await timePicker.getByRole('spinbutton', { name: 'Minutos', exact: true }).fill('00')
    await timePicker.getByRole('button', { name: 'Aplicar hora', exact: true }).click()
    await weekly.getByLabel('Duración (minutos)').fill('45')
    await pickDate(page, scope, /Vigente desde/, '2099-01-01')
    await weekly.getByLabel('Motivo (opcional)').fill(marker)
    const weeklyResponse = page.waitForResponse(
      (r) => r.request().method() === 'POST' && r.url().endsWith(`${API}/time-block-series`),
    )
    await weekly.getByRole('button', { name: 'Guardar serie' }).click()
    const weeklyId = (await (await weeklyResponse).json()).id as string
    created.push(weeklyId)
    await dismissNotices(page)
    const weeklyRecord = page
      .locator('.blocks-panel__record')
      .filter({ hasText: marker })
      .filter({ hasText: 'Semanal' })
    await expect(weeklyRecord).toBeVisible()

    await weeklyRecord.getByRole('button', { name: /^Editar serie/ }).click()
    const edit = page.getByRole('dialog', { name: 'Editar serie de bloqueo' })
    await expect(edit).toBeVisible()
    await expect(edit.getByLabel('Duración (minutos)')).toHaveValue('45')
    await edit.getByLabel('Duración (minutos)').fill('50')
    await checkAxe(page)
    await page.screenshot({
      path: path.join(evidence, `${testInfo.project.name}-editar-serie.png`),
    })
    const patchWhole = page.waitForResponse(
      (r) => r.request().method() === 'PATCH' && r.url().endsWith(`/time-block-series/${weeklyId}`),
    )
    await edit.getByRole('button', { name: 'Guardar cambios' }).click()
    expect((await patchWhole).status()).toBe(200)
    await expect(edit).not.toBeVisible()
    await expect(weeklyRecord).toContainText('50 min')
    await dismissNotices(page)

    await weeklyRecord.getByRole('button', { name: /^Editar serie/ }).click()
    await choose(scope, 'Qué cambia', 'Esta fecha y las siguientes')
    // Un corte anterior al inicio de la serie se explica antes de enviar.
    await pickDate(page, scope, /Aplicar desde/, '2099-01-01')
    await edit.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(edit.getByText(/El corte debe ser posterior al inicio/)).toBeVisible()
    await pickDate(page, scope, /Aplicar desde/, '2099-06-01')
    await edit.getByLabel('Duración (minutos)').fill('30')
    await expect(edit.getByText(/La serie actual termina el .*31/)).toBeVisible()
    await checkAxe(page)
    await page.screenshot({
      path: path.join(evidence, `${testInfo.project.name}-editar-desde.png`),
    })
    const patchSplit = page.waitForResponse(
      (r) => r.request().method() === 'PATCH' && r.url().endsWith(`/time-block-series/${weeklyId}`),
    )
    await edit.getByRole('button', { name: 'Guardar cambios' }).click()
    const split = await patchSplit
    expect(split.status()).toBe(200)
    const newId = (await split.json()).id as string
    created.push(newId)
    expect(newId).not.toBe(weeklyId)
    await expect(
      page
        .locator('.blocks-panel__record')
        .filter({ hasText: marker })
        .filter({ hasText: 'Semanal' }),
    ).toHaveCount(2)

    const series = await listSeries(page)
    expect(series.find((item) => item.id === weeklyId)).toMatchObject({
      effectiveUntil: '2099-05-31',
      durationMinutes: 50,
    })
    expect(series.find((item) => item.id === newId)).toMatchObject({
      effectiveFrom: '2099-06-01',
      durationMinutes: 30,
    })

    // Excepción sobre una serie semanal (sin lista de fechas).
    const newRecord = page
      .locator('.blocks-panel__record')
      .filter({ hasText: marker })
      .filter({ hasText: '30 min' })
    await newRecord.getByRole('button', { name: /^Excepciones de la serie/ }).click()
    const weeklyInstances = page.getByRole('dialog', { name: 'Excepciones de la serie' })
    await expect(weeklyInstances.getByRole('list', { name: 'Fechas de la serie' })).toHaveCount(0)
    await pickDate(page, scope, /Instancia que no se bloquea/, '2099-06-03')
    await weeklyInstances.getByRole('button', { name: 'Agregar excepción' }).click()
    await expect(
      weeklyInstances.getByRole('list', { name: 'Excepciones de la serie' }).getByRole('listitem'),
    ).toHaveCount(1)
    await weeklyInstances.getByRole('button', { name: 'Cerrar', exact: true }).click()
    await expect(newRecord).toContainText('1 excepción')

    // --- Recarga: todo persiste en el servidor ---------------------------
    await page.reload()
    await openPanel(page)
    await expect(page.locator('.blocks-panel__record').filter({ hasText: marker })).toHaveCount(3)
    await expect(
      page
        .locator('.blocks-panel__record')
        .filter({ hasText: marker })
        .filter({ hasText: '2 fechas' }),
    ).toBeVisible()

    // --- Responsive, teclado y desbordamiento --------------------------
    for (const viewport of [
      { width: 320, height: 740 },
      { width: 360, height: 800 },
      { width: 768, height: 1024 },
      { width: 1280, height: 900 },
    ]) {
      await page.setViewportSize(viewport)
      const overflow = () => page.evaluate(() => document.documentElement.scrollWidth > innerWidth)
      expect(await overflow()).toBe(false)
      await checkAxe(page, '.blocks-panel')
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-panel.png`),
      })
      const record = page
        .locator('.blocks-panel__record')
        .filter({ hasText: marker })
        .filter({ hasText: '2 fechas' })
      await record.scrollIntoViewIfNeeded()
      for (const name of [/^Fechas y excepciones de la serie/, /^Editar serie/, /^Retirar serie/]) {
        await expect(record.getByRole('button', { name })).toBeInViewport()
        const box = await record.getByRole('button', { name }).boundingBox()
        expect(box!.x).toBeGreaterThanOrEqual(0)
        expect(box!.x + box!.width).toBeLessThanOrEqual(viewport.width)
        expect(box!.height).toBeGreaterThanOrEqual(44)
      }
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-registro.png`),
      })

      await record.getByRole('button', { name: /^Fechas y excepciones de la serie/ }).click()
      const dialog = page.getByRole('dialog', { name: 'Fechas y excepciones' })
      await expect(dialog).toBeVisible()
      expect(await dialog.evaluate((el) => el.scrollWidth > el.clientWidth)).toBe(false)
      expect(await overflow()).toBe(false)
      await checkAxe(page)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-fechas.png`),
      })
      await page.keyboard.press('Escape')
      await expect(dialog).not.toBeVisible()
      await expect(
        record.getByRole('button', { name: /^Fechas y excepciones de la serie/ }),
      ).toBeFocused()

      await page.getByRole('button', { name: 'Agregar bloqueo por fechas' }).click()
      const alta = page.getByRole('dialog', { name: 'Agregar bloqueo por fechas' })
      await expect(alta).toBeVisible()
      expect(await alta.evaluate((el) => el.scrollWidth > el.clientWidth)).toBe(false)
      await checkAxe(page)
      await page.screenshot({
        path: path.join(evidence, `${testInfo.project.name}-${viewport.width}-alta.png`),
        fullPage: true,
      })
      await page.keyboard.press('Escape')
      await expect(alta).not.toBeVisible()
    }
    expect(pageErrors).toEqual([])
  } finally {
    // Retiro lógico (RN-BLQ-04): una serie retirada deja de mostrarse y
    // el siguiente arranque la ignora.
    for (const id of created) await page.request.delete(`${API}/time-block-series/${id}`)
  }
})
