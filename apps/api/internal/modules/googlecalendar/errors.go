package googlecalendar

import "errors"

// ErrTokenRevoked indica que Google rechazó el refresh token de forma
// permanente (invalid_grant): hay que reconectar. El adaptador de Google lo
// devuelve y el servicio lo traduce al estado `reauth_required`.
var ErrTokenRevoked = errors.New("googlecalendar: el permiso de Google fue revocado o caducó")
