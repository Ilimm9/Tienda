package cuenta

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	application "tienda/backend/internal/application/cuenta"
	"tienda/backend/internal/config"
	transporthttp "tienda/backend/internal/interfaces/http"
	"time"
)

type AuthHandler struct {
	auth     *application.AuthService
	sessions *application.SessionService
	cfg      config.Config
}

func NewAuthHandler(auth *application.AuthService, sessions *application.SessionService, cfg config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, sessions: sessions, cfg: cfg}
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
	if err := h.auth.Register(input.NombreCompleto, input.Correo, input.Telefono, input.Contrasena); err != nil {
		if err == application.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"mensaje": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible crear la cuenta"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"mensaje": "Cuenta creada correctamente"})
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
	session, err := h.sessions.Create(user.ID, input.Recordarme, requestIP(c.Request), c.Request.UserAgent())
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

func requestIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return host
	}
	return request.RemoteAddr
}
