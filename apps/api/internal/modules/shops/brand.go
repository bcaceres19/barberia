package shops

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Gender es el género gramatical de un término del vocabulario: solo sirve
// para que los artículos de la interfaz concuerden ("este estilista" /
// "esta estilista"). No describe a ninguna persona.
type Gender string

const (
	GenderMasculine Gender = "masculine"
	GenderFeminine  Gender = "feminine"
)

// IsValid informa si g es uno de los dos valores que barbershop_*_gender_ck
// admite en la base.
func (g Gender) IsValid() bool {
	return g == GenderMasculine || g == GenderFeminine
}

// PanelProfile es el perfil del panel de la barbería (issue #294, DEC-115):
// `shop` es el panel completo de una barbería con equipo y `solo` el de un
// barbero individual, sin gestión de equipo ni selector de barbero. Es
// presentación pura: no limita cuántos barberos existen ni cambia ninguna
// regla de agenda, porque un independiente sigue siendo una barbería con un
// solo barbero (DEC-019).
type PanelProfile string

const (
	PanelProfileShop PanelProfile = "shop"
	PanelProfileSolo PanelProfile = "solo"
)

// IsValid informa si p es uno de los dos valores que
// barbershop_panel_profile_ck admite en la base.
func (p PanelProfile) IsValid() bool {
	return p == PanelProfileShop || p == PanelProfileSolo
}

// Brand es la marca y el vocabulario que una barbería elige para su panel
// (issue #292, DEC-110). Es un recurso propio de /private/settings/brand,
// separado de Barbershop a propósito: el contrato de HU-020 declara
// exactamente cuatro campos (CA-020-07) y no se amplía.
//
// Accent es una clave de AllowedAccents, nunca un color: el cliente la
// traduce a una paleta con contraste AA comprobado en cada modo. Los
// términos se guardan recortados y en minúsculas; la interfaz capitaliza
// donde corresponde.
//
// PanelProfile viaja con la marca por ser otra preferencia de presentación de
// la barbería (DEC-115). En la entrada de BrandRepository.Update, vacío
// significa «conservar el perfil guardado»: así una solicitud que no lo
// declara nunca lo reinicia.
type Brand struct {
	Accent                 string
	BusinessTerm           string
	BusinessTermGender     Gender
	ProfessionalTerm       string
	ProfessionalTermPlural string
	ProfessionalTermGender Gender
	PanelProfile           PanelProfile
}

// Largos de un término, iguales a barbershop_business_term_ck,
// barbershop_professional_term_ck y barbershop_professional_term_plural_ck:
// el servicio no inventa un límite más laxo que la base ni al revés.
const (
	TermMinLength = 2
	TermMaxLength = 30
)

// AllowedAccents es la lista cerrada de barbershop_brand_accent_ck. Añadir
// una clave exige una migración y la misma clave en la paleta del cliente
// (apps/web/src/shared/model/brandPalette.ts), que prueba su contraste.
var AllowedAccents = [...]string{"brass", "emerald", "sapphire", "ruby", "amethyst", "copper"}

// DefaultBrand reproduce la interfaz anterior a DEC-110 y coincide con los
// DEFAULT de la migración 20261003120000_add_barbershop_brand.sql: una
// barbería que nunca toca esta configuración no cambia de aspecto.
var DefaultBrand = Brand{
	Accent:                 "brass",
	BusinessTerm:           "barbería",
	BusinessTermGender:     GenderFeminine,
	ProfessionalTerm:       "barbero",
	ProfessionalTermPlural: "barberos",
	ProfessionalTermGender: GenderMasculine,
	PanelProfile:           PanelProfileShop,
}

// IsAllowedAccent informa si key pertenece a la lista cerrada de acentos.
func IsAllowedAccent(key string) bool {
	for _, allowed := range AllowedAccents {
		if key == allowed {
			return true
		}
	}
	return false
}

// termShapePattern es la forma de un término: empieza con una letra y sigue
// con letras Unicode, espacios, guion o apóstrofo ("salón de belleza",
// "peluquería canina", "o'brien"). Sin dígitos, signos ni marcado: el valor
// se pinta en títulos y botones, nunca se interpreta como HTML.
var termShapePattern = regexp.MustCompile(`^\p{L}[\p{L} '\-]*$`)

// NormalizeTerm recorta, colapsa los espacios repetidos y pasa a minúsculas,
// igual que barbershop_*_term_ck exige (lower(btrim(...))). No valida: el
// largo y la forma se comprueban aparte con IsValidTerm.
func NormalizeTerm(raw string) string {
	return strings.ToLower(strings.Join(strings.Fields(raw), " "))
}

// IsValidTerm informa si term, ya normalizado, tiene un largo permitido y la
// forma de termShapePattern.
func IsValidTerm(term string) bool {
	length := utf8.RuneCountInString(term)
	if length < TermMinLength || length > TermMaxLength {
		return false
	}
	return termShapePattern.MatchString(term)
}
