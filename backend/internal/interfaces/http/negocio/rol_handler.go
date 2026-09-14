package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RolHandler struct {
	roles *application.RolService
}

func NewRolHandler(roles *application.RolService) *RolHandler {
	return &RolHandler{roles: roles}
}

// Permisos expone el catálogo global para construir la matriz de la interfaz.
func (h *RolHandler) Permisos(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.roles.ListarPermisos(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

// MisPermisos devuelve los permisos efectivos de la sesión sobre el negocio indicado.
func (h *RolHandler) MisPermisos(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	codigos, err := h.roles.PermisosEfectivos(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": codigos, "total": len(codigos)})
}

func (h *RolHandler) Listar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.roles.Listar(c.Request.Context(), usuarioID, negocioID, c.Query("estado") == "todos")
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *RolHandler) Crear(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	var input domain.CrearRolInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	detalle, err := h.roles.Crear(c.Request.Context(), usuarioID, negocioID, input)
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusCreated, detalle)
}

func (h *RolHandler) Obtener(c *gin.Context) {
	usuarioID, negocioID, rolID, ok := idsRol(c)
	if !ok {
		return
	}
	detalle, err := h.roles.Obtener(c.Request.Context(), usuarioID, negocioID, rolID)
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func (h *RolHandler) Actualizar(c *gin.Context) {
	usuarioID, negocioID, rolID, ok := idsRol(c)
	if !ok {
		return
	}
	var input domain.ActualizarRolInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	detalle, err := h.roles.Actualizar(c.Request.Context(), usuarioID, negocioID, rolID, input)
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusOK, detalle)
}

func (h *RolHandler) Eliminar(c *gin.Context) {
	usuarioID, negocioID, rolID, ok := idsRol(c)
	if !ok {
		return
	}
	if err := h.roles.Eliminar(c.Request.Context(), usuarioID, negocioID, rolID); err != nil {
		responderErrorRol(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RolHandler) ListarMiembros(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.roles.ListarMiembros(c.Request.Context(), usuarioID, negocioID)
	if err != nil {
		responderErrorRol(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *RolHandler) AsignarRoles(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	membresiaID, err := uuid.Parse(c.Param("membresiaId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador de la membresía no es válido.",
			"campos": gin.H{"membresiaId": "debe ser UUID"},
		})
		return
	}
	var input struct {
		Roles []uuid.UUID `json:"roles"`
	}
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	if err := h.roles.AsignarRoles(c.Request.Context(), usuarioID, negocioID, membresiaID, input.Roles); err != nil {
		responderErrorRol(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func idsRol(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	rolID, err := uuid.Parse(c.Param("rolId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del rol no es válido.",
			"campos": gin.H{"rolId": "debe ser UUID"},
		})
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	return usuarioID, negocioID, rolID, true
}

func responderErrorRol(c *gin.Context, err error) {
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
	case errors.Is(err, application.ErrRolNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "ROL_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el rol solicitado.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrMembresiaNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "MEMBRESIA_NO_ENCONTRADA", "mensaje": "No fue posible encontrar la membresía solicitada.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrRolProhibido):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrRolSinCambios):
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "SIN_CAMBIOS", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrRolConflicto),
		errors.Is(err, application.ErrRolSistemaInmutable),
		errors.Is(err, application.ErrRolEnUso),
		errors.Is(err, application.ErrPropietarioSinRol),
		errors.Is(err, application.ErrEstadoNegocio):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CONFLICTO_ROL", "mensaje": err.Error(), "campos": gin.H{},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{},
		})
	}
}
