// CA-062-08 / CA-063-08: selector de profesional, Anterior, Fecha y Siguiente miden 44 px.
// jsdom no calcula layout, así que la regresión (AGENDA-H-02, issue #409: los cuatro medían
// 40 px, copiados del atlas) se fija leyendo las reglas de las SFC: ninguna vuelve a una
// altura literal por debajo del objetivo táctil.
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const read = (relative: string) => readFileSync(resolve(__dirname, relative), 'utf-8')

describe('objetivos táctiles de la barra de controles de la agenda (AGENDA-H-02)', () => {
  it('el selector compacto usa la altura de control, no 40 px', () => {
    const source = read('../BarberSelect.vue')
    const rule = source.match(/\.barber-select--compact \.barber-select__trigger \{[^}]*\}/)?.[0]
    expect(rule).toContain('min-height: var(--control-height)')
    expect(rule).not.toMatch(/min-height:\s*40px/)
  })

  it('el selector de fecha usa la altura de control, no 40 px', () => {
    const source = read('../../../../shared/ui/BaseDatePicker.vue')
    const rule = source.match(/\.agenda-date-picker__trigger \{[^}]*\}/)?.[0]
    expect(rule).toContain('min-height: var(--control-height)')
    expect(rule).not.toMatch(/min-height:\s*40px/)
  })

  it('Anterior y Siguiente no recortan la altura del botón fantasma', () => {
    const source = read('../../pages/DailyAgendaPage.vue')
    const rule = source.match(
      /\.daily-agenda-page__date-nav :deep\(\.base-button--ghost\) \{[^}]*\}/,
    )?.[0]
    expect(rule).toContain('--btn-height: var(--control-height)')
    expect(rule).not.toMatch(/--btn-height:\s*40px/)
  })
})
