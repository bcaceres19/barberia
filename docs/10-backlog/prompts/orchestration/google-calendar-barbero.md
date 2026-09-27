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
pr: null
pr_url: null
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

# Orquestación: Google Calendar bidireccional por barbero

## Instrucción para el agente

Este archivo solo define orden, dependencias y criterio de avance. No implementa nada ni mezcla ramas o PR: cada entrega tiene su prompt, su issue, su rama y su PR (`AGENTS.md`, «Un prompt atiende una preocupación primaria»). Su propio issue (#284) cubre únicamente el registro documental.

## Principio arquitectónico

NAVA es la autoridad sobre el negocio de las citas. Google Calendar es una vista e integración de la agenda personal del barbero. Las citas de NAVA se sincronizan como eventos; los eventos externos de Google afectan la disponibilidad como bloqueos y nunca se convierten en citas. Google nunca se llama dentro de una transacción de negocio, y una caída de Google jamás pierde una reserva (`DEC-099`–`DEC-102`).

## Entregas y orden

| Orden | Prompt | Preocupación | Depende de |
| --- | --- | --- | --- |
| 0 | Este PR (#284) | Decisiones, dudas, alcance, prompts | — |
| 1 | [PROMPT-FEAT-GCAL-01-v1](../hu/gcal-01-vinculo-barbero-usuario.md) | Vínculo barbero–usuario | #284 integrado, `DP-INT-01` resuelta |
| 2 | [PROMPT-FEAT-GCAL-02-v1](../hu/gcal-02-schedule-fuente-google-calendar.md) | Fuente `google_calendar` en `schedule` | #284 integrado |
| 3 | [PROMPT-FEAT-GCAL-03-v1](../hu/gcal-03-booking-accion-integracion.md) | Acciones de `booking` con actor sistema | #284 integrado |
| 4 | [PROMPT-FEAT-GCAL-04-v1](../hu/gcal-04-conexion-oauth.md) | OAuth, conexión y tokens | 01 |
| 5 | [PROMPT-FEAT-GCAL-05-v1](../hu/gcal-05-nava-a-google.md) | NAVA → Google con cola propia | 04 |
| 6 | [PROMPT-FEAT-GCAL-06-v1](../hu/gcal-06-google-a-nava.md) | Google → NAVA (webhook, `syncToken`) | 02, 03, 05 |
| 7 | [PROMPT-FEAT-GCAL-07-v1](../hu/gcal-07-frontend-conexion.md) | Interfaz del barbero | 04, 06 |
| 8 | [PROMPT-OPS-GCAL-08-v1](../ops/gcal-08-puesta-en-marcha.md) | Revisión integral y guía de puesta en marcha | 06, 07 |

Las entregas 01, 02 y 03 son independientes entre sí y pueden ejecutarse en paralelo en ramas separadas. Ninguna arranca hasta que #284 esté integrado en `main`, tenga un issue real con criterios verificables y su prompt pase a `ready`.

## Criterio de avance

Una entrega se da por terminada solo con su PR integrado en `main`, CI verde y su tabla `Criterio | Estado | Prueba` sin criterios sin prueba. Un bloqueo se registra en el prompt afectado (`blocked`) y en `dudas-pendientes.md`.

## Casos obligatorios y entrega que los cubre

| Caso | Entrega |
| --- | --- |
| 1 reserva pública → evento; 2 cita manual → evento; 3 reprogramación NAVA; 4 y 5 cancelaciones; 13 bloque NAVA → Google | 05 |
| 15 token expirado; 16 refresh revocado (`reauth_required`); 17 Google caído sin perder la cita | 04 y 05 |
| 6 mover desde Google; 7 mover a horario ocupado; 8 eliminar cita desde Google; 19 evento terminal | 03 y 06 |
| 9, 11 y 12 evento personal crear/modificar/eliminar; 10 la disponibilidad pública lo excluye | 02 y 06 |
| 14 webhook duplicado; 20 eco de sincronización | 05 y 06 |
| 18 aislamiento entre tenants | 01, 04, 05 y 06 |

## Decisiones aún abiertas

- `DP-INT-01`: quién asigna el vínculo barbero–usuario. Bloquea la entrega 01.
- `DP-INT-02`: series de bloqueo de NAVA hacia Google (fuera de la primera fase por defecto).
- `DP-INT-03`: destino de los eventos ya creados en Google al desconectar (permanecen por defecto).
