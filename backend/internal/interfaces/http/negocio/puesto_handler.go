package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PuestoHandler struct {
	puestos *application.PuestoService
}

func NewPuestoHandler(puestos *application.PuestoService) *PuestoHandler {
	return &PuestoHandler{puestos: puestos}
}

func (h *PuestoHandler) Listar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.puestos.Listar(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		responderErrorPuesto(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *PuestoHandler) Crear(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	var input domain.CrearPuestoInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	puesto, err := h.puestos.Crear(c.Request.Context(), usuarioID, negocioID, input)
	if err != nil {
		responderErrorPuesto(c, err)
		return
	}
	c.JSON(http.StatusCreated, puesto)
}

func (h *PuestoHandler) Actualizar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	puestoID, err := uuid.Parse(c.Param("puestoId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del puesto no es válido.",
			"campos": gin.H{"puestoId": "debe ser UUID"},
		})
		return
	}
	var input domain.ActualizarPuestoInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	puesto, err := h.puestos.Actualizar(c.Request.Context(), usuarioID, negocioID, puestoID, input)
	if err != nil {
		responderErrorPuesto(c, err)
		return
	}
	c.JSON(http.StatusOK, puesto)
}

func responderErrorPuesto(c *gin.Context, err error) {
	switch {
	case errors.Is(err, application.ErrPuestoNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "PUESTO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el puesto solicitado.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrPuestoDuplicado):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "PUESTO_DUPLICADO", "mensaje": "Ya existe un puesto con ese nombre.",
			"campos": gin.H{"nombre": "ya existe en el catálogo"},
		})
	default:
		responderErrorEmpleado(c, err)
	}
}
