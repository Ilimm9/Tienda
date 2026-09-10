package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"tienda/backend/internal/application"
	"tienda/backend/internal/domain"
)

type PrecioCheckClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type precioCheckResponse struct {
	OK       bool `json:"ok"`
	Producto struct {
		CodigoBarras   string   `json:"codigo_barras"`
		Nombre         *string  `json:"nombre"`
		Descripcion    *string  `json:"descripcion"`
		Marca          *string  `json:"marca"`
		Categoria      *string  `json:"categoria"`
		PrecioSugerido *float64 `json:"precio_sugerido"`
		ContenidoNeto  *string  `json:"contenido_neto"`
		ImagenURL      *string  `json:"imagen_url"`
	} `json:"producto"`
}

var netContentPattern = regexp.MustCompile(`^\s*(\d+(?:[.,]\d+)?)\s*([[:alpha:]µμ]+)\s*$`)

func NewPrecioCheckClient(baseURL, apiKey string) *PrecioCheckClient {
	return &PrecioCheckClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  strings.TrimSpace(apiKey),
		client:  &http.Client{Timeout: 7 * time.Second},
	}
}

func (p *PrecioCheckClient) LookupProduct(barcode string) (domain.ProductLookup, error) {
	if p.apiKey == "" {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}

	endpoint := p.baseURL + "/producto/" + url.PathEscape(barcode)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return domain.ProductLookup{}, application.ErrProductNotFound
	case http.StatusTooManyRequests:
		return domain.ProductLookup{}, application.ErrLookupRateLimited
	case http.StatusUnauthorized:
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	case http.StatusOK:
	default:
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}

	var payload precioCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil || !payload.OK {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}

	result := domain.ProductLookup{
		CodigoBarras:   barcode,
		Nombre:         cleanString(payload.Producto.Nombre),
		Descripcion:    cleanString(payload.Producto.Descripcion),
		Marca:          cleanString(payload.Producto.Marca),
		Categoria:      cleanString(payload.Producto.Categoria),
		PrecioSugerido: payload.Producto.PrecioSugerido,
		Fuentes:        []string{"PrecioCheck"},
	}
	result.Contenido, result.UnidadContenido = parseNetContent(payload.Producto.ContenidoNeto)
	if payload.Producto.ImagenURL == nil || strings.TrimSpace(*payload.Producto.ImagenURL) == "" {
		return result, nil
	}
	imageURL := strings.TrimSpace(*payload.Producto.ImagenURL)
	parsed, err := url.Parse(imageURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return domain.ProductLookup{}, fmt.Errorf("%w: URL de imagen inválida", application.ErrLookupUnavailable)
	}
	result.ImagenURL = &imageURL
	return result, nil
}

func cleanString(value *string) *string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	cleaned := strings.TrimSpace(*value)
	return &cleaned
}

func parseNetContent(value *string) (*float64, *string) {
	cleaned := cleanString(value)
	if cleaned == nil {
		return nil, nil
	}
	matches := netContentPattern.FindStringSubmatch(*cleaned)
	if len(matches) != 3 {
		return nil, nil
	}
	amount, err := strconv.ParseFloat(strings.Replace(matches[1], ",", ".", 1), 64)
	if err != nil {
		return nil, nil
	}
	unit := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(matches[2], "µ", "u"), "μ", "u"))
	return &amount, &unit
}

var _ application.ProductLookupProvider = (*PrecioCheckClient)(nil)
