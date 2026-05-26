package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// ProfileHandler handles HTTP requests related to user profiles.
type ProfileHandler struct {
	profileService *service.ProfileService
}

// NewProfileHandler creates a new instance of ProfileHandler.
func NewProfileHandler(ps *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: ps}
}

// GetDashboard returns the aggregated statistics and data for the user profile.
func (h *ProfileHandler) GetDashboard(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	timeRange := c.DefaultQuery("range", "month")

	// Parse optional year/month for exact calendar-month filtering
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if y, err := strconv.Atoi(c.Query("year")); err == nil && y > 2000 {
		year = y
	}
	if m, err := strconv.Atoi(c.Query("month")); err == nil && m >= 1 && m <= 12 {
		month = m
	}

	stats, err := h.profileService.GetDashboardStats(c.Request.Context(), userIDVal.(string), timeRange, year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch profile stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateGoalsRequest represents the incoming payload for updating goals.
type UpdateGoalsRequest struct {
	ShortTerm  string `json:"short_term"`
	LongTerm   string `json:"long_term"`
	TargetDays int    `json:"target_days"`
}

// UpdateGoals handles the request to update a user's fitness goals.
func (h *ProfileHandler) UpdateGoals(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateGoalsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userUUID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	goal := &model.UserGoal{
		UserID:            userUUID,
		ShortTerm:         req.ShortTerm,
		LongTerm:          req.LongTerm,
		TargetDaysPerWeek: req.TargetDays,
	}

	if err := h.profileService.UpdateGoals(c.Request.Context(), goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update goals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "goals updated successfully"})
}

// AddBodyMetric handles the request to log a new body metric entry.
func (h *ProfileHandler) AddBodyMetric(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.AddBodyMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if err := h.profileService.AddBodyMetric(c.Request.Context(), userIDVal.(string), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add body metric"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "body metric added successfully"})
}
