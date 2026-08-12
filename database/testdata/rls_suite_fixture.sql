-- Fixture de dos tenants con al menos una fila por tabla protegida por RLS,
-- para database/tests/rls_suite.sql (issue #7, DDL-RLS-01).
--
-- REQUIERE modelo-fisico-referencia.sql cargado ADEMÁS de las cinco
-- migraciones de database/migrations/: inserta en staff_credential,
-- staff_session y el resto de las 23 tablas restantes, ninguna de las
-- cuales existe en el esquema realmente aplicado (database/README.md las
-- lista como diseño de referencia, no migración). Ejecutar este fixture
-- contra una base con solo las migraciones aplicadas falla con
-- `relation "staff_credential" does not exist`. Esta suite certifica el
-- diseño del modelo de referencia, NO el esquema desplegable hoy.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/rls_suite_fixture.sql
--
-- Se ejecuta con el rol migrador (política administrativa sobre las tablas
-- con RLS forzada). Identificadores fijos e idempotente (ON CONFLICT DO
-- NOTHING). No cubre working_hour_override_segment, time_block_series_date
-- ni time_block_series_exception con una fila propia: comparten política y
-- patrón con su tabla padre/hermana ya cubierta, y la suite genérica de
-- estructura (RLS habilitada/forzada, política admin, sin BYPASSRLS) sí las
-- alcanza a todas por catálogo, sin necesitar datos.
--
-- Ningún dato de este archivo corresponde a una persona real (AGENTS.md).

BEGIN;

-- ---------------------------------------------------------------------------
-- Barbería A (11111111-...)
-- ---------------------------------------------------------------------------

INSERT INTO staff_credential (staff_user_id, barbershop_id, password_hash) VALUES
  ('aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '11111111-1111-1111-1111-111111111111',
   repeat('a', 64))
ON CONFLICT (staff_user_id) DO NOTHING;

INSERT INTO staff_session (id, barbershop_id, staff_user_id, token_hash, expires_at) VALUES
  ('50550000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   repeat('a', 64), '2030-01-01 00:00:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_recovery_code (id, barbershop_id, staff_user_id, code_hash, expires_at) VALUES
  ('c0de0000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2',
   repeat('a', 64), '2030-01-01 00:00:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('ba5b0000-0000-0000-0000-00000000000a', '11111111-1111-1111-1111-111111111111', 'Barbero RLS A')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('5e5b0000-0000-0000-0000-00000000000a', '11111111-1111-1111-1111-111111111111',
   'Servicio RLS A', 30, 20000.00, 'COP')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES
  ('11111111-1111-1111-1111-111111111111',
   'ba5b0000-0000-0000-0000-00000000000a', '5e5b0000-0000-0000-0000-00000000000a')
ON CONFLICT DO NOTHING;

INSERT INTO working_hour (id, barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes) VALUES
  ('90570000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'ba5b0000-0000-0000-0000-00000000000a',
   1, '09:00', 480)
ON CONFLICT (id) DO NOTHING;

INSERT INTO working_hour_override (id, barbershop_id, barber_id, effective_date, is_closed) VALUES
  ('090e0000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'ba5b0000-0000-0000-0000-00000000000a',
   '2030-01-01', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO time_block_series (id, barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from) VALUES
  ('5e520000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'ba5b0000-0000-0000-0000-00000000000a',
   'lunch', 'weekly', 1, '12:00', 60, '2030-01-01')
ON CONFLICT (id) DO NOTHING;

INSERT INTO time_block (id, barbershop_id, barber_id, block_type, starts_at, ends_at) VALUES
  ('b10c0000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'ba5b0000-0000-0000-0000-00000000000a',
   'break', '2030-01-01 15:00:00-05', '2030-01-01 15:15:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO customer (id, barbershop_id, full_name) VALUES
  ('c0570000-0000-0000-0000-00000000000a'::uuid, '11111111-1111-1111-1111-111111111111', 'Cliente RLS A')
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES (
  'a99a0000-0000-0000-0000-00000000000a',
  '11111111-1111-1111-1111-111111111111', 'ba5b0000-0000-0000-0000-00000000000a',
  '5e5b0000-0000-0000-0000-00000000000a', 'c0570000-0000-0000-0000-00000000000a'::uuid,
  'Cliente RLS A', '2030-01-05 10:00:00-05', '2030-01-05 10:30:00-05', 'manual',
  'Servicio RLS A', 30, 20000.00, 'COP'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment_history (id, barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id) VALUES
  ('7157000a-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'a99a0000-0000-0000-0000-00000000000a',
   'appointment_created', 'staff', 'aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2')
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, new_value) VALUES
  ('11111111-1111-1111-1111-111111111111', '7157000a-0000-0000-0000-00000000000a', 'status', 'confirmed')
ON CONFLICT DO NOTHING;

-- token_hash es único GLOBAL (H.5): 'c'/'d' en vez de 'a'/'b' a propósito,
-- para no colisionar con los valores que ya usa
-- tests/customer_anonymization.sql en sus propios escenarios.
INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash) VALUES
  ('11111111-1111-1111-1111-111111111111', 'a99a0000-0000-0000-0000-00000000000a', repeat('c', 64))
ON CONFLICT DO NOTHING;

INSERT INTO notification_channel_setting (barbershop_id, event_type, channel) VALUES
  ('11111111-1111-1111-1111-111111111111', 'confirmation', 'email')
ON CONFLICT DO NOTHING;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for) VALUES
  ('501e0000-0000-0000-0000-00000000000a',
   '11111111-1111-1111-1111-111111111111', 'a99a0000-0000-0000-0000-00000000000a',
   'reminder', 'email', '2030-01-05 09:30:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO notification_attempt (barbershop_id, notification_schedule_id, attempt_number, result, template_key, template_version) VALUES
  ('11111111-1111-1111-1111-111111111111', '501e0000-0000-0000-0000-00000000000a',
   1, 'delivered', 'reminder_email', 'v1')
ON CONFLICT DO NOTHING;

INSERT INTO idempotency_record (barbershop_id, idempotency_key, operation, request_fingerprint, expires_at) VALUES
  ('11111111-1111-1111-1111-111111111111', 'rls-suite-key-a', 'create_appointment',
   repeat('a', 64), '2030-01-01 00:00:00-05')
ON CONFLICT (barbershop_id, idempotency_key) DO NOTHING;

-- ---------------------------------------------------------------------------
-- Barbería B (22222222-...) — mismas tablas, mismo patrón
-- ---------------------------------------------------------------------------

INSERT INTO staff_credential (staff_user_id, barbershop_id, password_hash) VALUES
  ('bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2', '22222222-2222-2222-2222-222222222222',
   repeat('b', 64))
ON CONFLICT (staff_user_id) DO NOTHING;

INSERT INTO staff_session (id, barbershop_id, staff_user_id, token_hash, expires_at) VALUES
  ('50550000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2',
   repeat('b', 64), '2030-01-01 00:00:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_recovery_code (id, barbershop_id, staff_user_id, code_hash, expires_at) VALUES
  ('c0de0000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2',
   repeat('b', 64), '2030-01-01 00:00:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name) VALUES
  ('ba5b0000-0000-0000-0000-00000000000b', '22222222-2222-2222-2222-222222222222', 'Barbero RLS B')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, duration_minutes, price_amount, price_currency) VALUES
  ('5e5b0000-0000-0000-0000-00000000000b', '22222222-2222-2222-2222-222222222222',
   'Servicio RLS B', 30, 20000.00, 'COP')
ON CONFLICT (id) DO NOTHING;

INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES
  ('22222222-2222-2222-2222-222222222222',
   'ba5b0000-0000-0000-0000-00000000000b', '5e5b0000-0000-0000-0000-00000000000b')
ON CONFLICT DO NOTHING;

INSERT INTO working_hour (id, barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes) VALUES
  ('90570000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'ba5b0000-0000-0000-0000-00000000000b',
   1, '09:00', 480)
ON CONFLICT (id) DO NOTHING;

INSERT INTO working_hour_override (id, barbershop_id, barber_id, effective_date, is_closed) VALUES
  ('090e0000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'ba5b0000-0000-0000-0000-00000000000b',
   '2030-01-01', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO time_block_series (id, barbershop_id, barber_id, block_type, recurrence_kind, iso_weekday, starts_time, duration_minutes, effective_from) VALUES
  ('5e520000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'ba5b0000-0000-0000-0000-00000000000b',
   'lunch', 'weekly', 1, '12:00', 60, '2030-01-01')
ON CONFLICT (id) DO NOTHING;

INSERT INTO time_block (id, barbershop_id, barber_id, block_type, starts_at, ends_at) VALUES
  ('b10c0000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'ba5b0000-0000-0000-0000-00000000000b',
   'break', '2030-01-01 15:00:00-05', '2030-01-01 15:15:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO customer (id, barbershop_id, full_name) VALUES
  ('c0570000-0000-0000-0000-00000000000b', '22222222-2222-2222-2222-222222222222', 'Cliente RLS B')
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES (
  'a99a0000-0000-0000-0000-00000000000b',
  '22222222-2222-2222-2222-222222222222', 'ba5b0000-0000-0000-0000-00000000000b',
  '5e5b0000-0000-0000-0000-00000000000b', 'c0570000-0000-0000-0000-00000000000b',
  'Cliente RLS B', '2030-01-05 10:00:00-05', '2030-01-05 10:30:00-05', 'manual',
  'Servicio RLS B', 30, 20000.00, 'COP'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment_history (id, barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id) VALUES
  ('71570000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'a99a0000-0000-0000-0000-00000000000b',
   'appointment_created', 'staff', 'bbbbbbb2-bbbb-bbbb-bbbb-bbbbbbbbbbb2')
ON CONFLICT (id) DO NOTHING;

INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, new_value) VALUES
  ('22222222-2222-2222-2222-222222222222', '71570000-0000-0000-0000-00000000000b', 'status', 'confirmed')
ON CONFLICT DO NOTHING;

INSERT INTO appointment_access_token (barbershop_id, appointment_id, token_hash) VALUES
  ('22222222-2222-2222-2222-222222222222', 'a99a0000-0000-0000-0000-00000000000b', repeat('d', 64))
ON CONFLICT DO NOTHING;

INSERT INTO notification_channel_setting (barbershop_id, event_type, channel) VALUES
  ('22222222-2222-2222-2222-222222222222', 'confirmation', 'email')
ON CONFLICT DO NOTHING;

INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for) VALUES
  ('501e0000-0000-0000-0000-00000000000b',
   '22222222-2222-2222-2222-222222222222', 'a99a0000-0000-0000-0000-00000000000b',
   'reminder', 'email', '2030-01-05 09:30:00-05')
ON CONFLICT (id) DO NOTHING;

INSERT INTO notification_attempt (barbershop_id, notification_schedule_id, attempt_number, result, template_key, template_version) VALUES
  ('22222222-2222-2222-2222-222222222222', '501e0000-0000-0000-0000-00000000000b',
   1, 'delivered', 'reminder_email', 'v1')
ON CONFLICT DO NOTHING;

INSERT INTO idempotency_record (barbershop_id, idempotency_key, operation, request_fingerprint, expires_at) VALUES
  ('22222222-2222-2222-2222-222222222222', 'rls-suite-key-b', 'create_appointment',
   repeat('b', 64), '2030-01-01 00:00:00-05')
ON CONFLICT (barbershop_id, idempotency_key) DO NOTHING;

COMMIT;
