package cuenta

import "unicode"

const (
	MinLongitudContrasena = 8
	MaxLongitudContrasena = 72
)

// ContrasenaSegura exige longitud, mayúscula, número y símbolo para las cuentas nuevas.
func ContrasenaSegura(contrasena string) bool {
	if len(contrasena) < MinLongitudContrasena || len(contrasena) > MaxLongitudContrasena {
		return false
	}

	tieneMayuscula := false
	tieneNumero := false
	tieneEspecial := false
	for _, caracter := range contrasena {
		switch {
		case unicode.IsUpper(caracter):
			tieneMayuscula = true
		case unicode.IsNumber(caracter):
			tieneNumero = true
		case !unicode.IsLetter(caracter) && !unicode.IsNumber(caracter) && !unicode.IsSpace(caracter):
			tieneEspecial = true
		}
	}

	return tieneMayuscula && tieneNumero && tieneEspecial
}
