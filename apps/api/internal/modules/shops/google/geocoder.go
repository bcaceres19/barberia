// Package google es el adaptador de shops hacia Google Geocoding (DEC-130). Es
// el ÚNICO paquete de shops que habla con Google: el núcleo y las pruebas del
// servicio usan el puerto shops.Geocoder.
package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"system-barbershop/internal/modules/shops"
)

// DefaultGeocodeURL es el endpoint de Google Geocoding.
const DefaultGeocodeURL = "https://maps.googleapis.com/maps/api/geocode/json"

// maxResponseBytes acota la lectura de la respuesta: una respuesta de
// geocodificación pesa unos pocos KiB.
const maxResponseBytes = 1 << 20

// Config es la configuración del adaptador.
type Config struct {
	// APIKey es la clave de SERVIDOR de Google Maps Platform (DEC-129/DEC-130):
	// nunca se envía al navegador. Vacía deja el adaptador sin poder consultar.
	APIKey string
	// URL existe para apuntar a un Google falso en pruebas; vacía usa la real.
	URL string
	// HTTPClient permite inyectar transporte y tiempo de espera; nil usa uno
	// con timeout de 8 s.
	HTTPClient *http.Client
	// Logger registra la causa real de un fallo (nunca la dirección ni la
	// clave: RN-DAT-02). nil descarta los registros.
	Logger *slog.Logger
}

// Geocoder implementa shops.Geocoder sobre la API de Geocoding de Google.
type Geocoder struct {
	apiKey string
	url    string
	http   *http.Client
	logger *slog.Logger
}

var _ shops.Geocoder = (*Geocoder)(nil)

// New construye el adaptador.
func New(cfg Config) *Geocoder {
	g := &Geocoder{apiKey: cfg.APIKey, url: cfg.URL, http: cfg.HTTPClient, logger: cfg.Logger}
	if g.url == "" {
		g.url = DefaultGeocodeURL
	}
	if g.http == nil {
		g.http = &http.Client{Timeout: 8 * time.Second}
	}
	if g.logger == nil {
		g.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return g
}

type geocodeResponse struct {
	Status  string `json:"status"`
	Results []struct {
		PartialMatch bool `json:"partial_match"`
		Geometry     struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
		} `json:"geometry"`
	} `json:"results"`
}

// Geocode implementa shops.Geocoder.Geocode.
//
// Un resultado se acepta solo si Google lo marca como coincidencia completa y
// lo ubica con precisión de calle o mejor: un `partial_match` o un punto
// `APPROXIMATE` (el centro de una ciudad o un barrio) pondría el marcador en un
// lugar que no es la barbería, y mostrar un punto erróneo es peor que no
// mostrar el mapa (riesgo de HU-102). Esos casos son ErrAddressUnresolvable; un
// fallo de red, de cuota o de credencial es ErrGeocoderUnavailable.
func (g *Geocoder) Geocode(ctx context.Context, address string) (shops.Coordinates, error) {
	if g.apiKey == "" {
		g.logger.Warn("shops.geocode.sin_clave")
		return shops.Coordinates{}, shops.ErrGeocoderUnavailable
	}

	query := url.Values{}
	query.Set("address", address)
	query.Set("key", g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.url+"?"+query.Encode(), nil)
	if err != nil {
		return shops.Coordinates{}, fmt.Errorf("shops/google: armar solicitud: %w", err)
	}

	resp, err := g.http.Do(req)
	if err != nil {
		// El error de net/http incluye la URL completa (con la clave): no se
		// registra tal cual, solo su tipo.
		g.logger.Warn("shops.geocode.red", "error_kind", classifyTransportError(err))
		return shops.Coordinates{}, shops.ErrGeocoderUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		g.logger.Warn("shops.geocode.http", "status", resp.StatusCode)
		return shops.Coordinates{}, shops.ErrGeocoderUnavailable
	}

	var body geocodeResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&body); err != nil {
		g.logger.Warn("shops.geocode.cuerpo_invalido")
		return shops.Coordinates{}, shops.ErrGeocoderUnavailable
	}

	switch body.Status {
	case "OK":
	case "ZERO_RESULTS":
		return shops.Coordinates{}, shops.ErrAddressUnresolvable
	default:
		// REQUEST_DENIED, OVER_QUERY_LIMIT, OVER_DAILY_LIMIT, INVALID_REQUEST,
		// UNKNOWN_ERROR: no dependen de la dirección.
		g.logger.Warn("shops.geocode.estado", "status", body.Status)
		return shops.Coordinates{}, shops.ErrGeocoderUnavailable
	}

	if len(body.Results) == 0 {
		return shops.Coordinates{}, shops.ErrAddressUnresolvable
	}
	best := body.Results[0]
	if best.PartialMatch || best.Geometry.LocationType == "APPROXIMATE" {
		return shops.Coordinates{}, shops.ErrAddressUnresolvable
	}
	return shops.Coordinates{
		Latitude:  best.Geometry.Location.Lat,
		Longitude: best.Geometry.Location.Lng,
	}, nil
}

func classifyTransportError(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "cancelado"
	default:
		return "transporte"
	}
}
