---
titulo: "Mapa documental y fuentes de verdad"
version: "2.1"
estado: "Vigente"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-12"
documentos_relacionados:
  - "00-control/registro-decisiones.md"
  - "00-control/contradicciones.md"
  - "00-control/matriz-trazabilidad.md"
  - "00-control/historial-cambios.md"
  - "03-desarrollo/flujo-git-github.md"
  - "03-desarrollo/estandar-diseno-visual.md"
  - "10-backlog/prompts/README.md"
---

# Mapa documental y fuentes de verdad

## 1. Propósito

Este documento indica dónde vive cada definición, cuál prevalece cuando hay diferencias y qué debe actualizarse al aprobar una decisión. No reemplaza el contenido especializado de los documentos enlazados.

## 2. Orden de precedencia

Cuando dos afirmaciones difieran, se aplica este orden:

1. [Registro de decisiones](00-control/registro-decisiones.md): autoridad sobre lo que el propietario aprobó, rechazó o aplazó.
2. [Contradicciones](00-control/contradicciones.md): impide tratar como resuelto un conflicto todavía abierto.
3. [Reglas de negocio](01-producto/reglas-negocio.md): fuente de verdad del comportamiento, solo para reglas marcadas como `Decisión confirmada`.
4. [Alcance del MVP](01-producto/alcance-mvp.md) y [prioridades](01-producto/prioridades.md): fuente de verdad de qué se construye ahora y con qué prioridad.
5. Documentos de requisitos: detalle verificable que debe derivar de las reglas y del alcance.
6. [Glosario](00-control/glosario.md): fuente de verdad de nombres y códigos.
7. [Dudas pendientes](00-control/dudas-pendientes.md) y [supuestos](00-control/supuestos.md): permiten avanzar, pero no equivalen a aprobación.

Los archivos de `respuesta-manuales/` son evidencia de origen. Si difieren de una decisión formalizada, prevalece `registro-decisiones.md`, que debe conservar el enlace a la fuente y explicar la normalización realizada.

## 3. Mapa actual

| Área | Documento | Función | Estado actual |
| --- | --- | --- | --- |
| Entrada | [`../README.md`](../README.md) | Propósito, estado y guía de inicio | Vigente |
| Entrada | [`../CONTRIBUTING.md`](../CONTRIBUTING.md) | Guía operativa breve para contribuir | Vigente |
| Control | [registro-decisiones.md](00-control/registro-decisiones.md) | Decisiones formales `DEC-*` | Vigente |
| Control | [contradicciones.md](00-control/contradicciones.md) | Conflictos abiertos o resueltos `CT-*` | Vigente; `CT-001` resuelta |
| Control | [matriz-trazabilidad.md](00-control/matriz-trazabilidad.md) | Función → regla → historia → criterio → API/datos → prueba → métrica | Cobertura hasta `DEC-039` |
| Control | [historial-cambios.md](00-control/historial-cambios.md) | Cambios relevantes entre versiones | Vigente |
| Control | [dudas-pendientes.md](00-control/dudas-pendientes.md) | Dudas y resoluciones `DP-*` | Tres dudas abiertas que bloquean historias de B0 |
| Control | [supuestos.md](00-control/supuestos.md) | Suposiciones temporales `SUP-*` | Borrador vivo |
| Control | [glosario.md](00-control/glosario.md) | Vocabulario y convenciones de códigos | Borrador mixto |
| Producto | [alcance-mvp.md](01-producto/alcance-mvp.md) | Funciones incluidas, futuras y excluidas | 45 funciones P0 confirmadas |
| Producto | [prioridades.md](01-producto/prioridades.md) | Método y clasificación de prioridades | Vigente |
| Producto | [reglas-negocio.md](01-producto/reglas-negocio.md) | Conducta normativa `RN-*` | Reglas confirmadas |
| Requisitos | [estados-citas.md](02-requisitos/estados-citas.md) | Máquina de estados y transiciones | Confirmada |
| Requisitos | [historias-usuario.md](02-requisitos/historias-usuario.md) | Historias `HU-*` y criterios `CA-*` | Propuesta; B0 y `HU-020`/`HU-021` redactados |
| Backlog | [plan-bloques.md](10-backlog/plan-bloques.md) | Orden de construcción de las 45 funciones P0 en siete bloques | Propuesta |
| Backlog | [prompts-implementacion.md](10-backlog/prompts-implementacion.md) | Prompts de implementación y revisión de B0 y primeras historias de B1 | Propuesta; herramienta, no norma |
| Backlog | [prompts/README.md](10-backlog/prompts/README.md) | Catálogo canónico, plantilla, estado y trazabilidad de prompts reutilizables | Obligatorio para prompts nuevos o revisados; herramienta, no norma |
| Backlog | [prompt-endurecimiento-ddl.md](10-backlog/prompt-endurecimiento-ddl.md) | Ejecución por fases de las correcciones de seguridad e integridad del DDL | Propuesta; herramienta, no norma |
| Desarrollo | [estandar-backend-go.md](03-desarrollo/estandar-backend-go.md) | Código limpio, paquetes y documentación Go | Obligatorio |
| Desarrollo | [estandar-frontend-vue.md](03-desarrollo/estandar-frontend-vue.md) | Código limpio, módulos, TypeScript y accesibilidad | Obligatorio |
| Desarrollo | [estandar-diseno-visual.md](03-desarrollo/estandar-diseno-visual.md) | Paleta, tokens, tipografía, medidas, componentes y pantallas | Obligatorio |
| Desarrollo | [estrategia-pruebas.md](03-desarrollo/estrategia-pruebas.md) | Unitarias, componente, integración, E2E y puertas de calidad | Obligatorio |
| Desarrollo | [flujo-git-github.md](03-desarrollo/flujo-git-github.md) | Ramas, commits, issues, PR, rulesets, squash, hotfix y versiones | Obligatorio; activación remota pendiente |
| Arquitectura | [stack-despliegue-operacion.md](04-arquitectura/stack-despliegue-operacion.md) | Stack, módulos, trabajos y despliegue | Confirmada |
| Arquitectura | [frontend.md](04-arquitectura/frontend.md) | Vue 3, Vite, estructura y reglas de rendimiento | Confirmada |
| Arquitectura | [backend-go.md](04-arquitectura/backend-go.md) | Chi, `net/http`, estructura y middleware | Confirmada |
| Backend | [base-datos.md](05-backend/base-datos.md) | PostgreSQL, RLS, consultas y retención | Confirmada |
| Backend | [estandar-base-datos.md](05-backend/estandar-base-datos.md) | Normalización, migraciones, integridad y pruebas | Obligatorio |
| Backend | [migraciones-atlas.md](05-backend/migraciones-atlas.md) | Atlas, promoción entre ambientes y trazabilidad de esquema | Confirmada |
| Backend | [revision-ddl-seguridad-2026-08-11.md](05-backend/revision-ddl-seguridad-2026-08-11.md) | Auditoría técnica de migraciones y modelo físico de referencia | Revisión técnica; no normativa |
| API | [estandar-openapi.md](06-api/estandar-openapi.md) | OpenAPI, errores, seguridad, ejemplos y compatibilidad | Obligatorio |

## 4. Flujo obligatorio de una decisión

1. Conservar la respuesta o evidencia original.
2. Crear una entrada `DEC-*` con fecha, responsable, motivo y alternativas descartadas.
3. Resolver o crear la `DP-*` o `CT-*` relacionada sin borrar su historia.
4. Actualizar reglas, alcance, requisitos y glosario afectados.
5. Añadir o corregir la fila correspondiente en la matriz de trazabilidad.
6. Registrar los documentos y versiones cambiados en el historial.

Una decisión no está incorporada por completo mientras queden documentos afectados que todavía la contradigan.

## 5. Estados y vacíos conocidos

- Existen las historias `HU-001`–`HU-012` de B0 y las propuestas `HU-020`/`HU-021` de B1 con criterios de aceptación. Las dos de B1 se prepararon anticipadamente por solicitud del propietario, pero no se implementan antes de que B0 cumpla su criterio de salida; el resto de B1 a B6 sigue pendiente.
- Existen dos migraciones de B0 y un modelo físico de referencia todavía no promovido por completo. Aún faltan el contrato OpenAPI concreto, componentes visuales implementados, el catálogo formal de métricas y las pruebas de datos para el resto del modelo.
- La matriz registra esos vacíos como `Pendiente`; no asigna códigos ficticios para aparentar cobertura.
- El registro formal no contiene contradicciones abiertas. La revisión DDL del 2026-08-11 detectó candidatas que deben registrarse antes de convertir las secciones afectadas en migraciones. Del lote respondido el 2026-08-05 no queda ninguna duda; el 2026-08-07 se abrieron `DP-SEG-04`, `DP-SEG-05` y `DP-SEG-06`, que bloquean cinco historias de B0.
- Antes del piloto faltan los textos legales, la verificación de proveedores de notificación y una restauración completa probada.
- Git y el repositorio remoto todavía no están inicializados; el estándar y las plantillas existen, pero el ruleset y los checks se activarán al crear GitHub.
