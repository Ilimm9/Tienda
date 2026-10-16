package plataforma

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	application "tienda/backend/internal/application/plataforma"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Resumen(c *gin.Context) {
	desde, hasta, ok := periodo(c)
	if !ok {
		return
	}
	item, err := h.service.Resumen(c.Request.Context(), desde, hasta)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListarNegocios(c *gin.Context) {
	page := entero(c.Query("page"), 1)
	pageSize := entero(c.Query("page_size"), 20)
	items, err := h.service.ListarNegocios(c.Request.Context(), c.Query("q"), c.Query("estado"), page, pageSize)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) ObtenerNegocio(c *gin.Context) {
	id, ok := negocioID(c)
	if !ok {
		return
	}
	desde, hasta, ok := periodo(c)
	if !ok {
		return
	}
	item, err := h.service.ObtenerNegocio(c.Request.Context(), id, desde, hasta)
	if errors.Is(err, application.ErrNegocioNoEncontrado) {
		c.JSON(http.StatusNotFound, gin.H{"codigo": "NEGOCIO_NO_ENCONTRADO", "mensaje": "El negocio no existe.", "campos": gin.H{}})
		return
	}
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListarUsuarios(c *gin.Context) {
	id, ok := negocioID(c)
	if !ok {
		return
	}
	items, err := h.service.ListarUsuarios(c.Request.Context(), id)
	if err != nil {
		internalError(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": len(items)})
}

func negocioID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("negocioId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "IDENTIFICADOR_INVALIDO", "mensaje": "El identificador del negocio no es válido.", "campos": gin.H{"negocioId": "debe ser UUID"}})
		return uuid.Nil, false
	}
	return id, true
}

func periodo(c *gin.Context) (time.Time, time.Time, bool) {
	desde, ok := fecha(c, "desde")
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	hasta, ok := fecha(c, "hasta")
	if !ok {
		return time.Time{}, time.Time{}, false
	}
	if !desde.IsZero() && !hasta.IsZero() && desde.After(hasta) {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "PERIODO_INVALIDO", "mensaje": "La fecha desde no puede ser posterior a hasta.", "campos": gin.H{"desde": "periodo inválido"}})
		return time.Time{}, time.Time{}, false
	}
	return desde, hasta, true
}

func fecha(c *gin.Context, key string) (time.Time, bool) {
	value := c.Query(key)
	if value == "" {
		return time.Time{}, true
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"codigo": "FECHA_INVALIDA", "mensaje": "La fecha debe usar formato YYYY-MM-DD.", "campos": gin.H{key: "formato inválido"}})
		return time.Time{}, false
	}
	return parsed, true
}

func entero(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func internalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{"codigo": "ERROR_INTERNO", "mensaje": "No fue posible completar la operación.", "campos": gin.H{}})
}
