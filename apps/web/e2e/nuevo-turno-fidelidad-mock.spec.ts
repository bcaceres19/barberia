import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  chooseBarber,
  fillCompleteForm,
  fillCustomerAndSchedule,
  installNewAppointmentMock,
  type Scenario,
} from './nuevo-turno-mock'

/**
 * Evidencia visual del issue #190 (fidelidad de `/panel/turnos/nuevo` con el
 * atlas `nuevo-turno-eventos`, segundo pase del 2026-09-04). Mismo criterio
 * que `panel-fidelidad-mock.spec.ts`: los datos del doble
 * (`nuevo-turno-mock.ts`) existen SOLO dentro de Playwright, así que cada uno
 * de los doce eventos del atlas se puede revisar sin iniciar sesión ni tocar
 * API, PostgreSQL, OTP o rate limiting locales.
 *
 * Un escenario por evento, con el MISMO número y nombre de archivo que el
 * atlas, capturado en los dos viewports de referencia (1440×1024 y
 * 420×935 @2x). La comparación contra la lámina (lado a lado y diff) la
 * produce `nuevo-turno-fidelidad-comparacion.mjs` a partir de estas capturas.
 */
const evidenceRoot = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'nuevo-turno',
  'fidelidad-190',
  'mock',
)

async function driveScenario(page: Page, scenario: Scenario) {
  switch (scenario) {
    case 'context-loading':
      await expect(page.locator('.nt-skeleton')).toBeVisible()
      return
    case 'context-error':
      await expect(page.getByText('No pudimos cargar esta sección')).toBeVisible()
      return
    case 'no-barbers':
      await expect(page.getByText('Aún no tienes barberos registrados')).toBeVisible()
      return
    case 'empty-form':
      await expect(page.getByRole('button', { name: 'Barbero', exact: true })).toBeVisible()
      return
    case 'services-loading':
      await chooseBarber(page)
      await expect(page.getByText('Cargando servicios')).toBeVisible()
      return
    case 'services-empty':
      await chooseBarber(page)
      await expect(
        page.getByText('Este barbero no tiene servicios activos asignados.'),
      ).toBeVisible()
      return
    case 'services-error':
      await chooseBarber(page)
      await expect(page.getByText('No pudimos cargar los servicios de este barbero.')).toBeVisible()
      return
    case 'filled-form':
      await fillCompleteForm(page)
      await expect(page.locator('.new-appointment-page__resumen')).toBeVisible()
      return
    case 'validation-error':
      await fillCustomerAndSchedule(page)
      await page.getByRole('button', { name: 'Registrar turno' }).click()
      await expect(page.getByText('Elige un barbero.')).toBeVisible()
      return
    case 'saving':
      await fillCompleteForm(page)
      await page.getByRole('button', { name: 'Registrar turno' }).click()
      await expect(page.getByRole('button', { name: 'Guardando…' })).toBeDisabled()
      return
    case 'schedule-conflict':
      await fillCompleteForm(page)
      await page.getByRole('button', { name: 'Registrar turno' }).click()
      await expect(
        page.getByText('el barbero ya tiene una cita en ese intervalo').first(),
      ).toBeVisible()
      return
    case 'created':
      await fillCompleteForm(page)
      await page.getByRole('button', { name: 'Registrar turno' }).click()
      await expect(page.getByText('Turno registrado')).toBeVisible()
  }
}

const scenarios: { readonly file: string; readonly state: Scenario }[] = [
  { file: '01-carga-contexto', state: 'context-loading' },
  { file: '02-error-contexto', state: 'context-error' },
  { file: '03-sin-barberos', state: 'no-barbers' },
  { file: '04-formulario-vacio', state: 'empty-form' },
  { file: '05-cargando-servicios', state: 'services-loading' },
  { file: '06-sin-servicios-asignados', state: 'services-empty' },
  { file: '07-error-servicios', state: 'services-error' },
  { file: '08-formulario-completo', state: 'filled-form' },
  { file: '09-error-validacion', state: 'validation-error' },
  { file: '10-guardando', state: 'saving' },
  { file: '11-conflicto-horario', state: 'schedule-conflict' },
  { file: '12-turno-registrado', state: 'created' },
]

for (const viewport of [
  { name: 'desktop', width: 1440, height: 1024 },
  { name: 'mobile', width: 420, height: 935 },
] as const) {
  test.describe(`fidelidad mock de /panel/turnos/nuevo en ${viewport.name}`, () => {
    // es-CO fija el formato que el control nativo `type="date"` dibuja
    // (04/09/2026, como en el atlas) y la zona evita que "hoy" dependa de la
    // máquina que corre la evidencia.
    test.use({
      locale: 'es-CO',
      timezoneId: 'America/Bogota',
      deviceScaleFactor: viewport.name === 'mobile' ? 2 : 1,
    })

    for (const scenario of scenarios) {
      test(`${scenario.file} usa datos mock sin backend`, async ({ page }) => {
        await page.setViewportSize(viewport)
        const pending = await installNewAppointmentMock(page, scenario.state)
        await page.goto('/panel/turnos/nuevo', { waitUntil: 'domcontentloaded' })
        await driveScenario(page, scenario.state)

        expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
        expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)

        // Mismas dos guardas que el generador del atlas
        // (tools/mockups/nuevo-turno-eventos/render.mjs): ninguna línea se
        // sale de su viewport y, en escritorio, los doce eventos entran en
        // 1024px de alto. Se mide también el área de contenido del cascarón,
        // que es la que puede desplazarse por su cuenta.
        const box = await page.evaluate(() => {
          const content = document.querySelector('.private-shell__content')
          return {
            docWidth: document.documentElement.scrollWidth,
            docHeight: document.documentElement.scrollHeight,
            contentScrollHeight: content?.scrollHeight ?? 0,
            contentClientHeight: content?.clientHeight ?? 0,
          }
        })
        expect(box.docWidth, `${scenario.file} desborda a lo ancho`).toBeLessThanOrEqual(
          viewport.width,
        )
        if (viewport.name === 'desktop') {
          expect(box.docHeight, `${scenario.file} no cabe en 1024px de alto`).toBeLessThanOrEqual(
            viewport.height,
          )
          expect(
            box.contentScrollHeight,
            `${scenario.file} exige scroll dentro del cascarón`,
          ).toBeLessThanOrEqual(box.contentClientHeight)
        }

        // En móvil el atlas registra el recorrido vertical entero (contenido
        // y, tras él, el dock). Quien hace scroll en la app no es el
        // documento sino `.private-shell__content`, así que `fullPage` por sí
        // solo devolvería solo el primer viewport: esta hoja de estilo —
        // aplicada DESPUÉS de verificar el viewport efectivo y solo para la
        // captura— deja crecer el documento igual que en el mockup. Solo se
        // aplica cuando el contenido realmente desborda: un estado corto
        // (error, vacío, éxito) se centra en el alto disponible, y forzarle
        // altura automática lo pegaría al encabezado.
        const contentOverflows = box.contentScrollHeight > box.contentClientHeight
        if (viewport.name === 'mobile' && contentOverflows) {
          await page.addStyleTag({
            // `!important` porque el cascarón fija su alto desde un estilo
            // con ámbito (`[data-v-…]`), de mayor especificidad que esta
            // clase suelta.
            content:
              '.private-shell{height:auto!important;min-height:100dvh!important}' +
              '.private-shell__content{overflow:visible!important}',
          })
        }

        await page.screenshot({
          path: path.join(evidenceRoot, viewport.name, `${scenario.file}.png`),
          fullPage: viewport.name === 'mobile',
          // El atlas dibuja cada evento en reposo: sin congelar animaciones,
          // la entrada de la alerta global y el pulso del esqueleto se
          // capturan a media transición y la evidencia deja de comparar
          // color contra color.
          animations: 'disabled',
        })

        pending?.release()
      })
    }
  })
}

test('vista interactiva local de nuevo turno mock', async ({ page }) => {
  test.skip(
    process.env.SHOW_NEW_APPOINTMENT_MOCK !== '1',
    'Solo se abre cuando se solicita una revisión local.',
  )

  await page.setViewportSize({ width: 1440, height: 1024 })
  await installNewAppointmentMock(page, 'empty-form')
  await page.goto('/panel/turnos/nuevo')
  await expect(page.getByRole('heading', { name: 'Nuevo turno' })).toBeVisible()
  await page.pause()
})
