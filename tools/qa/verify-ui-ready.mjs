// Smoke de preparación de #300, no campaña completa. Credenciales solo en memoria.
import { randomBytes } from 'node:crypto'
import { readFileSync, mkdirSync, writeFileSync, chmodSync, rmSync } from 'node:fs'
import { resolve, dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { createRequire } from 'node:module'
const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..')
const require = createRequire(join(root, 'apps/web/package.json'))
const { chromium } = require('@playwright/test')
const baseURL = process.env.APP_BASE_URL || 'http://127.0.0.1:5173'
if (!['127.0.0.1', 'localhost', '[::1]'].includes(new URL(baseURL).hostname)) throw new Error('Smoke solo local')
const manifest = JSON.parse(readFileSync(join(root, 'apps/web/.auth/nava-qa/current.json'), 'utf8'))
const output = join(root, 'apps/web/test-results/ui-qa', manifest.campaign, 'preparation')
mkdirSync(output, { recursive: true, mode: 0o700 })
chmodSync(output, 0o700)
const cases = []
const ipPrefix = "10." + randomBytes(2).join(".")
let browser
let lastPage
let step = "launch"
function assert(condition, message) {
  if (!condition) throw new Error(message)
}
async function login(account, index) {
  const context = await browser.newContext()
  const page = await context.newPage()
  page.setDefaultTimeout(15000)
  lastPage = page
  step = "LOGIN-" + account.scope
  await page.route('**/api/v1/public/auth/**', route => route.continue({
    headers: { ...route.request().headers(), 'x-forwarded-for': ipPrefix + '.' + (index + 1) },
  }))
  await page.goto(baseURL + '/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(account.email)
  await page.getByLabel('Contraseña', { exact: true }).fill(account.password)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await page.waitForURL('**/panel', { timeout: 15000 })
  await page.getByRole('heading', { name: 'Agenda', exact: true }).waitFor()
  cases.push({ id: 'LOGIN-' + account.scope, status: 'PASS', mode: 'REAL' })
  return { context, page }
}
try {
  const health = await fetch('http://127.0.0.1:8080/health')
  const database = await fetch('http://127.0.0.1:8080/health/db')
  assert(health.status === 200 && database.status === 200, 'API o base no listas')
  cases.push({ id: 'HEALTH-DB', status: 'PASS', mode: 'REAL' })
  browser = await chromium.launch({ headless: true })
  for (const [index, account] of manifest.accounts.entries()) {
    const { context, page } = await login(account, index)
    const status = await page.evaluate(async () => (await fetch('/api/v1/private/barbers')).status)
    assert(status === 200, 'Consulta privada de profesionales falló en ' + account.scope)
    const daily = await page.evaluate(async id => (await fetch('/api/v1/private/barbers/' + id + '/appointments/daily-agenda?date=' + new Date().toLocaleDateString('en-CA', { timeZone: 'America/Bogota' }))).status, account.barberIds[0])
    assert(daily === 200, 'Agenda diaria no disponible en ' + account.scope)
    cases.push({ id: 'AGENDA-' + account.scope, status: 'PASS', mode: 'REAL' })
    await context.close()
  }
  step = 'INACTIVE-REJECTED'
  const inactiveContext = await browser.newContext()
  const inactivePage = await inactiveContext.newPage()
  lastPage = inactivePage
  await inactivePage.route('**/api/v1/public/auth/**', route => route.continue({ headers: { ...route.request().headers(), 'x-forwarded-for': ipPrefix + '.26' } }))
  await inactivePage.goto(baseURL + '/acceso')
  await inactivePage.getByLabel('Correo', { exact: true }).fill(manifest.inactive.email)
  await inactivePage.getByLabel('Contraseña', { exact: true }).fill(manifest.inactive.password)
  const rejected = inactivePage.waitForResponse(r => r.url().endsWith('/public/auth/login') && r.request().method() === 'POST')
  await inactivePage.getByRole('button', { name: 'Iniciar sesión' }).click()
  assert((await rejected).status() === 401, 'La cuenta inactiva no fue rechazada')
  cases.push({ id: 'INACTIVE-REJECTED', status: 'PASS', mode: 'REAL' })
  await inactiveContext.close()
  const account = manifest.accounts.find(a => a.scope === 'staff')
  const { context, page } = await login(account, 30)
  step = 'SAVE-RELOAD-STAFF'
  await page.goto(baseURL + '/panel/barberos')
  await page.getByRole('heading', { name: 'Barberos', exact: true }).waitFor()
  const name = 'QA preparación ' + Date.now()
  await page.getByRole('button', { name: /Agregar barbero$/ }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
  await dialog.getByRole('textbox', { name: /^Nombre/ }).fill(name)
  await dialog.getByRole('button', { name: 'Guardar', exact: true }).click()
  await dialog.waitFor({ state: 'hidden' })
  await page.reload()
  // El alta se ordena al final; recargar vuelve a página 1. Buscar con la UI.
  await page.getByRole('button', { name: 'Página siguiente', exact: true }).waitFor()
  await page.getByRole('button', { name: /^Página \d+$/ }).last().click()
  await page.getByText(name, { exact: true }).waitFor()
  cases.push({ id: 'SAVE-RELOAD-STAFF', status: 'PASS', mode: 'REAL' })
  // La evidencia de preparación no contiene credenciales ni tokens.
  await page.setViewportSize({ width: 1280, height: 900 })
  await page.screenshot({ path: join(output, 'staff-1280.png'), fullPage: true, animations: 'disabled' })
  await page.setViewportSize({ width: 360, height: 800 })
  await page.screenshot({ path: join(output, 'staff-360.png'), fullPage: true, animations: 'disabled' })
  await context.close()
  for (const scope of ['isolation-a', 'isolation-b']) {
    const account = manifest.accounts.find(a => a.scope === scope)
    const { context, page } = await login(account, scope === 'isolation-a' ? 31 : 32)
    const result = await page.evaluate(async () => {
      const r = await fetch('/api/v1/private/auth/session')
      return { status: r.status, body: await r.json() }
    })
    assert(result.status === 200 && result.body.barbershop?.id === account.shopId, 'Tenant incorrecto')
    const wrong = manifest.accounts.find(a => a.scope === (scope === 'isolation-a' ? 'isolation-b' : 'isolation-a'))
    const crossed = await page.evaluate(async id => (await fetch('/api/v1/private/barbers/' + id + '/appointments/daily-agenda?date=' + new Date().toLocaleDateString('en-CA', { timeZone: 'America/Bogota' }) + '')).status, wrong.barberIds[0])
    assert(crossed === 404, 'Recurso cruzado no recibió 404')
    cases.push({ id: 'TENANT-' + scope, status: 'PASS', mode: 'REAL' })
    await context.close()
  }
  const publicAccount = manifest.accounts.find(a => a.scope === 'public-entry')
  const publicContext = await browser.newContext()
  const publicPage = await publicContext.newPage()
  await publicPage.goto(baseURL + '/reservar/' + publicAccount.slug)
  const publicStatus = await publicPage.evaluate(async slug =>
    (await fetch('/api/v1/public/barbershops/' + slug)).status, publicAccount.slug)
  assert(publicStatus === 200, 'Entrada pública falló')
  await publicContext.close()
  cases.push({ id: 'PUBLIC-ENTRY', status: 'PASS', mode: 'REAL' })
  const report = { kind: 'PREPARATION_ONLY', campaign: manifest.campaign, date: new Date().toISOString(), cases }
  writeFileSync(join(output, 'results.json'), JSON.stringify(report, null, 2) + '\n', { mode: 0o600 })
  rmSync(join(output, 'error.local'), { force: true })
  rmSync(join(output, 'failure-private.png'), { force: true })
  console.log('Preparación UI comprobada: ' + cases.length + ' comprobaciones; informe ' + output.replace(root + '/', ''))
} catch (error) {
  writeFileSync(join(output, 'results.json'), JSON.stringify({ kind: 'PREPARATION_ONLY', campaign: manifest.campaign,
    date: new Date().toISOString(), cases, blocked: true, step }, null, 2) + '\n', { mode: 0o600 })
  writeFileSync(join(output, 'error.local'), step + '\n' + String(error.stack), { mode: 0o600 })
  if (lastPage && !lastPage.isClosed()) await lastPage.screenshot({ path: join(output, 'failure-private.png') }).catch(() => {})
  // No imprimir mensaje bruto de Playwright: puede contener valores de formularios.
  console.error('Smoke de preparación falló; casos completados: ' + cases.length + '. Revisar solo salida privada.')
  process.exitCode = 1
} finally { if (browser) await browser.close() }
