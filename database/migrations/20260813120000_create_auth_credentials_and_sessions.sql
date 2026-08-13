-- Propósito
--   HU-005: material de autenticación (`staff_credential`) y sesiones de
--   larga duración (`staff_session`) del área privada, más las dos
--   funciones `SECURITY DEFINER` estrechas que resuelven el tenant ANTES de
--   que exista contexto (`app.barbershop_id`): al iniciar sesión por correo
--   y al validar una cookie de sesión existente. Copia revisada de las
--   secciones A.0-A.2 de `database/modelo-fisico-referencia.sql` contra el
--   estado real aplicado (cinco migraciones: `staff_user`,
--   `idempotency_record`, `barberia_owner`/`barberia_worker` y el protocolo
--   de idempotencia). NO incluye A.3 (`staff_recovery_code`, HU-008) ni A.4
--   (`login_throttle`, HU-007): fuera de alcance de HU-005.
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-DAT-02 (nada
--   sensible en logs), DEC-024 (esquema compartido con RLS), DEC-026
--   (autenticación por correo/contraseña), DEC-040 (modelo de roles:
--   `barberia_owner` NOLOGIN, `barberia_migrator` INHERIT miembro de
--   `barberia_owner`, `barberia_app`/`barberia_worker` separados), DEC-050
--   (cookie opaca revocable, 30 días con renovación por uso), DEC-055 (login
--   público en `/api/v1/public/auth/login`), DDL-AUT-01 (sin SELECT directo
--   de `staff_credential` para `barberia_app`: solo `auth_get_credential`,
--   estrecha y revisada).
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (`INHERIT`, sin `SET ROLE`, igual que
--   `20260811145252_harden_roles_and_definer_functions.sql`). Los objetos
--   nuevos quedan owned por `barberia_migrator`; las políticas RLS
--   administrativas se escriben `FOR ALL TO barberia_owner` y
--   `barberia_migrator` las satisface por membresía heredada, exactamente el
--   patrón ya probado contra PostgreSQL real por la migración de
--   endurecimiento de roles. Ninguna función `SECURITY DEFINER` de este
--   archivo devuelve material secreto ni dato personal por sí sola: solo
--   identificadores de barbería o, para `auth_get_credential`, el hash ya
--   codificado (nunca la contraseña) de UN único usuario ya conocido dentro
--   del tenant vigente.
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando esta
--   (DEC-036). No existe archivo `down`: el mecanismo normal es roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_migrator') THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260811145252_harden_roles_and_definer_functions.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'staff_user'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260807170000_create_tenant_foundation.sql (staff_user); '
      'esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. A.0 · Resolver el tenant de un correo de acceso ANTES de tener contexto
-- ---------------------------------------------------------------------------
--
-- RLS forzada exige `app.barbershop_id` fijado, pero el inicio de sesión debe
-- encontrar al usuario a partir del correo cuando todavía no se sabe a qué
-- barbería pertenece. La salida es una función `SECURITY DEFINER` mínima que
-- devuelve exclusivamente el identificador de barbería (o NULL). No expone el
-- hash de la contraseña ni ningún dato personal: solo permite abrir la
-- transacción con el contexto correcto y volver a la ruta protegida por RLS
-- (estandar-base-datos.md §9.10).

CREATE OR REPLACE FUNCTION authn_resolve_login_tenant(p_email text)
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT barbershop_id
  FROM public.staff_user
  WHERE email = pg_catalog.lower(p_email)
    AND is_active
$$;

REVOKE ALL     ON FUNCTION authn_resolve_login_tenant(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION authn_resolve_login_tenant(text) TO barberia_app;

COMMENT ON FUNCTION authn_resolve_login_tenant(text) IS
  'SECURITY DEFINER acotada: resuelve la barbería de un correo de acceso para poder '
  'fijar app.barbershop_id. Devuelve NULL si no existe o está inactivo; quien la llama '
  'debe responder igual en ambos casos y consumir el mismo tiempo (CA-005-02, CA-005-07). '
  'No expone credenciales ni datos personales.';

-- ---------------------------------------------------------------------------
-- 3. A.1 · `staff_credential`
-- ---------------------------------------------------------------------------
-- La credencial vive separada de `staff_user` para que ninguna consulta
-- ordinaria de perfil arrastre el material secreto.

CREATE TABLE staff_credential (
  staff_user_id       uuid        NOT NULL,
  barbershop_id       uuid        NOT NULL,
  password_hash       text        NOT NULL,
  password_algorithm  text        NOT NULL DEFAULT 'argon2id',
  password_updated_at timestamptz NOT NULL DEFAULT now(),
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT staff_credential_staff_user_id_pk PRIMARY KEY (staff_user_id),

  -- Composición tenant-aware: la FK simple por staff_user_id no demostraría
  -- que el usuario pertenece a esta barbería (estandar-base-datos.md §6).
  CONSTRAINT staff_credential_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  -- CASCADE autorizado: la credencial no tiene vida ni retención propias.
  CONSTRAINT staff_credential_password_hash_ck CHECK (
    char_length(password_hash) BETWEEN 32 AND 512
    -- El hash codificado de argon2id incluye sal y parámetros; no existe
    -- columna de sal separada a propósito (CA-005-03).
  ),
  CONSTRAINT staff_credential_password_algorithm_ck CHECK (
    password_algorithm IN ('argon2id', 'bcrypt')
  )
);

COMMENT ON TABLE staff_credential IS
  'Material de autenticación del área privada. Propietario funcional: barbería. '
  'Retención: mientras exista el usuario. Clasificación: secreto. Nunca se lee en '
  'consultas de listado ni se registra.';

CREATE TRIGGER staff_credential_set_updated_at
  BEFORE UPDATE ON staff_credential
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

ALTER TABLE staff_credential ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_credential FORCE  ROW LEVEL SECURITY;

CREATE POLICY staff_credential_all_admin_policy ON staff_credential
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY staff_credential_insert_tenant_policy ON staff_credential
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_credential_update_tenant_policy ON staff_credential
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- DDL-AUT-01: sin política ni GRANT de SELECT para barberia_app. Un `SELECT *
-- FROM staff_credential` directo (por ejemplo, desde una inyección SQL)
-- deja de ser posible; la lectura pasa por auth_get_credential(), que
-- devuelve el material de un único usuario, no de toda la barbería.
GRANT INSERT, UPDATE ON TABLE staff_credential TO barberia_app;

-- Lectura estrecha y revisada (DDL-AUT-01): un único staff_user_id, siempre
-- dentro del tenant vigente. No hay forma de pedir "todas las credenciales".
CREATE OR REPLACE FUNCTION auth_get_credential(p_staff_user_id uuid)
RETURNS TABLE (password_hash text, password_algorithm text)
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT c.password_hash, c.password_algorithm
  FROM public.staff_credential c
  WHERE c.staff_user_id = p_staff_user_id
    AND c.barbershop_id = current_setting('app.barbershop_id')::uuid
$$;

REVOKE ALL     ON FUNCTION auth_get_credential(uuid) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION auth_get_credential(uuid) TO barberia_app;

COMMENT ON FUNCTION auth_get_credential(uuid) IS
  'Lectura estrecha de material de autenticación (DDL-AUT-01): un único staff_user_id, '
  'acotado al tenant vigente por app.barbershop_id. Sustituye el SELECT directo sobre '
  'staff_credential, que queda sin GRANT ni política de SELECT para barberia_app.';

-- ---------------------------------------------------------------------------
-- 4. A.2 · `staff_session`
-- ---------------------------------------------------------------------------
-- Token opaco revocable en base de datos (DEC-050): la única forma compatible
-- con la revocación inmediata que exigirá HU-006. Vigencia de 30 días desde
-- el último uso; la aplicación extiende `expires_at` en cada solicitud
-- autenticada (renovación deslizante, fuera de alcance de HU-005). El plazo
-- vive en configuración de la aplicación, no en una columna: `expires_at` ya
-- lo expresa por fila.

CREATE TABLE staff_session (
  id            uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id uuid        NOT NULL,
  staff_user_id uuid        NOT NULL,
  token_hash    text        NOT NULL,
  issued_at     timestamptz NOT NULL DEFAULT now(),
  expires_at    timestamptz NOT NULL,
  last_used_at  timestamptz NOT NULL DEFAULT now(),
  revoked_at    timestamptz,

  CONSTRAINT staff_session_id_pk PRIMARY KEY (id),

  CONSTRAINT staff_session_barbershop_id_staff_user_id_fk
    FOREIGN KEY (barbershop_id, staff_user_id)
    REFERENCES staff_user (barbershop_id, id) ON DELETE CASCADE,

  -- Unicidad GLOBAL, por el mismo motivo que el correo de acceso: la sesión
  -- se resuelve antes de conocer la barbería.
  CONSTRAINT staff_session_token_hash_uk UNIQUE (token_hash),

  CONSTRAINT staff_session_token_hash_ck   CHECK (char_length(token_hash) = 64),
  CONSTRAINT staff_session_expires_at_ck   CHECK (expires_at > issued_at),
  CONSTRAINT staff_session_last_used_at_ck CHECK (last_used_at >= issued_at),
  CONSTRAINT staff_session_revoked_at_ck   CHECK (revoked_at IS NULL OR revoked_at >= issued_at)
);

COMMENT ON TABLE staff_session IS
  'Sesiones vigentes del área privada. Propietario funcional: barbería. '
  'Retención: hasta expires_at o revoked_at + ventana de auditoría. Clasificación: secreto. '
  'Se almacena solo el hash SHA-256 del token opaco; el valor original vive únicamente en '
  'la cookie del navegador.';

-- Consulta de "mis sesiones activas" (HU-006), revocación masiva al cambiar
-- contraseña (HU-008) y auditoría de sesiones vencidas/revocadas del mismo
-- usuario: sin predicado parcial a propósito. Un índice parcial `WHERE
-- revoked_at IS NULL` junto a este sería redundante -el planificador ya sirve
-- la consulta de sesiones activas con este mismo índice completo, filtrando
-- revoked_at en el mismo escaneo- y crearía dos índices que cubren el mismo
-- prefijo.
CREATE INDEX idx_staff_session_shop_user
  ON staff_session (barbershop_id, staff_user_id);

-- Limpieza de sesiones vencidas.
CREATE INDEX idx_staff_session_expires_at ON staff_session (expires_at);

ALTER TABLE staff_session ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_session FORCE  ROW LEVEL SECURITY;

CREATE POLICY staff_session_all_admin_policy ON staff_session
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);

CREATE POLICY staff_session_select_tenant_policy ON staff_session
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_session_insert_tenant_policy ON staff_session
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_session_update_tenant_policy ON staff_session
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_session_delete_tenant_policy ON staff_session
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE staff_session TO barberia_app;

-- staff_session NO lleva trigger de set_updated_at: no tiene columna
-- updated_at. last_used_at cumple ese rol para esta tabla y lo actualiza la
-- aplicación explícitamente en cada uso válido (HU-006), no un disparador.

-- Resolución de sesión previa al contexto, análoga a la sección 2 (A.0):
-- válida solo cuando el token no está revocado y no ha vencido. HU-005 no
-- consume esta función todavía (no valida cookies entrantes: eso es
-- HU-006); se crea aquí porque pertenece a la sección A.2 del modelo de
-- referencia, igual que la tabla que resuelve.
CREATE OR REPLACE FUNCTION authn_resolve_session_tenant(p_token_hash text)
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT barbershop_id
  FROM public.staff_session
  WHERE token_hash = p_token_hash
    AND revoked_at IS NULL
    AND expires_at > pg_catalog.now()
$$;

REVOKE ALL     ON FUNCTION authn_resolve_session_tenant(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION authn_resolve_session_tenant(text) TO barberia_app;

COMMENT ON FUNCTION authn_resolve_session_tenant(text) IS
  'SECURITY DEFINER acotada: resuelve la barbería de un token de sesión ya hasheado, '
  'sin exponer staff_user_id ni ninguna otra columna. Devuelve NULL para un token '
  'desconocido, revocado o vencido; HU-006 la usará para fijar app.barbershop_id '
  'antes de reabrir la transacción tenant-aware que confirma sesión, usuario y barbería.';
