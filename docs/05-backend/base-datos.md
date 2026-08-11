---
titulo: "PostgreSQL, multi-tenancy y consultas"
version: "1.2"
estado: "Decisión confirmada"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-06"
documentos_relacionados:
  - "../00-control/registro-decisiones.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
  - "../04-arquitectura/stack-despliegue-operacion.md"
  - "estandar-base-datos.md"
  - "migraciones-atlas.md"
---

# PostgreSQL, multi-tenancy y consultas

## 1. Decisiones estructurales

- motor: PostgreSQL;
- modelo multi-tenant: esquema compartido;
- unidad de aislamiento: barbería;
- `barbershop_id` obligatorio en toda tabla de negocio;
- seguridad a nivel de fila (RLS) en tablas con datos de una barbería;
- `barbershop` y `barber` son entidades distintas;
- la exclusión de turnos cruzados opera por `barber_id`;
- intervalos temporales semiabiertos `[starts_at, ends_at)`;
- instantes almacenados con zona horaria;
- retención personal de 24 meses por defecto, configurable y seguida de anonimización.

## 2. Patrón obligatorio de RLS

Cada solicitud autenticada abre una transacción y fija el identificador de barbería con alcance local a esa transacción. Las políticas comparan la fila contra ese contexto. No se reutiliza una conexión con contexto de otra barbería fuera de una transacción.

Patrón conceptual:

```sql
ALTER TABLE appointment ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment FORCE ROW LEVEL SECURITY;

CREATE POLICY appointment_tenant_policy ON appointment
USING (barbershop_id = current_setting('app.barbershop_id')::uuid)
WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
```

El rol normal de la aplicación no es propietario de las tablas ni posee `BYPASSRLS`. Las tareas administrativas usan un rol separado, auditable y no accesible desde peticiones ordinarias.

## 3. Integridad temporal

PostgreSQL debe impedir en base de datos que dos citas que ocupan agenda se crucen para el mismo barbero. La implementación usa un rango `tstzrange(starts_at, ends_at, '[)')` y una restricción de exclusión parcial sobre estados que ocupan agenda.

Patrón conceptual:

```sql
CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE appointment
ADD CONSTRAINT appointment_no_overlap
EXCLUDE USING gist (
  barbershop_id WITH =,
  barber_id WITH =,
  tstzrange(starts_at, ends_at, '[)') WITH &&
)
WHERE (status IN ('confirmed', 'completed', 'no_show'));
```

Los bloqueos no comparten esa restricción con las citas: un bloqueo urgente puede solaparse con citas existentes para abrir el flujo asistido definido por `RN-BLQ-03`.

Validaciones mínimas:

- `ends_at > starts_at`;
- barbero y cita pertenecen a la misma barbería;
- servicio, horario y bloqueos consultados pertenecen al tenant vigente;
- una cita que cruza medianoche cabe por completo en un tramo laboral configurado;
- una cita cancelada no entra en la restricción de exclusión.

## 4. Buenas prácticas de consultas

### Selección y volumen

1. Seleccionar columnas explícitas; no usar `SELECT *` en código de aplicación.
2. Paginar listas y poner límite incluso en pantallas que hoy tienen pocos registros.
3. Consultar por rango de fecha/hora, no aplicar funciones a la columna indexada en el filtro.
4. Evitar N+1: traer relaciones necesarias en una consulta o en lotes acotados.
5. No cargar historial, intentos de envío o contenido de mensajes en la vista diaria si no se muestran.

### Aislamiento

6. Toda operación ocurre dentro de una transacción con `app.barbershop_id` fijado.
7. Repetir `barbershop_id` en claves únicas e índices relevantes; RLS protege, pero el índice debe permitir filtrar barato.
8. Probar cada consulta y endpoint con dos barberías y recursos de identificadores conocidos.
9. Responder “no encontrado” ante recursos de otro tenant; no revelar su existencia.
10. Prohibir utilidades genéricas que acepten omitir el contexto de barbería.

### Índices

11. Crear índices desde consultas reales, no por intuición ni uno por cada columna.
12. Índices iniciales esperados:

   - `appointment (barbershop_id, barber_id, starts_at)`;
   - `appointment (barbershop_id, customer_id, starts_at DESC)`;
   - `time_block (barbershop_id, barber_id, starts_at)` donde no esté eliminado;
   - `notification_schedule (status, scheduled_for)` para trabajos pendientes;
   - índices únicos tenant-aware para enlaces, correos o claves que deban ser únicas dentro de la barbería.

13. Usar índices parciales para estados activos o filas no eliminadas cuando la consulta siempre filtra así.
14. Revisar duplicación y costo de escritura antes de agregar un índice nuevo.

### Planes y observabilidad

15. Revisar con `EXPLAIN (ANALYZE, BUFFERS)` las consultas críticas usando datos de volumen representativo.
16. Registrar duración, filas y nombre lógico de la operación; nunca nombres, teléfonos, correos ni SQL con valores sensibles.
17. Definir alerta para consultas lentas y revisar primero frecuencia total, no solo el peor caso aislado.
18. Mantener estadísticas y autovacuum activos; no “optimizar” deshabilitándolos.

### Escrituras y concurrencia

19. Mantener transacciones cortas y no hacer llamadas de red mientras una transacción está abierta.
20. Usar restricciones y claves idempotentes como última defensa, no una comprobación `SELECT` seguida de `INSERT` sin garantía.
21. Para trabajadores concurrentes, reclamar lotes pequeños con bloqueo de filas y `SKIP LOCKED`.
22. Insertar evento/programación de notificación en la misma transacción del cambio de negocio.
23. Evitar actualizaciones masivas sin filtro tenant, límite operativo y revisión previa de filas afectadas.

## 5. Revisión mínima antes de aprobar una consulta

- [ ] Selecciona solo las columnas necesarias.
- [ ] Tiene contexto de barbería y queda protegida por RLS.
- [ ] Su filtro principal puede usar un índice.
- [ ] Tiene límite o devuelve por diseño una sola fila.
- [ ] No introduce N+1.
- [ ] No expone datos personales en registros.
- [ ] Se probó con otra barbería y con concurrencia cuando modifica agenda.
- [ ] Su plan se revisó si participa en disponibilidad, agenda diaria o envío de recordatorios.

## 6. Retención y anonimización

Un parámetro de configuración conserva el plazo efectivo por barbería, con 24 meses como valor inicial. El proceso de anonimización:

1. identifica datos vencidos;
2. sustituye nombre, teléfono y correo por valores anónimos no reversibles;
3. revoca tokens públicos y credenciales de acceso a la cita;
4. conserva servicio, barbero, intervalo, estado, importes informativos e historial operativo;
5. registra el proceso sin copiar el dato personal anterior a los logs.

Un cambio de retención requiere motivo, actor, fecha y revisión de su base legal.

## 7. Estándar de implementación

La normalización, nomenclatura, tipos, integridad tenant-aware, migraciones y pruebas obligatorias se rigen por [estandar-base-datos.md](estandar-base-datos.md). Si una optimización exige desnormalizar, debe cumplir el proceso de excepción definido allí y conservar las decisiones de este documento.

Atlas CLI administra y rastrea las migraciones SQL versionadas según [migraciones-atlas.md](migraciones-atlas.md). Ningún proceso de la aplicación ejecuta DDL al arrancar.
