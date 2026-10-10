-- Propósito
--   Issue #353 (DEC-125): retira `auth_phone_challenge_destination`. La creó la
--   migración 20260919220000 para que un proveedor externo validara el OTP con
--   el teléfono guardado; NAVA genera y valida siempre el código (DEC-123), así
--   que ninguna capa vuelve a leer ese destino y la función solo ampliaba la
--   superficie SECURITY DEFINER.
--
-- Reglas y decisiones
--   DEC-036 (Atlas), DEC-123, DEC-125. La migración 20260919220000 es inmutable
--   y se conserva; esta la deshace hacia delante.
--
-- Impacto
--   Sin datos: es una función de solo lectura. Ningún código de la aplicación la
--   invoca desde el mismo PR. Revertir exige una migración nueva que la recree.

DROP FUNCTION IF EXISTS auth_phone_challenge_destination(text, text);
