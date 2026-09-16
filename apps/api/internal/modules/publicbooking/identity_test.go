// Pruebas de dominio de identidad (HU-096): normalización/validación de
// CustomerIdentityInput y la política de reconciliación pura de DEC-085
// (ReconcilePublicCustomer). La lectura real contra PostgreSQL
// (FindCustomerMatches, dos tenants) se prueba en
// postgres/identity_repository_test.go.
package publicbooking_test

import (
	"testing"

	"system-barbershop/internal/modules/publicbooking"
	"system-barbershop/internal/platform/apperr"
)

func strPtr(s string) *string { return &s }

func validIdentityInput() publicbooking.CustomerIdentityInput {
	return publicbooking.CustomerIdentityInput{
		FullName: "Ana Ríos",
		Phone:    "+573001234567",
		Email:    "Ana@Example.com",
	}
}

func TestNormalizeAndValidateCustomerIdentity_ForSelf_DerivesAttendeeNameWithoutAsking(t *testing.T) {
	got, err := publicbooking.NormalizeAndValidateCustomerIdentity(validIdentityInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AttendeeName != "Ana Ríos" {
		t.Fatalf("expected attendee name derived from customer, got %q", got.AttendeeName)
	}
	if got.Email != "ana@example.com" {
		t.Fatalf("expected email normalized to lowercase, got %q", got.Email)
	}
}

func TestNormalizeAndValidateCustomerIdentity_ForSomeoneElse_RequiresNonEmptyAttendeeName(t *testing.T) {
	in := validIdentityInput()
	in.ForSomeoneElse = true

	_, err := publicbooking.NormalizeAndValidateCustomerIdentity(in)
	assertValidation(t, err)

	in.AttendeeName = strPtr("   ")
	_, err = publicbooking.NormalizeAndValidateCustomerIdentity(in)
	assertValidation(t, err)

	in.AttendeeName = strPtr("Mateo Ruiz")
	got, err := publicbooking.NormalizeAndValidateCustomerIdentity(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AttendeeName != "Mateo Ruiz" {
		t.Fatalf("expected attendee name %q, got %q", "Mateo Ruiz", got.AttendeeName)
	}
}

func TestNormalizeAndValidateCustomerIdentity_RequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*publicbooking.CustomerIdentityInput)
	}{
		{"empty full name", func(in *publicbooking.CustomerIdentityInput) { in.FullName = "   " }},
		{"full name too long", func(in *publicbooking.CustomerIdentityInput) { in.FullName = repeat("a", 121) }},
		{"empty phone", func(in *publicbooking.CustomerIdentityInput) { in.Phone = "" }},
		{"invalid phone", func(in *publicbooking.CustomerIdentityInput) { in.Phone = "3001234567" }},
		{"empty email", func(in *publicbooking.CustomerIdentityInput) { in.Email = "" }},
		{"invalid email", func(in *publicbooking.CustomerIdentityInput) { in.Email = "no-arroba" }},
		{"note too long", func(in *publicbooking.CustomerIdentityInput) { in.Note = strPtr(repeat("a", 501)) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validIdentityInput()
			tt.mutate(&in)
			_, err := publicbooking.NormalizeAndValidateCustomerIdentity(in)
			assertValidation(t, err)
		})
	}
}

func TestNormalizeAndValidateCustomerIdentity_OptionalNoteBlankBecomesNil(t *testing.T) {
	in := validIdentityInput()
	in.Note = strPtr("   ")
	got, err := publicbooking.NormalizeAndValidateCustomerIdentity(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Note != nil {
		t.Fatalf("expected blank note to normalize to nil, got %q", *got.Note)
	}
}

func assertValidation(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("expected apperr.KindValidation, got %v", err)
	}
}

func repeat(s string, n int) string {
	out := make([]byte, 0, n)
	for len(out) < n {
		out = append(out, s...)
	}
	return string(out[:n])
}

func TestReconcilePublicCustomer_NoMatch_CreatesNew(t *testing.T) {
	got := publicbooking.ReconcilePublicCustomer(nil, nil, "+573001234567", "ana@example.com")
	if got.ReuseCustomerID != nil {
		t.Fatalf("expected no reuse, got %v", *got.ReuseCustomerID)
	}
}

func TestReconcilePublicCustomer_PhoneMatchOnly_ReusesAndUpdatesEmail(t *testing.T) {
	phoneID := "customer-1"
	got := publicbooking.ReconcilePublicCustomer(&phoneID, nil, "+573001234567", "nuevo@example.com")
	if got.ReuseCustomerID == nil || *got.ReuseCustomerID != phoneID {
		t.Fatalf("expected reuse of %q, got %v", phoneID, got.ReuseCustomerID)
	}
	if got.UpdateEmail == nil || *got.UpdateEmail != "nuevo@example.com" {
		t.Fatalf("expected UpdateEmail to overwrite with new email, got %v", got.UpdateEmail)
	}
	if got.UpdatePhone != nil {
		t.Fatalf("expected UpdatePhone nil, got %v", *got.UpdatePhone)
	}
}

func TestReconcilePublicCustomer_EmailMatchOnly_ReusesAndUpdatesPhone(t *testing.T) {
	emailID := "customer-2"
	got := publicbooking.ReconcilePublicCustomer(nil, &emailID, "+573009999999", "ana@example.com")
	if got.ReuseCustomerID == nil || *got.ReuseCustomerID != emailID {
		t.Fatalf("expected reuse of %q, got %v", emailID, got.ReuseCustomerID)
	}
	if got.UpdatePhone == nil || *got.UpdatePhone != "+573009999999" {
		t.Fatalf("expected UpdatePhone to overwrite with new phone, got %v", got.UpdatePhone)
	}
}

func TestReconcilePublicCustomer_BothMatchSameRow_Reuses(t *testing.T) {
	id := "customer-3"
	got := publicbooking.ReconcilePublicCustomer(&id, &id, "+573001234567", "ana@example.com")
	if got.ReuseCustomerID == nil || *got.ReuseCustomerID != id {
		t.Fatalf("expected reuse of %q, got %v", id, got.ReuseCustomerID)
	}
}

func TestReconcilePublicCustomer_ConflictBetweenDifferentRows_CreatesNewWithoutMerging(t *testing.T) {
	phoneID, emailID := "customer-4", "customer-5"
	got := publicbooking.ReconcilePublicCustomer(&phoneID, &emailID, "+573001234567", "ana@example.com")
	if got.ReuseCustomerID != nil {
		t.Fatalf("expected no reuse on conflict, got %v", *got.ReuseCustomerID)
	}
	if got.UpdatePhone != nil || got.UpdateEmail != nil {
		t.Fatalf("expected no update on conflict, got phone=%v email=%v", got.UpdatePhone, got.UpdateEmail)
	}
}
