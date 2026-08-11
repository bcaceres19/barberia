---
titulo: "Stack, despliegue y operación"
version: "1.6"
estado: "Decisión confirmada"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/alcance-mvp.md"
  - "../05-backend/base-datos.md"
  - "frontend.md"
  - "backend-go.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../05-backend/estandar-base-datos.md"
  - "../05-backend/migraciones-atlas.md"
  - "../06-api/estandar-openapi.md"
  - "../03-desarrollo/flujo-git-github.md"
---

# Stack, despliegue y operación

## 1. Decisión

La arquitectura inicial es un **monolito modular**:

| Capa | Elección | Razón |
| --- | --- | --- |
| Backend | Go + Chi v5 sobre `net/http` | Routing y middleware componibles sin framework pesado ni tipos HTTP propietarios |
| Frontend | Vue 3 + TypeScript + Vite | Desarrollo rápido, runtime contenido y aplicación responsive sin instalación |
| API | HTTP/JSON con OpenAPI 3.1.2 | Contrato verificable y generable entre frontend y backend |
| Base de datos | PostgreSQL | Transacciones, rangos temporales, restricciones de exclusión y RLS |
| Trabajos programados | Proceso trabajador del mismo código Go | Ejecuta recordatorios ya persistidos sin introducir microservicios |
| Desarrollo y piloto | Servicios gestionados gratuitos, mientras cumplan respaldos y disponibilidad mínima | Minimiza costo antes de validar el producto |
| Producción comercial | VPS dedicado económico, con PostgreSQL gestionado o aislado y copias externas | Control de costo sin compartir el proceso de aplicación con proyectos ajenos |

Fuente: `DEC-023`, `DEC-024`, `DEC-031`, `DEC-033`, `DEC-034`, `DEC-036`, `DEC-037` y `DEC-038`.

### Perfil mínimo del frontend

- Vue 3 con Composition API y `<script setup>`;
- Vite como servidor de desarrollo y herramienta de compilación;
- Vue Router con carga diferida por ruta;
- estado local y composables por defecto; Pinia solo cuando exista estado compartido real entre rutas;
- cliente HTTP delgado sobre `fetch`, con contratos TypeScript derivados de la API;
- sin SSR, framework full-stack, biblioteca visual pesada ni servicio offline en el MVP;
- páginas comerciales estáticas separadas de la aplicación operativa.

El detalle y las reglas de dependencias están en [frontend.md](frontend.md).

### Perfil mínimo del backend

- Chi v5 limitado a routing, subrouters y composición de middleware;
- handlers estándar `net/http`;
- codificación JSON con herramientas compatibles con la biblioteca estándar;
- servicios de aplicación independientes del transporte HTTP;
- repositorios explícitos y transacciones controladas por caso de uso;
- sin ORM, contenedor de inyección, framework MVC ni abstracciones globales en el MVP.

La estructura, el orden de middleware y los límites están en [backend-go.md](backend-go.md).

### Estructura del repositorio

```text
apps/
  api/        # módulo Go, API y worker
  web/        # Vue, pruebas de componente y E2E
database/     # migraciones, seeds ficticios y pruebas SQL
docs/         # producto, arquitectura y estándares
```

Los estándares detallados están en `docs/03-desarrollo/`; la base de datos se rige además por [estandar-base-datos.md](../05-backend/estandar-base-datos.md).

### Herramientas de calidad

| Área | Controles mínimos |
| --- | --- |
| Go | `gofmt`, `go vet`, `go test`, detector de carreras y `govulncheck` |
| Vue/TypeScript | Prettier, ESLint, `vue-tsc`, Vitest y build de producción |
| Componentes | Vue Test Utils sobre comportamiento visible |
| Sistema | Playwright para recorridos P0 |
| Datos | Atlas CLI, migraciones versionadas y pruebas contra PostgreSQL real |
| Contrato HTTP | Redocly CLI para lint, bundle y documentación OpenAPI |
| Control de cambios | GitHub Flow, Conventional Commits, PR y squash sobre `main` protegida |

Los momentos de ejecución, umbrales y escenarios obligatorios están en [estrategia-pruebas.md](../03-desarrollo/estrategia-pruebas.md).

Las migraciones se ejecutan en una etapa de despliegue separada usando [Atlas](../05-backend/migraciones-atlas.md), nunca durante el arranque del API o del worker.

OpenAPI es contract-first y se administra según [estandar-openapi.md](../06-api/estandar-openapi.md). El bundle validado alimenta documentación, tipos del frontend y pruebas de contrato.

Los cambios se entregan mediante GitHub Flow según [flujo-git-github.md](../03-desarrollo/flujo-git-github.md). No existen ramas por ambiente: cada despliegue promueve el mismo commit o tag que superó los checks.

## 2. Módulos del backend

El monolito mantiene límites internos explícitos:

- identidad y sesiones;
- barberías, usuarios y barberos;
- servicios y horarios;
- disponibilidad;
- turnos/citas;
- bloqueos y festivos;
- notificaciones y recordatorios;
- auditoría y operación.

Un módulo no lee tablas de otro de forma improvisada: usa una función o servicio interno con contexto de barbería. La separación prepara el código para crecer sin asumir el costo operativo de varios servicios desplegados.

## 3. Recordatorios y notificaciones

Al confirmar o modificar una cita, la misma transacción:

1. persiste el cambio;
2. invalida programaciones anteriores cuando corresponda;
3. crea las nuevas filas de recordatorio;
4. registra el evento que producirá avisos inmediatos.

El trabajador consulta filas vencidas, bloquea un lote pequeño con semántica equivalente a `FOR UPDATE SKIP LOCKED`, vuelve a leer el estado vigente y envía por los canales configurados. Los reintentos nunca crean un segundo recordatorio lógico.

Canales del MVP:

- correo electrónico transaccional;
- interfaz oficial de WhatsApp;
- selección por barbería y por tipo de evento.

No se usan automatizaciones de WhatsApp no oficiales.

## 4. Entornos

| Entorno | Datos | Propósito |
| --- | --- | --- |
| Local | Ficticios | Desarrollo y pruebas |
| Piloto | Reales, minimizados | Validación con 2 o 3 participantes |
| Producción | Reales | Operación comercial posterior |

No se copian datos personales de piloto o producción al entorno local. Secretos, tokens y contraseñas se inyectan por configuración del entorno y nunca se almacenan en el repositorio.

## 5. Copias y recuperación

- copia automática diaria;
- retención de 30 días;
- almacenamiento separado del proceso de aplicación;
- restauración completa probada antes del piloto y después de todo cambio material en el mecanismo;
- registro de fecha, duración, resultado y responsable de cada prueba de recuperación.

Una copia cuya restauración no se haya probado no satisface el requisito operativo.

## 6. Paso del piloto a producción

Antes de migrar a un VPS:

- medir memoria, CPU, almacenamiento y volumen de envíos reales;
- verificar que el proveedor soporta TLS, copias externas y monitoreo;
- ejecutar una restauración en un entorno limpio;
- documentar el procedimiento de despliegue y reversión;
- revisar la política de datos y los acuerdos con criterio jurídico;
- confirmar costos vigentes de correo y WhatsApp oficial.

“VPS dedicado” describe el servidor de aplicación. No significa que la base de datos deba vivir sin respaldo en el mismo disco.
