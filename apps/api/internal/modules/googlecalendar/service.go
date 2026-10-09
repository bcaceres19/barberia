package googlecalendar

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
)

// StateTTL es la vigencia del estado OAuth: lo que tarda una persona en
// consentir en Google. Corta a propósito (DEC-102): un enlace de autorización
// robado caduca pronto y de todos modos queda ligado a la sesión.
const StateTTL = 10 * time.Minute

// ErrReauthRequired indica que la conexión ya no puede publicar: Google
// revocó el permiso (o nunca hubo credenciales) y el barbero debe reconectar.
var ErrReauthRequired = errors.New("googlecalendar: se requiere volver a autorizar con Google")

// CallbackResult es el desenlace del regreso de Google, expresado como un
// código corto que la capa HTTP anexa a la URL de retorno del frontend.
type CallbackResult string

const (
	CallbackConnected CallbackResult = "connected"
	CallbackDenied    CallbackResult = "denied"
	CallbackFailed    CallbackResult = "failed"
)

// StatusView es lo que ve el barbero sobre su integración. Connection es nil
// cuando nunca conectó.
type StatusView struct {
	Enabled      bool
	BarberLinked bool
	Connection   *Connection
	// Jobs resume la cola de la conexión (pendientes y fallidos).
	Jobs JobCounts
}

// Service implementa los casos de uso de la conexión. Con la integración
// desactivada (faltan credenciales de Google) provider y cipher son nil y toda
// operación que lo necesite responde que no está disponible, sin fallar el
// arranque del API.
type Service struct {
	repo     Repository
	provider OAuthProvider
	cipher   *Cipher
	clock    clock.Clock
	random   Randomness
	log      *slog.Logger
}

// Deps agrupa las dependencias del servicio.
type Deps struct {
	Repo     Repository
	Provider OAuthProvider // nil = integración desactivada
	Cipher   *Cipher       // nil = integración desactivada
	Clock    clock.Clock
	Random   Randomness // nil = aleatoriedad criptográfica del sistema
	Logger   *slog.Logger
}

// NewService construye el servicio.
func NewService(deps Deps) *Service {
	random := deps.Random
	if random == nil {
		random = cryptoRandom{}
	}
	logger := deps.Logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Service{
		repo:     deps.Repo,
		provider: deps.Provider,
		cipher:   deps.Cipher,
		clock:    deps.Clock,
		random:   random,
		log:      logger,
	}
}

// Enabled informa si la integración tiene credenciales configuradas.
func (s *Service) Enabled() bool {
	return s.provider != nil && s.cipher != nil
}

type cryptoRandom struct{}

func (cryptoRandom) Token(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashState(state string) string {
	sum := sha256.Sum256([]byte(state))
	return hex.EncodeToString(sum[:])
}

// verifierAAD ata el verificador PKCE cifrado a su estado: copiarlo a otra
// fila no permite descifrarlo.
func verifierAAD(stateHash string) []byte { return []byte("gcal-pkce|" + stateHash) }

// tokenAAD ata el refresh token cifrado a su barbero: trasplantarlo a la fila
// de otro barbero (o de otra barbería) no permite descifrarlo.
func tokenAAD(barbershopID, barberID string) []byte {
	return []byte("gcal-refresh|" + barbershopID + "|" + barberID)
}

// Status devuelve el estado de la integración del usuario autenticado.
func (s *Service) Status(ctx context.Context, barbershopID, staffUserID string) (StatusView, error) {
	view := StatusView{Enabled: s.Enabled()}
	barberID, linked, err := s.repo.BarberOfUser(ctx, barbershopID, staffUserID)
	if err != nil {
		return StatusView{}, apperr.Internal(fmt.Errorf("googlecalendar: resolver barbero del usuario: %w", err))
	}
	if !linked {
		return view, nil
	}
	view.BarberLinked = true

	conn, found, err := s.repo.GetConnection(ctx, barbershopID, barberID)
	if err != nil {
		return StatusView{}, apperr.Internal(fmt.Errorf("googlecalendar: leer conexión: %w", err))
	}
	if found {
		view.Connection = &conn
		counts, err := s.repo.JobCounts(ctx, barbershopID, barberID)
		if err != nil {
			return StatusView{}, apperr.Internal(fmt.Errorf("googlecalendar: contar trabajos: %w", err))
		}
		view.Jobs = counts
	}
	return view, nil
}

// SyncNow es «Sincronizar ahora»: hace vencer ya los trabajos pendientes de la
// conexión del barbero y reintenta los fallidos. No llama a Google (lo hace el
// worker) ni toca el dominio de NAVA, solo reordena la cola de ESA conexión, y es
// segura ante varios clics.
func (s *Service) SyncNow(ctx context.Context, barbershopID, staffUserID string) (JobCounts, error) {
	if !s.Enabled() {
		return JobCounts{}, errIntegrationDisabled()
	}
	barberID, linked, err := s.repo.BarberOfUser(ctx, barbershopID, staffUserID)
	if err != nil {
		return JobCounts{}, apperr.Internal(fmt.Errorf("googlecalendar: resolver barbero del usuario: %w", err))
	}
	if !linked {
		return JobCounts{}, errNoLinkedBarber()
	}
	counts, found, err := s.repo.RequeueConnection(ctx, barbershopID, barberID, s.clock.Now())
	if err != nil {
		return JobCounts{}, apperr.Internal(fmt.Errorf("googlecalendar: sincronizar ahora: %w", err))
	}
	if !found {
		return JobCounts{}, errNotConnected()
	}
	return counts, nil
}

// StartConnect inicia la autorización: guarda un estado de un solo uso ligado a
// barbería, barbero, usuario y sesión, y devuelve la URL de consentimiento.
func (s *Service) StartConnect(ctx context.Context, barbershopID, staffUserID, sessionID string) (string, error) {
	if !s.Enabled() {
		return "", errIntegrationDisabled()
	}
	barberID, linked, err := s.repo.BarberOfUser(ctx, barbershopID, staffUserID)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: resolver barbero del usuario: %w", err))
	}
	if !linked {
		return "", errNoLinkedBarber()
	}

	state, err := s.random.Token(32)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: generar state: %w", err))
	}
	// 32 bytes en base64 URL = 43 caracteres: el largo mínimo que RFC 7636 admite.
	verifier, err := s.random.Token(32)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: generar verificador PKCE: %w", err))
	}

	stateHash := hashState(state)
	encrypted, keyID, err := s.cipher.Encrypt([]byte(verifier), verifierAAD(stateHash))
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: cifrar verificador PKCE: %w", err))
	}

	now := s.clock.Now()
	if err := s.repo.CreateState(ctx, OAuthState{
		BarbershopID:      barbershopID,
		BarberID:          barberID,
		StaffUserID:       staffUserID,
		SessionID:         sessionID,
		StateHash:         stateHash,
		VerifierEncrypted: encrypted,
		VerifierKeyID:     keyID,
		ExpiresAt:         now.Add(StateTTL),
	}, now); err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: guardar estado OAuth: %w", err))
	}
	return s.provider.AuthorizationURL(state, verifier), nil
}

// CompleteCallback procesa el regreso de Google. Nunca devuelve un error
// detallado al navegador: cualquier estado inválido responde el mismo
// apperr.Invalid y un fallo del intercambio, CallbackFailed. El estado se
// consume ANTES de cualquier otra comprobación, así que un intento fallido
// tampoco puede reintentarse con el mismo enlace.
func (s *Service) CompleteCallback(
	ctx context.Context,
	barbershopID, staffUserID, sessionID, rawState, code, providerError string,
) (CallbackResult, error) {
	if !s.Enabled() {
		return "", errIntegrationDisabled()
	}
	if rawState == "" {
		return "", errAuthorizationInvalid()
	}

	state, found, err := s.repo.ConsumeState(ctx, barbershopID, hashState(rawState), s.clock.Now())
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: consumir estado OAuth: %w", err))
	}
	// Ajeno (otra sesión o usuario) se trata igual que inexistente.
	if !found || state.StaffUserID != staffUserID || state.SessionID != sessionID {
		return "", errAuthorizationInvalid()
	}

	// El barbero debe seguir siendo el del usuario: si cambió su vínculo entre
	// iniciar y terminar, la autorización ya no corresponde.
	barberID, linked, err := s.repo.BarberOfUser(ctx, barbershopID, staffUserID)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: resolver barbero del usuario: %w", err))
	}
	if !linked || barberID != state.BarberID {
		return "", errAuthorizationInvalid()
	}

	if providerError != "" || code == "" {
		return CallbackDenied, nil
	}

	verifier, err := s.cipher.Decrypt(state.VerifierEncrypted, state.VerifierKeyID, verifierAAD(state.StateHash))
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: descifrar verificador PKCE: %w", err))
	}

	tokens, err := s.provider.Exchange(ctx, code, string(verifier))
	if err != nil || tokens.RefreshToken == "" {
		// Sin detalle: el error del proveedor puede traer partes del código.
		s.log.Warn("google_calendar.connect.exchange_failed", "barber_id", barberID)
		return CallbackFailed, nil
	}

	previous, hadPrevious, err := s.repo.GetConnection(ctx, barbershopID, barberID)
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: leer conexión previa: %w", err))
	}

	encrypted, keyID, err := s.cipher.Encrypt([]byte(tokens.RefreshToken), tokenAAD(barbershopID, barberID))
	if err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: cifrar refresh token: %w", err))
	}
	if _, err := s.repo.SaveConnected(ctx, barbershopID, barberID, ConnectedData{
		AccountEmail: tokens.AccountEmail,
		Credentials:  encrypted,
		KeyID:        keyID,
		At:           s.clock.Now(),
	}); err != nil {
		return "", apperr.Internal(fmt.Errorf("googlecalendar: guardar conexión: %w", err))
	}

	// Reconectar con otra cuenta deja un token anterior vivo en Google: se
	// revoca (mejor esfuerzo) para no acumular permisos olvidados.
	if hadPrevious && previous.HasCredentials() {
		s.revokeStored(ctx, previous, barbershopID, barberID)
	}
	s.log.Info("google_calendar.connect.connected", "barber_id", barberID)
	return CallbackConnected, nil
}

// UpdateReminder guarda la anticipación del recordatorio del barbero.
func (s *Service) UpdateReminder(ctx context.Context, barbershopID, staffUserID string, minutes *int) (Connection, error) {
	if err := ValidateReminderMinutes(minutes); err != nil {
		return Connection{}, err
	}
	barberID, linked, err := s.repo.BarberOfUser(ctx, barbershopID, staffUserID)
	if err != nil {
		return Connection{}, apperr.Internal(fmt.Errorf("googlecalendar: resolver barbero del usuario: %w", err))
	}
	if !linked {
		return Connection{}, errNotConnected()
	}
	conn, found, err := s.repo.SetReminder(ctx, barbershopID, barberID, minutes)
	if err != nil {
		return Connection{}, apperr.Internal(fmt.Errorf("googlecalendar: guardar recordatorio: %w", err))
	}
	if !found {
		return Connection{}, errNotConnected()
	}
	return conn, nil
}

// Disconnect desconecta al barbero del usuario. Es idempotente: sin barbero o
// sin conexión también es un éxito.
func (s *Service) Disconnect(ctx context.Context, barbershopID, staffUserID string) error {
	barberID, linked, err := s.repo.BarberOfUser(ctx, barbershopID, staffUserID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("googlecalendar: resolver barbero del usuario: %w", err))
	}
	if !linked {
		return nil
	}
	return s.disconnectBarber(ctx, barbershopID, barberID)
}

// BarberLinkReleased implementa staff.LinkObserver (DEC-100): si el usuario
// cambia o quita su vínculo, se desconecta la conexión del barbero anterior,
// revocando y borrando sus credenciales, para que quien tome ese barbero
// después no herede una cuenta de Google ajena.
func (s *Service) BarberLinkReleased(ctx context.Context, barbershopID, barberID string) {
	if err := s.disconnectBarber(ctx, barbershopID, barberID); err != nil {
		// El vínculo ya cambió: no hay a quién devolverle el error. Se deja
		// constancia sin datos personales para reconciliarlo.
		s.log.Error("google_calendar.disconnect.after_link_released_failed", "barber_id", barberID, "error", err)
	}
}

func (s *Service) disconnectBarber(ctx context.Context, barbershopID, barberID string) error {
	conn, found, err := s.repo.GetConnection(ctx, barbershopID, barberID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("googlecalendar: leer conexión: %w", err))
	}
	if !found || conn.Status == StatusDisconnected {
		return nil
	}
	if conn.HasCredentials() {
		s.revokeStored(ctx, conn, barbershopID, barberID)
	}
	if _, err := s.repo.MarkStatus(ctx, barbershopID, barberID, StatusDisconnected, "", s.clock.Now()); err != nil {
		return apperr.Internal(fmt.Errorf("googlecalendar: desconectar: %w", err))
	}
	s.log.Info("google_calendar.disconnect.disconnected", "barber_id", barberID)
	return nil
}

// revokeStored revoca en Google el refresh token guardado. Es de mejor
// esfuerzo: desconectar localmente NUNCA depende de que Google responda.
func (s *Service) revokeStored(ctx context.Context, conn Connection, barbershopID, barberID string) {
	if !s.Enabled() {
		return
	}
	token, err := s.cipher.Decrypt(conn.Credentials, conn.KeyID, tokenAAD(barbershopID, barberID))
	if err != nil {
		s.log.Warn("google_calendar.revoke.undecryptable", "barber_id", barberID)
		return
	}
	if err := s.provider.Revoke(ctx, string(token)); err != nil {
		s.log.Warn("google_calendar.revoke.failed", "barber_id", barberID)
	}
}

// AccessToken obtiene un access token vigente del barbero para llamar a la API
// de Google Calendar; lo usa el publicador (issue #324). El access token no se
// persiste (DEC-102): se renueva bajo demanda con el refresh token. Si Google
// ya no acepta el refresh token, la conexión pasa a `reauth_required` con las
// credenciales borradas y se devuelve ErrReauthRequired; un fallo transitorio
// se devuelve tal cual y NO cambia el estado.
func (s *Service) AccessToken(ctx context.Context, barbershopID, barberID string) (string, time.Time, error) {
	if !s.Enabled() {
		return "", time.Time{}, errIntegrationDisabled()
	}
	conn, found, err := s.repo.GetConnection(ctx, barbershopID, barberID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("googlecalendar: leer conexión: %w", err)
	}
	if !found || !conn.HasCredentials() {
		return "", time.Time{}, ErrReauthRequired
	}
	refreshToken, err := s.cipher.Decrypt(conn.Credentials, conn.KeyID, tokenAAD(barbershopID, barberID))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("googlecalendar: descifrar refresh token: %w", err)
	}

	access, expiresAt, err := s.provider.RefreshAccessToken(ctx, string(refreshToken))
	if errors.Is(err, ErrTokenRevoked) {
		if _, markErr := s.repo.MarkStatus(ctx, barbershopID, barberID, StatusReauthRequired, "token_revoked", s.clock.Now()); markErr != nil {
			return "", time.Time{}, fmt.Errorf("googlecalendar: marcar reauth_required: %w", markErr)
		}
		s.log.Warn("google_calendar.token.revoked", "barber_id", barberID)
		return "", time.Time{}, ErrReauthRequired
	}
	if err != nil {
		return "", time.Time{}, fmt.Errorf("googlecalendar: renovar access token: %w", err)
	}
	return access, expiresAt, nil
}
