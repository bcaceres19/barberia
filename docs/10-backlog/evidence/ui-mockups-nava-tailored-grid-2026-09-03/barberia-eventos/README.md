# Atlas visual de `/panel/barberia`

Paquete raster de diseño para el issue #193 y `HU-020`. Representa únicamente
la configuración básica real de la barbería activa: nombre, zona horaria IANA,
correo opcional, teléfono opcional y los estados de lectura/guardado que ya
expone la pantalla. Los PNG no implementan la ruta ni prueban persistencia,
aislamiento, accesibilidad o comportamiento funcional.

## Viewports y exportaciones

| Objetivo de composición | Exportación | Archivo principal |
| --- | --- | --- |
| 320 × 800 CSS px | 800 × 2000 px | `mobile/barberia/03a-principal-320.png` |
| 360 × 800 CSS px | 840 × 1870 px | `mobile/barberia/03-principal-sin-cambios.png` |
| 768 × 1024 CSS px | 1080 × 1440 px | `tablet/barberia/03-principal-768.png` |
| 1280 × 1024 CSS px | 1280 × 1024 px | `desktop/barberia/03a-principal-1280.png` |

Los estados secundarios mantienen el formato del atlas NAVA ya existente:
`desktop/barberia/` a 1440 × 1024 px y `mobile/barberia/` a 840 × 1870 px.
La variante principal exacta de 1280 × 1024 queda separada para no confundirla
con la serie de escritorio usada para comparar todos los eventos.

## Matriz de eventos

| Nº | Evento | Escritorio | Móvil | Revisión visual |
| --- | --- | --- | --- | --- |
| 01 | Carga inicial | `desktop/barberia/01-carga-inicial.png` | `mobile/barberia/01-carga-inicial.png` | `PageState` de espera idéntico al resto del atlas; sin acción disponible. |
| 02 | Error recuperable de carga | `desktop/barberia/02-error-carga.png` | `mobile/barberia/02-error-carga.png` | Conserva contexto y muestra `Reintentar`. |
| 03 | Principal sin cambios | `desktop/barberia/03-principal-sin-cambios.png` | `mobile/barberia/03-principal-sin-cambios.png` | Cuatro campos, hints y navegación completos. |
| 04 | Contactos opcionales ausentes | `desktop/barberia/04-contactos-vacios.png` | `mobile/barberia/04-contactos-vacios.png` | Vacío distinguible de un valor guardado y de `null`. |
| 05 | Cambios pendientes y foco visible | `desktop/barberia/05-cambios-pendientes-foco.png` | `mobile/barberia/05-cambios-pendientes-foco.png` | Foco por forma/contorno y acción habilitada. |
| 06 | Validación de nombre, zona, correo y teléfono | `desktop/barberia/06-validacion-campos.png` | `mobile/barberia/06-validacion-campos.png` | Cada error tiene texto y cambio de borde/superficie; no depende sólo del rojo. |
| 07 | Guardado en curso | `desktop/barberia/07-guardando.png` | `mobile/barberia/07-guardando.png` | Campos y acción bloqueados; estado textual explícito. |
| 08 | Zona IANA rechazada por el servidor | `desktop/barberia/08-zona-iana-no-reconocida.png` | `mobile/barberia/08-zona-iana-no-reconocida.png` | Conserva `COT` y explica `America/Bogota` como ejemplo. |
| 09 | Error de red al guardar | `desktop/barberia/09-error-red.png` | `mobile/barberia/09-error-red.png` | Conserva nombre y correo escritos; permite reintento. |
| 10 | Error inesperado al guardar | `desktop/barberia/10-error-inesperado.png` | `mobile/barberia/10-error-inesperado.png` | Conserva datos y usa mensaje recuperable. |
| 11 | Guardado confirmado | `desktop/barberia/11-guardado-confirmado.png` | `mobile/barberia/11-guardado-confirmado.png` | Copy vigente y nombre confirmado actualizado en cabecera. |
| 12 | Nombre largo y reflow | `desktop/barberia/12-nombre-largo-reflow.png` | `mobile/barberia/12-nombre-largo-reflow.png` | Cabecera truncada de forma controlada; campo sin solape. |

## Contrato visual

- Cascarón y proporciones compartidos literalmente con Agenda, Nuevo turno y
  Detalle: cabecera NAVA plana, contenido desde el mismo margen y dock inferior.
- Firma NAVA: tinta `#101B2B`, marfil `#F4F0E7`, latón `#B8955A`, líneas finas,
  voz editorial serif y controles sans-serif.
- Las alertas reutilizan las superficies, filetes y rótulos semánticos del atlas:
  peligro, advertencia y confirmación no se reinterpretan por ruta.
- En móvil, `Más` aparece activo porque Configuración es un destino secundario
  del cascarón real. En escritorio se activa `Configuración` directamente.
- La zona horaria se mantiene como entrada textual IANA. Se retiró el chevrón
  exploratorio de las primeras variantes porque el contrato actual no ofrece
  catálogo ni selector de zonas.
- Los datos son ficticios y usan el dominio reservado `.example`; no proceden
  de producción ni corresponden a personas reales.

## Fuentes

- `docs/02-requisitos/historias-usuario.md`, `HU-020` y `CA-020-01`–`08`.
- `docs/01-producto/reglas-negocio.md`: `RN-DIS-07`, `RN-TEN-01`, `RN-DAT-02`.
- `docs/00-control/registro-decisiones.md`: `DEC-007`, `DEC-024`, `DEC-033`,
  `DEC-037`, `DEC-077`–`DEC-080`.
- `docs/03-desarrollo/estandar-diseno-visual.md` y
  `docs/03-desarrollo/especificacion-frontend-nava.md`.
- `apps/web/src/modules/settings/pages/SettingsPage.vue`, sus validadores y sus
  pruebas de componente.
- `../panel-agenda-eventos/`, `../nuevo-turno-eventos/`,
  `../detalle-turno-eventos/` y `../barberos-eventos/` como referencias del
  sistema común. No se copiaron sus funciones.

## Exclusiones

No se representan logo de tenant, enlace público, tema o colores configurables,
políticas, horarios, recordatorios, canales OTP/notificación, proveedores,
usuarios, permisos, eliminación de barbería ni otra capacidad futura. El atlas
no modifica API, base de datos, frontend o datos.

## Procedencia

Rasterizado con Chromium y las mismas fuentes autoalojadas, tokens, shell,
`PageState`, `BaseAlert`, campos, botones y navegación que los atlas existentes.
No se usó interpretación generativa para color, alertas o componentes. Cada
variante aplica una sola diferencia de estado sobre el principal: carga, error
de carga, contactos vacíos, edición/foco, validación, guardando, rechazo de zona
IANA, error de red, error inesperado, éxito o nombre largo.

## Revisión y estado del paquete

Se abrieron muestras representativas de principal, `PageState`, validación,
alerta y diálogo en ambos viewports. Los 27 PNG pueden decodificarse y coinciden
con las dimensiones declaradas. Esta revisión es visual del artefacto, no QA de
la aplicación ni certificación WCAG.

El atlas raster queda disponible como avance solicitado. El prompt Penpot de
#193 permanece `blocked`: #192 sigue abierto y no existe todavía fuente editable,
rama ni PR de este paquete. Por ello estos PNG no deben marcar #193 como cerrado.
