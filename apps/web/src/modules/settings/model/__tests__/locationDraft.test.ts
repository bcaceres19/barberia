import { describe, expect, it } from 'vitest'
import { DEFAULT_LOCATION, type LocationSettings } from '../brandOutcome'
import {
  addressChanged,
  draftFromSaved,
  locationChange,
  normalizeAddress,
  sameLocation,
  validateAddress,
  validateGooglePlaceId,
} from '../locationDraft'

const saved: LocationSettings = {
  hasPhysicalLocation: true,
  address: 'Calle 10 # 5-20, Bogotá',
  latitude: 4.598056,
  longitude: -74.076111,
  offersHomeService: false,
  googlePlaceId: null,
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
    expect(draftFromSaved(DEFAULT_LOCATION)).toEqual({
      hasPhysicalLocation: true,
      address: '',
      offersHomeService: false,
      googlePlaceId: '',
    })
  })

  it('una dirección sin cambios reales (solo espacios) no cuenta como cambio', () => {
    const draft = {
      hasPhysicalLocation: true,
      address: '  Calle 10   # 5-20, Bogotá ',
      offersHomeService: false,
      googlePlaceId: '',
    }
    expect(addressChanged(draft, saved)).toBe(false)
    expect(sameLocation(draft, saved)).toBe(true)
    expect(locationChange(draft, saved)).toEqual({})
  })

  it('solo envía lo que cambió', () => {
    expect(
      locationChange(
        {
          hasPhysicalLocation: false,
          address: 'Calle 10 # 5-20, Bogotá',
          offersHomeService: false,
          googlePlaceId: '',
        },
        saved,
      ),
    ).toEqual({ hasPhysicalLocation: false })
    expect(
      locationChange(
        {
          hasPhysicalLocation: true,
          address: 'Carrera 7 # 32-16',
          offersHomeService: false,
          googlePlaceId: '',
        },
        saved,
      ),
    ).toEqual({ address: 'Carrera 7 # 32-16' })
  })

  it('el interruptor de domicilio se envía solo si cambió (DEC-131)', () => {
    const base = { hasPhysicalLocation: true, address: 'Calle 10 # 5-20, Bogotá' }
    expect(locationChange({ ...base, offersHomeService: true, googlePlaceId: '' }, saved)).toEqual({
      offersHomeService: true,
    })
    expect(sameLocation({ ...base, offersHomeService: true, googlePlaceId: '' }, saved)).toBe(false)
    expect(locationChange({ ...base, offersHomeService: false, googlePlaceId: '' }, saved)).toEqual(
      {},
    )
  })

  it('una dirección vacía se envía como null para quitarla', () => {
    expect(
      locationChange(
        { hasPhysicalLocation: true, address: '  ', offersHomeService: false, googlePlaceId: '' },
        saved,
      ),
    ).toEqual({
      address: null,
    })
  })

  it('vaciar un campo que ya estaba sin dirección no es un cambio', () => {
    expect(
      locationChange(
        { hasPhysicalLocation: true, address: '', offersHomeService: false, googlePlaceId: '' },
        DEFAULT_LOCATION,
      ),
    ).toEqual({})
  })

  describe('ficha de Google (DEC-129, HU-101)', () => {
    const base = {
      hasPhysicalLocation: true,
      address: 'Calle 10 # 5-20, Bogotá',
      offersHomeService: false,
    }

    it('acepta letras, números, guion y guion bajo, y también vacío', () => {
      expect(validateGooglePlaceId('ChIJN1t_tDeuEmsRUsoyG83frY4')).toBeUndefined()
      expect(validateGooglePlaceId('  ')).toBeUndefined()
      expect(validateGooglePlaceId('')).toBeUndefined()
    })

    it('rechaza un enlace pegado, espacios intermedios o más de 200 caracteres', () => {
      expect(validateGooglePlaceId('https://maps.google.com/?cid=1')).toContain(
        'no pegues un enlace',
      )
      expect(validateGooglePlaceId('ChIJ abc')).toBeDefined()
      expect(validateGooglePlaceId('x'.repeat(201))).toBeDefined()
    })

    it('se envía recortado solo si cambió y vacío se envía como null para quitarla', () => {
      expect(locationChange({ ...base, googlePlaceId: ' ChIJabc ' }, saved)).toEqual({
        googlePlaceId: 'ChIJabc',
      })
      expect(sameLocation({ ...base, googlePlaceId: 'ChIJabc' }, saved)).toBe(false)
      const withListing = { ...saved, googlePlaceId: 'ChIJabc' }
      expect(locationChange({ ...base, googlePlaceId: 'ChIJabc ' }, withListing)).toEqual({})
      expect(locationChange({ ...base, googlePlaceId: '  ' }, withListing)).toEqual({
        googlePlaceId: null,
      })
      expect(locationChange({ ...base, googlePlaceId: '' }, saved)).toEqual({})
    })
  })
})
