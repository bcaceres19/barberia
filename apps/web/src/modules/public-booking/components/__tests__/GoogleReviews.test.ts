/**
 * Reseñas de Google del negocio (HU-101, DEC-129): la sección solo existe con una
 * calificación con datos (CA-101-01, CA-101-03), muestra la atribución y el enlace a la
 * ficha real (CA-101-02) y es accesible (CA-101-06).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { axe } from 'vitest-axe'

const getReviewsMock = vi.hoisted(() => vi.fn())
vi.mock('../../api/getPublicGoogleReviewsApi', () => ({ getPublicGoogleReviews: getReviewsMock }))

const { default: GoogleReviews } = await import('../GoogleReviews.vue')

const reviews = {
  rating: 4.7,
  ratingCount: 1280,
  mapsUrl: 'https://maps.google.com/?cid=1',
  reviews: [
    {
      authorName: 'Camila R.',
      authorUrl: 'https://www.google.com/maps/contrib/1',
      rating: 5,
      text: 'Excelente atención.',
      relativeTime: 'hace un mes',
    },
    { authorName: 'Andrés M.', authorUrl: null, rating: 4, text: null, relativeTime: null },
  ],
}

async function mountIt() {
  const wrapper = mount(GoogleReviews, { attachTo: document.body, props: { slug: 'barberia-x' } })
  await flushPromises()
  return wrapper
}

beforeEach(() => getReviewsMock.mockReset())

describe('GoogleReviews', () => {
  it('renders nothing when there is nothing to show (no listing, no key, failed query)', async () => {
    getReviewsMock.mockResolvedValueOnce({ kind: 'empty' })
    const wrapper = await mountIt()

    expect(wrapper.find('section').exists()).toBe(false)
    expect(wrapper.text()).toBe('')
  })

  it('shows the rating, the count and the Google attribution with a link to the real listing', async () => {
    getReviewsMock.mockResolvedValueOnce({ kind: 'success', reviews })
    const wrapper = await mountIt()

    expect(getReviewsMock).toHaveBeenCalledWith('barberia-x')
    expect(wrapper.get('h2').text()).toBe('Opiniones en Google')
    expect(wrapper.text()).toContain('4,7')
    expect(wrapper.text()).toContain('1.280 calificaciones en Google')
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe(
      '4,7 de 5 estrellas en Google',
    )
    expect(wrapper.text()).toContain('Reseñas de Google')
    const link = wrapper.get('.pb-google__attribution a')
    expect(link.attributes('href')).toBe('https://maps.google.com/?cid=1')
    expect(link.attributes('rel')).toContain('noopener')
    expect(link.attributes('target')).toBe('_blank')
  })

  it('lists the reviews on demand, with author attribution and optional text', async () => {
    getReviewsMock.mockResolvedValueOnce({ kind: 'success', reviews })
    const wrapper = await mountIt()

    expect(wrapper.find('.pb-google__list').exists()).toBe(false)
    const toggle = wrapper.get('.pb-google__toggle')
    expect(toggle.attributes('aria-expanded')).toBe('false')
    expect(toggle.text()).toBe('Ver reseñas (2)')

    await toggle.trigger('click')

    expect(toggle.attributes('aria-expanded')).toBe('true')
    const items = wrapper.findAll('.pb-google__item')
    expect(items).toHaveLength(2)
    expect(items[0]!.text()).toContain('Camila R.')
    expect(items[0]!.text()).toContain('hace un mes')
    expect(items[0]!.text()).toContain('Excelente atención.')
    expect(items[0]!.get('a').attributes('href')).toBe('https://www.google.com/maps/contrib/1')
    // Sin enlace de autor, el nombre va como texto; sin texto, no hay párrafo.
    expect(items[1]!.find('a').exists()).toBe(false)
    expect(items[1]!.text()).toContain('Andrés M.')
    expect(items[1]!.find('.pb-google__text').exists()).toBe(false)
  })

  it('omits the toggle when the listing has a rating but no reviews to read', async () => {
    getReviewsMock.mockResolvedValueOnce({ kind: 'success', reviews: { ...reviews, reviews: [] } })
    const wrapper = await mountIt()

    expect(wrapper.find('.pb-google__toggle').exists()).toBe(false)
    expect(wrapper.text()).toContain('4,7')
  })

  it('has no accessibility violations, collapsed and expanded', async () => {
    getReviewsMock.mockResolvedValueOnce({ kind: 'success', reviews })
    const wrapper = await mountIt()
    expect(await axe(wrapper.element)).toHaveNoViolations()

    await wrapper.get('.pb-google__toggle').trigger('click')
    expect(await axe(wrapper.element)).toHaveNoViolations()
  })
})
