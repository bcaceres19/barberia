---
titulo: "Exploración combinatoria y checklists de bugs"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../03-desarrollo/estandar-diseno-visual.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../02-requisitos/historias-usuario.md"
  - "../00-control/matriz-trazabilidad.md"
  - "../../AGENTS.md"
---

# Exploración combinatoria y checklists de bugs

## 1. Qué es esta carpeta y qué no es

Esta carpeta guarda **checklists de exploración manual o dirigida por IA** del aplicativo ya construido, orientados a encontrar defectos que una prueba automatizada puntual no cubre: combinaciones raras de campos, condiciones de carrera de interfaz, responsive roto, datos mal almacenados y errores 500/no controlados que solo aparecen al combinar varias cosas a la vez.

**No sustituye** [`estrategia-pruebas.md`](../03-desarrollo/estrategia-pruebas.md): esa es la fuente normativa de qué prueba automatizada es obligatoria para cada cambio de código, y sigue mandando. Esta carpeta es el complemento exploratorio: encuentra el caso raro que ninguna prueba dirigida imaginó, y su resultado (cuando confirma un defecto real) se convierte en un issue y, si corresponde, en una prueba de regresión permanente en la capa adecuada — nunca en una prueba nueva dentro de esta carpeta.

Tampoco es el catálogo formal `07-calidad/casos-prueba-funcionales.md` que mencionan [`estados-citas.md`](../02-requisitos/estados-citas.md) §13 y otros documentos como destino futuro de casos ligados uno a uno con una `RN-*`/`CA-*` y con un identificador `E2E-*`/`IT-BD-*` en la matriz de trazabilidad. Ese archivo, cuando exista, será normativo y trazable. Los archivos de esta carpeta son **hipótesis de exploración**: una casilla marcada significa "no se encontró falla explorando esta combinación en esta sesión", nunca "está garantizado". Por eso no se enlazan a `matriz-trazabilidad.md`.

## 2. Quién usa esto y cómo

- **Una persona** ejecuta un archivo como guion de sesión: 25–40 minutos, un módulo, resultado registrado con [`06-plantilla-hallazgo.md`](06-plantilla-hallazgo.md).
- **Un agente de IA** (por ejemplo vía la skill `/run` + `claude-in-chrome`, o Playwright dirigido por un agente) recorre el mismo checklist de forma sistemática contra una instancia local con datos de prueba sembrados — nunca contra producción ni con datos reales (`RN-DAT-01`, [`estrategia-pruebas.md`](../03-desarrollo/estrategia-pruebas.md) §7: "Solo datos ficticios"). Ver el protocolo exacto en [`01-metodologia-y-uso.md`](01-metodologia-y-uso.md) §4.

Empieza siempre por [`01-metodologia-y-uso.md`](01-metodologia-y-uso.md): explica la técnica de combinación por pares, la taxonomía de severidad y el protocolo para agentes de IA.

## 3. Índice

| Archivo | Contenido |
| --- | --- |
| [01-metodologia-y-uso.md](01-metodologia-y-uso.md) | Técnica de combinación, cómo correr una sesión (humano o IA), severidad, escalamiento a issue/prueba de regresión |
| [02-checklist-transversal.md](02-checklist-transversal.md) | Checks que aplican a **toda** pantalla y flujo: responsive, errores HTTP, aislamiento de tenant, idempotencia, teclado/accesibilidad, almacenamiento |
| [03-checklist-autenticacion-sesion.md](03-checklist-autenticacion-sesion.md) | Login, reto telefónico, recuperación en 3 pasos, sesión, guardia del panel |
| [04-checklist-modulos-catalogo.md](04-checklist-modulos-catalogo.md) | Configuración de la barbería, barberos, catálogo de servicios (incl. ciclo de vida), servicios por barbero |
| [05-matriz-combinaciones.md](05-matriz-combinaciones.md) | Tablas concretas de combinaciones por pares, listas para ejecutar, con valores límite reales tomados del código |
| [06-plantilla-hallazgo.md](06-plantilla-hallazgo.md) | Formato para reportar un defecto encontrado durante una sesión |
| [07-modulos-futuros-pendientes.md](07-modulos-futuros-pendientes.md) | Qué falta cubrir cuando existan agenda, reserva pública, bloqueos y recordatorios (hoy son solo carpetas `index.ts` vacías) |

## 4. Alcance actual (2026-08-25)

Los checklists de `03` y `04` cubren exactamente lo que existe en `apps/web`/`apps/api` hoy: autenticación y sesión (`HU-005`–`HU-012`), configuración de la barbería (`HU-020`), barberos (`HU-021`), catálogo de servicios y su ciclo de vida (`HU-022`, `HU-024`) y asignación de servicios a barberos (`HU-023`). Agenda, reserva pública, bloqueos y notificaciones (`schedule`, `booking`, `notification` en el backend; `agenda`, `public-booking`, `schedules`, `notifications` en el frontend) son hoy solo el esqueleto de un módulo — ver [`07-modulos-futuros-pendientes.md`](07-modulos-futuros-pendientes.md) en vez de inventar un flujo que no existe todavía.
