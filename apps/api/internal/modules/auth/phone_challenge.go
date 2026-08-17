package auth

import (
	"context"
	"crypto/rand"
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

// PhoneChallengeRepository es el puerto de persistencia del reto telefónico.
// Ambas operaciones corren ANTES de resolver contexto de tenant (igual que
// Repository.ResolveLoginTenant): auth_phone_challenge_request y
// auth_phone_challenge_verify son SECURITY DEFINER estrechas.
type PhoneChallengeRepository interface {
	// RequestChallenge intenta registrar un reto para email desde ipHash
	// con codeHash. accepted=false cubre, indistinguiblemente, cuenta
	// inexistente, teléfono no verificado, IP no escalada y límite propio
	// de reenvío/solicitudes activas excedido; en ese caso phone queda
	// vacío y NINGÚN envío debe intentarse.
	RequestChallenge(ctx context.Context, email, ipHash, codeHash string, cfg PhoneChallengeConfig) (accepted bool, phone string, err error)

	// VerifyChallenge valida codeHash contra el reto vigente de email
	// atado a ipHash. ok=true ya limpió el escalamiento de esa IP en
	// login_throttle, en la misma transacción.
	VerifyChallenge(ctx context.Context, email, ipHash, codeHash string) (ok bool, err error)
}

// PhoneChallengeService implementa el reto telefónico de HU-007 (DEC-062).
// No conoce Chi, net/http ni PostgreSQL.
type PhoneChallengeService struct {
	repo   PhoneChallengeRepository
	codes  PhoneCodeGenerator
	sender PhoneCodeSender
	cfg    PhoneChallengeConfig
	secret []byte
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
	return &PhoneChallengeService{repo: repo, codes: codes, sender: sender, cfg: cfg, secret: secret}
}

// Request genera un código nuevo y, si el repositorio acepta la solicitud
// (cuenta con teléfono verificado, IP realmente escalada, límite propio no
// excedido), lo envía por el puerto de entrega. El resultado de esta
// llamada NUNCA debe cambiar la respuesta HTTP: el handler responde 202
// siempre, sin importar qué devuelva Request (no enumeración, DEC-062). El
// error que sí puede devolver es exclusivamente para diagnóstico interno
// (logging), nunca para decidir el status code.
func (s *PhoneChallengeService) Request(ctx context.Context, rawEmail, ipHash string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de solicitar el reto: %w", err))
	}

	code, err := s.codes.New()
	if err != nil {
		return apperr.Internal(err)
	}
	codeHash := HMACHex(code, s.secret)

	email := NormalizeEmail(rawEmail)
	accepted, phone, err := s.repo.RequestChallenge(ctx, email, ipHash, codeHash, s.cfg)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: solicitar reto telefónico: %w", err))
	}
	if !accepted || phone == "" {
		return nil
	}

	if err := s.sender.SendCode(ctx, phone, code); err != nil {
		// El fallo de entrega se reporta al llamador solo para que lo
		// registre sin destinatario ni código (RN-DAT-02); la respuesta al
		// cliente ya se decidió antes de invocar este método y no cambia.
		return apperr.Internal(fmt.Errorf("auth: enviar código del reto telefónico: %w", err))
	}
	return nil
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
	codeHash := HMACHex(code, s.secret)

	ok, err := s.repo.VerifyChallenge(ctx, email, ipHash, codeHash)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: verificar reto telefónico: %w", err))
	}
	if !ok {
		return errInvalidChallenge()
	}
	return nil
}
