---
name: ui-app-testing
description: "Prueba toda la app NAVA o una pantalla desde la interfaz real con Playwright, subagentes GPT-6 Luna medium, cuentas sintéticas aisladas y hallazgos reproducibles. Se activa con prueben esta app o prueba toda la app. Excluye fixes de producto y revisiones solo de código."
---

# Pruebas de interfaz de NAVA

Fuente: solicitud explícita del propietario, DEC-120, issue #300. Activa «prueben esta app», «prueba toda la app» o una pantalla. Entrega hallazgos/cobertura; no cambia producto.

## Preparación

1. Leer AGENTS.md y [entrada](../../../docs/07-calidad/pruebas-ui/README.md). Detectar rutas actuales y SHA, no heredar alcance histórico de agosto.
2. Aplicar [entorno y cuentas](../../../docs/07-calidad/pruebas-ui/01-entorno-cuentas.md) y local-app-startup si no está listo: API/PostgreSQL reales, Atlas sin pendientes y recorrido autenticado. Wrapper QA desactiva proveedores sin cambiar .env.local.
3. Probar Chromium con el smoke del kit. Scripts Playwright locales bastan si no hay MCP de navegador; node_modules no acredita control del navegador.
4. Seleccionar issue real y [orquestación](../../../docs/10-backlog/prompts/orchestration/pruebas-ui-luna.md) vigente. Primera campaña #301; si cerrada, el coordinador crea otra y nuevos prompts según protocolo. Si GitHub no responde, continuar preflight local y registrar inicio de campaña BLOCKED. Persistir versiones y actualizar índice antes de ejecutar; pending sigue draft. No reescribir cuerpos ya ejecutados.

## Delegación y presupuesto

Leer [reparto](../../../docs/07-calidad/pruebas-ui/03-inventario.md) y [protocolo](../../../docs/07-calidad/pruebas-ui/02-protocolo-agentes.md). Máximo dos agentes, una pantalla por tarea, cuenta/tenant/contexto/salida distintos. Serializar access/recovery, luego cola de pantallas, finalmente journeys/aislamiento.

Codex spawn_agent: model=gpt-6-luna, reasoning_effort=medium, fork_turns=none. Enviar solo ruta de prompt persistido, scope, campaña, SHA, URL local, ruta del manifiesto privado y salida exclusiva. No credenciales/historial completo. Encargos guardados antes de entregar; preferencias explícitas del propietario prevalecen.

Un skill no cambia modelo del coordinador actual. Si el runtime no puede escoger Luna/subagentes, comunicar y registrar modelo real; no sustituir Astra/Sol automáticamente. Ejecutar en serie solo si es compatible con la petición; delegación faltante queda pendiente.

## Ejecución y cierre

Cada agente lee solo ficha/transversales/fuentes de capacidad y usa UI para negocio; éxito/guardado sin mocks. Escribe solo salida ignorada, no producto. F5 acredita recuperación del dato; invariantes internos requieren apoyo SQL autorizado o suites existentes.

Guardar checkpoint privado tras cada mutación y retomar desde estado verificado, según protocolo. Tanda hasta 12 casos/20 min, reencolar NOT_RUN. Dos intentos por bloqueo. Escalar únicamente la pregunta que Luna no resuelve, sin supervisor caro adicional.

Consolidar [informe](../../../docs/07-calidad/pruebas-ui/06-informe.md), [recorridos](../../../docs/07-calidad/pruebas-ui/05-recorridos.md) y [límites](../../../docs/07-calidad/pruebas-ui/08-cobertura.md). Sanitizar evidencia. PASS requiere evidencia, no ausencia de bugs. Actualizar metadatos/índice y enlazar informes/pendientes. Fix confirmado = otro issue/rama/PR.
