package booking_test

import (
	"strings"
	"testing"
	"time"

	"system-barbershop/internal/modules/booking"
)

func TestLooksLikeAppointmentID(t *testing.T) {
	cases := []struct {
		name string
		id   string
		want bool
	}{
		{"uuid válido", "8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", true},
		{"vacío", "", false},
		{"sin guiones", "8f3ac2b1e4d546f6a7c8d9e0f1a2b3c4", false},
		{"con espacio", " 8f3ac2b1-e4d5-46f6-a7c8-d9e0f1a2b3c4", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := booking.LooksLikeAppointmentID(c.id); got != c.want {
				t.Fatalf("LooksLikeAppointmentID(%q) = %v, want %v", c.id, got, c.want)
			}
		})
	}
}

// TestEncodeVersionToken_SameInput_IsDeterministic verifica que dos
// lecturas de la misma cita sin ninguna escritura entre medias producen el
// mismo token de concurrencia.
func TestEncodeVersionToken_SameInput_IsDeterministic(t *testing.T) {
	updatedAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	a := booking.EncodeVersionToken(validAppointmentID, updatedAt)
	b := booking.EncodeVersionToken(validAppointmentID, updatedAt)
	if a != b {
		t.Fatalf("EncodeVersionToken no es determinista: %q != %q", a, b)
	}
	if a == "" {
		t.Fatalf("EncodeVersionToken devolvió un token vacío")
	}
}

// TestEncodeVersionToken_DifferentUpdatedAt_ChangesToken cubre "cualquier
// escritura que toque updated_at produce un token distinto".
func TestEncodeVersionToken_DifferentUpdatedAt_ChangesToken(t *testing.T) {
	a := booking.EncodeVersionToken(validAppointmentID, time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC))
	b := booking.EncodeVersionToken(validAppointmentID, time.Date(2026, 8, 30, 10, 0, 0, 1, time.UTC))
	if a == b {
		t.Fatalf("EncodeVersionToken produjo el mismo token para updated_at distintos")
	}
}

// TestEncodeVersionToken_DifferentAppointment_ChangesToken cubre que el
// token nunca sea intercambiable entre dos citas distintas aunque
// coincidiera el instante de actualización.
func TestEncodeVersionToken_DifferentAppointment_ChangesToken(t *testing.T) {
	updatedAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	a := booking.EncodeVersionToken(validAppointmentID, updatedAt)
	b := booking.EncodeVersionToken("11111111-1111-1111-1111-111111111111", updatedAt)
	if a == b {
		t.Fatalf("EncodeVersionToken produjo el mismo token para dos citas distintas")
	}
}

// TestEncodeVersionToken_NeverContainsRawTimestamp cubre "no conviertas su
// contenido en API pública": el token nunca debe traer, en texto plano, el
// año/mes/día del updated_at que codifica (irreversible, no una
// codificación reversible como base64 de RFC3339).
func TestEncodeVersionToken_NeverContainsRawTimestamp(t *testing.T) {
	updatedAt := time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC)
	token := booking.EncodeVersionToken(validAppointmentID, updatedAt)
	if strings.Contains(token, "2026") {
		t.Fatalf("el token expuso el año de updated_at en claro: %q", token)
	}
}

func TestHistoryCursor_EncodeDecode_RoundTrips(t *testing.T) {
	want := booking.HistoryCursor{OccurredAt: time.Date(2026, 8, 30, 10, 15, 0, 0, time.UTC), ID: "history-1"}
	token := booking.EncodeHistoryCursor(want)
	if token == "" {
		t.Fatalf("EncodeHistoryCursor devolvió un token vacío")
	}
	got, err := booking.DecodeHistoryCursor(token)
	if err != nil {
		t.Fatalf("err inesperado: %v", err)
	}
	if got.ID != want.ID || !got.OccurredAt.Equal(want.OccurredAt) {
		t.Fatalf("cursor decodificado = %+v, want %+v", got, want)
	}
}

func TestDecodeHistoryCursor_InvalidToken_RejectsAsInvalid(t *testing.T) {
	cases := []string{"", "no-es-base64url-!!", "AAAA"}
	for _, raw := range cases {
		if _, err := booking.DecodeHistoryCursor(raw); err == nil {
			t.Fatalf("DecodeHistoryCursor(%q) no devolvió error", raw)
		}
	}
}

// TestDecodeHistoryCursor_ManipulatedButValidBase64_RejectsAsInvalid cubre
// un valor que decodifica base64 correctamente pero nunca salió de
// EncodeHistoryCursor (id vacío): se rechaza igual, mismo criterio que
// staff.DecodeCursor.
func TestDecodeHistoryCursor_ManipulatedButValidBase64_RejectsAsInvalid(t *testing.T) {
	// Este valor decodifica a JSON válido pero incompleto: {"a":1}, sin las
	// claves occurredAt/id que historyCursorWire espera.
	if _, err := booking.DecodeHistoryCursor("eyJhIjoxfQ"); err == nil {
		t.Fatalf("DecodeHistoryCursor aceptó un valor que nunca salió de EncodeHistoryCursor")
	}
}
