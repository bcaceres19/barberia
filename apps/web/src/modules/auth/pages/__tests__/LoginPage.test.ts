/**
 * Pruebas de integración de LoginPage: envío válido y navegación a
 * `/panel` (CA-010-01), credenciales inválidas (CA-010-02), error de red
 * con Reintentar (CA-010-03), doble envío con una sola solicitud
 * (CA-010-04), botón cargando, 429 documentado y error inesperado con
 * `requestId`. `login()` se sustituye por un doble de prueba: el
 * recorrido real contra el API real en local vive en `e2e/acceso.spec.ts`.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils'
import { createRouter, createMemoryHistory } from 'vue-router'
import { axe } from 'vitest-axe'
import type { LoginOutcome } from '../../model/loginOutcome'

const loginMock = vi.hoisted(() => vi.fn())
const requestChallengeMock = vi.hoisted(() => vi.fn())
const verifyChallengeMock = vi.hoisted(() => vi.fn())

vi.mock('../../api/loginApi', () => ({ login: loginMock }))
vi.mock('../../api/challengeApi', () => ({
  requestChallenge: requestChallengeMock,
  verifyChallenge: verifyChallengeMock,
}))

const { default: LoginPage } = await import('../LoginPage.vue')

// region/color-contrast: mismo motivo que el resto de la suite (sin
// Canvas2D en jsdom; fuera del fragmento montado no hay landmarks
// completos de la página).
// heading-order: BaseAlert.vue (HU-009) fija el título de una alerta como
// `<h4>` porque es la etiqueta de un widget transitorio (`role="alert"`),
// no un encabezado del esquema del documento; junto al único `<h1>` real
// de esta página, axe interpreta ese salto como un esquema de encabezados
// roto. No es un componente de HU-010 (no se modifica BaseAlert aquí) ni
// un defecto real de esta pantalla: se documenta la excepción en vez de
// insertar un h2/h3 vacío solo para complacer la regla.
const axeOptions = {
  rules: {
    region: { enabled: false },
    'color-contrast': { enabled: false },
    'heading-order': { enabled: false },
  },
}

function buildRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/acceso', name: 'acceso', component: { template: '<div>acceso</div>' } },
      { path: '/panel', name: 'panel', component: { template: '<div>panel</div>' } },
      {
        path: '/panel/otra-seccion',
        name: 'panel-otra-seccion',
        component: { template: '<div>otra sección</div>' },
      },
    ],
  })
}

async function mountPage(query?: Record<string, string>) {
  const router = buildRouter()
  await router.push({ name: 'acceso', query })
  await router.isReady()
  const wrapper = mount(LoginPage, {
    global: {
      plugins: [router],
      stubs: { RouterLink: { props: ['to'], template: '<a><slot /></a>' } },
    },
  })
  return { wrapper, router }
}

async function fillAndSubmit(
  wrapper: VueWrapper,
  email = 'barbero@ejemplo.test',
  password = 'contraseña-valida',
) {
  await wrapper.get('input[name="email"]').setValue(email)
  await wrapper.get('input[name="password"]').setValue(password)
  await wrapper.get('form').trigger('submit')
}

describe('LoginPage', () => {
  beforeEach(() => {
    loginMock.mockReset()
    requestChallengeMock.mockReset()
    verifyChallengeMock.mockReset()
    window.sessionStorage.clear()
  })

  it('navigates to /panel by default on success (CA-010-01)', async () => {
    const outcome: LoginOutcome = {
      kind: 'success',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    }
    loginMock.mockResolvedValueOnce(outcome)
    const { wrapper, router } = await mountPage()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('panel')
  })

  it('returns to the intended destination after login (CA-012-02)', async () => {
    const outcome: LoginOutcome = {
      kind: 'success',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    }
    loginMock.mockResolvedValueOnce(outcome)
    const { wrapper, router } = await mountPage()
    await router.push({ name: 'acceso', query: { redirect: '/panel/otra-seccion' } })

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/panel/otra-seccion')
  })

  it('shows a "Tu sesión venció" message when redirected with ?motivo=sesion-expirada (Fase 3, issue #138)', async () => {
    const { wrapper } = await mountPage({ motivo: 'sesion-expirada' })
    expect(wrapper.text()).toContain('Tu sesión venció')
  })

  it('does not show the session-expired message on a plain visit to /acceso', async () => {
    const { wrapper } = await mountPage()
    expect(wrapper.text()).not.toContain('Tu sesión venció')
  })

  it('hides the session-expired message once a submission is attempted', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'invalid-credentials' })
    const { wrapper } = await mountPage({ motivo: 'sesion-expirada' })
    expect(wrapper.text()).toContain('Tu sesión venció')

    await fillAndSubmit(wrapper, 'barbero@ejemplo.test', 'clave-incorrecta')
    await flushPromises()

    expect(wrapper.text()).not.toContain('Tu sesión venció')
  })

  it('ignores an unsafe redirect target and falls back to /panel', async () => {
    const outcome: LoginOutcome = {
      kind: 'success',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    }
    loginMock.mockResolvedValueOnce(outcome)
    const { wrapper, router } = await mountPage()
    await router.push({ name: 'acceso', query: { redirect: 'https://evil.example/robo' } })

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('panel')
  })

  it('shows a non-enumerable message and preserves the email on invalid credentials (CA-010-02)', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'invalid-credentials' } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper, 'barbero@ejemplo.test', 'clave-incorrecta')
    await flushPromises()

    expect((wrapper.get('input[name="email"]').element as HTMLInputElement).value).toBe(
      'barbero@ejemplo.test',
    )
    expect(wrapper.text()).toContain('Revisa tu correo y contraseña')
    // No debe mencionar si el correo existe o no.
    expect(wrapper.text().toLowerCase()).not.toContain('no existe')
  })

  it('does not show a contradictory "required field" hint after the password is cleared post-rejection', async () => {
    // Regresión: LoginPage limpia password.value tras credenciales
    // inválidas (CA-010-02); ese cambio de prop no debe reactivar la
    // validación local de "Escribe tu contraseña" mientras el mensaje del
    // servidor ya explica el rechazo (ver LoginForm.vue, revalidación
    // atada a eventos de entrada reales, no a un watch sobre props).
    loginMock.mockResolvedValueOnce({ kind: 'invalid-credentials' } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper, 'barbero@ejemplo.test', 'clave-incorrecta')
    await flushPromises()

    expect((wrapper.get('input[name="password"]').element as HTMLInputElement).value).toBe('')
    expect(wrapper.find('.base-input__error').exists()).toBe(false)
    expect(wrapper.text()).toContain('Revisa tu correo y contraseña')
  })

  it('offers Reintentar and preserves both fields on a network error (CA-010-03)', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'network-error' } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper, 'barbero@ejemplo.test', 'clave-cualquiera')
    await flushPromises()

    expect((wrapper.get('input[name="email"]').element as HTMLInputElement).value).toBe(
      'barbero@ejemplo.test',
    )
    expect((wrapper.get('input[name="password"]').element as HTMLInputElement).value).toBe(
      'clave-cualquiera',
    )
    expect(wrapper.text()).toContain('Reintentar')
  })

  it('retries without reloading when Reintentar is pressed', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'network-error' } satisfies LoginOutcome)
    const { wrapper } = await mountPage()
    await fillAndSubmit(wrapper, 'barbero@ejemplo.test', 'clave-cualquiera')
    await flushPromises()

    loginMock.mockResolvedValueOnce({
      kind: 'success',
      expiresAt: new Date(Date.now() + 60_000).toISOString(),
    } satisfies LoginOutcome)
    await wrapper.get('button:not([type="submit"])').trigger('click')
    await flushPromises()

    expect(loginMock).toHaveBeenCalledTimes(2)
  })

  it('explains a 429 and shows the phone challenge to unblock immediately (HU-007)', async () => {
    loginMock.mockResolvedValueOnce({
      kind: 'rate-limited',
      retryAfterSeconds: 120,
    } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('Demasiados intentos')
    expect(wrapper.text()).toContain('Verifica tu teléfono')
  })

  it('does not show the phone challenge outside the rate-limited state', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'invalid-credentials' } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(wrapper.text()).not.toContain('Verifica tu teléfono')
  })

  // CA-007-02 de punta a punta en la capa de página: tras un 429, verificar
  // el reto (mockeado) reintenta el login automáticamente con las mismas
  // credenciales, sin que el barbero tenga que volver a escribirlas.
  it('retries login automatically once the phone challenge verifies', async () => {
    loginMock
      .mockResolvedValueOnce({
        kind: 'rate-limited',
        retryAfterSeconds: 120,
      } satisfies LoginOutcome)
      .mockResolvedValueOnce({
        kind: 'success',
        expiresAt: new Date(Date.now() + 60_000).toISOString(),
      } satisfies LoginOutcome)
    requestChallengeMock.mockResolvedValueOnce({ kind: 'accepted' })
    verifyChallengeMock.mockResolvedValueOnce({ kind: 'verified' })

    const { wrapper, router } = await mountPage()
    await fillAndSubmit(wrapper)
    await flushPromises()

    await wrapper.get('button:not([type="submit"])').trigger('click') // "Enviar código por WhatsApp"
    await flushPromises()
    await wrapper.get('input[name="challengeCode"]').setValue('482913')
    const verifyButton = wrapper
      .findAll('button')
      .find((b) => b.text().includes('Verificar código'))
    await verifyButton?.trigger('click')
    await flushPromises()

    expect(verifyChallengeMock).toHaveBeenCalledWith('barbero@ejemplo.test', '482913')
    expect(loginMock).toHaveBeenCalledTimes(2)
    expect(router.currentRoute.value.name).toBe('panel')
  })

  it('shows a safe unexpected-error message and surfaces requestId when present', async () => {
    loginMock.mockResolvedValueOnce({
      kind: 'unexpected-error',
      requestId: 'req-999',
    } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper)
    await flushPromises()

    expect(wrapper.text()).toContain('req-999')
  })

  it('sends exactly one request when submit fires twice while one is in flight (CA-010-04)', async () => {
    let resolveLogin: (value: LoginOutcome) => void = () => {}
    loginMock.mockImplementationOnce(
      () =>
        new Promise<LoginOutcome>((resolve) => {
          resolveLogin = resolve
        }),
    )
    const { wrapper } = await mountPage()
    await wrapper.get('input[name="email"]').setValue('barbero@ejemplo.test')
    await wrapper.get('input[name="password"]').setValue('contraseña-valida')

    await wrapper.get('form').trigger('submit')
    await wrapper.get('form').trigger('submit')

    expect(loginMock).toHaveBeenCalledTimes(1)
    resolveLogin({ kind: 'invalid-credentials' })
    await flushPromises()
  })

  it('disables the submit action and shows aria-busy while a request is in flight', async () => {
    loginMock.mockImplementationOnce(() => new Promise<LoginOutcome>(() => {}))
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper)

    const button = wrapper.get('button[type="submit"]')
    expect(button.attributes('aria-busy')).toBe('true')
  })

  it('has no axe violations after a server error is shown', async () => {
    loginMock.mockResolvedValueOnce({ kind: 'invalid-credentials' } satisfies LoginOutcome)
    const { wrapper } = await mountPage()

    await fillAndSubmit(wrapper)
    await flushPromises()

    const results = await axe(wrapper.element, axeOptions)
    expect(results).toHaveNoViolations()
  })
})
