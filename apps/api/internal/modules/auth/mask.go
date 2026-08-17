package auth

import "strings"

// MaskPhone enmascara un teléfono E.164 (formato validado por
// staff_user_phone_ck: '+' seguido de 8 a 15 dígitos) conservando el
// prefijo y los dos últimos dígitos, nunca el valor completo (CA-008-06,
// DEC-065). Un valor que no cumple ese formato mínimo se enmascara por
// completo en vez de arriesgar una fuga parcial.
func MaskPhone(phone string) string {
	if len(phone) < 5 || phone[0] != '+' {
		return "***"
	}
	prefixLen := 3
	if len(phone) < prefixLen+2 {
		prefixLen = 1
	}
	return phone[:prefixLen] + " *** *** " + phone[len(phone)-2:]
}

// MaskEmail enmascara un correo conservando el primer carácter de la parte
// local y del dominio, nunca el valor completo (CA-008-06, DEC-065). Un
// valor sin '@' se enmascara por completo.
func MaskEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 || at == len(email)-1 {
		return "***"
	}
	local, domain := email[:at], email[at+1:]

	maskedLocal := local[:1] + "***"

	dot := strings.IndexByte(domain, '.')
	var maskedDomain string
	if dot > 0 {
		maskedDomain = domain[:1] + "***" + domain[dot:]
	} else {
		maskedDomain = domain[:1] + "***"
	}

	return maskedLocal + "@" + maskedDomain
}
