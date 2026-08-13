package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
)

// --- doble de prueba ----------------------------------------------------

type resolveSessionCall struct{ tokenHash string }

type renewCall struct {
	shop, tokenHash   string
	now, newExpiresAt time.Time
}

type revokeCall struct {
	shop, sessionID string
	now             time.Time
}

// fakeSessionRepository es un doble en memoria de auth.SessionRepository:
// estas pruebas verifican el CASO DE USO en aislamiento, sin PostgreSQL
// real (esa cobertura vive en
// internal/modules/auth/postgres/session_repository_test.go).
type fakeSessionRepository struct {
	resolveShop  string
	resolveFound bool
	resolveErr   error
	resolveCalls []resolveSessionCall

	renewPrincipal auth.Principal
	renewValid     bool
	renewErr       error
	renewCalls     []renewCall

	revokeErr   error
	revokeCalls []revokeCall
}

func (f *fakeSessionRepository) ResolveSessionTenant(_ context.Context, tokenHash string) (string, bool, error) {
	f.resolveCalls = append(f.resolveCalls, resolveSessionCall{tokenHash})
	return f.resolveShop, f.resolveFound, f.resolveErr
}

func (f *fakeSessionRepository) ValidateAndRenewSession(_ context.Context, barbershopID, tokenHash string, now, newExpiresAt time.Time) (auth.Principal, bool, error) {
	f.renewCalls = append(f.renewCalls, renewCall{barbershopID, tokenHash, now, newExpiresAt})
	return f.renewPrincipal, f.renewValid, f.renewErr
}

func (f *fakeSessionRepository) RevokeSession(_ context.Context, barbershopID, sessionID string, now time.Time) error {
	f.revokeCalls = append(f.revokeCalls, revokeCall{barbershopID, sessionID, now})
	return f.revokeErr
}

var _ auth.SessionRepository = (*fakeSessionRepository)(nil)

const rawTestToken = "token-de-prueba-para-hash-consistente"

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected apperr.KindUnauthorized, got %v", err)
	}
}

// --- Validate -------------------------------------------------------------

// TestSessionService_Validate_UnknownToken_ReturnsUniformUnauthorized cubre
// un token que ResolveSessionTenant no resuelve (desconocido, vencido o
// revocado a nivel de la función SQL): CA-006-02.
func TestSessionService_Validate_UnknownToken_ReturnsUniformUnauthorized(t *testing.T) {
	repo := &fakeSessionRepository{resolveFound: false}
	svc := auth.NewSessionService(repo, fixedClock{now: time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)})

	_, err := svc.Validate(context.Background(), rawTestToken)
	assertUnauthorized(t, err)

	if len(repo.resolveCalls) != 1 {
		t.Fatalf("expected exactly one ResolveSessionTenant call, got %d", len(repo.resolveCalls))
	}
	if repo.resolveCalls[0].tokenHash != auth.HashToken(rawTestToken) {
		t.Fatalf("expected the resolver to receive the hashed token, not the raw value")
	}
	if len(repo.renewCalls) != 0 {
		t.Fatal("ValidateAndRenewSession must not be called when the tenant never resolved")
	}
}

// TestSessionService_Validate_ValidSession_RenewsAndReturnsPrincipal cubre
// el camino feliz: renovación deslizante con el reloj inyectado, nunca
// now() real.
func TestSessionService_Validate_ValidSession_RenewsAndReturnsPrincipal(t *testing.T) {
	shop := "11111111-1111-1111-1111-111111111111"
	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: shop}
	repo := &fakeSessionRepository{
		resolveShop: shop, resolveFound: true,
		renewPrincipal: principal, renewValid: true,
	}
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	svc := auth.NewSessionService(repo, fixedClock{now: now})

	got, err := svc.Validate(context.Background(), rawTestToken)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if got != principal {
		t.Fatalf("expected principal %+v, got %+v", principal, got)
	}

	if len(repo.renewCalls) != 1 {
		t.Fatalf("expected exactly one ValidateAndRenewSession call, got %d", len(repo.renewCalls))
	}
	call := repo.renewCalls[0]
	if call.shop != shop {
		t.Fatalf("expected shop %q, got %q", shop, call.shop)
	}
	if call.tokenHash != auth.HashToken(rawTestToken) {
		t.Fatal("expected the renewal to receive the hashed token")
	}
	if !call.now.Equal(now) {
		t.Fatalf("expected now=%v (injected clock), got %v", now, call.now)
	}
	wantExpiry := now.Add(auth.SessionDuration)
	if !call.newExpiresAt.Equal(wantExpiry) {
		t.Fatalf("expected newExpiresAt=%v (30 days from now), got %v", wantExpiry, call.newExpiresAt)
	}
}

// TestSessionService_Validate_ExpiredRevokedOrInactive_AllReturnSameUnauthorized
// cubre CA-006-02/CA-006-03: sesión vencida, revocada o de un usuario
// inactivo -detectadas por ValidateAndRenewSession, no por
// ResolveSessionTenant- producen exactamente el mismo error uniforme, y
// nunca se renueva ni se reabre una sesión que ya no es válida.
func TestSessionService_Validate_ExpiredRevokedOrInactive_AllReturnSameUnauthorized(t *testing.T) {
	shop := "11111111-1111-1111-1111-111111111111"
	repo := &fakeSessionRepository{
		resolveShop: shop, resolveFound: true,
		renewValid: false, // el repositorio ya reconfirmó bajo RLS y rechazó
	}
	svc := auth.NewSessionService(repo, fixedClock{now: time.Now()})

	_, err := svc.Validate(context.Background(), rawTestToken)
	assertUnauthorized(t, err)
}

// TestSessionService_Validate_ResolveRepositoryFailure_ReturnsInternal cubre
// un fallo de infraestructura (no un rechazo de negocio): apperr.KindInternal,
// nunca KindUnauthorized, para no fingir que el motivo fue una credencial
// inválida.
func TestSessionService_Validate_ResolveRepositoryFailure_ReturnsInternal(t *testing.T) {
	repo := &fakeSessionRepository{resolveErr: errors.New("boom")}
	svc := auth.NewSessionService(repo, fixedClock{now: time.Now()})

	_, err := svc.Validate(context.Background(), rawTestToken)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

// TestSessionService_Validate_RenewRepositoryFailure_ReturnsInternal cubre
// el mismo caso para el segundo paso (ValidateAndRenewSession).
func TestSessionService_Validate_RenewRepositoryFailure_ReturnsInternal(t *testing.T) {
	repo := &fakeSessionRepository{
		resolveShop: "11111111-1111-1111-1111-111111111111", resolveFound: true,
		renewErr: errors.New("boom"),
	}
	svc := auth.NewSessionService(repo, fixedClock{now: time.Now()})

	_, err := svc.Validate(context.Background(), rawTestToken)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

// TestSessionService_Validate_CancelledContext_ReturnsInternalWithoutQuerying
// confirma que un contexto ya cancelado nunca llega al repositorio.
func TestSessionService_Validate_CancelledContext_ReturnsInternalWithoutQuerying(t *testing.T) {
	repo := &fakeSessionRepository{}
	svc := auth.NewSessionService(repo, fixedClock{now: time.Now()})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Validate(ctx, rawTestToken)
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
	if len(repo.resolveCalls) != 0 {
		t.Fatal("expected no repository call with an already-cancelled context")
	}
}

// --- Logout -----------------------------------------------------------

// TestSessionService_Logout_RevokesExactlyThePrincipalsOwnSession confirma
// que Logout nunca acepta un identificador ajeno: solo usa
// principal.BarbershopID/SessionID, que el middleware ya validó para ESTA
// solicitud (CA-006-06, CA-006-07).
func TestSessionService_Logout_RevokesExactlyThePrincipalsOwnSession(t *testing.T) {
	repo := &fakeSessionRepository{}
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	svc := auth.NewSessionService(repo, fixedClock{now: now})

	principal := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: "shop-a"}
	if err := svc.Logout(context.Background(), principal); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	if len(repo.revokeCalls) != 1 {
		t.Fatalf("expected exactly one RevokeSession call, got %d", len(repo.revokeCalls))
	}
	call := repo.revokeCalls[0]
	if call.shop != principal.BarbershopID || call.sessionID != principal.SessionID {
		t.Fatalf("expected shop=%q session=%q, got shop=%q session=%q",
			principal.BarbershopID, principal.SessionID, call.shop, call.sessionID)
	}
	if !call.now.Equal(now) {
		t.Fatalf("expected the injected clock's now, got %v", call.now)
	}
}

// TestSessionService_Logout_RepositoryFailure_ReturnsInternal cubre un
// fallo de infraestructura durante la revocación.
func TestSessionService_Logout_RepositoryFailure_ReturnsInternal(t *testing.T) {
	repo := &fakeSessionRepository{revokeErr: errors.New("boom")}
	svc := auth.NewSessionService(repo, fixedClock{now: time.Now()})

	err := svc.Logout(context.Background(), auth.Principal{SessionID: "s-1", BarbershopID: "shop-a"})
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected apperr.KindInternal, got %v", err)
	}
}

// --- Principal en contexto ----------------------------------------------

func TestPrincipalFromContext_RoundTrips(t *testing.T) {
	p := auth.Principal{SessionID: "s-1", StaffUserID: "u-1", BarbershopID: "shop-a"}
	ctx := auth.ContextWithPrincipal(context.Background(), p)

	got, ok := auth.PrincipalFromContext(ctx)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got != p {
		t.Fatalf("expected %+v, got %+v", p, got)
	}
}

func TestPrincipalFromContext_AbsentContext_ReturnsFalse(t *testing.T) {
	_, ok := auth.PrincipalFromContext(context.Background())
	if ok {
		t.Fatal("expected ok=false when no principal was set")
	}
}
