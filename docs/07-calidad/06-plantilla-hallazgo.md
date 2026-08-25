---
titulo: "Plantilla de hallazgo"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "README.md"
  - "01-metodologia-y-uso.md"
---

# Plantilla de hallazgo

Copia este bloque por cada hallazgo real durante una sesión de exploración. No se registra una casilla marcada sin defecto; solo se registra cuando algo salió distinto de lo esperado.

```markdown
### Hallazgo: <título corto y específico, no "error en formulario">

- **Fecha:** YYYY-MM-DD
- **Sesión:** manual | IA (agente/skill usados)
- **Módulo / pantalla:** <archivo o ruta>
- **Checklist / fila de matriz:** <enlace al archivo y al ítem, p. ej. `04-checklist-modulos-catalogo.md` §3>
- **Combinación exacta:**
  - Dimensión 1: valor
  - Dimensión 2: valor
  - Ancho: valor
  - (todas las dimensiones relevantes, con el valor literal usado, no una descripción vaga)
- **Pasos para reproducir:**
  1. ...
  2. ...
- **Resultado esperado:** <cita la RN-*, CA-* o sección del estándar visual que lo respalda; si no hay una regla explícita, dilo>
- **Resultado obtenido:** <qué pasó realmente>
- **Evidencia:** <captura, `request_id`, mensaje de consola/red — nunca datos personales reales>
- **Severidad:** Bloqueante | Alto | Medio | Bajo (criterio en `01-metodologia-y-uso.md` §3)
- **Alcance:** ¿afecta a un solo tenant, a todos, solo a un ancho, solo a un navegador?
- **¿Reproducible?** Sí / No / Intermitente — si es intermitente, cuántos intentos de cuántos tuvo el defecto
- **Próxima acción:** issue a crear | prompt de corrección en `docs/10-backlog/prompts/fix/` | duda registrada en `dudas-pendientes.md` | checklist corregido (si el checklist estaba desactualizado, no el producto)
```

## Ejemplo ilustrativo

> Este ejemplo es hipotético, para mostrar el nivel de detalle esperado. No es un hallazgo real confirmado ni un issue abierto; antes de crear un issue a partir de un caso así hay que ejecutarlo de verdad contra el entorno local.

```markdown
### Hallazgo: reactivar un servicio cuyo nombre choca con uno activo nuevo no da error claro

- **Fecha:** 2026-08-25
- **Sesión:** manual
- **Módulo / pantalla:** `apps/web/src/modules/catalog/pages/CatalogPage.vue`
- **Checklist / fila de matriz:** `05-matriz-combinaciones.md` §3, fila 5
- **Combinación exacta:**
  - Estado inicial: servicio "Corte clásico" desactivado
  - Acción intermedia: se crea un nuevo servicio activo también llamado "Corte clásico"
  - Acción: reactivar el original
  - Ancho: 768 px
- **Pasos para reproducir:**
  1. Desactivar "Corte clásico" (servicio A).
  2. Crear un nuevo servicio activo con el mismo nombre "Corte clásico" (servicio B).
  3. Reactivar el servicio A original.
- **Resultado esperado:** según `DEC-067` el nombre debe ser único solo entre servicios activos; reactivar A debería producir un conflicto de nombre igual de claro que crear o renombrar (`name-conflict`), no un error genérico.
- **Resultado obtenido:** `ReactivateServiceOutcome` no declara `name-conflict` en su tipo (solo `not-found`, `transition-conflict`, `idempotency-conflict`); el servidor respondió 409 pero la interfaz lo mostró como `unexpected-error` genérico.
- **Evidencia:** captura adjunta, `request_id: 8f3a...`
- **Severidad:** Medio (el rechazo ocurre, pero el mensaje no orienta a la persona sobre qué hacer)
- **Alcance:** todos los tenants, todos los anchos
- **¿Reproducible?** Sí, 3/3 intentos
- **Próxima acción:** issue a crear + duda registrada en `dudas-pendientes.md` (¿`ReactivateServiceOutcome` debería declarar `name-conflict`?) antes de codificar la corrección
```
