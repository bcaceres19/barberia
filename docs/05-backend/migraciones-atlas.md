---
titulo: "Gestión de migraciones PostgreSQL con Atlas"
version: "1.2"
estado: "Decisión confirmada"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-11"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../00-control/matriz-trazabilidad.md"
  - "base-datos.md"
  - "estandar-base-datos.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../04-arquitectura/stack-despliegue-operacion.md"
---

# Gestión de migraciones PostgreSQL con Atlas

## 1. Decisión

El proyecto usa **Atlas CLI** con un flujo de **migraciones SQL versionadas** para crear, validar, aplicar y rastrear cambios de PostgreSQL.

El SQL versionado es la fuente de verdad. No se aplica el estado declarativo directamente sobre piloto o producción, porque RLS, restricciones de exclusión, funciones, roles y cambios de datos requieren revisión humana del SQL exacto.

El flujo obligatorio usa capacidades locales de Atlas y no requiere Atlas Cloud ni una suscripción Pro. Las funciones pagas de linting avanzado, pruebas administradas, aprobaciones o detección de drift pueden evaluarse después, pero no son una dependencia del MVP.

Fuente normativa: `DEC-036`.

## 2. Comparación realizada

| Herramienta | Ventajas relevantes | Límite para este proyecto | Resultado |
| --- | --- | --- | --- |
| Atlas | SQL versionado, estado por base, `dry-run`, validación, generación/diff y `atlas.sum` para integridad lineal | algunas capacidades avanzadas requieren Pro; se excluyen del flujo obligatorio | Elegida |
| Goose | binario Go ligero, migraciones SQL/Go y comandos `up`, `down`, `status`, `version` | su conjunto estable se concentra en ejecutar versiones, no en planear y proteger todo el directorio | Alternativa simple |
| `golang-migrate` | CLI y biblioteca Go madura, varios drivers, funcionamiento deliberadamente mínimo | aporta aplicación y estado sucio, pero menos gobierno preventivo del esquema | Alternativa mínima |
| Flyway | tabla de historia detallada, checksums y validación madura | mayor peso operativo para este monolito y capacidades de modelo avanzadas ligadas a ediciones comerciales | No elegida |

Atlas ofrece el mejor equilibrio para una persona: conserva SQL legible, automatiza la promoción y agrega controles de integridad sin introducir un servicio permanente ni acoplar las migraciones al binario Go.

## 3. Estructura del repositorio

```text
database/
  atlas.hcl
  migrations/
    atlas.sum
    20260806143000_create_barbershop.sql
  seeds/
  testdata/
  tests/
  README.md
```

Reglas:

1. `atlas.hcl` define nombres de ambientes y obtiene URLs únicamente de variables de entorno o secretos de CI.
2. `database/migrations` y `atlas.sum` se versionan en Git.
3. La versión de Atlas se fija en el entorno de desarrollo y CI; no se usa automáticamente “latest”.
4. Las migraciones usan el timestamp ordenable generado por Atlas y una descripción `snake_case`.
5. Cada versión contiene un archivo SQL ascendente. No se crean parejas `.up.sql`/`.down.sql`.
6. Seeds ficticios y pruebas no se mezclan con migraciones productivas, salvo datos de referencia indispensables.
7. No se guarda URL, contraseña, token o certificado dentro de `atlas.hcl` ni del repositorio.

## 4. Flujo de creación

Los comandos se ejecutan desde `database/` o indicando su configuración explícitamente.

```text
atlas migrate new create_barbershop
# editar y revisar el SQL generado
atlas migrate hash
atlas migrate validate --env local
```

Secuencia obligatoria:

1. partir de la última rama principal y revisar `atlas migrate status` en local;
2. crear el archivo con `atlas migrate new <descripcion>`;
3. escribir o revisar manualmente el SQL, incluidos RLS, roles, constraints y datos existentes;
4. actualizar `atlas.sum` con `atlas migrate hash`;
5. ejecutar `atlas migrate validate` contra PostgreSQL efímero;
6. aplicar desde cero y desde la versión anterior con datos representativos;
7. ejecutar pruebas de integridad, RLS y concurrencia;
8. revisar `dry-run`, bloqueos, plan de recuperación y compatibilidad con la aplicación;
9. integrar juntos migración, código compatible, pruebas y documentación afectada.

`atlas.sum` nunca se resuelve aceptando ciegamente un hash nuevo. Si cambia una migración ya compartida, el cambio se rechaza y se crea otra migración.

## 5. Configuración por ambientes

`atlas.hcl` mantiene una sola definición de migraciones y configuraciones separadas para `local`, `test`, `pilot` y `production`. Las URLs reales llegan desde el entorno.

| Ambiente | Aplicación | Autorización |
| --- | --- | --- |
| Local | manual, recreable y con datos ficticios | desarrollador |
| Test/CI | automático sobre PostgreSQL efímero | pipeline del commit |
| Piloto | automático desde artefacto aprobado | aprobación del responsable del piloto |
| Producción | etapa separada antes del despliegue del API | aprobación manual y respaldo verificado |

La misma migración y el mismo `atlas.sum` avanzan entre ambientes. Está prohibido regenerar SQL distinto para piloto o producción.

## 6. Promoción y automatización

### Pull request

El pipeline:

1. verifica que `atlas.sum` coincide con el directorio;
2. valida la semántica SQL contra la versión PostgreSQL del proyecto;
3. aplica todas las migraciones sobre una base vacía;
4. actualiza una base en la versión anterior con datos de prueba;
5. ejecuta pruebas SQL y de integración con dos barberías;
6. falla ante archivos fuera de orden, modificados o SQL inválido.

### Despliegue

La etapa de migración ocurre una sola vez antes de desplegar procesos de aplicación:

```text
atlas migrate status --env pilot
atlas migrate apply --env pilot --dry-run
atlas migrate apply --env pilot
atlas migrate status --env pilot
```

Para producción se cambia el ambiente, pero no el artefacto. El pipeline debe:

- confirmar copia o punto de recuperación según el riesgo;
- guardar versión previa y objetivo;
- requerir aprobación manual;
- usar un rol migrador separado con DDL, nunca el rol ordinario del API;
- capturar salida, commit, actor, ambiente, inicio, fin y resultado;
- ejecutar smoke tests antes de continuar el despliegue;
- detenerse sin desplegar la aplicación si la migración falla.

Las migraciones no se ejecutan al arrancar el API o el worker. Así el rol de aplicación carece de DDL y varias instancias no compiten por modificar el esquema.

### Bootstrap de roles (`DEC-040`)

Los cuatro roles de PostgreSQL se aprovisionan **fuera de Atlas**, con un
administrador (o superusuario de bootstrap), **antes** de que Atlas se
conecte por primera vez a un ambiente nuevo. Ninguna migración crea estos
roles base a partir de `20260811145252_harden_roles_and_definer_functions.sql`
en adelante; esa migración fue la corrección única y de una sola vez para
ambientes que ya tenían `20260807170000_create_tenant_foundation.sql`
aplicada sin este modelo.

| Rol | Atributos | Uso |
| --- | --- | --- |
| `barberia_owner` | `NOLOGIN`, sin `SUPERUSER`/`CREATEDB`/`CREATEROLE`/`BYPASSRLS` | Dueño real de tablas, índices, secuencias y funciones existentes. Nunca se conecta. |
| `barberia_migrator` | `LOGIN`, `INHERIT`, miembro de `barberia_owner`, sin `SUPERUSER`/`CREATEDB`/`CREATEROLE`/`BYPASSRLS` | Ejecuta Atlas. Hereda automáticamente los privilegios de `barberia_owner` en cada sesión, sin `SET ROLE` (ver nota de diseño más abajo). |
| `barberia_app` | `LOGIN`, tenant-scoped, sin `SUPERUSER`/`CREATEDB`/`CREATEROLE`/`BYPASSRLS` | Rol del API. Fija `app.barbershop_id` por transacción. |
| `barberia_worker` | `LOGIN`, sin `SUPERUSER`/`CREATEDB`/`CREATEROLE`/`BYPASSRLS` | Procesos en segundo plano (notificaciones, retención). Solo `EXECUTE` en funciones de claim/finalización; sin `GRANT` general sobre tablas de negocio. |

Procedimiento de bootstrap por ambiente nuevo (desarrollo, piloto, producción):

1. Un administrador se conecta como superusuario o con `CREATEROLE` y crea los
   cuatro roles con los atributos de la tabla anterior.
2. `ALTER ROLE barberia_migrator INHERIT; GRANT barberia_owner TO
   barberia_migrator;` (con `INHERIT`, la membresía basta: no hace falta
   `SET ROLE`).
3. Fija la contraseña de `barberia_migrator`, `barberia_app` y
   `barberia_worker` con `ALTER ROLE ... PASSWORD '...'` desde el secreto del
   ambiente (gestor de secretos de despliegue); nunca en el repositorio ni en
   `atlas.hcl`.
4. **El mismo administrador con `CREATEROLE` (no `barberia_migrator`
   todavía) aplica las primeras TRES migraciones**
   (`20260807170000_create_tenant_foundation.sql`,
   `20260807170100_create_idempotency_record.sql`,
   `20260811145252_harden_roles_and_definer_functions.sql`), en ese orden.
   Confirmado contra PostgreSQL 14 real, con los cuatro roles ya creados
   (paso 1): `20260807170000` exige `CREATEROLE` de todas formas —`ALTER
   ROLE barberia_app SET search_path = ...` altera la configuración de OTRO
   rol, igual que crear uno, y eso requiere `CREATEROLE`/superusuario aunque
   el rol ya exista—; `20260807170100` exige `REFERENCES` sobre `barbershop`
   para su FK, privilegio que `barberia_migrator` no tiene hasta que
   `20260811145252` transfiera `barbershop`/`staff_user`/`idempotency_record`
   a `barberia_owner` (de quien `barberia_migrator` sí hereda). Por eso las
   tres van juntas bajo el ejecutor con `CREATEROLE`, no solo la primera y la
   tercera. Este paso es manual y puntual: no se automatiza con Atlas
   conectado como `barberia_migrator`, y no vuelve a repetirse para
   migraciones futuras (estas tres son las únicas que tocan roles u
   ownership).
5. Recién entonces Atlas se conecta con `barberia_migrator` (sin
   `CREATEROLE`) y aplica el resto del directorio de migraciones. Verificado
   contra PostgreSQL 14 real: `barberia_migrator` sin `CREATEROLE` aplica
   `20260811154100` y `20260811220000` sin error a partir de aquí.

**Por qué `INHERIT` y no `SET ROLE` por migración.** Se probó primero un
diseño con `barberia_migrator` `NOINHERIT` y `SET ROLE barberia_owner; ...
RESET ROLE;` dentro de cada migración que necesitara actuar como propietario.
Contra PostgreSQL real, esto rompe la propia aplicación de Atlas: mientras
aplica una migración, Atlas escribe su progreso en el esquema
`atlas_schema_revisions` (visible solo para quien lo creó, normalmente
`barberia_migrator`); con `SET ROLE barberia_owner` todavía activo, esa
escritura falla con `permission denied for schema atlas_schema_revisions` y
la migración completa se revierte. La alternativa robusta es que
`barberia_migrator` sea `INHERIT` y miembro de `barberia_owner`: adquiere sus
privilegios automáticamente, sin cambiar nunca de rol activo. Consecuencia
aceptada: los objetos que cree una migración nueva quedan `owned` por
`barberia_migrator` (quien la ejecuta), no por `barberia_owner`; es
reproducible porque `barberia_migrator` es el único rol que corre
migraciones. Las políticas RLS administrativas de una tabla nueva se escriben
igual `FOR ALL TO barberia_owner`: por membresía heredada, `barberia_migrator`
las satisface sin que el objeto sea literalmente suyo.

Rotar la contraseña de `barberia_migrator`, `barberia_app` o `barberia_worker`
no requiere una migración: es un cambio operativo sobre el rol existente.

## 7. Trazabilidad

La evidencia de un cambio se compone de cuatro niveles:

1. **Git:** autor, revisión, decisión relacionada y contenido SQL.
2. **`atlas.sum`:** orden e integridad criptográfica del directorio de migraciones.
3. **`atlas_schema_revisions`:** versiones y ejecución registradas por Atlas en cada base.
4. **CI/CD:** commit exacto, ambiente, actor de aprobación, tiempos, salida y resultado del despliegue.

`atlas_schema_revisions` pertenece a Atlas y no se modifica mediante SQL manual. `atlas migrate status` es la forma normal de consultarla. Una intervención excepcional usa comandos documentados de Atlas, aprobación y registro de incidente.

El historial de la herramienta no reemplaza la auditoría funcional del sistema ni el respaldo.

## 8. Transacciones y migraciones especiales

- DDL compatible se ejecuta transaccionalmente.
- `CREATE INDEX CONCURRENTLY` vive solo en una migración marcada sin transacción:

```sql
-- atlas:txmode none

CREATE INDEX CONCURRENTLY idx_appointment_shop_starts_at
ON appointment (barbershop_id, starts_at);
```

- Una migración no transaccional incluye precondiciones, forma de detectar aplicación parcial y avance correctivo.
- Transformaciones grandes se separan en expansión, backfill por lotes, validación y contracción posterior.
- Un cambio destructivo no se mezcla con el despliegue que deja de leer la estructura anterior.
- SQL dependiente de datos incluye consultas previas que fallen claramente si la condición de seguridad no se cumple.

## 9. Reversión

El proyecto usa **roll-forward** como mecanismo normal:

1. desplegar código compatible con estructura anterior y nueva;
2. aplicar expansión y migrar datos;
3. corregir con una nueva migración si aparece un defecto;
4. retirar estructura antigua en una versión posterior.

No se depende de migraciones `down` preescritas para producción. Si hubo pérdida o corrupción posible, se detiene el despliegue y se decide entre migración correctiva y restauración. Un rollback de la aplicación solo es seguro si el esquema mantiene compatibilidad hacia atrás.

En local, la base puede recrearse desde cero para volver a un estado conocido.

## 10. Límites de la decisión

- Atlas administra esquema y migraciones; no es un ORM ni reemplaza repositorios SQL.
- El `dry-run` permite revisar lo pendiente, pero no predice por sí solo duración, bloqueos o validez de datos productivos.
- `atlas migrate validate` comprueba integridad y semántica en una base de desarrollo; las pruebas del proyecto siguen siendo obligatorias.
- El linting avanzado de Atlas es una función Pro desde la versión indicada por su documentación actual y no forma parte de las puertas mínimas.
- Cualquier futura dependencia de Atlas Cloud, Pro o flujos declarativos en producción requiere una decisión nueva con costo y plan de salida.

## 11. Lista de revisión

- [ ] El archivo fue creado por Atlas y tiene descripción clara.
- [ ] El SQL es ascendente, determinista y compatible con PostgreSQL del proyecto.
- [ ] `atlas.sum` fue actualizado y revisado.
- [ ] No se modificó una migración ya compartida o aplicada.
- [ ] Pasa validación en vacío y actualización con datos.
- [ ] RLS, restricciones, roles e índices tienen pruebas.
- [ ] Se evaluaron bloqueo, duración y aplicación parcial.
- [ ] La aplicación es compatible durante toda la promoción.
- [ ] `dry-run`, versión inicial y versión final quedan en evidencia de CI.
- [ ] Existe plan de avance correctivo o restauración.

## 12. Referencias oficiales

- [Migraciones versionadas con Atlas](https://atlasgo.io/versioned/apply)
- [Integridad del directorio y `atlas.sum`](https://atlasgo.io/concepts/migration-directory-integrity)
- [`atlas migrate validate`](https://atlasgo.io/cli-reference#atlas-migrate-validate)
- [Ediciones y capacidades de Atlas](https://atlasgo.io/community-edition)
- [Comandos estables de Goose](https://pressly.github.io/goose/documentation/cli-commands/)
- [`golang-migrate`](https://github.com/golang-migrate/migrate)
- [Historial de esquema de Flyway](https://documentation.red-gate.com/flyway/flyway-concepts/migrations/flyway-schema-history-table)
