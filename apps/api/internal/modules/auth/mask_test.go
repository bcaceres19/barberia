package auth_test

import (
	"testing"

	"system-barbershop/internal/modules/auth"
)

func TestMaskPhone(t *testing.T) {
	cases := []struct {
		name  string
		phone string
		want  string
	}{
		{"E164 colombiano", "+573001234567", "+57 *** *** 67"},
		{"vacío", "", "***"},
		{"sin +", "573001234567", "***"},
		{"muy corto", "+57", "***"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := auth.MaskPhone(c.phone)
			if got != c.want {
				t.Fatalf("MaskPhone(%q) = %q, want %q", c.phone, got, c.want)
			}
			if c.phone != "" && c.phone != got {
				// El valor completo nunca debe aparecer intacto en la salida.
				if got == c.phone {
					t.Fatalf("MaskPhone(%q) devolvió el valor completo sin enmascarar", c.phone)
				}
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		name  string
		email string
		want  string
	}{
		{"correo normal", "bcaceres@correo.test", "b***@c***.test"},
		{"dominio sin punto", "a@localhost", "a***@l***"},
		{"sin arroba", "correo-invalido", "***"},
		{"arroba al final", "usuario@", "***"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := auth.MaskEmail(c.email)
			if got != c.want {
				t.Fatalf("MaskEmail(%q) = %q, want %q", c.email, got, c.want)
			}
		})
	}
}

// TestMaskEmail_NeverReturnsFullValue cubre CA-008-06 de forma general: para
// ningún correo con arroba interior el resultado debe ser igual al valor
// completo.
func TestMaskEmail_NeverReturnsFullValue(t *testing.T) {
	emails := []string{"duena.a@ejemplo.test", "x@y.co", "propietario@dominio.largo.test"}
	for _, e := range emails {
		if got := auth.MaskEmail(e); got == e {
			t.Fatalf("MaskEmail(%q) devolvió el valor completo sin enmascarar", e)
		}
	}
}
