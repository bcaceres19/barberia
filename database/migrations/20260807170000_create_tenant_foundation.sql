-- Propósito
--   Andamiaje del esquema multi-tenant exigido por HU-001: extensiones, roles
--   separados, tabla `barbershop`, tabla de usuarios del área privada
--   (`staff_user`), RLS habilitada y forzada, y políticas de lectura/escritura
--   contra `app.barbershop_id`.
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-DIS-07 (todo
--   instante con zona horaria), DEC-024 (esquema compartido con RLS),
--   DEC-035 (estándar de base de datos), DEC-036 (Atlas, migración inmutable).
--
-- Condición de seguridad
--   Se ejecuta con el rol migrador, que queda como propietario de los objetos.
--   El rol de aplicación se crea SIN contraseña, SIN BYPASSRLS y SIN propiedad
--   sobre las tablas; su contraseña se fija fuera del repositorio con
--   `ALTER ROLE barberia_app PASSWORD ...` desde el secreto del ambiente.
--
-- Plan de avance
--   Una corrección posterior se hace con otra migración, nunca editando esta
--   (DEC-036). No existe archivo `down`: el mecanismo normal es roll-forward.

-- ---------------------------------------------------------------------------
-- 1. Precondiciones
-- ---------------------------------------------------------------------------

-- `FORCE ROW LEVEL SECURITY`, `gen_random_uuid()` en el núcleo y `btree_gist`
-- sobre `uuid` requieren PostgreSQL 14 o superior. Fallar aquí es preferible a
-- aplicar la mitad del esquema sobre un motor que no puede sostener la
-- restricción de exclusión de B3.
DO $$
BEGIN
  IF current_setting('server_version_num')::integer < 140000 THEN
    RAISE EXCEPTION
      'Se requiere PostgreSQL 14 o superior; el servidor reporta %.',
      current_setting('server_version');
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Extensiones
-- ---------------------------------------------------------------------------

-- `btree_gist` permite combinar igualdad sobre `uuid` con solapamiento de
-- rangos en la restricción de exclusión de `appointment` (base-datos.md §3).
-- Se crea aquí, con el resto del andamiaje, porque crear extensiones exige
-- privilegios que el rol de aplicación no tiene y no debe repetirse por bloque.
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- ---------------------------------------------------------------------------
-- 3. Roles
-- ---------------------------------------------------------------------------

-- `CREATE ROLE` no admite `IF NOT EXISTS`; el bloque hace la migración
-- repetible sobre una base donde los roles ya fueron aprovisionados
-- (CA-001-05).
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_migrator') THEN
    CREATE ROLE barberia_migrator
      LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app') THEN
    CREATE ROLE barberia_app
      LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
  END IF;
END
$$;

-- Verificación explícita de CA-001-04: si alguien elevó el rol de aplicación
-- fuera de las migraciones, la migración se detiene en lugar de legitimarlo.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_roles
    WHERE rolname = 'barberia_app'
      AND (rolsuper OR rolbypassrls OR rolcreaterole OR rolcreatedb)
  ) THEN
    RAISE EXCEPTION
      'barberia_app no puede ser superusuario ni tener BYPASSRLS (CA-001-04).';
  END IF;
END
$$;

-- `search_path` explícito por rol: ningún objeto se resuelve desde un esquema
-- escribible por terceros (estandar-base-datos.md §9.9).
ALTER ROLE barberia_migrator SET search_path = public, pg_catalog;
ALTER ROLE barberia_app      SET search_path = public, pg_catalog;

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT  USAGE  ON SCHEMA public TO barberia_app;

-- ---------------------------------------------------------------------------
-- 4. Función auxiliar de `updated_at`
-- ---------------------------------------------------------------------------

-- Un disparador es la única forma de garantizar `updated_at` frente a
-- cualquier escritura, incluida una corrección manual del rol migrador. Es la
-- excepción declarada al criterio de "restricciones antes que triggers"
-- (estandar-base-datos.md §7): no es expresable de forma declarativa.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $$
BEGIN
  NEW.updated_at := now();
  RETURN NEW;
END;
$$;

REVOKE ALL     ON FUNCTION set_updated_at() FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION set_updated_at() TO barberia_app;

COMMENT ON FUNCTION set_updated_at() IS
  'Fija updated_at en cada UPDATE. Sin lógica de negocio: solo marca de tiempo.';

-- ---------------------------------------------------------------------------
-- 5. Tabla `barbershop`
-- ---------------------------------------------------------------------------

CREATE TABLE barbershop (
  id          uuid        NOT NULL DEFAULT gen_random_uuid(),
  name        text        NOT NULL,
  timezone    text        NOT NULL DEFAULT 'America/Bogota',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT barbershop_id_pk       PRIMARY KEY (id),
  CONSTRAINT barbershop_name_ck     CHECK (btrim(name) <> '' AND char_length(name) <= 120),
  CONSTRAINT barbershop_timezone_ck CHECK (btrim(timezone) <> '' AND char_length(timezone) <= 64)
);

COMMENT ON TABLE barbershop IS
  'Unidad de aislamiento del sistema (DEC-024). Propietario funcional: propietario del proyecto. '
  'Retención: mientras exista la cuenta. Clasificación: datos de negocio, sin datos personales de clientes.';

COMMENT ON COLUMN barbershop.timezone IS
  'Zona IANA en la que se presenta toda hora de esta barbería (RN-DIS-07). '
  'La validación contra pg_timezone_names no puede expresarse en un CHECK '
  '(no es inmutable); la aplica el servicio de configuración y la verifica su prueba.';

CREATE TRIGGER barbershop_set_updated_at
  BEFORE UPDATE ON barbershop
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 6. Tabla `staff_user`
-- ---------------------------------------------------------------------------

CREATE TABLE staff_user (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  email          text        NOT NULL,
  full_name      text        NOT NULL,
  is_active      boolean     NOT NULL DEFAULT true,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT staff_user_id_pk                PRIMARY KEY (id),
  CONSTRAINT staff_user_barbershop_id_fk     FOREIGN KEY (barbershop_id)
                                             REFERENCES barbershop (id) ON DELETE RESTRICT,
  CONSTRAINT staff_user_barbershop_id_id_uk  UNIQUE (barbershop_id, id),
  CONSTRAINT staff_user_email_uk             UNIQUE (email),
  CONSTRAINT staff_user_email_ck             CHECK (
                                               email = lower(email)
                                               AND email LIKE '_%@_%._%'
                                               AND char_length(email) <= 254
                                             ),
  CONSTRAINT staff_user_full_name_ck         CHECK (
                                               btrim(full_name) <> ''
                                               AND char_length(full_name) <= 120
                                             )
);

COMMENT ON TABLE staff_user IS
  'Usuario del área privada. Propietario funcional: barbería. Retención: mientras exista la cuenta. '
  'Clasificación: datos personales (correo, nombre).';

COMMENT ON COLUMN staff_user.email IS
  'Identidad de acceso, normalizada en minúsculas. Excepción documentada a la unicidad tenant-aware '
  '(estandar-base-datos.md §6): la unicidad es GLOBAL porque el inicio de sesión debe resolver la '
  'barbería antes de existir contexto de tenant. La unicidad dentro de una barbería queda implicada. '
  'No hay auto-registro en el MVP, por lo que la unicidad global no expone una vía de enumeración.';

COMMENT ON COLUMN staff_user.is_active IS
  'Un usuario inactivo no puede iniciar sesión ni conservar sesiones vigentes (CA-005-07).';

CREATE TRIGGER staff_user_set_updated_at
  BEFORE UPDATE ON staff_user
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- 7. Seguridad a nivel de fila
-- ---------------------------------------------------------------------------

-- `current_setting('app.barbershop_id')` sin segundo argumento lanza error
-- cuando el ajuste no fue fijado. Es deliberado: una transacción sin contexto
-- falla de forma explícita en lugar de devolver el conjunto completo
-- (CA-001-03).

ALTER TABLE barbershop ENABLE ROW LEVEL SECURITY;
ALTER TABLE barbershop FORCE  ROW LEVEL SECURITY;

ALTER TABLE staff_user ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_user FORCE  ROW LEVEL SECURITY;

-- Rol administrativo: `FORCE ROW LEVEL SECURITY` también alcanza al
-- propietario, así que el rol migrador necesita una política explícita para
-- cargar datos de referencia, semillas y correcciones auditadas
-- (estandar-base-datos.md §9.5). Es un camino único y visible, no un
-- `BYPASSRLS` implícito.
CREATE POLICY barbershop_all_admin_policy ON barbershop
  FOR ALL TO barberia_migrator
  USING (true) WITH CHECK (true);

CREATE POLICY staff_user_all_admin_policy ON staff_user
  FOR ALL TO barberia_migrator
  USING (true) WITH CHECK (true);

-- `barbershop`: la aplicación lee y actualiza su propia barbería. No se le
-- concede crear ni borrar barberías; el aprovisionamiento es administrativo.
CREATE POLICY barbershop_select_tenant_policy ON barbershop
  FOR SELECT TO barberia_app
  USING (id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY barbershop_update_tenant_policy ON barbershop
  FOR UPDATE TO barberia_app
  USING      (id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (id = current_setting('app.barbershop_id')::uuid);

-- `staff_user`: lectura y escritura completas dentro de la barbería vigente.
CREATE POLICY staff_user_select_tenant_policy ON staff_user
  FOR SELECT TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_user_insert_tenant_policy ON staff_user
  FOR INSERT TO barberia_app
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_user_update_tenant_policy ON staff_user
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

CREATE POLICY staff_user_delete_tenant_policy ON staff_user
  FOR DELETE TO barberia_app
  USING (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- ---------------------------------------------------------------------------
-- 8. Privilegios del rol de aplicación
-- ---------------------------------------------------------------------------

-- RLS limita filas; los privilegios limitan operaciones. Se necesitan ambos.
GRANT SELECT, UPDATE                 ON TABLE barbershop TO barberia_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE staff_user TO barberia_app;
