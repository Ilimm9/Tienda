package negocio

import (
	"errors"
	"reflect"
	"testing"
)

func TestNormalizarSeccionesOcultasOrdenaYQuitaRepetidas(t *testing.T) {
	got, err := NormalizarSeccionesOcultas([]string{" Marcas", "categorias", "marcas"})
	if err != nil || !reflect.DeepEqual(got, []string{"categorias", "marcas"}) {
		t.Fatalf("got = %#v, err = %v", got, err)
	}
	got, err = NormalizarSeccionesOcultas([]string{})
	if err != nil || len(got) != 0 {
		t.Fatalf("lista vacía: got = %#v, err = %v", got, err)
	}
}

func TestNormalizarSeccionesOcultasRechazaSeccionesFijas(t *testing.T) {
	for _, seccion := range []string{"productos", "unidades", "proveedores", ""} {
		if _, err := NormalizarSeccionesOcultas([]string{seccion}); !errors.Is(err, ErrSeccionNoOcultable) {
			t.Fatalf("%q: err = %v, want ErrSeccionNoOcultable", seccion, err)
		}
	}
}

func TestLeerYEscribirSeccionesOcultas(t *testing.T) {
	if got := LeerSeccionesOcultas(""); len(got) != 0 {
		t.Fatalf("vacío = %#v", got)
	}
	if got := LeerSeccionesOcultas("marcas,obsoleta,categorias"); !reflect.DeepEqual(got, []string{"categorias", "marcas"}) {
		t.Fatalf("got = %#v", got)
	}
	if got := EscribirSeccionesOcultas([]string{"categorias", "marcas"}); got != "categorias,marcas" {
		t.Fatalf("got = %q", got)
	}
}
