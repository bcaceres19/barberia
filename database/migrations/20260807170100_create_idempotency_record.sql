-- Propósito
--   Registro de claves de idempotencia exigido por HU-004, disponible antes de
--   que exista la primera operación de escritura crítica.
--
-- Reglas y decisiones
--   RN-IDE-01 (operaciones críticas seguras ante reintentos), DEC-035, DEC-037.
--
-- Protocolo que esta tabla sostiene
--   1. La solicitud intenta `INSERT ... ON CONFLICT DO NOTHING` con
--      `status = 'in_progress'` y la huella del contenido.
--   2. Si insertó, ejecuta el efecto y actualiza la fila a `completed` con la
--      respuesta almacenada, dentro de la misma transacción del efecto.
--   3. Si no insertó, lee la fila vigente:
--      · `completed` + misma huella  -> devuelve la respuesta original (CA-004-01);
--      · huella distinta             -> 409, sin ejecutar nada (CA-004-02);
--      · `in_progress`               -> 409 de operación en curso (CA-004-03).
--   4. Si el efecto falla, la fila se BORRA para no bloquear un reintento
--      legítimo (CA-004-06). Por eso no existe un estado `failed`.
--   5. Una fila vencida no se lee; se trata como inexistente (CA-004-05).
--
-- Condición de seguridad
--   La tabla es tenant-aware: una clave literal idéntica en dos barberías son
--   dos registros distintos y RLS impide verlos entre sí (CA-004-04).

CREATE TABLE idempotency_record (
  id                    uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id         uuid        NOT NULL,
  idempotency_key       text        NOT NULL,
  operation             text        NOT NULL,
  request_fingerprint   text        NOT NULL,
  status                text        NOT NULL DEFAULT 'in_progress',
  response_status       smallint,
  response_content_type text,
  response_body         text,
  expires_at            timestamptz NOT NULL,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT idempotency_record_id_pk PRIMARY KEY (id),

  CONSTRAINT idempotency_record_barbershop_id_fk FOREIGN KEY (barbershop_id)
    REFERENCES barbershop (id) ON DELETE RESTRICT,

  -- Alcance de la clave: la barbería. No incluye `operation` a propósito:
  -- RN-IDE-01 exige que una clave ya usada NO se reutilice para otra
  -- operación, y debe responder 409 en lugar de ejecutarse.
  CONSTRAINT idempotency_record_barbershop_id_key_uk
    UNIQUE (barbershop_id, idempotency_key),

  CONSTRAINT idempotency_record_idempotency_key_ck CHECK (
    btrim(idempotency_key) <> '' AND char_length(idempotency_key) <= 255
  ),
  CONSTRAINT idempotency_record_operation_ck CHECK (
    btrim(operation) <> '' AND char_length(operation) <= 120
  ),
  CONSTRAINT idempotency_record_request_fingerprint_ck CHECK (
    char_length(request_fingerprint) BETWEEN 32 AND 128
  ),
  CONSTRAINT idempotency_record_status_ck CHECK (
    status IN ('in_progress', 'completed')
  ),
  CONSTRAINT idempotency_record_expires_at_ck CHECK (expires_at > created_at),

  -- La respuesta almacenada existe si y solo si la operación terminó. Evita
  -- que un reintento reproduzca una respuesta a medio escribir.
  CONSTRAINT idempotency_record_response_ck CHECK (
    (status = 'completed'
       AND response_status IS NOT NULL
       AND response_content_type IS NOT NULL
       AND response_body IS NOT NULL)
    OR
    (status = 'in_progress'
       AND response_status IS NULL
       AND response_content_type IS NULL
       AND response_body IS NULL)
  ),
  CONSTRAINT idempotency_record_response_status_ck CHECK (
    response_status IS NULL OR response_status BETWEEN 100 AND 599
  )
);

COMMENT ON TABLE idempotency_record IS
  'Claves de idempotencia de operaciones de escritura críticas (RN-IDE-01). '
  'Propietario funcional: plataforma. Retención: hasta expires_at; una tarea de '
  'limpieza borra las vencidas. Clasificación: técnico. El cuerpo almacenado puede '
  'contener datos personales del recurso creado, por lo que la tabla queda sujeta a '
  'la misma anonimización que el recurso al que responde.';

COMMENT ON COLUMN idempotency_record.request_fingerprint IS
  'Huella del contenido de la solicitud (hash hexadecimal de método, ruta y cuerpo '
  'canonizado). Comparar huellas es lo que distingue "mismo contenido" de '
  '"contenido distinto" sin volver a guardar la solicitud completa.';

COMMENT ON COLUMN idempotency_record.response_body IS
  'Respuesta original en texto, no en jsonb: se reproduce byte a byte y no debe '
  'reordenarse ni normalizarse al almacenarla.';

CREATE TRIGGER idempotency_record_set_updated_at
  BEFORE UPDATE ON idempotency_record
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Índice para la tarea de limpieza. Parcial no aplica: toda fila vence.
CREATE INDEX idx_idempotency_record_expires_at
  ON idempotency_record (expires_at);

-- ---------------------------------------------------------------------------
-- Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

ALTER TABLE idempotency_record ENABLE ROW LEVEL SECURITY;
ALTER TABLE idempotency_record FORCE  ROW LEVEL SECURITY;

CREATE POLICY idempotency_record_all_admin_policy ON idempotency_record
  FOR ALL TO barberia_migrator
  USING (true) WITH CHECK (true);

CREATE POLICY idempotency_record_select_tenant_policy ON idempotency_record
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY idempotency_record_insert_tenant_policy ON idempotency_record
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY idempotency_record_update_tenant_policy ON idempotency_record
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY idempotency_record_delete_tenant_policy ON idempotency_record
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE idempotency_record TO barberia_app;
