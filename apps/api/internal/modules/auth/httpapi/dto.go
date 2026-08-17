package httpapi

import "time"

// LoginRequest es el cuerpo de POST /api/v1/public/auth/login. Cerrado: el
// handler rechaza cualquier campo desconocido (json.Decoder con
// DisallowUnknownFields), según docs/06-api/estandar-openapi.md sección 9.4.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse es el cuerpo de una respuesta 200 exitosa. No incluye
// ningún secreto: el token de sesión viaja únicamente en la cabecera
// Set-Cookie, nunca en el body (CA-005-04).
type LoginResponse struct {
	ExpiresAt time.Time `json:"expiresAt"`
}

// BarbershopSummary es el payload mínimo de barbería activa de HU-012
// (DEC-060). Nunca incluye zona horaria, contacto ni ningún otro campo de
// HU-020: esos llegan con esa historia, no se adelantan aquí.
type BarbershopSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SessionContextResponse es el cuerpo de la respuesta 200 de
// GET /private/auth/session (HU-012, DEC-060). Payload mínimo: nunca
// StaffUserID, correo ni nombre del barbero (RN-DAT-02).
type SessionContextResponse struct {
	Barbershop BarbershopSummary `json:"barbershop"`
	ExpiresAt  time.Time         `json:"expiresAt"`
}

// ChallengeRequest es el cuerpo de POST /api/v1/public/auth/challenge
// (HU-007, DEC-062).
type ChallengeRequest struct {
	Email string `json:"email"`
}

// ChallengeAcceptedResponse es el cuerpo de la respuesta 202, siempre el
// mismo mensaje fijo (no enumeración, DEC-062).
type ChallengeAcceptedResponse struct {
	Message string `json:"message"`
}

// ChallengeVerifyRequest es el cuerpo de
// POST /api/v1/public/auth/challenge/verify (HU-007, DEC-062).
type ChallengeVerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
