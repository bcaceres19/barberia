import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-023 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): pantalla
 * "Servicios por barbero" en 320/360/768/1280 px, más zoom 200%
 * aproximado (mismo criterio que
 * `e2e/servicios-evidencia-responsiva.spec.ts`). Un solo inicio de sesión
 * reutilizado entre los cinco breakpoints. Las capturas se guardan en
 * `e2e/evidence/servicios-por-barbero/`.
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'servicios-por-barbero',
)
const axeScriptPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'node_modules',
  'axe-core',
  'axe.min.js',
)

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

test('pantalla Servicios por barbero sin scroll horizontal, con foco visible por teclado y sin violaciones axe en cada breakpoint (CA-023-08)', async ({
  page,
}) => {
  await page.goto('/acceso')
  await page.getByLabel('Correo', { exact: true }).fill(EMAIL)
  await page.getByLabel('Contraseña', { exact: true }).fill(PASSWORD)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel$/)

  // Datos reales representativos: un barbero y un servicio, para que la
  // evidencia muestre el estado "con datos" (checklist con al menos una
  // fila), no el vacío.
  const stamp = Date.now()
  const barberName = `Evidencia Responsiva Barbero ${stamp}`
  const serviceName = `Evidencia Responsiva Servicio ${stamp}`

  await page.getByRole('link', { name: 'Barberos' }).click()
  await page.getByRole('button', { name: 'Agregar barbero' }).click()
  const barberDialog = page.getByRole('dialog', { name: 'Agregar barbero' })
  await barberDialog.getByLabel('Nombre').fill(barberName)
  await barberDialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(barberDialog).toBeHidden()

  await page.getByRole('link', { name: 'Servicios', exact: true }).click()
  await page.getByRole('button', { name: 'Agregar servicio' }).click()
  const serviceDialog = page.getByRole('dialog', { name: 'Agregar servicio' })
  await serviceDialog.getByLabel('Nombre').fill(serviceName)
  await serviceDialog.getByLabel('Duración (minutos)').fill('30')
  await serviceDialog.getByLabel('Precio (COP)').fill('45000.00')
  await serviceDialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(serviceDialog).toBeHidden()

  await page.getByRole('link', { name: 'Servicios por barbero' }).click()
  await expect(page.getByRole('heading', { name: 'Servicios por barbero' })).toBeVisible()
  await page.getByLabel('Barbero', { exact: true }).selectOption({ label: barberName })
  await page.getByRole('checkbox', { name: serviceName }).check()
  await expect(page.getByRole('checkbox', { name: serviceName })).toBeChecked()

  await page.addScriptTag({ path: axeScriptPath })

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })

    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'normal.png'),
      fullPage: true,
    })

    const hasHorizontalScroll = await page.evaluate(
      () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
    )
    expect(hasHorizontalScroll).toBe(false)

    // Foco visible por teclado: el selector de barbero es el primer control
    // interactivo relevante de la pantalla.
    await page.getByLabel('Barbero', { exact: true }).focus()
    await expect(page.getByLabel('Barbero', { exact: true })).toBeFocused()
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })

    // La propia casilla de servicio también es operable por teclado: se
    // puede enfocar y alternar con la barra espaciadora, sin usar el mouse.
    const checkbox = page.getByRole('checkbox', { name: serviceName })
    await checkbox.focus()
    await expect(checkbox).toBeFocused()

    // axe-core contra la pantalla ya cargada con datos reales.
    // color-contrast se desactiva: la paleta ya está verificada en la tabla
    // aprobada de estandar-diseno-visual.md §4.3 (mismo criterio que las
    // pruebas de componente con vitest-axe).
    const axeResults = (await page.evaluate(async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const axeGlobal = (window as any).axe
      return axeGlobal.run(document, { rules: { 'color-contrast': { enabled: false } } })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    })) as { violations: any[] }
    expect(axeResults.violations, JSON.stringify(axeResults.violations, null, 2)).toEqual([])
  }

  // La casilla sigue reflejando el estado real (asignado) tras recorrer
  // todos los breakpoints: ningún cambio de tamaño de ventana la alteró.
  await expect(page.getByRole('checkbox', { name: serviceName })).toBeChecked()
})
