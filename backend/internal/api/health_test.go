package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/invulnerable/backend/internal/db"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_Health(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	handler := NewHealthHandler(database)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Health(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"healthy"`)
}

func TestHealthHandler_Ready(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	handler := NewHealthHandler(database)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Ready(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"status":"ready"`)
}
