package auth

import (
	"context"
	"fmt"
	"strings"

	"system-barbershop/internal/platform/apperr"
)

// invalidRecoveryCodeMessage es el ÚNICO texto que RecoveryService.Verify
// devuelve cuando la verificación falla, sin importar la causa (código
// incorrecto, vencido, agotado o cuenta inexistente): DEC-065 exige que
// ninguna de esas causas sea distinguible desde la respuesta (CA-008-01).
const invalidRecoveryCodeMessage = "código de recuperación inválido o expirado"

// invalidResetTokenMessage es el ÚNICO texto que RecoveryService.ChangePassword
// devuelve cuando el token de reinicio no autoriza el cambio: desconocido,
// de otra cuenta, ya consumido o vencido (DEC-064).
const invalidResetTokenMessage = "token de restablecimiento inválido o expirado"

func errInvalidRecoveryCode() error { return apperr.Unauthorized(invalidRecoveryCodeMessage) }
func errInvalidResetToken() error   { return apperr.Unauthorized(invalidResetTokenMessage) }

// RecoveryConfig agrupa los valores configurables del flujo de recuperación
// (DEC-064). Ningún valor vive incrustado en el código.
type RecoveryConfig struct {
	CodeExpiresSeconds       int
	CodeMaxAttempts          int
	ResendCooldownSeconds    int
	ResendWindowSeconds      int
	ResendMaxPerWindow       int
	ResetTokenExpiresSeconds int
}

// RecoveryCodeSender entrega el código en claro por WhatsApp oficial y
// correo a un usuario ya resuelto por el repositorio (DEC-066). El
// adaptador concreto (Meta WhatsApp Cloud API + Resend/SES) vive en el
// módulo notification; el núcleo de auth no conoce ninguno de los dos
// proveedores.
type RecoveryCodeSender interface {
	SendCode(ctx context.Context, phone, email, code string) error
}

// RecoveryRepository es el puerto de persistencia del flujo de
// recuperación. Las tres primeras operaciones corren ANTES de resolver
// contexto de tenant (igual que Repository.ResolveLoginTenant y
// PhoneChallengeRepository): auth_recovery_request/verify/
// current_credential/change_password son SECURITY DEFINER estrechas.
type RecoveryRepository interface {
	// RequestRecovery intenta registrar un código para email con codeHash.
	// accepted=false cubre, indistinguiblemente, cuenta inexistente,
	// teléfono no verificado y límite propio de cooldown/reenvío excedido;
	// en ese caso phone/resolvedEmail quedan vacíos y NINGÚN envío debe
	// intentarse.
	RequestRecovery(ctx context.Context, email, codeHash string, cfg RecoveryConfig) (accepted bool, phone, resolvedEmail string, err error)

	// VerifyRecovery valida codeHash contra el código vigente de email y,
	// si coincide, persiste resetTokenHash con vigencia
	// resetExpiresSeconds en la misma fila. ok=false cubre código
	// incorrecto, vencido, agotado o cuenta inexistente; phone/resolvedEmail
	// solo se completan cuando ok=true.
	VerifyRecovery(ctx context.Context, email, codeHash, resetTokenHash string, resetExpiresSeconds int) (ok bool, phone, resolvedEmail string, err error)

	// CurrentCredential expone el hash codificado vigente de la cuenta
	// autorizada por resetTokenHash (sin consumirlo), para que el servicio
	// pueda rechazar una contraseña nueva idéntica a la actual (DEC-063).
	// found=false cubre token desconocido, de otra cuenta, ya consumido o
	// vencido.
	CurrentCredential(ctx context.Context, email, resetTokenHash string) (found bool, passwordHash, passwordAlgorithm string, err error)

	// ChangePassword consume resetTokenHash y, en la misma operación
	// atómica, actualiza la credencial y revoca TODAS las sesiones activas
	// del usuario (CA-008-05). ok=false cubre token desconocido, de otra
	// cuenta, ya consumido o vencido: ninguna fila se modifica en ese caso.
	ChangePassword(ctx context.Context, email, resetTokenHash, newPasswordHash string) (ok bool, err error)
}

// RecoveryService implementa el caso de uso de HU-008: solicitar,
// verificar y establecer una contraseña nueva. No conoce Chi, net/http,
// JSON ni PostgreSQL.
type RecoveryService struct {
	repo   RecoveryRepository
	codes  PhoneCodeGenerator // mismo generador de 6 dígitos que HU-007 (DEC-064: mismo patrón criptográfico)
	tokens TokenGenerator     // mismo generador opaco que la sesión (DEC-064)
	sender RecoveryCodeSender
	hasher PasswordHasher
	cfg    RecoveryConfig
	secret []byte
}

// NewRecoveryService construye el servicio. secret firma el HMAC del
// código (DEC-064); es el mismo secreto de despliegue que ThrottleService/
// PhoneChallengeService usan, hasheando cada valor por separado.
func NewRecoveryService(
	repo RecoveryRepository,
	codes PhoneCodeGenerator,
	tokens TokenGenerator,
	sender RecoveryCodeSender,
	hasher PasswordHasher,
	cfg RecoveryConfig,
	secret []byte,
) *RecoveryService {
	return &RecoveryService{repo: repo, codes: codes, tokens: tokens, sender: sender, hasher: hasher, cfg: cfg, secret: secret}
}

// Request genera un código nuevo y, si el repositorio acepta la solicitud
// (cuenta activa con teléfono verificado, cooldown/límite de reenvío no
// excedidos), lo envía por WhatsApp y correo. El resultado de esta llamada
// NUNCA debe cambiar la respuesta HTTP: el handler responde 202 siempre,
// sin importar qué devuelva Request (no enumeración, DEC-065). El error
// que sí puede devolver es exclusivamente para diagnóstico interno
// (logging), nunca para decidir el status code.
func (s *RecoveryService) Request(ctx context.Context, rawEmail string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de solicitar recuperación: %w", err))
	}

	code, err := s.codes.New()
	if err != nil {
		return apperr.Internal(err)
	}
	codeHash := HMACHex(code, s.secret)

	email := NormalizeEmail(rawEmail)
	accepted, phone, resolvedEmail, err := s.repo.RequestRecovery(ctx, email, codeHash, s.cfg)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: solicitar recuperación: %w", err))
	}
	if !accepted || phone == "" {
		return nil
	}

	if err := s.sender.SendCode(ctx, phone, resolvedEmail, code); err != nil {
		// El fallo de entrega se reporta al llamador solo para que lo
		// registre sin destinatario ni código (RN-DAT-02); la respuesta al
		// cliente ya se decidió antes de invocar este método y no cambia
		// (DEC-066: fallo parcial de un canal no filtra existencia de cuenta).
		return apperr.Internal(fmt.Errorf("auth: enviar código de recuperación: %w", err))
	}
	return nil
}

// Verify comprueba code contra el código vigente de rawEmail. Éxito emite
// un token de reinicio opaco (mismo patrón CryptoTokenGenerator+HashToken
// que la sesión, DEC-064) y devuelve el destino enmascarado (DEC-065).
// Cualquier fallo —código incorrecto, vencido, agotado o cuenta
// inexistente— devuelve exactamente el mismo error uniforme.
func (s *RecoveryService) Verify(ctx context.Context, rawEmail, code string) (resetToken, maskedPhone, maskedEmail string, err error) {
	if err := ctx.Err(); err != nil {
		return "", "", "", apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de verificar recuperación: %w", err))
	}

	rawToken, err := s.tokens.New()
	if err != nil {
		return "", "", "", apperr.Internal(fmt.Errorf("auth: generar token de reinicio: %w", err))
	}
	tokenHash := HashToken(rawToken)

	email := NormalizeEmail(rawEmail)
	codeHash := HMACHex(code, s.secret)

	ok, phone, resolvedEmail, err := s.repo.VerifyRecovery(ctx, email, codeHash, tokenHash, s.cfg.ResetTokenExpiresSeconds)
	if err != nil {
		return "", "", "", apperr.Internal(fmt.Errorf("auth: verificar recuperación: %w", err))
	}
	if !ok {
		return "", "", "", errInvalidRecoveryCode()
	}

	return rawToken, MaskPhone(phone), MaskEmail(resolvedEmail), nil
}

// ChangePassword valida el token de reinicio, la política de contraseña
// (DEC-063) y, si ambas pasan, deriva el hash nuevo y lo persiste junto con
// la revocación de todas las sesiones del usuario en una misma operación
// atómica (CA-008-05). El artefacto se consume una vez incluso bajo dos
// solicitudes concurrentes: el repositorio serializa la segunda detrás de
// la primera.
func (s *RecoveryService) ChangePassword(ctx context.Context, rawEmail, resetToken, newPassword string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("auth: contexto cancelado antes de cambiar contraseña: %w", err))
	}

	email := NormalizeEmail(rawEmail)
	tokenHash := HashToken(resetToken)

	found, currentHash, _, err := s.repo.CurrentCredential(ctx, email, tokenHash)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: leer credencial vigente: %w", err))
	}
	if !found {
		return errInvalidResetToken()
	}

	if err := ValidateNewPassword(newPassword, email, currentHash, s.hasher); err != nil {
		return err
	}

	newHash, err := s.hasher.Hash(newPassword)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: derivar hash de contraseña nueva: %w", err))
	}

	ok, err := s.repo.ChangePassword(ctx, email, tokenHash, newHash)
	if err != nil {
		return apperr.Internal(fmt.Errorf("auth: cambiar contraseña: %w", err))
	}
	if !ok {
		// El token se consumió o venció entre CurrentCredential y esta
		// llamada (carrera legítima entre dos solicitudes concurrentes con
		// el mismo token, CA-008-05): mismo error uniforme que un token
		// inválido, nunca un estado distinto.
		return errInvalidResetToken()
	}
	return nil
}

// passwordMinLen y passwordMaxLen fijan la política de DEC-063: 10-128
// caracteres, límite defensivo de tamaño de entrada para Argon2id, no una
// regla de seguridad por sí misma.
const (
	passwordMinLen = 10
	passwordMaxLen = 128
)

// ValidateNewPassword aplica la política de contraseña de DEC-063: longitud
// 10-128, distinta del correo de la cuenta y de la contraseña actual. El
// mensaje indica exactamente qué regla incumple (CA-008-08). currentHash
// puede ser cadena vacía cuando no hay credencial que comparar (no debería
// ocurrir en producción, pero ValidateNewPassword no asume su presencia).
func ValidateNewPassword(password, email, currentHash string, hasher PasswordHasher) error {
	switch {
	case len(password) < passwordMinLen:
		return apperr.Validation(fmt.Sprintf("la contraseña debe tener al menos %d caracteres", passwordMinLen))
	case len(password) > passwordMaxLen:
		return apperr.Validation(fmt.Sprintf("la contraseña no puede superar %d caracteres", passwordMaxLen))
	case strings.EqualFold(password, email):
		return apperr.Validation("la contraseña no puede ser igual al correo de la cuenta")
	}
	if currentHash != "" && hasher.Verify(currentHash, password) {
		return apperr.Validation("la contraseña no puede ser igual a la contraseña actual")
	}
	return nil
}
