import { test, expect } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de HU-021 (docs/03-desarrollo/
 * estandar-diseno-visual.md §16 y estrategia-pruebas.md §5.4): pantalla
 * "Barberos" en 320/360/768/1280 px, más zoom 200% aproximado (mismo
 * criterio que `e2e/configuracion-barberia-evidencia-responsiva.spec.ts`).
 * Un solo inicio de sesión reutilizado entre los cinco breakpoints: el
 * umbral de intentos de HU-007 es compartido por todo este entorno de
 * pruebas. Las capturas se guardan en `e2e/evidence/barberos/`.
 */
const EMAIL = process.env.E2E_EMAIL ?? 'duena.a@ejemplo.test'
const PASSWORD = process.env.E2E_PASSWORD ?? 'ClaveDePruebaHU010!'

const evidenceDir = path.join(path.dirname(fileURLToPath(import.meta.url)), 'evidence', 'barberos')
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

test('pantalla Barberos sin scroll horizontal, con foco visible por teclado y sin violaciones axe en cada breakpoint (CA-021-08)', async ({
  page,
}) => {
  await page.goto('/acceso')
  await page.getByLabel('Correo').fill(EMAIL)
  await page.getByLabel('Contraseña').fill(PASSWORD)
  await page.getByRole('button', { name: 'Iniciar sesión' }).click()
  await expect(page).toHaveURL(/\/panel$/)

  await page.getByRole('link', { name: 'Barberos' }).click()
  await expect(page.getByRole('heading', { name: 'Barberos' })).toBeVisible()

  // Asegura al menos un barbero real visible en todos los breakpoints
  // (evidencia representativa del estado "con datos", no del vacío).
  const stamp = Date.now()
  const seedName = `Evidencia Responsiva ${stamp}`
  await page.getByRole('button', { name: 'Agregar barbero' }).click()
  const dialog = page.getByRole('dialog', { name: 'Agregar barbero' })
  await dialog.getByLabel('Nombre').fill(seedName)
  await dialog.getByRole('button', { name: 'Guardar' }).click()
  await expect(dialog).toBeHidden()
  await expect(page.getByText(seedName)).toBeVisible()

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

    // Foco visible por teclado: el botón "Agregar barbero" es el primer
    // control interactivo relevante de la pantalla.
    await page.getByRole('button', { name: 'Agregar barbero' }).focus()
    await expect(page.getByRole('button', { name: 'Agregar barbero' })).toBeFocused()
    await page.screenshot({ path: path.join(evidenceDir, viewport.name, 'foco.png') })

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

  // Diálogo de alta también inspeccionado (accesible, sin scroll horizontal,
  // foco atrapado en el primer campo) al menos en el viewport móvil más
  // angosto y en escritorio.
  for (const viewport of [viewports[0], viewports[3]]) {
    await page.setViewportSize({ width: viewport!.width, height: viewport!.height })
    await page.getByRole('button', { name: 'Agregar barbero' }).click()
    const createDialog = page.getByRole('dialog', { name: 'Agregar barbero' })
    await expect(createDialog).toBeVisible()
    // BaseDialog atrapa el foco dentro del diálogo (shared/ui/BaseDialog.vue,
    // trapFocus, tras un nextTick): se confirma que el foco quedó DENTRO
    // del diálogo, en cualquiera de sus controles, no fuera de él.
    await expect
      .poll(async () =>
        page.evaluate((dialogSelector) => {
          const dialog = document.querySelector(dialogSelector)
          return !!dialog && dialog.contains(document.activeElement)
        }, '.base-dialog--open'),
      )
      .toBe(true)

    const axeResults = (await page.evaluate(async () => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const axeGlobal = (window as any).axe
      return axeGlobal.run(document, { rules: { 'color-contrast': { enabled: false } } })
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    })) as { violations: any[] }
    expect(axeResults.violations, JSON.stringify(axeResults.violations, null, 2)).toEqual([])

    await createDialog.getByRole('button', { name: 'Cancelar' }).click()
    await expect(createDialog).toBeHidden()
  }
})
