package infrastructure

import (
	"testing"

	"github.com/google/uuid"
)

func TestProductImportIdentityKeyNormalizesValues(t *testing.T) {
	first := productImportIdentityKey("  Café molido ", " SKU-01 ", " Caja ")
	second := productImportIdentityKey("cafe MOLIDO", "sku-01", "caja")
	if first != second {
		t.Fatalf("keys = %q and %q, want equal normalized keys", first, second)
	}
	if first == productImportIdentityKey("Café molido", "sku-01", "Bolsa") {
		t.Fatal("different presentations must not share a duplicate key")
	}
	if first == productImportIdentityKey("Café molido", "sku-02", "Caja") {
		t.Fatal("different SKUs must not share a duplicate key")
	}
}

func TestProductImportNamePresentationKeyNormalizesValues(t *testing.T) {
	first := productImportNamePresentationKey("  Café molido ", " Caja ")
	second := productImportNamePresentationKey("cafe MOLIDO", "caja")
	if first != second {
		t.Fatalf("keys = %q and %q, want equal normalized keys", first, second)
	}
	if first == productImportNamePresentationKey("Café molido", "Bolsa") {
		t.Fatal("different presentations must not share a fallback duplicate key")
	}
}

func TestProductBaseKeyUsesOnlyConfiguredBaseIdentity(t *testing.T) {
	category := uuid.New()
	brand := uuid.New()
	first := productBaseKey("Peñafiel 600 ml", stringPointer("Botella"), &brand, &category)
	second := productBaseKey("penafiel 600 ML", stringPointer("botella"), &brand, &category)
	if first != second {
		t.Fatal("same name, presentation, brand and category must share a base key")
	}
	if first == productBaseKey("Peñafiel 600 ml", stringPointer("Lata"), &brand, &category) {
		t.Fatal("different presentations must not share a base key")
	}
	differentCategory := uuid.New()
	if first == productBaseKey("Peñafiel 600 ml", stringPointer("Botella"), &brand, &differentCategory) {
		t.Fatal("different categories must not share a base key")
	}
	withoutCategory := productBaseKey("Peñafiel 600 ml", stringPointer("Botella"), &brand, nil)
	if withoutCategory != productBaseKey("penafiel 600 ML", stringPointer("botella"), &brand, nil) {
		t.Fatal("rows without a category must share the same base key")
	}
	if withoutCategory == first {
		t.Fatal("a missing category and an assigned category must not share a base key")
	}
}

func TestSameVariantBaseInputRejectsSharedDataConflicts(t *testing.T) {
	content := 600.0
	first := variantBaseInput{Descripcion: stringPointer("Refresco"), Contenido: &content, UnidadContenido: stringPointer("ml")}
	second := variantBaseInput{Descripcion: stringPointer("Refresco"), Contenido: &content, UnidadContenido: stringPointer("ml")}
	if !sameVariantBaseInput(first, second) {
		t.Fatal("equal base data should be compatible")
	}
	second.Descripcion = stringPointer("Otra descripción")
	if sameVariantBaseInput(first, second) {
		t.Fatal("different shared descriptions must be rejected")
	}
}
