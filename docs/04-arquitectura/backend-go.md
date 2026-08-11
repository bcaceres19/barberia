---
titulo: "Arquitectura del backend en Go"
version: "1.2"
estado: "Decisión confirmada"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/reglas-negocio.md"
  - "../05-backend/base-datos.md"
  - "../03-desarrollo/estandar-backend-go.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../06-api/estandar-openapi.md"
  - "stack-despliegue-operacion.md"
---

# Arquitectura del backend en Go

## 1. Decisión

El backend usa **Chi v5** como router y compositor de middleware sobre la biblioteca estándar `net/http`.

Chi no gobierna la lógica de negocio, las transacciones ni la persistencia. Todos los handlers conservan las firmas estándar de Go y los módulos internos no importan Chi.

Fuente normativa: `DEC-034`.

## 2. Por qué no usar solo `net/http`

`net/http.ServeMux` ya soporta métodos y parámetros de ruta, por lo que vanilla es técnicamente viable. Sin embargo, este proyecto necesita varios árboles y cadenas de middleware:

- rutas públicas de reserva;
- acceso del cliente mediante token;
- rutas privadas autenticadas;
- contexto de barbería y RLS;
- limitación de abuso;
- idempotencia;
- endpoints operativos y de salud.

Chi agrega grupos, subrouters y middleware componible manteniendo `http.Handler`. Esa reducción de código repetitivo supera el costo de una única dependencia pequeña y reemplazable.

## 3. Por qué no un framework más amplio

No se adopta Gin, Echo, Fiber ni un framework MVC porque el MVP no necesita:

- tipos propios de contexto y respuesta;
- un ORM incluido;
- inyección de dependencias automática;
- generación de controladores;
- ciclo de vida administrado por un framework;
- mecanismos alternativos a `net/http`.

La decisión no afirma que esos frameworks sean lentos o inadecuados en general. Solo evita superficie que este monolito pequeño no necesita.

## 4. Estructura inicial

```text
apps/api/
  cmd/
    api/
    worker/
  internal/
    platform/
      clock/
      config/
      database/
      httpserver/
      observability/
    modules/
      auth/
      shops/
      staff/
      catalog/
      schedule/
      booking/
      notification/
      audit/
```

Dentro de cada módulo:

```text
module/
  domain.go
  errors.go
  ports.go
  service.go
  httpapi/
    dto.go
    handler.go
    routes.go
  postgres/
    repository.go
```

Las interfaces se declaran cerca del código que las consume. No se crea una carpeta global de interfaces o utilidades.

## 5. Flujo de una solicitud

`http.Handler` → decodificación y validación de forma → servicio de aplicación → repositorio/transacción → respuesta HTTP.

Reglas:

1. El handler no contiene reglas de negocio ni SQL.
2. El servicio no conoce Chi, rutas, encabezados ni códigos HTTP.
3. El repositorio no decide permisos de negocio.
4. Toda dependencia llega mediante constructores explícitos.
5. Los errores de dominio se traducen a HTTP en un punto común.
6. Los procesos `api` y `worker` reutilizan servicios y repositorios, no handlers.

## 6. Orden de middleware

Orden base:

1. identificador de solicitud;
2. recuperación controlada de `panic`;
3. límite de tamaño del cuerpo;
4. timeout y cancelación mediante `context.Context`;
5. registro estructurado con datos personales redactados;
6. encabezados de seguridad;
7. límite de solicitudes para rutas públicas;
8. autenticación para rutas privadas;
9. selección autorizada de barbería;
10. contexto transaccional de tenant/RLS cuando el caso de uso toca datos;
11. idempotencia en operaciones críticas.

La IP del cliente solo se toma de `RemoteAddr` o de encabezados emitidos por proxies expresamente confiables. No se confía ciegamente en `X-Forwarded-For`.

## 7. Routing

Estructura conceptual:

```text
/health
/api/v1/public/barbershops/{slug}/...
/api/v1/customer/appointments/{token}/...
/api/v1/private/auth/...
/api/v1/private/appointments/...
/api/v1/private/services/...
/api/v1/private/schedules/...
/api/v1/private/settings/...
```

Los subrouters públicos, de cliente y privados reciben cadenas de middleware distintas. El tenant no se acepta como dato confiable del cuerpo: se deriva del recurso público o de la identidad autorizada.

## 8. Dependencias permitidas

Chi se usa únicamente mediante:

- `github.com/go-chi/chi/v5`;
- middlewares concretos revisados individualmente.

Agregar otro paquete HTTP requiere justificar:

- problema que resuelve;
- alternativa con biblioteca estándar;
- dependencias transitivas;
- tratamiento de contexto y cancelación;
- impacto en seguridad y observabilidad.

La versión menor se fija en `go.mod` y se actualiza mediante revisión, pruebas y notas de cambios; la decisión arquitectónica solo fija la rama mayor v5.

## 9. Criterios de aceptación técnicos

- handlers y middleware satisfacen interfaces `net/http`;
- pruebas de servicios no levantan servidor HTTP;
- cada ruta privada prueba aislamiento entre dos barberías;
- timeouts cancelan consultas PostgreSQL;
- el mismo `request_id` conecta logs HTTP, transacción e intento de notificación;
- ningún log contiene nombre, teléfono, correo, token o contenido de mensaje;
- rutas y middlewares pueden inventariarse automáticamente;
- cada operación HTTP implementada existe en OpenAPI 3.1.2 y pasa pruebas de contrato;
- retirar Chi exigiría cambiar routing, no servicios ni repositorios.

## 10. Estándares de implementación

La estructura de paquetes, las reglas de código limpio, documentación y revisión son obligatorias según [estandar-backend-go.md](../03-desarrollo/estandar-backend-go.md). Las pruebas por capa y riesgos P0 se definen en [estrategia-pruebas.md](../03-desarrollo/estrategia-pruebas.md). El contrato HTTP se rige por [estandar-openapi.md](../06-api/estandar-openapi.md).

La organización detallada, incluida la regla de crear solo carpetas necesarias, está en el estándar de desarrollo.

## 11. Referencias oficiales

- [Chi](https://github.com/go-chi/chi)
- [Mejoras de routing en Go 1.22](https://go.dev/blog/routing-enhancements)
- [`net/http`](https://pkg.go.dev/net/http)
