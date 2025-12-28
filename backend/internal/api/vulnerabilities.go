package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
	"go.uber.org/zap"
)

type VulnerabilityHandler struct {
	logger   *zap.Logger
	vulnRepo *db.VulnerabilityRepository
}

func NewVulnerabilityHandler(logger *zap.Logger, vulnRepo *db.VulnerabilityRepository) *VulnerabilityHandler {
	return &VulnerabilityHandler{
		logger:   logger,
		vulnRepo: vulnRepo,
	}
}

// ListVulnerabilities handles GET /api/v1/vulnerabilities
// Returns vulnerabilities with image context for compliance tracking
func (h *VulnerabilityHandler) ListVulnerabilities(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	var severity, status *string
	if s := c.QueryParam("severity"); s != "" {
		severity = &s
	}
	if st := c.QueryParam("status"); st != "" {
		status = &st
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

	// Parse image_id parameter for filtering by image
	var imageID *int
	if imageIDStr := c.QueryParam("image_id"); imageIDStr != "" {
		id, err := strconv.Atoi(imageIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid image_id parameter")
		}
		imageID = &id
	}

	// Use ListWithImageInfo to get vulnerability+image combinations for compliance
	vulns, err := h.vulnRepo.ListWithImageInfo(c.Request().Context(), limit, offset, severity, status, hasFix, imageID)
	if err != nil {
		h.logger.Error("failed to list vulnerabilities", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list vulnerabilities")
	}

	return c.JSON(http.StatusOK, vulns)
}

// GetVulnerabilityByCVE handles GET /api/v1/vulnerabilities/:cve
func (h *VulnerabilityHandler) GetVulnerabilityByCVE(c echo.Context) error {
	cveID := c.Param("cve")

	vulns, err := h.vulnRepo.GetByCVE(c.Request().Context(), cveID)
	if err != nil {
		h.logger.Error("failed to get vulnerability", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get vulnerability")
	}

	if len(vulns) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "vulnerability not found")
	}

	return c.JSON(http.StatusOK, vulns)
}

// UpdateVulnerability handles PATCH /api/v1/vulnerabilities/:id
func (h *VulnerabilityHandler) UpdateVulnerability(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid vulnerability ID")
	}

	var update models.VulnerabilityUpdate
	if err := c.Bind(&update); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.vulnRepo.Update(c.Request().Context(), id, &update); err != nil {
		h.logger.Error("failed to update vulnerability", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update vulnerability")
	}

	// Get updated vulnerability
	vuln, err := h.vulnRepo.GetByID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get updated vulnerability", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get vulnerability")
	}

	return c.JSON(http.StatusOK, vuln)
}
