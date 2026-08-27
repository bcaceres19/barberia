-- Pruebas SQL de HU-060: núcleo persistente de citas sin cruces. Cubren el
-- esquema de las cuatro tablas (customer, appointment, appointment_history,
-- appointment_history_change), sus CHECK, el criterio derivado
-- occupies_schedule, la restricción de exclusión EXCLUDE USING gist contra
-- cruces del mismo barbero, RLS/grants exactos (sin DELETE en customer ni
-- appointment; sin UPDATE ni DELETE en el historial), aislamiento por
-- tenant (RN-TEN-01) y DEC-070 (ON DELETE RESTRICT, nunca CASCADE) a nivel
-- de base de datos. La primitiva transaccional Go (cliente+cita+historial
-- en una sola InTenantTx) se prueba aparte contra PostgreSQL real en
-- apps/api/internal/modules/booking/postgres, incluida la carrera de dos
-- conexiones (CA-060-*, sin dobles).
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/hu060_citas.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu060_citas.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- y `RESET ROLE` hacia un rol con privilegios administrativos (mismo
-- criterio que hu040_horario.sql/hu041_excepciones.sql/hu042_bloqueos.sql).
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-060 · pruebas del núcleo de citas ==='

-- ---------------------------------------------------------------------------
-- Esquema mínimo
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'customer') THEN
    RAISE EXCEPTION 'customer debe existir.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'appointment') THEN
    RAISE EXCEPTION 'appointment debe existir.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'appointment_history') THEN
    RAISE EXCEPTION 'appointment_history debe existir.';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'appointment_history_change') THEN
    RAISE EXCEPTION 'appointment_history_change debe existir.';
  END IF;
END
$$;
\echo 'esquema OK · las cuatro tablas existen'

-- ---------------------------------------------------------------------------
-- CHECK de `customer`: full_name, phone, email, anonymized_at coherente
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
BEGIN
  BEGIN
    INSERT INTO customer (barbershop_id, full_name)
    VALUES ('c17a0001-c17a-c17a-c17a-c17a00010001', '   ');
    RAISE EXCEPTION 'customer_full_name_ck: se aceptó un nombre solo de espacios.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO customer (barbershop_id, full_name, phone)
    VALUES ('c17a0001-c17a-c17a-c17a-c17a00010001', 'Cliente de prueba', 'no-es-telefono');
    RAISE EXCEPTION 'customer_phone_ck: se aceptó un teléfono con formato inválido.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO customer (barbershop_id, full_name, email)
    VALUES ('c17a0001-c17a-c17a-c17a-c17a00010001', 'Cliente de prueba', 'no-es-correo');
    RAISE EXCEPTION 'customer_email_ck: se aceptó un correo con formato inválido.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO customer (barbershop_id, full_name, email)
    VALUES ('c17a0001-c17a-c17a-c17a-c17a00010001', 'Cliente de prueba', 'Con.Mayusculas@Ejemplo.test');
    RAISE EXCEPTION 'customer_email_ck: se aceptó un correo sin forma canónica (mayúsculas).';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO customer (barbershop_id, full_name, phone, anonymized_at)
    VALUES ('c17a0001-c17a-c17a-c17a-c17a00010001', 'Cliente de prueba', '+573001234567', now());
    RAISE EXCEPTION 'customer_anonymized_ck: se aceptó anonymized_at con teléfono todavía presente.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- Forma válida: sin teléfono ni correo (RN-CIT-02, cita manual futura).
  INSERT INTO customer (barbershop_id, full_name)
  VALUES ('c17a0001-c17a-c17a-c17a-c17a00010001', 'Cliente sin contacto de prueba');
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · customer_full_name_ck/phone_ck/email_ck/anonymized_ck rechazan valores inválidos; forma sin contacto se acepta'

-- ---------------------------------------------------------------------------
-- CHECK de `appointment`: status, origin, interval, attendee_name,
-- duración/precio/moneda/nombre de servicio en el snapshot, customer_note,
-- duración coherente con el intervalo, cancellation_reason, resolved_at
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
DECLARE
  v_shop    uuid := 'c17a0001-c17a-c17a-c17a-c17a00010001';
  v_barber  uuid := 'c17a1001-1001-1001-1001-100110011001';
  v_service uuid := 'c17a5001-5001-5001-5001-500150015001';
  v_customer uuid;
BEGIN
  INSERT INTO customer (barbershop_id, full_name)
  VALUES (v_shop, 'Cliente de prueba CHECK')
  RETURNING id INTO v_customer;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'scheduled', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_status_ck: se aceptó un estado fuera de los cinco cerrados.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'walk_in',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_origin_ck: se aceptó un origen fuera de public/manual.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T10:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 0, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_interval_ck: se aceptó ends_at = starts_at (semiabierto).';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, '   ',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_attendee_name_ck: se aceptó un nombre solo de espacios.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 1500, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_duration_snapshot_ck: se aceptó una duración de 1500 minutos.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, -1.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_price_snapshot_ck: se aceptó un precio negativo.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'cop'
    );
    RAISE EXCEPTION 'appointment_currency_snapshot_ck: se aceptó una moneda en minúsculas.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      '', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_service_name_snapshot_ck: se aceptó un nombre de servicio vacío.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
      customer_note
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP', repeat('a', 501)
    );
    RAISE EXCEPTION 'appointment_customer_note_ck: se aceptó una nota de 501 caracteres.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- La duración del snapshot (45) no coincide con el intervalo real (60 min).
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 45, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_duration_matches_snapshot_ck: se aceptó un snapshot de 45 minutos con un intervalo de 60.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- cancellation_reason solo tiene sentido en un estado cancelado.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
      cancellation_reason
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP', 'motivo improcedente'
    );
    RAISE EXCEPTION 'appointment_cancellation_reason_ck: se aceptó un motivo de cancelación en una cita confirmada.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  -- resolved_at: NULL exigido en confirmed, obligatorio (y >= created_at) en cualquier otro estado.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
      resolved_at
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP', now()
    );
    RAISE EXCEPTION 'appointment_resolved_at_ck: se aceptó resolved_at en una cita confirmed.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber, v_service, v_customer, 'Cliente de prueba CHECK',
      '2027-05-10T10:00:00Z', '2027-05-10T11:00:00Z', 'no_show', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_resolved_at_ck: se aceptó no_show sin resolved_at.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · appointment_status_ck/origin_ck/interval_ck/attendee_name_ck/duration_snapshot_ck/price_snapshot_ck/currency_snapshot_ck/service_name_snapshot_ck/customer_note_ck/duration_matches_snapshot_ck/cancellation_reason_ck/resolved_at_ck rechazan valores inválidos'

-- ---------------------------------------------------------------------------
-- Cita base persistida (A1, confirmed, 10:00-11:00, barbero Q-1) para las
-- pruebas de exclusión, occupies_schedule, historial y RESTRICT que siguen
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

INSERT INTO customer (id, barbershop_id, full_name, phone)
VALUES (
  'c17ac001-0001-0001-0001-000100010001', 'c17a0001-c17a-c17a-c17a-c17a00010001',
  'Cliente Q-1 de prueba', '+573001112233'
);

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES (
  'c17aa001-0001-0001-0001-000100010001', 'c17a0001-c17a-c17a-c17a-c17a00010001',
  'c17a1001-1001-1001-1001-100110011001', 'c17a5001-5001-5001-5001-500150015001',
  'c17ac001-0001-0001-0001-000100010001', 'Cliente Q-1 de prueba',
  '2027-05-11T10:00:00Z', '2027-05-11T11:00:00Z', 'confirmed', 'manual',
  'Corte de prueba Q', 60, 20000.00, 'COP'
);

RESET ROLE;
COMMIT;
\echo 'fixture OK · cita base A1 (confirmed, 10:00-11:00, barbero Q-1) persistida'

-- ---------------------------------------------------------------------------
-- Restricción de exclusión (RN-CON-01, RN-CON-03, F-DISP-02): cruce total,
-- parcial por ambos extremos, un minuto, contigüidad válida, barberos
-- distintos y tenants distintos aceptados
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
DECLARE
  v_shop     uuid := 'c17a0001-c17a-c17a-c17a-c17a00010001';
  v_barber1  uuid := 'c17a1001-1001-1001-1001-100110011001';
  v_barber2  uuid := 'c17a1002-1002-1002-1002-100210021002';
  v_service  uuid := 'c17a5001-5001-5001-5001-500150015001';
  v_customer uuid := 'c17ac001-0001-0001-0001-000100010001';
BEGIN
  -- Cruce total: mismo intervalo exacto.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber1, v_service, v_customer, 'Cruce total',
      '2027-05-11T10:00:00Z', '2027-05-11T11:00:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_barber_interval_excl: se aceptó un cruce total con A1.';
  EXCEPTION
    WHEN exclusion_violation THEN NULL;
  END;

  -- Parcial por el inicio: 09:30-10:30 se cruza con 10:00-11:00.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber1, v_service, v_customer, 'Cruce parcial inicio',
      '2027-05-11T09:30:00Z', '2027-05-11T10:30:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_barber_interval_excl: se aceptó un cruce parcial por el inicio.';
  EXCEPTION
    WHEN exclusion_violation THEN NULL;
  END;

  -- Parcial por el final: 10:30-11:30 se cruza con 10:00-11:00.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber1, v_service, v_customer, 'Cruce parcial final',
      '2027-05-11T10:30:00Z', '2027-05-11T11:30:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_barber_interval_excl: se aceptó un cruce parcial por el final.';
  EXCEPTION
    WHEN exclusion_violation THEN NULL;
  END;

  -- Cruce de un minuto: 10:59-11:59 comparte [10:59, 11:00) con A1.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber1, v_service, v_customer, 'Cruce de un minuto',
      '2027-05-11T10:59:00Z', '2027-05-11T11:59:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_barber_interval_excl: se aceptó un cruce de un minuto.';
  EXCEPTION
    WHEN exclusion_violation THEN NULL;
  END;

  -- Contigüidad válida: 11:00-12:00 empieza justo cuando A1 termina
  -- (semiabierto [inicio, fin), RN-DIS-05). Debe aceptarse.
  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Contigua',
    '2027-05-11T11:00:00Z', '2027-05-11T12:00:00Z', 'confirmed', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP'
  );

  -- Mismo intervalo exacto, pero otro barbero de la misma barbería: la
  -- exclusión es por (barbershop_id, barber_id), no solo por barbershop_id.
  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
  ) VALUES (
    v_shop, v_barber2, v_service, v_customer, 'Otro barbero, mismo horario',
    '2027-05-11T10:00:00Z', '2027-05-11T11:00:00Z', 'confirmed', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP'
  );

  -- Cancelada: mismo barbero, mismo intervalo exacto que A1. occupies_schedule
  -- es false, así que no debe entrar en la exclusión.
  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
    resolved_at
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Cancelada solapada',
    '2027-05-11T10:00:00Z', '2027-05-11T11:00:00Z', 'cancelled_by_barber', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP', now()
  );
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'EXCLUDE OK · cruce total/parcial/un minuto rechazados; contigüidad, otro barbero y cita cancelada solapada aceptados'

-- ---------------------------------------------------------------------------
-- occupies_schedule (CA-060-03): confirmed/completed/no_show = true;
-- cancelled_by_customer/cancelled_by_barber = false. Un estado que ocupa
-- agenda (completed) también entra en la exclusión.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
DECLARE
  v_shop     uuid := 'c17a0001-c17a-c17a-c17a-c17a00010001';
  v_barber1  uuid := 'c17a1001-1001-1001-1001-100110011001';
  v_service  uuid := 'c17a5001-5001-5001-5001-500150015001';
  v_customer uuid := 'c17ac001-0001-0001-0001-000100010001';
  v_completed_id uuid;
  v_no_show_id   uuid;
  v_cancelled_customer_id uuid;
  v_occupies boolean;
BEGIN
  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
    resolved_at
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Atendida de prueba',
    '2027-05-12T14:00:00Z', '2027-05-12T15:00:00Z', 'completed', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP', now()
  ) RETURNING id INTO v_completed_id;

  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
    resolved_at
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Inasistencia de prueba',
    '2027-05-12T16:00:00Z', '2027-05-12T17:00:00Z', 'no_show', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP', now()
  ) RETURNING id INTO v_no_show_id;

  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot,
    resolved_at
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Cancelada por cliente de prueba',
    '2027-05-12T18:00:00Z', '2027-05-12T19:00:00Z', 'cancelled_by_customer', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP', now()
  ) RETURNING id INTO v_cancelled_customer_id;

  SELECT occupies_schedule INTO v_occupies FROM appointment WHERE id = v_completed_id;
  IF v_occupies IS DISTINCT FROM true THEN
    RAISE EXCEPTION 'occupies_schedule: completed debe ser true.';
  END IF;

  SELECT occupies_schedule INTO v_occupies FROM appointment WHERE id = v_no_show_id;
  IF v_occupies IS DISTINCT FROM true THEN
    RAISE EXCEPTION 'occupies_schedule: no_show debe ser true.';
  END IF;

  SELECT occupies_schedule INTO v_occupies FROM appointment WHERE id = v_cancelled_customer_id;
  IF v_occupies IS DISTINCT FROM false THEN
    RAISE EXCEPTION 'occupies_schedule: cancelled_by_customer debe ser false.';
  END IF;

  -- Un estado terminal que ocupa agenda (completed) también entra en la
  -- exclusión: crear otra cita encima de una ya atendida se rechaza.
  BEGIN
    INSERT INTO appointment (
      barbershop_id, barber_id, service_id, customer_id, attendee_name,
      starts_at, ends_at, status, origin,
      service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
    ) VALUES (
      v_shop, v_barber1, v_service, v_customer, 'Encima de una atendida',
      '2027-05-12T14:30:00Z', '2027-05-12T15:30:00Z', 'confirmed', 'manual',
      'Corte de prueba Q', 60, 20000.00, 'COP'
    );
    RAISE EXCEPTION 'appointment_barber_interval_excl: se aceptó una cita encima de una completed.';
  EXCEPTION
    WHEN exclusion_violation THEN NULL;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'occupies_schedule OK · completed/no_show=true, cancelled_by_customer=false; completed sigue ocupando la exclusión'

-- ---------------------------------------------------------------------------
-- Medianoche y DST: la duración es la diferencia real de instantes
-- absolutos (timestamptz), no una suma calendárica dependiente de la zona
-- de sesión (CA-060-04)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';
SET LOCAL TIME ZONE 'America/New_York';

DO $$
DECLARE
  v_shop     uuid := 'c17a0001-c17a-c17a-c17a-c17a00010001';
  v_barber1  uuid := 'c17a1001-1001-1001-1001-100110011001';
  v_service  uuid := 'c17a5001-5001-5001-5001-500150015001';
  v_customer uuid := 'c17ac001-0001-0001-0001-000100010001';
BEGIN
  -- Cruza medianoche UTC: 2027-05-20T23:30:00Z a 2027-05-21T00:30:00Z, 60 min.
  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Cruza medianoche',
    '2027-05-20T23:30:00+00', '2027-05-21T00:30:00+00', 'confirmed', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP'
  );

  -- Instantes UTC fijos alrededor del cambio de horario de verano de EE.UU.
  -- (segundo domingo de marzo). La sesión está en America/New_York
  -- (SET LOCAL TIME ZONE arriba); si el CHECK dependiera de una resta de
  -- horas locales en vez de EXTRACT(EPOCH) sobre timestamptz, fallaría o
  -- exigiría un valor de duración distinto justo este día.
  INSERT INTO appointment (
    barbershop_id, barber_id, service_id, customer_id, attendee_name,
    starts_at, ends_at, status, origin,
    service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
  ) VALUES (
    v_shop, v_barber1, v_service, v_customer, 'Cambio de horario DST',
    '2027-03-14T05:30:00+00', '2027-03-14T06:30:00+00', 'confirmed', 'manual',
    'Corte de prueba Q', 60, 20000.00, 'COP'
  );
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'DEC-007 OK · duración cruzando medianoche y en fecha de cambio de horario de verano, con zona de sesión distinta de UTC'

-- ---------------------------------------------------------------------------
-- CHECK de `appointment_history`: event_type, actor_type, forma del actor,
-- reason; INSERT permitido y UPDATE/DELETE rechazados (RN-HIS-02, DEC-014)
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
DECLARE
  v_shop  uuid := 'c17a0001-c17a-c17a-c17a-c17a00010001';
  v_appt  uuid := 'c17aa001-0001-0001-0001-000100010001';
  v_staff uuid := 'c17a9001-9001-9001-9001-900190019001';
BEGIN
  BEGIN
    INSERT INTO appointment_history (barbershop_id, appointment_id, event_type, actor_type)
    VALUES (v_shop, v_appt, 'appointment_booked', 'system');
    RAISE EXCEPTION 'appointment_history_event_type_ck: se aceptó un event_type fuera del vocabulario cerrado.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment_history (barbershop_id, appointment_id, event_type, actor_type)
    VALUES (v_shop, v_appt, 'appointment_created', 'owner');
    RAISE EXCEPTION 'appointment_history_actor_type_ck: se aceptó un actor_type fuera de staff/customer/system.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment_history (barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id)
    VALUES (v_shop, v_appt, 'appointment_created', 'system', v_staff);
    RAISE EXCEPTION 'appointment_history_actor_shape_ck: se aceptó actor_type=system con actor_staff_user_id presente.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment_history (barbershop_id, appointment_id, event_type, actor_type)
    VALUES (v_shop, v_appt, 'appointment_created', 'staff');
    RAISE EXCEPTION 'appointment_history_actor_shape_ck: se aceptó actor_type=staff sin actor_staff_user_id.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;

  BEGIN
    INSERT INTO appointment_history (barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id)
    VALUES (v_shop, v_appt, 'appointment_status_corrected', 'staff', v_staff);
    RAISE EXCEPTION 'appointment_history_reason_ck: se aceptó appointment_status_corrected sin motivo.';
  EXCEPTION
    WHEN check_violation THEN NULL;
  END;
END
$$;

-- Fila válida: appointment_created, actor staff. Se conserva para las
-- pruebas de RESTRICT/limpieza más abajo.
INSERT INTO appointment_history (id, barbershop_id, appointment_id, event_type, actor_type, actor_staff_user_id)
VALUES (
  'c17a4001-0001-0001-0001-000100010001', 'c17a0001-c17a-c17a-c17a-c17a00010001',
  'c17aa001-0001-0001-0001-000100010001', 'appointment_created', 'staff',
  'c17a9001-9001-9001-9001-900190019001'
);

INSERT INTO appointment_history_change (barbershop_id, history_id, field_name, new_value)
VALUES (
  'c17a0001-c17a-c17a-c17a-c17a00010001', 'c17a4001-0001-0001-0001-000100010001',
  'status', 'confirmed'
);

DO $$
BEGIN
  BEGIN
    UPDATE appointment_history SET reason = 'intento de edición' WHERE id = 'c17a4001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'RN-HIS-02: se permitió UPDATE sobre appointment_history con el rol de aplicación.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;
  END;

  BEGIN
    DELETE FROM appointment_history WHERE id = 'c17a4001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'RN-HIS-02: se permitió DELETE sobre appointment_history con el rol de aplicación.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;
  END;

  BEGIN
    UPDATE appointment_history_change SET new_value = 'otro' WHERE history_id = 'c17a4001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'RN-HIS-02: se permitió UPDATE sobre appointment_history_change con el rol de aplicación.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;
  END;

  BEGIN
    DELETE FROM appointment_history_change WHERE history_id = 'c17a4001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'RN-HIS-02: se permitió DELETE sobre appointment_history_change con el rol de aplicación.';
  EXCEPTION
    WHEN insufficient_privilege THEN NULL;
  END;
END
$$;

RESET ROLE;
COMMIT;
\echo 'CHECK/append-only OK · appointment_history rechaza event_type/actor_type/forma de actor/reason inválidos; UPDATE y DELETE rechazados en ambas tablas'

-- ---------------------------------------------------------------------------
-- RLS forzada, sin BYPASSRLS (las cuatro tablas)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_relrowsecurity boolean;
  v_relforcerowsecurity boolean;
BEGIN
  FOR v_relrowsecurity, v_relforcerowsecurity IN
    SELECT relrowsecurity, relforcerowsecurity FROM pg_class
    WHERE relname IN ('customer', 'appointment', 'appointment_history', 'appointment_history_change')
      AND relnamespace = 'public'::regnamespace
  LOOP
    IF NOT v_relrowsecurity OR NOT v_relforcerowsecurity THEN
      RAISE EXCEPTION 'las cuatro tablas de HU-060 deben tener RLS habilitada Y forzada.';
    END IF;
  END LOOP;

  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app' AND rolbypassrls) THEN
    RAISE EXCEPTION 'barberia_app no puede tener BYPASSRLS.';
  END IF;
END
$$;
\echo 'RLS OK · las cuatro tablas tienen RLS habilitada y forzada, barberia_app sin BYPASSRLS'

-- ---------------------------------------------------------------------------
-- Grants exactos de barberia_app: sin DELETE en customer/appointment;
-- sin UPDATE ni DELETE en el historial (RN-DAT-03, RN-HIS-02)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_missing text;
  v_table text;
BEGIN
  FOREACH v_table IN ARRAY ARRAY['customer', 'appointment']
  LOOP
    SELECT string_agg(expected, ', ') INTO v_missing
    FROM unnest(ARRAY['SELECT', 'INSERT', 'UPDATE']) AS expected
    WHERE NOT EXISTS (
      SELECT 1 FROM information_schema.table_privileges
      WHERE table_schema = 'public' AND table_name = v_table
        AND grantee = 'barberia_app' AND privilege_type = expected
    );
    IF v_missing IS NOT NULL THEN
      RAISE EXCEPTION 'barberia_app debe tener % sobre %.', v_missing, v_table;
    END IF;
    IF EXISTS (
      SELECT 1 FROM information_schema.table_privileges
      WHERE table_schema = 'public' AND table_name = v_table
        AND grantee = 'barberia_app' AND privilege_type = 'DELETE'
    ) THEN
      RAISE EXCEPTION 'RN-DAT-03/RN-HIS-02: barberia_app NO debe tener DELETE sobre %.', v_table;
    END IF;
  END LOOP;

  FOREACH v_table IN ARRAY ARRAY['appointment_history', 'appointment_history_change']
  LOOP
    SELECT string_agg(expected, ', ') INTO v_missing
    FROM unnest(ARRAY['SELECT', 'INSERT']) AS expected
    WHERE NOT EXISTS (
      SELECT 1 FROM information_schema.table_privileges
      WHERE table_schema = 'public' AND table_name = v_table
        AND grantee = 'barberia_app' AND privilege_type = expected
    );
    IF v_missing IS NOT NULL THEN
      RAISE EXCEPTION 'barberia_app debe tener % sobre %.', v_missing, v_table;
    END IF;
    IF EXISTS (
      SELECT 1 FROM information_schema.table_privileges
      WHERE table_schema = 'public' AND table_name = v_table
        AND grantee = 'barberia_app' AND privilege_type IN ('UPDATE', 'DELETE')
    ) THEN
      RAISE EXCEPTION 'RN-HIS-02: barberia_app NO debe tener UPDATE ni DELETE sobre %.', v_table;
    END IF;
  END LOOP;
END
$$;
\echo 'grants OK · sin DELETE en customer/appointment; sin UPDATE ni DELETE en appointment_history/appointment_history_change'

-- ---------------------------------------------------------------------------
-- Aislamiento de tenant (RN-TEN-01): el contexto de Q no ve ni edita la
-- cita de R
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0002-c17a-c17a-c17a-c17a00020002';

INSERT INTO customer (id, barbershop_id, full_name)
VALUES (
  'c17ac002-0002-0002-0002-000200020002', 'c17a0002-c17a-c17a-c17a-c17a00020002',
  'Cliente R-1 de prueba'
);

INSERT INTO appointment (
  id, barbershop_id, barber_id, service_id, customer_id, attendee_name,
  starts_at, ends_at, status, origin,
  service_name_snapshot, duration_minutes_snapshot, price_amount_snapshot, price_currency_snapshot
) VALUES (
  'c17aa010-0010-0010-0010-001000100010', 'c17a0002-c17a-c17a-c17a-c17a00020002',
  'c17a2001-2001-2001-2001-200120012001', 'c17a5002-5002-5002-5002-500250025002',
  'c17ac002-0002-0002-0002-000200020002', 'Cliente R-1 de prueba',
  '2027-05-11T10:00:00Z', '2027-05-11T10:45:00Z', 'confirmed', 'manual',
  'Corte de prueba R', 45, 25000.00, 'COP'
);

RESET ROLE;
COMMIT;

BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
DECLARE
  v_appt_r_id uuid := 'c17aa010-0010-0010-0010-001000100010';
  v_count integer;
  v_updated integer;
BEGIN
  SELECT count(*) INTO v_count FROM appointment WHERE id = v_appt_r_id;
  IF v_count <> 0 THEN
    RAISE EXCEPTION 'RN-TEN-01: el contexto de Q puede ver la cita de R por su id real.';
  END IF;

  UPDATE appointment SET attendee_name = 'intento cruzado' WHERE id = v_appt_r_id;
  GET DIAGNOSTICS v_updated = ROW_COUNT;
  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'RN-TEN-01: el contexto de Q pudo editar la cita de R (% filas).', v_updated;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'RN-TEN-01 OK · el contexto de Q no ve ni edita la cita de R'

-- ---------------------------------------------------------------------------
-- DEC-070 · Las FK hacia barber/service/customer/staff_user/appointment
-- usan ON DELETE RESTRICT, no CASCADE
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  BEGIN
    DELETE FROM barber WHERE id = 'c17a1001-1001-1001-1001-100110011001';
    RAISE EXCEPTION 'DEC-070: se permitió borrar un barber con una appointment asociada (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  BEGIN
    DELETE FROM service WHERE id = 'c17a5001-5001-5001-5001-500150015001';
    RAISE EXCEPTION 'DEC-070: se permitió borrar un service con una appointment asociada (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  BEGIN
    DELETE FROM customer WHERE id = 'c17ac001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'DEC-070: se permitió borrar un customer con una appointment asociada (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  BEGIN
    DELETE FROM staff_user WHERE id = 'c17a9001-9001-9001-9001-900190019001';
    RAISE EXCEPTION 'DEC-070: se permitió borrar un staff_user referenciado por appointment_history.actor_staff_user_id.';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  BEGIN
    DELETE FROM appointment WHERE id = 'c17aa001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'DEC-070: se permitió borrar una appointment con appointment_history asociado (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;

  BEGIN
    DELETE FROM appointment_history WHERE id = 'c17a4001-0001-0001-0001-000100010001';
    RAISE EXCEPTION 'DEC-070: se permitió borrar un appointment_history con appointment_history_change asociado (CASCADE en vez de RESTRICT).';
  EXCEPTION
    WHEN foreign_key_violation THEN NULL;
  END;
END
$$;
\echo 'DEC-070 OK · ninguna FK de HU-060 permite un borrado en cascada; todas rechazan (ON DELETE RESTRICT)'

-- ---------------------------------------------------------------------------
-- Trigger de updated_at sobre customer y appointment
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = 'c17a0001-c17a-c17a-c17a-c17a00010001';

DO $$
DECLARE
  v_updated_before timestamptz;
  v_updated_after timestamptz;
BEGIN
  SELECT updated_at INTO v_updated_before FROM customer WHERE id = 'c17ac001-0001-0001-0001-000100010001';
  PERFORM pg_sleep(0.01);
  UPDATE customer SET full_name = 'Cliente Q-1 actualizado' WHERE id = 'c17ac001-0001-0001-0001-000100010001'
  RETURNING updated_at INTO v_updated_after;
  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'customer_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;

  SELECT updated_at INTO v_updated_before FROM appointment WHERE id = 'c17aa001-0001-0001-0001-000100010001';
  PERFORM pg_sleep(0.01);
  UPDATE appointment SET customer_note = 'Nota de prueba' WHERE id = 'c17aa001-0001-0001-0001-000100010001'
  RETURNING updated_at INTO v_updated_after;
  IF v_updated_after <= v_updated_before THEN
    RAISE EXCEPTION 'appointment_set_updated_at: updated_at no avanzó tras el UPDATE (antes=%, después=%).',
      v_updated_before, v_updated_after;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'trigger OK · customer_set_updated_at/appointment_set_updated_at avanzan updated_at en cada UPDATE'

-- ---------------------------------------------------------------------------
-- Limpieza de las filas persistidas por este archivo (A1, R1, historial y
-- sus clientes). Orden que respeta ON DELETE RESTRICT: cambios de
-- historial -> historial -> citas -> clientes. Sin SET ROLE barberia_app:
-- ni customer ni appointment ni el historial le conceden DELETE a
-- propósito (RN-DAT-03/RN-HIS-02); limpiar como el rol administrativo
-- conectado (superusuario/migrador en este arnés) es correcto aquí, mismo
-- criterio que hu042_bloqueos.sql.
-- ---------------------------------------------------------------------------
BEGIN;
DELETE FROM appointment_history_change WHERE history_id = 'c17a4001-0001-0001-0001-000100010001';
DELETE FROM appointment_history WHERE id = 'c17a4001-0001-0001-0001-000100010001';
DELETE FROM appointment WHERE id IN (
  'c17aa001-0001-0001-0001-000100010001', 'c17aa010-0010-0010-0010-001000100010'
);
DELETE FROM customer WHERE id IN (
  'c17ac001-0001-0001-0001-000100010001', 'c17ac002-0002-0002-0002-000200020002'
);
COMMIT;

\echo '=== HU-060 · todas las comprobaciones pasaron ==='
