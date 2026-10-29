// Package google es el adaptador de publicbooking hacia Google Places (New)
// para las reseñas de la ficha del negocio (DEC-129, HU-101). Es el ÚNICO
// paquete de publicbooking que habla con Google: el núcleo y las pruebas del
// servicio usan el puerto publicbooking.GoogleReviewsProvider.
//
// Términos de Google Maps Platform: el resultado vive solo en memoria durante
// 6 horas (CacheTTL) y nunca se persiste. La clave es de SERVIDOR: viaja en una
// cabecera hacia Google y no se registra ni se devuelve al navegador (CA-101-04).
package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"system-barbershop/internal/modules/publicbooking"
)

// DefaultBaseURL es la API de Google Places (New).
const DefaultBaseURL = "https://places.googleapis.com/v1"

const (
	// DefaultCacheTTL es la ventana de caché de una consulta exitosa (DEC-129).
	DefaultCacheTTL = 6 * time.Hour
	// DefaultFailureTTL es cuánto se recuerda un fallo para no golpear a Google
	// en cada visita mientras la cuota o la ficha siguen mal: sin reintento
	// síncrono (mismo criterio que DEC-066).
	DefaultFailureTTL = time.Minute
	// DefaultTimeout es el tiempo máximo de una consulta: la reserva pública no
	// debe esperar a Google más que esto.
	DefaultTimeout = 3 * time.Second

	maxResponseBytes = 1 << 20
	maxCacheEntries  = 1000
	maxAuthorName    = 200
	maxReviewText    = 2000
	maxRelativeTime  = 100
	maxURL           = 500
	fieldMask        = "rating,userRatingCount,googleMapsUri,reviews"
)

// Config es la configuración del adaptador.
type Config struct {
	// APIKey es la clave de SERVIDOR de Google Maps Platform con Places API (New)
	// habilitada. Nunca llega al navegador ni se registra.
	APIKey string
	// BaseURL existe para apuntar a un Google falso en pruebas; vacía usa la real.
	BaseURL string
	// HTTPClient permite inyectar transporte; nil usa uno con DefaultTimeout.
	HTTPClient *http.Client
	// CacheTTL y FailureTTL; cero usa los valores por defecto.
	CacheTTL   time.Duration
	FailureTTL time.Duration
	// Now permite fijar el reloj en pruebas; nil usa time.Now.
	Now func() time.Time
}

// Client implementa publicbooking.GoogleReviewsProvider con caché en memoria.
type Client struct {
	apiKey     string
	baseURL    string
	http       *http.Client
	cacheTTL   time.Duration
	failureTTL time.Duration
	now        func() time.Time

	mu       sync.Mutex
	cache    map[string]cacheEntry
	inflight map[string]*call
}

type cacheEntry struct {
	reviews publicbooking.GoogleReviews
	err     error
	expires time.Time
}

// call comparte una consulta en vuelo entre visitas simultáneas de la misma
// ficha: una sola llamada a Google, no una por visitante.
type call struct {
	done    chan struct{}
	reviews publicbooking.GoogleReviews
	err     error
}

var _ publicbooking.GoogleReviewsProvider = (*Client)(nil)

// New construye el adaptador.
func New(cfg Config) *Client {
	c := &Client{
		apiKey: cfg.APIKey, baseURL: strings.TrimRight(cfg.BaseURL, "/"), http: cfg.HTTPClient,
		cacheTTL: cfg.CacheTTL, failureTTL: cfg.FailureTTL, now: cfg.Now,
		cache: map[string]cacheEntry{}, inflight: map[string]*call{},
	}
	if c.baseURL == "" {
		c.baseURL = DefaultBaseURL
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: DefaultTimeout}
	}
	if c.cacheTTL <= 0 {
		c.cacheTTL = DefaultCacheTTL
	}
	if c.failureTTL <= 0 {
		c.failureTTL = DefaultFailureTTL
	}
	if c.now == nil {
		c.now = time.Now
	}
	return c
}

// Fetch implementa publicbooking.GoogleReviewsProvider. Una segunda consulta
// dentro de la ventana de caché no vuelve a llamar a Google (CA-101-05).
func (c *Client) Fetch(ctx context.Context, placeID string) (publicbooking.GoogleReviews, error) {
	c.mu.Lock()
	if e, ok := c.cache[placeID]; ok && c.now().Before(e.expires) {
		c.mu.Unlock()
		return e.reviews, e.err
	}
	if in, ok := c.inflight[placeID]; ok {
		c.mu.Unlock()
		select {
		case <-in.done:
			return in.reviews, in.err
		case <-ctx.Done():
			return publicbooking.GoogleReviews{}, ctx.Err()
		}
	}
	in := &call{done: make(chan struct{})}
	c.inflight[placeID] = in
	c.mu.Unlock()

	reviews, err := c.fetch(ctx, placeID)

	c.mu.Lock()
	// Una consulta cancelada por quien la pidió no dice nada de Google: no se
	// recuerda como fallo de la ficha.
	if !errors.Is(err, context.Canceled) {
		ttl := c.cacheTTL
		if err != nil {
			ttl = c.failureTTL
		}
		c.store(placeID, cacheEntry{reviews: reviews, err: err, expires: c.now().Add(ttl)})
	}
	delete(c.inflight, placeID)
	in.reviews, in.err = reviews, err
	close(in.done)
	c.mu.Unlock()
	return reviews, err
}

// store guarda una entrada acotando el tamaño: los identificadores salen de la
// base (uno por barbería), pero la memoria nunca debe crecer sin límite.
func (c *Client) store(key string, e cacheEntry) {
	if len(c.cache) >= maxCacheEntries {
		now := c.now()
		for k, v := range c.cache {
			if !now.Before(v.expires) {
				delete(c.cache, k)
			}
		}
		if len(c.cache) >= maxCacheEntries {
			clear(c.cache)
		}
	}
	c.cache[key] = e
}

type placeResponse struct {
	Rating          *float64 `json:"rating"`
	UserRatingCount int      `json:"userRatingCount"`
	GoogleMapsURI   string   `json:"googleMapsUri"`
	Reviews         []struct {
		Rating                         int    `json:"rating"`
		RelativePublishTimeDescription string `json:"relativePublishTimeDescription"`
		Text                           *struct {
			Text string `json:"text"`
		} `json:"text"`
		AuthorAttribution struct {
			DisplayName string `json:"displayName"`
			URI         string `json:"uri"`
		} `json:"authorAttribution"`
	} `json:"reviews"`
}

func (c *Client) fetch(ctx context.Context, placeID string) (publicbooking.GoogleReviews, error) {
	if c.apiKey == "" {
		return publicbooking.GoogleReviews{}, fmt.Errorf("%w: sin credencial de servidor", publicbooking.ErrGoogleReviewsUnavailable)
	}
	// placeID ya viene validado (letras, dígitos, `_` y `-`), pero se escapa igual:
	// el adaptador no confía en quien lo llama para no construir una ruta ajena.
	endpoint := c.baseURL + "/places/" + url.PathEscape(placeID) + "?languageCode=es"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return publicbooking.GoogleReviews{}, fmt.Errorf("%w: armar la petición", publicbooking.ErrGoogleReviewsUnavailable)
	}
	req.Header.Set("X-Goog-Api-Key", c.apiKey)
	req.Header.Set("X-Goog-FieldMask", fieldMask)

	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return publicbooking.GoogleReviews{}, ctx.Err()
		}
		// *url.Error nunca lleva la clave (viaja en cabecera), pero se envuelve
		// sin el texto de la URL para no filtrar nada más del destino.
		return publicbooking.GoogleReviews{}, fmt.Errorf("%w: red", publicbooking.ErrGoogleReviewsUnavailable)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseBytes))
		return publicbooking.GoogleReviews{}, fmt.Errorf("%w: Google respondió %d", publicbooking.ErrGoogleReviewsUnavailable, resp.StatusCode)
	}
	var body placeResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(&body); err != nil {
		return publicbooking.GoogleReviews{}, fmt.Errorf("%w: respuesta ilegible", publicbooking.ErrGoogleReviewsUnavailable)
	}
	return toReviews(body), nil
}

// toReviews valida y recorta lo que devuelve Google: lo que no cumple el
// contrato (calificación fuera de 1–5, enlaces que no son https, textos largos)
// se descarta o se acota, nunca llega al navegador tal cual.
func toReviews(body placeResponse) publicbooking.GoogleReviews {
	out := publicbooking.GoogleReviews{Reviews: []publicbooking.GoogleReview{}}
	if body.Rating == nil || *body.Rating < 1 || *body.Rating > 5 || body.UserRatingCount < 0 {
		// Una ficha sin calificaciones todavía no tiene nada que mostrar.
		return out
	}
	rounded := float64(int(*body.Rating*10+0.5)) / 10
	out.Rating = &rounded
	out.RatingCount = body.UserRatingCount
	out.MapsURL = httpsURL(body.GoogleMapsURI)

	for _, r := range body.Reviews {
		if len(out.Reviews) == publicbooking.MaxGoogleReviews {
			break
		}
		name := truncate(strings.Join(strings.Fields(r.AuthorAttribution.DisplayName), " "), maxAuthorName)
		if name == "" || r.Rating < 1 || r.Rating > 5 {
			continue
		}
		review := publicbooking.GoogleReview{
			AuthorName: name, AuthorURL: httpsURL(r.AuthorAttribution.URI), Rating: r.Rating,
			RelativeTime: optional(truncate(strings.TrimSpace(r.RelativePublishTimeDescription), maxRelativeTime)),
		}
		if r.Text != nil {
			review.Text = optional(truncate(strings.TrimSpace(r.Text.Text), maxReviewText))
		}
		out.Reviews = append(out.Reviews, review)
	}
	return out
}

// httpsURL devuelve raw solo si es una URL https absoluta y corta: es el destino
// de un enlace, y un esquema como javascript: nunca debe llegar al navegador.
func httpsURL(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > maxURL {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil
	}
	return &raw
}

func truncate(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
