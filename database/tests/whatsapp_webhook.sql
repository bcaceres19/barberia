-- Pruebas SQL del issue #355 (DEC-126): webhook de Meta WhatsApp. Cubren los
-- privilegios (la API solo ejecuta funciones; la purga es exclusiva del worker),
-- la ventana de 24 horas por teléfono, la idempotencia de los estados de entrega
-- y la purga. Las tablas no tienen barbershop_id por diseño (el número emisor es
-- único de la plataforma), así que no aplican dos tenants: lo que se comprueba es
-- que ningún rol de aplicación toca las tablas directamente.
--
-- Uso:
--   psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -f tests/whatsapp_webhook.sql
--
-- Requisito del arnés: la conexión debe poder ejecutar `SET ROLE barberia_app` y
-- `SET ROLE barberia_worker`. Todo se revierte con ROLLBACK.

\set ON_ERROR_STOP on

\echo '=== Issue #355 · pruebas del webhook de Meta WhatsApp ==='

-- Privilegios ---------------------------------------------------------------
DO $$
DECLARE
  v_table text;
  v_role  text;
BEGIN
  FOREACH v_table IN ARRAY ARRAY['whatsapp_conversation_window', 'whatsapp_message_status'] LOOP
    FOREACH v_role IN ARRAY ARRAY['barberia_app', 'barberia_worker'] LOOP
      IF has_table_privilege(v_role, v_table, 'SELECT') OR has_table_privilege(v_role, v_table, 'INSERT')
         OR has_table_privilege(v_role, v_table, 'UPDATE') OR has_table_privilege(v_role, v_table, 'DELETE') THEN
        RAISE EXCEPTION '% no debe tener acceso directo a % (DDL-AUT-01).', v_role, v_table;
      END IF;
    END LOOP;
  END LOOP;

  IF NOT has_function_privilege('barberia_app', 'whatsapp_window_record(text, timestamptz)', 'EXECUTE')
     OR NOT has_function_privilege('barberia_app', 'whatsapp_window_is_open(text)', 'EXECUTE')
     OR NOT has_function_privilege('barberia_app', 'whatsapp_status_record(text, text, timestamptz, integer)', 'EXECUTE') THEN
    RAISE EXCEPTION 'barberia_app debe ejecutar las funciones del webhook.';
  END IF;
  IF has_function_privilege('barberia_app', 'whatsapp_webhook_purge_expired(integer, integer)', 'EXECUTE') THEN
    RAISE EXCEPTION 'barberia_app no debe ejecutar la purga.';
  END IF;
  IF NOT has_function_privilege('barberia_worker', 'whatsapp_webhook_purge_expired(integer, integer)', 'EXECUTE') THEN
    RAISE EXCEPTION 'barberia_worker debe ejecutar la purga.';
  END IF;
  IF has_function_privilege('barberia_worker', 'whatsapp_window_record(text, timestamptz)', 'EXECUTE')
     OR has_function_privilege('barberia_worker', 'whatsapp_status_record(text, text, timestamptz, integer)', 'EXECUTE') THEN
    RAISE EXCEPTION 'barberia_worker no debe escribir eventos del webhook.';
  END IF;
  IF has_function_privilege('public', 'whatsapp_window_record(text, timestamptz)', 'EXECUTE') THEN
    RAISE EXCEPTION 'PUBLIC no debe ejecutar las funciones del webhook.';
  END IF;
END
$$;
\echo 'Issue #355 OK · privilegios mínimos por rol'

-- Ventana de 24 horas -------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
DECLARE
  v_hash  text := encode(sha256('whatsapp-sql-test-window'::bytea), 'hex');
  v_other text := encode(sha256('whatsapp-sql-test-window-other'::bytea), 'hex');
  v_last  timestamptz;
BEGIN
  IF whatsapp_window_is_open(v_hash) THEN
    RAISE EXCEPTION 'Sin mensaje entrante no puede haber ventana abierta.';
  END IF;

  PERFORM whatsapp_window_record(v_hash, now() - interval '1 hour');
  IF NOT whatsapp_window_is_open(v_hash) THEN
    RAISE EXCEPTION 'Un mensaje de hace una hora debe abrir la ventana.';
  END IF;
  IF whatsapp_window_is_open(v_other) THEN
    RAISE EXCEPTION 'La ventana de un teléfono no abre la de otro.';
  END IF;

  -- Una notificación tardía o repetida no retrocede la ventana.
  PERFORM whatsapp_window_record(v_hash, now() - interval '20 hours');
  RESET ROLE;
  SELECT last_inbound_at INTO v_last FROM whatsapp_conversation_window WHERE phone_hash = v_hash;
  SET ROLE barberia_app;
  IF v_last < now() - interval '2 hours' THEN
    RAISE EXCEPTION 'Una notificación antigua retrocedió la última interacción.';
  END IF;

  -- Un reloj adelantado no abre la ventana más allá de ahora.
  PERFORM whatsapp_window_record(v_other, now() + interval '10 days');
  RESET ROLE;
  SELECT expires_at INTO v_last FROM whatsapp_conversation_window WHERE phone_hash = v_other;
  SET ROLE barberia_app;
  IF v_last > now() + interval '24 hours 5 seconds' THEN
    RAISE EXCEPTION 'Una fecha futura extendió la ventana más de 24 horas.';
  END IF;

  -- Un mensaje de hace más de 24 horas no abre nada.
  PERFORM whatsapp_window_record(encode(sha256('whatsapp-sql-test-window-old'::bytea), 'hex'), now() - interval '25 hours');
  IF whatsapp_window_is_open(encode(sha256('whatsapp-sql-test-window-old'::bytea), 'hex')) THEN
    RAISE EXCEPTION 'Un mensaje de hace 25 horas no debe abrir la ventana.';
  END IF;

  BEGIN
    PERFORM whatsapp_window_record('corto', now());
    RAISE EXCEPTION 'Se aceptó un hash de longitud inválida.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM LIKE 'Se aceptó%' THEN RAISE; END IF;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Issue #355 OK · la ventana abre, no retrocede y no se extiende con fechas futuras'

-- Estados de entrega --------------------------------------------------------
BEGIN;
SET ROLE barberia_app;

DO $$
BEGIN
  IF NOT whatsapp_status_record('wamid.test', 'sent', now(), NULL) THEN
    RAISE EXCEPTION 'El primer estado debe registrarse.';
  END IF;
  IF whatsapp_status_record('wamid.test', 'sent', now(), NULL) THEN
    RAISE EXCEPTION 'Un estado repetido no debe registrarse de nuevo.';
  END IF;
  IF NOT whatsapp_status_record('wamid.test', 'delivered', now(), NULL) THEN
    RAISE EXCEPTION 'Otro estado del mismo mensaje debe registrarse.';
  END IF;
  IF NOT whatsapp_status_record('wamid.test', 'failed', now(), 131047) THEN
    RAISE EXCEPTION 'Un fallo con código debe registrarse.';
  END IF;

  BEGIN
    PERFORM whatsapp_status_record('wamid.test', 'inventado', now(), NULL);
    RAISE EXCEPTION 'Se aceptó un estado desconocido.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM LIKE 'Se aceptó%' THEN RAISE; END IF;
  END;
  BEGIN
    PERFORM whatsapp_status_record('', 'sent', now(), NULL);
    RAISE EXCEPTION 'Se aceptó un wamid vacío.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM LIKE 'Se aceptó%' THEN RAISE; END IF;
  END;
END
$$;

RESET ROLE;
ROLLBACK;
\echo 'Issue #355 OK · los estados son idempotentes y validados'

-- Purga ---------------------------------------------------------------------
BEGIN;
INSERT INTO whatsapp_conversation_window (phone_hash, last_inbound_at, expires_at)
VALUES (encode(sha256('purge-vencida'::bytea), 'hex'), now() - interval '30 hours', now() - interval '6 hours'),
       (encode(sha256('purge-vigente'::bytea), 'hex'), now() - interval '1 hour', now() + interval '23 hours');
INSERT INTO whatsapp_message_status (wamid, status, occurred_at, received_at)
VALUES ('wamid.purge.viejo', 'sent', now() - interval '40 days', now() - interval '40 days'),
       ('wamid.purge.reciente', 'sent', now(), now());

SET ROLE barberia_worker;
DO $$
DECLARE
  v_deleted integer;
BEGIN
  SELECT whatsapp_webhook_purge_expired(100, 2592000) INTO v_deleted;
  IF v_deleted <> 2 THEN
    RAISE EXCEPTION 'La purga debía borrar 2 filas (una ventana vencida y un estado viejo), borró %.', v_deleted;
  END IF;
  BEGIN
    PERFORM whatsapp_webhook_purge_expired(0, 2592000);
    RAISE EXCEPTION 'Se aceptó un límite fuera de rango.';
  EXCEPTION WHEN raise_exception THEN
    IF SQLERRM LIKE 'Se aceptó%' THEN RAISE; END IF;
  END;
END
$$;
RESET ROLE;

DO $$
BEGIN
  IF (SELECT count(*) FROM whatsapp_conversation_window) <> 1
     OR (SELECT count(*) FROM whatsapp_message_status) <> 1 THEN
    RAISE EXCEPTION 'La purga debía conservar la ventana vigente y el estado reciente.';
  END IF;
END
$$;

SET ROLE barberia_app;
DO $$
BEGIN
  BEGIN
    PERFORM whatsapp_webhook_purge_expired(10, 2592000);
    RAISE EXCEPTION 'barberia_app ejecutó la purga.';
  EXCEPTION WHEN insufficient_privilege THEN
    NULL;
  END;
END
$$;
RESET ROLE;
ROLLBACK;
\echo 'Issue #355 OK · la purga es exclusiva del worker y respeta lo vigente'

\echo '=== Issue #355 · todas las pruebas pasaron ==='
