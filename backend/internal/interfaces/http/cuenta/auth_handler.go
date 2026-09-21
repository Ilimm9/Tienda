package cuenta

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	application "tienda/backend/internal/application/cuenta"
	"tienda/backend/internal/config"
	transporthttp "tienda/backend/internal/interfaces/http"
	"time"
)

type AuthHandler struct {
	auth         *application.AuthService
	sessions     *application.SessionService
	verification *application.VerificationService
	cfg          config.Config
}

func NewAuthHandler(auth *application.AuthService, sessions *application.SessionService, verification *application.VerificationService, cfg config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, sessions: sessions, verification: verification, cfg: cfg}
}

type loginRequest struct {
	Correo     string `json:"correo" binding:"required,email"`
	Contrasena string `json:"contrasena" binding:"required"`
	Recordarme bool   `json:"recordarme"`
}

type registerRequest struct {
	NombreCompleto string `json:"nombre_completo" binding:"required"`
	Correo         string `json:"correo" binding:"required,email"`
	Telefono       string `json:"telefono"`
	Contrasena     string `json:"contrasena" binding:"required,min=8"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input registerRequest
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Revisa los datos del formulario"})
		return
	}
	result, err := h.verification.Register(c.Request.Context(), input.NombreCompleto, input.Correo, input.Telefono, input.Contrasena, clientIP(c))
	if err != nil {
		if errors.Is(err, application.ErrVerificationTooSoon) || errors.Is(err, application.ErrVerificationLimited) {
			c.JSON(http.StatusTooManyRequests, gin.H{"mensaje": err.Error()})
			return
		}
		if errors.Is(err, application.ErrEmailDelivery) {
			c.JSON(http.StatusServiceUnavailable, registrationResponse(result, "La cuenta quedó pendiente, pero no fue posible enviar el código. Intenta reenviarlo."))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible crear la cuenta"})
		return
	}
	c.JSON(http.StatusAccepted, registrationResponse(result, "Si el correo puede registrarse, enviaremos un código de verificación."))
}

type verifyRequest struct {
	ChallengeID string `json:"desafio_id" binding:"required"`
	Code        string `json:"codigo" binding:"required,len=6,numeric"`
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var input verifyRequest
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": application.ErrVerificationInvalid.Error()})
		return
	}
	challengeID, err := uuid.Parse(input.ChallengeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": application.ErrVerificationInvalid.Error()})
		return
	}
	user, err := h.verification.Verify(c.Request.Context(), challengeID, input.Code)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"mensaje": application.ErrVerificationInvalid.Error()})
		return
	}
	session, err := h.sessions.Create(user.ID, false, clientIP(c), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "La cuenta fue verificada, pero no fue posible iniciar la sesión"})
		return
	}
	h.setCookies(c, session)
	c.JSON(http.StatusOK, gin.H{"mensaje": "Correo verificado correctamente", "usuario": gin.H{"id": user.ID, "correo": user.Correo}})
}

type resendRequest struct {
	ChallengeID string `json:"desafio_id" binding:"required"`
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var input resendRequest
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Solicitud inválida"})
		return
	}
	challengeID, err := uuid.Parse(input.ChallengeID)
	if err != nil {
		c.JSON(http.StatusAccepted, registrationResponse(application.RegistrationResult{ChallengeID: uuid.New(), ResendAfter: h.cfg.OTPResendWait}, "Si la cuenta sigue pendiente, enviaremos otro código."))
		return
	}
	result, err := h.verification.Resend(c.Request.Context(), challengeID, clientIP(c))
	if err != nil {
		if errors.Is(err, application.ErrVerificationTooSoon) || errors.Is(err, application.ErrVerificationLimited) {
			c.JSON(http.StatusTooManyRequests, gin.H{"mensaje": err.Error()})
			return
		}
		if errors.Is(err, application.ErrEmailDelivery) {
			c.JSON(http.StatusServiceUnavailable, registrationResponse(result, "No fue posible enviar el código. Intenta nuevamente."))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible procesar la solicitud"})
		return
	}
	c.JSON(http.StatusAccepted, registrationResponse(result, "Si la cuenta sigue pendiente, enviaremos otro código."))
}

func registrationResponse(result application.RegistrationResult, message string) gin.H {
	return gin.H{
		"mensaje":              message,
		"desafio_id":           result.ChallengeID,
		"correo_enmascarado":   result.MaskedEmail,
		"reenviar_en_segundos": int(result.ResendAfter.Seconds()),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input loginRequest
	if c.ShouldBindJSON(&input) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"mensaje": "Correo y contraseña son obligatorios"})
		return
	}
	user, err := h.auth.Login(input.Correo, input.Contrasena)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"mensaje": err.Error()})
		return
	}
	session, err := h.sessions.Create(user.ID, input.Recordarme, clientIP(c), c.Request.UserAgent())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible iniciar sesión"})
		return
	}
	h.setCookies(c, session)
	c.JSON(http.StatusOK, gin.H{"usuario": gin.H{"id": user.ID, "correo": user.Correo}})
}
func (h *AuthHandler) Logout(c *gin.Context) {
	if token, err := c.Cookie(h.cfg.SessionCookieName()); err == nil {
		if err := h.sessions.Revoke(token, "logout"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible cerrar la sesión"})
			return
		}
	}
	h.clearCookies(c)
	c.Status(http.StatusNoContent)
}
func (h *AuthHandler) Me(c *gin.Context) {
	userID, idOK := transporthttp.AuthenticatedUserID(c)
	email, emailOK := transporthttp.AuthenticatedEmail(c)
	if !idOK || !emailOK {
		c.JSON(http.StatusUnauthorized, gin.H{"autenticado": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"autenticado": true, "usuario": gin.H{"id": userID, "correo": email}})
}

func (h *AuthHandler) setCookies(c *gin.Context, session application.CreatedSession) {
	secure := h.cfg.AppEnv == "production"
	maxAge := int(time.Until(session.ExpiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(h.cfg.SessionCookieName(), session.Token, maxAge, "/", "", secure, true)
	c.SetCookie(transporthttp.CSRFCookieName, session.CSRFToken, maxAge, "/", "", secure, false)
}

func (h *AuthHandler) clearCookies(c *gin.Context) {
	secure := h.cfg.AppEnv == "production"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(h.cfg.SessionCookieName(), "", -1, "/", "", secure, true)
	c.SetCookie(transporthttp.CSRFCookieName, "", -1, "/", "", secure, false)
}

func clientIP(c *gin.Context) string {
	return c.ClientIP()
}
