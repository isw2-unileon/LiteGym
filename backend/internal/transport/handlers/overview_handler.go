package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// OverviewHandler handles overview-related HTTP requests.
type OverviewHandler struct {
	service *service.OverviewService
	nowFunc func() time.Time
}

// NewOverviewHandler creates a new OverviewHandler.
func NewOverviewHandler(svc *service.OverviewService) *OverviewHandler {
	return &OverviewHandler{
		service: svc,
		nowFunc: time.Now,
	}
}

// GetOverview returns the aggregated data required by the dashboard.
func (h *OverviewHandler) GetOverview(c *gin.Context) {
	userIDValue, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, _ := userIDValue.(string)
	statsReferenceDate := h.nowFunc()
	calendarDate := statsReferenceDate
	if month := c.Query("month"); month != "" {
		parsedMonth, parseErr := time.ParseInLocation("2006-01", month, statsReferenceDate.Location())
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
			return
		}
		calendarDate = parsedMonth
	}

	overview, err := h.service.GetOverviewForMonth(c.Request.Context(), userID, calendarDate, statsReferenceDate)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		slog.Error("failed to build dashboard overview", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve dashboard overview"})
		return
	}

	c.JSON(http.StatusOK, overview)
}
