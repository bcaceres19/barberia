package clientip_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"system-barbershop/internal/platform/clientip"
)

func mustTrusted(t *testing.T, cidrs ...string) clientip.TrustedProxies {
	t.Helper()
	tp, err := clientip.ParseTrustedProxies(cidrs)
	if err != nil {
		t.Fatalf("unexpected error parsing trusted proxies: %v", err)
	}
	return tp
}

func newRequest(remoteAddr string, headers map[string]string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.RemoteAddr = remoteAddr
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

func TestResolve_DirectConnectionNoProxy(t *testing.T) {
	r := newRequest("203.0.113.7:54321", nil)
	ip, err := clientip.Resolve(r, mustTrusted(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "203.0.113.7" {
		t.Fatalf("expected direct peer IP, got %q", ip)
	}
}

func TestResolve_XFFIgnoredWithoutTrustedProxies(t *testing.T) {
	r := newRequest("203.0.113.7:54321", map[string]string{"X-Forwarded-For": "198.51.100.1"})
	ip, err := clientip.Resolve(r, mustTrusted(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "203.0.113.7" {
		t.Fatalf("expected peer IP (XFF must be ignored with no trusted proxies), got %q", ip)
	}
}

func TestResolve_TrustedProxySingleHop(t *testing.T) {
	trusted := mustTrusted(t, "10.0.0.0/8")
	r := newRequest("10.1.2.3:443", map[string]string{"X-Forwarded-For": "198.51.100.1"})
	ip, err := clientip.Resolve(r, trusted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "198.51.100.1" {
		t.Fatalf("expected forwarded client IP, got %q", ip)
	}
}

func TestResolve_TrustedProxyMultiHop_WalksToFirstUntrustedFromRight(t *testing.T) {
	trusted := mustTrusted(t, "10.0.0.0/8")
	// Cadena: cliente real, proxy1 (confiable), proxy2 (confiable, es el
	// peer inmediato). El primer salto no confiable desde la derecha es el
	// cliente real.
	r := newRequest("10.2.2.2:443", map[string]string{"X-Forwarded-For": "198.51.100.1, 10.1.1.1"})
	ip, err := clientip.Resolve(r, trusted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "198.51.100.1" {
		t.Fatalf("expected the real client IP walking from the right, got %q", ip)
	}
}

func TestResolve_UntrustedPeerWithForgedHeader_UsesRemoteAddr(t *testing.T) {
	// El peer NO está en la lista de confiables: la cabecera, aunque
	// presente y con forma válida, se ignora por completo.
	trusted := mustTrusted(t, "10.0.0.0/8")
	r := newRequest("203.0.113.99:12345", map[string]string{"X-Forwarded-For": "1.2.3.4"})
	ip, err := clientip.Resolve(r, trusted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "203.0.113.99" {
		t.Fatalf("expected peer IP, spoofed header must not win: got %q", ip)
	}
}

func TestResolve_TrustedProxyEmptyXFF_UsesPeer(t *testing.T) {
	trusted := mustTrusted(t, "10.0.0.0/8")
	r := newRequest("10.1.2.3:443", nil)
	ip, err := clientip.Resolve(r, trusted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.1.2.3" {
		t.Fatalf("expected peer IP with no XFF header, got %q", ip)
	}
}

func TestResolve_TrustedProxyMalformedXFFEntry_FallsBackToPeer(t *testing.T) {
	trusted := mustTrusted(t, "10.0.0.0/8")
	r := newRequest("10.1.2.3:443", map[string]string{"X-Forwarded-For": "not-an-ip"})
	ip, err := clientip.Resolve(r, trusted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.1.2.3" {
		t.Fatalf("expected fallback to peer IP on malformed XFF, got %q", ip)
	}
}

func TestResolve_TrustedProxyEntireChainTrusted_FallsBackToPeer(t *testing.T) {
	trusted := mustTrusted(t, "10.0.0.0/8")
	r := newRequest("10.2.2.2:443", map[string]string{"X-Forwarded-For": "10.1.1.1"})
	ip, err := clientip.Resolve(r, trusted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.2.2.2" {
		t.Fatalf("expected peer fallback when entire chain is trusted, got %q", ip)
	}
}

func TestResolve_IPv6DirectConnection(t *testing.T) {
	r := newRequest("[2001:db8::1]:54321", nil)
	ip, err := clientip.Resolve(r, mustTrusted(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "2001:db8::1" {
		t.Fatalf("expected IPv6 peer, got %q", ip)
	}
}

func TestResolve_MalformedRemoteAddrErrors(t *testing.T) {
	r := newRequest("not-an-address", nil)
	if _, err := clientip.Resolve(r, mustTrusted(t)); err == nil {
		t.Fatal("expected error for malformed RemoteAddr, got nil")
	}
}

func TestParseTrustedProxies_InvalidCIDRErrors(t *testing.T) {
	if _, err := clientip.ParseTrustedProxies([]string{"not-a-cidr"}); err == nil {
		t.Fatal("expected error for invalid CIDR, got nil")
	}
}
