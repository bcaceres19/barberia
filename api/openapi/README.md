# api/openapi

Contrato HTTP contract-first entre `apps/web` y `apps/api`. Ver
[`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
antes de agregar una operación.

## Estado

`openapi.yaml` declara una operación: `POST /public/auth/login` (HU-005,
`paths/public-auth.yaml`). El resto de capacidades (`paths/public-booking.yaml`,
`customer-appointments.yaml`, `private-appointments.yaml`, `catalog.yaml`,
`schedules.yaml`, `settings.yaml`) sigue sin contenido.

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
