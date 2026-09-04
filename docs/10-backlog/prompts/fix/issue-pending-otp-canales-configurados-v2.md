---
prompt_id: "PROMPT-FIX-OTP-CANALES-CONFIGURADOS-v2"
version: "2.0"
kind: "fix"
status: "draft"
target_agents:
  - "claude"
  - "codex"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-007"
related_hu:
  - "HU-008"
  - "HU-011"
issue: "pending"
issue_url: null
suggested_issue_title: "fix(auth): configurar canales y destinos verificados para OTP"
branch: null
pr: null
pr_url: null
depends_on:
  - "DEC-081"
rules:
  - "RN-REC-04"
  - "RN-REC-05"
  - "RN-DAT-02"
decisions:
  - "DEC-052"
  - "DEC-062"
  - "DEC-064"
  - "DEC-065"
  - "DEC-066"
  - "DEC-081"
acceptance_criteria:
  - "Correo es el canal predeterminado y cada evento puede configurar correo, WhatsApp oficial o ambos."
  - "El servidor deriva los destinos desde contactos verificados; el cliente nunca aporta un destino arbitrario."
  - "Una entrega por ambos canales comparte código y operación lógica, con un intento técnico trazable por canal."
  - "Cuenta inexistente, inactiva, sin contacto utilizable y código inválido conservan respuestas no enumerables."
  - "Contrato, cliente tipado, persistencia, adaptadores, UI, pruebas y documentación permanecen coherentes."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/matriz-trazabilidad.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-03/auth-eventos/README.md"
  - "apps/api/internal/modules/auth"
  - "apps/api/internal/modules/notification"
  - "apps/api/cmd/api/main.go"
created_at: "2026-09-03"
updated_at: "2026-09-04"
supersedes:
  - "PROMPT-FIX-OTP-CANAL-UNICO-v1"
  - "PROMPT-FIX-LOGIN-OTP-DESTINO-VERIFICADO-v1"
  - "PROMPT-ORCH-OTP-CANAL-DESTINO-v1"
superseded_by: null
---

# Canales configurados y destinos verificados para OTP

## Instrucción

No ejecutes este prompt mientras `issue: pending`. Cuando exista un issue real, implementa únicamente la adaptación funcional exigida por `DEC-081`; no la mezcles con el rediseño visual de autenticación.

## Resultado requerido

- Configuración por evento con `email` predeterminado, `whatsapp` oficial o `both`.
- Destinos obtenidos solo en servidor desde contactos verificados y dentro del tenant de la cuenta candidata.
- En `both`, un único código, vigencia, contador de intentos e identidad de emisión; cada proveedor conserva su intento técnico independiente.
- Ningún request público acepta teléfono o correo de entrega. El correo de login/recuperación, cuando exista, identifica la cuenta; no autoriza por sí solo un destino.
- Respuestas, tiempos perceptibles y copy no enumeran existencia, estado, canal disponible, ausencia de contacto ni causa concreta del rechazo del código.

## Preflight obligatorio

1. Verifica el árbol, preserva cambios ajenos y lee completos todos los `source_docs`.
2. Confirma que existe un issue funcional propio con criterios observables. Mientras `issue: pending`, detente antes de crear rama o modificar código.
3. Comprueba que `DEC-081`, `RN-REC-05`, `HU-007` y `HU-008` siguen coherentes y que no existe otra decisión posterior sobre canales o destinos.
4. Audita el contrato y el comportamiento actual: el reto de login usa un destino telefónico devuelto por persistencia; recuperación construye un remitente dual o una excepción email-only de local/test. No confundas esa selección por credenciales de despliegue con la configuración de negocio por evento.
5. Revisa las migraciones aplicadas, roles, funciones `SECURITY DEFINER`, RLS, `notification_attempt` y límites de privacidad antes de diseñar persistencia nueva.
6. Si aparece una contradicción sobre propietario, granularidad, ausencia de contacto o fallo parcial de `both`, regístrala antes de implementar; no inventes un fallback.

## Trabajo obligatorio

1. Lee completos los `source_docs`, inspecciona cambios ajenos y confirma issue/rama antes de editar.
2. Actualiza OpenAPI primero; regenera el cliente tipado y rechaza campos desconocidos de destino.
3. Implementa la política en dominio/servicio y adapta persistencia mediante migración Atlas solo si es indispensable; una migración aplicada no se modifica.
4. Reutiliza los adaptadores oficiales de `DEC-066`. No agregues SMS, proveedor no oficial ni fallback a un canal no configurado.
5. Ajusta la UI a las variantes 07–12 del atlas de autenticación, conservando seis casillas OTP, bloqueo durante envío, foco y mensajes uniformes.
6. Actualiza trazabilidad y este prompt con issue, rama, PR y estado reales.

La configuración debe tener un dueño explícito y una representación normalizada. No uses `jsonb`, un mapa opaco ni variables de entorno como sustituto de configuración por barbería/evento. Si el esquema vigente no permite modelarla sin ambigüedad, crea una migración Atlas nueva y pruebas de actualización con datos existentes; nunca edites una migración aplicada.

## Pruebas obligatorias

- Dominio/servicio: email por defecto, WhatsApp, ambos, falta de contacto verificado, cuenta inactiva, código inválido y reenvío.
- HTTP/OpenAPI: payload sin destino, rechazo de campos desconocidos, respuesta uniforme e idempotencia.
- PostgreSQL real: dos tenants, RLS, concurrencia y una sola emisión lógica; en `both`, exactamente dos intentos de proveedor asociados al mismo código sin PII en claro.
- Adaptadores: destinos solo derivados por servidor; fallo parcial sin fallback no configurado ni filtración al cliente.
- Componentes/E2E: eventos 07–12 en desktop/mobile, teclado, pegado de seis dígitos, foco, `prefers-reduced-motion`, 320/360/768/1280 y zoom 200 %.

## Fuera de alcance

- Rediseñar otras pantallas, cambiar umbral/vigencia/intentos, agregar cuentas de cliente, mostrar destinos completos o guardar secretos/OTP/PII.
- Ejecutar el cambio dentro del issue visual `#188` o `#213`.

## Entrega

Entrega `Criterio | Estado | Evidencia`, ejecuta `git diff --check`, pruebas del módulo Go, OpenAPI, migraciones reales aplicables y suites frontend. No declares `PASS` sin demostrar aislamiento, no enumeración y la correlación correcta de ambos intentos cuando `both` esté activo.

Cuando exista el issue, crea `fix/<issue>-otp-canales-configurados` desde `main` actualizada. Commit/PR sugerido: `fix(auth): configura canales y destinos verificados para OTP`. Usa `Closes #<issue>` solo si contrato, backend, persistencia, proveedores, UI, pruebas y documentación cubren todos sus criterios; en otro caso usa `Refs #<issue>`. No hagas push directo, force push ni merge de `main`.
