import { test, expect, type Page, type BrowserContext } from '@playwright/test'
import axe from 'axe-core'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import path from 'node:path'

// #198: revisión integral final del rediseño NAVA / Tailored Grid. Recorre
// las rutas reales del producto con API y PostgreSQL reales, en los cuatro
// anchos de la guía visual, y registra por ruta: axe WCAG A/AA (contraste
// incluido), desbordamiento horizontal, errores de consola y de página, y la
// consistencia del cascarón (un solo <main>, un solo <h1>, marca y navegación,
// tipografía de títulos). Datos sintéticos de database/testdata/
// ui_bloqueos_barberos_286.sql. Las capturas de cada ruta (1280 y 320) y
// report.json son la evidencia versionada.
const evidence = path.resolve('e2e/evidence/revision-integral-198')
const VIEWPORTS = [
  { width: 1280, height: 900 },
  { width: 768, height: 1024 },
  { width: 360, height: 800 },
  { width: 320, height: 740 },
]
const PRIVATE_ROUTES = [
  { name: 'agenda', url: '/panel' },
  { name: 'nuevo-turno', url: '/panel/turnos/nuevo' },
  { name: 'barberos', url: '/panel/barberos' },
  { name: 'servicios', url: '/panel/servicios' },
  { name: 'servicios-por-barbero', url: '/panel/servicios-por-barbero' },
  { name: 'horarios', url: '/panel/horarios' },
  { name: 'barberia', url: '/panel/barberia' },
  { name: 'reserva-publica', url: '/panel/reserva-publica' },
]
const PUBLIC_ROUTES = [
  { name: 'acceso', url: '/acceso' },
  { name: 'recuperar-acceso', url: '/recuperar-acceso' },
]

interface RouteResult {
  route: string
  finalPath: string
  width: number
  axeViolations: { id: string; nodes: number }[]
  horizontalOverflow: boolean
  consoleErrors: string[]
  pageErrors: string[]
  mains: number
  h1s: number
  h1Font: string | null
  landmarksUnique: boolean
  truncatedNavLabels: string[]
}

const results: RouteResult[] = []

async function login(page: Page) {
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

async function settle(page: Page) {
  await page.waitForLoadState('networkidle')
  await page.evaluate(async () => {
    await document.fonts.ready
    await Promise.all(
      document
        .getAnimations()
        .filter((animation) => animation.effect?.getComputedTiming().iterations !== Infinity)
        .map((animation) => animation.finished.catch(() => undefined)),
    )
  })
}

async function inspect(
  page: Page,
  route: { name: string; url: string },
  viewport: { width: number; height: number },
  shot: boolean,
  errors: { console: string[]; page: string[] },
) {
  await page.setViewportSize(viewport)
  errors.console.length = 0
  errors.page.length = 0
  await page.goto(route.url)
  await settle(page)
  await page.evaluate(axe.source)
  const found = await page.evaluate(async () => {
    const run = await window.axe.run(document, {
      runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
    })
    return run.violations.map((v) => ({ id: v.id, nodes: v.nodes.length }))
  })
  const shell = await page.evaluate(() => {
    const h1 = document.querySelector('h1')
    const names = Array.from(document.querySelectorAll('nav, [role="banner"], header')).map(
      (el) => `${el.tagName}:${el.getAttribute('aria-label') ?? ''}`,
    )
    // Un rótulo del dock cortado con «…» oculta a dónde lleva el destino; el
    // dock móvil se mide aparte (display:none en escritorio y viceversa).
    const truncated = Array.from(document.querySelectorAll<HTMLElement>('.app-nav__label'))
      .filter((el) => el.offsetParent !== null && el.scrollWidth > el.clientWidth)
      .map((el) => el.textContent?.trim() ?? '')
    return {
      truncated,
      overflow: document.documentElement.scrollWidth > innerWidth,
      mains: document.querySelectorAll('main').length,
      h1s: document.querySelectorAll('h1').length,
      h1Font: h1 ? getComputedStyle(h1).fontFamily : null,
      landmarksUnique:
        new Set(names.filter((n) => n.endsWith(':') === false)).size ===
        names.filter((n) => n.endsWith(':') === false).length,
    }
  })
  results.push({
    route: route.name,
    finalPath: new URL(page.url()).pathname,
    width: viewport.width,
    axeViolations: found,
    horizontalOverflow: shell.overflow,
    consoleErrors: [...errors.console],
    pageErrors: [...errors.page],
    mains: shell.mains,
    h1s: shell.h1s,
    h1Font: shell.h1Font,
    landmarksUnique: shell.landmarksUnique,
    truncatedNavLabels: shell.truncated,
  })
  if (shot) {
    await page.screenshot({
      path: path.join(evidence, `${route.name}-${viewport.width}.png`),
      fullPage: false,
    })
  }
}

test.describe.configure({ mode: 'serial' })
test.use({ contextOptions: { reducedMotion: 'reduce' } })

test('rutas privadas del panel en los cuatro anchos', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium-desktop', 'La revisión integral corre en un motor')
  test.setTimeout(300_000)
  mkdirSync(evidence, { recursive: true })
  const errors = { console: [] as string[], page: [] as string[] }
  page.on('console', (message) => {
    if (message.type() === 'error') errors.console.push(message.text())
  })
  page.on('pageerror', (error) => errors.page.push(error.message))
  await login(page)
  for (const route of PRIVATE_ROUTES) {
    for (const viewport of VIEWPORTS) {
      await inspect(
        page,
        route,
        viewport,
        viewport.width === 1280 || viewport.width === 320,
        errors,
      )
    }
  }
  // /panel/bloqueos ya no es un destino propio (DEC-105): redirige a Barberos.
  await page.setViewportSize(VIEWPORTS[0]!)
  await page.goto('/panel/bloqueos')
  await expect(page).toHaveURL(/\/panel\/barberos$/)
})

test('acceso y recuperación en los cuatro anchos', async ({ browser }, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium-desktop', 'La revisión integral corre en un motor')
  test.setTimeout(120_000)
  const context = await browser.newContext({
    baseURL: testInfo.project.use.baseURL,
    reducedMotion: 'reduce',
  })
  const page = await context.newPage()
  const errors = { console: [] as string[], page: [] as string[] }
  page.on('console', (message) => {
    if (message.type() === 'error') errors.console.push(message.text())
  })
  page.on('pageerror', (error) => errors.page.push(error.message))
  try {
    for (const route of PUBLIC_ROUTES) {
      for (const viewport of VIEWPORTS) {
        await inspect(
          page,
          route,
          viewport,
          viewport.width === 1280 || viewport.width === 320,
          errors,
        )
      }
    }
  } finally {
    await context.close()
  }
})

test('resumen: sin violaciones, desbordamiento ni errores; cascarón consistente', async ({}, testInfo) => {
  test.skip(testInfo.project.name !== 'chromium-desktop', 'La revisión integral corre en un motor')
  writeFileSync(path.join(evidence, 'report.json'), `${JSON.stringify(results, null, 2)}\n`)
  const expected = (PRIVATE_ROUTES.length + PUBLIC_ROUTES.length) * VIEWPORTS.length
  expect(results).toHaveLength(expected)
  // Una sesión vencida redirigiría a /acceso y haría pasar todo en vacío.
  expect(
    [...PRIVATE_ROUTES, ...PUBLIC_ROUTES]
      .flatMap((route) =>
        results.filter((r) => r.route === route.name && r.finalPath !== route.url),
      )
      .map((r) => `${r.route}@${r.width} → ${r.finalPath}`),
  ).toEqual([])
  expect(results.filter((r) => r.axeViolations.length > 0)).toEqual([])
  expect(results.filter((r) => r.horizontalOverflow).map((r) => `${r.route}@${r.width}`)).toEqual(
    [],
  )
  expect(results.flatMap((r) => r.pageErrors)).toEqual([])
  expect(results.flatMap((r) => r.consoleErrors.map((e) => `${r.route}: ${e}`))).toEqual([])
  // A 1280 px (escritorio) ningún rótulo del dock queda truncado.
  expect(
    results
      .filter((r) => r.width === 1280 && r.truncatedNavLabels.length > 0)
      .map((r) => `${r.route}: ${r.truncatedNavLabels.join(', ')}`),
  ).toEqual([])
  // Cada ruta tiene un único <main> y un único <h1>.
  expect(results.filter((r) => r.mains !== 1).map((r) => `${r.route}@${r.width}`)).toEqual([])
  expect(results.filter((r) => r.h1s !== 1).map((r) => `${r.route}@${r.width}`)).toEqual([])
  // Los títulos del panel comparten una sola familia tipográfica.
  const families = new Set(
    results
      .filter((r) => r.route !== 'acceso' && r.route !== 'recuperar-acceso')
      .map((r) => r.h1Font),
  )
  expect([...families]).toHaveLength(1)
})
