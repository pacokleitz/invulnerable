package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type UserHandler struct {
	logger *zap.Logger
}

func NewUserHandler(logger *zap.Logger) *UserHandler {
	return &UserHandler{
		logger: logger,
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

	// If no email header, OAuth2 Proxy is not deployed (no authentication)
	if email == "" {
		h.logger.Debug("no X-Auth-Request-Email header found, OAuth2 Proxy not deployed")
		return c.NoContent(http.StatusNoContent)
	}

	h.logger.Debug("user info retrieved from OAuth2 headers",
		zap.String("email", email),
		zap.String("username", username))

	return c.JSON(http.StatusOK, UserResponse{
		Email:    email,
		Username: username,
	})
}
