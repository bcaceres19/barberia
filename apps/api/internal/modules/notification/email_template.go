package notification

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// Plantilla compartida de los correos transaccionales (issue #465).
//
// Los cuatro correos que envía el backend (código de recuperación, código de
// acceso, confirmación de turno y aviso de cambio de HU-130) comparten la misma
// estructura visual: banda de marca, insignia de estado opcional, título,
// párrafos, tarjeta de datos, código, botón y pie. Cada remitente solo decide
// el contenido; la presentación vive aquí para que cambiarla sea un único
// cambio.
//
// Decisiones de diseño:
//   - Identidad NAVA / Tailored Grid (DEC-077): tinta #101b2b, crema #f4f0e7 y
//     latón #b8955a, los mismos valores de apps/web/src/styles/tokens.css. Un
//     correo no puede leer variables CSS, por eso se repiten como literales.
//   - Compatibilidad con clientes de correo: tablas de maquetación con
//     role="presentation", estilos en línea, ancho máximo 600 px que se
//     contrae a la pantalla (320 px incluido) sin depender de media queries,
//     tipografías del sistema y ninguna imagen remota (no hay logo alojado que
//     un cliente pueda bloquear: la marca es un wordmark de texto).
//   - Modo oscuro tolerante: `color-scheme` más una hoja <style> con
//     prefers-color-scheme. Los clientes que la ignoran muestran la versión
//     clara, que por sí sola cumple contraste AA.
//   - Texto alternativo en la misma petición (Text): mejora entrega y lo leen
//     los clientes sin HTML y los lectores de pantalla que lo prefieren.
//   - Todo dato dinámico pasa por html/template, que escapa por contexto (texto
//     y atributos) y filtra las URL peligrosas; nombres de clientes y barberías
//     son entrada de usuarios.

// emailTone elige la paleta de la insignia de estado. Los pares fondo/texto son
// los de las superficies de estado de tokens.css y superan 4.5:1.
type emailTone int

const (
	toneNeutral emailTone = iota
	toneSuccess
	toneWarning
	toneDanger
)

// emailBadge es la etiqueta de estado sobre el título («Turno cancelado»).
type emailBadge struct {
	Label string
	Tone  emailTone
}

// emailDetail es una fila de la tarjeta de datos del turno.
type emailDetail struct {
	Label string
	Value string
}

// emailButton es la llamada a la acción. URL debe ser http(s); si llega vacía
// el botón no se dibuja y el remitente decide qué texto lo sustituye.
type emailButton struct {
	Label string
	URL   string
}

// emailContent es todo lo que un remitente aporta a la plantilla.
type emailContent struct {
	// Preheader es el texto de vista previa de la bandeja; no se muestra en el
	// cuerpo.
	Preheader string
	// Brand es el nombre de la barbería bajo el wordmark, cuando el correo
	// pertenece a una (confirmación y avisos). Vacío en los correos de cuenta.
	Brand      string
	Badge      *emailBadge
	Title      string
	Greeting   string
	Paragraphs []string
	Details    []emailDetail
	// Code es un código de un solo uso, mostrado grande y espaciado.
	Code      string
	Button    *emailButton
	Footnotes []string
}

// emailPalette son los literales de marca. El tono de la insignia se resuelve
// con badgeColors.
const (
	emailInk        = "#101b2b"
	emailCream      = "#f4f0e7"
	emailBrass      = "#b8955a"
	emailBrassText  = "#765c2f"
	emailTextMain   = "#2a2d32"
	emailTextMuted  = "#5e625f"
	emailBorder     = "#c9c0b2"
	emailSurfaceAlt = "#e8e2d8"
)

// badgeColors devuelve (fondo, texto, borde) de la insignia.
func badgeColors(t emailTone) (bg, fg, border template.CSS) {
	switch t {
	case toneSuccess:
		return "#eaf0eb", "#325d43", "#748477"
	case toneWarning:
		return "#f8f1df", "#775019", "#9a6a24"
	case toneDanger:
		return "#f8edec", "#8a2c2c", "#a43a3a"
	default:
		return "#e9eef3", "#23405b", "#667d93"
	}
}

type emailView struct {
	emailContent
	BadgeBg, BadgeFg, BadgeBorder template.CSS
	Ink, Cream, Brass, BrassText  template.CSS
	TextMain, TextMuted, Border   template.CSS
	SurfaceAlt                    template.CSS
	HasGreeting                   bool
}

// emailTemplate se compila una vez al cargar el paquete: una plantilla rota es
// un defecto de programación y debe fallar en las pruebas, no en producción.
var emailTemplate = template.Must(template.New("email").Parse(emailHTML))

// HTML renderiza la versión HTML. Con la plantilla ya compilada solo puede
// fallar si un dato no se puede escribir en el buffer, lo que no ocurre con un
// bytes.Buffer; aun así el error se propaga en vez de ocultarse.
func (c emailContent) HTML() (string, error) {
	v := emailView{
		emailContent: c,
		Ink:          emailInk, Cream: emailCream, Brass: emailBrass, BrassText: emailBrassText,
		TextMain: emailTextMain, TextMuted: emailTextMuted, Border: emailBorder,
		SurfaceAlt:  emailSurfaceAlt,
		HasGreeting: c.Greeting != "",
	}
	if c.Badge != nil {
		v.BadgeBg, v.BadgeFg, v.BadgeBorder = badgeColors(c.Badge.Tone)
	}
	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, v); err != nil {
		return "", fmt.Errorf("notification: renderizar plantilla de correo: %w", err)
	}
	return buf.String(), nil
}

// Text renderiza la versión de texto plano con el mismo contenido y orden.
func (c emailContent) Text() string {
	var b strings.Builder
	if c.Brand != "" {
		b.WriteString(c.Brand)
		b.WriteString("\n\n")
	}
	if c.Badge != nil {
		b.WriteString(strings.ToUpper(c.Badge.Label))
		b.WriteString("\n\n")
	}
	if c.Title != "" {
		b.WriteString(c.Title)
		b.WriteString("\n\n")
	}
	if c.Greeting != "" {
		b.WriteString(c.Greeting)
		b.WriteString("\n\n")
	}
	for _, p := range c.Paragraphs {
		b.WriteString(p)
		b.WriteString("\n\n")
	}
	if c.Code != "" {
		b.WriteString(c.Code)
		b.WriteString("\n\n")
	}
	if len(c.Details) > 0 {
		for _, d := range c.Details {
			fmt.Fprintf(&b, "%s: %s\n", d.Label, d.Value)
		}
		b.WriteString("\n")
	}
	if c.Button != nil && c.Button.URL != "" {
		fmt.Fprintf(&b, "%s: %s\n\n", c.Button.Label, c.Button.URL)
	}
	for _, f := range c.Footnotes {
		b.WriteString(f)
		b.WriteString("\n\n")
	}
	b.WriteString("NAVA · gestión para barberías")
	return b.String()
}

// emailHTML usa el sistema de tablas anidadas habitual en correo. Las clases
// solo existen para la hoja <style> (modo oscuro y móvil estrecho); la versión
// base funciona con los estilos en línea.
const emailHTML = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light dark">
<meta name="supported-color-schemes" content="light dark">
<title>{{.Title}}</title>
<style>
  body { margin:0; padding:0; }
  a { color: inherit; }
  @media (max-width: 480px) {
    .px { padding-left:20px !important; padding-right:20px !important; }
    .code { font-size:28px !important; letter-spacing:6px !important; }
    .title { font-size:22px !important; }
  }
  @media (prefers-color-scheme: dark) {
    .bg { background:#0b1420 !important; }
    .card { background:#16243a !important; border-color:#2a3a52 !important; }
    .ink { color:#f4f0e7 !important; }
    .muted { color:#a8b0bd !important; }
    .detail { background:#0b1420 !important; border-color:#b8955a !important; }
    .code { background:#0b1420 !important; border:1px solid #b8955a !important; }
    .btn { background:#b8955a !important; }
    .btn a { color:#101b2b !important; }
  }
</style>
</head>
<body class="bg" style="margin:0;padding:0;background:{{.Cream}};">
{{- if .Preheader}}
<div style="display:none;max-height:0;overflow:hidden;opacity:0;color:transparent;mso-hide:all;">{{.Preheader}}&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;&#847;&zwnj;&nbsp;</div>
{{- end}}
<main>
<table role="presentation" class="bg" width="100%" cellpadding="0" cellspacing="0" border="0" style="background:{{.Cream}};">
<tr><td align="center" style="padding:24px 12px;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0" style="max-width:600px;">
    <tr><td class="px" style="background:{{.Ink}};padding:26px 32px 22px 32px;border-radius:12px 12px 0 0;">
      <div style="font-family:Georgia,'Times New Roman',serif;font-size:26px;line-height:1;letter-spacing:8px;font-weight:700;color:{{.Cream}};">NAVA</div>
      {{- if .Brand}}
      <div style="margin-top:10px;font-family:-apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;font-size:13px;line-height:1.4;letter-spacing:1.5px;text-transform:uppercase;color:#cbd0d8;">{{.Brand}}</div>
      {{- end}}
    </td></tr>
    <tr><td style="background:{{.Brass}};height:4px;line-height:4px;font-size:4px;">&nbsp;</td></tr>
    <tr><td class="card px" style="background:#ffffff;padding:32px 32px 28px 32px;border:1px solid {{.Border}};border-top:0;border-radius:0 0 12px 12px;font-family:-apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
      {{- if .Badge}}
      <div style="margin:0 0 16px 0;"><span style="display:inline-block;padding:5px 12px;border-radius:999px;background:{{.BadgeBg}};color:{{.BadgeFg}};border:1px solid {{.BadgeBorder}};font-size:12px;line-height:1.3;font-weight:700;letter-spacing:1px;text-transform:uppercase;">{{.Badge.Label}}</span></div>
      {{- end}}
      <h1 class="title ink" style="margin:0 0 18px 0;font-family:Georgia,'Times New Roman',serif;font-size:26px;line-height:1.25;font-weight:700;color:{{.Ink}};">{{.Title}}</h1>
      {{- if .HasGreeting}}
      <p class="ink" style="margin:0 0 14px 0;font-size:16px;line-height:1.6;color:{{.TextMain}};">{{.Greeting}}</p>
      {{- end}}
      {{- range .Paragraphs}}
      <p class="ink" style="margin:0 0 16px 0;font-size:16px;line-height:1.6;color:{{$.TextMain}};">{{.}}</p>
      {{- end}}
      {{- if .Code}}
      <div class="code" style="margin:8px 0 22px 0;padding:18px 12px;border-radius:10px;background:{{.Ink}};color:{{.Cream}};text-align:center;font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,monospace;font-size:34px;line-height:1.2;font-weight:700;letter-spacing:10px;">{{.Code}}</div>
      {{- end}}
      {{- if .Details}}
      <table role="presentation" class="detail" width="100%" cellpadding="0" cellspacing="0" border="0" style="margin:6px 0 22px 0;background:{{.Cream}};border-left:4px solid {{.Brass}};border-radius:0 8px 8px 0;">
        <tr><td style="padding:14px 18px;">
          {{- range $i, $d := .Details}}
          <div class="muted" style="margin:{{if $i}}12px{{else}}0{{end}} 0 2px 0;font-size:12px;line-height:1.3;letter-spacing:1.2px;text-transform:uppercase;font-weight:700;color:{{$.BrassText}};">{{$d.Label}}</div>
          <div class="ink" style="font-size:17px;line-height:1.4;font-weight:600;color:{{$.Ink}};">{{$d.Value}}</div>
          {{- end}}
        </td></tr>
      </table>
      {{- end}}
      {{- if and .Button .Button.URL}}
      <table role="presentation" cellpadding="0" cellspacing="0" border="0" style="margin:6px 0 22px 0;">
        <tr><td class="btn" align="center" style="background:{{.Ink}};border-radius:8px;">
          <a href="{{.Button.URL}}" style="display:inline-block;padding:14px 28px;font-size:16px;line-height:1.2;font-weight:700;color:{{.Cream}};text-decoration:none;">{{.Button.Label}}</a>
        </td></tr>
      </table>
      {{- end}}
      {{- range .Footnotes}}
      <p class="muted" style="margin:0 0 8px 0;font-size:14px;line-height:1.55;color:{{$.TextMuted}};">{{.}}</p>
      {{- end}}
    </td></tr>
    <tr><td class="px" align="center" style="padding:20px 32px 4px 32px;font-family:-apple-system,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;">
      <div class="muted" style="font-size:12px;line-height:1.5;letter-spacing:1.5px;text-transform:uppercase;color:{{.TextMuted}};">NAVA · gestión para barberías</div>
    </td></tr>
  </table>
</td></tr>
</table>
</main>
</body>
</html>
`
