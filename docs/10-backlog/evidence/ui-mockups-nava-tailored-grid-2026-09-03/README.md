# Atlas NAVA / Tailored Grid por viewport y evento

Este directorio es la entrada canónica a los mockups raster del producto. Cada
subdirectorio contiene un `README.md` con la matriz de estados, alcance,
decisiones visuales, exclusiones y modo de uso. Los PNG son evidencia de diseño:
no crean funciones y ceden ante HU, reglas, decisiones, contratos, seguridad,
privacidad y accesibilidad.

## Inventario

| Ruta o familia | Issue | PNG | Cobertura |
| --- | ---: | ---: | --- |
| [`auth-eventos`](auth-eventos/README.md) | #188 | 46 | Acceso y recuperación. |
| [`panel-agenda-eventos`](panel-agenda-eventos/README.md) | #189 | 24 | Agenda diaria y navegación por fecha. |
| [`nuevo-turno-eventos`](nuevo-turno-eventos/README.md) | #190 | 24 | Creación manual de turnos. |
| [`detalle-turno-eventos`](detalle-turno-eventos/README.md) | #191 | 32 | Detalle, historial y reprogramación. |
| [`barberos-eventos`](barberos-eventos/README.md) | #192 | 47 | Listado, alta y cambio de nombre. |
| [`barberia-eventos`](barberia-eventos/README.md) | #193 | 27 | Configuración de la barbería. |
| [`servicios-eventos`](servicios-eventos/README.md) | #194 | 79 | Catálogo y ciclo de vida de servicios. |
| [`servicios-por-barbero-eventos`](servicios-por-barbero-eventos/README.md) | #195 | 37 | Asignación de servicios a barberos. |
| [`horarios-eventos`](horarios-eventos/README.md) | #196 | 35 | Horarios, excepciones y festivos. |
| [`bloqueos-eventos`](bloqueos-eventos/README.md) | #197 | 33 | Bloqueos puntuales y series. |
| **Total** |  | **384** | Diez familias con pares responsive y estados materiales. |

## Uso

1. Lee el `README.md` de la familia antes de abrir imágenes.
2. Selecciona el archivo del evento y viewport exactos; no uses una lámina
   compuesta como sustituto cuando exista un PNG individual.
3. Si un issue o prompt asigna ese PNG, activa fidelidad medible y compara la
   app real en el mismo viewport con capturas, lado a lado y overlay o diff.
4. Si no existe referencia exacta para un estado, conserva identidad NAVA /
   Tailored Grid y aplica el modo de identidad guiada del estándar visual.
5. Los datos representados son sintéticos. No copies nombres, contactos,
   precios o identificadores de una imagen como datos reales o fixtures.

## Integridad del paquete

- El inventario se valida contando PNG reales, no por la cantidad declarada en
  este archivo.
- Cada imagen debe tener firma PNG válida, dimensiones positivas y tamaño menor
  al límite individual de GitHub.
- Los archivos fuente o generadores solo se versionan cuando ya forman parte de
  la procedencia documentada por la familia; no son dependencias de runtime.
