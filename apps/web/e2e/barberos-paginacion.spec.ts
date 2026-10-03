import { test, expect, type Page } from '@playwright/test'
import axe from 'axe-core'
import { mkdirSync, existsSync, readFileSync } from 'node:fs'
import path from 'node:path'

// #288: servidor y PostgreSQL reales; solo fixtures sintéticos dedicados.
const evidence = path.resolve('e2e/evidence/barberos/paginacion-288')
async function login(page: Page, tenant: 1 | 2) {
  const file = path.resolve(`.auth/paginacion-288-${tenant}.json`)
  if (existsSync(file)) {
    const state = JSON.parse(readFileSync(file, 'utf8')) as Awaited<
      ReturnType<ReturnType<Page['context']>['storageState']>
    >
    await page.context().addCookies(state.cookies)
    if ((await page.request.get('/api/v1/private/auth/session')).ok()) return
  }
  const response = await page.request.post('/api/v1/public/auth/login', {
    data: { email: `paginacion.288.${tenant}@ejemplo.test`, password: 'PruebaBloqueos286!' },
  })
  expect(response.status()).toBe(200)
  mkdirSync(path.dirname(file), { recursive: true })
  await page.context().storageState({ path: file })
}
async function settled(page: Page) {
  await expect(page.locator('.staff-page__row').first()).toBeVisible()
  await expect(page.locator('.staff-page__ready')).toHaveAttribute('aria-busy', 'false')
  await expect(page.locator('.staff-page__ready--fitting')).toHaveCount(0)
  await page.evaluate(async () => {
    await document.fonts.ready
    await Promise.all(
      document
        .getAnimations()
        .filter((a) => a.effect?.getComputedTiming().iterations !== Infinity)
        .map((a) => a.finished.catch(() => undefined)),
    )
  })
}
async function fits(page: Page, width: number, height: number) {
  await expect.poll(() => page.evaluate(() => [innerWidth, innerHeight])).toEqual([width, height])
  await expect
    .poll(() =>
      page.evaluate(() => {
        const el = document.querySelector<HTMLElement>('.private-shell__content')!
        const footer = document.querySelector<HTMLElement>('.staff-page__footer')!
        const ready = document.querySelector('.staff-page__ready')!
        const row = document.querySelector<HTMLElement>('.staff-page__row')
        const next = document.querySelector<HTMLButtonElement>('[aria-label="Página siguiente"]')!
        const section = document.querySelector<HTMLElement>('.staff-page')!
        const spare =
          el.getBoundingClientRect().bottom -
          footer.getBoundingClientRect().bottom -
          parseFloat(getComputedStyle(section).paddingBottom)
        const fits =
          ready.getAttribute('aria-busy') === 'false' &&
          !!row &&
          (next.disabled || spare < row.getBoundingClientRect().height + 2) &&
          el.scrollHeight <= el.clientHeight + 1 &&
          el.scrollWidth <= el.clientWidth + 1 &&
          footer.getBoundingClientRect().bottom <= el.getBoundingClientRect().bottom + 1
        return {
          fits,
          spare,
          rowHeight: row?.getBoundingClientRect().height,
          rows: document.querySelectorAll('.staff-page__row').length,
          scroll: el.scrollHeight,
          client: el.clientHeight,
          busy: ready.getAttribute('aria-busy'),
          next: next.disabled,
          viewport: [innerWidth, innerHeight],
        }
      }),
    )
    .toMatchObject({ fits: true })
}

test('paginador real: reemplazo, reintento, resize, accesibilidad y aislamiento', async ({
  page,
  browser,
}, testInfo) => {
  test.setTimeout(120000)
  const errors: string[] = []
  page.on('pageerror', (e) => errors.push(e.message))
  await login(page, 1)
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.goto('/panel/barberos')
  await settled(page)
  await expect(page.locator('.staff-page__count')).toHaveText('17 barberos')
  const first = await page.locator('.staff-page__item-name').allTextContents()
  await page.getByRole('button', { name: 'Página siguiente', exact: true }).click()
  await settled(page)
  await expect(page.getByRole('button', { name: 'Página 2', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  )
  const second = await page.locator('.staff-page__item-name').allTextContents()
  expect(second.every((name) => !first.includes(name))).toBe(true)
  await page.getByRole('button', { name: 'Página anterior', exact: true }).click()
  await settled(page)
  expect(await page.locator('.staff-page__item-name').allTextContents()).toEqual(first)
  let fail = true
  await page.route('**/api/v1/private/barbers?*', async (route) => {
    if (fail && new URL(route.request().url()).searchParams.get('page') === '2') {
      fail = false
      await route.fulfill({
        status: 503,
        contentType: 'application/problem+json',
        body: JSON.stringify({ title: 'Fallo sintético' }),
      })
    } else await route.continue()
  })
  await page.getByRole('button', { name: 'Página siguiente', exact: true }).click()
  await expect(page.getByText('No pudimos cargar esta página', { exact: true })).toBeVisible()
  expect(await page.locator('.staff-page__item-name').allTextContents()).toEqual(first)
  await page.getByRole('button', { name: 'Reintentar página', exact: true }).click()
  await settled(page)
  await expect(page.getByRole('button', { name: 'Página 2', exact: true })).toHaveAttribute(
    'aria-current',
    'page',
  )
  await page.unroute('**/api/v1/private/barbers?*')
  mkdirSync(evidence, { recursive: true })
  const counts: number[] = []
  for (const [width, height] of [
    [320, 720],
    [360, 800],
    [768, 900],
    [1280, 900],
    [1280, 650],
    [640, 450],
  ]) {
    await page.setViewportSize({ width: width!, height: height! })
    await fits(page, width!, height!)
    await settled(page)
    counts.push(await page.locator('.staff-page__row').count())
    await page.screenshot({
      path: path.join(evidence, `${testInfo.project.name}-${width}x${height}.png`),
    })
    await page.evaluate(axe.source)
    const violations = await page.evaluate(async () =>
      (
        await window.axe.run(document, {
          runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
        })
      ).violations.map((v) => ({ id: v.id, targets: v.nodes.map((n) => n.target) })),
    )
    expect(violations).toEqual([])
  }
  expect(counts[3]).toBeGreaterThan(counts[4]!)
  await page.setViewportSize({ width: 320, height: 720 })
  await fits(page, 320, 720)
  await settled(page)
  // Página intermedia: también debe caber cuando hay puntos suspensivos a ambos lados.
  const last = await page.request.get('/api/v1/private/barbers?page=3&pageSize=3')
  expect(last.status()).toBe(200)
  for (let n = 0; n < 2; n++) {
    const next = page.getByRole('button', { name: 'Página siguiente', exact: true })
    if (await next.isEnabled()) {
      await next.click()
      await settled(page)
    }
  }
  await fits(page, 320, 720)
  const previous = page.getByRole('button', { name: 'Página anterior', exact: true })
  await previous.focus()
  await page.keyboard.press('Enter')
  await settled(page)
  await expect(page.locator('.staff-page__pagination-nav [aria-current="page"]')).toBeFocused()
  await page
    .getByRole('button', { name: /Bloquear a/ })
    .first()
    .click()
  await expect(page.getByRole('dialog', { name: /Bloqueos de/ })).toBeVisible()
  await page.keyboard.press('Escape')
  await expect(page.getByRole('dialog', { name: /Bloqueos de/ })).toBeHidden()
  const context = await browser.newContext({
    baseURL: testInfo.project.use.baseURL ?? process.env.APP_BASE_URL ?? 'http://localhost:5173',
  })
  const other = await context.newPage()
  await login(other, 2)
  const response = await other.request.get('/api/v1/private/barbers?page=1&pageSize=2')
  expect(response.status()).toBe(200)
  const body = (await response.json()) as { total: number; items: { fullName: string }[] }
  expect(body.total).toBe(3)
  expect(body.items.every((b) => b.fullName.startsWith('Otra barbería'))).toBe(true)
  expect(
    (
      await other.request.get('/api/v1/private/barbers/288b0001-0000-4000-8000-000000000001')
    ).status(),
  ).toBe(404)
  const cursor = await other.request.get('/api/v1/private/barbers?limit=1')
  expect(cursor.status()).toBe(200)
  const legacy = (await cursor.json()) as { nextCursor: string | null; total?: number }
  expect(legacy.nextCursor).not.toBeNull()
  expect(legacy.total).toBeUndefined()
  await context.close()
  expect(errors).toEqual([])
})
