-- Pruebas SQL del issue #292 (DEC-110): marca y vocabulario de la barbería.
-- Cubren a nivel de base de datos los valores por defecto, las restricciones
-- CHECK de la lista cerrada de acentos y de la forma de un término, y el
-- aislamiento RLS entre dos barberías. La validación fina de la forma
-- (solo letras) y la orquestación del PATCH viven en el servicio Go
-- (internal/modules/shops), no aquí.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f testdata/dos_barberias.sql
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/brand_marca_vocabulario.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app`
-- (estandar-base-datos.md §9.8), igual que hu020_configuracion_barberia.sql.
--
-- Cada comprobación falla lanzando una excepción. Si el archivo termina, pasó.

\set ON_ERROR_STOP on

\echo '=== Issue #292 · pruebas de marca y vocabulario de la barbería ==='

-- ---------------------------------------------------------------------------
-- Columnas NOT NULL con el default que reproduce la interfaz anterior
-- ---------------------------------------------------------------------------
DO $$
DECLARE
  v_col record;
  v_expected text[][] := ARRAY[
    ['brand_accent',             '''brass''::text'],
    ['business_term',            '''barbería''::text'],
    ['business_term_gender',     '''feminine''::text'],
    ['professional_term',        '''barbero''::text'],
    ['professional_term_plural', '''barberos''::text'],
    ['professional_term_gender', '''masculine''::text']
  ];
  i integer;
BEGIN
  FOR i IN 1 .. array_length(v_expected, 1) LOOP
    SELECT is_nullable, column_default INTO v_col
    FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop'
      AND column_name = v_expected[i][1];

    IF NOT FOUND THEN
      RAISE EXCEPTION 'barbershop.% debe existir.', v_expected[i][1];
    END IF;
    IF v_col.is_nullable <> 'NO' THEN
      RAISE EXCEPTION 'barbershop.% debe ser NOT NULL.', v_expected[i][1];
    END IF;
    IF v_col.column_default <> v_expected[i][2] THEN
      RAISE EXCEPTION 'barbershop.%: default % distinto del esperado %.',
        v_expected[i][1], v_col.column_default, v_expected[i][2];
    END IF;
  END LOOP;
END
$$;
\echo 'columnas OK · seis columnas NOT NULL con el default de la interfaz anterior'

-- ---------------------------------------------------------------------------
-- Una barbería que nunca tocó la marca lee los defaults
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  b barbershop%ROWTYPE;
BEGIN
  SELECT * INTO b FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF b.brand_accent <> 'brass' OR b.business_term <> 'barbería'
     OR b.professional_term <> 'barbero' OR b.professional_term_plural <> 'barberos'
     OR b.business_term_gender <> 'feminine' OR b.professional_term_gender <> 'masculine' THEN
    RAISE EXCEPTION 'la barbería existente debe leer los defaults de marca (leyó % / % / %).',
      b.brand_accent, b.business_term, b.professional_term;
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'defaults OK · una barbería existente conserva latón, barbería y barbero'

-- ---------------------------------------------------------------------------
-- Update propio válido dentro del tenant
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  b barbershop%ROWTYPE;
BEGIN
  UPDATE barbershop
     SET brand_accent = 'emerald',
         business_term = 'salón de belleza',
         business_term_gender = 'masculine',
         professional_term = 'estilista',
         professional_term_plural = 'estilistas',
         professional_term_gender = 'feminine'
   WHERE id = '11111111-1111-1111-1111-111111111111';

  SELECT * INTO b FROM barbershop WHERE id = '11111111-1111-1111-1111-111111111111';
  IF b.brand_accent <> 'emerald' OR b.business_term <> 'salón de belleza'
     OR b.professional_term <> 'estilista' OR b.professional_term_plural <> 'estilistas'
     OR b.professional_term_gender <> 'feminine' THEN
    RAISE EXCEPTION 'la marca propia no se guardó como se esperaba.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'update OK · la marca válida se persiste dentro del propio tenant'

-- ---------------------------------------------------------------------------
-- CHECK: cada valor fuera de contrato se rechaza en la base
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_case record;
BEGIN
  FOR v_case IN
    SELECT * FROM (VALUES
      ('brand_accent',             'rojo-libre',                'barbershop_brand_accent_ck'),
      ('brand_accent',             '#b8955a',                   'barbershop_brand_accent_ck'),
      ('business_term_gender',     'neutral',                   'barbershop_business_term_gender_ck'),
      ('professional_term_gender', '',                          'barbershop_professional_term_gender_ck'),
      ('business_term',            'Salón',                     'barbershop_business_term_ck'),
      ('business_term',            ' salón',                    'barbershop_business_term_ck'),
      ('business_term',            's',                         'barbershop_business_term_ck'),
      ('business_term',            repeat('a', 31),             'barbershop_business_term_ck'),
      ('business_term',            'salón 2',                   'barbershop_business_term_ck'),
      ('professional_term',        E'estilista\n',              'barbershop_professional_term_ck'),
      ('professional_term',        'ESTILISTA',                 'barbershop_professional_term_ck'),
      ('professional_term_plural', 'estilistas9',               'barbershop_professional_term_plural_ck'),
      ('professional_term_plural', '',                          'barbershop_professional_term_plural_ck')
    ) AS t(col, val, expected_constraint)
  LOOP
    BEGIN
      EXECUTE format(
        'UPDATE barbershop SET %I = %L WHERE id = %L',
        v_case.col, v_case.val, '11111111-1111-1111-1111-111111111111');
      RAISE EXCEPTION 'se aceptó % = % (esperaba rechazo por %).',
        v_case.col, v_case.val, v_case.expected_constraint;
    EXCEPTION
      WHEN check_violation THEN
        NULL; -- Esperado.
    END;
  END LOOP;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'CHECK OK · acento fuera de lista, género inválido y términos mal formados se rechazan'

-- ---------------------------------------------------------------------------
-- RLS: un tenant no lee ni modifica la marca de otro
-- ---------------------------------------------------------------------------
BEGIN;
SET ROLE barberia_app;
SET LOCAL app.barbershop_id = '11111111-1111-1111-1111-111111111111';

DO $$
DECLARE
  v_rows integer;
  v_visible integer;
BEGIN
  UPDATE barbershop SET brand_accent = 'ruby'
   WHERE id = '22222222-2222-2222-2222-222222222222';
  GET DIAGNOSTICS v_rows = ROW_COUNT;
  IF v_rows <> 0 THEN
    RAISE EXCEPTION 'RLS: el tenant A modificó la marca del tenant B (% filas).', v_rows;
  END IF;

  SELECT count(*) INTO v_visible FROM barbershop
   WHERE id = '22222222-2222-2222-2222-222222222222';
  IF v_visible <> 0 THEN
    RAISE EXCEPTION 'RLS: el tenant A puede leer la marca del tenant B.';
  END IF;
END
$$;

RESET ROLE;
ROLLBACK;

-- La barbería B conserva su marca inicial tras el intento anterior.
DO $$
DECLARE
  v_accent text;
BEGIN
  SELECT brand_accent INTO v_accent FROM barbershop
   WHERE id = '22222222-2222-2222-2222-222222222222';
  IF v_accent <> 'brass' THEN
    RAISE EXCEPTION 'la marca del tenant B cambió a % tras el intento del tenant A.', v_accent;
  END IF;
END
$$;
\echo 'RLS OK · el tenant A no lee ni escribe la marca del tenant B'

\echo '=== Issue #292 · todas las pruebas pasaron ==='
