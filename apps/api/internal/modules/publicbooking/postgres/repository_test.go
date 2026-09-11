// Package postgres_test (pruebas de integración, HU-090/HU-091) requiere
// PostgreSQL REAL con todas las migraciones aplicadas (incluida
// 20260911045044_add_barbershop_public_slug.sql) y
// database/testdata/{hu090_reserva_publica.sql,hu091_catalogo_publico.sql}
// cargados. Conéctate como barberia_app (docs/03-desarrollo/
// estrategia-pruebas.md §2 prohíbe mocks para RLS/resolución de tenant sin
// contexto).
//
//	export TEST_DATABASE_URL="postgres://barberia_app:app_test@localhost:5432/barberia_test?sslmode=disable"
//	go test -race ./internal/modules/publicbooking/postgres/...
package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"system-barbershop/internal/modules/publicbooking"
	publicbookingpostgres "system-barbershop/internal/modules/publicbooking/postgres"
	"system-barbershop/internal/platform/config"
	"system-barbershop/internal/platform/database"
)

const (
	testDatabaseURL = "postgres://barberia_app@localhost:5432/barberia_test?sslmode=disable"

	// Par dedicado de hu090_reserva_publica.sql: shopUno tiene contacto
	// completo y slug, shopDos tiene slug pero sin contacto (contactos
	// NULL, CA-090-04), shopNoPublicable tiene public_slug NULL (existe,
	// pero no es publicable).
	slugUno         = "barberia-hu090-uno"
	slugDos         = "barberia-hu090-dos"
	nameUno         = "Barbería de prueba HU-090 Uno"
	nameDos         = "Barbería de prueba HU-090 Dos"
	contactEmailUno = "contacto.hu090uno@ejemplo.test"
	contactPhoneUno = "+573000000001"

	// Par dedicado de hu091_catalogo_publico.sql (HU-091): slugCatalogoUno
	// tiene un servicio activo+asignado, uno activo sin asignar y uno
	// inactivo+asignado; slugCatalogoDos tiene un único servicio activo y
	// asignado, usado exclusivamente para CA-091-03 (aislamiento de tenant).
	slugCatalogoUno          = "barberia-hu091-uno"
	slugCatalogoDos          = "barberia-hu091-dos"
	servicioActivoAsignadoID = "00910101-0091-0091-0091-009101010101"
	servicioActivoAsignado   = "Corte activo asignado HU-091"
	servicioDosActivoID      = "00910201-0091-0091-0091-009102010201"
	servicioDosActivo        = "Corte activo asignado HU-091 Dos"

	// Par dedicado de hu092_seleccion_barbero.sql (HU-092): slugBarberoUno
	// tiene un servicio con exactamente un barbero (matriz "1"), uno con dos
	// (matriz "N"), uno sin ninguno (matriz "0") y uno inactivo con un
	// barbero asignado (CA-092-02); slugBarberoDos existe exclusivamente
	// para el aislamiento de tenant (CA-092-02/CA-092-03).
	slugBarberoUno         = "barberia-hu092-uno"
	slugBarberoDos         = "barberia-hu092-dos"
	servicioUnBarberoID    = "00920101-0092-0092-0092-009201010101"
	servicioVariosID       = "00920102-0092-0092-0092-009201020102"
	servicioSinBarberosID  = "00920103-0092-0092-0092-009201030103"
	servicioInactivoID     = "00920104-0092-0092-0092-009201040104"
	servicioDosBarberoID   = "00920201-0092-0092-0092-009202010201"
	barberoUnBarberoID     = "00920011-0092-0092-0092-009200110011"
	barberoUnBarberoNombre = "Barbero HU-092 Uno-1"
	barberoVariosSegundoID = "00920012-0092-0092-0092-009200120012"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = testDatabaseURL
	}
	cfg := config.Config{
		Environment:              "test",
		DatabaseMaxConns:         10,
		DatabaseMinConns:         2,
		DatabaseMaxConnLifetime:  time.Hour,
		DatabaseMaxConnIdleTime:  30 * time.Minute,
		DatabaseConnectTimeout:   5 * time.Second,
		DatabaseStatementTimeout: 10 * time.Second,
	}
	db, err := database.NewDB(config.DatabaseDSN(url), cfg)
	if err != nil {
		t.Fatalf("database.NewDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestResolveBySlug_ValidSlug_ReturnsPublicProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	profile, found, err := repo.ResolveBySlug(context.Background(), slugUno)
	if err != nil {
		t.Fatalf("ResolveBySlug: %v", err)
	}
	if !found {
		t.Fatal("expected slugUno to resolve")
	}
	if profile.Name != nameUno || profile.Timezone != "America/Bogota" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if profile.ContactEmail == nil || *profile.ContactEmail != contactEmailUno {
		t.Fatalf("expected contactEmail=%q, got %v", contactEmailUno, profile.ContactEmail)
	}
	if profile.ContactPhone == nil || *profile.ContactPhone != contactPhoneUno {
		t.Fatalf("expected contactPhone=%q, got %v", contactPhoneUno, profile.ContactPhone)
	}
}

// TestResolveBySlug_IsCaseInsensitive cubre DEC-082: el slug se compara sin
// distinguir mayúsculas (idx_barbershop_public_slug es sobre lower()).
func TestResolveBySlug_IsCaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	profile, found, err := repo.ResolveBySlug(context.Background(), "BARBERIA-HU090-UNO")
	if err != nil {
		t.Fatalf("ResolveBySlug: %v", err)
	}
	if !found || profile.Name != nameUno {
		t.Fatalf("expected case-insensitive match for shopUno, got found=%v profile=%+v", found, profile)
	}
}

// TestResolveBySlug_NoContact_ReturnsNilPointersNotEmptyStrings cubre
// CA-090-04: shopDos no tiene contacto configurado.
func TestResolveBySlug_NoContact_ReturnsNilPointersNotEmptyStrings(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	profile, found, err := repo.ResolveBySlug(context.Background(), slugDos)
	if err != nil {
		t.Fatalf("ResolveBySlug: %v", err)
	}
	if !found || profile.Name != nameDos {
		t.Fatalf("expected shopDos to resolve, got found=%v profile=%+v", found, profile)
	}
	if profile.ContactEmail != nil || profile.ContactPhone != nil {
		t.Fatalf("expected nil contact pointers, got %+v", profile)
	}
}

// TestResolveBySlug_UnknownSlug_ReturnsNotFound cubre CA-090-02: un slug
// bien formado pero que no existe.
func TestResolveBySlug_UnknownSlug_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	_, found, err := repo.ResolveBySlug(context.Background(), "barberia-que-no-existe")
	if err != nil {
		t.Fatalf("ResolveBySlug: %v", err)
	}
	if found {
		t.Fatal("expected an unknown slug not to resolve")
	}
}

// TestResolveBySlug_MalformedSlug_ReturnsNotFound cubre CA-090-02: una
// forma que ningún public_slug real puede tener (mayúsculas Y símbolos)
// produce la MISMA respuesta que un slug desconocido, nunca un error.
func TestResolveBySlug_MalformedSlug_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	for _, malformed := range []string{"Bad Slug!", "-empieza-con-guion", "termina-con-guion-", "ab"} {
		_, found, err := repo.ResolveBySlug(context.Background(), malformed)
		if err != nil {
			t.Fatalf("ResolveBySlug(%q): unexpected error %v", malformed, err)
		}
		if found {
			t.Fatalf("ResolveBySlug(%q): expected not found, got a match", malformed)
		}
	}
}

// TestResolveBySlug_NotPublicableBarbershop_ReturnsNotFound cubre
// CA-090-02/DEC-082: una barbería que EXISTE pero no tiene public_slug
// (no publicable) responde exactamente igual que un slug inexistente.
func TestResolveBySlug_NotPublicableBarbershop_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	// La barbería "No Publicable" del fixture no tiene ningún slug propio
	// que probar (public_slug es NULL): confirmamos que ningún slug la
	// alcanza, en particular no por casualidad con el nombre.
	_, found, err := repo.ResolveBySlug(context.Background(), "barberia-de-prueba-hu-090-no-publicable")
	if err != nil {
		t.Fatalf("ResolveBySlug: %v", err)
	}
	if found {
		t.Fatal("expected a barbershop without public_slug never to resolve")
	}
}

// TestResolveBySlug_TenantAIsolatedFromTenantB confirma que resolver un
// slug nunca filtra ni mezcla datos de la otra barbería del par (RN-TEN-01):
// cada resolución trae exactamente el perfil de SU propia barbería.
func TestResolveBySlug_TenantAIsolatedFromTenantB(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	uno, foundUno, err := repo.ResolveBySlug(context.Background(), slugUno)
	if err != nil || !foundUno {
		t.Fatalf("ResolveBySlug(slugUno): found=%v err=%v", foundUno, err)
	}
	dos, foundDos, err := repo.ResolveBySlug(context.Background(), slugDos)
	if err != nil || !foundDos {
		t.Fatalf("ResolveBySlug(slugDos): found=%v err=%v", foundDos, err)
	}
	if uno.Name == dos.Name {
		t.Fatalf("test fixture assumption broken: both slugs resolved to the same name %q", uno.Name)
	}
	if dos.ContactEmail != nil {
		t.Fatalf("RN-TEN-01: resolving slugDos leaked shopUno's contactEmail: %+v", dos)
	}
}

// TestListPublicServices_ActiveAndAssigned_ExcludesInactiveAndUnassigned
// cubre CA-091-01: de los cuatro servicios de la barbería HU-091 Uno (dos
// activos+asignados, uno activo sin asignar, uno inactivo+asignado), solo
// los dos activos+asignados aparecen.
func TestListPublicServices_ActiveAndAssigned_ExcludesInactiveAndUnassigned(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	result, found, err := repo.ListPublicServices(context.Background(), slugCatalogoUno, nil, 20)
	if err != nil {
		t.Fatalf("ListPublicServices: %v", err)
	}
	if !found {
		t.Fatal("expected slugCatalogoUno to resolve")
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected exactly 2 public services (active+assigned only), got %d: %+v", len(result.Items), result.Items)
	}
	svc := result.Items[0]
	if svc.ID != servicioActivoAsignadoID || svc.Name != servicioActivoAsignado {
		t.Fatalf("unexpected first service: %+v", svc)
	}
	if svc.DurationMinutes != 30 || svc.PriceCents != 3500000 || svc.Currency != "COP" {
		t.Fatalf("unexpected service fields: %+v", svc)
	}
	if result.NextCursor != "" {
		t.Fatalf("expected no next page for a full 20-item page with only 2 rows, got cursor %q", result.NextCursor)
	}
}

// TestListPublicServices_UnknownOrNotPublicSlug_ReturnsNotFound cubre la
// misma respuesta uniforme que ResolveBySlug (CA-090-02, reutilizada aquí).
func TestListPublicServices_UnknownOrNotPublicSlug_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	_, found, err := repo.ListPublicServices(context.Background(), "barberia-que-no-existe", nil, 20)
	if err != nil {
		t.Fatalf("ListPublicServices: %v", err)
	}
	if found {
		t.Fatal("expected an unknown slug not to resolve")
	}
}

// TestListPublicServices_TenantAIsolatedFromTenantB cubre CA-091-03: el
// catálogo de una barbería nunca incluye el servicio de la otra, verificado
// con dos tenants reales.
func TestListPublicServices_TenantAIsolatedFromTenantB(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	uno, foundUno, err := repo.ListPublicServices(context.Background(), slugCatalogoUno, nil, 20)
	if err != nil || !foundUno {
		t.Fatalf("ListPublicServices(slugCatalogoUno): found=%v err=%v", foundUno, err)
	}
	dos, foundDos, err := repo.ListPublicServices(context.Background(), slugCatalogoDos, nil, 20)
	if err != nil || !foundDos {
		t.Fatalf("ListPublicServices(slugCatalogoDos): found=%v err=%v", foundDos, err)
	}

	for _, svc := range uno.Items {
		if svc.ID == servicioDosActivoID {
			t.Fatalf("RN-TEN-01: slugCatalogoUno's catalog leaked shopDos's service: %+v", svc)
		}
	}
	if len(dos.Items) != 1 || dos.Items[0].ID != servicioDosActivoID || dos.Items[0].Name != servicioDosActivo {
		t.Fatalf("unexpected shopDos catalog: %+v", dos.Items)
	}
}

// TestListPublicServices_LimitOne_PaginatesAcrossTwoPages cubre la
// paginación por cursor del catálogo público: con límite 1, la barbería
// HU-091 Uno (exactamente dos servicios públicos) entrega su primera
// página con nextCursor, y la segunda página (pedida con ese cursor) trae
// el segundo servicio sin repetir el primero y sin nextCursor (ya no queda
// una tercera página).
func TestListPublicServices_LimitOne_PaginatesAcrossTwoPages(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	first, found, err := repo.ListPublicServices(context.Background(), slugCatalogoUno, nil, 1)
	if err != nil || !found || len(first.Items) != 1 {
		t.Fatalf("ListPublicServices (first page): found=%v items=%d err=%v", found, len(first.Items), err)
	}
	if first.NextCursor == "" {
		t.Fatal("expected a nextCursor: a second public service exists")
	}
	if first.Items[0].ID != servicioActivoAsignadoID {
		t.Fatalf("expected the first page to be the earliest-created service, got %+v", first.Items[0])
	}

	firstCursor, err := publicbooking.DecodeServiceCursor(first.NextCursor)
	if err != nil {
		t.Fatalf("DecodeServiceCursor(first.NextCursor): %v", err)
	}
	second, found, err := repo.ListPublicServices(context.Background(), slugCatalogoUno, &firstCursor, 1)
	if err != nil || !found || len(second.Items) != 1 {
		t.Fatalf("ListPublicServices (second page): found=%v items=%d err=%v", found, len(second.Items), err)
	}
	if second.Items[0].ID == first.Items[0].ID {
		t.Fatalf("expected the second page not to repeat the first item, got %+v twice", second.Items[0])
	}
	if second.NextCursor != "" {
		t.Fatalf("expected no nextCursor on the last page, got %q", second.NextCursor)
	}
}

// TestListPublicBarbers_OneBarber_ReturnsSingleItem cubre CA-092-01 (caso
// "1" de la matriz 0/1/N): exactamente un barbero elegible.
func TestListPublicBarbers_OneBarber_ReturnsSingleItem(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	result, found, err := repo.ListPublicBarbers(context.Background(), slugBarberoUno, servicioUnBarberoID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if !found {
		t.Fatal("expected slugBarberoUno to resolve")
	}
	if len(result.Items) != 1 || result.Items[0].ID != barberoUnBarberoID || result.Items[0].FullName != barberoUnBarberoNombre {
		t.Fatalf("expected exactly one barber (%s), got %+v", barberoUnBarberoID, result.Items)
	}
}

// TestListPublicBarbers_ManyBarbers_ReturnsAllInStableOrder cubre CA-092-01
// (caso "N"): varios barberos elegibles, en orden estable
// (created_at de la asignación, id del barbero).
func TestListPublicBarbers_ManyBarbers_ReturnsAllInStableOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	result, found, err := repo.ListPublicBarbers(context.Background(), slugBarberoUno, servicioVariosID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if !found {
		t.Fatal("expected slugBarberoUno to resolve")
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected exactly 2 barbers, got %+v", result.Items)
	}
	if result.Items[0].ID != barberoUnBarberoID || result.Items[1].ID != barberoVariosSegundoID {
		t.Fatalf("expected a stable order (%s, %s), got %+v", barberoUnBarberoID, barberoVariosSegundoID, result.Items)
	}
}

// TestListPublicBarbers_ZeroBarbers_ReturnsEmptySuccess cubre CA-092-01
// (caso "0"): un servicio activo sin ninguna asignación es un resultado
// exitoso vacío, nunca un error.
func TestListPublicBarbers_ZeroBarbers_ReturnsEmptySuccess(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	result, found, err := repo.ListPublicBarbers(context.Background(), slugBarberoUno, servicioSinBarberosID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if !found {
		t.Fatal("expected slugBarberoUno to resolve")
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected zero barbers, got %+v", result.Items)
	}
}

// TestListPublicBarbers_InactiveService_ReturnsEmpty cubre CA-092-02: un
// servicio inactivo nunca expone sus barberos asignados, aunque la
// asignación exista.
func TestListPublicBarbers_InactiveService_ReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	result, found, err := repo.ListPublicBarbers(context.Background(), slugBarberoUno, servicioInactivoID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if !found {
		t.Fatal("expected slugBarberoUno to resolve")
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected zero barbers for an inactive service, got %+v", result.Items)
	}
}

// TestListPublicBarbers_MalformedServiceID_ReturnsEmptySuccess cubre
// CA-092-03: un serviceID sin forma de UUID nunca produce un error de tipo
// de PostgreSQL, y la barbería sigue resolviendo.
func TestListPublicBarbers_MalformedServiceID_ReturnsEmptySuccess(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	result, found, err := repo.ListPublicBarbers(context.Background(), slugBarberoUno, "no-es-un-uuid")
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if !found {
		t.Fatal("expected slugBarberoUno to resolve regardless of a malformed serviceID")
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected zero barbers for a malformed serviceID, got %+v", result.Items)
	}
}

// TestListPublicBarbers_UnknownSlug_ReturnsNotFound cubre la misma
// respuesta uniforme que ResolveBySlug (CA-090-02, reutilizada aquí).
func TestListPublicBarbers_UnknownSlug_ReturnsNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	_, found, err := repo.ListPublicBarbers(context.Background(), "barberia-que-no-existe", servicioUnBarberoID)
	if err != nil {
		t.Fatalf("ListPublicBarbers: %v", err)
	}
	if found {
		t.Fatal("expected an unknown slug not to resolve")
	}
}

// TestListPublicBarbers_TenantAIsolatedFromTenantB cubre CA-092-02/
// CA-092-03: un serviceId de la barbería Dos nunca devuelve barberos al
// resolverse a través del slug de la barbería Uno, y viceversa, verificado
// con dos tenants reales y asignaciones cruzadas.
func TestListPublicBarbers_TenantAIsolatedFromTenantB(t *testing.T) {
	db := setupTestDB(t)
	repo := publicbookingpostgres.New(db)

	crossUno, foundUno, err := repo.ListPublicBarbers(context.Background(), slugBarberoUno, servicioDosBarberoID)
	if err != nil || !foundUno {
		t.Fatalf("ListPublicBarbers(slugBarberoUno, servicioDosBarberoID): found=%v err=%v", foundUno, err)
	}
	if len(crossUno.Items) != 0 {
		t.Fatalf("RN-TEN-01: slugBarberoUno exposed barbershop Dos's service barbers: %+v", crossUno.Items)
	}

	crossDos, foundDos, err := repo.ListPublicBarbers(context.Background(), slugBarberoDos, servicioUnBarberoID)
	if err != nil || !foundDos {
		t.Fatalf("ListPublicBarbers(slugBarberoDos, servicioUnBarberoID): found=%v err=%v", foundDos, err)
	}
	if len(crossDos.Items) != 0 {
		t.Fatalf("RN-TEN-01: slugBarberoDos exposed barbershop Uno's service barbers: %+v", crossDos.Items)
	}

	// Confirma además que cada barbería SÍ resuelve su propio servicio
	// (el aislamiento no es un falso negativo generalizado).
	own, foundOwn, err := repo.ListPublicBarbers(context.Background(), slugBarberoDos, servicioDosBarberoID)
	if err != nil || !foundOwn || len(own.Items) != 1 {
		t.Fatalf("ListPublicBarbers(slugBarberoDos, servicioDosBarberoID): found=%v items=%d err=%v", foundOwn, len(own.Items), err)
	}
}
