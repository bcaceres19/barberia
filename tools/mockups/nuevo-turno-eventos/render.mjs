/*
 * Generador determinista de los mockups de `/panel/turnos/nuevo` (Nuevo
 * turno).
 *
 * Comparte criterio y lenguaje visual con `tools/mockups/panel-agenda-eventos`
 * y `tools/mockups/auth-eventos`: HTML compuesto con los tokens reales del
 * sistema y las fuentes auto-hosteadas del proyecto, rasterizado con el
 * Chromium de Playwright que ya usa la suite e2e. El encabezado y el dock
 * reutilizan exactamente el cascarón de `panel-agenda-eventos`.
 *
 * Revisión 2026-09-04 (pulido contra los atlas hermanos). Lo que cambia y
 * por qué:
 *
 *  - El formulario deja de ser una pila de tarjetas de pergamino con campos
 *    blancos. Sobre el canvas de tinta, el sistema ya tiene una forma
 *    resuelta de control (`panel-agenda-eventos`: superficie translúcida,
 *    filete de latón inferior y rótulo en versalitas de latón) y el
 *    pergamino está reservado para REGISTROS (fichas y filas de turno), no
 *    para el cromo de un formulario. Los campos blancos sobre pergamino
 *    sobre tinta apilaban tres materiales para una sola tarea.
 *  - Cada sección pasa a componerse en dos columnas (rótulo numerado a la
 *    izquierda, campos a la derecha): el formulario cabía en 900 px
 *    centrados y dejaba media pantalla vacía, con el encabezado
 *    desalineado del de `/panel`. Ahora el encabezado comparte gutter con
 *    el resto del panel y el alto de escritorio entra en un viewport.
 *  - La espera deja de ser un spinner solo en una pantalla vacía y pasa a
 *    esqueleto con la geometría del formulario por llegar, igual que el
 *    evento `05` de la agenda.
 *  - El resumen usa el título y las cuatro entradas reales del `<dl>` del
 *    código (`Resumen`; Barbero, Servicio, Persona atendida, Fecha y hora)
 *    y aparece siempre que `hasSummaryContent` sea cierto, no solo con el
 *    formulario completo. En escritorio vive en una columna lateral; en
 *    móvil, antes del CTA (§7.3).
 *  - El éxito se compone como pantalla completa (divisor NAVA, titular en
 *    serif y ficha de pergamino con el turno) en vez de una alerta pequeña
 *    flotando en el vacío.
 *
 * Uso:
 *   node tools/mockups/nuevo-turno-eventos/render.mjs
 */
import playwright from '../../../apps/web/node_modules/@playwright/test/index.js'
import { readFileSync, mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { COPY, SCREENS, NAV_PRIMARY, NAV_SECONDARY, SHOP_NAME } from './content.mjs'

const { chromium } = playwright

const HERE = dirname(fileURLToPath(import.meta.url))
const REPO = resolve(HERE, '../../..')
const OUT = resolve(
  REPO,
  'docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/nuevo-turno-eventos',
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
}
const navIcon = (n) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">${NAV_ICONS[n]}</svg>`

const ALERT_EYEBROW = { danger: 'Error', warning: 'Atención', info: 'Nota', success: 'Confirmación' }

/*
 * Retrato del barbero, tomado tal cual de `panel-agenda-eventos`: dondequiera
 * que un barbero se ve o se elige, su nombre va acompañado de un monograma
 * cuadrado con filete de latón. Aquí aparece en el selector cerrado y en la
 * entrada «Barbero» del resumen.
 */
const initials = (name) =>
  name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0])
    .join('')
    .toUpperCase()

const avatarHtml = (name, size = 'sm') =>
  `<span class="avatar avatar--${size}" aria-hidden="true">${esc(initials(name))}</span>`

const alertHtml = (a) => `
  <div class="alert alert--${a.variant}" role="alert">
    <div class="alert__body">
      <p class="alert__eyebrow">${esc(ALERT_EYEBROW[a.variant])}</p>
      <p class="alert__title">${esc(a.title)}</p>
      ${a.body ? `<p class="alert__text">${esc(a.body)}</p>` : ''}
    </div>
  </div>`

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

const spinnerHtml = (size = 'lg') => `<span class="spinner spinner--${size}"><i></i><b></b></span>`

// Un destino del producto nombrado dentro de un mensaje se resalta en latón,
// igual que «Barberos» en el evento 04 de la agenda. El código lo escribe
// entre comillas rectas, así que se aceptan ambas formas sin tocar el copy.
const highlightDest = (t) =>
  t
    .replace(/«([^»]+)»/g, '<span class="pagestate__dest">«$1»</span>')
    .replace(/"([^"]+)"/g, '<span class="pagestate__dest">"$1"</span>')

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

// ------------------------------------------------------------- campos

const reqMark = '<span class="field__req" aria-hidden="true">*</span>'

const NOTE_TONE = { loading: 'loading', empty: 'warning', error: 'danger' }

/*
 * Control de formulario sobre tinta: misma familia que los controles de la
 * agenda (superficie translúcida, filete inferior de latón, rótulo en
 * versalitas). Un campo con error cambia el filete al rojo del sistema y
 * agrega el mensaje debajo; uno deshabilitado se atenúa y pierde el latón.
 */
const fieldHtml = ({
  label,
  required,
  value,
  placeholder,
  error,
  note,
  noteTone,
  disabled,
  avatar,
  type = 'text',
}) => `
  <div class="field${error ? ' field--error' : ''}">
    <span class="field__label">${esc(label)}${required ? reqMark : ''}</span>
    <div class="control${value && !disabled ? ' control--filled' : ''}${disabled ? ' control--off' : ''}${type === 'select' ? ' control--select' : ''}">
      ${avatar && value ? avatarHtml(value) : ''}
      <span class="control__value${value ? '' : ' control__value--ph'}">${esc(value || placeholder)}</span>
      ${type === 'select' ? '<span class="control__caret">▾</span>' : ''}
    </div>
    ${
      note
        ? `<p class="field__note field__note--${noteTone ?? 'muted'}">
             ${noteTone === 'loading' ? spinnerHtml('xs') : ''}<span>${esc(note)}</span>
           </p>`
        : ''
    }
    ${error ? `<p class="field__error" role="alert">${esc(error)}</p>` : ''}
  </div>`

const noteFieldHtml = ({ label, value }) => `
  <div class="field">
    <span class="field__label">${esc(label)}</span>
    <div class="control control--textarea${value ? ' control--filled' : ''}">
      <span class="control__value${value ? '' : ' control__value--ph'}">${esc(value)}</span>
    </div>
    <div class="field__foot">
      <span></span>
      <span class="field__counter">${value.length}/${COPY.noteMax}</span>
    </div>
  </div>`

const sectionHtml = (n, title, hint, body) => `
  <section class="sec">
    <div class="sec__head">
      <span class="sec__num" aria-hidden="true">${n}</span>
      <div class="sec__stack">
        <h2 class="sec__title">${esc(title)}</h2>
        <p class="sec__hint">${esc(hint)}</p>
      </div>
    </div>
    <div class="sec__body">${body}</div>
  </section>`

// ------------------------------------------------------------- resumen

/*
 * Resumen del turno. Refleja literalmente el `<dl>` del código: aparece en
 * cuanto hay UNA selección hecha (`hasSummaryContent`) y cada entrada se
 * dibuja solo si su dato existe. No se agregan teléfono, correo ni zona
 * horaria: el resumen real no los lista.
 */
const hasSummary = (f) => !!(f.barber || f.service || f.attendee || f.dateFull || f.time)

const resumenItem = (label, value, avatar = false) => `
  <div class="resumen__item">
    <dt class="resumen__label">${esc(label)}</dt>
    <dd class="resumen__value">${avatar ? avatarHtml(value) : ''}<span>${esc(value)}</span></dd>
  </div>`

const scheduleValue = (f) => [f.dateFull, f.time].filter(Boolean).join(' · ')

/*
 * Columna lateral. Lleva la alerta global y el resumen, en ese orden: en
 * escritorio quedan junto al CTA sin empujar el formulario hacia abajo, y en
 * móvil —donde la retícula colapsa a una columna— caen exactamente entre el
 * último campo y el CTA, que es donde §7.3 pide el resumen.
 */
const railHtml = (f, alert) => `
  <aside class="rail">
    ${alert ? alertHtml(alert) : ''}
    ${
      hasSummary(f)
        ? `<div class="resumen">
             <p class="resumen__title">${esc(COPY.resumenTitle)}</p>
             <dl class="resumen__list">
               ${f.barber ? resumenItem(COPY.resumenBarber, f.barber, true) : ''}
               ${f.service ? resumenItem(COPY.resumenService, f.service) : ''}
               ${f.attendee ? resumenItem(COPY.resumenAttendee, f.attendee) : ''}
               ${f.dateFull || f.time ? resumenItem(COPY.resumenSchedule, scheduleValue(f)) : ''}
             </dl>
           </div>`
        : ''
    }
  </aside>`

// ----------------------------------------------------------- contenido

const formHtml = (screen) => {
  const f = screen.fields
  const fe = screen.fieldErrors ?? {}
  const attempted = !!screen.attempted
  const rail = hasSummary(f) || !!screen.alert

  const barberField = fieldHtml({
    label: COPY.barberLabel,
    type: 'select',
    value: f.barber,
    placeholder: COPY.barberPlaceholder,
    avatar: true,
    error: attempted ? fe.barber : undefined,
  })
  const serviceNote =
    screen.servicesStatus === 'loading'
      ? COPY.servicesLoading
      : screen.servicesStatus === 'empty'
        ? COPY.servicesEmpty
        : screen.servicesStatus === 'error'
          ? COPY.servicesError
          : undefined
  const serviceField = fieldHtml({
    label: COPY.serviceLabel,
    type: 'select',
    value: f.service,
    placeholder: f.barber ? COPY.servicePlaceholder : COPY.servicePlaceholderNoBarber,
    disabled: !f.barber || screen.servicesStatus === 'loading',
    note: serviceNote,
    noteTone: NOTE_TONE[screen.servicesStatus],
    error: attempted ? fe.service : undefined,
  })

  const section1 = sectionHtml(
    1,
    COPY.section1Title,
    COPY.section1Hint,
    `<div class="row">${barberField}${serviceField}</div>`,
  )

  const section2 = sectionHtml(
    2,
    COPY.section2Title,
    COPY.section2Hint,
    `<div class="row">
       ${fieldHtml({ label: COPY.attendeeLabel, required: true, value: f.attendee, placeholder: COPY.attendeeLabel, error: attempted ? fe.attendee : undefined })}
       ${fieldHtml({ label: COPY.customerNameLabel, required: true, value: f.customerName, placeholder: COPY.customerNameLabel, error: attempted ? fe.customerName : undefined })}
     </div>
     <div class="row">
       ${fieldHtml({ label: COPY.phoneLabel, value: f.phone, placeholder: COPY.phonePlaceholder, error: attempted ? fe.phone : undefined })}
       ${fieldHtml({ label: COPY.emailLabel, value: f.email, placeholder: COPY.emailLabel, error: attempted ? fe.email : undefined })}
     </div>
     ${!f.phone && !f.email ? `<p class="sec__note">${esc(COPY.noContactHint)}</p>` : ''}`,
  )

  const section3 = sectionHtml(
    3,
    COPY.section3Title,
    COPY.section3Hint,
    `<div class="row">
       ${fieldHtml({ label: COPY.dateLabel, required: true, value: f.date, placeholder: 'dd/mm/aaaa' })}
       ${fieldHtml({ label: COPY.timeLabel, required: true, value: f.time, placeholder: '--:--', error: attempted ? fe.time : undefined })}
     </div>`,
  )

  const section4 = sectionHtml(
    4,
    COPY.section4Title,
    COPY.section4Hint,
    noteFieldHtml({ label: COPY.noteLabel, value: f.note }),
  )

  return `
    <div class="layout${rail ? '' : ' layout--solo'}">
      <div class="form"><div class="sheet">${section1}${section2}${section3}${section4}</div></div>
      ${rail ? railHtml(f, screen.alert) : ''}
      <div class="actions">
        <span class="btn btn--primary${screen.saving ? ' btn--disabled' : ''}">
          ${esc(screen.saving ? COPY.submitting : COPY.submit)}
        </span>
      </div>
    </div>`
}

/*
 * Éxito. La confirmación es la pantalla completa, no una nota al margen: el
 * turno recién creado se muestra como ficha de pergamino —el mismo material
 * con el que la agenda dibuja un turno vigente— bajo el divisor y el titular
 * en serif del sistema.
 */
const successHtml = (screen) => `
  <div class="pagestate pagestate--success success">
    <span class="pagestate__divider"><i></i><b></b><i></i></span>
    <p class="pagestate__eyebrow">${esc(ALERT_EYEBROW.success)}</p>
    <p class="pagestate__title">${esc(COPY.successTitle)}</p>
    <p class="ticket">${esc(COPY.successSummary(screen.created))}</p>
    <p class="pagestate__body">${esc(COPY.successNote)}</p>
    <span class="btn btn--primary pagestate__action">${esc(COPY.successAgain)}</span>
  </div>`

/*
 * Esqueleto del formulario. Con el contexto todavía sin resolver, la espera
 * conserva la geometría de lo que va a llegar —cuatro secciones numeradas y
 * sus campos— en vez de dejar la pantalla vacía con un indicador suelto.
 */
const skeletonHtml = () => {
  const bars = (n) =>
    Array.from({ length: n })
      .map(
        () => `<span class="skel__field">
                 <span class="skel__bar skel__bar--label"></span>
                 <span class="skel__bar skel__bar--control"></span>
               </span>`,
      )
      .join('')
  const sec = (rows) => `
    <div class="sec sec--skel">
      <div class="sec__head">
        <span class="skel__bar skel__bar--num"></span>
        <span class="sec__stack">
          <span class="skel__bar skel__bar--title"></span>
          <span class="skel__bar skel__bar--hint"></span>
        </span>
      </div>
      <div class="sec__body">
        ${rows.map((n) => `<div class="row">${bars(n)}</div>`).join('')}
      </div>
    </div>`

  return `
    <div class="skel" aria-hidden="true">
      <p class="skel__status">${spinnerHtml('sm')}<span>${esc(COPY.loadingBarbers)}</span></p>
      <div class="layout layout--solo">
        <div class="form">
          <div class="sheet">
            ${sec([2])}
            ${sec([2, 2])}
            ${sec([2])}
            ${sec([1])}
          </div>
        </div>
        <div class="actions"><span class="skel__bar skel__bar--cta"></span></div>
      </div>
    </div>`
}

const contentHtml = (screen) => {
  if (screen.page === 'loading') return skeletonHtml()
  if (screen.page === 'load-error') {
    return pageStateHtml({
      tone: 'warning',
      eyebrow: ALERT_EYEBROW.warning,
      title: COPY.loadErrorTitle,
      body: COPY.loadErrorBody,
      action: COPY.retry,
    })
  }
  if (screen.page === 'empty') {
    return pageStateHtml({
      tone: 'note',
      title: COPY.noBarbersTitle,
      body: COPY.noBarbersBody,
    })
  }
  if (screen.page === 'success') return successHtml(screen)
  return formHtml(screen)
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
  /* Tintes de estado LEVANTADOS al canvas de tinta. Los pares del sistema
     están calculados para superficies claras: #a43a3a sobre #101b2b no
     alcanza AA. Estos tres son el mismo matiz aclarado hasta pasar AA sobre
     tinta, y solo se usan para texto y filetes sobre tinta; las alertas y
     fichas siguen usando los pares claros del sistema. */
  --danger-ink:#e3928d; --warning-ink:#dfb063; --success-ink:#9dc2a9;
  --serif:'Instrument Serif',Georgia,serif;
  --sans:'Instrument Sans',system-ui,sans-serif;
}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:var(--sans);background:var(--ink);color:var(--on-ink);-webkit-font-smoothing:antialiased}
.shell{display:flex;flex-direction:column;min-height:100vh}

/* ---- barra de aplicación (idéntica a panel-agenda-eventos) ---- */
.appbar{display:flex;align-items:center;justify-content:space-between;
  border-bottom:1px solid rgba(184,149,90,.45)}
.appbar__brand{display:flex;align-items:baseline;gap:14px}
.appbar__mark{font-family:var(--serif);color:var(--on-ink);letter-spacing:.02em;line-height:1}
.appbar__shop{color:var(--on-ink-2)}

/* ---- encabezado de página ---- */
.main{flex:1;display:flex;flex-direction:column}
.pagehead{display:flex;flex-direction:column}
.pagehead__title{font-family:var(--serif);color:var(--on-ink);line-height:1.05}
.pagehead__meta{color:var(--on-ink-2)}

/* ---- retícula: formulario + resumen ---- */
.layout{display:grid;align-items:start}
.form{grid-area:form;display:flex;flex-direction:column}
.rail{grid-area:rail;display:flex;flex-direction:column}
.actions{grid-area:acts;display:flex;flex-direction:column}

/* ---- hoja del formulario ----
   Las cuatro secciones son bandas de UNA sola hoja separadas por filete, no
   cuatro tarjetas apiladas: con borde propio cada una, el formulario se leía
   como una pila rayada de objetos sueltos en vez de un documento. */
.sheet{background:rgba(244,240,231,.035);border:1px solid rgba(244,240,231,.1);
  border-left:3px solid rgba(184,149,90,.55);border-radius:2px}
.sec + .sec{border-top:1px solid rgba(244,240,231,.1)}
.sec__head{display:flex;align-items:flex-start}
.sec__num{flex:none;display:flex;align-items:center;justify-content:center;
  border:1px solid rgba(184,149,90,.6);border-radius:2px;color:var(--brass);
  font-family:var(--serif);line-height:1}
.sec__stack{display:flex;flex-direction:column;min-width:0}
.sec__title{font-family:var(--serif);color:var(--on-ink);line-height:1.15;font-weight:400}
/* Equilibra la última línea: el único subtítulo que no cabe en un renglón
   partía dejando «necesario.» solo. */
.sec__hint{color:var(--on-ink-2);text-wrap:balance}
.sec__body{display:flex;flex-direction:column}
.sec__note{color:var(--on-ink-2)}

/* ---- campos sobre tinta (misma familia que los controles de la agenda) ---- */
.row{display:flex}
.field{display:flex;flex-direction:column;flex:1;min-width:0}
.field__label{color:var(--brass);font-weight:600;text-transform:uppercase}
.field__req{color:var(--brass);opacity:.8}
/* El filete inferior de latón marca el campo YA RESUELTO; uno vacío lleva
   filete neutro. Ocho subrayados dorados a la vez convertían el formulario
   en un muestrario de oro y ya no distinguían lo hecho de lo pendiente. */
.control{display:flex;align-items:center;justify-content:space-between;
  background:rgba(244,240,231,.04);border:1px solid rgba(244,240,231,.12);
  border-bottom:2px solid rgba(244,240,231,.3);border-radius:2px}
.control--filled{background:rgba(244,240,231,.06);border-bottom-color:var(--brass)}
.control__value{flex:1;color:var(--on-ink);line-height:1.3;overflow:hidden;
  text-overflow:ellipsis;white-space:nowrap}
.control__value--ph{color:var(--on-ink-2)}
.control__caret{color:var(--brass);line-height:1;flex:none}
.control--off{opacity:.45;border-bottom-color:rgba(244,240,231,.2)}
.control--textarea{align-items:flex-start}
.control--textarea .control__value{white-space:normal}
.field--error .control{border-bottom-color:var(--danger-ink);
  background:rgba(227,146,141,.07)}
.field__error{color:var(--danger-ink);font-weight:600}
.field__note{display:flex;align-items:center;color:var(--on-ink-2)}
.field__note--loading{color:var(--brass);font-weight:600;text-transform:uppercase}
.field__note--warning{color:var(--warning-ink)}
.field__note--danger{color:var(--danger-ink)}
.field__foot{display:flex;align-items:center;justify-content:space-between}
.field__counter{color:var(--on-ink-2);font-variant-numeric:tabular-nums}

/* ---- retrato del barbero (mismo componente que la agenda) ---- */
.avatar{flex:none;display:flex;align-items:center;justify-content:center;
  background:var(--ink-2);border:1px solid rgba(184,149,90,.55);border-radius:2px;
  font-family:var(--serif);color:var(--brass);line-height:1;letter-spacing:.04em}

/* ---- resumen antes del CTA ---- */
.resumen{background:var(--ink-2);border:1px solid rgba(244,240,231,.12);
  border-top:2px solid var(--brass);border-radius:2px}
.resumen__title{color:var(--brass);font-weight:600;text-transform:uppercase}
.resumen__list{display:flex;flex-direction:column}
.resumen__item{display:flex;flex-direction:column}
.resumen__item + .resumen__item{border-top:1px solid rgba(244,240,231,.1)}
.resumen__label{color:var(--on-ink-2);text-transform:uppercase;font-weight:600}
.resumen__value{display:flex;align-items:center;color:var(--on-ink);font-weight:600}

/* ---- acciones ---- */
.btn{display:inline-flex;align-items:center;justify-content:center;border-radius:2px;
  font-weight:600;letter-spacing:.02em;white-space:nowrap}
.btn--primary{background:var(--brass);color:var(--ink)}
.btn--secondary{background:var(--canvas);color:var(--ink)}
.btn--disabled{opacity:.55}

/* ---- estados de página ---- */
.pagestate{display:flex;flex-direction:column;align-items:center;text-align:center;
  margin:auto;max-width:520px}
.pagestate__mark{display:flex;align-items:center;justify-content:center}
.pagestate__divider{display:flex;align-items:center;justify-content:center}
.pagestate__divider i{display:block;height:1px;background:var(--brass);opacity:.65}
.pagestate__divider b{display:block;background:var(--brass);transform:rotate(45deg);flex:none}
.pagestate--warning .pagestate__divider b{background:var(--warning-ink)}
.pagestate--success .pagestate__divider b{background:var(--success-ink)}
.pagestate__eyebrow{font-weight:600;text-transform:uppercase;color:var(--brass)}
.pagestate--warning .pagestate__eyebrow{color:var(--warning-ink)}
.pagestate--success .pagestate__eyebrow{color:var(--success-ink)}
.pagestate__title{font-family:var(--serif);color:var(--on-ink);line-height:1.15}
.pagestate__body{color:var(--on-ink-2)}
.pagestate__dest{color:var(--brass);font-weight:600}
/* La espera y el vacío se componen como titular; un párrafo largo no se
   compone al tamaño de un titular de error. */
.pagestate--note .pagestate__title{font-family:var(--serif)}

/* Ficha del turno recién creado: pergamino, el mismo material con el que la
   agenda dibuja un turno vigente. */
.ticket{background:var(--paper);color:var(--text);border-radius:2px;
  border-left:3px solid var(--brass-deep);font-weight:600;
  font-variant-numeric:tabular-nums;text-align:left}

.spinner{position:relative;display:inline-block;flex:none}
.spinner i{position:absolute;inset:0;border-radius:50%;
  border:2px solid rgba(184,149,90,.22);border-top-color:var(--brass);
  border-right-color:var(--brass);transform:rotate(-38deg)}
.spinner b{position:absolute;top:50%;left:50%;background:var(--brass);
  transform:translate(-50%,-50%) rotate(45deg)}
.spinner--xs i{border-width:1.5px}

/* ---- esqueleto del formulario ---- */
.skel{display:flex;flex-direction:column}
.skel__status{display:flex;align-items:center;color:var(--brass);font-weight:600;
  text-transform:uppercase}
.skel__field{flex:1;display:flex;flex-direction:column;min-width:0}
.skel__bar{display:block;background:rgba(244,240,231,.14);border-radius:2px}
.skel__bar--control{background:rgba(244,240,231,.07);
  border-bottom:2px solid rgba(184,149,90,.45)}
.skel__bar--hint{background:rgba(244,240,231,.09)}
.skel__bar--cta{background:rgba(184,149,90,.35)}

/* ---- alertas: misma nota al margen de los atlas hermanos ---- */
.alert{display:flex;border:1px solid;border-left-width:4px;border-radius:2px}
.alert__body{flex:1;display:flex;flex-direction:column}
.alert__eyebrow{font-weight:600;text-transform:uppercase;opacity:.85}
.alert__title{font-weight:600;color:var(--ink)}
.alert--danger{background:var(--danger-s);border-color:var(--danger-b);color:var(--danger-t)}
.alert--warning{background:var(--warning-s);border-color:var(--warning-b);color:var(--warning-t)}
.alert--success{background:var(--success-s);border-color:var(--success-b);color:var(--success-t)}

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
    .main{padding:22px 40px 24px;gap:18px}
    .pagehead{gap:5px}
    .pagehead__title{font-size:38px}
    .pagehead__meta{font-size:15px}

    .layout{grid-template-columns:minmax(0,1fr) 340px;
      grid-template-areas:"form rail" "acts rail";column-gap:32px}
    .layout--solo{grid-template-columns:minmax(0,980px);
      grid-template-areas:"form" "acts"}
    /* Rejilla de la hoja: la misma medida gobierna la columna de rótulos de
       cada banda y la sangría del CTA, para que el botón caiga exactamente
       bajo los campos y no bajo el título de la sección. */
    .layout{--label-col:288px;--band-gap:28px;--band-pad:26px}
    .form{gap:10px}
    .rail{gap:14px}
    .actions{gap:14px;margin-top:16px;
      padding-left:calc(var(--band-pad) + var(--label-col) + var(--band-gap) + 3px)}

    .sec{display:grid;grid-template-columns:var(--label-col) minmax(0,1fr);
      gap:var(--band-gap);padding:20px var(--band-pad)}
    .sec__head{gap:12px}
    .sec__num{width:27px;height:27px;font-size:15px}
    .sec__title{font-size:21px}
    .sec__hint{font-size:13px;line-height:18px;margin-top:3px}
    .sec__body{gap:14px}
    .sec__note{font-size:13px;line-height:18px}

    .row{gap:20px}
    .field{gap:7px}
    .field__label{font-size:11px;line-height:15px;letter-spacing:.14em}
    .control{height:46px;padding:0 14px;gap:10px}
    .control__value{font-size:15px}
    .control--textarea{height:74px;padding:12px 14px}
    .field__note{gap:8px;font-size:12px;line-height:16px;margin-top:1px}
    .field__note--loading{letter-spacing:.12em}
    .field__error{font-size:12px;line-height:16px}
    .field__foot{margin-top:2px}
    .field__counter{font-size:11px}
    .avatar{width:26px;height:26px;font-size:12px}

    .resumen{padding:18px 20px}
    .resumen__title{font-size:11px;line-height:15px;letter-spacing:.16em}
    .resumen__list{margin-top:14px}
    .resumen__item{gap:5px;padding:12px 0}
    .resumen__item:first-child{padding-top:0}
    .resumen__item:last-child{padding-bottom:0}
    .resumen__label{font-size:10px;line-height:14px;letter-spacing:.14em}
    .resumen__value{gap:10px;font-size:16px;line-height:22px}

    .btn{height:50px;padding:0 30px;font-size:16px}
    .btn--primary{align-self:flex-start;min-width:240px}
    .pagestate .btn{align-self:center;min-width:0}

    .pagestate{gap:14px;padding:40px 0}
    .pagestate__divider{gap:12px;margin:2px 0 4px}
    .pagestate__divider i{width:98px}
    .pagestate__divider b{width:7px;height:7px}
    .pagestate__eyebrow{font-size:11px;line-height:15px;letter-spacing:.16em}
    .pagestate__title{font-size:30px}
    .pagestate--note .pagestate__title{font-size:28px;line-height:36px;max-width:460px}
    .pagestate__body{font-size:16px;line-height:24px;max-width:470px}
    .pagestate__action{margin-top:10px}
    .ticket{padding:14px 18px;font-size:16px;line-height:23px;margin-top:4px}
    .success .pagestate__body{max-width:480px}

    .spinner--lg{width:54px;height:54px}
    .spinner--lg b{width:10px;height:10px}
    .spinner--sm{width:18px;height:18px}
    .spinner--sm b{width:5px;height:5px}
    .spinner--xs{width:13px;height:13px}
    .spinner--xs b{width:4px;height:4px}

    .skel{gap:18px}
    .skel__status{gap:12px;font-size:12px;letter-spacing:.14em}
    .skel__field{gap:9px}
    .skel__bar--num{width:28px;height:28px}
    .skel__bar--title{width:150px;height:15px}
    .skel__bar--hint{width:190px;height:10px;margin-top:9px}
    .skel__bar--label{width:88px;height:9px}
    .skel__bar--control{height:46px;width:100%}
    .skel__bar--cta{width:240px;height:50px}

    .alert{padding:15px 18px;max-width:640px}
    .alert__body{gap:5px}
    .alert__eyebrow{font-size:11px;line-height:15px;letter-spacing:.15em}
    .alert__title{font-size:17px;line-height:24px}
    .alert__text{font-size:15px;line-height:22px}

    .dock{padding:0 40px}
    .dock__item{gap:10px;height:64px;font-size:14px}
    .dock__icon svg{width:19px;height:19px}
  `,
  mobile: `
    .appbar{padding:16px 20px}
    .appbar__mark{font-size:26px}
    .appbar__shop{font-size:13px}
    .main{padding:20px;gap:18px}
    .pagehead{gap:5px}
    .pagehead__title{font-size:32px}
    .pagehead__meta{font-size:14px;line-height:20px}

    /* Una sola columna: el resumen queda entre el formulario y el CTA, que
       es donde §7.3 lo pide en móvil. */
    .layout{--band-pad:16px;grid-template-columns:minmax(0,1fr);
      grid-template-areas:"form" "rail" "acts";row-gap:16px}
    .layout--solo{grid-template-areas:"form" "acts"}
    .form{gap:12px}
    .rail{gap:14px}
    .actions{gap:14px;padding-left:0}

    .sec{display:flex;flex-direction:column;gap:14px;padding:18px var(--band-pad)}
    .sec__head{gap:11px}
    .sec__num{width:26px;height:26px;font-size:14px}
    .sec__title{font-size:19px}
    .sec__hint{font-size:13px;line-height:18px;margin-top:2px}
    .sec__body{gap:13px}
    .sec__note{font-size:12px;line-height:17px}

    .row{flex-direction:column;gap:13px}
    .field{gap:6px}
    .field__label{font-size:10px;line-height:14px;letter-spacing:.14em}
    .control{height:46px;padding:0 13px;gap:10px}
    .control__value{font-size:15px}
    .control--textarea{height:70px;padding:12px 13px}
    .field__note{gap:8px;font-size:12px;line-height:16px}
    .field__note--loading{letter-spacing:.12em}
    .field__error{font-size:12px;line-height:16px}
    .field__counter{font-size:11px}
    .avatar{width:26px;height:26px;font-size:12px}

    .resumen{padding:16px}
    .resumen__title{font-size:10px;line-height:14px;letter-spacing:.16em}
    .resumen__list{margin-top:12px}
    .resumen__item{gap:5px;padding:11px 0}
    .resumen__item:first-child{padding-top:0}
    .resumen__item:last-child{padding-bottom:0}
    .resumen__label{font-size:10px;line-height:14px;letter-spacing:.14em}
    .resumen__value{gap:9px;font-size:15px;line-height:21px}

    .btn{height:48px;padding:0 22px;font-size:15px}
    .btn--primary{width:100%}

    .pagestate{gap:12px;padding:34px 0;max-width:330px}
    .pagestate__divider{gap:10px;margin:2px 0 4px}
    .pagestate__divider i{width:78px}
    .pagestate__divider b{width:6px;height:6px}
    .pagestate__eyebrow{font-size:10px;line-height:14px;letter-spacing:.16em}
    .pagestate__title{font-size:25px}
    .pagestate--note .pagestate__title{font-size:23px;line-height:30px}
    .pagestate__body{font-size:15px;line-height:22px}
    .pagestate__action{margin-top:8px}
    .success .btn--primary{width:auto;align-self:stretch}
    .ticket{padding:13px 15px;font-size:15px;line-height:22px;margin-top:2px}

    .spinner--lg{width:46px;height:46px}
    .spinner--lg b{width:9px;height:9px}
    .spinner--sm{width:16px;height:16px}
    .spinner--sm b{width:5px;height:5px}
    .spinner--xs{width:13px;height:13px}
    .spinner--xs b{width:4px;height:4px}

    .skel{gap:16px}
    .skel__status{gap:10px;font-size:11px;letter-spacing:.14em}
    .skel__field{gap:8px}
    .skel__bar--num{width:26px;height:26px}
    .skel__bar--title{width:130px;height:14px}
    .skel__bar--hint{width:170px;height:10px;margin-top:8px}
    .skel__bar--label{width:80px;height:9px}
    .skel__bar--control{height:46px;width:100%}
    .skel__bar--cta{width:100%;height:48px}

    .alert{padding:13px 15px}
    .alert__body{gap:4px}
    .alert__eyebrow{font-size:10px;line-height:14px;letter-spacing:.15em}
    .alert__title{font-size:16px;line-height:23px}
    .alert__text{font-size:14px;line-height:20px}

    .dock{padding:0 8px}
    .dock__item{flex-direction:column;gap:5px;height:66px;font-size:11px}
    .dock__icon svg{width:19px;height:19px}
  `,
}

const page = (viewport, screen) => `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><style>${CSS}${SCALES[viewport]}</style></head>
<body>
  <div class="shell">
    ${headerHtml()}
    <main class="main">
      <div class="pagehead">
        <h1 class="pagehead__title">${esc(COPY.title)}</h1>
        ${screen.timezone ? `<p class="pagehead__meta">${esc(COPY.timezoneLine(screen.timezone))}</p>` : ''}
      </div>
      ${contentHtml(screen)}
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
  const heights = []

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
      // El formulario de escritorio debe caber en un viewport: si vuelve a
      // necesitar scroll, la composición se salió de presupuesto.
      if (viewport === 'desktop' && box.h > cfg.height) {
        problems.push(`${viewport}/${name}: desborde vertical ${box.h}px`)
      }
      if (viewport === 'desktop') heights.push(`${name}: ${box.h}px`)

      const dir = resolve(OUT, viewport, 'nuevo-turno')
      mkdirSync(dir, { recursive: true })
      await tab.screenshot({ path: resolve(dir, `${name}.png`), fullPage: viewport === 'mobile' })
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
