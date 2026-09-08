/*
 * Comparación visual del issue #193. El atlas reúne dos rutas en una lámina;
 * se recortan los paneles de Configuración y se genera referencia | app +
 * diff absoluto. No forma parte de la suite de pruebas.
 *
 *   node apps/web/e2e/configuracion-barberia-fidelidad-comparacion.mjs
 */
import playwright from '@playwright/test'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const { chromium } = playwright
const here = dirname(fileURLToPath(import.meta.url))
const repo = resolve(here, '../../..')
const atlas = resolve(
  repo,
  'docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/06-barberos-y-barberia.png',
)
const captures = resolve(here, 'evidence/configuracion-barberia/fidelidad-193')
const output = resolve(captures, 'comparacion')
const dataUri = (file) => `data:image/png;base64,${readFileSync(file).toString('base64')}`

const panels = [
  // Rectángulos medidos en la lámina original 1536×1024.
  { name: 'desktop', crop: [771, 10, 744, 514], actual: 'desktop/guardado.png' },
  { name: 'mobile', crop: [1247, 552, 266, 440], actual: 'mobile/guardado.png' },
]

async function main() {
  const browser = await chromium.launch()
  const page = await browser.newPage()
  await page.setContent('<canvas></canvas>')
  const report = []

  for (const panel of panels) {
    const actual = resolve(captures, panel.actual)
    if (!existsSync(actual)) throw new Error(`Falta evidencia: ${actual}`)
    const result = await page.evaluate(
      async ([referenceSource, actualSource, crop]) => {
        const load = (source) =>
          new Promise((done, fail) => {
            const image = new Image()
            image.onload = () => done(image)
            image.onerror = fail
            image.src = source
          })
        const [reference, implementation] = await Promise.all([
          load(referenceSource),
          load(actualSource),
        ])
        const [x, y, width, height] = crop
        const ref = document.createElement('canvas')
        ref.width = width
        ref.height = height
        ref.getContext('2d').drawImage(reference, x, y, width, height, 0, 0, width, height)
        const app = document.createElement('canvas')
        app.width = implementation.naturalWidth
        app.height = implementation.naturalHeight
        app.getContext('2d').drawImage(implementation, 0, 0)
        const label = 34
        const gap = 24
        const side = document.createElement('canvas')
        side.width = ref.width + gap + app.width
        side.height = label + Math.max(ref.height, app.height)
        const sideContext = side.getContext('2d')
        sideContext.fillStyle = '#101b2b'
        sideContext.fillRect(0, 0, side.width, side.height)
        sideContext.fillStyle = '#b8955a'
        sideContext.font = '600 18px system-ui, sans-serif'
        sideContext.fillText('ATLAS (referencia)', 8, 23)
        sideContext.fillText('APP (implementación)', ref.width + gap + 8, 23)
        sideContext.drawImage(ref, 0, label)
        sideContext.drawImage(app, ref.width + gap, label)
        const widthCommon = Math.min(ref.width, app.width)
        const heightCommon = Math.min(ref.height, app.height)
        const left = ref.getContext('2d').getImageData(0, 0, widthCommon, heightCommon)
        const right = app.getContext('2d').getImageData(0, 0, widthCommon, heightCommon)
        const diff = document.createElement('canvas')
        diff.width = widthCommon
        diff.height = heightCommon
        const diffContext = diff.getContext('2d')
        const pixels = diffContext.createImageData(widthCommon, heightCommon)
        let changed = 0
        for (let index = 0; index < left.data.length; index += 4) {
          const delta = Math.max(
            Math.abs(left.data[index] - right.data[index]),
            Math.abs(left.data[index + 1] - right.data[index + 1]),
            Math.abs(left.data[index + 2] - right.data[index + 2]),
          )
          if (delta > 16) changed += 1
          pixels.data[index] = delta > 16 ? 255 : 0
          pixels.data[index + 1] = delta > 16 ? 40 : 0
          pixels.data[index + 2] = delta > 16 ? 40 : 0
          pixels.data[index + 3] = 255
        }
        diffContext.putImageData(pixels, 0, 0)
        return {
          side: side.toDataURL('image/png'),
          diff: diff.toDataURL('image/png'),
          changedRatio: changed / (widthCommon * heightCommon),
          reference: [ref.width, ref.height],
          implementation: [app.width, app.height],
        }
      },
      [dataUri(atlas), dataUri(actual), panel.crop],
    )
    const directory = resolve(output, panel.name)
    mkdirSync(directory, { recursive: true })
    const save = (name, uri) =>
      writeFileSync(resolve(directory, name), Buffer.from(uri.split(',')[1], 'base64'))
    save('guardado-lado-a-lado.png', result.side)
    save('guardado-diff.png', result.diff)
    report.push(
      `${panel.name}: ${result.changedRatio.toFixed(4)} · atlas ${result.reference.join('×')} · app ${result.implementation.join('×')}`,
    )
  }
  await browser.close()
  mkdirSync(output, { recursive: true })
  writeFileSync(resolve(output, 'resumen.txt'), `${report.join('\n')}\n`)
  console.log(report.join('\n'))
}

await main()
