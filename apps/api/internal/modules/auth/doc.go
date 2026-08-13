// Package auth administra identidad, sesiones y recuperación de acceso del
// barbero, según docs/03-desarrollo/estandar-backend-go.md.
//
// HU-005 entrega el inicio de sesión: [LoginService] verifica correo y
// contraseña (hash argon2id, ver [Argon2Hasher]) sin distinguir un correo
// inexistente de una contraseña incorrecta (CA-005-02), y emite un token de
// sesión opaco (ver [CryptoTokenGenerator]) que el adaptador HTTP
// (httpapi/) entrega como cookie. La persistencia vive en postgres/, detrás
// del puerto [Repository] declarado en ports.go: el núcleo de este paquete
// no importa Chi, net/http, pgx ni internal/platform/database
// (CA-002-06), para que domain.go, errors.go, ports.go y service.go se
// puedan probar sin PostgreSQL real.
//
// Cierre de sesión, renovación deslizante completa, límite por IP y
// recuperación de acceso llegan con HU-006, HU-007 y HU-008
// respectivamente; no se adelantan aquí.
package auth
