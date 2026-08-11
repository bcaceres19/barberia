---
titulo: "Estándar de documentación OpenAPI"
version: "1.0"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../00-control/matriz-trazabilidad.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../03-desarrollo/estandar-backend-go.md"
  - "../03-desarrollo/estandar-frontend-vue.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../04-arquitectura/backend-go.md"
---

# Estándar de documentación OpenAPI

## 1. Propósito y autoridad

OpenAPI es la fuente de verdad del contrato HTTP entre el frontend Vue y el backend Go. Describe lo que un consumidor puede enviar, recibir y asumir; no documenta detalles internos de servicios, tablas o proveedores.

El contrato se diseña antes o junto con cada entrega vertical. Un handler no se considera terminado si su comportamiento observable contradice OpenAPI, y un cambio en OpenAPI no se considera terminado sin implementación y pruebas compatibles.

Fuente normativa: `DEC-037`.

## 2. Versión y herramientas

| Pieza | Elección |
| --- | --- |
| Especificación | OpenAPI 3.1.2 |
| Esquemas | JSON Schema Draft 2020-12 mediante el dialecto de OpenAPI 3.1 |
| Fuente | YAML 1.2 multiarchivo, UTF-8 |
| Lint y bundle | Redocly CLI v2 fijado en el lockfile |
| Documentación navegable | HTML estático generado por Redocly para uso interno |
| Contrato distribuible | Un único `openapi.yaml` bundled como artefacto de CI |

OpenAPI 3.2 se evaluará cuando las herramientas obligatorias de lint, bundle, documentación y generación usadas por el proyecto lo soporten de forma estable. No se cambia la versión solo por ser la más reciente.

No se ejecuta `npx ...@latest` en CI. La versión de Redocly CLI se fija como dependencia de desarrollo y se actualiza mediante revisión.

## 3. Organización del contrato

```text
api/
  openapi/
    openapi.yaml
    paths/
      public-booking.yaml
      customer-appointments.yaml
      private-appointments.yaml
      catalog.yaml
      schedules.yaml
      settings.yaml
    components/
      headers/
      parameters/
      responses/
      schemas/
      security-schemes/
    examples/
    CHANGELOG.md
    dist/                 # generado; no es fuente
redocly.yaml
```

Reglas:

1. `api/openapi/openapi.yaml` es el documento de entrada.
2. Los archivos `paths` se agrupan por capacidad y audiencia, no uno por método HTTP.
3. Los componentes reutilizables tienen un dueño semántico y nombre único.
4. Las referencias `$ref` son relativas al archivo que las contiene y deben resolver en el bundle.
5. No crear un archivo para cada esquema trivial; dividir cuando mejore navegación y propiedad.
6. `dist/` se regenera y publica como artefacto; no se edita manualmente.
7. La configuración `redocly.yaml` vive en la raíz y define el alias `core@v1`, reglas y documento de entrada.
8. El contrato no se genera a partir de comentarios Go. Puede generar tipos o interfaces de transporte, pero la descripción OpenAPI sigue siendo la fuente.

## 4. Audiencias, base URL y versionado

El API se divide conceptualmente en tres audiencias:

```text
/api/v1/public/...      # reserva sin sesión del barbero
/api/v1/customer/...    # acceso mediante credencial del enlace de la cita
/api/v1/private/...     # sesión autenticada del barbero
```

En OpenAPI se usa preferiblemente un servidor relativo:

```yaml
servers:
  - url: /api/v1
```

Así el mismo contrato sirve en local, CI, piloto y producción sin introducir hosts o secretos por ambiente.

Reglas de versión:

- `info.version` usa SemVer y representa la versión del contrato publicado.
- `/api/v1` representa la versión mayor del protocolo HTTP.
- Durante el desarrollo previo al primer contrato publicado pueden usarse versiones `0.x`.
- Después de publicar v1, un cambio incompatible exige coordinación explícita y normalmente `/api/v2`.
- Una versión nueva no elimina la anterior hasta que todos los consumidores hayan migrado y exista fecha de retiro comunicada.
- La versión del API no se ata a la versión interna del binario ni a una migración de base de datos.

## 5. Idioma y nomenclatura

- Identificadores técnicos, paths, propiedades, esquemas y `operationId`: inglés.
- `summary`, `description`, ejemplos explicativos y mensajes de documentación: español.
- JSON: `camelCase`.
- Parámetros de path: nombres completos en `camelCase`, como `{appointmentId}`.
- Esquemas: `PascalCase`, con intención y dirección: `CreateAppointmentRequest`, `AppointmentResponse`.
- `operationId`: único, estable y en `camelCase`, con verbo e intención: `createPublicAppointment`, `cancelCustomerAppointment`.
- Tags: capacidades estables y en `PascalCase`: `PublicBooking`, `Appointments`, `Schedules`.
- Códigos de problema: minúscula con guiones, por ejemplo `slot-conflict`.
- No usar nombres genéricos como `Data`, `Payload`, `Model`, `Object`, `Response` o `Error1` sin contexto.

El usuario ve “turno”; los identificadores técnicos conservan `appointment`, según `DEC-016`.

## 6. Diseño de paths y métodos

1. Usar sustantivos plurales para colecciones: `/appointments`, `/services`.
2. Evitar verbos CRUD en la URL: no `/getAppointments` ni `/createAppointment`.
3. Modelar transiciones de negocio como subrecursos o comandos explícitos y consistentes cuando un simple `PATCH` oculte precondiciones importantes.
4. No anidar más de lo necesario. La pertenencia a barbería se deriva de identidad o recurso, no de un `barbershopId` confiado del body.
5. `GET` no tiene body y no produce efectos de negocio.
6. `POST` crea recursos o ejecuta comandos no idempotentes por naturaleza, protegidos con clave de idempotencia cuando sean críticos.
7. `PUT` reemplaza una representación completa; `PATCH` aplica una modificación parcial con campos permitidos explícitos.
8. `DELETE` no implica borrado físico cuando las reglas exigen historial; la descripción explica el efecto observable.
9. Filtros y orden se documentan mediante parámetros definidos, no mediante una cadena libre.
10. Listas potencialmente crecientes usan paginación por cursor; `limit` tiene mínimo, máximo y valor inicial.

## 7. Contenido obligatorio de cada operación

Toda operación incluye:

- un tag declarado;
- `summary` breve e inequívoco;
- `description` con precondiciones, efectos y límites no evidentes;
- `operationId` único y estable;
- `security` explícito, incluso `security: []` para una operación pública;
- parámetros con descripción, tipo, formato, rango y ejemplo;
- `requestBody` con `required` explícito cuando exista;
- todas las respuestas normales y de error que el consumidor deba manejar;
- ejemplos de éxito y al menos un error relevante;
- headers relevantes como `X-Request-Id`, `Idempotency-Key`, `Location` o `Retry-After`;
- extensiones de trazabilidad `x-business-rules` y `x-decisions` cuando corresponda;
- indicación de idempotencia, rate limit o efecto asíncrono cuando aplique.

La descripción explica comportamiento, no implementación. No menciona nombres de tablas, SQL, structs Go, funciones privadas o detalles del proveedor de WhatsApp.

## 8. Trazabilidad del negocio

Se permiten extensiones OpenAPI con prefijo `x-`:

```yaml
x-business-rules:
  - RN-DIS-02
  - RN-CON-05
x-decisions:
  - DEC-005
  - DEC-020
```

Reglas:

1. Las extensiones enlazan comportamiento normativo, no sustituyen `summary` o `description`.
2. Solo se citan códigos existentes.
3. Un cambio de regla actualiza la operación, las pruebas y la matriz en el mismo cambio.
4. Las extensiones internas pueden retirarse del bundle público mediante una transformación, pero permanecen en la fuente interna.
5. OpenAPI no es la fuente de verdad de una regla: la regla vive en los documentos normativos enlazados.

## 9. Esquemas y JSON Schema

### Reglas generales

1. Separar request, response y modelo interno. No reutilizar un esquema solo porque hoy comparte campos.
2. Declarar `required` de forma explícita y distinguir propiedad ausente de valor `null`.
3. En OpenAPI 3.1, representar `null` mediante JSON Schema, por ejemplo `type: [string, "null"]`; no usar `nullable: true`.
4. Usar `additionalProperties: false` en requests cerrados. En responses, permitir propiedades adicionales solo cuando sea una decisión de evolución consciente.
5. Para mapas reales, declarar el esquema de `additionalProperties`; no aceptar objetos libres sin justificación.
6. Usar `readOnly` y `writeOnly` para separar campos generados o secretos.
7. Proporcionar `description` para toda propiedad cuya unidad, zona, origen o semántica no sea obvia.
8. Usar `examples` compatibles con el schema; no depender del campo singular obsoleto `example` dentro de Schema Objects.
9. Evitar `allOf`, `oneOf`, discriminadores y herencia salvo que representen una variante real del contrato y las herramientas la soporten.
10. No usar `format` como única validación: agregar patrón, rango o descripción cuando el negocio lo requiera.

### Tipos del proyecto

| Concepto | Representación HTTP |
| --- | --- |
| ID de dominio | `string`, `format: uuid` |
| Instante | `string`, `format: date-time`, RFC 3339 |
| Fecha local | `string`, `format: date` |
| Zona horaria | nombre IANA, como `America/Bogota` |
| Duración | entero en minutos con mínimo y máximo |
| Importe exacto | string decimal más código de moneda, salvo decisión posterior |
| Teléfono | string normalizado; nunca número |
| Estado | enum documentado y alineado con la máquina de estados |
| Token o contraseña | `writeOnly: true`, sin ejemplo real |

Los instantes se transportan con offset. Una fecha u hora mostrada al cliente siempre se interpreta en la zona de la barbería.

## 10. Requests y validación

- `Content-Type` normal: `application/json`.
- Rechazar campos desconocidos en requests cerrados.
- Documentar límites de longitud, valores, patrones y cardinalidad que el backend aplica realmente.
- Un campo obligatorio en producto aparece en `required`; no basta una descripción que diga “obligatorio”.
- No incluir `barbershopId`, permisos, estado derivado o actor como entrada confiable cuando el servidor puede derivarlo.
- Requests de creación crítica declaran `Idempotency-Key` como header requerido.
- La misma clave con el mismo contenido devuelve el mismo resultado lógico; con contenido distinto produce conflicto documentado.
- Archivos o contenido binario solo se agregan cuando exista un caso aprobado; no usar base64 dentro de JSON por conveniencia.

## 11. Responses y códigos HTTP

| Código | Uso en este proyecto |
| --- | --- |
| `200` | lectura o modificación con representación de respuesta |
| `201` | creación completada; incluir `Location` cuando corresponda |
| `202` | trabajo aceptado que todavía no terminó |
| `204` | éxito sin body |
| `400` | JSON, parámetro o sintaxis inválida |
| `401` | credencial ausente, inválida o expirada |
| `403` | identidad conocida sin permiso, solo cuando no revele un recurso sensible |
| `404` | recurso inexistente o perteneciente a otra barbería |
| `409` | conflicto de agenda, estado, versión o idempotencia |
| `422` | contenido bien formado que incumple validaciones de campo o regla procesable |
| `429` | límite de solicitudes; documentar `Retry-After` cuando exista |
| `500` | fallo inesperado sin detalles internos |
| `503` | dependencia o capacidad temporalmente no disponible |

No documentar `default` como sustituto de errores conocidos. Cada operación lista los errores que el cliente debe manejar y puede reutilizar responses en `components`.

Un `204` no contiene body. Un `201` no se usa para operaciones que solo programan trabajo; un `202` no afirma que la operación finalizó.

## 12. Formato uniforme de errores

Los errores usan RFC 9457 con `application/problem+json`.

Campos base:

| Campo | Uso |
| --- | --- |
| `type` | URI estable del tipo de problema |
| `title` | resumen estable y legible |
| `status` | código HTTP |
| `detail` | explicación segura de esta ocurrencia |
| `instance` | identificador o URI de la ocurrencia cuando sea seguro |
| `code` | código estable del proyecto para lógica del cliente |
| `requestId` | correlación con logs y soporte |
| `errors` | lista opcional de errores por campo |

Reglas:

- El frontend decide por `status`, `type` o `code`, nunca comparando `detail`.
- `detail` no contiene SQL, stack traces, tokens, teléfonos, correos ni existencia de recursos de otro tenant.
- `errors[].field` usa una ruta estable del request y cada elemento tiene código y mensaje.
- Cada tipo de problema documenta cuándo ocurre y un ejemplo ficticio.
- Los problem types publicados son estables; cambiar su significado es un cambio incompatible.

## 13. Seguridad y privacidad en la documentación

1. Cada operación declara su seguridad; no confiar en una seguridad global implícita para mezclar rutas públicas y privadas.
2. Los esquemas de seguridad se definen una sola vez en `components/securitySchemes` y reflejan la implementación real.
3. El acceso público usa `security: []`; eso no elimina rate limiting, idempotencia ni controles de abuso.
4. La descripción privada indica que el tenant se deriva de la identidad autorizada.
5. Recursos de otra barbería responden como no encontrados y no tienen ejemplo que revele su existencia.
6. No incluir credenciales, cookies, tokens funcionales, hosts internos ni datos reales en ejemplos.
7. Contraseñas, códigos y tokens se marcan `writeOnly`; respuestas nunca los reutilizan salvo el único momento expresamente diseñado.
8. La documentación completa de rutas privadas se publica solo en un artefacto con acceso controlado.
9. Si se publica documentación externa, se genera una variante filtrada; no se mantiene un segundo contrato manual.
10. Markdown y HTML en descripciones se mantienen simples y sanitizables.

## 14. Ejemplos

Todo endpoint importante incluye ejemplos coherentes de request y response.

- Usar UUID reservados para documentación o claramente ficticios.
- Usar nombres como “Cliente Ejemplo”; no copiar datos de pruebas reales.
- Fechas deben mostrar offset y zona de barbería coherentes.
- Un ejemplo debe validar contra su schema.
- Proporcionar casos de frontera importantes: conflicto de franja, token expirado, límite de abuso y validación por campo.
- No usar valores que parezcan secretos reales.
- Mantener ejemplos grandes en `api/openapi/examples/` y referenciarlos.

## 15. Ejemplo de operación

```yaml
post:
  tags: [PublicBooking]
  summary: Reservar un turno disponible
  description: >
    Crea una cita confirmada si la franja continúa disponible. La operación es
    idempotente y la disponibilidad final siempre la decide el backend.
  operationId: createPublicAppointment
  security: []
  x-business-rules: [RN-DIS-02, RN-CON-05, RN-IDE-01]
  x-decisions: [DEC-019, DEC-022]
  parameters:
    - $ref: ../components/parameters/IdempotencyKey.yaml
  requestBody:
    required: true
    content:
      application/json:
        schema:
          $ref: ../components/schemas/CreateAppointmentRequest.yaml
  responses:
    "201":
      $ref: ../components/responses/AppointmentCreated.yaml
    "409":
      $ref: ../components/responses/SlotConflict.yaml
    "422":
      $ref: ../components/responses/ValidationProblem.yaml
    "429":
      $ref: ../components/responses/RateLimited.yaml
```

El ejemplo muestra estructura, no define por sí solo los campos definitivos del contrato.

## 16. Lint, bundle y documentación generada

Controles mínimos una vez exista el scaffold:

```text
pnpm exec redocly check-config
pnpm exec redocly lint core@v1
pnpm exec redocly bundle core@v1 --output api/openapi/dist/openapi.yaml
pnpm exec redocly build-docs core@v1 --output api/openapi/dist/index.html
```

Reglas de automatización:

1. CI usa el ruleset `recommended-strict` como base y reglas propias del proyecto.
2. Errores de lint o referencias rotas bloquean el cambio.
3. Warnings nuevos se corrigen o se convierten en una excepción explícita, localizada, justificada y con referencia.
4. Está prohibido generar un archivo global de ignores para hacer pasar una especificación incompleta.
5. El bundle se genera después del lint y se usa para codegen, pruebas de contrato y publicación.
6. CI comprueba que ejemplos validen y que todos los `operationId` sean únicos.
7. La documentación HTML es un artefacto, no la fuente editable.
8. El bundle y la documentación no incluyen extensiones internas o rutas privadas cuando el destino sea público.

## 17. Pruebas de contrato

- Cada handler se prueba contra status, media type, headers y schema documentados.
- Requests válidos del contrato deben ser aceptados salvo una regla de negocio documentada.
- Requests inválidos deben producir el problem type documentado.
- El cliente TypeScript se genera o tipa desde el bundle; no mantiene DTO duplicados manualmente.
- Código generado no se edita.
- Un E2E afectado verifica el recorrido, pero no reemplaza la validación de todos los schemas.
- CI detecta endpoints implementados sin documentar y operaciones documentadas sin implementación mediante la herramienta que se seleccione al crear el servidor.
- Cambios en ejemplos o schemas ejecutan nuevamente pruebas de integración y frontend afectadas.

## 18. Compatibilidad y cambios

### Generalmente compatibles

- agregar endpoint;
- agregar parámetro opcional;
- agregar campo opcional de response;
- agregar nuevo código de error solo si el cliente ya trata errores desconocidos de esa clase;
- ampliar una descripción sin cambiar comportamiento.

### Incompatibles o potencialmente incompatibles

- eliminar o renombrar path, operación, propiedad o enum;
- convertir un campo opcional en requerido;
- cambiar tipo, formato, unidad, zona horaria o semántica;
- restringir rangos o patrones aceptados;
- agregar valor a un enum consumido exhaustivamente por código generado;
- cambiar seguridad, idempotencia, status exitoso o media type;
- reutilizar `operationId`, problem type o código con otro significado.

Todo cambio incompatible incluye motivo, consumidores afectados, plan de transición y decisión. `CHANGELOG.md` registra versiones publicadas; Git conserva el detalle.

Una corrección de documentación que revela que la implementación ya era distinta no se clasifica automáticamente como “solo documentación”: primero se determina cuál comportamiento es normativo.

## 19. Revisión mínima de una operación

- [ ] Tiene tag, summary, description y `operationId` único.
- [ ] La audiencia y `security` son explícitas.
- [ ] Tenant, permisos e idempotencia están correctamente descritos.
- [ ] Request y response usan schemas direccionales y campos `required` reales.
- [ ] Tipos, rangos, fechas, zonas y enums coinciden con el backend.
- [ ] Respuestas de éxito y error son completas y usan media types correctos.
- [ ] Los errores usan RFC 9457 y no filtran información sensible.
- [ ] Ejemplos ficticios validan contra el schema.
- [ ] `RN-*` y `DEC-*` relevantes están enlazados.
- [ ] El cambio es compatible o tiene plan de transición.
- [ ] Lint, bundle, pruebas de contrato y consumidor frontend pasan.
- [ ] La documentación publicada no expone rutas o extensiones internas sin autorización.

## 20. Referencias oficiales

- [OpenAPI Specification 3.1.2](https://spec.openapis.org/oas/v3.1.2.html)
- [Versiones publicadas de OpenAPI](https://spec.openapis.org/oas/)
- [Redocly CLI](https://redocly.com/docs/cli)
- [Lint con Redocly](https://redocly.com/docs/cli/commands/lint)
- [Lint y bundle](https://redocly.com/docs/cli/guides/lint-and-bundle)
- [RFC 9457: Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457.html)

