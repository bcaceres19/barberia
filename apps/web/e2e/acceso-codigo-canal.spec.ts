import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Canal del código de acceso en Configuración (DEC-139, HU-007). Respuestas
 * simuladas del API: que el servidor derive los destinos, que cada barbería tenga
 * su canal, el CHECK de la base y la entrega real por cada canal los cubren las
 * pruebas Go y SQL con PostgreSQL real. Aquí se recorre lo visible:
 *  1. el panel «Código de acceso» muestra los tres canales con el vigente marcado;
 *  2. elegir uno envía solo `{ channel }` (jamás un correo o teléfono) y confirma;
 *  3. si el guardado falla, se explica y se conserva el canal anterior.
 * En 320/360/768/1280 px, sin scroll horizontal y con axe-core. Evidencia en
 * e2e/evidence/configuracion/acceso-codigo/.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const evidenceDir = path.join(here, 'evidence', 'configuracion', 'acceso-codigo')
const axeScriptPath = path.join(here, '..', 'node_modules', 'axe-core', 'axe.min.js')

// Sin esto axe mide fotogramas a medio fundir y da falsos fallos de contraste.
test.use({ contextOptions: { reducedMotion: 'reduce' } })

const viewports = [
  { name: '320', width: 320, height: 720 },
  { name: '360', width: 360, height: 800 },
  { name: '768', width: 768, height: 1024 },
  { name: '1280', width: 1280, height: 900 },
] as const

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
  panelProfile: 'shop',
  hasPhysicalLocation: false,
  address: null,
  latitude: null,
  longitude: null,
  offersHomeService: false,
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

interface Api {
  patches: unknown[]
}

async function openSettings(
  page: Page,
  options: { initial?: string; patchStatus?: number } = {},
): Promise<Api> {
  const api: Api = { patches: [] }
  let channel = options.initial ?? 'email'
  await page.route('**/api/v1/private/**', async (route) => {
    const request = route.request()
    const { pathname } = new URL(request.url())
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
      await route.fulfill(json({ slug: 'navaqalocalk7x2m9q4' }))
    } else if (pathname.endsWith('/settings/login-otp')) {
      if (request.method() === 'PATCH') {
        const body = request.postDataJSON() as { channel: string }
        api.patches.push(body)
        if ((options.patchStatus ?? 200) !== 200) {
          await route.fulfill(json({ title: 'Error' }, options.patchStatus))
          return
        }
        channel = body.channel
      }
      await route.fulfill(json({ channel }))
    } else {
      await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
    }
  })
  await page.goto('/panel/barberia')
  await expect(page.getByRole('heading', { name: 'Código de acceso', level: 2 })).toBeVisible()
  await expect(page.getByRole('radiogroup', { name: 'Canal del código de acceso' })).toBeVisible()
  return api
}

async function expectAccessibleWithoutOverflow(page: Page) {
  const overflow = await page.evaluate(() => {
    const shell = document.querySelector('.private-shell__content') as HTMLElement
    return {
      document: document.documentElement.scrollWidth > document.documentElement.clientWidth,
      content: shell.scrollWidth > shell.clientWidth,
    }
  })
  expect(overflow).toEqual({ document: false, content: false })

  await page.addScriptTag({ path: axeScriptPath })
  const violations = await page.evaluate(async () => {
    const result = await (
      window as unknown as {
        axe: { run: (ctx: unknown, opts: unknown) => Promise<{ violations: unknown[] }> }
      }
    ).axe.run(document, {
      runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21aa', 'wcag22aa'] },
    })
    return result.violations
  })
  expect(violations).toEqual([])
}

test('muestra los tres canales con el vigente marcado', async ({ page }) => {
  await openSettings(page, { initial: 'both' })

  await expect(page.getByRole('radio', { name: /^Correo y WhatsApp/ })).toBeChecked()
  await expect(page.getByRole('radio', { name: /^Correo\b(?! y)/ })).not.toBeChecked()
  await expect(page.getByRole('radio', { name: /^WhatsApp/ })).not.toBeChecked()
})

test('elegir un canal envía solo el canal y lo confirma', async ({ page }) => {
  const api = await openSettings(page)

  await page.getByRole('radio', { name: /^WhatsApp/ }).check({ force: true })

  await expect(page.getByText('Guardado.')).toBeVisible()
  expect(api.patches).toEqual([{ channel: 'whatsapp' }])
  await expect(page.getByRole('radio', { name: /^WhatsApp/ })).toBeChecked()
})

test('si no se puede guardar, se explica y se conserva el canal anterior', async ({ page }) => {
  const api = await openSettings(page, { initial: 'email', patchStatus: 500 })

  await page.getByRole('radio', { name: /^WhatsApp/ }).check({ force: true })

  await expect(page.getByText(/se mantiene la opción anterior/)).toBeVisible()
  expect(api.patches).toEqual([{ channel: 'whatsapp' }])
  await expect(page.getByRole('radio', { name: /^Correo\b(?! y)/ })).toBeChecked()
})

for (const viewport of viewports) {
  test(`el panel es accesible y sin scroll horizontal a ${viewport.name} px`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await openSettings(page, { initial: 'both' })
    await page.waitForTimeout(900)

    const panel = page.locator('#configuracion-acceso')
    await panel.scrollIntoViewIfNeeded()
    await expectAccessibleWithoutOverflow(page)
    await panel.screenshot({ path: path.join(evidenceDir, `acceso-codigo-${viewport.name}.png`) })
  })
}
