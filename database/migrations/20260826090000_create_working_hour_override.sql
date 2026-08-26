-- Propósito
--   HU-041: agrega `barber.holiday_calendar_enabled` (interruptor de
--   calendario colombiano de festivos, independiente por barbero) y crea
--   `working_hour_override`/`working_hour_override_segment`: la excepción
--   de jornada para una fecha concreta (día cerrado, o abierto con uno o
--   varios tramos propios). Prevalece sobre el bloqueo automático de
--   festivo y sobre `working_hour` para esa fecha (RN-BLQ-02).
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-BLQ-02
--   (festivos activables/bloqueables por barbero, excepción manual
--   prevalece), RN-DIS-05/RN-DIS-07 (horas civiles interpretadas con la
--   zona IANA de la barbería), RN-DAT-02 (registros técnicos sin datos
--   personales), DEC-007/DEC-019/DEC-024/DEC-033-DEC-040 (mismo estándar
--   de esquema, RLS y roles que `working_hour`), DEC-020 (tramos
--   nocturnos: inicio + duración, nunca ends_time), DEC-070 (CT-008: las
--   FK hacia `barber`/`working_hour_override` usan ON DELETE RESTRICT,
--   no CASCADE).
--
-- Relación con database/modelo-fisico-referencia.sql §C.1/§C.3/§C.3b
--   Copia literal de las tres secciones: ya incorporan DEC-070 (RESTRICT)
--   desde que se resolvió CT-008, así que no hay diferencias de fondo
--   entre el modelo de referencia y esta migración.
--
-- Qué NO hace esta migración
--   No agrega el cálculo del calendario colombiano de festivos en sí
--   (Ley 51 de 1983, "Ley Emiliani"): esa resolución es determinista y
--   vive en Go (internal/modules/schedule), sin tabla ni dependencia
--   nueva -no existe una lista de festivos que persistir ni mantener-.
--   No agrega bloqueos de descanso/almuerzo/vacaciones/emergencia
--   (HU-042), citas, disponibilidad pública ni agenda.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: por membresía,
--   `barberia_migrator` la satisface sin que el objeto quede literalmente
--   owned por `barberia_owner` (mismo criterio documentado en
--   20260825160000_create_working_hour.sql).
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
    WHERE table_schema = 'public' AND table_name = 'working_hour'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260825160000_create_working_hour.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber'
      AND column_name = 'holiday_calendar_enabled'
  ) THEN
    RAISE EXCEPTION 'barber.holiday_calendar_enabled ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'working_hour_override'
  ) THEN
    RAISE EXCEPTION 'working_hour_override ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. `barber.holiday_calendar_enabled`
-- ---------------------------------------------------------------------------

ALTER TABLE barber
  ADD COLUMN holiday_calendar_enabled boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN barber.holiday_calendar_enabled IS
  'Activa el calendario colombiano de festivos para ESTE barbero (RN-BLQ-02). '
  'La decisión de un barbero no altera el calendario de otro de la misma barbería.';

-- ---------------------------------------------------------------------------
-- 3. `working_hour_override` — cabecera de excepción por fecha
-- ---------------------------------------------------------------------------

CREATE TABLE working_hour_override (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  barber_id      uuid        NOT NULL,
  effective_date date        NOT NULL,
  is_closed      boolean     NOT NULL,
  reason         text,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT working_hour_override_id_pk               PRIMARY KEY (id),
  CONSTRAINT working_hour_override_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- RESTRICT, no CASCADE (CT-008/DEC-070): mismo criterio que working_hour.
  CONSTRAINT working_hour_override_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,

  -- Una sola cabecera por barbero y fecha (CA-041-05): nunca admite dos
  -- filas cerradas ni una cerrada y otra abierta para la misma fecha.
  CONSTRAINT working_hour_override_shop_barber_date_uk
    UNIQUE (barbershop_id, barber_id, effective_date),

  CONSTRAINT working_hour_override_reason_ck CHECK (
    reason IS NULL OR char_length(reason) <= 200
  )
);

COMMENT ON TABLE working_hour_override IS
  'Cabecera de excepción de jornada para una fecha concreta (RN-BLQ-02, HU-041): is_closed '
  'decide si el día está cerrado o tiene tramos propios en working_hour_override_segment. '
  'Propietario funcional: barbería. Retención: mientras exista el barbero. Clasificación: negocio.';

CREATE INDEX idx_working_hour_override_shop_barber_date
  ON working_hour_override (barbershop_id, barber_id, effective_date);

CREATE TRIGGER working_hour_override_set_updated_at
  BEFORE UPDATE ON working_hour_override
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE working_hour_override ENABLE ROW LEVEL SECURITY;
ALTER TABLE working_hour_override FORCE  ROW LEVEL SECURITY;

CREATE POLICY working_hour_override_all_admin_policy ON working_hour_override
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY working_hour_override_select_tenant_policy ON working_hour_override
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_override_insert_tenant_policy ON working_hour_override
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_override_update_tenant_policy ON working_hour_override
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY working_hour_override_delete_tenant_policy ON working_hour_override
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE working_hour_override TO barberia_app;

-- ---------------------------------------------------------------------------
-- 4. `working_hour_override_segment` — tramos de una cabecera abierta
-- ---------------------------------------------------------------------------
-- Solo tiene sentido bajo una cabecera con is_closed = false; el
-- disparador de más abajo lo exige porque ninguna restricción declarativa
-- puede mirar otra tabla. A diferencia de working_hour (tramo semanal
-- recurrente, sin restricción de exclusión por la fragilidad de una
-- envolvente semanal con tramos nocturnos), esta tabla SÍ usa una
-- restricción de exclusión real: la fecha vive fija en la cabecera, así
-- que comparar tramos con un ancla común (2000-01-01) es seguro y no
-- ambiguo, incluido un tramo nocturno que cruza medianoche.

CREATE TABLE working_hour_override_segment (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  override_id      uuid        NOT NULL,
  starts_time      time        NOT NULL,
  duration_minutes integer     NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT working_hour_override_segment_id_pk PRIMARY KEY (id),

  -- RESTRICT, no CASCADE (CT-008/DEC-070): un tramo se retira
  -- explícitamente antes de poder retirar su cabecera; nunca por efecto
  -- colateral de borrar la cabecera.
  CONSTRAINT working_hour_override_segment_barbershop_id_override_id_fk
    FOREIGN KEY (barbershop_id, override_id)
    REFERENCES working_hour_override (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT working_hour_override_segment_duration_minutes_ck CHECK (
    duration_minutes BETWEEN 1 AND 1440
  ),
  CONSTRAINT working_hour_override_segment_override_start_uk
    UNIQUE (override_id, starts_time)
);

COMMENT ON TABLE working_hour_override_segment IS
  'Tramo horario de una cabecera de excepción abierta (RN-BLQ-02, HU-041). Clasificación: negocio.';

-- Ningún tramo se solapa con otro de la MISMA cabecera. Semántica
-- semiabierta [inicio, fin), igual que working_hour_shop_barber_weekday
-- y appointment (sección D.1).
ALTER TABLE working_hour_override_segment
  ADD CONSTRAINT working_hour_override_segment_no_overlap_excl
  EXCLUDE USING gist (
    override_id WITH =,
    tsrange(
      date '2000-01-01' + starts_time,
      date '2000-01-01' + starts_time + make_interval(mins => duration_minutes),
      '[)'
    ) WITH &&
  );

CREATE OR REPLACE FUNCTION working_hour_override_segment_check_open()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
DECLARE
  v_is_closed boolean;
BEGIN
  SELECT is_closed INTO v_is_closed
  FROM public.working_hour_override
  WHERE id = NEW.override_id AND barbershop_id = NEW.barbershop_id;

  IF v_is_closed IS NULL THEN
    RAISE EXCEPTION 'working_hour_override % no existe en la barbería %',
      NEW.override_id, NEW.barbershop_id;
  END IF;

  IF v_is_closed THEN
    RAISE EXCEPTION 'working_hour_override % está cerrada (is_closed=true): no admite tramos',
      NEW.override_id;
  END IF;

  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION working_hour_override_segment_check_open() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION working_hour_override_segment_check_open() TO barberia_app;

COMMENT ON FUNCTION working_hour_override_segment_check_open() IS
  'Impide insertar/mover un tramo bajo una cabecera cerrada (HU-041). La FK ya garantiza que la '
  'cabecera existe en la misma barbería; esto añade la condición is_closed = false.';

CREATE TRIGGER working_hour_override_segment_check_open_trg
  BEFORE INSERT OR UPDATE ON working_hour_override_segment
  FOR EACH ROW EXECUTE FUNCTION working_hour_override_segment_check_open();

ALTER TABLE working_hour_override_segment ENABLE ROW LEVEL SECURITY;
ALTER TABLE working_hour_override_segment FORCE  ROW LEVEL SECURITY;

CREATE POLICY working_hour_override_segment_all_admin_policy ON working_hour_override_segment
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY working_hour_override_segment_select_tenant_policy ON working_hour_override_segment
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY working_hour_override_segment_insert_tenant_policy ON working_hour_override_segment
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY working_hour_override_segment_update_tenant_policy ON working_hour_override_segment
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY working_hour_override_segment_delete_tenant_policy ON working_hour_override_segment
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE working_hour_override_segment TO barberia_app;
