// CA-020-04: un instante conocido, formateado con la zona CONFIGURADA de la
// barbería, produce la misma hora sin importar la zona del dispositivo que
// ejecuta el código. formatInstantInTimezone nunca lee la zona del entorno
// (Intl.DateTimeFormat().resolvedOptions().timeZone); solo usa la que
// recibe explícitamente.
import { describe, expect, it } from 'vitest'
import {
  formatFullDateInTimezone,
  formatInstantInTimezone,
  formatTimeInTimezone,
} from '../formatInstant'

const knownInstant = '2026-09-12T17:00:00Z'

describe('formatInstantInTimezone', () => {
  it('produces the same formatted hour for the same explicit timezone, called repeatedly', () => {
    // Modela "dos dispositivos": dos llamadas independientes, la misma zona
    // configurada de la barbería en ambas. Ninguna depende de dónde corre
    // el proceso.
    const deviceOne = formatInstantInTimezone(knownInstant, 'America/Bogota')
    const deviceTwo = formatInstantInTimezone(knownInstant, 'America/Bogota')
    expect(deviceOne).toBe(deviceTwo)
    expect(deviceOne).toContain('12:00')
  })

  it('changes the formatted hour when the configured timezone changes, proving timeZone (not the environment) governs the result', () => {
    const bogota = formatInstantInTimezone(knownInstant, 'America/Bogota')
    const madrid = formatInstantInTimezone(knownInstant, 'Europe/Madrid')
    // Control negativo (trabajo requerido §5, CA-020-04): si esta función
    // ignorara `timezone` y usara la zona del dispositivo en su lugar,
    // ambos resultados serían siempre iguales entre sí sin importar la
    // zona pedida. Que difieran demuestra que `timezone` sí gobierna la
    // salida.
    expect(bogota).not.toBe(madrid)
  })

  it('never reads Intl.DateTimeFormat().resolvedOptions().timeZone (the environment default)', () => {
    const environmentDefault = new Intl.DateTimeFormat().resolvedOptions().timeZone
    // Se elige a propósito una zona distinta de la que tenga el entorno de
    // pruebas, para que el resultado solo pueda coincidir con "America/Bogota"
    // si la función de verdad usó el parámetro explícito.
    const configuredTimezone =
      environmentDefault === 'America/Bogota' ? 'Europe/Madrid' : 'America/Bogota'
    const got = formatInstantInTimezone(knownInstant, configuredTimezone)
    const usingEnvironmentInstead = formatInstantInTimezone(knownInstant, environmentDefault)
    if (environmentDefault !== configuredTimezone) {
      expect(got).not.toBe(usingEnvironmentInstead)
    }
  })
})

describe('formatTimeInTimezone', () => {
  it('formats only the time, in the explicit timezone (HU-062)', () => {
    const got = formatTimeInTimezone(knownInstant, 'America/Bogota')
    expect(got).toContain('12:00')
    // Nunca incluye el año, ni fragmentos numéricos del mes/día que
    // pudieran confundirse con una fecha (dateStyle nunca se aplica aquí).
    expect(got).not.toContain('2026')
  })

  it('changes with the explicit timezone, proving it governs the result, not the environment', () => {
    const bogota = formatTimeInTimezone(knownInstant, 'America/Bogota')
    const madrid = formatTimeInTimezone(knownInstant, 'Europe/Madrid')
    expect(bogota).not.toBe(madrid)
  })
})

describe('formatFullDateInTimezone', () => {
  it('formats the full civil date in the explicit timezone, never the device default', () => {
    // 2026-08-29T02:30:00Z es 28 de agosto en America/Bogota (UTC-5) y 29 de
    // agosto en Europe/Madrid (UTC+2 en verano): el mismo instante produce
    // fechas civiles DISTINTAS según la zona explícita, nunca la del
    // dispositivo que ejecuta la prueba.
    const at = new Date('2026-08-29T02:30:00Z')
    const bogota = formatFullDateInTimezone('America/Bogota', at)
    const madrid = formatFullDateInTimezone('Europe/Madrid', at)
    expect(bogota.toLowerCase()).toContain('28 de agosto')
    expect(madrid.toLowerCase()).toContain('29 de agosto')
  })
})
