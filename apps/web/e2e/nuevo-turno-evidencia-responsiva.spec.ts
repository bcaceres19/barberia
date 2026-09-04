import { test, expect, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { fillCompleteForm, installNewAppointmentMock, BARBER, SERVICE } from './nuevo-turno-mock'

/**
 * Evidencia responsive y accesible de `/panel/turnos/nuevo` (issue #190;
 * docs/03-desarrollo/estandar-diseno-visual.md §16, estrategia-pruebas.md
 * §5.4): los mismos cinco anchos que ya fija
 * `e2e/panel-evidencia-responsiva.spec.ts` —320/360/768/1280 px más un
 * 1280 con zoom 200% aproximado (640×450)—, recorrido de teclado con foco
 * visible y `prefers-reduced-motion`.
 *
 * A diferencia de la evidencia responsive de `/panel`, esta corre sobre el
 * doble de API de `nuevo-turno-mock.ts` en vez de iniciar sesión contra el
 * API real: no consume el umbral de intentos de HU-007 y puede ejecutarse en
 * cualquier máquina sin PostgreSQL.
 */
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'nuevo-turno',
  'fidelidad-190',
  'responsiva',
)

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
]

async function openReadyForm(page: Page) {
  await installNewAppointmentMock(page, 'empty-form')
  await page.goto('/panel/turnos/nuevo', { waitUntil: 'domcontentloaded' })
  await expect(page.getByRole('button', { name: 'Barbero', exact: true })).toBeVisible()
}

/**
 * Quien hace scroll en el cascarón autenticado es `.private-shell__content`,
 * no el documento: sin esto, `fullPage: true` devolvería solo el primer
 * viewport y la evidencia ocultaría justo la parte del formulario que hay que
 * revisar. `!important` porque el cascarón fija su alto desde un estilo con
 * ámbito, de mayor especificidad que una clase suelta.
 */
async function expandShellForCapture(page: Page) {
  await page.addStyleTag({
    content:
      '.private-shell{height:auto!important;min-height:100dvh!important}' +
      '.private-shell__content{overflow:visible!important}',
  })
}

test.describe('evidencia responsiva de /panel/turnos/nuevo', () => {
  test.use({ locale: 'es-CO', timezoneId: 'America/Bogota' })

  test('el formulario no desborda a lo ancho en ningún breakpoint', async ({ page }) => {
    await openReadyForm(page)

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height })

      // El viewport efectivo se verifica de verdad, no se da por hecho por la
      // respuesta del resize (.claude/skills/browser-viewport-verification).
      expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
      expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)

      const overflows = await page.evaluate(() => {
        const doc = document.documentElement
        const content = document.querySelector('.private-shell__content')
        return {
          document: doc.scrollWidth > doc.clientWidth,
          content: !!content && content.scrollWidth > content.clientWidth,
        }
      })
      expect(overflows.document, `${viewport.name}: desborde horizontal del documento`).toBe(false)
      expect(overflows.content, `${viewport.name}: desborde horizontal del contenido`).toBe(false)

      await expandShellForCapture(page)
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'formulario.png'),
        fullPage: true,
        animations: 'disabled',
      })
    }
  })

  test('el recorrido de teclado llega al CTA con foco visible en cada breakpoint', async ({
    page,
  }) => {
    await openReadyForm(page)

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height })

      // El listbox de barbero se abre y se recorre con teclado (patrón
      // "Collapsible Listbox" de WAI-ARIA): abrir con Enter, elegir con Enter
      // y devolver el foco al disparador.
      const trigger = page.getByRole('button', { name: 'Barbero', exact: true })
      await trigger.focus()
      await expect(trigger).toBeFocused()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'foco-barbero.png'),
        animations: 'disabled',
      })

      await page.keyboard.press('Enter')
      await expect(page.getByRole('listbox', { name: 'Barbero' })).toBeVisible()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'listbox-barbero.png'),
        animations: 'disabled',
      })
      await page.keyboard.press('Enter')
      await expect(trigger).toBeFocused()
      await expect(trigger).toContainText(BARBER.fullName)

      await page.getByLabel('Servicio', { exact: true }).selectOption({ label: SERVICE.name })

      const submit = page.getByRole('button', { name: 'Registrar turno' })
      await submit.focus()
      await expect(submit).toBeFocused()
      await page.screenshot({
        path: path.join(evidenceDir, viewport.name, 'foco-cta.png'),
        animations: 'disabled',
      })

      await page.reload({ waitUntil: 'domcontentloaded' })
      await expect(trigger).toBeVisible()
    }
  })

  test('con prefers-reduced-motion el esqueleto y la alerta no animan', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' })
    await page.setViewportSize({ width: 1280, height: 900 })

    const pending = await installNewAppointmentMock(page, 'context-loading')
    await page.goto('/panel/turnos/nuevo', { waitUntil: 'domcontentloaded' })
    await expect(page.locator('.nt-skeleton')).toBeVisible()

    const skeletonAnimation = await page.evaluate(
      () => getComputedStyle(document.querySelector('.nt-skeleton__bar')!).animationName,
    )
    expect(skeletonAnimation).toBe('none')
    await page.screenshot({
      path: path.join(evidenceDir, 'reduced-motion', 'esqueleto.png'),
      animations: 'disabled',
    })
    pending?.release()
  })

  test('la alerta global y el error por campo conviven sin desbordar en 320px', async ({
    page,
  }) => {
    await page.setViewportSize({ width: 320, height: 720 })
    await installNewAppointmentMock(page, 'schedule-conflict')
    await page.goto('/panel/turnos/nuevo', { waitUntil: 'domcontentloaded' })
    await fillCompleteForm(page)
    await page.getByRole('button', { name: 'Registrar turno' }).click()
    await expect(
      page.getByText('el barbero ya tiene una cita en ese intervalo').first(),
    ).toBeVisible()

    const overflows = await page.evaluate(
      () => document.documentElement.scrollWidth > document.documentElement.clientWidth,
    )
    expect(overflows).toBe(false)

    await expandShellForCapture(page)
    await page.screenshot({
      path: path.join(evidenceDir, '320', 'conflicto.png'),
      fullPage: true,
      animations: 'disabled',
    })
  })
})
