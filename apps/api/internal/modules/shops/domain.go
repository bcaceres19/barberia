package shops

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

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

// slugFoldReplacer pliega vocales acentuadas, diéresis, ñ y cedilla a su
// equivalente ASCII antes de derivar el slug público (HU-090, DEC-082).
// Mismo criterio que el backfill SQL de
// 20260911045044_add_barbershop_public_slug.sql, sin depender de la
// extensión unaccent (AGENTS.md prohíbe una dependencia nueva sin
// justificación; translate()/strings.NewReplacer alcanzan para el alfabeto
// español).
var slugFoldReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	"à", "a", "è", "e", "ì", "i", "ò", "o", "ù", "u",
	"â", "a", "ê", "e", "î", "i", "ô", "o", "û", "u",
	"ä", "a", "ë", "e", "ï", "i", "ö", "o",
	"ç", "c",
)

var (
	slugNonAlnumPattern   = regexp.MustCompile(`[^a-z0-9]+`)
	slugTrimHyphenPattern = regexp.MustCompile(`(^-+|-+$)`)
)

const (
	// SlugMaxLength/SlugMinLength coinciden con barbershop_public_slug_ck
	// (^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$): 3 a 40 caracteres.
	SlugMaxLength = 40
	SlugMinLength = 3
	// slugFallback es el respaldo determinista cuando el nombre no aporta
	// ningún carácter latino alfanumérico (caso patológico: nombre vacío
	// tras normalizar, solo símbolos, solo un alfabeto no latino, etc.).
	slugFallback = "barberia"
)

// SlugBase deriva la base del slug público (DEC-082, resuelve DP-PUB-01) a
// partir de un nombre YA normalizado (NormalizeName): minúsculas, plegado
// ASCII de acentos, cualquier carácter fuera de [a-z0-9] colapsado a un
// solo guion, sin guion inicial/final, truncado a SlugMaxLength. Pura: no
// toca la base de datos ni decide la unicidad global -eso lo resuelve el
// repositorio reintentando con SlugWithSuffix ante un unique_violation real
// sobre idx_barbershop_public_slug-.
func SlugBase(name string) string {
	folded := slugFoldReplacer.Replace(strings.ToLower(name))
	base := slugNonAlnumPattern.ReplaceAllString(folded, "-")
	base = slugTrimHyphenPattern.ReplaceAllString(base, "")
	if utf8.RuneCountInString(base) > SlugMaxLength {
		runes := []rune(base)
		base = string(runes[:SlugMaxLength])
		base = slugTrimHyphenPattern.ReplaceAllString(base, "")
	}
	if utf8.RuneCountInString(base) < SlugMinLength {
		return slugFallback
	}
	return base
}

// SlugWithSuffix agrega el sufijo numérico determinístico que DEC-082 exige
// ante una colisión real de unicidad global ('-2', '-3', ...), recortando
// base lo necesario para que el resultado nunca exceda SlugMaxLength.
func SlugWithSuffix(base string, attempt int) string {
	suffix := fmt.Sprintf("-%d", attempt)
	maxBase := SlugMaxLength - utf8.RuneCountInString(suffix)
	if utf8.RuneCountInString(base) > maxBase {
		runes := []rune(base)
		base = string(runes[:maxBase])
	}
	return base + suffix
}
