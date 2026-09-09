package application

import (
	"errors"
	"testing"

	"tienda/backend/internal/domain"
)

type lookupStub struct {
	calledWith string
}

func (s *lookupStub) LookupProduct(barcode string) (domain.ProductLookup, error) {
	s.calledWith = barcode
	return domain.ProductLookup{CodigoBarras: barcode}, nil
}

func TestLookupProductValidatesBarcode(t *testing.T) {
	lookup := &lookupStub{}
	service := NewProductService(nil, lookup)

	if _, err := service.LookupProduct("abc"); !errors.Is(err, ErrInvalidBarcode) {
		t.Fatalf("error = %v, se esperaba ErrInvalidBarcode", err)
	}
	if lookup.calledWith != "" {
		t.Fatal("no se debía consultar al proveedor con un código inválido")
	}

	if _, err := service.LookupProduct(" 7501055303038 "); err != nil {
		t.Fatalf("código válido rechazado: %v", err)
	}
	if lookup.calledWith != "7501055303038" {
		t.Fatalf("código enviado = %q", lookup.calledWith)
	}
}
