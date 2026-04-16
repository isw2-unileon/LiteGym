package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
	ContextUsernameKey  = "username"
)

type AuthMiddleware struct {
	tokenService *service.TokenService
	cookieName   string
}

func NewAuthMiddleware(tokenService *service.TokenService, cookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		cookieName:   cookieName,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(m.cookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			c.Abort()
			return
		}

		claims, err := m.tokenService.ParseToken(cookie)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired authentication token",
			})
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextUsernameKey, claims.Username)
		c.Next()
	}
}
