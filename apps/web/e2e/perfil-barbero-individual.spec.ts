import { test, expect, type Page, type BrowserContext } from '@playwright/test'

// #294 (DEC-115): perfil «barbero individual» contra el API y PostgreSQL reales.
// Datos dedicados en database/testdata/ui_barbero_individual_294.sql: una barbería
// con UN solo barbero, perfil `solo`, tres de cuatro servicios ofrecidos y tres
// turnos hoy. No comparte registros con las otras suites. Apagar «Lo ofrezco» retira
// la última asignación de un servicio activo, que solo es posible con DEC-114: por eso
// estas pruebas encienden y comprueban, pero no apagan contra el API real (el apagado
// y su rechazo se cubren con dobles en CatalogWorkspacePage.test.ts).
const BARBER = '294b0001-0000-4000-8000-000000000001'
const EMAIL = 'mateo.demo@ejemplo.test'
const PASSWORD = 'Nava-Demo-2026!'

// Las pruebas comparten UNA sesión y una barbería: cada una deja el perfil como lo encontró.
test.describe.configure({ mode: 'serial' })

// UNA sesión y UNA página para todo el archivo: cada prueba reutiliza lo que dejó la anterior
// (modo serial) y el acceso se pide una sola vez, para no agotar el umbral de HU-007.
let context: BrowserContext
let page: Page

test.beforeAll(async ({ browser }, testInfo) => {
  context = await browser.newContext({ baseURL: testInfo.project.use.baseURL })
  const response = await context.request.post('/api/v1/public/auth/login', {
    data: { email: EMAIL, password: PASSWORD },
  })
  expect(response.status()).toBe(200)
  page = await context.newPage()
})

// Aunque una prueba falle a mitad del cambio de perfil, la barbería vuelve a `solo` para la
// siguiente ejecución (el servidor conserva lo demás de la marca si se omite).
test.afterAll(async () => {
  await page.goto('/panel')
  await page.evaluate(async () => {
    const current = await (await fetch('/api/v1/private/settings/brand')).json()
    await fetch('/api/v1/private/settings/brand', {
      method: 'PATCH',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify({ ...current, panelProfile: 'solo' }),
    })
  })
  await context.close()
})

async function desktopDock() {
  return page
    .locator('.app-nav__list--desktop .app-nav__label')
    .evaluateAll((labels) => labels.map((l) => l.textContent?.trim()))
}

// Consulta al API con la cookie de la propia página: es `Secure` y el cliente de peticiones
// de Node no la envía a http://127.0.0.1, mientras que el navegador sí.
async function apiGet<T>(url: string): Promise<{ status: number; body: T }> {
  return page.evaluate(async (target) => {
    const response = await fetch(target, { credentials: 'same-origin' })
    return { status: response.status, body: await response.json() }
  }, url)
}

async function offeredServiceNames() {
  const assigned = await apiGet<{ items: { serviceId: string }[] }>(
    `/api/v1/private/barbers/${BARBER}/services`,
  )
  expect(assigned.status).toBe(200)
  const ids = new Set(assigned.body.items.map((i) => i.serviceId))
  const all = await apiGet<{ items: { id: string; name: string }[] }>(
    '/api/v1/private/services?pageSize=50',
  )
  return all.body.items.filter((s) => ids.has(s.id)).map((s) => s.name)
}

test('el dock del barbero individual solo trae lo necesario', async () => {
  await page.goto('/panel')
  await expect(page.getByRole('heading', { name: 'Agenda', level: 1 })).toBeVisible()

  expect(await desktopDock()).toEqual([
    'Agenda',
    'Servicios',
    'Mi perfil',
    'Horarios',
    'Configuración',
    'Reserva pública',
  ])
  // Sin gestión de equipo ni la matriz «Servicios por barbero».
  await expect(page.getByRole('link', { name: 'Barberos' })).toHaveCount(0)
  await expect(page.getByRole('link', { name: 'Servicios por barbero' })).toHaveCount(0)
})

test('una ruta retirada por el perfil lleva a Servicios, no a una pantalla en blanco', async () => {
  await page.goto('/panel/servicios-por-barbero')

  await expect(page).toHaveURL(/\/panel\/servicios$/)
  await expect(page.getByRole('heading', { name: 'Servicios', level: 1 })).toBeVisible()
})

test('la agenda no pide elegir barbero y dice qué queda del día', async () => {
  await page.goto('/panel')
  await expect(page.getByRole('heading', { name: 'Agenda', level: 1 })).toBeVisible()

  await expect(page.getByLabel('Barbero', { exact: true })).toHaveCount(0)
  // La selección sigue en la URL (DEC-074): solo desaparece el control.
  await expect(page).toHaveURL(new RegExp(`barberId=${BARBER}`))
  // En móvil la línea temporal de escritorio está oculta: se busca el turno visible.
  await expect(
    page.getByText('Cliente de prueba Uno').locator('visible=true').first(),
  ).toBeVisible()
  // Los turnos del fixture son de hoy; según la hora quedan algunos o ninguno.
  await expect(page.locator('.daily-agenda-page__summary')).toContainText(
    /turnos? por atender|No te quedan/,
  )
})

test('Nuevo turno no pide elegir barbero y ya lista sus servicios', async () => {
  await page.goto('/panel/turnos/nuevo')
  await expect(page.getByRole('heading', { name: 'Nuevo turno', level: 1 })).toBeVisible()

  await expect(page.locator('#new-appointment-barber')).toHaveCount(0)
  const options = page.locator('#new-appointment-service option')
  await expect(options.filter({ hasText: 'Corte clásico' })).toHaveCount(1)
  await expect(options.filter({ hasText: 'Barba perfilada' })).toHaveCount(1)
})

test('Horarios no pide elegir barbero y muestra su horario', async () => {
  await page.goto('/panel/horarios')

  await expect(page.locator('#schedules-barber-select')).toHaveCount(0)
  await expect(page.getByText('09:00').first()).toBeVisible()
})

test('Mi perfil es la ficha del único barbero, sin controles de equipo', async () => {
  await page.goto('/panel/barberos')

  await expect(page.getByRole('heading', { name: 'Mi perfil', level: 1 })).toBeVisible()
  await expect(page.getByText('Mateo de prueba').first()).toBeVisible()
  await expect(page.getByRole('button', { name: /Agregar barbero/ })).toHaveCount(0)
  await expect(page.getByRole('navigation', { name: /Paginación/ })).toHaveCount(0)
  await expect(page.getByRole('button', { name: 'Bloquear a Mateo de prueba' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Editar Mateo de prueba' })).toBeVisible()
})

// La tabla ajusta cuántas filas caben al alto de la ventana; el buscador de la propia
// pantalla lleva directo al servicio sin depender de ese ajuste.
async function openServiceBySearch(name: string) {
  await page.goto('/panel/servicios')
  await page.getByRole('searchbox', { name: 'Buscar' }).fill(name)
  return page.getByRole('switch', { name: `Lo ofrezco: ${name}` })
}

test('«Lo ofrezco» refleja y cambia la asignación real del barbero', async () => {
  await page.goto('/panel/servicios')
  await expect(page.getByRole('switch', { name: 'Lo ofrezco: Corte clásico' })).toHaveAttribute(
    'aria-checked',
    'true',
  )

  const design = await openServiceBySearch('Diseño con navaja')
  await expect(design).toBeVisible()
  if ((await design.getAttribute('aria-checked')) === 'false') {
    await design.click()
    await expect(design).toHaveAttribute('aria-checked', 'true')
    await expect(page.getByText('Ahora lo ofreces')).toBeVisible()
  }
  // Lo que se ve es lo guardado: una recarga y el API real coinciden.
  await expect(await openServiceBySearch('Diseño con navaja')).toHaveAttribute(
    'aria-checked',
    'true',
  )
  expect(await offeredServiceNames()).toContain('Diseño con navaja')
})

test('un servicio nuevo queda ofrecido sin pasar por la matriz', async () => {
  const name = `E2E ${Date.now().toString(36)}`
  await page.goto('/panel/servicios')
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const dialog = page.getByRole('dialog')
  await dialog.getByLabel('Nombre').fill(name)
  await dialog.getByLabel(/Duración/).fill('25')
  await dialog.getByLabel(/Precio/).fill('18000')
  await dialog
    .getByRole('button', { name: /Crear|Guardar|Agregar/ })
    .last()
    .click()

  await expect(page.getByText('Servicio creado')).toBeVisible()
  await expect(page.getByRole('switch', { name: `Lo ofrezco: ${name}` })).toHaveAttribute(
    'aria-checked',
    'true',
  )
  expect(await offeredServiceNames()).toContain(name)
})

// Las entradas del dock de escritorio siguen en el DOM en móvil (solo se ocultan con CSS):
// se comprueba la presencia de cada entrada, no su visibilidad.
const desktopLabel = (text: string) =>
  page.locator('.app-nav__list--desktop .app-nav__label', { hasText: text })

test('cambiar a «Barbería con equipo» devuelve las pantallas de equipo y se puede volver', async () => {
  await page.goto('/panel/barberia')
  await expect(page.getByRole('heading', { name: 'Configuración', level: 1 })).toBeVisible()
  const group = page.getByRole('radiogroup', { name: 'Perfil del panel' })
  await expect(group.getByRole('radio', { name: /Barbero individual/ })).toBeChecked()

  await group.getByRole('radio', { name: /Barbería con equipo/ }).check()
  await page.getByRole('button', { name: /Guardar/ }).click()
  await expect(desktopLabel('Servicios por barbero')).toHaveCount(1)
  await expect(desktopLabel('Barberos')).toHaveCount(1)
  // Ningún dato cambió: sigue siendo la misma barbería con el mismo barbero.
  const barbers = await apiGet<{ items: unknown[] }>('/api/v1/private/barbers')
  expect(barbers.body.items).toHaveLength(1)

  await group.getByRole('radio', { name: /Barbero individual/ }).check()
  await page.getByRole('button', { name: /Guardar/ }).click()
  await expect(desktopLabel('Mi perfil')).toHaveCount(1)
  await expect(desktopLabel('Servicios por barbero')).toHaveCount(0)
  const brand = await apiGet<{ panelProfile: string }>('/api/v1/private/settings/brand')
  expect(brand.body.panelProfile).toBe('solo')
})
