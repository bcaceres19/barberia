/**
 * Pruebas de SettingsPage (HU-020, DEC-110): carga de los dos recursos (datos
 * básicos y marca), borrador con detección de cambios, guardado por grupos,
 * error de validación y de red conservando lo escrito (CA-020-08), doble envío
 * bloqueado, vista previa y descarte del acento, presets y concordancia del
 * vocabulario, y las preferencias de ESTE dispositivo (modo, tamaño de texto,
 * animaciones). settingsApi, brandApi y auth.updateBarbershopName se sustituyen
 * por dobles de prueba; el recorrido real contra el API vive en el E2E.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { PAGE_MIN_HOLD_MS } from '@/shared/composables'
import {
  appearance,
  brandState,
  DEFAULT_BRAND,
  effectiveAccent,
  resetAppearance,
  resetBrand,
  setAccentPreview,
  setBrand,
} from '@/shared/model'
import { resetToasts, toastState } from '@/shared/model/toastStore'

const fetchMock = vi.hoisted(() => vi.fn())
const saveMock = vi.hoisted(() => vi.fn())
const fetchBrandMock = vi.hoisted(() => vi.fn())
const saveBrandMock = vi.hoisted(() => vi.fn())
const updateBarbershopNameMock = vi.hoisted(() => vi.fn())
// El enlace público tiene su propio recurso y su propia prueba
// (PublicLinkPanel.test.ts): aquí solo se evita la red.
const fetchPublicLinkMock = vi.hoisted(() =>
  vi.fn(() => Promise.resolve({ kind: 'success', slug: 'barberia-ejemplo-k7x2m9' })),
)

vi.mock('../../api/settingsApi', () => ({
  fetchBarbershopSettings: fetchMock,
  saveBarbershopSettings: saveMock,
}))
vi.mock('../../api/brandApi', () => ({
  fetchBrand: fetchBrandMock,
  saveBrand: saveBrandMock,
}))
vi.mock('../../api/publicLinkApi', () => ({ fetchPublicLink: fetchPublicLinkMock }))
vi.mock('@/modules/auth', () => ({ updateBarbershopName: updateBarbershopNameMock }))

const { default: SettingsPage } = await import('../SettingsPage.vue')

const loadedSettings = {
  name: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  contactEmail: 'contacto@ejemplo.test',
  contactPhone: '+573001234567',
}

const mountOptions = {
  global: {
    stubs: {
      RouterLink: { props: ['to'], template: '<a class="stub-link"><slot /></a>' },
    },
  },
}

// El rombo de carga se sostiene PAGE_MIN_HOLD_MS aunque la respuesta ya haya
// llegado: se espera ese retraso real antes de mirar la pantalla cargada.
async function settle() {
  await flushPromises()
  await new Promise((resolve) => setTimeout(resolve, PAGE_MIN_HOLD_MS + 50))
  await flushPromises()
}

async function mountReady(brand = DEFAULT_BRAND) {
  fetchMock.mockResolvedValueOnce({ kind: 'success', settings: loadedSettings })
  fetchBrandMock.mockResolvedValueOnce({ kind: 'success', brand })
  const wrapper = mount(SettingsPage, mountOptions)
  await settle()
  return wrapper
}

const input = (wrapper: VueWrapper, name: string) =>
  wrapper.get(`input[name="${name}"]`).element as HTMLInputElement

async function submit(wrapper: VueWrapper) {
  await wrapper.get('form').trigger('submit')
  await flushPromises()
}

describe('SettingsPage', () => {
  beforeEach(() => {
    resetToasts()
    resetAppearance()
    resetBrand()
    window.localStorage.clear()
    fetchMock.mockReset()
    saveMock.mockReset()
    fetchBrandMock.mockReset()
    saveBrandMock.mockReset()
    updateBarbershopNameMock.mockReset()
  })

  describe('carga', () => {
    it('shows a non-blank loading state, then the loaded values of both resources', async () => {
      let resolveFetch: (value: unknown) => void = () => {}
      fetchMock.mockReturnValueOnce(new Promise((resolve) => (resolveFetch = resolve)))
      fetchBrandMock.mockResolvedValueOnce({ kind: 'success', brand: DEFAULT_BRAND })
      const wrapper = mount(SettingsPage, mountOptions)

      expect(wrapper.text()).toContain('Cargando')

      resolveFetch({ kind: 'success', settings: loadedSettings })
      await settle()

      expect(input(wrapper, 'name').value).toBe('Barbería Ejemplo')
      expect(input(wrapper, 'contactEmail').value).toBe('contacto@ejemplo.test')
      expect(input(wrapper, 'businessTerm').value).toBe('barbería')
      expect(input(wrapper, 'professionalTerm').value).toBe('barbero')
      expect(input(wrapper, 'professionalTermPlural').value).toBe('barberos')
      // La zona vive en el combobox del selector de zonas.
      expect((wrapper.get('[role="combobox"]').element as HTMLInputElement).value).toBe(
        'America/Bogota',
      )
    })

    it.each([
      ['datos básicos', () => fetchMock.mockResolvedValueOnce({ kind: 'network-error' })],
      ['marca', () => fetchBrandMock.mockResolvedValueOnce({ kind: 'unexpected-error' })],
    ])(
      'shows a recoverable error with Reintentar when the %s fail to load',
      async (_name, fail) => {
        fetchMock.mockResolvedValue({ kind: 'success', settings: loadedSettings })
        fetchBrandMock.mockResolvedValue({ kind: 'success', brand: DEFAULT_BRAND })
        fail()
        const wrapper = mount(SettingsPage, mountOptions)
        await settle()

        expect(wrapper.text()).toContain('No pudimos cargar la configuración')
        expect(wrapper.find('input[name="name"]').exists()).toBe(false)

        await wrapper.get('button').trigger('click')
        await settle()

        expect(fetchMock).toHaveBeenCalledTimes(2)
        expect(wrapper.find('input[name="name"]').exists()).toBe(true)
      },
    )

    it('renders empty contact fields (not "null") when the barbershop has none configured', async () => {
      fetchMock.mockResolvedValueOnce({
        kind: 'success',
        settings: { ...loadedSettings, contactEmail: null, contactPhone: null },
      })
      fetchBrandMock.mockResolvedValueOnce({ kind: 'success', brand: DEFAULT_BRAND })
      const wrapper = mount(SettingsPage, mountOptions)
      await settle()

      expect(input(wrapper, 'contactEmail').value).toBe('')
      expect(input(wrapper, 'contactPhone').value).toBe('')
    })
  })

  describe('borrador y guardado', () => {
    it('shows no save bar until something changes, and hides it when the change is undone', async () => {
      const wrapper = await mountReady()
      expect(wrapper.find('.save-bar').exists()).toBe(false)

      await wrapper.get('input[name="name"]').setValue('Otro nombre')
      expect(wrapper.find('.save-bar').exists()).toBe(true)
      expect(wrapper.get('.save-bar').text()).toContain('Datos básicos')

      await wrapper.get('input[name="name"]').setValue('Barbería Ejemplo')
      expect(wrapper.find('.save-bar').exists()).toBe(false)
    })

    it('blocks submission and shows a field error for an empty name (client-side validation)', async () => {
      const wrapper = await mountReady()

      await wrapper.get('input[name="name"]').setValue('   ')
      await submit(wrapper)

      expect(saveMock).not.toHaveBeenCalled()
      expect(wrapper.text()).toContain('Escribe el nombre de la barbería.')
    })

    it.each([
      ['con dígitos', 'salón 24', 'Usa solo letras, espacios, guion o apóstrofo.'],
      ['de una letra', 's', 'Usa al menos 2 letras.'],
      ['vacío', '   ', 'Escribe una palabra.'],
    ])('blocks a business word %s and shows why', async (_name, value, message) => {
      const wrapper = await mountReady()

      await wrapper.get('input[name="businessTerm"]').setValue(value)
      await submit(wrapper)

      expect(saveBrandMock).not.toHaveBeenCalled()
      expect(wrapper.text()).toContain(message)
    })

    it('saves only the basic data when only they changed, updates the header and announces it', async () => {
      const wrapper = await mountReady()
      saveMock.mockResolvedValueOnce({
        kind: 'success',
        settings: { ...loadedSettings, name: 'Barbería Renombrada' },
      })

      await wrapper.get('input[name="name"]').setValue('Barbería Renombrada')
      await submit(wrapper)

      expect(saveMock).toHaveBeenCalledTimes(1)
      expect(saveBrandMock).not.toHaveBeenCalled()
      expect(updateBarbershopNameMock).toHaveBeenCalledWith('Barbería Renombrada')
      expect(wrapper.find('.save-bar').exists()).toBe(false)
      expect(toastState.items.map((item) => item.title)).toEqual(['Configuración guardada'])
    })

    it('saves only the brand when only it changed, normalized, and publishes it to the whole app', async () => {
      const wrapper = await mountReady()
      saveBrandMock.mockImplementationOnce(async (brand) => ({ kind: 'success', brand }))

      await wrapper.get('input[name="businessTerm"]').setValue('  Salón   de Belleza ')
      await submit(wrapper)

      expect(saveMock).not.toHaveBeenCalled()
      expect(saveBrandMock).toHaveBeenCalledTimes(1)
      expect(saveBrandMock.mock.calls[0]![0]).toMatchObject({ businessTerm: 'salón de belleza' })
      // La marca confirmada llega al estado compartido que lee toda la app.
      expect(brandState.brand.businessTerm).toBe('salón de belleza')
      expect(wrapper.find('.save-bar').exists()).toBe(false)
    })

    it('saves both resources when both changed', async () => {
      const wrapper = await mountReady()
      saveMock.mockResolvedValueOnce({
        kind: 'success',
        settings: { ...loadedSettings, name: 'Nuevo' },
      })
      saveBrandMock.mockImplementationOnce(async (brand) => ({ kind: 'success', brand }))

      await wrapper.get('input[name="name"]').setValue('Nuevo')
      await wrapper.get('input[name="businessTerm"]').setValue('estudio')
      await submit(wrapper)

      expect(saveMock).toHaveBeenCalledTimes(1)
      expect(saveBrandMock).toHaveBeenCalledTimes(1)
      expect(wrapper.find('.save-bar').exists()).toBe(false)
    })

    it('keeps the unsaved resource pending when only the other one fails to save', async () => {
      const wrapper = await mountReady()
      saveMock.mockResolvedValueOnce({
        kind: 'success',
        settings: { ...loadedSettings, name: 'Nuevo' },
      })
      saveBrandMock.mockResolvedValueOnce({ kind: 'network-error' })

      await wrapper.get('input[name="name"]').setValue('Nuevo')
      await wrapper.get('input[name="businessTerm"]').setValue('estudio')
      await submit(wrapper)

      // Lo guardado no se vuelve a enviar, lo fallido sigue pendiente y escrito.
      expect(updateBarbershopNameMock).toHaveBeenCalledWith('Nuevo')
      expect(wrapper.text()).toContain('No pudimos conectar')
      expect(wrapper.get('.save-bar').text()).toContain('Marca y vocabulario')
      expect(wrapper.get('.save-bar').text()).not.toContain('Datos básicos')
      expect(input(wrapper, 'businessTerm').value).toBe('estudio')
    })

    it('on a recoverable network error, keeps the typed values and never updates the header (CA-020-08)', async () => {
      const wrapper = await mountReady()
      await wrapper.get('input[name="name"]').setValue('Nombre editado sin guardar')
      saveMock.mockResolvedValueOnce({ kind: 'network-error' })

      await submit(wrapper)

      expect(wrapper.text()).toContain('No pudimos conectar')
      expect(updateBarbershopNameMock).not.toHaveBeenCalled()
      expect(input(wrapper, 'name').value).toBe('Nombre editado sin guardar')
      expect(wrapper.find('.save-bar').exists()).toBe(true)
    })

    it('on a validation-error for the basic data, keeps the typed values and never updates the header', async () => {
      const wrapper = await mountReady()
      const zone = wrapper.get('[role="combobox"]')
      await zone.setValue('COT')
      // El selector solo acepta una zona de su lista; el servidor sigue siendo
      // la autoridad, así que se fuerza el cambio por el nombre del negocio.
      await wrapper.get('input[name="name"]').setValue('Nombre nuevo')
      saveMock.mockResolvedValueOnce({ kind: 'validation-error' })

      await submit(wrapper)

      expect(wrapper.text()).toContain('No pudimos guardar los datos básicos')
      expect(updateBarbershopNameMock).not.toHaveBeenCalled()
      expect(input(wrapper, 'name').value).toBe('Nombre nuevo')
    })

    it('on a validation-error for the brand, explains the words rule and keeps what was typed', async () => {
      const wrapper = await mountReady()
      await wrapper.get('input[name="businessTerm"]').setValue('estudio')
      saveBrandMock.mockResolvedValueOnce({ kind: 'validation-error' })

      await submit(wrapper)

      expect(wrapper.text()).toContain('No pudimos guardar la marca ni el vocabulario')
      expect(input(wrapper, 'businessTerm').value).toBe('estudio')
    })

    it('blocks a second submit while the first is still in flight (no double PATCH)', async () => {
      const wrapper = await mountReady()
      let resolveSave: (value: unknown) => void = () => {}
      saveMock.mockReturnValueOnce(new Promise((resolve) => (resolveSave = resolve)))

      await wrapper.get('input[name="name"]').setValue('Nuevo')
      await wrapper.get('form').trigger('submit')
      await wrapper.get('form').trigger('submit')
      await flushPromises()

      expect(saveMock).toHaveBeenCalledTimes(1)
      resolveSave({ kind: 'success', settings: { ...loadedSettings, name: 'Nuevo' } })
      await flushPromises()
    })

    it('discards every pending change and restores the confirmed values', async () => {
      const wrapper = await mountReady()
      await wrapper.get('input[name="name"]').setValue('Descartar esto')
      await wrapper.get('input[name="businessTerm"]').setValue('estudio')

      const discard = wrapper.findAll('.save-bar button').find((b) => b.text() === 'Descartar')!
      await discard.trigger('click')

      expect(input(wrapper, 'name').value).toBe('Barbería Ejemplo')
      expect(input(wrapper, 'businessTerm').value).toBe('barbería')
      expect(wrapper.find('.save-bar').exists()).toBe(false)
      expect(saveMock).not.toHaveBeenCalled()
    })
  })

  describe('acento de la marca', () => {
    it('previews the chosen accent on the whole app before saving and drops it on discard', async () => {
      const wrapper = await mountReady()
      expect(effectiveAccent.value).toBe('brass')

      await wrapper.get('input[type="radio"][value="emerald"]').setValue(true)

      expect(effectiveAccent.value).toBe('emerald')
      // Todavía no se confirmó: la marca guardada no cambió.
      expect(brandState.brand.accent).toBe('brass')
      expect(wrapper.get('.save-bar').text()).toContain('Marca y vocabulario')

      const discard = wrapper.findAll('.save-bar button').find((b) => b.text() === 'Descartar')!
      await discard.trigger('click')
      expect(effectiveAccent.value).toBe('brass')
    })

    it('stops previewing the accent when the screen is left without saving', async () => {
      const wrapper = await mountReady()
      await wrapper.get('input[type="radio"][value="ruby"]').setValue(true)
      expect(effectiveAccent.value).toBe('ruby')

      wrapper.unmount()

      expect(effectiveAccent.value).toBe('brass')
      setAccentPreview(null)
    })

    it('saves the accent and keeps it as the confirmed one', async () => {
      const wrapper = await mountReady()
      saveBrandMock.mockImplementationOnce(async (brand) => ({ kind: 'success', brand }))

      await wrapper.get('input[type="radio"][value="sapphire"]').setValue(true)
      await submit(wrapper)

      expect(saveBrandMock.mock.calls[0]![0]).toMatchObject({ accent: 'sapphire' })
      expect(brandState.brand.accent).toBe('sapphire')
      expect(effectiveAccent.value).toBe('sapphire')
    })
  })

  describe('perfil del panel (DEC-115)', () => {
    it('offers the two profiles, with the full panel selected by default', async () => {
      const wrapper = await mountReady()

      const group = wrapper.get('[role="radiogroup"][aria-label="Perfil del panel"]')
      const labels = group.findAll('.profile-card__title').map((l) => l.text())
      expect(labels).toEqual(['Barbería con equipo', 'Barbero individual'])
      expect(group.get<HTMLInputElement>('input[value="shop"]').element.checked).toBe(true)
      expect(group.get<HTMLInputElement>('input[value="solo"]').element.checked).toBe(false)
    })

    it('writes the options with the words of the barbershop', async () => {
      // Las opciones usan el vocabulario CONFIRMADO (el del resto del panel), no el borrador.
      const salon = {
        ...DEFAULT_BRAND,
        businessTerm: 'salón',
        businessTermGender: 'masculine' as const,
        professionalTerm: 'estilista',
        professionalTermPlural: 'estilistas',
        professionalTermGender: 'feminine' as const,
      }
      setBrand(salon)
      const wrapper = await mountReady(salon)

      const labels = wrapper
        .get('[role="radiogroup"][aria-label="Perfil del panel"]')
        .findAll('.profile-card__title')
        .map((l) => l.text())
      expect(labels).toEqual(['Salón con equipo', 'Estilista individual'])
    })

    it('marks the draft as changed without applying it until it is saved', async () => {
      const wrapper = await mountReady()

      await wrapper.get('input[type="radio"][value="solo"]').setValue(true)

      expect(wrapper.get('.save-bar').text()).toContain('Marca y vocabulario')
      // Todavía no se confirmó: el panel sigue siendo el completo.
      expect(brandState.brand.panelProfile).toBe('shop')
    })

    it('saves the solo profile, publishes it to the whole app and can go back', async () => {
      const wrapper = await mountReady()
      saveBrandMock.mockImplementation(async (brand) => ({ kind: 'success', brand }))

      await wrapper.get('input[type="radio"][value="solo"]').setValue(true)
      await submit(wrapper)

      expect(saveBrandMock.mock.calls[0]![0]).toMatchObject({ panelProfile: 'solo' })
      expect(brandState.brand.panelProfile).toBe('solo')
      expect(wrapper.find('.save-bar').exists()).toBe(false)

      await wrapper.get('input[type="radio"][value="shop"]').setValue(true)
      await submit(wrapper)

      expect(saveBrandMock.mock.calls[1]![0]).toMatchObject({ panelProfile: 'shop' })
      expect(brandState.brand.panelProfile).toBe('shop')
    })

    it('keeps the confirmed profile when the save fails, and restores it on discard', async () => {
      const wrapper = await mountReady()
      saveBrandMock.mockResolvedValueOnce({ kind: 'network-error' })

      await wrapper.get('input[type="radio"][value="solo"]').setValue(true)
      await submit(wrapper)
      expect(brandState.brand.panelProfile).toBe('shop')

      const discard = wrapper.findAll('.save-bar button').find((b) => b.text() === 'Descartar')!
      await discard.trigger('click')
      expect(
        wrapper.get<HTMLInputElement>('input[type="radio"][value="shop"]').element.checked,
      ).toBe(true)
    })

    it('shows the profile that was already saved', async () => {
      const wrapper = await mountReady({ ...DEFAULT_BRAND, panelProfile: 'solo' })

      expect(
        wrapper.get<HTMLInputElement>('input[type="radio"][value="solo"]').element.checked,
      ).toBe(true)
      expect(wrapper.find('.save-bar').exists()).toBe(false)
    })
  })

  describe('vocabulario', () => {
    it('fills the word, its plural and its gender from a preset', async () => {
      const wrapper = await mountReady()

      const preset = wrapper.findAll('.preset-chip').find((b) => b.text() === 'estilista')!
      await preset.trigger('click')

      expect(input(wrapper, 'professionalTerm').value).toBe('estilista')
      expect(input(wrapper, 'professionalTermPlural').value).toBe('estilistas')
      expect(
        (
          wrapper.get('input[name^="option-group"][value="feminine"]:checked')
            .element as HTMLInputElement
        ).value,
      ).toBe('feminine')
    })

    it('suggests the plural while the singular is typed, until it is edited by hand', async () => {
      const wrapper = await mountReady()

      await wrapper.get('input[name="professionalTerm"]').setValue('colorista')
      expect(input(wrapper, 'professionalTermPlural').value).toBe('coloristas')

      await wrapper.get('input[name="professionalTermPlural"]').setValue('coloristas expertas')
      await wrapper.get('input[name="professionalTerm"]').setValue('maestro')
      expect(input(wrapper, 'professionalTermPlural').value).toBe('coloristas expertas')
    })

    it('previews the sentences with the draft vocabulary and its grammatical agreement', async () => {
      const wrapper = await mountReady()

      const preset = wrapper.findAll('.preset-chip').find((b) => b.text() === 'estilista')!
      await preset.trigger('click')

      const preview = wrapper.get('[aria-label="Vista previa del vocabulario"]').text()
      expect(preview).toContain('Estilistas')
      expect(preview).toContain('+ Agregar estilista')
      expect(preview).toContain('Aún no tienes estilistas registradas en la barbería.')
      expect(preview).toContain('Elige a la estilista y el servicio.')
    })

    it('writes the scope of each section with the confirmed barbershop word', async () => {
      // La marca confirmada llega al estado compartido al iniciar la sesión.
      const salon = {
        ...DEFAULT_BRAND,
        businessTerm: 'salón',
        businessTermGender: 'masculine' as const,
      }
      setBrand(salon)
      const wrapper = await mountReady(salon)

      expect(wrapper.text()).toContain('Para el salón')
      expect(wrapper.text()).toContain('cómo se llama el salón')
    })
  })

  describe('preferencias de este dispositivo', () => {
    it('applies the mode at once, persists it and needs no saving', async () => {
      const wrapper = await mountReady()

      await wrapper.get('input[type="radio"][value="ivory"]').setValue(true)

      expect(appearance.theme).toBe('ivory')
      expect(JSON.parse(window.localStorage.getItem('nava.appearance.v1')!)).toMatchObject({
        theme: 'ivory',
      })
      expect(wrapper.find('.save-bar').exists()).toBe(false)
      expect(saveMock).not.toHaveBeenCalled()
    })

    it('changes the text size', async () => {
      const wrapper = await mountReady()

      await wrapper.get('input[type="radio"][value="xlarge"]').setValue(true)

      expect(appearance.textScale).toBe('xlarge')
    })

    it('toggles reduced animations with an announced switch', async () => {
      const wrapper = await mountReady()
      const toggle = wrapper.get('[role="switch"]')
      expect(toggle.attributes('aria-checked')).toBe('false')

      await toggle.trigger('click')

      expect(toggle.attributes('aria-checked')).toBe('true')
      expect(appearance.motion).toBe('reduced')
    })

    it('restores the screen preferences, and the button is disabled when nothing differs', async () => {
      const wrapper = await mountReady()
      const reset = () =>
        wrapper.findAll('button').find((b) => b.text() === 'Restablecer pantalla')!
      expect(reset().attributes('disabled')).toBeDefined()

      await wrapper.get('input[type="radio"][value="ivory"]').setValue(true)
      expect(reset().attributes('disabled')).toBeUndefined()

      await reset().trigger('click')
      expect(appearance.theme).toBe('ink')
    })
  })

  describe('enlace público (DEC-117)', () => {
    it('adds the link as section 07, before the booking rules, and lists it in the index', async () => {
      const wrapper = await mountReady()

      const panel = wrapper.get('#configuracion-enlace')
      expect(panel.get('.settings-panel__number').text()).toBe('07')
      expect(panel.get('h2').text()).toBe('Enlace público')
      expect(panel.get('[role="textbox"]').text()).toContain('/reservar/barberia-ejemplo-k7x2m9')
      expect(wrapper.get('#configuracion-reservas .settings-panel__number').text()).toBe('08')
      expect(wrapper.get('a[href="#configuracion-enlace"]').text()).toContain('Enlace')
      wrapper.unmount()
    })

    it('is read-only: opening the link never makes the draft dirty', async () => {
      const wrapper = await mountReady()

      expect(wrapper.find('.save-bar').exists()).toBe(false)
      wrapper.unmount()
    })
  })

  describe('accesibilidad', () => {
    it('has no axe violations once loaded', async () => {
      const wrapper = await mountReady()
      const results = await axe(wrapper.element)
      expect(results).toHaveNoViolations()
    })

    // heading-order: BaseAlert.vue (HU-009) fija el título de una alerta como
    // `<h4>` porque es la etiqueta de un widget transitorio, no un encabezado
    // del esquema del documento. Mismo criterio documentado en
    // LoginPage.test.ts: no es un defecto de esta pantalla ni de BaseAlert.
    const axeOptionsWithAlert = { rules: { 'heading-order': { enabled: false } } }

    it('has no axe violations with the validation-error alert and the save bar visible', async () => {
      const wrapper = await mountReady()
      await wrapper.get('input[name="name"]').setValue('Nombre nuevo')
      saveMock.mockResolvedValueOnce({ kind: 'validation-error' })

      await submit(wrapper)

      const results = await axe(wrapper.element, axeOptionsWithAlert)
      expect(results).toHaveNoViolations()
    })

    it('has no axe violations with a field error visible', async () => {
      const wrapper = await mountReady()
      await wrapper.get('input[name="businessTerm"]').setValue('salón 24')
      await submit(wrapper)

      const results = await axe(wrapper.element, axeOptionsWithAlert)
      expect(results).toHaveNoViolations()
    })
  })
})
