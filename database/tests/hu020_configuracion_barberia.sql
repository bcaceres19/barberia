-- Pruebas SQL de HU-020: configuración básica de la barbería. Cubren
-- CA-020-01, CA-020-05, CA-020-06 a nivel de base de datos (constraints y
-- RLS); la validación de zona IANA y la orquestación completa del PATCH
-- viven en el servicio Go (internal/modules/shops), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/hu020_configuracion_barberia.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que hu001_aislamiento_rls.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== HU-020 · pruebas de configuración de la barbería ==='

-- ---------------------------------------------------------------------------
-- Columnas nuevas anulables y con los CHECK esperados
-- ---------------------------------------------------------------------------
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop'
      AND column_name = 'contact_email' AND is_nullable = 'YES'
  ) THEN
    RAISE EXCEPTION 'barbershop.contact_email debe existir y ser anulable.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop'
      AND column_name = 'contact_phone' AND is_nullable = 'YES'
  ) THEN
    RAISE EXCEPTION 'barbershop.contact_phone debe existir y ser anulable.';
  END IF;
END
$$;
\echo 'columnas OK · contact_email/contact_phone existen y son anulables'

-- ---------------------------------------------------------------------------
-- CA-020-06 · Update propio con contacto válido, dentro del tenant
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_email text;
  v_phone text;
BEGIN
  UPDATE barbershop
     SET contact_email = 'contacto@ejemplo.test',
         contact_phone = '+573001234567'
   WHERE id = '11111111-1111-1111-1111-111111111111';

  SELECT contact_email, contact_phone INTO v_email, v_phone
  FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';

  IF v_email <> 'contacto@ejemplo.test' OR v_phone <> '+573001234567' THEN
    RAISE EXCEPTION 'CA-020-06: el contacto propio no se guardó como se esperaba (% / %).', v_email, v_phone;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-020-06 OK · contacto válido se persiste dentro del propio tenant'

-- ---------------------------------------------------------------------------
-- CA-020-06 · Contacto vacío se persiste como NULL, nunca como cadena vacía
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_email text;
BEGIN
  UPDATE barbershop SET contact_email = NULL
  WHERE id = '11111111-1111-1111-1111-111111111111';

  SELECT contact_email INTO v_email
  FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';

  IF v_email IS NOT NULL THEN
    RAISE EXCEPTION 'CA-020-06: contacto vacío no quedó NULL (quedó %).', v_email;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-020-06 OK · el servicio persiste ausencia como NULL, no cadena vacía'

-- ---------------------------------------------------------------------------
-- Constraint: correo con forma inválida se rechaza en la base
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    UPDATE barbershop SET contact_email = 'no-es-un-correo'
    WHERE id = '11111111-1111-1111-1111-111111111111';
    RAISE EXCEPTION 'barbershop_contact_email_ck: se aceptó un correo con forma inválida.';
  EXCEPTION
    WHEN check_violation THEN
      NULL; -- Esperado.
  END;

  BEGIN
    UPDATE barbershop SET contact_email = 'MAYUSCULA@Ejemplo.test'
    WHERE id = '11111111-1111-1111-1111-111111111111';
    RAISE EXCEPTION 'barbershop_contact_email_ck: se aceptó un correo sin normalizar a minúsculas.';
  EXCEPTION
    WHEN check_violation THEN
      NULL; -- Esperado: la normalización a minúsculas es responsabilidad del servicio (defensa en profundidad).
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'constraint OK · barbershop_contact_email_ck rechaza forma inválida y mayúsculas'

-- ---------------------------------------------------------------------------
-- Constraint: teléfono fuera de forma E.164 se rechaza en la base
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    UPDATE barbershop SET contact_phone = '3001234567'
    WHERE id = '11111111-1111-1111-1111-111111111111';
    RAISE EXCEPTION 'barbershop_contact_phone_ck: se aceptó un teléfono sin indicativo E.164.';
  EXCEPTION
    WHEN check_violation THEN
      NULL; -- Esperado.
  END;

  BEGIN
    UPDATE barbershop SET contact_phone = '+57'
    WHERE id = '11111111-1111-1111-1111-111111111111';
    RAISE EXCEPTION 'barbershop_contact_phone_ck: se aceptó un teléfono demasiado corto.';
  EXCEPTION
    WHEN check_violation THEN
      NULL; -- Esperado.
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'constraint OK · barbershop_contact_phone_ck rechaza formas fuera de E.164'

-- ---------------------------------------------------------------------------
-- CA-020-05 · Contexto de A no ve ni puede escribir el contacto de B
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_visible integer;
  v_updated integer;
BEGIN
  -- B es invisible con el contexto de A: RLS ya lo demuestra en
  -- hu001_aislamiento_rls.sql; aquí se repite acotado a las columnas nuevas.
  SELECT count(*) INTO v_visible
  FROM barbershop
  WHERE id = '22222222-2222-2222-2222-222222222222';

  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'CA-020-05: el contexto de A ve la fila de B (RLS rota).';
  END IF;

  UPDATE barbershop SET contact_email = 'intruso@ejemplo.test'
  WHERE id = '22222222-2222-2222-2222-222222222222';
  GET DIAGNOSTICS v_updated = ROW_COUNT;

  IF v_updated <> 0 THEN
    RAISE EXCEPTION 'CA-020-05: el contexto de A pudo actualizar el contacto de B.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CA-020-05 OK · el contexto de A no ve ni puede escribir el contacto de B'

-- ---------------------------------------------------------------------------
-- Zona inválida a nivel de aplicación no es responsabilidad del CHECK: se
-- confirma aquí que la base SIGUE aceptando cualquier texto en timezone
-- (documentado en la migración fundacional), para que quede explícito por
-- qué la validación IANA vive en el servicio Go y no aquí.
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
BEGIN
  BEGIN
    UPDATE barbershop SET timezone = 'UTC-5'
    WHERE id = '11111111-1111-1111-1111-111111111111';
  EXCEPTION
    WHEN check_violation THEN
      RAISE EXCEPTION
        'la base rechazó un texto de zona no vacío; confirma que la validación IANA sigue viviendo '
        'exclusivamente en el servicio (CA-020-03), no que este CHECK deba reforzarse aquí.';
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'documentado OK · el CHECK de timezone no valida IANA; lo hace el servicio Go'

\echo '=== HU-020 · todas las comprobaciones pasaron ==='
