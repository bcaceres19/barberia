package google_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"system-barbershop/internal/modules/shops"
	"system-barbershop/internal/modules/shops/google"
)

func serve(t *testing.T, status int, body string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Query().Get("key") != "server-key" {
			t.Errorf("la consulta debe llevar la clave de servidor, llevó %q", r.URL.Query().Get("key"))
		}
		if r.URL.Query().Get("address") == "" {
			t.Errorf("la consulta debe llevar la dirección")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestGeocode(t *testing.T) {
	const ok = `{"status":"OK","results":[{"partial_match":false,"geometry":{"location":{"lat":4.598056,"lng":-74.076111},"location_type":"ROOFTOP"}}]}`
	tests := []struct {
		name    string
		status  int
		body    string
		want    shops.Coordinates
		wantErr error
	}{
		{name: "éxito", status: 200, body: ok, want: shops.Coordinates{Latitude: 4.598056, Longitude: -74.076111}},
		{name: "interpolada", status: 200,
			body: `{"status":"OK","results":[{"geometry":{"location":{"lat":1,"lng":2},"location_type":"RANGE_INTERPOLATED"}}]}`,
			want: shops.Coordinates{Latitude: 1, Longitude: 2}},
		{name: "no encontrada", status: 200, body: `{"status":"ZERO_RESULTS","results":[]}`, wantErr: shops.ErrAddressUnresolvable},
		{name: "ambigua (coincidencia parcial)", status: 200,
			body:    `{"status":"OK","results":[{"partial_match":true,"geometry":{"location":{"lat":1,"lng":2},"location_type":"ROOFTOP"}}]}`,
			wantErr: shops.ErrAddressUnresolvable},
		{name: "solo a nivel de ciudad", status: 200,
			body:    `{"status":"OK","results":[{"geometry":{"location":{"lat":1,"lng":2},"location_type":"APPROXIMATE"}}]}`,
			wantErr: shops.ErrAddressUnresolvable},
		{name: "OK sin resultados", status: 200, body: `{"status":"OK","results":[]}`, wantErr: shops.ErrAddressUnresolvable},
		{name: "cuota agotada", status: 200, body: `{"status":"OVER_QUERY_LIMIT"}`, wantErr: shops.ErrGeocoderUnavailable},
		{name: "credencial denegada", status: 200, body: `{"status":"REQUEST_DENIED"}`, wantErr: shops.ErrGeocoderUnavailable},
		{name: "http 500", status: 500, body: `oops`, wantErr: shops.ErrGeocoderUnavailable},
		{name: "cuerpo inválido", status: 200, body: `{no`, wantErr: shops.ErrGeocoderUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := serve(t, tt.status, tt.body)
			g := google.New(google.Config{APIKey: "server-key", URL: srv.URL})
			got, err := g.Geocode(context.Background(), "Calle 10 # 5-20, Bogotá")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, quería %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got != tt.want {
				t.Fatalf("coordenadas = %+v, quería %+v", got, tt.want)
			}
		})
	}
}

func TestGeocodeWithoutKeyNeverCallsGoogle(t *testing.T) {
	srv, calls := serve(t, 200, `{"status":"OK"}`)
	g := google.New(google.Config{URL: srv.URL})
	if _, err := g.Geocode(context.Background(), "Calle 10"); !errors.Is(err, shops.ErrGeocoderUnavailable) {
		t.Fatalf("sin clave debe ser ErrGeocoderUnavailable, fue %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("sin clave no debe haber ninguna llamada de red, hubo %d", calls.Load())
	}
}

func TestGeocodeUnreachableNeverLogsKeyOrAddress(t *testing.T) {
	srv, _ := serve(t, 200, `{}`)
	url := srv.URL
	srv.Close() // Nadie escucha: falla el transporte.

	var logs bytes.Buffer
	g := google.New(google.Config{
		APIKey: "server-key", URL: url,
		Logger: slog.New(slog.NewTextHandler(&logs, nil)),
	})
	if _, err := g.Geocode(context.Background(), "Calle Secreta 123"); !errors.Is(err, shops.ErrGeocoderUnavailable) {
		t.Fatalf("un fallo de red debe ser ErrGeocoderUnavailable, fue %v", err)
	}
	if out := logs.String(); strings.Contains(out, "server-key") || strings.Contains(out, "Secreta") {
		t.Fatalf("el log no debe contener la clave ni la dirección: %s", out)
	}
}

func TestSimulatedIsDeterministicAndRejectsShortText(t *testing.T) {
	var s google.Simulated
	a, err := s.Geocode(context.Background(), "Calle 10 # 5-20")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := s.Geocode(context.Background(), "  calle 10 # 5-20 ")
	if a != b {
		t.Fatalf("la misma dirección debe dar el mismo punto: %+v vs %+v", a, b)
	}
	if a.Latitude < 4.5 || a.Latitude > 4.7 || a.Longitude > -74.0 || a.Longitude < -74.2 {
		t.Fatalf("el punto simulado debe caer cerca de Bogotá: %+v", a)
	}
	if _, err := s.Geocode(context.Background(), "ab"); !errors.Is(err, shops.ErrAddressUnresolvable) {
		t.Fatalf("un texto muy corto debe ser irresoluble, fue %v", err)
	}
}
