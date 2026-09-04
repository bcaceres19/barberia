/*
 * Generador determinista de los mockups de `/panel` (agenda diaria).
 *
 * Comparte criterio y lenguaje visual con `tools/mockups/auth-eventos`:
 * HTML compuesto con los tokens reales del sistema y las fuentes
 * auto-hosteadas del proyecto, rasterizado con el Chromium de Playwright que
 * ya usa la suite e2e. El motor de maquetación resuelve el ajuste de línea,
 * y la posición de cada ficha en la línea temporal se calcula desde su propia
 * hora, así que una ficha no puede quedar desalineada de su horario ni una
 * marca puede taparle la etiqueta a una hora del eje.
 *
 * Uso:
 *   node tools/mockups/panel-agenda-eventos/render.mjs
 */
import playwright from '../../../apps/web/node_modules/@playwright/test/index.js'
import { readFileSync, mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { COPY, STATUS_LABELS, PANEL, NAV_PRIMARY, NAV_SECONDARY } from './content.mjs'

const { chromium } = playwright

const HERE = dirname(fileURLToPath(import.meta.url))
const REPO = resolve(HERE, '../../..')
const OUT = resolve(
  REPO,
  'docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/panel-agenda-eventos',
)
const FONTS = resolve(REPO, 'apps/web/src/assets/fonts')

const font = (f) => readFileSync(resolve(FONTS, f)).toString('base64')
const SERIF = font('instrument-serif-latin-400.woff2')
const SANS = font('instrument-sans-latin-400-700.woff2')

const esc = (s) => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

const hhmm = (m) =>
  `${String(Math.floor((m % 1440) / 60)).padStart(2, '0')}:${String(m % 60).padStart(2, '0')}`

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
}

const navIcon = (n) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">${NAV_ICONS[n]}</svg>`

// Misma familia de alerta que el atlas de autenticación: filete lateral del
// color de estado y palabra de estado en versalitas, no un glifo genérico.
const ALERT_EYEBROW = { danger: 'Error', warning: 'Atención', info: 'Nota', success: 'Confirmación' }

const alertHtml = (a) => `
  <div class="alert alert--${a.variant}">
    <div class="alert__body">
      <p class="alert__eyebrow">${esc(ALERT_EYEBROW[a.variant])}</p>
      <p class="alert__title">${esc(a.title)}</p>
      <p class="alert__text">${esc(a.body)}</p>
      ${a.action ? `<span class="alert__action">${esc(a.action)}</span>` : ''}
    </div>
  </div>`

/*
 * Retrato del barbero. Acompaña al nombre en todo lugar donde se ve o se
 * elige un barbero, y cambia de tamaño según el sitio: 'sm' en el selector
 * cerrado, 'md' en cada opción de la lista desplegada.
 *
 * Dos representaciones, y el orden importa:
 *
 *  - MONOGRAMA (por defecto). Iniciales en la serif del wordmark sobre
 *    superficie de tinta con filete de latón. Se deriva de `fullName`, que
 *    es lo único que el contrato declara hoy (`Barber` = id, fullName,
 *    createdAt, updatedAt), así que es implementable sin cambiar nada.
 *  - FOTOGRAFÍA. Requiere un campo que el contrato NO tiene. Se representa
 *    aquí porque el propietario lo pidió, pero necesita decisión e issue
 *    propios; hasta entonces el monograma es el estado real y también la
 *    reserva permanente para quien no suba foto.
 *
 * En el atlas la foto se dibuja como silueta de retrato: no se inventan
 * rostros de personas que no existen.
 */
const initials = (name) =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0])
    .join('')
    .toUpperCase()

const PORTRAIT =
  '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><circle cx="12" cy="9" r="3.6"/><path d="M4.8 20.5c0-3.7 3.2-6.2 7.2-6.2s7.2 2.5 7.2 6.2"/></svg>'

const avatarHtml = (barber, size) =>
  barber.photo
    ? `<span class="avatar avatar--${size} avatar--photo" aria-hidden="true">${PORTRAIT}</span>`
    : `<span class="avatar avatar--${size}" aria-hidden="true">${esc(initials(barber.name))}</span>`

// ------------------------------------------------------------- cascarón

const headerHtml = (s) => `
  <header class="appbar">
    <div class="appbar__brand">
      <span class="appbar__mark">NAVA</span>
      <span class="appbar__shop">${esc(s.shop)}</span>
    </div>
  </header>`

const pageHeadHtml = (s) => `
  <div class="pagehead">
    <div>
      <h1 class="pagehead__title">${esc(COPY.title)}</h1>
      ${
        s.date
          ? `<p class="pagehead__meta">${esc(s.date)}${s.timezone ? ` · Zona ${esc(s.timezone)}` : ''}</p>`
          : ''
      }
    </div>
    ${s.showCta ? `<span class="btn btn--secondary">${esc(COPY.cta)}</span>` : ''}
  </div>`

const controlsHtml = (s) => {
  if (!s.showControls) return ''
  const off = s.datesEnabled ? '' : ' control--off'
  return `
    <div class="controls">
      <div class="field field--barber">
        <span class="field__label">${esc(COPY.barberLabel)}</span>
        <div class="control control--select${s.open ? ' control--open' : ''}">
          ${s.barber ? avatarHtml(s.barber, 'sm') : ''}
          <span class="control__value${s.barber ? '' : ' control__value--ph'}">${esc(
            s.barber?.name ?? 'Selecciona un barbero',
          )}</span>
          <span class="control__caret">▾</span>
        </div>
        ${
          s.open
            ? `<ul class="options">${s.options
                .map(
                  (o) => `<li class="option${o.name === s.barber?.name ? ' option--selected' : ''}">
                    ${avatarHtml(o, 'md')}
                    <span class="option__name">${esc(o.name)}</span>
                  </li>`,
                )
                .join('')}</ul>`
            : ''
        }
      </div>
      <div class="datenav">
        <span class="btn btn--ghost${off}">${esc(COPY.prevDay)}</span>
        <div class="field">
          <span class="field__label">${esc(COPY.dateLabel)}</span>
          <div class="control${off}">
            <span class="control__value">2026-09-0${s.date === 'Hoy, 3 de septiembre' ? '3' : '4'}</span>
          </div>
        </div>
        <span class="btn btn--ghost${off}">${esc(COPY.nextDay)}</span>
      </div>
    </div>`
}

const navHtml = (viewport) => {
  const items = viewport === 'desktop' ? [...NAV_PRIMARY, ...NAV_SECONDARY] : NAV_PRIMARY
  const cells = items
    .map(
      (i) => `
      <span class="dock__item${i.active ? ' dock__item--active' : ''}">
        <span class="dock__icon">${navIcon(i.icon)}</span>
        <span class="dock__label">${esc(i.label)}</span>
      </span>`,
    )
    .join('')
  const more =
    viewport === 'mobile'
      ? `<span class="dock__item">
           <span class="dock__icon">${navIcon('mas')}</span>
           <span class="dock__label">Más</span>
         </span>`
      : ''
  return `<nav class="dock">${cells}${more}</nav>`
}

// ------------------------------------------------- línea temporal y lista

/*
 * Línea temporal de escritorio. Cada ficha se posiciona y dimensiona desde su
 * propia hora contra el eje declarado, de modo que su geometría no puede
 * contradecir su rótulo. Las etiquetas del eje viven en su propia banda y las
 * marcas ("Ahora", "Cambio de día") en otra: por construcción no se tapan.
 */
const timelineHtml = (screen) => {
  const { from, to } = screen.axis
  const span = to - from
  const pct = (m) => `${((m - from) / span) * 100}%`

  const ticks = []
  for (let m = from; m <= to; m += 60) {
    ticks.push(
      `<span class="axis__tick" style="left:${pct(m)}"><i></i><em>${hhmm(m)}</em></span>`,
    )
  }

  const marks = []
  if (screen.now != null) {
    marks.push(
      `<span class="mark mark--now" style="left:${pct(screen.now)}"><em>${esc(COPY.now)}</em></span>`,
    )
  }
  if (screen.dayChangeAt != null) {
    marks.push(
      `<span class="mark mark--day" style="left:${pct(screen.dayChangeAt)}"><em>${esc(
        COPY.dayChange,
      )}</em></span>`,
    )
  }

  const slips = screen.entries
    .map((e) => {
      const term = e.status !== 'confirmed'
      // El servicio solo aparece cuando la duración le da ancho suficiente:
      // truncarlo a dos letras no informa y ensucia la ficha.
      const wide = (e.end - e.start) / span > 0.09
      return `
      <span class="slip${term ? ' slip--terminal' : ''}"
            style="left:${pct(e.start)};width:${((e.end - e.start) / span) * 100}%">
        <em class="slip__time">${wide ? `${hhmm(e.start)}–${hhmm(e.end)}` : hhmm(e.start)}</em>
        <em class="slip__name">${esc(e.name)}</em>
        ${wide ? `<em class="slip__service">${esc(e.service)}</em>` : ''}
      </span>`
    })
    .join('')

  // Guías: una por hora, más una media hora tenue entre cada par. Dan lectura
  // de duración sin convertir el carril en una cuadrícula.
  const guides = []
  for (let m = from; m <= to; m += 30) {
    const half = (m - from) % 60 !== 0
    guides.push(
      `<span class="track__rule${half ? ' track__rule--half' : ''}" style="left:${pct(m)}"></span>`,
    )
  }

  return `
    <div class="timeline">
      <div class="axis">${ticks.join('')}</div>
      <div class="marks">${marks.join('')}</div>
      <div class="track">
        ${guides.join('')}
        ${screen.now != null ? `<span class="track__mark track__mark--now" style="left:${pct(screen.now)}"></span>` : ''}
        ${screen.dayChangeAt != null ? `<span class="track__mark track__mark--day" style="left:${pct(screen.dayChangeAt)}"></span>` : ''}
        ${slips}
      </div>
    </div>`
}

/** Lista cronológica: la representación accesible equivalente, en ambos
 *  viewports. No es un adorno de móvil. */
const listHtml = (entries) => `
  <ul class="list">
    ${entries
      .map((e) => {
        const term = e.status !== 'confirmed'
        return `
      <li class="row${term ? ' row--terminal' : ''}">
        <span class="row__time">${hhmm(e.start)}–${hhmm(e.end)}</span>
        <span class="row__body">
          <span class="row__name">${esc(e.name)}</span>
          <span class="row__service">${esc(e.service)}</span>
          ${e.endsNextDay ? `<span class="row__note">${esc(e.endsNextDay)}</span>` : ''}
        </span>
        <span class="badge badge--${e.status}">${esc(STATUS_LABELS[e.status])}</span>
      </li>`
      })
      .join('')}
  </ul>`

// ----------------------------------------------------------- contenido

/*
 * Indicador de progreso: anillo de latón interrumpido con el rombo del
 * divisor NAVA en el centro. No es un spinner genérico de librería; reusa el
 * mismo motivo que separa el panel oscuro y que marca el cambio de día, para
 * que la espera pertenezca al lenguaje de la marca. En la implementación gira
 * de forma continua y se detiene con `prefers-reduced-motion`.
 */
const spinnerHtml = (size = 'lg') => `
  <span class="spinner spinner--${size}"><i></i><b></b></span>`

/*
 * Estado de página: cuando la carga, el error o el vacío ocupan TODA el área
 * de contenido, un renglón suelto arriba a la izquierda deja la pantalla
 * pareciendo rota. Estos estados se componen centrados, con marca, título en
 * serif y una acción real, mientras que la alerta como nota al margen se
 * reserva para cuando acompaña a contenido que sigue visible (evento 10).
 */
const highlightDest = (t) =>
  t.replace(/«([^»]+)»/g, '<span class="pagestate__dest">«$1»</span>')

const pageStateHtml = (s) => `
  <div class="pagestate pagestate--${s.tone}">
    ${
      s.tone === 'loading'
        ? `<div class="pagestate__mark">${spinnerHtml()}</div>`
        : '<span class="pagestate__divider"><i></i><b></b><i></i></span>'
    }
    ${s.eyebrow ? `<p class="pagestate__eyebrow">${esc(s.eyebrow)}</p>` : ''}
    ${s.title ? `<p class="pagestate__title">${esc(s.title)}</p>` : ''}
    ${s.body ? `<p class="pagestate__body">${highlightDest(esc(s.body))}</p>` : ''}
    ${s.action ? `<span class="btn btn--secondary pagestate__action">${esc(s.action)}</span>` : ''}
  </div>`

/*
 * Esqueleto de la agenda. Cuando barbero y fecha ya están resueltos, la espera
 * conserva la geometría que va a ocupar el contenido real: así la pantalla no
 * salta al llegar los datos y la carga se lee como "esto se está llenando",
 * no como "aquí no hay nada".
 */
const skeletonHtml = (viewport, text) => {
  const bars = [22, 34, 18, 28]
  const track =
    viewport === 'desktop'
      ? `<div class="skel__track">
           ${bars
             .map(
               (w, i) =>
                 `<span class="skel__slip" style="left:${8 + i * 24}%;width:${w * 0.42}%"></span>`,
             )
             .join('')}
         </div>`
      : ''
  return `
    <div class="skel" aria-hidden="true">
      <p class="skel__status">${spinnerHtml('sm')}<span>${esc(text)}</span></p>
      ${track}
      <ul class="skel__list">
        ${bars
          .map(
            () => `<li class="skel__row">
              <span class="skel__bar skel__bar--time"></span>
              <span class="skel__stack">
                <span class="skel__bar skel__bar--name"></span>
                <span class="skel__bar skel__bar--service"></span>
              </span>
              <span class="skel__bar skel__bar--badge"></span>
            </li>`,
          )
          .join('')}
      </ul>
    </div>`
}

const contentHtml = (screen, viewport) => {
  const out = []

  if (screen.state?.kind === 'loading') {
    // Con contexto resuelto se muestra el esqueleto de lo que está por
    // llegar; sin contexto todavía no hay geometría que anticipar.
    out.push(
      screen.state.skeleton
        ? skeletonHtml(viewport, screen.state.text)
        // El texto de espera se compone como titular en serif: es el único
        // mensaje de la pantalla y merece el peso del titular, no el de una
        // nota al pie.
        : pageStateHtml({ tone: 'loading', title: screen.state.text }),
    )
  }
  // Un día válido pero vacío sí tiene calendario: se dibuja el eje del día
  // antes del mensaje, para que la pantalla muestre que la jornada existe y
  // está abierta en vez de parecer una sección sin contenido.
  const emptyDay = Array.isArray(screen.entries) && screen.entries.length === 0
  if (emptyDay) {
    out.push(`<div class="agenda agenda--open">
      ${viewport === 'desktop' ? timelineHtml(screen) : ''}
    </div>`)
  }

  if (screen.state?.kind === 'note') {
    out.push(
      pageStateHtml({
        tone: 'note',
        title: screen.state.text,
        body: screen.state.body,
        action: screen.state.action,
      }),
    )
  }
  if (screen.state?.kind === 'alert') {
    out.push(
      pageStateHtml({
        tone: screen.state.variant,
        eyebrow: ALERT_EYEBROW[screen.state.variant],
        title: screen.state.title,
        body: screen.state.body,
        action: screen.state.action,
      }),
    )
  }
  if (screen.alert) out.push(alertHtml(screen.alert))
  if (screen.status) out.push(`<p class="state state--updating">${esc(screen.status)}</p>`)

  if (screen.entries && screen.entries.length > 0) {
    const stale = screen.stale ? ' agenda--stale' : ''
    out.push(`<div class="agenda${stale}">
      ${viewport === 'desktop' ? timelineHtml(screen) : ''}
      ${listHtml(screen.entries)}
    </div>`)
  }

  return out.join('\n')
}

// ---------------------------------------------------------------- estilos

const CSS = `
@font-face{font-family:'Instrument Serif';font-weight:400;font-display:block;src:url(data:font/woff2;base64,${SERIF}) format('woff2')}
@font-face{font-family:'Instrument Sans';font-weight:400 700;font-display:block;src:url(data:font/woff2;base64,${SANS}) format('woff2')}

:root{
  --ink:#101b2b; --ink-2:#16243a; --canvas:#f4f0e7; --surface:#fff;
  /* Superficie de las fichas de turno sobre tinta. El blanco puro daba
     ~17:1 contra #101b2b y deslumbraba en una pantalla oscura que se mira
     durante toda la jornada; este pergamino es el --color-surface-muted del
     sistema y conserva 13:1 con la tinta, muy por encima de AA. */
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

/* ---- barra de aplicación ---- */
.appbar{display:flex;align-items:center;justify-content:space-between;
  border-bottom:1px solid rgba(184,149,90,.45)}
.appbar__brand{display:flex;align-items:baseline;gap:14px}
.appbar__mark{font-family:var(--serif);color:var(--on-ink);letter-spacing:.02em;line-height:1}
.appbar__shop{color:var(--on-ink-2)}

/* ---- encabezado de página ---- */
.main{flex:1;display:flex;flex-direction:column}
.pagehead{display:flex;align-items:flex-start;justify-content:space-between;gap:20px}
.pagehead__title{font-family:var(--serif);color:var(--on-ink);line-height:1.05}
.pagehead__meta{color:var(--on-ink-2)}

/* ---- controles reglados sobre tinta ---- */
.controls{display:flex;align-items:flex-end;flex-wrap:wrap}
.field{display:flex;flex-direction:column}
.field__label{color:var(--brass);font-weight:600;text-transform:uppercase}
.control{display:flex;align-items:flex-end;justify-content:space-between;gap:12px;
  background:rgba(244,240,231,.05);border:1px solid rgba(244,240,231,.16);
  border-bottom:2px solid var(--brass);border-radius:2px}
.control__value{color:var(--on-ink);white-space:nowrap;line-height:1}
.control__value--ph{color:var(--on-ink-2)}
.control__caret{color:var(--brass);line-height:1}
.control--off{opacity:.42;border-bottom-color:rgba(244,240,231,.28)}
.datenav{display:flex;align-items:flex-end}
.field--barber{position:relative}

/* ---- retrato del barbero ---- */
.avatar{flex:none;display:flex;align-items:center;justify-content:center;
  background:var(--ink-2);border:1px solid rgba(184,149,90,.55);border-radius:2px;
  font-family:var(--serif);color:var(--brass);line-height:1;overflow:hidden;
  letter-spacing:.04em}
.avatar--photo{color:rgba(184,149,90,.75)}
.avatar--photo svg{width:78%;height:78%}

/* ---- lista de barberos desplegada ---- */
.options{position:absolute;left:0;right:0;top:100%;list-style:none;z-index:5;
  background:var(--ink-2);border:1px solid rgba(244,240,231,.16);
  border-top:2px solid var(--brass);border-radius:2px;
  box-shadow:0 16px 40px rgb(0 0 0 / 45%)}
.option{display:flex;align-items:center;color:var(--on-ink)}
.option + .option{border-top:1px solid rgba(244,240,231,.08)}
.option--selected{background:rgba(184,149,90,.12)}
.option--selected .option__name{font-weight:600}
.option__name{flex:1;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.control--open{border-bottom-color:var(--brass)}
/* Con retrato y cursor el control tiene tres hijos: sin esto, space-between
   dejaba el nombre flotando en el centro en vez de junto a su retrato. */
.control--select .control__value{flex:1;text-align:left}

/* ---- acciones ---- */
.btn{display:inline-flex;align-items:center;justify-content:center;border-radius:2px;
  font-weight:600;letter-spacing:.02em;white-space:nowrap}
.btn--secondary{background:var(--canvas);color:var(--ink)}
.btn--ghost{background:transparent;color:var(--brass);
  border:1px solid rgba(184,149,90,.5);border-bottom:2px solid var(--brass)}
.btn--ghost.control--off{opacity:.42}

/* ---- estados ---- */
.state{color:var(--on-ink-2)}
.state--updating{color:var(--brass);font-weight:600;text-transform:uppercase}

/* Anillo de latón interrumpido con el rombo NAVA al centro. */
.spinner{position:relative;display:inline-block;flex:none}
.spinner i{position:absolute;inset:0;border-radius:50%;
  border:2px solid rgba(184,149,90,.22);border-top-color:var(--brass);
  border-right-color:var(--brass);transform:rotate(-38deg)}
.spinner b{position:absolute;top:50%;left:50%;background:var(--brass);
  transform:translate(-50%,-50%) rotate(45deg)}

/* ---- estado de página (carga, error o vacío a pantalla completa) ---- */
.pagestate{display:flex;flex-direction:column;align-items:center;text-align:center;
  margin:auto;max-width:460px}
.pagestate__mark{display:flex;align-items:center;justify-content:center}
/* Regla de latón bajo la marca: el mismo remate del wordmark y del divisor,
   para que un estado vacío o de error tenga la misma factura editorial que
   el resto de la interfaz y no quede como una frase suelta. */
.pagestate__divider{display:flex;align-items:center;justify-content:center}
.pagestate__divider i{display:block;height:1px;background:var(--brass);opacity:.65}
.pagestate__divider b{display:block;background:var(--brass);transform:rotate(45deg);flex:none}
.pagestate__dest{color:var(--brass);font-weight:600}
/* Rombo del divisor NAVA como marca del estado: misma familia que el
   indicador de progreso, teñido por el tono del estado. */
.pagestate__mark .diamond{display:block;transform:rotate(45deg)}
.pagestate--warning .pagestate__divider b{background:var(--warning-b)}
.pagestate--danger .pagestate__divider b{background:var(--danger-b)}
.pagestate--info .pagestate__divider b{background:var(--info-b)}

.pagestate__eyebrow{font-weight:600;text-transform:uppercase}
.pagestate--warning .pagestate__eyebrow{color:var(--warning-b)}
.pagestate--danger .pagestate__eyebrow{color:var(--danger-b)}
.pagestate--info .pagestate__eyebrow{color:var(--info-b)}
.pagestate__title{font-family:var(--serif);color:var(--on-ink);line-height:1.15}
.pagestate__body{color:var(--on-ink-2)}
.pagestate--loading .pagestate__title,.pagestate--note .pagestate__title{
  font-size:inherit}

/* ---- esqueleto de la agenda ---- */
.skel{display:flex;flex-direction:column}
.skel__status{display:flex;align-items:center;color:var(--brass);font-weight:600;
  text-transform:uppercase}
.skel__track{position:relative;border-top:1px solid rgba(244,240,231,.14)}
.skel__slip{position:absolute;background:rgba(244,240,231,.08);
  border-left:3px solid rgba(184,149,90,.35);border-radius:2px}
.skel__list{list-style:none;display:flex;flex-direction:column}
.skel__row{display:flex;align-items:center;background:rgba(244,240,231,.05);
  border:1px solid rgba(244,240,231,.1);
  border-left:3px solid rgba(184,149,90,.35);border-radius:2px}
.skel__stack{flex:1;display:flex;flex-direction:column}
.skel__bar{display:block;background:rgba(244,240,231,.14);border-radius:2px}
.skel__bar--time{flex:none}
.skel__bar--service{background:rgba(244,240,231,.09)}
.skel__bar--badge{flex:none}

/* ---- alertas: misma nota al margen del atlas de autenticación ---- */
.alert{display:flex;border:1px solid;border-left-width:4px;border-radius:2px}
.alert__body{flex:1;display:flex;flex-direction:column}
.alert__eyebrow{font-weight:600;text-transform:uppercase;opacity:.85}
.alert__title{font-weight:600;color:var(--ink)}
.alert__action{align-self:flex-start;border:1px solid currentColor;border-bottom-width:2px;
  border-radius:2px;font-weight:600;text-transform:uppercase;letter-spacing:.08em}
.alert--danger{background:var(--danger-s);border-color:var(--danger-b);color:var(--danger-t)}
.alert--warning{background:var(--warning-s);border-color:var(--warning-b);color:var(--warning-t)}
.alert--info{background:var(--info-s);border-color:var(--info-b);color:var(--info-t)}

/* ---- línea temporal ---- */
.agenda{display:flex;flex-direction:column}
.agenda--stale{opacity:.5}
/* Día abierto: el carril se mantiene, más bajo, para mostrar que la jornada
   existe y está libre en vez de desaparecer con los turnos. */
.agenda--open .track{background:transparent}
.timeline{display:flex;flex-direction:column}
.axis{position:relative;height:18px}
.axis__tick{position:absolute;top:0;transform:translateX(-50%);display:flex;
  flex-direction:column;align-items:center}
.axis__tick em{font-style:normal;color:var(--on-ink-2);letter-spacing:.08em;
  font-variant-numeric:tabular-nums}
.marks{position:relative;height:22px}
.mark{position:absolute;top:0;transform:translateX(-50%);display:flex;align-items:center}
.mark em{font-style:normal;font-weight:600;text-transform:uppercase;letter-spacing:.12em;
  border-radius:2px;white-space:nowrap}
.mark--now em{background:var(--brass);color:var(--ink)}
.mark--day em{background:transparent;color:var(--brass);border:1px solid var(--brass)}
/* El carril es una región definida, no fichas flotando en el vacío. */
.track{position:relative;background:rgba(244,240,231,.035);
  border:1px solid rgba(244,240,231,.1);border-radius:2px}
.track__rule{position:absolute;top:0;bottom:0;width:1px;
  background:rgba(244,240,231,.12)}
.track__rule--half{background:rgba(244,240,231,.05)}
.track__mark{position:absolute;top:0;bottom:0;width:1.5px}
.track__mark--now{background:var(--brass)}
.track__mark--day{background:repeating-linear-gradient(to bottom,var(--brass) 0 6px,transparent 6px 12px)}
.slip{position:absolute;top:0;display:flex;flex-direction:column;justify-content:center;
  overflow:hidden;background:var(--paper);color:var(--ink);border-radius:2px;
  border-left:3px solid var(--brass-deep)}
.slip em{font-style:normal;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
  display:block}
.slip__time{color:var(--text-2);font-variant-numeric:tabular-nums;letter-spacing:.03em}
.slip__name{font-weight:600}
.slip__service{color:var(--text-2)}
/* Terminal: se atenúa por opacidad sobre la misma superficie clara. Un
   fondo gris propio lo haría pesar MÁS que un turno vigente sobre tinta,
   que es justo lo contrario de lo que pide el contrato. */
.slip--terminal{background:transparent;border:1px solid rgba(244,240,231,.2);
  border-left:3px solid rgba(244,240,231,.3);color:var(--on-ink-2)}
.slip--terminal .slip__time{color:var(--on-ink-2)}
.slip--terminal .slip__name{color:var(--on-ink);font-weight:500}

/* ---- lista cronológica accesible ---- */
.list{list-style:none;display:flex;flex-direction:column}
.row{display:flex;align-items:center;background:var(--paper);color:var(--text);
  border-radius:2px;border-left:3px solid var(--brass-deep)}
/* Terminal: deja de ser papel y pasa a contorno sobre la tinta. Un turno
   vigente es una superficie presente; uno cerrado es un registro. Atenuar
   solo con un relleno gris lo dejaba pesando igual que el vigente. */
.row--terminal{background:transparent;border:1px solid rgba(244,240,231,.2);
  border-left:3px solid rgba(244,240,231,.3);color:var(--on-ink-2)}
.row--terminal .row__name{color:var(--on-ink);font-weight:500}
.row--terminal .row__time,.row--terminal .row__service{color:var(--on-ink-2)}
.row--terminal .badge{background:transparent;border-color:rgba(244,240,231,.32);
  color:var(--on-ink-2)}
.row__time{font-variant-numeric:tabular-nums;color:var(--text-2);flex:none}
.row__body{flex:1;display:flex;flex-direction:column;min-width:0}
.row__name{font-weight:600;color:var(--ink)}
.row__service{color:var(--text-2)}
.row__note{color:var(--brass-deep);font-weight:600}
.badge{flex:none;border:1px solid;border-radius:2px;font-weight:600;white-space:nowrap}
.badge--confirmed{background:rgba(35,64,91,.08);border-color:var(--info-b);color:var(--info-t)}
.badge--completed{background:rgba(50,93,67,.08);border-color:var(--success-b);color:var(--success-t)}
.badge--cancelled_by_customer{background:transparent;border-color:var(--inactive-b);color:var(--inactive-t)}
.badge--cancelled_by_barber{background:rgba(138,44,44,.08);border-color:var(--danger-b);color:var(--danger-t)}
.badge--no_show{background:rgba(119,80,25,.08);border-color:var(--warning-b);color:var(--warning-t)}

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
    .main{padding:26px 40px 30px;gap:20px}
    .pagehead__title{font-size:40px}
    .pagehead__meta{font-size:15px;margin-top:6px}
    .controls{gap:28px}
    .field{gap:8px}
    .field__label{font-size:11px;line-height:15px;letter-spacing:.14em}
    .control{height:48px;padding:0 14px 11px;min-width:240px}
    .control--select{align-items:center;padding-bottom:0;gap:12px}
    .avatar--sm{width:28px;height:28px;font-size:13px}
    .avatar--md{width:34px;height:34px;font-size:16px}
    .options{margin-top:4px}
    .option{gap:12px;padding:10px 14px;font-size:16px}
    .control__value{font-size:16px}
    .control--select{min-width:280px}
    .datenav{gap:12px}
    .datenav .control{min-width:150px}
    .btn{height:48px;padding:0 22px;font-size:15px}
    .state{font-size:16px}
    .state--updating{font-size:12px;letter-spacing:.14em}
    .spinner--lg{width:54px;height:54px}
    .spinner--lg b{width:10px;height:10px}
    .spinner--sm{width:18px;height:18px}
    .spinner--sm b{width:5px;height:5px}
    .pagestate{gap:14px;padding:40px 0}
    .pagestate__divider{gap:12px;margin:2px 0 4px}
    .pagestate__divider i{width:98px}
    .pagestate__divider b{width:7px;height:7px}
    .pagestate__mark .diamond{width:13px;height:13px}
    .pagestate__eyebrow{font-size:11px;line-height:15px;letter-spacing:.16em}
    .pagestate__title{font-size:30px;margin-top:-2px}
    .pagestate--loading,.pagestate--note{font-size:26px}
    .pagestate--loading .pagestate__title{line-height:34px}
    .pagestate--note .pagestate__title{line-height:34px;max-width:440px}
    .pagestate__body{font-size:16px;line-height:24px;max-width:452px}
    .pagestate__action{margin-top:10px}
    .skel{gap:18px}
    .skel__status{gap:12px;font-size:12px;letter-spacing:.14em}
    .skel__track{height:96px}
    .skel__slip{top:10px;height:76px}
    .skel__list{gap:10px}
    .skel__row{gap:20px;padding:13px 18px;height:64px}
    .skel__bar--time{width:96px;height:12px}
    .skel__stack{gap:8px}
    .skel__bar--name{width:180px;height:13px}
    .skel__bar--service{width:110px;height:10px}
    .skel__bar--badge{width:96px;height:22px}
    .alert{padding:15px 18px;max-width:640px}
    .alert__body{gap:5px}
    .alert__eyebrow{font-size:11px;line-height:15px;letter-spacing:.15em}
    .alert__title{font-size:17px;line-height:24px}
    .alert__text{font-size:15px;line-height:22px}
    .alert__action{margin-top:8px;padding:9px 16px;font-size:12px}
    .agenda{gap:20px}
    .timeline{gap:0}
    .axis__tick em{font-size:12px}
    .mark em{font-size:10px;padding:4px 9px}
    .track{height:96px;margin-top:2px}
    .agenda--open .track{height:72px}
    .slip{height:76px;top:10px;padding:0 12px;gap:2px}
    .slip__service{font-size:12px}
    .slip__time{font-size:11px}
    .slip__name{font-size:13px}
    .list{gap:10px}
    .row{gap:20px;padding:13px 18px}
    .row__time{font-size:15px;width:118px}
    .row__name{font-size:17px}
    .row__service{font-size:14px;margin-top:2px}
    .row__note{font-size:13px;margin-top:4px}
    .badge{padding:5px 12px;font-size:12px}
    .dock{padding:0 40px}
    .dock__item{gap:10px;height:64px;font-size:14px}
    .dock__icon svg{width:19px;height:19px}
  `,
  mobile: `
    .appbar{padding:16px 20px}
    .appbar__mark{font-size:26px}
    .appbar__shop{font-size:13px}
    .main{padding:20px;gap:20px}
    .pagehead{flex-direction:column;align-items:stretch;gap:14px}
    .pagehead__title{font-size:32px}
    .pagehead__meta{font-size:14px;margin-top:5px}
    .btn{height:46px;padding:0 18px;font-size:15px}
    .pagehead .btn{width:100%}
    .controls{flex-direction:column;align-items:stretch;gap:16px}
    .field{gap:7px}
    .field__label{font-size:10px;line-height:14px;letter-spacing:.14em}
    .control{height:46px;padding:0 13px 10px}
    .control--select{align-items:center;padding-bottom:0;gap:10px}
    .avatar--sm{width:26px;height:26px;font-size:12px}
    .avatar--md{width:32px;height:32px;font-size:15px}
    .options{margin-top:4px}
    .option{gap:11px;padding:10px 13px;font-size:15px}
    .control__value{font-size:15px}
    .datenav{gap:10px;align-items:flex-end}
    .datenav .field{flex:1}
    .datenav .control{min-width:0}
    .datenav .btn{padding:0 14px;font-size:14px}
    .state{font-size:15px}
    .state--updating{font-size:11px;letter-spacing:.14em}
    .spinner--lg{width:46px;height:46px}
    .spinner--lg b{width:9px;height:9px}
    .spinner--sm{width:16px;height:16px}
    .spinner--sm b{width:5px;height:5px}
    /* Se estrecha para que el titular en serif quiebre en dos líneas en vez
       de rozar los bordes de la pantalla. */
    .pagestate{gap:12px;padding:34px 0;max-width:320px}
    .pagestate__divider{gap:10px;margin:2px 0 4px}
    .pagestate__divider i{width:78px}
    .pagestate__divider b{width:6px;height:6px}
    .pagestate__mark .diamond{width:12px;height:12px}
    .pagestate__eyebrow{font-size:10px;line-height:14px;letter-spacing:.16em}
    .pagestate__title{font-size:25px;margin-top:-2px}
    .pagestate--loading,.pagestate--note{font-size:22px}
    .pagestate--loading .pagestate__title{line-height:29px}
    .pagestate--note .pagestate__title{line-height:29px}
    .pagestate__body{font-size:15px;line-height:22px}
    .pagestate__action{margin-top:8px}
    .skel{gap:16px}
    .skel__status{gap:10px;font-size:11px;letter-spacing:.14em}
    .skel__list{gap:10px}
    .skel__row{gap:14px;padding:13px 15px;height:60px}
    .skel__bar--time{width:74px;height:11px}
    .skel__stack{gap:7px}
    .skel__bar--name{width:130px;height:12px}
    .skel__bar--service{width:88px;height:9px}
    .skel__bar--badge{width:76px;height:20px}
    .alert{padding:13px 15px}
    .alert__body{gap:4px}
    .alert__eyebrow{font-size:10px;line-height:14px;letter-spacing:.15em}
    .alert__title{font-size:16px;line-height:23px}
    .alert__text{font-size:14px;line-height:20px}
    .alert__action{margin-top:8px;padding:9px 14px;font-size:11px}
    .agenda{gap:0}
    .list{gap:10px}
    /* En móvil la insignia baja a su propia línea: compartir renglón con
       el nombre obligaba a partir "Daniel López" en dos. */
    .row{flex-direction:column;align-items:flex-start;gap:6px;padding:13px 15px}
    .row__time{font-size:13px;width:auto;letter-spacing:.02em}
    .row__name{font-size:16px}
    .row__service{font-size:13px;margin-top:2px}
    .row__note{font-size:12px;margin-top:4px}
    .badge{padding:4px 9px;font-size:11px;margin-top:4px}
    .dock{padding:0 8px}
    .dock__item{flex-direction:column;gap:5px;height:66px;font-size:11px}
    .dock__icon svg{width:19px;height:19px}
  `,
}

const page = (viewport, screen) => `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><style>${CSS}${SCALES[viewport]}</style></head>
<body>
  <div class="shell">
    ${headerHtml(screen.shell)}
    <main class="main">
      ${pageHeadHtml(screen.shell)}
      ${controlsHtml(screen.shell)}
      ${contentHtml(screen, viewport)}
    </main>
    ${navHtml(viewport)}
  </div>
</body></html>`

const VIEWPORTS = {
  desktop: { width: 1440, height: 1024, deviceScaleFactor: 1 },
  mobile: { width: 420, height: 935, deviceScaleFactor: 2 },
}

async function main() {
  const browser = await chromium.launch()
  const problems = []

  for (const [viewport, cfg] of Object.entries(VIEWPORTS)) {
    const ctx = await browser.newContext({
      viewport: { width: cfg.width, height: cfg.height },
      deviceScaleFactor: cfg.deviceScaleFactor,
    })
    const tab = await ctx.newPage()

    for (const [name, screen] of Object.entries(PANEL.screens)) {
      await tab.setContent(page(viewport, screen), { waitUntil: 'load' })
      await tab.evaluate(() => document.fonts.ready)

      const box = await tab.evaluate(() => ({
        h: document.documentElement.scrollHeight,
        w: document.documentElement.scrollWidth,
      }))
      if (box.w > cfg.width) {
        problems.push(`${viewport}/${name}: desborde horizontal ${box.w}px`)
      }
      if (viewport === 'desktop' && box.h > cfg.height) {
        problems.push(`${viewport}/${name}: desborde vertical ${box.h}px`)
      }

      // Ninguna marca del eje puede tapar la etiqueta de una hora: fue el
      // defecto principal de la revisión anterior, así que se comprueba aquí.
      const collision = await tab.evaluate(() => {
        const marks = [...document.querySelectorAll('.mark em')]
        const ticks = [...document.querySelectorAll('.axis__tick em')]
        for (const m of marks) {
          const a = m.getBoundingClientRect()
          for (const t of ticks) {
            const b = t.getBoundingClientRect()
            if (a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom) {
              return `${m.textContent} tapa ${t.textContent}`
            }
          }
        }
        return null
      })
      if (collision) problems.push(`${viewport}/${name}: ${collision}`)

      const dir = resolve(OUT, viewport, 'panel')
      mkdirSync(dir, { recursive: true })
      await tab.screenshot({
        path: resolve(dir, `${name}.png`),
        fullPage: viewport === 'mobile',
      })
    }
    await ctx.close()
  }

  await browser.close()

  if (problems.length) {
    console.error('Problemas de composición:\n- ' + problems.join('\n- '))
    process.exitCode = 1
  } else {
    console.log(`${Object.keys(PANEL.screens).length * 2} mockups regenerados sin desbordes ni marcas superpuestas.`)
  }
}

await main()
