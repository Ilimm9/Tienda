package http

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"tienda/backend/internal/application"
	"tienda/backend/internal/config"
	"time"
)

type AuthHandler struct {
	auth *application.AuthService
	cfg  config.Config
}

func NewAuthHandler(auth *application.AuthService, cfg config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
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
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": user.ID.String(), "correo": user.Correo, "exp": jwt.NewNumericDate(time.Now().Add(h.cfg.JWTExpiration))})
	signed, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"mensaje": "No fue posible iniciar sesión"})
		return
	}
	maxAge := int(h.cfg.JWTExpiration.Seconds())
	if input.Recordarme {
		maxAge = int((30 * 24 * time.Hour).Seconds())
	}
	c.SetCookie("tienda_session", signed, maxAge, "/", "", h.cfg.AppEnv == "production", true)
	c.JSON(http.StatusOK, gin.H{"usuario": gin.H{"id": user.ID, "correo": user.Correo}})
}
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("tienda_session", "", -1, "/", "", h.cfg.AppEnv == "production", true)
	c.Status(http.StatusNoContent)
}
func (h *AuthHandler) Me(c *gin.Context) {
	value, err := c.Cookie("tienda_session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"autenticado": false})
		return
	}
	token, err := jwt.Parse(value, func(token *jwt.Token) (interface{}, error) { return []byte(h.cfg.JWTSecret), nil })
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"autenticado": false})
		return
	}
	claims, _ := token.Claims.(jwt.MapClaims)
	c.JSON(http.StatusOK, gin.H{"autenticado": true, "usuario": gin.H{"id": claims["sub"], "correo": claims["correo"]}})
}
