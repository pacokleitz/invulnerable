package api

import (
	"net/http"
	"strconv"

	"github.com/invulnerable/backend/internal/metrics"
	"github.com/labstack/echo/v4"
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
	// Parse has_fix parameter
	var hasFix *bool
	if hasFixStr := c.QueryParam("has_fix"); hasFixStr != "" {
		hasFixBool, err := strconv.ParseBool(hasFixStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid has_fix parameter")
		}
		hasFix = &hasFixBool
	}

	metrics, err := h.metricsService.GetDashboardMetrics(c.Request().Context(), hasFix)
	if err != nil {
		h.logger.Error("failed to get metrics", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get metrics")
	}

	return c.JSON(http.StatusOK, metrics)
}
