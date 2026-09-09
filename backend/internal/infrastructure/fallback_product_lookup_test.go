package infrastructure

import (
	"errors"
	"testing"

	"tienda/backend/internal/application"
	"tienda/backend/internal/domain"
)

type productLookupStub struct {
	product domain.ProductLookup
	err     error
	calls   int
}

func (s *productLookupStub) LookupProduct(string) (domain.ProductLookup, error) {
	s.calls++
	return s.product, s.err
}

func TestFallbackLookupSkipsFallbackForCompletePrimaryResult(t *testing.T) {
	image := "https://cdn.test/primary.jpg"
	primary := &productLookupStub{product: domain.ProductLookup{CodigoBarras: "1", ImagenURL: &image, Fuentes: []string{"PrecioCheck"}}}
	fallback := &productLookupStub{}
	lookup := NewFallbackProductLookup(primary, fallback)

	if _, err := lookup.LookupProduct("1"); err != nil {
		t.Fatal(err)
	}
	if fallback.calls != 0 {
		t.Fatalf("el fallback recibió %d llamadas", fallback.calls)
	}
	if _, err := lookup.LookupProduct("1"); err != nil || primary.calls != 1 {
		t.Fatalf("la segunda consulta no usó caché: error=%v, llamadas=%d", err, primary.calls)
	}
}

func TestFallbackLookupMergesMissingPrimaryFields(t *testing.T) {
	name := "Producto principal"
	secondaryDescription := "Descripción secundaria"
	secondaryImage := "https://cdn.test/secondary.jpg"
	price := 20.0
	primary := &productLookupStub{product: domain.ProductLookup{Nombre: &name, PrecioSugerido: &price, Fuentes: []string{"PrecioCheck"}}}
	fallback := &productLookupStub{product: domain.ProductLookup{Descripcion: &secondaryDescription, ImagenURL: &secondaryImage, Fuentes: []string{"UPCitemdb"}}}

	result, err := NewFallbackProductLookup(primary, fallback).LookupProduct("1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Nombre != &name || result.Descripcion == nil || *result.Descripcion != secondaryDescription || result.ImagenURL == nil {
		t.Fatalf("resultado combinado inesperado: %+v", result)
	}
	if result.PrecioSugerido == nil || *result.PrecioSugerido != price || len(result.Fuentes) != 2 {
		t.Fatalf("precedencia o fuentes inesperadas: %+v", result)
	}
}

func TestFallbackLookupUsesSecondaryAndCombinesErrors(t *testing.T) {
	secondaryName := "Solo UPCitemdb"
	primary := &productLookupStub{err: application.ErrProductNotFound}
	fallback := &productLookupStub{product: domain.ProductLookup{Nombre: &secondaryName, Fuentes: []string{"UPCitemdb"}}}
	result, err := NewFallbackProductLookup(primary, fallback).LookupProduct("1")
	if err != nil || result.Nombre == nil || *result.Nombre != secondaryName {
		t.Fatalf("fallback inesperado: resultado=%+v error=%v", result, err)
	}

	missing := NewFallbackProductLookup(
		&productLookupStub{err: application.ErrProductNotFound},
		&productLookupStub{err: application.ErrProductNotFound},
	)
	if _, err := missing.LookupProduct("2"); !errors.Is(err, application.ErrProductNotFound) {
		t.Fatalf("error combinado = %v", err)
	}
}
