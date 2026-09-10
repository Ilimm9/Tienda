package infrastructure

import (
	"errors"
	"net/http"
	"testing"

	"tienda/backend/internal/application"
)

func TestUPCItemDBClientLookupProduct(t *testing.T) {
	client := NewUPCItemDBClient("https://upcitemdb.test/prod/trial")
	client.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/prod/trial/lookup" || r.URL.Query().Get("upc") != "7501055303038" {
			t.Fatalf("solicitud inesperada: %s", r.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"total":1,"items":[{"title":"Producto UPC","description":"Descripción","brand":"Marca","category":"Food > Drinks","size":"355 ml","images":["http://insegura.test/a.jpg","https://cdn.test/producto.jpg"],"lowest_recorded_price":2.5,"currency":"USD"}]}`), nil
	})}

	result, err := client.LookupProduct("7501055303038")
	if err != nil {
		t.Fatalf("LookupProduct devolvió error: %v", err)
	}
	if result.Nombre == nil || *result.Nombre != "Producto UPC" || result.ImagenURL == nil || *result.ImagenURL != "https://cdn.test/producto.jpg" {
		t.Fatalf("resultado inesperado: %+v", result)
	}
	if result.Contenido == nil || *result.Contenido != 355 || result.PrecioSugerido != nil {
		t.Fatalf("contenido o precio inesperado: %+v", result)
	}
	if len(result.Fuentes) != 1 || result.Fuentes[0] != "UPCitemdb" {
		t.Fatalf("fuentes inesperadas: %v", result.Fuentes)
	}
}

func TestUPCItemDBClientMapsMissingAndLimitedResponses(t *testing.T) {
	tests := []struct {
		status int
		body   string
		want   error
	}{
		{status: http.StatusOK, body: `{"total":0,"items":[]}`, want: application.ErrProductNotFound},
		{status: http.StatusNotFound, want: application.ErrProductNotFound},
		{status: http.StatusTooManyRequests, want: application.ErrLookupRateLimited},
	}

	for _, tt := range tests {
		client := NewUPCItemDBClient("https://upcitemdb.test/prod/trial")
		client.client = &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return jsonResponse(tt.status, tt.body), nil
		})}
		if _, err := client.LookupProduct("7501055303038"); !errors.Is(err, tt.want) {
			t.Fatalf("status %d: error = %v, esperado %v", tt.status, err, tt.want)
		}
	}
}
