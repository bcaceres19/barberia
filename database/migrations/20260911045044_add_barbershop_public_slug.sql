-- Propósito
--   HU-090: agrega `barbershop.public_slug`, el identificador del enlace
--   público de reservas (F-PUB-01). Se genera automáticamente desde
--   `barbershop.name` (nunca lo escribe el barbero), es único GLOBALMENTE
--   (se resuelve antes de que exista contexto de tenant) y se compara sin
--   distinguir mayúsculas. Agrega también `public_resolve_barbershop_by_slug`,
--   la función SECURITY DEFINER estrecha que resuelve un slug a un
--   identificador de barbería sin RLS, siguiendo el mismo patrón ya
--   establecido por `authn_resolve_login_tenant`/`authn_resolve_session_tenant`
--   (20260813120000_create_auth_credentials_and_sessions.sql).
--
-- Reglas y decisiones
--   RN-TEN-01 (ninguna barbería accede a datos de otra), RN-DAT-01/RN-DAT-02
--   (minimización, registros técnicos sin datos personales), DEC-024
--   (esquema compartido con RLS), DEC-035 (estándar de base de datos),
--   DEC-036 (Atlas, migración inmutable), DEC-082 (resolución de
--   `DP-PUB-01`: generación automática, edición con ruptura del enlace
--   anterior sin redirección, respuesta uniforme al deshabilitar).
--
-- Relación con database/modelo-fisico-referencia.sql §B.1/§E.3
--   Adopta el formato (`^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$`), el índice
--   único parcial sobre `lower(public_slug)` y la firma de
--   `public_resolve_barbershop_by_slug` ya propuestos allí como insumo de
--   diseño. Difiere en un punto: aquí se hace el backfill determinista de
--   las barberías ya existentes (el modelo de referencia no lo necesitaba
--   por ser insumo de diseño, no una migración aplicada sobre datos reales).
--
-- Qué NO hace esta migración
--   No agrega ninguna capacidad de edición del slug desde HU-020 ni
--   ninguna otra pantalla (fuera del alcance de HU-090); la generación
--   automática al guardar el nombre por primera vez vive en
--   internal/modules/shops (Go), no aquí. No agrega `is_active`,
--   catálogo, barberos, disponibilidad, formulario, cita ni cancelación
--   pública (HU-091 en adelante). No agrega ninguna tabla nueva.
--
-- Condición de seguridad
--   Se ejecuta con `barberia_migrator`, que hereda los privilegios de
--   `barberia_owner` por membresía (INHERIT, sin SET ROLE,
--   20260811145252_harden_roles_and_definer_functions.sql). El backfill de
--   la sección 3 lee y escribe TODAS las barberías porque
--   `barbershop_all_admin_policy` ya cubre a `barberia_migrator`
--   (`FOR ALL TO barberia_migrator USING (true) WITH CHECK (true)`,
--   20260807170000_create_tenant_foundation.sql); no es un `BYPASSRLS`
--   nuevo, es la misma vía administrativa ya usada por seeds y
--   correcciones auditadas.
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
    WHERE table_schema = 'public' AND table_name = 'barbershop'
  ) THEN
    RAISE EXCEPTION
      'Esta migración depende de 20260807170000_create_tenant_foundation.sql; '
      'esa migración debe estar aplicada antes.';
  END IF;

  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'barbershop' AND column_name = 'public_slug'
  ) THEN
    RAISE EXCEPTION 'barbershop.public_slug ya existe: esta migración no puede aplicarse dos veces.';
  END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. Columna, formato y unicidad global
-- ---------------------------------------------------------------------------

ALTER TABLE barbershop
  ADD COLUMN public_slug text;

-- Minúsculas, dígitos y guiones internos; nunca empieza ni termina en
-- guion; 3 a 40 caracteres en total (DEC-082). NULL significa "todavía sin
-- generar o intencionalmente retirado": no es publicable (CA-090-02).
ALTER TABLE barbershop
  ADD CONSTRAINT barbershop_public_slug_ck CHECK (
    public_slug IS NULL OR public_slug ~ '^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$'
  );

-- Índice único en lugar de restricción de tabla: permite comparar el
-- identificador público sin distinguir mayúsculas. Único GLOBAL (no
-- incluye barbershop_id): el enlace público se resuelve ANTES de que
-- exista contexto de tenant, así que dos barberías nunca pueden competir
-- por el mismo slug dentro de un tenant que todavía no se conoce.
CREATE UNIQUE INDEX idx_barbershop_public_slug
  ON barbershop (lower(public_slug))
  WHERE public_slug IS NOT NULL;

COMMENT ON COLUMN barbershop.public_slug IS
  'Identificador del enlace público de reservas (F-PUB-01, HU-090). Generado automáticamente '
  'desde barbershop.name la primera vez que se guarda (internal/modules/shops, DEC-082); el '
  'barbero no lo escribe. Editable desde configuración: al cambiar, el valor anterior deja de '
  'resolver, sin redirección (DEC-082). NULL = no publicable, misma respuesta uniforme que un '
  'slug desconocido (CA-090-02). Único GLOBALMENTE, comparado sin distinguir mayúsculas '
  '(idx_barbershop_public_slug). Propietario funcional: barbería. Retención: mientras exista la '
  'cuenta. Clasificación: dato de negocio, no personal.';

-- ---------------------------------------------------------------------------
-- 3. Backfill determinista de las barberías ya existentes
-- ---------------------------------------------------------------------------

-- Cada fila con public_slug NULL recibe un slug calculado a partir de su
-- nombre vigente: minúsculas, vocales acentuadas/diéresis/ñ/cedilla
-- plegadas a ASCII (translate(), sin extensión unaccent -AGENTS.md prohíbe
-- una dependencia nueva sin justificación y translate() ya es núcleo de
-- PostgreSQL), cualquier otro carácter fuera de [a-z0-9] colapsado a un
-- solo guion, sin guion inicial/final, truncado a 40. Un resultado vacío o
-- menor de 3 caracteres (nombre sin ningún alfanumérico latino, caso
-- patológico) cae al respaldo determinista 'barberia'. Una colisión (dos
-- nombres que producen el mismo slug) se resuelve con sufijo numérico
-- '-2', '-3', ... determinista por orden de creación (DEC-082). Va como
-- barberia_migrator, que sí ve todas las filas (sección "Condición de
-- seguridad" arriba); barberia_app en producción nunca podría hacer este
-- barrido cruzado de tenants.
DO $$
DECLARE
  shop        RECORD;
  base_slug   text;
  candidate   text;
  suffix      integer;
  from_chars  CONSTANT text := 'áéíóúüñÁÉÍÓÚÜÑàèìòùÀÈÌÒÙâêîôûÂÊÎÔÛäëïöÄËÏÖçÇ';
  to_chars    CONSTANT text := 'aeiouunAEIOUUNaeiouAEIOUaeiouAEIOUaeioAEIOcC';
BEGIN
  FOR shop IN
    SELECT id, name FROM barbershop WHERE public_slug IS NULL ORDER BY created_at, id
  LOOP
    base_slug := regexp_replace(
      translate(lower(shop.name), from_chars, to_chars),
      '[^a-z0-9]+', '-', 'g'
    );
    base_slug := regexp_replace(base_slug, '(^-+|-+$)', '', 'g');
    base_slug := left(base_slug, 40);
    base_slug := regexp_replace(base_slug, '-+$', '', 'g');

    IF base_slug IS NULL OR char_length(base_slug) < 3 THEN
      base_slug := 'barberia';
    END IF;

    candidate := base_slug;
    suffix := 2;
    WHILE EXISTS (
      SELECT 1 FROM barbershop
      WHERE lower(public_slug) = lower(candidate) AND id <> shop.id
    ) LOOP
      candidate := left(base_slug, 40 - char_length('-' || suffix::text)) || '-' || suffix::text;
      suffix := suffix + 1;
    END LOOP;

    UPDATE barbershop SET public_slug = candidate WHERE id = shop.id;
  END LOOP;
END
$$;

-- ---------------------------------------------------------------------------
-- 4. Resolver el tenant de un slug público ANTES de tener contexto
-- ---------------------------------------------------------------------------
--
-- Mismo patrón que authn_resolve_login_tenant/authn_resolve_session_tenant
-- (20260813120000_create_auth_credentials_and_sessions.sql): una visita
-- pública llega sin sesión y sin `app.barbershop_id`. La función devuelve
-- exclusivamente el uuid de la barbería (o NULL); nunca nombre, contacto
-- ni ningún otro dato -quien la llama vuelve a la ruta protegida por RLS
-- (InTenantTx/ResolveTenant, estandar-base-datos.md §9.10) para leer lo
-- que CA-090-04 autoriza a mostrar-.
CREATE OR REPLACE FUNCTION public_resolve_barbershop_by_slug(p_slug text)
RETURNS uuid
LANGUAGE sql
STABLE
SECURITY DEFINER
SET search_path = ''
AS $$
  SELECT id
  FROM public.barbershop
  WHERE public_slug IS NOT NULL
    AND pg_catalog.lower(public_slug) = pg_catalog.lower(p_slug)
$$;

REVOKE ALL     ON FUNCTION public_resolve_barbershop_by_slug(text) FROM PUBLIC;
GRANT  EXECUTE ON FUNCTION public_resolve_barbershop_by_slug(text) TO barberia_app;

COMMENT ON FUNCTION public_resolve_barbershop_by_slug(text) IS
  'SECURITY DEFINER acotada: resuelve el slug del enlace público de reservas (HU-090) a un '
  'identificador de barbería para poder fijar app.barbershop_id. Devuelve NULL si el slug no '
  'existe o no es publicable (public_slug NULL); quien la llama responde igual en ambos casos '
  '(CA-090-02, RN-TEN-01). No expone nombre, contacto ni ningún otro dato.';
