---
titulo: "Metodología de exploración combinatoria"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "README.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../01-producto/reglas-negocio.md"
  - "../../AGENTS.md"
---

# Metodología de exploración combinatoria

## 1. Objetivo

Encontrar defectos que emergen de la **combinación** de condiciones, no de una condición aislada:

1. **Almacenamiento de datos** — algo se guarda distinto de lo que se envió, se pierde, se duplica o queda visible desde el tenant equivocado.
2. **Diseño / responsive** — algo se rompe, se desborda, se trunca o se vuelve inalcanzable en un ancho, zoom o modo de interacción concreto.
3. **500 y errores no controlados por combinaciones raras** — una entrada válida por sí sola pero inesperada en conjunto con otro estado (doble clic, red cortada, sesión vencida a mitad de formulario) tumba el servidor o deja la interfaz en un estado sin salida.
4. **Flujos no esperados** — atrás del navegador, URL directa, refrescar a mitad de un paso, dos pestañas, abrir un enlace de otro tenant.

Ninguna de las cuatro categorías se cubre bien con pruebas dirigidas una por una: el catálogo de `RN-*` y las pruebas unitarias/E2E de [`estrategia-pruebas.md`](../03-desarrollo/estrategia-pruebas.md) ya prueban cada regla por separado. Lo que falta es la intersección.

## 2. Técnica: combinación por pares, no producto cartesiano completo

Cada flujo se describe como una tabla de **dimensiones** (rol, ancho de pantalla, clase de valor de un campo, estado de red, temporización) con varios **valores** posibles cada una. Probar el producto cartesiano completo es inviable (un formulario con 5 campos y 4 clases de valor cada uno ya son 1024 combinaciones); en cambio se sigue la regla práctica de **combinación por pares** (*pairwise*): elegir un conjunto de filas tal que **cada par de valores de dos dimensiones distintas aparezca junto al menos una vez**. Esto captura la enorme mayoría de defectos de interacción con una fracción del esfuerzo, y es coherente con por qué el proyecto ya prioriza el bloque `RN-CON` como el de mayor riesgo: los bugs más caros de este producto son de interacción (dos citas, dos tenants, dos pestañas), no de un campo aislado.

[`05-matriz-combinaciones.md`](05-matriz-combinaciones.md) trae tablas ya reducidas por pares para los flujos actuales. Al extender un checklist para un módulo nuevo:

1. Lista las dimensiones relevantes (mínimo: clase de valor de cada campo, ancho de pantalla, rol/tenant, estado de red, temporización de doble acción).
2. Lista 3–5 valores representativos por dimensión, incluyendo siempre: vacío/límite mínimo, límite máximo exacto, límite máximo + 1, un valor típico válido y un valor "hostil" (script/HTML, unicode, muy largo).
3. Arma filas de combinación cubriendo todos los pares — a mano basta para 3–5 dimensiones; para más, herramientas como PICT generan la tabla, pero no es obligatorio instalar nada.
4. Cada fila de la matriz es una sesión de prueba concreta, no una descripción vaga.

## 3. Severidad

| Severidad | Criterio | Ejemplo |
| --- | --- | --- |
| **Bloqueante** | Corrupción o fuga de datos entre barberías, 500 sin recuperación, pérdida irreversible de datos ya guardados | Un servicio de la barbería A aparece en el selector de la barbería B |
| **Alto** | Dato persistido incorrecto sin fuga entre tenants, flujo roto sin forma de continuar | Guardar configuración deja `contactEmail` como cadena vacía en vez de `null` |
| **Medio** | Mensaje de error incorrecto o genérico donde debía ser específico, responsive roto que dificulta pero no impide la tarea | A 320 px el botón "Guardar" queda parcialmente tapado por la barra inferior |
| **Bajo** | Detalle visual, copy, inconsistencia menor sin efecto funcional | Un ícono decorativo queda descentrado 2 px |

Un hallazgo **Bloqueante** o **Alto** se reporta de inmediato al propietario, no se espera a terminar la sesión completa.

## 4. Protocolo para una sesión humana

1. Elige un checklist (`03`, `04`) o una fila de [`05-matriz-combinaciones.md`](05-matriz-combinaciones.md).
2. Levanta el entorno local (`apps/api` + `apps/web` + PostgreSQL real) con datos de prueba sembrados — nunca contra producción ni con datos reales de una barbería piloto.
3. Fija un tiempo (25–40 min) y recorre las combinaciones en orden, marcando cada una.
4. Ante un resultado inesperado, no lo descartes por parecer menor: regístralo con [`06-plantilla-hallazgo.md`](06-plantilla-hallazgo.md) antes de seguir.
5. Al terminar, si algún hallazgo es real y reproducible, créalo como issue. Si además requiere una corrección de código, el trabajo de corrección se guarda como un prompt persistente `fix/issue-<numero>-<slug>.md` en [`docs/10-backlog/prompts/`](../10-backlog/prompts/README.md), siguiendo `AGENTS.md` — esta carpeta nunca reemplaza esa persistencia.

## 5. Protocolo para un agente de IA

Un agente (por ejemplo mediante la skill `/run` con `claude-in-chrome`, o Playwright programático) puede ejecutar estos checklists de forma sistemática. Reglas obligatorias:

1. **Nunca contra producción.** Solo contra `apps/web` servido localmente contra `apps/api` local y una base PostgreSQL de prueba, sembrada con datos ficticios propios de la sesión (`estrategia-pruebas.md` §7). Si no existe un entorno local levantado, el agente lo indica y no continúa inventando resultados.
2. **Nunca datos personales reales.** Los datos de prueba (nombres, correos, teléfonos) son ficticios y reconocibles como tales, igual que exige `estrategia-pruebas.md` §7 y `RN-DAT-02` para no dejarlos en capturas ni en logs.
3. **Recorre las combinaciones en el orden de la tabla**, sin saltarse pares por parecer redundantes: la redundancia aparente es a menudo donde vive el bug de interacción.
4. **Ante cualquier 500, traza no controlada o dato visiblemente incorrecto tras recargar la página**, captura evidencia inmediatamente (captura de pantalla, `request_id` de la respuesta si existe, mensaje de consola/red) antes de continuar — la evidencia se pierde si el agente sigue navegando.
5. **Registra cada hallazgo** con [`06-plantilla-hallazgo.md`](06-plantilla-hallazgo.md) tal como lo haría una persona; un agente no “corrige sobre la marcha” un defecto encontrado durante exploración sin que el propietario lo apruebe — reporta primero.
6. **No modifica código de producto durante la sesión de exploración.** Encontrar y corregir son responsabilidades separadas en este proyecto (ver `AGENTS.md` — un prompt de `fix` es un artefacto distinto con su propio issue y rama).
7. Al terminar, entrega un resumen: combinaciones cubiertas, hallazgos con severidad, y cuáles quedan pendientes de esta sesión.

## 6. Qué hacer con un hallazgo confirmado

1. Si es reproducible y real: crear el issue en GitHub con los pasos exactos de [`06-plantilla-hallazgo.md`](06-plantilla-hallazgo.md).
2. Si requiere corrección de código: seguir `AGENTS.md` — rama `fix/<issue>-<descripcion>`, prueba de regresión en la capa más baja que reproduzca el defecto fielmente (`estrategia-pruebas.md` §1), y el prompt persistente correspondiente en `docs/10-backlog/prompts/fix/`.
3. Si revela una regla de negocio ambigua o no cubierta (`RN-*` inexistente para el caso encontrado): registrar la duda en `docs/00-control/dudas-pendientes.md` antes de asumir un comportamiento — nunca se inventa la regla durante la exploración.
4. Si el hallazgo confirma que un checklist de esta carpeta quedó desactualizado (el flujo cambió, un límite de validación cambió), corregir el archivo correspondiente en la misma sesión.
