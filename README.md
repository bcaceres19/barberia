# Sistema de agenda para barberías

Este repositorio reúne la definición de producto y requisitos de un sistema de agenda para barberías. El objetivo del MVP es que un barbero pueda operar su jornada, permitir reservas públicas, evitar cruces de horario y conservar trazabilidad de los cambios.

## Estado actual

**Fase: base técnica del backend y del frontend.** El alcance P0, las reglas principales, la máquina de estados, el stack y la estrategia de datos ya tienen decisiones confirmadas. Sobre esa base existe ahora el esqueleto ejecutable descrito en la arquitectura: `apps/api` (Go, arranca y expone `/health`), `apps/web` (Vue 3 + TypeScript + Vite, compila y pasa lint), `database/` (Atlas configurado, sin migraciones) y `api/openapi/` (documento de entrada válido, sin operaciones). Ningún módulo de dominio, endpoint de negocio ni tabla existe todavía; se agregan uno por uno siguiendo `AGENTS.md`.

El 7 de agosto de 2026 se agregó el primer backlog verificable: [`docs/10-backlog/plan-bloques.md`](docs/10-backlog/plan-bloques.md) ordena las 45 funciones P0 en siete bloques y [`docs/02-requisitos/historias-usuario.md`](docs/02-requisitos/historias-usuario.md) contiene las historias `HU-001`–`HU-012` del bloque base, con sus criterios de aceptación. Cinco de esas historias están bloqueadas por las dudas `DP-SEG-04`, `DP-SEG-05` y `DP-SEG-06`, que corresponden al propietario.

El 5 y 6 de agosto de 2026 se formalizaron las respuestas del propietario contenidas en [`respuesta-manuales/respuesta-propuestas-oc.txt`](respuesta-manuales/respuesta-propuestas-oc.txt) y [`respuesta-manuales/respuesta-dudas-pendientes.txt`](respuesta-manuales/respuesta-dudas-pendientes.txt), junto con instrucciones posteriores. Las decisiones `DEC-001`–`DEC-039` viven en [`docs/00-control/registro-decisiones.md`](docs/00-control/registro-decisiones.md); los archivos manuales quedan como evidencia de origen.

## Cómo revisar o ejecutar el proyecto

Para trabajar con la documentación:

1. Abra [`docs/README.md`](docs/README.md) para conocer el mapa documental y la precedencia de las fuentes.
2. Consulte el [registro de decisiones](docs/00-control/registro-decisiones.md) antes de cambiar reglas o alcance.
3. Revise las [dudas pendientes](docs/00-control/dudas-pendientes.md) y las [contradicciones](docs/00-control/contradicciones.md) antes de implementar algo que dependa de ellas.
4. Aplique la [guía de contribución](CONTRIBUTING.md) y el [flujo Git/GitHub](docs/03-desarrollo/flujo-git-github.md) para todo cambio.
5. Use un visor de Markdown para navegar los enlaces.

Comandos opcionales de inventario en PowerShell:

```powershell
rg --files docs
rg -n "Estado:|DEC-|CT-|DP-" docs
```

Para ejecutar la base técnica actual (requiere Go 1.23+, Node.js 20+ y pnpm):

```bash
# Backend: arranca en :8080 y expone GET /health
cd apps/api && go run ./cmd/api

# Frontend: servidor de desarrollo Vite
cd apps/web && pnpm install && pnpm dev

# Contrato OpenAPI: lint, bundle y documentación
pnpm install && pnpm run openapi:lint
```

No hay base de datos, migraciones ni PostgreSQL configurados todavía; ver
[`database/README.md`](database/README.md). Los detalles de cada parte
están en `apps/api/README.md`, `apps/web/README.md` y
`api/openapi/README.md`.

## Dónde está cada cosa

| Ruta | Contenido |
| --- | --- |
| [`apps/api/`](apps/api/) | Backend Go (Chi v5 pendiente): procesos `api` y `worker`, base de plataforma y módulos vacíos |
| [`apps/web/`](apps/web/) | Frontend Vue 3 + TypeScript + Vite: arranque, router y tokens del sistema visual |
| [`database/`](database/) | Configuración de Atlas y carpetas de migraciones, seeds, testdata y pruebas |
| [`api/openapi/`](api/openapi/) | Documento de entrada OpenAPI 3.1.2 y estructura de paths/components |
| [`docs/README.md`](docs/README.md) | Mapa documental, fuentes de verdad y flujo de cambios |
| [`docs/00-control/`](docs/00-control/) | Decisiones, contradicciones, trazabilidad, cambios, dudas, supuestos y glosario |
| [`docs/01-producto/`](docs/01-producto/) | Alcance del MVP, prioridades y reglas de negocio |
| [`docs/02-requisitos/`](docs/02-requisitos/) | Requisitos detallados, estados de las citas e historias de usuario |
| [`docs/03-desarrollo/`](docs/03-desarrollo/) | Estándares de código, diseño visual y estrategia de pruebas |
| [`docs/04-arquitectura/`](docs/04-arquitectura/) | Go + Chi v5, Vue 3 + TypeScript + Vite, despliegue y operación |
| [`docs/05-backend/`](docs/05-backend/) | PostgreSQL, RLS, normalización y migraciones administradas con Atlas |
| [`docs/06-api/`](docs/06-api/) | Estándar OpenAPI y gobierno del contrato HTTP |
| [`docs/10-backlog/`](docs/10-backlog/) | Plan de bloques y prompts de implementación |
| [`AGENTS.md`](AGENTS.md) | Reglas operativas obligatorias para implementar cambios |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Flujo resumido de issues, ramas, commits y pull requests |
| [`.github/`](.github/) | Plantillas de issues y pull requests |
| [`respuesta-manuales/`](respuesta-manuales/) | Evidencia original de respuestas; no es fuente normativa |

## Regla de gobierno

Una respuesta informal no cambia el producto por sí sola. Debe registrarse como `DEC-*`, actualizar los documentos afectados, enlazarse en la matriz de trazabilidad y anotarse en el historial de cambios. Si dos documentos no pueden cumplirse a la vez, se registra una `CT-*` y no se oculta la diferencia.
