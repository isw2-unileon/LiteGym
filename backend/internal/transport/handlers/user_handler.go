package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{
		service: svc,
	}
}

// CreateUserRequest represents the payload to create a user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username, email and password are required",
		})
		return
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password,
	}

	if err := h.service.Create(c.Request.Context(), user); err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "username, email and password are required",
			})
			return
		}

		slog.Error(
			"failed to create user",
			"error", err,
			"username", req.Username,
			"email", req.Email,
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUserByID retrieves a user by ID.
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id.String())
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user id",
			})
			return
		}

		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		slog.Error("failed to retrieve user", "error", err, "id", id.String())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}
