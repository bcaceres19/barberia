import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import { buildVocabularyFromTerms } from '@/shared/model'
import AppointmentCancellation from '../AppointmentCancellation.vue'
import type { CustomerAppointment } from '../../model/customerAppointmentOutcome'

const STARTS_AT = '2027-03-15T15:00:00Z'
const startsMs = Date.parse(STARTS_AT)
const limitMs = startsMs - 20 * 60_000

function appointment(over: Partial<CustomerAppointment> = {}): CustomerAppointment {
  return {
    barbershopName: 'Barbería Ejemplo',
    timezone: 'America/Bogota',
    attendeeName: 'Cliente Ejemplo',
    serviceName: 'Corte clásico',
    durationMinutes: 30,
    barberName: 'Barbero Ejemplo',
    startsAt: STARTS_AT,
    endsAt: '2027-03-15T15:30:00Z',
    status: 'confirmed',
    cancellationDeadlineMinutes: 20,
    lateCancellationClientAllowed: true,
    lateCancellationReasonRequired: true,
    versionToken: 'v-1',
    barbershopContactEmail: 'contacto@barberia-ejemplo.test',
    barbershopContactPhone: '+573001234567',
    vocabulary: {
      businessTerm: 'barbería',
      businessTermGender: 'feminine',
      professionalTerm: 'barbero',
      professionalTermPlural: 'barberos',
      professionalTermGender: 'masculine',
    },
    canReview: false,
    review: null,
    ...over,
  }
}

function mountCancel(
  over: Partial<CustomerAppointment> = {},
  props: Record<string, unknown> = {},
  nowMs = limitMs - 60_000,
) {
  const appt = appointment(over)
  return mount(AppointmentCancellation, {
    props: {
      appointment: appt,
      vocabulary: buildVocabularyFromTerms(appt.vocabulary),
      nowMs,
      ...props,
    },
    attachTo: document.body,
  })
}

describe('AppointmentCancellation', () => {
  it('dentro de plazo ofrece el botón y, al abrirlo, explica la consecuencia con el motivo opcional', async () => {
    const wrapper = mountCancel()
    expect(wrapper.find('form').exists()).toBe(false)

    await wrapper.find('button').trigger('click')

    expect(wrapper.find('form').exists()).toBe(true)
    expect(wrapper.text()).toContain('No se puede deshacer')
    expect(wrapper.text()).toContain('(opcional)')
    expect(document.activeElement?.id).toBe('appointment-cancel-reason')
    wrapper.unmount()
  })

  it('dentro de plazo envía sin motivo (null) sin pedirlo', async () => {
    const wrapper = mountCancel()
    await wrapper.find('button').trigger('click')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')).toEqual([[{ reason: null }]])
    wrapper.unmount()
  })

  it('fuera de plazo con motivo obligatorio lo exige y no envía en blanco', async () => {
    const wrapper = mountCancel({}, {}, limitMs + 60_000)
    await wrapper.find('button').trigger('click')
    expect(wrapper.text()).toContain('(obligatorio)')

    await wrapper.find('textarea').setValue('   ')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.find('[role="alert"]').text()).toContain('motivo')
    expect(wrapper.find('textarea').attributes('aria-invalid')).toBe('true')

    await wrapper.find('textarea').setValue('  me surgió algo ')
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')).toEqual([[{ reason: 'me surgió algo' }]])
    wrapper.unmount()
  })

  it('un rechazo reason-required del servidor vuelve obligatorio el motivo y enfoca el campo', async () => {
    const wrapper = mountCancel()
    await wrapper.find('button').trigger('click')
    await wrapper.setProps({ rejection: 'reason-required' })
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('(obligatorio)')
    expect(wrapper.text()).toContain('Cuéntanos brevemente el motivo')
    expect(document.activeElement?.id).toBe('appointment-cancel-reason')
    wrapper.unmount()
  })

  it('no deja superar los 500 caracteres del motivo', async () => {
    const wrapper = mountCancel()
    await wrapper.find('button').trigger('click')
    await wrapper.find('textarea').setValue('a'.repeat(501))
    await wrapper.find('form').trigger('submit')
    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('500')
    wrapper.unmount()
  })

  it('"Mantener mi turno" cierra el formulario sin enviar', async () => {
    const wrapper = mountCancel()
    await wrapper.find('button').trigger('click')
    const keep = wrapper.findAll('button').find((b) => b.text() === 'Mantener mi turno')!
    await keep.trigger('click')
    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.emitted('submit')).toBeUndefined()
    wrapper.unmount()
  })

  it('con la política que no permite cancelar tarde no hay botón: explica y da el contacto', () => {
    const wrapper = mountCancel(
      { lateCancellationClientAllowed: false, lateCancellationReasonRequired: false },
      {},
      limitMs + 60_000,
    )
    expect(wrapper.find('button').exists()).toBe(false)
    expect(wrapper.text()).toContain('El plazo para cancelar por tu cuenta ya venció')
    expect(wrapper.text()).toContain('a la barbería')
    expect(wrapper.find('a[href="tel:+573001234567"]').exists()).toBe(true)
    expect(wrapper.find('a[href="mailto:contacto@barberia-ejemplo.test"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('con el turno ya iniciado dice que no puede cancelarlo y sin contacto remite al canal de reserva', () => {
    const wrapper = mountCancel(
      { barbershopContactEmail: null, barbershopContactPhone: null },
      {},
      startsMs + 1,
    )
    expect(wrapper.find('button').exists()).toBe(false)
    expect(wrapper.text()).toContain('ya empezó')
    expect(wrapper.text()).toContain('mismo medio por el que reservaste')
    wrapper.unmount()
  })

  it('un window-closed del servidor reemplaza el formulario aunque el reloj local dijera otra cosa', async () => {
    const wrapper = mountCancel()
    await wrapper.find('button').trigger('click')
    await wrapper.setProps({ rejection: 'window-closed' })

    expect(wrapper.find('form').exists()).toBe(false)
    expect(wrapper.text()).toContain('El plazo para cancelar por tu cuenta ya venció')
    wrapper.unmount()
  })

  it('muestra el error del envío y bloquea los botones mientras envía', async () => {
    const wrapper = mountCancel({}, { errorMessage: 'No pudimos conectar.', submitting: true })
    await wrapper.find('button').trigger('click')
    expect(wrapper.text()).toContain('No pudimos conectar.')
    const keep = wrapper.findAll('button').find((b) => b.text() === 'Mantener mi turno')!
    expect(keep.attributes('disabled')).toBeDefined()
    expect(wrapper.find('textarea').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })

  it('no tiene violaciones de accesibilidad en ninguno de sus estados', async () => {
    const closed = mountCancel({ lateCancellationClientAllowed: false }, {}, limitMs + 1)
    expect(await axe(closed.element)).toHaveNoViolations()
    closed.unmount()

    const open = mountCancel()
    expect(await axe(open.element)).toHaveNoViolations()
    await open.find('button').trigger('click')
    expect(await axe(open.element)).toHaveNoViolations()
    await open.setProps({ errorMessage: 'Algo falló.', rejection: 'reason-required' })
    expect(await axe(open.element)).toHaveNoViolations()
    open.unmount()
  })
})
