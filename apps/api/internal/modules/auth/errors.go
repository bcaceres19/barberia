package auth

import "system-barbershop/internal/platform/apperr"

// invalidCredentialsMessage es el ÚNICO texto que LoginService devuelve para
// cualquier fallo de autenticación: correo inexistente, contraseña
// incorrecta o usuario inactivo (CA-005-02, CA-005-07). Ninguna otra parte
// del paquete construye un mensaje distinto para estos casos.
const invalidCredentialsMessage = "correo o contraseña incorrectos"

// errInvalidCredentials es el error uniforme de credenciales inválidas.
func errInvalidCredentials() error {
	return apperr.Unauthorized(invalidCredentialsMessage)
}
