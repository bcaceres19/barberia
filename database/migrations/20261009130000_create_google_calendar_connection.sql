-- Propósito
--   Issue #323 (DEC-099, DEC-102): crea `google_calendar_connection`, la
--   conexión de un barbero con SU Google Calendar (una fila como máximo por
--   barbero), y `google_calendar_oauth_state`, el estado de un solo uso del
--   flujo OAuth (PKCE) que liga la autorización a barbería, barbero, usuario
--   y sesión. La publicación de citas y la cola de trabajos llegan en el
--   issue #324 y no viven aquí.
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), DEC-019/DEC-024
--   (esquema compartido con RLS), DEC-035/DEC-036 (estándar de base de datos,
--   Atlas, migración inmutable), DEC-040 (modelo de roles), DEC-099 (la
--   conexión pertenece a (barbería, barbero); NAVA → Google), DEC-100 (el
--   barbero se resuelve por el vínculo con el usuario), DEC-102 (refresh
--   token cifrado con AES-256-GCM en la aplicación, con identificador de
--   clave para rotación; el access token nunca se persiste).
--
-- Credenciales
--   La base solo guarda el texto cifrado del refresh token y el `key_id` de
--   la clave que lo cifró; nunca el token en claro ni el access token. Un
--   CHECK liga las credenciales al estado: `connected` y `error` (fallo
--   transitorio, el token sigue siendo válido) las conservan; `reauth_required`
--   (Google revocó o caducó el permiso) y `disconnected` las tienen borradas.
--   Desconectar NO borra la fila: conserva barbero, preferencia de recordatorio
--   y fechas, pero sin credenciales. Reconectar reutiliza la misma fila.
--
-- Qué NO hace esta migración
--   No crea eventos, vínculos de evento ni cola (issue #324), no concede
--   privilegios al rol `barberia_worker` (el worker los recibe con la cola) y
--   no agrega DELETE a la conexión: se desconecta, no se elimina. El estado
--   OAuth sí admite DELETE para purgar los vencidos.
--
-- Clasificación de datos
--   `google_account_email` es dato personal (correo de la cuenta de Google del
--   barbero); el texto cifrado del refresh token es un secreto. Retención:
--   mientras el barbero mantenga la conexión; al desconectar se borran las
--   credenciales y el correo. El estado OAuth es efímero (vigencia corta,
--   un solo uso) y se purga al vencer.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE, ver
--   20260811145252_harden_roles_and_definer_functions.sql).
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
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barber' AND column_name = 'staff_user_id'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20261009120000_add_barber_staff_user_link.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public'
      AND table_name IN ('google_calendar_connection', 'google_calendar_oauth_state')
  ) THEN
    RAISE EXCEPTION 'Las tablas de Google Calendar ya existen: esta migración no puede aplicarse dos veces.';
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql '
      '(rol barberia_owner); esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Conexión del barbero con su Google Calendar
-- ---------------------------------------------------------------------------

CREATE TABLE google_calendar_connection (
  id                       uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id            uuid        NOT NULL,
  barber_id                uuid        NOT NULL,
  status                   text        NOT NULL,
  google_account_email     text,
  calendar_id              text        NOT NULL DEFAULT 'primary',
  reminder_minutes         integer,
  refresh_token_ciphertext bytea,
  token_key_id             text,
  connected_at             timestamptz,
  disconnected_at          timestamptz,
  last_synced_at           timestamptz,
  last_error_code          text,
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT google_calendar_connection_id_pk PRIMARY KEY (id),

  -- Referenciable por la cola y los vínculos de evento del issue #324.
  CONSTRAINT google_calendar_connection_barbershop_id_id_uk UNIQUE (barbershop_id, id),

  -- Una conexión como máximo por barbero: la propia restricción impide una
  -- segunda fila y sirve de índice para toda consulta por barbero.
  CONSTRAINT google_calendar_connection_barbershop_id_barber_id_uk UNIQUE (barbershop_id, barber_id),

  -- FK compuesta por barbershop_id: impide conectar a un barbero de otra
  -- barbería AUN con SQL directo bajo el rol de aplicación (RN-TEN-01). Sin
  -- CASCADE: `barber` no se borra (DEC-047).
  CONSTRAINT google_calendar_connection_barber_fk FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,

  CONSTRAINT google_calendar_connection_status_ck
    CHECK (status IN ('connected', 'reauth_required', 'error', 'disconnected')),

  -- 0 a 40320 minutos (4 semanas, el límite de Google); NULL = recordatorios
  -- predeterminados del calendario del barbero (DEC-101).
  CONSTRAINT google_calendar_connection_reminder_minutes_ck
    CHECK (reminder_minutes IS NULL OR reminder_minutes BETWEEN 0 AND 40320),

  CONSTRAINT google_calendar_connection_calendar_id_ck
    CHECK (btrim(calendar_id) <> '' AND char_length(calendar_id) <= 255),

  CONSTRAINT google_calendar_connection_account_email_ck
    CHECK (google_account_email IS NULL OR char_length(google_account_email) <= 254),

  -- Las credenciales existen exactamente cuando el token sigue siendo válido.
  CONSTRAINT google_calendar_connection_credentials_ck CHECK (
    (status IN ('connected', 'error')
       AND refresh_token_ciphertext IS NOT NULL AND token_key_id IS NOT NULL)
    OR
    (status IN ('reauth_required', 'disconnected')
       AND refresh_token_ciphertext IS NULL AND token_key_id IS NULL)
  )
);

COMMENT ON TABLE google_calendar_connection IS
  'Conexión de un barbero con su propio Google Calendar (issue #323, DEC-099): una fila como '
  'máximo por (barbería, barbero); desconectar conserva la fila sin credenciales y reconectar la '
  'reutiliza. Propietario funcional: módulo googlecalendar. Clasificación: datos personales '
  '(correo de la cuenta de Google) y secreto (texto cifrado del refresh token). Retención: '
  'mientras el barbero mantenga la conexión.';

COMMENT ON COLUMN google_calendar_connection.status IS
  'connected: publica con normalidad. error: fallo transitorio, el token sigue siendo válido. '
  'reauth_required: Google revocó o caducó el permiso; credenciales borradas hasta reconectar. '
  'disconnected: el barbero desconectó; credenciales borradas.';

COMMENT ON COLUMN google_calendar_connection.refresh_token_ciphertext IS
  'Refresh token cifrado con AES-256-GCM en la aplicación (nonce + texto cifrado + etiqueta). '
  'Nunca el token en claro; el access token no se persiste (DEC-102). NULL sin credenciales.';

COMMENT ON COLUMN google_calendar_connection.token_key_id IS
  'Identificador de la clave que cifró el refresh token, para rotar la clave sin perder '
  'conexiones (DEC-102). NULL sin credenciales.';

COMMENT ON COLUMN google_calendar_connection.reminder_minutes IS
  'Anticipación del recordatorio emergente de cada evento, en minutos (0 a 40320). NULL usa los '
  'recordatorios predeterminados del calendario del barbero (DEC-101).';

COMMENT ON COLUMN google_calendar_connection.calendar_id IS
  'Calendario donde se publica. `primary` es el calendario principal de la cuenta conectada.';

COMMENT ON COLUMN google_calendar_connection.last_error_code IS
  'Código corto de la última falla permanente o transitoria, sin detalle del proveedor ni '
  'datos personales. NULL cuando la conexión está sana.';

CREATE TRIGGER google_calendar_connection_set_updated_at
  BEFORE UPDATE ON google_calendar_connection
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE google_calendar_connection ENABLE ROW LEVEL SECURITY;
ALTER TABLE google_calendar_connection FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: FORCE ROW LEVEL SECURITY también alcanza al propietario
-- (estandar-base-datos.md §9.5, mismo patrón que barber/barber_photo).
CREATE POLICY google_calendar_connection_all_admin_policy ON google_calendar_connection
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY google_calendar_connection_select_tenant_policy ON google_calendar_connection
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY google_calendar_connection_insert_tenant_policy ON google_calendar_connection
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY google_calendar_connection_update_tenant_policy ON google_calendar_connection
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin política ni GRANT de DELETE: se desconecta, no se elimina.

GRANT SELECT, INSERT, UPDATE ON TABLE google_calendar_connection TO barberia_app;

-- ---------------------------------------------------------------------------
-- 3. Estado de un solo uso del flujo OAuth (PKCE)
-- ---------------------------------------------------------------------------

CREATE TABLE google_calendar_oauth_state (
  id                         uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id              uuid        NOT NULL,
  barber_id                  uuid        NOT NULL,
  staff_user_id              uuid        NOT NULL,
  session_id                 uuid        NOT NULL,
  state_hash                 text        NOT NULL,
  code_verifier_ciphertext   bytea       NOT NULL,
  verifier_key_id            text        NOT NULL,
  expires_at                 timestamptz NOT NULL,
  consumed_at                timestamptz,
  created_at                 timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT google_calendar_oauth_state_id_pk PRIMARY KEY (id),

  -- Solo se guarda el hash del `state`; el valor en claro viaja por Google y
  -- por la URL de retorno, nunca por la base. Único global: el callback lo
  -- busca dentro del tenant de la sesión, pero dos filas jamás comparten hash.
  CONSTRAINT google_calendar_oauth_state_state_hash_uk UNIQUE (state_hash),
  CONSTRAINT google_calendar_oauth_state_state_hash_ck CHECK (char_length(state_hash) = 64),

  CONSTRAINT google_calendar_oauth_state_barber_fk FOREIGN KEY (barbershop_id, barber_id)
    REFERENCES barber (barbershop_id, id) ON DELETE RESTRICT,
  CONSTRAINT google_calendar_oauth_state_staff_user_fk FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,
  -- La autorización queda ligada a la sesión que la inició: si la sesión se
  -- purga, el estado pendiente desaparece con ella.
  CONSTRAINT google_calendar_oauth_state_session_fk FOREIGN KEY (session_id)
    REFERENCES staff_session (id) ON DELETE CASCADE,

  CONSTRAINT google_calendar_oauth_state_expires_at_ck CHECK (expires_at > created_at)
);

COMMENT ON TABLE google_calendar_oauth_state IS
  'Estado de un solo uso del flujo OAuth con Google (issue #323, DEC-102): liga la autorización a '
  'barbería, barbero, usuario y sesión, y guarda cifrado el verificador PKCE. Vigencia corta; se '
  'consume una sola vez y se purga al vencer. Clasificación: secreto efímero.';

COMMENT ON COLUMN google_calendar_oauth_state.state_hash IS
  'SHA-256 hexadecimal del `state` enviado a Google. El valor en claro nunca se persiste.';

COMMENT ON COLUMN google_calendar_oauth_state.code_verifier_ciphertext IS
  'Verificador PKCE cifrado con AES-256-GCM (mismo mecanismo y clave que el refresh token).';

COMMENT ON COLUMN google_calendar_oauth_state.consumed_at IS
  'Instante en que el callback usó este estado. Un estado consumido jamás vuelve a ser válido.';

CREATE INDEX idx_google_calendar_oauth_state_expires_at
  ON google_calendar_oauth_state (expires_at);

ALTER TABLE google_calendar_oauth_state ENABLE ROW LEVEL SECURITY;
ALTER TABLE google_calendar_oauth_state FORCE  ROW LEVEL SECURITY;

CREATE POLICY google_calendar_oauth_state_all_admin_policy ON google_calendar_oauth_state
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY google_calendar_oauth_state_select_tenant_policy ON google_calendar_oauth_state
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY google_calendar_oauth_state_insert_tenant_policy ON google_calendar_oauth_state
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY google_calendar_oauth_state_update_tenant_policy ON google_calendar_oauth_state
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY google_calendar_oauth_state_delete_tenant_policy ON google_calendar_oauth_state
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE google_calendar_oauth_state TO barberia_app;
