---
prompt_id: "PROMPT-FIX-LOGIN-OTP-DESTINO-VERIFICADO-v1"
version: "1.0"
kind: "fix"
status: "superseded"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-007"
related_hu:
  - "HU-005"
  - "HU-008"
  - "HU-011"
issue: "pending"
issue_url: null
suggested_issue_title: "fix(auth): derivar el destino del reto OTP desde contactos verificados"
branch: null
pr: null
pr_url: null
depends_on:
  - "DP-SEG-13"
  - "CT-010"
  - "PROMPT-FIX-OTP-CANAL-UNICO-v1"
  - "DEC-026"
  - "DEC-052"
  - "DEC-062"
rules:
  - "RN-DAT-02"
  - "RN-REC-04"
decisions:
  - "DEC-026"
  - "DEC-052"
  - "DEC-062"
acceptance_criteria:
  - "El cliente usa el correo solo para identificar la cuenta cuando corresponda; nunca elige ni envía un teléfono o correo de destino arbitrario."
  - "El servidor deriva el destino exclusivamente desde un contacto verificado almacenado en la cuenta y aplica la precedencia aprobada."
  - "La respuesta es uniforme cuando la cuenta no existe, está inactiva o no tiene un contacto verificado utilizable."
  - "Una solicitud escalada no evalúa la contraseña antes de superar el reto y no permite evadir el límite cambiando el destino."
  - "No aparecen contactos completos, códigos ni datos personales en respuestas, URLs, logs, métricas o trazas."
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
  - "docs/10-backlog/prompts/hu/hu-007-defensa-abuso.md"
created_at: "2026-09-02"
updated_at: "2026-09-03"
supersedes: null
superseded_by: "PROMPT-FIX-OTP-CANALES-CONFIGURADOS-v2"
---

# Destino verificado para el OTP del login

## Instrucción para el agente

Implementa únicamente el cambio de seguridad del destino del reto adicional de login descrito aquí. Este prompt permanece en `draft`: no se ejecuta hasta que exista un issue real, se resuelvan `DP-SEG-13` y `CT-010`, y se actualicen los metadatos con la política aprobada.

## Objetivo

Cuando el límite de login exige un reto OTP, el servidor debe localizar la cuenta mediante el identificador de acceso permitido y derivar el destino desde un correo o teléfono verificado almacenado en `staff_user`. El cliente no debe pedir, aceptar ni controlar un destino de entrega.

La decisión debe fijar la precedencia entre correo y teléfono, el canal aplicable, el tratamiento de cuentas sin contacto verificado y si se permite elegir entre varios contactos ya verificados. “No pedir el destino” no elimina necesariamente el correo como identificador de la cuenta; impide usarlo como parámetro de entrega arbitrario.

## Preflight obligatorio

1. Verifica cambios ajenos en el árbol y no los sobrescribas.
2. Lee `AGENTS.md`, `CLAUDE.md` y todos los `source_docs` completos.
3. Comprueba `DP-SEG-13` y `CT-010`; si siguen abiertos, no implementes.
4. Confirma el issue real, la decisión nueva y la dependencia del prompt de canal único antes de crear una rama.
5. Revisa `POST /api/v1/public/auth/challenge`, `challenge/verify`, `auth_phone_challenge`, las funciones `SECURITY DEFINER`, el adaptador de notificaciones y la obtención del tenant.
6. No uses un correo, teléfono, OTP, secreto o dato personal real en código, fixture, evidencia o log.

## Alcance incluido

- Resolución server-side del contacto verificado y de la política de canal del reto de `HU-007`.
- Contrato OpenAPI, cliente tipado y UI necesarios para dejar de solicitar un destino de entrega.
- Respuesta uniforme ante cuenta inexistente/inactiva, contacto no verificado o canal no utilizable.
- Protección contra manipulación del payload, enumeración, cambio de destino, doble solicitud y carreras.
- Auditoría técnica sanitizada del intento y del canal, sin exponer el contacto completo ni el OTP.

## Fuera de alcance

- Recuperación de contraseña de `HU-008`/`HU-011`, salvo la reutilización explícita de una política común ya aprobada.
- Login sin correo/identificador de cuenta, passwordless, SMS, proveedores no oficiales o bypass del rate limit.
- Permitir que el cliente envíe un teléfono o correo para “recibir” el código.
- Cambiar la sesión, la contraseña, el umbral de abuso o la identidad visual fuera de los estados afectados.

## Estado existente que debe conservarse

- `DEC-052` y `DEC-062`: ventana, escalamiento, vigencia, intentos, consumo y límites propios del reto.
- El reto debe continuar atado a la cuenta candidata y a la IP escalada, con respuestas no enumerables.
- `RN-DAT-02`, `RN-REC-04`, RLS, HMAC de IP/código, idempotencia y funciones estrechas de autenticación.
- El correo usado para identificar la cuenta no se convierte automáticamente en un destino hasta que la política lo considere contacto verificado.

## Trabajo requerido

1. Propaga la decisión aprobada a `HU-007`, reglas, contrato, cliente, adaptador de notificación, modelo de datos solo si es indispensable, matriz y README afectados.
2. Elimina del contrato cualquier campo de destino que permita que el cliente elija un número o correo ajeno; si se conserva un identificador de cuenta, documenta su finalidad separada.
3. Resuelve el destino dentro del límite de confianza correcto, con normalización, verificación y contexto de tenant; no uses una consulta RLS insegura ni un dato no verificado.
4. Mantén respuestas uniformes para no revelar existencia, estado o ausencia de contacto y no devuelvas el destino completo.
5. Prueba que cambiar el payload, repetir la solicitud, cambiar IP o usar un tenant distinto no cambia la cuenta ni permite entregar el reto a otro contacto.

## Pruebas y evidencia

- Dominio y servicio: precedencia aprobada, contacto no verificado, cuenta inactiva, destino manipulado y política de canal único.
- HTTP/OpenAPI: contrato sin destino arbitrario, respuesta uniforme, no evaluación prematura de contraseña y errores sanitizados.
- PostgreSQL real: dos tenants, RLS, funciones `SECURITY DEFINER`, aislamiento, carrera de solicitudes e idempotencia.
- Integración de proveedores: el doble de correo/WhatsApp recibe únicamente el contacto derivado de la fixture; nunca un valor del request.
- E2E: login normal, escalamiento, reto válido, código inválido, reenvío y navegación sin mostrar el destino completo.
- Evidencia responsive y accesible en 320, 360, 768, 1280 y zoom 200 %, con logs sin PII.

## Documentación y trazabilidad

- Registra una nueva `DEC-*` al resolver `DP-SEG-13`/`CT-010`; no alteres silenciosamente `DEC-062`.
- Actualiza `dudas-pendientes.md`, `contradicciones.md`, `registro-decisiones.md`, `historias-usuario.md`, `matriz-trazabilidad.md`, contrato y prompts afectados.
- Actualiza este prompt y el catálogo con issue, rama, PR y estado reales; crea una nueva versión si cambia el alcance después de iniciar.

## Verificación final

```text
git diff --check
Atlas lint/validate y migraciones aplicadas según docs/05-backend/migraciones-atlas.md
formato, lint, typecheck y build del frontend
pruebas Go del módulo auth/notification
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web test:e2e
pruebas PostgreSQL reales con dos tenants, RLS y concurrencia
```

Entrega una tabla:

`Criterio | Estado | Prueba o evidencia`

No declares seguro el flujo sin demostrar que el destino solo puede provenir de un contacto verificado de la cuenta candidata.

## Git y PR

- Commit/PR: `fix(auth): derivar el destino del reto OTP desde contactos verificados`.
- Usa `Closes #<issue>` solo si se cubre el issue completo; de lo contrario, `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
