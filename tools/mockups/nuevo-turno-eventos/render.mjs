/*
 * Generador determinista de los mockups de `/panel/turnos/nuevo` (Nuevo
 * turno).
 *
 * Comparte criterio y lenguaje visual con `tools/mockups/panel-agenda-eventos`
 * y `tools/mockups/auth-eventos`: HTML compuesto con los tokens reales del
 * sistema y las fuentes auto-hosteadas del proyecto, rasterizado con el
 * Chromium de Playwright que ya usa la suite e2e. El encabezado y el dock
 * reutilizan exactamente el cascarón de `panel-agenda-eventos`: la lámina
 * compuesta `03-nuevo-turno.png` mostraba botones de cuenta decorativos y
 * navegación duplicada en el encabezado, la misma inconsistencia que ya se
 * corrigió al construir `auth-eventos` y `panel-agenda-eventos`.
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

const FIELD_ICONS = {
  person:
    '<circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 3.6-7 8-7s8 3 8 7"/>',
  service:
    '<path d="M6 4 18 16M18 4 6 16"/><circle cx="6" cy="18" r="2.2"/><circle cx="18" cy="18" r="2.2"/>',
  schedule: '<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M3 10h18M8 3v4M16 3v4"/>',
  phone:
    '<path d="M6 3h3l1.5 4-2 1.5a12 12 0 0 0 6 6l1.5-2 4 1.5v3a2 2 0 0 1-2.2 2A17 17 0 0 1 4 6.2 2 2 0 0 1 6 3z"/>',
  timezone: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a13 13 0 0 1 0 18 13 13 0 0 1 0-18z"/>',
}
const fieldIcon = (n) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">${FIELD_ICONS[n]}</svg>`

const ALERT_EYEBROW = { danger: 'Error', warning: 'Atención', info: 'Nota', success: 'Confirmación' }

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

const highlightDest = (t) => t.replace(/«([^»]+)»/g, '<span class="pagestate__dest">«$1»</span>')

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

const fieldHtml = ({ label, required, value, placeholder, error, hint, disabled, type = 'text' }) => `
  <div class="field${error ? ' field--error' : ''}">
    <label class="field__label">${esc(label)}${required ? reqMark : ''}</label>
    <div class="input${disabled ? ' input--disabled' : ''}${type === 'select' ? ' input--select' : ''}">
      ${
        value
          ? `<span class="input__value">${esc(value)}</span>`
          : `<span class="input__value input__value--ph">${esc(placeholder)}</span>`
      }
      ${type === 'select' ? '<span class="input__caret">▾</span>' : ''}
    </div>
    ${hint ? `<p class="field__hint">${esc(hint)}</p>` : ''}
    ${error ? `<p class="field__error-text" role="alert">${esc(error)}</p>` : ''}
  </div>`

const textareaHtml = ({ label, value, hint, counter }) => `
  <div class="field">
    <label class="field__label">${esc(label)}</label>
    <div class="input input--textarea">
      <span class="input__value${value ? '' : ' input__value--ph'}">${esc(value || '')}</span>
    </div>
    <div class="field__foot">
      ${hint ? `<p class="field__hint">${esc(hint)}</p>` : '<span></span>'}
      <span class="field__counter">${esc(counter)}</span>
    </div>
  </div>`

const cardHtml = (n, title, hint, body) => `
  <section class="card">
    <div class="card__head">
      <span class="card__badge" aria-hidden="true">${n}</span>
      <div>
        <h2 class="card__title">${esc(title)}</h2>
        <p class="card__hint">${esc(hint)}</p>
      </div>
    </div>
    <div class="card__body">${body}</div>
  </section>`

const resumenRow = (icon, label, value) => `
  <div class="resumen__item">
    <span class="resumen__icon">${fieldIcon(icon)}</span>
    <span class="resumen__stack">
      <span class="resumen__value">${esc(value)}</span>
      <span class="resumen__label">${esc(label)}</span>
    </span>
  </div>`

const resumenHtml = (f, tz) => `
  <div class="resumen">
    <p class="resumen__title">${esc(COPY.resumenTitle)}</p>
    <div class="resumen__grid">
      ${resumenRow('person', COPY.resumenBarber, f.barber)}
      ${resumenRow('service', COPY.resumenService, f.service)}
      ${resumenRow('schedule', COPY.resumenSchedule, `${f.date} · ${f.time}`)}
      ${resumenRow('person', COPY.resumenAttendee, f.attendee)}
      ${f.phone ? resumenRow('phone', COPY.resumenPhone, f.phone) : ''}
      ${resumenRow('timezone', COPY.resumenTimezone, tz)}
    </div>
  </div>`

// ----------------------------------------------------------- contenido

const formHtml = (screen) => {
  const f = screen.fields
  const fe = screen.fieldErrors ?? {}
  const attempted = !!screen.attempted

  const barberField = fieldHtml({
    label: COPY.barberLabel,
    type: 'select',
    value: f.barber,
    placeholder: COPY.barberPlaceholder,
    error: attempted ? fe.barber : undefined,
  })
  const serviceHint =
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
    hint: serviceHint,
    error: attempted ? fe.service : undefined,
  })

  const section1 = cardHtml(
    1,
    COPY.section1Title,
    COPY.section1Hint,
    `<div class="form-row">${barberField}${serviceField}</div>`,
  )

  const section2 = cardHtml(
    2,
    COPY.section2Title,
    COPY.section2Hint,
    `${fieldHtml({ label: COPY.attendeeLabel, required: true, value: f.attendee, placeholder: COPY.attendeeLabel, error: attempted ? fe.attendee : undefined })}
     ${fieldHtml({ label: COPY.customerNameLabel, required: true, value: f.customerName, placeholder: COPY.customerNameLabel, error: attempted ? fe.customerName : undefined })}
     <div class="form-row">
       ${fieldHtml({ label: COPY.phoneLabel, value: f.phone, placeholder: COPY.phonePlaceholder, error: attempted ? fe.phone : undefined })}
       ${fieldHtml({ label: COPY.emailLabel, value: f.email, placeholder: COPY.emailLabel, error: attempted ? fe.email : undefined })}
     </div>
     ${!f.phone && !f.email ? `<p class="field__hint">${esc(COPY.noContactHint)}</p>` : ''}`,
  )

  const section3 = cardHtml(
    3,
    COPY.section3Title,
    COPY.section3Hint,
    `<div class="form-row">
       ${fieldHtml({ label: COPY.dateLabel, required: true, value: f.date, placeholder: COPY.dateLabel })}
       ${fieldHtml({ label: COPY.timeLabel, required: true, value: f.time, placeholder: COPY.timeLabel, error: attempted ? fe.time : undefined })}
     </div>`,
  )

  const section4 = cardHtml(
    4,
    COPY.section4Title,
    COPY.section4Hint,
    textareaHtml({ label: COPY.noteLabel, value: f.note, counter: `${f.note.length}/300` }),
  )

  const resumen = screen.resumen ? resumenHtml(f, screen.timezone) : ''
  const alert = screen.alert ? alertHtml(screen.alert) : ''

  return `
    ${section1}${section2}${section3}${section4}
    ${resumen}
    ${alert}
    <span class="btn btn--primary btn--submit${screen.saving ? ' btn--disabled' : ''}">
      ${esc(screen.saving ? COPY.submitting : COPY.submit)}
    </span>`
}

const successHtml = (screen) => `
  <div class="success">
    ${alertHtml({ variant: 'success', title: COPY.successTitle, body: COPY.successSummary(screen.created) })}
    <p class="success__note">${esc(COPY.successNote)}</p>
    <span class="btn btn--primary btn--submit">${esc(COPY.successAgain)}</span>
  </div>`

const contentHtml = (screen) => {
  if (screen.page === 'loading') {
    return pageStateHtml({ tone: 'loading', title: COPY.loadingBarbers })
  }
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
    return pageStateHtml({ tone: 'note', title: COPY.noBarbers })
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
  --serif:'Instrument Serif',Georgia,serif;
  --sans:'Instrument Sans',system-ui,sans-serif;
}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:var(--sans);background:var(--ink);color:var(--on-ink);-webkit-font-smoothing:antialiased}
.shell{display:flex;flex-direction:column;min-height:100vh}

.appbar{display:flex;align-items:center;justify-content:space-between;
  border-bottom:1px solid rgba(184,149,90,.45)}
.appbar__brand{display:flex;align-items:baseline;gap:14px}
.appbar__mark{font-family:var(--serif);color:var(--on-ink);letter-spacing:.02em;line-height:1}
.appbar__shop{color:var(--on-ink-2)}

.main{flex:1;display:flex;flex-direction:column}
.pagehead{display:flex;flex-direction:column}
.pagehead__title{font-family:var(--serif);color:var(--on-ink);line-height:1.05}
.pagehead__meta{color:var(--on-ink-2)}

/* ---- tarjetas numeradas (formulario sobre pergamino) ---- */
.card{background:var(--paper);border-radius:6px;display:flex;flex-direction:column}
.card__head{display:flex;align-items:flex-start}
.card__badge{flex:none;display:flex;align-items:center;justify-content:center;
  border:1.5px solid var(--brass);border-radius:999px;color:var(--brass-deep);
  font-family:var(--serif)}
.card__title{font-family:var(--serif);color:var(--text);line-height:1.2}
.card__hint{color:var(--text-2)}
.card__body{display:flex;flex-direction:column}

.form-row{display:flex}
.field{display:flex;flex-direction:column;flex:1}
.field__label{font-weight:600;color:var(--text)}
.field__req{color:var(--brass-deep)}
.field__hint{color:var(--text-2)}
.field__error-text{color:var(--danger-t);font-weight:600}
.field__foot{display:flex;align-items:center;justify-content:space-between}
.field__counter{color:var(--text-2)}

.input{display:flex;align-items:center;justify-content:space-between;background:#fff;
  border:1px solid var(--border);border-radius:4px;color:var(--text)}
.input__value--ph{color:var(--text-2)}
.input__caret{color:var(--text-2)}
.input--disabled{opacity:.55;background:var(--paper-2)}
.input--textarea{align-items:flex-start;min-height:64px}
.field--error .input{border-color:var(--danger-b)}

/* ---- resumen ---- */
.resumen{background:var(--ink-2);border-radius:6px;border:1px solid rgba(184,149,90,.3)}
.resumen__title{color:var(--on-ink-2);font-weight:600;text-transform:uppercase}
.resumen__grid{display:flex;flex-wrap:wrap}
.resumen__item{display:flex;align-items:center}
.resumen__icon{flex:none;color:var(--brass);display:flex}
.resumen__stack{display:flex;flex-direction:column;min-width:0}
.resumen__value{color:var(--on-ink);font-weight:600;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.resumen__label{color:var(--on-ink-2)}

/* ---- acciones ---- */
.btn{display:inline-flex;align-items:center;justify-content:center;border-radius:4px;
  font-weight:600;letter-spacing:.02em}
.btn--primary{background:var(--brass);color:var(--ink)}
.btn--secondary{background:var(--canvas);color:var(--ink)}
.btn--submit{width:100%}
.btn--disabled{opacity:.45}

/* ---- estados de página ---- */
.pagestate{display:flex;flex-direction:column;align-items:center;text-align:center;
  margin:auto;max-width:460px}
.pagestate__mark{display:flex;align-items:center;justify-content:center}
.pagestate__divider{display:flex;align-items:center;justify-content:center}
.pagestate__divider i{display:block;height:1px;background:var(--brass);opacity:.65}
.pagestate__divider b{display:block;background:var(--brass);transform:rotate(45deg);flex:none}
.pagestate--warning .pagestate__divider b{background:var(--warning-b)}
.pagestate__eyebrow{font-weight:600;text-transform:uppercase;color:var(--warning-b)}
.pagestate__title{font-family:var(--serif);color:var(--on-ink);line-height:1.15}
.pagestate__body{color:var(--on-ink-2)}
.spinner{position:relative;display:inline-block;flex:none}
.spinner i{position:absolute;inset:0;border-radius:50%;
  border:2px solid rgba(184,149,90,.22);border-top-color:var(--brass);
  border-right-color:var(--brass);transform:rotate(-38deg)}
.spinner b{position:absolute;top:50%;left:50%;background:var(--brass);
  transform:translate(-50%,-50%) rotate(45deg)}

/* ---- alertas ---- */
.alert{display:flex;border:1px solid;border-left-width:4px;border-radius:2px}
.alert__body{flex:1;display:flex;flex-direction:column}
.alert__eyebrow{font-weight:600;text-transform:uppercase;opacity:.85}
.alert__title{font-weight:600;color:var(--ink)}
.alert--danger{background:var(--danger-s);border-color:var(--danger-b);color:var(--danger-t)}
.alert--warning{background:var(--warning-s);border-color:var(--warning-b);color:var(--warning-t)}
.alert--success{background:var(--success-s);border-color:var(--success-b);color:var(--success-t)}

.success{display:flex;flex-direction:column}
.success__note{color:var(--on-ink-2)}

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
    .main{padding:26px 40px 30px;gap:22px;max-width:900px;margin:0 auto;width:100%}
    .pagehead{gap:6px}
    .pagehead__title{font-size:38px}
    .pagehead__meta{font-size:15px}
    .card{gap:20px;padding:24px 28px}
    .card__head{gap:16px}
    .card__badge{width:36px;height:36px;font-size:16px}
    .card__title{font-size:22px}
    .card__hint{font-size:14px;margin-top:2px}
    .card__body{gap:16px}
    .form-row{gap:20px}
    .field{gap:6px}
    .field__label{font-size:14px}
    .field__hint{font-size:13px}
    .field__error-text{font-size:13px}
    .field__counter{font-size:12px}
    .input{height:46px;padding:0 14px;font-size:15px}
    .input--textarea{padding:12px 14px;font-size:15px}
    .resumen{padding:20px 24px;gap:14px;display:flex;flex-direction:column}
    .resumen__title{font-size:12px;letter-spacing:.12em}
    .resumen__grid{gap:22px 32px}
    .resumen__item{gap:12px}
    .resumen__icon svg{width:20px;height:20px}
    .resumen__value{font-size:15px;max-width:220px}
    .resumen__label{font-size:12px}
    .btn--submit{height:52px;font-size:16px}
    .pagestate{gap:14px;padding:40px 0}
    .pagestate__divider{gap:12px;margin:2px 0 4px}
    .pagestate__divider i{width:98px}
    .pagestate__divider b{width:7px;height:7px}
    .pagestate__eyebrow{font-size:11px;line-height:15px;letter-spacing:.16em}
    .pagestate__title{font-size:28px}
    .pagestate__body{font-size:16px;line-height:24px;max-width:452px}
    .pagestate__action{margin-top:10px;height:48px;padding:0 22px;font-size:15px}
    .spinner--lg{width:54px;height:54px}
    .spinner--lg b{width:10px;height:10px}
    .alert{padding:15px 18px}
    .alert__body{gap:5px}
    .alert__eyebrow{font-size:11px;line-height:15px;letter-spacing:.15em}
    .alert__title{font-size:17px;line-height:24px}
    .alert__text{font-size:15px;line-height:22px}
    .success{gap:18px;max-width:520px}
    .success__note{font-size:15px;line-height:22px}
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
    .pagehead__title{font-size:30px}
    .pagehead__meta{font-size:13px}
    .card{gap:16px;padding:18px}
    .card__head{gap:12px}
    .card__badge{width:30px;height:30px;font-size:14px}
    .card__title{font-size:18px}
    .card__hint{font-size:13px;margin-top:2px}
    .card__body{gap:14px}
    .form-row{flex-direction:column;gap:14px}
    .field{gap:5px}
    .field__label{font-size:13px}
    .field__hint{font-size:12px}
    .field__error-text{font-size:12px}
    .field__counter{font-size:11px}
    .input{height:44px;padding:0 12px;font-size:14px}
    .input--textarea{padding:11px 12px;font-size:14px}
    .resumen{padding:16px;gap:12px;display:flex;flex-direction:column}
    .resumen__title{font-size:11px;letter-spacing:.1em}
    .resumen__grid{flex-direction:column;gap:12px}
    .resumen__item{gap:10px}
    .resumen__icon svg{width:18px;height:18px}
    .resumen__value{font-size:14px;max-width:240px}
    .resumen__label{font-size:11px}
    .btn--submit{height:48px;font-size:15px}
    .pagestate{gap:12px;padding:34px 0;max-width:300px}
    .pagestate__divider{gap:10px;margin:2px 0 4px}
    .pagestate__divider i{width:78px}
    .pagestate__divider b{width:6px;height:6px}
    .pagestate__eyebrow{font-size:10px;line-height:14px;letter-spacing:.16em}
    .pagestate__title{font-size:23px}
    .pagestate__body{font-size:14px;line-height:21px}
    .pagestate__action{margin-top:8px;height:46px;padding:0 18px;font-size:14px}
    .spinner--lg{width:46px;height:46px}
    .spinner--lg b{width:9px;height:9px}
    .alert{padding:13px 15px}
    .alert__body{gap:4px}
    .alert__eyebrow{font-size:10px;line-height:14px;letter-spacing:.15em}
    .alert__title{font-size:16px;line-height:23px}
    .alert__text{font-size:14px;line-height:20px}
    .success{gap:16px}
    .success__note{font-size:14px;line-height:21px}
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
      ${
        screen.page === 'form' || screen.page === 'success'
          ? `<div class="pagehead">
               <h1 class="pagehead__title">${esc(COPY.title)}</h1>
               ${screen.timezone ? `<p class="pagehead__meta">${esc(COPY.timezoneLine(screen.timezone))}</p>` : ''}
             </div>`
          : `<div class="pagehead"><h1 class="pagehead__title">${esc(COPY.title)}</h1></div>`
      }
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

      const dir = resolve(OUT, viewport, 'nuevo-turno')
      mkdirSync(dir, { recursive: true })
      await tab.screenshot({ path: resolve(dir, `${name}.png`), fullPage: true })
    }
    await ctx.close()
  }

  await browser.close()

  if (problems.length) {
    console.error('Problemas de composición:\n- ' + problems.join('\n- '))
    process.exitCode = 1
  } else {
    console.log(`${Object.keys(SCREENS).length * 2} mockups regenerados sin desbordes.`)
  }
}

await main()
