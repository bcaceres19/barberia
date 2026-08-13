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
