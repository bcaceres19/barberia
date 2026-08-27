-- Propósito
--   HU-060: base persistente de B3. Crea `customer` (persona que reserva,
--   también en la cita manual), `appointment` (la cita/turno, con el
--   criterio derivado `occupies_schedule` y la restricción de exclusión que
--   hace imposible persistir citas cruzadas del mismo barbero),
--   `appointment_history` y `appointment_history_change` (historial
--   append-only). Ningún endpoint HTTP, formulario ni pantalla se agrega en
--   esta migración: HU-060 solo entrega la primitiva transaccional interna
--   del módulo Go `booking`.
--
-- Reglas y decisiones
--   RN-CON-01/RN-CON-03 (la restricción de exclusión es la última defensa
--   contra cruces), RN-DIS-05 (intervalo semiabierto [inicio, fin)),
--   RN-DIS-07 (instantes timestamptz inequívocos), RN-HIS-01/RN-HIS-02
--   (historial completo e inmutable), RN-RES-02/RN-RES-03 (persona que
--   reserva y persona atendida pueden diferir), RN-TEN-01 (aislamiento por
--   barbería), DEC-002 (duración planificada, no real), DEC-004 (snapshots
--   de servicio, sin sincronización automática), DEC-007 (zonas horarias:
--   timestamptz sin depender de zona de sesión), DEC-014 (historial
--   inmutable), DEC-016 (vocabulario técnico en inglés/snake_case), DEC-019
--   (modelo multi-barbero), DEC-020 (intervalos [inicio, fin)), DEC-024
--   (RLS por fila), DEC-035-DEC-040 (estándares de ingeniería, Atlas,
--   OpenAPI, Git, diseño visual, roles), DEC-041 (vocabulario cerrado de
--   `event_type`), DEC-045 (identidad de `customer` por teléfono dentro de
--   la barbería), DEC-046 (unicidad de correo por barbería), DEC-070
--   (`ON DELETE RESTRICT`, nunca CASCADE).
--
-- Relación con database/modelo-fisico-referencia.sql §D.0-D.2
--   Copia reconciliada de esas tres secciones contra las quince migraciones
--   ya aplicadas: los nombres de columna de `barber`/`service` coinciden
--   exactamente (20260823130000_create_barber.sql,
--   20260824140000_create_service.sql), así que no hubo diferencia de
--   fondo que resolver. Se agregan las precondiciones de esta migración
--   (sección 1) y se documenta aquí, no en el modelo de referencia, que
--   `origin` acepta 'public' y 'manual' aunque HU-060 no expone ninguno de
--   los dos flujos todavía (B4 y HU-061, respectivamente).
--
-- Qué NO hace esta migración
--   No agrega `appointment_access_token`, programación de recordatorios,
--   intentos de notificación ni columnas de anonimización adicionales
--   (pertenecen a B4-B6, secciones E-G del modelo de referencia). No decide
--   cómo se reutiliza un cliente manual sin teléfono (DP-CIT-01): la forma
--   de la tabla permite `phone`/`email` nulos, pero la política de
--   reconciliación es de HU-061. No valida jornada, bloqueos ni asignación
--   servicio-barbero: eso lo hace el caso de uso de HU-061 antes de invocar
--   la primitiva de este módulo, nunca esta migración.
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

  IF NOT EXISTS (
    SELECT 1 FROM pg_extension WHERE extname = 'btree_gist'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de la extensión btree_gist '
      '(20260807170000_create_tenant_foundation.sql).';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'customer'
  ) THEN
    RAISE EXCEPTION 'customer ya existe: esta migración no puede aplicarse dos veces.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment'
  ) THEN
    RAISE EXCEPTION 'appointment ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Tabla `customer`
-- ---------------------------------------------------------------------------
-- Toda cita referencia un cliente, también la manual: un único lugar donde
-- anonimizar (RN-DAT-03) en vez de datos personales repartidos entre
-- tablas. `phone` y `email` son opcionales porque la cita manual puede
-- omitirlos (RN-CIT-02); la política de reconciliación sin teléfono queda
-- para HU-061 (DP-CIT-01), fuera de esta migración.

CREATE TABLE customer (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  full_name      text        NOT NULL,
  phone          text,
  email          text,
  anonymized_at  timestamptz,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT customer_id_pk               PRIMARY KEY (id),
  CONSTRAINT customer_barbershop_id_id_uk UNIQUE (barbershop_id, id),
  CONSTRAINT customer_barbershop_id_fk    FOREIGN KEY (barbershop_id)
                                          REFERENCES barbershop (id) ON DELETE RESTRICT,

  CONSTRAINT customer_full_name_ck CHECK (
    btrim(full_name) <> '' AND char_length(full_name) <= 120
  ),
  CONSTRAINT customer_phone_ck CHECK (phone IS NULL OR phone ~ '^\+[1-9][0-9]{7,14}$'),
  -- Forma canónica explícita (sin espacios, ya recortado), no solo el
  -- patrón mínimo de arroba y punto (mismo criterio que staff_user.email).
  CONSTRAINT customer_email_ck CHECK (
    email IS NULL
    OR (
      email = lower(email)
      AND email = btrim(email)
      AND email !~ '\s'
      AND email LIKE '_%@_%._%'
      AND char_length(email) <= 254
    )
  ),

  -- La anonimización futura (RN-DAT-03, DEC-025, fuera de alcance de
  -- HU-060) será idempotente y verificable: una fila anonimizada no
  -- conservará teléfono, correo ni nombre real. El marcador se fija ahora
  -- para que el CHECK no deba reescribirse cuando B6 la implemente.
  CONSTRAINT customer_anonymized_ck CHECK (
    anonymized_at IS NULL
    OR (phone IS NULL AND email IS NULL AND full_name = 'Cliente anonimizado')
  )
);

COMMENT ON TABLE customer IS
  'Persona que reserva. Propietario funcional: barbería. Retención: configurable, contada '
  'desde la última actividad (DEC-042, B6); después, anonimización (RN-DAT-03, fuera de '
  'alcance de HU-060). Clasificación: DATO PERSONAL. '
  'La persona atendida NO se guarda aquí: vive en appointment.attendee_name (RN-RES-03).';

-- Único por teléfono dentro de la barbería (DEC-045): permite reutilizar el
-- cliente sin una carrera entre SELECT e INSERT en el flujo público futuro.
CREATE UNIQUE INDEX idx_customer_shop_phone
  ON customer (barbershop_id, phone)
  WHERE phone IS NOT NULL AND anonymized_at IS NULL;

-- DEC-046: el correo es único por barbería, no global.
CREATE UNIQUE INDEX idx_customer_shop_email
  ON customer (barbershop_id, email)
  WHERE email IS NOT NULL AND anonymized_at IS NULL;

CREATE TRIGGER customer_set_updated_at
  BEFORE UPDATE ON customer
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE customer ENABLE ROW LEVEL SECURITY;
ALTER TABLE customer FORCE  ROW LEVEL SECURITY;

CREATE POLICY customer_all_admin_policy ON customer
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY customer_select_tenant_policy ON customer
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY customer_insert_tenant_policy ON customer
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY customer_update_tenant_policy ON customer
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: los datos personales se anonimizan, no se borran (RN-DAT-03).
GRANT SELECT, INSERT, UPDATE ON TABLE customer TO barberia_app;

-- ---------------------------------------------------------------------------
-- 3. Tabla `appointment`
-- ---------------------------------------------------------------------------

CREATE TABLE appointment (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  barber_id     uuid        NOT NULL,
  service_id    uuid        NOT NULL,
  customer_id   uuid        NOT NULL,

  -- RN-RES-02/RN-RES-03: quien reserva y quien es atendido pueden diferir.
  -- Se resuelve al crear la cita y se guarda aquí, no se deriva del cliente
  -- en tiempo de lectura.
  attendee_name text        NOT NULL,

  starts_at     timestamptz NOT NULL,
  ends_at       timestamptz NOT NULL,

  status        text        NOT NULL DEFAULT 'confirmed',

  -- estados-citas.md §11: columna derivada y almacenada para que la
  -- restricción de exclusión y el cálculo de disponibilidad (B4) usen
  -- exactamente el mismo criterio y no puedan divergir.
  occupies_schedule boolean GENERATED ALWAYS AS (
    status IN ('confirmed', 'completed', 'no_show')
  ) STORED,

  -- 'public' (B4, reserva del cliente) y 'manual' (HU-061, barbero). HU-060
  -- no expone ninguno de los dos flujos: solo persiste el origen que el
  -- llamador ya decidió.
  origin text NOT NULL,

  -- Snapshots autorizados por DEC-004: se fijan al crear o al aplicar un
  -- cambio explícito; NUNCA se sincronizan en segundo plano con `service`.
  service_name_snapshot     text           NOT NULL,
  duration_minutes_snapshot integer        NOT NULL,
  price_amount_snapshot     numeric(12, 2) NOT NULL,
  price_currency_snapshot   char(3)        NOT NULL,

  customer_note       text,
  cancellation_reason text,

  -- Instante efectivo del resultado, para métricas futuras (B3 posterior).
  -- HU-060 nunca lo fija: toda cita que crea queda 'confirmed'.
  resolved_at timestamptz,

  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT appointment_id_pk               PRIMARY KEY (id),
  CONSTRAINT appointment_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- Pertenencia tenant-aware en ambos lados: una FK simple por barber_id no
  -- demostraría que el barbero es de esta barbería (estandar-base-datos.md §6).
  CONSTRAINT appointment_barbershop_id_barber_id_fk
    FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_barbershop_id_service_id_fk
    FOREIGN KEY (barbershop_id, service_id)
    REFERENCES service (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_barbershop_id_customer_id_fk
    FOREIGN KEY (barbershop_id, customer_id)
    REFERENCES customer (barbershop_id, id) ON DELETE RESTRICT,

  -- Los cinco estados confirmados de estados-citas.md §2. Texto, nunca
  -- números: un `estado = 3` es ilegible en soporte (estados-citas.md §11).
  CONSTRAINT appointment_status_ck CHECK (
    status IN ('confirmed', 'completed', 'cancelled_by_customer', 'cancelled_by_barber', 'no_show')
  ),
  CONSTRAINT appointment_origin_ck CHECK (origin IN ('public', 'manual')),

  CONSTRAINT appointment_interval_ck      CHECK (ends_at > starts_at),
  CONSTRAINT appointment_attendee_name_ck CHECK (
    btrim(attendee_name) <> '' AND char_length(attendee_name) <= 120
  ),
  CONSTRAINT appointment_duration_snapshot_ck CHECK (duration_minutes_snapshot BETWEEN 1 AND 1440),
  CONSTRAINT appointment_price_snapshot_ck    CHECK (price_amount_snapshot >= 0),
  CONSTRAINT appointment_currency_snapshot_ck CHECK (price_currency_snapshot ~ '^[A-Z]{3}$'),
  CONSTRAINT appointment_service_name_snapshot_ck CHECK (
    btrim(service_name_snapshot) <> '' AND char_length(service_name_snapshot) <= 120
  ),
  CONSTRAINT appointment_customer_note_ck CHECK (
    customer_note IS NULL OR char_length(customer_note) <= 500
  ),

  -- La resta de dos timestamptz es la diferencia exacta entre dos instantes
  -- absolutos, no una suma calendárica: no depende de la zona de sesión ni
  -- de DST (probado contra PostgreSQL real cruzando medianoche y un cambio
  -- de horario de verano, CA-060-04).
  CONSTRAINT appointment_duration_matches_snapshot_ck CHECK (
    EXTRACT(EPOCH FROM (ends_at - starts_at)) = duration_minutes_snapshot * 60
  ),

  -- Un motivo de cancelación solo tiene sentido en una cita cancelada, con
  -- cota. HU-060 nunca inserta uno (no cancela), pero la forma ya está
  -- lista para HU-061 en adelante.
  CONSTRAINT appointment_cancellation_reason_ck CHECK (
    cancellation_reason IS NULL
    OR (
      status IN ('cancelled_by_customer', 'cancelled_by_barber')
      AND char_length(cancellation_reason) <= 500
    )
  ),

  -- `confirmed` es el único estado no terminal; los cuatro restantes tienen
  -- instante de resolución, que nunca precede a la creación de la cita.
  CONSTRAINT appointment_resolved_at_ck CHECK (
    (status = 'confirmed' AND resolved_at IS NULL)
    OR
    (status <> 'confirmed' AND resolved_at IS NOT NULL AND resolved_at >= created_at)
  )
);

COMMENT ON TABLE appointment IS
  'Cita (turno en la interfaz, DEC-016). Propietario funcional: barbería. '
  'Retención: permanente en lo operativo. Clasificación: negocio + dato personal '
  '(attendee_name, customer_note). Nunca se borra en cascada desde barbería, servicio ni barbero.';

COMMENT ON COLUMN appointment.occupies_schedule IS
  'Criterio ÚNICO de "ocupa agenda" (estados-citas.md §4 y §11). Lo usan la restricción de '
  'exclusión y el cálculo de disponibilidad de B4. Cambiar el conjunto de estados que ocupan '
  'agenda se hace aquí y en ningún otro lugar.';

COMMENT ON COLUMN appointment.duration_minutes_snapshot IS
  'Duración con la que se creó la cita (DEC-004). La igualdad '
  'ends_at = starts_at + duration_minutes_snapshot no puede expresarse en un CHECK porque la '
  'suma timestamptz + interval es STABLE, no IMMUTABLE: la garantiza el servicio de citas y '
  'la verifica su prueba de integración.';

-- ---------------------------------------------------------------------------
-- La restricción central del producto (RN-CON-01, RN-CON-03, F-DISP-02)
-- ---------------------------------------------------------------------------
-- Última línea de defensa: aunque dos procesos comprueben a la vez que la
-- franja está libre, PostgreSQL acepta uno solo (RN-CON-02). No puede
-- descansar en el frontend, en un SELECT previo ni en un bloqueo en memoria.
--
-- `'[)'` implementa el intervalo semiabierto: dos citas contiguas
-- (10:00-11:00 y 11:00-12:00) NO se cruzan (RN-DIS-05).
--
-- El predicado usa la columna generada, no una lista de estados repetida:
-- si mañana `pending` ocupa agenda (estados-citas.md §2.3), se cambia la
-- columna y la restricción sigue siendo correcta sin reescribirla.
ALTER TABLE appointment
  ADD CONSTRAINT appointment_barber_interval_excl
  EXCLUDE USING gist (
    barbershop_id WITH =,
    barber_id     WITH =,
    tstzrange(starts_at, ends_at, '[)') WITH &&
  ) WHERE (occupies_schedule);

-- Agenda diaria del barbero (B3 posterior) y cálculo de disponibilidad (B4).
CREATE INDEX idx_appointment_shop_barber_starts_at
  ON appointment (barbershop_id, barber_id, starts_at);

-- Historial del cliente y vista de sus turnos, del más reciente al más antiguo.
CREATE INDEX idx_appointment_shop_customer_starts_at
  ON appointment (barbershop_id, customer_id, starts_at DESC);

-- "Citas por servicio" (por ejemplo, para desactivarlo con RN-SER-03, B1).
CREATE INDEX idx_appointment_shop_service
  ON appointment (barbershop_id, service_id);

-- Cola del cierre automático de citas vencidas (RN-CIT-05, historia
-- futura): solo las activas.
CREATE INDEX idx_appointment_open_ends_at
  ON appointment (barbershop_id, ends_at)
  WHERE status = 'confirmed';

CREATE TRIGGER appointment_set_updated_at
  BEFORE UPDATE ON appointment
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE appointment ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_all_admin_policy ON appointment
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY appointment_select_tenant_policy ON appointment
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY appointment_insert_tenant_policy ON appointment
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY appointment_update_tenant_policy ON appointment
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni privilegio de DELETE: una cita no se borra, se cancela
-- (historia futura).
GRANT SELECT, INSERT, UPDATE ON TABLE appointment TO barberia_app;

-- ---------------------------------------------------------------------------
-- 4. Tabla `appointment_history`
-- ---------------------------------------------------------------------------
-- Append-only para el rol de aplicación, garantizado por dos mecanismos
-- independientes: no se conceden privilegios UPDATE/DELETE, y no existe
-- política RLS que los permita. Una entrada errónea se corrige con otra
-- entrada, nunca editando la anterior (RN-HIS-02, DEC-014).

CREATE TABLE appointment_history (
  id                   uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id        uuid        NOT NULL,
  appointment_id       uuid        NOT NULL,
  event_type           text        NOT NULL,
  actor_type           text        NOT NULL,
  actor_staff_user_id  uuid,
  actor_customer_id    uuid,
  reason               text,
  occurred_at          timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT appointment_history_id_pk               PRIMARY KEY (id),
  CONSTRAINT appointment_history_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  CONSTRAINT appointment_history_barbershop_id_appointment_id_fk
    FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_history_barbershop_id_actor_staff_user_id_fk
    FOREIGN KEY (barbershop_id, actor_staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_history_barbershop_id_actor_customer_id_fk
    FOREIGN KEY (barbershop_id, actor_customer_id)
    REFERENCES customer (barbershop_id, id) ON DELETE RESTRICT,

  -- Vocabulario en inglés y snake_case (DEC-041): los ocho valores
  -- confirmados; HU-060 solo emite 'appointment_created'.
  CONSTRAINT appointment_history_event_type_ck CHECK (
    event_type IN (
      'appointment_created',
      'appointment_rescheduled',
      'appointment_service_changed',
      'appointment_completed',
      'appointment_cancelled_by_customer',
      'appointment_cancelled_by_barber',
      'appointment_no_show',
      'appointment_status_corrected'
    )
  ),

  -- RN-HIS-01: el actor puede ser el barbero, el cliente o el sistema.
  CONSTRAINT appointment_history_actor_type_ck CHECK (
    actor_type IN ('staff', 'customer', 'system')
  ),
  CONSTRAINT appointment_history_actor_shape_ck CHECK (
    (actor_type = 'staff'    AND actor_staff_user_id IS NOT NULL AND actor_customer_id IS NULL)
    OR
    (actor_type = 'customer' AND actor_customer_id IS NOT NULL   AND actor_staff_user_id IS NULL)
    OR
    (actor_type = 'system'   AND actor_staff_user_id IS NULL     AND actor_customer_id IS NULL)
  ),

  -- La corrección auditada (T8, historia futura) exigirá motivo
  -- obligatorio; en los demás casos es opcional pero acotado.
  CONSTRAINT appointment_history_reason_ck CHECK (
    (event_type = 'appointment_status_corrected'
       AND reason IS NOT NULL AND btrim(reason) <> '' AND char_length(reason) <= 500)
    OR
    (event_type <> 'appointment_status_corrected'
       AND (reason IS NULL OR (btrim(reason) <> '' AND char_length(reason) <= 500)))
  )
);

COMMENT ON TABLE appointment_history IS
  'Historial inmutable de cambios de una cita (RN-HIS-01, RN-HIS-02, DEC-014). '
  'Propietario funcional: barbería. Retención: permanente, incluso tras anonimizar al cliente. '
  'Clasificación: auditoría. APPEND-ONLY: el rol de aplicación no tiene UPDATE ni DELETE.';

CREATE INDEX idx_appointment_history_shop_appointment_occurred
  ON appointment_history (barbershop_id, appointment_id, occurred_at);

CREATE INDEX idx_appointment_history_shop_actor_customer
  ON appointment_history (barbershop_id, actor_customer_id)
  WHERE actor_customer_id IS NOT NULL;

CREATE INDEX idx_appointment_history_shop_actor_staff
  ON appointment_history (barbershop_id, actor_staff_user_id)
  WHERE actor_staff_user_id IS NOT NULL;

ALTER TABLE appointment_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_history FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_history_all_admin_policy ON appointment_history
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY appointment_history_select_tenant_policy ON appointment_history
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY appointment_history_insert_tenant_policy ON appointment_history
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Deliberadamente sin políticas de UPDATE ni DELETE.
GRANT SELECT, INSERT ON TABLE appointment_history TO barberia_app;
REVOKE UPDATE, DELETE ON TABLE appointment_history FROM barberia_app;

-- ---------------------------------------------------------------------------
-- 5. Tabla `appointment_history_change`
-- ---------------------------------------------------------------------------
-- Valores anterior y nuevo de cada campo tocado (RN-HIS-01). Tabla hija en
-- lugar de `jsonb`: mantiene 1FN y evita la dependencia prohibida por
-- AGENTS.md. HU-060 no inserta ninguna fila aquí todavía (appointment_created
-- no tiene "valor anterior"); la tabla queda lista para HU-061 en adelante.

CREATE TABLE appointment_history_change (
  id              uuid NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id   uuid NOT NULL,
  history_id      uuid NOT NULL,
  field_name      text NOT NULL,
  previous_value  text,
  new_value       text,

  CONSTRAINT appointment_history_change_id_pk PRIMARY KEY (id),

  CONSTRAINT appointment_history_change_barbershop_id_history_id_fk
    FOREIGN KEY (barbershop_id, history_id)
    REFERENCES appointment_history (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT appointment_history_change_field_name_uk
    UNIQUE (barbershop_id, history_id, field_name),

  CONSTRAINT appointment_history_change_field_name_ck CHECK (
    btrim(field_name) <> '' AND char_length(field_name) <= 60
  ),
  -- Una entrada sin ningún valor no aporta nada.
  CONSTRAINT appointment_history_change_value_ck CHECK (
    previous_value IS NOT NULL OR new_value IS NOT NULL
  ),
  CONSTRAINT appointment_history_change_previous_value_ck CHECK (
    previous_value IS NULL OR char_length(previous_value) <= 1000
  ),
  CONSTRAINT appointment_history_change_new_value_ck CHECK (
    new_value IS NULL OR char_length(new_value) <= 1000
  )
);

COMMENT ON TABLE appointment_history_change IS
  'Campos modificados con su valor anterior y nuevo (RN-HIS-01). APPEND-ONLY, igual que su '
  'padre. Clasificación: auditoría; puede contener datos personales.';

ALTER TABLE appointment_history_change ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_history_change FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_history_change_all_admin_policy ON appointment_history_change
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY appointment_history_change_select_tenant_policy ON appointment_history_change
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_history_change_insert_tenant_policy ON appointment_history_change
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT ON TABLE appointment_history_change TO barberia_app;
REVOKE UPDATE, DELETE ON TABLE appointment_history_change FROM barberia_app;
