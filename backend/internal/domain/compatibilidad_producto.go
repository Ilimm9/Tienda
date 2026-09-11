package domain

// Compatibilidad temporal para producto y catálogo.
// Eliminar cuando esos dominios migren sus imports a cuenta y negocio.
import (
	cuentadomain "tienda/backend/internal/domain/cuenta"
	negociodomain "tienda/backend/internal/domain/negocio"

	"github.com/google/uuid"
)

type Usuario = cuentadomain.Usuario
type Negocio = negociodomain.Negocio

// setID permanece temporalmente para hooks de modelos legacy de catálogo.
func setID(id *uuid.UUID) {
	if *id == uuid.Nil {
		*id = uuid.New()
	}
}
