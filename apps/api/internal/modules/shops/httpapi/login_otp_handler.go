package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"system-barbershop/internal/modules/auth"
	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/httpserver"
)

// LoginOTPResponse es la representación de GET y PATCH
// /private/settings/login-otp (DEC-139): solo el canal.
type LoginOTPResponse struct {
	Channel string `json:"channel"`
}

// UpdateLoginOTPRequest es el cuerpo de PATCH /private/settings/login-otp. Cerrado:
// cualquier campo desconocido se rechaza, en particular un destino (correo o
// teléfono), que nunca lo fija el cliente (DEC-081).
type UpdateLoginOTPRequest struct {
	Channel string `json:"channel"`
}

// GetLoginOTPHandler expone GET /private/settings/login-otp.
type GetLoginOTPHandler struct {
	service *shops.LoginOTPService
}

// NewGetLoginOTPHandler construye el handler de lectura.
func NewGetLoginOTPHandler(service *shops.LoginOTPService) *GetLoginOTPHandler {
	return &GetLoginOTPHandler{service: service}
}

// ServeHTTP implementa http.Handler. La barbería sale exclusivamente de
// auth.Principal.
func (h *GetLoginOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}
	channel, err := h.service.Get(r.Context(), principal.BarbershopID)
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeLoginOTP(w, channel)
}

// UpdateLoginOTPHandler expone PATCH /private/settings/login-otp.
type UpdateLoginOTPHandler struct {
	service *shops.LoginOTPService
}

// NewUpdateLoginOTPHandler construye el handler de actualización.
func NewUpdateLoginOTPHandler(service *shops.LoginOTPService) *UpdateLoginOTPHandler {
	return &UpdateLoginOTPHandler{service: service}
}

// ServeHTTP implementa http.Handler.
func (h *UpdateLoginOTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())
	principal, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		httpserver.WriteProblem(w, httpserver.Translate(
			apperr.Internal(errors.New("shops: falta el principal de sesión en el contexto")), requestID))
		return
	}

	var req UpdateLoginOTPRequest
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

	channel, err := h.service.Update(r.Context(), principal.BarbershopID, shops.LoginOTPChannel(req.Channel))
	if err != nil {
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}
	writeLoginOTP(w, channel)
}

func writeLoginOTP(w http.ResponseWriter, channel shops.LoginOTPChannel) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginOTPResponse{Channel: string(channel)})
}
