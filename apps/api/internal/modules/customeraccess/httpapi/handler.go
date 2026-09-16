// Package httpapi es el adaptador HTTP del módulo customeraccess: decodifica,
// invoca customeraccess.Service y traduce el resultado a HTTP
// (docs/03-desarrollo/estandar-backend-go.md §4). No importa Chi: la
// composición de rutas es responsabilidad exclusiva de cmd/api sobre el
// router de audiencia customer que internal/platform/httpserver ya
// construye.
package httpapi

import (
	"encoding/json"
	"net/http"

	"system-barbershop/internal/modules/customeraccess"
	"system-barbershop/internal/platform/httpserver"
)

// tokenParam es el nombre del parámetro de ruta que cmd/api.buildRouter
// registra para /customer/appointments/{token} (HU-098).
const tokenParam = "token"

// GetAppointmentHandler expone GET /customer/appointments/{token}
// (CA-098-01 a CA-098-05): lectura mínima del turno del cliente autenticada
// únicamente por la credencial del enlace, sin sesión.
type GetAppointmentHandler struct {
	service *customeraccess.Service
}

// NewGetAppointmentHandler construye el handler de lectura del cliente.
func NewGetAppointmentHandler(service *customeraccess.Service) *GetAppointmentHandler {
	return &GetAppointmentHandler{service: service}
}

// ServeHTTP implementa http.Handler. Deliberadamente NO lee ningún
// principal de sesión: el único dato de entrada es el token de la ruta, y
// el servicio lo hashea y resuelve el tenant por sí mismo antes de leer
// nada (mismo criterio que publicbooking.ResolveBarbershopHandler frente a
// slug). Cache-Control: no-store evita que un caché compartido o del
// navegador retenga el turno de otra persona que use el mismo equipo
// (RN-DAT-02, "políticas de caché" del alcance de HU-098); el resto de
// cabeceras de seguridad ya las fija httpserver.SecurityHeaders para toda
// respuesta, incluida Referrer-Policy: no-referrer.
func (h *GetAppointmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := httpserver.RequestIDFromContext(r.Context())

	token := httpserver.URLParam(r, tokenParam)
	view, err := h.service.GetAppointmentByToken(r.Context(), token)
	if err != nil {
		w.Header().Set("Cache-Control", "no-store")
		httpserver.WriteProblem(w, httpserver.Translate(err, requestID))
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(newCustomerAppointmentResponse(view))
}

func newCustomerAppointmentResponse(v customeraccess.AppointmentView) CustomerAppointmentResponse {
	return CustomerAppointmentResponse{
		BarbershopName:                 v.BarbershopName,
		Timezone:                       v.Timezone,
		AttendeeName:                   v.AttendeeName,
		ServiceName:                    v.ServiceName,
		DurationMinutes:                v.DurationMinutes,
		BarberName:                     v.BarberName,
		StartsAt:                       v.StartsAt,
		EndsAt:                         v.EndsAt,
		Status:                         v.Status,
		CancellationDeadlineMinutes:    v.CancellationDeadlineMinutes,
		LateCancellationClientAllowed:  v.LateCancellationClientAllowed,
		LateCancellationReasonRequired: v.LateCancellationReasonRequired,
	}
}
