---
prompt_id: "PROMPT-AUDIT-UI-COMPONENTES-ALERTAS-NAVA-v1"
version: "1.0"
kind: "audit"
status: "executed"
target_agents:
  - "codex"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-009"
related_hu:
  - "HU-010"
  - "HU-011"
  - "HU-012"
  - "HU-020"
  - "HU-022"
  - "HU-061"
issue: null
issue_url: null
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "DEC-039"
  - "DEC-077"
rules:
  - "RN-DAT-02"
decisions:
  - "DEC-039"
  - "DEC-077"
acceptance_criteria:
  - "Existe un inventario visual de botones, inputs, tipografía, navegación, modales y alertas en estados inicial, error y éxito."
  - "Cada captura se puede abrir desde el repositorio y no contiene secretos ni datos personales reales."
  - "Las observaciones distinguen evidencia visible, hipótesis de diseño y defectos funcionales confirmados."
  - "La auditoría no modifica código, contrato, base de datos ni tokens visuales."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/10-backlog/prompts/hu/hu-009-sistema-visual.md"
  - "docs/10-backlog/prompts/README.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Auditoría visual de componentes y alertas NAVA

## Instrucción para el agente

Realiza una auditoría visual de referencia sobre la interfaz local de NAVA. Captura estados representativos y separa con claridad lo observado en pantalla de cualquier propuesta de rediseño. No implementes correcciones en esta auditoría.

## Resultado ejecutado

Se inspeccionaron, mediante el navegador integrado, el acceso, recuperación, panel, nuevo turno, servicios y configuración. Las capturas se tomaron con datos sintéticos del tenant QA; no contienen credenciales, códigos, teléfonos ni correos reales.

## Evidencia visual

| Captura | Estado observado |
| --- | --- |
| [`01-nuevo-turno-alertas.png`](../../evidence/ui-visual-2026-09-02/01-nuevo-turno-alertas.png) | Formulario de nuevo turno con errores de campos y alerta resumen. |
| [`02-servicios-lista.png`](../../evidence/ui-visual-2026-09-02/02-servicios-lista.png) | Lista de servicios con estado activo, acción primaria y acción destructiva. |
| [`03-servicio-modal.png`](../../evidence/ui-visual-2026-09-02/03-servicio-modal.png) | Modal de alta con inputs, helper text y botones secundarios/primarios. |
| [`04-servicio-modal-validacion.png`](../../evidence/ui-visual-2026-09-02/04-servicio-modal-validacion.png) | Modal con validación inline de nombre y precio. |
| [`05-acceso-inicial.png`](../../evidence/ui-visual-2026-09-02/05-acceso-inicial.png) | Acceso inicial con marca, título, inputs, CTA y enlace de recuperación. |
| [`06-acceso-validacion.png`](../../evidence/ui-visual-2026-09-02/06-acceso-validacion.png) | Acceso con correo inválido y contraseña enmascarada. |
| [`07-recuperacion-inicial.png`](../../evidence/ui-visual-2026-09-02/07-recuperacion-inicial.png) | Paso inicial de recuperación; el copy actual menciona ambos canales. |
| [`08-configuracion-inicial.png`](../../evidence/ui-visual-2026-09-02/08-configuracion-inicial.png) | Configuración con campos requeridos, opcionales y ayudas. |
| [`09-configuracion-exito.png`](../../evidence/ui-visual-2026-09-02/09-configuracion-exito.png) | Estado exitoso de guardado con alerta verde. |

## Hallazgos de referencia para el futuro diseño

Estas observaciones describen oportunidades visuales, no defectos funcionales confirmados:

- **Jerarquía de botones:** el CTA oscuro aparece como botón compacto en cabecera, ancho completo en formularios y compacto en modales. Conviene definir una escala única de tamaños, prioridad y ancho por contexto.
- **Inputs:** los campos comparten borde gris y fondo blanco, pero el ritmo entre etiqueta, control, ayuda y error cambia según la pantalla. Conviene definir una anatomía única de campo y una altura/token común.
- **Tipografía:** la marca serif y el texto de interfaz sans se distinguen, pero la relación entre títulos, etiquetas, helper text y errores se percibe poco graduada. Conviene fijar una escala tipográfica y pesos explícitos para display, heading, body, label, help y error.
- **Alertas:** los errores inline aparecen debajo de cada campo, mientras el resumen usa una caja rosada con icono y el éxito una caja verde con icono. Conviene unificar iconografía, padding, bordes, contraste, rol semántico y distancia vertical.
- **Acciones destructivas:** “Desactivar” usa rojo sólido y “Cancelar” usa contorno, pero falta una regla visual documentada para diferenciar peligro, secundario y primario en todos los módulos.
- **Modal:** el overlay, cierre, encabezado, formulario y acciones tienen una composición clara, pero necesitan tokens compartidos para ancho, padding, radios y separación entre controles.
- **Shell y navegación:** la navegación inferior contiene seis entradas y la última se trunca visualmente en el ancho observado. Debe decidirse si se conserva como navegación desplazable o se rediseña para garantizar legibilidad y equilibrio.
- **Copy visible:** la pantalla de recuperación todavía anuncia envío por WhatsApp y correo porque esa es la política vigente. El cambio solicitado está registrado aparte como `DP-NOT-06`/`CT-009`; esta auditoría no lo modifica.

## Límites

- No se concluye incumplimiento WCAG únicamente por apariencia; contraste, foco, teclado, zoom y lector de pantalla requieren pruebas específicas.
- No se concluye un bug de layout para todos los anchos a partir de estas capturas; la campaña responsive debe verificar 320, 360, 768, 1280 y zoom 200 %.
- No se cambian tokens, componentes, copy, contrato, base de datos ni lógica.

## Siguiente trabajo recomendado

Crear un issue de diseño que use esta auditoría como evidencia, defina tokens y variantes de `Button`, `Field`, `Alert`, `Modal` y navegación, y después aplique las pantallas por lotes independientes con pruebas de componente, accesibilidad y responsive.
