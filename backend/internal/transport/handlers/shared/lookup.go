package shared

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ParseUUIDParam parses the `id` path parameter and returns a bad request
// response when the value is not a valid UUID.
func ParseUUIDParam(c *gin.Context, entityLabel string) (uuid.UUID, bool) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid " + entityLabel + " id",
		})
		return uuid.UUID{}, false
	}

	return id, true
}

// HandleLookupError sends the appropriate HTTP response for lookup errors.
func HandleLookupError(c *gin.Context, err error, entityLabel string) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": entityLabel + " not found",
		})
		return true
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "failed to retrieve " + entityLabel,
	})
	return true
}

// GetByIDHandler implements a generic UUID lookup flow for context-aware
// service methods.
func GetByIDHandler[T any](c *gin.Context, entityLabel string, lookup func(*gin.Context, string) (*T, error)) {
	id, ok := ParseUUIDParam(c, entityLabel)
	if !ok {
		return
	}

	entity, err := lookup(c, id.String())
	if err != nil {
		if HandleLookupError(c, err, entityLabel) {
			return
		}
	}

	c.JSON(http.StatusOK, entity)
}
