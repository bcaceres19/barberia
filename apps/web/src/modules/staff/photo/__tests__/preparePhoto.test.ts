/**
 * Pruebas de la preparación de la fotografía (DEC-104). jsdom no decodifica
 * imágenes ni implementa canvas, así que el recorrido de decodificación se
 * sustituye por dobles (createImageBitmap, getContext, toBlob) y se verifica lo
 * que esta función decide: qué archivos rechaza, qué recorte calcula, qué
 * tamaño y tipo produce y que libera lo que abre.
 */
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import {
  PHOTO_MIN_SIDE,
  PHOTO_OUTPUT_QUALITY,
  PHOTO_OUTPUT_SIDE,
  PHOTO_OUTPUT_TYPE,
  PHOTO_SOURCE_MAX_BYTES,
  centerSquare,
  preparePhoto,
} from '../preparePhoto'

function file(type: string, size = 1024): File {
  return new File([new Uint8Array(size)], 'foto', { type })
}

describe('centerSquare', () => {
  it('crops a landscape image to its centered square', () => {
    expect(centerSquare(1600, 900)).toEqual({ sx: 350, sy: 0, side: 900 })
  })

  it('crops a portrait image to its centered square', () => {
    expect(centerSquare(900, 1600)).toEqual({ sx: 0, sy: 350, side: 900 })
  })

  it('keeps an already square image whole', () => {
    expect(centerSquare(600, 600)).toEqual({ sx: 0, sy: 0, side: 600 })
  })
})

describe('preparePhoto', () => {
  const close = vi.fn()
  const drawImage = vi.fn()
  const fillRect = vi.fn()
  let bitmapSize = { width: 1600, height: 900 }
  let toBlobResult: Blob | null = new Blob(['x'], { type: PHOTO_OUTPUT_TYPE })
  let canvasContext: unknown

  beforeEach(() => {
    close.mockReset()
    drawImage.mockReset()
    fillRect.mockReset()
    bitmapSize = { width: 1600, height: 900 }
    toBlobResult = new Blob(['x'], { type: PHOTO_OUTPUT_TYPE })
    canvasContext = { drawImage, fillRect, fillStyle: '', imageSmoothingQuality: 'low' }

    vi.stubGlobal(
      'createImageBitmap',
      vi.fn(async () => ({ ...bitmapSize, close })),
    )
    vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockImplementation(
      () => canvasContext as CanvasRenderingContext2D,
    )
    vi.spyOn(HTMLCanvasElement.prototype, 'toBlob').mockImplementation(function (callback) {
      callback(toBlobResult)
    })
    vi.stubGlobal('URL', Object.assign(URL, { createObjectURL: vi.fn(() => 'blob:preview') }))
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
  })

  it.each(['image/gif', 'application/pdf', 'text/html', ''])(
    'rejects an unsupported type (%s) without decoding it',
    async (type) => {
      expect(await preparePhoto(file(type))).toEqual({ kind: 'error', reason: 'unsupported-type' })
      expect(createImageBitmap).not.toHaveBeenCalled()
    },
  )

  it('rejects a source file over the limit', async () => {
    const result = await preparePhoto(file('image/jpeg', PHOTO_SOURCE_MAX_BYTES + 1))
    expect(result).toEqual({ kind: 'error', reason: 'too-large' })
  })

  it('rejects an image smaller than the server minimum', async () => {
    bitmapSize = { width: PHOTO_MIN_SIDE - 1, height: 300 }
    expect(await preparePhoto(file('image/png'))).toEqual({ kind: 'error', reason: 'too-small' })
    expect(close).toHaveBeenCalledTimes(1)
  })

  it('crops to the center, downsizes to the output side and re-encodes as JPEG', async () => {
    const result = await preparePhoto(file('image/png'))

    expect(result).toEqual({ kind: 'ready', blob: toBlobResult, previewUrl: 'blob:preview' })
    // Recorte cuadrado centrado de 900 px de un 1600 × 900, dibujado a 512 × 512.
    expect(drawImage).toHaveBeenCalledWith(
      expect.anything(),
      350,
      0,
      900,
      900,
      0,
      0,
      PHOTO_OUTPUT_SIDE,
      PHOTO_OUTPUT_SIDE,
    )
    expect(HTMLCanvasElement.prototype.toBlob).toHaveBeenCalledWith(
      expect.any(Function),
      PHOTO_OUTPUT_TYPE,
      PHOTO_OUTPUT_QUALITY,
    )
    // Fondo blanco antes de dibujar: un PNG transparente no debe salir negro.
    expect(fillRect).toHaveBeenCalled()
    expect(close).toHaveBeenCalledTimes(1)
  })

  it('never upscales a small image beyond its real resolution', async () => {
    bitmapSize = { width: 200, height: 300 }
    await preparePhoto(file('image/jpeg'))
    expect(drawImage).toHaveBeenCalledWith(expect.anything(), 0, 50, 200, 200, 0, 0, 200, 200)
  })

  it('reports unreadable when the browser cannot decode the file', async () => {
    vi.stubGlobal(
      'createImageBitmap',
      vi.fn(async () => {
        throw new Error('decode')
      }),
    )
    expect(await preparePhoto(file('image/jpeg'))).toEqual({ kind: 'error', reason: 'unreadable' })
  })

  it('reports unreadable when the canvas cannot encode the result', async () => {
    toBlobResult = null
    expect(await preparePhoto(file('image/jpeg'))).toEqual({ kind: 'error', reason: 'unreadable' })
    expect(close).toHaveBeenCalledTimes(1)
  })

  it('reports unreadable when there is no 2D context', async () => {
    canvasContext = null
    expect(await preparePhoto(file('image/jpeg'))).toEqual({ kind: 'error', reason: 'unreadable' })
  })
})
