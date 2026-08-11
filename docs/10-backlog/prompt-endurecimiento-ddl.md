---
titulo: "Prompt maestro para endurecer el DDL PostgreSQL"
version: "1.0"
estado: "Propuesta; herramienta de ejecución, no norma"
responsable: "Pendiente de aprobación del propietario"
ultima_actualizacion: "2026-08-11"
documentos_relacionados:
  - "../05-backend/revision-ddl-seguridad-2026-08-11.md"
  - "../05-backend/estandar-base-datos.md"
  - "../05-backend/migraciones-atlas.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../00-control/registro-decisiones.md"
---

# Prompt maestro para endurecer el DDL PostgreSQL

Copiar desde “Inicio del prompt” hasta “Fin del prompt”. Este prompt organiza el
trabajo, pero no autoriza a inventar decisiones ni a modificar migraciones aplicadas.

## Inicio del prompt

Trabaja como responsable senior de PostgreSQL, seguridad y backend Go en este
repositorio. Tu objetivo es corregir y endurecer el DDL conforme al proyecto, sin
ampliar el alcance funcional ni ocultar decisiones pendientes.

### Resultado exigido

Entrega el esquema mediante migraciones Atlas roll-forward, pruebas reales,
documentación normativa coherente y evidencia de upgrade. El trabajo termina solo
cuando:

1. los roles y propietarios efectivos son reproducibles y de mínimo privilegio;
2. API y worker no comparten privilegios globales;
3. todas las funciones `SECURITY DEFINER` están endurecidas y probadas;
4. idempotencia y reclamación de trabajos son correctas bajo concurrencia real;
5. las invariantes de horarios, citas, historial y tenant están protegidas en la
   capa más baja adecuada;
6. la anonimización elimina o transforma todas las copias aprobadas de datos
   personales y revoca tokens;
7. las migraciones funcionan tanto desde una base vacía como al actualizar una base
   que ya tiene las migraciones `20260807170000` y `20260807170100` aplicadas.

### Reglas inviolables

- Lee completamente `AGENTS.md`, `docs/00-control/registro-decisiones.md`, el
  alcance P0, las reglas afectadas, `docs/05-backend/estandar-base-datos.md`,
  `docs/05-backend/migraciones-atlas.md`,
  `docs/03-desarrollo/estrategia-pruebas.md` y
  `docs/05-backend/revision-ddl-seguridad-2026-08-11.md` antes de editar.
- Confirma que existe un issue y crea una rama corta desde `main` actualizado según
  `docs/03-desarrollo/flujo-git-github.md`. Si Git aún no está inicializado, detén
  las mutaciones y entrega solo el preflight; no simules el flujo obligatorio.
- No edites las dos migraciones aplicadas ni recalcules su contenido. Toda corrección
  sobre ellas se realiza en migraciones nuevas y actualiza `atlas.sum` con Atlas.
- `database/modelo-fisico-referencia.sql` no es una migración aplicada: puede
  corregirse para que refleje el destino aprobado, pero nunca debe contradecir las
  migraciones nuevas ni la documentación normativa.
- No uses ORM, `jsonb`, borrado en cascada sobre datos con vida/retención propia,
  DDL al arrancar, secretos, datos personales reales o una dependencia nueva sin
  justificación aprobada.
- No hagas llamadas de red dentro de una transacción. No reduzcas pruebas ni cambies
  snapshots para esconder fallos.
- Si una conclusión depende de una duda o contradicción no aprobada, regístrala en el
  artefacto canónico y detente en esa parte. Continúa solo con correcciones
  independientes y reversibles.

### Fase 0 — Preflight obligatorio, sin cambios de esquema

1. Ejecuta `git status`, identifica rama/base y comprueba que no vas a mezclar cambios
   ajenos.
2. Ejecuta `atlas version` y valida `database/migrations/atlas.sum`.
3. Arranca o conecta una instancia desechable de la misma versión mayor de PostgreSQL
   soportada. No uses una base compartida ni datos reales.
4. Aplica las migraciones desde cero y captura desde catálogos:
   - versión de PostgreSQL y extensiones;
   - atributos y membresías de roles;
   - propietario de esquema, tablas, índices, secuencias y funciones;
   - privilegios explícitos y por defecto;
   - RLS habilitada/forzada y políticas por tabla;
   - definición, propietario, `prosecdef` y `proconfig` de funciones.
5. Crea otro escenario que empiece exactamente con las dos migraciones actuales y
   datos ficticios de dos tenants. Será la prueba de upgrade.
6. Compara el estado real con los comentarios de las migraciones. No asumas que
   `barberia_migrator` posee objetos: demuéstralo en `pg_class`, `pg_proc` y
   `pg_namespace`.

Entrega un resumen del preflight antes de implementar. Si no hay PostgreSQL real,
Docker/servicio equivalente o permiso para crear la rama, informa el bloqueo; no
declares validado el DDL solo porque el checksum sea correcto.

### Fase 1 — Registrar decisiones y dividir el trabajo

Abre issues separados para no mezclar preocupaciones:

1. bootstrap, ownership, roles y `SECURITY DEFINER`;
2. idempotencia y limpieza;
3. integridad del modelo de horarios/citas/historial;
4. workers de notificación y retención;
5. privacidad y anonimización;
6. índices, validaciones y pruebas RLS.

Antes de codificar, registra y solicita aprobación para:

- versión mayor mínima de PostgreSQL;
- idioma/vocabulario de `appointment_history.event_type`;
- fecha ancla y matriz completa de anonimización;
- identidad/reutilización de `customer` por teléfono;
- correo global frente a pertenencia multi-barbería;
- ciclo de vida de `barber` y vínculo con `staff_user`;
- segundo/tercer recordatorio;
- semántica concurrente de idempotencia;
- modelo definitivo de roles.

No selecciones tú una alternativa de producto. En cada duda muestra impacto, dos o
tres alternativas concretas y recomendación técnica, pero espera aprobación.

### Fase 2 — Roles, propietarios y privilegios

Implementa el resultado aprobado con estas propiedades mínimas:

- un propietario de objetos `NOLOGIN` es preferible; el login migrador no se usa por
  el API y el rol API no posee objetos;
- `barberia_app` permanece sin superusuario, `CREATEDB`, `CREATEROLE` ni
  `BYPASSRLS`;
- crea un rol `barberia_worker` separado, sin acceso general de handlers y sin
  `BYPASSRLS`;
- los roles se aprovisionan antes de que Atlas se conecte; documenta el procedimiento
  bootstrap para un servicio gestionado y no dependas de que una migración cree el
  rol con el que ya está conectada;
- una migración correctiva transfiere ownership/grants solo después de consultar los
  propietarios reales; debe ser segura si el estado esperado ya existe;
- fija privilegios por defecto del propietario, especialmente `REVOKE EXECUTE ON
  FUNCTIONS FROM PUBLIC`, y concede cada capacidad de forma explícita;
- elimina `EXECUTE` innecesario del API sobre funciones trigger;
- si los tests necesitan `SET ROLE`, aprovisiona la membresía exacta o conecta con el
  rol real; no uses superusuario para afirmar que RLS funciona.

Para toda función `SECURITY DEFINER`:

- propietario controlado, preferiblemente `NOLOGIN`;
- `SET search_path = ''` y referencias `public.objeto`/`pg_catalog.función`
  completamente calificadas;
- `REVOKE ALL ... FROM PUBLIC` antes de conceder `EXECUTE`;
- parámetros nulos, negativos, tamaños máximos y formato validados;
- retorno mínimo, sin datos personales ni secretos;
- concesión exclusiva a API o worker según el caso;
- prueba de que un objeto homónimo en un esquema escribible no altera la ejecución.

### Fase 3 — Corrección de HU-001 y HU-004 por roll-forward

No edites `20260807170000_create_tenant_foundation.sql` ni
`20260807170100_create_idempotency_record.sql`.

Para HU-001:

- revoca `INSERT`/`DELETE` de `staff_user` si ninguna operación P0 aprobada los
  necesita todavía; concede solo columnas/operaciones usadas;
- elimina la política de borrado no aprobada;
- normaliza datos de entrada de forma coherente (`btrim`, minúsculas, E.164) entre
  API y restricciones;
- añade baseline verificable de privilegios y prueba todos los atributos de roles,
  owners y RLS con dos tenants.

Para HU-004:

- implementa la semántica concurrente aprobada:
  - opción A: la segunda solicitud espera con límite y reproduce el resultado
    confirmado; o
  - opción B: usa `pg_try_advisory_xact_lock` derivado de tenant+clave y devuelve un
    conflicto temporal mientras la primera transacción continúa;
- no persistas una reserva en otra transacción salvo que diseñes lease, expiración y
  recuperación de caída;
- revoca `DELETE` y su política al rol API; la limpieza de vencidos pertenece al
  worker/mantenimiento mediante una operación acotada; una función estrecha debe
  permitir reservar de forma atómica una clave ya vencida aunque la limpieza aún no
  haya corrido, sin abrir borrado arbitrario al API;
- considera almacenar un digest de la clave de idempotencia en vez del valor recibido
  en claro, conservando la unicidad por tenant;
- valida la huella como hash con algoritmo y codificación definidos;
- limita `response_content_type` y tamaño de `response_body`; guarda solo la respuesta
  mínima necesaria y aplica su tratamiento de privacidad;
- demuestra que un fallo revierte efecto y registro, que un reintento legítimo puede
  continuar y que una clave completada no puede borrarse desde el API.

### Fase 4 — Corregir el modelo físico antes de nuevas migraciones

Aplica solo lo que ya esté aprobado:

1. Corrige el orden de dependencias: crea `customer` antes de `appointment` o añade
   la FK `appointment -> customer` después de que ambas tablas existan.
2. Renombra `working_hours` y `working_hours_override` a singular antes de migrarlas.
3. Remodela la excepción diaria para que una fecha sea cerrada o tenga segmentos,
   nunca ambas; impide solapes incluso con escrituras concurrentes.
4. Protege también el solape de jornadas semanales usando una representación
   declarativa o bloqueo/constraint trigger mínimo y determinista.
5. Asegura que una fila `time_block_series_date` solo pertenezca a una serie
   `date_list` y esté dentro de su rango efectivo.
6. Sustituye los `ON DELETE SET NULL` compuestos que intentarían anular
   `barbershop_id`; usa `RESTRICT` y desvinculación explícita si corresponde.
7. Sustituye `CASCADE` sobre bloqueos con retención permanente por `RESTRICT`.
8. Mantén `barber` dentro de HU-021: id, tenant, nombre y timestamps. No añadas
   borrado, desactivación, orden ni vínculo automático sin decisión aprobada.
9. Añade la FK `(barbershop_id, actor_customer_id) -> customer` y su índice.
10. Garantiza que una transición de cita, su historial, cambios y programación de
   notificaciones se persistan atómicamente; impide cambios a id, tenant,
   `created_at` y snapshots fuera de comandos aprobados.
11. Valida la igualdad entre intervalo y duración snapshot si PostgreSQL permite una
    expresión inmutable y correcta para DST; si no, conserva servicio + prueba y
    documenta la razón con evidencia.
12. Al cambiar el teléfono verificado de un usuario, invalida
    `phone_verified_at` de forma atómica.
13. Inserta/backfill la regla ordinal 1 a 30 minutos exigida por `DEC-018`, sin
    impedir que el usuario elija explícitamente cero recordatorios después.
14. Elimina `notification_attempt.channel` o impide por diseño que difiera del padre.
15. Endurece `notification_schedule_outcome_ck`: cada estado debe tener exactamente
    los timestamps que le corresponden.

### Fase 5 — Workers correctos y de mínimo privilegio

Rediseña `notification_claim_due` como protocolo de lease:

1. en una transacción corta, selecciona con `FOR UPDATE SKIP LOCKED` y cambia
   `pending/retryable -> processing` con `claim_token`, `claimed_at`,
   `lease_expires_at` y contador;
2. devuelve solo identificadores y el token de reclamación;
3. confirma la transacción antes de llamar correo/WhatsApp;
4. revalida cita, canal y destinatario bajo el tenant antes de enviar;
5. finaliza con `UPDATE ... WHERE claim_token=? AND status='processing'`;
6. registra intento append-only y siguiente reintento/backoff en una transacción
   corta;
7. recupera leases vencidos con un límite pequeño;
8. si el proveedor admite clave idempotente, úsala para reducir duplicados tras una
   caída entre envío y confirmación.

`p_limit` debe rechazar `NULL`, valores menores de 1 y superiores al máximo aprobado.
Solo `barberia_worker` puede reclamar/finalizar trabajos globales. Prueba dos workers,
caída después de claim, lease vencido, fallo temporal, fallo permanente y ausencia de
doble claim vigente.

Para retención, no uses un claim efímero. Si toda la anonimización cabe en PostgreSQL,
hazla atómicamente en lotes pequeños. Si es multietapa, usa el mismo patrón de lease.

### Fase 6 — Privacidad completa

Implementa la matriz aprobada y prueba al menos:

- `customer.full_name`, `phone`, `email`;
- `appointment.attendee_name` y `customer_note`;
- `appointment_history.reason` y `appointment_history_change.previous_value/new_value`;
- hashes y filas de `appointment_access_token`;
- `idempotency_record.response_body` y cualquier clave que permita correlación;
- identificadores o errores de proveedor que puedan contener datos enviados.

La anonimización debe ser idempotente, conservar solo hechos operativos autorizados,
revocar tokens públicos y no romper FKs ni historial. No llames “anónimo” a un registro
que conserve el nombre original. Prueba solicitud individual, vencimiento masivo, dos
tenants, repetición y caída a mitad de lote.

### Fase 7 — Índices y validaciones

- Inventaría todas las FKs y demuestra si existe un índice cuyo prefijo sirve para
  búsquedas/borrados del padre, incluidas filas inactivas o borradas lógicamente.
- Revisa como mínimo sesiones y recuperación por usuario, bloques/series por barbero,
  citas por servicio e historial por actor.
- No añadas índices por intuición: usa las consultas reales y `EXPLAIN (ANALYZE,
  BUFFERS)` con datos ficticios representativos.
- Añade checks temporales y de forma: revocación/emisión, eliminación/creación,
  outcomes de notificación, configuración de ventana/anticipación y hash/token.
- Evalúa índices parciales solo para consultas cuyo predicado coincide; no los cuentes
  como cobertura universal de una FK si excluyen filas históricas.

### Fase 8 — Pruebas y evidencia obligatorias

Ejecuta en PostgreSQL real y con al menos dos tenants:

1. migración desde cero;
2. upgrade desde las dos migraciones actuales con datos existentes;
3. validación Atlas, `atlas.sum` y chequeo de inmutabilidad;
4. catálogo de owners, roles, membresías, grants/default privileges y funciones;
5. RLS por cada tabla y operación con el rol API real: sin contexto, tenant A y B,
   lectura y escritura cruzadas, `FORCE RLS` y ausencia de `BYPASSRLS`;
6. rol worker: solo funciones autorizadas y sin lectura general cross-tenant;
7. carreras repetibles de citas, horarios, idempotencia, claim y retención;
8. append-only de historia e intentos;
9. anonimización completa y revocación de tokens;
10. planes de las consultas e índices nuevos;
11. límites de lote, timeouts, rollback y recuperación de lease.

No aceptes una prueba ejecutada como owner/superusuario como sustituto de una prueba
con `barberia_app`. No pongas correos, teléfonos, tokens ni secretos reales en fixtures
o logs.

### Documentación y entrega

- Actualiza el contrato API, diccionario/diagrama, matriz, decisiones, dudas,
  contradicciones, historial y README afectados.
- Cada tabla conserva propósito, propietario funcional, retención y clasificación.
- Documenta bootstrap de roles, rotación de credenciales, timeouts y recuperación.
- Mantén commits Conventional Commits y un PR por preocupación; no hagas push directo
  ni reescribas `main`.
- En la entrega enumera archivos, migraciones, decisiones, pruebas ejecutadas y
  limitaciones. Separa claramente “verificado” de “pendiente”.

No declares terminado mientras código, pruebas, DDL y documentación normativa se
contradigan.

## Fin del prompt
