package api

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/invulnerable/backend/internal/auth"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger       *zap.Logger
	jwtValidator *auth.JWTValidator
	oauthEnabled bool // Whether OAuth2 Proxy is enabled in deployment
}

func NewUserHandler(logger *zap.Logger, jwtValidator *auth.JWTValidator, oauthEnabled bool) *UserHandler {
	return &UserHandler{
		logger:       logger,
		jwtValidator: jwtValidator,
		oauthEnabled: oauthEnabled,
	}
}

type UserResponse struct {
	Email    string `json:"email"`
	Username string `json:"username,omitempty"`
}

// GetCurrentUser handles GET /api/v1/user/me - returns current authenticated user info
func (h *UserHandler) GetCurrentUser(c echo.Context) error {
	// OAuth2 Proxy injects user information in headers
	email := c.Request().Header.Get("X-Auth-Request-Email")
	username := c.Request().Header.Get("X-Auth-Request-User")
	accessToken := c.Request().Header.Get("X-Auth-Request-Access-Token")

	// If OAuth is not enabled, return 204 (no user info available)
	if !h.oauthEnabled {
		h.logger.Debug("OAuth2 not enabled - no authentication required")
		return c.NoContent(http.StatusNoContent)
	}

	// OAuth is enabled - authentication is REQUIRED
	// If no email header, OAuth2 Proxy is misconfigured
	if email == "" {
		h.logger.Warn("OAuth2 enabled but no authentication headers present",
			zap.String("remote_addr", c.RealIP()))
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	// Security: When OAuth is enabled, JWT validation is MANDATORY
	if h.jwtValidator == nil {
		h.logger.Error("OAuth2 enabled but JWT validator not configured - this is a configuration error",
			zap.String("email", email),
			zap.String("remote_addr", c.RealIP()))
		return echo.NewHTTPError(http.StatusInternalServerError, "authentication not properly configured")
	}

	// Validate access token cryptographically
	// Validation MUST succeed - no fallback!
	token, err := h.jwtValidator.ValidateToken(accessToken)
	if err != nil {
		h.logger.Warn("invalid access token",
			zap.Error(err),
			zap.String("email", email),
			zap.String("remote_addr", c.RealIP()))
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid access token")
	}

	// Extract email from token and verify it matches header
	// This provides additional defense against header injection
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if tokenEmail, ok := claims["email"].(string); ok && tokenEmail != email {
			h.logger.Warn("email mismatch between token and header",
				zap.String("token_email", tokenEmail),
				zap.String("header_email", email),
				zap.String("remote_addr", c.RealIP()))
			return echo.NewHTTPError(http.StatusUnauthorized, "email mismatch")
		}
	}

	h.logger.Debug("user authenticated via JWT token",
		zap.String("email", email),
		zap.String("username", username))

	return c.JSON(http.StatusOK, UserResponse{
		Email:    email,
		Username: username,
	})
}
