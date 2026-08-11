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

## Estructura obligatoria

- Backend Go: `apps/api`, un módulo Go, comandos `api` y `worker`, paquetes internos por capacidad.
- Frontend Vue: `apps/web`, módulos por funcionalidad y dependencias `app → modules → shared`.
- Base de datos: `database/migrations`, `database/seeds`, `database/testdata` y `database/tests`.
- No crear carpetas globales `utils`, `helpers`, `controllers`, `models` o `services` sin un dueño y una responsabilidad concreta.

## Calidad

- Mantener dominio y servicios Go independientes de Chi y PostgreSQL.
- Mantener componentes Vue independientes de la forma interna del API mediante un cliente tipado.
- Aplicar `docs/03-desarrollo/estandar-diseno-visual.md`: los módulos no inventan paletas, tamaños, tipografías ni variantes fuera de los tokens y patrones aprobados.
- No usar `any`, estado global, ORM, borrado en cascada, `jsonb` o una dependencia nueva sin justificación conforme al estándar correspondiente.
- Comentarios explican decisiones e invariantes; no repiten el código. Todo `TODO` tiene referencia rastreable.
- Ningún log, fixture, comentario o prueba contiene datos personales reales, secretos o tokens.

## Pruebas obligatorias

- Cada regla o defecto modificado conserva una prueba en la capa más baja adecuada.
- SQL, RLS, migraciones y concurrencia se prueban con PostgreSQL real y al menos dos tenants.
- Componentes interactivos tienen prueba de componente; recorridos P0 afectados tienen E2E.
- Todo cambio visible incluye evidencia responsive y verificación accesible en los anchos definidos por el estándar visual.
- Antes de integrar deben pasar los controles definidos en `docs/03-desarrollo/estrategia-pruebas.md`.
- No reducir umbrales, ignorar pruebas inestables ni actualizar snapshots sin revisar el comportamiento.

## Base de datos y documentación

- Diseñar en tercera forma normal por defecto; toda desnormalización necesita motivo, consistencia y prueba.
- Una migración aplicada es inmutable. Evaluar bloqueo, datos existentes y recuperación antes de cambios destructivos.
- Crear, validar y aplicar migraciones con Atlas según `docs/05-backend/migraciones-atlas.md`; versionar `atlas.sum` y no ejecutar DDL al arrancar la aplicación.
- Actualizar contrato API, diagrama/diccionario de datos, matriz y decisiones cuando el cambio los afecte.
- Toda operación HTTP debe seguir `docs/06-api/estandar-openapi.md`; OpenAPI es contract-first y se actualiza junto con handler, cliente y pruebas.
- Un cambio no está terminado si código, pruebas y documentación normativa se contradicen.
