package venta

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ventaapp "tienda/backend/internal/application/venta"
	ventadomain "tienda/backend/internal/domain/venta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type accesoFalso struct{ sucursales []uuid.UUID }

func (a accesoFalso) ValidarSucursalActiva(_ context.Context, _, _, sucursalID uuid.UUID) error {
	if !contiene(a.sucursales, sucursalID) {
		return context.Canceled
	}
	return nil
}
func (a accesoFalso) SucursalesAccesibles(context.Context, uuid.UUID, uuid.UUID) ([]uuid.UUID, error) {
	return a.sucursales, nil
}

type permisosFalsos []string

func (p permisosFalsos) PermisosEfectivos(context.Context, uuid.UUID, uuid.UUID) ([]string, error) {
	return p, nil
}

type repoFalso struct {
	filtro ventaapp.Filtro
	venta  ventadomain.Venta
	costo  ventadomain.Centavos
}

func (r *repoFalso) BuscarPorCodigo(uuid.UUID, uuid.UUID, string) ([]ventaapp.ProductoVendible, error) {
	costo := r.costo
	return []ventaapp.ProductoVendible{{Nombre: "Coca-Cola 600 ml", Precio: 1850, Costo: &costo}}, nil
}
func (r *repoFalso) BuscarPorNombre(uuid.UUID, uuid.UUID, []string, int) ([]ventaapp.ProductoVendible, error) {
	return nil, nil
}
func (r *repoFalso) Registrar(ventaapp.RegistrarComando) (ventadomain.Venta, error) {
	return r.venta, nil
}
func (r *repoFalso) Cancelar(ventaapp.CancelarComando) (ventadomain.Venta, error) {
	return r.venta, nil
}
func (r *repoFalso) Listar(_ uuid.UUID, filtro ventaapp.Filtro) (ventaapp.ListadoVentas, error) {
	r.filtro = filtro
	return ventaapp.ListadoVentas{Items: []ventadomain.Venta{}}, nil
}
func (r *repoFalso) Obtener(uuid.UUID, uuid.UUID) (ventadomain.Venta, error) { return r.venta, nil }

func TestPermisosDeVentas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	negocio, usuario, otroCajero, sucursal := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	ventaID := uuid.New()
	// Cada llamada recibe un ticket nuevo, como una lectura real de la base.
	ventaAjena := func() ventadomain.Venta {
		costo := ventadomain.Centavos(1200)
		return ventadomain.Venta{ID: ventaID, SucursalID: sucursal, UsuarioID: otroCajero, Lineas: []ventadomain.VentaLinea{{CostoUnitario: &costo}}}
	}

	llamar := func(permisos []string, repo *repoFalso, metodo, ruta string) *httptest.ResponseRecorder {
		h := NewHandler(ventaapp.NewService(repo), accesoFalso{sucursales: []uuid.UUID{sucursal}}, permisosFalsos(permisos))
		router := gin.New()
		router.Use(func(c *gin.Context) { c.Set("authenticated_user_id", usuario); c.Next() })
		router.GET("/negocios/:negocioId/ventas", h.List)
		router.GET("/negocios/:negocioId/ventas/productos", h.Productos)
		router.GET("/negocios/:negocioId/ventas/:ventaId", h.Get)
		respuesta := httptest.NewRecorder()
		router.ServeHTTP(respuesta, httptest.NewRequest(metodo, ruta, nil))
		return respuesta
	}
	base := "/negocios/" + negocio.String() + "/ventas"

	t.Run("sin permisos de ventas responde 403", func(t *testing.T) {
		if r := llamar([]string{"compras.ver"}, &repoFalso{}, http.MethodGet, base); r.Code != http.StatusForbidden {
			t.Fatalf("estado %d", r.Code)
		}
	})

	t.Run("el cajero solo lista sus propias ventas aunque pida otro cajero", func(t *testing.T) {
		repo := &repoFalso{}
		r := llamar([]string{"ventas.registrar"}, repo, http.MethodGet, base+"?usuario_id="+otroCajero.String())
		if r.Code != http.StatusOK || repo.filtro.UsuarioID == nil || *repo.filtro.UsuarioID != usuario {
			t.Fatalf("estado %d filtro %+v", r.Code, repo.filtro)
		}
		repo = &repoFalso{}
		llamar([]string{"ventas.ver"}, repo, http.MethodGet, base+"?usuario_id="+otroCajero.String())
		if repo.filtro.UsuarioID == nil || *repo.filtro.UsuarioID != otroCajero {
			t.Fatalf("con ventas.ver se filtra por el cajero elegido: %+v", repo.filtro)
		}
	})

	t.Run("el periodo se valida", func(t *testing.T) {
		repo := &repoFalso{}
		if r := llamar([]string{"ventas.ver"}, repo, http.MethodGet, base+"?periodo=semana"); r.Code != http.StatusOK || repo.filtro.Periodo != "semana" {
			t.Fatalf("estado %d filtro %+v", r.Code, repo.filtro)
		}
		if r := llamar([]string{"ventas.ver"}, &repoFalso{}, http.MethodGet, base+"?periodo=anio"); r.Code != http.StatusBadRequest {
			t.Fatalf("periodo inválido: %d", r.Code)
		}
	})

	t.Run("una sucursal ajena responde 404", func(t *testing.T) {
		if r := llamar([]string{"ventas.ver"}, &repoFalso{}, http.MethodGet, base+"?sucursal_id="+uuid.NewString()); r.Code != http.StatusNotFound {
			t.Fatalf("estado %d", r.Code)
		}
	})

	t.Run("el ticket de otro cajero responde 404 sin ventas.ver", func(t *testing.T) {
		ruta := base + "/" + ventaID.String()
		if r := llamar([]string{"ventas.registrar"}, &repoFalso{venta: ventaAjena()}, http.MethodGet, ruta); r.Code != http.StatusNotFound {
			t.Fatalf("estado %d", r.Code)
		}
		r := llamar([]string{"ventas.ver"}, &repoFalso{venta: ventaAjena()}, http.MethodGet, ruta)
		var cuerpo ventadomain.Venta
		_ = json.Unmarshal(r.Body.Bytes(), &cuerpo)
		if r.Code != http.StatusOK || cuerpo.Lineas[0].CostoUnitario != nil {
			t.Fatalf("con ventas.ver se ve, sin costos: %d %s", r.Code, r.Body.String())
		}
		r = llamar([]string{"ventas.ver", "ventas.ver_costos"}, &repoFalso{venta: ventaAjena()}, http.MethodGet, ruta)
		_ = json.Unmarshal(r.Body.Bytes(), &cuerpo)
		if cuerpo.Lineas[0].CostoUnitario == nil || *cuerpo.Lineas[0].CostoUnitario != 1200 {
			t.Fatalf("con ventas.ver_costos se envía el costo: %s", r.Body.String())
		}
	})

	t.Run("la búsqueda solo envía costos con ventas.ver_costos", func(t *testing.T) {
		ruta := base + "/productos?q=750&sucursal_id=" + sucursal.String()
		r := llamar([]string{"ventas.registrar"}, &repoFalso{costo: 1200}, http.MethodGet, ruta)
		var cuerpo ventaapp.ResultadoBusqueda
		_ = json.Unmarshal(r.Body.Bytes(), &cuerpo)
		if r.Code != http.StatusOK || len(cuerpo.Items) != 1 || cuerpo.Items[0].Costo != nil {
			t.Fatalf("sin costos: %d %s", r.Code, r.Body.String())
		}
		r = llamar([]string{"ventas.registrar", "ventas.ver_costos"}, &repoFalso{costo: 1200}, http.MethodGet, ruta)
		_ = json.Unmarshal(r.Body.Bytes(), &cuerpo)
		if cuerpo.Items[0].Costo == nil || cuerpo.Items[0].NivelMargen == nil {
			t.Fatalf("con costos: %s", r.Body.String())
		}
	})
}
