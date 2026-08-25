-- Propósito
--   HU-040: crea `working_hour`, el tramo recurrente de la jornada laboral
--   de un barbero por día ISO de la semana. Varios tramos del mismo
--   barbero y día representan una jornada partida; un tramo cuyo
--   starts_time + duration_minutes cruza medianoche representa una
--   jornada nocturna (DEC-020). Base para el cálculo futuro de
--   disponibilidad (F-DISP-*, B3/B4); esta migración no calcula
--   disponibilidad ni crea citas.
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-DIS-05/RN-DIS-07
--   (horas civiles interpretadas con la zona IANA de la barbería, nunca la
--   del dispositivo), RN-DAT-02 (registros técnicos sin datos personales),
--   DEC-007/DEC-019/DEC-024/DEC-033-DEC-040 (mismo estándar de esquema, RLS
--   y roles que `barber`), DEC-020 (tramos nocturnos: se almacena inicio +
--   duración, nunca ends_time, para no ser ambiguo cuando el fin cae al día
--   siguiente), DEC-070 (CT-008: la FK hacia `barber` usa
--   ON DELETE RESTRICT, no CASCADE).
--
-- Relación con database/modelo-fisico-referencia.sql §C.2
--   Toma como base §C.2 con una diferencia deliberada: la FK hacia `barber`
--   usa ON DELETE RESTRICT en vez del ON DELETE CASCADE que proponía la
--   versión previa del modelo de referencia, por DEC-070 (CT-008). El resto
--   de la forma (columnas, CHECK de día/duración, UNIQUE de tramo exacto)
--   coincide.
--
-- Qué NO hace esta migración
--   No agrega una restricción de exclusión (EXCLUDE) para el solape entre
--   tramos del mismo barbero y día: un cálculo con envolvente semanal y
--   jornadas nocturnas la haría frágil (mismo criterio documentado en
--   modelo-fisico-referencia.sql §C.2). El solape se valida en el caso de
--   uso Go de internal/modules/schedule, dentro de la misma transacción que
--   el INSERT/UPDATE, con SELECT ... FOR UPDATE sobre los tramos existentes
--   del mismo (barbershop_id, barber_id, iso_weekday) para resistir la
--   carrera de dos altas concurrentes que se solaparían entre sí. La
--   invariante dura que sí vive en la base de datos -que dos citas no se
--   crucen- pertenece a `appointment` (B3), no a esta tabla.
--   No agrega excepciones por fecha, festivos, bloqueos, disponibilidad,
--   citas ni agenda (HU-041, HU-042, B3, B4).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: por membresía,
--   `barberia_migrator` la satisface sin que el objeto quede literalmente
--   owned por `barberia_owner` (mismo criterio documentado en
--   20260823130000_create_barber.sql).
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

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'working_hour'
  ) THEN
    RAISE EXCEPTION 'working_hour ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;

  -- working_hour necesita el mismo UNIQUE (barbershop_id, id) de barber
  -- para poder referenciarla con FK compuesta por tenant.
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'barber_barbershop_id_id_uk'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de barber_barbershop_id_id_uk '
      '(20260823130000_create_barber.sql); esa restricción debe existir antes.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_proc WHERE proname = 'set_updated_at'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de la función set_updated_at '
      '(20260807170000_create_tenant_foundation.sql); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla `working_hour`
-- ---------------------------------------------------------------------------

CREATE TABLE working_hour (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  barber_id        uuid        NOT NULL,
  iso_weekday      smallint    NOT NULL,
  starts_time      time        NOT NULL,
  duration_minutes integer     NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT working_hour_id_pk PRIMARY KEY (id),

  -- FK compuesta por barbershop_id: impide asociar un tramo a un barbero de
  -- otra barbería aun con SQL directo bajo el rol de aplicación (RN-TEN-01).
  -- ON DELETE RESTRICT, no CASCADE (DEC-070/CT-008): barber no tiene
  -- borrado físico en su alcance vigente (HU-021, DDL-BIZ-02); si una
  -- historia futura lo agrega, deberá decidir explícitamente qué pasa con
  -- el horario del barbero, en vez de perderlo por un efecto colateral de
  -- la FK.
  CONSTRAINT working_hour_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id)
    ON DELETE RESTRICT,

  -- ISO 8601: 1 = lunes … 7 = domingo. No se usa 0-6 para no heredar la
  -- ambigüedad de qué día es el cero.
  CONSTRAINT working_hour_iso_weekday_ck      CHECK (iso_weekday BETWEEN 1 AND 7),
  CONSTRAINT working_hour_duration_minutes_ck CHECK (duration_minutes BETWEEN 1 AND 1440),

  -- Un mismo barbero no repite el mismo inicio en el mismo día de semana
  -- (CA-040-04, parte de "solape"). El solape parcial entre dos tramos que
  -- no comparten starts_time se valida en la aplicación Go (ver "Qué NO
  -- hace esta migración").
  CONSTRAINT working_hour_shop_barber_weekday_start_uk
    UNIQUE (barbershop_id, barber_id, iso_weekday, starts_time)
);

COMMENT ON TABLE working_hour IS
  'Tramo recurrente de la jornada laboral de un barbero, por día ISO de la semana (HU-040, '
  'F-HOR-01). Propietario funcional: barbería. Retención: mientras exista el barbero '
  '(working_hour_barbershop_id_barber_id_fk usa ON DELETE RESTRICT, DEC-070). '
  'Clasificación: negocio. Varios tramos por barbero y día representan jornada partida.';

COMMENT ON COLUMN working_hour.starts_time IS
  'Hora civil de inicio del tramo, interpretada junto con barbershop.timezone; no se '
  'almacena un instante calculado (estandar-base-datos.md §5.5).';

COMMENT ON COLUMN working_hour.duration_minutes IS
  'Duración del tramo desde starts_time. Un valor que empuje el fin más allá de medianoche '
  'representa una jornada nocturna, permitida por DEC-020.';

-- ---------------------------------------------------------------------------
-- 3. Disparador de updated_at
-- ---------------------------------------------------------------------------

CREATE TRIGGER working_hour_set_updated_at
  BEFORE UPDATE ON working_hour
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 4. Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

ALTER TABLE working_hour ENABLE ROW LEVEL SECURITY;
ALTER TABLE working_hour FORCE  ROW LEVEL SECURITY;

CREATE POLICY working_hour_all_admin_policy ON working_hour
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY working_hour_select_tenant_policy ON working_hour
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_insert_tenant_policy ON working_hour
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_update_tenant_policy ON working_hour
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_delete_tenant_policy ON working_hour
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- ---------------------------------------------------------------------------
-- 5. Índices
-- ---------------------------------------------------------------------------

-- Sirve tanto la lista paginada por cursor (CA-040-01, orden estable por
-- iso_weekday, starts_time y luego id) como la verificación de solape
-- dentro de la transacción de alta/edición (SELECT ... FOR UPDATE acotado a
-- un barbero y día).
CREATE INDEX idx_working_hour_shop_barber_weekday_start
  ON working_hour (barbershop_id, barber_id, iso_weekday, starts_time, id);

-- ---------------------------------------------------------------------------
-- 6. Privilegios del rol de aplicación
-- ---------------------------------------------------------------------------

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE working_hour TO barberia_app;
