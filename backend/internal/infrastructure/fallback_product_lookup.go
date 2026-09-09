package infrastructure

import (
	"errors"
	"sync"
	"time"

	"tienda/backend/internal/application"
	"tienda/backend/internal/domain"
)

type lookupCacheEntry struct {
	product   domain.ProductLookup
	err       error
	expiresAt time.Time
}

type FallbackProductLookup struct {
	primary  application.ProductLookupProvider
	fallback application.ProductLookupProvider
	mu       sync.RWMutex
	cache    map[string]lookupCacheEntry
	now      func() time.Time
}

func NewFallbackProductLookup(primary, fallback application.ProductLookupProvider) *FallbackProductLookup {
	return &FallbackProductLookup{
		primary: primary, fallback: fallback,
		cache: make(map[string]lookupCacheEntry), now: time.Now,
	}
}

func (f *FallbackProductLookup) LookupProduct(barcode string) (domain.ProductLookup, error) {
	if product, err, ok := f.cached(barcode); ok {
		return product, err
	}

	primaryProduct, primaryErr := f.primary.LookupProduct(barcode)
	if primaryErr == nil && primaryProduct.ImagenURL != nil {
		f.store(barcode, primaryProduct, nil, 24*time.Hour)
		return primaryProduct, nil
	}

	fallbackProduct, fallbackErr := f.fallback.LookupProduct(barcode)
	if fallbackErr == nil {
		if primaryErr == nil {
			merged := mergeProductLookups(primaryProduct, fallbackProduct)
			f.store(barcode, merged, nil, 24*time.Hour)
			return merged, nil
		}
		f.store(barcode, fallbackProduct, nil, 24*time.Hour)
		return fallbackProduct, nil
	}

	if primaryErr == nil {
		return primaryProduct, nil
	}
	err := combinedLookupError(primaryErr, fallbackErr)
	if errors.Is(err, application.ErrProductNotFound) {
		f.store(barcode, domain.ProductLookup{}, err, 15*time.Minute)
	}
	return domain.ProductLookup{}, err
}

func mergeProductLookups(primary, fallback domain.ProductLookup) domain.ProductLookup {
	result := primary
	if result.Nombre == nil {
		result.Nombre = fallback.Nombre
	}
	if result.Descripcion == nil {
		result.Descripcion = fallback.Descripcion
	}
	if result.Marca == nil {
		result.Marca = fallback.Marca
	}
	if result.Categoria == nil {
		result.Categoria = fallback.Categoria
	}
	if result.Contenido == nil {
		result.Contenido, result.UnidadContenido = fallback.Contenido, fallback.UnidadContenido
	}
	if result.ImagenURL == nil {
		result.ImagenURL = fallback.ImagenURL
	}
	result.Fuentes = appendUnique(result.Fuentes, fallback.Fuentes...)
	return result
}

func appendUnique(values []string, candidates ...string) []string {
	for _, candidate := range candidates {
		found := false
		for _, value := range values {
			if value == candidate {
				found = true
				break
			}
		}
		if !found {
			values = append(values, candidate)
		}
	}
	return values
}

func combinedLookupError(primaryErr, fallbackErr error) error {
	if errors.Is(primaryErr, application.ErrProductNotFound) && errors.Is(fallbackErr, application.ErrProductNotFound) {
		return application.ErrProductNotFound
	}
	if errors.Is(primaryErr, application.ErrLookupRateLimited) || errors.Is(fallbackErr, application.ErrLookupRateLimited) {
		return application.ErrLookupRateLimited
	}
	return application.ErrLookupUnavailable
}

func (f *FallbackProductLookup) cached(barcode string) (domain.ProductLookup, error, bool) {
	f.mu.RLock()
	entry, ok := f.cache[barcode]
	f.mu.RUnlock()
	if !ok || !f.now().Before(entry.expiresAt) {
		if ok {
			f.mu.Lock()
			delete(f.cache, barcode)
			f.mu.Unlock()
		}
		return domain.ProductLookup{}, nil, false
	}
	return entry.product, entry.err, true
}

func (f *FallbackProductLookup) store(barcode string, product domain.ProductLookup, err error, ttl time.Duration) {
	f.mu.Lock()
	f.cache[barcode] = lookupCacheEntry{product: product, err: err, expiresAt: f.now().Add(ttl)}
	f.mu.Unlock()
}

var _ application.ProductLookupProvider = (*FallbackProductLookup)(nil)
