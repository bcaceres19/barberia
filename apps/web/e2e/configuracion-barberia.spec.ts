import { test, expect, type Browser, type BrowserContext, type Page } from '@playwright/test'

/**
 * Recorrido E2E de Configuración (HU-020, DEC-110): datos básicos, marca,
 * vocabulario, zona horaria y preferencias de este dispositivo. Corre contra
 * el API real en local (`go run ./cmd/api` sobre PostgreSQL real con todas las
 * migraciones aplicadas y usuarios con un hash argon2id REAL, mismo criterio
 * que `e2e/panel.spec.ts`).
 *
 * Variables de entorno:
 *   E2E_EMAIL / E2E_PASSWORD: credenciales de la barbería A.
 *   E2E_EMAIL_B / E2E_PASSWORD_B: credenciales de la barbería B (aislamiento).
 *   APP_BASE_URL: base del frontend (por defecto http://localhost:5173).
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'
const EMAIL_B = process.env.E2E_EMAIL_B ?? 'dueno.b@ejemplo.test'
const PASSWORD_B = process.env.E2E_PASSWORD_B ?? 'ClaveDePruebaHU010!'

async function login(page: Page, email = EMAIL, password = PASSWORD) {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(email)
  await page.getByLabel('Contraseña', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel(\?.*)?$/)
}

async function openSettings(page: Page) {
  await page.getByRole('link', { name: 'Configuración', exact: true }).click()
  await expect(page).toHaveURL(/\/panel\/barberia$/)
  await expect(page.getByRole('heading', { name: 'Configuración', level: 1 })).toBeVisible()
  await expect(page.getByLabel('Nombre', { exact: true })).toBeVisible()
}

async function save(page: Page) {
  await page.getByRole('button', { name: 'Guardar cambios' }).click()
  await expect(page.getByText('Configuración guardada').first()).toBeVisible()
  await expect(page.getByText('Cambios sin guardar')).toBeHidden()
}

// Devuelve la barbería A a su marca inicial para que las corridas no dependan
// del estado que dejó la anterior.
async function restoreInitialBrand(page: Page) {
  await page.getByRole('button', { name: 'barbería', exact: true }).click()
  await page.getByRole('button', { name: 'barbero', exact: true }).click()
  await page.getByRole('radio', { name: /Latón/ }).check()
  if (await page.getByText('Cambios sin guardar').isVisible()) await save(page)
}

test.describe('Configuración (HU-020, DEC-110)', () => {
  // Serial a propósito: todas las pruebas escriben sobre la MISMA fila real de
  // barbershop (barbería única de prueba); ejecutarlas en paralelo produciría
  // una carrera entre sus propias escrituras, no un defecto del producto.
  test.describe.configure({ mode: 'serial' })
  // El dock móvil agrupa Configuración dentro de "Más"; la evidencia móvil
  // vive en configuracion-barberia-evidencia-responsiva.spec.ts.
  test.skip(({ isMobile }) => isMobile, 'recorrido de escritorio')

  // Una sola sesión para todo el recorrido: el umbral de intentos de HU-007 es
  // compartido por todo el entorno de pruebas y un acceso por prueba lo agota.
  let context: BrowserContext
  let page: Page

  test.beforeAll(async ({ browser }: { browser: Browser }) => {
    context = await browser.newContext()
    page = await context.newPage()
    await login(page)
  })

  test.afterAll(async () => {
    await context.close()
  })

  test('editar y guardar los datos básicos actualiza la cabecera y persiste tras recargar (CA-020-01, CA-020-02)', async () => {
    await openSettings(page)

    // Nombre único por corrida: evita depender del estado de una ejecución anterior.
    const newName = `Barbería E2E ${Date.now()}`
    await page.getByLabel('Nombre', { exact: true }).fill(newName)
    await page.getByLabel('Correo de contacto').fill('contacto-e2e@ejemplo.test')
    await page.getByLabel('Teléfono de contacto').fill('+573009998877')

    await save(page)

    // La cabecera refleja el nuevo nombre SIN recargar la aplicación (CA-020-02).
    await expect(page.getByTestId('barbershop-name')).toHaveText(newName)

    // Persistencia real: una recarga vuelve a consultar el servidor.
    await page.reload()
    await expect(page.getByTestId('barbershop-name')).toHaveText(newName)
    await expect(page.getByLabel('Nombre', { exact: true })).toHaveValue(newName)
    await expect(page.getByLabel('Correo de contacto')).toHaveValue('contacto-e2e@ejemplo.test')
    await expect(page.getByLabel('Teléfono de contacto')).toHaveValue('+573009998877')
  })

  test('la zona se elige con el buscador y el reloj muestra la hora de esa zona; descartar la devuelve (CA-020-03, RN-DIS-07)', async () => {
    await openSettings(page)

    const zone = page.getByRole('combobox', { name: 'Zona horaria' })
    const before = await zone.inputValue()
    await zone.click()
    await zone.fill('madrid')
    await page.getByRole('option', { name: /Europe\/Madrid/ }).click()

    await expect(zone).toHaveValue('Europe/Madrid')
    await expect(page.getByText('Hora en Madrid')).toBeVisible()
    await expect(page.getByText('Cambios sin guardar')).toBeVisible()

    await page.getByRole('button', { name: 'Descartar' }).click()
    await expect(zone).toHaveValue(before)
    await expect(page.getByText('Cambios sin guardar')).toBeHidden()
  })

  test('una barra de guardado solo aparece con cambios, y un valor inválido se explica sin escribir (CA-020-08)', async () => {
    await openSettings(page)
    await expect(page.getByText('Cambios sin guardar')).toBeHidden()

    const before = await page.getByLabel('Nombre', { exact: true }).inputValue()
    await page.getByLabel('Nombre', { exact: true }).fill('   ')
    await expect(page.getByText('Cambios sin guardar')).toBeVisible()
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Escribe el nombre de la barbería.')).toBeVisible()

    await page.getByLabel('Cómo llamas a tu negocio').fill('salón 24')
    await page.getByRole('button', { name: 'Guardar cambios' }).click()
    await expect(page.getByText('Usa solo letras, espacios, guion o apóstrofo.')).toBeVisible()

    // Cero escritura: recargar muestra lo de ANTES del intento.
    await page.reload()
    await expect(page.getByLabel('Nombre', { exact: true })).toHaveValue(before)
  })

  test('la marca y el vocabulario se guardan, cambian el panel entero y persisten tras recargar (DEC-110)', async () => {
    await openSettings(page)
    await restoreInitialBrand(page)

    await page.getByRole('radio', { name: /Esmeralda/ }).check()
    // La vista previa es inmediata, antes de guardar.
    await expect
      .poll(() =>
        page.evaluate(() =>
          document.documentElement.style.getPropertyValue('--color-brand-accent-surface'),
        ),
      )
      .toBe('#4cb58a')

    await page.getByRole('button', { name: 'salón de belleza', exact: true }).click()
    await page.getByRole('button', { name: 'estilista', exact: true }).click()
    await expect(page.getByText('Elige a la estilista y el servicio.')).toBeVisible()
    await save(page)

    // El vocabulario llega a la navegación y a las demás pantallas, con su concordancia.
    await expect(page.getByRole('link', { name: 'Estilistas' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Barberos', exact: true })).toHaveCount(0)
    await page.getByRole('link', { name: 'Estilistas' }).click()
    await expect(page.getByRole('heading', { name: 'Estilistas', level: 1 })).toBeVisible()

    // Persistencia real tras recargar, incluido el acento.
    await page.reload()
    await expect(page.getByRole('heading', { name: 'Estilistas', level: 1 })).toBeVisible()
    expect(
      await page.evaluate(() =>
        document.documentElement.style.getPropertyValue('--color-brand-accent-surface'),
      ),
    ).toBe('#4cb58a')

    await openSettings(page)
    await expect(page.getByLabel('Cómo llamas a tu negocio')).toHaveValue('salón de belleza')
    await expect(page.getByRole('radio', { name: /Esmeralda/ })).toBeChecked()

    await restoreInitialBrand(page)
    await expect(page.getByRole('link', { name: 'Barberos', exact: true })).toBeVisible()
  })

  test('la marca de una barbería nunca llega a otra (aislamiento entre tenants)', async ({
    browser,
  }) => {
    await openSettings(page)
    await restoreInitialBrand(page)
    await page.getByRole('radio', { name: /Rubí/ }).check()
    await page.getByRole('button', { name: 'peluquera', exact: true }).click()
    await save(page)

    const other = await browser.newContext()
    const pageB = await other.newPage()
    await login(pageB, EMAIL_B, PASSWORD_B)
    await expect(pageB.getByRole('link', { name: 'Barberos', exact: true })).toBeVisible()
    await expect(pageB.getByRole('link', { name: 'Peluqueras' })).toHaveCount(0)
    expect(
      await pageB.evaluate(() =>
        document.documentElement.style.getPropertyValue('--color-brand-accent-surface'),
      ),
    ).not.toBe('#e8808c')
    await other.close()

    await restoreInitialBrand(page)
  })

  test('el enlace público es único, estable, se copia y cada uno abre solo su barbería (DEC-117)', async ({
    browser,
  }) => {
    await openSettings(page)
    const linkField = page.getByRole('textbox', { name: 'Tu enlace de reservas' })
    await expect(linkField).toHaveText(/\/reservar\/[a-z0-9-]+$/)
    const linkA = (await linkField.innerText()).trim()
    // Un slug generado ahora lleva el código aleatorio; uno anterior a DEC-117 no.
    expect(linkA).toMatch(/\/reservar\/[a-z0-9][a-z0-9-]*[a-z0-9]$/)

    // Estable: otra lectura (recarga) devuelve exactamente el mismo enlace.
    await page.reload()
    await expect(linkField).toHaveText(linkA)

    // Copiar entrega el mismo texto que se ve.
    await page.context().grantPermissions(['clipboard-read', 'clipboard-write'])
    await page.getByRole('button', { name: 'Copiar enlace' }).click()
    await expect(page.getByText('Enlace copiado.')).toBeVisible()
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(linkA)

    // Sin sesión, el enlace abre la reserva de ESTA barbería.
    const nameA = (await page.getByTestId('barbershop-name').innerText()).trim()
    const visitor = await browser.newContext()
    const visitorPage = await visitor.newPage()
    await visitorPage.goto(linkA)
    await expect(visitorPage.getByRole('heading', { level: 1 })).toHaveText(nameA)

    // La barbería B recibe otro enlace, y el de A sigue mostrando a A.
    const other = await browser.newContext()
    const pageB = await other.newPage()
    await login(pageB, EMAIL_B, PASSWORD_B)
    await openSettings(pageB)
    const linkB = (
      await pageB.getByRole('textbox', { name: 'Tu enlace de reservas' }).innerText()
    ).trim()
    expect(linkB).not.toBe(linkA)
    const nameB = (await pageB.getByTestId('barbershop-name').innerText()).trim()
    expect(nameB).not.toBe(nameA)
    await visitorPage.goto(linkB)
    await expect(visitorPage.getByRole('heading', { level: 1 })).toHaveText(nameB)
    await visitorPage.goto(linkA)
    await expect(visitorPage.getByRole('heading', { level: 1 })).toHaveText(nameA)

    await other.close()
    await visitor.close()
  })

  test('las preferencias de este dispositivo se aplican al instante, persisten y no tocan el acceso (DEC-110)', async () => {
    await openSettings(page)

    await page.getByRole('radio', { name: /Marfil/ }).check()
    await expect(page.locator('html')).toHaveAttribute('data-app-theme', 'ivory')
    await page.getByRole('radio', { name: 'Muy grande' }).check()
    await page.getByRole('switch', { name: 'Reducir animaciones' }).click()
    await expect(page.locator('html')).toHaveAttribute('data-motion', 'reduced')
    // No hay nada que guardar: son del dispositivo.
    await expect(page.getByText('Cambios sin guardar')).toBeHidden()

    await page.reload()
    await expect(page.locator('html')).toHaveAttribute('data-app-theme', 'ivory')
    await expect(page.locator('html')).toHaveAttribute('data-motion', 'reduced')

    // El acceso conserva su propio diseño: el tema del panel no se filtra.
    await page.getByRole('button', { name: 'Cerrar sesión' }).click()
    await expect(page).toHaveURL(/\/acceso$/)
    await expect(page.locator('html')).not.toHaveAttribute('data-app-theme', /.+/)

    await login(page)
    await openSettings(page)
    await page.getByRole('button', { name: 'Restablecer pantalla' }).click()
    await expect(page.locator('html')).toHaveAttribute('data-app-theme', 'ink')
  })
})
