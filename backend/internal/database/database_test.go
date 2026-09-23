package database

import "testing"

func TestNowUTCReturnsUTC(t *testing.T) {
	if location := nowUTC().Location(); location.String() != "UTC" {
		t.Fatalf("ubicación del reloj = %s; se esperaba UTC", location)
	}
}
