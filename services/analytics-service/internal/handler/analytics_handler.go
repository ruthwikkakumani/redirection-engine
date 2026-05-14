package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ruthwikkakumani/redirection-engine/services/analytics-service/internal/service"
	"go.uber.org/zap"
)

type AnalyticsHandler struct {
	svc    *service.AnalyticsService
	logger *zap.Logger
}

func NewAnalyticsHandler(svc *service.AnalyticsService, logger *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, logger: logger}
}

// respond is a tiny helper to keep handlers DRY.
func (h *AnalyticsHandler) respond(c *gin.Context, data any, err error) {
	if err != nil {
		h.logger.Error("analytics handler error",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch analytics"})
		return
	}
	if data == nil {
		c.JSON(http.StatusOK, gin.H{"data": []any{}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GET /api/analytics/:code/summary
// Returns total clicks and unique visitors.
// Summary godoc
// @Summary Get URL summary
// @Description Returns total clicks and unique visitors for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/summary [get]
func (h *AnalyticsHandler) Summary(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Summary(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/over-time?interval=hour|day|week
// Returns time-series click data.
// OverTime godoc
// @Summary Get time-series clicks
// @Description Returns time-series click data for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Param interval query string false "Interval (hour, day, week)" default(day)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/over-time [get]
func (h *AnalyticsHandler) OverTime(c *gin.Context) {
	code := c.Param("code")
	interval := c.DefaultQuery("interval", "day")
	data, err := h.svc.OverTime(c.Request.Context(), code, interval)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/countries
// Countries godoc
// @Summary Get country breakdown
// @Description Returns click breakdown by country for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/countries [get]
func (h *AnalyticsHandler) Countries(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Countries(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/cities
// Cities godoc
// @Summary Get city breakdown
// @Description Returns click breakdown by city for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/cities [get]
func (h *AnalyticsHandler) Cities(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Cities(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/devices
// Devices godoc
// @Summary Get device breakdown
// @Description Returns click breakdown by device type for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/devices [get]
func (h *AnalyticsHandler) Devices(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Devices(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/os
// OS godoc
// @Summary Get OS breakdown
// @Description Returns click breakdown by operating system for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/os [get]
func (h *AnalyticsHandler) OS(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.OSBreakdown(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/browsers
// Browsers godoc
// @Summary Get browser breakdown
// @Description Returns click breakdown by browser for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/browsers [get]
func (h *AnalyticsHandler) Browsers(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.Browsers(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/peak-hours
// Returns hourly traffic distribution (0–23).
// PeakHours godoc
// @Summary Get peak hours
// @Description Returns hourly traffic distribution (0–23) for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/peak-hours [get]
func (h *AnalyticsHandler) PeakHours(c *gin.Context) {
	code := c.Param("code")
	data, err := h.svc.PeakHours(c.Request.Context(), code)
	h.respond(c, data, err)
}

// GET /api/analytics/:code/recent?limit=20
// Returns recent click events.
// RecentClicks godoc
// @Summary Get recent clicks
// @Description Returns recent click events for a short code
// @Tags Analytics
// @Accept json
// @Produce json
// @Param code path string true "Short Code"
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /{code}/recent [get]
func (h *AnalyticsHandler) RecentClicks(c *gin.Context) {
	code := c.Param("code")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	data, err := h.svc.RecentClicks(c.Request.Context(), code, limit)
	h.respond(c, data, err)
}
