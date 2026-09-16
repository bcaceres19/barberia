/**
 * Pruebas de PublicCustomerDetailsPage (HU-096): formulario condicional
 * para sí mismo/otra persona (CA-096-01), campos obligatorios y opcionales
 * (CA-096-02), validación con conservación de datos ante error (CA-096-03),
 * resumen local antes de confirmar (sin llamar al servidor, HU-097 no
 * existe todavía) y ausencia de violaciones de accesibilidad (CA-096-06).
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { axe } from 'vitest-axe'
import PublicCustomerDetailsPage from '../PublicCustomerDetailsPage.vue'

const BASE_PROPS = {
  slug: 'barberia-ejemplo',
  serviceId: '8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4',
  barberId: 'a1111111-1111-1111-1111-111111111111',
  startsAt: '2026-09-20T14:30:00Z',
}

function mountPage() {
  return mount(PublicCustomerDetailsPage, { props: BASE_PROPS })
}

async function fillValidForm(wrapper: ReturnType<typeof mountPage>) {
  await wrapper.find('input[type="text"]').setValue('Ana Ríos')
  await wrapper.find('input[type="tel"]').setValue('+573001234567')
  await wrapper.find('input[type="email"]').setValue('ana@example.com')
}

describe('PublicCustomerDetailsPage', () => {
  it('shows required-field errors and preserves entered data on an invalid attempt (CA-096-02, CA-096-03)', async () => {
    const wrapper = mountPage()
    await wrapper.find('input[type="text"]').setValue('Ana Ríos')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('El teléfono es obligatorio.')
    expect(wrapper.text()).toContain('El correo es obligatorio.')
    expect((wrapper.find('input[type="text"]').element as HTMLInputElement).value).toBe('Ana Ríos')
  })

  it('derives the attendee name from the customer without asking twice when booking for self (CA-096-01)', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    // "Para mí" nunca pide un segundo nombre: un solo campo de texto en el
    // formulario (el nombre del cliente).
    expect(wrapper.findAll('input[type="text"]')).toHaveLength(1)

    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('Ana Ríos')
    const rows = wrapper.findAll('.customer-details__summary-row')
    const attendedRow = rows.find((row) => row.text().includes('Atiende a'))!
    expect(attendedRow.text()).toContain('Ana Ríos')
  })

  it('requires a non-empty attendee name when booking for someone else (CA-096-01)', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.findAll('input[type="radio"]')[1]!.setValue()

    expect(wrapper.findAll('input[type="text"]')).toHaveLength(2)

    await wrapper.find('form').trigger('submit')
    expect(wrapper.text()).toContain('El nombre de la persona atendida es obligatorio.')

    const attendeeInput = wrapper.findAll('input[type="text"]')[1]!
    await attendeeInput.setValue('Mateo Ruiz')
    await wrapper.find('form').trigger('submit')

    const rows = wrapper.findAll('.customer-details__summary-row')
    const attendedRow = rows.find((row) => row.text().includes('Atiende a'))!
    expect(attendedRow.text()).toContain('Mateo Ruiz')
    const clientRow = rows.find((row) => row.text().includes('Cliente'))!
    expect(clientRow.text()).toContain('Ana Ríos')
  })

  it('lets the customer go back to editing without losing the data already entered', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    expect(wrapper.find('form').exists()).toBe(false)

    await wrapper.find('button[type="button"]').trigger('click')
    expect(wrapper.find('form').exists()).toBe(true)
    expect((wrapper.find('input[type="tel"]').element as HTMLInputElement).value).toBe(
      '+573001234567',
    )
  })

  it('rejects an invalid phone/email format', async () => {
    const wrapper = mountPage()
    await wrapper.find('input[type="text"]').setValue('Ana Ríos')
    await wrapper.find('input[type="tel"]').setValue('3001234567')
    await wrapper.find('input[type="email"]').setValue('no-arroba')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.text()).toContain('indicativo de país')
    expect(wrapper.text()).toContain('formato inválido')
  })

  it('has no accessibility violations on the form', async () => {
    const wrapper = mountPage()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on the form with attempted errors', async () => {
    const wrapper = mountPage()
    await wrapper.find('form').trigger('submit')
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  it('has no accessibility violations on the summary', async () => {
    const wrapper = mountPage()
    await fillValidForm(wrapper)
    await wrapper.find('form').trigger('submit')
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })
})
