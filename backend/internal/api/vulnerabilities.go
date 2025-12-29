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

func getUserFromHeaders(c echo.Context) string {
	// Try X-Auth-Request-Email first (more specific)
	if email := c.Request().Header.Get("X-Auth-Request-Email"); email != "" {
		return email
	}
	// Fall back to X-Auth-Request-User
	if user := c.Request().Header.Get("X-Auth-Request-User"); user != "" {
		return user
	}
	return "unknown"
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

	// Parse image_name parameter for filtering by image name
	var imageName *string
	if imageNameStr := c.QueryParam("image_name"); imageNameStr != "" {
		imageName = &imageNameStr
	}

	// Parse cve_id parameter for filtering by specific CVE
	var cveID *string
	if cveIDStr := c.QueryParam("cve_id"); cveIDStr != "" {
		cveID = &cveIDStr
	}

	// Use ListWithImageInfo to get vulnerability+image combinations for compliance
	vulns, err := h.vulnRepo.ListWithImageInfo(c.Request().Context(), limit, offset, severity, status, hasFix, imageID, imageName, cveID)
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

	// Get user from OAuth2 Proxy headers
	updatedBy := getUserFromHeaders(c)

	// Create update with context
	updateWithContext := &models.VulnerabilityUpdateWithContext{
		Status:    update.Status,
		Notes:     update.Notes,
		UpdatedBy: updatedBy,
	}

	if err := h.vulnRepo.Update(c.Request().Context(), id, updateWithContext); err != nil {
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

// BulkUpdateVulnerabilities handles PATCH /api/v1/vulnerabilities/bulk
func (h *VulnerabilityHandler) BulkUpdateVulnerabilities(c echo.Context) error {
	var req models.BulkUpdateRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if len(req.VulnerabilityIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "no vulnerability IDs provided")
	}

	if len(req.VulnerabilityIDs) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot update more than 100 vulnerabilities at once")
	}

	// Get user from headers
	updatedBy := getUserFromHeaders(c)

	updateWithContext := &models.VulnerabilityUpdateWithContext{
		Status:    req.Status,
		Notes:     req.Notes,
		UpdatedBy: updatedBy,
	}

	if err := h.vulnRepo.BulkUpdate(c.Request().Context(), req.VulnerabilityIDs, updateWithContext); err != nil {
		h.logger.Error("failed to bulk update vulnerabilities", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update vulnerabilities")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"updated_count": len(req.VulnerabilityIDs),
		"status":        req.Status,
	})
}

// GetVulnerabilityHistory handles GET /api/v1/vulnerabilities/:id/history
func (h *VulnerabilityHandler) GetVulnerabilityHistory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid vulnerability ID")
	}

	history, err := h.vulnRepo.GetHistory(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get vulnerability history", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get history")
	}

	return c.JSON(http.StatusOK, history)
}
