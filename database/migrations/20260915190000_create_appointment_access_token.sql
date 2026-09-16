-- Propósito
--   HU-097: agrega `appointment_access_token` (F-PUB-07), la credencial del
--   enlace aleatorio largo con el que el cliente público consultará y
--   cancelará su turno más adelante (HU-098/HU-099). Esta migración SOLO
--   crea la tabla y persiste el token que HU-097 emite dentro de la misma
--   transacción atómica de confirmación (DEC-089, resuelve `DP-PUB-05`); la
--   función de resolución pública por token (`public_resolve_appointment_
--   token_tenant`, sección E.2 del modelo de referencia) y su consumo por
--   HTTP son alcance de HU-098, que aún no existe.
--
-- Reglas y decisiones
--   RN-DAT-01/RN-DAT-02 (minimización, sin dato personal en texto plano),
--   RN-DAT-03 (revocación al anonimizar), RN-TEN-01 (aislamiento por
--   barbería), DEC-022 (token de acceso al turno), DEC-024 (esquema
--   compartido con RLS), DEC-035/DEC-036 (estándar y roles de Atlas),
--   DEC-089 (resolución de `DP-PUB-05`: 256 bits `crypto/rand`, solo se
--   almacena `SHA-256(token)`, vigencia 90 días sin rotación, revocación al
--   anonimizar o cancelar, emisión única dentro de la transacción de
--   HU-097).
--
-- Relación con database/modelo-fisico-referencia.sql §E.1
--   Copia literal de la tabla, su comentario, su índice, RLS y grants: no
--   hay diferencias de fondo entre el modelo de referencia y esta
--   migración. La sección E.2 (funciones de resolución pública por token)
--   NO se incluye aquí: pertenece a HU-098, que la agregará con su propia
--   migración cuando exista esa historia.
--
-- Qué NO hace esta migración
--   No agrega la función `public_resolve_appointment_token_tenant` ni
--   ningún endpoint de lectura/cancelación por token (HU-098/HU-099). No
--   modifica `appointment`, `customer` ni ninguna tabla existente. No
--   agrega notificaciones, recordatorios ni ninguna tabla de B5.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). La política
--   administrativa se escribe `FOR ALL TO barberia_owner`: por membresía,
--   `barberia_migrator` la satisface sin que el objeto quede literalmente
--   owned por `barberia_owner` (mismo criterio documentado en
--   20260826100000_create_time_block.sql).
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
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260827110000_create_appointment_core.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment_access_token'
  ) THEN
    RAISE EXCEPTION 'appointment_access_token ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. `appointment_access_token` — F-PUB-07, DEC-022, DEC-089
-- ---------------------------------------------------------------------------

CREATE TABLE appointment_access_token (
  id             uuid        NOT NULL DEFAULT gen_random_uuid(),
  barbershop_id  uuid        NOT NULL,
  appointment_id uuid        NOT NULL,
  token_hash     text        NOT NULL,
  issued_at      timestamptz NOT NULL DEFAULT now(),
  expires_at     timestamptz,
  revoked_at     timestamptz,

  CONSTRAINT appointment_access_token_id_pk PRIMARY KEY (id),

  CONSTRAINT appointment_access_token_barbershop_id_appointment_id_fk
    FOREIGN KEY (barbershop_id, appointment_id)
    REFERENCES appointment (barbershop_id, id) ON DELETE RESTRICT,

  -- Unicidad global: el enlace se resuelve antes de conocer la barbería
  -- (mismo criterio que barbershop.public_slug).
  CONSTRAINT appointment_access_token_token_hash_uk UNIQUE (token_hash),
  -- DDL-VAL-01: hexadecimal minúscula de 64 caracteres (SHA-256), no solo longitud.
  CONSTRAINT appointment_access_token_token_hash_ck CHECK (token_hash ~ '^[0-9a-f]{64}$'),
  CONSTRAINT appointment_access_token_expires_at_ck CHECK (
    expires_at IS NULL OR expires_at > issued_at
  ),
  -- DDL-TMP-01: la revocación no puede preceder a la emisión.
  CONSTRAINT appointment_access_token_revoked_at_ck CHECK (
    revoked_at IS NULL OR revoked_at >= issued_at
  )
);

COMMENT ON TABLE appointment_access_token IS
  'Token del enlace aleatorio largo con el que el cliente consulta o cancela su turno (DEC-022, '
  'DEC-089). Propietario funcional: barbería. Retención: hasta la anonimización de la cita, que lo '
  'revoca (RN-DAT-03). Clasificación: secreto. Solo se almacena el hash; el valor viaja una sola vez, '
  'por correo, dentro de la transacción de confirmación de HU-097.';

-- DDL-PER-01: sin predicado parcial. Un `WHERE revoked_at IS NULL` aparte
-- duplicaría el prefijo; este índice completo sirve tanto la resolución
-- pública futura (HU-098, busca uno vigente) como la auditoría de tokens ya
-- revocados de una cita.
CREATE INDEX idx_appointment_access_token_shop_appointment
  ON appointment_access_token (barbershop_id, appointment_id);

ALTER TABLE appointment_access_token ENABLE ROW LEVEL SECURITY;
ALTER TABLE appointment_access_token FORCE  ROW LEVEL SECURITY;

CREATE POLICY appointment_access_token_all_admin_policy ON appointment_access_token
  FOR ALL TO barberia_owner USING (true) WITH CHECK (true);
CREATE POLICY appointment_access_token_select_tenant_policy ON appointment_access_token
  FOR SELECT TO barberia_app USING (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_access_token_insert_tenant_policy ON appointment_access_token
  FOR INSERT TO barberia_app WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);
CREATE POLICY appointment_access_token_update_tenant_policy ON appointment_access_token
  FOR UPDATE TO barberia_app
  USING      (barbershop_id = current_setting('app.barbershop_id')::uuid)
  WITH CHECK (barbershop_id = current_setting('app.barbershop_id')::uuid);

-- Sin DELETE: RN-DAT-03 exige revocación (UPDATE revoked_at), nunca borrado
-- físico -mismo criterio que appointment/customer.
GRANT SELECT, INSERT, UPDATE ON TABLE appointment_access_token TO barberia_app;
