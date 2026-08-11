---
titulo: "Estándar de código del backend en Go"
version: "1.1"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/reglas-negocio.md"
  - "../04-arquitectura/backend-go.md"
  - "estrategia-pruebas.md"
  - "../05-backend/estandar-base-datos.md"
  - "../06-api/estandar-openapi.md"
---

# Estándar de código del backend en Go

## 1. Propósito y alcance

Estas reglas aplican al API, al trabajador de notificaciones, a tareas administrativas y a cualquier paquete Go del MVP. Adaptan código limpio a un monolito modular pequeño: claridad, cohesión y comportamiento verificable tienen prioridad sobre patrones, capas o abstracciones por sí mismas.

Una excepción es válida cuando simplifica un caso real y queda explicada en la revisión. No se aceptan excepciones para aislamiento entre barberías, integridad de agenda, protección de datos o pruebas de reglas P0.

Fuente normativa: `DEC-035`.

## 2. Ubicación y estructura del módulo Go

El backend vive en `apps/api` y contiene un único módulo Go mientras no exista una razón demostrable para dividirlo.

```text
apps/api/
  cmd/
    api/
      main.go
    worker/
      main.go
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
  go.mod
  go.sum
```

Significado de los módulos:

| Paquete | Responsabilidad |
| --- | --- |
| `auth` | identidad, sesiones y recuperación de acceso |
| `shops` | barbería, configuración y zona horaria |
| `staff` | barberos y su pertenencia a la barbería |
| `catalog` | servicios ofrecidos, duración y precio informativo |
| `schedule` | horarios, excepciones, festivos y bloqueos |
| `booking` | disponibilidad, creación y ciclo de vida de citas |
| `notification` | programación, intentos y adaptadores de canal |
| `audit` | historial inmutable y evidencia operativa |

`main.go` solo carga configuración, construye dependencias, registra rutas o trabajos y controla el apagado. No contiene reglas de negocio, SQL ni lectura directa de variables de entorno fuera del paquete de configuración.

## 3. Organización interna de un módulo

El núcleo de un módulo permanece en su paquete raíz. Los adaptadores dependen del núcleo; el núcleo no depende de ellos.

```text
internal/modules/booking/
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

No todas las carpetas se crean de antemano. Se agregan cuando el módulo tenga ese adaptador. Un archivo se divide cuando mejora la localización de responsabilidades, no para cumplir una cantidad fija de líneas.

Dirección permitida de dependencias:

```text
cmd → adaptadores → núcleo del módulo
cmd → platform
adaptadores → platform
núcleo del módulo → biblioteca estándar y contratos mínimos consumidos
```

Reglas de paquetes:

1. Organizar por capacidad de negocio, no en carpetas globales `controllers`, `models`, `services` y `repositories`.
2. Usar nombres cortos, en minúscula y sin guion bajo; el nombre debe describir una sola responsabilidad.
3. Declarar interfaces junto al consumidor. No crear `interfaces`, `common`, `helpers` o `utils` globales.
4. Evitar ciclos. Una interacción entre módulos pasa por una operación explícita o por un contrato pequeño, no por acceso a tablas ajenas.
5. Mantener `internal/platform` libre de reglas de barbería, agenda o citas.
6. Exportar solo lo necesario para componer adaptadores y módulos. No exportar para facilitar una prueba.
7. Introducir una abstracción después de identificar variación real; la similitud accidental de dos funciones no basta.

## 4. Responsabilidades por capa local

### Dominio y aplicación

- Los tipos representan términos del glosario: `Appointment`, `Barbershop`, `TimeBlock`, no nombres genéricos como `Item`, `Data` o `Manager`.
- El dominio protege invariantes que pueden comprobarse sin infraestructura.
- El servicio de aplicación coordina permisos, reloj, transacción, repositorios y efectos persistidos.
- Los servicios no conocen Chi, encabezados HTTP, JSON, SQL ni códigos de estado.
- Los DTO de entrada y salida no se reutilizan como entidades de dominio.
- Una regla no se duplica entre handler, servicio y repositorio; existe una autoridad clara y las demás capas solo aplican defensas complementarias.

### HTTP

- Chi se limita a rutas, parámetros y composición de middleware.
- Los handlers decodifican, validan la forma, invocan un caso de uso y traducen el resultado a HTTP.
- Cada cuerpo tiene límite de tamaño y rechaza campos desconocidos cuando el contrato sea controlado por el proyecto.
- Las respuestas de error usan una forma estable y no filtran SQL, stack traces, tokens o existencia de recursos de otro tenant.
- El handler no abre transacciones manuales ni ejecuta consultas.
- Autenticación, tenant y `request_id` llegan mediante contexto tipado; nunca mediante claves de texto repetidas.

### PostgreSQL

- El repositorio traduce entre filas y tipos del módulo; no contiene decisiones de negocio.
- Las consultas tienen columnas explícitas y contexto de barbería.
- El caso de uso define la frontera transaccional. No se hace una llamada de red con una transacción abierta.
- No se incorpora ORM en el MVP. Las consultas SQL permanecen visibles, revisables y probadas contra PostgreSQL real.
- Las restricciones de base de datos son la última defensa ante concurrencia; un `SELECT` previo no sustituye una restricción.

## 5. Reglas de código limpio

### Nombres y funciones

1. Usar nombres que expresen intención y vocabulario del dominio.
2. Nombrar funciones con verbos: `CreateAppointment`, `Cancel`, `FindAvailableSlots`.
3. Usar `is`, `has`, `can` o una pregunta de dominio para booleanos; evitar banderas ambiguas.
4. Mantener una función en un nivel de abstracción. Extraer pasos cuando tengan un nombre útil o una prueba propia.
5. Preferir retornos tempranos a condicionales profundamente anidados.
6. Evitar parámetros booleanos que cambian por completo el comportamiento; usar operaciones u opciones tipadas.
7. No usar valores mágicos. Duraciones, límites y estados se expresan con tipos, constantes o configuración validada.
8. No crear funciones genéricas antes de tener usos reales y semánticamente equivalentes.

### Tipos y estado

9. Hacer inválidos los estados imposibles cuando resulte simple: estados cerrados, duraciones positivas e intervalos con fin posterior al inicio.
10. Preferir tipos de dominio a cadenas o enteros sueltos en fronteras importantes.
11. No usar mapas sin tipo para transportar datos internos.
12. Evitar estado global mutable. Configuración, reloj, generador de identificadores y clientes externos se inyectan por constructor.
13. Copiar o proteger estructuras compartidas; todo acceso concurrente debe tener propiedad y sincronización evidentes.

### Errores y contexto

14. Devolver errores; no usar `panic` para entradas, fallos de proveedor o condiciones de negocio esperables.
15. Envolver con `%w` y contexto operativo breve. No comparar texto de errores.
16. Definir errores de dominio o tipos cuando el llamador necesite decidir una respuesta.
17. Registrar un error una sola vez en la frontera que conoce la operación y su resultado.
18. `context.Context` es el primer parámetro, no se guarda en structs y se propaga hasta PostgreSQL o proveedores.
19. Respetar cancelación y timeout; no iniciar goroutines huérfanas desde handlers.

### Simplicidad y dependencias

20. Preferir biblioteca estándar y dependencias pequeñas con propósito concreto.
21. Toda dependencia nueva debe justificar necesidad, mantenimiento, licencia, superficie transitiva y efecto en seguridad.
22. No introducir contenedor de inyección, reflexión para registrar módulos ni generación que oculte el flujo principal.
23. Duplicación pequeña y clara es preferible a una abstracción equivocada; al tercer caso estable se evalúa extraer.
24. El código generado se identifica y no se edita manualmente.

## 6. Documentación dentro del código

1. Cada paquete tiene un comentario que explica responsabilidad y límites; en paquetes amplios se usa `doc.go`.
2. Todo identificador exportado tiene comentario Go Doc que comienza con su nombre y explica su contrato.
3. Los comentarios explican el **porqué**, la invariante, el riesgo o una decisión no obvia; no traducen línea por línea el código.
4. Reglas temporales, RLS, idempotencia y decisiones de concurrencia incluyen referencia a la regla `RN-*` o decisión `DEC-*` cuando ayude a evitar una modificación peligrosa.
5. Un `TODO` incluye condición concreta y referencia rastreable: `TODO(#123): ...`. No se aceptan `TODO` anónimos permanentes.
6. Los ejemplos públicos se mantienen compilables cuando sean necesarios.
7. Cambiar una operación HTTP exige actualizar el contrato OpenAPI en el mismo cambio; el comentario del handler no lo sustituye.
8. No incluir secretos, datos personales ni ejemplos con información real en comentarios, fixtures o documentación generada.

## 7. Formato y controles automáticos

Antes de integrar un cambio Go deben pasar, cuando el código ya exista:

```text
gofmt sin diferencias
go vet ./...
go test ./...
go test -race ./...        # en CI programada y antes de una versión
govulncheck ./...          # en CI programada y antes de desplegar
```

- El código no formateado no se revisa manualmente.
- Una advertencia no se silencia sin motivo escrito y alcance mínimo.
- `//nolint` solo puede acompañar la regla exacta y una explicación.
- Dependencias y herramientas quedan fijadas por `go.mod` y por la configuración reproducible de CI.

## 8. Lista de revisión del backend

- [ ] El nombre y el paquete corresponden a una capacidad del dominio.
- [ ] La dirección de dependencias conserva el núcleo independiente de HTTP y PostgreSQL.
- [ ] El caso de uso aplica tenant, permisos y transacción en un único lugar claro.
- [ ] No hay SQL ni regla de negocio en el handler.
- [ ] Errores, cancelación e idempotencia tienen comportamiento explícito.
- [ ] No se registran datos personales, tokens o cuerpos sensibles.
- [ ] El contrato OpenAPI, sus ejemplos y las pruebas del handler coinciden.
- [ ] Se agregaron las pruebas exigidas por [estrategia-pruebas.md](estrategia-pruebas.md).
- [ ] Los comentarios explican decisiones no evidentes y no repiten el código.
- [ ] Una dependencia o abstracción nueva resuelve un problema actual demostrable.
- [ ] El cambio preserva las reglas `RN-*`, las decisiones `DEC-*` y el aislamiento entre barberías.

## 9. Referencias oficiales

- [Organización de un módulo Go](https://go.dev/doc/modules/layout)
- [Comentarios de documentación en Go](https://go.dev/doc/comment)
- [Effective Go](https://go.dev/doc/effective_go)
- [Detector de carreras de Go](https://go.dev/doc/articles/race_detector)
