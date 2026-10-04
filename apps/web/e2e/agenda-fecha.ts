import type { Page } from '@playwright/test'

/**
 * Elige una fecha civil AAAA-MM-DD en el calendario propio de /panel
 * (AgendaDatePicker). Sustituye a `getByLabel('Fecha').fill(...)`: el campo ya
 * no es un `input[type=date]` sino un disparador que abre un calendario.
 * Se abre en el mes de la fecha vigente, navega los meses de diferencia y
 * pulsa el día; el calendario se cierra solo al elegirlo.
 */
export async function pickAgendaDate(page: Page, isoDate: string): Promise<void> {
  await page.locator('#daily-agenda-date-picker').click()
  const focused = await page
    .locator('.agenda-date-picker__day[tabindex="0"]')
    .getAttribute('data-date')
  if (!focused) throw new Error('El calendario no expone un día enfocado')

  const monthIndex = (date: string) => {
    const [year, month] = date.split('-').map(Number)
    return year! * 12 + (month! - 1)
  }
  const delta = monthIndex(isoDate) - monthIndex(focused)
  const nav = page.getByRole('button', { name: delta < 0 ? 'Mes anterior' : 'Mes siguiente' })
  for (let i = 0; i < Math.abs(delta); i++) await nav.click()

  await page.locator(`.agenda-date-picker__day[data-date="${isoDate}"]`).click()
}
