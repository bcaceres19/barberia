---
titulo: "Estándar de diseño visual y experiencia de interfaz"
version: "2.0"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-01"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../04-arquitectura/frontend.md"
  - "especificacion-frontend-nava.md"
  - "estandar-frontend-vue.md"
  - "estrategia-pruebas.md"
---

# Estándar de diseño visual y experiencia de interfaz

## 1. Propósito, alcance y autoridad

Este documento es la fuente de verdad de los fundamentos visuales para el flujo público y el panel del barbero. Define colores, tipografía, medidas, componentes base, estados e interacción. La composición exhaustiva y el inventario de pantallas viven en [especificacion-frontend-nava.md](especificacion-frontend-nava.md); ambos documentos se aplican juntos.

El estándar aplica a toda interfaz en `apps/web`. Una pantalla puede componer los patrones aquí definidos, pero no crear una variante visual nueva sin documentar primero el caso de uso. Las reglas de negocio conservan autoridad sobre el comportamiento; este documento gobierna cómo se presenta ese comportamiento.

Fuentes normativas: `DEC-039` y `DEC-077`. `DEC-077` conserva de `DEC-039` el gobierno por tokens, el enfoque móvil primero, WCAG 2.2 AA, el objetivo táctil de 44 × 44 px y el tema claro único, y sustituye sus elecciones visuales concretas por la identidad NAVA / Tailored Grid.

## 2. Principios del sistema

1. **Móvil primero:** el barbero trabaja principalmente en un celular de gama media, a veces con una sola mano, y el cliente llega desde un enlace. La experiencia de 320–767 px se diseña antes de expandirla.
2. **La tarea domina la decoración:** la agenda, la hora, la persona atendida y la acción siguiente tienen prioridad visual. Los adornos no compiten con ellos.
3. **Una acción principal por región:** cada formulario, diálogo o bloque de decisión destaca una sola acción primaria. Las demás son secundarias o de texto.
4. **Consistencia semántica:** un color representa siempre la misma intención. Tinta NAVA significa acción principal o selección; rojo significa peligro; los estados del turno conservan su mapeo en todo el producto.
5. **Información además de color:** texto, icono, forma o posición acompañan siempre al color en estados, errores y selecciones.
6. **Progreso explícito y errores recuperables:** los flujos indican el paso actual, conservan datos no sensibles y muestran cómo continuar tras un error.
7. **Ligero por defecto:** se usan CSS, dos familias tipográficas self-hosted y componentes propios pequeños. Una biblioteca visual, una fuente adicional o una dependencia de iconos requiere la justificación definida en la arquitectura frontend.
8. **Accesibilidad como criterio de terminado:** el objetivo mínimo es WCAG 2.2 nivel AA; no es una revisión opcional posterior.

## 3. Gobierno de tokens

### 3.1 Fuente única

Cuando exista el frontend, los tokens canónicos vivirán en:

```text
apps/web/src/styles/
  tokens.css       # valores primitivos y semánticos
  base.css         # reset mínimo, tipografía y elementos HTML
```

Los componentes consumen variables semánticas como `--color-action-primary` o `--space-4`; no escriben hexadecimales, sombras, radios o espaciados arbitrarios. Los valores primitivos solo se usan dentro de `tokens.css`.

Orden de decisión:

```text
valor primitivo → token semántico → variante del componente → pantalla
```

No se crean tokens con el nombre de una pantalla (`--agenda-blue`) ni con una apariencia sin intención (`--dark-gray-2`). Los nombres expresan función: `--color-text-muted`, `--color-status-confirmed`.

### 3.2 Excepciones

Una excepción visual requiere en el mismo cambio:

1. necesidad que no cubre un token o componente existente;
2. revisión de contraste, móvil, teclado y estados;
3. nuevo token o variante reutilizable si el patrón puede repetirse;
4. evidencia visual y prueba en la capa adecuada;
5. actualización de este documento si cambia una regla global.

No es una excepción válida “se ve mejor en esta pantalla”.

## 4. Color

### 4.1 Paleta primitiva aprobada

La identidad NAVA combina tinta azul marino, marfil cálido, grafito, latón, salvia y piedra. El latón decorativo aporta calidez editorial; no compite con la tinta como color de acción ni reemplaza colores semánticos.

| Familia | Token | Valor | Uso permitido |
| --- | --- | --- | --- |
| Marca | `nava-ink` | `#101B2B` | Wordmark, texto fuerte, navegación y acción primaria |
| Marca | `nava-ink-hover` | `#18283D` | Acción primaria en hover |
| Marca | `nava-ink-active` | `#0A1420` | Acción primaria activa |
| Base | `nava-ivory` | `#F4F0E7` | Lienzo general cálido |
| Base | `white` | `#FFFFFF` | Controles y superficie de contraste |
| Neutro | `nava-graphite` | `#2A2D32` | Texto principal sobre fondos claros |
| Neutro | `nava-muted` | `#5E625F` | Texto secundario mínimo permitido |
| Neutro | `nava-stone` | `#C9C0B2` | Bordes y superficies de apoyo |
| Neutro | `nava-stone-soft` | `#E8E2D8` | Divisores y fondo secundario |
| Acento | `nava-brass` | `#B8955A` | Fondo o detalle editorial con texto tinta |
| Acento | `nava-brass-dark` | `#765C2F` | Foco, texto/acento funcional y borde seleccionado |
| Acento | `nava-sage` | `#748477` | Apoyo decorativo sobrio; no reemplaza éxito |

### 4.2 Tokens semánticos obligatorios

| Intención | Fondo | Texto / icono | Borde |
| --- | --- | --- | --- |
| Lienzo | `#F4F0E7` | `#2A2D32` | — |
| Superficie | `#FFFFFF` | `#2A2D32` | `#C9C0B2` |
| Superficie tinta | `#101B2B` | `#F4F0E7` | `#101B2B` |
| Texto secundario | — | `#5E625F` | — |
| Control | `#FFFFFF` | `#2A2D32` | `#748477` |
| Acción primaria | `#101B2B` | `#F4F0E7` | `#101B2B` |
| Acción hover | `#18283D` | `#F4F0E7` | `#18283D` |
| Acción activa | `#0A1420` | `#F4F0E7` | `#0A1420` |
| Acción suave / selección | `#F4F0E7` | `#101B2B` | `#765C2F` |
| Foco | transparente | — | `#765C2F` |
| Éxito | `#EAF0EB` | `#325D43` | `#748477` |
| Advertencia | `#F8F1DF` | `#775019` | `#9A6A24` |
| Peligro | `#F8EDEC` | `#8A2C2C` | `#A43A3A` |
| Información | `#E9EEF3` | `#23405B` | `#667D93` |
| Inactivo | `#EEECE8` | `#56514A` | `#C9C0B2` |

Nombres canónicos para la implementación:

```css
:root {
  --color-canvas: #f4f0e7;
  --color-surface: #ffffff;
  --color-surface-muted: #e8e2d8;
  --color-surface-strong: #101b2b;
  --color-text-primary: #2a2d32;
  --color-text-secondary: #5e625f;
  --color-border-subtle: #c9c0b2;
  --color-border-control: #748477;

  --color-action-primary: #101b2b;
  --color-action-primary-hover: #18283d;
  --color-action-primary-active: #0a1420;
  --color-action-soft: #f4f0e7;
  --color-action-soft-border: #765c2f;
  --color-action-soft-active: #e8e2d8;
  --color-focus: #765c2f;
  --color-brand-accent-surface: #b8955a;
  --color-brand-accent-text: #101b2b;
  --color-brand-sage: #748477;

  --color-success-surface: #eaf0eb;
  --color-success-text: #325d43;
  --color-success-border: #748477;
  --color-warning-surface: #f8f1df;
  --color-warning-text: #775019;
  --color-warning-border: #9a6a24;
  --color-danger-surface: #f8edec;
  --color-danger-text: #8a2c2c;
  --color-danger-border: #a43a3a;
  --color-danger-action: #8a2c2c;
  --color-info-surface: #e9eef3;
  --color-info-text: #23405b;
  --color-info-border: #667d93;
  --color-inactive-surface: #eeece8;
  --color-inactive-text: #56514a;
  --color-inactive-border: #c9c0b2;
  --color-on-strong: #f4f0e7;

  --color-status-confirmed-surface: var(--color-info-surface);
  --color-status-confirmed-text: var(--color-info-text);
  --color-status-confirmed-border: var(--color-info-border);
  --color-status-completed-surface: var(--color-success-surface);
  --color-status-completed-text: var(--color-success-text);
  --color-status-completed-border: var(--color-success-border);
  --color-status-cancelled-customer-surface: var(--color-inactive-surface);
  --color-status-cancelled-customer-text: var(--color-inactive-text);
  --color-status-cancelled-customer-border: var(--color-inactive-border);
  --color-status-cancelled-barber-surface: var(--color-danger-surface);
  --color-status-cancelled-barber-text: var(--color-danger-text);
  --color-status-cancelled-barber-border: var(--color-danger-border);
  --color-status-no-show-surface: var(--color-warning-surface);
  --color-status-no-show-text: var(--color-warning-text);
  --color-status-no-show-border: var(--color-warning-border);
}
```

El latón decorativo `#B8955A` no se usa para texto pequeño sobre marfil. Puede funcionar como superficie con tinta NAVA encima. El latón oscuro `#765C2F` se reserva para foco y énfasis secundario; la tinta continúa siendo la acción primaria. Ningún módulo agrega un color de marca propio.

### 4.3 Contraste verificado

Combinaciones mínimas de referencia:

| Combinación | Contraste aproximado |
| --- | ---: |
| tinta `#101B2B` sobre marfil | 15.21:1 |
| marfil sobre tinta `#101B2B` | 15.21:1 |
| grafito `#2A2D32` sobre marfil | 12.15:1 |
| texto secundario `#5E625F` sobre marfil | 5.45:1 |
| latón oscuro `#765C2F` sobre marfil | 5.52:1 |
| tinta sobre latón decorativo `#B8955A` | 6.17:1 |
| texto de confirmado sobre su fondo | 9.20:1 |
| texto de atendido sobre su fondo | 6.53:1 |
| texto de cancelación por cliente sobre su fondo | 6.66:1 |
| texto de cancelación por barbería sobre su fondo | 7.39:1 |
| texto de no asistencia sobre su fondo | 6.33:1 |

Todo texto normal mantiene al menos 4.5:1. Texto grande y señales visuales necesarias para identificar controles o estados mantienen al menos 3:1. No se supone que una combinación cumple por pertenecer a la paleta: se verifica la pareja real de primer plano y fondo.

### 4.4 Estados de turno

| Estado técnico | Etiqueta visible | Color | Señal adicional |
| --- | --- | --- | --- |
| `confirmed` | Confirmada | Información / azul tinta | Icono de calendario con marca y borde izquierdo sólido |
| `completed` | Atendida | Éxito / verde | Icono de verificación |
| `cancelled_by_customer` | Cancelada por cliente | Neutro | Icono de cierre y texto tachado solo en resúmenes compactos |
| `cancelled_by_barber` | Cancelada por la barbería | Peligro / rojo | Icono de cierre y etiqueta completa |
| `no_show` | No asistió | Advertencia / ámbar | Icono de persona ausente |

La interfaz nunca muestra los valores técnicos en inglés. Una insignia de estado incluye texto; un punto de color aislado no es suficiente.

## 5. Tipografía

### 5.1 Familia y pesos

NAVA usa una pareja tipográfica deliberada:

```css
:root {
  --font-display: "Instrument Serif", Georgia, "Times New Roman", serif;
  --font-sans: "Instrument Sans", Inter, system-ui, -apple-system,
    BlinkMacSystemFont, "Segoe UI", sans-serif;
}
```

`Instrument Serif` 400 se usa solo para el wordmark, display y títulos editoriales seleccionados. `Instrument Sans` usa 400 para cuerpo, 500 para controles y 600 para títulos o énfasis; 700 se reserva a cifras o casos excepcionales. No se usa serif en formularios, tablas, navegación o texto largo.

Los archivos se self-hostean como WOFF2 y se precarga únicamente lo necesario para el primer render. El issue que incorpore las fuentes debe documentar licencia, peso de archivos, `font-display`, fallback y medición antes/después. No se cargan desde un CDN ni desde un componente. Hasta integrar esos assets, los fallbacks anteriores son el comportamiento válido.

### 5.2 Escala tipográfica

| Rol | Tamaño / alto de línea | Peso | Uso |
| --- | --- | --- | --- |
| `display` | 48 / 52 px escritorio; 36 / 40 px móvil | Serif 400 | Wordmark o encabezado editorial; nunca formulario operativo |
| `h1` | 32 / 38 px | Sans 600 o serif 400 si la pantalla es editorial | Título principal de pantalla |
| `h2` | 24 / 30 px | 600 | Sección principal |
| `h3` | 19 / 26 px | 600 | Subgrupo o ficha relevante |
| `body-lg` | 18 / 28 px | 400 | Introducción o confirmación |
| `body` | 16 / 24 px | 400 | Texto y formularios públicos |
| `body-sm` | 14 / 20 px | 400 o 500 | Panel operativo y metadatos |
| `caption` | 12 / 16 px | 500 | Etiqueta auxiliar no esencial |

En el flujo público no se reduce el cuerpo por debajo de 16 px. Información necesaria para completar una tarea nunca depende solo de `caption`. Horas, importes y métricas usan cifras tabulares. Mayúsculas completas solo se permiten en el wordmark y microetiquetas de hasta cuatro palabras con tracking; nunca en párrafos, botones o estados.

## 6. Espaciado, tamaño, forma y profundidad

### 6.1 Escala de espaciado

La unidad base es 4 px.

| Token | Valor | Uso habitual |
| --- | ---: | --- |
| `space-0` | 0 | Reinicio explícito |
| `space-1` | 4 px | Separación interna mínima |
| `space-2` | 8 px | Icono con texto, elementos estrechos |
| `space-3` | 12 px | Controles compactos |
| `space-4` | 16 px | Separación y padding móvil normal |
| `space-5` | 20 px | Tarjeta compacta |
| `space-6` | 24 px | Secciones y tarjetas |
| `space-8` | 32 px | Separación de bloques |
| `space-10` | 40 px | Encabezados amplios |
| `space-12` | 48 px | Secciones de página |
| `space-16` | 64 px | Separación máxima habitual |

No se introducen valores como 13, 18, 22 o 30 px para “ajustar” una pantalla. Si un caso real no cabe en la escala, se corrige la composición o se agrega un token global justificado.

### 6.2 Tamaños de control

| Elemento | Altura o área mínima |
| --- | ---: |
| Botón e input estándar | 44 px |
| Botón principal móvil | 48 px |
| Botón de icono | 44 × 44 px |
| Opción de hora | 44 px de alto |
| Navegación inferior | 64 px más área segura del dispositivo |
| Avatar operativo | 32 o 40 px |
| Iconos | 16, 20 o 24 px |

El objetivo táctil del proyecto es 44 × 44 px, superior al mínimo de WCAG 2.2. Un icono puede medir 20 px dentro de un botón de 44 px; el área interactiva no se reduce al dibujo.

### 6.3 Radios, bordes y sombras

| Token | Valor | Uso |
| --- | ---: | --- |
| `radius-sm` | 4 px | Inputs, botones y controles compactos |
| `radius-md` | 6 px | Fichas y superficies |
| `radius-lg` | 8 px | Diálogos y paneles elevados |
| `radius-pill` | 999 px | Insignias; no botones generales |
| Borde normal | 1 px | Controles y agrupaciones |
| Borde de énfasis | 2 px | Selección y error |

Se permiten dos sombras: una sutil para navegación o superficies sticky y otra para diálogos. Las fichas normales se separan con fondo, regla y espacio, no con sombras distintas por módulo. No se usan desenfoques decorativos, neomorfismo ni vidrio translúcido.

```css
:root {
  --shadow-raised: 0 1px 2px rgb(16 27 43 / 8%), 0 4px 12px rgb(16 27 43 / 6%);
  --shadow-dialog: 0 16px 40px rgb(16 27 43 / 18%);
  --layer-base: 0;
  --layer-sticky: 10;
  --layer-menu: 20;
  --layer-overlay: 30;
  --layer-dialog: 40;
  --layer-toast: 50;

  /* Superposiciones translúcidas derivadas de nava-ink; nunca un color
     nuevo. --color-overlay-scrim es el fondo no interactivo detrás de un
     diálogo; --color-overlay-hover/--color-overlay-active son el realce
     de un control transparente (cerrar, descartar) en hover y active. */
  --color-overlay-scrim: rgb(16 27 43 / 48%);
  --color-overlay-hover: rgb(16 27 43 / 8%);
  --color-overlay-active: rgb(16 27 43 / 16%);
}
```

Ningún componente inventa un `z-index`; consume una capa semántica y un diálogo no puede aparecer debajo de una barra fija.

### 6.4 Movimiento

Toda transición o animación que explica un cambio de estado usa una de dos duraciones; ningún componente escribe un número de milisegundos suelto:

```css
:root {
  --motion-duration-fast: 120ms;
  --motion-duration-base: 200ms;
  --motion-easing-standard: ease;
}
```

Un indicador de progreso indeterminado (por ejemplo, el giro continuo de un botón en estado de carga) no es una transición de estado y no usa estos tokens: señala trabajo en curso de duración desconocida. Igual se apaga o reduce con `prefers-reduced-motion` (§12).

## 7. Composición responsiva

### 7.1 Anchos de referencia

Los componentes responden al espacio disponible; los puntos de referencia no representan modelos de dispositivo específicos.

| Rango | Composición |
| --- | --- |
| 320–767 px | Una columna, navegación móvil, padding lateral de 16 px |
| 768–1023 px | Una o dos columnas según contenido, padding de 24 px |
| 1024–1279 px | Panel NAVA con dock inferior, contenido de 8–12 columnas |
| 1280 px o más | Contenedor centrado; no se estira texto o formulario sin límite |

Anchos máximos:

- contenido general privado: 1440 px;
- reserva pública: 720 px;
- formulario legible: 640 px;
- texto largo: 72 caracteres aproximadamente.

Se prueba reflow sin pérdida funcional a 320 CSS px. Solo la representación espacial de una agenda puede requerir desplazamiento horizontal; siempre ofrece una vista de lista equivalente en móvil.

### 7.2 Estructuras base

Flujo público:

```text
┌──────────────────────────────┐
│ Nombre de barbería           │
│ Reservas con NAVA            │
├──────────────────────────────┤
│ Paso actual y progreso       │
│ Título + ayuda breve         │
│                              │
│ Contenido o formulario       │
│ Estado vacío/error contextual│
│                              │
├──────────────────────────────┤
│ Atrás       Acción principal │
└──────────────────────────────┘
```

Panel privado en escritorio:

```text
┌───────────────────────────────────────────────┐
│ NAVA · barbería       Contexto · Nuevo turno │
├───────────────────────────────────────────────┤
│ Título + acción local                         │
│ Filtros o resumen autorizado                  │
│ Contenido principal en grid                   │
├───────────────────────────────────────────────┤
│ Agenda · Servicios · Barberos · Horarios · … │
└───────────────────────────────────────────────┘
```

Panel privado en móvil:

```text
┌──────────────────────────────┐
│ Barra superior + fecha       │
├──────────────────────────────┤
│ Título + acción principal    │
│ Contenido en lista           │
│                              │
├──────────────────────────────┤
│ Agenda · Nuevo · Horarios · Más │
└──────────────────────────────┘
```

La navegación inferior muestra icono y texto. “Nuevo” inicia la creación manual; no se representa solo con un símbolo ambiguo. La composición completa del shell, destinos permitidos y variaciones están en [especificacion-frontend-nava.md](especificacion-frontend-nava.md) §5.

## 8. Componentes base

### 8.1 Botones y enlaces

Variantes permitidas:

| Variante | Uso |
| --- | --- |
| Primario | Fondo `action-primary`, texto `on-strong`; una acción principal por formulario o región |
| Secundario | Fondo de superficie, texto principal y borde de control; alternativa válida sin prioridad |
| Texto | Sin superficie, texto `action-primary`; navegación o acción de baja jerarquía |
| Peligro | Fondo `danger-action`, texto `on-strong`; acción destructiva confirmada, nunca navegación normal |

Todos cubren `default`, `hover`, `active`, `focus-visible`, `loading` y `disabled`. En carga mantienen el ancho, cambian el texto a una acción explícita como “Confirmando…” y bloquean envíos repetidos sin sustituir la idempotencia del servidor.

Los enlaces dentro de texto están subrayados. No se hace parecer enlace a un botón ni botón a texto estático.

### 8.2 Formularios

Orden de un campo:

```text
Etiqueta
Ayuda opcional
[ control                                   ]
Mensaje de error o confirmación
```

Reglas:

- etiqueta visible encima del control; `placeholder` nunca sustituye la etiqueta;
- texto “Opcional” cuando aplique, en vez de llenar el formulario de asteriscos;
- error junto al campo, con icono, texto y relación accesible;
- resumen de errores solo cuando hay varios y mueve el foco al fallar el envío;
- formato y ejemplo visibles antes del error cuando evitan ambigüedad;
- selección con tarjetas solo si cada opción necesita más que una etiqueta;
- fecha y hora se muestran en la zona de la barbería y la zona aparece junto al selector;
- los datos no sensibles se conservan ante error recuperable o conflicto de horario.

### 8.3 Tarjetas, listas y tablas

- Una tarjeta agrupa una unidad accionable; no envuelve cada párrafo.
- En móvil, los turnos y configuraciones se presentan como lista de tarjetas o filas de al menos 56 px.
- Las tablas quedan para comparaciones de varias columnas en escritorio y se transforman en filas etiquetadas en móvil.
- Acciones repetidas se agrupan en un menú con nombre accesible; la acción primaria frecuente permanece visible.
- Una fila completa puede ser interactiva si tiene foco, rol y nombre correctos; no contiene otra zona clicable que compita sin necesidad.

### 8.4 Insignias, alertas y mensajes

- Una insignia etiqueta un estado; no funciona como botón salvo que se implemente expresamente como filtro.
- Una alerta en línea conserva contexto y puede contener una acción de recuperación.
- Los errores de campo no se muestran como notificación temporal.
- Las confirmaciones no críticas pueden usar un aviso temporal, pero la nueva situación debe quedar visible en la pantalla.
- Los mensajes críticos no desaparecen solos.
- El error inesperado muestra una explicación segura, “Reintentar” cuando aplique y el `request_id` copiable para soporte.

### 8.5 Diálogos y paneles

- Diálogo para una decisión breve; página o flujo dedicado para más de una etapa.
- Título descriptivo, consecuencia, acción primaria y cancelación visible.
- Una acción destructiva nombra el objeto: “Cancelar turno”, no “Aceptar”.
- El foco entra, queda contenido, vuelve al disparador y nunca queda detrás de una capa.
- En móvil puede presentarse como panel inferior accesible; conserva la misma semántica y no depende de arrastrar.

### 8.6 Iconos

Se usa un conjunto pequeño de SVG de línea, con caja de 24 px y trazo de 2 px. Los SVG viven bajo un dueño claro en `shared/ui`; no se copian variantes entre módulos. Iconos decorativos se ocultan del árbol accesible; iconos que actúan como control reciben nombre mediante su botón.

No se usan emoji como iconografía de producto. Incorporar una dependencia de iconos exige revisar tamaño, licencia y alternativa conforme al estándar frontend.

## 9. Patrones de estado de una pantalla

Toda pantalla o región asíncrona contempla, cuando aplique:

| Estado | Patrón |
| --- | --- |
| Inicial | Contenido listo o instrucción concreta; no una superficie vacía |
| Carga inicial | Esqueleto con la geometría aproximada; no bloqueador a pantalla completa |
| Actualización | Contenido anterior permanece con indicador local |
| Vacío | Motivo, siguiente acción y sin ilustración dominante |
| Error recuperable | Qué ocurrió en lenguaje simple, datos conservados y “Reintentar” |
| Error de campo | Junto al control y resumido si hay varios |
| Conflicto | Explicación y alternativas; no código técnico |
| Éxito | Resultado persistente, resumen y siguiente acción |

Los esqueletos no simulan contenido que la pantalla nunca tendrá y respetan reducción de movimiento.

## 10. Plantillas de las pantallas P0

Esta tabla es un índice compacto. La especificación obligatoria, estados, responsive y frontera P0/P1/P2 de cada familia viven en [especificacion-frontend-nava.md](especificacion-frontend-nava.md) §7.

| Pantalla | Composición obligatoria | Acción principal |
| --- | --- | --- |
| Acceso | Marca discreta, título, formulario estrecho, recuperación visible | Iniciar sesión |
| Recuperación | Paso actual, destino enmascarado, código, reenvío con estado | Verificar código |
| Agenda diaria | Selector obligatorio de un barbero, fecha y navegación, resumen autorizado, lista cronológica y estado por turno | Nuevo turno |
| Detalle de turno | Hora, persona atendida, servicio, contacto, estado, historial y acciones permitidas | Acción contextual según estado |
| Crear/editar turno | Secciones cortas en orden: servicio, persona, fecha/hora, contacto, resumen | Guardar turno |
| Servicios | Lista de nombre, duración, precio y estado; turnos futuros visibles al desactivar | Nuevo servicio |
| Barberos | Lista por persona con servicios y horario asociados | Añadir barbero |
| Horarios y bloqueos | Selector de barbero, semana base, excepciones y bloqueos próximos | Crear bloqueo |
| Configuración | Secciones separadas: barbería, reserva, cancelación, recordatorios y canales | Guardar sección |
| Reserva pública | Progreso, una decisión principal por paso, resumen persistente | Continuar / Confirmar turno |
| Confirmación pública | Estado de éxito, fecha/hora/zona, servicio, barbero, acceso enviado y siguiente acción | Gestionar este turno |
| Gestión pública | Resumen del turno, política vigente y acción permitida | Cancelar turno, si procede |

Una pantalla de configuración no mezcla todas las secciones en un formulario con un único “Guardar”. Cada sección guarda y comunica su estado de manera independiente.

## 11. Flujos prioritarios

### 11.1 Reserva pública

Secuencia canónica:

1. **Servicio y barbero:** se elige servicio; el selector de barbero aparece cuando hay varios. Con uno, queda preseleccionado y se muestra en el resumen sin agregar una decisión.
2. **Fecha y hora:** días disponibles primero, luego franjas de al menos 44 px. Se muestra “Hora de la barbería · America/Bogota” o la zona configurada.
3. **Tus datos:** nombre, teléfono, correo, nota opcional y control “Reservo para otra persona” que revela el nombre adicional.
4. **Revisión:** resumen completo, política relevante y confirmación explícita.
5. **Resultado:** confirmación inmediata, sin crear cuenta, con instrucciones sobre el enlace enviado.

El indicador recalcula el total de pasos si una futura variante agrega o elimina una pantalla. No se muestra un paso vacío para un único barbero. Volver atrás conserva la selección mientras siga disponible.

Si la franja se ocupa al confirmar, el error queda en la región de fecha y hora, conserva los datos escritos y presenta alternativas cercanas según `RN-CON-05`.

### 11.2 Agenda diaria

- La pantalla abre en hoy y muestra la fecha completa, no solo “Hoy”.
- Exige un barbero seleccionado y no ofrece una vista consolidada inicial, según `DEC-074`.
- Anterior, selector de fecha y siguiente conservan posiciones estables.
- En móvil se usa una lista cronológica; en escritorio puede añadirse una línea temporal, sin quitar la vista de lista accesible.
- Cada turno muestra en este orden: hora, persona atendida, servicio, barbero cuando aporte contexto y estado.
- El detalle contiene los datos de contacto; la lista evita exponerlos de forma innecesaria.
- “Ahora” señala la posición temporal; no crea el estado `in_progress`. Los conflictos se señalan con texto, borde e icono, no solo con fondo.
- Los huecos no se convierten en decenas de tarjetas vacías. “Nuevo turno” permite iniciar desde una hora elegida.
- Los estados terminales se separan o atenúan sin hacer ilegible el historial del día.

### 11.3 Bloqueos con turnos afectados

Tras crear el bloqueo, se confirma primero que el bloqueo existe. Si afecta turnos, aparece una lista con cantidad, hora, persona y decisiones individuales. Mantener es la opción inicial; reprogramar o cancelar requiere acción explícita. Nunca se presenta una acción masiva preseleccionada que contradiga `RN-BLQ-03`.

### 11.4 Acciones destructivas o de alto impacto

Cancelar, desactivar o aplicar cambios a turnos existentes usa:

1. objeto afectado y cantidad exacta;
2. consecuencia en lenguaje directo;
3. alternativas permitidas por la regla;
4. selección explícita, sin casillas peligrosas preactivadas;
5. confirmación con verbo y objeto;
6. resultado y notificaciones que se generarán.

## 12. Interacción, foco y movimiento

- El foco visible usa un anillo exterior de 2 px `var(--color-focus)` con separación de 2 px. En fondo tinta se añade una separación marfil para conservar contraste.
- `hover` nunca es el único modo de revelar una acción necesaria.
- Orden de tabulación coincide con el orden visual y semántico.
- Cambiar un selector no navega ni guarda sin aviso; las acciones ocurren por solicitud explícita.
- La aplicación respeta `prefers-reduced-motion`.
- Transiciones opcionales duran 120–200 ms y solo explican cambio de estado o jerarquía. No hay animaciones continuas, parallax ni entradas que retrasen la tarea.
- Ninguna tarea depende de arrastrar; siempre existe una acción equivalente por toque o teclado.
- Una barra fija no oculta el foco ni el último contenido y considera las áreas seguras del dispositivo.

## 13. Lenguaje visual y contenido

- La interfaz usa **turno**; API, datos y código usan `appointment`, según `DEC-016`.
- Títulos describen la tarea: “Elige una hora”, no “Disponibilidad”.
- Botones usan verbo y resultado: “Confirmar turno”, “Guardar horario”, “Reintentar”.
- Los errores explican qué ocurrió y qué puede hacer la persona; nunca muestran nombres de excepciones o proveedores.
- Fechas públicas usan formato legible en español, por ejemplo “jueves, 14 de agosto”; formularios pueden acompañarlo con formato corto.
- Las horas usan una convención única por configuración regional y siempre incluyen minutos. No se mezclan “9 AM”, “09:00” y “9:00 a. m.” en una misma experiencia.
- No se usan imágenes de stock como relleno. Logotipo o ilustración futura declara dimensiones, texto alternativo y versión adecuada para fondo claro.

## 14. Tema y personalización de barbería

El MVP tiene **un único tema claro NAVA**. NAVA identifica la plataforma; cada barbería muestra su nombre como contexto. Todas comparten tokens y componentes y no pueden escoger colores, fuentes, radios o estilos por pantalla. Un logotipo propio solo se muestra cuando una capacidad futura defina carga, almacenamiento, accesibilidad y fallback; no se inventa ni se exige en P0.

No forman parte del MVP:

- modo oscuro;
- selector de tema;
- colores libres por tenant;
- CSS personalizado;
- tipografías por barbería;
- variantes de componentes por marca.

Una futura personalización solo podrá sobrescribir un conjunto limitado de tokens semánticos, con contraste validado y valores de respaldo. Nunca permitirá CSS arbitrario ni cambiar los colores de éxito, advertencia o peligro.

## 15. Implementación futura en Vue

1. `tokens.css` se importa una sola vez desde el arranque de la aplicación.
2. Las fuentes NAVA, una vez aprobadas en un issue, se declaran una sola vez en estilos globales con WOFF2 self-hosted; ningún componente carga una fuente remota.
3. `BaseButton`, `BaseInput`, `BaseSelect`, `BaseDialog`, `BaseAlert`, `BaseBadge` y `BaseIconButton` implementan las variantes compartidas cuando exista uso real; no se crean todos por anticipado.
4. Los componentes de negocio —por ejemplo `AppointmentStatusBadge`— viven en su módulo y componen primitivas de `shared/ui`.
5. Props describen intención (`tone="danger"`), no valores libres (`color="#8a2c2c"`).
6. Se prefieren CSS Grid y Flexbox; no se posiciona la estructura principal con coordenadas absolutas.
7. No se usan estilos globales para alcanzar internos de otro componente.
8. Los estilos locales no redefinen tipografía, controles o estados ya gobernados por un componente base.
9. La agenda puede usar cálculo visual específico, pero la autoridad de disponibilidad permanece en el backend.

Ejemplo de consumo permitido:

```css
.booking-card {
  padding: var(--space-6);
  color: var(--color-text-primary);
  background: var(--color-surface);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
}
```

No se copian los valores hexadecimales o numéricos detrás de esas variables al componente.

## 16. Verificación visual y accesible

Todo cambio visible aporta evidencia proporcional al riesgo:

- capturas de antes/después o del nuevo estado en 360 px y 1280 px;
- 320 px para verificar reflow y 768 px cuando cambie la composición;
- estados normal, foco, error, carga, vacío, éxito e inactivo afectados;
- zoom de texto al 200 % sin pérdida de contenido;
- teclado completo y foco no oculto;
- lector semántico en formularios, diálogos y avisos críticos;
- contraste real de combinaciones nuevas;
- tema claro con nombre o logotipo corto y largo de prueba;
- wordmark NAVA y nombre de barbería con longitudes extremas sin solaparse;
- teléfono real del piloto antes de una versión.

Las pruebas de componente validan nombre, rol, estado y comportamiento. Las E2E cubren el recorrido P0 afectado. Una captura no sustituye esas pruebas y un snapshot no se actualiza sin revisar el cambio visual.

## 17. Lista de revisión

- [ ] La pantalla usa tokens y componentes existentes; no contiene colores o tamaños arbitrarios.
- [ ] Hay una jerarquía clara y una sola acción primaria por región.
- [ ] Funciona desde 320 px sin pérdida, salvo la excepción espacial documentada de agenda con alternativa en lista.
- [ ] Controles táctiles alcanzan 44 × 44 px y no quedan cubiertos por barras fijas.
- [ ] Texto, controles y foco cumplen los contrastes mínimos.
- [ ] Color no es la única señal de estado, selección, conflicto o error.
- [ ] Carga, vacío, error, conflicto y éxito están diseñados donde aplican.
- [ ] Formularios conservan datos no sensibles y asocian etiquetas y errores.
- [ ] Horas muestran la zona de la barbería y la interfaz usa “turno”.
- [ ] El flujo no agrega pasos, campos o decisiones fuera del alcance P0.
- [ ] Teclado, reducción de movimiento y zoom funcionan.
- [ ] La evidencia visual y las pruebas aplicables acompañan el cambio.

## 18. Referencias oficiales

- [WCAG 2.2](https://www.w3.org/TR/WCAG22/)
- [Contraste mínimo de texto](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html)
- [Contraste no textual](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html)
- [Reflow a 320 CSS px](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html)
- [Tamaño mínimo del objetivo](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html)
- [Foco visible](https://www.w3.org/WAI/WCAG22/Understanding/focus-visible.html)
