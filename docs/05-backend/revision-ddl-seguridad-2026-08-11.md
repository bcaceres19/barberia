---
titulo: "Revisión técnica del DDL y seguridad PostgreSQL"
version: "1.0"
estado: "Revisión técnica; pendiente de aprobación"
responsable: "Pendiente de aprobación del propietario"
ultima_actualizacion: "2026-08-11"
documentos_relacionados:
  - "estandar-base-datos.md"
  - "migraciones-atlas.md"
  - "../00-control/registro-decisiones.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/historias-usuario.md"
  - "../03-desarrollo/estrategia-pruebas.md"
  - "../10-backlog/prompt-endurecimiento-ddl.md"
---

# Revisión técnica del DDL y seguridad PostgreSQL

## 1. Resultado ejecutivo

El diseño tiene una base técnicamente buena: usa RLS habilitada y forzada, claves
foráneas tenant-aware, `timestamptz`, restricciones con nombre, comentarios de
clasificación y retención, una exclusión GiST para impedir cruces de citas y un
historial sin `UPDATE` ni `DELETE` para el rol de aplicación.

No se recomienda promover todavía el modelo completo a piloto o producción. Antes
de crear nuevas migraciones deben resolverse cuatro bloqueos principales:

1. el aprovisionamiento y la propiedad real de roles/objetos no son reproducibles;
2. las funciones `SECURITY DEFINER` y los privilegios del rol compartido amplían el
   impacto de una inyección SQL o de una credencial comprometida;
3. la reclamación de notificaciones no conserva una posesión durable del trabajo y
   puede producir envíos duplicados;
4. la anonimización no alcanza todos los datos personales conservados por citas,
   historial, tokens e idempotencia.

También hay errores de integridad en el modelo de referencia que deben corregirse
antes de convertirlo en migraciones: jornadas especiales contradictorias,
`ON DELETE SET NULL` sobre claves compuestas que incluyen `barbershop_id NOT NULL`,
actor cliente sin FK y relaciones de series que no prueban la forma del padre.

Esta revisión no modifica el DDL. Las dos migraciones versionadas se consideran
inmutables; cualquier corrección debe ser roll-forward mediante archivos nuevos.

## 2. Alcance y límites de la revisión

Se revisaron:

- `database/migrations/20260807170000_create_tenant_foundation.sql`;
- `database/migrations/20260807170100_create_idempotency_record.sql`;
- `database/modelo-fisico-referencia.sql`;
- `database/tests/hu001_aislamiento_rls.sql` y `database/testdata/`;
- los estándares, decisiones, reglas de negocio e historias afectadas.

`atlas migrate validate --dir file://migrations` terminó correctamente y
`atlas.sum` es consistente. No se pudo ejecutar la aplicación completa del esquema,
las pruebas con PostgreSQL real ni el lint semántico contra una base efímera porque
no había una URL de prueba configurada y el motor Docker no estaba disponible. Por
eso los hallazgos de ejecución deben confirmarse en PostgreSQL real antes de cerrar
los issues.

## 3. Fortalezas que deben conservarse

- RLS `ENABLE` + `FORCE` y políticas separadas por operación.
- El rol `barberia_app` se declara sin superusuario, DDL ni `BYPASSRLS`.
- Las relaciones de negocio principales incluyen `barbershop_id` en sus FKs.
- La cita usa snapshots y una exclusión por `(barbershop_id, barber_id, intervalo)`.
- Los intervalos son semiabiertos y los instantes usan `timestamptz`.
- Las tablas de historial e intentos son append-only para el rol de aplicación.
- Las migraciones declaran roll-forward y conservan checksum con Atlas.
- El modelo evita `jsonb` como contenedor genérico y mantiene una estructura cercana
  a tercera forma normal.

## 4. Hallazgos priorizados

### 4.1 Bloqueantes de seguridad, concurrencia o privacidad

| ID | Severidad | Hallazgo y evidencia | Riesgo | Corrección recomendada |
| --- | --- | --- | --- | --- |
| `DDL-INT-00` | Alta | El archivo de referencia crea `appointment` con una FK a `customer` antes de crear `customer` ([líneas 1044-1104](../../database/modelo-fisico-referencia.sql#L1044) frente a [1371-1399](../../database/modelo-fisico-referencia.sql#L1371)). | El archivo no es ejecutable en su orden actual y una migración copiada por secciones fallaría. | Crear primero `customer` o añadir la FK de cita en un `ALTER TABLE` posterior. Validar siempre el SQL derivado desde una base vacía. |
| `DDL-SEC-01` | Crítica | La primera migración afirma ejecutarse como `barberia_migrator`, pero ese mismo archivo crea el rol. Si se conecta un administrador para crearlo, el administrador queda como propietario de tablas y funciones; no hay `SET ROLE` ni `ALTER OWNER`. Véanse [líneas 12-20](../../database/migrations/20260807170000_create_tenant_foundation.sql#L12) y [57-69](../../database/migrations/20260807170000_create_tenant_foundation.sql#L57). | El rol migrador puede no controlar los objetos, los tests con `SET ROLE` pueden fallar y los privilegios efectivos dependen de cómo se ejecutó el primer despliegue. | Aprovisionar roles fuera del historial Atlas con un administrador; preferir un rol propietario `NOLOGIN`, un login migrador y roles separados de API/worker. Inventariar propietarios reales antes de emitir una migración correctiva de ownership y grants. |
| `DDL-SEC-02` | Crítica | Las funciones `SECURITY DEFINER` usan `SET search_path = public, pg_catalog` y nombres sin calificar; además, las funciones globales de worker se conceden a `barberia_app`. Ejemplos: [resolución de login](../../database/modelo-fisico-referencia.sql#L64), [reclamación de notificaciones](../../database/modelo-fisico-referencia.sql#L1762) y [retención](../../database/modelo-fisico-referencia.sql#L1869). | Una resolución inesperada de objetos o una credencial del API comprometida gana una vía cross-tenant. | Usar `SET search_path = ''` y nombres totalmente calificados, propietario controlado `NOLOGIN`, `REVOKE ... FROM PUBLIC`, validación estricta de argumentos y `EXECUTE` solo para el rol que lo necesita. Crear `barberia_worker` separado del API. |
| `DDL-CON-01` | Crítica | `notification_claim_due` hace `SELECT ... FOR UPDATE SKIP LOCKED`, pero no cambia estado ni crea lease/token de reclamación ([líneas 1762-1779](../../database/modelo-fisico-referencia.sql#L1762)). | Al confirmar la transacción la fila vuelve a estar disponible; si se mantiene abierta durante el envío se viola la regla de transacciones cortas. Ambos caminos permiten duplicados o bloqueos largos. | Reclamar con un `UPDATE ... SET status='processing', claim_token, claimed_at, lease_expires_at ... RETURNING`, confirmar, enviar fuera de la transacción y finalizar mediante CAS por `claim_token`. Recuperar leases vencidos y probar varios workers. |
| `DDL-PRI-01` | Crítica | `customer_anonymized_ck` solo exige borrar teléfono y correo; conserva `customer.full_name`. Además quedan `appointment.attendee_name`, `customer_note`, valores de historial y posibles respuestas de idempotencia ([líneas 1371-1404](../../database/modelo-fisico-referencia.sql#L1371), [1044-1079](../../database/modelo-fisico-referencia.sql#L1044) y [82-96](../../database/migrations/20260807170100_create_idempotency_record.sql#L82)). | Incumple `DEC-025`/`RN-DAT-03`: una supuesta anonimización sigue permitiendo identificar a una persona. | Aprobar primero una matriz de datos personales y una fecha ancla de retención. La operación debe anonimizar o redactar todas las copias, revocar tokens, tratar respuestas idempotentes y ser idempotente/transaccional. |
| `DDL-INT-01` | Alta | Los dos índices parciales de `working_hours_override` no impiden que un mismo día tenga simultáneamente una fila `is_closed=true` y segmentos `is_closed=false`; tampoco impiden solapes entre segmentos ([líneas 715-759](../../database/modelo-fisico-referencia.sql#L715)). | La disponibilidad puede considerar el mismo día cerrado y abierto o devolver horarios duplicados/inválidos. | Separar la cabecera de excepción diaria de sus segmentos, o imponer una exclusión/constraint trigger transaccional que garantice una única forma y ausencia de solapes. Aplicar también serialización a jornadas recurrentes concurrentes. |
| `DDL-INT-02` | Alta | `barber` y `time_block` declaran una FK `(barbershop_id, usuario_id) ON DELETE SET NULL` mientras `barbershop_id` es `NOT NULL` ([líneas 408-430](../../database/modelo-fisico-referencia.sql#L408) y [948-983](../../database/modelo-fisico-referencia.sql#L948)). | Al borrar el usuario PostgreSQL intenta poner en `NULL` ambas columnas y la operación falla. | Usar `RESTRICT` y un flujo explícito de desvinculación, o rediseñar la referencia. No habilitar borrado de barberos: HU-021 lo excluye. |
| `DDL-IDEM-01` | Alta | El protocolo dice que una segunda inserción concurrente observará `in_progress` y responderá `409`, pero la restricción única hace que el segundo `INSERT ... ON CONFLICT` espere la resolución de la primera transacción. También concede `DELETE` al API aunque un fallo dentro de la misma transacción revierte la fila ([líneas 8-19](../../database/migrations/20260807170100_create_idempotency_record.sql#L8) y [121-134](../../database/migrations/20260807170100_create_idempotency_record.sql#L121)). | Semántica documentada inalcanzable, espera no acotada y posibilidad de borrar una respuesta completada para repetir un efecto crítico. Una fila vencida sigue ocupando la clave única hasta ser reciclada o eliminada. | Elegir y documentar una de dos semánticas: espera acotada y replay, o bloqueo asesor transaccional no bloqueante con `409`. Revocar `DELETE` directo al API; una función estrecha debe reciclar atómicamente una clave vencida para cumplir CA-004-05 y el mantenimiento debe limpiar el resto. Validar el hash, considerar almacenar el digest de la clave y limitar el cuerpo almacenado. |
| `DDL-AUT-01` | Alta | `barberia_app` recibe `SELECT` directo sobre `staff_credential` y DML completo sobre `login_throttle` ([líneas 92-148](../../database/modelo-fisico-referencia.sql#L92) y [342-363](../../database/modelo-fisico-referencia.sql#L342)). | Una inyección SQL dentro del proceso puede extraer hashes de toda la barbería o reiniciar el control de fuerza bruta. | Exponer operaciones estrechas mediante funciones revisadas, separar privilegios de autenticación y evitar DML directo de handlers generales. Para IP usar HMAC con secreto de despliegue, no hash simple de un espacio de baja entropía. |
| `DDL-HIS-01` | Alta | `appointment_history.actor_customer_id` no referencia `customer`, y el rol del API puede actualizar una cita sin crear historial o insertar historial arbitrario ([líneas 1227-1310](../../database/modelo-fisico-referencia.sql#L1227)). | Actores huérfanos e incumplimiento de la invariante de auditoría atómica. | Añadir FK tenant-aware e índice. Encapsular transiciones críticas en una función/comando transaccional o imponer una garantía equivalente que actualice cita, historial, cambios y notificaciones de forma atómica. |

### 4.2 Integridad y consistencia funcional

| ID | Severidad | Hallazgo | Corrección recomendada |
| --- | --- | --- | --- |
| `DDL-INT-03` | Alta | `time_block_series_date` no prueba que el padre sea `date_list`, ni que la fecha pertenezca al rango efectivo. | Dividir modelos por tipo o usar un constraint trigger pequeño, documentado y probado. |
| `DDL-INT-04` | Media | La cita no garantiza en BD que `ends_at - starts_at` corresponda a `duration_minutes_snapshot`. | Evaluar un `CHECK` basado en la diferencia de instantes, independiente de la zona de sesión, y probar DST/cruce de medianoche. No copiar la expresión sin validarla en la versión PostgreSQL soportada. |
| `DDL-INT-05` | Alta | Cambiar `staff_user.phone` puede conservar el antiguo `phone_verified_at`. | Centralizar el cambio de teléfono y poner la verificación en `NULL` cuando cambie el valor, con prueba de integración. |
| `DDL-BIZ-01` | Alta | `DEC-018` exige un recordatorio inicial a 30 minutos, pero cero filas significa ninguno y no existe backfill ni provisión por defecto ([líneas 1621-1644](../../database/modelo-fisico-referencia.sql#L1621)). | Insertar la regla ordinal 1/30 al crear una barbería y migrar barberías existentes; mantener 0 como elección explícita posterior. |
| `DDL-BIZ-02` | Alta | El modelo de `barber` incluye borrado, desactivación, orden y vínculo automático con `staff_user`; HU-021 solo autoriza alta, lectura y cambio de nombre y excluye ese ciclo de vida. | Retirar del primer DDL de HU-021 lo no aprobado o registrar una decisión antes de incluirlo. No conceder `DELETE`. |
| `DDL-BIZ-03` | Media | La unicidad de `customer` por teléfono presupone que un teléfono identifica permanentemente a una persona que reserva. | Registrar una decisión de identidad/reutilización antes de implementar el upsert; contemplar teléfonos compartidos o reasignados. |
| `DDL-NOR-01` | Media | `notification_attempt.channel` repite `notification_schedule.channel` y puede discrepar. | Eliminar la columna duplicada o garantizar consistencia mediante una clave compuesta; preferir derivarla del padre. |
| `DDL-VAL-01` | Media | Hashes y tokens se validan solo por longitud; correo/nombres no siempre exigen forma canónica con `btrim`; varios textos sensibles no tienen cota. | Preferir `bytea` para hashes binarios o regex hexadecimal minúscula; fijar límites y normalización; acotar `response_body`, `reason`, valores históricos e identificadores de proveedor según contrato real. |
| `DDL-TMP-01` | Media | Faltan varias relaciones temporales: revocación posterior a emisión, eliminación posterior a creación y outcomes de notificación mutuamente consistentes. | Añadir `CHECK` donde la regla sea local y estable; cubrir cada estado permitido y rechazado. |
| `DDL-CFG-01` | Media | `min_lead_minutes` puede superar toda la ventana pública expresada en días. | Añadir una restricción cruzada coherente con la semántica aprobada o documentar explícitamente si se permite una barbería sin disponibilidad pública. |
| `DDL-NAM-01` | Baja | `working_hours` y `working_hours_override` incumplen la convención de tablas en singular. | Renombrarlas antes de que existan migraciones aplicadas, usando nombres de dominio claros. |
| `DDL-DOC-01` | Baja | Algunas tablas auxiliares no documentan de forma completa propietario funcional, retención y clasificación, como exige el estándar. | Completar los comentarios al convertir cada sección en migración y mantener el diccionario/diagrama de datos sincronizado. |

### 4.3 Privilegios, rendimiento y pruebas

| ID | Severidad | Hallazgo | Corrección recomendada |
| --- | --- | --- | --- |
| `DDL-SEC-03` | Alta | No hay baseline de `ALTER DEFAULT PRIVILEGES`; el trigger `set_updated_at()` concede `EXECUTE` al API sin necesidad demostrada. | Revocar privilegios públicos por defecto sobre funciones del propietario, conceder solo objetos explícitos y retirar grants innecesarios. Validar en catálogos, no solo leyendo SQL. |
| `DDL-SEC-04` | Alta | API y worker comparten rol, aunque el worker requiere alcance global y el API debe permanecer tenant-scoped. | Separar roles y credenciales; el worker solo ejecuta funciones de claim/finalización y no recibe acceso general cross-tenant. |
| `DDL-PER-01` | Media | Faltan índices completos en lados referenciantes que deben buscar también filas históricas: sesiones y códigos por usuario, bloques/series borrados lógicamente por barbero, citas por servicio e historial por actor. | Confirmar con consultas reales y `EXPLAIN`; añadir los índices mínimos cuyo prefijo coincide con la FK/consulta. No crear índices duplicados. |
| `DDL-RLS-01` | Alta | La prueba actual verifica dos tablas y parte de las operaciones, pero el modelo completo necesitará probar cada política y cada rol; el caso sin contexto depende de un tipo de error específico. | Suite generada/inventariada por tabla: sin contexto, tenant A/B, cada CRUD permitido/denegado, `FORCE RLS`, propietarios, `BYPASSRLS`, funciones definer y roles API/worker. Aceptar cualquier fallo fail-closed esperado, no un único SQLSTATE accidental. |
| `DDL-CON-02` | Alta | No hay pruebas concurrentes ejecutables del DDL de jornadas, reclamos, idempotencia, exclusión ni retención. | PostgreSQL real, al menos dos conexiones y dos tenants; repetir carreras y comprobar exactamente una reclamación/efecto. Nunca simular estas garantías con mocks. |
| `DDL-OPS-01` | Media | `p_limit` de funciones globales no se valida y los timeouts no están definidos por rol/operación. | Rechazar `NULL`, cero, negativos y lotes excesivos; usar lotes pequeños. Definir `lock_timeout`, `statement_timeout` e `idle_in_transaction_session_timeout` apropiados en despliegue y probar recuperación. |
| `DDL-PER-02` | Baja | Las políticas repiten `current_setting(...)`; el valor es constante en la transacción, pero no se ha medido el plan en tablas grandes. | Conservar el fail-closed actual y medir. Si el plan lo justifica, evaluar una función estable segura o la forma escalar que PostgreSQL ejecute una vez, sin relajar RLS. |

## 5. Decisiones que deben registrarse antes de codificar

No deben resolverse dentro de una migración ni por preferencia del implementador:

1. versión mayor mínima y ventana de soporte de PostgreSQL;
2. vocabulario almacenado de `appointment_history.event_type`, hoy contradictorio
   entre dos secciones de `estados-citas.md`;
3. fecha ancla de los 24 meses: creación del cliente, última cita, última
   interacción u otra base aprobada;
4. matriz exacta de anonimización para nombre del cliente, persona atendida,
   notas, historial, tokens y respuestas idempotentes;
5. reutilización/identidad de clientes por teléfono;
6. unicidad global del correo si una persona puede pertenecer a varias barberías;
7. ciclo de vida de `barber` y relación opcional con `staff_user`;
8. valores/orden del segundo y tercer recordatorio;
9. respuesta concurrente de idempotencia: esperar y reproducir o responder que la
   operación sigue en curso;
10. adopción del modelo de roles con propietario `NOLOGIN`, API y worker separados.

Las contradicciones ya visibles en el propio DDL de referencia deben trasladarse a
`docs/00-control/contradicciones.md` o `dudas-pendientes.md` antes de crear código.

## 6. Orden recomendado de ejecución

1. **Preflight y decisiones:** abrir issues separados, confirmar entorno/versiones,
   registrar dudas/contradicciones y obtener aprobación.
2. **Roles y superficie de privilegios:** resolver bootstrap, ownership, default
   privileges y endurecimiento de funciones; hacerlo antes de agregar tablas.
3. **Migración correctiva de HU-001/HU-004:** privilegios mínimos, idempotencia y
   pruebas de upgrade; no editar las migraciones existentes.
4. **Corregir el modelo de referencia:** integridad de horarios, borrados, claves,
   normalización y alcance real de HU-020/HU-021.
5. **Diseñar workers y privacidad:** leases de notificación y anonimización completa
   con roles separados.
6. **Promover por bloques:** una migración/issue pequeño por preocupación, con
   ejecución desde vacío y desde el último estado aplicado.

El prompt autocontenido para realizar este trabajo está en
[prompt-endurecimiento-ddl.md](../10-backlog/prompt-endurecimiento-ddl.md).
