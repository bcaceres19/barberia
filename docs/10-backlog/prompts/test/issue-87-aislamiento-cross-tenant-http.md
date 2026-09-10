---
prompt_id: "PROMPT-TEST-87-AISLAMIENTO-CROSS-TENANT-HTTP-v1"
version: "1.0"
kind: "test"
status: "ready"
target_agents:
  - "codex"
  - "claude"
  - "human"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-021"
related_hu:
  - "HU-022"
  - "HU-023"
issue: 87
issue_url: "https://github.com/bcaceres19/barberia/issues/87"
suggested_issue_title: null
branch: null
pr: null
pr_url: null
depends_on:
  - "Dos tenants sintéticos y sesiones QA separadas en un entorno local controlado"
rules:
  - "RN-TEN-01"
  - "RN-DAT-02"
decisions:
  - "DEC-024"
  - "DEC-035"
acceptance_criteria:
  - "Una sesión del tenant A que consulta IDs existentes del tenant B recibe 404 uniforme."
  - "La prueba cubre lectura y mutación en barberos, servicios y asignaciones sin alterar datos del tenant B."
  - "Network confirma status, método y ruta; cuerpo y logs no revelan existencia, tenant ni datos personales."
  - "La evidencia usa datos sintéticos y no publica cookies, tokens, correos, teléfonos ni credenciales."
  - "Las pruebas automatizadas PostgreSQL/HTTP equivalentes siguen en verde."
source_docs:
  - "AGENTS.md"
  - "CONTRIBUTING.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/05-backend/estandar-base-datos.md"
  - "docs/10-backlog/prompts/test/2026-08-25-qa-manual-b0-b1-chrome-mcp-report.md"
  - "apps/api/cmd/api/staff_integration_test.go"
  - "apps/api/cmd/api/catalog_integration_test.go"
  - "apps/api/cmd/api/barber_services_integration_test.go"
created_at: "2026-09-04"
updated_at: "2026-09-04"
supersedes: null
superseded_by: null
---

# Verificar aislamiento cross-tenant por HTTP real

## Instrucción

Ejecuta la comprobación directa pendiente del issue [#87](https://github.com/bcaceres19/barberia/issues/87) contra la aplicación real. El backend de `HU-021`, `HU-022` y `HU-023` ya está implementado y tiene pruebas automatizadas; este trabajo aporta evidencia manual/HTTP adicional y solo corrige código si primero reproduce un defecto real.

## Objetivo

Demostrar que conocer un ID válido de otra barbería no permite leerlo ni modificarlo: una sesión del tenant A recibe el mismo `404` que recibiría por un recurso inexistente, y el recurso del tenant B queda intacto.

## Preflight

1. Lee completos los `source_docs`, comprueba que #87 sigue abierto y preserva cambios ajenos.
2. Arranca el stack oficial con PostgreSQL 14 real, RLS forzada y dos tenants sintéticos. No uses producción.
3. Crea o identifica en el tenant B un barbero, un servicio y una asignación con IDs conocidos de forma segura.
4. Inicia sesión por la UI como tenant A. Introduce credenciales solo en el navegador; no las copies a comandos, archivos, logs o evidencia.
5. Abre DevTools Network/Console o un controlador equivalente que ejecute `fetch` en el contexto autenticado sin revelar la cookie `HttpOnly`.

## Matriz mínima

Ejecuta con la sesión A:

- leer el barbero B;
- renombrar o actualizar el barbero B;
- leer/listar el servicio B por una operación que acepte ID;
- actualizar o cambiar el ciclo de vida del servicio B si el contrato vigente lo permite;
- listar o modificar asignaciones usando el barbero/servicio B;
- enviar un cuerpo con `barbershopId` ajeno donde el schema lo prohíbe y confirmar `400`, sin confundirlo con la prueba de RLS `404`.

Para cada caso registra método, ruta sanitizada, status, `problem.code` y comprobación posterior desde la sesión B. El resultado esperado para un recurso ajeno bien formado es `404`; autenticación ausente usa `401` y un cuerpo con campo desconocido puede usar `400`, según el contrato.

## Fuera de alcance

- Desactivar RLS, consultar como propietario de tablas o inyectar cookies desde archivos.
- Probar solo ocultamiento de UI; la evidencia debe incluir solicitudes HTTP reales.
- Cambiar IDs, permisos, rutas o respuestas para fabricar el resultado.
- Publicar cuerpos con datos personales, cookies, cabeceras de autorización, DSN o credenciales.
- Ampliar la prueba a módulos no relacionados sin un issue propio.

## Verificación y tratamiento de hallazgos

1. Ejecuta las pruebas Go de integración afectadas con `-race` y PostgreSQL real.
2. Confirma que la sesión B sigue leyendo los valores originales después de cada mutación rechazada.
3. Revisa logs y respuestas para asegurar que no exponen `barbershopId`, nombre sensible, SQL o causa RLS.
4. Si todo pasa, persiste un informe breve y comenta #87 con evidencia segura; cierra el issue solo si cubre toda su próxima acción.
5. Si un caso devuelve datos, `200`, `403`, `500` o modifica B, detén la campaña, conserva evidencia sanitizada y abre un fix de seguridad separado. No corrijas silenciosamente dentro de este prompt de prueba.

Entrega:

`Operación | ID propio/ajeno | Esperado | Real | Integridad posterior | Evidencia`

`Dato sensible revisado | Resultado`

Si solo se produce el informe, usa una rama `test/87-aislamiento-cross-tenant-http` y un PR `test(api): verifica aislamiento cross-tenant por HTTP`, salvo que el proceso del repositorio permita cerrar el issue únicamente con comentario y evidencia ya versionada. No inventes rama ni PR si no existen.
