---
prompt_id: "PROMPT-FIX-OTP-CANAL-UNICO-v1"
version: "1.0"
kind: "fix"
status: "draft"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-008"
related_hu:
  - "HU-011"
  - "HU-020"
issue: "pending"
issue_url: null
suggested_issue_title: "fix(auth): permitir un único canal para cada código de recuperación"
branch: null
pr: null
pr_url: null
depends_on:
  - "DP-NOT-06"
  - "CT-009"
  - "DEC-027"
  - "DEC-051"
  - "DEC-065"
  - "DEC-066"
rules:
  - "RN-REC-04"
  - "RN-REC-05"
  - "RN-DAT-02"
decisions:
  - "DEC-027"
  - "DEC-051"
  - "DEC-065"
  - "DEC-066"
acceptance_criteria:
  - "La decisión aprobada define si el canal se elige por solicitud, por configuración o mediante otra política explícita."
  - "Un código de recuperación se entrega por un solo canal aprobado: correo electrónico o WhatsApp oficial, nunca por ambos en la misma emisión."
  - "El backend valida el canal contra un contacto verificado y no acepta un destino arbitrario proporcionado por el cliente."
  - "No se filtra existencia de cuenta, canal disponible, destino completo ni contenido del código en respuestas o registros técnicos."
  - "La operación conserva idempotencia, reenvío controlado y evidencia por intento sin datos personales en claro."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/prompts/README.md"
  - "docs/10-backlog/prompts/hu/hu-008-recuperacion-acceso.md"
created_at: "2026-09-02"
updated_at: "2026-09-02"
supersedes: null
superseded_by: null
---

# Canal único para códigos de recuperación

## Instrucción para el agente

Implementa únicamente la política de entrega de códigos de recuperación descrita en este archivo. Este prompt permanece en `draft`: no se ejecuta ni autoriza cambios hasta que exista un issue real, se resuelvan `DP-NOT-06` y `CT-009`, y se actualicen los metadatos con la decisión aprobada.

## Objetivo

Permitir que el flujo de recuperación ofrezca un canal controlado para cada código OTP —correo electrónico o WhatsApp oficial— y garantice que una emisión concreta se entregue por exactamente un canal, sin envíos simultáneos.

La decisión pendiente debe fijar si el canal lo elige el usuario en el flujo, lo determina la configuración de la barbería o se aplica otra política. El agente no puede inferir esa elección ni introducir fallback automático.

## Preflight obligatorio

1. Verifica que el árbol de trabajo ya contiene cambios ajenos y no los sobrescribas.
2. Lee `AGENTS.md`, `CLAUDE.md` y todos los `source_docs` completos.
3. Comprueba el estado de `DP-NOT-06` y `CT-009`; si siguen abiertos, detén la ejecución de implementación y conserva este prompt en `draft`.
4. Comprueba que el issue real y la decisión `DEC-*` nueva estén enlazados aquí, en el índice y en la trazabilidad antes de crear una rama.
5. Revisa el contrato OpenAPI, el cliente tipado, `staff_recovery_code`, `notification_attempt` y los adaptadores existentes antes de proponer cambios de esquema.
6. No uses cuentas, correos, teléfonos, códigos, claves ni destinatarios reales en fixtures, pruebas, logs, comentarios o documentación.

## Alcance incluido

- Selección o representación del canal aprobado para el flujo de `HU-008`/`HU-011`, según la decisión nueva.
- Contrato HTTP y cliente tipado necesarios para transportar únicamente el canal permitido.
- Validación server-side de que el canal tiene un contacto verificado utilizable.
- Envío exclusivo por un adaptador en cada emisión y persistencia de la evidencia técnica por canal, sin datos personales en claro.
- Estados de interfaz para canal no disponible, envío en curso, error recuperable, reenvío y éxito, conservando no enumeración.

## Fuera de alcance

- Cambiar el reto de login de `HU-007`; ese trabajo pertenece a `PROMPT-FIX-LOGIN-OTP-DESTINO-VERIFICADO-v1`.
- Añadir SMS, proveedores no oficiales, enlaces mágicos, recuperación sin prueba de posesión o envío a dos canales.
- Cambiar la política de contraseña, la sesión, el límite por IP o la navegación general de NAVA.
- Ejecutar una migración o modificar el modelo de datos sin justificarlo contra el contrato y el estándar Atlas.

## Estado existente que debe conservarse

- `DEC-065`: la solicitud de recuperación no enumera cuentas y el destino solo se revela enmascarado después de verificar un código válido.
- `DEC-064`: código de 6 dígitos, HMAC-SHA256, vigencia, intentos, cooldown, reenvío y token opaco de reinicio.
- `RN-REC-04` y `RN-DAT-02`: cada intento deja evidencia técnica, sin correo, teléfono, nombre ni contenido en claro.
- PostgreSQL, RLS por barbería, idempotencia y adaptadores oficiales ya existentes.

## Trabajo requerido

1. Propaga la decisión aprobada a `HU-008`, `HU-011`, reglas de negocio, contrato OpenAPI, cliente tipado, modelo de datos si aplica, matriz de trazabilidad y prompt operativo.
2. Diseña la operación para que el canal y el destinatario no puedan ser usados como selector arbitrario de otra cuenta o contacto.
3. Garantiza que una misma solicitud idempotente no envíe por segunda vez ni cambie de canal silenciosamente.
4. Registra cada intento con estado, proveedor y canal; nunca guardes el código ni datos personales en claro.
5. Implementa la UI solo con los patrones y tokens vigentes de NAVA, con foco, teclado, estados accesibles y reflow.

## Pruebas y evidencia

- Dominio y servicios: selección válida, canal no disponible, canal no permitido, reenvío y doble envío.
- HTTP/OpenAPI: schema, errores uniformes, no enumeración y rechazo de campos de destino no autorizados.
- PostgreSQL real: dos tenants, RLS, idempotencia y carrera de solicitudes; exactamente un intento efectivo por emisión.
- Adaptadores: dobles deterministas para correo/WhatsApp y prueba de fallo parcial sin filtrar existencia de cuenta.
- Componente y E2E: pasos de recuperación, teclado, foco, objetivos táctiles y anchos 320, 360, 768, 1280 y zoom 200 %.
- Evidencia: informe con request IDs no sensibles, estados de proveedor sanitizados, capturas y resultado de cada criterio.

## Documentación y trazabilidad

- Al resolver la duda, registra una nueva `DEC-*`; no reescribas `DEC-051`/`DEC-066` sin enlazar la resolución.
- Actualiza `dudas-pendientes.md`, `contradicciones.md`, `registro-decisiones.md`, `historias-usuario.md`, `matriz-trazabilidad.md`, contrato y README afectados.
- Actualiza este prompt y el catálogo con issue, rama, PR y estado reales; si el alcance cambia después de iniciar, crea una versión sucesora.

## Verificación final

```text
git diff --check
Atlas lint/validate y migraciones aplicadas según docs/05-backend/migraciones-atlas.md
formato, lint, typecheck y build del frontend
pruebas Go del módulo auth/notification
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web test:e2e
pruebas PostgreSQL reales con dos tenants y concurrencia
```

Entrega una tabla:

`Criterio | Estado | Prueba o evidencia`

No declares cumplido el envío de un canal ni la ausencia de envío doble sin evidencia observable del proveedor o del doble de prueba equivalente.

## Git y PR

- Commit/PR: `fix(auth): permitir un único canal para cada código de recuperación`.
- Usa `Closes #<issue>` solo si se cubre el issue completo; de lo contrario, `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
