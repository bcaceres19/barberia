# api/openapi

Contrato HTTP contract-first entre `apps/web` y `apps/api`. Ver
[`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
antes de agregar una operación.

## Estado

Solo existe la base: `openapi.yaml` sin operaciones, las carpetas `paths/` y
`components/` con la agrupación documentada y un `README.md` explicando qué
irá en cada una. No hay endpoints ni esquemas todavía.

## Comandos

Desde la raíz del repositorio, con las dependencias de `package.json`
instaladas (`pnpm install`):

```bash
pnpm run openapi:check-config
pnpm run openapi:lint
pnpm run openapi:bundle
pnpm run openapi:docs
```

`api/openapi/dist/` es un artefacto generado; no se edita ni se versiona.
