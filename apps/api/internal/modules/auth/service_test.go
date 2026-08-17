package auth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
)

// --- dobles de prueba -------------------------------------------------

type resolveCall struct{ email string }

type lookupCall struct {
	shop     string
	resolved bool
	email    string
}

type createSessionCall struct {
	shop, staffUserID, tokenHash string
	issuedAt, expiresAt          time.Time
}

// fakeRepository es un doble en memoria de auth.Repository: las pruebas de
// este archivo verifican el CASO DE USO en aislamiento, sin PostgreSQL real
// (esa cobertura vive en internal/modules/auth/postgres/repository_test.go
// y en database/tests/hu005_aislamiento_credenciales_sesiones.sql).
type fakeRepository struct {
	resolveShop  string
	resolveFound bool
	resolveErr   error
	resolveCalls []resolveCall

	lookupCred  auth.Credential
	lookupErr   error
	lookupCalls []lookupCall

	createSessionErr   error
	createSessionCalls []createSessionCall
}

func (f *fakeRepository) ResolveLoginTenant(_ context.Context, email string) (string, bool, error) {
	f.resolveCalls = append(f.resolveCalls, resolveCall{email})
	return f.resolveShop, f.resolveFound, f.resolveErr
}

func (f *fakeRepository) LookupCredential(_ context.Context, shop string, resolved bool, email string) (auth.Credential, error) {
	f.lookupCalls = append(f.lookupCalls, lookupCall{shop, resolved, email})
	return f.lookupCred, f.lookupErr
}

func (f *fakeRepository) CreateSession(_ context.Context, shop, staffUserID, tokenHash string, issuedAt, expiresAt time.Time) error {
	f.createSessionCalls = append(f.createSessionCalls, createSessionCall{shop, staffUserID, tokenHash, issuedAt, expiresAt})
	return f.createSessionErr
}

type verifyCall struct{ encodedHash, password string }

// fakeHasher permite observar exactamente cuántas veces y con qué
// argumentos se llamó Verify, sin pagar el costo real de argon2id en cada
// prueba: es lo que hace posible la prueba estructural de CA-005-02 (mismo
// número de llamadas, mismo "hash a verificar" en forma) sin depender de una
// aserción frágil de nanosegundos.
type fakeHasher struct {
	hashPrefix   string
	verifyResult bool
	verifyCalls  []verifyCall
}

func (h *fakeHasher) Hash(password string) (string, error) {
	return h.hashPrefix + password, nil
}

func (h *fakeHasher) Verify(encodedHash, password string) bool {
	h.verifyCalls = append(h.verifyCalls, verifyCall{encodedHash, password})
	return h.verifyResult
}

type fakeTokens struct {
	token string
	err   error
	calls int
}

func (f *fakeTokens) New() (string, error) {
	f.calls++
	return f.token, f.err
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// --- helpers ------------------------------------------------------------

// testIP es la IP usada en las pruebas de este archivo que NO ejercitan
// HU-007 (throttle=nil): su valor es irrelevante porque LoginService.Login
// solo la usa cuando throttle no es nil.
const testIP = "203.0.113.1"

func newTestService(t *testing.T, repo *fakeRepository, hasher *fakeHasher, tokens *fakeTokens, clk fixedClock) *auth.LoginService {
	t.Helper()
	svc, err := auth.NewLoginService(repo, hasher, tokens, clk, nil)
	if err != nil {
		t.Fatalf("NewLoginService: %v", err)
	}
	return svc
}

func requireUnauthorized(t *testing.T, err error) *apperr.Error {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperr.Error, got %v (%T)", err, err)
	}
	if appErr.Kind != apperr.KindUnauthorized {
		t.Fatalf("expected KindUnauthorized, got %v", appErr.Kind)
	}
	return appErr
}

func requireInternal(t *testing.T, err error) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperr.Error, got %v (%T)", err, err)
	}
	if appErr.Kind != apperr.KindInternal {
		t.Fatalf("expected KindInternal, got %v", appErr.Kind)
	}
}

// --- pruebas --------------------------------------------------------------

func TestLogin_Success_CreatesSessionWithHashedToken(t *testing.T) {
	repo := &fakeRepository{
		resolveShop:  "11111111-1111-1111-1111-111111111111",
		resolveFound: true,
		lookupCred: auth.Credential{
			Found: true, StaffUserID: "aaaaaaa1-aaaa-aaaa-aaaa-aaaaaaaaaaa1",
			PasswordHash: "hash-real-almacenado", PasswordAlgorithm: "argon2id",
		},
	}
	hasher := &fakeHasher{verifyResult: true}
	tokens := &fakeTokens{token: "token-en-claro-de-prueba"}
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	svc := newTestService(t, repo, hasher, tokens, fixedClock{now})

	session, err := svc.Login(context.Background(), "duena.a@ejemplo.test", "contraseña-correcta", testIP)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if session.Token != "token-en-claro-de-prueba" {
		t.Fatalf("expected the raw token to flow through to the caller, got %q", session.Token)
	}
	if session.BarbershopID != repo.resolveShop {
		t.Fatalf("expected barbershop %q, got %q", repo.resolveShop, session.BarbershopID)
	}
	if session.StaffUserID != repo.lookupCred.StaffUserID {
		t.Fatalf("expected staff user %q, got %q", repo.lookupCred.StaffUserID, session.StaffUserID)
	}
	if !session.ExpiresAt.Equal(now.Add(auth.SessionDuration)) {
		t.Fatalf("expected expiry 30 days from now, got %v", session.ExpiresAt)
	}

	if len(repo.createSessionCalls) != 1 {
		t.Fatalf("expected exactly one CreateSession call, got %d", len(repo.createSessionCalls))
	}
	call := repo.createSessionCalls[0]
	if call.tokenHash != auth.HashToken("token-en-claro-de-prueba") {
		t.Fatalf("expected the persisted hash to be HashToken(rawToken), got %q", call.tokenHash)
	}
	if call.tokenHash == "token-en-claro-de-prueba" {
		t.Fatal("the raw token must never be passed to the repository as-is")
	}
}

func TestLogin_NormalizesEmail_TrimsAndLowercases(t *testing.T) {
	repo := &fakeRepository{resolveFound: false}
	hasher := &fakeHasher{verifyResult: false}
	svc := newTestService(t, repo, hasher, &fakeTokens{}, fixedClock{time.Now()})

	_, _ = svc.Login(context.Background(), "  DUENA.A@Ejemplo.TEST  ", "cualquiera", testIP)

	if len(repo.resolveCalls) != 1 {
		t.Fatalf("expected exactly one ResolveLoginTenant call, got %d", len(repo.resolveCalls))
	}
	if got := repo.resolveCalls[0].email; got != "duena.a@ejemplo.test" {
		t.Fatalf("expected a normalized email, got %q", got)
	}
}

func TestLogin_WrongPassword_ReturnsUnauthorizedWithoutCreatingSession(t *testing.T) {
	repo := &fakeRepository{
		resolveShop: "11111111-1111-1111-1111-111111111111", resolveFound: true,
		lookupCred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "hash-real"},
	}
	hasher := &fakeHasher{verifyResult: false}
	svc := newTestService(t, repo, hasher, &fakeTokens{}, fixedClock{time.Now()})

	_, err := svc.Login(context.Background(), "duena.a@ejemplo.test", "incorrecta", testIP)

	appErr := requireUnauthorized(t, err)
	if appErr.Message == "" {
		t.Fatal("expected a non-empty safe message")
	}
	if len(repo.createSessionCalls) != 0 {
		t.Fatal("expected no session to be created for a wrong password")
	}
}

func TestLogin_UnknownEmail_VerifiesAgainstDummyHash(t *testing.T) {
	repo := &fakeRepository{resolveFound: false}
	hasher := &fakeHasher{verifyResult: false}
	svc := newTestService(t, repo, hasher, &fakeTokens{}, fixedClock{time.Now()})

	_, err := svc.Login(context.Background(), "no-existe@ejemplo.test", "cualquiera", testIP)

	requireUnauthorized(t, err)

	if len(repo.lookupCalls) != 1 {
		t.Fatalf("expected LookupCredential to still be called once (equalized shape), got %d", len(repo.lookupCalls))
	}
	if repo.lookupCalls[0].resolved {
		t.Fatal("expected resolved=false to reach LookupCredential")
	}
	if len(hasher.verifyCalls) != 1 {
		t.Fatalf("expected exactly one Verify call, got %d", len(hasher.verifyCalls))
	}
	// El hash señuelo es Hash(dummyPassword), calculado una sola vez al
	// construir el servicio; con fakeHasher.Hash(p) = prefix+p, su forma es
	// predecible sin exportar la constante interna dummyPassword.
	if hasher.verifyCalls[0].encodedHash == "" {
		t.Fatal("expected a non-empty dummy hash to be used for verification")
	}
}

// TestLogin_InactiveUser_IsIndistinguishableFromUnknownEmail cubre CA-005-07
// a nivel de servicio: authn_resolve_login_tenant (SQL, probado en
// database/tests/hu005_aislamiento_credenciales_sesiones.sql) ya filtra
// is_active y devuelve found=false para un usuario inactivo exactamente
// igual que para un correo inexistente. Desde la perspectiva de
// LoginService, "usuario inactivo" y "correo inexistente" son la MISMA
// entrada (resolved=false): no hay una rama de código que los distinga.
func TestLogin_InactiveUser_IsIndistinguishableFromUnknownEmail(t *testing.T) {
	repoInactive := &fakeRepository{resolveFound: false}
	repoUnknown := &fakeRepository{resolveFound: false}
	hasher := &fakeHasher{verifyResult: false}

	svcInactive := newTestService(t, repoInactive, hasher, &fakeTokens{}, fixedClock{time.Now()})
	_, errInactive := svcInactive.Login(context.Background(), "barbero.b@ejemplo.test", "cualquiera", testIP)

	svcUnknown := newTestService(t, repoUnknown, hasher, &fakeTokens{}, fixedClock{time.Now()})
	_, errUnknown := svcUnknown.Login(context.Background(), "no-existe@ejemplo.test", "cualquiera", testIP)

	a, b := requireUnauthorized(t, errInactive), requireUnauthorized(t, errUnknown)
	if a.Message != b.Message {
		t.Fatalf("expected identical messages, got %q vs %q", a.Message, b.Message)
	}
}

// TestLogin_StructuralNonEnumeration_SameShapeForUnknownEmailAndWrongPassword
// es la evidencia ESTABLE de CA-005-02 que pide el prompt en vez de una
// aserción frágil de nanosegundos: verifica que ambos caminos de fallo
// ejecutan exactamente el mismo número de llamadas a cada dependencia
// (una resolución, una búsqueda de credencial, una verificación) y producen
// el mismo error observable. Como LookupCredential ya iguala la forma de
// las consultas SQL entre tenant real y señuelo (repository.go), y Verify
// siempre se ejecuta contra un hash con el mismo costo (real o señuelo,
// password.go), esto es evidencia de que el trabajo criptográfico y de base
// de datos es equivalente en ambos casos, sin medir tiempo de reloj.
func TestLogin_StructuralNonEnumeration_SameShapeForUnknownEmailAndWrongPassword(t *testing.T) {
	unknownRepo := &fakeRepository{resolveFound: false}
	unknownHasher := &fakeHasher{verifyResult: false}
	unknownSvc := newTestService(t, unknownRepo, unknownHasher, &fakeTokens{}, fixedClock{time.Now()})
	_, unknownErr := unknownSvc.Login(context.Background(), "no-existe@ejemplo.test", "cualquiera", testIP)

	wrongRepo := &fakeRepository{
		resolveShop: "11111111-1111-1111-1111-111111111111", resolveFound: true,
		lookupCred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "hash-real"},
	}
	wrongHasher := &fakeHasher{verifyResult: false}
	wrongSvc := newTestService(t, wrongRepo, wrongHasher, &fakeTokens{}, fixedClock{time.Now()})
	_, wrongErr := wrongSvc.Login(context.Background(), "duena.a@ejemplo.test", "incorrecta", testIP)

	if len(unknownRepo.resolveCalls) != len(wrongRepo.resolveCalls) {
		t.Fatalf("expected the same number of ResolveLoginTenant calls, got %d vs %d",
			len(unknownRepo.resolveCalls), len(wrongRepo.resolveCalls))
	}
	if len(unknownRepo.lookupCalls) != len(wrongRepo.lookupCalls) {
		t.Fatalf("expected the same number of LookupCredential calls, got %d vs %d",
			len(unknownRepo.lookupCalls), len(wrongRepo.lookupCalls))
	}
	if len(unknownHasher.verifyCalls) != len(wrongHasher.verifyCalls) {
		t.Fatalf("expected the same number of Verify calls, got %d vs %d",
			len(unknownHasher.verifyCalls), len(wrongHasher.verifyCalls))
	}

	a, b := requireUnauthorized(t, unknownErr), requireUnauthorized(t, wrongErr)
	if a.Message != b.Message || a.Kind != b.Kind {
		t.Fatalf("expected identical apperr for both failure paths, got %+v vs %+v", a, b)
	}
}

func TestLogin_ResolveTenantError_ReturnsInternal(t *testing.T) {
	repo := &fakeRepository{resolveErr: errors.New("db unavailable")}
	svc := newTestService(t, repo, &fakeHasher{}, &fakeTokens{}, fixedClock{time.Now()})

	_, err := svc.Login(context.Background(), "duena.a@ejemplo.test", "x", testIP)
	requireInternal(t, err)
}

func TestLogin_LookupCredentialError_ReturnsInternal(t *testing.T) {
	repo := &fakeRepository{resolveFound: true, resolveShop: "shop-1", lookupErr: errors.New("db unavailable")}
	svc := newTestService(t, repo, &fakeHasher{}, &fakeTokens{}, fixedClock{time.Now()})

	_, err := svc.Login(context.Background(), "duena.a@ejemplo.test", "x", testIP)
	requireInternal(t, err)
}

func TestLogin_CreateSessionError_ReturnsInternal(t *testing.T) {
	repo := &fakeRepository{
		resolveFound: true, resolveShop: "shop-1",
		lookupCred:       auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"},
		createSessionErr: errors.New("db unavailable"),
	}
	hasher := &fakeHasher{verifyResult: true}
	tokens := &fakeTokens{token: "tok"}
	svc := newTestService(t, repo, hasher, tokens, fixedClock{time.Now()})

	_, err := svc.Login(context.Background(), "duena.a@ejemplo.test", "correcta", testIP)
	requireInternal(t, err)
}

func TestLogin_TokenGeneratorFails_NeverCreatesSession(t *testing.T) {
	repo := &fakeRepository{
		resolveFound: true, resolveShop: "shop-1",
		lookupCred: auth.Credential{Found: true, StaffUserID: "u-1", PasswordHash: "h"},
	}
	hasher := &fakeHasher{verifyResult: true}
	tokens := &fakeTokens{err: fmt.Errorf("crypto/rand agotado")}
	svc := newTestService(t, repo, hasher, tokens, fixedClock{time.Now()})

	_, err := svc.Login(context.Background(), "duena.a@ejemplo.test", "correcta", testIP)

	requireInternal(t, err)
	if len(repo.createSessionCalls) != 0 {
		t.Fatal("expected no session to be created when token generation fails")
	}
}

func TestLogin_CancelledContext_NeverCallsRepository(t *testing.T) {
	repo := &fakeRepository{resolveFound: true, resolveShop: "shop-1"}
	svc := newTestService(t, repo, &fakeHasher{}, &fakeTokens{}, fixedClock{time.Now()})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Login(ctx, "duena.a@ejemplo.test", "x", testIP)

	requireInternal(t, err)
	if len(repo.resolveCalls) != 0 {
		t.Fatal("expected no repository call once the context was already cancelled")
	}
}

func TestNewLoginService_DummyHashFailure_ReturnsError(t *testing.T) {
	_, err := auth.NewLoginService(&fakeRepository{}, failingDummyHasher{err: errors.New("boom")}, &fakeTokens{}, fixedClock{time.Now()}, nil)
	if err == nil {
		t.Fatal("expected NewLoginService to fail when the hasher cannot compute the dummy hash")
	}
}

// failingDummyHasher es un doble mínimo, separado de fakeHasher, solo para
// forzar el fallo de construcción de NewLoginService (Hash siempre error).
type failingDummyHasher struct{ err error }

func (h failingDummyHasher) Hash(string) (string, error)              { return "", h.err }
func (h failingDummyHasher) Verify(encodedHash, password string) bool { return false }
