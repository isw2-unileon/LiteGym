package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
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
	Role     string `json:"role"`
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

	if req.Username == "" || req.Email == "" || req.Password == "" || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username, email and password are required",
		})
		return
	}

	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: req.Password,
		Role:         req.Role,
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
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), id)
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

		slog.Error("failed to retrieve user", "error", err, "id", id)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve user",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetMe returns the authenticated user profile
func (h *UserHandler) GetMe(c *gin.Context) {
	userIDVal, exists := c.Get(middleware.ContextUserIDKey)

	if !exists {
		userIDVal, exists = c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
	}

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid id format in context"})
		return
	}

	user, err := h.service.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found in db"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListAllUsers handles the request to retrieve all registered users.
// This endpoint is meant to be consumed by the Admin dashboard.
func (h *UserHandler) ListAllUsers(c *gin.Context) {
	users, err := h.service.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve users list",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// DeleteUser handles the request to delete a specific user.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}
