package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clientip"
	"system-barbershop/internal/platform/httpserver"
)

// challengeAcceptedMessage es el ÚNICO texto que POST /auth/challenge
// devuelve, sin importar si la cuenta existe, si el teléfono está
// verificado o si la IP está realmente escalada (DEC-062, no enumeración).
const challengeAcceptedMessage = "Si la cuenta existe y su teléfono está verificado, recibirá un código por WhatsApp."

var challengeCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

// ChallengeHandler decodifica, valida la forma e invoca
// auth.PhoneChallengeService.Request. No contiene reglas de negocio.
type ChallengeHandler struct {
	service        *auth.PhoneChallengeService
	throttle       *auth.ThrottleService
	trustedProxies clientip.TrustedProxies
}

// NewChallengeHandler construye el handler de solicitud del reto. throttle
// solo se usa para derivar el mismo ip_hash que el login (HashIP), nunca
// para registrar un intento nuevo: pedir el reto no cuenta como un intento
// de acceso.
func NewChallengeHandler(service *auth.PhoneChallengeService, throttle *auth.ThrottleService, trustedProxies clientip.TrustedProxies) *ChallengeHandler {
	return &ChallengeHandler{service: service, throttle: throttle, trustedProxies: trustedProxies}
}

// ServeHTTP implementa http.Handler. Responde SIEMPRE 202 tras una forma de
// solicitud válida, sin importar el resultado interno del servicio
// (DEC-062): el resultado de auth.PhoneChallengeService.Request nunca
// decide el status code, solo se usa para diagnóstico interno si algún día
// el handler agrega logging explícito.
func (h *ChallengeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	var req ChallengeRequest
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

	rawIP, err := clientip.Resolve(r, h.trustedProxies)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Internal(err), requestID))
		return
	}
	ipHash := h.throttle.HashIP(rawIP)

	// El resultado se descarta deliberadamente para la respuesta (DEC-062);
	// Request ya registra internamente cualquier fallo de entrega sin
	// destinatario ni código.
	_ = h.service.Request(r.Context(), req.Email, ipHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(ChallengeAcceptedResponse{Message: challengeAcceptedMessage})
}

// ChallengeVerifyHandler decodifica, valida la forma e invoca
// auth.PhoneChallengeService.Verify.
type ChallengeVerifyHandler struct {
	service        *auth.PhoneChallengeService
	throttle       *auth.ThrottleService
	trustedProxies clientip.TrustedProxies
}

// NewChallengeVerifyHandler construye el handler de verificación del reto.
func NewChallengeVerifyHandler(service *auth.PhoneChallengeService, throttle *auth.ThrottleService, trustedProxies clientip.TrustedProxies) *ChallengeVerifyHandler {
	return &ChallengeVerifyHandler{service: service, throttle: throttle, trustedProxies: trustedProxies}
}

// ServeHTTP implementa http.Handler.
func (h *ChallengeVerifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	var req ChallengeVerifyRequest
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
	case !challengeCodePattern.MatchString(req.Code):
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Validation("code debe tener exactamente 6 dígitos"), requestID))
		return
	}

	rawIP, err := clientip.Resolve(r, h.trustedProxies)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(apperr.Internal(err), requestID))
		return
	}
	ipHash := h.throttle.HashIP(rawIP)

	if err := h.service.Verify(r.Context(), req.Email, ipHash, req.Code); err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
