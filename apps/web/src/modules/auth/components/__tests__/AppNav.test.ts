/**
 * Pruebas de AppNav (HU-012; dock NAVA desde la Fase 2, issue #135; carril
 * "Más" desde la Fase 2 del rediseño Tailored Grid, issue #187):
 * navegación semántica, marca la entrada activa con clase y con el
 * aria-current nativo de RouterLink, icono decorativo por entrada,
 * separación primary/secondary y menú "Más" con cierre de sesión.
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'
import type { NavItem } from '@/shared/navigation/navItem'

const postMock = vi.hoisted(() => vi.fn())
vi.mock('@/shared/api/httpClient', () => ({
  httpClient: { GET: vi.fn(), POST: postMock },
}))

const { default: AppNav } = await import('../AppNav.vue')
const { resetForFreshLogin } = await import('../../model/sessionStore')
const { openDialogStack } = await import('@/shared/ui/BaseDialog.vue')

async function mountNav(initialRoute = '/panel', extraItems?: NavItem[]) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/panel', name: 'panel', component: { template: '<div />' } },
      { path: '/panel/horarios', name: 'schedules-horarios', component: { template: '<div />' } },
      { path: '/acceso', name: 'acceso', component: { template: '<div />' } },
    ],
  })
  await router.push(initialRoute)
  await router.isReady()
  const wrapper = mount(AppNav, { props: { extraItems }, global: { plugins: [router] } })
  return { wrapper, router }
}

describe('AppNav', () => {
  beforeEach(() => {
    postMock.mockReset()
    resetForFreshLogin()
  })

  // El "Más" de AppNav usa BaseDialog, que hace Teleport a document.body:
  // los wrappers de esta suite no se desmontan entre pruebas (mismo
  // criterio que BaseDialog.test.ts), así que se limpia a mano para que
  // ninguna prueba encuentre el diálogo (cerrado) de la anterior.
  afterEach(() => {
    document.querySelectorAll('.base-dialog__overlay').forEach((el) => el.remove())
    document.body.style.overflow = ''
    openDialogStack.length = 0
  })

  it('renders a semantic nav landmark with a labelled link (Fase 2: "Agenda", not "Panel")', async () => {
    const { wrapper } = await mountNav()
    expect(wrapper.get('nav').attributes('aria-label')).toBeTruthy()
    expect(wrapper.get('a').text()).toBe('Agenda')
  })

  it('marks the current route as active, including native aria-current from RouterLink', async () => {
    const { wrapper } = await mountNav('/panel')
    const link = wrapper.get('a')
    expect(link.classes()).toContain('app-nav__link--active')
    expect(link.attributes('aria-current')).toBe('page')
  })

  it('renders a decorative icon (aria-hidden) alongside each label', async () => {
    const { wrapper } = await mountNav()
    const icon = wrapper.get('.app-nav__icon')
    expect(icon.attributes('aria-hidden')).toBe('true')
    expect(icon.find('svg').exists()).toBe(true)
  })

  it('composes only primary extra items directly in the mobile dock, after "Agenda" (issue #189: desktop shows every destination flat instead)', async () => {
    const { wrapper } = await mountNav('/panel', [
      { to: { name: 'panel' }, label: 'Servicios', primary: true },
      { to: { name: 'schedules-horarios' }, label: 'Horarios' },
    ])
    const links = wrapper.get('.app-nav__list--mobile').findAll('a')
    expect(links.map((l) => l.text())).toEqual(['Agenda', 'Servicios'])
  })

  it('shows every destination flat in the desktop list, with no "Más" grouping (issue #189, atlas panel-agenda-eventos)', async () => {
    const { wrapper } = await mountNav('/panel', [
      { to: { name: 'panel' }, label: 'Servicios', primary: true },
      { to: { name: 'schedules-horarios' }, label: 'Horarios' },
    ])
    const links = wrapper.get('.app-nav__list--desktop').findAll('a')
    expect(links.map((l) => l.text())).toEqual(['Agenda', 'Servicios', 'Horarios'])
  })

  it('does not render a "Más" trigger when every item is primary', async () => {
    const { wrapper } = await mountNav('/panel', [
      { to: { name: 'panel' }, label: 'Servicios', primary: true },
    ])
    expect(wrapper.find('[aria-haspopup="dialog"]').exists()).toBe(false)
  })

  it('has no axe violations in the closed dock', async () => {
    const { wrapper } = await mountNav()
    const results = await axe(wrapper.element)
    expect(results).toHaveNoViolations()
  })

  describe('"Más" con destinos secundarios y cierre de sesión', () => {
    function extraItems(): NavItem[] {
      return [{ to: { name: 'schedules-horarios' }, label: 'Horarios' }]
    }

    it('renders a "Más" trigger when a secondary item exists', async () => {
      const { wrapper } = await mountNav('/panel', extraItems())
      expect(wrapper.get('[aria-haspopup="dialog"]').text()).toContain('Más')
    })

    // BaseDialog usa Teleport a document.body (mismo patrón que
    // BaseDialog.test.ts): su contenido queda fuera del árbol de
    // `wrapper.element`, así que se consulta con `document` directamente.
    it('opens the dialog with the secondary items and a logout action', async () => {
      const { wrapper } = await mountNav('/panel', extraItems())
      await wrapper.get('[aria-haspopup="dialog"]').trigger('click')
      await flushPromises()

      const dialog = document.querySelector('[role="dialog"]')
      expect(dialog?.textContent).toContain('Horarios')
      expect(document.querySelector('.app-nav__more-logout')?.textContent?.trim()).toBe(
        'Cerrar sesión',
      )
    })

    it('closes the dialog after choosing a secondary destination', async () => {
      const { wrapper } = await mountNav('/panel', extraItems())
      await wrapper.get('[aria-haspopup="dialog"]').trigger('click')
      await flushPromises()

      document.querySelector<HTMLAnchorElement>('.app-nav__more-link')?.click()
      await flushPromises()

      expect(wrapper.get('[aria-haspopup="dialog"]').attributes('aria-expanded')).toBe('false')
    })

    it('logs out, revokes the server session and closes the dialog', async () => {
      postMock.mockResolvedValueOnce({ response: new Response(null, { status: 204 }) })
      const { wrapper, router } = await mountNav('/panel', extraItems())
      await wrapper.get('[aria-haspopup="dialog"]').trigger('click')
      await flushPromises()

      document.querySelector<HTMLButtonElement>('.app-nav__more-logout')?.click()
      await flushPromises()

      expect(postMock).toHaveBeenCalledWith('/private/auth/logout')
      expect(router.currentRoute.value.name).toBe('acceso')
      expect(wrapper.get('[aria-haspopup="dialog"]').attributes('aria-expanded')).toBe('false')
    })

    it('has no axe violations with the dialog open', async () => {
      const { wrapper } = await mountNav('/panel', extraItems())
      await wrapper.get('[aria-haspopup="dialog"]').trigger('click')
      await flushPromises()

      // El diálogo vive en document.body (Teleport), fuera de wrapper.element.
      const results = await axe(document.body, { rules: { region: { enabled: false } } })
      expect(results).toHaveNoViolations()
    })
  })
})
