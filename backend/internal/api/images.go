package api

import (
	"net/http"
	"strconv"

	"github.com/invulnerable/backend/internal/db"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ImageHandler struct {
	logger    *zap.Logger
	imageRepo *db.ImageRepository
}

func NewImageHandler(logger *zap.Logger, imageRepo *db.ImageRepository) *ImageHandler {
	return &ImageHandler{
		logger:    logger,
		imageRepo: imageRepo,
	}
}

// ListImages handles GET /api/v1/images
func (h *ImageHandler) ListImages(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	// Parse has_fix parameter
	var hasFix *bool
	if hasFixStr := c.QueryParam("has_fix"); hasFixStr != "" {
		hasFixBool, err := strconv.ParseBool(hasFixStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid has_fix parameter")
		}
		hasFix = &hasFixBool
	}

	images, err := h.imageRepo.List(c.Request().Context(), limit, offset, hasFix)
	if err != nil {
		h.logger.Error("failed to list images", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list images")
	}

	return c.JSON(http.StatusOK, images)
}

// GetImageHistory handles GET /api/v1/images/:id/history
func (h *ImageHandler) GetImageHistory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid image ID")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Parse has_fix parameter
	var hasFix *bool
	if hasFixStr := c.QueryParam("has_fix"); hasFixStr != "" {
		hasFixBool, err := strconv.ParseBool(hasFixStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid has_fix parameter")
		}
		hasFix = &hasFixBool
	}

	scans, err := h.imageRepo.GetScanHistory(c.Request().Context(), id, limit, hasFix)
	if err != nil {
		h.logger.Error("failed to get image history", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get image history")
	}

	return c.JSON(http.StatusOK, scans)
}
