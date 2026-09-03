/*
 * Generador determinista de los mockups de `/acceso` y `/recuperar-acceso`.
 *
 * Compone cada evento como HTML con los tokens reales del sistema
 * (`apps/web/src/styles/tokens.css`) y las fuentes auto-hosteadas del
 * proyecto, y lo rasteriza con el Chromium de Playwright que ya usa la
 * suite e2e. Se prefiere HTML sobre SVG a mano porque el motor de
 * maquetación resuelve el ajuste de línea real: la revisión anterior
 * dejaba texto desbordado fuera de su alerta en móvil porque cada línea
 * estaba posicionada a mano.
 *
 * Uso:
 *   node tools/mockups/auth-eventos/render.mjs
 */
import playwright from '../../../apps/web/node_modules/@playwright/test/index.js'
import { readFileSync, mkdirSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { ACCESO, RECUPERACION } from './content.mjs'

const { chromium } = playwright

const HERE = dirname(fileURLToPath(import.meta.url))
const REPO = resolve(HERE, '../../..')
const OUT = resolve(REPO, 'docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos')
const FONTS = resolve(REPO, 'apps/web/src/assets/fonts')

const font = (file) => readFileSync(resolve(FONTS, file)).toString('base64')
const SERIF = font('instrument-serif-latin-400.woff2')
const SANS = font('instrument-sans-latin-400-700.woff2')

const esc = (s) =>
  String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')

// --------------------------------------------------------------- bloques

/*
 * Campo reglado. El lenguaje NAVA es editorial y se construye con reglas
 * (la de latón del wordmark, la del divisor del panel oscuro), no con
 * cajas redondeadas: el campo es una superficie de papel apoyada sobre una
 * línea base de tinta, con el rótulo en versalitas espaciadas como el resto
 * de rótulos cortos de la interfaz. Se retiran los iconos de sobre y
 * candado —el rótulo ya nombra el campo— y el conmutador de contraseña pasa
 * a ser la palabra `Mostrar`, que además se lee en un lector de pantalla
 * sin depender de un glifo.
 */
const fieldHtml = (f) => {
  const state = f.state ?? 'idle'
  const filled = Boolean(f.value)
  return `
    <div class="field field--${state}">
      <span class="field__label">${esc(f.label)}</span>
      <div class="control">
        <span class="control__value${filled ? '' : ' control__value--ph'}">${esc(filled ? f.value : f.placeholder)}</span>
        ${f.toggle ? '<span class="control__toggle">Mostrar</span>' : ''}
      </div>
      ${f.error ? `<p class="field__error">${esc(f.error)}</p>` : ''}
    </div>`
}

/*
 * Casillas de código: seis ranuras regladas, no seis cajas. Cada dígito se
 * apoya en su propia línea base de tinta, con la misma serif del wordmark,
 * de modo que el bloque pertenece a la misma familia que los campos y no
 * parece un componente traído de otro sistema. El valor lógico sigue siendo
 * uno solo de seis dígitos (ver contrato del README).
 */
const otpHtml = ({ label, digits, state, error }) => {
  const cells = Array.from({ length: 6 }, (_, i) => {
    const d = digits?.[i] ?? ''
    return `<span class="otp__slot">${esc(d)}</span>`
  }).join('')
  return `
    <div class="otp otp--${state}">
      ${label ? `<span class="field__label">${esc(label)}</span>` : ''}
      <div class="otp__row">${cells}</div>
      ${error ? `<p class="field__error">${esc(error)}</p>` : ''}
    </div>`
}

/*
 * Alerta como nota al margen: filete lateral del color de estado, fondo
 * apenas teñido y una palabra de estado en versalitas. Esa palabra es la
 * que cumple "icono, texto y estructura además del color" (WCAG 2.2 AA,
 * 1.4.1) sin recurrir a un glifo genérico de kit de formularios.
 */
const ALERT_EYEBROW = {
  danger: 'Error',
  warning: 'Atención',
  info: 'Nota',
  success: 'Confirmación',
}

const alertHtml = (a) => `
  <div class="alert alert--${a.variant}${a.plain ? ' alert--plain' : ''}">
    <div class="alert__body">
      <p class="alert__eyebrow">${esc(a.eyebrow ?? (a.plain ? a.title : ALERT_EYEBROW[a.variant]))}</p>
      ${a.plain ? '' : `<p class="alert__title">${esc(a.title)}</p>`}
      <p class="alert__text">${esc(a.body)}</p>
      ${a.action ? `<span class="alert__action">${esc(a.action)}</span>` : ''}
    </div>
    ${a.dismissible ? '<span class="alert__close">Descartar</span>' : ''}
  </div>`

const blockHtml = (b) => {
  switch (b.type) {
    case 'fields':
      return `<div class="fields">${b.fields.map(fieldHtml).join('')}</div>`
    case 'submit':
      return `<div class="btn btn--primary btn--${b.state ?? 'idle'}">${
        b.state === 'loading' ? '<span class="spinner"></span>' : ''
      }<span>${esc(b.label)}</span></div>`
    case 'ghost':
      return `<div class="btn btn--ghost">${esc(b.label)}</div>`
    case 'link':
      return `<p class="textlink${b.state === 'muted' ? ' textlink--muted' : ''}"><span>${esc(b.label)}</span></p>`
    case 'note':
      return `<p class="note">${esc(b.text)}</p>`
    case 'alert':
      return alertHtml(b)
    case 'otp':
      return otpHtml(b)
    // El reto deja de ser una tarjeta sobrepuesta y pasa a ser una sección
    // de la misma columna, abierta por la regla de latón con rombo que ya
    // separa contenido en el panel oscuro. Así se lee como continuación del
    // formulario y no como un widget pegado encima.
    case 'challenge':
      return `
        <section class="section">
          <div class="section__rule"><span></span><i></i><span></span></div>
          <p class="section__eyebrow">${esc(b.chip)}</p>
          <p class="section__note">${esc(b.note)}</p>
          ${otpHtml({ label: 'Código de 6 dígitos', ...b.otp, error: b.error })}
          <div class="section__actions">
            <div class="btn btn--primary">${esc(b.primary)}</div>
            <div class="btn btn--ghost">${esc(b.secondary)}</div>
          </div>
        </section>`
    default:
      throw new Error(`bloque desconocido: ${b.type}`)
  }
}

// ----------------------------------------------------------------- shell

const wordmark = (size) => `
  <div class="wordmark" style="--wm:${size}px">
    <span class="wordmark__text">NAVA</span>
    <span class="wordmark__rule"></span>
  </div>`

const headHtml = (route, screen, cfg) => {
  if (route === 'acceso') {
    return `
      <header class="head">
        ${wordmark(cfg.wordmark)}
        <h1 class="head__title">${esc(ACCESO.title)}</h1>
        <p class="head__sub">${esc(ACCESO.subtitle)}</p>
      </header>`
  }
  return `
    <header class="head">
      ${wordmark(cfg.wordmark)}
      <p class="head__step">Paso ${screen.step} de 3</p>
      <h1 class="head__title head__title--step">${esc(screen.title)}</h1>
    </header>`
}

const asideHtml = (route) => `
  <aside class="aside">
    <div class="aside__mark">NAVA</div>
    <div class="aside__divider"><span></span><i></i><span></span></div>
    <p class="aside__tagline">Gestión precisa<br />para tu barbería.</p>
    <p class="aside__caption">${route === 'acceso' ? ACCESO.caption : RECUPERACION.caption} · NAVA</p>
  </aside>`

const CSS = `
@font-face{font-family:'Instrument Serif';font-weight:400;font-display:block;src:url(data:font/woff2;base64,${SERIF}) format('woff2')}
@font-face{font-family:'Instrument Sans';font-weight:400 700;font-display:block;src:url(data:font/woff2;base64,${SANS}) format('woff2')}

:root{
  --ink:#101b2b; --canvas:#f4f0e7; --surface:#fff; --surface-muted:#e8e2d8;
  --text:#2a2d32; --text-2:#5e625f; --border:#c9c0b2; --brass:#b8955a; --brass-deep:#765c2f;
  --danger-s:#f8edec; --danger-t:#8a2c2c; --danger-b:#a43a3a;
  --warning-s:#f8f1df; --warning-t:#775019; --warning-b:#9a6a24;
  --info-s:#e9eef3; --info-t:#23405b; --info-b:#667d93;
  --success-s:#eaf0eb; --success-t:#325d43; --success-b:#748477;
  --serif:'Instrument Serif',Georgia,serif;
  --sans:'Instrument Sans',system-ui,sans-serif;
}
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:var(--sans);color:var(--text);background:var(--canvas);-webkit-font-smoothing:antialiased}

/* ---- cascarón ---- */
.shell{display:grid;min-height:100vh}
.shell--desktop{grid-template-columns:620px 1fr}
.shell--mobile{grid-template-columns:1fr}

.aside{background:var(--ink);color:var(--canvas);display:flex;flex-direction:column;
  align-items:center;justify-content:center;gap:44px;padding:64px 48px}
.aside__mark{font-family:var(--serif);font-size:132px;line-height:1;letter-spacing:.02em}
.aside__divider{display:flex;align-items:center;gap:14px;width:440px}
.aside__divider span{flex:1;height:1.5px;background:var(--brass)}
.aside__divider i{width:20px;height:20px;background:var(--brass);transform:rotate(45deg)}
.aside__tagline{font-family:var(--serif);font-size:44px;line-height:1.24;text-align:center}
.aside__caption{font-size:14px;letter-spacing:.14em;color:#a9a9a3;margin-top:12px}

.pane{display:flex;flex-direction:column;align-items:center;justify-content:center;
  padding:56px 0;background:var(--canvas)}
.shell--mobile .pane{justify-content:flex-start;padding:40px 24px}
.col{display:flex;flex-direction:column;width:100%}

/* ---- encabezado ---- */
.head{display:flex;flex-direction:column;align-items:center;text-align:center}
.wordmark{display:inline-flex;flex-direction:column;align-items:stretch;margin-bottom:22px}
.wordmark__text{font-family:var(--serif);font-size:var(--wm);line-height:1.06;
  letter-spacing:.02em;color:var(--ink)}
.wordmark__rule{height:2px;background:var(--brass);margin-top:6px}
.head__step{font-size:13px;font-weight:600;letter-spacing:.14em;color:var(--brass-deep);
  text-transform:uppercase;margin-bottom:8px}
.head__title{font-family:var(--serif);color:var(--ink);letter-spacing:-.01em}
.head__sub{color:var(--text-2);margin-top:8px}

/* ---- campos reglados ----
   Papel apoyado en una línea base de tinta. El rótulo va en versalitas
   espaciadas, igual que PASO 1 DE 3 y ACCESO SEGURO, para que el
   formulario hable el mismo idioma que el resto de la composición. */
.fields{display:flex;flex-direction:column}
.field{display:flex;flex-direction:column}
.field__label{font-weight:600;color:var(--brass-deep);text-transform:uppercase}
.control{display:flex;align-items:flex-end;background:var(--surface);
  border:1px solid rgba(16,27,43,.12);border-bottom:2px solid var(--ink);border-radius:2px}
.control__value{flex:1;color:var(--ink);white-space:nowrap;overflow:hidden}
.control__value--ph{color:#9a978f}
.control__toggle{flex:none;color:var(--brass-deep);font-weight:600;
  text-transform:uppercase;border-bottom:1px solid var(--brass)}
.field--error .control{border-color:rgba(164,58,58,.35);border-bottom-color:var(--danger-b)}
.field--disabled .control{background:transparent;border-color:rgba(16,27,43,.08);
  border-bottom-color:var(--border)}
.field--disabled .control__value,.field--disabled .control__toggle{color:#9a978f}
.field--disabled .control__toggle{border-bottom-color:var(--border)}
.field__error{color:var(--danger-t);font-weight:500}

/* ---- ranuras de código ----
   Seis dígitos en la serif del wordmark, cada uno sobre su línea base. No
   son cajas: pertenecen a la misma familia reglada que los campos. */
.otp{display:flex;flex-direction:column}
.otp__row{display:grid;grid-template-columns:repeat(6,1fr)}
.otp__slot{display:flex;align-items:flex-end;justify-content:center;
  font-family:var(--serif);color:var(--ink);background:var(--surface);
  border:1px solid rgba(16,27,43,.12);border-bottom:2px solid var(--ink);border-radius:2px}
.otp--error .otp__slot{border-color:rgba(164,58,58,.35);border-bottom-color:var(--danger-b)}

/* ---- acciones ---- */
.btn{display:flex;align-items:center;justify-content:center;gap:12px;border-radius:2px;
  font-weight:600;text-align:center;letter-spacing:.02em}
.btn--primary{background:var(--ink);color:var(--canvas)}
.btn--disabled{background:transparent;color:#9a978f;
  border:1px solid rgba(16,27,43,.14);border-bottom:2px solid var(--border)}
.btn--loading{background:#4d5560;color:var(--canvas)}
.btn--ghost{background:transparent;color:var(--brass-deep);
  border:1px solid rgba(118,92,47,.5);border-bottom:2px solid var(--brass-deep)}
.spinner{width:18px;height:18px;border:2.5px solid rgba(244,240,231,.35);
  border-top-color:var(--canvas);border-radius:50%}
.textlink{text-align:center;color:var(--brass-deep);font-weight:500}
.textlink span{border-bottom:1px solid var(--brass);padding-bottom:2px}
.textlink--muted{color:#9a978f}
.textlink--muted span{border-bottom-color:#c3bcb0}
.note{color:var(--text-2)}

/* ---- alertas como nota al margen ----
   Filete lateral del color de estado sobre un fondo apenas teñido. La
   palabra de estado en versalitas es el signo no cromático; sustituye al
   glifo circular genérico de la revisión anterior. */
.alert{display:flex;align-items:flex-start;border:1px solid;border-left-width:4px;
  border-radius:2px}
.alert__body{flex:1;display:flex;flex-direction:column}
.alert__eyebrow{font-weight:600;text-transform:uppercase;opacity:.85}
.alert__title{font-weight:600;color:var(--ink)}
.alert__action{align-self:flex-start;border:1px solid currentColor;
  border-bottom-width:2px;border-radius:2px;font-weight:600;
  text-transform:uppercase;letter-spacing:.08em}
.alert__close{flex:none;font-weight:600;text-transform:uppercase;opacity:.72}
.alert--danger{background:var(--danger-s);border-color:var(--danger-b);color:var(--danger-t)}
.alert--warning{background:var(--warning-s);border-color:var(--warning-b);color:var(--warning-t)}
.alert--info{background:var(--info-s);border-color:var(--info-b);color:var(--info-t)}
.alert--success{background:var(--success-s);border-color:var(--success-b);color:var(--success-t)}
.alert--plain{background:transparent;border-color:var(--border);
  border-left-color:var(--brass);color:var(--text-2)}
.alert--plain .alert__eyebrow{color:var(--brass-deep)}

/* ---- sección del reto ----
   Misma regla de latón con rombo que divide el panel oscuro: abre una
   sección de la columna en vez de sobreponer una tarjeta. */
.section{display:flex;flex-direction:column}
.section__rule{display:flex;align-items:center;gap:12px}
.section__rule span{flex:1;height:1px;background:var(--brass)}
.section__rule i{width:9px;height:9px;background:var(--brass);transform:rotate(45deg)}
.section__eyebrow{color:var(--brass-deep);font-weight:600;text-transform:uppercase}
.section__note{color:var(--text-2)}
.challenge__actions{display:grid;grid-template-columns:1fr auto}
`

// Escala por viewport. Un solo lugar decide tamaños y ritmo vertical, así
// que escritorio y móvil no pueden desincronizarse por un ajuste manual.
const SCALES = {
  desktop: `
    .col{max-width:578px;gap:22px}
    .head__title{font-size:44px;line-height:52px}
    .head__title--step{font-size:36px;line-height:44px}
    .head__sub{font-size:17px;line-height:24px}
    .fields{gap:22px}
    .field{gap:9px}
    .field__label{font-size:12px;line-height:16px;letter-spacing:.13em}
    .control{height:58px;padding:0 16px 13px;gap:16px}
    .control__value{font-size:18px;line-height:1}
    .control__toggle{font-size:11px;line-height:1;letter-spacing:.11em}
    .field__error{font-size:14px;line-height:19px;margin-top:4px}
    .otp{gap:12px}
    .otp__row{gap:14px}
    .otp__slot{height:62px;font-size:34px;line-height:1;padding-bottom:10px}
    .btn{height:58px;font-size:18px;padding:0 20px}
    .btn--ghost{height:58px;font-size:16px}
    .textlink{font-size:17px;line-height:24px}
    .note{font-size:16px;line-height:24px}
    .alert{padding:15px 18px}
    .alert__body{gap:5px}
    .alert__eyebrow{font-size:11px;line-height:15px;letter-spacing:.15em}
    .alert__title{font-size:18px;line-height:25px}
    .alert__text{font-size:15px;line-height:22px}
    .alert__action{margin-top:8px;padding:9px 16px;font-size:12px}
    .alert__close{font-size:11px;letter-spacing:.12em;margin-top:2px}
    .section{gap:16px}
    .section__eyebrow{font-size:12px;line-height:16px;letter-spacing:.15em;margin-top:-2px}
    .section__note{font-size:15px;line-height:22px;margin-top:-6px}
    .section__actions{display:grid;grid-template-columns:1fr auto;gap:12px;margin-top:10px}

    /* El reto suma una sección completa al formulario ya existente. Se
       comprime el ritmo en vez de recortar contenido, para que el evento
       siga cabiendo en una pantalla sin desplazamiento. */
    .pane--dense{padding:36px 0}
    .col--dense{gap:16px}
    .col--dense .wordmark{margin-bottom:12px}
    .col--dense .fields{gap:18px}
    .col--dense .control{height:54px;padding-bottom:12px}
    .col--dense .btn{height:54px}
    .col--dense .otp__slot{height:56px;font-size:30px}
    .col--dense .section{gap:14px}
  `,
  mobile: `
    .col{max-width:100%;gap:20px}
    .head__title{font-size:36px;line-height:42px}
    .head__title--step{font-size:30px;line-height:36px}
    .head__sub{font-size:15px;line-height:22px}
    .fields{gap:20px}
    .field{gap:8px}
    .field__label{font-size:11px;line-height:15px;letter-spacing:.13em}
    .control{height:54px;padding:0 14px 12px;gap:14px}
    .control__value{font-size:17px;line-height:1}
    .control__toggle{font-size:10px;line-height:1;letter-spacing:.11em}
    .field__error{font-size:13px;line-height:18px;margin-top:4px}
    .otp{gap:10px}
    .otp__row{gap:10px}
    .otp__slot{height:56px;font-size:30px;line-height:1;padding-bottom:9px}
    .btn{height:54px;font-size:16px;padding:0 16px}
    .btn--ghost{height:52px;font-size:15px}
    .textlink{font-size:15px;line-height:22px}
    .note{font-size:15px;line-height:22px}
    .alert{padding:13px 15px}
    .alert__body{gap:4px}
    .alert__eyebrow{font-size:10px;line-height:14px;letter-spacing:.15em}
    .alert__title{font-size:16px;line-height:23px}
    .alert__text{font-size:14px;line-height:20px}
    .alert__action{margin-top:8px;padding:9px 14px;font-size:11px}
    .alert__close{font-size:10px;letter-spacing:.12em;margin-top:2px}
    .section{gap:14px}
    .section__eyebrow{font-size:11px;line-height:15px;letter-spacing:.15em;margin-top:-2px}
    .section__note{font-size:14px;line-height:20px;margin-top:-4px}
    .section__actions{display:grid;grid-template-columns:1fr;gap:10px;margin-top:8px}
  `,
}

const page = (viewport, route, screen, cfg) => {
  const dense = viewport === 'desktop' && screen.blocks.some((b) => b.type === 'challenge')
  return `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><style>${CSS}${SCALES[viewport]}</style></head>
<body>
  <div class="shell shell--${viewport}">
    ${viewport === 'desktop' ? asideHtml(route) : ''}
    <main class="pane${dense ? ' pane--dense' : ''}">
      <div class="col${dense ? ' col--dense' : ''}">
        ${headHtml(route, screen, cfg)}
        ${screen.blocks.map(blockHtml).join('\n')}
        ${viewport === 'mobile' ? `<p class="aside__caption" style="color:#8a8880;text-align:center;margin-top:8px">${route === 'acceso' ? ACCESO.caption : RECUPERACION.caption} · NAVA</p>` : ''}
      </div>
    </main>
  </div>
</body></html>`
}

const VIEWPORTS = {
  desktop: { width: 1440, height: 1024, deviceScaleFactor: 1, wordmark: 54 },
  mobile: { width: 420, height: 935, deviceScaleFactor: 2, wordmark: 54 },
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

    for (const [route, spec] of [
      ['acceso', ACCESO],
      ['recuperacion', RECUPERACION],
    ]) {
      for (const [name, screen] of Object.entries(spec.screens)) {
        await tab.setContent(page(viewport, route, screen, cfg), { waitUntil: 'load' })
        await tab.evaluate(() => document.fonts.ready)

        // Guardia: el desbordamiento silencioso fue el defecto principal de
        // la revisión anterior, así que se detecta aquí y no en la lámina.
        const overflow = await tab.evaluate(() => ({
          h: document.documentElement.scrollHeight,
          w: document.documentElement.scrollWidth,
        }))
        if (overflow.w > cfg.width) {
          problems.push(`${viewport}/${route}/${name}: desborde horizontal ${overflow.w}px`)
        }
        if (viewport === 'desktop' && overflow.h > cfg.height) {
          problems.push(`${viewport}/${route}/${name}: desborde vertical ${overflow.h}px`)
        }

        const dir = resolve(OUT, viewport, route)
        mkdirSync(dir, { recursive: true })
        await tab.screenshot({
          path: resolve(dir, `${name}.png`),
          fullPage: viewport === 'mobile',
        })
      }
    }
    await ctx.close()
  }

  await browser.close()

  if (problems.length) {
    console.error('Problemas de composición:\n- ' + problems.join('\n- '))
    process.exitCode = 1
  } else {
    console.log('46 mockups regenerados sin desbordes.')
  }
}

await main()
