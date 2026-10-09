package postgres_test

import (
	"context"
	"sync"
	"testing"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/database"
)

// Pruebas de integración del vínculo barbero–usuario (DEC-100) contra
// PostgreSQL real y DOS barberías (shopC/shopD). Los usuarios viven en
// testdata/hu021_barberos.sql; cada prueba libera sus vínculos al terminar
// para ser repetible sobre una base persistente.

const (
	userC1 = "cccccc01-cccc-4ccc-8ccc-cccccccccc01"
	userC2 = "cccccc02-cccc-4ccc-8ccc-cccccccccc02"
	userD1 = "dddddd01-dddd-4ddd-8ddd-dddddddddd01"
)

func releaseOnCleanup(t *testing.T, repo staff.Repository, shop database.BarbershopID, users ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, user := range users {
			if _, err := repo.Unlink(context.Background(), string(shop), user); err != nil {
				t.Errorf("cleanup unlink %s: %v", user, err)
			}
		}
	})
}

func TestLink_LinksReadsAndIsIdempotent(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1)
	barber := createBarber(t, repo, shopC, "Vínculo propio "+uniqueSuffix(t))

	if _, found, err := repo.GetLinked(ctx, string(shopC), userC1); err != nil || found {
		t.Fatalf("antes de vincular no hay barbero: found=%v err=%v", found, err)
	}

	first, err := repo.Link(ctx, string(shopC), userC1, barber.ID)
	if err != nil || !first.Found || first.Taken || first.Barber.ID != barber.ID {
		t.Fatalf("Link: %+v err=%v", first, err)
	}
	got, found, err := repo.GetLinked(ctx, string(shopC), userC1)
	if err != nil || !found || got.ID != barber.ID || got.FullName != barber.FullName {
		t.Fatalf("GetLinked: %+v found=%v err=%v", got, found, err)
	}

	again, err := repo.Link(ctx, string(shopC), userC1, barber.ID)
	if err != nil || !again.Found || again.Taken || again.ReleasedBarberID != "" {
		t.Fatalf("repetir la selección es idempotente y no libera nada: %+v err=%v", again, err)
	}
}

func TestLink_SwitchingBarberReleasesThePreviousOne(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1)
	first := createBarber(t, repo, shopC, "Primero "+uniqueSuffix(t))
	second := createBarber(t, repo, shopC, "Segundo "+uniqueSuffix(t))

	if _, err := repo.Link(ctx, string(shopC), userC1, first.ID); err != nil {
		t.Fatal(err)
	}
	moved, err := repo.Link(ctx, string(shopC), userC1, second.ID)
	if err != nil || moved.Taken || moved.ReleasedBarberID != first.ID {
		t.Fatalf("cambiar de barbero libera el anterior: %+v err=%v", moved, err)
	}
	got, _, err := repo.GetLinked(ctx, string(shopC), userC1)
	if err != nil || got.ID != second.ID {
		t.Fatalf("el usuario debe quedar con el segundo barbero: %+v err=%v", got, err)
	}

	// El primero quedó libre: otro usuario puede tomarlo.
	releaseOnCleanup(t, repo, shopC, userC2)
	taken, err := repo.Link(ctx, string(shopC), userC2, first.ID)
	if err != nil || taken.Taken || !taken.Found {
		t.Fatalf("el barbero liberado debe poder tomarlo otro usuario: %+v err=%v", taken, err)
	}
}

func TestLink_BarberOfAnotherUser_IsTakenAndKeepsThePreviousLink(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1, userC2)
	mine := createBarber(t, repo, shopC, "Mío "+uniqueSuffix(t))
	theirs := createBarber(t, repo, shopC, "Ajeno "+uniqueSuffix(t))

	if _, err := repo.Link(ctx, string(shopC), userC1, mine.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Link(ctx, string(shopC), userC2, theirs.ID); err != nil {
		t.Fatal(err)
	}

	result, err := repo.Link(ctx, string(shopC), userC1, theirs.ID)
	if err != nil || !result.Found || !result.Taken || result.ReleasedBarberID != "" {
		t.Fatalf("tomar el barbero de otro usuario es conflicto y no libera nada: %+v err=%v", result, err)
	}
	got, found, err := repo.GetLinked(ctx, string(shopC), userC1)
	if err != nil || !found || got.ID != mine.ID {
		t.Fatalf("el vínculo anterior del solicitante debe conservarse: %+v found=%v err=%v", got, found, err)
	}
	owner, _, err := repo.GetLinked(ctx, string(shopC), userC2)
	if err != nil || owner.ID != theirs.ID {
		t.Fatalf("el dueño del barbero no debe cambiar: %+v err=%v", owner, err)
	}
}

func TestLink_AcrossTenants_IsNotFoundAndDoesNotLeak(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1)
	releaseOnCleanup(t, repo, shopD, userD1)
	inC := createBarber(t, repo, shopC, "En C "+uniqueSuffix(t))
	inD := createBarber(t, repo, shopD, "En D "+uniqueSuffix(t))

	// Un usuario de C pidiendo un barbero de D (y al revés) no lo encuentra.
	if result, err := repo.Link(ctx, string(shopC), userC1, inD.ID); err != nil || result.Found {
		t.Fatalf("un barbero de otra barbería no existe para este tenant: %+v err=%v", result, err)
	}
	if result, err := repo.Link(ctx, string(shopD), userD1, inC.ID); err != nil || result.Found {
		t.Fatalf("un barbero de otra barbería no existe para este tenant: %+v err=%v", result, err)
	}

	// El vínculo de C es invisible y no liberable desde D.
	if _, err := repo.Link(ctx, string(shopC), userC1, inC.ID); err != nil {
		t.Fatal(err)
	}
	if _, found, err := repo.GetLinked(ctx, string(shopD), userC1); err != nil || found {
		t.Fatalf("D no debe ver el vínculo de C: found=%v err=%v", found, err)
	}
	if released, err := repo.Unlink(ctx, string(shopD), userC1); err != nil || released != "" {
		t.Fatalf("D no debe liberar el vínculo de C: released=%q err=%v", released, err)
	}
	if got, found, err := repo.GetLinked(ctx, string(shopC), userC1); err != nil || !found || got.ID != inC.ID {
		t.Fatalf("el vínculo de C debe seguir intacto: %+v found=%v err=%v", got, found, err)
	}
}

func TestUnlink_ReleasesIsIdempotentAndKeepsTheBarber(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1)
	barber := createBarber(t, repo, shopC, "Se libera "+uniqueSuffix(t))

	if _, err := repo.Link(ctx, string(shopC), userC1, barber.ID); err != nil {
		t.Fatal(err)
	}
	released, err := repo.Unlink(ctx, string(shopC), userC1)
	if err != nil || released != barber.ID {
		t.Fatalf("Unlink debe devolver el barbero liberado: %q err=%v", released, err)
	}
	if _, found, _ := repo.GetLinked(ctx, string(shopC), userC1); found {
		t.Fatal("tras quitarlo el usuario no debe tener barbero")
	}
	if stillThere, found, err := repo.Get(ctx, string(shopC), barber.ID); err != nil || !found || stillThere.FullName != barber.FullName {
		t.Fatalf("quitar el vínculo no borra ni cambia al barbero: %+v found=%v err=%v", stillThere, found, err)
	}
	if released, err := repo.Unlink(ctx, string(shopC), userC1); err != nil || released != "" {
		t.Fatalf("quitar sin vínculo es un éxito: released=%q err=%v", released, err)
	}
}

// Dos usuarios piden a la vez el mismo barbero libre: exactamente uno gana.
func TestLink_ConcurrentRaceForOneBarber_HasExactlyOneWinner(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1, userC2)
	barber := createBarber(t, repo, shopC, "Carrera "+uniqueSuffix(t))

	results := make([]staff.LinkResult, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i, user := range []string{userC1, userC2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = repo.Link(ctx, string(shopC), user, barber.ID)
		}()
	}
	wg.Wait()

	winners := 0
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("link %d: %v", i, errs[i])
		}
		if results[i].Found && !results[i].Taken {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("exactamente un usuario debe ganar el barbero, ganaron %d: %+v", winners, results)
	}
}

// Una misma persona lanzando dos vínculos simultáneos a barberos distintos
// termina con exactamente uno, sin violar la unicidad parcial.
func TestLink_ConcurrentSameUserTwoBarbers_EndsWithExactlyOne(t *testing.T) {
	repo := newRepository(setupTestDB(t))
	ctx := context.Background()
	releaseOnCleanup(t, repo, shopC, userC1)
	a := createBarber(t, repo, shopC, "Doble A "+uniqueSuffix(t))
	b := createBarber(t, repo, shopC, "Doble B "+uniqueSuffix(t))

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, id := range []string{a.ID, b.ID} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = repo.Link(ctx, string(shopC), userC1, id)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("link %d no debe fallar por la carrera: %v", i, err)
		}
	}
	got, found, err := repo.GetLinked(ctx, string(shopC), userC1)
	if err != nil || !found || (got.ID != a.ID && got.ID != b.ID) {
		t.Fatalf("el usuario debe terminar con exactamente un barbero: %+v found=%v err=%v", got, found, err)
	}
}
