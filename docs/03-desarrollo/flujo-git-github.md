---
titulo: "Flujo de Git y GitHub"
version: "1.0"
estado: "Obligatorio para desarrollo"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../../CONTRIBUTING.md"
  - "../../.github/PULL_REQUEST_TEMPLATE.md"
  - "../00-control/registro-decisiones.md"
  - "estrategia-pruebas.md"
  - "../05-backend/migraciones-atlas.md"
  - "../06-api/estandar-openapi.md"
---

# Flujo de Git y GitHub

## 1. Objetivo y decisión

Este estándar mantiene `main` estable, hace revisable cada cambio y conserva una traza entre necesidad, código, pruebas y despliegue. La estrategia es **GitHub Flow** con una sola rama permanente, ramas cortas por cambio, pull request obligatorio y merge por squash.

No se crean `develop`, ramas por ambiente ni ramas `release/*` en el MVP. Los ambientes despliegan un commit o tag verificable, no una rama distinta. Se revisará esta decisión solo si aparecen versiones soportadas en paralelo que realmente lo exijan.

Fuente normativa: `DEC-038`.

## 2. Ramas

### 2.1 Rama permanente

- `main` es la única fuente de verdad y debe permanecer desplegable.
- No se hacen commits ni pushes directos a `main`, aunque exista una sola persona desarrollando.
- No se permite force push, reescritura ni eliminación de `main`.
- Todo cambio llega mediante un pull request con sus controles aprobados.

### 2.2 Ramas de trabajo

Cada rama nace de `main` actualizada, atiende una sola preocupación y se elimina al integrar. Debe durar horas o pocos días; si crece, se divide mediante cambios compatibles o feature flags, no se mantiene abierta durante semanas.

Formato:

```text
<tipo>/<issue>-<descripcion-kebab-case>
```

Tipos permitidos:

| Prefijo | Uso | Ejemplo |
| --- | --- | --- |
| `feat/` | Función o comportamiento nuevo | `feat/123-reserva-publica` |
| `fix/` | Corrección de un defecto | `fix/148-evitar-cita-duplicada` |
| `refactor/` | Cambio interno sin alterar conducta | `refactor/166-extraer-politica-cancelacion` |
| `test/` | Pruebas sin cambio funcional | `test/171-concurrencia-reservas` |
| `docs/` | Documentación | `docs/175-contrato-disponibilidad` |
| `chore/` | Mantenimiento o herramientas | `chore/180-actualizar-toolchain-go` |
| `ci/` | Integración o despliegue continuo | `ci/182-validar-openapi` |
| `hotfix/` | Incidente urgente de producción | `hotfix/201-revocar-sesiones` |

Reglas de nombre:

- minúsculas, caracteres ASCII y palabras separadas por guiones;
- número de issue obligatorio para funciones, defectos, seguridad, API, datos y operación;
- sin nombres de personas, fechas, `final`, `prueba`, `cambios` ni ramas personales permanentes;
- un ajuste editorial trivial puede omitir issue, por ejemplo `docs/corregir-enlace-readme`.

## 3. Ciclo de trabajo local

Ejemplo de inicio:

```bash
git switch main
git pull --ff-only
git switch -c feat/123-reserva-publica
```

Antes de cada commit:

```bash
git status --short
git diff
git add <rutas-especificas>
git diff --cached
```

Se agregan rutas explícitas cuando el árbol contiene cambios no relacionados. No se usa `git add .` sin revisar previamente qué incluirá. Antes de abrir o actualizar el PR se ejecutan los controles locales afectados.

Para incorporar cambios recientes de `main`:

```bash
git fetch origin
git rebase origin/main
```

Después de reescribir una rama ya publicada, solo se permite:

```bash
git push --force-with-lease
```

`--force-with-lease` se usa únicamente en una rama de trabajo que pertenece al autor y después de comprobar que nadie más depende de ella. Nunca se usa `--force`, nunca se reescribe una rama compartida y nunca se fuerza `main`.

Flujo opcional con GitHub CLI después del primer push:

```bash
git push -u origin feat/123-reserva-publica
gh pr create --draft --template .github/PULL_REQUEST_TEMPLATE.md
gh pr checks --watch
gh pr ready
gh pr merge --squash --delete-branch
```

Cada comando se ejecuta en su momento, no como una secuencia automática: el PR solo se marca listo cuando cumple su definición de terminado y solo se integra cuando GitHub confirma las reglas. No se usa `gh pr merge --admin` para saltarse controles ordinarios.

## 4. Commits

### 4.1 Formato

Se adopta Conventional Commits 1.0.0:

```text
<tipo>(<alcance>): <descripcion>

<cuerpo opcional que explica por qué y las restricciones>

<pies opcionales>
```

Tipos aceptados: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`, `chore`, `revert`.

Alcances iniciales: `booking`, `schedule`, `auth`, `notifications`, `api`, `web`, `db`, `ops`, `docs`, `ci`. Se añade otro alcance solo si representa una capacidad estable del repositorio.

Ejemplos:

```text
feat(booking): crea una reserva pública idempotente
fix(db): impide cruces de citas del mismo barbero
docs(api): documenta errores de disponibilidad
test(auth): cubre la expiración de sesiones
```

Un cambio incompatible usa `!` y explica la transición:

```text
feat(api)!: reemplaza el campo start por starts_at

BREAKING CHANGE: los consumidores deben enviar starts_at en formato date-time.
```

### 4.2 Calidad de cada commit

- Expresa una sola intención y deja el repositorio en un estado coherente cuando sea viable.
- Incluye la prueba y documentación necesarias para esa intención; no separa artificialmente código y evidencia si juntos forman el cambio completo.
- La descripción es breve, imperativa, sin punto final y de hasta 72 caracteres.
- El cuerpo explica el motivo, las restricciones y decisiones que no se deducen del diff; no enumera cada archivo modificado.
- No mezcla reformateo masivo, renombres o dependencias con una función no relacionada.
- No incluye secretos, `.env`, datos personales reales, binarios, salidas de build ni archivos generados innecesarios.
- `WIP` y `fixup!` solo se toleran mientras el PR sea borrador; se limpian antes de marcarlo listo.
- No se modifica la autoría ni se reescribe historia compartida para ocultar un problema ya revisado.

El título del PR debe seguir el mismo formato porque será el mensaje final al hacer squash.

## 5. Issues antes de desarrollar

Se crea o enlaza un issue para toda función, defecto, migración, cambio de contrato, seguridad, operación o deuda no trivial. El issue registra:

- problema y resultado esperado, no una solución prematura;
- alcance incluido y explícitamente excluido;
- reglas `RN-*`, decisiones `DEC-*` y dependencias relevantes;
- criterios de aceptación observables;
- riesgos de datos, API, seguridad, interfaz y operación;
- estrategia mínima de prueba.

Un PR enlaza el issue en su descripción con `Closes #123` cuando lo resuelve por completo. Si solo entrega una parte, usa `Refs #123` y explica qué falta; no cierra el issue antes de cumplir todos sus criterios.

## 6. Pull requests

### 6.1 Apertura y tamaño

- Abrir un draft temprano cuando haga falta validar dirección o visibilizar trabajo.
- Cambiar a listo solo cuando el alcance esté completo, los controles locales pasen y no queden `TODO` sin referencia.
- Un PR atiende una sola preocupación y debe poder revertirse como unidad.
- Se prefieren menos de unas 400 líneas significativas modificadas, excluyendo archivos generados, lockfiles y migraciones mecánicas. Si no es posible dividirlo sin romper una transición, se explica en el PR.
- No se agregan arreglos oportunistas: se crea otro issue o rama.

### 6.2 Contenido obligatorio

La plantilla de `.github/PULL_REQUEST_TEMPLATE.md` exige:

- resumen del problema y de la solución;
- issue y trazabilidad normativa;
- cambios incluidos y fuera de alcance;
- comandos y resultado de pruebas;
- impacto en OpenAPI, frontend, PostgreSQL/Atlas, seguridad y operación;
- evidencia visual para cambios de interfaz;
- riesgo, despliegue y recuperación o reversión.

La lista de archivos no reemplaza un resumen. Las capturas no reemplazan pruebas y nunca muestran datos personales o secretos.

### 6.3 Revisión

El autor realiza una auto-revisión del diff completo antes de solicitar revisión. Un revisor comprueba, en este orden:

1. comportamiento, reglas y alcance;
2. seguridad, autorización y aislamiento por barbería;
3. integridad de datos, concurrencia y compatibilidad;
4. pruebas y evidencia;
5. claridad y mantenibilidad.

Toda conversación se responde y resuelve. Una observación válida pero fuera de alcance se convierte en issue enlazado; no se oculta ni infla el PR. Después de cambios materiales se solicita una nueva revisión.

Durante el trabajo individual, GitHub puede exigir PR con **cero aprobaciones**, pues el autor no puede sustituir una revisión independiente. Cuando participe otra persona con contexto, se exige al menos **una aprobación ajena al último cambio**, y se descartan aprobaciones obsoletas tras modificaciones materiales.

## 7. Controles antes de integrar

Los checks deben tener nombres únicos y pasar antes del merge. El conjunto se activa conforme exista cada parte del código:

| Check lógico | Evidencia mínima |
| --- | --- |
| Documentación | enlaces Markdown y estructura documental válidos |
| Go | formato, análisis estático, unitarias y build |
| Vue | formato, lint, tipos, unitarias/componentes y build |
| PostgreSQL | integración, RLS y concurrencia con PostgreSQL real |
| Atlas | hash, lint/validate, base vacía y actualización con datos |
| OpenAPI | lint, bundle, ejemplos y contrato frente a handlers/cliente |
| Sistema | E2E smoke de recorridos P0 afectados |
| Seguridad | análisis de dependencias y secretos según la etapa |

La fuente detallada es [estrategia-pruebas.md](estrategia-pruebas.md). No se desactiva un check, se reduce cobertura ni se marca una prueba como ignorada para integrar. Una excepción urgente sigue el procedimiento de la sección 11.

## 8. Integración e historial

- El único método habilitado es **Squash and merge**.
- El título final usa Conventional Commits y la descripción del squash resume el motivo o referencia el issue.
- Se deshabilitan merge commits y rebase merge en la interfaz para evitar historiales distintos según cada autor.
- Después del merge se elimina automáticamente la rama remota y el autor elimina la local cuando ya no la necesita.
- No se reutiliza una rama ya integrada para otra tarea.
- Para revertir producción se usa un nuevo PR con `revert:` y referencia al commit o PR original; no se borra historia.

## 9. Configuración de GitHub para `main`

Se prefiere un ruleset del repositorio. Si el plan o visibilidad del repositorio no ofrece rulesets, se configura una regla de protección equivalente. Si GitHub no permite ninguna de las dos, el proceso manual es una mitigación temporal que debe registrarse como riesgo y revisarse antes de sumar colaboradores o iniciar el piloto.

Configuración objetivo:

| Regla | MVP individual | Con colaboradores |
| --- | --- | --- |
| Requerir pull request | Sí | Sí |
| Aprobaciones requeridas | 0 | 1 como mínimo |
| Resolver conversaciones | Sí | Sí |
| Checks requeridos | Los ya implementados | Todos los aplicables |
| Rama actualizada con `main` | Opcional mientras no haya PR concurrentes | Sí; merge queue si el volumen lo justifica |
| Historial lineal | Sí | Sí |
| Método permitido | Squash | Squash |
| Bloquear force push y eliminación | Sí | Sí |
| Descartar aprobaciones obsoletas | No aplica | Sí |
| Revisión de CODEOWNERS | Pendiente de responsables reales | Sí en áreas sensibles |
| Commits firmados | Recomendado, no bloqueante al inicio | Activar cuando personas y bots estén preparados |

También se configura:

- `main` como rama por defecto;
- eliminación automática de ramas después del merge;
- título del PR como mensaje predeterminado del squash;
- lista de bypass vacía por defecto;
- nombres de checks únicos entre workflows.

No se crea `CODEOWNERS` hasta conocer usuarios o equipos reales con permisos de escritura. Al incorporarlos, como mínimo se asignan responsables para `.github/`, `database/`, `api/openapi/` y configuración de despliegue.

## 10. Cambios especiales

### 10.1 OpenAPI

Un cambio HTTP actualiza en la misma entrega coherente el contrato, implementación, cliente derivado y pruebas. Un cambio incompatible requiere versión o período de compatibilidad según [estandar-openapi.md](../06-api/estandar-openapi.md); no se oculta como `fix`.

### 10.2 Base de datos y Atlas

- Una migración aplicada no se edita ni se reemplaza.
- `atlas.sum` y el SQL viajan juntos.
- Se usa evolución expandir/migrar/contraer cuando código nuevo y anterior deban convivir.
- La eliminación o conversión destructiva se separa de la entrega que deja de usar el dato.
- El PR declara bloqueo esperado, compatibilidad, validación con datos y procedimiento de recuperación.

La norma completa está en [migraciones-atlas.md](../05-backend/migraciones-atlas.md).

### 10.3 Dependencias y archivos generados

Una dependencia nueva explica necesidad, alternativas, licencia, mantenimiento, tamaño y riesgo. El lockfile se incluye. El código generado identifica su fuente y comando; no se edita manualmente. Los diffs de generación se aíslan para poder revisar el cambio de origen.

## 11. Hotfix, bypass y recuperación

Un `hotfix/*` nace de `main`, conserva PR y ejecuta todos los checks que el tiempo permita sin poner datos o seguridad en mayor riesgo. Se prioriza un cambio mínimo y reversible.

El bypass de una regla es último recurso para indisponibilidad, pérdida/corrupción de datos o incidente de seguridad activo. Requiere:

1. issue o incidente con hora, responsable y motivo;
2. revisión del diff por otra persona si está disponible;
3. evidencia de los checks ejecutados;
4. plan explícito de reversión;
5. PR de regularización y retrospectiva en el siguiente día hábil.

“Tengo prisa”, un check incómodo o una rama desactualizada no justifican bypass.

## 12. Versiones y despliegues

- Cada despliegue identifica el SHA exacto; no se despliega “lo último” sin evidencia.
- Antes de `1.0.0` se usan tags SemVer `v0.x.y`; producción estable usa `vMAJOR.MINOR.PATCH`.
- Los tags de versión se crean sobre `main` después de los checks y se acompañan de notas de cambios y migraciones.
- No se reutiliza ni mueve un tag publicado.
- Se protege el patrón `v*` cuando el repositorio permita rulesets para tags.
- Los ambientes no tienen ramas propias: promueven el mismo artefacto validado.

## 13. Higiene y seguridad del repositorio

- Nunca versionar contraseñas, tokens, claves, certificados privados, dumps ni datos reales.
- Si un secreto se publica, se revoca y rota de inmediato; borrarlo en otro commit no lo elimina de la historia.
- Mantener `.gitignore`, `.gitattributes` y, al existir herramientas, archivos de versión y lockfiles.
- Usar LF para código, YAML, SQL y Markdown; PowerShell puede conservar CRLF.
- No subir salidas de build, cobertura, reportes locales o archivos grandes sin una necesidad documentada.
- Revisar cambios de lockfiles y acciones de terceros; fijar acciones a una versión o SHA confiable según el riesgo.
- Evitar datos personales en nombres de ramas, commits, issues, PR, capturas y logs de CI.

## 14. Definición de terminado para un PR

- [ ] El issue y el alcance están claros.
- [ ] El título cumple Conventional Commits.
- [ ] El diff contiene una sola preocupación y fue auto-revisado.
- [ ] Código, pruebas, contrato y documentación coinciden.
- [ ] Los checks requeridos pasan sin excepciones ocultas.
- [ ] Los impactos de tenant, seguridad, API, datos y operación fueron evaluados.
- [ ] Migraciones y recuperación están documentadas cuando aplican.
- [ ] Las conversaciones están resueltas y existe la aprobación requerida.
- [ ] El PR puede integrarse por squash y revertirse como unidad.
- [ ] No contiene secretos, datos personales ni artefactos locales.

## 15. Antipatrones prohibidos

- commits directos o force push a `main`;
- ramas permanentes `develop`, `qa`, `staging` o por persona;
- ramas grandes que mezclan función, refactor y mantenimiento;
- mensajes como `cambios`, `fix`, `final`, `prueba` o `WIP` en historial listo para integrar;
- PR sin issue para cambios de comportamiento o sin evidencia de pruebas;
- integrar con checks fallidos, conversaciones abiertas o contrato desactualizado;
- editar migraciones aplicadas, mover tags o desplegar un commit diferente al aprobado;
- usar un PR para reformatear archivos ajenos a la tarea;
- exponer secretos o datos de clientes en cualquier metadato de GitHub.

## 16. Referencias oficiales

- [GitHub Flow](https://docs.github.com/en/get-started/using-github/github-flow)
- [Reglas disponibles en rulesets](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/available-rules-for-rulesets)
- [Configuración de squash merge](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/configuring-commit-squashing-for-pull-requests)
- [Vincular un pull request con un issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/linking-a-pull-request-to-an-issue)
- [Code owners](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners)
- [Conventional Commits 1.0.0](https://www.conventionalcommits.org/es/v1.0.0/)
- [Manual de pull requests de GitHub CLI](https://cli.github.com/manual/gh_pr)
