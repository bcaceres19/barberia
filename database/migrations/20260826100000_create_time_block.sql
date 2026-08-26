-- Propósito
--   HU-042: bloqueos de agenda del barbero. Crea `time_block` (bloqueo
--   puntual, incluida la emergencia), `time_block_series` (definición
--   recurrente semanal o lista explícita de fechas), `time_block_series_date`
--   (fechas explícitas de una serie `date_list`) y `time_block_series_exception`
--   (instancias suprimidas de una serie, "esta instancia no").
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-BLQ-01 (siete
--   tipos, punto/recurrencia/lista con excepciones), RN-BLQ-03 (un bloqueo
--   nunca falla por chocar con citas; sin restricción de exclusión contra
--   `appointment`), RN-BLQ-04 (eliminación lógica, nunca DELETE físico),
--   RN-DIS-05/RN-DIS-07 (horas civiles/zona IANA de la barbería), RN-IDE-01
--   (semántica segura ante reintentos, vive en el servicio Go, no aquí),
--   DEC-007/DEC-008/DEC-009/DEC-013/DEC-019/DEC-020/DEC-024/DEC-033-DEC-040
--   (mismo estándar de esquema, RLS y roles que `working_hour`/
--   `working_hour_override`), DEC-043 (protocolo de idempotencia), DEC-070
--   (CT-008: las FK hacia `barber`/`staff_user`/`time_block_series` usan
--   ON DELETE RESTRICT, no CASCADE).
--
-- Relación con database/modelo-fisico-referencia.sql §C.4/§C.5
--   Copia literal de ambas secciones: ya incorporan DEC-070 (RESTRICT) desde
--   que se resolvió CT-008, así que no hay diferencias de fondo entre el
--   modelo de referencia y esta migración.
--
-- Qué NO hace esta migración
--   No crea, modifica, reprograma, cancela ni notifica citas: `appointment`
--   pertenece a B3. No calcula la lista de citas afectadas por un bloqueo
--   (RN-BLQ-03): esa consulta la implementa B3/B5 sin que `time_block` la
--   conozca. No agrega disponibilidad pública ni reserva.
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
    WHERE table_schema = 'public' AND table_name = 'working_hour_override'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260826090000_create_working_hour_override.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'time_block_series'
  ) THEN
    RAISE EXCEPTION 'time_block_series ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. `time_block_series` — bloqueos recurrentes y listas de fechas
-- ---------------------------------------------------------------------------
-- DEC-020 exige tres formas: instancia puntual, recurrencia y lista explícita
-- de fechas, con excepciones individuales. AGENTS.md prohíbe representar
-- horarios en `jsonb`, así que la recurrencia se modela como entidad con dos
-- tablas hijas.
--
-- La disponibilidad expande las series en la consulta del rango pedido; NO se
-- materializan instancias en `time_block`. Materializarlas obligaría a
-- mantener un horizonte y a repararlo tras cada cambio de configuración.

CREATE TABLE time_block_series (
  id               uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id    uuid        NOT NULL,
  barber_id        uuid        NOT NULL,
  block_type       text        NOT NULL,
  recurrence_kind  text        NOT NULL,
  iso_weekday      smallint,
  starts_time      time        NOT NULL,
  duration_minutes integer     NOT NULL,
  effective_from   date        NOT NULL,
  effective_until  date,
  reason           text,
  deleted_at       timestamptz,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT time_block_series_id_pk               PRIMARY KEY (id),
  CONSTRAINT time_block_series_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- RESTRICT, no CASCADE (CT-008/DEC-070): mismo criterio que working_hour.
  CONSTRAINT time_block_series_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,

  -- Los siete tipos del criterio de salida de B2 (RN-BLQ-01).
  CONSTRAINT time_block_series_block_type_ck CHECK (
    block_type IN ('break', 'lunch', 'unavailable', 'day_off', 'holiday', 'vacation', 'emergency')
  ),
  CONSTRAINT time_block_series_recurrence_kind_ck CHECK (
    recurrence_kind IN ('weekly', 'date_list')
  ),

  -- Una recurrencia semanal necesita día de semana; una lista de fechas no.
  CONSTRAINT time_block_series_weekday_shape_ck CHECK (
    (recurrence_kind = 'weekly'    AND iso_weekday IS NOT NULL AND iso_weekday BETWEEN 1 AND 7)
    OR
    (recurrence_kind = 'date_list' AND iso_weekday IS NULL)
  ),
  CONSTRAINT time_block_series_duration_minutes_ck CHECK (duration_minutes BETWEEN 1 AND 1440),
  CONSTRAINT time_block_series_effective_range_ck  CHECK (
    effective_until IS NULL OR effective_until >= effective_from
  ),
  CONSTRAINT time_block_series_reason_ck CHECK (reason IS NULL OR char_length(reason) <= 200),

  -- DDL-TMP-01: la eliminación lógica no puede preceder a la creación.
  CONSTRAINT time_block_series_deleted_at_ck CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

COMMENT ON TABLE time_block_series IS
  'Definición de un bloqueo recurrente o por lista de fechas (RN-BLQ-01, DEC-020). '
  'Propietario funcional: barbería. Retención: permanente; eliminación lógica (RN-BLQ-04). '
  'Clasificación: negocio. Nunca se materializa en time_block: la disponibilidad la expande.';

-- DDL-PER-01: sin predicado parcial. Un `WHERE deleted_at IS NULL` aparte
-- sería un segundo índice con el mismo prefijo (duplicado prohibido por el
-- hallazgo); este mismo índice completo sirve tanto la agenda vigente como
-- la auditoría de series borradas lógicamente por barbero.
CREATE INDEX idx_time_block_series_shop_barber
  ON time_block_series (barbershop_id, barber_id, effective_from);

CREATE TRIGGER time_block_series_set_updated_at
  BEFORE UPDATE ON time_block_series
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE time_block_series ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block_series FORCE  ROW LEVEL SECURITY;

CREATE POLICY time_block_series_all_admin_policy ON time_block_series
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY time_block_series_select_tenant_policy ON time_block_series
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_series_insert_tenant_policy ON time_block_series
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_series_update_tenant_policy ON time_block_series
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: RN-BLQ-04 exige eliminación lógica.
GRANT SELECT, INSERT, UPDATE ON TABLE time_block_series TO barberia_app;

-- ---------------------------------------------------------------------------
-- 3. `time_block_series_date` — fechas explícitas de una serie `date_list`
-- ---------------------------------------------------------------------------

CREATE TABLE time_block_series_date (
  barbershop_id uuid NOT NULL,
  series_id     uuid NOT NULL,
  block_date    date NOT NULL,

  CONSTRAINT time_block_series_date_pk PRIMARY KEY (barbershop_id, series_id, block_date),
  -- RESTRICT, no CASCADE (CT-008/DEC-070): time_block_series usa eliminación
  -- lógica (deleted_at, RN-BLQ-04); sus hijas nunca se borran por cascada.
  CONSTRAINT time_block_series_date_barbershop_id_series_id_fk
    FOREIGN KEY (barbershop_id, series_id)
    REFERENCES time_block_series (barbershop_id, id) ON DELETE RESTRICT
);

COMMENT ON TABLE time_block_series_date IS
  'Fechas explícitas de una serie de tipo date_list (por ejemplo, vacaciones del 15 al 30). '
  'Clasificación: negocio.';

-- DDL-INT-03: ninguna restricción declarativa puede comprobar a la vez que el
-- padre es de tipo date_list y que la fecha cae dentro de su rango efectivo
-- (effective_from/effective_until). Disparador pequeño y documentado, como
-- pide el hallazgo, en vez de dividir el modelo por tipo.
CREATE OR REPLACE FUNCTION time_block_series_date_check_parent()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
DECLARE
  v_series public.time_block_series%ROWTYPE;
BEGIN
  SELECT * INTO v_series
  FROM public.time_block_series
  WHERE id = NEW.series_id AND barbershop_id = NEW.barbershop_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'time_block_series % no existe en la barbería %', NEW.series_id, NEW.barbershop_id;
  END IF;

  IF v_series.recurrence_kind <> 'date_list' THEN
    RAISE EXCEPTION 'time_block_series % no es date_list (es %): no admite fechas explícitas',
      NEW.series_id, v_series.recurrence_kind;
  END IF;

  IF NEW.block_date < v_series.effective_from
     OR (v_series.effective_until IS NOT NULL AND NEW.block_date > v_series.effective_until) THEN
    RAISE EXCEPTION 'block_date % fuera del rango efectivo [%, %] de time_block_series %',
      NEW.block_date, v_series.effective_from, v_series.effective_until, NEW.series_id;
  END IF;

  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION time_block_series_date_check_parent() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION time_block_series_date_check_parent() TO barberia_app;

COMMENT ON FUNCTION time_block_series_date_check_parent() IS
  'Valida que la fecha explícita pertenezca a una serie date_list vigente en ese rango '
  '(DDL-INT-03). No puede expresarse como CHECK: necesita leer otra tabla.';

CREATE TRIGGER time_block_series_date_check_parent_trg
  BEFORE INSERT OR UPDATE ON time_block_series_date
  FOR EACH ROW EXECUTE FUNCTION time_block_series_date_check_parent();

-- ---------------------------------------------------------------------------
-- 4. `time_block_series_exception` — "esta instancia no"
-- ---------------------------------------------------------------------------

CREATE TABLE time_block_series_exception (
  barbershop_id uuid        NOT NULL,
  series_id     uuid        NOT NULL,
  excluded_date date        NOT NULL,
  reason        text,
  created_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT time_block_series_exception_pk PRIMARY KEY (barbershop_id, series_id, excluded_date),
  -- RESTRICT, no CASCADE (CT-008/DEC-070): mismo criterio que
  -- time_block_series_date.
  CONSTRAINT time_block_series_exception_barbershop_id_series_id_fk
    FOREIGN KEY (barbershop_id, series_id)
    REFERENCES time_block_series (barbershop_id, id) ON DELETE RESTRICT,
  CONSTRAINT time_block_series_exception_reason_ck CHECK (
    reason IS NULL OR char_length(reason) <= 200
  )
);

COMMENT ON TABLE time_block_series_exception IS
  'Instancias suprimidas de una serie recurrente ("esta instancia", RN-BLQ-01). '
  'Clasificación: negocio.';

ALTER TABLE time_block_series_date      ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block_series_date      FORCE  ROW LEVEL SECURITY;
ALTER TABLE time_block_series_exception ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block_series_exception FORCE  ROW LEVEL SECURITY;

CREATE POLICY time_block_series_date_all_admin_policy ON time_block_series_date
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY time_block_series_date_select_tenant_policy ON time_block_series_date
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_date_insert_tenant_policy ON time_block_series_date
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_date_delete_tenant_policy ON time_block_series_date
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_series_exception_all_admin_policy ON time_block_series_exception
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY time_block_series_exception_select_tenant_policy ON time_block_series_exception
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_exception_insert_tenant_policy ON time_block_series_exception
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY time_block_series_exception_delete_tenant_policy ON time_block_series_exception
  FOR DELETE TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, DELETE ON TABLE time_block_series_date      TO barberia_app;
GRANT SELECT, INSERT, DELETE ON TABLE time_block_series_exception TO barberia_app;

-- ---------------------------------------------------------------------------
-- 5. `time_block` — bloqueo puntual, incluida la emergencia (RN-BLQ-03)
-- ---------------------------------------------------------------------------

CREATE TABLE time_block (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  block_type    text        NOT NULL,
  source        text        NOT NULL DEFAULT 'manual',
  starts_at     timestamptz NOT NULL,
  ends_at       timestamptz NOT NULL,
  reason        text,
  deleted_at    timestamptz,
  deleted_by    uuid,
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT time_block_id_pk PRIMARY KEY (id),

  -- RESTRICT, no CASCADE (CT-008/DEC-070): mismo criterio que working_hour;
  -- time_block ya usa eliminación lógica propia (deleted_at/deleted_by).
  CONSTRAINT time_block_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,

  -- RESTRICT, no SET NULL (DDL-INT-02): barbershop_id es NOT NULL y una FK
  -- compuesta con SET NULL anularía ambas columnas a la vez, violando el
  -- NOT NULL y haciendo fallar el DELETE del staff_user. staff_user nunca se
  -- borra físicamente en este diseño (se desactiva, no se elimina).
  CONSTRAINT time_block_barbershop_id_deleted_by_fk
    FOREIGN KEY (barbershop_id, deleted_by)
    REFERENCES staff_user (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT time_block_block_type_ck CHECK (
    block_type IN ('break', 'lunch', 'unavailable', 'day_off', 'holiday', 'vacation', 'emergency')
  ),
  CONSTRAINT time_block_source_ck CHECK (source IN ('manual', 'holiday_calendar')),

  -- Intervalo semiabierto [starts_at, ends_at) con fin estrictamente posterior
  -- (RN-DIS-05).
  CONSTRAINT time_block_interval_ck   CHECK (ends_at > starts_at),
  CONSTRAINT time_block_reason_ck     CHECK (reason IS NULL OR char_length(reason) <= 200),
  CONSTRAINT time_block_deleted_by_ck CHECK (
    (deleted_at IS NULL AND deleted_by IS NULL) OR deleted_at IS NOT NULL
  ),
  -- DDL-TMP-01: la eliminación lógica no puede preceder a la creación.
  CONSTRAINT time_block_deleted_at_ck CHECK (deleted_at IS NULL OR deleted_at >= created_at)
);

COMMENT ON TABLE time_block IS
  'Bloqueo puntual de agenda, incluida la emergencia (RN-BLQ-01, RN-BLQ-03). '
  'Propietario funcional: barbería. Retención: permanente; eliminación lógica (RN-BLQ-04, DEC-009). '
  'Clasificación: negocio.';

COMMENT ON COLUMN time_block.starts_at IS
  'DELIBERADAMENTE SIN restricción de exclusión contra appointment: RN-BLQ-03 y DEC-008 exigen '
  'que un bloqueo urgente SIEMPRE se pueda crear aunque haya citas encima. Impedirlo sería lo '
  'contrario de lo que el barbero necesita en una emergencia. El flujo asistido de citas '
  'afectadas vive en la aplicación.';

-- Índice de base-datos.md §12. DDL-PER-01: sin predicado parcial, por el
-- mismo motivo que idx_time_block_series_shop_barber -sirve la agenda
-- vigente y la auditoría de bloqueos borrados lógicamente por barbero sin
-- duplicar el índice.
CREATE INDEX idx_time_block_shop_barber_starts_at
  ON time_block (barbershop_id, barber_id, starts_at);

-- DDL-PER-01: lado referenciante de deleted_by sin índice ("¿qué bloqueos
-- eliminó este miembro del staff?"). Parcial: la mayoría de filas nunca se
-- eliminan (deleted_by IS NULL).
CREATE INDEX idx_time_block_shop_deleted_by
  ON time_block (barbershop_id, deleted_by)
  WHERE deleted_by IS NOT NULL;

CREATE TRIGGER time_block_set_updated_at
  BEFORE UPDATE ON time_block
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE time_block ENABLE ROW LEVEL SECURITY;
ALTER TABLE time_block FORCE  ROW LEVEL SECURITY;

CREATE POLICY time_block_all_admin_policy ON time_block
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY time_block_select_tenant_policy ON time_block
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_insert_tenant_policy ON time_block
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY time_block_update_tenant_policy ON time_block
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: RN-BLQ-04.
GRANT SELECT, INSERT, UPDATE ON TABLE time_block TO barberia_app;
