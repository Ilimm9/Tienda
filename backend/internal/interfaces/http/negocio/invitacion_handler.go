package negocio

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/negocio"
	domain "tienda/backend/internal/domain/negocio"
	transporthttp "tienda/backend/internal/interfaces/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InvitacionHandler struct {
	invitaciones *application.InvitacionService
}

func NewInvitacionHandler(invitaciones *application.InvitacionService) *InvitacionHandler {
	return &InvitacionHandler{invitaciones: invitaciones}
}

func (h *InvitacionHandler) Listar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	items, err := h.invitaciones.Listar(c.Request.Context(), usuarioID, negocioID, c.Query("estado"))
	if err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func (h *InvitacionHandler) Crear(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	var input domain.CrearInvitacionInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	creada, err := h.invitaciones.Crear(c.Request.Context(), usuarioID, negocioID, input)
	if err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.JSON(http.StatusCreated, creada)
}

func (h *InvitacionHandler) Cancelar(c *gin.Context) {
	usuarioID, negocioID, ok := idsNegocio(c)
	if !ok {
		return
	}
	invitacionID, err := uuid.Parse(c.Param("invitacionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador de la invitación no es válido.",
			"campos": gin.H{"invitacionId": "debe ser UUID"},
		})
		return
	}
	if err := h.invitaciones.Cancelar(c.Request.Context(), usuarioID, negocioID, invitacionID); err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Consultar es público: quien abre el enlace todavía puede no tener cuenta.
func (h *InvitacionHandler) Consultar(c *gin.Context) {
	publica, err := h.invitaciones.Consultar(c.Request.Context(), c.Param("token"))
	if err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.JSON(http.StatusOK, publica)
}

// Aceptar exige sesión iniciada con el correo invitado.
func (h *InvitacionHandler) Aceptar(c *gin.Context) {
	usuarioID, ok := transporthttp.AuthenticatedUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"codigo": "SESION_REQUERIDA", "mensaje": "Debes iniciar sesión.", "campos": gin.H{},
		})
		return
	}
	if err := h.invitaciones.Aceptar(c.Request.Context(), usuarioID, c.Param("token")); err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func responderErrorInvitacion(c *gin.Context, err error) {
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
	case errors.Is(err, application.ErrEmpleadoNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "EMPLEADO_NO_ENCONTRADO", "mensaje": "No fue posible encontrar el empleado solicitado.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{
			"codigo": "INVITACION_NO_ENCONTRADA", "mensaje": "No fue posible encontrar la invitación solicitada.", "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionProhibida):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionCorreoDistinto):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "CORREO_NO_COINCIDE", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionNoVigente),
		errors.Is(err, application.ErrInvitacionYaVinculado),
		errors.Is(err, application.ErrInvitacionCorreoRequerido),
		errors.Is(err, application.ErrInvitacionRolAjeno),
		errors.Is(err, application.ErrEstadoNegocio):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CONFLICTO_INVITACION", "mensaje": err.Error(), "campos": gin.H{},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{},
		})
	}
}
