/**
 * Pruebas de SettingsPage (HU-020): carga, edición, guardado exitoso
 * (actualiza la cabecera vía updateBarbershopName), error de validación y
 * error de red conservando lo escrito (CA-020-08), doble envío bloqueado.
 * settingsApi y auth.updateBarbershopName se sustituyen por dobles de
 * prueba; el recorrido real contra el API vive en el E2E de HU-020.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const fetchMock = vi.hoisted(() => vi.fn())
const saveMock = vi.hoisted(() => vi.fn())
const updateBarbershopNameMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/settingsApi', () => ({
  fetchBarbershopSettings: fetchMock,
  saveBarbershopSettings: saveMock,
}))
vi.mock('@/modules/auth', () => ({ updateBarbershopName: updateBarbershopNameMock }))

const { default: SettingsPage } = await import('../SettingsPage.vue')

const loadedSettings = {
  name: 'Barbería Ejemplo',
  timezone: 'America/Bogota',
  contactEmail: 'contacto@ejemplo.test',
  contactPhone: '+573001234567',
}

async function mountReady() {
  fetchMock.mockResolvedValueOnce({ kind: 'success', settings: loadedSettings })
  const wrapper = mount(SettingsPage)
  await flushPromises()
  return wrapper
}

describe('SettingsPage', () => {
  beforeEach(() => {
    fetchMock.mockReset()
    saveMock.mockReset()
    updateBarbershopNameMock.mockReset()
  })

  it('shows a non-blank loading state, then the loaded values', async () => {
    let resolveFetch: (value: unknown) => void = () => {}
    fetchMock.mockReturnValueOnce(new Promise((resolve) => (resolveFetch = resolve)))
    const wrapper = mount(SettingsPage)

    expect(wrapper.text()).toContain('Cargando')

    resolveFetch({ kind: 'success', settings: loadedSettings })
    await flushPromises()

    expect((wrapper.find('input[name="name"]').element as HTMLInputElement).value).toBe(
      'Barbería Ejemplo',
    )
    expect((wrapper.find('input[name="timezone"]').element as HTMLInputElement).value).toBe(
      'America/Bogota',
    )
    expect((wrapper.find('input[name="contactEmail"]').element as HTMLInputElement).value).toBe(
      'contacto@ejemplo.test',
    )
  })

  it('shows a recoverable error with Reintentar when the initial load fails', async () => {
    fetchMock.mockResolvedValueOnce({ kind: 'network-error' })
    const wrapper = mount(SettingsPage)
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos cargar la configuración')
    expect(fetchMock).toHaveBeenCalledTimes(1)

    fetchMock.mockResolvedValueOnce({ kind: 'success', settings: loadedSettings })
    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(wrapper.find('input[name="name"]').exists()).toBe(true)
  })

  it('renders empty contact fields (not "null") when the barbershop has none configured', async () => {
    fetchMock.mockResolvedValueOnce({
      kind: 'success',
      settings: { ...loadedSettings, contactEmail: null, contactPhone: null },
    })
    const wrapper = mount(SettingsPage)
    await flushPromises()

    expect((wrapper.find('input[name="contactEmail"]').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('input[name="contactPhone"]').element as HTMLInputElement).value).toBe('')
  })

  it('blocks submission and shows a field error for an empty name (client-side validation)', async () => {
    const wrapper = await mountReady()

    await wrapper.get('input[name="name"]').setValue('   ')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Escribe el nombre de la barbería.')
  })

  it('on success, updates the header via updateBarbershopName and shows a saved confirmation', async () => {
    const wrapper = await mountReady()
    saveMock.mockResolvedValueOnce({
      kind: 'success',
      settings: { ...loadedSettings, name: 'Barbería Renombrada' },
    })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).toHaveBeenCalledTimes(1)
    expect(updateBarbershopNameMock).toHaveBeenCalledWith('Barbería Renombrada')
    expect(wrapper.text()).toContain('Guardado')
    expect((wrapper.find('input[name="name"]').element as HTMLInputElement).value).toBe(
      'Barbería Renombrada',
    )
  })

  it('on a recoverable network error, keeps the typed values and never updates the header (CA-020-08)', async () => {
    const wrapper = await mountReady()
    await wrapper.get('input[name="name"]').setValue('Nombre editado sin guardar')
    saveMock.mockResolvedValueOnce({ kind: 'network-error' })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos conectar')
    expect(updateBarbershopNameMock).not.toHaveBeenCalled()
    expect((wrapper.find('input[name="name"]').element as HTMLInputElement).value).toBe(
      'Nombre editado sin guardar',
    )
  })

  it('on a validation-error (e.g. unrecognized timezone), keeps the typed values and never updates the header', async () => {
    const wrapper = await mountReady()
    await wrapper.get('input[name="timezone"]').setValue('COT')
    saveMock.mockResolvedValueOnce({ kind: 'validation-error' })

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('No pudimos guardar los cambios')
    expect(updateBarbershopNameMock).not.toHaveBeenCalled()
    expect((wrapper.find('input[name="timezone"]').element as HTMLInputElement).value).toBe('COT')
  })

  it('blocks a second submit while the first is still in flight (no double POST)', async () => {
    const wrapper = await mountReady()
    let resolveSave: (value: unknown) => void = () => {}
    saveMock.mockReturnValueOnce(new Promise((resolve) => (resolveSave = resolve)))

    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(saveMock).toHaveBeenCalledTimes(1)
    resolveSave({ kind: 'success', settings: loadedSettings })
    await flushPromises()
  })

  it('has no axe violations once loaded', async () => {
    const wrapper = await mountReady()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
