package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/invulnerable/backend/internal/analyzer"
	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
	"github.com/invulnerable/backend/internal/notifier"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ScanHandler struct {
	logger    *zap.Logger
	imageRepo *db.ImageRepository
	scanRepo  *db.ScanRepository
	vulnRepo  *db.VulnerabilityRepository
	sbomRepo  *db.SBOMRepository
	analyzer  *analyzer.Analyzer
	notifier  *notifier.Notifier
}

func NewScanHandler(
	logger *zap.Logger,
	imageRepo *db.ImageRepository,
	scanRepo *db.ScanRepository,
	vulnRepo *db.VulnerabilityRepository,
	sbomRepo *db.SBOMRepository,
	analyzer *analyzer.Analyzer,
	notifier *notifier.Notifier,
) *ScanHandler {
	return &ScanHandler{
		logger:    logger,
		imageRepo: imageRepo,
		scanRepo:  scanRepo,
		vulnRepo:  vulnRepo,
		sbomRepo:  sbomRepo,
		analyzer:  analyzer,
		notifier:  notifier,
	}
}

type ScanRequest struct {
	Image            string                   `json:"image"`
	ImageDigest      *string                  `json:"image_digest,omitempty"`
	GrypeResult      models.GrypeResult       `json:"grype_result"`
	SBOM             json.RawMessage          `json:"sbom"`
	SBOMFormat       string                   `json:"sbom_format"`
	SBOMVersion      *string                  `json:"sbom_version,omitempty"`
	WebhookConfig    *WebhookConfig           `json:"webhook_config,omitempty"`
	SLAConfig        *SLAConfig               `json:"sla_config,omitempty"`
	ImageScanContext *models.ImageScanContext `json:"imagescan_context,omitempty"`
}

type WebhookConfig struct {
	URL         string `json:"url"`
	Format      string `json:"format"`
	MinSeverity string `json:"min_severity"`
	OnlyFixable bool   `json:"only_fixable"`
}

type SLAConfig struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
}

// CreateScan handles POST /api/v1/scans - receives scan results from CronJob
func (h *ScanHandler) CreateScan(c echo.Context) error {
	var req ScanRequest
	if err := c.Bind(&req); err != nil {
		h.logger.Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	// Log ImageScan context for debugging
	if req.ImageScanContext != nil {
		h.logger.Info("received scan with ImageScan context",
			zap.String("namespace", req.ImageScanContext.Namespace),
			zap.String("name", req.ImageScanContext.Name),
			zap.String("image", req.Image))
	} else {
		h.logger.Warn("received scan WITHOUT ImageScan context",
			zap.String("image", req.Image))
	}

	ctx := c.Request().Context()

	// Parse image name (registry/repository:tag)
	registry, repository, tag := parseImageName(req.Image)

	// Get or create image (Create method handles upsert with digest update)
	image := &models.Image{
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
		Digest:     req.ImageDigest,
	}
	if err := h.imageRepo.Create(ctx, image); err != nil {
		h.logger.Error("failed to create/update image", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create/update image")
	}

	// Create scan record
	syftVersion := "unknown"
	grypeVersion := req.GrypeResult.Descriptor.Version

	// Set SLA values with defaults
	slaCritical := 7
	slaHigh := 30
	slaMedium := 90
	slaLow := 180
	if req.SLAConfig != nil {
		slaCritical = req.SLAConfig.Critical
		slaHigh = req.SLAConfig.High
		slaMedium = req.SLAConfig.Medium
		slaLow = req.SLAConfig.Low
	}

	scan := &models.Scan{
		ImageID:      image.ID,
		ScanDate:     time.Now(),
		SyftVersion:  &syftVersion,
		GrypeVersion: &grypeVersion,
		Status:       "completed",
		SLACritical:  slaCritical,
		SLAHigh:      slaHigh,
		SLAMedium:    slaMedium,
		SLALow:       slaLow,
	}

	// Add ImageScan context if provided
	if req.ImageScanContext != nil {
		scan.ImageScanNamespace = &req.ImageScanContext.Namespace
		scan.ImageScanName = &req.ImageScanContext.Name
	}

	if err := h.scanRepo.Create(ctx, scan); err != nil {
		h.logger.Error("failed to create scan", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create scan")
	}

	// Store SBOM
	sbom := &models.SBOM{
		ScanID:  scan.ID,
		Format:  req.SBOMFormat,
		Version: req.SBOMVersion,
	}

	if err := h.sbomRepo.Create(ctx, sbom, []byte(req.SBOM)); err != nil {
		h.logger.Error("failed to create SBOM", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create SBOM")
	}

	// Track which vulnerabilities we've already reverted in this scan to avoid duplicates
	revertedVulns := make(map[string]bool)

	// Process vulnerabilities
	for _, match := range req.GrypeResult.Matches {
		// Determine fix version
		var fixVersion *string
		if match.Vulnerability.Fix != nil && len(match.Vulnerability.Fix.Versions) > 0 {
			fv := match.Vulnerability.Fix.Versions[0]
			fixVersion = &fv
		}

		// Get primary URL
		var url *string
		if len(match.Vulnerability.URLs) > 0 {
			url = &match.Vulnerability.URLs[0]
		}

		vuln := &models.Vulnerability{
			CVEID:           match.Vulnerability.ID,
			PackageName:     match.Artifact.Name,
			PackageVersion:  match.Artifact.Version,
			PackageType:     &match.Artifact.Type,
			Severity:        normalizeSeverity(match.Vulnerability.Severity),
			FixVersion:      fixVersion,
			URL:             url,
			Description:     &match.Vulnerability.Description,
			Status:          "active",
			FirstDetectedAt: time.Now(),
			LastSeenAt:      time.Now(),
		}

		// Add ImageScan context if provided (will update on every scan via Upsert)
		if req.ImageScanContext != nil {
			vuln.ImageScanNamespace = &req.ImageScanContext.Namespace
			vuln.ImageScanName = &req.ImageScanContext.Name
			h.logger.Debug("assigning ImageScan context to vulnerability",
				zap.String("cve_id", vuln.CVEID),
				zap.String("namespace", req.ImageScanContext.Namespace),
				zap.String("name", req.ImageScanContext.Name))
		} else {
			h.logger.Warn("no ImageScan context to assign to vulnerability",
				zap.String("cve_id", vuln.CVEID))
		}

		// Check if vulnerability already exists
		existing, err := h.vulnRepo.GetByUniqueKey(ctx, vuln.CVEID, vuln.PackageName, vuln.PackageVersion)
		if err != nil {
			h.logger.Error("failed to check existing vulnerability", zap.Error(err))
			continue
		}

		if existing != nil {
			// Update last_seen_at
			vuln.ID = existing.ID
			vuln.FirstDetectedAt = existing.FirstDetectedAt

			// Log existing vulnerability status for debugging
			updatedByStr := "nil"
			if existing.UpdatedBy != nil {
				updatedByStr = *existing.UpdatedBy
			}
			h.logger.Debug("processing existing vulnerability",
				zap.String("cve_id", vuln.CVEID),
				zap.String("package", vuln.PackageName),
				zap.String("current_status", existing.Status),
				zap.String("updated_by", updatedByStr))

			// Check if this CVE was manually marked as fixed
			// If so, revert it back to active since it's still being detected
			// We revert if: status is "fixed" AND (updated_by is NULL OR updated_by is not "system")
			// This handles both manually fixed CVEs and CVEs fixed before the audit migration

			// Create unique key for this vulnerability (cve_id + package_name + package_version)
			vulnKey := fmt.Sprintf("%s|%s|%s", vuln.CVEID, vuln.PackageName, vuln.PackageVersion)

			shouldRevert := existing.Status == models.StatusFixed &&
				(existing.UpdatedBy == nil || *existing.UpdatedBy != "system") &&
				!revertedVulns[vulnKey] // Only revert once per scan

			if shouldRevert {
				h.logger.Info("reverting fixed CVE back to active (still detected in scan)",
					zap.String("cve_id", vuln.CVEID),
					zap.String("package", vuln.PackageName),
					zap.String("previous_updated_by", updatedByStr))

				// Revert status to active
				newStatus := models.StatusActive
				updateCtx := &models.VulnerabilityUpdateWithContext{
					Status:    &newStatus,
					UpdatedBy: "system",
				}

				if err := h.vulnRepo.Update(ctx, existing.ID, updateCtx); err != nil {
					h.logger.Error("failed to revert manually fixed CVE",
						zap.Error(err),
						zap.Int("vuln_id", existing.ID))
				} else {
					// Mark as reverted to prevent duplicate history entries
					revertedVulns[vulnKey] = true
				}
				// Note: Update() method already creates history entry, no need to duplicate
			}
		}

		if err := h.vulnRepo.Upsert(ctx, vuln); err != nil {
			h.logger.Error("failed to upsert vulnerability", zap.Error(err))
			continue
		}

		// Link vulnerability to scan
		if err := h.vulnRepo.LinkToScan(ctx, scan.ID, vuln.ID); err != nil {
			h.logger.Error("failed to link vulnerability to scan", zap.Error(err))
		}
	}

	// Send webhook notification if configured
	if req.WebhookConfig != nil && req.WebhookConfig.URL != "" {
		go func() {
			// Get all vulnerabilities for this scan with their current status from the database
			// This allows us to filter out ignored/accepted CVEs
			vulnsFromDB, err := h.scanRepo.GetVulnerabilities(context.Background(), scan.ID)
			if err != nil {
				h.logger.Error("failed to get vulnerabilities for webhook notification",
					zap.Error(err),
					zap.Int("scan_id", scan.ID))
				return
			}

			// Create a map of CVE+Package to status for quick lookup
			vulnStatusMap := make(map[string]string)
			for _, v := range vulnsFromDB {
				key := fmt.Sprintf("%s|%s|%s", v.CVEID, v.PackageName, v.PackageVersion)
				vulnStatusMap[key] = v.Status
			}

			// Filter matches to exclude ignored/accepted CVEs
			matchesToNotify := []models.GrypeMatch{}
			for _, match := range req.GrypeResult.Matches {
				key := fmt.Sprintf("%s|%s|%s", match.Vulnerability.ID, match.Artifact.Name, match.Artifact.Version)
				status, exists := vulnStatusMap[key]

				// Skip if CVE is ignored or accepted (triaged as not actionable)
				if exists && (status == models.StatusIgnored || status == models.StatusAccepted) {
					continue
				}

				// Apply onlyFixable filter if configured
				if req.WebhookConfig.OnlyFixable && len(match.Vulnerability.Fix.Versions) == 0 {
					continue
				}

				matchesToNotify = append(matchesToNotify, match)
			}

			// Calculate severity counts for notification (only for actionable vulnerabilities)
			severityCounts := notifier.SeverityCounts{}
			for _, match := range matchesToNotify {
				switch match.Vulnerability.Severity {
				case "Critical":
					severityCounts.Critical++
				case "High":
					severityCounts.High++
				case "Medium":
					severityCounts.Medium++
				case "Low":
					severityCounts.Low++
				default:
					severityCounts.Negligible++
				}
			}

			webhookConfig := notifier.WebhookConfig{
				URL:         req.WebhookConfig.URL,
				Format:      req.WebhookConfig.Format,
				MinSeverity: req.WebhookConfig.MinSeverity,
				OnlyFixable: req.WebhookConfig.OnlyFixable,
			}

			notificationPayload := notifier.NotificationPayload{
				Image:          req.Image,
				ImageDigest:    req.ImageDigest,
				ScanID:         scan.ID,
				TotalVulns:     len(matchesToNotify),
				SeverityCounts: severityCounts,
			}

			if err := h.notifier.SendNotification(context.Background(), webhookConfig, notificationPayload); err != nil {
				h.logger.Error("failed to send webhook notification",
					zap.Error(err),
					zap.String("webhook_url", req.WebhookConfig.URL),
					zap.Int("scan_id", scan.ID))
			} else if len(matchesToNotify) == 0 {
				h.logger.Info("no actionable vulnerabilities to notify about (all ignored/accepted)",
					zap.Int("scan_id", scan.ID),
					zap.Int("total_detected", len(req.GrypeResult.Matches)),
					zap.Int("actionable", 0))
			}
		}()
	}

	h.logger.Info("scan created successfully",
		zap.Int("scan_id", scan.ID),
		zap.String("image", req.Image),
		zap.Int("vulnerabilities", len(req.GrypeResult.Matches)))

	return c.JSON(http.StatusCreated, scan)
}

// ListScans handles GET /api/v1/scans
func (h *ScanHandler) ListScans(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	var imageID *int
	if imageIDStr := c.QueryParam("image_id"); imageIDStr != "" {
		id, err := strconv.Atoi(imageIDStr)
		if err == nil {
			imageID = &id
		}
	}

	var imageName *string
	if imageNameStr := c.QueryParam("image"); imageNameStr != "" {
		imageName = &imageNameStr
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

	// Get total count
	total, err := h.scanRepo.Count(c.Request().Context(), imageID, imageName)
	if err != nil {
		h.logger.Error("failed to count scans", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to count scans")
	}

	scans, err := h.scanRepo.List(c.Request().Context(), limit, offset, imageID, imageName, hasFix)
	if err != nil {
		h.logger.Error("failed to list scans", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list scans")
	}

	response := map[string]interface{}{
		"data":   scans,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	return c.JSON(http.StatusOK, response)
}

// GetScan handles GET /api/v1/scans/:id
func (h *ScanHandler) GetScan(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid scan ID")
	}

	// Parse has_fix parameter for counts
	var hasFix *bool
	if hasFixStr := c.QueryParam("has_fix"); hasFixStr != "" {
		hasFixBool, err := strconv.ParseBool(hasFixStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid has_fix parameter")
		}
		hasFix = &hasFixBool
	}

	scan, err := h.scanRepo.GetWithDetails(c.Request().Context(), id, hasFix)
	if err != nil {
		h.logger.Error("failed to get scan", zap.Error(err))
		return echo.NewHTTPError(http.StatusNotFound, "scan not found")
	}

	// Get vulnerabilities for this scan
	vulns, err := h.scanRepo.GetVulnerabilities(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get vulnerabilities", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get vulnerabilities")
	}

	response := map[string]interface{}{
		"scan":            scan,
		"vulnerabilities": vulns,
	}

	return c.JSON(http.StatusOK, response)
}

// GetSBOM handles GET /api/v1/scans/:id/sbom
func (h *ScanHandler) GetSBOM(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid scan ID")
	}

	document, err := h.sbomRepo.GetDocumentByScanID(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get SBOM", zap.Error(err))
		return echo.NewHTTPError(http.StatusNotFound, "SBOM not found")
	}

	return c.JSONBlob(http.StatusOK, document)
}

// GetScanDiff handles GET /api/v1/scans/:id/diff
func (h *ScanHandler) GetScanDiff(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid scan ID")
	}

	diff, err := h.analyzer.CompareScan(c.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to compare scan", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to compare scan")
	}

	return c.JSON(http.StatusOK, diff)
}

// Helper functions

func parseImageName(fullName string) (registry, repository, tag string) {
	// Default tag
	tag = "latest"
	repoPath := fullName

	// Find the last '/' to separate registry/repo from tag
	lastSlash := strings.LastIndex(fullName, "/")

	// Find the last ':' after the last '/' (if any)
	// This is the tag separator, not a port separator
	tagSeparator := strings.LastIndex(fullName, ":")
	if tagSeparator > lastSlash {
		// There's a tag
		tag = fullName[tagSeparator+1:]
		repoPath = fullName[:tagSeparator]
	}

	// Split registry/repository
	slashParts := strings.Split(repoPath, "/")

	if len(slashParts) == 1 {
		// No registry, just repository
		registry = "docker.io"
		repository = slashParts[0]
	} else if strings.Contains(slashParts[0], ".") || strings.Contains(slashParts[0], ":") {
		// Has registry (either has a dot like gcr.io, or has a port like localhost:5000)
		registry = slashParts[0]
		repository = strings.Join(slashParts[1:], "/")
	} else {
		// Docker Hub format (e.g., library/nginx)
		registry = "docker.io"
		repository = repoPath
	}

	return
}

func normalizeSeverity(severity string) string {
	severity = strings.ToUpper(severity)
	switch severity {
	case "CRITICAL":
		return "Critical"
	case "HIGH":
		return "High"
	case "MEDIUM":
		return "Medium"
	case "LOW":
		return "Low"
	default:
		return "Unknown"
	}
}
