package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

type AuthHandler struct {
	userService  *service.UserService
	tokenService *service.TokenService
	cookieName   string
	cookieSecure bool
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to authenticate user",
		})
		return
	}

	token, err := h.tokenService.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate auth token",
		})
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.tokenService.TTL().Seconds()),
	})

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
