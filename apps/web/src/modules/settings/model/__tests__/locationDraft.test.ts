import { describe, expect, it } from 'vitest'
import { DEFAULT_LOCATION, type LocationSettings } from '../brandOutcome'
import {
  addressChanged,
  draftFromSaved,
  locationChange,
  normalizeAddress,
  sameLocation,
  validateAddress,
} from '../locationDraft'

const saved: LocationSettings = {
  hasPhysicalLocation: true,
  address: 'Calle 10 # 5-20, Bogotá',
  latitude: 4.598056,
  longitude: -74.076111,
}

describe('locationDraft', () => {
  it('normaliza la dirección como el servidor: recorta y colapsa espacios', () => {
    expect(normalizeAddress('  Calle 10   # 5-20 \n Bogotá ')).toBe('Calle 10 # 5-20 Bogotá')
    expect(normalizeAddress('   ')).toBe('')
  })

  it('valida el largo máximo de 200 caracteres contando letras, no bytes', () => {
    expect(validateAddress('ñ'.repeat(200))).toBeUndefined()
    expect(validateAddress('ñ'.repeat(201))).toContain('200')
    expect(validateAddress('')).toBeUndefined()
  })

  it('el borrador de una barbería sin dirección empieza con el campo vacío', () => {
    expect(draftFromSaved(DEFAULT_LOCATION)).toEqual({ hasPhysicalLocation: true, address: '' })
  })

  it('una dirección sin cambios reales (solo espacios) no cuenta como cambio', () => {
    const draft = { hasPhysicalLocation: true, address: '  Calle 10   # 5-20, Bogotá ' }
    expect(addressChanged(draft, saved)).toBe(false)
    expect(sameLocation(draft, saved)).toBe(true)
    expect(locationChange(draft, saved)).toEqual({})
  })

  it('solo envía lo que cambió', () => {
    expect(
      locationChange({ hasPhysicalLocation: false, address: 'Calle 10 # 5-20, Bogotá' }, saved),
    ).toEqual({ hasPhysicalLocation: false })
    expect(
      locationChange({ hasPhysicalLocation: true, address: 'Carrera 7 # 32-16' }, saved),
    ).toEqual({ address: 'Carrera 7 # 32-16' })
  })

  it('una dirección vacía se envía como null para quitarla', () => {
    expect(locationChange({ hasPhysicalLocation: true, address: '  ' }, saved)).toEqual({
      address: null,
    })
  })

  it('vaciar un campo que ya estaba sin dirección no es un cambio', () => {
    expect(locationChange({ hasPhysicalLocation: true, address: '' }, DEFAULT_LOCATION)).toEqual({})
  })
})
