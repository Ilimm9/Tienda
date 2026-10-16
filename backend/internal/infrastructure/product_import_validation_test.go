package infrastructure

import (
	"testing"

	"github.com/google/uuid"
)

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

func TestImportedUnitSymbolIsStableAndAvoidsExistingSymbols(t *testing.T) {
	used := map[string]struct{}{}
	first := importedUnitSymbol("Pieza", used)
	if first == "" || len(first) > 20 {
		t.Fatalf("symbol = %q, want a generated symbol within the database limit", first)
	}
	used[catalogImportKey(first)] = struct{}{}
	second := importedUnitSymbol("Pieza", used)
	if second == first {
		t.Fatalf("symbol = %q, want a distinct value when the generated symbol is already used", second)
	}
}

func TestSimpleProductDuplicateIssueRequiresVariantForDifferentBarcode(t *testing.T) {
	field, reason := simpleProductDuplicateIssue("7501055303038", map[string]struct{}{}, map[string]struct{}{})
	if field != "Variantes" || reason == "" {
		t.Fatalf("issue = %q, %q; want a variant instruction", field, reason)
	}
	field, _ = simpleProductDuplicateIssue("7501055303038", map[string]struct{}{"7501055303038": {}}, map[string]struct{}{})
	if field != "Producto" {
		t.Fatalf("field = %q, want Producto for a repeated barcode", field)
	}
}
