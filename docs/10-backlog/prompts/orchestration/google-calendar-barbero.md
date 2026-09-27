---
prompt_id: "PROMPT-ORCH-GCAL-BARBERO-v1"
version: "1.0"
kind: "orchestration"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "284"
issue_url: "https://github.com/bcaceres19/barberia/issues/284"
suggested_issue_title: "docs(integraciones): registrar decisiones de Google Calendar por barbero"
branch: "docs/284-google-calendar-decisiones"
pr: 285
pr_url: "https://github.com/bcaceres19/barberia/pull/285"
depends_on:
  - "DEC-099"
  - "DEC-100"
  - "DEC-101"
  - "DEC-102"
rules:
  - "RN-TEN-01"
  - "RN-CON-03"
decisions:
  - "DEC-019"
  - "DEC-040"
  - "DEC-047"
  - "DEC-073"
  - "DEC-076"
  - "DEC-099"
  - "DEC-100"
  - "DEC-101"
  - "DEC-102"
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/10-backlog/prompts/README.md"
created_at: "2026-09-26"
updated_at: "2026-09-26"
supersedes: null
superseded_by: null
---

# Orquestación: publicación de NAVA en Google Calendar por barbero

## Instrucción para el agente

Este archivo solo define orden, dependencias y criterio de avance. No implementa nada ni mezcla ramas o PR: cada entrega tiene su prompt, su issue, su rama y su PR (`AGENTS.md`, «Un prompt atiende una preocupación primaria»). Su propio issue (#284) cubre únicamente el registro documental.

## Principio arquitectónico

La sincronización es **unidireccional, de NAVA hacia Google Calendar** (`DEC-099`). NAVA es la única autoridad sobre las citas y sus reglas; Google Calendar muestra la agenda del barbero. Lo que se cambie o elimine en Google no afecta a NAVA, y las citas solo se cancelan o reprograman desde la app; NAVA únicamente restaura los eventos que se borren por error. Google nunca se llama dentro de una transacción de negocio, y una caída de Google jamás pierde una reserva (`DEC-101`, `DEC-102`).

La propuesta inicial era bidireccional (webhook, `syncToken`, eventos externos como bloqueos, reprogramación desde Google). El propietario la descartó el 2026-09-26; esos prompts no existen y no deben recrearse sin una decisión nueva.

## Entregas y orden

| Orden | Prompt | Preocupación | Depende de |
| --- | --- | --- | --- |
| 0 | Este PR (#284) | Decisiones, dudas, alcance, prompts | — |
| 1 | [PROMPT-FEAT-GCAL-01-v1](../hu/gcal-01-vinculo-barbero-usuario.md) | Vínculo barbero–usuario | #284 integrado |
| 2 | [PROMPT-FEAT-GCAL-02-v1](../hu/gcal-02-conexion-oauth.md) | OAuth, conexión y tokens cifrados | 01 |
| 3 | [PROMPT-FEAT-GCAL-03-v1](../hu/gcal-03-publicacion-nava-a-google.md) | Publicación NAVA → Google con cola propia | 02 |
| 4 | [PROMPT-FEAT-GCAL-04-v1](../hu/gcal-04-frontend-conexion.md) | Interfaz del barbero | 02, 03 |
| 5 | [PROMPT-OPS-GCAL-05-v1](../ops/gcal-05-puesta-en-marcha.md) | Revisión integral y guía de puesta en marcha | 03, 04 |

Ninguna entrega arranca hasta que #284 esté integrado en `main`, tenga un issue real con criterios verificables y su prompt pase a `ready`.

## Criterio de avance

Una entrega se da por terminada solo con su PR integrado en `main`, CI verde y su tabla `Criterio | Estado | Prueba` sin criterios sin prueba. Un bloqueo se registra en el prompt afectado (`blocked`) y en `dudas-pendientes.md`.

## Casos de prueba que siguen vigentes y entrega que los cubre

| Caso | Entrega |
| --- | --- |
| 1 reserva pública → evento; 2 cita manual → evento; 3 reprogramación NAVA; 4 y 5 cancelaciones; 13 bloque NAVA → Google | 03 |
| 15 token expirado; 16 refresh revocado (`reauth_required`); 17 Google caído sin perder la cita | 02 y 03 |
| 18 aislamiento entre tenants | 01, 02 y 03 |
| Nuevo: un cambio o borrado hecho en Google no modifica citas ni bloqueos de NAVA | 03 |
| Nuevo: un evento borrado en Google se restaura; una cita cancelada en la app no se recrea | 03 |
| Nuevo: el recordatorio usa `reminder_minutes` del barbero | 02, 03 y 04 |

Los casos 6 a 12, 14, 19 y 20 de la petición original (mover, eliminar o crear eventos desde Google, webhook duplicado, evento terminal y eco) quedan descartados por `DEC-099`.

## Dudas

No quedan dudas abiertas de esta integración: `DP-INT-01` (el usuario selecciona su barbero, `DEC-100`), `DP-INT-02` (series de bloqueo fuera de la primera fase) y `DP-INT-03` (los eventos permanecen en Google al desconectar) están resueltas.
