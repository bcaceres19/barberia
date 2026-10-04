-- Solo datos sintéticos para ver y probar el perfil «barbero individual» (#294,
-- DEC-115): una barbería con UN solo barbero, el perfil `solo`, cuatro servicios
-- (tres ya asignados a él), horario de lunes a sábado y tres turnos hoy en la zona
-- de la barbería. Aplicar con el rol migrador únicamente en una BD local de
-- desarrollo o de pruebas; no usar en producción.
--
-- Cuenta ficticia: mateo.demo@ejemplo.test · contraseña de desarrollo
-- Nava-Demo-2026! (hash argon2id real, mismo valor que las cuentas demo locales).
--
-- Es repetible: los turnos de hoy se borran y se vuelven a crear con la fecha
-- actual; el resto usa ON CONFLICT DO NOTHING con identificadores fijos.
-- Ningún dato corresponde a una persona real (AGENTS.md).
BEGIN;

INSERT INTO barbershop (id, name, timezone, panel_profile)
VALUES ('294a0001-0000-4000-8000-000000000001', 'Mateo · Barbero', 'America/Bogota', 'solo')
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_user (id, barbershop_id, email, full_name, is_active)
VALUES ('294a0001-0000-4000-8000-000000000002', '294a0001-0000-4000-8000-000000000001',
        'mateo.demo@ejemplo.test', 'Mateo de prueba', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO staff_credential (staff_user_id, barbershop_id, password_hash, password_algorithm)
VALUES ('294a0001-0000-4000-8000-000000000002', '294a0001-0000-4000-8000-000000000001',
        '$argon2id$v=19$m=19456,t=2,p=1$85G7e0c0elS58Y0r6XwbTQ$BDLtYxmfV8IV6H5Vvuj0djdhbHmnCcpBq4AECbiHga4',
        'argon2id')
ON CONFLICT (staff_user_id) DO NOTHING;

INSERT INTO barber (id, barbershop_id, full_name)
VALUES ('294b0001-0000-4000-8000-000000000001', '294a0001-0000-4000-8000-000000000001', 'Mateo de prueba')
ON CONFLICT (id) DO NOTHING;

INSERT INTO service (id, barbershop_id, name, description, duration_minutes, price_amount) VALUES
  ('294c0001-0000-4000-8000-000000000001', '294a0001-0000-4000-8000-000000000001',
   'Corte clásico', 'Corte a tijera o máquina con acabado a navaja.', 40, 30000),
  ('294c0001-0000-4000-8000-000000000002', '294a0001-0000-4000-8000-000000000001',
   'Barba perfilada', 'Perfilado y toalla caliente.', 30, 20000),
  ('294c0001-0000-4000-8000-000000000003', '294a0001-0000-4000-8000-000000000001',
   'Corte y barba', NULL, 60, 45000),
  ('294c0001-0000-4000-8000-000000000004', '294a0001-0000-4000-8000-000000000001',
   'Diseño con navaja', NULL, 20, 15000)
ON CONFLICT (id) DO NOTHING;

-- Tres de los cuatro servicios quedan ofrecidos; «Diseño con navaja» queda sin
-- asignar a propósito para que se vea el control «Lo ofrezco» apagado.
INSERT INTO barber_service (barbershop_id, barber_id, service_id) VALUES
  ('294a0001-0000-4000-8000-000000000001', '294b0001-0000-4000-8000-000000000001', '294c0001-0000-4000-8000-000000000001'),
  ('294a0001-0000-4000-8000-000000000001', '294b0001-0000-4000-8000-000000000001', '294c0001-0000-4000-8000-000000000002'),
  ('294a0001-0000-4000-8000-000000000001', '294b0001-0000-4000-8000-000000000001', '294c0001-0000-4000-8000-000000000003')
ON CONFLICT DO NOTHING;

-- Lunes a viernes 09:00–18:00 y sábado 09:00–14:00.
INSERT INTO working_hour (barbershop_id, barber_id, iso_weekday, starts_time, duration_minutes)
SELECT '294a0001-0000-4000-8000-000000000001', '294b0001-0000-4000-8000-000000000001', d,
       time '09:00', CASE WHEN d = 6 THEN 300 ELSE 540 END
  FROM generate_series(1, 6) AS d
ON CONFLICT DO NOTHING;

INSERT INTO customer (id, barbershop_id, full_name, phone) VALUES
  ('294d0001-0000-4000-8000-000000000001', '294a0001-0000-4000-8000-000000000001', 'Cliente de prueba Uno',  '+573000000011'),
  ('294d0001-0000-4000-8000-000000000002', '294a0001-0000-4000-8000-000000000001', 'Cliente de prueba Dos',  '+573000000012'),
  ('294d0001-0000-4000-8000-000000000003', '294a0001-0000-4000-8000-000000000001', 'Cliente de prueba Tres', '+573000000013')
ON CONFLICT (id) DO NOTHING;

-- Turnos de hoy en la zona de la barbería, recreados en cada aplicación.
DELETE FROM appointment_history WHERE barbershop_id = '294a0001-0000-4000-8000-000000000001';
DELETE FROM appointment WHERE barbershop_id = '294a0001-0000-4000-8000-000000000001';

WITH day AS (
  SELECT (now() AT TIME ZONE 'America/Bogota')::date AS d
), slot(id, customer, service, hh, mm, name, duration, price, label) AS (
  VALUES
    ('294e0001-0000-4000-8000-000000000001'::uuid, '294d0001-0000-4000-8000-000000000001'::uuid, '294c0001-0000-4000-8000-000000000001'::uuid, 10, 0,  'Corte clásico',    40, 30000, 'Cliente de prueba Uno'),
    ('294e0001-0000-4000-8000-000000000002'::uuid, '294d0001-0000-4000-8000-000000000002'::uuid, '294c0001-0000-4000-8000-000000000003'::uuid, 11, 30, 'Corte y barba',    60, 45000, 'Cliente de prueba Dos'),
    ('294e0001-0000-4000-8000-000000000003'::uuid, '294d0001-0000-4000-8000-000000000003'::uuid, '294c0001-0000-4000-8000-000000000002'::uuid, 15, 0,  'Barba perfilada',  30, 20000, 'Cliente de prueba Tres')
)
INSERT INTO appointment (id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
                         starts_at, ends_at, status, origin, service_name_snapshot,
                         duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot)
SELECT slot.id, '294a0001-0000-4000-8000-000000000001', '294b0001-0000-4000-8000-000000000001',
       slot.service, slot.customer, slot.label,
       ((day.d + make_time(slot.hh, slot.mm, 0)) AT TIME ZONE 'America/Bogota'),
       ((day.d + make_time(slot.hh, slot.mm, 0) + make_interval(mins => slot.duration)) AT TIME ZONE 'America/Bogota'),
       'confirmed', 'manual', slot.name, slot.duration, slot.price, 'COP'
  FROM slot, day;

COMMIT;
