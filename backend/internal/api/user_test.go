package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestUserHandler_GetCurrentUser_WithAuthHeaders(t *testing.T) {
	logger := zap.NewNop()
	handler := NewUserHandler(logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	req.Header.Set("X-Auth-Request-Email", "test@example.com")
	req.Header.Set("X-Auth-Request-User", "testuser")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetCurrentUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"email":"test@example.com"`)
	assert.Contains(t, rec.Body.String(), `"username":"testuser"`)
}

func TestUserHandler_GetCurrentUser_WithoutAuthHeaders(t *testing.T) {
	logger := zap.NewNop()
	handler := NewUserHandler(logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetCurrentUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestUserHandler_GetCurrentUser_OnlyEmail(t *testing.T) {
	logger := zap.NewNop()
	handler := NewUserHandler(logger)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	req.Header.Set("X-Auth-Request-Email", "test@example.com")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetCurrentUser(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"email":"test@example.com"`)
}
