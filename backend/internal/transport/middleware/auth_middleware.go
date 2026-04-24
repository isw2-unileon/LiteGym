package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
)

// Context keys used by the authentication middleware to store user values in the request context.
//
// These keys are placed into the Gin context by the middleware and can be read by
// handlers to obtain the authenticated user's ID, email and username.
const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
	ContextUsernameKey  = "username"
	ContextUserRoleKey  = "user_role"
)

// AuthMiddleware provides HTTP middleware for authentication.
//
// It uses a TokenService to parse and validate authentication tokens stored in a cookie,
// and injects the parsed claims into the request context for downstream handlers.
type AuthMiddleware struct {
	tokenService *service.TokenService
	cookieName   string
}

// NewAuthMiddleware constructs a new AuthMiddleware.
//
// tokenService is used to parse/verify tokens and cookieName specifies which cookie
// contains the authentication token.
func NewAuthMiddleware(tokenService *service.TokenService, cookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
		cookieName:   cookieName,
	}
}

// RequireAuth returns a Gin middleware handler that enforces authentication.
//
// The returned handler reads the configured cookie, validates the token using the
// TokenService, and on success stores the claims (subject, email, username) in the
// request context.
// It reads the JWT from the cookie, validates it, and injects the user data into the context.
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(m.cookieName)
		if err != nil {
			// If there is no cookie, reject the request immediately
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: missing cookie"})
			return
		}

		claims, err := m.tokenService.ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: invalid or expired token"})
			return
		}

		c.Set(ContextUserIDKey, claims.Subject)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextUsernameKey, claims.Username)

		c.Next()
	}
}
