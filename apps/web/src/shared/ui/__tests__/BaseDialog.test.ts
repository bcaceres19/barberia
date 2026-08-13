/**
 * Tests for BaseDialog component
 * Verifies: open/close, focus trap, ESC, backdrop click, sizes, accessibility
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { defineComponent, nextTick, ref } from 'vue'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import BaseDialog, { openDialogStack } from '../BaseDialog.vue'

// BaseDialog usa Teleport a document.body: axe debe analizar el overlay
// teletransportado, no el elemento raíz (vacío) del wrapper de VTU. region
// se desactiva porque el fragmento no tiene el resto de la página alrededor.
// color-contrast se desactiva: jsdom no implementa Canvas2D (ver
// BaseButton.test.ts); el contraste ya está verificado en la tabla
// aprobada de estandar-diseno-visual.md §4.3.
const axeOptions = { rules: { region: { enabled: false }, 'color-contrast': { enabled: false } } }

// useId() genera ids únicos DENTRO de una misma instancia de aplicación
// Vue; mount() de Vue Test Utils crea una app aislada por llamada, así que
// dos mount(BaseDialog) SEPARADOS no comparten secuencia de useId (cada uno
// reinicia su propio contador) y no reproducen el escenario real de dos
// diálogos dentro de la MISMA aplicación. TwoDialogsHost los monta como
// hermanos en un único árbol, igual que ocurriría en producción.
const TwoDialogsHost = defineComponent({
  components: { BaseDialog },
  setup() {
    return { bottomOpen: ref(true), topOpen: ref(true) }
  },
  template: `
    <BaseDialog v-model="bottomOpen" title="Fondo" />
    <BaseDialog v-model="topOpen" title="Encima" />
  `,
})

describe('BaseDialog', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    // Limpiar el DOM del body antes de cada test
    const existing = document.getElementById('test-dialog-root')
    if (existing) existing.remove()
    // openDialogStack es un módulo compartido entre TODAS las instancias
    // (por diseño: así Escape solo actúa en el diálogo más alto). Los
    // wrappers de esta suite no se desmontan explícitamente entre pruebas
    // (mismo patrón que el resto del archivo), así que se limpia a mano
    // para que ninguna prueba herede el diálogo abierto de la anterior.
    openDialogStack.length = 0
  })

  afterEach(() => {
    vi.useRealTimers()
    // Limpiar body después de cada test
    const overlays = document.querySelectorAll('.base-dialog__overlay')
    overlays.forEach((el) => el.remove())
    document.body.style.overflow = ''
    openDialogStack.length = 0
  })

  describe('Rendering', () => {
    it('does not render when closed', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: false } })
      void wrapper
      expect(wrapper.find('.base-dialog').exists()).toBe(false)
    })

    it('renders when open', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      void wrapper
      // Teleport renders to body, so we check the whole document
      const dialogs = document.querySelectorAll('.base-dialog')
      expect(dialogs.length).toBeGreaterThan(0)
    })

    it('renders title', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, title: 'Dialog Title' } })
      void wrapper
      const titles = document.querySelectorAll('.base-dialog__title')
      expect(titles.length).toBeGreaterThan(0)
      expect(titles[0].textContent).toBe('Dialog Title')
    })

    it('renders description', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, description: 'Description' } })
      void wrapper
      const descriptions = document.querySelectorAll('.base-dialog__description')
      expect(descriptions.length).toBeGreaterThan(0)
      expect(descriptions[0].textContent).toBe('Description')
    })

    it('renders default slot', () => {
      const wrapper = mount(BaseDialog, {
        props: { modelValue: true },
        slots: { default: 'Content text' },
      })
      void wrapper
      expect(document.body.textContent).toContain('Content text')
    })

    it('renders footer slot', () => {
      const wrapper = mount(BaseDialog, {
        props: { modelValue: true },
        slots: { footer: '<button>OK</button>' },
      })
      void wrapper
      expect(document.querySelector('.base-dialog__footer')).toBeTruthy()
      expect(document.body.textContent).toContain('OK')
    })

    it('renders header slot', () => {
      const wrapper = mount(BaseDialog, {
        props: { modelValue: true },
        slots: { header: '<h2>Custom</h2>' },
      })
      void wrapper
      expect(document.querySelector('.base-dialog__header')).toBeTruthy()
    })

    it('shows close button when showClose true', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, showClose: true } })
      void wrapper
      expect(document.querySelector('.base-dialog__close')).toBeTruthy()
    })

    it('hides close button when showClose false', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, showClose: false } })
      void wrapper
      expect(document.querySelector('.base-dialog__close')).toBeFalsy()
    })
  })

  describe('Sizes', () => {
    it('applies size classes', () => {
      const sizes = ['sm', 'md', 'lg', 'full'] as const
      sizes.forEach((size) => {
        // Limpiar antes de cada uno
        document.querySelectorAll('.base-dialog').forEach((el) => el.remove())
        const wrapper = mount(BaseDialog, { props: { modelValue: true, size } })
        void wrapper
        const dialogs = document.querySelectorAll('.base-dialog')
        expect(dialogs[0].classList).toContain(`base-dialog--${size}`)
      })
    })
  })

  describe('Opening and Closing', () => {
    it('emits update:modelValue false on close', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      await wrapper.vm.close()
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe(false)
    })

    it('emits close event on close', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      await wrapper.vm.close()
      expect(wrapper.emitted('close')).toBeTruthy()
    })

    it('emits open event on open', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: false } })
      await wrapper.vm.open()
      expect(wrapper.emitted('open')).toBeTruthy()
    })

    it('closes when close button clicked', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, showClose: true } })
      const closeButton = document.querySelector('.base-dialog__close')
      expect(closeButton).toBeTruthy()
      await closeButton!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe(false)
    })
  })

  describe('Backdrop click', () => {
    it('closes on backdrop click when closeOnBackdrop true', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, closeOnBackdrop: true } })
      const overlay = document.querySelector('.base-dialog__overlay')
      expect(overlay).toBeTruthy()
      await overlay!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe(false)
    })

    it('does not close on backdrop click when closeOnBackdrop false', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, closeOnBackdrop: false } })
      const overlay = document.querySelector('.base-dialog__overlay')
      await overlay!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      expect(wrapper.emitted('update:modelValue')).toBeFalsy()
    })

    it('does not close when clicking dialog content', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, closeOnBackdrop: true } })
      const container = document.querySelector('.base-dialog__container')
      expect(container).toBeTruthy()
      await container!.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      expect(wrapper.emitted('update:modelValue')).toBeFalsy()
    })
  })

  describe('ESC key', () => {
    it('closes on ESC when closeOnEscape true', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, closeOnEscape: true } })
      const dialog = document.querySelector('.base-dialog')
      await dialog!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toBe(false)
    })

    it('does not close on ESC when closeOnEscape false', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, closeOnEscape: false } })
      const dialog = document.querySelector('.base-dialog')
      await dialog!.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      expect(wrapper.emitted('update:modelValue')).toBeFalsy()
    })
  })

  describe('Accessibility', () => {
    it('has role dialog', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      void wrapper
      const dialog = document.querySelector('.base-dialog')
      expect(dialog?.getAttribute('role')).toBe('dialog')
    })

    it('has aria-modal true', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      void wrapper
      const dialog = document.querySelector('.base-dialog')
      expect(dialog?.getAttribute('aria-modal')).toBe('true')
    })

    it('has aria-labelledby when title provided', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, title: 'Title' } })
      void wrapper
      const dialog = document.querySelector('.base-dialog')
      const labelledby = dialog?.getAttribute('aria-labelledby')
      expect(labelledby).toBeTruthy()
      expect(labelledby).toContain('-title')
    })

    it('has aria-describedby when description provided', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, description: 'Desc' } })
      void wrapper
      const dialog = document.querySelector('.base-dialog')
      const describedby = dialog?.getAttribute('aria-describedby')
      expect(describedby).toBeTruthy()
      expect(describedby).toContain('-description')
    })

    it('close button has aria-label', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, showClose: true } })
      void wrapper
      const closeButton = document.querySelector('.base-dialog__close')
      expect(closeButton?.getAttribute('aria-label')).toBe('Cerrar diálogo')
    })

    it('close icon has aria-hidden', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true, showClose: true } })
      void wrapper
      const closeIcon = document.querySelector('.base-dialog__close-icon')
      expect(closeIcon?.getAttribute('aria-hidden')).toBe('true')
    })
  })

  describe('Body scroll lock', () => {
    it('locks body scroll when open', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      void wrapper
      await vi.runAllTimersAsync()
      expect(document.body.style.overflow).toBe('hidden')
    })

    it('restores body scroll when closed', async () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      await vi.runAllTimersAsync()
      expect(document.body.style.overflow).toBe('hidden')
      // Simulate closing by updating the prop
      await wrapper.setProps({ modelValue: false })
      await nextTick()
      await vi.runAllTimersAsync()
      expect(document.body.style.overflow).toBe('')
    })
  })

  describe('Accesibilidad automatizada (axe-core, CA-009-05)', () => {
    it('sin violaciones abierto con título, contenido y pie', async () => {
      mount(BaseDialog, {
        props: {
          modelValue: true,
          title: 'Cancelar turno',
          description: 'Esta acción no se puede deshacer.',
        },
        slots: {
          default: '<p>¿Seguro que quieres cancelar?</p>',
          footer: '<button>Cancelar turno</button>',
        },
      })
      await vi.runAllTimersAsync()
      const overlay = document.querySelector('.base-dialog__overlay') as HTMLElement
      // axe-core programa trabajo interno con temporizadores reales; con
      // los falsos activos (beforeEach de este archivo) su promesa nunca
      // se resuelve y la prueba agota el tiempo de espera.
      vi.useRealTimers()
      expect(await axe(overlay, axeOptions)).toHaveNoViolations()
    })
  })

  describe('Fallthrough de clase', () => {
    // class ya no es un prop propio (auditoría HU-009). BaseDialog
    // teletransporta su contenido, así que el fallthrough automático de
    // Vue no aplica al overlay: no existe un gancho de clase por instancia,
    // decisión documentada en la auditoría (el tamaño y la apariencia del
    // diálogo ya están gobernados por completo por el prop size).
    it('no declara class como prop propio', () => {
      const wrapper = mount(BaseDialog, { props: { modelValue: true } })
      expect(wrapper.props()).not.toHaveProperty('class')
    })
  })

  describe('Regresión de auditoría HU-009', () => {
    it('el overlay no queda aria-hidden mientras el diálogo está abierto y enfocable', async () => {
      mount(BaseDialog, {
        props: { modelValue: true, title: 'Título' },
        slots: { default: '<button>Acción</button>' },
      })
      await vi.runAllTimersAsync()
      const overlay = document.querySelector('.base-dialog__overlay')
      // Bug encontrado en la auditoría: el overlay quedaba aria-hidden="true"
      // mientras contenía el foco atrapado, ocultando de la tecnología de
      // asistencia un diálogo que el teclado sí podía alcanzar.
      expect(overlay?.getAttribute('aria-hidden')).toBeNull()
    })

    it('genera un id único por instancia aunque haya dos diálogos montados a la vez', () => {
      mount(TwoDialogsHost)
      const ids = Array.from(document.querySelectorAll('.base-dialog')).map((el) => el.id)
      expect(ids.length).toBe(2)
      expect(new Set(ids).size).toBe(2)
    })

    it('Escape solo cierra el diálogo visualmente más alto cuando hay dos abiertos', async () => {
      const host = mount(TwoDialogsHost)
      await vi.runAllTimersAsync()

      document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
      await nextTick()

      const dialogs = host.findAllComponents(BaseDialog)
      const [bottom, top] = dialogs
      expect(top.emitted('update:modelValue')?.[0]?.[0]).toBe(false)
      expect(bottom.emitted('update:modelValue')).toBeFalsy()
    })

    it('mantiene el scroll del body bloqueado si un segundo diálogo se cierra mientras el primero sigue abierto', async () => {
      const host = mount(TwoDialogsHost)
      await vi.runAllTimersAsync()
      expect(document.body.style.overflow).toBe('hidden')

      host.vm.topOpen = false
      await nextTick()
      await vi.runAllTimersAsync()

      // El primer diálogo sigue abierto: el body no debe desbloquearse.
      expect(document.body.style.overflow).toBe('hidden')
    })
  })
})
