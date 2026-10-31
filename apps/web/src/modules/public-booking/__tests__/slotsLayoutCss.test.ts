// CA-095-05 (PSLOTS-F-01, issue #415) y PCONF-F-02 (issue #417): jsdom no calcula layout, así que la regresión se fija
// sobre las reglas del CSS. La tira de fechas dejó de contraerse a ~16 px (chips de 24 px) y
// la lista de horas dejó de desbordar sobre el pie en ventanas bajas.
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const css = readFileSync(resolve(__dirname, '../styles/publicBooking.css'), 'utf-8')

function rule(selector: string): string {
  const start = css.indexOf(`${selector} {`)
  expect(start, `regla ${selector}`).toBeGreaterThanOrEqual(0)
  return css.slice(start, css.indexOf('}', start))
}

describe('pantalla de horario público: layout en ventanas bajas', () => {
  it('la tira de fechas conserva su alto natural y no cede alto a la lista', () => {
    expect(rule('.pb-days')).toMatch(/flex:\s*none/)
  })

  it('el contenedor del horario tiene un suelo de alto: bajo él la página se desplaza', () => {
    expect(rule('.pb-layout:has(.pb-page--scroll-slots)')).toMatch(/height:\s*max\(100dvh,/)
  })

  it('el horario ya no comparte la regla de alto fijo de 100dvh', () => {
    const shared = rule(
      '.pb-layout:has(.pb-page--scroll-choices),\n.pb-layout:has(.pb-page--scroll-form)',
    )
    expect(shared).toMatch(/height:\s*100dvh/)
    expect(shared).not.toContain('scroll-slots')
  })

  it('el resumen de la reserva tiene su propio suelo de alto, así el pie no tapa «Confirmar turno»', () => {
    expect(rule('.pb-layout:has(.pb-page--scroll-form .customer-details__summary-body)')).toMatch(
      /height:\s*max\(100dvh,/,
    )
  })

  it('el formulario de datos tiene un suelo de alto: a 200 % de zoom la hoja no se aplasta', () => {
    expect(rule('.pb-layout:has(.pb-page--scroll-form .customer-details__fields)')).toMatch(
      /height:\s*max\(100dvh,/,
    )
  })
})
