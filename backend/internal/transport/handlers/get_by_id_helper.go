package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type getByIDConfig struct {
	invalidIDMessage string
	notFoundMessage  string
	logMessage       string
	internalMessage  string
	invalidInputErr  error
	notFoundErr      error
}

func respondWithResourceByID[T any](
	c *gin.Context,
	getByID func(context.Context, string) (T, error),
	cfg getByIDConfig,
) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": cfg.invalidIDMessage})
		return
	}

	resource, err := getByID(c.Request.Context(), id.String())
	if err != nil {
		if errors.Is(err, cfg.invalidInputErr) {
			c.JSON(http.StatusBadRequest, gin.H{"error": cfg.invalidIDMessage})
			return
		}

		if errors.Is(err, cfg.notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": cfg.notFoundMessage})
			return
		}

		slog.Error(cfg.logMessage, "error", err, "id", id.String())

		c.JSON(http.StatusInternalServerError, gin.H{"error": cfg.internalMessage})
		return
	}

	c.JSON(http.StatusOK, resource)
}
