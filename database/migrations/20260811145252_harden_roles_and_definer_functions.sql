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
-- Por qué no se usa `SET ROLE` / `RESET ROLE`
--   Se intentó primero un diseño con `SET ROLE barberia_owner; ... RESET
--   ROLE;` alrededor de las operaciones que exigen ser propietario. Probado
--   contra PostgreSQL real, Atlas escribe su propio registro de progreso en
--   `atlas_schema_revisions` (un esquema que solo puede ver quien lo creó,
--   normalmente `barberia_migrator`) DURANTE la aplicación de la migración,
--   no solo al final; con `SET ROLE` activo, esa escritura de Atlas falla
--   con "permission denied for schema atlas_schema_revisions" y toda la
--   migración se revierte. La solución robusta es que `barberia_migrator`
--   herede los privilegios de `barberia_owner` de forma automática
--   (`INHERIT` + membresía), sin cambiar de rol nunca: eso alcanza tanto
--   para las operaciones que exigen ser propietario (confirmado: pertenecer
--   con herencia basta, igual que para las políticas RLS que apuntan al
--   propietario) como para que Atlas siga escribiendo su propio registro
--   sin fricción.
--
-- Plan de avance
--   Toda migración futura que cree objetos nuevos los deja owned por
--   `barberia_migrator` (quien la ejecuta), no por `barberia_owner`; es
--   aceptable porque `barberia_migrator` es el único rol que corre
--   migraciones (reproducible por diseño) y hereda los privilegios de
--   `barberia_owner`. Las políticas RLS administrativas de una tabla nueva
--   se escriben `FOR ALL TO barberia_owner`, igual que las de aquí: por
--   membresía, `barberia_migrator` las satisface sin necesitar que el
--   objeto quede literalmente owned por `barberia_owner`.

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

-- barberia_migrator pasa de NOINHERIT (migración original) a INHERIT: con
-- membresía en barberia_owner, adquiere sus privilegios automáticamente en
-- cada sesión, sin `SET ROLE` (confirmado en PostgreSQL real: necesario
-- para no romper el registro de progreso de Atlas, ver encabezado). Sigue
-- sin `SUPERUSER`, `CREATEDB`, `CREATEROLE` ni `BYPASSRLS` propios: todo lo
-- que gana es exactamente lo que barberia_owner tiene.
ALTER ROLE barberia_migrator INHERIT;
GRANT barberia_owner TO barberia_migrator;

-- El worker solo invoca funciones SECURITY DEFINER con nombres ya
-- calificados (public.*, pg_catalog.*); no necesita resolver nada por
-- search_path.
ALTER ROLE barberia_worker SET search_path = '';

-- `REVOKE CREATE ON SCHEMA public FROM PUBLIC` ya lo hizo la migración
-- original; aquí solo se concede USAGE al rol nuevo.
GRANT USAGE ON SCHEMA public TO barberia_worker;

-- Poseer una tabla no implica poder verla: el propietario también necesita
-- USAGE sobre el esquema que la contiene (hallazgo confirmado en PostgreSQL
-- real: sin esta línea, ni siquiera puede hacer SELECT de sus propias
-- tablas). CREATE porque las migraciones futuras, ejecutadas por
-- barberia_migrator con los privilegios heredados de barberia_owner, crean
-- objetos nuevos en este esquema.
GRANT USAGE, CREATE ON SCHEMA public TO barberia_owner;

-- ---------------------------------------------------------------------------
-- 3. Transferencia de ownership (DDL-SEC-01)
-- ---------------------------------------------------------------------------

-- Transferir ownership solo exige poseer el objeto (barberia_migrator, como
-- lo creó la migración original) y ser miembro del rol destino.
ALTER TABLE    barbershop         OWNER TO barberia_owner;
ALTER TABLE    staff_user         OWNER TO barberia_owner;
ALTER TABLE    idempotency_record OWNER TO barberia_owner;
ALTER FUNCTION set_updated_at()   OWNER TO barberia_owner;

-- ---------------------------------------------------------------------------
-- 4. Políticas administrativas: de barberia_migrator a barberia_owner
-- ---------------------------------------------------------------------------

-- FORCE ROW LEVEL SECURITY alcanza también al propietario. Ahora que
-- barberia_owner es el propietario real, las políticas "admin" deben
-- apuntarle a él. barberia_migrator las sigue satisfaciendo por membresía
-- heredada (confirmado en PostgreSQL real), sin necesitar `SET ROLE`.

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

-- Los privilegios por defecto se indexan por el rol que CREA el objeto, no
-- por quien hereda sus privilegios (confirmado en PostgreSQL real: una
-- función creada por barberia_migrator, aunque herede de barberia_owner,
-- seguía concediendo EXECUTE a PUBLIC porque el baseline solo cubría a
-- barberia_owner). Como barberia_migrator es quien de hecho crea los objetos
-- de las migraciones futuras, el baseline debe fijarse para ambos roles.
--
-- Sin `IN SCHEMA`, no con `IN SCHEMA public`: probado en PostgreSQL real, la
-- forma acotada a un esquema no se aplicó a una función nueva creada sin
-- calificar el esquema (`CREATE FUNCTION f()`, el caso real de este
-- proyecto), mientras que la forma global sí. La forma global es además más
-- robusta: cubre cualquier esquema futuro, no solo `public`. Esta línea es
-- un refuerzo adicional, no la única defensa: cada función SECURITY DEFINER
-- sigue llevando su propio `REVOKE ALL ... FROM PUBLIC` explícito
-- inmediatamente después de crearse (ver modelo-fisico-referencia.sql), que
-- es la protección primaria y sí probada de forma aislada por objeto.
ALTER DEFAULT PRIVILEGES FOR ROLE barberia_owner
  REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
ALTER DEFAULT PRIVILEGES FOR ROLE barberia_migrator
  REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;

-- `set_updated_at()` es un disparador interno; el API nunca lo invoca de
-- forma directa (solo dispara UPDATE sobre tablas que ya tienen el trigger
-- creado por el propietario). PostgreSQL no vuelve a comprobar EXECUTE del
-- rol que dispara un trigger en tiempo de ejecución, solo lo comprobó al
-- crear el trigger: confirmado en PostgreSQL real (INSERT + UPDATE como
-- barberia_app tras esta revocación, updated_at sí cambió).
REVOKE EXECUTE ON FUNCTION set_updated_at() FROM barberia_app;
