---
name: browser-viewport-verification
description: "Alias heredado de la verificación de viewport de visual-qa. Se conserva para que los prompts persistentes que citan esta ruta sigan resolviendo; no se activa automáticamente."
disable-model-invocation: true
---

# browser-viewport-verification · alias heredado

Este skill fue absorbido por el skill compartido `visual-qa` (issue #259, `DEC-086`). Su contenido ya no se mantiene aquí.

Si un issue, un prompt o el usuario te remitió a este archivo, lee y aplica `.agents/skills/visual-qa/references/viewport-verification.md`: nunca confíes en la respuesta de un resize sin medir `window.innerWidth` y `window.innerHeight`, y usa el arnés Playwright `apps/web/e2e/*-evidencia-responsiva.spec.ts` cuando `resize_window` no cambie el ancho real. Si además hay un mockup asignado, sigue `.agents/skills/visual-qa/SKILL.md` en modo fidelidad.
