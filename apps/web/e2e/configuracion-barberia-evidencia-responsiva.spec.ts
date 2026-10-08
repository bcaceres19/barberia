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
  publicLinkStatus = 200,
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
    } else if (pathname.endsWith('/settings/public-link')) {
      await route.fulfill(
        publicLinkStatus === 200
          ? json({ slug: 'nava-qa-local-k7x2m9' })
          : json({ title: 'Error' }, publicLinkStatus),
      )
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

test.describe('enlace público (DEC-117)', () => {
  for (const viewport of viewports) {
    test(`${viewport.name}: el enlace completo cabe, se copia y se anuncia`, async ({
      page,
      context,
    }) => {
      await context.grantPermissions(['clipboard-read', 'clipboard-write'])
      await openSettings(page, viewport.width, viewport.height)

      const panel = page.locator('#configuracion-enlace')
      await panel.scrollIntoViewIfNeeded()
      await expect(panel.getByRole('textbox', { name: 'Tu enlace de reservas' })).toHaveText(
        /\/reservar\/nava-qa-local-k7x2m9$/,
      )

      await panel.getByRole('button', { name: 'Copiar enlace' }).click()
      await expect(panel.getByText('Enlace copiado.')).toBeVisible()
      expect(await page.evaluate(() => navigator.clipboard.readText())).toMatch(
        /\/reservar\/nava-qa-local-k7x2m9$/,
      )

      await expectNoHorizontalScroll(page)
      await panel.screenshot({
        path: path.join(evidenceDir, 'enlace-publico', `${viewport.name}-copiado.png`),
      })
    })
  }

  test('sin permiso de portapapeles queda seleccionado y lo dice', async ({ page }) => {
    await openSettings(page, 1280, 900)
    await page.evaluate(() => {
      Object.defineProperty(navigator, 'clipboard', {
        value: { writeText: () => Promise.reject(new Error('denied')) },
        configurable: true,
      })
    })

    const panel = page.locator('#configuracion-enlace')
    await panel.getByRole('button', { name: 'Copiar enlace' }).click()

    await expect(panel.getByText(/cópialo con Ctrl\+C/)).toBeVisible()
    await expect(panel.getByRole('textbox', { name: 'Tu enlace de reservas' })).toBeFocused()
    await panel.screenshot({
      path: path.join(evidenceDir, 'enlace-publico', '1280-manual.png'),
    })
  })

  test('si el servidor falla, avisa y se puede reintentar', async ({ page }) => {
    await openSettings(page, 1280, 900, {}, 500)
    const panel = page.locator('#configuracion-enlace')
    await expect(panel.getByText('No pudimos cargar tu enlace')).toBeVisible()
    await panel.screenshot({ path: path.join(evidenceDir, 'enlace-publico', '1280-error.png') })
    expect(await axeViolations(page)).toEqual([])
  })
})

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

        // El `zoom` del texto no debe dejar el documento más alto que la ventana: si lo
        // hace, el scroll general destapa una franja marfil bajo el dock (#298).
        expect(
          await page.evaluate(
            () => document.scrollingElement!.scrollHeight > document.scrollingElement!.clientHeight,
          ),
        ).toBe(false)

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
