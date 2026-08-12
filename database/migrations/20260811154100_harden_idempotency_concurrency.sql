-- Propósito
--   Corrección roll-forward de DDL-IDEM-01
--   (docs/05-backend/revision-ddl-seguridad-2026-08-11.md): implementa la
--   semántica concurrente de idempotencia aprobada en DEC-043 (bloqueo
--   consultivo transaccional, sin espera no acotada), revoca el acceso
--   directo de escritura del API sobre `idempotency_record` y expone
--   funciones estrechas y revisadas en su lugar, con límites de tamaño en
--   la respuesta almacenada.
--
-- Reglas y decisiones
--   DEC-043 (semántica concurrente), DEC-040 (modelo de roles), DDL-IDEM-01.
--
-- Qué NO hace esta migración
--   No edita `20260807170100_create_idempotency_record.sql` (DEC-036:
--   inmutable). No cambia el alcance de la clave de idempotencia (sigue
--   sin incluir `operation`, según el diseño original). No cambia
--   `idempotency_key` a un digest almacenado: sigue como texto plano
--   recibido del cliente; esa es una decisión de producto/privacidad
--   pendiente (prompt-endurecimiento-ddl.md, Fase 3), no algo que este PR
--   deba decidir.
--
-- Protocolo resultante
--   1. `idempotency_begin` toma `pg_try_advisory_xact_lock` derivado de
--      (barbershop_id, idempotency_key). Si no puede tomarlo, responde
--      'locked' de inmediato: nadie espera a que termine otra transacción.
--   2. Con el lock tomado, busca la fila. Si no existe o está vencida, la
--      reclama atómicamente (inserta una nueva) y responde 'proceed'.
--   3. Si existe, vigente y de otra operación, responde 'conflict_operation'
--      (RN-IDE-01: una clave no se reutiliza para otra operación).
--   4. Si existe, completada y la huella coincide, responde 'replay' con la
--      respuesta original. Si la huella difiere, 'conflict_fingerprint'.
--   5. Si existe, en curso y vigente, responde 'conflict_in_progress'.
--   6. El llamador ejecuta el efecto en la misma transacción y confirma con
--      `idempotency_complete` (solo si sigue 'in_progress') o limpia con
--      `idempotency_abort` (solo borra 'in_progress'; nunca una fila
--      'completed', cerrando el hallazgo de DDL-IDEM-01).
--   7. `idempotency_purge_expired` es la limpieza de mantenimiento para
--      claves vencidas que nadie reclamó; exclusiva de `barberia_worker`.

-- ---------------------------------------------------------------------------
-- 1. Precondición
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'idempotency_record'
  ) THEN
    RAISE EXCEPTION
      'Esta migración corrige 20260807170100_create_idempotency_record.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_worker') THEN
    RAISE EXCEPTION
      'Esta migración depende de barberia_worker '
      '(20260811145252_harden_roles_and_definer_functions.sql).';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Límites de la respuesta almacenada (DDL-IDEM-01)
-- ---------------------------------------------------------------------------

-- Formato mínimo de tipo de medio (`tipo/subtipo`, sin parámetros); acota
-- longitud para no depender de un valor arbitrario del proveedor o cliente.
ALTER TABLE idempotency_record
  ADD CONSTRAINT idempotency_record_response_content_type_ck CHECK (
    response_content_type IS NULL
    OR (
      char_length(response_content_type) <= 100
      AND response_content_type ~ '^[a-z0-9][a-z0-9!#$&^_.+-]*/[a-z0-9][a-z0-9!#$&^_.+-]*$'
    )
  );

-- 4000 caracteres alcanza para confirmar el recurso creado (RN-IDE-01) sin
-- convertir la tabla en almacén general de respuestas HTTP.
ALTER TABLE idempotency_record
  ADD CONSTRAINT idempotency_record_response_body_len_ck CHECK (
    response_body IS NULL OR char_length(response_body) <= 4000
  );

COMMENT ON COLUMN idempotency_record.response_body IS
  'Respuesta original en texto, no en jsonb: se reproduce byte a byte y no debe '
  'reordenarse ni normalizarse al almacenarla. Máximo 4000 caracteres: guarda solo '
  'la confirmación mínima necesaria, no un espejo general de la respuesta HTTP.';

-- ---------------------------------------------------------------------------
-- 3. Revocar escritura directa del API (DDL-IDEM-01)
-- ---------------------------------------------------------------------------

-- Toda escritura pasa por las funciones de la sección 4, que validan
-- argumentos y nunca borran una fila `completed`. SELECT se conserva: es
-- inofensivo bajo RLS y útil para observabilidad sin exponer una vía de
-- escritura.
REVOKE INSERT, UPDATE, DELETE ON TABLE idempotency_record FROM barberia_app;

DROP POLICY idempotency_record_insert_tenant_policy ON idempotency_record;
DROP POLICY idempotency_record_update_tenant_policy ON idempotency_record;
DROP POLICY idempotency_record_delete_tenant_policy ON idempotency_record;

-- ---------------------------------------------------------------------------
-- 4. Funciones SECURITY DEFINER del protocolo
-- ---------------------------------------------------------------------------

CREATE OR REPLACE FUNCTION idempotency_begin(
  p_barbershop_id       uuid,
  p_idempotency_key     text,
  p_operation           text,
  p_request_fingerprint text,
  p_ttl_seconds         integer
)
RETURNS TABLE (
  outcome                text,
  response_status         smallint,
  response_content_type   text,
  response_body           text
)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_lock_key bigint;
  v_locked   boolean;
  v_row      public.idempotency_record%ROWTYPE;
BEGIN
  IF p_barbershop_id IS NULL OR p_idempotency_key IS NULL OR p_operation IS NULL
     OR p_request_fingerprint IS NULL OR p_ttl_seconds IS NULL THEN
    RAISE EXCEPTION 'idempotency_begin: ningún argumento admite NULL.';
  END IF;
  IF p_ttl_seconds < 1 OR p_ttl_seconds > 86400 THEN
    RAISE EXCEPTION 'idempotency_begin: p_ttl_seconds fuera de rango (1-86400).';
  END IF;
  IF btrim(p_idempotency_key) = '' OR char_length(p_idempotency_key) > 255 THEN
    RAISE EXCEPTION 'idempotency_begin: idempotency_key inválida.';
  END IF;
  IF btrim(p_operation) = '' OR char_length(p_operation) > 120 THEN
    RAISE EXCEPTION 'idempotency_begin: operation inválida.';
  END IF;
  IF p_request_fingerprint !~ '^[0-9a-f]{32,128}$' THEN
    RAISE EXCEPTION 'idempotency_begin: request_fingerprint inválido.';
  END IF;

  -- Bloqueo consultivo transaccional (DEC-043): se libera solo al terminar
  -- la transacción (COMMIT o ROLLBACK), nunca queda huérfano entre
  -- solicitudes. `pg_try_advisory_xact_lock` no espera: si otra sesión ya
  -- lo tiene, retorna false de inmediato.
  v_lock_key := pg_catalog.hashtextextended(
    p_barbershop_id::text || ':' || p_idempotency_key, 0
  );
  v_locked := pg_catalog.pg_try_advisory_xact_lock(v_lock_key);

  IF NOT v_locked THEN
    RETURN QUERY SELECT 'locked'::text, NULL::smallint, NULL::text, NULL::text;
    RETURN;
  END IF;

  SELECT * INTO v_row
  FROM public.idempotency_record
  WHERE barbershop_id = p_barbershop_id
    AND idempotency_key = p_idempotency_key;

  IF NOT FOUND OR v_row.expires_at <= pg_catalog.now() THEN
    -- No existe, o existe vencida: se reclama atómicamente aunque el
    -- mantenimiento (idempotency_purge_expired) todavía no la haya
    -- limpiado. El lock ya tomado impide que dos sesiones reclamen la
    -- misma clave a la vez.
    DELETE FROM public.idempotency_record
    WHERE barbershop_id = p_barbershop_id AND idempotency_key = p_idempotency_key;

    INSERT INTO public.idempotency_record (
      barbershop_id, idempotency_key, operation, request_fingerprint,
      status, expires_at
    ) VALUES (
      p_barbershop_id, p_idempotency_key, p_operation, p_request_fingerprint,
      'in_progress', pg_catalog.now() + pg_catalog.make_interval(secs => p_ttl_seconds)
    );

    RETURN QUERY SELECT 'proceed'::text, NULL::smallint, NULL::text, NULL::text;
    RETURN;
  END IF;

  IF v_row.operation <> p_operation THEN
    -- RN-IDE-01: una clave ya usada no se reutiliza para otra operación.
    RETURN QUERY SELECT 'conflict_operation'::text, NULL::smallint, NULL::text, NULL::text;
    RETURN;
  END IF;

  IF v_row.status = 'completed' THEN
    IF v_row.request_fingerprint = p_request_fingerprint THEN
      RETURN QUERY SELECT
        'replay'::text, v_row.response_status, v_row.response_content_type, v_row.response_body;
    ELSE
      RETURN QUERY SELECT 'conflict_fingerprint'::text, NULL::smallint, NULL::text, NULL::text;
    END IF;
    RETURN;
  END IF;

  -- status = 'in_progress', vigente, misma operación: otra ejecución
  -- legítima sigue en curso (poco frecuente bajo el lock, pero posible si
  -- el TTL de una reclamación previa aún no venció).
  RETURN QUERY SELECT 'conflict_in_progress'::text, NULL::smallint, NULL::text, NULL::text;
END;
$$;

CREATE OR REPLACE FUNCTION idempotency_complete(
  p_barbershop_id         uuid,
  p_idempotency_key       text,
  p_response_status       integer,
  p_response_content_type text,
  p_response_body         text
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_updated integer;
BEGIN
  IF p_barbershop_id IS NULL OR p_idempotency_key IS NULL OR p_response_status IS NULL
     OR p_response_content_type IS NULL OR p_response_body IS NULL THEN
    RAISE EXCEPTION 'idempotency_complete: ningún argumento admite NULL.';
  END IF;
  IF p_response_status NOT BETWEEN 100 AND 599 THEN
    RAISE EXCEPTION 'idempotency_complete: response_status fuera de rango.';
  END IF;

  -- Solo transiciona una fila que sigue 'in_progress': confirmar dos veces
  -- la misma clave no duplica nada, y nunca sobrescribe una ya completada.
  UPDATE public.idempotency_record
  SET status                 = 'completed',
      response_status        = p_response_status,
      response_content_type  = p_response_content_type,
      response_body          = p_response_body
  WHERE barbershop_id = p_barbershop_id
    AND idempotency_key = p_idempotency_key
    AND status = 'in_progress';

  GET DIAGNOSTICS v_updated = ROW_COUNT;
  RETURN v_updated = 1;
END;
$$;

CREATE OR REPLACE FUNCTION idempotency_abort(
  p_barbershop_id   uuid,
  p_idempotency_key text
)
RETURNS boolean
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_deleted integer;
BEGIN
  IF p_barbershop_id IS NULL OR p_idempotency_key IS NULL THEN
    RAISE EXCEPTION 'idempotency_abort: ningún argumento admite NULL.';
  END IF;

  -- Cierra DDL-IDEM-01: solo borra una fila 'in_progress' (el efecto
  -- fracasó y no debe bloquear un reintento legítimo). Una fila
  -- 'completed' jamás se borra desde aquí ni desde ningún otro camino del
  -- API, así que un reintento no puede repetir un efecto crítico ya hecho.
  DELETE FROM public.idempotency_record
  WHERE barbershop_id = p_barbershop_id
    AND idempotency_key = p_idempotency_key
    AND status = 'in_progress';

  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  RETURN v_deleted = 1;
END;
$$;

CREATE OR REPLACE FUNCTION idempotency_purge_expired(p_limit integer)
RETURNS integer
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
  v_deleted integer;
BEGIN
  IF p_limit IS NULL OR p_limit < 1 OR p_limit > 1000 THEN
    RAISE EXCEPTION 'idempotency_purge_expired: p_limit fuera de rango (1-1000).';
  END IF;

  -- Mantenimiento acotado (DDL-OPS-01): limpia lo que idempotency_begin no
  -- alcanzó a reciclar porque nadie volvió a usar esa clave. Cualquier
  -- estado (in_progress o completed) es elegible una vez vencida.
  WITH due AS (
    SELECT id FROM public.idempotency_record
    WHERE expires_at <= pg_catalog.now()
    ORDER BY expires_at
    LIMIT p_limit
    FOR UPDATE SKIP LOCKED
  )
  DELETE FROM public.idempotency_record
  WHERE id IN (SELECT id FROM due);

  GET DIAGNOSTICS v_deleted = ROW_COUNT;
  RETURN v_deleted;
END;
$$;

REVOKE ALL ON FUNCTION idempotency_begin(uuid, text, text, text, integer) FROM PUBLIC;
REVOKE ALL ON FUNCTION idempotency_complete(uuid, text, integer, text, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION idempotency_abort(uuid, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION idempotency_purge_expired(integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION idempotency_begin(uuid, text, text, text, integer)        TO barberia_app;
GRANT EXECUTE ON FUNCTION idempotency_complete(uuid, text, integer, text, text)    TO barberia_app;
GRANT EXECUTE ON FUNCTION idempotency_abort(uuid, text)                            TO barberia_app;
GRANT EXECUTE ON FUNCTION idempotency_purge_expired(integer)                       TO barberia_worker;

COMMENT ON FUNCTION idempotency_begin(uuid, text, text, text, integer) IS
  'Punto de entrada único del protocolo de idempotencia (DEC-043). Toma un bloqueo '
  'consultivo transaccional por (barbershop_id, idempotency_key); si no puede, '
  'responde "locked" sin esperar. Nunca deja la tabla en un estado a medio escribir.';

COMMENT ON FUNCTION idempotency_complete(uuid, text, integer, text, text) IS
  'Confirma una operación en curso con su respuesta. Solo transiciona una fila '
  'in_progress; no reescribe una fila ya completed.';

COMMENT ON FUNCTION idempotency_abort(uuid, text) IS
  'Limpia una reclamación cuyo efecto falló. Solo borra in_progress: nunca borra '
  'una fila completed (cierra DDL-IDEM-01).';

COMMENT ON FUNCTION idempotency_purge_expired(integer) IS
  'Mantenimiento del worker: recicla en lote las claves vencidas que nadie volvió '
  'a reclamar. Exclusivo de barberia_worker.';
