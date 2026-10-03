// Cambio de fotografía pendiente dentro de un diálogo (DEC-104): nada se envía
// al servidor hasta pulsar «Guardar». `new` lleva la imagen ya recortada y
// reducida (ver photo/preparePhoto.ts) y su URL de vista previa; `remove`
// pide quitar la fotografía actual.
export type PhotoDraft =
  { kind: 'none' } | { kind: 'new'; blob: Blob; previewUrl: string } | { kind: 'remove' }

export const NO_PHOTO_CHANGE: PhotoDraft = { kind: 'none' }
