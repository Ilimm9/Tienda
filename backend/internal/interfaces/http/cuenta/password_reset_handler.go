package cuenta

import (
	"errors"
	"net/http"

	application "tienda/backend/internal/application/cuenta"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const mensajeSolicitudRecuperacion = "Si existe una cuenta con ese correo, enviaremos un enlace para restablecer la contraseña."

type PasswordResetHandler struct {
	resets *application.PasswordResetService
}

func NewPasswordResetHandler(resets *application.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{resets: resets}
}

type passwordResetRequest struct {
	Correo string `json:"correo" binding:"required,email"`
}

func (h *PasswordResetHandler) Request(c *gin.Context) {
	var input passwordResetRequest
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Ingresa un correo válido"})
		return
	}
	if err := h.resets.Request(c.Request.Context(), input.Correo, c.ClientIP()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible procesar la solicitud"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"mensaje": mensajeSolicitudRecuperacion})
}

type passwordResetConfirm struct {
	ChallengeID string `json:"desafio_id" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Contrasena  string `json:"contrasena" binding:"required"`
}

func (h *PasswordResetHandler) Reset(c *gin.Context) {
	var input passwordResetConfirm
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": application.ErrPasswordResetInvalid.Error()})
		return
	}
	challengeID, err := uuid.Parse(input.ChallengeID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"mensaje": application.ErrPasswordResetInvalid.Error()})
		return
	}
	err = h.resets.Reset(c.Request.Context(), challengeID, input.Token, input.Contrasena)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"mensaje": "Tu contraseña se actualizó. Inicia sesión con la nueva contraseña."})
	case errors.Is(err, application.ErrPasswordPolicy):
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": err.Error()})
	case errors.Is(err, application.ErrPasswordResetInvalid):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"mensaje": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible actualizar la contraseña"})
	}
}
