// Package clientip resuelve la IP real de una solicitud HTTP (HU-007). Vive
// en platform, no en el módulo auth, porque resolver una IP es un concern de
// transporte reutilizable por cualquier límite futuro, no una regla de
// negocio de autenticación (docs/04-arquitectura/backend-go.md).
package clientip

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

// TrustedProxies es la lista cerrada de redes cuyo peer inmediato puede
// declarar la IP real del cliente mediante X-Forwarded-For. Vacía por
// defecto: sin proxies confiables, Resolve ignora esa cabecera por completo,
// sin importar su contenido.
type TrustedProxies struct {
	nets []*net.IPNet
}

// ParseTrustedProxies valida cada CIDR de cidrs. Un valor mal formado es un
// error de configuración de despliegue, no un valor que deba descartarse en
// silencio.
func ParseTrustedProxies(cidrs []string) (TrustedProxies, error) {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, ipNet, err := net.ParseCIDR(c)
		if err != nil {
			return TrustedProxies{}, fmt.Errorf("clientip: CIDR inválido %q: %w", c, err)
		}
		nets = append(nets, ipNet)
	}
	return TrustedProxies{nets: nets}, nil
}

func (t TrustedProxies) contains(ip net.IP) bool {
	for _, n := range t.nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// Resolve determina la IP del cliente real. Por defecto usa r.RemoteAddr (el
// peer TCP inmediato, imposible de falsificar). Solo si ese peer pertenece a
// trusted acepta X-Forwarded-For, y en ese caso toma el valor MÁS A LA
// DERECHA de la lista que no pertenezca a trusted: en un despliegue con N
// proxies encadenados y confiables, cada uno antepone su propio peer a la
// cabecera: el primer salto no confiable, empezando desde el final, es el
// cliente real. Un peer no confiable con una cabecera falsificada nunca
// llega a leerse: RemoteAddr manda.
func Resolve(r *http.Request, trusted TrustedProxies) (string, error) {
	peerHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr sin puerto (poco común, pero net/http no lo garantiza
		// en todas las rutas de prueba): se intenta como IP desnuda antes
		// de fallar.
		peerHost = r.RemoteAddr
	}
	peerIP := net.ParseIP(peerHost)
	if peerIP == nil {
		return "", fmt.Errorf("clientip: RemoteAddr no contiene una IP válida: %q", r.RemoteAddr)
	}

	if len(trusted.nets) == 0 || !trusted.contains(peerIP) {
		return peerIP.String(), nil
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff == "" {
		return peerIP.String(), nil
	}

	parts := strings.Split(xff, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := net.ParseIP(strings.TrimSpace(parts[i]))
		if candidate == nil {
			// Un valor no parseable en la cadena reenviada no es una IP de
			// un proxy confiable: se detiene la caminata y se usa el peer
			// inmediato, nunca un valor que no se pudo validar.
			return peerIP.String(), nil
		}
		if !trusted.contains(candidate) {
			return candidate.String(), nil
		}
	}

	// Toda la cadena reenviada pertenece a proxies confiables (sin un salto
	// de cliente real visible): se usa el peer inmediato como última
	// defensa, nunca un valor vacío.
	return peerIP.String(), nil
}
