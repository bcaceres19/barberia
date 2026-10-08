/**
 * Cobertura del vocabulario (DEC-110, DEC-119): ninguna pantalla del panel ni de la reserva pública
 * escribe «barbero» ni «barbería» como texto fijo; las dice con `useVocabulary`.
 * Si esta prueba falla, la palabra que el negocio eligió (manicurista, estudio...)
 * se perdería en esa pantalla. Es una red de seguridad estática: lee el código
 * fuente, quita los comentarios y busca la palabra en lo que queda.
 */
import { describe, expect, it } from 'vitest'

const sources = import.meta.glob(
  '/src/modules/{agenda,schedules,staff,barberServices,settings,catalog,auth,public-booking,customer-access}/**/*.{vue,ts}',
  { query: '?raw', import: 'default', eager: true },
) as Record<string, string>

// Identificadores que contienen la palabra y no son texto de pantalla: nombres de
// ruta, rutas, claves técnicas del contrato (`cancelled_by_barber`, `barber_id`)
// y los valores iniciales declarados como respaldo.
const ALLOWED = [
  /staff-barberos|configuracion-barberia|reserva-publica-barbero|barber-services/,
  /servicios-por-barbero|path: '([^']*\/)?barber(o|os|ia)?[/']/,
  /cancelled_by_barber|appointment_cancelled_by_barber|barber_id/,
  /\b(label|term|singular|plural): '(Barberos|Servicios por barbero|barber[a-zí]*)'/,
  /label: 'Servicios por barbero'|label: 'Barberos'/,
]

// Archivos donde la palabra es dato y no texto: los valores iniciales y las
// sugerencias del selector de vocabulario.
const EXEMPT = [/\/__tests__\//, /\.test\.ts$/, /settingsDraft\.ts$/]

const WORD = /barber(o|a|os|as|ía|ías|ia|ias)\b/i

function stripComments(source: string): string {
  return source
    .replace(/<!--[\s\S]*?-->/g, '')
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/(^|[^:'"`\w])\/\/.*$/gm, '$1')
}

describe('cobertura del vocabulario en el panel', () => {
  it('lee los archivos del panel', () => {
    expect(Object.keys(sources).length).toBeGreaterThan(40)
  })

  it('no deja «barbero» ni «barbería» como texto fijo', () => {
    const offenders: string[] = []
    for (const [path, raw] of Object.entries(sources)) {
      if (EXEMPT.some((re) => re.test(path))) continue
      stripComments(raw)
        .split('\n')
        .forEach((line, index) => {
          if (!WORD.test(line)) return
          if (ALLOWED.some((re) => re.test(line))) return
          offenders.push(`${path}:${index + 1}: ${line.trim().slice(0, 120)}`)
        })
    }
    expect(offenders).toEqual([])
  })
})
