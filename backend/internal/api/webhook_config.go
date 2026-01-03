package api

import (
	"net/http"

	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// WebhookConfigHandler handles webhook configuration API endpoints
type WebhookConfigHandler struct {
	repo   *db.WebhookConfigRepository
	logger *zap.Logger
}

// NewWebhookConfigHandler creates a new webhook config handler
func NewWebhookConfigHandler(repo *db.WebhookConfigRepository, logger *zap.Logger) *WebhookConfigHandler {
	return &WebhookConfigHandler{
		repo:   repo,
		logger: logger,
	}
}

// UpsertWebhookConfig handles PUT /api/v1/webhook-configs/:namespace/:name
func (h *WebhookConfigHandler) UpsertWebhookConfig(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "namespace and name are required")
	}

	var req models.WebhookConfigRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to bind webhook config request",
			zap.Error(err),
			zap.String("namespace", namespace),
			zap.String("name", name))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Validate required fields
	if req.WebhookURL == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "webhook_url is required")
	}

	// Set defaults
	if req.WebhookFormat == "" {
		req.WebhookFormat = "slack"
	}
	if req.ScanMinSeverity == "" {
		req.ScanMinSeverity = "High"
	}
	if req.StatusChangeMinSeverity == "" {
		req.StatusChangeMinSeverity = "High"
	}
	if req.StatusChangeTransitions == nil {
		req.StatusChangeTransitions = []string{}
	}

	err := h.repo.Upsert(c.Request().Context(), namespace, name, &req)
	if err != nil {
		h.logger.Error("failed to upsert webhook config",
			zap.Error(err),
			zap.String("namespace", namespace),
			zap.String("name", name))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to save webhook config")
	}

	h.logger.Info("webhook config upserted",
		zap.String("namespace", namespace),
		zap.String("name", name),
		zap.String("url", req.WebhookURL))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "webhook config saved successfully",
	})
}

// GetWebhookConfig handles GET /api/v1/webhook-configs/:namespace/:name
func (h *WebhookConfigHandler) GetWebhookConfig(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "namespace and name are required")
	}

	config, err := h.repo.Get(c.Request().Context(), namespace, name)
	if err != nil {
		h.logger.Error("failed to get webhook config",
			zap.Error(err),
			zap.String("namespace", namespace),
			zap.String("name", name))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get webhook config")
	}

	if config == nil {
		return echo.NewHTTPError(http.StatusNotFound, "webhook config not found")
	}

	return c.JSON(http.StatusOK, config)
}

// DeleteWebhookConfig handles DELETE /api/v1/webhook-configs/:namespace/:name
func (h *WebhookConfigHandler) DeleteWebhookConfig(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "namespace and name are required")
	}

	err := h.repo.Delete(c.Request().Context(), namespace, name)
	if err != nil {
		h.logger.Error("failed to delete webhook config",
			zap.Error(err),
			zap.String("namespace", namespace),
			zap.String("name", name))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete webhook config")
	}

	h.logger.Info("webhook config deleted",
		zap.String("namespace", namespace),
		zap.String("name", name))

	return c.JSON(http.StatusOK, map[string]string{
		"message": "webhook config deleted successfully",
	})
}
