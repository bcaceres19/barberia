/**
 * Pruebas de BarberPhotoField (DEC-104): elegir y soltar una foto, quitarla,
 * deshacer, errores de preparación, liberación de la vista previa y
 * accesibilidad. La preparación real (canvas) se sustituye por un doble: aquí
 * se verifica el contrato del campo, no el recorte.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import type { PhotoDraft } from '../../model/photoDraft'

const prepareMock = vi.hoisted(() => vi.fn())

vi.mock('../../photo/preparePhoto', () => ({
  PHOTO_ACCEPT_ATTRIBUTE: 'image/jpeg,image/png,image/webp',
  preparePhoto: prepareMock,
}))

const { default: BarberPhotoField } = await import('../BarberPhotoField.vue')

const revokeMock = vi.fn()
const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

function mountField(
  props: Partial<{ currentUrl: string | null; draft: PhotoDraft; disabled: boolean }> = {},
) {
  return mount(BarberPhotoField, {
    props: { fullName: 'Carlos Ramírez', currentUrl: null, draft: { kind: 'none' }, ...props },
  })
}

function chooseFile(wrapper: ReturnType<typeof mountField>, file: File) {
  const input = wrapper.get('[data-testid="barber-photo-input"]').element as HTMLInputElement
  Object.defineProperty(input, 'files', { value: [file], configurable: true })
  return wrapper.get('[data-testid="barber-photo-input"]').trigger('change')
}

function buttonByText(wrapper: ReturnType<typeof mountField>, text: string) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

const jpeg = new File(['x'], 'foto.jpg', { type: 'image/jpeg' })

describe('BarberPhotoField', () => {
  beforeEach(() => {
    prepareMock.mockReset()
    revokeMock.mockReset()
    URL.revokeObjectURL = revokeMock
  })

  it('offers "Elegir foto" and shows the monogram while there is no photo', () => {
    const wrapper = mountField()
    expect(buttonByText(wrapper, 'Elegir foto')).toBeTruthy()
    expect(buttonByText(wrapper, 'Quitar foto')).toBeUndefined()
    expect(wrapper.get('.barber-avatar__monogram').text()).toBe('CR')
  })

  it('shows the current photo and offers to change or remove it', () => {
    const wrapper = mountField({ currentUrl: '/photo?v=1' })
    expect(wrapper.get('.barber-avatar__photo').attributes('src')).toBe('/photo?v=1')
    expect(buttonByText(wrapper, 'Cambiar foto')).toBeTruthy()
    expect(buttonByText(wrapper, 'Quitar foto')).toBeTruthy()
    expect(buttonByText(wrapper, 'Deshacer')).toBeUndefined()
  })

  it('opens the hidden file input from the visible button', async () => {
    const wrapper = mountField()
    const click = vi.spyOn(
      wrapper.get('[data-testid="barber-photo-input"]').element as HTMLInputElement,
      'click',
    )
    await buttonByText(wrapper, 'Elegir foto')!.trigger('click')
    expect(click).toHaveBeenCalledTimes(1)
  })

  it('emits a "new" draft with the prepared blob and preview URL', async () => {
    const blob = new Blob(['prepared'], { type: 'image/jpeg' })
    prepareMock.mockResolvedValueOnce({ kind: 'ready', blob, previewUrl: 'blob:preview-1' })
    const wrapper = mountField()

    await chooseFile(wrapper, jpeg)
    await flushPromises()

    expect(prepareMock).toHaveBeenCalledWith(jpeg)
    expect(wrapper.emitted('update:draft')![0]).toEqual([
      { kind: 'new', blob, previewUrl: 'blob:preview-1' },
    ])
  })

  it('shows the pending new photo and explains it will be saved with the form', async () => {
    const wrapper = mountField({
      currentUrl: '/photo?v=1',
      draft: { kind: 'new', blob: new Blob(['x']), previewUrl: 'blob:preview-2' },
    })
    expect(wrapper.get('.barber-avatar__photo').attributes('src')).toBe('blob:preview-2')
    expect(wrapper.text()).toContain('Se guardará al pulsar «Guardar»')
    expect(buttonByText(wrapper, 'Deshacer')).toBeTruthy()
  })

  it('reports a preparation failure in an alert and emits nothing', async () => {
    prepareMock.mockResolvedValueOnce({ kind: 'error', reason: 'unsupported-type' })
    const wrapper = mountField()

    await chooseFile(wrapper, new File(['x'], 'a.gif', { type: 'image/gif' }))
    await flushPromises()

    expect(wrapper.emitted('update:draft')).toBeUndefined()
    const alert = wrapper.get('[role="alert"]')
    expect(alert.text()).toContain('JPG, PNG o WebP')
  })

  it('asks to remove the current photo (draft "remove") and can undo it', async () => {
    const wrapper = mountField({ currentUrl: '/photo?v=1' })
    await buttonByText(wrapper, 'Quitar foto')!.trigger('click')
    expect(wrapper.emitted('update:draft')![0]).toEqual([{ kind: 'remove' }])

    await wrapper.setProps({ draft: { kind: 'remove' } })
    expect(wrapper.find('.barber-avatar__photo').exists()).toBe(false)
    expect(wrapper.text()).toContain('La foto se quitará al pulsar «Guardar»')

    await buttonByText(wrapper, 'Deshacer')!.trigger('click')
    expect(wrapper.emitted('update:draft')![1]).toEqual([{ kind: 'none' }])
  })

  it('discards a picked photo straight to "none" when there was no photo before', async () => {
    const wrapper = mountField({
      draft: { kind: 'new', blob: new Blob(['x']), previewUrl: 'blob:preview-3' },
    })
    await buttonByText(wrapper, 'Quitar foto')!.trigger('click')
    expect(wrapper.emitted('update:draft')![0]).toEqual([{ kind: 'none' }])
    expect(buttonByText(wrapper, 'Deshacer')).toBeUndefined()
  })

  it('accepts a photo dropped on the frame', async () => {
    const blob = new Blob(['prepared'], { type: 'image/jpeg' })
    prepareMock.mockResolvedValueOnce({ kind: 'ready', blob, previewUrl: 'blob:dropped' })
    const wrapper = mountField()

    const frame = wrapper.get('.barber-photo-field__frame')
    await frame.trigger('dragover')
    expect(frame.classes()).toContain('barber-photo-field__frame--dragging')

    await frame.trigger('drop', { dataTransfer: { files: [jpeg] } })
    await flushPromises()

    expect(wrapper.emitted('update:draft')![0]).toEqual([
      { kind: 'new', blob, previewUrl: 'blob:dropped' },
    ])
    expect(frame.classes()).not.toContain('barber-photo-field__frame--dragging')
  })

  it('ignores a drop and disables every control while disabled', async () => {
    const wrapper = mountField({ currentUrl: '/photo?v=1', disabled: true })
    for (const button of wrapper.findAll('button')) {
      expect(button.attributes('disabled')).toBeDefined()
    }
    await wrapper
      .get('.barber-photo-field__frame')
      .trigger('drop', { dataTransfer: { files: [jpeg] } })
    expect(prepareMock).not.toHaveBeenCalled()
  })

  it('releases the preview URL when the draft is replaced and when unmounted', async () => {
    const wrapper = mountField({
      draft: { kind: 'new', blob: new Blob(['a']), previewUrl: 'blob:a' },
    })
    await wrapper.setProps({ draft: { kind: 'none' } })
    expect(revokeMock).toHaveBeenCalledWith('blob:a')

    await wrapper.setProps({ draft: { kind: 'new', blob: new Blob(['b']), previewUrl: 'blob:b' } })
    wrapper.unmount()
    expect(revokeMock).toHaveBeenCalledWith('blob:b')
  })

  it('has no obvious accessibility violations', async () => {
    const wrapper = mountField({ currentUrl: '/photo?v=1' })
    expect(await axe(wrapper.element, axeOptions)).toHaveNoViolations()
  })
})
