package infrastructure

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"tienda/backend/internal/application"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func testPrecioCheckClient(handler roundTripFunc) *PrecioCheckClient {
	client := NewPrecioCheckClient("https://preciocheck.test/api/v1", "secret-test-key")
	client.client = &http.Client{Transport: handler}
	return client
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestPrecioCheckClientLookupProduct(t *testing.T) {
	client := testPrecioCheckClient(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/producto/7501055303038" {
			t.Fatalf("ruta inesperada: %s", r.URL.Path)
		}
		if got := r.Header.Get("X-API-Key"); got != "secret-test-key" {
			t.Fatalf("API key inesperada: %q", got)
		}
		return jsonResponse(http.StatusOK, `{"ok":true,"producto":{"codigo_barras":"7501055303038","nombre":" Coca-Cola 600ml ","descripcion":"Refresco de cola","marca":"Coca-Cola","categoria":"Bebidas","precio_sugerido":18.5,"contenido_neto":"600 ml","imagen_url":"https://cdn.example.com/producto.jpg"}}`), nil
	})

	result, err := client.LookupProduct("7501055303038")
	if err != nil {
		t.Fatalf("LookupProduct devolvió error: %v", err)
	}
	if result.CodigoBarras != "7501055303038" || result.Nombre == nil || *result.Nombre != "Coca-Cola 600ml" {
		t.Fatalf("datos descriptivos inesperados: %+v", result)
	}
	if result.PrecioSugerido == nil || *result.PrecioSugerido != 18.5 || result.Contenido == nil || *result.Contenido != 600 {
		t.Fatalf("precio o contenido inesperado: %+v", result)
	}
	if result.UnidadContenido == nil || *result.UnidadContenido != "ml" || result.ImagenURL == nil || *result.ImagenURL != "https://cdn.example.com/producto.jpg" {
		t.Fatalf("resultado inesperado: %+v", result)
	}
}

func TestParseNetContent(t *testing.T) {
	tests := []struct {
		input  *string
		amount *float64
		unit   *string
	}{
		{input: stringPointer("1,5 L"), amount: floatPointer(1.5), unit: stringPointer("l")},
		{input: stringPointer("250 µg"), amount: floatPointer(250), unit: stringPointer("ug")},
		{input: stringPointer("6 x 100 ml")},
		{input: nil},
	}

	for _, tt := range tests {
		amount, unit := parseNetContent(tt.input)
		if !equalFloatPointer(amount, tt.amount) || !equalStringPointer(unit, tt.unit) {
			t.Fatalf("parseNetContent(%v) = %v, %v; esperado %v, %v", tt.input, amount, unit, tt.amount, tt.unit)
		}
	}
}

func stringPointer(value string) *string  { return &value }
func floatPointer(value float64) *float64 { return &value }

func equalStringPointer(a, b *string) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func equalFloatPointer(a, b *float64) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func TestPrecioCheckClientAllowsProductWithoutImage(t *testing.T) {
	client := testPrecioCheckClient(func(_ *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"ok":true,"producto":{"codigo_barras":"7501055303038","imagen_url":null}}`), nil
	})

	result, err := client.LookupProduct("7501055303038")
	if err != nil {
		t.Fatalf("LookupProduct devolvió error: %v", err)
	}
	if result.ImagenURL != nil {
		t.Fatalf("se esperaba imagen nil: %+v", result)
	}
}

func TestPrecioCheckClientMapsUpstreamErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{name: "not found", status: http.StatusNotFound, want: application.ErrProductNotFound},
		{name: "rate limited", status: http.StatusTooManyRequests, want: application.ErrLookupRateLimited},
		{name: "unauthorized", status: http.StatusUnauthorized, want: application.ErrLookupUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testPrecioCheckClient(func(_ *http.Request) (*http.Response, error) {
				return jsonResponse(tt.status, ""), nil
			})
			_, err := client.LookupProduct("7501055303038")
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, se esperaba %v", err, tt.want)
			}
		})
	}
}

func TestPrecioCheckClientRejectsMissingKeyAndInvalidImageURL(t *testing.T) {
	if _, err := NewPrecioCheckClient("https://preciocheck.com/api/v1", "").LookupProduct("7501055303038"); !errors.Is(err, application.ErrLookupUnavailable) {
		t.Fatalf("API key vacía: error = %v", err)
	}

	client := testPrecioCheckClient(func(_ *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{"ok":true,"producto":{"imagen_url":"http://insecure.example.com/producto.jpg"}}`), nil
	})
	if _, err := client.LookupProduct("7501055303038"); !errors.Is(err, application.ErrLookupUnavailable) {
		t.Fatalf("URL insegura: error = %v", err)
	}
}
