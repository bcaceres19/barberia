-- Propósito
--   HU-098: agrega `public_resolve_appointment_token_tenant`, la función
--   `SECURITY DEFINER` estrecha que resuelve el hash de un token de acceso
--   al turno (`appointment_access_token`, HU-097) a un identificador de
--   barbería, sin RLS y sin exponer ningún otro dato -mismo patrón exacto
--   que `public_resolve_barbershop_by_slug`
--   (20260911045044_add_barbershop_public_slug.sql, HU-090). Solo resuelve
--   un token vigente y no revocado: la lectura protegida por RLS de la cita
--   asociada la hace, después, una transacción tenant-aware normal
--   (InTenantTx), igual que `publicbooking.Repository.ResolveBySlug`.
--
-- Reglas y decisiones
--   RN-DAT-01/RN-DAT-02 (minimización, sin dato personal en texto plano),
--   RN-DAT-03 (una fila revocada o vencida no resuelve), RN-TEN-01
--   (aislamiento por barbería), RN-CNF-01/RN-CNF-02, DEC-022 (token de
--   acceso al turno), DEC-024 (esquema compartido con RLS), DEC-035/
--   DEC-036 (estándar y roles de Atlas), DEC-089 (resolución de
--   `DP-PUB-05`: 256 bits `crypto/rand`, solo se almacena
--   `SHA-256(token)`, vigencia 90 días sin rotación, revocación al
--   anonimizar o cancelar).
--
-- Relación con database/modelo-fisico-referencia.sql §E.2
--   Copia literal de `public_resolve_appointment_token_tenant`, su
--   `REVOKE`/`GRANT` y su comentario: no hay diferencias de fondo entre el
--   modelo de referencia y esta migración. `public_resolve_barbershop_by_slug`
--   (misma sección E.2) NO se toca aquí: ya está migrada por
--   20260911045044_add_barbershop_public_slug.sql y esta migración no la
--   redefine.
--
-- Qué NO hace esta migración
--   No agrega ningún endpoint HTTP, no modifica `appointment_access_token`
--   ni ninguna tabla existente, no implementa la cancelación del cliente
--   (HU-099) ni la anonimización que revoca tokens vencidos (worker de
--   retención, B6, `DEC-025`/`DEC-049`): esta función solo consume
--   `revoked_at`/`expires_at`, no decide quién los fija.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
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
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'appointment_access_token'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260915190000_create_appointment_access_token.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM pg_proc
    WHERE proname = 'public_resolve_appointment_token_tenant'
  ) THEN
    RAISE EXCEPTION 'public_resolve_appointment_token_tenant ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. `public_resolve_appointment_token_tenant` — F-PUB-07, DEC-089
-- ---------------------------------------------------------------------------

CREATE FUNCTION public_resolve_appointment_token_tenant(p_token_hash text)
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT barbershop_id
  FROM public.appointment_access_token
  WHERE token_hash = p_token_hash
    AND revoked_at IS NULL
    AND (expires_at IS NULL OR expires_at > pg_catalog.now())
$$;

REVOKE ALL     ON FUNCTION public_resolve_appointment_token_tenant(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION public_resolve_appointment_token_tenant(text) TO barberia_app;

COMMENT ON FUNCTION public_resolve_appointment_token_tenant(text) IS
  'SECURITY DEFINER acotada: resuelve el hash del token de acceso al turno (HU-098) a un '
  'identificador de barbería para poder fijar app.barbershop_id. Devuelve NULL si el hash no '
  'existe, ya expiró o ya fue revocado; quien la llama responde igual en los tres casos '
  '(CA-098-02, RN-TEN-01). No expone la cita, el cliente ni ningún otro dato.';
