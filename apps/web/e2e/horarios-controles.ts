import { expect, type Locator, type Page } from '@playwright/test'

/**
 * Controles propios de Horarios (selector de barbero, hora y fecha). Sustituyen
 * a `selectOption(...)` y `getByLabel(...).fill(...)` de los controles
 * nativos que la pantalla tenía antes del rediseño: ahora son disparadores que
 * abren un listbox, un panel de hora o un calendario.
 */

// Elige un barbero en el listbox de la pantalla (BaseSelect).
export async function selectBarber(page: Page, fullName: string): Promise<void> {
  await page.locator('#schedules-barber-select').click()
  await page.getByRole('option', { name: fullName }).click()
  await expect(page.locator('#schedules-barber-select')).toContainText(fullName)
}

// Fija una hora HH:MM en un BaseTimePicker cuyo rótulo es `label`.
export async function pickTime(scope: Locator, page: Page, label: string, time: string, index = 0) {
  await scope
    .getByRole('button', { name: new RegExp(label) })
    .nth(index)
    .click()
  const picker = page.getByRole('dialog', { name: label, exact: true })
  await picker.getByRole('spinbutton', { name: 'Hora', exact: true }).fill(time.slice(0, 2))
  await picker.getByRole('spinbutton', { name: 'Minutos', exact: true }).fill(time.slice(3))
  await picker.getByRole('button', { name: 'Aplicar hora', exact: true }).click()
}

// Elige una fecha civil AAAA-MM-DD en el BaseDatePicker cuyo disparador tiene
// el id `triggerId`: navega por teclado hasta el mes y pulsa el día.
export async function pickDate(page: Page, triggerId: string, date: string): Promise<void> {
  await page.locator(`#${triggerId}`).click()
  const calendar = page.getByRole('dialog', { name: 'Elegir fecha', exact: true })
  const focused = await calendar.locator('[data-date][tabindex="0"]').getAttribute('data-date')
  if (!focused) throw new Error('Calendario sin día enfocado')
  const monthIndex = (value: string) => Number(value.slice(0, 4)) * 12 + Number(value.slice(5, 7))
  const delta = monthIndex(date) - monthIndex(focused)
  const years = Math.trunc(delta / 12)
  for (let index = 0; index < Math.abs(years); index++)
    await page.keyboard.press(years > 0 ? 'Shift+PageDown' : 'Shift+PageUp')
  const months = delta - years * 12
  for (let index = 0; index < Math.abs(months); index++)
    await page.keyboard.press(months > 0 ? 'PageDown' : 'PageUp')
  await calendar.locator(`[data-date="${date}"]`).click()
}

// Abre un apartado del panel lateral (acordeón) si está cerrado; los apartados
// cerrados no reciben foco ni se exponen a lectores de pantalla.
export async function openAsidePanel(page: Page, title: string): Promise<void> {
  const trigger = page.getByRole('button', { name: new RegExp(`^${title}`) })
  if ((await trigger.getAttribute('aria-expanded')) !== 'true') await trigger.click()
  await expect(trigger).toHaveAttribute('aria-expanded', 'true')
}
