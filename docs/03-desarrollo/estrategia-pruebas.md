---
titulo: "Estrategia de pruebas y controles de calidad"
version: "1.3"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/matriz-trazabilidad.md"
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "estandar-backend-go.md"
  - "estandar-frontend-vue.md"
  - "estandar-diseno-visual.md"
  - "../05-backend/estandar-base-datos.md"
  - "../05-backend/migraciones-atlas.md"
  - "../06-api/estandar-openapi.md"
---

# Estrategia de pruebas y controles de calidad

## 1. Objetivo

Las pruebas deben demostrar que el MVP permite operar la agenda sin cruces, fugas entre barberías, duplicados ni pérdida silenciosa del historial. Se prioriza una pirámide práctica: muchas pruebas unitarias y de componente, integración suficiente con PostgreSQL real y un conjunto pequeño de E2E para recorridos completos.

Toda corrección de defecto comienza con una prueba que falle por el defecto y termina conservando esa prueba como regresión.

Fuentes normativas: `DEC-035` y `DEC-039`.

## 2. Herramientas confirmadas

| Área | Herramienta | Uso |
| --- | --- | --- |
| Backend unitario | `testing` de Go | dominio, servicios y utilidades puras |
| Backend HTTP | `net/http/httptest` | handlers, middleware y contrato HTTP |
| Backend concurrencia | `go test -race` | accesos concurrentes y estado compartido |
| Backend entradas hostiles | fuzzing nativo de Go | parsers, validadores y fronteras expuestas |
| Frontend unitario | Vitest | modelos, validación, composables y cliente API |
| Componentes Vue | Vitest + Vue Test Utils | comportamiento visible de componentes |
| E2E | Playwright | recorridos reales en navegador |
| Persistencia | PostgreSQL real | repositorios, RLS, restricciones y migraciones |
| Contrato HTTP | Redocly CLI + bundle OpenAPI | lint, referencias, schemas y documentación |

No se usa SQLite como sustituto de PostgreSQL: no reproduce RLS, rangos, exclusiones ni los mismos bloqueos.

## 3. Regla para elegir el tipo de prueba

| Cambio | Prueba mínima requerida |
| --- | --- |
| Regla de dominio o cálculo | unitaria con casos de frontera |
| Servicio de aplicación | unitaria con dobles solo en puertos externos |
| Handler o middleware | prueba HTTP con `httptest` |
| Consulta o repositorio | integración contra PostgreSQL real |
| Migración, restricción o RLS | prueba de base de datos y dos tenants |
| Componente Vue con estados o interacción | prueba de componente |
| Composable, validación o modelo de vista | unitaria con Vitest |
| Cambio de contrato API | pruebas del handler, cliente frontend y contrato cuando exista |
| Recorrido P0 que cruza frontend, API y datos | E2E o integración de sistema |
| Carrera de agenda o trabajador | integración concurrente repetible |
| Defecto corregido | prueba de regresión en la capa más baja que lo reproduzca fielmente |

Una prueba E2E no reemplaza una prueba unitaria que localiza una regla, y un mock de repositorio no demuestra que una consulta o política RLS sea correcta.

## 4. Pruebas del backend

### 4.1 Unitarias de dominio y aplicación

Se escriben junto al paquete como `*_test.go`. Se prefieren tablas de casos cuando varias entradas verifican la misma regla.

Cobertura obligatoria por comportamiento:

- intervalos `[inicio, fin)`, contigüidad y cruces de medianoche;
- duración, paso, anticipación y ventana de reserva;
- horario recurrente, excepciones, festivos y bloqueos;
- transiciones de cita, `no_show`, corrección y cierre;
- política de cancelación por actor y plazo;
- permisos y pertenencia a barbería;
- programación, invalidación y reintentos de recordatorios;
- anonimización y revocación de tokens;
- idempotencia de creación y cambios críticos.

Reglas:

1. Congelar el reloj mediante un puerto `Clock`; no depender de la hora del equipo.
2. Usar identificadores y zonas horarias explícitas en cada caso.
3. Probar resultado y efecto observable, no cantidad de llamadas internas salvo que la llamada sea el contrato.
4. Preferir fakes pequeños a mocks profundos. Solo simular fronteras como repositorio, reloj o proveedor externo.
5. Un caso de uso con transacción prueba éxito, rechazo de negocio, error de persistencia y cancelación de contexto.

### 4.2 HTTP y middleware

Cada endpoint prueba como mínimo:

- solicitud válida y forma exacta de la respuesta;
- JSON inválido, campo desconocido, cuerpo excesivo y dato faltante;
- autenticación ausente, inválida y válida cuando corresponda;
- recurso de otra barbería tratado como no encontrado;
- traducción estable de conflicto, validación y error interno;
- propagación de `request_id`, timeout y cancelación;
- idempotencia repetida con igual y diferente contenido;
- ausencia de detalles sensibles en la respuesta.

Las pruebas invocan un `http.Handler` real. No se llama directamente a métodos privados del handler para aparentar cobertura.

### 4.3 Contrato OpenAPI

- Redocly valida y bundlea la fuente multiarchivo antes de ejecutar codegen o pruebas.
- Cada operación implementada coincide en método, path, seguridad, status, headers, media type y schema.
- Requests y responses de las pruebas HTTP se validan contra el bundle OpenAPI.
- Ejemplos del contrato validan contra sus schemas y nunca contienen datos reales.
- CI detecta operaciones documentadas sin implementación y rutas implementadas sin contrato cuando exista el verificador del servidor.
- Un cambio incompatible requiere la transición definida en [estandar-openapi.md](../06-api/estandar-openapi.md).

### 4.4 Integración con PostgreSQL

Los repositorios se prueban con el esquema migrado desde cero en una base aislada. Cada grupo crea sus propios datos o usa transacciones que pueda revertir sin interferir con otros grupos.

Casos obligatorios:

- crear, consultar y modificar dentro del tenant autorizado;
- usar el mismo identificador conocido desde otro tenant y obtener cero filas;
- intentar insertar una relación cruzada y recibir rechazo de integridad;
- validar restricciones `NOT NULL`, `CHECK`, `UNIQUE`, FK y exclusión temporal;
- crear citas contiguas y rechazar cruces reales;
- ejecutar carreras reserva/reserva y reserva/bloqueo;
- reclamar trabajos concurrentes sin doble procesamiento;
- cancelar contexto y comprobar que la consulta termina;
- verificar planes de consultas críticas con un volumen representativo antes del piloto.

RLS se prueba con el rol normal de la aplicación, nunca solo con el propietario de las tablas.

### 4.5 Worker y proveedores

- El worker prueba reclamación, éxito, fallo temporal, fallo permanente, backoff e idempotencia.
- La programación y el cambio de negocio se verifican en la misma transacción.
- Los adaptadores de correo y WhatsApp usan servidores falsos o contratos controlados; las pruebas automáticas no envían mensajes reales.
- El contrato del proveedor se prueba por separado sin incluir credenciales en fixtures o resultados.

### 4.6 Carrera, fuzzing y seguridad

- `go test -race ./...` corre de forma programada y antes de una versión.
- Los validadores de tokens, parámetros, JSON o intervalos que acepten muchas combinaciones son candidatos a fuzzing.
- Todo caso encontrado por fuzzing se conserva en el corpus.
- `govulncheck ./...` corre de forma programada y antes de desplegar.
- Autenticación, rate limit y recuperación incluyen pruebas de abuso, expiración y no enumeración.

## 5. Pruebas del frontend

### 5.1 Unitarias

Vitest cubre:

- validadores y transformación de formularios;
- formato de fecha en la zona de la barbería;
- cálculo exclusivamente visual, nunca la autoridad de disponibilidad;
- composables sin y con ciclo de vida;
- estados discriminados de carga, vacío, éxito y error;
- cliente API: headers, `request_id`, idempotencia, cancelación y mapeo de errores.

No se prueba una asignación trivial. Se prueban decisiones, ramas, casos límite y contratos.

### 5.2 Componentes

Todo componente con interacción, condición o estado asincrónico necesita prueba. Se monta con Vue Test Utils y se observa desde la perspectiva del usuario.

Casos habituales:

- contenido y controles accesibles por rol o etiqueta;
- props de entrada y eventos emitidos;
- escritura, selección, envío y prevención de doble toque;
- carga, vacío, error por campo, error global y reintento;
- permisos que ocultan o deshabilitan una acción;
- foco y teclado en diálogos o mensajes críticos;
- zona horaria y texto de “turno” correctos.
- variantes visuales con nombre accesible, color acompañado por otra señal y objetivo táctil conforme al estándar.

No se afirman métodos privados, estructura interna de `ref` ni clases CSS salvo que la clase sea el contrato visual. Los snapshots no se usan como única evidencia.

### 5.3 E2E con Playwright

Los E2E prueban comportamiento visible, usan locators por rol, etiqueta o texto y mantienen datos aislados. No dependen de orden, cuenta compartida ni espera fija.

Recorridos P0 mínimos antes del piloto:

1. cliente consulta disponibilidad y reserva un turno;
2. doble envío o reintento conserva una sola cita;
3. cliente abre su enlace y cancela según la política;
4. barbero inicia sesión y recupera acceso;
5. barbero crea una cita manual fuera de la rejilla pública válida;
6. barbero reprograma, cancela y cambia estado con historial;
7. bloqueo urgente muestra citas afectadas sin cancelarlas automáticamente;
8. una barbería no ve ni modifica datos de otra;
9. un cambio de cita reprograma recordatorios sin duplicarlos;
10. conexión o respuesta fallida muestra reintento seguro y conserva datos no sensibles.

En cada cambio se ejecutan los recorridos afectados. En cada PR se mantiene un smoke crítico en Chromium para escritorio y viewport móvil. Antes de una versión se ejecuta la suite P0 en Chromium, Firefox y WebKit; el teléfono real del piloto complementa, pero no reemplaza, la automatización.

Terceros como WhatsApp o correo se interceptan. El E2E comprueba que el sistema programó el efecto correcto, no la disponibilidad del proveedor externo.

### 5.4 Verificación visual y accesible

Un cambio visible conserva evidencia en 360 y 1280 px; se agrega 320 px para reflow y 768 px cuando cambia la composición. Se revisan los estados afectados —foco, carga, vacío, error, conflicto, éxito e inactivo—, zoom de texto al 200 %, teclado y contraste.

Las capturas verifican composición, no sustituyen las aserciones de comportamiento. Los snapshots visuales, si se incorporan después, se actualizan solo tras revisar la diferencia.

Herramienta de automatización accesible, elegida en la auditoría HU-009: **axe-core**, vía el paquete `vitest-axe` (envoltorio del matcher `toHaveNoViolations` para Vitest). Necesidad: motor de reglas WCAG 2.1/2.2 de referencia de la industria (Deque), integrable en la prueba de componente sin navegador real. Mantenimiento: `axe-core` tiene desarrollo activo y adopción amplia; `vitest-axe` es un envoltorio delgado (el registro automático de su versión publicada resultó incompleto — `dist/extend-expect.js` vacío en 0.1.0 — así que el proyecto registra el matcher a mano en `apps/web/vitest.setup.ts` contra el `matchers.js` de la misma librería, sin parchear su código). Licencia: MIT. Superficie transitiva: pequeña (`axe-core`, `aria-query`, utilidades de formato). La regla `color-contrast` se desactiva en las pruebas de componente porque jsdom no implementa Canvas2D y ese chequeo no puede medir contraste real ahí; el contraste sigue verificado por la tabla aprobada de [estandar-diseno-visual.md](estandar-diseno-visual.md) sección 4.3. No sustituye la revisión manual por teclado que cada prueba de componente ya cubre, ni la verificación en navegador real antes del piloto.

## 6. Pruebas de migraciones y base de datos

En cada PR que toque SQL deben pasar:

1. validar `atlas.sum` y la semántica SQL con `atlas migrate validate` sobre PostgreSQL efímero;
2. aplicar todas las migraciones con Atlas sobre una base vacía;
3. actualizar una base en la versión anterior con datos representativos;
4. verificar datos, restricciones, índices, funciones, roles y políticas RLS esperados;
5. ejecutar la suite de integración con el rol de aplicación;
6. comprobar el plan de avance correctivo o restauración descrito por la migración;
7. medir bloqueos y duración cuando modifique una tabla con datos;
8. comprobar que `atlas migrate status` no deja versiones pendientes al terminar.

Una migración no se considera probada solo porque ejecuta sin error sobre una base vacía.

El flujo de promoción, trazabilidad y `dry-run` se define en [migraciones-atlas.md](../05-backend/migraciones-atlas.md).

## 7. Datos de prueba y aislamiento

- Solo datos ficticios; nunca copiar información del piloto o producción.
- Builders declaran los campos relevantes y proporcionan valores válidos para el resto.
- Cada prueba controla su barbería, usuario, reloj y zona horaria.
- Las pruebas pueden ejecutarse solas, en otro orden y en paralelo cuando estén marcadas como paralelas.
- No usar `sleep` para coordinar concurrencia; usar barreras, canales o condiciones observables.
- Semillas del entorno local son repetibles y no forman parte de migraciones de producción salvo datos de referencia.
- Un fallo conserva logs técnicos, traza o captura sin tokens ni datos personales.

## 8. Cobertura y calidad mínima

La cobertura es una alarma, no el objetivo principal.

- Toda regla o criterio modificado tiene al menos una prueba de éxito y sus fronteras relevantes.
- Los paquetes críticos de dominio/aplicación (`booking`, `schedule`, `auth`, programación de notificaciones y aislamiento tenant) mantienen al menos 85 % de cobertura de sentencias.
- Los demás paquetes Go no generados con lógica mantienen al menos 75 %.
- Modelos, validación, composables y cliente API del frontend mantienen al menos 80 % de sentencias y 75 % de ramas.
- Componentes se evalúan por comportamiento; no se agrega una prueba vacía para elevar el porcentaje.
- Bootstrap, código generado y adaptadores declarativos pueden excluirse con justificación explícita.
- El reporte incluye archivos no importados para que un archivo sin pruebas no desaparezca de la medición.
- El umbral nunca baja en un PR para hacerlo pasar.

Un 100 % de cobertura que no prueba concurrencia, RLS o efectos visibles no satisface este estándar.

## 9. Puertas de calidad

### Cada cambio local

- formato, lint y tipos;
- pruebas unitarias afectadas;
- prueba de regresión si corrige un defecto.

### Cada pull request

- build de backend y frontend;
- todas las unitarias y de componente;
- integración PostgreSQL, migraciones y RLS;
- lint, bundle y pruebas del contrato OpenAPI;
- E2E smoke de recorridos críticos afectados;
- evidencia visual y revisión accesible para cambios de interfaz;
- cobertura sin reducción;
- análisis estático sin errores nuevos.

### Programado y antes de versión

- detector de carreras;
- fuzzing con tiempo acotado sobre objetivos elegidos;
- revisión de vulnerabilidades;
- E2E P0 en los tres motores;
- restauración de respaldo y pruebas de migración sobre copia anonimizada de volumen representativo;
- revisión de consultas críticas y bundle frontend.

No se permiten pruebas inestables ignoradas. Una prueba flaky se corrige o se aísla temporalmente con responsable, referencia y fecha límite; no se reintenta indefinidamente para ocultarla.

## 10. Convenciones y trazabilidad

- Go: `TestUnidad_Condicion_Resultado` o nombres equivalentes claros; subpruebas describen el escenario.
- Frontend: `describe` nombra unidad y `it` expresa comportamiento visible.
- E2E: título con actor, acción y resultado.
- Los archivos viven junto al código salvo E2E y pruebas globales de base de datos.
- Al implementarse un caso formal se asigna un identificador estable, por ejemplo `UT-BE-*`, `IT-BD-*`, `CT-FE-*` o `E2E-*`, y se enlaza a la matriz. No se reservan identificadores para pruebas inexistentes.
- Una prueba relacionada con una regla compleja cita su `RN-*` en el nombre o comentario cuando aporte trazabilidad.

## 11. Lista de revisión de pruebas

- [ ] La capa elegida reproduce el riesgo sin depender innecesariamente de todo el sistema.
- [ ] Existen casos de éxito, frontera, error y autorización relevantes.
- [ ] Tenant, tiempo e identificadores son explícitos.
- [ ] La prueba falla si se elimina la conducta que pretende proteger.
- [ ] No depende de orden, espera fija, red externa o datos compartidos.
- [ ] No contiene datos personales ni secretos.
- [ ] Las integraciones usan PostgreSQL y roles reales.
- [ ] El recorrido P0 afectado conserva prueba E2E o de sistema.
- [ ] La evidencia se enlaza cuando exista un criterio codificado.

## 12. Referencias oficiales

- [Pruebas con tablas en Go](https://go.dev/wiki/TableDrivenTests)
- [Fuzzing en Go](https://go.dev/doc/security/fuzz/)
- [Detector de carreras de Go](https://go.dev/doc/articles/race_detector)
- [Pruebas en Vue](https://vuejs.org/guide/scaling-up/testing.html)
- [Vue Test Utils](https://test-utils.vuejs.org/guide/)
- [Cobertura en Vitest](https://vitest.dev/guide/coverage.html)
- [Buenas prácticas de Playwright](https://playwright.dev/docs/best-practices)
