/**
 * Vocabulario de la barbería en la reserva pública (DEC-119): el cascarón lo lee del
 * perfil público del enlace y lo ofrece a las pantallas del recorrido, que dicen
 * «manicurista» y «estudio» en vez de «barbero» y «barbería». Con los valores
 * iniciales cada texto sale idéntico al anterior.
 */
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import PublicBookingLayout from '../PublicBookingLayout.vue'
import BookingProgress from '../../components/BookingProgress.vue'
import { resetPublicProfileCache } from '../../model/publicProfile'
import { usePublicVocabulary } from '../../model/publicVocabulary'
import { defineComponent, h } from 'vue'

const resolveBarbershopMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/resolveBarbershopApi', () => ({ resolveBarbershop: resolveBarbershopMock }))

const profile = (vocabulary: Record<string, string>) => ({
  kind: 'success' as const,
  profile: {
    name: 'Estudio Lila',
    timezone: 'America/Bogota',
    contactEmail: null,
    contactPhone: null,
    vocabulary,
  },
})

const nails = {
  businessTerm: 'estudio de uñas',
  businessTermGender: 'masculine',
  professionalTerm: 'manicurista',
  professionalTermPlural: 'manicuristas',
  professionalTermGender: 'feminine',
}

// Pantalla de prueba: dice lo que dirían las reales, a partir del vocabulario inyectado.
const Probe = defineComponent({
  setup() {
    const v = usePublicVocabulary()
    return () =>
      h('div', [
        h('p', { class: 'probe' }, `Elige tu ${v.value.professional} de ${v.value.theBusiness}`),
        h(BookingProgress, { step: 2 }),
      ])
  },
})

async function mountAt(slug: string, subpath = 'servicios/s-1/barbero') {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      {
        path: '/reservar/:slug',
        component: PublicBookingLayout,
        children: [
          { path: '', name: 'reserva-publica-entrada', component: Probe },
          { path: 'servicios/:serviceId/barbero', name: 'probe', component: Probe },
        ],
      },
    ],
  })
  await router.push(`/reservar/${slug}/${subpath}`)
  const wrapper = mount(
    { template: '<RouterView />' },
    { global: { plugins: [router], stubs: { BookingBackdrop: true } } },
  )
  await flushPromises()
  return wrapper
}

beforeEach(() => {
  resolveBarbershopMock.mockReset()
  resetPublicProfileCache()
})

describe('vocabulario de la reserva pública (DEC-119)', () => {
  it('dice la palabra configurada en el cascarón, el progreso y las pantallas de estado', async () => {
    resolveBarbershopMock.mockResolvedValueOnce(profile(nails))
    const wrapper = await mountAt('estudiolila')

    expect(resolveBarbershopMock).toHaveBeenCalledWith('estudiolila')
    expect(wrapper.get('.probe').text()).toBe('Elige tu manicurista de el estudio de uñas')
    expect(wrapper.get('.pb-progress').text()).toContain('Manicurista')
    expect(wrapper.text()).not.toMatch(/barber(o|a|os|as|ía|ías)\b/i)
  })

  it('conserva los valores iniciales si la barbería nunca los cambió', async () => {
    resolveBarbershopMock.mockResolvedValueOnce(
      profile({
        businessTerm: 'barbería',
        businessTermGender: 'feminine',
        professionalTerm: 'barbero',
        professionalTermPlural: 'barberos',
        professionalTermGender: 'masculine',
      }),
    )
    const wrapper = await mountAt('barberia')

    expect(wrapper.get('.probe').text()).toBe('Elige tu barbero de la barbería')
    expect(wrapper.get('.pb-progress').text()).toContain('Barbero')
  })

  it('no monta las pantallas hasta saber el vocabulario, para no mostrar un instante la palabra equivocada', async () => {
    let release!: (value: unknown) => void
    resolveBarbershopMock.mockReturnValueOnce(new Promise((resolve) => (release = resolve)))
    const wrapper = await mountAt('estudiolila')

    expect(wrapper.find('.probe').exists()).toBe(false)
    release(profile(nails))
    await flushPromises()
    expect(wrapper.find('.probe').exists()).toBe(true)
  })

  it('usa los valores iniciales si el enlace no resuelve y deja que cada pantalla muestre su error', async () => {
    resolveBarbershopMock.mockResolvedValueOnce({ kind: 'not-found' })
    const wrapper = await mountAt('desconocido')

    expect(wrapper.get('.probe').text()).toBe('Elige tu barbero de la barbería')
  })

  it('no hace esperar a la entrada: ella misma lee el perfil y comparte la petición', async () => {
    let release!: (value: unknown) => void
    resolveBarbershopMock.mockReturnValueOnce(new Promise((resolve) => (release = resolve)))
    const wrapper = await mountAt('estudiolila', '')

    expect(wrapper.find('.probe').exists()).toBe(true)
    release(profile(nails))
    await flushPromises()
    expect(wrapper.get('.probe').text()).toBe('Elige tu manicurista de el estudio de uñas')
  })
})
