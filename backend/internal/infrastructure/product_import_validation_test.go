package infrastructure

import "testing"

func TestProductImportNamePresentationKeyNormalizesValues(t *testing.T) {
	first := productImportNamePresentationKey("  Café molido ", " Caja ")
	second := productImportNamePresentationKey("cafe MOLIDO", "caja")
	if first != second {
		t.Fatalf("keys = %q and %q, want equal normalized keys", first, second)
	}
	if first == productImportNamePresentationKey("Café molido", "Bolsa") {
		t.Fatal("different presentations must not share a duplicate key")
	}
}
