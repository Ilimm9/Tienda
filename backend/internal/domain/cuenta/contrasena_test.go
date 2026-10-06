package cuenta

import (
	"strings"
	"testing"
)

func TestContrasenaSegura(t *testing.T) {
	for _, prueba := range []struct {
		nombre     string
		contrasena string
		segura     bool
	}{
		{"válida", "ClaveSegura1!", true},
		{"sin longitud mínima", "A1!cort", false},
		{"sin mayúscula", "clavesegura1!", false},
		{"sin número", "ClaveSegura!", false},
		{"sin símbolo", "ClaveSegura1", false},
		{"espacio no es símbolo", "Clave Segura1", false},
		{"límite máximo", "A1!" + strings.Repeat("a", MaxLongitudContrasena-3), true},
	} {
		t.Run(prueba.nombre, func(t *testing.T) {
			if segura := ContrasenaSegura(prueba.contrasena); segura != prueba.segura {
				t.Fatalf("ContrasenaSegura(%q)=%t, se esperaba %t", prueba.contrasena, segura, prueba.segura)
			}
		})
	}
}
