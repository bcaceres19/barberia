-- Propósito
--   Corrección roll-forward de DDL-SEC-01/02/03/04
--   (docs/05-backend/revision-ddl-seguridad-2026-08-11.md): introduce un
--   propietario de objetos NOLOGIN reproducible, separa el rol del worker del
--   rol del API, y cierra la superficie de privilegios por defecto que
--   `20260807170000_create_tenant_foundation.sql` dejó abierta.
--
-- Reglas y decisiones
--   DEC-040 (modelo de roles), DDL-SEC-01 (ownership no reproducible),
--   DDL-SEC-02 (search_path de SECURITY DEFINER), DDL-SEC-03 (sin baseline de
--   default privileges), DDL-SEC-04 (API y worker comparten privilegios).
--
-- Qué NO hace esta migración
--   No edita `20260807170000_create_tenant_foundation.sql` ni
--   `20260807170100_create_idempotency_record.sql` (DEC-036: inmutables).
--   No rediseña `notification_claim_due` ni `retention_claim_due_customers`
--   (DDL-CON-01, DEC-042): esas funciones viven en el modelo de referencia,
--   no están aplicadas todavía, y su corrección de lease/fecha ancla
--   pertenece al issue de workers, no a este de bootstrap de roles.
--
-- Condición de seguridad
--   Igual que la migración original, esta requiere un ejecutor con
--   privilegio para crear/alterar roles y conceder sobre el esquema `public`
--   (`CREATEROLE` como mínimo, típicamente el administrador de bootstrap):
--   `barberia_migrator` no tiene `CREATEROLE` ni debería tenerlo. Es
--   exactamente el hueco que `DDL-SEC-01` señaló y que el procedimiento de
--   bootstrap de `migraciones-atlas.md` cierra hacia adelante: a partir de
--   este PR, los roles se aprovisionan con un administrador antes de que
--   Atlas se conecte, y esta es la última vez que una migración crea roles.
--   `ALTER ... OWNER TO` es idempotente: si el objeto ya pertenece a
--   `barberia_owner`, no cambia nada.
--
-- Plan de avance
--   Toda migración posterior que cree objetos nuevos debe abrir con
--   `SET ROLE barberia_owner;` y cerrar con `RESET ROLE;` (documentado en
--   migraciones-atlas.md), para que el propietario real sea siempre
--   `barberia_owner` y no `barberia_migrator`.

-- ---------------------------------------------------------------------------
-- 1. Precondición
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_migrator')
     OR NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_app') THEN
    RAISE EXCEPTION
      'Esta migración corrige 20260807170000_create_tenant_foundation.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Roles nuevos: propietario NOLOGIN y worker separado del API (DEC-040)
-- ---------------------------------------------------------------------------

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_owner') THEN
    CREATE ROLE barberia_owner
      NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'barberia_worker') THEN
    CREATE ROLE barberia_worker
      LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT NOREPLICATION NOBYPASSRLS;
  END IF;
END
$$;

-- Igual que CA-001-04 para barberia_app: si algo elevó a barberia_worker
-- fuera de esta migración, se detiene en vez de legitimarlo.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_roles
    WHERE rolname = 'barberia_worker'
      AND (rolsuper OR rolbypassrls OR rolcreaterole OR rolcreatedb)
  ) THEN
    RAISE EXCEPTION
      'barberia_worker no puede ser superusuario ni tener BYPASSRLS.';
  END IF;
END
$$;

-- barberia_migrator es NOINHERIT a propósito (migración original): no debe
-- adquirir en automático los privilegios de barberia_owner en cada consulta.
-- La membresía sí le permite transferir ownership hacia barberia_owner (esa
-- operación concreta solo exige pertenencia, no herencia activa) y le permite
-- `SET ROLE barberia_owner` explícito cuando de verdad necesite actuar como
-- propietario (sección 4 en adelante).
GRANT barberia_owner TO barberia_migrator;

-- El worker solo invoca funciones SECURITY DEFINER con nombres ya
-- calificados (public.*, pg_catalog.*); no necesita resolver nada por
-- search_path.
ALTER ROLE barberia_worker SET search_path = '';

-- `REVOKE CREATE ON SCHEMA public FROM PUBLIC` ya lo hizo la migración
-- original; aquí solo se concede USAGE al rol nuevo.
GRANT USAGE ON SCHEMA public TO barberia_worker;

-- ---------------------------------------------------------------------------
-- 3. Transferencia de ownership (DDL-SEC-01)
-- ---------------------------------------------------------------------------

-- Se ejecuta todavía como barberia_migrator (dueño actual de los objetos):
-- transferir ownership solo exige poseer el objeto y ser miembro del rol
-- destino, no exige `SET ROLE` previo.
ALTER TABLE    barbershop         OWNER TO barberia_owner;
ALTER TABLE    staff_user         OWNER TO barberia_owner;
ALTER TABLE    idempotency_record OWNER TO barberia_owner;
ALTER FUNCTION set_updated_at()   OWNER TO barberia_owner;

-- A partir de aquí los objetos ya NO pertenecen a barberia_migrator. Como es
-- NOINHERIT, necesita actuar explícitamente como barberia_owner para las
-- operaciones que siguen (gestionar políticas, fijar default privileges,
-- revocar EXECUTE de una función que ahora es suya).
SET ROLE barberia_owner;

-- ---------------------------------------------------------------------------
-- 4. Políticas administrativas: de barberia_migrator a barberia_owner
-- ---------------------------------------------------------------------------

-- FORCE ROW LEVEL SECURITY alcanza también al propietario. Ahora que
-- barberia_owner es el propietario real, las políticas "admin" deben
-- apuntarle a él, no al login migrador.

DROP POLICY barbershop_all_admin_policy         ON barbershop;
DROP POLICY staff_user_all_admin_policy         ON staff_user;
DROP POLICY idempotency_record_all_admin_policy ON idempotency_record;

CREATE POLICY barbershop_all_admin_policy ON barbershop
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY staff_user_all_admin_policy ON staff_user
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

CREATE POLICY idempotency_record_all_admin_policy ON idempotency_record
  FOR ALL TO barberia_owner
  USING (true) WITH CHECK (true);

-- ---------------------------------------------------------------------------
-- 5. Privilegios por defecto (DDL-SEC-03)
-- ---------------------------------------------------------------------------

-- Sin esta línea, PostgreSQL concede EXECUTE a PUBLIC en toda función nueva
-- creada por barberia_owner. Cierra el hueco para todo lo que se cree de aquí
-- en adelante, no solo lo ya existente.
ALTER DEFAULT PRIVILEGES FOR ROLE barberia_owner IN SCHEMA public
  REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;

-- `set_updated_at()` es un disparador interno; el API nunca lo invoca de
-- forma directa (solo dispara UPDATE sobre tablas que ya tienen el trigger
-- creado por el propietario). PostgreSQL no vuelve a comprobar EXECUTE del
-- rol que dispara un trigger en tiempo de ejecución, solo lo comprobó al
-- crear el trigger (verificado en la tarea de validación contra PostgreSQL
-- real antes de fusionar este PR; si el trigger dejara de dispararse para
-- barberia_app, esta línea se revierte en este mismo PR).
REVOKE EXECUTE ON FUNCTION set_updated_at() FROM barberia_app;

-- ---------------------------------------------------------------------------
-- 6. Volver al rol de conexión
-- ---------------------------------------------------------------------------

RESET ROLE;
