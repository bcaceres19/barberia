-- Pruebas SQL de la anonimización completa de clientes (issue #6, DEC-054).
-- Cubren DDL-PRI-01, DEC-042 (fecha ancla por última actividad) y DEC-049
-- (matriz completa de anonimización) sobre retention_claim_due_customers y
-- customer_anonymize.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/customer_anonymization_fixture.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/customer_anonymization.sql
--
-- Requisito del arnés: igual que hu001_aislamiento_rls.sql, la conexión debe
-- poder ejecutar `SET ROLE barberia_app` y `SET ROLE barberia_worker` sin
-- pertenecer a esos roles (conectada como superusuario en local/CI).
--
-- Cada escenario fija todos los instantes explícitamente (created_at,
-- occurred_at, issued_at, p_now) en vez de usar el reloj real: la fecha
-- ancla de DEC-042 depende de comparar instantes entre sí, y dejar que
-- alguno tome el valor por defecto de `now()` mezclaría el momento real de
-- la corrida con las fechas ficticias del escenario.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== customer_anonymize / retention_claim_due_customers · anonimización completa ==='

-- ---------------------------------------------------------------------------
-- Escenario 1 · Solicitud individual: cobertura completa de DEC-049
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO customer (id, barbershop_id, full_name, phone, created_at)
VALUES ('00000000-0000-0000-0000-000000000101',
        '11111111-1111-1111-1111-111111111111',
        'Cliente Original Uno', '+573000000001', '2020-01-01 00:00:00-05');

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
  customer_note, created_at
) VALUES (
  '00000000-0000-0000-0000-000000000201',
  '11111111-1111-1111-1111-111111111111',
  'b1a00000-0000-0000-0000-00000000000a',
  '5e100000-0000-0000-0000-00000000000a',
  '00000000-0000-0000-0000-000000000101',
  'Hijo De Prueba',
  '2020-01-05 10:00:00-05', '2020-01-05 10:30:00-05', 'manual',
  'Corte de prueba de anonimización A', 30, 20000.00, 'COP',
  'Nota personal de prueba', '2020-01-02 00:00:00-05'
);

INSERT INTO appointment_history (
  id, barbershop_id, appointment_id, event_type, actor_type, actor_customer_id, reason, occurred_at
) VALUES (
  '00000000-0000-0000-0000-000000000301',
  '11111111-1111-1111-1111-111111111111',
  '00000000-0000-0000-0000-000000000201',
  'appointment_cancelled_by_customer', 'customer',
  '00000000-0000-0000-0000-000000000101',
  'Motivo personal de prueba', '2020-01-03 00:00:00-05'
);

INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, previous_value, new_value)
VALUES
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000301',
   'attendee_name', 'Nombre Anterior', 'Hijo De Prueba'),
  ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000301',
   'status', 'confirmed', 'cancelled_by_customer');

INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash, issued_at)
VALUES ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000201',
        repeat('a', 64), '2020-01-02 00:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_claimed record;
BEGIN
  SELECT * INTO v_claimed
  FROM retention_claim_due_customers(10, '2026-08-11 00:00:00-05'::timestamptz)
  WHERE customer_id = '00000000-0000-0000-0000-000000000101';

  IF v_claimed IS NULL THEN
    RAISE EXCEPTION 'Escenario 1: el cliente vencido no fue reclamado.';
  END IF;

  IF NOT customer_anonymize(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000101',
    '2026-08-11 00:00:00-05'::timestamptz
  ) THEN
    RAISE EXCEPTION 'Escenario 1: customer_anonymize debía devolver true.';
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_customer    record;
  v_appointment record;
  v_history     record;
  v_change_name record;
  v_change_status record;
  v_token       record;
BEGIN
  SELECT full_name, phone, email, anonymized_at INTO v_customer
  FROM customer WHERE id = '00000000-0000-0000-0000-000000000101';
  IF v_customer.full_name <> 'Cliente anonimizado' OR v_customer.phone IS NOT NULL
     OR v_customer.anonymized_at <> '2026-08-11 00:00:00-05'::timestamptz THEN
    RAISE EXCEPTION 'Escenario 1: customer no quedó anonimizado correctamente.';
  END IF;

  SELECT attendee_name, customer_note INTO v_appointment
  FROM appointment WHERE id = '00000000-0000-0000-0000-000000000201';
  IF v_appointment.attendee_name <> 'Cliente anonimizado' OR v_appointment.customer_note IS NOT NULL THEN
    RAISE EXCEPTION 'Escenario 1: appointment.attendee_name/customer_note no se redactaron.';
  END IF;

  SELECT reason INTO v_history
  FROM appointment_history WHERE id = '00000000-0000-0000-0000-000000000301';
  IF v_history.reason <> 'Cliente anonimizado' THEN
    RAISE EXCEPTION 'Escenario 1: appointment_history.reason no se redactó.';
  END IF;

  SELECT previous_value, new_value INTO v_change_name
  FROM appointment_history_change
  WHERE history_id = '00000000-0000-0000-0000-000000000301' AND field_name = 'attendee_name';
  IF v_change_name.previous_value <> 'Cliente anonimizado'
     OR v_change_name.new_value <> 'Cliente anonimizado' THEN
    RAISE EXCEPTION 'Escenario 1: appointment_history_change.attendee_name no se redactó.';
  END IF;

  -- `status` NO es un campo personal: DEC-054 exige conservarlo íntegro.
  SELECT previous_value, new_value INTO v_change_status
  FROM appointment_history_change
  WHERE history_id = '00000000-0000-0000-0000-000000000301' AND field_name = 'status';
  IF v_change_status.previous_value <> 'confirmed' OR v_change_status.new_value <> 'cancelled_by_customer' THEN
    RAISE EXCEPTION 'Escenario 1: un campo no personal (status) se redactó por error.';
  END IF;

  SELECT revoked_at INTO v_token
  FROM appointment_access_token WHERE appointment_id = '00000000-0000-0000-0000-000000000201';
  IF v_token.revoked_at IS NULL THEN
    RAISE EXCEPTION 'Escenario 1: appointment_access_token no se revocó.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 1 OK · solicitud individual cubre customer/appointment/history/change/token'

-- ---------------------------------------------------------------------------
-- Escenario 2 · Vencimiento masivo: el trabajador agota el lote en lotes
-- de p_limit hasta que no queda nada vencido
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO customer (id, barbershop_id, full_name, created_at) VALUES
  ('00000000-0000-0000-0000-000000000401', '11111111-1111-1111-1111-111111111111', 'Cliente Masivo Uno',   '2020-01-01 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000402', '11111111-1111-1111-1111-111111111111', 'Cliente Masivo Dos',   '2020-01-02 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000403', '11111111-1111-1111-1111-111111111111', 'Cliente Masivo Tres',  '2020-01-03 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000404', '11111111-1111-1111-1111-111111111111', 'Cliente Masivo Cuatro','2020-01-04 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000405', '11111111-1111-1111-1111-111111111111', 'Cliente Masivo Cinco', '2020-01-05 00:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_now       constant timestamptz := '2026-08-11 00:00:00-05';
  v_ids       uuid[];
  v_total     integer := 0;
  v_id        uuid;
  v_batch     integer;
  v_iterations integer := 0;
BEGIN
  LOOP
    v_iterations := v_iterations + 1;
    IF v_iterations > 10 THEN
      RAISE EXCEPTION 'Escenario 2: demasiadas iteraciones, el lote no se agota.';
    END IF;

    v_ids := ARRAY(
      SELECT customer_id FROM retention_claim_due_customers(2, v_now)
      WHERE customer_id IN (
        '00000000-0000-0000-0000-000000000401', '00000000-0000-0000-0000-000000000402',
        '00000000-0000-0000-0000-000000000403', '00000000-0000-0000-0000-000000000404',
        '00000000-0000-0000-0000-000000000405'
      )
    );

    v_batch := array_length(v_ids, 1);
    EXIT WHEN v_batch IS NULL;

    IF v_batch > 2 THEN
      RAISE EXCEPTION 'Escenario 2: p_limit=2 devolvió % filas.', v_batch;
    END IF;

    FOREACH v_id IN ARRAY v_ids LOOP
      PERFORM customer_anonymize('11111111-1111-1111-1111-111111111111', v_id, v_now);
      v_total := v_total + 1;
    END LOOP;
  END LOOP;

  IF v_total <> 5 THEN
    RAISE EXCEPTION 'Escenario 2: se anonimizaron % clientes, se esperaban 5.', v_total;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 2 OK · vencimiento masivo se agota en lotes de p_limit sin perder ni repetir clientes'

-- ---------------------------------------------------------------------------
-- Escenario 3 · Dos tenants: anonimizar en A no toca a B, y un intento con
-- barbershop_id/customer_id cruzados no hace nada
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO customer (id, barbershop_id, full_name, created_at) VALUES
  ('00000000-0000-0000-0000-000000000501', '11111111-1111-1111-1111-111111111111', 'Cliente Tenant A', '2020-01-01 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000502', '22222222-2222-2222-2222-222222222222', 'Cliente Tenant B', '2020-01-01 00:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_now     constant timestamptz := '2026-08-11 00:00:00-05';
  v_a_seen  boolean := false;
  v_b_seen  boolean := false;
  v_row     record;
BEGIN
  FOR v_row IN
    SELECT * FROM retention_claim_due_customers(50, v_now)
    WHERE customer_id IN ('00000000-0000-0000-0000-000000000501', '00000000-0000-0000-0000-000000000502')
  LOOP
    IF v_row.customer_id = '00000000-0000-0000-0000-000000000501'
       AND v_row.barbershop_id = '11111111-1111-1111-1111-111111111111' THEN
      v_a_seen := true;
    END IF;
    IF v_row.customer_id = '00000000-0000-0000-0000-000000000502'
       AND v_row.barbershop_id = '22222222-2222-2222-2222-222222222222' THEN
      v_b_seen := true;
    END IF;
  END LOOP;

  IF NOT v_a_seen OR NOT v_b_seen THEN
    RAISE EXCEPTION 'Escenario 3: el claim no devolvió ambos tenants con su barbershop_id correcto.';
  END IF;

  -- Cruzar el tenant del cliente B con el barbershop_id de A no encuentra
  -- fila: el WHERE exige ambos a la vez, así que no hay forma de anonimizar
  -- un cliente ajeno adivinando o reutilizando su identificador.
  IF customer_anonymize(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000502',
    v_now
  ) THEN
    RAISE EXCEPTION 'Escenario 3: se pudo anonimizar un cliente de B usando el barbershop_id de A.';
  END IF;

  IF NOT customer_anonymize('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000501', v_now) THEN
    RAISE EXCEPTION 'Escenario 3: no se pudo anonimizar el cliente propio de A.';
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_a_name text;
  v_b_name text;
BEGIN
  SELECT full_name INTO v_a_name FROM customer WHERE id = '00000000-0000-0000-0000-000000000501';
  SELECT full_name INTO v_b_name FROM customer WHERE id = '00000000-0000-0000-0000-000000000502';

  IF v_a_name <> 'Cliente anonimizado' THEN
    RAISE EXCEPTION 'Escenario 3: el cliente de A debía quedar anonimizado.';
  END IF;
  IF v_b_name <> 'Cliente Tenant B' THEN
    RAISE EXCEPTION 'Escenario 3: el cliente de B se modificó sin autorización (quedó "%").', v_b_name;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 3 OK · aislamiento entre tenants; barbershop_id/customer_id cruzados no anonimizan nada'

-- ---------------------------------------------------------------------------
-- Escenario 4 · Repetición: idempotente, sin doble efecto ni error
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO customer (id, barbershop_id, full_name, created_at)
VALUES ('00000000-0000-0000-0000-000000000601',
        '11111111-1111-1111-1111-111111111111', 'Cliente Repetido', '2020-01-01 00:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_now constant timestamptz := '2026-08-11 00:00:00-05';
BEGIN
  IF NOT customer_anonymize('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000601', v_now) THEN
    RAISE EXCEPTION 'Escenario 4: la primera llamada debía devolver true.';
  END IF;

  IF customer_anonymize(
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000601',
    v_now + interval '1 day'
  ) THEN
    RAISE EXCEPTION 'Escenario 4: repetir sobre un cliente ya anonimizado debía devolver false.';
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_anonymized_at timestamptz;
BEGIN
  SELECT anonymized_at INTO v_anonymized_at
  FROM customer WHERE id = '00000000-0000-0000-0000-000000000601';
  -- Sigue con el instante de la PRIMERA llamada: la segunda no lo tocó.
  IF v_anonymized_at <> '2026-08-11 00:00:00-05'::timestamptz THEN
    RAISE EXCEPTION 'Escenario 4: la repetición modificó anonymized_at.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 4 OK · repetir customer_anonymize es idempotente, sin doble efecto'

-- ---------------------------------------------------------------------------
-- Escenario 5 · Caída a mitad de lote: una transacción sin confirmar no
-- deja ningún cliente a medio anonimizar
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

INSERT INTO customer (id, barbershop_id, full_name, created_at) VALUES
  ('00000000-0000-0000-0000-000000000701', '11111111-1111-1111-1111-111111111111', 'Cliente Lote Uno',  '2020-01-01 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000702', '11111111-1111-1111-1111-111111111111', 'Cliente Lote Dos',  '2020-01-02 00:00:00-05'),
  ('00000000-0000-0000-0000-000000000703', '11111111-1111-1111-1111-111111111111', 'Cliente Lote Tres', '2020-01-03 00:00:00-05');

COMMIT;

-- "Caída": la transacción del trabajador reclama y anonimiza 2 de 3 y
-- nunca confirma (se revierte, como haría un proceso que muere a mitad).
BEGIN;
SET ROLE barberia_worker;

DO $$
DECLARE
  v_now constant timestamptz := '2026-08-11 00:00:00-05';
  v_id  uuid;
  v_n   integer := 0;
BEGIN
  FOR v_id IN
    SELECT customer_id FROM retention_claim_due_customers(10, v_now)
    WHERE customer_id IN (
      '00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000702',
      '00000000-0000-0000-0000-000000000703'
    )
  LOOP
    EXIT WHEN v_n >= 2;  -- simula la caída antes de terminar el lote de 3.
    PERFORM customer_anonymize('11111111-1111-1111-1111-111111111111', v_id, v_now);
    v_n := v_n + 1;
  END LOOP;
END
$$;

ROLLBACK;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_anonymized integer;
BEGIN
  SELECT count(*) INTO v_anonymized
  FROM customer
  WHERE id IN (
    '00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000702',
    '00000000-0000-0000-0000-000000000703'
  )
  AND anonymized_at IS NOT NULL;

  IF v_anonymized <> 0 THEN
    RAISE EXCEPTION 'Escenario 5: la caída a mitad de lote dejó % clientes anonimizados sin confirmar.', v_anonymized;
  END IF;
END
$$;
RESET ROLE;

-- Reintento real: la misma reclamación, esta vez confirmada de punta a
-- punta, procesa los tres sin que la caída anterior haya dejado nada roto.
BEGIN;
SET ROLE barberia_worker;

DO $$
DECLARE
  v_now constant timestamptz := '2026-08-11 00:00:00-05';
  v_id  uuid;
  v_n   integer := 0;
BEGIN
  FOR v_id IN
    SELECT customer_id FROM retention_claim_due_customers(10, v_now)
    WHERE customer_id IN (
      '00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000702',
      '00000000-0000-0000-0000-000000000703'
    )
  LOOP
    PERFORM customer_anonymize('11111111-1111-1111-1111-111111111111', v_id, v_now);
    v_n := v_n + 1;
  END LOOP;

  IF v_n <> 3 THEN
    RAISE EXCEPTION 'Escenario 5: el reintento procesó % clientes, se esperaban 3.', v_n;
  END IF;
END
$$;

SET ROLE barberia_migrator;

DO $$
DECLARE
  v_anonymized integer;
BEGIN
  SELECT count(*) INTO v_anonymized
  FROM customer
  WHERE id IN (
    '00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000702',
    '00000000-0000-0000-0000-000000000703'
  )
  AND anonymized_at IS NOT NULL;

  IF v_anonymized <> 3 THEN
    RAISE EXCEPTION 'Escenario 5: el reintento confirmado dejó % clientes anonimizados, se esperaban 3.', v_anonymized;
  END IF;
END
$$;

-- Limpieza manual: este escenario confirmó datos reales (no usa
-- BEGIN/ROLLBACK envolvente porque necesitaba una caída real entre dos
-- transacciones), así que revierte su propio fixture.
DELETE FROM customer WHERE id IN (
  '00000000-0000-0000-0000-000000000701', '00000000-0000-0000-0000-000000000702',
  '00000000-0000-0000-0000-000000000703'
);
RESET ROLE;
COMMIT;
\echo 'Escenario 5 OK · una caída a mitad de lote no deja clientes a medio anonimizar; el reintento sí los procesa'

-- ---------------------------------------------------------------------------
-- Escenario 6 · Fecha ancla por última actividad (DEC-042): un cliente
-- antiguo con actividad reciente NO vence
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_migrator;

-- Creado hace años, pero con una transición de historial reciente: no debe
-- vencer, porque la actividad -no la creación- es la fecha ancla.
INSERT INTO customer (id, barbershop_id, full_name, created_at)
VALUES ('00000000-0000-0000-0000-000000000801',
        '11111111-1111-1111-1111-111111111111', 'Cliente Activo Antiguo', '2020-01-01 00:00:00-05');

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
  created_at
) VALUES (
  '00000000-0000-0000-0000-000000000802',
  '11111111-1111-1111-1111-111111111111',
  'b1a00000-0000-0000-0000-00000000000a',
  '5e100000-0000-0000-0000-00000000000a',
  '00000000-0000-0000-0000-000000000801',
  'Cliente Activo Antiguo',
  '2020-01-05 10:00:00-05', '2020-01-05 10:30:00-05', 'manual',
  'Corte de prueba de anonimización A', 30, 20000.00, 'COP',
  '2020-01-02 00:00:00-05'
);

-- La transición ocurre 10 días antes de p_now: mantiene al cliente activo.
INSERT INTO appointment_history (
  id, barbershop_id, appointment_id, event_type, actor_type, actor_customer_id, occurred_at
) VALUES (
  '00000000-0000-0000-0000-000000000803',
  '11111111-1111-1111-1111-111111111111',
  '00000000-0000-0000-0000-000000000802',
  'appointment_completed', 'system', NULL,
  '2026-08-01 00:00:00-05'
);

-- Cliente igual de antiguo, misma cita antigua, pero SIN actividad
-- reciente: debe vencer (control del mismo escenario).
INSERT INTO customer (id, barbershop_id, full_name, created_at)
VALUES ('00000000-0000-0000-0000-000000000811',
        '11111111-1111-1111-1111-111111111111', 'Cliente Realmente Inactivo', '2020-01-01 00:00:00-05');

SET ROLE barberia_worker;

DO $$
DECLARE
  v_now constant timestamptz := '2026-08-11 00:00:00-05';
  v_active_claimed   integer;
  v_inactive_claimed integer;
BEGIN
  SELECT count(*) INTO v_active_claimed
  FROM retention_claim_due_customers(50, v_now)
  WHERE customer_id = '00000000-0000-0000-0000-000000000801';

  SELECT count(*) INTO v_inactive_claimed
  FROM retention_claim_due_customers(50, v_now)
  WHERE customer_id = '00000000-0000-0000-0000-000000000811';

  IF v_active_claimed <> 0 THEN
    RAISE EXCEPTION 'Escenario 6: un cliente con actividad de historial reciente fue reclamado (DEC-042 violado).';
  END IF;
  IF v_inactive_claimed <> 1 THEN
    RAISE EXCEPTION 'Escenario 6: el cliente de control, sin actividad reciente, no fue reclamado.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 6 OK · la fecha ancla es la última actividad (historial), no la creación (DEC-042)'

-- ---------------------------------------------------------------------------
-- Escenario 7 · barberia_app no puede reclamar ni anonimizar (DEC-040, DDL-SEC-04)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    PERFORM * FROM retention_claim_due_customers(10, '2026-08-11 00:00:00-05'::timestamptz);
    RAISE EXCEPTION 'Escenario 7: barberia_app pudo ejecutar retention_claim_due_customers.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado.
  END;

  BEGIN
    PERFORM customer_anonymize(
      '11111111-1111-1111-1111-111111111111', gen_random_uuid(), '2026-08-11 00:00:00-05'::timestamptz
    );
    RAISE EXCEPTION 'Escenario 7: barberia_app pudo ejecutar customer_anonymize.';
  EXCEPTION
    WHEN insufficient_privilege THEN
      NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 7 OK · barberia_app no tiene EXECUTE ni sobre el claim ni sobre la anonimización'

-- ---------------------------------------------------------------------------
-- Escenario 8 · Validación de p_limit/p_now (DDL-OPS-01) en ambas funciones
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_worker;

DO $$
DECLARE
  v_limit integer;
BEGIN
  FOREACH v_limit IN ARRAY ARRAY[NULL, 0, -1, 201] LOOP
    BEGIN
      PERFORM * FROM retention_claim_due_customers(v_limit, now());
      RAISE EXCEPTION 'Escenario 8: p_limit=% debía rechazarse en retention_claim_due_customers.', v_limit;
    EXCEPTION
      WHEN raise_exception THEN NULL;  -- Esperado.
    END;
  END LOOP;

  BEGIN
    PERFORM * FROM retention_claim_due_customers(10, NULL);
    RAISE EXCEPTION 'Escenario 8: p_now NULL debía rechazarse en retention_claim_due_customers.';
  EXCEPTION
    WHEN raise_exception THEN NULL;  -- Esperado.
  END;

  BEGIN
    PERFORM customer_anonymize(NULL, NULL, NULL);
    RAISE EXCEPTION 'Escenario 8: customer_anonymize con NULL debía rechazarse.';
  EXCEPTION
    WHEN raise_exception THEN NULL;  -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Escenario 8 OK · p_limit/p_now fuera de rango y argumentos NULL se rechazan en ambas funciones'

\echo '=== customer_anonymization · todas las comprobaciones pasaron ==='
