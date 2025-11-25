package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/invulnerable/backend/internal/metrics"
	"go.uber.org/zap"
)

type MetricsHandler struct {
	logger         *zap.Logger
	metricsService *metrics.Service
}

func NewMetricsHandler(logger *zap.Logger, metricsService *metrics.Service) *MetricsHandler {
	return &MetricsHandler{
		logger:         logger,
		metricsService: metricsService,
	}
}

// GetMetrics handles GET /api/v1/metrics
func (h *MetricsHandler) GetMetrics(c echo.Context) error {
	metrics, err := h.metricsService.GetDashboardMetrics(c.Request().Context())
	if err != nil {
		h.logger.Error("failed to get metrics", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get metrics")
	}

	return c.JSON(http.StatusOK, metrics)
}
