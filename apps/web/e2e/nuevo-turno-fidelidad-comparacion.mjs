/*
 * Comparación visual del issue #190: para cada uno de los doce eventos del
 * atlas `nuevo-turno-eventos` y sus dos viewports, compone la lámina de
 * referencia y la captura real de `nuevo-turno-fidelidad-mock.spec.ts` lado a
 * lado y genera además la imagen de diferencia absoluta.
 *
 * No es una suite de pruebas (extensión .mjs a propósito: `testDir` solo
 * recoge `*.spec.ts`): es el paso de comparación que la skill
 * `nava-mockup-fidelity` exige antes de declarar una pantalla terminada, y se
 * ejecuta después de capturar la evidencia.
 *
 *   node apps/web/e2e/nuevo-turno-fidelidad-comparacion.mjs
 *
 * El diff se calcula sobre la región común: en móvil las dos capturas son de
 * página completa y su alto puede diferir, y esa diferencia de alto se
 * reporta por consola en vez de disimularse escalando una sobre la otra.
 */
import playwright from '@playwright/test'
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const { chromium } = playwright

const HERE = dirname(fileURLToPath(import.meta.url))
const REPO = resolve(HERE, '../../..')
const ATLAS = resolve(
  REPO,
  'docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos',
)
const CAPTURES = resolve(HERE, 'evidence/nuevo-turno/fidelidad-190/mock')
const OUT = resolve(HERE, 'evidence/nuevo-turno/fidelidad-190/comparacion')

const EVENTS = [
  '01-carga-contexto',
  '02-error-contexto',
  '03-sin-barberos',
  '04-formulario-vacio',
  '05-cargando-servicios',
  '06-sin-servicios-asignados',
  '07-error-servicios',
  '08-formulario-completo',
  '09-error-validacion',
  '10-guardando',
  '11-conflicto-horario',
  '12-turno-registrado',
]

const dataUri = (file) => `data:image/png;base64,${readFileSync(file).toString('base64')}`

async function main() {
  const browser = await chromium.launch()
  const page = await browser.newPage()
  await page.setContent('<canvas id="c"></canvas>')

  const report = []

  for (const viewport of ['desktop', 'mobile']) {
    for (const event of EVENTS) {
      const reference = resolve(ATLAS, viewport, 'nuevo-turno', `${event}.png`)
      const actual = resolve(CAPTURES, viewport, `${event}.png`)
      if (!existsSync(reference) || !existsSync(actual)) {
        report.push(`${viewport}/${event}: FALTA una de las dos imágenes`)
        continue
      }

      const result = await page.evaluate(
        async ([refSrc, actSrc]) => {
          const load = (src) =>
            new Promise((done, fail) => {
              const img = new Image()
              img.onload = () => done(img)
              img.onerror = fail
              img.src = src
            })
          const [a, b] = await Promise.all([load(refSrc), load(actSrc)])

          const draw = (img) => {
            const c = document.createElement('canvas')
            c.width = img.naturalWidth
            c.height = img.naturalHeight
            c.getContext('2d').drawImage(img, 0, 0)
            return c
          }
          const ca = draw(a)
          const cb = draw(b)

          // --- lado a lado: referencia | implementación, con rótulo ---
          const gap = 24
          const label = 34
          const side = document.createElement('canvas')
          side.width = ca.width + gap + cb.width
          side.height = label + Math.max(ca.height, cb.height)
          const sx = side.getContext('2d')
          sx.fillStyle = '#101b2b'
          sx.fillRect(0, 0, side.width, side.height)
          sx.fillStyle = '#b8955a'
          sx.font = '600 18px system-ui, sans-serif'
          sx.fillText('ATLAS (referencia)', 8, 23)
          sx.fillText('APP (implementación)', ca.width + gap + 8, 23)
          sx.drawImage(ca, 0, label)
          sx.drawImage(cb, ca.width + gap, label)

          // --- diferencia absoluta sobre la región común ---
          const w = Math.min(ca.width, cb.width)
          const h = Math.min(ca.height, cb.height)
          const da = ca.getContext('2d').getImageData(0, 0, w, h)
          const db = cb.getContext('2d').getImageData(0, 0, w, h)
          const diff = document.createElement('canvas')
          diff.width = w
          diff.height = h
          const dctx = diff.getContext('2d')
          const out = dctx.createImageData(w, h)
          let changed = 0
          for (let i = 0; i < da.data.length; i += 4) {
            const dr = Math.abs(da.data[i] - db.data[i])
            const dg = Math.abs(da.data[i + 1] - db.data[i + 1])
            const dbl = Math.abs(da.data[i + 2] - db.data[i + 2])
            const delta = Math.max(dr, dg, dbl)
            // Umbral de 16/255: por debajo es ruido de antialias del texto,
            // no una diferencia de composición.
            if (delta > 16) changed += 1
            out.data[i] = delta > 16 ? 255 : 0
            out.data[i + 1] = delta > 16 ? 40 : 0
            out.data[i + 2] = delta > 16 ? 40 : 0
            out.data[i + 3] = 255
          }
          dctx.putImageData(out, 0, 0)

          return {
            side: side.toDataURL('image/png'),
            diff: diff.toDataURL('image/png'),
            refSize: [ca.width, ca.height],
            actSize: [cb.width, cb.height],
            changedRatio: changed / (w * h),
          }
        },
        [dataUri(reference), dataUri(actual)],
      )

      const dir = resolve(OUT, viewport)
      mkdirSync(dir, { recursive: true })
      const save = (name, uri) =>
        writeFileSync(resolve(dir, name), Buffer.from(uri.split(',')[1], 'base64'))
      save(`${event}-lado-a-lado.png`, result.side)
      save(`${event}-diff.png`, result.diff)

      const sizeNote =
        result.refSize[1] === result.actSize[1]
          ? ''
          : ` · alto atlas ${result.refSize[1]} vs app ${result.actSize[1]}`
      report.push(
        `${viewport}/${event}: ${(result.changedRatio * 100).toFixed(1)}% de píxeles distintos${sizeNote}`,
      )
    }
  }

  await browser.close()

  const summary = report.join('\n')
  console.log(summary)
  mkdirSync(OUT, { recursive: true })
  writeFileSync(resolve(OUT, 'resumen.txt'), `${summary}\n`)
}

await main()
