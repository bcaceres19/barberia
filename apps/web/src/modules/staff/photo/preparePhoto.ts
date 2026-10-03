// Prepara la fotografía de un barbero en el navegador antes de enviarla
// (DEC-104): la recorta al centro en un cuadrado, la reduce a 512 px y la
// re-codifica como JPEG. Así el servidor recibe siempre una imagen pequeña
// (decenas de KiB), sin metadatos EXIF del dispositivo (ubicación, modelo de
// cámara) y con la orientación ya aplicada. El servidor la vuelve a validar:
// esto mejora la experiencia, no es la defensa.

/** Lado del cuadrado resultante. Un retrato de lista o de detalle no necesita
 * más, y con 512 px un JPEG de calidad 0,86 pesa muy por debajo del tope de
 * 512 KiB del servidor. */
export const PHOTO_OUTPUT_SIDE = 512
export const PHOTO_OUTPUT_TYPE = 'image/jpeg'
export const PHOTO_OUTPUT_QUALITY = 0.86

/** Tipos que el usuario puede elegir; WebP/HEIC se admiten de entrada porque el
 * navegador los decodifica y salen siempre como JPEG. */
export const PHOTO_ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/webp']
export const PHOTO_ACCEPT_ATTRIBUTE = PHOTO_ACCEPTED_TYPES.join(',')

/** Archivo de origen máximo (antes de reducirlo): una foto de móvil pesa unos
 * pocos MiB; por encima de esto casi seguro no es una fotografía. */
export const PHOTO_SOURCE_MAX_BYTES = 12 * 1024 * 1024

/** Lado mínimo aceptado por el servidor (staff.MinPhotoSide): una imagen más
 * pequeña se rechaza aquí con un mensaje claro en vez de esperar un 422. */
export const PHOTO_MIN_SIDE = 64

export type PreparePhotoFailure = 'unsupported-type' | 'too-large' | 'too-small' | 'unreadable'

export type PreparePhotoResult =
  { kind: 'ready'; blob: Blob; previewUrl: string } | { kind: 'error'; reason: PreparePhotoFailure }

/** Rectángulo de origen del recorte cuadrado centrado. Pura y exportada para
 * probarla sin canvas. */
export function centerSquare(
  width: number,
  height: number,
): { sx: number; sy: number; side: number } {
  const side = Math.min(width, height)
  return { sx: Math.round((width - side) / 2), sy: Math.round((height - side) / 2), side }
}

async function decode(file: File): Promise<{
  source: CanvasImageSource
  width: number
  height: number
  release: () => void
}> {
  // imageOrientation: 'from-image' aplica la rotación EXIF (una foto vertical
  // de móvil no sale de lado).
  const bitmap = await createImageBitmap(file, { imageOrientation: 'from-image' })
  return {
    source: bitmap,
    width: bitmap.width,
    height: bitmap.height,
    release: () => bitmap.close(),
  }
}

function toBlob(canvas: HTMLCanvasElement): Promise<Blob | null> {
  return new Promise((resolve) => canvas.toBlob(resolve, PHOTO_OUTPUT_TYPE, PHOTO_OUTPUT_QUALITY))
}

export async function preparePhoto(file: File): Promise<PreparePhotoResult> {
  if (!PHOTO_ACCEPTED_TYPES.includes(file.type)) {
    return { kind: 'error', reason: 'unsupported-type' }
  }
  if (file.size > PHOTO_SOURCE_MAX_BYTES) {
    return { kind: 'error', reason: 'too-large' }
  }

  try {
    const image = await decode(file)
    try {
      const { sx, sy, side } = centerSquare(image.width, image.height)
      if (side < PHOTO_MIN_SIDE) return { kind: 'error', reason: 'too-small' }
      // No se amplía una imagen pequeña: se conserva su resolución real.
      const output = Math.min(PHOTO_OUTPUT_SIDE, side)

      const canvas = document.createElement('canvas')
      canvas.width = output
      canvas.height = output
      const context = canvas.getContext('2d')
      if (!context) return { kind: 'error', reason: 'unreadable' }

      // Fondo blanco: un PNG con transparencia no debe salir negro en JPEG.
      context.fillStyle = '#ffffff'
      context.fillRect(0, 0, output, output)
      context.imageSmoothingQuality = 'high'
      context.drawImage(image.source, sx, sy, side, side, 0, 0, output, output)

      const blob = await toBlob(canvas)
      if (!blob) return { kind: 'error', reason: 'unreadable' }
      return { kind: 'ready', blob, previewUrl: URL.createObjectURL(blob) }
    } finally {
      image.release()
    }
  } catch {
    return { kind: 'error', reason: 'unreadable' }
  }
}
