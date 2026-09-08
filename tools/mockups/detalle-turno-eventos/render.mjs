/*
 * Generador determinista de los mockups de `/panel/turnos/:appointmentId`
 * (detalle, historial y reprogramación).
 *
 * Comparte criterio y lenguaje visual con `tools/mockups/panel-agenda-eventos`,
 * `tools/mockups/nuevo-turno-eventos` y `tools/mockups/auth-eventos`: HTML
 * compuesto con los tokens reales del sistema y las fuentes auto-hosteadas del
 * proyecto, rasterizado con el Chromium de Playwright que ya usa la suite e2e.
 * El encabezado y el dock son exactamente el cascarón ya validado en esos
 * atlas, sin controles de cuenta decorativos ni navegación duplicada.
 *
 * Decisiones de composición propias de esta pantalla:
 *
 *  - La ficha del turno es PERGAMINO, el mismo material con el que la agenda
 *    dibuja un turno vigente: quien toca una fila de papel en `/panel` abre
 *    aquí esa misma fila en grande. El historial se queda sobre tinta, así
 *    que el registro (papel) y el libro de eventos (tinta) se distinguen sin
 *    necesidad de un rótulo que lo explique.
 *  - Un turno TERMINAL pierde el papel y pasa a contorno sobre tinta, igual
 *    que su fila en la agenda: un turno vigente es una superficie presente,
 *    uno cerrado es un registro. Esa misma regla explica sin texto por qué el
 *    encabezado se queda sin la acción de reprogramar.
 *  - El eje del historial usa el rombo del divisor NAVA, no el punto redondo
 *    del código actual: el círculo era el único elemento redondo del sistema.
 *    El rombo sólido marca el evento más reciente YA CARGADO, y solo cuando
 *    no queda cursor pendiente; si el historial sigue paginado, el eje
 *    continúa hasta «Cargar más» con un rombo hueco.
 *  - El diálogo de reprogramación es un formulario, así que vive sobre tinta
 *    (superficie `--ink-2` con filete de latón) con los mismos controles de
 *    los atlas hermanos, nunca sobre pergamino.
 *
 * Uso:
 *   node tools/mockups/detalle-turno-eventos/render.mjs
 */
import playwright from '../../../apps/web/node_modules/@playwright/test/index.js'
import { readFileSync, mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import {
  COPY,
  SCREENS,
  STATUS_LABELS,
  STATUS_VARIANT,
  HISTORY_EVENT_LABELS,
  HISTORY_FIELD_LABELS,
  INSTANTS,
  NAV_PRIMARY,
  NAV_SECONDARY,
  SHOP_NAME,
} from './content.mjs'

const { chromium } = playwright

const HERE = dirname(fileURLToPath(import.meta.url))
const REPO = resolve(HERE, '../../..')
const OUT = resolve(
  REPO,
  'docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/detalle-turno-eventos',
)
const FONTS = resolve(REPO, 'apps/web/src/assets/fonts')

const font = (f) => readFileSync(resolve(FONTS, f)).toString('base64')
const SERIF = font('instrument-serif-latin-400.woff2')
const SANS = font('instrument-sans-latin-400-700.woff2')

const esc = (s) => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

// --------------------------------------------------------------- iconos

const NAV_ICONS = {
  agenda:
    '<rect x="3" y="4" width="18" height="17" rx="2"/><line x1="3" y1="9" x2="21" y2="9"/><line x1="8" y1="2" x2="8" y2="6"/><line x1="16" y1="2" x2="16" y2="6"/>',
  servicios:
    '<path d="M4 4h10l6 6v10a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1z"/><path d="M14 4v6h6"/><line x1="8" y1="13" x2="15" y2="13"/><line x1="8" y1="17" x2="12" y2="17"/>',
  barberos: '<circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 3.6-7 8-7s8 3 8 7"/>',
  horarios: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3.5 2"/>',
  bloqueos: '<rect x="4" y="10" width="16" height="10" rx="1.5"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/>',
  configuracion:
    '<path d="M4 6h11M19 6h1M4 12h1M9 12h11M4 18h6M14 18h6"/><circle cx="17" cy="6" r="2"/><circle cx="7" cy="12" r="2"/><circle cx="12" cy="18" r="2"/>',
  'servicios-barbero':
    '<circle cx="9" cy="7" r="3"/><path d="M2 21c0-3.5 3-6 7-6s7 2.5 7 6"/><path d="M16 4a3 3 0 0 1 0 6"/><path d="M22 21c0-2.7-1.7-4.8-4-5.6"/>',
  mas: '<circle cx="5" cy="12" r="1.5"/><circle cx="12" cy="12" r="1.5"/><circle cx="19" cy="12" r="1.5"/>',
  // El «←» del enlace real y la «✕» del diálogo se dibujan con el mismo
  // trazo que el resto de la iconografía, en vez de como glifos sueltos.
  back: '<line x1="20" y1="12" x2="5" y2="12"/><polyline points="11,18 5,12 11,6"/>',
  close: '<line x1="6" y1="6" x2="18" y2="18"/><line x1="18" y1="6" x2="6" y2="18"/>',
}

const icon = (n) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">${NAV_ICONS[n]}</svg>`

const ALERT_EYEBROW = { danger: 'Error', warning: 'Atención', info: 'Nota', success: 'Confirmación' }

const alertHtml = (a) => `
  <div class="alert alert--${a.variant}" role="alert">
    <div class="alert__body">
      <p class="alert__eyebrow">${esc(ALERT_EYEBROW[a.variant])}</p>
      <p class="alert__title">${esc(a.title)}</p>
      ${a.body ? `<p class="alert__text">${esc(a.body)}</p>` : ''}
      ${a.action ? `<span class="alert__action">${esc(a.action)}</span>` : ''}
    </div>
  </div>`

const spinnerHtml = (size = 'lg') => `<span class="spinner spinner--${size}"><i></i><b></b></span>`

// ------------------------------------------------------------- cascarón

const headerHtml = () => `
  <header class="appbar">
    <div class="appbar__brand">
      <span class="appbar__mark">NAVA</span>
      <span class="appbar__shop">${esc(SHOP_NAME)}</span>
    </div>
  </header>`

const navHtml = (viewport) => {
  const items = viewport === 'desktop' ? [...NAV_PRIMARY, ...NAV_SECONDARY] : NAV_PRIMARY
  const cells = items
    .map(
      (i) => `
      <span class="dock__item${i.label === 'Agenda' ? ' dock__item--active' : ''}">
        <span class="dock__icon">${icon(i.icon)}</span>
        <span class="dock__label">${esc(i.label)}</span>
      </span>`,
    )
    .join('')
  const more =
    viewport === 'mobile'
      ? `<span class="dock__item">
           <span class="dock__icon">${icon('mas')}</span>
           <span class="dock__label">Más</span>
         </span>`
      : ''
  return `<nav class="dock">${cells}${more}</nav>`
}

/*
 * «Volver a la agenda» conserva la fecha y el barbero de origen (HU-063), así
 * que es el único camino de salida y encabeza la pantalla, por encima del
 * nombre. Es un enlace, no un botón: no compite con la acción primaria.
 */
const crumbHtml = () => `
  <p class="crumb"><span class="crumb__icon">${icon('back')}</span>${esc(COPY.back)}</p>`

const pageStateHtml = (s) => `
  <div class="pagestate pagestate--${s.tone}">
    <span class="pagestate__divider"><i></i><b></b><i></i></span>
    <p class="pagestate__eyebrow">${esc(s.eyebrow)}</p>
    <p class="pagestate__title">${esc(s.title)}</p>
    <p class="pagestate__body">${esc(s.body)}</p>
    ${s.action ? `<span class="btn btn--secondary pagestate__action">${esc(s.action)}</span>` : ''}
  </div>`

// --------------------------------------------------------- ficha del turno

const TERMINAL = (status) => status !== 'confirmed'

const cellHtml = (label, value, wide) => `
  <div class="cell${wide ? ' cell--wide' : ''}">
    <dt class="cell__label">${esc(label)}</dt>
    <dd class="cell__value">${esc(value)}</dd>
  </div>`

/*
 * Ficha del turno. Es el `<dl>` real, con sus mismos rótulos y en el mismo
 * orden del template, pero compuesto como una ficha y no como una tabla de
 * siete renglones idénticos: siete filas con el mismo peso no tenían foco y
 * dejaban una columna de rótulos vacía a lo ancho de media hoja.
 *
 *  - «Hora» encabeza en serif y a tamaño de titular. Es el primer `<dt>` del
 *    código y el dato por el que un barbero abre esta pantalla; el resto son
 *    datos de apoyo.
 *  - Los cinco hechos siguientes caen en dos columnas, en el mismo orden de
 *    lectura del template, así que la hoja se llena a lo ancho.
 *  - La nota del cliente cierra la ficha en serif: es una cita textual de una
 *    persona, no un campo más de la retícula.
 *
 * Contacto y nota solo existen cuando su dato existe, igual que en el código,
 * y ningún dato del cliente aparece en ninguna otra superficie (RN-HIS-01).
 */
const recordHtml = (a) => {
  const contact = [a.customerPhone, a.customerEmail].filter(Boolean).join(' · ')
  return `
    <dl class="ticket${TERMINAL(a.status) ? ' ticket--terminal' : ''}">
      <div class="ticket__hero">
        <dt class="ticket__label">${esc(COPY.factTime)}</dt>
        <dd class="ticket__when">
          <span class="ticket__time">${esc(a.timeRange)}</span>
          ${a.timezone ? `<span class="ticket__zone">${esc(COPY.zoneSuffix(a.timezone))}</span>` : ''}
        </dd>
      </div>
      <div class="ticket__grid">
        ${cellHtml(COPY.factAttendee, a.attendeeName)}
        ${cellHtml(
          COPY.factService,
          `${a.serviceName} · ${a.durationMinutes} min · ${a.priceAmount} ${a.currency}`,
        )}
        ${cellHtml(COPY.factBarber, a.barberFullName)}
        ${cellHtml(COPY.factCustomer, a.customerFullName)}
        ${contact ? cellHtml(COPY.factContact, contact, true) : ''}
      </div>
      ${
        a.customerNote
          ? `<div class="ticket__note">
               <dt class="ticket__label">${esc(COPY.factNote)}</dt>
               <dd class="ticket__quote">${esc(a.customerNote)}</dd>
             </div>`
          : ''
      }
    </dl>`
}

// ------------------------------------------------------------- historial

/*
 * Un evento del historial. El eje es el rombo del divisor NAVA sobre una línea
 * de 1 px; los campos modificados se componen con el rótulo del campo en
 * versalitas y los dos valores debajo, para que una línea larga no se parta en
 * medio de una fecha. El texto conserva lo que dibuja el código —rótulo del
 * campo, valor anterior, flecha, valor nuevo— y el orden real de la API.
 */
const changeHtml = (c) => `
  <li class="change">
    <span class="change__field">${esc(HISTORY_FIELD_LABELS[c.field] ?? c.field)}</span>
    <span class="change__values">
      <span class="change__from">${esc(c.from ?? '—')}</span>
      <span class="change__arrow" aria-hidden="true">→</span>
      <span class="change__to">${esc(c.to ?? '—')}</span>
    </span>
  </li>`

const eventHtml = (e, { current, tail }) => `
  <li class="event${current ? ' event--current' : ''}${tail ? ' event--tail' : ''}">
    <span class="event__mark" aria-hidden="true"></span>
    <span class="event__head">
      <span class="event__name">${esc(HISTORY_EVENT_LABELS[e.eventType] ?? e.eventType)}</span>
      <span class="event__meta">${esc(e.occurredAt)} · ${esc(e.actorLabel)}</span>
    </span>
    ${e.reason ? `<p class="event__reason">${esc(e.reason)}</p>` : ''}
    ${
      e.changes.length > 0
        ? `<ul class="changes">${e.changes.map(changeHtml).join('')}</ul>`
        : ''
    }
  </li>`

const historySkeletonHtml = () => `
  <div class="skel" aria-hidden="true">
    <p class="skel__status">${spinnerHtml('sm')}<span>${esc(COPY.historyLoading)}</span></p>
    <ul class="timeline">
      ${[0, 1]
        .map(
          () => `<li class="event">
            <span class="event__mark" aria-hidden="true"></span>
            <span class="event__head">
              <span class="skel__bar skel__bar--name"></span>
              <span class="skel__bar skel__bar--meta"></span>
            </span>
          </li>`,
        )
        .join('')}
    </ul>
  </div>`

const historyHtml = (h) => {
  let body = ''
  if (h.status === 'loading') {
    body = historySkeletonHtml()
  } else if (h.status === 'error') {
    body = alertHtml({
      variant: 'warning',
      title: COPY.historyErrorTitle,
      body: COPY.historyErrorBody,
      action: COPY.retry,
    })
  } else if (h.items.length === 0) {
    body = `<p class="history__empty">${esc(COPY.historyEmpty)}</p>`
  } else {
    const last = h.items.length - 1
    body = `
      <ul class="timeline">
        ${h.items
          .map((e, i) =>
            eventHtml(e, { current: i === last && !h.more, tail: i === last && !h.more }),
          )
          .join('')}
        ${
          h.more
            ? `<li class="event event--more event--tail">
                 <span class="event__mark event__mark--open" aria-hidden="true"></span>
                 <span class="btn btn--ghost btn--sm">${esc(COPY.loadMore)}</span>
               </li>`
            : ''
        }
      </ul>`
  }

  return `
    <section class="history">
      <h2 class="history__title">${esc(COPY.historyTitle)}</h2>
      ${body}
    </section>`
}

// --------------------------------------------------------------- diálogo

const fieldHtml = ({ label, value, disabled }) => `
  <div class="field">
    <span class="field__label">${esc(label)}<span class="field__req" aria-hidden="true">*</span></span>
    <div class="control control--filled${disabled ? ' control--off' : ''}">
      <span class="control__value">${esc(value)}</span>
    </div>
  </div>`

const dialogHtml = (d) => `
  <div class="modal">
    <div class="modal__scrim"></div>
    <div class="modal__panel" role="dialog" aria-modal="true">
      <div class="modal__head">
        <p class="modal__title">${esc(COPY.dialogTitle)}</p>
        <span class="modal__close" aria-label="Cerrar">${icon('close')}</span>
      </div>
      <div class="modal__body">
        ${d.alert ? alertHtml(d.alert) : ''}
        <p class="modal__current">${esc(COPY.dialogCurrent(d.currentRange))}</p>
        <div class="row">
          ${fieldHtml({ label: COPY.dialogDateLabel, value: d.date, disabled: d.saving })}
          ${fieldHtml({ label: COPY.dialogTimeLabel, value: d.time, disabled: d.saving })}
        </div>
        <p class="modal__preview">${esc(COPY.dialogPreview(d.previewFrom, d.previewTo))}</p>
        <div class="modal__actions">
          <span class="btn btn--ghost">${esc(COPY.cancel)}</span>
          <span class="btn btn--primary${d.saving ? ' btn--disabled' : ''}">
            ${d.saving ? spinnerHtml('xs') : ''}${esc(COPY.confirm)}
          </span>
        </div>
      </div>
    </div>
  </div>`

// ----------------------------------------------------------- contenido

const detailHtml = (screen) => {
  const a = screen.appointment
  return `
    ${crumbHtml()}
    <div class="pagehead">
      <div class="pagehead__stack">
        <h1 class="pagehead__title">${esc(a.attendeeName)}</h1>
        <span class="badge badge--${STATUS_VARIANT[a.status]}">
          <i class="badge__dot" aria-hidden="true"></i>${esc(STATUS_LABELS[a.status])}
        </span>
      </div>
      ${
        a.status === 'confirmed'
          ? `<span class="btn btn--secondary">${esc(COPY.reschedule)}</span>`
          : ''
      }
    </div>
    <div class="body">
      ${recordHtml(a)}
      ${historyHtml(screen.history)}
    </div>`
}

/*
 * Esqueleto de la pantalla. Con el detalle todavía sin resolver, la espera
 * conserva la geometría de lo que va a llegar —la ficha de hechos y la
 * columna del historial— en vez de dejar un renglón suelto en una pantalla
 * vacía, igual que el evento `05` de `panel-agenda-eventos`.
 */
const skeletonHtml = () => `
  ${crumbHtml()}
  <div class="skel" aria-hidden="true">
    <p class="skel__status">${spinnerHtml('sm')}<span>${esc(COPY.loadingDetail)}</span></p>
    <span class="skel__bar skel__bar--title"></span>
    <div class="body">
      <div class="ticket ticket--skel">
        <div class="ticket__hero">
          <span class="skel__bar skel__bar--label"></span>
          <span class="skel__bar skel__bar--hero"></span>
        </div>
        <div class="ticket__grid">
          ${Array.from({ length: 5 })
            .map(
              (_, i) => `<div class="cell${i === 4 ? ' cell--wide' : ''}">
                <span class="skel__bar skel__bar--label"></span>
                <span class="skel__bar skel__bar--value"></span>
              </div>`,
            )
            .join('')}
        </div>
        <div class="ticket__note">
          <span class="skel__bar skel__bar--label"></span>
          <span class="skel__bar skel__bar--quote"></span>
        </div>
      </div>
      <section class="history">
        <p class="history__title history__title--skel">
          <span class="skel__bar skel__bar--h2"></span>
        </p>
        <ul class="timeline">
          ${[0, 1, 2]
            .map(
              () => `<li class="event">
                <span class="event__mark" aria-hidden="true"></span>
                <span class="event__head">
                  <span class="skel__bar skel__bar--name"></span>
                  <span class="skel__bar skel__bar--meta"></span>
                </span>
              </li>`,
            )
            .join('')}
        </ul>
      </section>
    </div>
  </div>`

const contentHtml = (screen) => {
  if (screen.page === 'loading') return skeletonHtml()
  if (screen.page === 'not-found') {
    return `${crumbHtml()}${pageStateHtml({
      tone: 'warning',
      eyebrow: ALERT_EYEBROW.warning,
      title: COPY.notFoundTitle,
      body: COPY.notFoundBody,
    })}`
  }
  if (screen.page === 'load-error') {
    return `${crumbHtml()}${pageStateHtml({
      tone: 'warning',
      eyebrow: ALERT_EYEBROW.warning,
      title: COPY.loadErrorTitle,
      body: COPY.loadErrorBody,
      action: COPY.retry,
    })}`
  }
  return detailHtml(screen)
}

// ---------------------------------------------------------------- estilos

const CSS = `
@font-face{font-family:'Instrument Serif';font-weight:400;font-display:block;src:url(data:font/woff2;base64,${SERIF}) format('woff2')}
@font-face{font-family:'Instrument Sans';font-weight:400 700;font-display:block;src:url(data:font/woff2;base64,${SANS}) format('woff2')}

:root{
  --ink:#101b2b; --ink-2:#16243a; --canvas:#f4f0e7; --surface:#fff;
  --paper:#e8e2d8; --paper-2:#d9d2c5;
  --on-ink:#f4f0e7; --on-ink-2:#a8b0bd; --brass:#b8955a; --brass-deep:#765c2f;
  --text:#2a2d32; --text-2:#5e625f; --border:#c9c0b2;
  --danger-s:#f8edec; --danger-t:#8a2c2c; --danger-b:#a43a3a;
  --warning-s:#f8f1df; --warning-t:#775019; --warning-b:#9a6a24;
  --info-s:#e9eef3; --info-t:#23405b; --info-b:#667d93;
  --success-s:#eaf0eb; --success-t:#325d43; --success-b:#748477;
  --inactive-s:#eeece8; --inactive-t:#56514a; --inactive-b:#c9c0b2;
  --serif:'Instrument Serif',Georgia,serif;
  --sans:'Instrument Sans',system-ui,sans-serif;
}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:var(--sans);background:var(--ink);color:var(--on-ink);-webkit-font-smoothing:antialiased}
.shell{display:flex;flex-direction:column;min-height:100vh}

/* ---- barra de aplicación (idéntica a los atlas hermanos) ---- */
.appbar{display:flex;align-items:center;justify-content:space-between;
  border-bottom:1px solid rgba(184,149,90,.45)}
.appbar__brand{display:flex;align-items:baseline;gap:14px}
.appbar__mark{font-family:var(--serif);color:var(--on-ink);letter-spacing:.02em;line-height:1}
.appbar__shop{color:var(--on-ink-2)}

/* ---- regreso a la agenda ---- */
.main{flex:1;display:flex;flex-direction:column}
.crumb{display:inline-flex;align-items:center;align-self:flex-start;color:var(--brass);
  font-weight:600;text-transform:uppercase}
.crumb__icon{display:flex}

/* ---- encabezado de página ---- */
.pagehead{display:flex;align-items:flex-start;justify-content:space-between}
/* La insignia acompaña al nombre en su misma línea de base, como en el
   header real del template: colgada debajo quedaba como etiqueta suelta. */
.pagehead__stack{display:flex;min-width:0}
.pagehead__title{font-family:var(--serif);color:var(--on-ink);line-height:1.05}

/* Sobre tinta la insignia usa el par CLARO SÓLIDO del sistema: el tinte al
   8 % de la agenda está calculado para leerse sobre pergamino y su texto no
   alcanza AA sobre #101b2b. */
.badge{display:inline-flex;align-items:center;border:1px solid;border-radius:2px;
  font-weight:600;white-space:nowrap}
.badge__dot{display:block;transform:rotate(45deg);background:currentColor;flex:none}
.badge--info{background:var(--info-s);border-color:var(--info-b);color:var(--info-t)}
.badge--success{background:var(--success-s);border-color:var(--success-b);color:var(--success-t)}
.badge--neutral{background:var(--inactive-s);border-color:var(--inactive-b);color:var(--inactive-t)}
.badge--danger{background:var(--danger-s);border-color:var(--danger-b);color:var(--danger-t)}
.badge--warning{background:var(--warning-s);border-color:var(--warning-b);color:var(--warning-t)}

/* ---- retícula: ficha + historial ----
   El bloque ocupa el alto restante y las columnas se estiran, de modo que el
   filete que separa la ficha del historial recorre la página entera: sin eso,
   un turno con pocos hechos dejaba el divisor colgando a media altura y la
   mitad inferior de la pantalla sin pertenecer a nada. El papel, en cambio, no
   se estira: una hoja vacía a media página pesa más que el registro que
   contiene. */
.body{display:grid;flex:1;align-items:start}

/* ---- ficha del turno: pergamino, el material de los REGISTROS ---- */
.ticket{background:var(--paper);color:var(--text);border-radius:2px;
  border-left:3px solid var(--brass-deep)}
.ticket__label{color:var(--brass-deep);font-weight:600;text-transform:uppercase;display:block}
/* El horario encabeza la ficha en serif: es el primer hecho de la lista de
   definición y el dato por el que se abre esta pantalla. */
.ticket__when{display:flex;flex-direction:column}
.ticket__time{font-family:var(--serif);color:var(--text);line-height:1.08;
  font-variant-numeric:tabular-nums}
.ticket__zone{color:var(--text-2)}
.ticket__grid{display:grid;border-top:1px solid rgba(42,45,50,.16)}
.cell{min-width:0}
.cell--wide{grid-column:1 / -1}
.cell__label{color:var(--brass-deep);font-weight:600;text-transform:uppercase;display:block}
.cell__value{color:var(--text)}
/* La nota es una cita textual de una persona, no un campo más de la retícula:
   cierra la ficha en la serif del sistema. */
.ticket__note{border-top:1px solid rgba(42,45,50,.16)}
.ticket__quote{font-family:var(--serif);color:var(--text)}
/* Terminal: deja de ser papel y pasa a contorno sobre tinta, igual que su
   fila en la agenda. Es también lo que explica, sin texto, que el encabezado
   se quede sin la acción de reprogramar. */
.ticket--terminal{background:transparent;border:1px solid rgba(244,240,231,.2);
  border-left:3px solid rgba(244,240,231,.3)}
.ticket--terminal .ticket__grid,.ticket--terminal .ticket__note{
  border-top-color:rgba(244,240,231,.14)}
.ticket--terminal .ticket__label,.ticket--terminal .cell__label,
.ticket--terminal .ticket__zone{color:var(--on-ink-2)}
.ticket--terminal .ticket__time,.ticket--terminal .cell__value,
.ticket--terminal .ticket__quote{color:var(--on-ink)}
.ticket--skel{background:rgba(244,240,231,.035);border:1px solid rgba(244,240,231,.1);
  border-left:3px solid rgba(184,149,90,.45)}
.ticket--skel .ticket__grid,.ticket--skel .ticket__note{
  border-top-color:rgba(244,240,231,.1)}

/* ---- historial sobre tinta ---- */
.history{display:flex;flex-direction:column}
/* Filete de latón bajo el título: el mismo remate editorial del wordmark y del
   divisor. Es lo que ancla la columna sin necesidad de un borde que recorra la
   página entera hasta el vacío. */
.history__title{font-family:var(--serif);color:var(--on-ink);line-height:1.15;font-weight:400;
  border-bottom:1px solid rgba(184,149,90,.45)}
.history__title--skel{line-height:0}
.history__empty{color:var(--on-ink-2)}
.timeline{list-style:none;display:flex;flex-direction:column}
.event{position:relative;display:flex;flex-direction:column}
/* Eje continuo entre eventos. El último evento cargado solo corta el eje
   cuando ya no queda historial por traer. */
.event::before{content:'';position:absolute;left:3px;top:11px;bottom:0;width:1px;
  background:rgba(244,240,231,.16)}
.event--tail::before{display:none}
.event__mark{position:absolute;left:0;top:7px;width:7px;height:7px;
  border:1px solid var(--brass);transform:rotate(45deg)}
.event--current .event__mark{background:var(--brass)}
.event__mark--open{border-style:dashed}
/* «Cargar más» es un control, no un evento: su rombo se centra con el botón y
   el botón se mide por su texto. */
.event--more{align-items:flex-start}
.event--more .event__mark{top:14px}
/* Nombre y sello del evento comparten renglón, como el __history-main real:
   apilarlos alargaba cada evento sin aportar jerarquía. */
.event__head{display:flex;flex-wrap:wrap;align-items:baseline;justify-content:space-between}
.event__name{font-weight:600;color:var(--on-ink)}
.event__meta{color:var(--on-ink-2)}
.event__reason{color:var(--on-ink-2)}
.changes{list-style:none;display:flex;flex-direction:column;
  border-left:1px solid rgba(244,240,231,.16)}
.change{display:flex;flex-direction:column}
/* El rótulo del campo va en gris, no en latón: con cuatro rótulos dorados por
   evento el historial se volvía un muestrario de oro y el latón dejaba de
   señalar nada. Aquí solo la flecha —el hecho del cambio— es de latón. */
.change__field{color:var(--on-ink-2);font-weight:600;text-transform:uppercase}
.change__values{display:flex;flex-wrap:wrap;align-items:baseline}
.change__from{color:var(--on-ink-2)}
.change__arrow{color:var(--brass)}
.change__to{color:var(--on-ink);font-weight:600}

/* ---- acciones ---- */
.btn{display:inline-flex;align-items:center;justify-content:center;border-radius:2px;
  font-weight:600;letter-spacing:.02em;white-space:nowrap}
.btn--primary{background:var(--brass);color:var(--ink)}
.btn--secondary{background:var(--canvas);color:var(--ink)}
.btn--ghost{background:transparent;color:var(--brass);
  border:1px solid rgba(184,149,90,.5);border-bottom:2px solid var(--brass)}
.btn--disabled{opacity:.55}

/* ---- estados de página ---- */
.pagestate{display:flex;flex-direction:column;align-items:center;text-align:center;
  margin:auto;max-width:520px}
.pagestate__divider{display:flex;align-items:center;justify-content:center}
.pagestate__divider i{display:block;height:1px;background:var(--brass);opacity:.65}
.pagestate__divider b{display:block;background:var(--warning-b);transform:rotate(45deg);flex:none}
.pagestate__eyebrow{font-weight:600;text-transform:uppercase;color:var(--warning-b)}
.pagestate__title{font-family:var(--serif);color:var(--on-ink);line-height:1.15}
.pagestate__body{color:var(--on-ink-2)}

.spinner{position:relative;display:inline-block;flex:none}
.spinner i{position:absolute;inset:0;border-radius:50%;
  border:2px solid rgba(184,149,90,.22);border-top-color:var(--brass);
  border-right-color:var(--brass);transform:rotate(-38deg)}
.spinner b{position:absolute;top:50%;left:50%;background:var(--brass);
  transform:translate(-50%,-50%) rotate(45deg)}
.spinner--xs i{border-width:1.5px}
.btn--primary .spinner i{border-color:rgba(16,27,43,.25);border-top-color:var(--ink);
  border-right-color:var(--ink)}
.btn--primary .spinner b{background:var(--ink)}

/* ---- esqueletos ---- */
.skel{display:flex;flex-direction:column}
.skel__status{display:flex;align-items:center;color:var(--brass);font-weight:600;
  text-transform:uppercase}
.skel__bar{display:block;background:rgba(244,240,231,.14);border-radius:2px}
.skel__bar--meta{background:rgba(244,240,231,.09)}
.skel__bar--value{background:rgba(244,240,231,.09)}
/* Los siete hechos no miden lo mismo: barras idénticas anuncian una tabla
   que no va a llegar. */
.record--skel .fact:nth-child(2n) .skel__bar--value{width:44%}
.record--skel .fact:nth-child(3n) .skel__bar--value{width:70%}

/* ---- alertas: misma nota al margen de los atlas hermanos ---- */
.alert{display:flex;border:1px solid;border-left-width:4px;border-radius:2px}
.alert__body{flex:1;display:flex;flex-direction:column}
.alert__eyebrow{font-weight:600;text-transform:uppercase;opacity:.85}
.alert__title{font-weight:600;color:var(--ink)}
.alert__action{align-self:flex-start;border:1px solid currentColor;border-bottom-width:2px;
  border-radius:2px;font-weight:600;text-transform:uppercase;letter-spacing:.08em}
.alert--danger{background:var(--danger-s);border-color:var(--danger-b);color:var(--danger-t)}
.alert--warning{background:var(--warning-s);border-color:var(--warning-b);color:var(--warning-t)}

/* ---- diálogo de reprogramación: formulario, luego vive sobre tinta ---- */
.modal{position:fixed;inset:0;display:flex;align-items:center;justify-content:center}
.modal__scrim{position:absolute;inset:0;background:rgba(7,12,20,.74)}
.modal__panel{position:relative;width:100%;background:var(--ink-2);
  border:1px solid rgba(244,240,231,.14);border-top:2px solid var(--brass);
  border-radius:2px;box-shadow:0 28px 70px rgb(0 0 0 / 55%)}
.modal__head{display:flex;align-items:center;justify-content:space-between;
  border-bottom:1px solid rgba(244,240,231,.12)}
.modal__title{font-family:var(--serif);color:var(--on-ink);line-height:1.1}
.modal__close{display:flex;color:var(--on-ink-2);flex:none}
.modal__body{display:flex;flex-direction:column}
/* El horario vigente es el contexto contra el que se decide, así que se
   enmarca con filete de latón en vez de quedar como un renglón gris más. */
.modal__current{color:var(--on-ink);background:rgba(244,240,231,.04);
  border-left:2px solid var(--brass);border-radius:2px}
.modal__preview{color:var(--on-ink-2)}
.modal__actions{display:flex;justify-content:flex-end}
.row{display:flex}
.field{display:flex;flex-direction:column;flex:1;min-width:0}
.field__label{color:var(--brass);font-weight:600;text-transform:uppercase}
.field__req{color:var(--brass);opacity:.8}
.control{display:flex;align-items:center;background:rgba(244,240,231,.04);
  border:1px solid rgba(244,240,231,.12);border-bottom:2px solid rgba(244,240,231,.3);
  border-radius:2px}
.control--filled{background:rgba(244,240,231,.06);border-bottom-color:var(--brass)}
.control--off{opacity:.45;border-bottom-color:rgba(244,240,231,.2)}
.control__value{flex:1;color:var(--on-ink);line-height:1.3;overflow:hidden;
  text-overflow:ellipsis;white-space:nowrap;font-variant-numeric:tabular-nums}

/* ---- dock ---- */
.dock{display:flex;border-top:1px solid rgba(244,240,231,.14);margin-top:auto}
.dock__item{flex:1;display:flex;align-items:center;justify-content:center;
  color:var(--on-ink-2)}
.dock__item--active{color:var(--brass);font-weight:600}
.dock__icon{display:flex}
`

const SCALES = {
  desktop: `
    .appbar{padding:20px 40px}
    .appbar__mark{font-size:34px}
    .appbar__shop{font-size:16px}
    .main{padding:22px 40px 26px;gap:16px}

    .crumb{gap:9px;font-size:11px;letter-spacing:.14em}
    .crumb__icon svg{width:15px;height:15px}

    .pagehead{gap:24px}
    .pagehead__stack{align-items:baseline;gap:18px}
    .pagehead__title{font-size:40px}
    .badge{gap:9px;padding:5px 12px;font-size:12px}
    .badge__dot{width:6px;height:6px}
    .btn{height:48px;padding:0 24px;font-size:15px;gap:10px}
    .btn--sm{height:36px;padding:0 16px;font-size:12px;text-transform:uppercase;
      letter-spacing:.1em}

    /* La ficha se mide por su contenido, no por la mitad de la pantalla. No
       hay filete entre columnas: el papel y la tinta ya son dos materiales
       distintos, y una línea que cruzara la página hasta el borde inferior
       solo subrayaba el vacío de abajo. */
    .body{grid-template-columns:minmax(0,760px) minmax(0,1fr);column-gap:56px;margin-top:6px}

    .ticket{padding:0 34px}
    .ticket__label{font-size:11px;line-height:15px;letter-spacing:.15em}
    .ticket__hero{padding:26px 0 28px}
    .ticket__when{margin-top:11px;gap:7px}
    .ticket__time{font-size:33px}
    .ticket__zone{font-size:13px;line-height:18px}
    .ticket__grid{grid-template-columns:repeat(2,minmax(0,1fr));gap:24px 44px;padding:26px 0}
    .cell__label{font-size:11px;line-height:15px;letter-spacing:.15em}
    .cell__value{font-size:17px;line-height:24px;margin-top:7px}
    .ticket__note{padding:24px 0 28px}
    .ticket__quote{font-size:21px;line-height:29px;margin-top:10px}

    .history{gap:20px}
    .history__title{font-size:25px;padding-bottom:13px}
    .history__empty{font-size:15px}
    .timeline{gap:22px}
    .event{padding-left:26px;padding-bottom:22px;gap:4px}
    .event:last-child{padding-bottom:0}
    .event__head{gap:8px 16px}
    .event__name{font-size:15px;line-height:21px}
    .event__meta{font-size:13px;line-height:18px}
    .event__reason{font-size:13px;line-height:18px;margin-top:5px}
    .changes{gap:10px;margin-top:12px;padding-left:14px}
    .change{gap:4px}
    .change__field{font-size:10px;line-height:13px;letter-spacing:.15em}
    .change__values{gap:9px;font-size:13px;line-height:18px}
    .event--more{padding-bottom:0}

    .pagestate{gap:14px;padding:40px 0}
    .pagestate__divider{gap:12px;margin:2px 0 4px}
    .pagestate__divider i{width:98px}
    .pagestate__divider b{width:7px;height:7px}
    .pagestate__eyebrow{font-size:11px;line-height:15px;letter-spacing:.16em}
    .pagestate__title{font-size:30px}
    .pagestate__body{font-size:16px;line-height:24px;max-width:470px}
    .pagestate__action{margin-top:10px}

    .spinner--sm{width:18px;height:18px}
    .spinner--sm b{width:5px;height:5px}
    .spinner--xs{width:14px;height:14px}
    .spinner--xs b{width:4px;height:4px}

    .skel{gap:16px}
    .skel__status{gap:12px;font-size:12px;letter-spacing:.14em}
    .skel__bar--title{width:262px;height:36px}
    .skel__bar--h2{width:120px;height:20px}
    .skel__bar--label{width:118px;height:11px}
    .skel__bar--hero{width:62%;height:30px;margin-top:13px}
    .skel__bar--value{width:70%;height:15px;margin-top:9px}
    .skel__bar--quote{width:78%;height:20px;margin-top:11px}
    .skel__bar--name{width:150px;height:13px}
    .skel__bar--meta{width:210px;height:10px}

    .alert{padding:15px 18px}
    .alert__body{gap:5px}
    .alert__eyebrow{font-size:11px;line-height:15px;letter-spacing:.15em}
    .alert__title{font-size:17px;line-height:24px}
    .alert__text{font-size:15px;line-height:22px}
    .alert__action{margin-top:8px;padding:9px 16px;font-size:12px}

    .modal{padding:32px}
    .modal__panel{max-width:560px}
    .modal__head{padding:18px 24px;gap:16px}
    .modal__title{font-size:25px}
    .modal__close svg{width:20px;height:20px}
    .modal__body{padding:20px 24px 22px;gap:16px}
    .modal__current{font-size:14px;line-height:20px;padding:11px 14px}
    .modal__preview{font-size:13px;line-height:18px;margin-top:-2px}
    .modal__actions{gap:12px;margin-top:6px}
    .row{gap:16px}
    .field{gap:7px}
    .field__label{font-size:11px;line-height:15px;letter-spacing:.14em}
    .control{height:46px;padding:0 14px}
    .control__value{font-size:15px}

    .dock{padding:0 40px}
    .dock__item{gap:10px;height:64px;font-size:14px}
    .dock__icon svg{width:19px;height:19px}
  `,
  mobile: `
    .appbar{padding:16px 20px}
    .appbar__mark{font-size:26px}
    .appbar__shop{font-size:13px}
    .main{padding:18px 20px 24px;gap:14px}

    .crumb{gap:8px;font-size:10px;letter-spacing:.14em}
    .crumb__icon svg{width:14px;height:14px}

    .pagehead{flex-direction:column;align-items:stretch;gap:16px}
    .pagehead__stack{flex-direction:column;align-items:flex-start;gap:11px}
    .pagehead__title{font-size:32px}
    .badge{gap:8px;padding:5px 11px;font-size:11px}
    .badge__dot{width:6px;height:6px}
    .btn{height:46px;padding:0 20px;font-size:15px;gap:10px}
    .pagehead .btn{width:100%}
    .btn--sm{height:36px;padding:0 14px;font-size:11px;text-transform:uppercase;
      letter-spacing:.1em}

    .body{grid-template-columns:minmax(0,1fr);row-gap:26px}

    /* Una sola columna de hechos: en 420 px, dos no dejan medida para el
       nombre del servicio ni para el contacto. */
    .ticket{padding:0 18px}
    .ticket__label{font-size:10px;line-height:14px;letter-spacing:.15em}
    .ticket__hero{padding:20px 0 22px}
    .ticket__when{margin-top:9px;gap:6px}
    .ticket__time{font-size:25px}
    .ticket__zone{font-size:12px;line-height:17px}
    .ticket__grid{grid-template-columns:minmax(0,1fr);gap:18px;padding:20px 0}
    .cell__label{font-size:10px;line-height:14px;letter-spacing:.15em}
    .cell__value{font-size:16px;line-height:22px;margin-top:6px}
    .ticket__note{padding:20px 0 22px}
    .ticket__quote{font-size:19px;line-height:26px;margin-top:8px}

    .history{gap:16px}
    .history__title{font-size:22px;padding-bottom:11px}
    .history__empty{font-size:14px}
    .timeline{gap:20px}
    .event{padding-left:22px;padding-bottom:20px;gap:4px}
    .event:last-child{padding-bottom:0}
    .event__head{flex-direction:column;gap:3px}
    .event__name{font-size:15px;line-height:21px}
    .event__meta{font-size:12px;line-height:17px}
    .event__reason{font-size:12px;line-height:17px;margin-top:5px}
    .changes{gap:9px;margin-top:11px;padding-left:12px}
    .change{gap:3px}
    .change__field{font-size:10px;line-height:13px;letter-spacing:.15em}
    .change__values{gap:8px;font-size:12px;line-height:17px}
    .event--more{padding-bottom:0}

    .pagestate{gap:12px;padding:34px 0;max-width:330px}
    .pagestate__divider{gap:10px;margin:2px 0 4px}
    .pagestate__divider i{width:78px}
    .pagestate__divider b{width:6px;height:6px}
    .pagestate__eyebrow{font-size:10px;line-height:14px;letter-spacing:.16em}
    .pagestate__title{font-size:25px}
    .pagestate__body{font-size:15px;line-height:22px}
    .pagestate__action{margin-top:8px}

    .spinner--sm{width:16px;height:16px}
    .spinner--sm b{width:5px;height:5px}
    .spinner--xs{width:14px;height:14px}
    .spinner--xs b{width:4px;height:4px}

    .skel{gap:14px}
    .skel__status{gap:10px;font-size:11px;letter-spacing:.14em}
    .skel__bar--title{width:190px;height:30px}
    .skel__bar--h2{width:110px;height:18px}
    .skel__bar--label{width:104px;height:10px}
    .skel__bar--hero{width:76%;height:24px;margin-top:11px}
    .skel__bar--value{width:64%;height:14px;margin-top:8px}
    .skel__bar--quote{width:86%;height:18px;margin-top:9px}
    .skel__bar--name{width:140px;height:13px}
    .skel__bar--meta{width:190px;height:10px;margin-top:7px}

    .alert{padding:13px 15px}
    .alert__body{gap:4px}
    .alert__eyebrow{font-size:10px;line-height:14px;letter-spacing:.15em}
    .alert__title{font-size:16px;line-height:23px}
    .alert__text{font-size:14px;line-height:20px}
    .alert__action{margin-top:8px;padding:9px 14px;font-size:11px}

    .modal{padding:16px}
    .modal__panel{max-width:none}
    .modal__head{padding:16px 18px;gap:14px}
    .modal__title{font-size:22px}
    .modal__close svg{width:19px;height:19px}
    .modal__body{padding:18px;gap:15px}
    .modal__current{font-size:13px;line-height:19px;padding:10px 12px}
    .modal__preview{font-size:12px;line-height:17px;margin-top:-2px}
    .modal__actions{gap:10px;margin-top:4px}
    .modal__actions .btn{flex:1}
    /* Dos campos de fecha y hora no caben en un renglón de 360 px: la
       retícula real (flex-wrap con base de 10rem) los apila ahí, y el atlas
       representa ese quiebre. */
    .row{flex-direction:column;gap:14px}
    .field{gap:6px}
    .field__label{font-size:10px;line-height:14px;letter-spacing:.14em}
    .control{height:46px;padding:0 13px}
    .control__value{font-size:15px}

    .dock{padding:0 8px}
    .dock__item{flex-direction:column;gap:5px;height:66px;font-size:11px}
    .dock__icon svg{width:19px;height:19px}
  `,
}

const page = (viewport, screen) => {
  const dialog = screen.dialog
    ? dialogHtml({ ...screen.dialog, currentRange: screen.appointment.timeRange })
    : ''
  return `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><style>${CSS}${SCALES[viewport]}</style></head>
<body>
  <div class="shell">
    ${headerHtml()}
    <main class="main">
      ${contentHtml(screen)}
    </main>
    ${navHtml(viewport)}
  </div>
  ${dialog}
</body></html>`
}

const VIEWPORTS = {
  desktop: { width: 1440, height: 1024, deviceScaleFactor: 1 },
  mobile: { width: 420, height: 935, deviceScaleFactor: 2 },
}

async function main() {
  const browser = await chromium.launch()
  const problems = []
  const heights = []

  // Los rótulos de fecha y hora del atlas se calculan en Node con el mismo
  // `Intl.DateTimeFormat` que usa `formatInstantInTimezone`. Se vuelven a
  // calcular aquí, dentro del Chromium que rasteriza, para que una diferencia
  // de ICU no deje el atlas mostrando un formato que el navegador no produce.
  const checker = await browser.newPage()
  const rendered = await checker.evaluate((spec) => {
    const out = {}
    for (const [key, v] of Object.entries(spec.values)) {
      const options =
        v.style === 'clock'
          ? { timeZone: spec.timezone, timeStyle: 'short' }
          : { timeZone: spec.timezone, dateStyle: 'medium', timeStyle: 'short' }
      out[key] = new Intl.DateTimeFormat('es-CO', options).format(new Date(v.iso))
    }
    return out
  }, INSTANTS)
  for (const [key, v] of Object.entries(INSTANTS.values)) {
    if (rendered[key] !== v.text) {
      problems.push(`intl/${key}: Node "${v.text}" ≠ Chromium "${rendered[key]}"`)
    }
  }
  await checker.close()

  for (const [viewport, cfg] of Object.entries(VIEWPORTS)) {
    const ctx = await browser.newContext({
      viewport: { width: cfg.width, height: cfg.height },
      deviceScaleFactor: cfg.deviceScaleFactor,
    })
    const tab = await ctx.newPage()

    for (const [name, screen] of Object.entries(SCREENS)) {
      await tab.setContent(page(viewport, screen), { waitUntil: 'load' })
      await tab.evaluate(() => document.fonts.ready)

      const box = await tab.evaluate(() => ({
        h: document.documentElement.scrollHeight,
        w: document.documentElement.scrollWidth,
      }))
      if (box.w > cfg.width) {
        problems.push(`${viewport}/${name}: desborde horizontal ${box.w}px`)
      }
      // La pantalla de escritorio debe caber en un viewport: si vuelve a
      // necesitar scroll, la composición se salió de presupuesto.
      if (viewport === 'desktop' && box.h > cfg.height) {
        problems.push(`${viewport}/${name}: desborde vertical ${box.h}px`)
      }
      if (viewport === 'desktop') heights.push(`${name}: ${box.h}px`)

      // Ningún texto puede salirse de su contenedor: el defecto más caro de
      // un atlas es prometer una caja donde el copy real no cabe.
      const spill = await tab.evaluate(() => {
        for (const el of document.querySelectorAll(
          '.ticket__time,.ticket__zone,.cell__value,.ticket__quote,.event__name,.event__meta,' +
            '.change__values,.modal__current,.modal__preview,.control__value,.btn,.badge',
        )) {
          if (el.scrollWidth > el.clientWidth + 1) return el.textContent.trim().slice(0, 40)
        }
        return null
      })
      if (spill) problems.push(`${viewport}/${name}: texto desbordado en «${spill}»`)

      const dir = resolve(OUT, viewport, 'detalle-turno')
      mkdirSync(dir, { recursive: true })
      // Un diálogo es una superficie anclada al viewport: capturarlo con
      // fullPage lo dejaría flotando sobre una página más alta que la real.
      await tab.screenshot({
        path: resolve(dir, `${name}.png`),
        fullPage: viewport === 'mobile' && !screen.dialog,
      })
    }
    await ctx.close()
  }

  await browser.close()

  if (problems.length) {
    console.error('Problemas de composición:\n- ' + problems.join('\n- '))
    process.exitCode = 1
  } else {
    console.log(`${Object.keys(SCREENS).length * 2} mockups regenerados sin desbordes.`)
    console.log('Alto de escritorio:\n- ' + heights.join('\n- '))
  }
}

await main()
