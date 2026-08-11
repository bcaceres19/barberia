# Cómo contribuir

Este proyecto usa GitHub Flow: `main` es la única rama permanente y todo cambio entra por pull request con squash. La norma completa está en [`docs/03-desarrollo/flujo-git-github.md`](docs/03-desarrollo/flujo-git-github.md).

## Antes de cambiar código

1. Lea `AGENTS.md`, el registro de decisiones y los estándares del área afectada.
2. Cree o seleccione un issue con criterios de aceptación para cualquier cambio no trivial.
3. Actualice `main` y cree una rama `<tipo>/<issue>-<descripcion>`.

```bash
git switch main
git pull --ff-only
git switch -c feat/123-reserva-publica
```

Prefijos permitidos: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `ci` y `hotfix`.

## Commits

Use Conventional Commits y una descripción breve en imperativo:

```text
feat(booking): crea una reserva pública idempotente
fix(db): impide cruces de citas del mismo barbero
docs(api): documenta errores de disponibilidad
```

Cada commit expresa una intención coherente, incluye sus pruebas o documentación necesarias y no mezcla cambios ajenos. Revise siempre el staged diff antes de confirmar:

```bash
git status --short
git diff --cached
```

Nunca incluya secretos, archivos `.env`, datos personales reales, dumps, reportes locales ni salidas de build.

## Pull request

Abra un draft si necesita retroalimentación temprana. Antes de marcarlo listo:

- complete la plantilla y enlace el issue con `Closes #123` o `Refs #123`;
- use un título Conventional Commits, porque será el commit final del squash;
- ejecute formato, lint, tipos, pruebas y build aplicables;
- actualice OpenAPI, Atlas, pruebas y documentación en el mismo cambio coherente;
- revise el diff completo y retire `WIP`, `fixup!` y cambios no relacionados;
- documente riesgo, despliegue y recuperación cuando afecte datos u operación.

No se integra con checks fallidos, conversaciones abiertas o contradicción entre código, contrato, migraciones y documentación. Al terminar, se elimina la rama.

## Normas por área

- Go: [`docs/03-desarrollo/estandar-backend-go.md`](docs/03-desarrollo/estandar-backend-go.md)
- Vue: [`docs/03-desarrollo/estandar-frontend-vue.md`](docs/03-desarrollo/estandar-frontend-vue.md)
- Diseño visual: [`docs/03-desarrollo/estandar-diseno-visual.md`](docs/03-desarrollo/estandar-diseno-visual.md)
- Pruebas: [`docs/03-desarrollo/estrategia-pruebas.md`](docs/03-desarrollo/estrategia-pruebas.md)
- PostgreSQL: [`docs/05-backend/estandar-base-datos.md`](docs/05-backend/estandar-base-datos.md)
- Atlas: [`docs/05-backend/migraciones-atlas.md`](docs/05-backend/migraciones-atlas.md)
- OpenAPI: [`docs/06-api/estandar-openapi.md`](docs/06-api/estandar-openapi.md)
