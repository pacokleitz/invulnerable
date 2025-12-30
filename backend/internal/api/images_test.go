package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/invulnerable/backend/internal/db"
	"github.com/invulnerable/backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestImageHandler_ListImages(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	imageRepo := db.NewImageRepository(database)
	handler := NewImageHandler(logger, imageRepo)

	// Create test images
	image1 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image1)
	require.NoError(t, err)

	image2 := &models.Image{
		Registry:   "docker.io",
		Repository: "library/postgres",
		Tag:        "15",
	}
	err = imageRepo.Create(context.Background(), image2)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = handler.ListImages(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "nginx")
	assert.Contains(t, rec.Body.String(), "postgres")
}

func TestImageHandler_ListImages_WithLimit(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	imageRepo := db.NewImageRepository(database)
	handler := NewImageHandler(logger, imageRepo)

	// Create test images
	for i := 0; i < 5; i++ {
		image := &models.Image{
			Registry:   "docker.io",
			Repository: "library/test",
			Tag:        string(rune('a' + i)),
		}
		err := imageRepo.Create(context.Background(), image)
		require.NoError(t, err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images?limit=2", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListImages(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestImageHandler_ListImages_InvalidHasFix(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	imageRepo := db.NewImageRepository(database)
	handler := NewImageHandler(logger, imageRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images?has_fix=invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListImages(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "invalid has_fix parameter")
}

func TestImageHandler_GetImageHistory(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	imageRepo := db.NewImageRepository(database)
	scanRepo := db.NewScanRepository(database)
	handler := NewImageHandler(logger, imageRepo)

	// Create test image
	image := &models.Image{
		Registry:   "docker.io",
		Repository: "library/nginx",
		Tag:        "latest",
	}
	err := imageRepo.Create(context.Background(), image)
	require.NoError(t, err)

	// Create scans
	for i := 0; i < 3; i++ {
		scan := &models.Scan{
			ImageID:     image.ID,
			ScanDate:    time.Now().Add(time.Duration(-i) * time.Hour),
			Status:      "completed",
			SLACritical: 7,
			SLAHigh:     30,
			SLAMedium:   90,
			SLALow:      180,
		}
		err = scanRepo.Create(context.Background(), scan)
		require.NoError(t, err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+string(rune(image.ID+48))+"/history", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(string(rune(image.ID + 48)))

	err = handler.GetImageHistory(c)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestImageHandler_GetImageHistory_InvalidID(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	imageRepo := db.NewImageRepository(database)
	handler := NewImageHandler(logger, imageRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/invalid/history", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := handler.GetImageHistory(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "invalid image ID")
}

func TestImageHandler_GetImageHistory_InvalidHasFix(t *testing.T) {
	database := db.SetupTestDatabase(t)
	defer database.Close()

	logger := zap.NewNop()
	imageRepo := db.NewImageRepository(database)
	handler := NewImageHandler(logger, imageRepo)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/1/history?has_fix=notabool", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := handler.GetImageHistory(c)
	require.Error(t, err)

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	assert.Contains(t, httpErr.Message, "invalid has_fix parameter")
}
