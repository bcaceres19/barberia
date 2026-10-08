// Pruebas EN AISLAMIENTO de shops.SlugBase/SlugWithCode/NewSlugCode
// (HU-090, DEC-082, DEC-117): funciones puras, sin PostgreSQL. La generación dentro de una
// transacción real, el reintento ante unique_violation y la unicidad
// GLOBAL entre dos tenants se prueban contra PostgreSQL real en
// internal/modules/shops/postgres/repository_test.go
// (docs/03-desarrollo/estrategia-pruebas.md §2).
package shops_test

import (
	"regexp"
	"strings"
	"testing"

	"system-barbershop/internal/modules/shops"
)

func TestSlugBase_SimpleName_LowercasesAndHyphenates(t *testing.T) {
	got := shops.SlugBase(shops.NormalizeName("Barbería Ejemplo"))
	if got != "barberia-ejemplo" {
		t.Fatalf("expected %q, got %q", "barberia-ejemplo", got)
	}
}

func TestSlugBase_AccentedCharacters_FoldToASCII(t *testing.T) {
	got := shops.SlugBase(shops.NormalizeName("Peluquería José Núñez Ñoño"))
	if got != "peluqueria-jose-nunez-nono" {
		t.Fatalf("expected %q, got %q", "peluqueria-jose-nunez-nono", got)
	}
}

func TestSlugBase_NonAlphanumericRuns_CollapseToSingleHyphen(t *testing.T) {
	got := shops.SlugBase(shops.NormalizeName("  Barber@Shop!!   #1  "))
	if strings.Contains(got, "--") {
		t.Fatalf("expected no consecutive hyphens, got %q", got)
	}
	if strings.HasPrefix(got, "-") || strings.HasSuffix(got, "-") {
		t.Fatalf("expected no leading/trailing hyphen, got %q", got)
	}
}

func TestSlugBase_TooLong_TruncatesToMaxLength(t *testing.T) {
	longName := strings.Repeat("a", 100)
	got := shops.SlugBase(longName)
	if len(got) > shops.SlugMaxLength {
		t.Fatalf("expected at most %d chars, got %d (%q)", shops.SlugMaxLength, len(got), got)
	}
}

func TestSlugBase_TruncationLandsOnHyphen_TrimsIt(t *testing.T) {
	// 39 letras + un espacio en la posición 40: al truncar a 40 el corte
	// cae justo en el guion que reemplazó ese espacio.
	name := strings.Repeat("a", 39) + " " + strings.Repeat("b", 10)
	got := shops.SlugBase(name)
	if strings.HasSuffix(got, "-") {
		t.Fatalf("expected no trailing hyphen after truncation, got %q", got)
	}
}

func TestSlugBase_NoAlphanumericCharacters_FallsBackToBarberia(t *testing.T) {
	got := shops.SlugBase(shops.NormalizeName("!!! ??? ..."))
	if got != "barberia" {
		t.Fatalf("expected fallback %q, got %q", "barberia", got)
	}
}

func TestSlugBase_ShorterThanMinLength_FallsBackToBarberia(t *testing.T) {
	got := shops.SlugBase("A")
	if got != "barberia" {
		t.Fatalf("expected fallback %q, got %q", "barberia", got)
	}
}

func TestSlugBase_ResultMatchesDatabaseFormat(t *testing.T) {
	cases := []string{"Barbería Ejemplo", "A B", "!!!", strings.Repeat("x", 200), "Café & Corte 123"}
	for _, name := range cases {
		got := shops.SlugBase(shops.NormalizeName(name))
		if len(got) < shops.SlugMinLength || len(got) > shops.SlugMaxLength {
			t.Fatalf("SlugBase(%q) = %q: largo fuera de [%d,%d]", name, got, shops.SlugMinLength, shops.SlugMaxLength)
		}
		if strings.HasPrefix(got, "-") || strings.HasSuffix(got, "-") {
			t.Fatalf("SlugBase(%q) = %q: no debe empezar ni terminar en guion", name, got)
		}
		for _, r := range got {
			if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' {
				t.Fatalf("SlugBase(%q) = %q: carácter fuera de [a-z0-9-]: %q", name, got, r)
			}
		}
	}
}

func TestSlugWithCode_JoinsBaseAndCodeWithAHyphen(t *testing.T) {
	got := shops.SlugWithCode("barberia-ejemplo", "k7x2m9")
	if got != "barberia-ejemplo-k7x2m9" {
		t.Fatalf("expected %q, got %q", "barberia-ejemplo-k7x2m9", got)
	}
}

func TestSlugWithCode_TruncatesBaseToStayWithinMaxLength(t *testing.T) {
	base := strings.Repeat("a", shops.SlugMaxLength)
	got := shops.SlugWithCode(base, "k7x2m9")
	if len(got) != shops.SlugMaxLength {
		t.Fatalf("expected exactly %d chars, got %d (%q)", shops.SlugMaxLength, len(got), got)
	}
	if !strings.HasSuffix(got, "-k7x2m9") {
		t.Fatalf("expected code preserved, got %q", got)
	}
}

func TestSlugWithCode_TruncationLandsOnHyphen_DoesNotLeaveDoubleHyphen(t *testing.T) {
	// El corte cae justo en un guion de la base: sin recorte quedaría "--".
	base := strings.Repeat("a", 32) + "-" + strings.Repeat("b", 7)
	got := shops.SlugWithCode(base, "k7x2m9")
	if strings.Contains(got, "--") {
		t.Fatalf("expected no consecutive hyphens, got %q", got)
	}
}

var slugFormat = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{1,38}[a-z0-9])$`)

func TestSlugWithCode_ResultMatchesDatabaseFormat(t *testing.T) {
	for _, name := range []string{"Barbería Ejemplo", "A B", "!!!", strings.Repeat("x", 200), "Café & Corte 123"} {
		got := shops.SlugWithCode(shops.SlugBase(shops.NormalizeName(name)), "k7x2m9")
		if !slugFormat.MatchString(got) {
			t.Fatalf("SlugWithCode(SlugBase(%q)) = %q: no cumple barbershop_public_slug_ck", name, got)
		}
	}
}

func TestNewSlugCode_HasFixedLengthAndSafeAlphabet(t *testing.T) {
	for range 200 {
		code, err := shops.NewSlugCode()
		if err != nil {
			t.Fatalf("NewSlugCode: %v", err)
		}
		if len(code) != shops.SlugCodeLength {
			t.Fatalf("expected %d chars, got %q", shops.SlugCodeLength, code)
		}
		if strings.ContainsAny(code, "ilo01-") {
			t.Fatalf("code %q contains an ambiguous character or a hyphen", code)
		}
		if !regexp.MustCompile(`^[a-z0-9]+$`).MatchString(code) {
			t.Fatalf("code %q outside [a-z0-9]", code)
		}
	}
}

func TestNewSlugCode_IsNotSequential(t *testing.T) {
	seen := make(map[string]struct{})
	for range 500 {
		code, err := shops.NewSlugCode()
		if err != nil {
			t.Fatalf("NewSlugCode: %v", err)
		}
		seen[code] = struct{}{}
	}
	// 500 sorteos sobre ~887 millones: una repetición indicaría un
	// generador predecible o con muy poca entropía.
	if len(seen) != 500 {
		t.Fatalf("expected 500 distinct codes, got %d", len(seen))
	}
}
