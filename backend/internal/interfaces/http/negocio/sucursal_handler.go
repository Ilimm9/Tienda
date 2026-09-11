package negocio

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SucursalHandler struct {
	sucursales *application.SucursalService
}

func NewSucursalHandler(sucursales *application.SucursalService) *SucursalHandler {
	return &SucursalHandler{sucursales: sucursales}
}

func (h *SucursalHandler) Listar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.sucursales.Listar(
		c.Request.Context(), usuarioID, negocioID,
		c.DefaultQuery("estado", "activo"), c.Query("buscar"),
	)
	if err != nil {
		responderErrorSucursal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *SucursalHandler) Crear(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	var input domain.CrearSucursalInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	detalle, err := h.sucursales.Crear(c.Request.Context(), usuarioID, negocioID, input)
	if err != nil {
		responderErrorSucursal(c, err)
		return
	}
	c.JSON(http.StatusCreated, detalle)
}

func (h *SucursalHandler) Obtener(c *gin.Context) {
	usuarioID, negocioID, sucursalID, ok := idsSucursal(c)
	if !ok {
		return
	}
	detalle, err := h.sucursales.Obtener(c.Request.Context(), usuarioID, negocioID, sucursalID)
	if err != nil {
		responderErrorSucursal(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func (h *SucursalHandler) Actualizar(c *gin.Context) {
	usuarioID, negocioID, sucursalID, ok := idsSucursal(c)
	if !ok {
		return
	}
	var input domain.ActualizarSucursalInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	detalle, err := h.sucursales.Actualizar(c.Request.Context(), usuarioID, negocioID, sucursalID, input)
	if err != nil {
		responderErrorSucursal(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func decodificarJSONSucursal(c *gin.Context, target interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("se recibió más de un valor JSON")
		}
		return err
	}
	return nil
}

func (h *SucursalHandler) Archivar(c *gin.Context) {
	usuarioID, negocioID, sucursalID, ok := idsSucursal(c)
	if !ok {
		return
	}
	if err := h.sucursales.Archivar(c.Request.Context(), usuarioID, negocioID, sucursalID); err != nil {
		responderErrorSucursal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SucursalHandler) Restaurar(c *gin.Context) {
	usuarioID, negocioID, sucursalID, ok := idsSucursal(c)
	if !ok {
		return
	}
	detalle, err := h.sucursales.Restaurar(c.Request.Context(), usuarioID, negocioID, sucursalID)
	if err != nil {
		responderErrorSucursal(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func idsSucursal(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	sucursalID, err := uuid.Parse(c.Param("sucursalId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador de la sucursal no es válido.",
			"campos": gin.H{"sucursalId": "debe ser UUID"},
		})
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return usuarioID, negocioID, sucursalID, true
}

func responderDatosInvalidosSucursal(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"codigo": "DATOS_INVALIDOS", "mensaje": "Revisa los datos enviados.", "campos": gin.H{},
	})
}

func responderErrorSucursal(c *gin.Context, err error) {
	var validation *application.ErrorValidacion
	switch {
	case errors.As(err, &validation):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"codigo": "DATOS_INVALIDOS", "mensaje": validation.Error(), "campos": validation.Campos,
		})
	case errors.Is(err, application.ErrNegocioNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "NEGOCIO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el negocio solicitado.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrSucursalNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "SUCURSAL_NO_ENCONTRADA", "mensaje": "No fue posible encontrar la sucursal solicitada.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrSucursalProhibida):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrSucursalSinCambios):
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "SIN_CAMBIOS", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrEstadoSucursal),
		errors.Is(err, application.ErrSucursalConflicto),
		errors.Is(err, application.ErrSucursalPrincipalRequerida),
		errors.Is(err, application.ErrEstadoNegocio):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CONFLICTO_SUCURSAL", "mensaje": err.Error(), "campos": gin.H{},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{},
		})
	}
}
