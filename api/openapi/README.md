# api/openapi

Contrato HTTP contract-first entre `apps/web` y `apps/api`. Ver
[`docs/06-api/estandar-openapi.md`](../../docs/06-api/estandar-openapi.md)
antes de agregar una operación.

## Estado

`openapi.yaml` declara: autenticación pública y privada (`paths/public-auth.yaml`,
`paths/private-auth.yaml`, HU-005/006/007/008/012), configuración de la
barbería (`paths/settings.yaml`, HU-020) y registro/listado de barberos
(`paths/staff.yaml`, HU-021: `GET`/`POST /private/barbers`,
`GET`/`PATCH /private/barbers/{barberId}`). El resto de capacidades
(`paths/public-booking.yaml`, `customer-appointments.yaml`,
`private-appointments.yaml`, `catalog.yaml`, `schedules.yaml`) sigue sin
contenido.

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
