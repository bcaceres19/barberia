package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// recoveryAcceptedMessage es el ÚNICO texto que POST /auth/recovery/request
// devuelve, sin importar si la cuenta existe o si el teléfono está
// verificado (DEC-065, no enumeración).
const recoveryAcceptedMessage = "Si la cuenta existe, se envió un código a los medios registrados."

// maxNewPasswordLen es un límite defensivo de tamaño de cuerpo HTTP, no la
// política de contraseña: esa la aplica auth.ValidateNewPassword (DEC-063,
// 10-128 caracteres) dentro del servicio, con el mensaje específico que
// CA-008-08 exige.
const maxNewPasswordLen = 512

var recoveryCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

// RecoveryRequestHandler decodifica, valida la forma e invoca
// auth.RecoveryService.Request. No contiene reglas de negocio.
type RecoveryRequestHandler struct {
	service *auth.RecoveryService
	logger  *slog.Logger
}

// NewRecoveryRequestHandler construye el handler de solicitud de
// recuperación. logger registra únicamente un fallo de entrega, sin
// destinatario ni código (RN-DAT-02, DEC-066); nunca decide la respuesta.
func NewRecoveryRequestHandler(service *auth.RecoveryService, logger *slog.Logger) *RecoveryRequestHandler {
	return &RecoveryRequestHandler{service: service, logger: logger}
}

// ServeHTTP implementa http.Handler. Responde SIEMPRE 202 tras una forma de
// solicitud válida, sin importar el resultado interno del servicio
// (DEC-065): el resultado de auth.RecoveryService.Request nunca decide el
// status code, solo se registra internamente para diagnóstico.
func (h *RecoveryRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	var req RecoveryRequestRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}
	if dec.More() {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}

	switch {
	case req.Email == "":
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("email es obligatorio"), requestID))
		return
	case len(req.Email) > maxEmailLen:
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("email excede el largo máximo"), requestID))
		return
	}

	if err := h.service.Request(r.Context(), req.Email); err != nil {
		// El resultado se descarta deliberadamente para la respuesta
		// (DEC-065); solo se registra sin destinatario, código ni detalle
		// de proveedor (RN-DAT-02).
		h.logger.WarnContext(r.Context(), "auth: fallo al procesar solicitud de recuperación", "requestId", requestID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(RecoveryRequestAcceptedResponse{Message: recoveryAcceptedMessage})
}

// RecoveryVerifyHandler decodifica, valida la forma e invoca
// auth.RecoveryService.Verify.
type RecoveryVerifyHandler struct {
	service *auth.RecoveryService
}

// NewRecoveryVerifyHandler construye el handler de verificación.
func NewRecoveryVerifyHandler(service *auth.RecoveryService) *RecoveryVerifyHandler {
	return &RecoveryVerifyHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *RecoveryVerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	var req RecoveryVerifyRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}
	if dec.More() {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}

	switch {
	case req.Email == "":
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("email es obligatorio"), requestID))
		return
	case len(req.Email) > maxEmailLen:
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("email excede el largo máximo"), requestID))
		return
	case req.Code == "":
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("code es obligatorio"), requestID))
		return
	case !recoveryCodePattern.MatchString(req.Code):
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("code debe tener exactamente 6 dígitos"), requestID))
		return
	}

	resetToken, maskedPhone, maskedEmail, err := h.service.Verify(r.Context(), req.Email, req.Code)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(RecoveryVerifyResponse{
		ResetToken:  resetToken,
		MaskedPhone: maskedPhone,
		MaskedEmail: maskedEmail,
	})
}

// RecoveryResetPasswordHandler decodifica, valida la forma e invoca
// auth.RecoveryService.ChangePassword.
type RecoveryResetPasswordHandler struct {
	service *auth.RecoveryService
}

// NewRecoveryResetPasswordHandler construye el handler de cambio de contraseña.
func NewRecoveryResetPasswordHandler(service *auth.RecoveryService) *RecoveryResetPasswordHandler {
	return &RecoveryResetPasswordHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *RecoveryResetPasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	var req RecoveryResetPasswordRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
			return
		}
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}
	if dec.More() {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Invalid("cuerpo JSON inválido o con un campo desconocido"), requestID))
		return
	}

	switch {
	case req.Email == "":
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("email es obligatorio"), requestID))
		return
	case len(req.Email) > maxEmailLen:
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("email excede el largo máximo"), requestID))
		return
	case req.ResetToken == "":
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("resetToken es obligatorio"), requestID))
		return
	case req.NewPassword == "":
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("newPassword es obligatorio"), requestID))
		return
	case len(req.NewPassword) > maxNewPasswordLen:
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("newPassword excede el largo máximo"), requestID))
		return
	}

	if err := h.service.ChangePassword(r.Context(), req.Email, req.ResetToken, req.NewPassword); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
