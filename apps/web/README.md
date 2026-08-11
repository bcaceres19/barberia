# apps/web

Aplicación Vue 3 + TypeScript + Vite del flujo público y del panel del
barbero. Ver
[`docs/04-arquitectura/frontend.md`](../../docs/04-arquitectura/frontend.md)
y [`docs/03-desarrollo/estandar-frontend-vue.md`](../../docs/03-desarrollo/estandar-frontend-vue.md)
antes de agregar código.

## Estado

Solo existe la base: arranque de la aplicación, router con una pantalla
provisional, tokens del sistema visual (`src/styles/tokens.css`) y las
carpetas de `modules/` y `shared/` con un `index.ts` que documenta su
responsabilidad futura. No hay componentes de negocio, cliente API ni
pantallas reales todavía.

## Requisitos

- Node.js 20 o superior.
- pnpm.

## Comandos

```bash
pnpm install
pnpm dev
pnpm build
pnpm typecheck
pnpm lint
pnpm format
pnpm test:unit
pnpm test:e2e
```

`shared/api/generated` se agrega cuando exista un bundle OpenAPI que
generar; Pinia se agrega solo si aparece estado compartido real entre rutas,
según `DEC-033`.
