import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Evidencia visual responsive/accesible de Configuración (HU-020, DEC-110):
 * la pantalla en 320/360/768/1280 px más zoom 200% aproximado, en los dos
 * modos (Tinta y Marfil), con los cuatro tamaños de texto, el selector de zona
 * abierto, la barra de cambios sin guardar, foco visible, sin scroll
 * horizontal y con axe-core (incluido el contraste real de color) sobre cada
 * combinación. Corre con respuestas simuladas del API: el recorrido contra el
 * API real vive en configuracion-barberia.spec.ts. Las capturas van a
 * e2e/evidence/configuracion/rediseno/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'configuracion', 'rediseno')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

const settings = {
  name: 'NAVA QA Local',
  timezone: 'America/Bogota',
  contactEmail: 'contacto@nava.qa',
  contactPhone: '+573001234567',
}
const brand = {
  accent: 'brass',
  businessTerm: 'barbería',
  businessTermGender: 'feminine',
  professionalTerm: 'barbero',
  professionalTermPlural: 'barberos',
  professionalTermGender: 'masculine',
}

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
  { name: '1280-zoom200', width: 640, height: 450 },
] as const

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

async function openSettings(
  page: Page,
  width: number,
  height: number,
  prefs: { theme?: string; textScale?: string; motion?: string } = {},
) {
  await page.setViewportSize({ width, height })
  await page.addInitScript((stored) => {
    window.localStorage.setItem(
      'nava.appearance.v1',
      JSON.stringify({ theme: 'ink', textScale: 'normal', motion: 'full', ...stored }),
    )
  }, prefs)
  await page.route('**/api/v1/private/**', async (route) => {
    const { pathname } = new URL(route.request().url())
    if (pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'nava', name: 'NAVA QA Local' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
    } else if (pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json(settings))
    } else if (pathname.endsWith('/settings/brand')) {
      await route.fulfill(json(brand))
    } else {
      await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
    }
  })
  await page.goto('/panel/barberia')
  await expect(page.getByRole('heading', { name: 'Configuración', level: 1 })).toBeVisible()
  await expect(page.getByLabel('Nombre', { exact: true })).toBeVisible()
  // Las entradas escalonadas terminan antes de medir o capturar.
  await page.waitForTimeout(900)
}

async function expectNoHorizontalScroll(page: Page) {
  const overflow = await page.evaluate(() => {
    const shell = document.querySelector('.private-shell__content') as HTMLElement
    return {
      document: document.documentElement.scrollWidth > document.documentElement.clientWidth,
      content: shell.scrollWidth > shell.clientWidth,
    }
  })
  expect(overflow).toEqual({ document: false, content: false })
}

async function axeViolations(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  return page.evaluate(async () => {
    const results = await (
      window as unknown as {
        axe: {
          run: (
            c: Document,
            o: object,
          ) => Promise<{ violations: Array<{ id: string; nodes: Array<{ target: unknown }> }> }>
        }
      }
    ).axe.run(document, {
      runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
    })
    return results.violations.map((v) => ({ id: v.id, targets: v.nodes.map((n) => n.target) }))
  })
}

for (const theme of ['ink', 'ivory'] as const) {
  test.describe(`modo ${theme}`, () => {
    for (const viewport of viewports) {
      test(`${viewport.name}: sin scroll horizontal, foco visible y accesible`, async ({
        page,
      }) => {
        await openSettings(page, viewport.width, viewport.height, { theme })

        await page.screenshot({
          path: path.join(evidenceDir, theme, viewport.name, 'normal.png'),
          fullPage: true,
        })
        await expectNoHorizontalScroll(page)

        await page.getByLabel('Nombre', { exact: true }).focus()
        await expect(page.getByLabel('Nombre', { exact: true })).toBeFocused()
        await page.screenshot({ path: path.join(evidenceDir, theme, viewport.name, 'foco.png') })

        expect(await axeViolations(page)).toEqual([])
      })
    }

    test('con cambios sin guardar: barra de guardado, acento en vivo y accesible', async ({
      page,
    }) => {
      await openSettings(page, 1280, 900, { theme })
      await page.getByRole('radio', { name: /Esmeralda/ }).check()
      await page.getByRole('button', { name: 'estilista', exact: true }).click()
      await expect(page.getByText('Cambios sin guardar')).toBeVisible()
      await page.waitForTimeout(500)

      await page.screenshot({
        path: path.join(evidenceDir, theme, '1280', 'cambios-sin-guardar.png'),
        fullPage: true,
      })
      await expectNoHorizontalScroll(page)
      expect(await axeViolations(page)).toEqual([])
    })

    test('selector de zona horaria abierto', async ({ page }) => {
      await openSettings(page, 1280, 900, { theme })
      await page.getByRole('combobox', { name: 'Zona horaria' }).click()
      await expect(page.getByRole('listbox')).toBeVisible()
      await page.waitForTimeout(400)

      await page.screenshot({ path: path.join(evidenceDir, theme, '1280', 'zona-abierta.png') })
      expect(await axeViolations(page)).toEqual([])
    })
  })
}

test.describe('tamaño de texto', () => {
  for (const textScale of ['small', 'large', 'xlarge'] as const) {
    for (const viewport of [viewports[1], viewports[3]]) {
      test(`${textScale} a ${viewport.name}: el dock sigue a la vista y no hay scroll horizontal`, async ({
        page,
      }) => {
        await openSettings(page, viewport.width, viewport.height, { textScale })

        await page.screenshot({
          path: path.join(evidenceDir, 'texto', `${textScale}-${viewport.name}.png`),
        })
        await expectNoHorizontalScroll(page)

        // El escalado compensa 100dvh: el dock queda dentro de la ventana, no
        // empujado fuera por el zoom.
        const dock = await page.locator('.app-nav').boundingBox()
        expect(dock).not.toBeNull()
        expect(dock!.y + dock!.height).toBeLessThanOrEqual(viewport.height + 1)
        expect(dock!.y).toBeGreaterThanOrEqual(0)
      })
    }
  }

  test('a 320 px, "Muy grande" se limita para no pasar de 320 px efectivos', async ({ page }) => {
    await openSettings(page, 320, 720, { textScale: 'xlarge' })

    const zoom = await page.evaluate(() =>
      Number(document.documentElement.style.getPropertyValue('--ui-zoom') || '1'),
    )
    expect(zoom).toBe(1)
    await expectNoHorizontalScroll(page)
  })
})

test('Marfil no se filtra al acceso, que conserva su propio diseño', async ({ page }) => {
  await openSettings(page, 1280, 900, { theme: 'ivory' })
  expect(await page.evaluate(() => document.documentElement.dataset.appTheme)).toBe('ivory')

  // Salir del panel retira el tema de <html>.
  await page.route('**/api/v1/public/**', (route) => route.fulfill(json({}, 404)))
  await page.evaluate(() => (window as unknown as { __router?: unknown }).__router)
  await page.goto('/acceso')
  await expect(page.getByRole('heading', { name: 'Accede a NAVA' })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.dataset.appTheme ?? null)).toBeNull()
})

test('con animaciones reducidas ninguna transición dura más de un instante', async ({ page }) => {
  await openSettings(page, 1280, 900, { motion: 'reduced' })

  expect(await page.evaluate(() => document.documentElement.dataset.motion)).toBe('reduced')
  const longest = await page.evaluate(() => {
    const durations = [...document.querySelectorAll('.settings-panel, .settings-page__title')].map(
      (el) => parseFloat(getComputedStyle(el).animationDuration),
    )
    return Math.max(...durations)
  })
  expect(longest).toBeLessThan(0.01)
})
