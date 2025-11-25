package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/invulnerable/backend/internal/db"
)

type HealthHandler struct {
	db *db.Database
}

func NewHealthHandler(db *db.Database) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health handles GET /health
func (h *HealthHandler) Health(c echo.Context) error {
	if err := h.db.Health(c.Request().Context()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Ready handles GET /ready
func (h *HealthHandler) Ready(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}
