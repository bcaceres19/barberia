// Pruebas del caso de uso EN AISLAMIENTO, sin PostgreSQL real: cubren
// normalización y validación de campo. El aislamiento de tenant, la
// confirmación de zona IANA contra pg_timezone_names y la atomicidad de la
// escritura se prueban contra PostgreSQL real en
// internal/modules/shops/postgres/repository_test.go
// (docs/03-desarrollo/estrategia-pruebas.md §2: nada de eso se prueba con
// un doble).
package shops_test

import (
	"context"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/apperr"
)

type updateCall struct {
	barbershopID string
	input        shops.UpdateInput
}

type fakeRepository struct {
	getBarbershop shops.Barbershop
	getFound      bool
	getErr        error
	getCalls      []string

	updateResult shops.UpdateResult
	updateErr    error
	updateCalls  []updateCall
}

func (f *fakeRepository) Get(_ context.Context, barbershopID string) (shops.Barbershop, bool, error) {
	f.getCalls = append(f.getCalls, barbershopID)
	return f.getBarbershop, f.getFound, f.getErr
}

func (f *fakeRepository) Update(_ context.Context, barbershopID string, input shops.UpdateInput) (shops.UpdateResult, error) {
	f.updateCalls = append(f.updateCalls, updateCall{barbershopID, input})
	return f.updateResult, f.updateErr
}

var _ shops.Repository = (*fakeRepository)(nil)

func assertValidation(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindValidation {
		t.Fatalf("expected apperr.KindValidation, got %v", err)
	}
}

const shopID = "11111111-1111-1111-1111-111111111111"

func validInput() shops.Input {
	return shops.Input{
		Name:         "Barbería Ejemplo",
		Timezone:     "America/Bogota",
		ContactEmail: "Contacto@Ejemplo.test",
		ContactPhone: "+573001234567",
	}
}

// --- Get --------------------------------------------------------------

func TestService_Get_Found_ReturnsBarbershop(t *testing.T) {
	want := shops.Barbershop{Name: "Barbería A", Timezone: "America/Bogota"}
	repo := &fakeRepository{getBarbershop: want, getFound: true}
	svc := shops.NewService(repo)

	got, err := svc.Get(context.Background(), shopID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
	if len(repo.getCalls) != 1 || repo.getCalls[0] != shopID {
		t.Fatalf("expected exactly one Get call with %q, got %v", shopID, repo.getCalls)
	}
}

func TestService_Get_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{getFound: false}
	svc := shops.NewService(repo)

	_, err := svc.Get(context.Background(), shopID)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

// --- Update: validación de campo (CA-020-03, CA-020-06) -----------------

func TestService_Update_EmptyName_ReturnsValidationWithoutCallingRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := shops.NewService(repo)

	input := validInput()
	input.Name = "   "
	_, err := svc.Update(context.Background(), shopID, input)
	assertValidation(t, err)
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be called for an invalid name")
	}
}

func TestService_Update_NameTooLong_ReturnsValidation(t *testing.T) {
	repo := &fakeRepository{}
	svc := shops.NewService(repo)

	longName := ""
	for i := 0; i < shops.NameMaxLength+1; i++ {
		longName += "a"
	}
	input := validInput()
	input.Name = longName
	_, err := svc.Update(context.Background(), shopID, input)
	assertValidation(t, err)
}

func TestService_Update_EmptyTimezone_ReturnsValidation(t *testing.T) {
	repo := &fakeRepository{}
	svc := shops.NewService(repo)

	input := validInput()
	input.Timezone = "   "
	_, err := svc.Update(context.Background(), shopID, input)
	assertValidation(t, err)
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be called for an empty timezone")
	}
}

func TestService_Update_InvalidContactEmailShape_ReturnsValidation(t *testing.T) {
	repo := &fakeRepository{}
	svc := shops.NewService(repo)

	input := validInput()
	input.ContactEmail = "no-es-un-correo"
	_, err := svc.Update(context.Background(), shopID, input)
	assertValidation(t, err)
	if len(repo.updateCalls) != 0 {
		t.Fatal("expected Repository.Update to never be called for a malformed contact email")
	}
}

func TestService_Update_InvalidContactPhoneShape_ReturnsValidation(t *testing.T) {
	for _, phone := range []string{"3001234567", "+57", "0123456789"} {
		repo := &fakeRepository{}
		svc := shops.NewService(repo)

		input := validInput()
		input.ContactPhone = phone
		_, err := svc.Update(context.Background(), shopID, input)
		assertValidation(t, err)
	}
}

// TestService_Update_EmptyContacts_NormalizeToNilBeforeCallingRepository
// cubre CA-020-06: un contacto vacío del formulario llega al repositorio
// como ausencia (nil), nunca como cadena vacía.
func TestService_Update_EmptyContacts_NormalizeToNilBeforeCallingRepository(t *testing.T) {
	repo := &fakeRepository{updateResult: shops.UpdateResult{TimezoneValid: true, Found: true}}
	svc := shops.NewService(repo)

	input := validInput()
	input.ContactEmail = "   "
	input.ContactPhone = ""
	if _, err := svc.Update(context.Background(), shopID, input); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if len(repo.updateCalls) != 1 {
		t.Fatalf("expected exactly one Update call, got %d", len(repo.updateCalls))
	}
	got := repo.updateCalls[0].input
	if got.ContactEmail != nil {
		t.Fatalf("expected ContactEmail to normalize to nil, got %q", *got.ContactEmail)
	}
	if got.ContactPhone != nil {
		t.Fatalf("expected ContactPhone to normalize to nil, got %q", *got.ContactPhone)
	}
}

// TestService_Update_ContactEmail_NormalizesToLowercaseAndTrimmed cubre
// CA-020-06.
func TestService_Update_ContactEmail_NormalizesToLowercaseAndTrimmed(t *testing.T) {
	repo := &fakeRepository{updateResult: shops.UpdateResult{TimezoneValid: true, Found: true}}
	svc := shops.NewService(repo)

	input := validInput()
	input.ContactEmail = "  Contacto@Ejemplo.TEST  "
	if _, err := svc.Update(context.Background(), shopID, input); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got := repo.updateCalls[0].input.ContactEmail
	if got == nil || *got != "contacto@ejemplo.test" {
		t.Fatalf("expected normalized lowercase email, got %v", got)
	}
}

// --- Update: desenlace de la zona (CA-020-03) ----------------------------

// TestService_Update_RepositoryReportsInvalidTimezone_ReturnsValidationNoWrite
// cubre CA-020-03: cero escritura del lado del repositorio se traduce en
// un 422 de campo, no en un 500 ni en un éxito silencioso.
func TestService_Update_RepositoryReportsInvalidTimezone_ReturnsValidationNoWrite(t *testing.T) {
	repo := &fakeRepository{updateResult: shops.UpdateResult{TimezoneValid: false}}
	svc := shops.NewService(repo)

	_, err := svc.Update(context.Background(), shopID, validInput())
	assertValidation(t, err)
	if len(repo.updateCalls) != 1 {
		t.Fatalf("expected the service to still delegate to the repository once, got %d calls", len(repo.updateCalls))
	}
}

func TestService_Update_RepositoryReportsNotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeRepository{updateResult: shops.UpdateResult{TimezoneValid: true, Found: false}}
	svc := shops.NewService(repo)

	_, err := svc.Update(context.Background(), shopID, validInput())
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindNotFound {
		t.Fatalf("expected apperr.KindNotFound, got %v", err)
	}
}

func TestService_Update_Success_ReturnsSavedRepresentation(t *testing.T) {
	saved := shops.Barbershop{Name: "Barbería Ejemplo", Timezone: "America/Bogota"}
	repo := &fakeRepository{updateResult: shops.UpdateResult{Barbershop: saved, TimezoneValid: true, Found: true}}
	svc := shops.NewService(repo)

	got, err := svc.Update(context.Background(), shopID, validInput())
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got != saved {
		t.Fatalf("expected %+v, got %+v", saved, got)
	}
}

func TestService_Update_NeverPassesBarbershopIDFromInput(t *testing.T) {
	// El propio tipo shops.Input no declara un campo BarbershopID: esta
	// prueba documenta en código que no hay canal para que un valor del
	// cuerpo llegue a sustituir el identificador ya autenticado (CA-020-05).
	repo := &fakeRepository{updateResult: shops.UpdateResult{TimezoneValid: true, Found: true}}
	svc := shops.NewService(repo)

	if _, err := svc.Update(context.Background(), shopID, validInput()); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updateCalls[0].barbershopID != shopID {
		t.Fatalf("expected the repository to receive the authenticated shopID %q, got %q", shopID, repo.updateCalls[0].barbershopID)
	}
}
