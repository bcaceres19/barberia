---
prompt_id: "PROMPT-ORCH-OTP-CANAL-DESTINO-v1"
version: "1.0"
kind: "orchestration"
status: "superseded"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-007"
  - "HU-008"
  - "HU-011"
issue: "pending"
issue_url: null
suggested_issue_title: "security(auth): coordinar canal único y destino verificado de OTP"
branch: null
pr: null
pr_url: null
depends_on:
  - "PROMPT-FIX-OTP-CANAL-UNICO-v1"
  - "PROMPT-FIX-LOGIN-OTP-DESTINO-VERIFICADO-v1"
  - "DP-NOT-06"
  - "DP-SEG-13"
  - "CT-009"
  - "CT-010"
rules:
  - "RN-DAT-02"
  - "RN-REC-04"
  - "RN-REC-05"
decisions:
  - "DEC-026"
  - "DEC-027"
  - "DEC-051"
  - "DEC-052"
  - "DEC-062"
acceptance_criteria:
  - "Cada preocupación tiene un issue, rama, PR y pruebas propios; la orquestación no mezcla commits."
  - "Las decisiones nuevas resuelven explícitamente canal, precedencia, fallback, contacto verificado y respuestas uniformes."
  - "La recuperación y el reto de login nunca envían el mismo código por dos canales a la vez."
  - "El cliente no puede controlar el destino de entrega ni obtener datos personales completos."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/prompts/README.md"
created_at: "2026-09-02"
updated_at: "2026-09-03"
supersedes: null
superseded_by: "PROMPT-FIX-OTP-CANALES-CONFIGURADOS-v2"
---

# Orquestación: canal único y destino verificado de OTP

## Instrucción para el agente

Coordina únicamente los dos prompts independientes enlazados. Este archivo no autoriza cambios de código ni sustituye los issues, decisiones o PR propios de cada preocupación.

## Objetivo

Cerrar y ejecutar de forma trazable dos cambios de autenticación solicitados por el propietario:

1. Recuperación: elegir o aplicar una política que envíe cada código por un solo canal entre correo electrónico y WhatsApp, sin envío simultáneo.
2. Login: derivar el destino desde un contacto verificado de la cuenta, sin pedir ni aceptar un destino arbitrario del cliente.

## Preflight obligatorio

1. Leer `AGENTS.md`, `CLAUDE.md` y los `source_docs`.
2. Resolver `DP-NOT-06`/`CT-009` y `DP-SEG-13`/`CT-010` con nuevas decisiones `DEC-*`; no ejecutar mientras sigan abiertas.
3. Crear o confirmar un issue real por cada prompt de implementación.
4. Confirmar que el cambio de canal compartido no se implementa dos veces ni se mezclan ramas.

## Orden y dependencias

1. Resolver las decisiones de canal, contacto verificado, precedencia, fallback, ausencia de contacto y no enumeración.
2. Ejecutar `PROMPT-FIX-OTP-CANAL-UNICO-v1` en su propia rama y PR; integrar antes de usar sus contratos compartidos.
3. Ejecutar `PROMPT-FIX-LOGIN-OTP-DESTINO-VERIFICADO-v1` en su propia rama y PR, reutilizando únicamente lo que el primer PR haya integrado.
4. Ejecutar la campaña de autenticación completa, incluida recuperación, login escalado, dos tenants, concurrencia, privacidad, responsive y accesibilidad.

## Fuera de alcance

- Un PR combinado para ambas preocupaciones.
- Cambiar decisiones vigentes sin una nueva entrada `DEC-*` enlazada.
- Añadir proveedores o almacenar credenciales, códigos o datos personales reales.

## Criterio de avance

Cada etapa avanza solo con issue real, decisión resuelta, pruebas de la capa adecuada y evidencia. Si una etapa falla, se documenta como bloqueada o inconclusa y no se declara aprobada la plataforma.

## Entrega

Actualizar el catálogo, la matriz, historial, decisiones, dudas, contradicciones, contratos, requisitos y prompts con el estado real de cada rama y PR. La orquestación termina cuando ambos PR están integrados o cuando un bloqueo externo queda registrado con evidencia.
