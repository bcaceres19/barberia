import { expect, test, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

// Evidencia determinista de #191. Los datos viven solo dentro de Playwright:
// no inician sesión real ni contienen información de producción.
const evidenceDir = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'agenda-detalle',
  'fidelidad-191',
)

const detail = {
  id: 'turno-mateo',
  barberId: 'barbero-julian',
  barberFullName: 'Julián Rodríguez',
  attendeeName: 'Mateo Rojas',
  customerFullName: 'Mateo Rojas',
  customerPhone: '+57 321 456 7890',
  customerEmail: 'mateo.rojas@example.test',
  customerNote: 'Prefiere degradado bajo en los laterales.',
  startsAt: '2026-05-24T14:00:00Z',
  endsAt: '2026-05-24T14:45:00Z',
  status: 'confirmed',
  origin: 'manual',
  serviceName: 'Corte clásico',
  durationMinutes: 45,
  priceAmount: '28000.00',
  currency: 'COP',
  versionToken: 'opaque-version-token',
  createdAt: '2026-05-24T15:30:00Z',
}

const history = {
  items: [
    {
      id: 'event-created',
      eventType: 'appointment_created',
      actorType: 'system',
      actorLabel: 'Sistema',
      reason: null,
      occurredAt: '2026-05-24T15:30:00Z',
      changes: [],
    },
    {
      id: 'event-rescheduled',
      eventType: 'appointment_rescheduled',
      actorType: 'staff',
      actorLabel: 'Julián Rodríguez',
      reason: null,
      occurredAt: '2026-05-24T15:45:00Z',
      changes: [
        {
          fieldName: 'starts_at',
          previousValue: '24 may 2026, 08:00 AM',
          newValue: '24 may 2026, 09:00 AM',
        },
      ],
    },
  ],
  nextCursor: 'more-events',
}

function json(body: unknown, status = 200) {
  return { status, contentType: 'application/json', body: JSON.stringify(body) }
}

// cancelOutcome controla la respuesta de POST .../cancel (HU-066, T6):
// 'success' (por defecto) cancela de verdad -la lectura GET posterior ya
// refleja cancelled_by_barber, mismo criterio que el turno real tras
// confirmar-, 'version-conflict' reproduce el 409 distinguible sin mutar
// nada, para la evidencia del diálogo de conflicto. closeOutcome hace lo
// mismo para POST .../complete y .../no-show (HU-067, T4 manual/T7):
// 'success' (por defecto) aplica el estado terminal correspondiente,
// 'version-conflict'/'invalid-state'/'not-started' reproducen cada
// conflicto distinguible sin mutar nada. correctOutcome hace lo mismo para
// POST .../correct-status (HU-068, T8): 'success' aplica el destino que el
// cuerpo de la solicitud pidió, 'version-conflict'/'invalid-state'/
// 'conflict'/'not-started' reproducen cada conflicto distinguible sin
// mutar nada. initialStatus permite abrir el detalle YA en un terminal
// (T8 solo se ofrece sobre un terminal, CA-068-08), sin tener que cerrar
// primero el turno dentro de la misma prueba.
async function installDetailMock(
  page: Page,
  options: {
    cancelOutcome?: 'success' | 'version-conflict'
    cancelDelayMs?: number
    closeOutcome?: 'success' | 'version-conflict' | 'invalid-state' | 'not-started'
    closeDelayMs?: number
    correctOutcome?: 'success' | 'version-conflict' | 'invalid-state' | 'conflict' | 'not-started'
    correctDelayMs?: number
    initialStatus?:
      'confirmed' | 'completed' | 'no_show' | 'cancelled_by_customer' | 'cancelled_by_barber'
  } = {},
) {
  const cancelOutcome = options.cancelOutcome ?? 'success'
  const cancelDelayMs = options.cancelDelayMs ?? 0
  const closeOutcome = options.closeOutcome ?? 'success'
  const closeDelayMs = options.closeDelayMs ?? 0
  const correctOutcome = options.correctOutcome ?? 'success'
  const correctDelayMs = options.correctDelayMs ?? 0
  const state = { cancelRequests: 0, completeRequests: 0, noShowRequests: 0, correctRequests: 0 }
  let finalStatus:
    'confirmed' | 'cancelled_by_barber' | 'cancelled_by_customer' | 'completed' | 'no_show' =
    options.initialStatus ?? 'confirmed'

  function versionConflictProblem() {
    return json(
      {
        type: 'https://nava.example/problems/version-conflict',
        title: 'Conflicto de versión',
        status: 409,
        detail: 'La representación del turno cambió.',
        code: 'version-conflict',
      },
      409,
    )
  }

  function invalidStateProblem() {
    return json(
      {
        type: 'https://nava.example/problems/invalid-state',
        title: 'Estado inválido',
        status: 409,
        detail:
          'el turno ya tiene un resultado terminal registrado; una futura corrección (T8) permitirá cambiarlo',
        code: 'invalid-state',
      },
      409,
    )
  }

  function notStartedProblem() {
    return json(
      {
        type: 'https://nava.example/problems/validation-error',
        title: 'Error de validación',
        status: 422,
        detail: 'el turno todavía no comienza; espera hasta su hora de inicio',
        code: 'validation-error',
      },
      422,
    )
  }

  // notTerminalProblem cubre HU-068 (CA-068-01): el turno todavía está
  // `confirmed`, sin ningún resultado terminal que T8 pueda corregir.
  function notTerminalProblem() {
    return json(
      {
        type: 'https://nava.example/problems/invalid-state',
        title: 'Estado inválido',
        status: 409,
        detail: 'el turno todavía no tiene un resultado terminal para corregir',
        code: 'invalid-state',
      },
      409,
    )
  }

  // scheduleConflictProblem cubre HU-068 (CA-068-04): el destino elegido
  // volvería a ocupar una franja que otra cita del mismo barbero ya ocupa.
  function scheduleConflictProblem() {
    return json(
      {
        type: 'https://nava.example/problems/conflict',
        title: 'Conflicto',
        status: 409,
        detail: 'el barbero ya tiene una cita en ese intervalo',
        code: 'conflict',
      },
      409,
    )
  }

  function currentDetail() {
    return { ...detail, status: finalStatus, versionToken: 'opaque-version-token-2' }
  }

  await page.route('**/api/v1/private/**', async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname.endsWith('/auth/session')) {
      await route.fulfill(
        json({
          barbershop: { id: 'shop-nava', name: 'NAVA QA Local' },
          expiresAt: '2099-01-01T00:00:00Z',
        }),
      )
      return
    }
    if (url.pathname.endsWith('/settings/barbershop')) {
      await route.fulfill(json({ timezone: 'America/Bogota' }))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/history')) {
      await route.fulfill(json(history))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/reschedule')) {
      await route.fulfill(versionConflictProblem())
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/cancel')) {
      state.cancelRequests += 1
      if (cancelDelayMs > 0) {
        await new Promise((resolve) => setTimeout(resolve, cancelDelayMs))
      }
      if (cancelOutcome === 'version-conflict') {
        await route.fulfill(versionConflictProblem())
        return
      }
      finalStatus = 'cancelled_by_barber'
      await route.fulfill(json(currentDetail()))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/complete')) {
      state.completeRequests += 1
      if (closeDelayMs > 0) {
        await new Promise((resolve) => setTimeout(resolve, closeDelayMs))
      }
      if (closeOutcome === 'version-conflict') {
        await route.fulfill(versionConflictProblem())
        return
      }
      if (closeOutcome === 'invalid-state') {
        await route.fulfill(invalidStateProblem())
        return
      }
      if (closeOutcome === 'not-started') {
        await route.fulfill(notStartedProblem())
        return
      }
      finalStatus = 'completed'
      await route.fulfill(json(currentDetail()))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/no-show')) {
      state.noShowRequests += 1
      if (closeDelayMs > 0) {
        await new Promise((resolve) => setTimeout(resolve, closeDelayMs))
      }
      if (closeOutcome === 'version-conflict') {
        await route.fulfill(versionConflictProblem())
        return
      }
      if (closeOutcome === 'invalid-state') {
        await route.fulfill(invalidStateProblem())
        return
      }
      if (closeOutcome === 'not-started') {
        await route.fulfill(notStartedProblem())
        return
      }
      finalStatus = 'no_show'
      await route.fulfill(json(currentDetail()))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo/correct-status')) {
      state.correctRequests += 1
      if (correctDelayMs > 0) {
        await new Promise((resolve) => setTimeout(resolve, correctDelayMs))
      }
      if (correctOutcome === 'version-conflict') {
        await route.fulfill(versionConflictProblem())
        return
      }
      if (correctOutcome === 'invalid-state') {
        await route.fulfill(notTerminalProblem())
        return
      }
      if (correctOutcome === 'conflict') {
        await route.fulfill(scheduleConflictProblem())
        return
      }
      if (correctOutcome === 'not-started') {
        await route.fulfill(notStartedProblem())
        return
      }
      const body = route.request().postDataJSON() as { status: typeof finalStatus }
      finalStatus = body.status
      await route.fulfill(json(currentDetail()))
      return
    }
    if (url.pathname.endsWith('/appointments/turno-mateo')) {
      await route.fulfill(json(finalStatus === 'confirmed' ? detail : currentDetail()))
      return
    }
    await route.fulfill(json({ title: 'Mock endpoint not found' }, 404))
  })
  return state
}

async function openDetail(
  page: Page,
  width: number,
  height: number,
  mockOptions: Parameters<typeof installDetailMock>[1] = {},
) {
  await page.setViewportSize({ width, height })
  const mockState = await installDetailMock(page, mockOptions)
  await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
  await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
  await expect(page.getByText('Turno reprogramado')).toBeVisible()
  expect(await page.evaluate(() => window.innerWidth)).toBe(width)
  expect(await page.evaluate(() => window.innerHeight)).toBe(height)
  return mockState
}

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`fidelidad #191: detalle listo en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth > window.innerWidth,
    )
    expect(overflow).toBe(false)
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'detalle-listo.png'),
      animations: 'disabled',
    })
  })

  test(`fidelidad #191: diálogo de conflicto en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Reprogramar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Reprogramar turno' })
    await dialog.getByRole('button', { name: 'Confirmar' }).click()
    await expect(dialog.getByText('Este turno cambió mientras lo editabas')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDir, viewport.name, 'reprogramar-conflicto.png'),
      animations: 'disabled',
    })
  })
}

// --- HU-066 (T6): cancelación por el barbero, issue #225 -----------------
// El estado terminal (badge "Cancelado por el barbero") reutiliza el mismo
// componente/variant que el atlas ya fija para cancelled_by_barber
// (dailyAgenda.ts APPOINTMENT_STATUS_BADGE_VARIANT = 'danger', mismo
// tratamiento que detalle-turno-eventos/10-turno-cancelado.png); el diálogo
// de confirmación no tiene mockup exacto y se diseñó dentro de NAVA.
const evidenceDirCancel = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'agenda-detalle',
  'cancelacion-225',
)

const axeScriptPath = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '..',
  'node_modules',
  'axe-core',
  'axe.min.js',
)

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`HU-066: diálogo de confirmación de cancelación en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
    await expect(dialog.getByText('no se puede deshacer')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirCancel, viewport.name, 'confirmar-cancelacion.png'),
      animations: 'disabled',
    })
  })

  test(`HU-066: turno cancelado (estado terminal) en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
    await dialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)
    await expect(page.getByRole('status').getByText('Cancelado por el barbero')).toBeVisible()
    // El turno ya cancelado no vuelve a ofrecer "Cancelar turno" (solo
    // aplica sobre confirmed, CA-066-08).
    await expect(page.getByRole('button', { name: 'Cancelar turno' })).toHaveCount(0)
    await page.screenshot({
      path: path.join(evidenceDirCancel, viewport.name, 'turno-cancelado.png'),
      animations: 'disabled',
    })
  })

  test(`HU-066: conflicto de versión al cancelar en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, { cancelOutcome: 'version-conflict' })
    await page.getByRole('button', { name: 'Cancelar turno' }).click()
    const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
    await dialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
    await expect(dialog.getByText('cambió mientras lo revisabas')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirCancel, viewport.name, 'conflicto-version.png'),
      animations: 'disabled',
    })
  })
}

test('HU-066: doble toque no duplica la solicitud de cancelación', async ({ page }) => {
  const mockState = await openDetail(page, 1280, 900, { cancelDelayMs: 200 })

  await page.getByRole('button', { name: 'Cancelar turno' }).click()
  const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
  const confirm = dialog.getByRole('button', { name: 'Sí, cancelar turno' })
  await confirm.click()
  // El segundo toque llega mientras la solicitud sigue en curso: el botón
  // ya está deshabilitado (BaseButton :disabled), así que un clic nativo
  // real no dispara un segundo evento — la protección real (cancelStatus
  // === 'saving' en AppointmentDetailPage.vue) es la que importa aquí.
  await confirm.click({ force: true }).catch(() => {})
  await expect(page.getByRole('status').getByText('Cancelado por el barbero')).toBeVisible()

  expect(mockState.cancelRequests).toBe(1)
})

test('HU-066: recorrido completo por teclado, sin ratón', async ({ page }) => {
  await openDetail(page, 1280, 900)
  await page.getByRole('button', { name: 'Cancelar turno' }).focus()
  await page.keyboard.press('Enter')
  const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
  await expect(dialog).toBeVisible()

  await page.keyboard.press('Tab')
  await page.keyboard.press('Tab')
  await expect(dialog.getByRole('button', { name: 'Sí, cancelar turno' })).toBeFocused()
  await page.keyboard.press('Enter')

  await expect(page.getByRole('status').getByText('Cancelado por el barbero')).toBeVisible()
})

async function runAxe(page: Page) {
  await page.addScriptTag({ path: axeScriptPath })
  return page.evaluate(async () => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const axeGlobal = (window as any).axe
    return axeGlobal.run(document, { rules: { 'color-contrast': { enabled: false } } })
  }) as Promise<{ violations: unknown[] }>
}

test('HU-066: axe-core sin violaciones en el diálogo y en el estado terminal', async ({ page }) => {
  await openDetail(page, 1280, 900)

  await page.getByRole('button', { name: 'Cancelar turno' }).click()
  const dialogResults = await runAxe(page)
  expect(dialogResults.violations, JSON.stringify(dialogResults.violations, null, 2)).toEqual([])

  const dialog = page.getByRole('dialog', { name: 'Cancelar turno' })
  await dialog.getByRole('button', { name: 'Sí, cancelar turno' }).click()
  await expect(page.getByRole('status').getByText('Cancelado por el barbero')).toBeVisible()
  const terminalResults = await runAxe(page)
  expect(terminalResults.violations, JSON.stringify(terminalResults.violations, null, 2)).toEqual(
    [],
  )
})

test('HU-066: reflow y foco del botón Cancelar en los viewports obligatorios', async ({ page }) => {
  await installDetailMock(page)
  const viewports = [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    // El harness del repositorio aproxima el zoom de texto al 200 % con este viewport.
    { name: '1280-zoom200', width: 640, height: 450 },
  ]

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
    await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
    expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
    expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)
    expect(
      await page.evaluate(() => {
        const content = document.querySelector('.private-shell__content')
        return (
          document.documentElement.scrollWidth > document.documentElement.clientWidth ||
          (!!content && content.scrollWidth > content.clientWidth)
        )
      }),
      `${viewport.name}: no hay desborde horizontal`,
    ).toBe(false)

    const action = page.getByRole('button', { name: 'Cancelar turno' })
    await action.focus()
    await expect(action).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDirCancel, 'responsive', viewport.name, 'foco-cancelar.png'),
      animations: 'disabled',
    })
  }
})

test('HU-066: movimiento reducido elimina la transición del diálogo de cancelación', async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openDetail(page, 1280, 900)
  await page.getByRole('button', { name: 'Cancelar turno' }).click()
  await expect(page.getByRole('dialog', { name: 'Cancelar turno' })).toBeVisible()
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-dialog') as HTMLElement)
            .transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})

// --- HU-067 (T4 manual/T7): cierre manual como atendido o no asistió,
// issue #226 ----------------------------------------------------------
// Los dos estados terminales (badge "Completado"/"No se presentó") reutilizan
// el mismo componente/variant que el atlas ya fija (dailyAgenda.ts
// APPOINTMENT_STATUS_BADGE_VARIANT = 'success'/'warning', mismo tratamiento
// que detalle-turno-eventos/09-turno-completado.png y su variante no_show
// documentada); los diálogos de confirmación no tienen mockup exacto y se
// diseñaron dentro de NAVA, mismo criterio que el de cancelar (HU-066).
const evidenceDirClose = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'agenda-detalle',
  'cierre-manual-226',
)

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`HU-067: diálogo de confirmación al marcar como atendido en ${viewport.name}`, async ({
    page,
  }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Marcar como atendido' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar como atendido' })
    await expect(dialog.getByText('no se puede deshacer')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'confirmar-completar.png'),
      animations: 'disabled',
    })
  })

  test(`HU-067: turno completado (estado terminal) en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Marcar como atendido' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar como atendido' })
    await dialog.getByRole('button', { name: 'Sí, marcar como atendido' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)
    await expect(page.getByRole('status').getByText('Completado')).toBeVisible()
    // Un turno ya completed no vuelve a ofrecer ninguna acción de cierre
    // (T4/T7 solo aplican sobre confirmed, CA-067-08).
    await expect(page.getByRole('button', { name: 'Marcar como atendido' })).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'Marcar que no asistió' })).toHaveCount(0)
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'turno-completado.png'),
      animations: 'disabled',
    })
  })

  test(`HU-067: diálogo de confirmación al marcar que no asistió en ${viewport.name}`, async ({
    page,
  }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Marcar que no asistió' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar que no asistió' })
    await expect(dialog.getByText('no se puede deshacer')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'confirmar-no-show.png'),
      animations: 'disabled',
    })
  })

  test(`HU-067: turno no asistió (estado terminal) en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height)
    await page.getByRole('button', { name: 'Marcar que no asistió' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar que no asistió' })
    await dialog.getByRole('button', { name: 'Sí, marcar que no asistió' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)
    await expect(page.getByRole('status').getByText('No se presentó')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'turno-no-show.png'),
      animations: 'disabled',
    })
  })

  test(`HU-067: conflicto de versión al completar en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, { closeOutcome: 'version-conflict' })
    await page.getByRole('button', { name: 'Marcar como atendido' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar como atendido' })
    await dialog.getByRole('button', { name: 'Sí, marcar como atendido' }).click()
    await expect(dialog.getByText('cambió mientras lo revisabas')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'conflicto-version.png'),
      animations: 'disabled',
    })
  })

  // CA-067-05: el resultado contrario o cualquier otro terminal responde el
  // mismo invalid-state, orientando a la futura corrección T8.
  test(`HU-067: resultado contrario ya registrado en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, { closeOutcome: 'invalid-state' })
    await page.getByRole('button', { name: 'Marcar que no asistió' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar que no asistió' })
    await dialog.getByRole('button', { name: 'Sí, marcar que no asistió' }).click()
    await expect(dialog.getByText('ya tiene un resultado registrado')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'resultado-contrario.png'),
      animations: 'disabled',
    })
  })

  // CA-067-03: antes de starts_at el servidor responde 422 (distinto de un
  // conflicto): sin persistir nada, con un mensaje que explica cuándo
  // reintentar.
  test(`HU-067: turno todavía no comienza en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, { closeOutcome: 'not-started' })
    await page.getByRole('button', { name: 'Marcar como atendido' }).click()
    const dialog = page.getByRole('dialog', { name: 'Marcar como atendido' })
    await dialog.getByRole('button', { name: 'Sí, marcar como atendido' }).click()
    await expect(dialog.getByText('todavía no comienza')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirClose, viewport.name, 'aun-no-comienza.png'),
      animations: 'disabled',
    })
  })
}

test('HU-067: doble toque no duplica la solicitud de completar', async ({ page }) => {
  const mockState = await openDetail(page, 1280, 900, { closeDelayMs: 200 })

  await page.getByRole('button', { name: 'Marcar como atendido' }).click()
  const dialog = page.getByRole('dialog', { name: 'Marcar como atendido' })
  const confirm = dialog.getByRole('button', { name: 'Sí, marcar como atendido' })
  await confirm.click()
  await confirm.click({ force: true }).catch(() => {})
  await expect(page.getByRole('status').getByText('Completado')).toBeVisible()

  expect(mockState.completeRequests).toBe(1)
})

test('HU-067: recorrido completo por teclado, sin ratón, para ambas acciones', async ({ page }) => {
  await openDetail(page, 1280, 900)
  await page.getByRole('button', { name: 'Marcar que no asistió' }).focus()
  await page.keyboard.press('Enter')
  const dialog = page.getByRole('dialog', { name: 'Marcar que no asistió' })
  await expect(dialog).toBeVisible()

  await page.keyboard.press('Tab')
  await page.keyboard.press('Tab')
  await expect(dialog.getByRole('button', { name: 'Sí, marcar que no asistió' })).toBeFocused()
  await page.keyboard.press('Enter')

  await expect(page.getByRole('status').getByText('No se presentó')).toBeVisible()
})

test('HU-067: axe-core sin violaciones en ambos diálogos y en ambos estados terminales', async ({
  page,
}) => {
  await openDetail(page, 1280, 900)

  await page.getByRole('button', { name: 'Marcar como atendido' }).click()
  const completeDialogResults = await runAxe(page)
  expect(
    completeDialogResults.violations,
    JSON.stringify(completeDialogResults.violations, null, 2),
  ).toEqual([])

  const completeDialog = page.getByRole('dialog', { name: 'Marcar como atendido' })
  await completeDialog.getByRole('button', { name: 'Sí, marcar como atendido' }).click()
  await expect(page.getByRole('status').getByText('Completado')).toBeVisible()
  const completedResults = await runAxe(page)
  expect(completedResults.violations, JSON.stringify(completedResults.violations, null, 2)).toEqual(
    [],
  )
})

test('HU-067: axe-core sin violaciones marcando inasistencia', async ({ page }) => {
  await openDetail(page, 1280, 900)

  await page.getByRole('button', { name: 'Marcar que no asistió' }).click()
  const noShowDialogResults = await runAxe(page)
  expect(
    noShowDialogResults.violations,
    JSON.stringify(noShowDialogResults.violations, null, 2),
  ).toEqual([])

  const noShowDialog = page.getByRole('dialog', { name: 'Marcar que no asistió' })
  await noShowDialog.getByRole('button', { name: 'Sí, marcar que no asistió' }).click()
  await expect(page.getByRole('status').getByText('No se presentó')).toBeVisible()
  const noShowResults = await runAxe(page)
  expect(noShowResults.violations, JSON.stringify(noShowResults.violations, null, 2)).toEqual([])
})

test('HU-067: reflow y foco de las acciones de cierre en los viewports obligatorios', async ({
  page,
}) => {
  await installDetailMock(page)
  const viewports = [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    // El harness del repositorio aproxima el zoom de texto al 200 % con este viewport.
    { name: '1280-zoom200', width: 640, height: 450 },
  ]

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
    await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
    expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
    expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)
    expect(
      await page.evaluate(() => {
        const content = document.querySelector('.private-shell__content')
        return (
          document.documentElement.scrollWidth > document.documentElement.clientWidth ||
          (!!content && content.scrollWidth > content.clientWidth)
        )
      }),
      `${viewport.name}: no hay desborde horizontal`,
    ).toBe(false)

    const action = page.getByRole('button', { name: 'Marcar como atendido' })
    await action.focus()
    await expect(action).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDirClose, 'responsive', viewport.name, 'foco-completar.png'),
      animations: 'disabled',
    })
  }
})

test('HU-067: movimiento reducido elimina la transición de ambos diálogos de cierre', async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openDetail(page, 1280, 900)
  await page.getByRole('button', { name: 'Marcar como atendido' }).click()
  await expect(page.getByRole('dialog', { name: 'Marcar como atendido' })).toBeVisible()
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-dialog') as HTMLElement)
            .transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})

// Corrección auditada de un resultado terminal (HU-068, T8): sin mockup
// exacto asignado, el diálogo se diseñó libremente dentro de NAVA/Tailored
// Grid (mismo criterio que los de cancelar/cerrar). El estado terminal
// resultante reutiliza el mismo badge que ya fija el atlas (dailyAgenda.ts
// APPOINTMENT_STATUS_BADGE_VARIANT), sin cambios.
const evidenceDirCorrect = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  'evidence',
  'agenda-detalle',
  'correccion-227',
)

for (const viewport of [
  { name: 'desktop', width: 1280, height: 900 },
  { name: 'mobile', width: 360, height: 800 },
] as const) {
  test(`HU-068: diálogo de corrección en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, { initialStatus: 'completed' })
    await page.getByRole('button', { name: 'Corregir resultado' }).click()
    const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
    await dialog.getByLabel('Nuevo resultado').selectOption({ label: 'No se presentó' })
    await dialog
      .getByLabel('Motivo de la corrección')
      .fill('El barbero marcó completed por error; el cliente nunca llegó.')
    await expect(dialog.getByText('no borra ni edita el anterior')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirCorrect, viewport.name, 'confirmar-correccion.png'),
      animations: 'disabled',
    })
  })

  test(`HU-068: turno corregido (estado terminal) en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, { initialStatus: 'completed' })
    await page.getByRole('button', { name: 'Corregir resultado' }).click()
    const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
    await dialog.getByLabel('Nuevo resultado').selectOption({ label: 'No se presentó' })
    await dialog
      .getByLabel('Motivo de la corrección')
      .fill('El barbero marcó completed por error; el cliente nunca llegó.')
    await dialog.getByRole('button', { name: 'Confirmar corrección' }).click()
    await expect(page.getByRole('dialog')).toHaveCount(0)
    await expect(page.getByRole('status').getByText('No se presentó')).toBeVisible()
    // La corrección sigue disponible sobre el nuevo terminal (CA-068-08):
    // a diferencia de T4/T7, T8 nunca deja de ofrecerse sobre un terminal.
    await expect(page.getByRole('button', { name: 'Corregir resultado' })).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirCorrect, viewport.name, 'turno-corregido.png'),
      animations: 'disabled',
    })
  })

  test(`HU-068: conflicto de versión al corregir en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, {
      initialStatus: 'completed',
      correctOutcome: 'version-conflict',
    })
    await page.getByRole('button', { name: 'Corregir resultado' }).click()
    const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
    await dialog.getByLabel('Nuevo resultado').selectOption({ label: 'No se presentó' })
    await dialog.getByLabel('Motivo de la corrección').fill('motivo de prueba')
    await dialog.getByRole('button', { name: 'Confirmar corrección' }).click()
    await expect(dialog.getByText('cambió mientras lo revisabas')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirCorrect, viewport.name, 'conflicto-version.png'),
      animations: 'disabled',
    })
  })

  // CA-068-04: el destino elegido volvería a ocupar una franja que otra
  // cita del mismo barbero ya ocupa.
  test(`HU-068: cruce de agenda al corregir en ${viewport.name}`, async ({ page }) => {
    await openDetail(page, viewport.width, viewport.height, {
      initialStatus: 'cancelled_by_barber',
      correctOutcome: 'conflict',
    })
    await page.getByRole('button', { name: 'Corregir resultado' }).click()
    const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
    await dialog.getByLabel('Nuevo resultado').selectOption({ label: 'Completado' })
    await dialog.getByLabel('Motivo de la corrección').fill('motivo de prueba')
    await dialog.getByRole('button', { name: 'Confirmar corrección' }).click()
    await expect(dialog.getByText('el barbero ya tiene una cita en ese intervalo')).toBeVisible()
    await page.screenshot({
      path: path.join(evidenceDirCorrect, viewport.name, 'cruce-agenda.png'),
      animations: 'disabled',
    })
  })
}

test('HU-068: doble toque no duplica la solicitud de corrección', async ({ page }) => {
  const mockState = await openDetail(page, 1280, 900, {
    initialStatus: 'completed',
    correctDelayMs: 200,
  })

  await page.getByRole('button', { name: 'Corregir resultado' }).click()
  const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
  await dialog.getByLabel('Nuevo resultado').selectOption({ label: 'No se presentó' })
  await dialog.getByLabel('Motivo de la corrección').fill('motivo de prueba')
  const confirm = dialog.getByRole('button', { name: 'Confirmar corrección' })
  await confirm.click()
  await confirm.click({ force: true }).catch(() => {})
  await expect(page.getByRole('status').getByText('No se presentó')).toBeVisible()

  expect(mockState.correctRequests).toBe(1)
})

test('HU-068: recorrido completo por teclado, sin ratón', async ({ page }) => {
  await openDetail(page, 1280, 900, { initialStatus: 'completed' })
  await page.getByRole('button', { name: 'Corregir resultado' }).focus()
  await page.keyboard.press('Enter')
  const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
  await expect(dialog).toBeVisible()

  await page.keyboard.press('Tab')
  await expect(dialog.getByLabel('Nuevo resultado')).toBeFocused()
  // Selección nativa por teclado: el primer carácter salta a la opción que
  // empieza por él (comportamiento nativo de <select>, sin ratón).
  await page.keyboard.press('N')
  await expect(dialog.getByLabel('Nuevo resultado')).toHaveValue('no_show')
  await page.keyboard.press('Tab')
  await expect(dialog.getByLabel('Motivo de la corrección')).toBeFocused()
  await page.keyboard.type('El barbero marcó completed por error.')
  await page.keyboard.press('Tab')
  await page.keyboard.press('Tab')
  await expect(dialog.getByRole('button', { name: 'Confirmar corrección' })).toBeFocused()
  await page.keyboard.press('Enter')

  await expect(page.getByRole('dialog')).toHaveCount(0)
})

test('HU-068: axe-core sin violaciones en el diálogo y en el estado corregido', async ({
  page,
}) => {
  await openDetail(page, 1280, 900, { initialStatus: 'completed' })

  await page.getByRole('button', { name: 'Corregir resultado' }).click()
  const dialogResults = await runAxe(page)
  expect(dialogResults.violations, JSON.stringify(dialogResults.violations, null, 2)).toEqual([])

  const dialog = page.getByRole('dialog', { name: 'Corregir resultado' })
  await dialog.getByLabel('Nuevo resultado').selectOption({ label: 'No se presentó' })
  await dialog.getByLabel('Motivo de la corrección').fill('motivo de prueba')
  await dialog.getByRole('button', { name: 'Confirmar corrección' }).click()
  await expect(page.getByRole('status').getByText('No se presentó')).toBeVisible()
  const correctedResults = await runAxe(page)
  expect(correctedResults.violations, JSON.stringify(correctedResults.violations, null, 2)).toEqual(
    [],
  )
})

test('HU-068: reflow y foco de la acción de corrección en los viewports obligatorios', async ({
  page,
}) => {
  await installDetailMock(page, { initialStatus: 'completed' })
  const viewports = [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    // El harness del repositorio aproxima el zoom de texto al 200 % con este viewport.
    { name: '1280-zoom200', width: 640, height: 450 },
  ]

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
    await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
    expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
    expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)
    expect(
      await page.evaluate(() => {
        const content = document.querySelector('.private-shell__content')
        return (
          document.documentElement.scrollWidth > document.documentElement.clientWidth ||
          (!!content && content.scrollWidth > content.clientWidth)
        )
      }),
      `${viewport.name}: no hay desborde horizontal`,
    ).toBe(false)

    const action = page.getByRole('button', { name: 'Corregir resultado' })
    await action.focus()
    await expect(action).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDirCorrect, 'responsive', viewport.name, 'foco-corregir.png'),
      animations: 'disabled',
    })
  }
})

test('HU-068: movimiento reducido elimina la transición del diálogo de corrección', async ({
  page,
}) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openDetail(page, 1280, 900, { initialStatus: 'completed' })
  await page.getByRole('button', { name: 'Corregir resultado' }).click()
  await expect(page.getByRole('dialog', { name: 'Corregir resultado' })).toBeVisible()
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-dialog') as HTMLElement)
            .transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})

test('fidelidad #191: reflow, teclado y foco en los viewports obligatorios', async ({ page }) => {
  await installDetailMock(page)
  const viewports = [
    { name: '320', width: 320, height: 720 },
    { name: '360', width: 360, height: 800 },
    { name: '768', width: 768, height: 1024 },
    { name: '1280', width: 1280, height: 900 },
    // El harness del repositorio aproxima el zoom de texto al 200 % con este viewport.
    { name: '1280-zoom200', width: 640, height: 450 },
  ]

  for (const viewport of viewports) {
    await page.setViewportSize({ width: viewport.width, height: viewport.height })
    await page.goto('/panel/turnos/turno-mateo?date=2026-05-24&barberId=barbero-julian')
    await expect(page.getByRole('heading', { name: 'Mateo Rojas' })).toBeVisible()
    expect(await page.evaluate(() => window.innerWidth)).toBe(viewport.width)
    expect(await page.evaluate(() => window.innerHeight)).toBe(viewport.height)
    expect(
      await page.evaluate(() => {
        const content = document.querySelector('.private-shell__content')
        return (
          document.documentElement.scrollWidth > document.documentElement.clientWidth ||
          (!!content && content.scrollWidth > content.clientWidth)
        )
      }),
      `${viewport.name}: no hay desborde horizontal`,
    ).toBe(false)

    const action = page.getByRole('button', { name: 'Reprogramar turno' })
    await action.focus()
    await expect(action).toBeFocused()
    await page.screenshot({
      path: path.join(evidenceDir, 'responsive', viewport.name, 'foco-reprogramar.png'),
      animations: 'disabled',
    })
  }
})

test('fidelidad #191: movimiento reducido elimina la transición del diálogo', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' })
  await openDetail(page, 1280, 900)
  await page.getByRole('button', { name: 'Reprogramar turno' }).click()
  await expect(page.getByRole('dialog', { name: 'Reprogramar turno' })).toBeVisible()
  expect(
    Number.parseFloat(
      await page.evaluate(
        () =>
          getComputedStyle(document.querySelector('.base-dialog') as HTMLElement)
            .transitionDuration,
      ),
    ),
  ).toBeLessThanOrEqual(0.001)
})
