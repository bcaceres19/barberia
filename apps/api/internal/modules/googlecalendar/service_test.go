package googlecalendar_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"system-barbershop/internal/modules/googlecalendar"
	"system-barbershop/internal/platform/apperr"
)

const (
	shopA    = "11111111-1111-1111-1111-111111111111"
	shopB    = "22222222-2222-2222-2222-222222222222"
	userA    = "aaaaaaa2-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
	userA2   = "aaaaaaa3-aaaa-aaaa-aaaa-aaaaaaaaaaa3"
	barberA  = "c0000001-0000-4000-8000-000000000001"
	barberA2 = "c0000001-0000-4000-8000-000000000002"
	session1 = "5e550001-0000-4000-8000-000000000001"
	session2 = "5e550001-0000-4000-8000-000000000002"
)

type fakeClock struct{ now time.Time }

func (c *fakeClock) Now() time.Time { return c.now }

// fixedRandom entrega valores deterministas para poder reconstruir el state.
type fixedRandom struct {
	mu     sync.Mutex
	values []string
}

func (r *fixedRandom) Token(int) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v := r.values[0]
	r.values = r.values[1:]
	return v, nil
}

// fakeRepo es un doble en memoria de googlecalendar.Repository: prueba la
// orquestación del servicio, nunca SQL ni RLS (eso vive en postgres/).
type fakeRepo struct {
	mu          sync.Mutex
	barbers     map[string]string // shop|user -> barber
	connections map[string]googlecalendar.Connection
	states      map[string]*storedState
}

type storedState struct {
	googlecalendar.OAuthState
	consumed bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		barbers: map[string]string{
			shopA + "|" + userA:  barberA,
			shopA + "|" + userA2: barberA2,
		},
		connections: map[string]googlecalendar.Connection{},
		states:      map[string]*storedState{},
	}
}

func (r *fakeRepo) GetConnection(_ context.Context, shop, barber string) (googlecalendar.Connection, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.connections[shop+"|"+barber]
	return c, ok, nil
}

func (r *fakeRepo) SaveConnected(_ context.Context, shop, barber string, d googlecalendar.ConnectedData) (googlecalendar.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	at := d.At
	c := googlecalendar.Connection{
		BarbershopID: shop, BarberID: barber, Status: googlecalendar.StatusConnected,
		AccountEmail: d.AccountEmail, CalendarID: googlecalendar.CalendarPrimary,
		Credentials: d.Credentials, KeyID: d.KeyID, ConnectedAt: &at,
	}
	if prev, ok := r.connections[shop+"|"+barber]; ok {
		c.ReminderMinutes = prev.ReminderMinutes
	}
	r.connections[shop+"|"+barber] = c
	return c, nil
}

func (r *fakeRepo) MarkStatus(_ context.Context, shop, barber string, status googlecalendar.Status, code string, _ time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.connections[shop+"|"+barber]
	if !ok {
		return false, nil
	}
	c.Status, c.LastErrorCode = status, code
	if status == googlecalendar.StatusReauthRequired || status == googlecalendar.StatusDisconnected {
		c.Credentials, c.KeyID = nil, ""
	}
	if status == googlecalendar.StatusDisconnected {
		c.AccountEmail = ""
	}
	r.connections[shop+"|"+barber] = c
	return true, nil
}

func (r *fakeRepo) SetReminder(_ context.Context, shop, barber string, minutes *int) (googlecalendar.Connection, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.connections[shop+"|"+barber]
	if !ok || c.Status == googlecalendar.StatusDisconnected {
		return googlecalendar.Connection{}, false, nil
	}
	c.ReminderMinutes = minutes
	r.connections[shop+"|"+barber] = c
	return c, true, nil
}

func (r *fakeRepo) CreateState(_ context.Context, s googlecalendar.OAuthState, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states[s.StateHash] = &storedState{OAuthState: s}
	return nil
}

func (r *fakeRepo) ConsumeState(_ context.Context, shop, hash string, now time.Time) (googlecalendar.OAuthState, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.states[hash]
	if !ok || s.BarbershopID != shop || s.consumed || !s.ExpiresAt.After(now) {
		return googlecalendar.OAuthState{}, false, nil
	}
	s.consumed = true
	return s.OAuthState, true, nil
}

func (r *fakeRepo) BarberOfUser(_ context.Context, shop, user string) (string, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.barbers[shop+"|"+user]
	return b, ok, nil
}

// fakeProvider es un Google falso que registra lo que recibe.
type fakeProvider struct {
	mu            sync.Mutex
	exchangeCalls []exchangeCall
	exchangeTok   googlecalendar.Tokens
	exchangeErr   error
	refreshErr    error
	revoked       []string
	revokeErr     error
	refreshed     int
}

type exchangeCall struct{ code, verifier string }

func (p *fakeProvider) AuthorizationURL(state, verifier string) string {
	return "https://accounts.example.test/auth?state=" + state + "&challenge=" + verifier
}

func (p *fakeProvider) Exchange(_ context.Context, code, verifier string) (googlecalendar.Tokens, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.exchangeCalls = append(p.exchangeCalls, exchangeCall{code, verifier})
	return p.exchangeTok, p.exchangeErr
}

func (p *fakeProvider) RefreshAccessToken(context.Context, string) (string, time.Time, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.refreshed++
	if p.refreshErr != nil {
		return "", time.Time{}, p.refreshErr
	}
	return "access-token-de-prueba", time.Unix(1_800_000_000, 0), nil
}

func (p *fakeProvider) Revoke(_ context.Context, token string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.revoked = append(p.revoked, token)
	return p.revokeErr
}

type harness struct {
	svc      *googlecalendar.Service
	repo     *fakeRepo
	provider *fakeProvider
	clock    *fakeClock
	cipher   *googlecalendar.Cipher
	random   *fixedRandom
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		repo:     newFakeRepo(),
		provider: &fakeProvider{exchangeTok: googlecalendar.Tokens{RefreshToken: "1//refresh-de-prueba", AccountEmail: "barbero@ejemplo.test"}},
		clock:    &fakeClock{now: time.Date(2026, 10, 9, 15, 0, 0, 0, time.UTC)},
		cipher:   mustCipher(t, "v1", map[string][]byte{"v1": key(1)}),
		random:   &fixedRandom{values: []string{"state-aleatorio-0001", "verificador-pkce-0001-aaaaaaaaaaaaaaaaaaaaaa"}},
	}
	h.svc = googlecalendar.NewService(googlecalendar.Deps{
		Repo: h.repo, Provider: h.provider, Cipher: h.cipher, Clock: h.clock, Random: h.random,
	})
	return h
}

func (h *harness) startAndGetState(t *testing.T) string {
	t.Helper()
	url, err := h.svc.StartConnect(context.Background(), shopA, userA, session1)
	if err != nil {
		t.Fatalf("StartConnect: %v", err)
	}
	const prefix = "https://accounts.example.test/auth?state="
	state := strings.TrimPrefix(strings.SplitN(url, "&", 2)[0], prefix)
	if state == "" {
		t.Fatalf("la URL no lleva state: %q", url)
	}
	return state
}

func requireKind(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != want {
		t.Fatalf("expected apperr kind %q, got %v", want, err)
	}
}

func (h *harness) connect(t *testing.T) {
	t.Helper()
	state := h.startAndGetState(t)
	result, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "codigo-google", "")
	if err != nil || result != googlecalendar.CallbackConnected {
		t.Fatalf("conectar: %v %v", result, err)
	}
}

// --- StartConnect -----------------------------------------------------------

func TestStartConnect_Disabled_IsInvalidState(t *testing.T) {
	svc := googlecalendar.NewService(googlecalendar.Deps{Repo: newFakeRepo(), Clock: &fakeClock{}})
	if svc.Enabled() {
		t.Fatal("sin proveedor ni cifrador la integración está desactivada")
	}
	_, err := svc.StartConnect(context.Background(), shopA, userA, session1)
	requireKind(t, err, apperr.KindInvalidState)
}

func TestStartConnect_WithoutLinkedBarber_IsInvalidState(t *testing.T) {
	h := newHarness(t)
	_, err := h.svc.StartConnect(context.Background(), shopA, "usuario-sin-barbero", session1)
	requireKind(t, err, apperr.KindInvalidState)
}

func TestStartConnect_PersistsOnlyHashedStateAndEncryptedVerifier(t *testing.T) {
	h := newHarness(t)
	url, err := h.svc.StartConnect(context.Background(), shopA, userA, session1)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(url, "state-aleatorio-0001") || !strings.Contains(url, "verificador-pkce-0001") {
		t.Fatalf("la URL debe llevar el state y el verificador para el proveedor: %q", url)
	}

	if len(h.repo.states) != 1 {
		t.Fatalf("debe guardarse un estado, hay %d", len(h.repo.states))
	}
	for hash, st := range h.repo.states {
		if strings.Contains(hash, "state-aleatorio-0001") || len(hash) != 64 {
			t.Fatalf("solo se guarda el hash del state, no el valor: %q", hash)
		}
		if bytes.Contains(st.VerifierEncrypted, []byte("verificador-pkce-0001")) {
			t.Fatal("el verificador PKCE no debe persistirse en claro")
		}
		if st.BarbershopID != shopA || st.BarberID != barberA || st.StaffUserID != userA || st.SessionID != session1 {
			t.Fatalf("el estado debe ligarse a barbería, barbero, usuario y sesión: %+v", st.OAuthState)
		}
		if got := st.ExpiresAt.Sub(h.clock.now); got != googlecalendar.StateTTL {
			t.Fatalf("vigencia = %v, se esperaba %v", got, googlecalendar.StateTTL)
		}
	}
}

// --- CompleteCallback -------------------------------------------------------

func TestCompleteCallback_Success_StoresEncryptedTokenAndUsesPKCEVerifier(t *testing.T) {
	h := newHarness(t)
	h.connect(t)

	if len(h.provider.exchangeCalls) != 1 {
		t.Fatalf("debe canjearse una vez, fueron %d", len(h.provider.exchangeCalls))
	}
	call := h.provider.exchangeCalls[0]
	if call.code != "codigo-google" || call.verifier != "verificador-pkce-0001-aaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("el canje debe usar el código y el verificador guardado: %+v", call)
	}

	conn, found, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
	if !found || conn.Status != googlecalendar.StatusConnected || conn.AccountEmail != "barbero@ejemplo.test" {
		t.Fatalf("conexión inesperada: %+v", conn)
	}
	if bytes.Contains(conn.Credentials, []byte("1//refresh-de-prueba")) {
		t.Fatal("el refresh token no debe persistirse en claro")
	}
	plain, err := h.cipher.Decrypt(conn.Credentials, conn.KeyID, []byte("gcal-refresh|"+shopA+"|"+barberA))
	if err != nil || string(plain) != "1//refresh-de-prueba" {
		t.Fatalf("el token cifrado debe descifrarse con el contexto del barbero: %q %v", plain, err)
	}
	// Trasplantado a otro barbero, ya no se puede leer.
	if _, err := h.cipher.Decrypt(conn.Credentials, conn.KeyID, []byte("gcal-refresh|"+shopA+"|"+barberA2)); err == nil {
		t.Fatal("el token no debe descifrarse en el contexto de otro barbero")
	}
}

func TestCompleteCallback_StateIsSingleUse(t *testing.T) {
	h := newHarness(t)
	state := h.startAndGetState(t)
	ctx := context.Background()

	if _, err := h.svc.CompleteCallback(ctx, shopA, userA, session1, state, "codigo", ""); err != nil {
		t.Fatal(err)
	}
	_, err := h.svc.CompleteCallback(ctx, shopA, userA, session1, state, "codigo", "")
	requireKind(t, err, apperr.KindInvalid)
	if len(h.provider.exchangeCalls) != 1 {
		t.Fatalf("un state ya usado no debe canjear de nuevo, canjes=%d", len(h.provider.exchangeCalls))
	}
}

func TestCompleteCallback_RejectsInvalidStatesWithTheSameError(t *testing.T) {
	cases := map[string]func(h *harness, state string) error{
		"state vacío": func(h *harness, _ string) error {
			_, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, "", "c", "")
			return err
		},
		"state inexistente": func(h *harness, _ string) error {
			_, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, "inventado", "c", "")
			return err
		},
		"otra sesión": func(h *harness, state string) error {
			_, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session2, state, "c", "")
			return err
		},
		"otro usuario": func(h *harness, state string) error {
			_, err := h.svc.CompleteCallback(context.Background(), shopA, userA2, session1, state, "c", "")
			return err
		},
		"otra barbería": func(h *harness, state string) error {
			_, err := h.svc.CompleteCallback(context.Background(), shopB, userA, session1, state, "c", "")
			return err
		},
		"vencido": func(h *harness, state string) error {
			h.clock.now = h.clock.now.Add(googlecalendar.StateTTL + time.Second)
			_, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "c", "")
			return err
		},
		"el vínculo del usuario cambió": func(h *harness, state string) error {
			h.repo.barbers[shopA+"|"+userA] = barberA2
			_, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "c", "")
			return err
		},
	}
	var messages []string
	for name, run := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			state := h.startAndGetState(t)
			err := run(h, state)
			requireKind(t, err, apperr.KindInvalid)
			appErr, _ := apperr.As(err)
			messages = append(messages, appErr.Message)
			if len(h.provider.exchangeCalls) != 0 {
				t.Fatal("un state inválido nunca debe llegar a Google")
			}
			if _, found, _ := h.repo.GetConnection(context.Background(), shopA, barberA); found {
				t.Fatal("un state inválido no debe crear conexión")
			}
		})
	}
	for _, m := range messages[1:] {
		if m != messages[0] {
			t.Fatalf("todos los rechazos deben compartir mensaje para no revelar la causa: %q vs %q", m, messages[0])
		}
	}
}

func TestCompleteCallback_UserDenied_ConnectsNothingAndConsumesState(t *testing.T) {
	h := newHarness(t)
	state := h.startAndGetState(t)
	result, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "", "access_denied")
	if err != nil || result != googlecalendar.CallbackDenied {
		t.Fatalf("result=%v err=%v", result, err)
	}
	if len(h.provider.exchangeCalls) != 0 {
		t.Fatal("un rechazo no debe canjear")
	}
	if _, found, _ := h.repo.GetConnection(context.Background(), shopA, barberA); found {
		t.Fatal("un rechazo no debe crear conexión")
	}
	_, err = h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "x", "")
	requireKind(t, err, apperr.KindInvalid)
}

func TestCompleteCallback_ExchangeFailure_ReportsFailedWithoutSavingOrLeaking(t *testing.T) {
	for name, setup := range map[string]func(p *fakeProvider){
		"Google rechaza el canje": func(p *fakeProvider) { p.exchangeErr = errors.New("invalid_grant: code=SECRETO-DEL-CODIGO") },
		"sin refresh token":       func(p *fakeProvider) { p.exchangeTok = googlecalendar.Tokens{AccountEmail: "x@ejemplo.test"} },
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			setup(h.provider)
			state := h.startAndGetState(t)

			result, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "codigo", "")
			if err != nil || result != googlecalendar.CallbackFailed {
				t.Fatalf("result=%v err=%v", result, err)
			}
			if _, found, _ := h.repo.GetConnection(context.Background(), shopA, barberA); found {
				t.Fatal("un canje fallido no debe crear conexión")
			}
		})
	}
}

func TestCompleteCallback_Reconnect_RevokesThePreviousToken(t *testing.T) {
	h := newHarness(t)
	h.connect(t)

	h.random.values = append(h.random.values, "state-aleatorio-0002", "verificador-pkce-0002-bbbbbbbbbbbbbbbbbbbbbb")
	h.provider.exchangeTok = googlecalendar.Tokens{RefreshToken: "1//refresh-NUEVO", AccountEmail: "otra@ejemplo.test"}
	state := h.startAndGetState(t)
	if _, err := h.svc.CompleteCallback(context.Background(), shopA, userA, session1, state, "codigo2", ""); err != nil {
		t.Fatal(err)
	}
	if len(h.provider.revoked) != 1 || h.provider.revoked[0] != "1//refresh-de-prueba" {
		t.Fatalf("debe revocarse el token anterior, no el nuevo: %v", h.provider.revoked)
	}
	conn, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
	if conn.AccountEmail != "otra@ejemplo.test" {
		t.Fatalf("la conexión debe reflejar la cuenta nueva: %+v", conn)
	}
}

// --- Status, recordatorio y desconexión -------------------------------------

func TestStatus_ReportsEnabledLinkedAndConnection(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()

	view, err := h.svc.Status(ctx, shopA, "sin-barbero")
	if err != nil || !view.Enabled || view.BarberLinked || view.Connection != nil {
		t.Fatalf("sin barbero: %+v %v", view, err)
	}
	view, _ = h.svc.Status(ctx, shopA, userA)
	if !view.BarberLinked || view.Connection != nil {
		t.Fatalf("vinculado sin conexión: %+v", view)
	}
	h.connect(t)
	view, _ = h.svc.Status(ctx, shopA, userA)
	if view.Connection == nil || view.Connection.Status != googlecalendar.StatusConnected {
		t.Fatalf("conectado: %+v", view)
	}
	// El otro usuario de la barbería no ve la conexión de este.
	view, _ = h.svc.Status(ctx, shopA, userA2)
	if view.Connection != nil {
		t.Fatal("cada barbero ve solo su conexión")
	}
}

func TestUpdateReminder_ValidatesRangeAndRequiresAConnection(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	minutes := func(v int) *int { return &v }

	_, err := h.svc.UpdateReminder(ctx, shopA, userA, minutes(30))
	requireKind(t, err, apperr.KindNotFound)

	h.connect(t)
	for _, bad := range []int{-1, 40321} {
		_, err := h.svc.UpdateReminder(ctx, shopA, userA, minutes(bad))
		requireKind(t, err, apperr.KindValidation)
	}
	for _, good := range []int{0, 30, 40320} {
		conn, err := h.svc.UpdateReminder(ctx, shopA, userA, minutes(good))
		if err != nil || conn.ReminderMinutes == nil || *conn.ReminderMinutes != good {
			t.Fatalf("recordatorio %d: %+v %v", good, conn, err)
		}
	}
	conn, err := h.svc.UpdateReminder(ctx, shopA, userA, nil)
	if err != nil || conn.ReminderMinutes != nil {
		t.Fatalf("nil usa los recordatorios predeterminados: %+v %v", conn, err)
	}
}

func TestDisconnect_RevokesClearsCredentialsAndIsIdempotent(t *testing.T) {
	h := newHarness(t)
	ctx := context.Background()
	h.connect(t)

	if err := h.svc.Disconnect(ctx, shopA, userA); err != nil {
		t.Fatal(err)
	}
	if len(h.provider.revoked) != 1 || h.provider.revoked[0] != "1//refresh-de-prueba" {
		t.Fatalf("debe revocar el token guardado: %v", h.provider.revoked)
	}
	conn, _, _ := h.repo.GetConnection(ctx, shopA, barberA)
	if conn.Status != googlecalendar.StatusDisconnected || conn.HasCredentials() || conn.AccountEmail != "" {
		t.Fatalf("debe quedar desconectada y sin credenciales: %+v", conn)
	}
	if err := h.svc.Disconnect(ctx, shopA, userA); err != nil {
		t.Fatalf("desconectar de nuevo es un éxito: %v", err)
	}
	if len(h.provider.revoked) != 1 {
		t.Fatal("desconectar dos veces no debe revocar dos veces")
	}
	if err := h.svc.Disconnect(ctx, shopA, "sin-barbero"); err != nil {
		t.Fatalf("sin barbero también es un éxito: %v", err)
	}
}

func TestDisconnect_GoogleDownOrTokenUnreadable_StillClearsLocally(t *testing.T) {
	t.Run("Google falla al revocar", func(t *testing.T) {
		h := newHarness(t)
		h.connect(t)
		h.provider.revokeErr = errors.New("503")
		if err := h.svc.Disconnect(context.Background(), shopA, userA); err != nil {
			t.Fatal(err)
		}
		conn, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
		if conn.Status != googlecalendar.StatusDisconnected || conn.HasCredentials() {
			t.Fatalf("desconectar nunca depende de Google: %+v", conn)
		}
	})
	t.Run("clave retirada del anillo", func(t *testing.T) {
		h := newHarness(t)
		h.connect(t)
		conn := h.repo.connections[shopA+"|"+barberA]
		conn.KeyID = "clave-perdida"
		h.repo.connections[shopA+"|"+barberA] = conn
		if err := h.svc.Disconnect(context.Background(), shopA, userA); err != nil {
			t.Fatal(err)
		}
		got, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
		if got.HasCredentials() {
			t.Fatal("aunque no se pueda descifrar, las credenciales deben borrarse")
		}
	})
}

func TestBarberLinkReleased_DisconnectsThePreviousBarber(t *testing.T) {
	h := newHarness(t)
	h.connect(t)

	h.svc.BarberLinkReleased(context.Background(), shopA, barberA)

	conn, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
	if conn.Status != googlecalendar.StatusDisconnected || conn.HasCredentials() {
		t.Fatalf("liberar el vínculo debe desconectar y borrar credenciales: %+v", conn)
	}
	if len(h.provider.revoked) != 1 {
		t.Fatalf("debe revocar el token del barbero anterior: %v", h.provider.revoked)
	}
	// Un barbero sin conexión no es un problema.
	h.svc.BarberLinkReleased(context.Background(), shopA, barberA2)
}

// --- AccessToken (token expirado y revocado) ---------------------------------

func TestAccessToken_RefreshesOnDemandWithoutPersistingIt(t *testing.T) {
	h := newHarness(t)
	h.connect(t)

	token, expiresAt, err := h.svc.AccessToken(context.Background(), shopA, barberA)
	if err != nil || token != "access-token-de-prueba" || expiresAt.IsZero() {
		t.Fatalf("AccessToken: %q %v %v", token, expiresAt, err)
	}
	conn, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
	if bytes.Contains(conn.Credentials, []byte(token)) {
		t.Fatal("el access token nunca se persiste")
	}
	if conn.Status != googlecalendar.StatusConnected {
		t.Fatalf("un refresh correcto no cambia el estado: %v", conn.Status)
	}
}

func TestAccessToken_RevokedRefreshToken_MarksReauthRequiredAndDropsCredentials(t *testing.T) {
	h := newHarness(t)
	h.connect(t)
	h.provider.refreshErr = googlecalendar.ErrTokenRevoked

	_, _, err := h.svc.AccessToken(context.Background(), shopA, barberA)
	if !errors.Is(err, googlecalendar.ErrReauthRequired) {
		t.Fatalf("se esperaba ErrReauthRequired, fue %v", err)
	}
	conn, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
	if conn.Status != googlecalendar.StatusReauthRequired || conn.HasCredentials() || conn.LastErrorCode != "token_revoked" {
		t.Fatalf("debe quedar en reauth_required sin credenciales: %+v", conn)
	}
	// Con las credenciales borradas, volver a pedir token ya no llama a Google.
	before := h.provider.refreshed
	if _, _, err := h.svc.AccessToken(context.Background(), shopA, barberA); !errors.Is(err, googlecalendar.ErrReauthRequired) {
		t.Fatalf("segunda llamada: %v", err)
	}
	if h.provider.refreshed != before {
		t.Fatal("sin credenciales no debe llamarse a Google")
	}
}

func TestAccessToken_TransientGoogleFailure_KeepsTheConnection(t *testing.T) {
	h := newHarness(t)
	h.connect(t)
	h.provider.refreshErr = errors.New("503 de Google")

	_, _, err := h.svc.AccessToken(context.Background(), shopA, barberA)
	if err == nil || errors.Is(err, googlecalendar.ErrReauthRequired) {
		t.Fatalf("un fallo transitorio se devuelve tal cual: %v", err)
	}
	conn, _, _ := h.repo.GetConnection(context.Background(), shopA, barberA)
	if conn.Status != googlecalendar.StatusConnected || !conn.HasCredentials() {
		t.Fatalf("un fallo transitorio no debe tocar la conexión: %+v", conn)
	}
}

func TestAccessToken_NoConnection_RequiresReauth(t *testing.T) {
	h := newHarness(t)
	if _, _, err := h.svc.AccessToken(context.Background(), shopA, barberA); !errors.Is(err, googlecalendar.ErrReauthRequired) {
		t.Fatalf("sin conexión: %v", err)
	}
}
