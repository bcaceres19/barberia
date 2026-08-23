package shops

import "strings"

// NameMaxLength y TimezoneMaxLength coinciden con barbershop_name_ck y
// barbershop_timezone_ck (20260807170000_create_tenant_foundation.sql): el
// servicio no inventa un límite más estricto que la base ya aplica
// (CA-020-07).
const (
	NameMaxLength     = 120
	TimezoneMaxLength = 64
)

// Barbershop es la configuración básica de la barbería activa que HU-020
// expone: exactamente los cuatro campos autorizados (CA-020-07), nunca
// public_slug, logotipo ni ningún otro dato de F-PUB-01. ContactEmail y
// ContactPhone son nil cuando la barbería no tiene ese contacto configurado
// (ausencia real, no cadena vacía).
type Barbershop struct {
	Name         string
	Timezone     string
	ContactEmail *string
	ContactPhone *string
}

// NormalizeName recorta espacios, igual que barbershop_name_ck espera
// (btrim). No trunca ni rechaza: el largo se valida aparte.
func NormalizeName(raw string) string {
	return strings.TrimSpace(raw)
}

// NormalizeContact recorta espacios y, si el resultado queda vacío, lo
// normaliza a ausencia (nil) en vez de cadena vacía (CA-020-06): un contacto
// vacío del formulario se persiste como NULL. No cambia mayúsculas ni
// minúsculas por sí sola: el correo se normaliza aparte con
// NormalizeContactEmail, porque un teléfono E.164 no tiene ese concepto.
func NormalizeContact(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// NormalizeContactEmail aplica NormalizeContact y, cuando queda un valor,
// lo pasa a minúsculas (mismo criterio que staff_user.email /
// auth.NormalizeEmail), para que dos formularios que difieren solo en
// mayúsculas guarden el mismo valor.
func NormalizeContactEmail(raw string) *string {
	normalized := NormalizeContact(raw)
	if normalized == nil {
		return nil
	}
	lower := strings.ToLower(*normalized)
	return &lower
}
