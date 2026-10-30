---
prompt_id: "PROMPT-TEST-87-AISLAMIENTO-CROSS-TENANT-HTTP-v1"
report_for: "docs/10-backlog/prompts/test/issue-87-aislamiento-cross-tenant-http.md"
issue: 87
issue_url: "https://github.com/bcaceres19/barberia/issues/87"
executed_at: "2026-10-30"
executed_by: "agente Claude (sesión interactiva), operador bcaceres19"
environment: "efímero: postgres:14 en :55432 con 39 migraciones y fixtures de CI; API compilado en :8081 (APP_ENVIRONMENT=local); Vite en :5174; Chromium de Playwright"
---

# Evidencia: aislamiento cross-tenant por HTTP real (issue #87)

## Resumen

Dos sesiones reales (tenant A `duena.a`, tenant B `dueno.b`, iniciadas por la UI, con el reto de
acceso resuelto con el código capturado del entorno local) y `fetch` ejecutado **dentro** de la
sesión A con `credentials: 'same-origin'`: la cookie `HttpOnly` nunca se lee ni se imprime.
Todas las operaciones de A sobre IDs válidos de B responden `404 not-found`, con el mismo cuerpo
que un ID inexistente, y B conserva exactamente sus valores originales.

Los recursos de B se crearon desde la sesión de B (el fixture no trae barberos ni servicios):
un barbero, un servicio y su asignación. Toda mutación incluyó `Idempotency-Key` para que un
`400` por cabecera ausente no pudiera enmascarar el `404`.

## Matriz

| Operación (sesión A, ID de B)                               | Esperado | Real | `problem.code`    | Integridad posterior |
| ----------------------------------------------------------- | -------- | ---- | ----------------- | -------------------- |
| `GET /private/barbers/{barberB}`                            | 404      | 404  | `not-found`       | B intacto            |
| `PATCH /private/barbers/{barberB}` (renombrar)              | 404      | 404  | `not-found`       | B intacto            |
| `GET /private/services/{serviceB}`                          | 404      | 404  | `not-found`       | B intacto            |
| `PATCH /private/services/{serviceB}`                        | 404      | 404  | `not-found`       | B intacto            |
| `POST /private/services/{serviceB}/deactivate`              | 404      | 404  | `not-found`       | B intacto            |
| `POST /private/services/{serviceB}/reactivate`              | 404      | 404  | `not-found`       | B intacto            |
| `GET /private/services/{serviceB}/deactivation-impact`      | 404      | 404  | `not-found`       | B intacto            |
| `GET /private/barbers/{barberB}/services`                   | 404      | 404  | `not-found`       | B intacto            |
| `PUT /private/barbers/{barberB}/services/{serviceB}`        | 404      | 404  | `not-found`       | B intacto            |
| `DELETE /private/barbers/{barberB}/services/{serviceB}`     | 404      | 404  | `not-found`       | B intacto            |
| `PATCH /private/barbers/{barberB}` con `barbershopId` ajeno | 400      | 400  | `invalid-request` | B intacto            |
| Control: `GET` barbero inexistente                          | 404      | 404  | `not-found`       | —                    |
| Control: `GET` servicio inexistente                         | 404      | 404  | `not-found`       | —                    |
| Control: `GET` barbero de B sin sesión                      | 401      | 401  | `unauthorized`    | —                    |

La comprobación posterior compara, desde la sesión B, el barbero, el servicio y la lista de
asignaciones antes y después de toda la matriz: idénticos byte a byte.

## Datos sensibles revisados

| Dato sensible revisado                                                      | Resultado                                                                                                    |
| --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| El cuerpo de los `404` contiene el ID de B o el `barbershopId` del tenant B | No (comprobado por búsqueda en cada cuerpo)                                                                  |
| Tamaño del cuerpo del `404` de recurso ajeno frente al de uno inexistente   | Igual en lectura y mutación (248/249 bytes); el `DELETE` de asignación mide 259 por su texto propio, sin IDs |
| Cookie, cabeceras de autorización, DSN o credenciales en la evidencia       | No se registran; las credenciales son sintéticas y la BD se destruye al terminar                             |
| Causa RLS, SQL o nombres de tablas en las respuestas                        | No                                                                                                           |

## Conclusión

Sin hallazgos: no hace falta un fix de seguridad. Cubre la próxima acción de #87.
