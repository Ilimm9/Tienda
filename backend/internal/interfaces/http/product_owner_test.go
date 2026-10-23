package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	negocioapplication "tienda/backend/internal/application/negocio"
	negociodomain "tienda/backend/internal/domain/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextoFalso struct {
	opciones []negociodomain.ContextoNegocio
}

func (f contextoFalso) ListarOpcionesContexto(context.Context, uuid.UUID) ([]negociodomain.ContextoNegocio, error) {
	return f.opciones, nil
}
func (contextoFalso) NegocioActivoAccesible(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}
func (contextoFalso) SucursalActivaAccesible(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}
func (contextoFalso) SucursalesAccesibles(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

// TestSoloPropietarioCreaProductos covers capability 31: only the business owner
// may create or import products.
func TestSoloPropietarioCreaProductos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	negocio, otroNegocio := uuid.New(), uuid.New()
	casos := []struct {
		nombre   string
		opciones []negociodomain.ContextoNegocio
		esperado int
	}{
		{"propietario", []negociodomain.ContextoNegocio{{ID: negocio, TipoMiembro: "propietario"}}, http.StatusNoContent},
		{"miembro con permisos", []negociodomain.ContextoNegocio{{ID: negocio, TipoMiembro: "miembro", Permisos: []string{"catalogo.gestionar", "precios.autorizar"}}}, http.StatusForbidden},
		{"propietario de otro negocio", []negociodomain.ContextoNegocio{{ID: otroNegocio, TipoMiembro: "propietario"}, {ID: negocio, TipoMiembro: "miembro"}}, http.StatusForbidden},
		{"sin membresía", nil, http.StatusForbidden},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			handler := &ProductHandler{contexto: negocioapplication.NewContextoService(contextoFalso{opciones: caso.opciones})}
			router := gin.New()
			router.POST("/negocios/:negocioId/catalogo/productos", func(c *gin.Context) {
				c.Set(authenticatedUserKey, uuid.New())
			}, handler.SoloPropietario(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
			res := httptest.NewRecorder()
			router.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/negocios/"+negocio.String()+"/catalogo/productos", nil))
			if res.Code != caso.esperado {
				t.Fatalf("status = %d, esperado %d", res.Code, caso.esperado)
			}
			if caso.esperado == http.StatusForbidden {
				var body map[string]string
				_ = json.Unmarshal(res.Body.Bytes(), &body)
				if body["codigo"] != "SOLO_PROPIETARIOS" {
					t.Fatalf("respuesta = %s", res.Body.String())
				}
			}
		})
	}
	t.Run("sin sesión", func(t *testing.T) {
		handler := &ProductHandler{contexto: negocioapplication.NewContextoService(contextoFalso{})}
		router := gin.New()
		router.POST("/negocios/:negocioId/catalogo/productos", handler.SoloPropietario())
		res := httptest.NewRecorder()
		router.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/negocios/"+negocio.String()+"/catalogo/productos", nil))
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d", res.Code)
		}
	})
}
