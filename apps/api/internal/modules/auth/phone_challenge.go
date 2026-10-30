package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"system-barbershop/internal/platform/apperr"
)

// invalidChallengeMessage es el ÚNICO texto que PhoneChallengeService
// devuelve cuando la verificación falla, sin importar la causa (código
// incorrecto, vencido, agotado o cuenta inexistente): DEC-062 exige que
// ninguna de esas causas sea distinguible desde la respuesta.
const invalidChallengeMessage = "código de verificación inválido o expirado"

func errInvalidChallenge() error {
	return apperr.Unauthorized(invalidChallengeMessage)
}

// PhoneChallengeConfig agrupa los valores configurables del reto telefónico
// (DEC-062). Ningún valor vive incrustado en el código.
type PhoneChallengeConfig struct {
	ExpiresSeconds        int
	RateWindowSeconds     int
	RateMaxActive         int
	ResendCooldownSeconds int
}

// PhoneCodeGenerator produce el código numérico de 6 dígitos del reto.
type PhoneCodeGenerator interface {
	New() (string, error)
}

// CryptoPhoneCodeGenerator implementa [PhoneCodeGenerator] con crypto/rand,
// distribución uniforme sobre 000000-999999 (DEC-062).
type CryptoPhoneCodeGenerator struct{}

// NewCryptoPhoneCodeGenerator construye el generador.
func NewCryptoPhoneCodeGenerator() CryptoPhoneCodeGenerator { return CryptoPhoneCodeGenerator{} }

var _ PhoneCodeGenerator = CryptoPhoneCodeGenerator{}

// New implementa [PhoneCodeGenerator]. rand.Int con un límite exacto de
// 1_000_000 evita el sesgo de módulo que produciría una reducción ingenua
// de bytes aleatorios.
func (CryptoPhoneCodeGenerator) New() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", fmt.Errorf("auth: generar código del reto telefónico: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// PhoneCodeSender entrega el código en claro al teléfono de un usuario ya
// resuelto por el repositorio. El adaptador concreto (Meta WhatsApp Cloud
// API, DEC-066) pertenece a HU-008; HU-007 define el puerto y una
// implementación mínima documentada como placeholder
// ([LoggingPhoneCodeSender]), para no adelantar una decisión de proveedor
// que no le corresponde a esta historia.
type PhoneCodeSender interface {
	SendCode(ctx context.Context, phone, code string) error
}

// ChallengeEmailSender entrega el código del reto de acceso por correo al
// correo de la cuenta ya resuelto por el repositorio (DEC-139). Mismo criterio
// que [PhoneCodeSender]: auth no conoce el proveedor (DEC-066).
type ChallengeEmailSender interface {
	SendCode(ctx context.Context, email, code string) error
}

// ChallengeDestinations son los destinos que el repositorio derivó de la
// cuenta según el canal configurado de su barbería (DEC-139). Un destino vacío
// significa «canal no configurado o sin contacto utilizable» y NUNCA se
// sustituye por el otro. Nunca provienen del llamador (DEC-081).
type ChallengeDestinations struct {
	Phone string
	Email string
}

// None informa si no hay ningún destino al que entregar.
func (d ChallengeDestinations) None() bool { return d.Phone == "" && d.Email == "" }

// PhoneChallengeRepository es el puerto de persistencia del reto telefónico.
// Ambas operaciones corren ANTES de resolver contexto de tenant (igual que
// Repository.ResolveLoginTenant): auth_phone_challenge_request y
// auth_phone_challenge_verify son SECURITY DEFINER estrechas.
type PhoneChallengeRepository interface {
	// RequestChallenge intenta registrar un reto para email desde ipHash
	// con codeHash. accepted=false cubre, indistinguiblemente, cuenta
	// inexistente o inactiva, canal de la barbería sin contacto utilizable,
	// IP no escalada y límite propio de reenvío/solicitudes activas
	// excedido; en ese caso los destinos quedan vacíos y NINGÚN envío debe
	// intentarse. Con accepted=true, los destinos son exactamente los del
	// canal configurado de la barbería de la cuenta (DEC-139).
	RequestChallenge(ctx context.Context, email, ipHash, codeHash string, cfg PhoneChallengeConfig) (accepted bool, dest ChallengeDestinations, err error)

	// VerifyChallenge valida codeHash contra el reto vigente de email
	// atado a ipHash. ok=true ya limpió el escalamiento de esa IP en
	// login_throttle, en la misma transacción.
	VerifyChallenge(ctx context.Context, email, ipHash, codeHash string) (ok bool, err error)
}

// PhoneChallengeService implementa el reto telefónico de HU-007 (DEC-062).
// No conoce Chi, net/http ni PostgreSQL.
type PhoneChallengeService struct {
	repo     PhoneChallengeRepository
	provider WhatsAppOTPProvider
	email    ChallengeEmailSender
	cfg      PhoneChallengeConfig
}

// NewPhoneChallengeService construye el servicio. secret firma el HMAC del
// código (DEC-062); es el mismo secreto de despliegue que ThrottleService
// usa para la IP, pero cada valor se hashea por separado, nunca combinados
// en una sola entrada.
func NewPhoneChallengeService(
	repo PhoneChallengeRepository,
	codes PhoneCodeGenerator,
	sender PhoneCodeSender,
	cfg PhoneChallengeConfig,
	secret []byte,
) *PhoneChallengeService {
	return NewPhoneChallengeServiceWithOTPProvider(repo, NewLocalWhatsAppOTPProvider(codes, sender, secret), cfg)
}

// NewPhoneChallengeServiceWithOTPProvider constructs the service with the
// provider selected by the composition root.
func NewPhoneChallengeServiceWithOTPProvider(
	repo PhoneChallengeRepository,
	provider WhatsAppOTPProvider,
	cfg PhoneChallengeConfig,
) *PhoneChallengeService {
	return &PhoneChallengeService{repo: repo, provider: provider, cfg: cfg}
}

// WithEmailSender habilita la entrega por correo (canales email y both de
// DEC-139). Sin remitente de correo, un destino de correo se reporta como fallo
// interno de ese canal y no se sustituye por WhatsApp.
func (s *PhoneChallengeService) WithEmailSender(sender ChallengeEmailSender) *PhoneChallengeService {
	s.email = sender
	return s
}

// Request genera un código nuevo y, si el repositorio acepta la solicitud
// (cuenta activa con un destino utilizable en el canal de su barbería, IP
// realmente escalada, límite propio no excedido), lo envía por cada canal
// configurado con el MISMO código y una sola operación lógica (DEC-139). El resultado de esta
// llamada NUNCA debe cambiar la respuesta HTTP: el handler responde 202
// siempre, sin importar qué devuelva Request (no enumeración, DEC-062). El
// error que sí puede devolver es exclusivamente para diagnóstico interno
// (logging), nunca para decidir el status code.
func (s *PhoneChallengeService) Request(ctx context.Context, rawEmail, ipHash string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de solicitar el reto: %w", err))
	}

	prepared, err := s.provider.Prepare(ctx)
	if err != nil {
		return apperr.Internal(err)
	}

	email := NormalizeEmail(rawEmail)
	accepted, dest, err := s.repo.RequestChallenge(ctx, email, ipHash, prepared.PersistenceDigest(), s.cfg)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: solicitar reto telefónico: %w", err))
	}
	if !accepted || dest.None() {
		return nil
	}

	return s.deliver(ctx, dest, prepared)
}

// deliver entrega el código por cada canal con destino. Los canales son
// independientes (DEC-139): el fallo de uno no impide el intento del otro y no
// hay respaldo a un canal no configurado. El error devuelto solo nombra el
// canal que falló —nunca el destinatario ni el código (RN-DAT-02)— y existe
// para diagnóstico interno: la respuesta al cliente ya se decidió antes de
// invocar este método y no cambia.
func (s *PhoneChallengeService) deliver(ctx context.Context, dest ChallengeDestinations, prepared PreparedWhatsAppOTP) error {
	var failures []error
	if dest.Phone != "" {
		if err := s.provider.Deliver(ctx, dest.Phone, prepared); err != nil {
			failures = append(failures, fmt.Errorf("canal whatsapp: %w", err))
		}
	}
	if dest.Email != "" {
		switch {
		case s.email == nil:
			failures = append(failures, errors.New("canal correo: sin remitente configurado"))
		default:
			if err := s.email.SendCode(ctx, dest.Email, prepared.Code()); err != nil {
				failures = append(failures, fmt.Errorf("canal correo: %w", err))
			}
		}
	}
	if len(failures) == 0 {
		return nil
	}
	return apperr.Internal(fmt.Errorf("auth: enviar código del reto: %w", errors.Join(failures...)))
}

// Verify comprueba code contra el reto vigente de rawEmail atado a ipHash.
// Éxito limpia el escalamiento de esa IP (ya lo hizo el repositorio, en la
// misma transacción). Cualquier fallo —código incorrecto, vencido, agotado
// o cuenta inexistente— devuelve exactamente el mismo error uniforme
// (CA-007-03 aplicado al reto: no enumeración).
func (s *PhoneChallengeService) Verify(ctx context.Context, rawEmail, ipHash, code string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de verificar el reto: %w", err))
	}

	email := NormalizeEmail(rawEmail)
	codeHash := s.provider.VerificationDigest(code)

	ok, err := s.repo.VerifyChallenge(ctx, email, ipHash, codeHash)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: verificar reto telefónico: %w", err))
	}
	if !ok {
		return errInvalidChallenge()
	}
	return nil
}
