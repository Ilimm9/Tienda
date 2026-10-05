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
	filtro := domain.FiltroInvitaciones{Estado: c.Query("estado"), SinAceptar: c.Query("sin_aceptar") == "true"}
	if valor := c.Query("sucursal_id"); valor != "" {
		sucursalID, err := uuid.Parse(valor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador de la sucursal no es válido.",
				"campos": gin.H{"sucursal_id": "debe ser UUID"},
			})
			return
		}
		filtro.SucursalID = &sucursalID
	}
	items, err := h.invitaciones.Listar(c.Request.Context(), usuarioID, negocioID, filtro)
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

// Reenviar emite un enlace nuevo para la misma sucursal y rol; con `correo` corrige al destinatario.
func (h *InvitacionHandler) Reenviar(c *gin.Context) {
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
	var input domain.ReenviarInvitacionInput
	// Un cuerpo vacío equivale a reenviar al mismo destinatario.
	if c.Request.ContentLength != 0 {
		if err := decodificarJSONSucursal(c, &input); err != nil {
			responderDatosInvalidosSucursal(c)
			return
		}
	}
	creada, err := h.invitaciones.Reenviar(c.Request.Context(), usuarioID, negocioID, invitacionID, input)
	if err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.JSON(http.StatusCreated, creada)
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

// Registrar es público: crea la cuenta pendiente del invitado y devuelve el desafío OTP.
func (h *InvitacionHandler) Registrar(c *gin.Context) {
	var input domain.RegistroInvitacionInput
	if err := decodificarJSONSucursal(c, &input); err != nil {
		responderDatosInvalidosSucursal(c)
		return
	}
	desafio, err := h.invitaciones.Registrar(c.Request.Context(), c.Param("token"), input, c.ClientIP())
	if errors.Is(err, application.ErrInvitacionEnvioCodigo) {
		// La cuenta quedó pendiente: el desafío permite pedir otro código sin repetir el registro.
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"codigo": "CODIGO_NO_ENVIADO", "mensaje": "La cuenta quedó pendiente, pero no fue posible enviar el código. Solicita otro.",
			"campos": gin.H{}, "desafio_id": desafio.DesafioID,
			"correo_enmascarado": desafio.CorreoEnmascarado, "reenviar_en_segundos": desafio.ReenviarEnSegundos,
		})
		return
	}
	if err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.JSON(http.StatusAccepted, desafio)
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
	aceptada, err := h.invitaciones.Aceptar(c.Request.Context(), usuarioID, c.Param("token"))
	if err != nil {
		responderErrorInvitacion(c, err)
		return
	}
	c.JSON(http.StatusOK, aceptada)
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
	case errors.Is(err, application.ErrRolNoDelegable):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ROL_NO_DELEGABLE", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionProhibida):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "ACCESO_DENEGADO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrEmpleadoCorreoDuplicado):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CORREO_EMPLEADO_DUPLICADO", "mensaje": "Ese correo ya está registrado en este negocio.",
			"campos": gin.H{"correo": "ya está registrado en este negocio"},
		})
	case errors.Is(err, application.ErrInvitacionCorreoDistinto):
		c.JSON(http.StatusForbidden, gin.H{
			"codigo": "CORREO_NO_COINCIDE", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionNoVigente):
		c.JSON(http.StatusGone, gin.H{
			"codigo": "INVITACION_NO_VIGENTE", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionLimite):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"codigo": "LIMITE_EXCEDIDO", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionYaAceptada):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "INVITACION_YA_ACEPTADA", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionCuentaExistente):
		c.JSON(http.StatusConflict, gin.H{
			"codigo": "CUENTA_EXISTENTE", "mensaje": err.Error(), "campos": gin.H{},
		})
	case errors.Is(err, application.ErrInvitacionReemplazada),
		errors.Is(err, application.ErrInvitacionSinSucursal),
		errors.Is(err, application.ErrInvitacionYaVinculado),
		errors.Is(err, application.ErrInvitacionCorreoRequerido),
		errors.Is(err, application.ErrInvitacionRolNoDisponible),
		errors.Is(err, application.ErrInvitacionSucursalNoDisponible),
		errors.Is(err, application.ErrInvitacionEmpleadoNoElegible),
		errors.Is(err, application.ErrInvitacionMembresiaInactiva),
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
