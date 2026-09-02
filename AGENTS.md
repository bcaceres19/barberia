# Reglas de desarrollo del proyecto

Estas reglas aplican a personas y agentes que modifiquen el repositorio.

## Antes de implementar

1. Leer `docs/00-control/registro-decisiones.md`, el alcance P0 y las reglas de negocio afectadas.
2. Aplicar los estándares de `docs/03-desarrollo/` y `docs/05-backend/estandar-base-datos.md`.
3. No inventar una respuesta a una duda o contradicción: registrarla antes de codificar.

## Git y entrega

- Aplicar `docs/03-desarrollo/flujo-git-github.md` y `CONTRIBUTING.md`.
- Todo cambio no trivial parte de un issue y una rama corta `<tipo>/<issue>-<descripcion>` creada desde `main` actualizada.
- Los commits y el título del PR siguen Conventional Commits; no mezclar preocupaciones ni incluir secretos o datos personales.
- No hacer push directo, force push ni reescribir `main`. Integrar únicamente por pull request y squash.
- Antes del merge deben pasar los checks aplicables y resolverse todas las conversaciones. Con colaboradores se exige al menos una aprobación ajena al último cambio.
- Eliminar la rama después del merge y desplegar el SHA o tag aprobado, nunca una rama de ambiente divergente.

## Prompts persistentes y relevo entre agentes

- Todo prompt redactado para ejecución futura —implementación de una `HU-*`, corrección, auditoría, revisión, pruebas, documentación, CI, operación u orquestación— se guarda antes de entregarlo en `docs/10-backlog/prompts/`, siguiendo su [catálogo y plantilla](docs/10-backlog/prompts/README.md). No se considera entregado si existe únicamente en un chat.
- Cada archivo es autocontenido y registra como mínimo: identificador y versión del prompt, tipo de trabajo, estado, agente objetivo, `HU-*` cuando aplique, issue real o `pending`, dependencias, reglas `RN-*`, decisiones `DEC-*`, criterios `CA-*`, alcance excluido, pruebas, rama y PR cuando existan.
- Un prompt que autoriza cambios en el repositorio solo puede pasar a `ready` y ejecutarse cuando tiene un issue real con criterios verificables. `issue: pending` obliga a conservarlo como `draft`; no se inventan números ni se omite la trazabilidad.
- Un prompt atiende una preocupación primaria y un issue. Para varias entregas se crea un prompt de orquestación que enlaza prompts independientes, sin combinar sus ramas o PR.
- Los prompts son herramientas derivadas, no fuentes normativas. Si contradicen `AGENTS.md`, una `RN-*`, una `DEC-*`, una `HU-*` o un estándar, prevalece la fuente normativa y el prompt se corrige o se sustituye antes de ejecutarlo.
- Al crear, cambiar de estado, ejecutar, bloquear o sustituir un prompt, se actualizan su metadato y el índice de `docs/10-backlog/prompts/`. Un cambio material después de iniciar la ejecución crea una nueva versión que enlaza `supersedes`; no se reescribe silenciosamente el cuerpo usado por otro agente.
- La respuesta final que entregue un prompt incluye un enlace al archivo guardado. Si el agente no puede escribir el repositorio, debe decir explícitamente que no quedó persistido y entregar el contenido como borrador pendiente de guardar.

## Estructura obligatoria

- Backend Go: `apps/api`, un módulo Go, comandos `api` y `worker`, paquetes internos por capacidad.
- Frontend Vue: `apps/web`, módulos por funcionalidad y dependencias `app → modules → shared`.
- Base de datos: `database/migrations`, `database/seeds`, `database/testdata` y `database/tests`.
- No crear carpetas globales `utils`, `helpers`, `controllers`, `models` o `services` sin un dueño y una responsabilidad concreta.

## Calidad

- Mantener dominio y servicios Go independientes de Chi y PostgreSQL.
- Mantener componentes Vue independientes de la forma interna del API mediante un cliente tipado.
- Aplicar `docs/03-desarrollo/estandar-diseno-visual.md`: todo rediseño o pantalla nueva respeta los mockups aprobados, la firma cromática y el lenguaje NAVA / Tailored Grid; composición, tamaños, grid, componentes y tecnología de estilos permanecen libres, con accesibilidad, responsive, rendimiento y pruebas obligatorios.
- No usar `any`, estado global, ORM, borrado en cascada, `jsonb` o una dependencia nueva sin justificación conforme al estándar correspondiente.
- Comentarios explican decisiones e invariantes; no repiten el código. Todo `TODO` tiene referencia rastreable.
- Ningún log, fixture, comentario o prueba contiene datos personales reales, secretos o tokens.

## Pruebas obligatorias

- Cada regla o defecto modificado conserva una prueba en la capa más baja adecuada.
- SQL, RLS, migraciones y concurrencia se prueban con PostgreSQL real y al menos dos tenants.
- Componentes interactivos tienen prueba de componente; recorridos P0 afectados tienen E2E.
- Todo cambio visible incluye evidencia responsive y verificación accesible en los anchos definidos por la guía visual y la estrategia de pruebas.
- Antes de integrar deben pasar los controles definidos en `docs/03-desarrollo/estrategia-pruebas.md`.
- No reducir umbrales, ignorar pruebas inestables ni actualizar snapshots sin revisar el comportamiento.

## Base de datos y documentación

- Diseñar en tercera forma normal por defecto; toda desnormalización necesita motivo, consistencia y prueba.
- Una migración aplicada es inmutable. Evaluar bloqueo, datos existentes y recuperación antes de cambios destructivos.
- Crear, validar y aplicar migraciones con Atlas según `docs/05-backend/migraciones-atlas.md`; versionar `atlas.sum` y no ejecutar DDL al arrancar la aplicación.
- Actualizar contrato API, diagrama/diccionario de datos, matriz y decisiones cuando el cambio los afecte.
- Toda operación HTTP debe seguir `docs/06-api/estandar-openapi.md`; OpenAPI es contract-first y se actualiza junto con handler, cliente y pruebas.
- Un cambio no está terminado si código, pruebas y documentación normativa se contradicen.
