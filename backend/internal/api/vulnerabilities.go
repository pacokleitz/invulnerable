package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
	"github.com/invulnerable/backend/internal/notifier"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type VulnerabilityHandler struct {
	logger            *zap.Logger
	vulnRepo          *db.VulnerabilityRepository
	notifier          *notifier.Notifier
	webhookConfigRepo *db.WebhookConfigRepository
}

func NewVulnerabilityHandler(
	logger *zap.Logger,
	vulnRepo *db.VulnerabilityRepository,
	notifier *notifier.Notifier,
	webhookConfigRepo *db.WebhookConfigRepository,
) *VulnerabilityHandler {
	return &VulnerabilityHandler{
		logger:            logger,
		vulnRepo:          vulnRepo,
		notifier:          notifier,
		webhookConfigRepo: webhookConfigRepo,
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

	// Get total count
	total, err := h.vulnRepo.CountWithImageInfo(c.Request().Context(), severity, status, hasFix, imageID, imageName, cveID)
	if err != nil {
		h.logger.Error("failed to count vulnerabilities", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count vulnerabilities")
	}

	// Use ListWithImageInfo to get vulnerability+image combinations for compliance
	vulns, err := h.vulnRepo.ListWithImageInfo(c.Request().Context(), limit, offset, severity, status, hasFix, imageID, imageName, cveID)
	if err != nil {
		h.logger.Error("failed to list vulnerabilities", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list vulnerabilities")
	}

	response := map[string]interface{}{
		"data":   vulns,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	return c.JSON(http.StatusOK, response)
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

	// Send status change webhook notification in background (if applicable)
	go h.sendStatusChangeWebhook(context.Background(), id, updatedBy)

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

	// Send status change webhook notifications for each vulnerability in background
	for _, vulnID := range req.VulnerabilityIDs {
		go h.sendStatusChangeWebhook(context.Background(), vulnID, updatedBy)
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

// sendStatusChangeWebhook sends webhook notification for vulnerability status changes
// This runs in a background goroutine and logs errors without failing the request
func (h *VulnerabilityHandler) sendStatusChangeWebhook(ctx context.Context, vulnID int, changedBy string) {
	// Get ImageScan context for this vulnerability
	imageScanCtx, err := h.vulnRepo.GetImageScanInfoForWebhook(ctx, vulnID)
	if err != nil {
		h.logger.Error("failed to get ImageScan info for webhook",
			zap.Error(err),
			zap.Int("vulnerability_id", vulnID))
		return
	}

	// If no ImageScan context, skip webhook (vulnerability wasn't discovered by an ImageScan)
	if imageScanCtx == nil {
		h.logger.Debug("no ImageScan context for vulnerability, skipping webhook",
			zap.Int("vulnerability_id", vulnID))
		return
	}

	// Get webhook config from database
	webhookConfig, err := h.webhookConfigRepo.Get(ctx, imageScanCtx.Namespace, imageScanCtx.Name)
	if err != nil {
		h.logger.Error("failed to get webhook config",
			zap.Error(err),
			zap.String("namespace", imageScanCtx.Namespace),
			zap.String("name", imageScanCtx.Name))
		return
	}

	// If no config or status change webhooks not enabled, skip
	if webhookConfig == nil || !webhookConfig.StatusChangeEnabled {
		h.logger.Debug("status change webhooks not enabled, skipping",
			zap.String("namespace", imageScanCtx.Namespace),
			zap.String("name", imageScanCtx.Name))
		return
	}

	// Get vulnerability details
	vuln, err := h.vulnRepo.GetByID(ctx, vulnID)
	if err != nil {
		h.logger.Error("failed to get vulnerability for webhook",
			zap.Error(err),
			zap.Int("vulnerability_id", vulnID))
		return
	}

	// Get history to find old status
	history, err := h.vulnRepo.GetHistory(ctx, vulnID)
	if err != nil {
		h.logger.Error("failed to get vulnerability history for webhook",
			zap.Error(err),
			zap.Int("vulnerability_id", vulnID))
		return
	}

	// Find previous status from history
	var oldStatus string
	if len(history) >= 2 {
		// Most recent entry is the current status, second most recent is old status
		if history[1].NewValue != nil {
			oldStatus = *history[1].NewValue
		} else {
			oldStatus = "active" // Default if no value
		}
	} else {
		// If only one history entry, it was the initial status
		oldStatus = "active" // Default initial status
	}

	// Get a representative image name for this vulnerability
	imageName, err := h.vulnRepo.GetImageNameForVulnerability(ctx, vulnID)
	if err != nil {
		// If no image found, use a placeholder
		imageName = "unknown"
		h.logger.Warn("could not find image for vulnerability",
			zap.Int("vulnerability_id", vulnID),
			zap.Error(err))
	}

	// Build notification payload
	payload := notifier.StatusChangeNotificationPayload{
		CVEID:           vuln.CVEID,
		PackageName:     vuln.PackageName,
		PackageVersion:  vuln.PackageVersion,
		Severity:        vuln.Severity,
		FixVersion:      vuln.FixVersion,
		OldStatus:       oldStatus,
		NewStatus:       vuln.Status,
		ChangedBy:       changedBy,
		Notes:           vuln.Notes,
		ImageName:       imageName,
		VulnerabilityID: vulnID,
		VulnURL:         "", // Will be constructed by notifier if frontend URL is configured
		Timestamp:       time.Now(),
	}

	// Build webhook config for notifier
	notifierConfig := notifier.StatusChangeWebhookConfig{
		URL:                webhookConfig.WebhookURL,
		Format:             webhookConfig.WebhookFormat,
		MinSeverity:        webhookConfig.StatusChangeMinSeverity,
		OnlyFixable:        webhookConfig.StatusChangeOnlyFixable,
		StatusTransitions:  webhookConfig.StatusChangeTransitions,
		IncludeNoteChanges: webhookConfig.StatusChangeIncludeNotes,
	}

	// Log what we're about to send
	h.logger.Info("attempting to send status change webhook",
		zap.Int("vulnerability_id", vulnID),
		zap.String("cve_id", vuln.CVEID),
		zap.String("old_status", oldStatus),
		zap.String("new_status", vuln.Status),
		zap.String("severity", vuln.Severity),
		zap.String("transition", fmt.Sprintf("%s→%s", oldStatus, vuln.Status)),
		zap.String("webhook_url", webhookConfig.WebhookURL))

	// Send notification (filtering happens in notifier)
	if err := h.notifier.SendStatusChangeNotification(ctx, notifierConfig, payload); err != nil {
		h.logger.Error("failed to send status change webhook",
			zap.Error(err),
			zap.Int("vulnerability_id", vulnID),
			zap.String("webhook_url", webhookConfig.WebhookURL))
		return
	}

	h.logger.Info("status change webhook sent successfully",
		zap.Int("vulnerability_id", vulnID),
		zap.String("cve_id", vuln.CVEID),
		zap.String("old_status", oldStatus),
		zap.String("new_status", vuln.Status))
}
