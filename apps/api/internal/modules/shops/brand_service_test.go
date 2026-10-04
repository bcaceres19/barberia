// Pruebas del caso de uso de marca y vocabulario EN AISLAMIENTO, sin
// PostgreSQL real: cubren normalización y validación de campo. El
// aislamiento de tenant y las restricciones CHECK se prueban contra
// PostgreSQL real en internal/modules/shops/postgres/brand_repository_test.go
// (docs/03-desarrollo/estrategia-pruebas.md §2).
package shops_test

import (
	"context"
	"strings"
	"testing"

	"system-barbershop/internal/modules/shops"
)

type brandUpdateCall struct {
	barbershopID string
	brand        shops.Brand
}

type fakeBrandRepository struct {
	getBrand shops.Brand
	getFound bool
	getErr   error

	updateResult shops.BrandUpdateResult
	updateErr    error
	updateCalls  []brandUpdateCall
}

func (f *fakeBrandRepository) Get(_ context.Context, _ string) (shops.Brand, bool, error) {
	return f.getBrand, f.getFound, f.getErr
}

func (f *fakeBrandRepository) Update(_ context.Context, barbershopID string, brand shops.Brand) (shops.BrandUpdateResult, error) {
	f.updateCalls = append(f.updateCalls, brandUpdateCall{barbershopID, brand})
	return f.updateResult, f.updateErr
}

var _ shops.BrandRepository = (*fakeBrandRepository)(nil)

func validBrand() shops.Brand {
	return shops.Brand{
		Accent:                 "emerald",
		BusinessTerm:           "Salón de Belleza",
		BusinessTermGender:     shops.GenderMasculine,
		ProfessionalTerm:       "  Estilista ",
		ProfessionalTermPlural: "estilistas",
		ProfessionalTermGender: shops.GenderFeminine,
	}
}

func TestBrandService_Get_Found_ReturnsBrand(t *testing.T) {
	want := shops.DefaultBrand
	svc := shops.NewBrandService(&fakeBrandRepository{getBrand: want, getFound: true})

	got, err := svc.Get(context.Background(), shopID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestBrandService_Get_NotFound_ReturnsNotFound(t *testing.T) {
	svc := shops.NewBrandService(&fakeBrandRepository{})

	_, err := svc.Get(context.Background(), shopID)
	if err == nil {
		t.Fatal("expected not found")
	}
}

func TestBrandService_Update_NormalizesTermsBeforeWriting(t *testing.T) {
	repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Found: true}}
	svc := shops.NewBrandService(repo)

	if _, err := svc.Update(context.Background(), shopID, validBrand()); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(repo.updateCalls) != 1 {
		t.Fatalf("expected exactly one repository write, got %d", len(repo.updateCalls))
	}
	got := repo.updateCalls[0]
	if got.barbershopID != shopID {
		t.Fatalf("expected the write to use the principal's barbershop, got %q", got.barbershopID)
	}
	// Recortado, en minúsculas y con los espacios repetidos colapsados:
	// exactamente lo que barbershop_*_term_ck exige en la base.
	if got.brand.BusinessTerm != "salón de belleza" || got.brand.ProfessionalTerm != "estilista" {
		t.Fatalf("terms were not normalized: %+v", got.brand)
	}
}

func TestBrandService_Update_CollapsesInnerWhitespace(t *testing.T) {
	repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Found: true}}
	svc := shops.NewBrandService(repo)

	in := validBrand()
	in.BusinessTerm = "salón   de \t belleza"
	if _, err := svc.Update(context.Background(), shopID, in); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updateCalls[0].brand.BusinessTerm != "salón de belleza" {
		t.Fatalf("expected collapsed spaces, got %q", repo.updateCalls[0].brand.BusinessTerm)
	}
}

func TestBrandService_Update_AcceptsEveryAllowedAccent(t *testing.T) {
	for _, accent := range shops.AllowedAccents {
		repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Found: true}}
		in := validBrand()
		in.Accent = accent
		if _, err := shops.NewBrandService(repo).Update(context.Background(), shopID, in); err != nil {
			t.Fatalf("accent %q should be accepted: %v", accent, err)
		}
	}
}

func TestBrandService_Update_InvalidField_ReturnsValidationAndNeverWrites(t *testing.T) {
	cases := map[string]func(*shops.Brand){
		"accent fuera de la lista":       func(b *shops.Brand) { b.Accent = "rojo" },
		"accent como color hexa":         func(b *shops.Brand) { b.Accent = "#b8955a" },
		"accent vacío":                   func(b *shops.Brand) { b.Accent = "" },
		"accent con otra capitalización": func(b *shops.Brand) { b.Accent = "Brass" },
		"negocio vacío":                  func(b *shops.Brand) { b.BusinessTerm = "   " },
		"negocio de una letra":           func(b *shops.Brand) { b.BusinessTerm = "s" },
		"negocio demasiado largo":        func(b *shops.Brand) { b.BusinessTerm = strings.Repeat("a", 31) },
		"negocio con dígitos":            func(b *shops.Brand) { b.BusinessTerm = "salón 24" },
		"negocio con etiqueta":           func(b *shops.Brand) { b.BusinessTerm = "<b>salón</b>" },
		"negocio que empieza con guion":  func(b *shops.Brand) { b.BusinessTerm = "-salón" },
		"género del negocio inválido":    func(b *shops.Brand) { b.BusinessTermGender = "neutral" },
		"profesional vacío":              func(b *shops.Brand) { b.ProfessionalTerm = "" },
		"plural con signos":              func(b *shops.Brand) { b.ProfessionalTermPlural = "estilistas!" },
		"género del profesional vacío":   func(b *shops.Brand) { b.ProfessionalTermGender = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Found: true}}
			in := validBrand()
			mutate(&in)

			_, err := shops.NewBrandService(repo).Update(context.Background(), shopID, in)
			assertValidation(t, err)
			if len(repo.updateCalls) != 0 {
				t.Fatalf("a rejected field must never reach the repository, got %d writes", len(repo.updateCalls))
			}
		})
	}
}

func TestBrandService_Update_AcceptsAccentsApostropheAndHyphen(t *testing.T) {
	for _, term := range []string{"peluquería canina", "o'brien", "centro de estética-spa", "ñandú", "salón"} {
		repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Found: true}}
		in := validBrand()
		in.BusinessTerm = term
		if _, err := shops.NewBrandService(repo).Update(context.Background(), shopID, in); err != nil {
			t.Fatalf("term %q should be accepted: %v", term, err)
		}
	}
}

func TestBrandService_Update_NotFound_ReturnsNotFound(t *testing.T) {
	repo := &fakeBrandRepository{updateResult: shops.BrandUpdateResult{Found: false}}

	_, err := shops.NewBrandService(repo).Update(context.Background(), shopID, validBrand())
	if err == nil {
		t.Fatal("expected not found when the barbershop is not visible")
	}
}

func TestBrandService_Update_CanceledContext_NeverWrites(t *testing.T) {
	repo := &fakeBrandRepository{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := shops.NewBrandService(repo).Update(ctx, shopID, validBrand()); err == nil {
		t.Fatal("expected an error for a canceled context")
	}
	if len(repo.updateCalls) != 0 {
		t.Fatal("a canceled context must never write")
	}
}

func TestDefaultBrand_IsValidAndMatchesTheMigrationDefaults(t *testing.T) {
	b := shops.DefaultBrand
	if !shops.IsAllowedAccent(b.Accent) || !shops.IsValidTerm(b.BusinessTerm) ||
		!shops.IsValidTerm(b.ProfessionalTerm) || !shops.IsValidTerm(b.ProfessionalTermPlural) ||
		!b.BusinessTermGender.IsValid() || !b.ProfessionalTermGender.IsValid() {
		t.Fatalf("DefaultBrand must satisfy its own validation: %+v", b)
	}
	// Los literales de la migración 20261003120000_add_barbershop_brand.sql.
	if b.Accent != "brass" || b.BusinessTerm != "barbería" || b.ProfessionalTerm != "barbero" ||
		b.ProfessionalTermPlural != "barberos" || b.BusinessTermGender != shops.GenderFeminine ||
		b.ProfessionalTermGender != shops.GenderMasculine {
		t.Fatalf("DefaultBrand drifted from the migration defaults: %+v", b)
	}
}
