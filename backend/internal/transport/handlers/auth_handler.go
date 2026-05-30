package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	userService  *service.UserService
	tokenService *service.TokenService
	cookieName   string
	cookieSecure bool
}

// LoginRequest represents the expected payload for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(
	userService *service.UserService,
	tokenService *service.TokenService,
	cookieName string,
	cookieSecure bool,
) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		tokenService: tokenService,
		cookieName:   cookieName,
		cookieSecure: cookieSecure,
	}
}

// Login authenticates a user and sets the authentication cookie.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, err := h.userService.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidUserInput) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "email and password are required",
			})
			return
		}

		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})
			return
		}

		slog.Error("failed to authenticate user", "error", err, "email", req.Email)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to authenticate user",
		})
		return
	}

	token, err := h.tokenService.GenerateToken(user.ID, user.Email, user.Username, user.Role)
	if err != nil {
		slog.Error("failed to generate auth token", "error", err, "user_id", user.ID)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate auth token",
		})
		return
	}

	sameSite := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSite = http.SameSiteNoneMode
	}

	//nolint:gosec
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: sameSite,
		MaxAge:   int(h.tokenService.TTL().Seconds()),
	})

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Me returns the current authenticated user from the request context.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	userIDString, _ := userID.(string)
	user, err := h.userService.GetByID(c.Request.Context(), userIDString)
	if err != nil {
		slog.Error("failed to retrieve authenticated user", "error", err, "user_id", userIDString)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve authenticated user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Logout clears the authentication cookie for the current client.
func (h *AuthHandler) Logout(c *gin.Context) {
	sameSite := http.SameSiteLaxMode
	if h.cookieSecure {
		sameSite = http.SameSiteNoneMode
	}

	//nolint:gosec
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: sameSite,
		MaxAge:   -1,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "session closed",
	})
}
