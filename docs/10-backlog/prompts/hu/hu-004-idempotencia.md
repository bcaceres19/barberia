---
prompt_id: "PROMPT-HU-004-v1"
version: "1.0"
kind: "hu"
status: "executed"
target_agents:
  - "claude"
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-004"
related_hu: []
issue: "40"
issue_url: "https://github.com/bcaceres19/barberia/issues/40"
suggested_issue_title: "feat(api): implementar HU-004 idempotencia reutilizable"
branch: "feat/40-hu004-idempotencia"
pr: 41
pr_url: "https://github.com/bcaceres19/barberia/pull/41"
depends_on:
  - "HU-003 integrada"
rules:
  - "RN-IDE-01"
decisions:
  - "DEC-035"
  - "DEC-037"
  - "DEC-038"
  - "DEC-043"
acceptance_criteria:
  - "CA-004-01"
  - "CA-004-02"
  - "CA-004-03"
  - "CA-004-04"
  - "CA-004-05"
  - "CA-004-06"
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/prioridades.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-backend-go.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/05-backend/migraciones-atlas.md"
  - "docs/06-api/estandar-openapi.md"
  - "database/migrations/20260807170100_create_idempotency_record.sql"
  - "database/migrations/20260811145252_harden_roles_and_definer_functions.sql"
  - "database/migrations/20260811154100_harden_idempotency_concurrency.sql"
  - "database/migrations/20260811220000_harden_idempotency_fingerprint_format.sql"
  - "database/tests/idempotency_concurrency.sql"
  - "database/tests/idempotency_concurrency_two_connections.sh"
created_at: "2026-08-12"
updated_at: "2026-08-13"
supersedes: "docs/10-backlog/prompts-implementacion.md#prompt-de-hu-004--idempotencia-reutilizable"
superseded_by: null
---

# Implementar HU-004: idempotencia reutilizable

## Instrucción para Claude o Codex

Implementa únicamente `HU-004` como capacidad reutilizable del API. Conserva el DDL y la semántica concurrente ya aprobados, y produce un pull request borrador verificable. No implementes citas ni publiques un endpoint ficticio.

## Objetivo

Entregar una única abstracción de idempotencia para futuras escrituras críticas: la misma clave y la misma huella reproducen la respuesta original sin repetir efectos; una huella distinta obtiene `409`; la concurrencia nunca ejecuta el efecto dos veces; los fallos permiten un reintento legítimo; y la misma clave literal permanece aislada por barbería.

## Preflight obligatorio

1. Verifica árbol limpio, `main` actualizada por fast-forward y ausencia de cambios ajenos que puedan sobrescribirse.
2. Consulta Graphify por idempotencia, transacciones tenant-aware, contrato HTTP y dependencias si existe el grafo.
3. Lee por completo cada `source_docs`. Confirma que `HU-003` está integrada y que sus errores uniformes, correlación y OpenAPI están disponibles.
4. Busca un issue abierto que cubra exactamente `HU-004`. Si no existe, créalo con el título sugerido y copia alcance, criterios, pruebas y exclusiones.
5. Sustituye `issue: pending` por el issue real, actualiza el índice y cambia a `ready`; no ejecutes modificaciones mientras el prompt siga en borrador.
6. Crea `feat/<issue>-hu004-idempotencia` desde `main` y registra rama y estado `in_progress`.

## Alcance incluido

- Adaptador HTTP o servicio reutilizable que interpreta y valida `Idempotency-Key`.
- Huella canónica del contenido y coordinación con las funciones PostgreSQL ya existentes.
- Reproducción exacta de la respuesta persistida, incluidos estado y cuerpo definidos por el contrato.
- Semántica explícita para primera ejecución, repetición, conflicto, operación en curso, vencimiento, aborto y error.
- Declaración de cabecera y respuestas en OpenAPI.
- Pruebas con un handler disponible solo dentro del paquete de pruebas.
- Documentación que haga obligatorio el mecanismo para toda escritura crítica futura.

## Fuera de alcance

- Crear citas, clientes, servicios u otra operación de negocio.
- Publicar un endpoint de demostración o prueba en el router o contrato de producción.
- Rediseñar el esquema aprobado sin una contradicción comprobada y una decisión nueva.
- Editar una migración ya aplicada, aunque parezca más simple que crear una nueva.
- Hacer llamadas de red o trabajo lento dentro de una transacción de base de datos.

## Estado existente que debes conservar

- `idempotency_record`, sus políticas RLS y las funciones `idempotency_begin`, `idempotency_complete`, `idempotency_abort` y `purge_expired_idempotency_records` ya existen mediante migraciones aplicadas.
- Las migraciones de endurecimiento fijan roles, bloqueo consultivo transaccional, respuestas inmediatas para conflictos concurrentes y formato de huella. Son inmutables.
- El fingerprint esperado es SHA-256 en hexadecimal minúsculo del contenido canónico acordado; valida longitud y formato antes de invocar PostgreSQL.
- Toda operación tenant-aware usa el límite transaccional de `HU-002`; dominio y servicios no importan PostgreSQL ni Chi.
- Los problemas HTTP y la correlación vienen de `HU-003` y deben reutilizarse.

## Trabajo requerido

1. Audita la API de las funciones SQL y las pruebas existentes. No dupliques en Go decisiones que PostgreSQL ya garantiza de forma atómica.
2. Diseña puertos en la capa de aplicación para comenzar, completar y abortar una ejecución, con un adaptador PostgreSQL concreto. Mantén las dependencias hacia adentro.
3. Valida `Idempotency-Key`: ausencia cuando sea obligatoria, límites, caracteres permitidos y normalización deben coincidir con OpenAPI. Nunca registres la clave literal si pudiera ser sensible; usa una referencia opaca.
4. Calcula la huella de una representación canónica estable. Documenta exactamente qué método, ruta y cuerpo participan para evitar colisiones semánticas.
5. Traduce el resultado de `idempotency_begin` a estados inequívocos: ejecutar, reproducir, conflicto de contenido y operación bloqueada/en curso. Aplica los códigos confirmados por `DEC-043`, incluidos `409`/`425` donde corresponda; no esperes indefinidamente.
6. Ejecuta el efecto una sola vez. Completa el registro únicamente después del éxito y persiste la respuesta necesaria para una reproducción coherente.
7. Ante error, pánico o cancelación de contexto, aborta de forma segura si corresponde. Nunca conviertas una ejecución ya completada en abortada ni dejes una clave fallida bloqueando un reintento legítimo.
8. Mantén corta cualquier transacción propia de coordinación. Si la atomicidad con el efecto requiere una decisión que las fuentes no resuelven, registra la contradicción y detén esa parte en vez de inventar una garantía.
9. Declara en OpenAPI `Idempotency-Key`, sus restricciones, la reproducción y todos los problemas posibles. No agregues un endpoint artificial al bundle.

## Pruebas y evidencia

Usa PostgreSQL real 14 o la versión fijada por el proyecto, al menos dos tenants y concurrencia real. Cubre:

- primera ejecución y persistencia de la respuesta;
- repetición con misma clave/huella que reproduce exactamente y no repite el contador de efecto;
- misma clave con huella distinta que devuelve `409` sin efecto;
- dos conexiones simultáneas con la misma clave: un solo efecto y respuestas coherentes;
- estado bloqueado/en curso con respuesta inmediata conforme a `DEC-043`;
- misma clave literal en dos barberías sin interferencia;
- registro vencido que permite una evaluación nueva;
- error, pánico y cancelación que dejan reintentar;
- clave y huella inválidas;
- garantía de que una ejecución completada no se aborta después;
- carrera con `go test -race` en el código Go aplicable;
- contrato OpenAPI sin endpoint ficticio de producción.

Entrega una tabla `Criterio | Estado | Prueba o evidencia` para `CA-004-01` a `CA-004-06` y separa evidencia SQL, integración HTTP y carrera.

## Documentación y trazabilidad

- Documenta el patrón obligatorio y un ejemplo seguro en `apps/api/README.md`.
- Actualiza OpenAPI, matriz, historial, diagrama/diccionario de datos solo si el resultado realmente los afecta.
- Si necesitas una migración correctiva, créala con Atlas, evalúa bloqueo/recuperación y versiona `atlas.sum`; nunca edites las existentes.
- Actualiza metadatos e índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

Ejecuta los comandos equivalentes del repositorio para:

```text
Atlas migrate lint y migrate validate
pruebas SQL de idempotencia y concurrencia a dos conexiones
formato, vet, pruebas y race del módulo Go
pruebas HTTP y de contrato
openapi:lint
openapi:bundle
git diff --check
graphify update .
```

No simules PostgreSQL con mocks para afirmar los criterios concurrentes y no reduzcas umbrales para obtener verde.

## Git y PR

- Commits y título: `feat(api): implementa idempotencia reutilizable` o equivalente Conventional Commits.
- Abre un PR borrador contra `main`; usa `Closes #<issue>` solo al cubrir el issue completo.
- Incluye resultados de concurrencia, estrategia de atomicidad, contrato, migraciones si las hubo, riesgo de bloqueo y recuperación.
- No hagas merge, push directo, force push ni reescribas `main`.
