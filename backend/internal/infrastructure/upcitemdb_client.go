package infrastructure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"tienda/backend/internal/application"
	"tienda/backend/internal/domain"
)

type UPCItemDBClient struct {
	baseURL string
	client  *http.Client
}

type upcItemDBResponse struct {
	Total int `json:"total"`
	Items []struct {
		Title       *string  `json:"title"`
		Description *string  `json:"description"`
		Brand       *string  `json:"brand"`
		Category    *string  `json:"category"`
		Size        *string  `json:"size"`
		Weight      *string  `json:"weight"`
		Images      []string `json:"images"`
	} `json:"items"`
}

func NewUPCItemDBClient(baseURL string) *UPCItemDBClient {
	return &UPCItemDBClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 7 * time.Second},
	}
}

func (u *UPCItemDBClient) LookupProduct(barcode string) (domain.ProductLookup, error) {
	endpoint := u.baseURL + "/lookup?upc=" + url.QueryEscape(barcode)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}
	req.Header.Set("Accept", "application/json")

	resp, err := u.client.Do(req)
	if err != nil {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return domain.ProductLookup{}, application.ErrProductNotFound
	case http.StatusTooManyRequests:
		return domain.ProductLookup{}, application.ErrLookupRateLimited
	case http.StatusOK:
	default:
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}

	var payload upcItemDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.ProductLookup{}, application.ErrLookupUnavailable
	}
	if payload.Total < 1 || len(payload.Items) == 0 {
		return domain.ProductLookup{}, application.ErrProductNotFound
	}

	item := payload.Items[0]
	result := domain.ProductLookup{
		CodigoBarras: barcode,
		Nombre:       cleanString(item.Title),
		Descripcion:  cleanString(item.Description),
		Marca:        cleanString(item.Brand),
		Categoria:    cleanString(item.Category),
		Fuentes:      []string{"UPCitemdb"},
	}
	result.Contenido, result.UnidadContenido = parseNetContent(item.Size)
	if result.Contenido == nil {
		result.Contenido, result.UnidadContenido = parseNetContent(item.Weight)
	}
	result.ImagenURL = firstHTTPSImage(item.Images)
	return result, nil
}

func firstHTTPSImage(images []string) *string {
	for _, candidate := range images {
		candidate = strings.TrimSpace(candidate)
		parsed, err := url.Parse(candidate)
		if err == nil && parsed.Scheme == "https" && parsed.Host != "" {
			return &candidate
		}
	}
	return nil
}

var _ application.ProductLookupProvider = (*UPCItemDBClient)(nil)
