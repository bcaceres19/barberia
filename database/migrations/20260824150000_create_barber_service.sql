-- Propósito
--   HU-023: crea `barber_service`, la asociación tenant-aware PURA entre un
--   barbero (`barber`) y un servicio (`service`) de la misma barbería: qué
--   servicios presta cada integrante del equipo. Sin columnas propias más
--   allá de las claves y `created_at`: nunca copia nombre, duración, precio
--   ni estado de ninguna de las dos tablas dueñas (CA-023-07).
--
-- Reglas y decisiones
--   RN-SER-03 (un servicio activo siempre reservable), RN-SER-04 (un cambio
--   de catálogo nunca reescribe una cita existente), RN-TEN-01 (ninguna
--   barbería accede a datos de otra), RN-DAT-02 (registros técnicos sin
--   datos personales), DEC-004/DEC-019/DEC-024/DEC-033-DEC-040 (mismo
--   estándar de esquema, RLS y roles que `barber`/`service`), DEC-068
--   (un servicio activo debe conservar al menos un barbero asignado:
--   retirar la última asignación activa se RECHAZA, verificado dentro de la
--   MISMA transacción que el DELETE, con bloqueo de fila para resistir la
--   carrera de dos desasignaciones concurrentes de las dos últimas filas de
--   un mismo servicio; la aplicación Go de esa regla vive en
--   internal/modules/catalog, esta migración solo prepara el esquema).
--
-- Relación con database/modelo-fisico-referencia.sql §B.4
--   Esta migración toma como base §B.4 (tabla de asociación en 2FN con FK
--   compuestas) con las mismas claves compuestas por `barbershop_id` que
--   `barber`/`service` ya establecieron en HU-021/HU-022
--   (`UNIQUE (barbershop_id, id)` en ambas), sin diferencias de fondo: la
--   propuesta de §B.4 ya coincidía con la decisión final (DEC-068 solo
--   resuelve la regla de negocio de "última asignación", no el esquema).
--
-- Qué NO hace esta migración
--   No agrega columnas de horario, disponibilidad, precio ni duración por
--   barbero (fuera de alcance de HU-023). No modifica `barber` ni `service`
--   (siguen siendo dueños exclusivos de su propio ciclo de vida). No agrega
--   `ON DELETE CASCADE` hacia `barber`/`service`: ninguna de las dos tablas
--   expone DELETE (RN-SER-03, CA-021-07), así que un borrado físico de esa
--   fila nunca puede ocurrir en la práctica; `ON DELETE RESTRICT` documenta
--   esa imposibilidad de forma explícita en vez de dejarla implícita.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: por membresía,
--   `barberia_migrator` la satisface sin que el objeto quede literalmente
--   owned por `barberia_owner` (mismo criterio documentado en
--   20260823130000_create_barber.sql y 20260824140000_create_service.sql).
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando
--   esta (DEC-036). No existe archivo `down`: el mecanismo normal es
--   roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260823130000_create_barber.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'service'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260824140000_create_service.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'barber_service'
  ) THEN
    RAISE EXCEPTION 'barber_service ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;

  -- barber_service necesita el mismo `UNIQUE (barbershop_id, id)` de las dos
  -- tablas dueñas para poder referenciarlas con FK compuesta por tenant.
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'barber_barbershop_id_id_uk'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de barber_barbershop_id_id_uk '
      '(20260823130000_create_barber.sql); esa restricción debe existir antes.';
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'service_barbershop_id_id_uk'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de service_barbershop_id_id_uk '
      '(20260824140000_create_service.sql); esa restricción debe existir antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla `barber_service`
-- ---------------------------------------------------------------------------

CREATE TABLE barber_service (
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  service_id    uuid        NOT NULL,
  created_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT barber_service_pk PRIMARY KEY (barbershop_id, barber_id, service_id),

  -- FK compuestas por barbershop_id: impiden asociar un barbero y un
  -- servicio de barberías distintas AUN con SQL directo bajo el rol de
  -- aplicación (CA-023-04, RN-TEN-01) -no solo por convención de la
  -- aplicación Go-, porque barber_id/service_id por sí solos no bastan para
  -- validar la pertenencia común al mismo tenant.
  --
  -- ON DELETE RESTRICT documenta que un borrado físico de `barber`/`service`
  -- nunca puede ocurrir en la práctica (ninguna de las dos tablas expone
  -- DELETE al rol de aplicación, RN-SER-03/CA-021-07): esta asociación no
  -- autoriza ni depende de un borrado en cascada que jamás sucederá.
  CONSTRAINT barber_service_barber_fk  FOREIGN KEY (barbershop_id, barber_id)
                                       REFERENCES barber (barbershop_id, id)
                                       ON DELETE RESTRICT,
  CONSTRAINT barber_service_service_fk FOREIGN KEY (barbershop_id, service_id)
                                       REFERENCES service (barbershop_id, id)
                                       ON DELETE RESTRICT
);

COMMENT ON TABLE barber_service IS
  'Asociación tenant-aware PURA entre barber y service (HU-023): qué servicios presta cada '
  'barbero. Propietario funcional: barbería, coordinado entre los módulos catalog (dueño de '
  'la operación) y staff (dueño de la existencia del barbero) mediante un puerto explícito, '
  'nunca acoplando el núcleo de un módulo al repositorio interno del otro. Sin columnas '
  'propias de negocio (CA-023-07): nunca nombre, duración, precio ni estado, que siguen '
  'siendo responsabilidad exclusiva de barber/service. DEC-068: un servicio ACTIVO debe '
  'conservar al menos una fila aquí; la aplicación (no esta tabla) rechaza la operación que '
  'lo dejaría en cero, dentro de una transacción que bloquea la fila de service para '
  'resistir la carrera de dos desasignaciones concurrentes.';

COMMENT ON COLUMN barber_service.created_at IS
  'Instante de la asignación. Sin updated_at: esta fila nunca se edita, solo se crea o se '
  'borra (RN-DAT-02); repetir la asignación no cambia este valor (ON CONFLICT DO NOTHING).';

-- ---------------------------------------------------------------------------
-- 3. Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

ALTER TABLE barber_service ENABLE ROW LEVEL SECURITY;
ALTER TABLE barber_service FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: FORCE ROW LEVEL SECURITY también alcanza al
-- propietario, así que barberia_owner necesita una política explícita para
-- cargar datos de referencia, semillas y correcciones auditadas
-- (estandar-base-datos.md §9.5, mismo patrón que barber/service).
CREATE POLICY barber_service_all_admin_policy ON barber_service
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY barber_service_select_tenant_policy ON barber_service
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barber_service_insert_tenant_policy ON barber_service
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- A diferencia de barber/service (sin DELETE, RN-SER-03/CA-021-07),
-- barber_service SÍ expone DELETE: retirar una asignación es el propio
-- propósito de esta tabla (DEC-068 la limita, no la prohíbe). Nunca borra
-- `barber` ni `service`: solo la fila de asociación.
CREATE POLICY barber_service_delete_tenant_policy ON barber_service
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni GRANT de UPDATE: esta fila no tiene ningún campo editable
-- (CA-023-07); cambiar una asignación es borrar e insertar de nuevo.

-- ---------------------------------------------------------------------------
-- 4. Índices
-- ---------------------------------------------------------------------------

-- La PK (barbershop_id, barber_id, service_id) ya sirve como índice para
-- "servicios de un barbero" (CA-023-01, orden por created_at además del id
-- para la paginación por cursor de la aplicación Go: se agrega un índice
-- adicional aparte porque la PK no ordena por created_at).
CREATE INDEX idx_barber_service_barber_created
  ON barber_service (barbershop_id, barber_id, created_at, service_id);

-- Sirve tanto "barberos que ofrecen un servicio" como, sobre todo, la
-- verificación de DEC-068 dentro de la transacción de desasignación: contar
-- cuántas filas activas quedan para un service_id dado, bajo el mismo
-- barbershop_id que ya bloqueó la fila de `service` (FOR UPDATE).
CREATE INDEX idx_barber_service_service
  ON barber_service (barbershop_id, service_id, barber_id);

-- ---------------------------------------------------------------------------
-- 5. Privilegios del rol de aplicación
-- ---------------------------------------------------------------------------

-- SELECT/INSERT/DELETE, nunca UPDATE (sin campo editable) ni el DELETE
-- amplio que barber/service prohíben (aquí SÍ se permite, acotado por RLS a
-- la propia barbería y por DEC-068 dentro de la aplicación Go).
GRANT SELECT, INSERT, DELETE ON TABLE barber_service TO barberia_app;
